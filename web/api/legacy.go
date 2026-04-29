// Package api — legacy public endpoints kept for backwards
// compatibility. The only remaining one is the IP-whitelist flow
// used by NPC clients to self-register their public IP via vkey +
// pass; the md5 signature flow has been removed in favour of API
// tokens (see web/api/token.go).
package api

import (
	"crypto/subtle"
	"html"
	"net"

	"ehang.io/nps/lib/file"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// LegacyAuthController exposes the public, unauthenticated helper
// endpoints that used to live in web/controllers/auth.go.
//
// It embeds beego.Controller directly (NOT baseController) so the
// session/auth-key middleware does not apply.
type LegacyAuthController struct {
	beego.Controller
}

// remoteIP returns the request's TCP source IP, ignoring any
// X-Forwarded-For / X-Real-IP headers (which a client could spoof
// when NPS is exposed directly without a trusted reverse proxy).
func remoteIP(remoteAddr string) string {
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return h
	}
	return remoteAddr
}

// IpWhiteAuth lets an unauthenticated visitor add their public IP to
// a client's whitelist by presenting the client's vkey + IP-whitelist
// password.
//
// GET|POST /auth/ipwhiteauth?vkey=...&pass=...
//
// SECURITY: the IP added is ALWAYS the request's source address as
// observed by the server (s.Ctx.Input.IP()). Any user-supplied "ip"
// query parameter is ignored — otherwise an attacker who learned a
// client's vkey + IpWhitePass could whitelist arbitrary addresses
// (audit finding #7).
func (s *LegacyAuthController) IpWhiteAuth() {
	// CORS: this endpoint is intentionally public so NPC peers can
	// self-register; restrict methods/headers but no Origin echo.
	s.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	s.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	vkey := html.EscapeString(s.GetString("vkey"))
	password := html.EscapeString(s.GetString("pass"))
	// Always trust only the server-observed source IP. The "ip" query
	// parameter (if any) is intentionally ignored. We also bypass
	// Beego's Input.IP() because it consults X-Forwarded-For /
	// X-Real-IP, which a direct caller could forge.
	ip := remoteIP(s.Ctx.Request.RemoteAddr)

	if vkey == "" || password == "" {
		s.Data["json"] = map[string]interface{}{"success": false, "message": "参数错误"}
		s.ServeJSON()
		return
	}

	c, err := file.GetDb().GetClientByVkey(vkey)
	if err != nil {
		s.Data["json"] = map[string]interface{}{"success": false, "message": "客户端密钥错误"}
		s.ServeJSON()
		logs.Error("ip-white auth failed, bad vkey: vkey=%s ip=%s", vkey, ip)
		return
	}

	if subtle.ConstantTimeCompare([]byte(c.IpWhitePass), []byte(password)) != 1 {
		s.Data["json"] = map[string]interface{}{"success": false, "message": "授权密码错误"}
		s.ServeJSON()
		logs.Error("ip-white auth failed, bad password: vkey=%s ip=%s", vkey, ip)
		return
	}

	exists := false
	for _, e := range c.IpWhiteList {
		if e == ip {
			exists = true
			break
		}
	}
	if !exists {
		c.IpWhiteList = append(c.IpWhiteList, ip)
		file.GetDb().UpdateClient(c)
	}

	s.Data["json"] = map[string]interface{}{"success": true, "message": "授权成功"}
	s.ServeJSON()
	logs.Info("ip-white auth ok: vkey=%s ip=%s", vkey, ip)
}
