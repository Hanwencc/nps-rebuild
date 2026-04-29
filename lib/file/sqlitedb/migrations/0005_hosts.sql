-- Phase 4 of JSON->SQLite migration: hosts table.
--
-- Schema mirrors the persistable subset of lib/file.Host. Runtime-only
-- fields are NOT stored:
--   Health.HealthNextTime / HealthMap / HealthRemoveArr,
--   Target.nowIndex / TargetArr
-- because they are recomputed at startup or maintained per-request.
--
-- inlet_flow / export_flow ARE persisted by the periodic flush ticker
-- (server/server.go), matching the legacy hosts.json behaviour.
--
-- client_id is a logical reference to clients.id but NOT a SQL FK, for
-- the same reason as tasks: the in-memory sync.Map is the runtime
-- registry and a CASCADE would silently drop rows behind it.

CREATE TABLE IF NOT EXISTS hosts (
    id                      INTEGER PRIMARY KEY,
    client_id               INTEGER NOT NULL,
    host                    TEXT    NOT NULL DEFAULT '',
    header_change           TEXT    NOT NULL DEFAULT '',
    host_change             TEXT    NOT NULL DEFAULT '',
    location                TEXT    NOT NULL DEFAULT '',
    remark                  TEXT    NOT NULL DEFAULT '',
    scheme                  TEXT    NOT NULL DEFAULT '',
    cert_file_path          TEXT    NOT NULL DEFAULT '',
    key_file_path           TEXT    NOT NULL DEFAULT '',
    no_store                INTEGER NOT NULL DEFAULT 0,
    is_close                INTEGER NOT NULL DEFAULT 0,
    auto_https              INTEGER NOT NULL DEFAULT 0,
    inlet_flow              INTEGER NOT NULL DEFAULT 0,
    export_flow             INTEGER NOT NULL DEFAULT 0,
    flow_limit              INTEGER NOT NULL DEFAULT 0,
    target_str              TEXT    NOT NULL DEFAULT '',
    target_local_proxy      INTEGER NOT NULL DEFAULT 0,
    health_check_timeout    INTEGER NOT NULL DEFAULT 0,
    health_max_fail         INTEGER NOT NULL DEFAULT 0,
    health_check_interval   INTEGER NOT NULL DEFAULT 0,
    http_health_url         TEXT    NOT NULL DEFAULT '',
    health_check_type       TEXT    NOT NULL DEFAULT '',
    health_check_target     TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_hosts_client_id ON hosts(client_id);
CREATE INDEX IF NOT EXISTS idx_hosts_host      ON hosts(host);
