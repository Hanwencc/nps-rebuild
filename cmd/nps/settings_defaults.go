package main

import (
	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/file/sqlitedb"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// settingsDefaults are the baseline values written to app_settings on
// first boot when the operator did not provide a populated nps.conf
// (the stripped Phase 6.1 nps.conf only carries the three Web bootstrap
// keys). Without these, downstream code paths like
// `beego.AppConfig.Int("bridge_port")` panic with strconv parse errors
// because the key is empty rather than absent.
//
// Values mirror the historical NPS defaults documented in
// `conf/nps.conf` prior to the SQLite migration. Bootstrap keys
// (web_*) are NOT seeded here — those live in nps.conf only.
var settingsDefaults = map[string]string{
	// runtime
	"runmode":             "dev",
	"log_level":           "7",
	"log_path":            "",
	"flow_store_interval": "1",
	"disconnect_timeout":  "60",

	// bridge
	"bridge_type":     "tcp",
	"bridge_ip":       "0.0.0.0",
	"bridge_port":     "8024",
	"tls_enable":      "false",
	"tls_bridge_port": "8025",

	// http(s) proxy
	"http_proxy_ip":          "0.0.0.0",
	"http_proxy_port":        "80",
	"https_proxy_port":       "443",
	"https_just_proxy":       "true",
	"http_add_origin_header": "false",
	"http_cache":             "false",
	"http_cache_length":      "100",

	// p2p / pprof
	"p2p_ip":   "",
	"p2p_port": "6000",

	// access control
	"ip_limit":    "false",
	"allow_ports": "",

	// limits
	"allow_flow_limit":           "false",
	"allow_rate_limit":           "false",
	"allow_tunnel_num_limit":     "false",
	"allow_local_proxy":          "true",
	"allow_connection_num_limit": "false",
	"allow_multi_ip":             "false",

	// multi-user
	"allow_user_login":           "false",
	"allow_user_register":        "false",
	"allow_user_change_username": "false",

	// SOCKS5 shared gateway (Phase 9)
	"socks5_shared_port": "0",
	"socks5_shared_ip":   "0.0.0.0",
}

// applySettingsDefaults fills in missing app_settings rows so a fresh
// install (no SQLite + minimal nps.conf) starts on sane defaults.
// Existing values (whether from a previous SQLite db or from nps.conf
// just imported into beego.AppConfig) are NEVER overwritten.
//
// Run once, immediately after BackfillOrLoadAppSettings.
func applySettingsDefaults(store *sqlitedb.Store) {
	if store == nil {
		return
	}
	cur, err := store.ListAppSettings()
	if err != nil {
		logs.Warn("settings defaults: list app_settings failed: %v", err)
		return
	}
	missing := make(map[string]string, len(settingsDefaults))
	for k, v := range settingsDefaults {
		if _, ok := cur[k]; ok {
			continue
		}
		// Honour any value that nps.conf already provided so we don't
		// later overwrite it with the in-memory beego copy.
		if existing := beego.AppConfig.String(k); existing != "" {
			continue
		}
		missing[k] = v
	}
	if len(missing) == 0 {
		return
	}
	applied, rejected, err := store.UpsertSettings(missing)
	if err != nil {
		logs.Warn("settings defaults: upsert failed: %v", err)
		return
	}
	for k, v := range missing {
		_ = beego.AppConfig.Set(k, v)
	}
	logs.Info("settings defaults: seeded %d row(s) (rejected=%d) into app_settings", applied, len(rejected))
}

// Ensure file package is referenced for documentation purposes; the
// import keeps this file tied to the same dependency graph as the
// other nps boot helpers.
var _ = file.GetDb
