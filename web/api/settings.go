package api

import (
	"encoding/json"
	"sort"
	"strings"

	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/file/sqlitedb"
	"github.com/astaxie/beego"
)

// SettingsController exposes the SQLite-backed app_settings table that
// mirrors nps.conf. Admin-only.
//
//	GET  /api/v1/settings        →  full snapshot (persisted + bootstrap)
//	PUT  /api/v1/settings        →  bulk update; takes a JSON object whose
//	                                values are coerced to strings
type SettingsController struct {
	baseController
}

// settingDescriptor describes a single editable key. The Vue UI uses
// `group`/`label`/`type` to render the form. Any key encountered at
// runtime that is NOT in this list is still returned (under group
// "其他"), so older or third-party keys remain editable.
type settingDescriptor struct {
	Key            string   `json:"key"`
	Label          string   `json:"label"`
	Group          string   `json:"group"`
	Type           string   `json:"type"`     // "string" | "int" | "bool" | "enum" | "password"
	Enum           []string `json:"enum,omitempty"`
	Help           string   `json:"help,omitempty"`
	NeedsRestart   bool     `json:"needsRestart"`   // port/IP/TLS — UI shows badge
	Bootstrap      bool     `json:"bootstrap"`      // lives in nps.conf only
	// ReadOnly marks fields whose value is computed by nps itself and
	// re-overwritten on every boot. The Update endpoint silently
	// rejects writes to these keys; the UI should render the input
	// disabled. Example: tls_cert_fingerprint is derived from the
	// on-disk server.pem and cannot be set by the operator.
	ReadOnly       bool     `json:"readOnly"`
}

