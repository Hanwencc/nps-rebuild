package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ehang.io/nps/lib/file/sqlitedb"
	"ehang.io/nps/server"
	"ehang.io/nps/server/proxy"
	"ehang.io/nps/server/tool"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// socks5_gateway.go owns the lifecycle of the global SOCKS5 gateway
// (Phase 9). The gateway listens on socks5_shared_ip:socks5_shared_port
// and dispatches each connection to the NPC client identified by the
// SOCKS5 username (matched against tasks.username for mode=socks5).
//
// Hot-restart contract is identical to hothttp.go's: settings hooks
// schedule a debounced restart; restart closes the previous listener,
// briefly waits for the socket to free, probes the new port, and brings
// up a fresh server.
//
// The gateway is OPTIONAL: if socks5_shared_port is zero (the default
// for fresh installs / upgrades) no listener is started and SOCKS5
// tasks behave as inert routing entries — admin UI surfaces this as
// "网关未启用" and offers a deep link to the settings page.

var (
	socks5GatewayMu sync.Mutex
	socks5GatewaySvc *proxy.Sock5SharedServer
	socks5GatewayAddr string

	socks5RestartTimerMu sync.Mutex
	socks5RestartTimer   *time.Timer
)

// startSocks5Gateway is called once from cmd/nps/nps.go after
// server.StartNewServer + adoptInitialHttpHostServer. Subsequent
// changes go through scheduleSocks5Restart.
func startSocks5Gateway() {
	port := beego.AppConfig.DefaultInt("socks5_shared_port", 0)
	if port <= 0 {
		logs.Info("socks5 gateway disabled (socks5_shared_port=0); set it via the admin UI to enable")
		return
	}
	if err := bringUpSocks5Gateway(port); err != nil {
		logs.Error("socks5 gateway start failed: %v", err)
	}
}

// scheduleSocks5Restart debounces multi-key saves so changing both
// IP and port from the UI only triggers one restart.
func scheduleSocks5Restart(reason string) {
	socks5RestartTimerMu.Lock()
	defer socks5RestartTimerMu.Unlock()
	if socks5RestartTimer != nil {
		socks5RestartTimer.Stop()
	}
	socks5RestartTimer = time.AfterFunc(200*time.Millisecond, func() {
		if err := restartSocks5Gateway(); err != nil {
			logs.Error("socks5 gateway hot-restart failed (%s): %v", reason, err)
			return
		}
		logs.Info("socks5 gateway hot-restarted (%s)", reason)
	})
}

func restartSocks5Gateway() error {
	socks5GatewayMu.Lock()
	defer socks5GatewayMu.Unlock()

	port := beego.AppConfig.DefaultInt("socks5_shared_port", 0)

	// Tear down whatever is running before deciding what to do next so
	// the new port is free to probe and the old config is gone if the
	// operator just disabled the gateway by setting port=0.
	if socks5GatewaySvc != nil {
		if err := socks5GatewaySvc.Close(); err != nil {
			logs.Warn("socks5 gateway: close old listener: %v", err)
		}
		socks5GatewaySvc = nil
		socks5GatewayAddr = ""
		// Brief pause so the OS can release the listening socket
		// before we try to probe / re-bind on the same port.
		time.Sleep(150 * time.Millisecond)
	}

	if port <= 0 {
		logs.Info("socks5 gateway: stopped (port=0)")
		return nil
	}
	return bringUpSocks5GatewayLocked(port)
}

// bringUpSocks5Gateway acquires the mutex and starts a fresh listener.
// Used for the initial start; restartSocks5Gateway uses the *Locked
// variant since it already owns the mutex.
func bringUpSocks5Gateway(port int) error {
	socks5GatewayMu.Lock()
	defer socks5GatewayMu.Unlock()
	return bringUpSocks5GatewayLocked(port)
}

func bringUpSocks5GatewayLocked(port int) error {
	if server.Bridge == nil {
		return errors.New("bridge not initialised yet")
	}
	if !tool.TestServerPort(port, "tcp") {
		return fmt.Errorf("socks5_shared_port %d is not bindable (in use or not in allow_ports)", port)
	}
	ip := beego.AppConfig.DefaultString("socks5_shared_ip", "0.0.0.0")
	addr := proxy.SafeBindAddr(ip, port)
	svc := proxy.NewSock5SharedServer(server.Bridge, addr)
	go func() {
		if err := svc.Start(); err != nil {
			logs.Error("socks5 gateway listener exited: %v", err)
		}
	}()
	socks5GatewaySvc = svc
	socks5GatewayAddr = addr
	logs.Info("socks5 gateway listening on %s", addr)
	return nil
}

func init() {
	// Publish the gateway status callback so web/api/socks5.go can
	// expose it through /api/v1/socks5/gateway without importing
	// cmd/nps (which would be a cycle).
	server.Socks5GatewayInfoFn = func() (bool, string, int, int) {
		st := GetSocks5GatewayStatus()
		return st.Listening, st.Addr, st.Port, st.Routes
	}
}

// Socks5GatewayStatus returns a snapshot for the admin UI banner.
type Socks5GatewayStatus struct {
	Listening bool   `json:"listening"`
	Addr      string `json:"addr"`
	Port      int    `json:"port"`
	Routes    int    `json:"routes"`
}

// GetSocks5GatewayStatus is exported so web/api/socks5.go can report it
// without taking a hard dependency on the cmd/nps package private vars.
func GetSocks5GatewayStatus() Socks5GatewayStatus {
	socks5GatewayMu.Lock()
	svc := socks5GatewaySvc
	addr := socks5GatewayAddr
	socks5GatewayMu.Unlock()

	st := Socks5GatewayStatus{
		Port: beego.AppConfig.DefaultInt("socks5_shared_port", 0),
	}
	if svc != nil {
		st.Listening = true
		st.Addr = addr
		st.Routes = svc.RouteCount()
	}
	return st
}

// registerSocks5GatewayHooks wires settings → restart. Called from
// registerSettingsHotHooks() so it's part of the same boot phase as
// the HTTP hot-restart wiring.
func registerSocks5GatewayHooks() {
	for _, k := range []string{"socks5_shared_port", "socks5_shared_ip"} {
		key := k
		sqlitedb.OnSettingChange(key, func(_, _, _ string) {
			scheduleSocks5Restart(key)
		})
	}
}
