package main

import (
	"strings"

	"ehang.io/nps/bridge"
	"ehang.io/nps/lib/file/sqlitedb"
	"ehang.io/nps/server"
	"ehang.io/nps/server/tool"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// registerSettingsHotHooks wires the SettingChange callbacks for keys
// that can take effect WITHOUT restarting nps.
//
// Three tiers of "hot":
//   1. AppConfig-only (handled implicitly by sqlitedb.UpsertSettings
//      calling beego.AppConfig.Set): every key whose runtime read site
//      uses beego.AppConfig.* gets the new value on the next read with
//      no extra wiring. This covers most allow_*_limit / allow_user_*
//      / open_captcha / system_info_display style toggles.
//   2. Field-mutation hooks (this file): values that are cached in a
//      package-level variable or unexported struct field after boot
//      need an explicit setter call here.
//   3. Listener rebind hooks (this file): http_proxy_* and
//      https_proxy_* trigger a debounced restart of the implicit
//      httpHostServer task via cmd/nps/hothttp.go.
//
// Keys that need a full process restart (bridge_*, tls_bridge_port,
// web_*, p2p_*, pprof_*, log_path, runmode, flow_store_interval) are
// intentionally NOT hooked — the UI surfaces a "需重启" badge for them.
//
// All hooks fire AFTER beego.AppConfig has been updated, so the new
// value is already visible via beego.AppConfig.{String,Int,Bool}.
func registerSettingsHotHooks() {
	// --- tier 2: field mutations -------------------------------------

	// log_level: re-arm the logger so existing log calls pick the new
	// verbosity. Mirrors the level-parse the boot path uses.
	sqlitedb.OnSettingChange("log_level", func(_, _, newVal string) {
		lvl := strings.TrimSpace(newVal)
		if lvl == "" {
			lvl = "7"
		}
		logs.SetLevel(parseLogLevel(lvl))
		logs.Info("log_level hot-applied -> %s", lvl)
	})

	// tls_enable: mutate the bridge global. New incoming bridge
	// connections see the flipped value on the next handshake.
	// Established tunnels are not affected.
	sqlitedb.OnSettingChange("tls_enable", func(_, _, newVal string) {
		bridge.ServerTlsEnable = beego.AppConfig.DefaultBool("tls_enable", false)
		logs.Info("tls_enable hot-applied -> %t (existing connections keep old setting)", bridge.ServerTlsEnable)
	})

	// ip_limit: mutate Bridge.ipVerify. Affects new connections only.
	sqlitedb.OnSettingChange("ip_limit", func(_, _, _ string) {
		if server.Bridge == nil {
			return
		}
		v, _ := beego.AppConfig.Bool("ip_limit")
		server.Bridge.SetIpVerify(v)
		logs.Info("ip_limit hot-applied -> %t", v)
	})

	// disconnect_timeout: mutate Bridge.disconnectTime. Read on every
	// ping tick (~10s).
	sqlitedb.OnSettingChange("disconnect_timeout", func(_, _, _ string) {
		if server.Bridge == nil {
			return
		}
		n, err := beego.AppConfig.Int("disconnect_timeout")
		if err != nil {
			n = 60
		}
		server.Bridge.SetDisconnectTime(n)
		logs.Info("disconnect_timeout hot-applied -> %d", n)
	})

	// allow_ports: re-parse the whitelist. Subsequent task creation
	// calls go through tool.TestServerPort which reads the refreshed
	// `ports` slice.
	sqlitedb.OnSettingChange("allow_ports", func(_, _, _ string) {
		tool.InitAllowPort()
		logs.Info("allow_ports hot-applied -> %s", beego.AppConfig.String("allow_ports"))
	})

	// --- tier 3: HTTP/HTTPS proxy listener rebind --------------------
	// All of these get baked into *httpServer at construction time, so
	// the only way to apply them is to tear down + rebuild that one
	// service. scheduleHttpRestart debounces multi-key saves into one
	// restart.
	for _, k := range []string{
		"http_proxy_ip",
		"http_proxy_port",
		"https_proxy_port",
		"https_just_proxy",
		"https_default_cert_file",
		"https_default_key_file",
		"http_cache",
		"http_cache_length",
		"http_add_origin_header",
	} {
		key := k // capture
		sqlitedb.OnSettingChange(key, func(_, _, _ string) {
			scheduleHttpRestart(key)
		})
	}

	// --- audit log (wildcard) ----------------------------------------
	// Fired AFTER any specific-key hook, once per changed key. Skip
	// password-shaped keys to avoid leaking secrets into nps.log.
	sqlitedb.OnSettingChange("", func(key, _, newVal string) {
		if isSecretKey(key) {
			logs.Info("setting hot-applied: %s = ***", key)
			return
		}
		logs.Info("setting hot-applied: %s = %s", key, newVal)
	})
}

func parseLogLevel(s string) int {
	switch strings.TrimSpace(s) {
	case "0":
		return logs.LevelEmergency
	case "1":
		return logs.LevelAlert
	case "2":
		return logs.LevelCritical
	case "3":
		return logs.LevelError
	case "4":
		return logs.LevelWarning
	case "5":
		return logs.LevelNotice
	case "6":
		return logs.LevelInformational
	case "7":
		return logs.LevelDebug
	}
	return logs.LevelDebug
}

func isSecretKey(k string) bool {
	switch k {
	case "public_vkey":
		return true
	}
	return false
}
