package sqlitedb

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"ehang.io/nps/lib/file"
)

// ----- column layout helpers ---------------------------------------------

const hostColumns = `id, client_id, host, header_change, host_change, location, remark,
	scheme, cert_file_path, key_file_path,
	no_store, is_close, auto_https,
	inlet_flow, export_flow, flow_limit,
	target_str, target_local_proxy,
	health_check_timeout, health_max_fail, health_check_interval,
	http_health_url, health_check_type, health_check_target`

// 24 columns => 24 placeholders.
const hostPlaceholders = `?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?`

func hostArgs(h *file.Host) []any {
	var (
		clientId                                int
		inlet, export, flowLimit                int64
		targetStr                               string
		targetLocalProxyInt                     int
	)
	if h.Client != nil {
		clientId = h.Client.Id
	}
	if h.Flow != nil {
		inlet = h.Flow.InletFlow
		export = h.Flow.ExportFlow
		flowLimit = h.Flow.FlowLimit
	}
	if h.Target != nil {
		targetStr = h.Target.TargetStr
		if h.Target.LocalProxy {
			targetLocalProxyInt = 1
		}
	}
	return []any{
		h.Id, clientId, h.Host, h.HeaderChange, h.HostChange, h.Location, h.Remark,
		h.Scheme, h.CertFilePath, h.KeyFilePath,
		boolInt(h.NoStore), boolInt(h.IsClose), boolInt(h.AutoHttps),
		inlet, export, flowLimit,
		targetStr, targetLocalProxyInt,
		h.HealthCheckTimeout, h.HealthMaxFail, h.HealthCheckInterval,
		h.HttpHealthUrl, h.HealthCheckType, h.HealthCheckTarget,
	}
}

func scanHost(scan func(...any) error) (*file.Host, int, error) {
	h := &file.Host{
		Flow:   &file.Flow{},
		Target: &file.Target{},
	}
	var (
		clientId                                            int
		noStoreInt, isCloseInt, autoHttpsInt, targetLocalInt int
	)
	if err := scan(
		&h.Id, &clientId, &h.Host, &h.HeaderChange, &h.HostChange, &h.Location, &h.Remark,
		&h.Scheme, &h.CertFilePath, &h.KeyFilePath,
		&noStoreInt, &isCloseInt, &autoHttpsInt,
		&h.Flow.InletFlow, &h.Flow.ExportFlow, &h.Flow.FlowLimit,
		&h.Target.TargetStr, &targetLocalInt,
		&h.HealthCheckTimeout, &h.HealthMaxFail, &h.HealthCheckInterval,
		&h.HttpHealthUrl, &h.HealthCheckType, &h.HealthCheckTarget,
	); err != nil {
		return nil, 0, err
	}
	h.NoStore = noStoreInt != 0
	h.IsClose = isCloseInt != 0
	h.AutoHttps = autoHttpsInt != 0
	h.Target.LocalProxy = targetLocalInt != 0
	return h, clientId, nil
}

// ----- HostPersister implementation --------------------------------------

func (s *Store) UpsertHost(h *file.Host) error {
	if h == nil || h.Id == 0 {
		return errors.New("upsert host: id is zero")
	}
	args := hostArgs(h)
	_, err := s.Writer.ExecContext(context.Background(),
		`INSERT INTO hosts(`+hostColumns+`) VALUES(`+hostPlaceholders+`)
		 ON CONFLICT(id) DO UPDATE SET
			client_id=excluded.client_id,
			host=excluded.host,
			header_change=excluded.header_change,
			host_change=excluded.host_change,
			location=excluded.location,
			remark=excluded.remark,
			scheme=excluded.scheme,
			cert_file_path=excluded.cert_file_path,
			key_file_path=excluded.key_file_path,
			no_store=excluded.no_store,
			is_close=excluded.is_close,
			auto_https=excluded.auto_https,
			inlet_flow=excluded.inlet_flow,
			export_flow=excluded.export_flow,
			flow_limit=excluded.flow_limit,
			target_str=excluded.target_str,
			target_local_proxy=excluded.target_local_proxy,
			health_check_timeout=excluded.health_check_timeout,
			health_max_fail=excluded.health_max_fail,
			health_check_interval=excluded.health_check_interval,
			http_health_url=excluded.http_health_url,
			health_check_type=excluded.health_check_type,
			health_check_target=excluded.health_check_target`,
		args...)
	if err != nil {
		return fmt.Errorf("upsert host id=%d: %w", h.Id, err)
	}
	return nil
}

