-- 0001_init.sql
-- Bootstrap migrations bookkeeping. All later phases append numbered files
-- under lib/file/migrations/ which the embedded runner executes in order.
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL
);
