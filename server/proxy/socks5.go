package proxy

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/conn"
	"ehang.io/nps/lib/file"
	"github.com/astaxie/beego/logs"
)

const (
	ipV4            = 1
	domainName      = 3
	ipV6            = 4
	connectMethod   = 1
	bindMethod      = 2
	associateMethod = 3
	// The maximum packet size of any udp Associate packet, based on ethernet's max size,
	// minus the IP and UDP headers. IPv4 has a 20 byte header, UDP adds an
	// additional 4 bytes.  This is a total overhead of 24 bytes.  Ethernet's
	// max packet size is 1500 bytes,  1500 - 24 = 1476.
	maxUDPPacketSize = 1476
)

const (
	succeeded uint8 = iota
	serverFailure
	notAllowed
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

const (
	UserPassAuth    = uint8(2)
	userAuthVersion = uint8(1)
	authSuccess     = uint8(0)
	authFailure     = uint8(1)
)

// socks5Conn carries the per-connection routing context shared by both
// the legacy single-tunnel listener (Sock5LocalServer, used by NPC's
// p2p socks5 visitor) and the multi-client gateway (Sock5SharedServer
// in socks5_shared.go).
//
// Phase 9 split socks5.go so the protocol parsing has no dependency on
// a single owning *Sock5ModeServer; client / flow / task are passed in
// per-connection. This keeps the shared gateway lock-free even when
// thousands of accounts are routed through one listener.
type socks5Conn struct {
	base   *BaseServer  // owner; provides BlackIP / DealClient plumbing
	bridge NetBridge    // pulled out of base for tests / NPC-only mode
	client *file.Client // resolved tunnel client (post-auth for shared)
	flow   *file.Flow   // per-route flow accumulator
	task   *file.Tunnel // per-route tunnel (used for SendLinkInfo)
}

// req
func (s *socks5Conn) handleRequest(c net.Conn) {
	header := make([]byte, 3)
	if _, err := io.ReadFull(c, header); err != nil {
		logs.Warn("illegal request", err)
		c.Close()
		return
	}
	switch header[1] {
	case connectMethod:
		s.handleConnect(c)
	case bindMethod:
		s.handleBind(c)
	case associateMethod:
		s.handleUDP(c)
	default:
		s.sendReply(c, commandNotSupported)
		c.Close()
	}
}

// reply
func (s *socks5Conn) sendReply(c net.Conn, rep uint8) {
	reply := []byte{5, rep, 0, 1}
	localAddr := c.LocalAddr().String()
	localHost, localPort, _ := net.SplitHostPort(localAddr)
	ipBytes := net.ParseIP(localHost).To4()
	nPort, _ := strconv.Atoi(localPort)
	reply = append(reply, ipBytes...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(nPort))
	reply = append(reply, portBytes...)
	c.Write(reply)
}

// do conn
func (s *socks5Conn) doConnect(c net.Conn, command uint8) {
	addrType := make([]byte, 1)
	c.Read(addrType)
	var host string
	switch addrType[0] {
	case ipV4:
		ipv4 := make(net.IP, net.IPv4len)
		c.Read(ipv4)
		host = ipv4.String()
	case ipV6:
		ipv6 := make(net.IP, net.IPv6len)
		c.Read(ipv6)
		host = ipv6.String()
	case domainName:
		var domainLen uint8
		binary.Read(c, binary.BigEndian, &domainLen)
		domain := make([]byte, domainLen)
		c.Read(domain)
		host = string(domain)
	default:
		s.sendReply(c, addrTypeNotSupported)
		return
	}

	var port uint16
	binary.Read(c, binary.BigEndian, &port)
	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
	var ltype string
	if command == associateMethod {
		ltype = common.CONN_UDP
	} else {
		ltype = common.CONN_TCP
	}
	localProxy := false
	if s.task != nil && s.task.Target != nil {
		localProxy = s.task.Target.LocalProxy
	}
	s.base.DealClient(conn.NewConn(c), s.client, addr, nil, ltype, func() {
		s.sendReply(c, succeeded)
	}, s.flow, localProxy, s.task)
}

// conn
func (s *socks5Conn) handleConnect(c net.Conn) {
	s.doConnect(c, connectMethod)
}

// passive mode
func (s *socks5Conn) handleBind(c net.Conn) {}

func (s *socks5Conn) sendUdpReply(writeConn net.Conn, c net.Conn, rep uint8, serverIp string) {
	reply := []byte{5, rep, 0, 1}
	_, localPort, _ := net.SplitHostPort(c.LocalAddr().String())
	ipBytes := net.ParseIP(serverIp).To4()
	nPort, _ := strconv.Atoi(localPort)
	reply = append(reply, ipBytes...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(nPort))
	reply = append(reply, portBytes...)
	writeConn.Write(reply)
}

func (s *socks5Conn) handleUDP(c net.Conn) {
	defer c.Close()
	addrType := make([]byte, 1)
	c.Read(addrType)
	var host string
	switch addrType[0] {
	case ipV4:
		ipv4 := make(net.IP, net.IPv4len)
		c.Read(ipv4)
		host = ipv4.String()
	case ipV6:
		ipv6 := make(net.IP, net.IPv6len)
		c.Read(ipv6)
		host = ipv6.String()
	case domainName:
		var domainLen uint8
		binary.Read(c, binary.BigEndian, &domainLen)
		domain := make([]byte, domainLen)
		c.Read(domain)
		host = string(domain)
	default:
		s.sendReply(c, addrTypeNotSupported)
		return
	}
	var port uint16
	binary.Read(c, binary.BigEndian, &port)
	logs.Warn(host, string(port))
	serverIp := ""
	if s.task != nil {
		serverIp = s.task.ServerIp
	}
	replyAddr, err := net.ResolveUDPAddr("udp", serverIp+":0")
	if err != nil {
		logs.Error("build local reply addr error", err)
		return
	}
	reply, err := net.ListenUDP("udp", replyAddr)
	if err != nil {
		s.sendReply(c, addrTypeNotSupported)
		logs.Error("listen local reply udp port error")
		return
	}
	s.sendUdpReply(c, reply, succeeded, common.GetServerIpByClientIp(c.RemoteAddr().(*net.TCPAddr).IP))
	defer reply.Close()
	link := conn.NewLink("udp5", "", s.client.Cnf.Crypt, s.client.Cnf.Compress, c.RemoteAddr().String(), false, "")
	target, err := s.bridge.SendLinkInfo(s.client.Id, link, s.task)
	if err != nil {
		logs.Warn("get connection from client id %d  error %s", s.client.Id, err.Error())
		return
	}

	var clientAddr net.Addr
	go func() {
		b := common.BufPoolUdp.Get().([]byte)
		defer common.BufPoolUdp.Put(b)
		defer c.Close()
		for {
			n, laddr, err := reply.ReadFrom(b)
			if err != nil {
				logs.Error("read data from %s err %s", reply.LocalAddr().String(), err.Error())
				return
			}
			if clientAddr == nil {
				clientAddr = laddr
			}
			if _, err := target.Write(b[:n]); err != nil {
				logs.Error("write data to client error", err.Error())
				return
			}
		}
	}()

	go func() {
		var l int32
		b := common.BufPoolUdp.Get().([]byte)
		defer common.BufPoolUdp.Put(b)
		defer c.Close()
		for {
			if err := binary.Read(target, binary.LittleEndian, &l); err != nil || l >= common.PoolSizeUdp || l <= 0 {
				logs.Warn("read len bytes error", err.Error())
				return
			}
			binary.Read(target, binary.LittleEndian, b[:l])
			if err != nil {
				logs.Warn("read data form client error", err.Error())
				return
			}
			if _, err := reply.WriteTo(b[:l], clientAddr); err != nil {
				logs.Warn("write data to user ", err.Error())
				return
			}
		}
	}()

	b := common.BufPoolUdp.Get().([]byte)
	defer common.BufPoolUdp.Put(b)
	defer target.Close()
	for {
		_, err := c.Read(b)
		if err != nil {
			c.Close()
			return
		}
	}
}

// negotiate reads the SOCKS5 method-negotiation greeting and writes
// the chosen auth method back. needAuth==true forces UserPassAuth even
// if the client offered "no auth"; used by the shared gateway where
// every connection MUST authenticate to be routable.
func (s *socks5Conn) negotiate(c net.Conn, needAuth bool) error {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(c, buf); err != nil {
		return err
	}
	if version := buf[0]; version != 5 {
		return errors.New("only socks5 supported")
	}
	nMethods := buf[1]
	methods := make([]byte, nMethods)
	if n, err := c.Read(methods); n != int(nMethods) || err != nil {
		return errors.New("wrong method")
	}
	if needAuth {
		buf[1] = UserPassAuth
		_, err := c.Write(buf)
		return err
	}
	buf[1] = 0
	_, err := c.Write(buf)
	return err
}

// readUserPass reads a RFC 1929 user/pass auth packet WITHOUT validating
// the credentials. The shared gateway uses this to extract the routing
// key (username) before deciding which Tunnel handles the connection;
// the local NPC server uses (user, pass) for the legacy single-account
// or MultiAccount comparison.
func socks5ReadUserPass(c net.Conn) (user, pass string, err error) {
	header := []byte{0, 0}
	if _, err = io.ReadAtLeast(c, header, 2); err != nil {
		return
	}
	if header[0] != userAuthVersion {
		err = errors.New("auth method not supported")
		return
	}
	userLen := int(header[1])
	userBuf := make([]byte, userLen)
	if _, err = io.ReadAtLeast(c, userBuf, userLen); err != nil {
		return
	}
	if _, err = c.Read(header[:1]); err != nil {
		err = errors.New("password length read error")
		return
	}
	passLen := int(header[0])
	passBuf := make([]byte, passLen)
	if _, err = io.ReadAtLeast(c, passBuf, passLen); err != nil {
		return
	}
	user = string(userBuf)
	pass = string(passBuf)
	return
}

func socks5WriteAuthResult(c net.Conn, ok bool) error {
	res := []byte{userAuthVersion, authSuccess}
	if !ok {
		res[1] = authFailure
	}
	_, err := c.Write(res)
	return err
}

// ----- legacy single-tunnel listener (NPC p2p visitor only) ---------------
//
// Sock5LocalServer is the per-port listener used by NPC's local "p2ps"
// mode (client/local.go). It binds task.ServerIp:task.Port, accepts
// SOCKS5 connections, and forwards everything to the single Tunnel
// it was constructed with. The server-side multi-client gateway lives
// in socks5_shared.go and does NOT use this type.

type Sock5LocalServer struct {
	BaseServer
	listener net.Listener
}

func (s *Sock5LocalServer) handleConn(c net.Conn) {
	sc := &socks5Conn{
		base:   &s.BaseServer,
		bridge: s.bridge,
		client: s.task.Client,
		flow:   s.task.Flow,
		task:   s.task,
	}
	useAuth := (s.task.Client.Cnf.U != "" && s.task.Client.Cnf.P != "") ||
		(s.task.MultiAccount != nil && len(s.task.MultiAccount.AccountMap) > 0)
	if err := sc.negotiate(c, useAuth); err != nil {
		logs.Warn("negotiation err", err)
		c.Close()
		return
	}
	if useAuth {
		user, pass, err := socks5ReadUserPass(c)
		if err != nil {
			_ = socks5WriteAuthResult(c, false)
			logs.Warn("Validation failed:", err)
			c.Close()
			return
		}
		var U, P string
		if s.task.MultiAccount != nil {
			if user == "" {
				_ = socks5WriteAuthResult(c, false)
				c.Close()
				return
			}
			var ok bool
			P, ok = s.task.MultiAccount.AccountMap[user]
			if !ok {
				_ = socks5WriteAuthResult(c, false)
				c.Close()
				return
			}
			U = user
		} else {
			U = s.task.Client.Cnf.U
			P = s.task.Client.Cnf.P
		}
		if user != U || pass != P {
			_ = socks5WriteAuthResult(c, false)
			c.Close()
			return
		}
		if err := socks5WriteAuthResult(c, true); err != nil {
			c.Close()
			return
		}
	}
	sc.handleRequest(c)
}

// start
func (s *Sock5LocalServer) Start() error {
	return conn.NewTcpListenerAndProcess(s.task.ServerIp+":"+strconv.Itoa(s.task.Port), func(c net.Conn) {
		if err := s.CheckFlowAndConnNum(s.task.Client); err != nil {
			logs.Warn("client id %d, task id %d, error %s, when socks5 connection", s.task.Client.Id, s.task.Id, err.Error())
			c.Close()
			return
		}
		logs.Trace("New socks5 connection,client %d,remote address %s", s.task.Client.Id, c.RemoteAddr())
		s.handleConn(c)
		s.task.Client.AddConn()
	}, &s.listener)
}

// new
func NewSock5LocalServer(bridge NetBridge, task *file.Tunnel) *Sock5LocalServer {
	s := new(Sock5LocalServer)
	s.bridge = bridge
	s.task = task
	return s
}

// close
func (s *Sock5LocalServer) Close() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}
