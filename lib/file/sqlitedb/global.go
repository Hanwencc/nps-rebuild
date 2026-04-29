package sqlitedb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"ehang.io/nps/lib/file"
)

// global is a single-row table (id=1).

// UpsertGlobal replaces the singleton row. The Glob's RWMutex is held by
// callers (see web/api), but to be safe we read fields under its RLock.
func (s *Store) UpsertGlobal(g *file.Glob) error {
	if g == nil {
		return errors.New("upsert global: nil glob")
	}
	g.RLock()
	serverURL := g.ServerUrl
	list := g.BlackIpList
	g.RUnlock()
	if list == nil {
		list = []string{}
	}
	listJSON, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("marshal black_ip_list: %w", err)
	}
	_, err = s.Writer.ExecContext(context.Background(),
		`INSERT INTO global(id, server_url, black_ip_list) VALUES(1, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			server_url=excluded.server_url,
			black_ip_list=excluded.black_ip_list`,
		serverURL, string(listJSON))
	if err != nil {
		return fmt.Errorf("upsert global: %w", err)
	}
	return nil
}

// BackfillOrLoadGlobal mirrors the contract of the other tables. Returns
// one of "backfill"/"load"/"empty" plus a count (0 or 1).
//
// "empty" means SQLite has no row AND in-memory Global is nil — caller
// should rely on whatever default the rest of the app picks.
func (s *Store) BackfillOrLoadGlobal(jdb *file.JsonDb) (string, int, error) {
	var (
		serverURL  sql.NullString
		listJSON   sql.NullString
	)
	row := s.Reader.QueryRowContext(context.Background(),
		`SELECT server_url, black_ip_list FROM global WHERE id=1`)
	switch err := row.Scan(&serverURL, &listJSON); err {
	case nil:
		// SQL has the row → load into memory.
		g := &file.Glob{ServerUrl: serverURL.String}
		if listJSON.Valid && listJSON.String != "" {
			if err := json.Unmarshal([]byte(listJSON.String), &g.BlackIpList); err != nil {
				return "", 0, fmt.Errorf("unmarshal black_ip_list: %w", err)
			}
		}
		jdb.Global = g
		return "load", 1, nil
	case sql.ErrNoRows:
		// No SQL row → backfill from memory if anything is there.
		if jdb.Global == nil {
			return "empty", 0, nil
		}
		if err := s.UpsertGlobal(jdb.Global); err != nil {
			return "", 0, err
		}
		return "backfill", 1, nil
	default:
		return "", 0, fmt.Errorf("read global: %w", err)
	}
}

// Compile-time interface check.
var _ file.GlobalPersister = (*Store)(nil)
