package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ehang.io/nps/lib/file"
	"ehang.io/nps/server"
	"ehang.io/nps/server/connection"
	"ehang.io/nps/server/proxy"
	"ehang.io/nps/server/tool"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// hothttp.go owns hot-restart of the implicit "httpHostServer" task
// that fronts the user's HTTP and HTTPS reverse-proxy domains.
//
// At boot, server.StartNewServer constructs a webServer mode service
// which in turn calls AddTask for an inline httpHostServer Tunnel
// (Id=0, Port=0). The resulting *httpServer ends up at RunList[0].
// We adopt that handle here once it appears.
//
// On change to any of the keys that NewMode reads at construction time
// (http_proxy_port, https_proxy_port, http_proxy_ip, https_just_proxy,
// https_default_cert_file, https_default_key_file, http_cache,
// http_cache_length, http_add_origin_header) we close the old service
// and build a fresh one with the new config.
//
// Bridge port collisions activate pMux at boot, in which case the
// HTTP/HTTPS listeners are sub-listeners of the bridge port and we
// CANNOT safely rebuild them — restart of the whole nps process is
// the only safe path. scheduleHttpRestart returns an error in that case.

var (
	httpHostMu  sync.Mutex
	httpHostSvc proxy.Service

	httpRestartTimerMu sync.Mutex
	httpRestartTimer   *time.Timer
)

// adoptInitialHttpHostServer waits for the bootstrap goroutines in
// server.StartNewServer to publish the implicit httpHostServer task at
// RunList[0]. Called from cmd/nps/nps.go after StartNewServer.
func adoptInitialHttpHostServer() {
	go func() {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if v, ok := server.RunList.Load(0); ok && v != nil {
				if svc, ok := v.(proxy.Service); ok {
					httpHostMu.Lock()
					httpHostSvc = svc
					httpHostMu.Unlock()
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		logs.Warn("hothttp: initial httpHostServer not adopted within 5s; HTTP/HTTPS hot-restart disabled")
	}()
}

// scheduleHttpRestart debounces repeated calls so a multi-key UI save
// triggers exactly one restart. The debounce window is short enough
// that the user perceives an immediate effect.
func scheduleHttpRestart(reason string) {
	httpRestartTimerMu.Lock()
	defer httpRestartTimerMu.Unlock()
	if httpRestartTimer != nil {
		httpRestartTimer.Stop()
	}
	httpRestartTimer = time.AfterFunc(200*time.Millisecond, func() {
		if err := restartHttpHostServer(); err != nil {
			logs.Error("httpHostServer hot-restart failed (%s): %v", reason, err)
			return
		}
		logs.Info("httpHostServer hot-restarted (%s)", reason)
	})
}

func restartHttpHostServer() error {
	httpHostMu.Lock()
	defer httpHostMu.Unlock()

	if connection.PMuxActive() {
		return errors.New("port multiplexing is active (http/https port shares bridge port); restart nps to apply changes")
	}
	if httpHostSvc == nil {
		return errors.New("no httpHostServer instance has been adopted yet")
	}

	newHttpPort := beego.AppConfig.DefaultInt("http_proxy_port", 0)
	newHttpsPort := beego.AppConfig.DefaultInt("https_proxy_port", 0)
	useCache, _ := beego.AppConfig.Bool("http_cache")
	cacheLen, _ := beego.AppConfig.Int("http_cache_length")
	addOrigin, _ := beego.AppConfig.Bool("http_add_origin_header")

	// Tear the old listener down first so its socket is freed before we
	// probe. The proxy.NewHttp Start path os.Exit's on bind failure, so
	// we have to be sure the port is free before we let it Start.
	old := httpHostSvc
	httpHostSvc = nil
	if err := old.Close(); err != nil {
		logs.Warn("hothttp: old httpHostServer Close returned: %v", err)
	}
	// Brief pause to let the OS release the TCP listening socket.
	time.Sleep(150 * time.Millisecond)

	connection.RefreshHttpPorts()

	if newHttpPort > 0 && !tool.TestServerPort(newHttpPort, "tcp") {
		// Best effort: roll forward. We cannot bring the old listener
		// back because Close already destroyed its underlying handle.
		return fmt.Errorf("http_proxy_port %d is not bindable; HTTP proxy is now stopped", newHttpPort)
	}
	if newHttpsPort > 0 && !tool.TestServerPort(newHttpsPort, "tcp") {
		return fmt.Errorf("https_proxy_port %d is not bindable; HTTPS proxy is now stopped", newHttpsPort)
	}

	task := &file.Tunnel{Port: 0, Mode: "httpHostServer", Status: true}
	svc := proxy.NewHttp(server.Bridge, task, newHttpPort, newHttpsPort, useCache, cacheLen, addOrigin)

	// proxy.NewHttp's Start spawns goroutines and returns nil; the
	// goroutines call os.Exit on bind failure. We pre-probed above so
	// this is best-effort safe.
	if err := svc.Start(); err != nil {
		return err
	}
	httpHostSvc = svc
	server.RunList.Store(0, svc)
	return nil
}
