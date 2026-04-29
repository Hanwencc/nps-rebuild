package sqlitedb

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"ehang.io/nps/lib/file"
	"github.com/astaxie/beego"
)

// Bootstrap keys live in nps.conf (and ONLY there) because the web
// server has to listen on web_port and authenticate web_username /
// web_password BEFORE SQLite is opened. Anything else is mirrored into
// app_settings and is editable at runtime via the admin UI.
var bootstrapKeys = map[string]struct{}{
	"web_port":     {},
	"web_username": {},
	"web_password": {},
}

// IsBootstrapKey reports whether a config key must stay in nps.conf.
func IsBootstrapKey(k string) bool {
	_, ok := bootstrapKeys[k]
	return ok
}

// ----- hot-apply registry -----------------------------------------------

// SettingChange callbacks fire AFTER a successful UpsertSetting and AFTER
// beego.AppConfig has been updated, so listeners can re-read via the
// usual beego.AppConfig.{String,Int,Bool} APIs.
type SettingChange func(key, oldVal, newVal string)

var (
	hookMu     sync.RWMutex
	keyHooks   = map[string][]SettingChange{}
	wildHooks  []SettingChange
)

// OnSettingChange registers a per-key listener. Multiple listeners may
// be registered for the same key; they fire in registration order.
// An empty key registers a wildcard listener that fires for every change.
func OnSettingChange(key string, fn SettingChange) {
	if fn == nil {
		return
	}
	hookMu.Lock()
	defer hookMu.Unlock()
	if key == "" {
		wildHooks = append(wildHooks, fn)
		return
	}
	keyHooks[key] = append(keyHooks[key], fn)
}

func fireHooks(key, oldVal, newVal string) {
	hookMu.RLock()
	keyed := append([]SettingChange(nil), keyHooks[key]...)
	wild := append([]SettingChange(nil), wildHooks...)
	hookMu.RUnlock()
	for _, fn := range keyed {
		safeFire(fn, key, oldVal, newVal)
	}
	for _, fn := range wild {
		safeFire(fn, key, oldVal, newVal)
	}
}

func safeFire(fn SettingChange, key, oldVal, newVal string) {
	defer func() { _ = recover() }()
	fn(key, oldVal, newVal)
}

// ----- store API --------------------------------------------------------

