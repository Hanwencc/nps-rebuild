package api

import (
	"encoding/json"
	"strings"

	"ehang.io/nps/lib/file"
)

// GlobalController exposes the global server settings (currently the
// black-IP list and server URL). Admin-only.
type GlobalController struct {
	baseController
}

type globalPayload struct {
	BlackIpList []string `json:"blackIpList"`
	ServerUrl   string   `json:"serverUrl"`
}

// Get GET /api/v1/global
func (c *GlobalController) Get() {
	if !c.currentIsAdmin() {
		c.forbidden("permission denied")
		return
	}
	g := file.GetDb().GetGlobal()
	if g == nil {
		c.ok(globalPayload{BlackIpList: []string{}})
		return
	}
	c.ok(globalPayload{
		BlackIpList: g.BlackIpList,
		ServerUrl:   g.ServerUrl,
	})
}

// Update PUT /api/v1/global
func (c *GlobalController) Update() {
	if !c.currentIsAdmin() {
		c.forbidden("permission denied")
		return
	}
	var p globalPayload
	body := c.Ctx.Input.RequestBody
	if len(body) > 0 {
		if err := json.Unmarshal(body, &p); err != nil {
			c.badRequest("invalid JSON body: " + err.Error())
			return
		}
	}
	cleaned := make([]string, 0, len(p.BlackIpList))
	seen := map[string]struct{}{}
	for _, raw := range p.BlackIpList {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		cleaned = append(cleaned, s)
	}
	g := &file.Glob{
		BlackIpList: cleaned,
		ServerUrl:   strings.TrimSpace(p.ServerUrl),
	}
	if err := file.GetDb().SaveGlobal(g); err != nil {
		c.serverErr(err.Error())
		return
	}
	c.okMsg("saved")
}
