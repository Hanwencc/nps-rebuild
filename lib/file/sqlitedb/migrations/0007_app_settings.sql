-- Phase 6.1: app_settings — generic key/value table mirroring nps.conf.
--
-- The original storage is conf/nps.conf (INI) which is loaded by Beego at
-- startup into beego.AppConfig. We mirror every key into this table so
-- the admin UI can edit them at runtime; nps.conf keeps only the
-- bootstrap subset (web_port / web_username / web_password) so the web
-- server can come up before SQLite is opened.
--
-- Values are stored as TEXT; type coercion happens at the read site via
-- beego.AppConfig.{Int,Bool,DefaultBool,...}, exactly as before.

CREATE TABLE IF NOT EXISTS app_settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at INTEGER NOT NULL DEFAULT 0
);
