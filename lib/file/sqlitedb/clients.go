package sqlitedb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/rate"
)

// ----- column layout helpers ---------------------------------------------

const clientColumns = `id, verify_key, addr, remark, status,
	rate_limit, flow_limit, inlet_flow, export_flow,
	max_conn, max_tunnel_num, web_user_name, web_password,
	config_conn_allow, no_store, no_display,
	ip_white, ip_white_pass, ip_white_list, black_ip_list,
	cnf_u, cnf_p, cnf_compress, cnf_crypt,
	create_time, last_online_time`

// 26 columns => 26 placeholders for INSERT
const clientPlaceholders = `?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?`

// clientArgs flattens a *file.Client into the positional arguments that
// match clientColumns. Used by both INSERT (UPSERT) and the backfill path.
func clientArgs(c *file.Client) []any {
	bl, _ := json.Marshal(c.BlackIpList)
	wl, _ := json.Marshal(c.IpWhiteList)
	var (
		inlet, export, flowLimit int64
		cnfU, cnfP               string
		cnfCompress, cnfCrypt    int
	)
	if c.Flow != nil {
		inlet = c.Flow.InletFlow
		export = c.Flow.ExportFlow
		flowLimit = c.Flow.FlowLimit
	}
	if c.Cnf != nil {
		cnfU = c.Cnf.U
		cnfP = c.Cnf.P
		if c.Cnf.Compress {
			cnfCompress = 1
		}
		if c.Cnf.Crypt {
			cnfCrypt = 1
		}
	}
	return []any{
		c.Id, c.VerifyKey, c.Addr, c.Remark, boolInt(c.Status),
		c.RateLimit, flowLimit, inlet, export,
		c.MaxConn, c.MaxTunnelNum, c.WebUserName, c.WebPassword,
		boolInt(c.ConfigConnAllow), boolInt(c.NoStore), boolInt(c.NoDisplay),
		boolInt(c.IpWhite), c.IpWhitePass, string(wl), string(bl),
		cnfU, cnfP, cnfCompress, cnfCrypt,
		c.CreateTime, c.LastOnlineTime,
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// scanClient reads a row produced by SELECT clientColumns into a fresh
// *file.Client. Runtime-only fields (IsConnect/NowConn/Rate/Version)
// stay at their zero values; the caller is responsible for spinning up
// a Rate limiter.
func scanClient(scan func(...any) error) (*file.Client, error) {
	c := &file.Client{Cnf: &file.Config{}, Flow: &file.Flow{}}
	var (
		statusInt, configConnInt, noStoreInt, noDisplayInt int
		ipWhiteInt, cnfCompressInt, cnfCryptInt            int
		ipWhiteListJSON, blackIpListJSON                   string
	)
	if err := scan(
		&c.Id, &c.VerifyKey, &c.Addr, &c.Remark, &statusInt,
		&c.RateLimit, &c.Flow.FlowLimit, &c.Flow.InletFlow, &c.Flow.ExportFlow,
		&c.MaxConn, &c.MaxTunnelNum, &c.WebUserName, &c.WebPassword,
		&configConnInt, &noStoreInt, &noDisplayInt,
		&ipWhiteInt, &c.IpWhitePass, &ipWhiteListJSON, &blackIpListJSON,
		&c.Cnf.U, &c.Cnf.P, &cnfCompressInt, &cnfCryptInt,
		&c.CreateTime, &c.LastOnlineTime,
	); err != nil {
		return nil, err
	}
	c.Status = statusInt != 0
	c.ConfigConnAllow = configConnInt != 0
	c.NoStore = noStoreInt != 0
	c.NoDisplay = noDisplayInt != 0
	c.IpWhite = ipWhiteInt != 0
	c.Cnf.Compress = cnfCompressInt != 0
	c.Cnf.Crypt = cnfCryptInt != 0
	if ipWhiteListJSON != "" {
		_ = json.Unmarshal([]byte(ipWhiteListJSON), &c.IpWhiteList)
	}
	if blackIpListJSON != "" {
		_ = json.Unmarshal([]byte(blackIpListJSON), &c.BlackIpList)
	}
	return c, nil
}

// ----- ClientPersister implementation ------------------------------------

// UpsertClient implements file.ClientPersister. Uses INSERT ... ON
// CONFLICT(id) DO UPDATE so callers don't have to distinguish create vs
// update paths.
func (s *Store) UpsertClient(c *file.Client) error {
	if c == nil || c.Id == 0 {
		return errors.New("upsert client: id is zero")
	}
	args := clientArgs(c)
	_, err := s.Writer.ExecContext(context.Background(),
		`INSERT INTO clients(`+clientColumns+`) VALUES(`+clientPlaceholders+`)
		 ON CONFLICT(id) DO UPDATE SET
			verify_key=excluded.verify_key,
			addr=excluded.addr,
			remark=excluded.remark,
			status=excluded.status,
			rate_limit=excluded.rate_limit,
			flow_limit=excluded.flow_limit,
			inlet_flow=excluded.inlet_flow,
			export_flow=excluded.export_flow,
			max_conn=excluded.max_conn,
			max_tunnel_num=excluded.max_tunnel_num,
			web_user_name=excluded.web_user_name,
			web_password=excluded.web_password,
			config_conn_allow=excluded.config_conn_allow,
			no_store=excluded.no_store,
			no_display=excluded.no_display,
			ip_white=excluded.ip_white,
			ip_white_pass=excluded.ip_white_pass,
			ip_white_list=excluded.ip_white_list,
			black_ip_list=excluded.black_ip_list,
			cnf_u=excluded.cnf_u,
			cnf_p=excluded.cnf_p,
			cnf_compress=excluded.cnf_compress,
			cnf_crypt=excluded.cnf_crypt,
			create_time=excluded.create_time,
			last_online_time=excluded.last_online_time`,
		args...)
	if err != nil {
		return fmt.Errorf("upsert client id=%d: %w", c.Id, err)
	}
	return nil
}

// DeleteClient implements file.ClientPersister.
func (s *Store) DeleteClient(id int) error {
	_, err := s.Writer.ExecContext(context.Background(),
		`DELETE FROM clients WHERE id=?`, id)
	return err
}

// ----- one-shot bootstrap ------------------------------------------------

// BackfillOrLoadClients reconciles the in-memory sync.Map with the
// SQLite table at server startup. Two modes:
//
//  1. SQLite empty AND sync.Map non-empty (first boot after upgrade) ->
//     INSERT every Client from sync.Map into SQLite; sync.Map kept as-is.
//  2. SQLite non-empty (steady state) -> CLEAR sync.Map and rehydrate
//     it from SQLite (SQLite is now the source of truth).
//
// In both cases ClientIncreaseId is reset to MAX(id) so subsequent
// auto-assigned IDs don't collide.
//
// Returns (mode, count) where mode is one of "backfill"/"load"/"empty".
func (s *Store) BackfillOrLoadClients(jdb *file.JsonDb) (string, int, error) {
	var sqlCount int
	if err := s.Reader.QueryRowContext(context.Background(),
		`SELECT COUNT(1) FROM clients`).Scan(&sqlCount); err != nil {
		return "", 0, fmt.Errorf("count clients: %w", err)
	}

	if sqlCount == 0 {
		// Mode 1: first boot — push the JSON-loaded map into SQLite.
		var memCount int
		jdb.Clients.Range(func(_, _ any) bool { memCount++; return true })
		if memCount == 0 {
			return "empty", 0, nil
		}
		tx, err := s.Writer.BeginTx(context.Background(), nil)
		if err != nil {
			return "", 0, err
		}
		defer func() { _ = tx.Rollback() }()

		stmt, err := tx.Prepare(`INSERT INTO clients(` + clientColumns + `) VALUES(` + clientPlaceholders + `)`)
		if err != nil {
			return "", 0, err
		}
		defer stmt.Close()

		var (
			inserted int
			maxID    int32
			rangeErr error
		)
		jdb.Clients.Range(func(_, value any) bool {
			c, _ := value.(*file.Client)
			if c == nil || c.NoStore {
				return true
			}
			if _, err := stmt.Exec(clientArgs(c)...); err != nil {
				rangeErr = fmt.Errorf("backfill client id=%d: %w", c.Id, err)
				return false
			}
			inserted++
			if int32(c.Id) > maxID {
				maxID = int32(c.Id)
			}
			return true
		})
		if rangeErr != nil {
			return "", inserted, rangeErr
		}
		if err := tx.Commit(); err != nil {
			return "", inserted, err
		}
		atomic.StoreInt32(&jdb.ClientIncreaseId, maxID)
		return "backfill", inserted, nil
	}

	// Mode 2: load SQLite into memory (replacing whatever JSON loaded).
	rows, err := s.Reader.QueryContext(context.Background(),
		`SELECT `+clientColumns+` FROM clients ORDER BY id ASC`)
	if err != nil {
		return "", 0, fmt.Errorf("load clients: %w", err)
	}
	defer rows.Close()

	// Wipe map first so a stale JSON entry that was deleted in SQL does
	// not linger in memory.
	jdb.Clients.Range(func(k, _ any) bool { jdb.Clients.Delete(k); return true })

	var (
		loaded int
		maxID  int32
	)
	for rows.Next() {
		c, err := scanClient(rows.Scan)
		if err != nil {
			return "", loaded, err
		}
		// Recreate the runtime rate limiter the same way
		// LoadClientFromJsonFile does, so behaviour is identical.
		if c.RateLimit > 0 {
			c.Rate = rate.NewRate(int64(c.RateLimit * 1024))
		} else {
			c.Rate = rate.NewRate((2 << 23) * 1024)
		}
		c.Rate.Start()
		c.NowConn = 0
		c.IsConnect = false

		jdb.Clients.Store(c.Id, c)
		if int32(c.Id) > maxID {
			maxID = int32(c.Id)
		}
		loaded++
	}
	if err := rows.Err(); err != nil {
		return "", loaded, err
	}
	atomic.StoreInt32(&jdb.ClientIncreaseId, maxID)
	return "load", loaded, nil
}

// Compile-time interface check.
var _ file.ClientPersister = (*Store)(nil)

// avoid unused import if all helpers are removed; sql is referenced by
// other files in this package but keep this anchor for readability.
var _ = sql.ErrNoRows
