package sqlitedb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ehang.io/nps/lib/file"
)

// From extracts the *Store stashed on a *file.DbUtils by the server
// entrypoint. Returns nil when SQLite has not been initialised
// (e.g. during tests or if Open failed at startup).
func From(db *file.DbUtils) *Store {
	if db == nil || db.JsonDb == nil || db.JsonDb.SQLite == nil {
		return nil
	}
	s, _ := db.JsonDb.SQLite.(*Store)
	return s
}

// ----- API token CRUD -----------------------------------------------------

// scanApiToken reads a row produced by selectApiTokenColumns into a fresh
// *file.ApiToken.
func scanApiToken(scan func(dest ...any) error) (*file.ApiToken, error) {
	t := &file.ApiToken{}
	var (
		methodsJSON  string
		ipsJSON      string
		disabledInt  int
	)
	if err := scan(
		&t.Id, &t.KeyId, &t.SecretHash, &t.Remark,
		&t.AllowedPathPrefix, &methodsJSON, &ipsJSON,
		&t.ExpiresAt, &t.CreatedAt, &t.LastUsedAt, &t.LastUsedIp,
		&disabledInt,
	); err != nil {
		return nil, err
	}
	if methodsJSON != "" {
		_ = json.Unmarshal([]byte(methodsJSON), &t.AllowedMethods)
	}
	if ipsJSON != "" {
		_ = json.Unmarshal([]byte(ipsJSON), &t.AllowIps)
	}
	t.Disabled = disabledInt != 0
	return t, nil
}

const apiTokenColumns = `id, key_id, secret_hash, remark,
		allowed_path_prefix, allowed_methods, allow_ips,
		expires_at, created_at, last_used_at, last_used_ip, disabled`

