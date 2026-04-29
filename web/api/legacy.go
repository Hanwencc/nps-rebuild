// Package api — legacy public endpoints kept for backwards
// compatibility. The only remaining one is the IP-whitelist flow
// used by NPC clients to self-register their public IP via vkey +
// pass; the md5(auth_key+ts) signature flow has been removed in
// favour of API tokens (see web/api/token.go).
package api

import (
	"html"

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

// IpWhiteAuth lets an unauthenticated visitor add their public IP to
// a client's whitelist by presenting the client's vkey + IP-whitelist
// password.
//
// GET|POST /auth/ipwhiteauth?vkey=...&pass=...&ip=...
func (s *LegacyAuthController) IpWhiteAuth() {
	s.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	s.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	s.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	vkey := html.EscapeString(s.GetString("vkey"))
	ip := html.EscapeString(s.GetString("ip"))
	password := html.EscapeString(s.GetString("pass"))

	if vkey == "" || password == "" {
		s.Data["json"] = map[string]interface{}{"success": false, "message": "参数错误"}
		s.ServeJSON()
		return
	}
	if ip == "" {
		ip = html.EscapeString(s.Ctx.Input.IP())
	}

	c, err := file.GetDb().GetClientByVkey(vkey)
	if err != nil {
		s.Data["json"] = map[string]interface{}{"success": false, "message": "客户端密钥错误"}
		s.ServeJSON()
		logs.Error("ip-white auth failed, bad vkey: vkey=%s ip=%s", vkey, ip)
		return
	}

	if c.IpWhitePass != password {
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
