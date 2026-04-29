-- Phase 1 of JSON->SQLite migration: API tokens table.
--
-- Schema mirrors lib/file.ApiToken. Slice fields are stored as JSON text
-- because the row count is tiny (admin-managed credentials) and we
-- never need to filter on individual list members.
--
-- key_id has a UNIQUE index because the auth middleware looks tokens up
-- by KeyId on every authenticated request.

CREATE TABLE IF NOT EXISTS api_tokens (
    id                  INTEGER PRIMARY KEY,
    key_id              TEXT    NOT NULL,
    secret_hash         TEXT    NOT NULL,
    remark              TEXT    NOT NULL DEFAULT '',
    allowed_path_prefix TEXT    NOT NULL DEFAULT '',
    allowed_methods     TEXT    NOT NULL DEFAULT '[]', -- JSON array
    allow_ips           TEXT    NOT NULL DEFAULT '[]', -- JSON array
    expires_at          INTEGER NOT NULL DEFAULT 0,
    created_at          INTEGER NOT NULL DEFAULT 0,
    last_used_at        INTEGER NOT NULL DEFAULT 0,
    last_used_ip        TEXT    NOT NULL DEFAULT '',
    disabled            INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_key_id ON api_tokens(key_id);