// settingsCatalog is the ordered, curated list of known nps.conf keys.
// Order here drives display order in the UI.
var settingsCatalog = []settingDescriptor{
	// ----- web (bootstrap) -----
	{Key: "web_port", Label: "Web 端口", Group: "Web 管理面板", Type: "int", Bootstrap: true, NeedsRestart: true, Help: "管理面板监听端口（编辑 nps.conf 后重启 nps 生效）"},
	{Key: "web_username", Label: "管理员用户名", Group: "Web 管理面板", Type: "string", Bootstrap: true, NeedsRestart: true, Help: "管理员账号（编辑 nps.conf 后重启 nps 生效）"},
	{Key: "web_password", Label: "管理员密码", Group: "Web 管理面板", Type: "password", Bootstrap: true, NeedsRestart: true, Help: "管理员密码（编辑 nps.conf 后重启 nps 生效）"},
	{Key: "web_ip", Label: "Web 监听 IP", Group: "Web 管理面板", Type: "string", NeedsRestart: true},
	{Key: "web_host", Label: "Web 域名", Group: "Web 管理面板", Type: "string", NeedsRestart: true},
	{Key: "web_base_url", Label: "Web 子路径", Group: "Web 管理面板", Type: "string", NeedsRestart: true, Help: "反向代理子路径，如 /nps"},
	{Key: "web_open_ssl", Label: "启用 HTTPS", Group: "Web 管理面板", Type: "bool", NeedsRestart: true},
	{Key: "web_cert_file", Label: "Web 证书路径", Group: "Web 管理面板", Type: "string", NeedsRestart: true},
	{Key: "web_key_file", Label: "Web 私钥路径", Group: "Web 管理面板", Type: "string", NeedsRestart: true},
	{Key: "open_captcha", Label: "登录验证码", Group: "Web 管理面板", Type: "bool"},

	// ----- runtime -----
	{Key: "runmode", Label: "运行模式", Group: "运行时", Type: "enum", Enum: []string{"dev", "prod"}},
	{Key: "log_level", Label: "日志级别", Group: "运行时", Type: "enum", Enum: []string{"0", "1", "2", "3", "4", "5", "6", "7"}, Help: "0=Emergency .. 7=Debug，保存后即时生效"},
	{Key: "log_path", Label: "日志路径", Group: "运行时", Type: "string", NeedsRestart: true},
	{Key: "flow_store_interval", Label: "流量持久化间隔(分钟)", Group: "运行时", Type: "int", Help: "0 表示不持久化"},
	{Key: "disconnect_timeout", Label: "断线超时(秒)", Group: "运行时", Type: "int", Help: "保存后下一个 ping 周期生效"},
	{Key: "system_info_display", Label: "显示系统信息", Group: "运行时", Type: "bool"},

	// ----- bridge -----
	{Key: "bridge_type", Label: "Bridge 类型", Group: "Bridge", Type: "enum", Enum: []string{"tcp", "kcp"}, NeedsRestart: true},
	{Key: "bridge_ip", Label: "Bridge 监听 IP", Group: "Bridge", Type: "string", NeedsRestart: true},
	{Key: "bridge_port", Label: "Bridge 端口", Group: "Bridge", Type: "int", NeedsRestart: true},
	{Key: "tls_enable", Label: "启用 TLS", Group: "Bridge", Type: "bool", NeedsRestart: true, Help: "对新连接生效，已建立的隧道保留旧值"},
	{Key: "tls_bridge_port", Label: "TLS Bridge 端口", Group: "Bridge", Type: "int", NeedsRestart: true},
	{Key: "tls_cert_fingerprint", Label: "Bridge 证书指纹 (SHA-256)", Group: "Bridge", Type: "string", ReadOnly: true, Help: "由 conf/server.pem 自动计算，每次启动重新写入；NPC 需将此值填入 tls_server_fingerprint 以防御中间人攻击"},
	{Key: "public_vkey", Label: "公共连接密钥", Group: "Bridge", Type: "password"},

	// ----- http(s) proxy (Phase 6.2: all hot-applied via httpHostServer rebind) -----
	{Key: "http_proxy_ip", Label: "HTTP 代理 IP", Group: "HTTP/HTTPS 代理", Type: "string", Help: "保存后重建 HTTP/HTTPS 监听器"},
	{Key: "http_proxy_port", Label: "HTTP 代理端口", Group: "HTTP/HTTPS 代理", Type: "int", Help: "保存后重建 HTTP 监听器"},
	{Key: "https_proxy_port", Label: "HTTPS 代理端口", Group: "HTTP/HTTPS 代理", Type: "int", Help: "保存后重建 HTTPS 监听器"},
	{Key: "https_just_proxy", Label: "仅 HTTPS 代理", Group: "HTTP/HTTPS 代理", Type: "bool"},
	{Key: "https_default_cert_file", Label: "默认 HTTPS 证书", Group: "HTTP/HTTPS 代理", Type: "string"},
	{Key: "https_default_key_file", Label: "默认 HTTPS 私钥", Group: "HTTP/HTTPS 代理", Type: "string"},
	{Key: "http_add_origin_header", Label: "保留真实 IP", Group: "HTTP/HTTPS 代理", Type: "bool"},
	{Key: "http_cache", Label: "启用 HTTP 缓存", Group: "HTTP/HTTPS 代理", Type: "bool"},
	{Key: "http_cache_length", Label: "HTTP 缓存条目数", Group: "HTTP/HTTPS 代理", Type: "int"},

	// ----- p2p / pprof -----
	// ----- shared SOCKS5 gateway (Phase 9) -----
	{Key: "socks5_shared_port", Label: "SOCKS5 共享端口", Group: "SOCKS5 共享网关", Type: "int", Help: "0 表示禁用；保存后立即重启监听器，所有 SOCKS5 任务复用此端口"},
	{Key: "socks5_shared_ip", Label: "SOCKS5 监听 IP", Group: "SOCKS5 共享网关", Type: "string", Help: "默认 0.0.0.0；保存后立即重启监听器"},

	// ----- p2p / pprof -----
	{Key: "p2p_ip", Label: "P2P 服务 IP", Group: "P2P / Debug", Type: "string", NeedsRestart: true},
	{Key: "p2p_port", Label: "P2P 端口", Group: "P2P / Debug", Type: "int", NeedsRestart: true},
	{Key: "pprof_ip", Label: "pprof IP", Group: "P2P / Debug", Type: "string", NeedsRestart: true},
	{Key: "pprof_port", Label: "pprof 端口", Group: "P2P / Debug", Type: "int", NeedsRestart: true},

	// ----- access control -----
	{Key: "ip_limit", Label: "限制 IP 访问", Group: "鉴权", Type: "bool"},
	{Key: "allow_ports", Label: "允许的端口范围", Group: "鉴权", Type: "string", Help: "如 9001-9009,10001"},

	// ----- multi-user -----
	{Key: "allow_user_login", Label: "允许多用户登录", Group: "多用户", Type: "bool"},
	{Key: "allow_user_register", Label: "允许注册", Group: "多用户", Type: "bool"},
	{Key: "allow_user_change_username", Label: "允许改用户名", Group: "多用户", Type: "bool"},

	// ----- limits -----
	{Key: "allow_flow_limit", Label: "启用流量限制", Group: "限流", Type: "bool"},
	{Key: "allow_rate_limit", Label: "启用带宽限制", Group: "限流", Type: "bool"},
	{Key: "allow_tunnel_num_limit", Label: "启用隧道数限制", Group: "限流", Type: "bool"},
	{Key: "allow_local_proxy", Label: "允许 local proxy", Group: "限流", Type: "bool"},
	{Key: "allow_connection_num_limit", Label: "启用连接数限制", Group: "限流", Type: "bool"},
	{Key: "allow_multi_ip", Label: "允许多 IP 隧道", Group: "限流", Type: "bool"},
}

