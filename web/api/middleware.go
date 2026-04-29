package api

import (
	"net"
	"strings"
	"time"

	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/file/sqlitedb"
	"github.com/astaxie/beego/logs"
)

// authFreePaths are public endpoints that bypass the auth check.
var authFreePaths = map[string]struct{}{
	"/api/v1/auth/login":  {},
	"/api/v1/auth/health": {},
}

// Prepare implements the dual-track authentication used after the
// migration. A request is authenticated if EITHER:
//
//  1. it carries a valid Beego session (auth=true), OR
//  2. it presents a valid API token (X-Api-Key + X-Api-Secret, or
//     `Authorization: Bearer <keyId>.<secret>`).
//
// The legacy md5 signature flow has been removed; machine-to-machine
// callers must use API tokens.
func (c *baseController) Prepare() {
	c.Ctx.Output.Header("Cache-Control", "no-store")

	path := c.Ctx.Request.URL.Path
	if _, ok := authFreePaths[path]; ok {
		return
	}

	// 1) API Token (preferred for machine-to-machine)
	if tok, ok := c.verifyApiToken(); ok {
		c.Data["isAdmin"] = true // tokens are admin-scope today
		c.Data["apiToken"] = tok
		return
	}

	// 2) session cookie
	if c.GetSession("auth") == true {
		isAdmin, _ := c.GetSession("isAdmin").(bool)
		c.Data["isAdmin"] = isAdmin
		if !isAdmin {
			if cid, ok := c.GetSession("clientId").(int); ok {
				c.Data["clientId"] = cid
			}
		}
		return
	}

	c.unauthorized("authentication required")
}

// verifyApiToken returns the matched token (and true) when the request
// presents a valid API credential. Accepted forms:
//
//	X-Api-Key: <keyId>          + X-Api-Secret: <secret>
//	Authorization: Bearer <keyId>.<secret>
//	Authorization: <keyId>.<secret>          (Bearer prefix optional)
//	Authorization: <keyId> <secret>          (space-separated)
//
// Path/method/IP scope is enforced here.
func (c *baseController) verifyApiToken() (*file.ApiToken, bool) {
	keyId := strings.TrimSpace(c.Ctx.Input.Header("X-Api-Key"))
	secret := strings.TrimSpace(c.Ctx.Input.Header("X-Api-Secret"))

	if keyId == "" || secret == "" {
		raw := strings.TrimSpace(c.Ctx.Input.Header("Authorization"))
		// strip optional "Bearer " / "Token " scheme prefix
		for _, prefix := range []string{"Bearer ", "bearer ", "Token ", "token "} {
			if strings.HasPrefix(raw, prefix) {
				raw = strings.TrimSpace(raw[len(prefix):])
				break
			}
		}
		if raw != "" {
			// Try "<keyId>.<secret>" first, then "<keyId> <secret>".
			if i := strings.IndexByte(raw, '.'); i > 0 && !strings.ContainsAny(raw[:i], " \t") {
				keyId = raw[:i]
				secret = raw[i+1:]
			} else if i := strings.IndexAny(raw, " \t"); i > 0 {
				keyId = raw[:i]
				secret = strings.TrimSpace(raw[i+1:])
			}
		}
	}
	if keyId == "" || secret == "" {
		return nil, false
	}

	store := sqlitedb.From(file.GetDb())
	if store == nil {
		logs.Warn("api-token verify: sqlite store unavailable")
		return nil, false
	}
	tok, err := store.FindApiTokenByKeyId(keyId)
	if err != nil {
		return nil, false
	}
	if !tok.VerifySecret(secret) {
		logs.Warn("api-token bad secret: keyId=%s ip=%s", keyId, c.Ctx.Input.IP())
		return nil, false
	}

	ip, _, _ := net.SplitHostPort(c.Ctx.Request.RemoteAddr)
	if ip == "" {
		ip = c.Ctx.Input.IP()
	}
	if err := tok.MatchesRequest(c.Ctx.Request.Method, c.Ctx.Request.URL.Path, ip); err != nil {
		logs.Warn("api-token rejected: keyId=%s reason=%s ip=%s", keyId, err.Error(), ip)
		return nil, false
	}
	// Persist last-used metadata. Conditional UPDATE inside the store
	// debounces writes to <=1/sec per (id, ip).
	if err := store.TouchApiToken(tok.Id, ip, time.Now().Unix()); err != nil {
		logs.Warn("api-token touch failed: keyId=%s err=%v", keyId, err)
	}
	return tok, true
}

func (c *baseController) currentIsAdmin() bool {
	v, _ := c.Data["isAdmin"].(bool)
	return v
}

func (c *baseController) currentClientId() int {
	v, _ := c.Data["clientId"].(int)
	return v
}