// ListAppSettings returns every persisted setting (excluding bootstrap
// keys, which never live in SQLite).
func (s *Store) ListAppSettings() (map[string]string, error) {
	rows, err := s.Reader.QueryContext(context.Background(),
		`SELECT key, value FROM app_settings ORDER BY key`)
	if err != nil {
		return nil, fmt.Errorf("list app_settings: %w", err)
	}
	defer rows.Close()
	out := make(map[string]string, 64)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// UpsertSetting persists (key,value), pushes the new value into the
// in-memory beego.AppConfig so existing call sites see it on next read,
// and fires registered hot-apply hooks. Bootstrap keys are rejected.
func (s *Store) UpsertSetting(key, value string) error {
	if key == "" {
		return fmt.Errorf("upsert setting: empty key")
	}
	if IsBootstrapKey(key) {
		return fmt.Errorf("upsert setting: %q is a bootstrap key (edit nps.conf and restart)", key)
	}
	old := beego.AppConfig.String(key)
	now := time.Now().Unix()
	if _, err := s.Writer.ExecContext(context.Background(),
		`INSERT INTO app_settings(key, value, updated_at) VALUES(?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET
			value=excluded.value,
			updated_at=excluded.updated_at`,
		key, value, now); err != nil {
		return fmt.Errorf("upsert setting %q: %w", key, err)
	}
	if err := beego.AppConfig.Set(key, value); err != nil {
		// Persisted but in-memory cache is stale — log via hook caller.
		return fmt.Errorf("apply setting %q to beego: %w", key, err)
	}
	if old != value {
		fireHooks(key, old, value)
	}
	return nil
}

// UpsertSettings applies a batch of settings inside a single transaction
// for the SQL writes; beego.AppConfig.Set + hooks fire per-key after the
// commit. Returns the keys that were rejected (bootstrap or empty) along
// with the count of successful writes.
func (s *Store) UpsertSettings(kv map[string]string) (applied int, rejected []string, err error) {
	if len(kv) == 0 {
		return 0, nil, nil
	}
	// Sort for deterministic ordering (test stability).
	keys := make([]string, 0, len(kv))
	for k := range kv {
		if k == "" || IsBootstrapKey(k) {
			rejected = append(rejected, k)
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tx, err := s.Writer.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, rejected, err
	}
	defer func() { _ = tx.Rollback() }()
	stmt, err := tx.Prepare(
		`INSERT INTO app_settings(key, value, updated_at) VALUES(?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`)
	if err != nil {
		return 0, rejected, err
	}
	defer stmt.Close()
	now := time.Now().Unix()
	type pending struct{ key, oldVal, newVal string }
	var fires []pending
	for _, k := range keys {
		v := kv[k]
		old := beego.AppConfig.String(k)
		if _, err := stmt.Exec(k, v, now); err != nil {
			return applied, rejected, fmt.Errorf("upsert setting %q: %w", k, err)
		}
		applied++
		if old != v {
			fires = append(fires, pending{k, old, v})
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, rejected, err
	}
	// Apply to in-memory beego.AppConfig + fire hooks AFTER commit.
	for _, p := range fires {
		_ = beego.AppConfig.Set(p.key, p.newVal)
	}
	for _, p := range fires {
		fireHooks(p.key, p.oldVal, p.newVal)
	}
	return applied, rejected, nil
}

// BackfillOrLoadAppSettings is the startup reconciliation step. Same
// contract as the other tables.
//
//   - SQLite empty → enumerate beego.AppConfig (which was just loaded
//     from nps.conf) and import every non-bootstrap key into SQLite.
//     mode = "backfill", count = inserted rows.
//   - SQLite non-empty → SELECT all rows and push each value into
//     beego.AppConfig via Set, OVERRIDING what nps.conf provided.
//     mode = "load", count = applied rows.
//
// Must run AFTER sqlitedb.Open() and AFTER beego.LoadAppConfig().
func (s *Store) BackfillOrLoadAppSettings(_ *file.JsonDb) (string, int, error) {
	var sqlCount int
	if err := s.Reader.QueryRowContext(context.Background(),
		`SELECT COUNT(1) FROM app_settings`).Scan(&sqlCount); err != nil {
		return "", 0, fmt.Errorf("count app_settings: %w", err)
	}
	if sqlCount == 0 {
		// First run on this DB → import from nps.conf via beego.AppConfig.
		section, err := beego.AppConfig.GetSection("default")
		if err != nil || len(section) == 0 {
			// Some INI loaders return the unsectioned area under "" .
			section, _ = beego.AppConfig.GetSection("")
		}
		if len(section) == 0 {
			return "empty", 0, nil
		}
		tx, err := s.Writer.BeginTx(context.Background(), nil)
		if err != nil {
			return "", 0, err
		}
		defer func() { _ = tx.Rollback() }()
		stmt, err := tx.Prepare(
			`INSERT INTO app_settings(key, value, updated_at) VALUES(?, ?, ?)`)
		if err != nil {
			return "", 0, err
		}
		defer stmt.Close()
		now := time.Now().Unix()
		var inserted int
		for k, v := range section {
			if k == "" || IsBootstrapKey(k) {
				continue
			}
			if _, err := stmt.Exec(k, v, now); err != nil {
				return "", inserted, fmt.Errorf("backfill setting %q: %w", k, err)
			}
			inserted++
		}
		if err := tx.Commit(); err != nil {
			return "", inserted, err
		}
		return "backfill", inserted, nil
	}

	// Subsequent runs → SQLite is the source of truth, push into beego.
	rows, err := s.Reader.QueryContext(context.Background(),
		`SELECT key, value FROM app_settings`)
	if err != nil {
		return "", 0, fmt.Errorf("load app_settings: %w", err)
	}
	defer rows.Close()
	var applied int
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return "", applied, err
		}
		if IsBootstrapKey(k) {
			// Should never happen (we filter on insert), but defensive.
			continue
		}
		if err := beego.AppConfig.Set(k, v); err != nil {
			return "", applied, fmt.Errorf("apply setting %q: %w", k, err)
		}
		applied++
	}
	return "load", applied, rows.Err()
}
