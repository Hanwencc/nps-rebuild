package sqlitedb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"ehang.io/nps/lib/file"
)

// ----- column layout helpers ---------------------------------------------

const taskColumns = `id, client_id, mode, port, server_ip, status,
	ports, password, remark, target_addr, local_path, strip_pre, proto_version,
	no_store, inlet_flow, export_flow, flow_limit,
	target_str, target_local_proxy, multi_account_map,
	health_check_timeout, health_max_fail, health_check_interval,
	http_health_url, health_check_type, health_check_target`

// 26 columns => 26 placeholders.
const taskPlaceholders = `?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?`

// taskArgs flattens a *file.Tunnel into the positional arguments matching
// taskColumns. Used by both UPSERT and the backfill path.
func taskArgs(t *file.Tunnel) []any {
	var (
		clientId                              int
		inlet, export, flowLimit              int64
		targetStr                             string
		targetLocalProxyInt                   int
		multiAccountJSON                      = "{}"
	)
	if t.Client != nil {
		clientId = t.Client.Id
	}
	if t.Flow != nil {
		inlet = t.Flow.InletFlow
		export = t.Flow.ExportFlow
		flowLimit = t.Flow.FlowLimit
	}
	if t.Target != nil {
		targetStr = t.Target.TargetStr
		if t.Target.LocalProxy {
			targetLocalProxyInt = 1
		}
	}
	if t.MultiAccount != nil && t.MultiAccount.AccountMap != nil {
		if b, err := json.Marshal(t.MultiAccount.AccountMap); err == nil {
			multiAccountJSON = string(b)
		}
	}
	return []any{
		t.Id, clientId, t.Mode, t.Port, t.ServerIp, boolInt(t.Status),
		t.Ports, t.Password, t.Remark, t.TargetAddr, t.LocalPath, t.StripPre, t.ProtoVersion,
		boolInt(t.NoStore), inlet, export, flowLimit,
		targetStr, targetLocalProxyInt, multiAccountJSON,
		t.HealthCheckTimeout, t.HealthMaxFail, t.HealthCheckInterval,
		t.HttpHealthUrl, t.HealthCheckType, t.HealthCheckTarget,
	}
}

// scanTask reads a row into a fresh *file.Tunnel. The Client pointer is
// resolved from the in-memory map by the caller (BackfillOrLoadTasks)
// because lib/file.Tunnel.Client is a *Client, not a numeric FK, and
// cross-table joining would duplicate state already loaded from the
// clients table.
func scanTask(scan func(...any) error) (*file.Tunnel, int, error) {
	t := &file.Tunnel{
		Flow:         &file.Flow{},
		Target:       &file.Target{},
		MultiAccount: &file.MultiAccount{},
	}
	var (
		clientId                                     int
		statusInt, noStoreInt, targetLocalProxyInt   int
		multiAccountJSON                             string
	)
	if err := scan(
		&t.Id, &clientId, &t.Mode, &t.Port, &t.ServerIp, &statusInt,
		&t.Ports, &t.Password, &t.Remark, &t.TargetAddr, &t.LocalPath, &t.StripPre, &t.ProtoVersion,
		&noStoreInt, &t.Flow.InletFlow, &t.Flow.ExportFlow, &t.Flow.FlowLimit,
		&t.Target.TargetStr, &targetLocalProxyInt, &multiAccountJSON,
		&t.HealthCheckTimeout, &t.HealthMaxFail, &t.HealthCheckInterval,
		&t.HttpHealthUrl, &t.HealthCheckType, &t.HealthCheckTarget,
	); err != nil {
		return nil, 0, err
	}
	t.Status = statusInt != 0
	t.NoStore = noStoreInt != 0
	t.Target.LocalProxy = targetLocalProxyInt != 0
	if multiAccountJSON != "" && multiAccountJSON != "{}" {
		_ = json.Unmarshal([]byte(multiAccountJSON), &t.MultiAccount.AccountMap)
	}
	return t, clientId, nil
}

// ----- TaskPersister implementation --------------------------------------

