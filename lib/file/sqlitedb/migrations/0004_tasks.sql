-- Phase 3 of JSON->SQLite migration: tasks (Tunnel) table.
--
-- Schema mirrors the persistable subset of lib/file.Tunnel. Runtime-only
-- fields are NOT stored:
--   RunStatus, Health.HealthNextTime, Health.HealthMap,
--   Health.HealthRemoveArr, Target.nowIndex, Target.TargetArr
-- because they are recomputed at startup or maintained per-request.
--
-- inlet_flow / export_flow ARE persisted: the periodic flush ticker in
-- server/server.go snapshots them roughly every minute, so they survive
-- a restart with at most a minute of accumulator loss (same behaviour
-- as the legacy tasks.json path).
--
-- Map / multi-target fields (multi_account_map, target_str) are stored
-- as TEXT (JSON for the map, raw newline-separated string for targets,
-- matching the in-memory representation).
--
-- client_id references clients.id but is NOT enforced as a SQL foreign
-- key: the application already prunes orphaned tunnels via
-- DelTunnelAndHostByClientId() at controller level, and a hard FK with
-- ON DELETE CASCADE would silently drop rows behind the in-memory
-- sync.Map, leaving stale Tunnel pointers in RunList. An index on
-- client_id keeps lookups by client cheap.

CREATE TABLE IF NOT EXISTS tasks (
    id                      INTEGER PRIMARY KEY,
    client_id               INTEGER NOT NULL,
    mode                    TEXT    NOT NULL DEFAULT '',
    port                    INTEGER NOT NULL DEFAULT 0,
    server_ip               TEXT    NOT NULL DEFAULT '',
    status                  INTEGER NOT NULL DEFAULT 1,
    ports                   TEXT    NOT NULL DEFAULT '',
    password                TEXT    NOT NULL DEFAULT '',
    remark                  TEXT    NOT NULL DEFAULT '',
    target_addr             TEXT    NOT NULL DEFAULT '',
    local_path              TEXT    NOT NULL DEFAULT '',
    strip_pre               TEXT    NOT NULL DEFAULT '',
    proto_version           TEXT    NOT NULL DEFAULT '',
    no_store                INTEGER NOT NULL DEFAULT 0,
    inlet_flow              INTEGER NOT NULL DEFAULT 0,
    export_flow             INTEGER NOT NULL DEFAULT 0,
    flow_limit              INTEGER NOT NULL DEFAULT 0,
    target_str              TEXT    NOT NULL DEFAULT '',
    target_local_proxy      INTEGER NOT NULL DEFAULT 0,
    multi_account_map       TEXT    NOT NULL DEFAULT '{}', -- JSON object
    health_check_timeout    INTEGER NOT NULL DEFAULT 0,
    health_max_fail         INTEGER NOT NULL DEFAULT 0,
    health_check_interval   INTEGER NOT NULL DEFAULT 0,
    http_health_url         TEXT    NOT NULL DEFAULT '',
    health_check_type       TEXT    NOT NULL DEFAULT '',
    health_check_target     TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_tasks_client_id ON tasks(client_id);
CREATE INDEX IF NOT EXISTS idx_tasks_password  ON tasks(password);
