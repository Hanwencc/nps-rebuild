package api

import (
	"encoding/json"
	"strconv"

	"ehang.io/nps/lib/file"
)

// HostController exposes domain (HTTP/HTTPS) host CRUD endpoints under
// /api/v1/hosts.
type HostController struct {
	baseController
}

type hostPayload struct {
	ClientId     int    `json:"clientId"`
	Host         string `json:"host"`
	Target       string `json:"target"`
	LocalProxy   bool   `json:"localProxy"`
	HeaderChange string `json:"header"`
	HostChange   string `json:"hostchange"`
	Remark       string `json:"remark"`
	Location     string `json:"location"`
	Scheme       string `json:"scheme"`
	KeyFilePath  string `json:"keyFilePath"`
	CertFilePath string `json:"certFilePath"`
	AutoHttps    bool   `json:"autoHttps"`
}

func (c *HostController) decode(target interface{}) bool {
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

func (c *HostController) parseId() (int, bool) {
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

func (c *HostController) canAccessClient(clientId int) bool {
	if c.currentIsAdmin() {
		return true
	}
	return clientId != 0 && clientId == c.currentClientId()
}

// List GET /api/v1/hosts?clientId=&offset=&limit=&search=
func (c *HostController) List() {
	offset, _ := c.GetInt("offset")
	limit, _ := c.GetInt("limit")
	if limit <= 0 {
		limit = 10
	}
	clientId, _ := c.GetInt("clientId")
	search := c.GetString("search")

	if !c.currentIsAdmin() {
		clientId = c.currentClientId()
	}

	list, total := file.GetDb().GetHost(offset, limit, clientId, search)
	c.ok(Page{Total: int64(total), Items: list})
}

// Get GET /api/v1/hosts/:id
func (c *HostController) Get() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	h, err := file.GetDb().GetHostById(id)
	if err != nil {
		c.notFound("host not found")
		return
	}
	if !c.canAccessClient(h.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	c.ok(h)
}

// Create POST /api/v1/hosts
func (c *HostController) Create() {
	var p hostPayload
	if !c.decode(&p) {
		return
	}
	if !c.canAccessClient(p.ClientId) {
		c.forbidden("permission denied")
		return
	}
	cli, err := file.GetDb().GetClient(p.ClientId)
	if err != nil {
		c.badRequest("client not found")
		return
	}
	if cli.MaxTunnelNum != 0 && cli.GetTunnelNum() >= cli.MaxTunnelNum {
		c.badRequest("the number of tunnels exceeds the limit")
		return
	}
	id := int(file.GetDb().JsonDb.GetHostId())
	h := &file.Host{
		Id:           id,
		Host:         p.Host,
		Target:       &file.Target{TargetStr: p.Target, LocalProxy: p.LocalProxy},
		HeaderChange: p.HeaderChange,
		HostChange:   p.HostChange,
		Remark:       p.Remark,
		Location:     p.Location,
		Flow:         &file.Flow{},
		Scheme:       p.Scheme,
		KeyFilePath:  p.KeyFilePath,
		CertFilePath: p.CertFilePath,
		AutoHttps:    p.AutoHttps,
		Client:       cli,
	}
	if h.Scheme == "http" {
		h.AutoHttps = false
	}
	if err := file.GetDb().NewHost(h); err != nil {
		c.badRequest(err.Error())
		return
	}
	c.ok(map[string]int{"id": id})
}

// Update PUT /api/v1/hosts/:id
func (c *HostController) Update() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	h, err := file.GetDb().GetHostById(id)
	if err != nil {
		c.notFound("host not found")
		return
	}
	if !c.canAccessClient(h.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	var p hostPayload
	if !c.decode(&p) {
		return
	}
	if h.Host != p.Host {
		tmp := &file.Host{Host: p.Host, Location: p.Location, Scheme: p.Scheme}
		if file.GetDb().IsHostExist(tmp) {
			c.badRequest("host has exist")
			return
		}
	}
	if p.ClientId != 0 && p.ClientId != h.Client.Id {
		if !c.currentIsAdmin() {
			c.forbidden("only admin can reassign client")
			return
		}
		newCli, err := file.GetDb().GetClient(p.ClientId)
		if err != nil {
			c.badRequest("client not found")
			return
		}
		h.Client = newCli
	}
	h.Host = p.Host
	h.Target = &file.Target{TargetStr: p.Target, LocalProxy: p.LocalProxy}
	h.HeaderChange = p.HeaderChange
	h.HostChange = p.HostChange
	h.Remark = p.Remark
	h.Location = p.Location
	h.Scheme = p.Scheme
	h.KeyFilePath = p.KeyFilePath
	h.CertFilePath = p.CertFilePath
	h.AutoHttps = p.AutoHttps
	if h.Scheme == "http" {
		h.AutoHttps = false
	}
	if err := file.GetDb().UpdateHost(h); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.okMsg("updated")
}

// Delete DELETE /api/v1/hosts/:id
func (c *HostController) Delete() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	h, err := file.GetDb().GetHostById(id)
	if err != nil {
		c.notFound("host not found")
		return
	}
	if !c.canAccessClient(h.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	if err := file.GetDb().DelHost(id); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.okMsg("deleted")
}

// ChangeStatus POST /api/v1/hosts/:id/status  body: { "status": true }
//
// `status: true` means enabled (IsClose=false).
func (c *HostController) ChangeStatus() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	h, err := file.GetDb().GetHostById(id)
	if err != nil {
		c.notFound("host not found")
		return
	}
	if !c.canAccessClient(h.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	var p changeStatusPayload
	if !c.decode(&p) {
		return
	}
	h.IsClose = !p.Status
	if err := file.GetDb().UpdateHost(h); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.okMsg("updated")
}