// ListApiTokens returns all rows ordered by id ascending.
func (s *Store) ListApiTokens() ([]*file.ApiToken, error) {
	rows, err := s.Reader.QueryContext(context.Background(),
		`SELECT `+apiTokenColumns+` FROM api_tokens ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list api_tokens: %w", err)
	}
	defer rows.Close()
	out := make([]*file.ApiToken, 0)
	for rows.Next() {
		t, err := scanApiToken(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetApiToken fetches a single token by primary key.
func (s *Store) GetApiToken(id int) (*file.ApiToken, error) {
	row := s.Reader.QueryRowContext(context.Background(),
		`SELECT `+apiTokenColumns+` FROM api_tokens WHERE id=?`, id)
	t, err := scanApiToken(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("api token not found")
	}
	return t, err
}

// FindApiTokenByKeyId fetches a token by its public KeyId. Used on every
// authenticated request, hence the unique index on key_id.
func (s *Store) FindApiTokenByKeyId(keyId string) (*file.ApiToken, error) {
	if keyId == "" {
		return nil, errors.New("empty keyId")
	}
	row := s.Reader.QueryRowContext(context.Background(),
		`SELECT `+apiTokenColumns+` FROM api_tokens WHERE key_id=?`, keyId)
	t, err := scanApiToken(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("api token not found")
	}
	return t, err
}

// NewApiToken inserts a new row, assigning t.Id from sqlite's rowid when
// the caller leaves it 0. CreatedAt is filled in if zero.
func (s *Store) NewApiToken(t *file.ApiToken) error {
	if t.CreatedAt == 0 {
		t.CreatedAt = time.Now().Unix()
	}
	methods, _ := json.Marshal(t.AllowedMethods)
	ips, _ := json.Marshal(t.AllowIps)
	disabled := 0
	if t.Disabled {
		disabled = 1
	}

	if t.Id != 0 {
		_, err := s.Writer.ExecContext(context.Background(),
			`INSERT INTO api_tokens(`+apiTokenColumns+`)
			 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
			t.Id, t.KeyId, t.SecretHash, t.Remark,
			t.AllowedPathPrefix, string(methods), string(ips),
			t.ExpiresAt, t.CreatedAt, t.LastUsedAt, t.LastUsedIp, disabled)
		return err
	}

	res, err := s.Writer.ExecContext(context.Background(),
		`INSERT INTO api_tokens(key_id, secret_hash, remark,
			allowed_path_prefix, allowed_methods, allow_ips,
			expires_at, created_at, last_used_at, last_used_ip, disabled)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		t.KeyId, t.SecretHash, t.Remark,
		t.AllowedPathPrefix, string(methods), string(ips),
		t.ExpiresAt, t.CreatedAt, t.LastUsedAt, t.LastUsedIp, disabled)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.Id = int(id)
	return nil
}

// UpdateApiToken overwrites every mutable column for the given id.
func (s *Store) UpdateApiToken(t *file.ApiToken) error {
	methods, _ := json.Marshal(t.AllowedMethods)
	ips, _ := json.Marshal(t.AllowIps)
	disabled := 0
	if t.Disabled {
		disabled = 1
	}
	_, err := s.Writer.ExecContext(context.Background(),
		`UPDATE api_tokens SET
			key_id=?, secret_hash=?, remark=?,
			allowed_path_prefix=?, allowed_methods=?, allow_ips=?,
			expires_at=?, last_used_at=?, last_used_ip=?, disabled=?
		 WHERE id=?`,
		t.KeyId, t.SecretHash, t.Remark,
		t.AllowedPathPrefix, string(methods), string(ips),
		t.ExpiresAt, t.LastUsedAt, t.LastUsedIp, disabled, t.Id)
	return err
}

// DelApiToken removes a row.
func (s *Store) DelApiToken(id int) error {
	_, err := s.Writer.ExecContext(context.Background(),
		`DELETE FROM api_tokens WHERE id=?`, id)
	return err
}

// TouchApiToken bumps last_used_at / last_used_ip but only when at least
// one second has elapsed since the previous touch from the same IP.
// Returns the post-update row when it was modified, or nil otherwise.
// Done in SQL because the in-memory ApiToken returned from a SELECT is
// not the same instance across requests, so the previous *ApiToken.mu
// rate-limit no longer works.
func (s *Store) TouchApiToken(id int, ip string, now int64) error {
	_, err := s.Writer.ExecContext(context.Background(),
		`UPDATE api_tokens
		 SET last_used_at=?, last_used_ip=?
		 WHERE id=? AND (last_used_at < ? OR last_used_ip <> ?)`,
		now, ip, id, now, ip)
	return err
}

// BackfillApiTokens copies the in-memory tokens (loaded from
// api_tokens.json by lib/file at startup) into SQLite — but only when
// the table is empty. Subsequent boots are no-ops. Idempotent.
func (s *Store) BackfillApiTokens(rows []*file.ApiToken) (int, error) {
	var n int
	if err := s.Reader.QueryRowContext(context.Background(),
		`SELECT COUNT(1) FROM api_tokens`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count api_tokens: %w", err)
	}
	if n > 0 || len(rows) == 0 {
		return 0, nil
	}
	tx, err := s.Writer.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(
		`INSERT INTO api_tokens(` + apiTokenColumns + `)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	inserted := 0
	for _, t := range rows {
		methods, _ := json.Marshal(t.AllowedMethods)
		ips, _ := json.Marshal(t.AllowIps)
		disabled := 0
		if t.Disabled {
			disabled = 1
		}
		if _, err := stmt.Exec(
			t.Id, t.KeyId, t.SecretHash, t.Remark,
			t.AllowedPathPrefix, string(methods), string(ips),
			t.ExpiresAt, t.CreatedAt, t.LastUsedAt, t.LastUsedIp, disabled,
		); err != nil {
			return inserted, fmt.Errorf("backfill api_token id=%d: %w", t.Id, err)
		}
		inserted++
	}
	if err := tx.Commit(); err != nil {
		return inserted, err
	}
	return inserted, nil
}
