package api

import (
	"encoding/json"
	"strconv"

	"ehang.io/nps/lib/file"
	"ehang.io/nps/server"
	"ehang.io/nps/server/tool"
)

// TunnelController exposes tunnel CRUD endpoints under /api/v1/tunnels.
type TunnelController struct {
	baseController
}

// tunnelPayload mirrors the legacy IndexController.Add/Edit form fields,
// using camelCase JSON keys so it lines up with the SPA's TS types.
type tunnelPayload struct {
	ClientId     int    `json:"clientId"`
	Mode         string `json:"mode"`
	Port         int    `json:"port"`
	ServerIp     string `json:"serverIp"`
	Target       string `json:"target"`
	LocalProxy   bool   `json:"localProxy"`
	Password     string `json:"password"`
	Remark       string `json:"remark"`
	LocalPath    string `json:"localPath"`
	StripPre     string `json:"stripPre"`
	ProtoVersion string `json:"protoVersion"`
}

func (c *TunnelController) decode(target interface{}) bool {
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

func (c *TunnelController) parseId() (int, bool) {
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

// canAccessClient returns true if the current session is allowed to
// touch resources belonging to the given client id.
func (c *TunnelController) canAccessClient(clientId int) bool {
	if c.currentIsAdmin() {
		return true
	}
	return clientId != 0 && clientId == c.currentClientId()
}

// List GET /api/v1/tunnels?mode=&clientId=&offset=&limit=&search=&sort=&order=
func (c *TunnelController) List() {
	offset, _ := c.GetInt("offset")
	limit, _ := c.GetInt("limit")
	if limit <= 0 {
		limit = 10
	}
	mode := c.GetString("mode")
	clientId, _ := c.GetInt("clientId")
	search := c.GetString("search")
	sortField := c.GetString("sort")
	order := c.GetString("order")

	if !c.currentIsAdmin() {
		// Sub-users may only list their own tunnels.
		clientId = c.currentClientId()
	}

	list, total := server.GetTunnel(offset, limit, mode, clientId, search, sortField, order)
	c.ok(Page{Total: int64(total), Items: list})
}

// Get GET /api/v1/tunnels/:id
func (c *TunnelController) Get() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	t, err := file.GetDb().GetTask(id)
	if err != nil {
		c.notFound("tunnel not found")
		return
	}
	if !c.canAccessClient(t.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	c.ok(t)
}

// Create POST /api/v1/tunnels
func (c *TunnelController) Create() {
	var p tunnelPayload
	if !c.decode(&p) {
		return
	}
	if p.Mode == "" {
		c.badRequest("mode is required")
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

	id := int(file.GetDb().JsonDb.GetTaskId())
	t := &file.Tunnel{
		Id:           id,
		Port:         p.Port,
		ServerIp:     p.ServerIp,
		Mode:         p.Mode,
		Status:       true,
		Remark:       p.Remark,
		Password:     p.Password,
		LocalPath:    p.LocalPath,
		StripPre:     p.StripPre,
		ProtoVersion: p.ProtoVersion,
		Target:       &file.Target{TargetStr: p.Target, LocalProxy: p.LocalProxy},
		Client:       cli,
		Flow:         &file.Flow{},
	}
	if t.Port <= 0 {
		t.Port = tool.GenerateServerPort(t.Mode)
	}
	if t.Mode != "secret" && t.Mode != "p2p" {
		if !tool.TestServerPort(t.Port, t.Mode) {
			c.badRequest("the port cannot be opened (occupied or not allowed)")
			return
		}
	}
	if err := file.GetDb().NewTask(t); err != nil {
		c.badRequest(err.Error())
		return
	}
	if err := server.AddTask(t); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.ok(map[string]int{"id": id})
}

// Update PUT /api/v1/tunnels/:id
func (c *TunnelController) Update() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	t, err := file.GetDb().GetTask(id)
	if err != nil {
		c.notFound("tunnel not found")
		return
	}
	if !c.canAccessClient(t.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	var p tunnelPayload
	if !c.decode(&p) {
		return
	}
	cli := t.Client
	if p.ClientId != 0 && p.ClientId != cli.Id {
		if !c.currentIsAdmin() {
			c.forbidden("only admin can reassign client")
			return
		}
		newCli, err := file.GetDb().GetClient(p.ClientId)
		if err != nil {
			c.badRequest("client not found")
			return
		}
		t.Client = newCli
	}
	if p.Port != 0 && p.Port != t.Port {
		t.Port = p.Port
		if t.Port <= 0 {
			t.Port = tool.GenerateServerPort(t.Mode)
		}
		if t.Mode != "secret" && t.Mode != "p2p" {
			if !tool.TestServerPort(t.Port, t.Mode) {
				c.badRequest("the port cannot be opened (occupied or not allowed)")
				return
			}
		}
	}
	if p.ServerIp != "" {
		t.ServerIp = p.ServerIp
	}
	if p.Mode != "" {
		t.Mode = p.Mode
	}
	t.Target = &file.Target{TargetStr: p.Target, LocalProxy: p.LocalProxy}
	t.Password = p.Password
	t.LocalPath = p.LocalPath
	t.StripPre = p.StripPre
	t.ProtoVersion = p.ProtoVersion
	t.Remark = p.Remark
	if err := file.GetDb().UpdateTask(t); err != nil {
		c.serverErr(err.Error())
		return
	}
	_ = server.StopServer(t.Id)
	_ = server.StartTask(t.Id)
	c.okMsg("updated")
}

// Delete DELETE /api/v1/tunnels/:id
func (c *TunnelController) Delete() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	t, err := file.GetDb().GetTask(id)
	if err != nil {
		c.notFound("tunnel not found")
		return
	}
	if !c.canAccessClient(t.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	if err := server.DelTask(id); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.okMsg("deleted")
}

// Start POST /api/v1/tunnels/:id/start
func (c *TunnelController) Start() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	t, err := file.GetDb().GetTask(id)
	if err != nil {
		c.notFound("tunnel not found")
		return
	}
	if !c.canAccessClient(t.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	if err := server.StartTask(id); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.okMsg("started")
}

// Stop POST /api/v1/tunnels/:id/stop
func (c *TunnelController) Stop() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	t, err := file.GetDb().GetTask(id)
	if err != nil {
		c.notFound("tunnel not found")
		return
	}
	if !c.canAccessClient(t.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	if err := server.StopServer(id); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.okMsg("stopped")
}

// Copy POST /api/v1/tunnels/:id/copy → clones an existing tunnel.
func (c *TunnelController) Copy() {
	id, ok := c.parseId()
	if !ok {
		return
	}
	old, err := file.GetDb().GetTask(id)
	if err != nil {
		c.notFound("tunnel not found")
		return
	}
	if !c.canAccessClient(old.Client.Id) {
		c.forbidden("permission denied")
		return
	}
	cli, err := file.GetDb().GetClient(old.Client.Id)
	if err != nil {
		c.badRequest("client not found")
		return
	}
	newId := int(file.GetDb().JsonDb.GetTaskId())
	newTask := &file.Tunnel{
		Id:           newId,
		Client:       cli,
		Port:         tool.GenerateServerPort(old.Mode),
		ServerIp:     old.ServerIp,
		Mode:         old.Mode,
		Status:       true,
		Remark:       old.Remark,
		Password:     old.Password,
		LocalPath:    old.LocalPath,
		StripPre:     old.StripPre,
		ProtoVersion: old.ProtoVersion,
		Target:       old.Target,
		Flow:         &file.Flow{},
	}
	if newTask.Mode != "secret" && newTask.Mode != "p2p" {
		if !tool.TestServerPort(newTask.Port, newTask.Mode) {
			c.badRequest("the port cannot be opened (occupied or not allowed)")
			return
		}
	}
	if cli.MaxTunnelNum != 0 && cli.GetTunnelNum() >= cli.MaxTunnelNum {
		c.badRequest("the number of tunnels exceeds the limit")
		return
	}
	if err := file.GetDb().NewTask(newTask); err != nil {
		c.badRequest(err.Error())
		return
	}
	if err := server.AddTask(newTask); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.ok(map[string]int{"id": newId})
}
