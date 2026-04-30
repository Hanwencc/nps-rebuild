// Package sqlitedb — SQLite store (Phase 0 of the JSON->SQLite migration).
//
// This package is intentionally separate from `lib/file` so that
// client-only binaries (npc) do not transitively pull in the
// `modernc.org/sqlite` + `modernc.org/libc` blobs (~6 MB stripped).
// Only the server entrypoint imports it.
//
// Design notes:
//   - Pure-Go driver `modernc.org/sqlite` keeps nps cgo-free.
//   - Two *sql.DB handles: a single-writer pool to avoid `database is
//     locked`, and an unrestricted reader pool. WAL mode plus a
//     reasonable busy_timeout makes concurrent reads non-blocking.
//   - Migrations are embedded SQL files under ./migrations and applied
//     in lexical order, tracked by the schema_migrations table.
package sqlitedb

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store wraps a SQLite database with split read/write pools.
type Store struct {
	Path   string
	Writer *sql.DB // single connection, serialised writes
	Reader *sql.DB // multi connection pool for reads
}

// Open opens (or creates) the SQLite database at `path`,
// applies pragmas and runs all embedded migrations.
func Open(path string) (*Store, error) {
	// Ensure the parent directory exists. modernc.org/sqlite returns
	// the cryptic SQLITE_CANTOPEN ("unable to open database file (14)")
	// when the containing directory is missing — common in fresh
	// Docker / first-run setups where /conf has not been bind-mounted
	// or pre-created. Create it now so a missing directory no longer
	// blocks first boot; the file itself is created by the driver.
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("sqlite mkdir %s: %w", dir, err)
		}
	}

	dsnW := buildDSN(path, true)
	dsnR := buildDSN(path, false)

	w, err := sql.Open("sqlite", dsnW)
	if err != nil {
		return nil, fmt.Errorf("sqlite open writer: %w", err)
	}
	w.SetMaxOpenConns(1)
	w.SetMaxIdleConns(1)
	w.SetConnMaxIdleTime(time.Hour)

	if err := w.Ping(); err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("sqlite ping writer: %w", err)
	}

	r, err := sql.Open("sqlite", dsnR)
	if err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("sqlite open reader: %w", err)
	}
	r.SetMaxOpenConns(8)
	r.SetMaxIdleConns(4)
	r.SetConnMaxIdleTime(time.Hour)

	if err := r.Ping(); err != nil {
		_ = w.Close()
		_ = r.Close()
		return nil, fmt.Errorf("sqlite ping reader: %w", err)
	}

	s := &Store{Path: path, Writer: w, Reader: r}
	if err := s.applyMigrations(context.Background()); err != nil {
		_ = w.Close()
		_ = r.Close()
		return nil, err
	}
	return s, nil
}

// Close releases both connection pools.
func (s *Store) Close() error {
	var firstErr error
	if s.Writer != nil {
		if err := s.Writer.Close(); err != nil {
			firstErr = err
		}
	}
	if s.Reader != nil {
		if err := s.Reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func buildDSN(path string, writer bool) string {
	pragmas := []string{
		"_pragma=journal_mode(WAL)",
		"_pragma=synchronous(NORMAL)",
		"_pragma=foreign_keys(ON)",
		"_pragma=busy_timeout(5000)",
		"_pragma=temp_store(MEMORY)",
	}
	if !writer {
		pragmas = append(pragmas, "_pragma=query_only(1)")
	}
	return "file:" + path + "?cache=shared&" + strings.Join(pragmas, "&")
}

func (s *Store) applyMigrations(ctx context.Context) error {
	const bootstrap = `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`
	if _, err := s.Writer.ExecContext(ctx, bootstrap); err != nil {
		return fmt.Errorf("schema_migrations bootstrap: %w", err)
	}

	applied, err := s.appliedVersions(ctx)
	if err != nil {
		return err
	}

	entries, err := listMigrations()
	if err != nil {
		return err
	}

	for _, m := range entries {
		if _, ok := applied[m.Version]; ok {
			continue
		}
		if err := s.runMigration(ctx, m); err != nil {
			return fmt.Errorf("migration %s: %w", m.Name, err)
		}
	}
	return nil
}

type migrationEntry struct {
	Version int
	Name    string
	SQL     string
}

func listMigrations() ([]migrationEntry, error) {
	dir, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	out := make([]migrationEntry, 0, len(dir))
	for _, e := range dir {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		base := filepath.Base(e.Name())
		under := strings.IndexByte(base, '_')
		if under <= 0 {
			continue
		}
		v, err := strconv.Atoi(base[:under])
		if err != nil {
			continue
		}
		body, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, err
		}
		out = append(out, migrationEntry{Version: v, Name: e.Name(), SQL: string(body)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

func (s *Store) appliedVersions(ctx context.Context) (map[int]struct{}, error) {
	rows, err := s.Writer.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	out := make(map[int]struct{})
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = struct{}{}
	}
	return out, rows.Err()
}

func (s *Store) runMigration(ctx context.Context, m migrationEntry) error {
	tx, err := s.Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, stmt := range splitStatements(m.SQL) {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec %q: %w", firstLine(stmt), err)
		}
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`,
		m.Version, time.Now().Unix(),
	); err != nil {
		return err
	}
	return tx.Commit()
}

func splitStatements(sqlText string) []string {
	// Split on ';' but ignore semicolons that appear inside `-- ...`
	// line comments or single-quoted string literals. The previous naive
	// strings.Split(sqlText, ";") tripped on `;` in comment headers.
	var (
		parts   []string
		buf     strings.Builder
		inLine  bool // inside `-- ...` until newline
		inStr   bool // inside '...' (sqlite uses '' to escape a quote)
	)
	for i := 0; i < len(sqlText); i++ {
		c := sqlText[i]
		switch {
		case inLine:
			buf.WriteByte(c)
			if c == '\n' {
				inLine = false
			}
		case inStr:
			buf.WriteByte(c)
			if c == '\'' {
				inStr = false
			}
		case c == '-' && i+1 < len(sqlText) && sqlText[i+1] == '-':
			buf.WriteByte(c)
			inLine = true
		case c == '\'':
			buf.WriteByte(c)
			inStr = true
		case c == ';':
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(c)
		}
	}
	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		// Strip leading line-comments and blank lines so a statement
		// preceded by an `-- ...` block (common at the top of a
		// migration file) is not mistaken for a pure-comment chunk.
		lines := strings.Split(p, "\n")
		i := 0
		for i < len(lines) {
			t := strings.TrimSpace(lines[i])
			if t == "" || strings.HasPrefix(t, "--") {
				i++
				continue
			}
			break
		}
		stmt := strings.TrimSpace(strings.Join(lines[i:], "\n"))
		if stmt == "" {
			continue
		}
		out = append(out, stmt)
	}
	return out
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
