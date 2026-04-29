-- Phase 5: global configuration (single Glob record).
--
-- The original storage is conf/global.json which always contains exactly
-- one JSON object (BlackIpList[] + ServerUrl). We model that as a single
-- row keyed by id=1; any UPSERT replaces it in place.
--
-- BlackIpList is stored as a JSON-encoded TEXT (e.g. ["1.1.1.1","2.2.2.2"]).
-- Empty/missing list is normalised to "[]".

CREATE TABLE IF NOT EXISTS global (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    server_url    TEXT NOT NULL DEFAULT '',
    black_ip_list TEXT NOT NULL DEFAULT '[]'
);