// settingsResponseItem is a catalog entry plus its current value.
type settingsResponseItem struct {
	settingDescriptor
	Value string `json:"value"`
}

func (c *SettingsController) store() *sqlitedb.Store {
	return sqlitedb.From(file.GetDb())
}

// Get GET /api/v1/settings
func (c *SettingsController) Get() {
	if !c.currentIsAdmin() {
		c.forbidden("permission denied")
		return
	}
	// Always read through beego.AppConfig — that's the source of truth
	// at runtime (kept in sync with SQLite by sqlitedb.UpsertSetting).
	knownKeys := map[string]struct{}{}
	items := make([]settingsResponseItem, 0, len(settingsCatalog))
	for _, d := range settingsCatalog {
		knownKeys[d.Key] = struct{}{}
		items = append(items, settingsResponseItem{
			settingDescriptor: d,
			Value:             beego.AppConfig.String(d.Key),
		})
	}
	// Surface any extra keys present in SQLite that aren't in the
	// curated catalog (e.g. user-added or third-party). Group "其他".
	if s := c.store(); s != nil {
		all, err := s.ListAppSettings()
		if err == nil {
			extras := make([]settingsResponseItem, 0)
			for k, v := range all {
				if _, ok := knownKeys[k]; ok {
					continue
				}
				if sqlitedb.IsBootstrapKey(k) {
					continue
				}
				extras = append(extras, settingsResponseItem{
					settingDescriptor: settingDescriptor{
						Key: k, Label: k, Group: "其他", Type: "string",
					},
					Value: v,
				})
			}
			sort.Slice(extras, func(i, j int) bool { return extras[i].Key < extras[j].Key })
			items = append(items, extras...)
		}
	}
	c.ok(items)
}

// Update PUT /api/v1/settings
//
// Body: a JSON object {"key":"value", ...}. Values may be string,
// number, or bool — they're stringified before being persisted.
// Bootstrap keys (web_port, web_username, web_password) are rejected.
func (c *SettingsController) Update() {
	if !c.currentIsAdmin() {
		c.forbidden("permission denied")
		return
	}
	body := c.Ctx.Input.RequestBody
	if len(body) == 0 {
		c.badRequest("empty body")
		return
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		c.badRequest("invalid JSON: " + err.Error())
		return
	}
	if len(raw) == 0 {
		c.badRequest("no keys to update")
		return
	}
	store := c.store()
	if store == nil {
		c.serverErr("sqlite store unavailable")
		return
	}
	kv := make(map[string]string, len(raw))
	for k, v := range raw {
		k = strings.TrimSpace(k)
		if isReadOnlySetting(k) {
			// Drop silently — same contract as bootstrap keys. Logged
			// so a curious operator can confirm why their edit didn't
			// stick. Avoids a 4xx that would make the bulk save flow
			// in the UI feel broken when only one key was rejected.
			continue
		}
		kv[k] = stringify(v)
	}
	applied, rejected, err := store.UpsertSettings(kv)
	if err != nil {
		c.serverErr(err.Error())
		return
	}
	c.ok(map[string]any{
		"applied":  applied,
		"rejected": rejected, // bootstrap keys silently dropped
	})
}

// isReadOnlySetting returns true for catalog entries marked
// ReadOnly:true. Linear scan is fine — the catalog is ~50 entries
// long and Update() is not on a hot path.
func isReadOnlySetting(key string) bool {
	for _, d := range settingsCatalog {
		if d.Key == key {
			return d.ReadOnly
		}
	}
	return false
}

func stringify(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(v)
		s := string(b)
		// strip surrounding quotes if marshal produced a string literal
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		return s
	}
}
