package api

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/rate"
	"ehang.io/nps/server"
	"github.com/astaxie/beego"
)

// ClientController exposes client CRUD endpoints under /api/v1/clients.
type ClientController struct {
	baseController
}

// ----- payloads ------------------------------------------------------------

type clientPayload struct {
	VerifyKey       string   `json:"vkey"`
	Remark          string   `json:"remark"`
	UserName        string   `json:"u"`
	Password        string   `json:"p"`
	Compress        bool     `json:"compress"`
	Crypt           bool     `json:"crypt"`
	ConfigConnAllow bool     `json:"configConnAllow"`
	RateLimit       int      `json:"rateLimit"`
	MaxConn         int      `json:"maxConn"`
	MaxTunnel       int      `json:"maxTunnel"`
	FlowLimit       int64    `json:"flowLimit"`
	WebUsername     string   `json:"webUsername"`
	WebPassword     string   `json:"webPassword"`
	BlackIpList     []string `json:"blackIpList"`
	IpWhite         bool     `json:"ipWhite"`
	IpWhitePass     string   `json:"ipWhitePass"`
	IpWhiteList     []string `json:"ipWhiteList"`
}

type changeStatusPayload struct {
	Status bool `json:"status"`
}

func (c *ClientController) decode(target interface{}) bool {
	body := c.Ctx.Input.RequestBody
	if len(body) == 0 {
		return true
	}
	if err := json.Unmarshal(body, target); err != nil {
		c.badRequest("invalid JSON body: " + err.Error())
		return false
	}
	return true
}

// ----- handlers ------------------------------------------------------------

// List GET /api/v1/clients?offset=&limit=&search=&sort=&order=
func (c *ClientController) List() {
	offset, _ := c.GetInt("offset")
	limit, _ := c.GetInt("limit")
	if limit <= 0 {
		limit = 10
	}
	search := c.GetString("search")
	sort := c.GetString("sort")
	order := c.GetString("order")

	clientId := 0
	if !c.currentIsAdmin() {
		clientId = c.currentClientId()
	}
	list, total := server.GetClientList(offset, limit, search, sort, order, clientId)
	c.ok(Page{Total: int64(total), Items: list})
}

// Get GET /api/v1/clients/:id
func (c *ClientController) Get() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	cli, err := file.GetDb().GetClient(id)
	if err != nil {
		c.notFound("client not found")
		return
	}
	if !c.currentIsAdmin() && cli.Id != c.currentClientId() {
		c.forbidden("permission denied")
		return
	}
	c.ok(cli)
}

// Create POST /api/v1/clients
func (c *ClientController) Create() {
	if !c.currentIsAdmin() {
		c.forbidden("admin only")
		return
	}
	var p clientPayload
	if !c.decode(&p) {
		return
	}
	id := int(file.GetDb().JsonDb.GetClientId())
	t := &file.Client{
		Id:        id,
		Status:    true,
		VerifyKey: p.VerifyKey,
		Remark:    p.Remark,
		Cnf: &file.Config{
			U:        p.UserName,
			P:        p.Password,
			Compress: p.Compress,
			Crypt:    p.Crypt,
		},
		ConfigConnAllow: p.ConfigConnAllow,
		RateLimit:       p.RateLimit,
		MaxConn:         p.MaxConn,
		MaxTunnelNum:    p.MaxTunnel,
		WebUserName:     p.WebUsername,
		WebPassword:     p.WebPassword,
		Flow:            &file.Flow{FlowLimit: p.FlowLimit},
		BlackIpList:     dedupNonEmpty(p.BlackIpList),
		IpWhite:         p.IpWhite,
		IpWhitePass:     p.IpWhitePass,
		IpWhiteList:     dedupNonEmpty(p.IpWhiteList),
		CreateTime:      time.Now().Format("2006-01-02 15:04:05"),
	}
	if err := file.GetDb().NewClient(t); err != nil {
		c.badRequest(err.Error())
		return
	}
	c.ok(map[string]int{"id": id})
}

