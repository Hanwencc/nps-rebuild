// Package routers wires the HTTP surface of nps after the migration
// to a Vue 3 single-page application.
//
// Layout:
//
//	/api/v1/...   — RESTful JSON API consumed by the SPA + 3rd-parties
//	/auth/...     — legacy public helpers (timestamp, auth-key, IP whitelist)
//	/ui/...       — embedded SPA (web/webui/dist via go:embed)
//	/             — 302 redirect to /ui/
package routers

import (
	"ehang.io/nps/web/api"
	"ehang.io/nps/web/webui"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func Init() {
	web_base_url := beego.AppConfig.String("web_base_url")

	// Beego does not populate Ctx.Input.RequestBody by default; required
	// by every JSON-decoding handler under /api/v1.
	beego.BConfig.CopyRequestBody = true

	// ---------- /api/v1 (RESTful) ----------------------------------------
	apiNs := beego.NewNamespace(web_base_url+"/api/v1",
		// auth
		beego.NSRouter("/auth/health", &api.AuthController{}, "get:Health"),
		beego.NSRouter("/auth/login", &api.AuthController{}, "post:Login"),
		beego.NSRouter("/auth/logout", &api.AuthController{}, "*:Logout"),
		beego.NSRouter("/auth/me", &api.AuthController{}, "get:Me"),
		// clients
		beego.NSRouter("/clients", &api.ClientController{}, "get:List;post:Create"),
		beego.NSRouter("/clients/:id", &api.ClientController{},
			"get:Get;put:Update;delete:Delete"),
		beego.NSRouter("/clients/:id/status", &api.ClientController{}, "post:ChangeStatus"),
		beego.NSRouter("/clients/:id/quickinfo", &api.ClientController{}, "get:QuickInfo"),
		// tunnels
		beego.NSRouter("/tunnels", &api.TunnelController{}, "get:List;post:Create"),
		beego.NSRouter("/tunnels/:id", &api.TunnelController{},
			"get:Get;put:Update;delete:Delete"),
		beego.NSRouter("/tunnels/:id/start", &api.TunnelController{}, "post:Start"),
		beego.NSRouter("/tunnels/:id/stop", &api.TunnelController{}, "post:Stop"),
		beego.NSRouter("/tunnels/:id/copy", &api.TunnelController{}, "post:Copy"),
		// hosts
		beego.NSRouter("/hosts", &api.HostController{}, "get:List;post:Create"),
		beego.NSRouter("/hosts/:id", &api.HostController{},
			"get:Get;put:Update;delete:Delete"),
		beego.NSRouter("/hosts/:id/status", &api.HostController{}, "post:ChangeStatus"),
		// dashboard
		beego.NSRouter("/dashboard/summary", &api.DashboardController{}, "get:Summary"),
		// global config
		beego.NSRouter("/global", &api.GlobalController{}, "get:Get;put:Update"),
		// app settings (mirrors nps.conf)
		beego.NSRouter("/settings", &api.SettingsController{}, "get:Get;put:Update"),
		// API tokens (multi-key + scope)
		beego.NSRouter("/tokens", &api.TokenController{}, "get:List;post:Create"),
		beego.NSRouter("/tokens/:id", &api.TokenController{},
			"get:Get;put:Update;delete:Delete"),
		beego.NSRouter("/tokens/:id/rotate", &api.TokenController{}, "post:Rotate"),
	)
	beego.AddNamespace(apiNs)

	// ---------- /auth/* legacy public endpoints --------------------------
	// Only the IP-whitelist flow remains; the md5(auth_key+ts) signature
	// has been removed in favour of API tokens (see /api/v1/tokens).
	beego.Router(web_base_url+"/auth/ipwhiteauth", &api.LegacyAuthController{}, "*:IpWhiteAuth")

	// ---------- SPA static handler ---------------------------------------
	spa := webui.Handler(web_base_url + "/ui")
	beego.Handler(web_base_url+"/ui", spa, true)
	beego.Handler(web_base_url+"/ui/*", spa, true)

	// "/" → "/ui/" so the historical landing page still works.
	beego.InsertFilter(web_base_url+"/", beego.BeforeRouter, func(ctx *context.Context) {
		p := ctx.Request.URL.Path
		if p == web_base_url || p == web_base_url+"/" {
			ctx.Redirect(302, web_base_url+"/ui/")
		}
	})
}
