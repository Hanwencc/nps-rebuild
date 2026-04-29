package api

import (
	"ehang.io/nps/lib/version"
	"ehang.io/nps/server"
	"github.com/astaxie/beego"
)

// DashboardController exposes server-wide summary statistics consumed
// by the SPA dashboard. Admin-only.
type DashboardController struct {
	baseController
}

// Summary GET /api/v1/dashboard/summary
//
// Returns a curated subset of the legacy `server.GetDashboardData()`
// payload, but with stable camelCase keys and explicit numeric types
// so the SPA can rely on the schema.
func (c *DashboardController) Summary() {
	if !c.currentIsAdmin() {
		c.forbidden("permission denied")
		return
	}
	raw := server.GetDashboardData()

	// helper to coerce numeric values from the raw map.
	num := func(key string) float64 {
		v, ok := raw[key]
		if !ok {
			return 0
		}
		switch x := v.(type) {
		case int:
			return float64(x)
		case int64:
			return float64(x)
		case float64:
			return x
		case uint64:
			return float64(x)
		}
		return 0
	}
	str := func(key string) string {
		s, _ := raw[key].(string)
		return s
	}

	out := map[string]interface{}{
		"version":     version.VERSION,
		"bridgeType":  str("bridgeType"),
		"bridgePort":  beego.AppConfig.String("bridge_port"),
		"tlsEnable":   beego.AppConfig.DefaultBool("tls_enable", false),
		"tlsBridgePort": beego.AppConfig.DefaultInt("tls_bridge_port", 8025),
		"serverIp":    str("serverIp"),
		"p2pPort":     str("p2pPort"),
		"logLevel":    str("logLevel"),
		"ipLimit":     str("ipLimit"),
		"flowStoreInterval": str("flowStoreInterval"),
		"httpProxyPort":     str("httpProxyPort"),
		"httpsProxyPort":    str("httpsProxyPort"),

		"clientCount":       num("clientCount"),
		"clientOnlineCount": num("clientOnlineCount"),
		"hostCount":         num("hostCount"),
		"tunnelCount": map[string]float64{
			"tcp":       num("tcpC"),
			"udp":       num("udpCount"),
			"socks5":    num("socks5Count"),
			"httpProxy": num("httpProxyCount"),
			"secret":    num("secretCount"),
			"p2p":       num("p2pCount"),
		},

		"flow": map[string]float64{
			"in":  num("inletFlowCount"),
			"out": num("exportFlowCount"),
		},
		"connections": num("tcpCount"),

		"system": map[string]float64{
			"cpu":     num("cpu"),
			"mem":     num("virtual_mem"),
			"swap":    num("swap_mem"),
			"ioSend":  num("io_send"),
			"ioRecv":  num("io_recv"),
		},
		"load": str("load"),
	}

	// chart history: sys1..sys10 (legacy) -> []sys with {cpu,mem,load}
	history := make([]map[string]interface{}, 0, 10)
	for i := 1; i <= 10; i++ {
		key := "sys" + itoa(i)
		v, ok := raw[key]
		if !ok {
			continue
		}
		if m, ok := v.(map[string]interface{}); ok {
			history = append(history, m)
		}
	}
	out["history"] = history

	c.ok(out)
}

func itoa(i int) string {
	// avoid importing strconv just for this; tiny fixed range 1..10.
	switch i {
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 4:
		return "4"
	case 5:
		return "5"
	case 6:
		return "6"
	case 7:
		return "7"
	case 8:
		return "8"
	case 9:
		return "9"
	case 10:
		return "10"
	}
	return ""
}