func (s *Store) UpsertTask(t *file.Tunnel) error {
	if t == nil || t.Id == 0 {
		return errors.New("upsert task: id is zero")
	}
	args := taskArgs(t)
	_, err := s.Writer.ExecContext(context.Background(),
		`INSERT INTO tasks(`+taskColumns+`) VALUES(`+taskPlaceholders+`)
		 ON CONFLICT(id) DO UPDATE SET
			client_id=excluded.client_id,
			mode=excluded.mode,
			port=excluded.port,
			server_ip=excluded.server_ip,
			status=excluded.status,
			ports=excluded.ports,
			password=excluded.password,
			remark=excluded.remark,
			target_addr=excluded.target_addr,
			local_path=excluded.local_path,
			strip_pre=excluded.strip_pre,
			proto_version=excluded.proto_version,
			no_store=excluded.no_store,
			inlet_flow=excluded.inlet_flow,
			export_flow=excluded.export_flow,
			flow_limit=excluded.flow_limit,
			target_str=excluded.target_str,
			target_local_proxy=excluded.target_local_proxy,
			multi_account_map=excluded.multi_account_map,
			health_check_timeout=excluded.health_check_timeout,
			health_max_fail=excluded.health_max_fail,
			health_check_interval=excluded.health_check_interval,
			http_health_url=excluded.http_health_url,
			health_check_type=excluded.health_check_type,
			health_check_target=excluded.health_check_target`,
		args...)
	if err != nil {
		return fmt.Errorf("upsert task id=%d: %w", t.Id, err)
	}
	return nil
}

func (s *Store) DeleteTask(id int) error {
	_, err := s.Writer.ExecContext(context.Background(),
		`DELETE FROM tasks WHERE id=?`, id)
	return err
}

// ----- one-shot bootstrap ------------------------------------------------

// BackfillOrLoadTasks reconciles the in-memory Tasks sync.Map with the
// SQLite table. Same two-mode contract as BackfillOrLoadClients.
//
// MUST run AFTER BackfillOrLoadClients so client_id lookups hit a
// populated Clients map.
//
// Returns (mode, count) where mode is one of "backfill"/"load"/"empty".
func (s *Store) BackfillOrLoadTasks(jdb *file.JsonDb) (string, int, error) {
	var sqlCount int
	if err := s.Reader.QueryRowContext(context.Background(),
		`SELECT COUNT(1) FROM tasks`).Scan(&sqlCount); err != nil {
		return "", 0, fmt.Errorf("count tasks: %w", err)
	}

	if sqlCount == 0 {
		// Mode 1: first boot — push the JSON-loaded map into SQLite.
		var memCount int
		jdb.Tasks.Range(func(_, _ any) bool { memCount++; return true })
		if memCount == 0 {
			return "empty", 0, nil
		}
		tx, err := s.Writer.BeginTx(context.Background(), nil)
		if err != nil {
			return "", 0, err
		}
		defer func() { _ = tx.Rollback() }()

		stmt, err := tx.Prepare(`INSERT INTO tasks(` + taskColumns + `) VALUES(` + taskPlaceholders + `)`)
		if err != nil {
			return "", 0, err
		}
		defer stmt.Close()

		var (
			inserted int
			maxID    int32
			rangeErr error
		)
		jdb.Tasks.Range(func(_, value any) bool {
			t, _ := value.(*file.Tunnel)
			if t == nil || t.NoStore {
				return true
			}
			if _, err := stmt.Exec(taskArgs(t)...); err != nil {
				rangeErr = fmt.Errorf("backfill task id=%d: %w", t.Id, err)
				return false
			}
			inserted++
			if int32(t.Id) > maxID {
				maxID = int32(t.Id)
			}
			return true
		})
		if rangeErr != nil {
			return "", inserted, rangeErr
		}
		if err := tx.Commit(); err != nil {
			return "", inserted, err
		}
		atomic.StoreInt32(&jdb.TaskIncreaseId, maxID)
		return "backfill", inserted, nil
	}

	// Mode 2: load SQLite into memory (replacing whatever JSON loaded).
	rows, err := s.Reader.QueryContext(context.Background(),
		`SELECT `+taskColumns+` FROM tasks ORDER BY id ASC`)
	if err != nil {
		return "", 0, fmt.Errorf("load tasks: %w", err)
	}
	defer rows.Close()

	// Wipe map first so a stale JSON entry deleted in SQL does not
	// linger in memory.
	jdb.Tasks.Range(func(k, _ any) bool { jdb.Tasks.Delete(k); return true })

	var (
		loaded int
		maxID  int32
	)
	for rows.Next() {
		t, clientId, err := scanTask(rows.Scan)
		if err != nil {
			return "", loaded, err
		}
		// Resolve Client pointer from the in-memory map. If the
		// referenced client has been deleted, drop the orphaned task
		// (matches LoadTaskFromJsonFile behaviour, which also skips on
		// GetClient error).
		c, gerr := jdb.GetClient(clientId)
		if gerr != nil {
			continue
		}
		t.Client = c
		jdb.Tasks.Store(t.Id, t)
		if int32(t.Id) > maxID {
			maxID = int32(t.Id)
		}
		loaded++
	}
	if err := rows.Err(); err != nil {
		return "", loaded, err
	}
	atomic.StoreInt32(&jdb.TaskIncreaseId, maxID)
	return "load", loaded, nil
}

// Compile-time interface check.
var _ file.TaskPersister = (*Store)(nil)
