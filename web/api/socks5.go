package api

import (
	"ehang.io/nps/server"
)

// Socks5Controller surfaces runtime info about the Phase 9 shared
// SOCKS5 gateway. Read-only; configuration goes through SettingsController.
//
//	GET  /api/v1/socks5/gateway   →  {listening, addr, port, routes}
//
// The endpoint is required by the SPA tunnel-edit form so it can warn
// the operator when socks5_shared_port=0 (no listener) before they
// create a routing entry.
type Socks5Controller struct {
	baseController
}

// Gateway returns the live status snapshot. Falls back to a disabled
// stub if cmd/nps hasn't published the callback yet (e.g. during
// early boot or in unit tests that import the api package directly).
func (c *Socks5Controller) Gateway() {
	resp := map[string]interface{}{
		"listening": false,
		"addr":      "",
		"port":      0,
		"routes":    0,
	}
	if server.Socks5GatewayInfoFn != nil {
		listening, addr, port, routes := server.Socks5GatewayInfoFn()
		resp["listening"] = listening
		resp["addr"] = addr
		resp["port"] = port
		resp["routes"] = routes
	}
	c.ok(resp)
}
