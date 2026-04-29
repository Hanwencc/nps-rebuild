-- Phase 2 of JSON->SQLite migration: clients table.
--
-- Schema mirrors the persistable subset of lib/file.Client. Runtime-only
-- fields are NOT stored:
--   IsConnect / NowConn / Rate / Version
-- because they either change on every byte transferred or are
-- reconstructed at startup.
--
-- inlet_flow / export_flow ARE persisted: the periodic flush ticker in
-- server/server.go snapshots them to disk roughly every minute, so they
-- survive a restart with at most a minute of accumulator loss (same
-- behaviour as the legacy clients.json path).
--
-- Slice fields (black_ip_list, ip_white_list) are stored as JSON TEXT
-- because the row count is small (admin-managed) and we never filter on
-- individual list members in SQL.
--
-- verify_key gets a non-unique index because the bridge looks clients up
-- by key on every connect. Uniqueness is enforced by VerifyVkey() in Go,
-- not at the schema level (legacy code allows blank verify_key for
-- transient/no-store clients).

CREATE TABLE IF NOT EXISTS clients (
    id                  INTEGER PRIMARY KEY,
    verify_key          TEXT    NOT NULL DEFAULT '',
    addr                TEXT    NOT NULL DEFAULT '',
    remark              TEXT    NOT NULL DEFAULT '',
    status              INTEGER NOT NULL DEFAULT 1,
    rate_limit          INTEGER NOT NULL DEFAULT 0,
    flow_limit          INTEGER NOT NULL DEFAULT 0,
    inlet_flow          INTEGER NOT NULL DEFAULT 0,
    export_flow         INTEGER NOT NULL DEFAULT 0,
    max_conn            INTEGER NOT NULL DEFAULT 0,
    max_tunnel_num      INTEGER NOT NULL DEFAULT 0,
    web_user_name       TEXT    NOT NULL DEFAULT '',
    web_password        TEXT    NOT NULL DEFAULT '',
    config_conn_allow   INTEGER NOT NULL DEFAULT 1,
    no_store            INTEGER NOT NULL DEFAULT 0,
    no_display          INTEGER NOT NULL DEFAULT 0,
    ip_white            INTEGER NOT NULL DEFAULT 0,
    ip_white_pass       TEXT    NOT NULL DEFAULT '',
    ip_white_list       TEXT    NOT NULL DEFAULT '[]', -- JSON array
    black_ip_list       TEXT    NOT NULL DEFAULT '[]', -- JSON array
    cnf_u               TEXT    NOT NULL DEFAULT '',
    cnf_p               TEXT    NOT NULL DEFAULT '',
    cnf_compress        INTEGER NOT NULL DEFAULT 0,
    cnf_crypt           INTEGER NOT NULL DEFAULT 0,
    create_time         TEXT    NOT NULL DEFAULT '',
    last_online_time    TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_clients_verify_key ON clients(verify_key);
