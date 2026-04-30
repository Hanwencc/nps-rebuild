package proxy

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"ehang.io/nps/lib/conn"
	"ehang.io/nps/lib/file"
	"github.com/astaxie/beego/logs"
)

// Sock5SharedServer is the Phase 9 multi-client SOCKS5 gateway. A
// single TCP listener accepts connections, demands SOCKS5
// username/password authentication, and routes each connection to the
// NPC client identified by the supplied username (matched against the
// `username` column of mode="socks5" Tunnel rows).
//
// Compared to the legacy per-port Sock5LocalServer this gateway
// trades one global listener for any number of routable accounts. The
// hot path remains: ① one map lookup over the in-memory Tasks
// sync.Map (linear scan, but typically <100 entries), ② per-route
// flow / quota check, ③ existing BaseServer.DealClient pipeline.
//
// Lifecycle is owned by cmd/nps/socks5_gateway.go which constructs a
// fresh instance on every settings change to socks5_shared_port or
// socks5_shared_ip.
type Sock5SharedServer struct {
	BaseServer
	listener net.Listener
	addr     string
}

// NewSock5SharedServer constructs an unbound gateway. The internal
// task is a synthetic placeholder (Id=-1) that owns flow accounting
// for unauthenticated handshakes; per-route flow goes onto the route's
// own Tunnel.Flow.
func NewSock5SharedServer(bridge NetBridge, addr string) *Sock5SharedServer {
	s := new(Sock5SharedServer)
	s.bridge = bridge
	// Synthetic owner task — never persisted, used only so BaseServer
	// helpers (CheckFlowAndConnNum etc.) have a non-nil pointer if
	// ever called via the embedded base. We pass a per-route task
	// into DealClient explicitly to avoid touching this one on the
	// hot path.
	s.task = &file.Tunnel{
		Id:    -1,
		Mode:  "socks5",
		Flow:  &file.Flow{},
		NoStore: true,
	}
	s.addr = addr
	return s
}

// Start binds the listener and begins accepting connections.
func (s *Sock5SharedServer) Start() error {
	return conn.NewTcpListenerAndProcess(s.addr, s.accept, &s.listener)
}

// Close stops the listener; safe to call multiple times.
func (s *Sock5SharedServer) Close() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

// Addr returns the bound address (for diagnostics / status endpoints).
func (s *Sock5SharedServer) Addr() string { return s.addr }

// RouteCount counts active SOCKS5 routing entries — used by the status
// endpoint so the UI can warn the operator if zero accounts exist.
func (s *Sock5SharedServer) RouteCount() int {
	var n int
	file.GetDb().JsonDb.Tasks.Range(func(_, v any) bool {
		t, _ := v.(*file.Tunnel)
		if t == nil || t.Mode != "socks5" {
			return true
		}
		if t.Status && t.Username != "" {
			n++
		}
		return true
	})
	return n
}

// accept is the per-connection entry point. It mirrors the structure
// of Sock5LocalServer.Start's inner closure but defers the conn-count
// bump until after auth so a misbehaving anonymous client cannot
// exhaust a routed client's NowConn quota by half-handshaking.
func (s *Sock5SharedServer) accept(c net.Conn) {
	sc := &socks5Conn{
		base:   &s.BaseServer,
		bridge: s.bridge,
	}
	if err := sc.negotiate(c, true); err != nil {
		logs.Trace("socks5 gateway negotiate failed from %s: %v", c.RemoteAddr(), err)
		c.Close()
		return
	}
	user, pass, err := socks5ReadUserPass(c)
	if err != nil {
		_ = socks5WriteAuthResult(c, false)
		logs.Trace("socks5 gateway auth read failed from %s: %v", c.RemoteAddr(), err)
		c.Close()
		return
	}
	tunnel := s.lookupRoute(user, pass)
	if tunnel == nil {
		_ = socks5WriteAuthResult(c, false)
		logs.Warn("socks5 gateway auth denied: user=%q remote=%s", user, c.RemoteAddr())
		c.Close()
		return
	}
	cli := tunnel.Client
	if cli == nil || !cli.Status {
		_ = socks5WriteAuthResult(c, false)
		logs.Warn("socks5 gateway: client for user %q is unavailable", user)
		c.Close()
		return
	}
	if err := s.checkRouteQuota(cli); err != nil {
		_ = socks5WriteAuthResult(c, false)
		logs.Warn("socks5 gateway: user %q over quota: %v", user, err)
		c.Close()
		return
	}
	if err := socks5WriteAuthResult(c, true); err != nil {
		cli.AddConn() // release the slot we just took
		c.Close()
		return
	}
	logs.Trace("socks5 gateway routed: user=%q client=%d remote=%s", user, cli.Id, c.RemoteAddr())

	// Per-conn dispatch state — pointer fields prevent the hot path
	// from racing on s.task (single shared instance).
	sc.client = cli
	sc.flow = tunnel.Flow
	sc.task = tunnel

	// Per-route base server so DealClient's bridge.SendLinkInfo carries
	// the correct task pointer (used by bridge for proto version).
	per := &BaseServer{bridge: s.bridge, task: tunnel}
	sc.base = per

	defer cli.AddConn() // release on connection close
	sc.handleRequest(c)
}

// lookupRoute walks the in-memory Tasks map for the first enabled
// socks5 entry whose username/password match. Linear scan is fine —
// real deployments rarely exceed a few dozen routes and the alternative
// (per-username index) duplicates state already kept by the JsonDb.
func (s *Sock5SharedServer) lookupRoute(user, pass string) *file.Tunnel {
	if user == "" {
		return nil
	}
	var match *file.Tunnel
	file.GetDb().JsonDb.Tasks.Range(func(_, v any) bool {
		t, _ := v.(*file.Tunnel)
		if t == nil || t.Mode != "socks5" || !t.Status {
			return true
		}
		if t.Username == user && t.Password == pass {
			match = t
			return false
		}
		return true
	})
	return match
}

// checkRouteQuota duplicates the parts of CheckFlowAndConnNum that
// matter for the gateway. Doing it inline avoids a second mutex
// acquisition on the hot path (BaseServer.CheckFlowAndConnNum walks
// through s.task which we don't want to touch here).
func (s *Sock5SharedServer) checkRouteQuota(client *file.Client) error {
	in := atomic.LoadInt64(&client.Flow.InletFlow)
	out := atomic.LoadInt64(&client.Flow.ExportFlow)
	if client.Flow.FlowLimit > 0 && (client.Flow.FlowLimit<<20) < (in+out) {
		return errors.New("traffic exceeded")
	}
	if !client.GetConn() {
		return errors.New("connections exceed the current client limit")
	}
	return nil
}

// SafeBindAddr renders host:port, accepting empty host as "0.0.0.0"
// for parity with the rest of nps.
func SafeBindAddr(ip string, port int) string {
	if ip == "" {
		ip = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%s", ip, strconv.Itoa(port))
}
