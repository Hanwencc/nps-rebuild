-- Phase 9: shared SOCKS5 gateway.
--
-- A single global SOCKS5 listener (configured via app_settings keys
-- socks5_shared_port / socks5_shared_ip) multiplexes traffic to many
-- NPC clients based on the SOCKS5 username. Each "socks5" task in the
-- tasks table degenerates into a routing entry whose key is the new
-- `username` column. The per-task Port / ServerIp columns are no longer
-- used by the runtime for socks5 mode but stay in the schema for
-- backward compatibility (older modes still rely on them).
--
-- An index on (mode, username) keeps the per-connection lookup O(log n).

ALTER TABLE tasks ADD COLUMN username TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_tasks_socks5_user ON tasks(mode, username);