// Update PUT /api/v1/clients/:id
func (c *ClientController) Update() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	cli, err := file.GetDb().GetClient(id)
	if err != nil {
		c.notFound("client not found")
		return
	}
	if !c.currentIsAdmin() && cli.Id != c.currentClientId() {
		c.forbidden("permission denied")
		return
	}
	var p clientPayload
	if !c.decode(&p) {
		return
	}

	if p.WebUsername != "" {
		if p.WebUsername == beego.AppConfig.String("web_username") ||
			!file.GetDb().VerifyUserName(p.WebUsername, cli.Id) {
			c.badRequest("web login username duplicate, please reset")
			return
		}
	}

	isAdmin := c.currentIsAdmin()
	if isAdmin {
		if p.VerifyKey != "" && !file.GetDb().VerifyVkey(p.VerifyKey, cli.Id) {
			c.badRequest("vkey duplicate, please reset")
			return
		}
		if p.VerifyKey != "" {
			cli.VerifyKey = p.VerifyKey
		}
		cli.Flow.FlowLimit = p.FlowLimit
		cli.RateLimit = p.RateLimit
		cli.MaxConn = p.MaxConn
		cli.MaxTunnelNum = p.MaxTunnel
	}
	cli.Remark = p.Remark
	cli.Cnf.U = p.UserName
	cli.Cnf.P = p.Password
	cli.Cnf.Compress = p.Compress
	cli.Cnf.Crypt = p.Crypt
	if allow, err := beego.AppConfig.Bool("allow_user_change_username"); isAdmin || (err == nil && allow) {
		cli.WebUserName = p.WebUsername
	}
	cli.WebPassword = p.WebPassword
	cli.ConfigConnAllow = p.ConfigConnAllow
	cli.IpWhite = p.IpWhite
	cli.IpWhitePass = p.IpWhitePass
	cli.IpWhiteList = dedupNonEmpty(p.IpWhiteList)
	cli.BlackIpList = dedupNonEmpty(p.BlackIpList)

	if cli.Rate != nil {
		cli.Rate.Stop()
	}
	if cli.RateLimit > 0 {
		cli.Rate = rate.NewRate(int64(cli.RateLimit * 1024))
	} else {
		cli.Rate = rate.NewRate((2 << 23) * 1024)
	}
	cli.Rate.Start()

	file.GetDb().JsonDb.StoreClientsToJsonFile()
	c.okMsg("updated")
}

// Delete DELETE /api/v1/clients/:id
func (c *ClientController) Delete() {
	if !c.currentIsAdmin() {
		c.forbidden("admin only")
		return
	}
	id, ok := c.parseId()
	if !ok {
		return
	}
	if err := file.GetDb().DelClient(id); err != nil {
		c.serverErr(err.Error())
		return
	}
	server.DelTunnelAndHostByClientId(id, false)
	server.DelClientConnect(id)
	c.okMsg("deleted")
}

// ChangeStatus POST /api/v1/clients/:id/status  body: { "status": true }
func (c *ClientController) ChangeStatus() {
	if !c.currentIsAdmin() {
		c.forbidden("admin only")
		return
	}
	id, ok := c.parseId()
	if !ok {
		return
	}
	cli, err := file.GetDb().GetClient(id)
	if err != nil {
		c.notFound("client not found")
		return
	}
	var p changeStatusPayload
	if !c.decode(&p) {
		return
	}
	cli.Status = p.Status
	if !p.Status {
		server.DelClientConnect(cli.Id)
	}
	file.GetDb().JsonDb.StoreClientsToJsonFile()
	c.okMsg("updated")
}

// QuickInfo GET /api/v1/clients/:id/quickinfo  → bridge meta useful for
// rendering "quickly install" command on the SPA.
func (c *ClientController) QuickInfo() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	cli, err := file.GetDb().GetClient(id)
	if err != nil {
		c.notFound("client not found")
		return
	}
	host := c.Ctx.Request.Host
	ip := common.GetIpByAddr(host)
	bridgePort := server.Bridge.TunnelPort
	bridgeType := beego.AppConfig.String("bridge_type")
	tlsPort := beego.AppConfig.DefaultInt("tls_bridge_port", 8025)
	c.ok(map[string]interface{}{
		"id":         cli.Id,
		"vkey":       cli.VerifyKey,
		"remark":     cli.Remark,
		"ip":         ip,
		"bridgePort": bridgePort,
		"bridgeType": bridgeType,
		"tlsPort":    tlsPort,
	})
}

// ----- helpers -------------------------------------------------------------

func (c *ClientController) parseId() (int, bool) {
	v := c.Ctx.Input.Param(":id")
	if v == "" {
		c.badRequest("missing id")
		return 0, false
	}
	id, err := strconv.Atoi(v)
	if err != nil {
		c.badRequest("invalid id")
		return 0, false
	}
	return id, true
}

func dedupNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