func (s *Store) DeleteHost(id int) error {
	_, err := s.Writer.ExecContext(context.Background(),
		`DELETE FROM hosts WHERE id=?`, id)
	return err
}

// ----- one-shot bootstrap ------------------------------------------------

// BackfillOrLoadHosts has the same two-mode contract as the clients/tasks
// variants. MUST run AFTER BackfillOrLoadClients so client_id lookups
// hit a populated Clients map.
//
// Returns (mode, count) where mode is one of "backfill"/"load"/"empty".
func (s *Store) BackfillOrLoadHosts(jdb *file.JsonDb) (string, int, error) {
	var sqlCount int
	if err := s.Reader.QueryRowContext(context.Background(),
		`SELECT COUNT(1) FROM hosts`).Scan(&sqlCount); err != nil {
		return "", 0, fmt.Errorf("count hosts: %w", err)
	}

	if sqlCount == 0 {
		var memCount int
		jdb.Hosts.Range(func(_, _ any) bool { memCount++; return true })
		if memCount == 0 {
			return "empty", 0, nil
		}
		tx, err := s.Writer.BeginTx(context.Background(), nil)
		if err != nil {
			return "", 0, err
		}
		defer func() { _ = tx.Rollback() }()

		stmt, err := tx.Prepare(`INSERT INTO hosts(` + hostColumns + `) VALUES(` + hostPlaceholders + `)`)
		if err != nil {
			return "", 0, err
		}
		defer stmt.Close()

		var (
			inserted int
			maxID    int32
			rangeErr error
		)
		jdb.Hosts.Range(func(_, value any) bool {
			h, _ := value.(*file.Host)
			if h == nil || h.NoStore {
				return true
			}
			if _, err := stmt.Exec(hostArgs(h)...); err != nil {
				rangeErr = fmt.Errorf("backfill host id=%d: %w", h.Id, err)
				return false
			}
			inserted++
			if int32(h.Id) > maxID {
				maxID = int32(h.Id)
			}
			return true
		})
		if rangeErr != nil {
			return "", inserted, rangeErr
		}
		if err := tx.Commit(); err != nil {
			return "", inserted, err
		}
		atomic.StoreInt32(&jdb.HostIncreaseId, maxID)
		return "backfill", inserted, nil
	}

	rows, err := s.Reader.QueryContext(context.Background(),
		`SELECT `+hostColumns+` FROM hosts ORDER BY id ASC`)
	if err != nil {
		return "", 0, fmt.Errorf("load hosts: %w", err)
	}
	defer rows.Close()

	jdb.Hosts.Range(func(k, _ any) bool { jdb.Hosts.Delete(k); return true })

	var (
		loaded int
		maxID  int32
	)
	for rows.Next() {
		h, clientId, err := scanHost(rows.Scan)
		if err != nil {
			return "", loaded, err
		}
		c, gerr := jdb.GetClient(clientId)
		if gerr != nil {
			continue
		}
		h.Client = c
		jdb.Hosts.Store(h.Id, h)
		if int32(h.Id) > maxID {
			maxID = int32(h.Id)
		}
		loaded++
	}
	if err := rows.Err(); err != nil {
		return "", loaded, err
	}
	atomic.StoreInt32(&jdb.HostIncreaseId, maxID)
	return "load", loaded, nil
}

// Compile-time interface check.
var _ file.HostPersister = (*Store)(nil)
