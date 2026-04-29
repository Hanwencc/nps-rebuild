package api

import (
	"encoding/json"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/server"
	"github.com/astaxie/beego"
)

// AuthController exposes session/login endpoints. It does NOT inherit
// baseController.Prepare() (auth required) because login itself must
// be reachable anonymously — instead it embeds beego.Controller
// directly.
type AuthController struct {
	beego.Controller
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	IsAdmin   bool   `json:"isAdmin"`
	Username  string `json:"username"`
	ClientId  int    `json:"clientId"`
	Authed    bool   `json:"authed"`
	WebVer    string `json:"webVersion,omitempty"`
}

func (c *AuthController) ok(data interface{}) {
	c.Data["json"] = Envelope{Code: 0, Message: "ok", Data: data}
	c.ServeJSON()
	c.StopRun()
}

func (c *AuthController) fail(httpStatus, code int, msg string) {
	c.Ctx.Output.SetStatus(httpStatus)
	c.Data["json"] = Envelope{Code: code, Message: msg}
	c.ServeJSON()
	c.StopRun()
}

// ---------- naive brute-force throttle (mirrors legacy login logic) ----------

var apiIpRecord sync.Map

type apiLoginRecord struct {
	fails    int
	lastSeen time.Time
}

func clearApiIpRecord() {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(100) != 1 {
		return
	}
	apiIpRecord.Range(func(k, v interface{}) bool {
		r := v.(*apiLoginRecord)
		if time.Since(r.lastSeen) >= time.Minute {
			apiIpRecord.Delete(k)
		}
		return true
	})
}

// Login authenticates with username + password. Sets a session cookie
// on success.
func (c *AuthController) Login() {
	clearApiIpRecord()
	ip, _, _ := net.SplitHostPort(c.Ctx.Request.RemoteAddr)
	if v, ok := apiIpRecord.Load(ip); ok {
		r := v.(*apiLoginRecord)
		if time.Since(r.lastSeen) >= time.Minute {
			r.fails = 0
		}
		if r.fails >= 10 {
			c.fail(429, 4290, "too many failed attempts, please retry in 1 minute")
			return
		}
	}

	var req loginRequest
	body := c.Ctx.Input.RequestBody
	if len(body) > 0 && strings.Contains(c.Ctx.Input.Header("Content-Type"), "application/json") {
		_ = json.Unmarshal(body, &req)
	}
	if req.Username == "" {
		req.Username = c.GetString("username")
	}
	if req.Password == "" {
		req.Password = c.GetString("password")
	}
	if req.Username == "" || req.Password == "" {
		c.fail(400, 4000, "username and password required")
		return
	}

	if c.attemptLogin(req.Username, req.Password) {
		apiIpRecord.Delete(ip)
		c.SetSession("auth", true)
		isAdmin, _ := c.GetSession("isAdmin").(bool)
		username, _ := c.GetSession("username").(string)
		cid, _ := c.GetSession("clientId").(int)
		if isAdmin {
			username = beego.AppConfig.String("web_username")
		}
		c.ok(meResponse{IsAdmin: isAdmin, Username: username, ClientId: cid, Authed: true})
		return
	}

	r, _ := apiIpRecord.LoadOrStore(ip, &apiLoginRecord{lastSeen: time.Now()})
	rec := r.(*apiLoginRecord)
	rec.fails++
	rec.lastSeen = time.Now()
	c.fail(401, 4010, "username or password incorrect")
}

func (c *AuthController) attemptLogin(username, password string) bool {
	if password == beego.AppConfig.String("web_password") &&
		username == beego.AppConfig.String("web_username") {
		c.SetSession("isAdmin", true)
		c.DelSession("clientId")
		c.DelSession("username")
		server.Bridge.Register.Store(common.GetIpByAddr(c.Ctx.Input.IP()),
			time.Now().Add(2*time.Hour))
		return true
	}
	allowUser, _ := beego.AppConfig.Bool("allow_user_login")
	if !allowUser {
		return false
	}
	var ok bool
	file.GetDb().JsonDb.Clients.Range(func(_, value interface{}) bool {
		v := value.(*file.Client)
		if !v.Status || v.NoDisplay {
			return true
		}
		match := false
		if v.WebUserName == "" && v.WebPassword == "" {
			if username == "user" && v.VerifyKey == password {
				match = true
			}
		} else if v.WebUserName == username && v.WebPassword == password {
			match = true
		}
		if match {
			ok = true
			c.SetSession("isAdmin", false)
			c.SetSession("clientId", v.Id)
			c.SetSession("username", v.WebUserName)
			return false
		}
		return true
	})
	return ok
}

// Health is a public probe used by load balancers / SPA bootstrap.
func (c *AuthController) Health() {
	c.ok(map[string]string{"status": "ok"})
}

// Logout clears session.
func (c *AuthController) Logout() {
	c.SetSession("auth", false)
	c.DelSession("isAdmin")
	c.DelSession("clientId")
	c.DelSession("username")
	c.ok(nil)
}

// Me returns current authenticated user information.
func (c *AuthController) Me() {
	if c.GetSession("auth") != true {
		c.fail(401, 4010, "unauthenticated")
		return
	}
	isAdmin, _ := c.GetSession("isAdmin").(bool)
	username, _ := c.GetSession("username").(string)
	cid, _ := c.GetSession("clientId").(int)
	if isAdmin {
		username = beego.AppConfig.String("web_username")
	}
	c.ok(meResponse{IsAdmin: isAdmin, Username: username, ClientId: cid, Authed: true})
}
