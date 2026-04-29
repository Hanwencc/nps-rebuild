# NPS 内网穿透项目 — 代码库速查文档

## 项目概览
- **项目名**: nps (ehang.io/nps)
- **版本**: 0.26.32（`lib/version/version.go`）
- **语言**: Go 1.24，Web框架: Beego v1.12
- **模块路径**: `ehang.io/nps`
- **功能**: 轻量级高性能内网穿透代理服务器，支持 TCP/UDP/HTTP(S)/Socks5/P2P/文件服务

---

## 目录结构总览

```
bridge/         服务端网桥，处理客户端连接与信令
client/         NPC 客户端核心逻辑
cmd/nps/        服务端入口 (nps.go)
cmd/npc/        客户端入口 (npc.go, sdk.go)
cmd/npc/npc-gui/ GUI 客户端 (Wails + Fyne)
conf/           配置文件 (nps.conf, npc.conf, tasks/clients/hosts/global.json)
lib/
  cache/        LRU 缓存
  common/       常量、工具函数、对象池、日志
  config/       npc 配置文件解析
  conn/         连接封装、KCP/TCP 监听器、snappy压缩
  crypt/        TLS/AES/MD5 加密
  daemon/       守护进程、热重载
  file/         数据持久化 (JSON文件数据库，obj/db/file)
  goroutine/    协程池 (ants)
  install/      安装/卸载系统服务
  nps_mux/      多路复用器 (单TCP连接承载多逻辑连接)
  pmux/         端口复用
  rate/         带宽/流量限速
  version/      版本常量
server/
  connection/   监听器工厂 (HTTP/HTTPS/Bridge/Web)
  proxy/        各协议代理服务器实现
  tool/         工具函数 (黑名单等)
  server.go     服务端主逻辑 (任务启停、Dashboard)
web/
  controllers/  Beego 控制器 (client/index/auth/login/global)
  routers/      路由注册
  static/       前端静态资源
  views/        Beego 模板
gui/npc/        Android 客户端
```

---

## 核心数据结构（`lib/file/obj.go`）

### Client（NPC客户端）
```go
type Client struct {
  Id              int
  VerifyKey       string     // 连接密钥
  Remark          string
  Status          bool       // 是否允许连接
  IsConnect       bool
  RateLimit       int        // KB/s
  Flow            *Flow      // 流量统计/限制
  Rate            *rate.Rate
  MaxConn         int        // 最大并发连接数
  NowConn         int32
  WebUserName     string     // 子用户web登录
  WebPassword     string
  ConfigConnAllow bool       // 允许配置文件连接
  MaxTunnelNum    int
  BlackIpList     []string
  IpWhite         bool
  IpWhitePass     string
  IpWhiteList     []string
  Cnf             *Config    // 压缩/加密配置
}
```

### Tunnel（隧道任务）
- 存储于 `conf/tasks.json`，字段含 `Mode`、`Port`、`ServerIp`、`Target`、`Client`、`Status`、`Flow` 等

### Host（域名解析）
- 存储于 `conf/hosts.json`，域名→目标地址映射，支持自定义证书

### Glob（全局配置）
- 存储于 `conf/global.json`，含 `BlackIpList` 等全局策略

---

## 持久化（`lib/file/`）

- **数据库**: `JsonDb`（sync.Map + JSON文件），单例 `file.GetDb()`
- **文件**: `tasks.json` / `clients.json` / `hosts.json` / `global.json`
- **加载**: 启动时 `LoadXxxFromJsonFile()`，增量写入

---

## 网桥（`bridge/bridge.go`）

- `Bridge` struct：服务端监听 `bridge_port`(默认8024)，接受 NPC 连接
- 支持 TCP / KCP / TLS 三种隧道类型
- 内部通道：`OpenTask` / `CloseTask` / `CloseClient` / `SecretChan`
- 客户端连接后建立 `nps_mux.Mux` 多路复用隧道

---

## 多路复用（`lib/nps_mux/mux.go`）

- 单物理连接上多路复用逻辑连接
- Ping/Pong 心跳检测（KCP默认20次，TCP默认60次超时断连）
- 带宽/延迟统计

---

## 代理服务器（`server/proxy/`）

| 文件 | 类型 | 说明 |
|------|------|------|
| `base.go` | `BaseServer` | 公共基类，流量统计、黑名单检查、`DealClient` |
| `tcp.go` | `TunnelModeServer` | TCP隧道、Web管理服务器 |
| `udp.go` | `UdpModeServer` | UDP转发 |
| `http.go` | `httpServer` | HTTP域名代理，支持LRU缓存 |
| `https.go` | `HttpsServer` | HTTPS，SNI证书路由 |
| `socks5.go` | `Sock5ModeServer` | Socks5代理（支持用户名密码认证） |
| `p2p.go` | `P2PServer` | P2P UDP打洞 |
| `websocket.go` | WebSocket代理 |  |

所有代理实现 `Service` 接口：`Start() error` / `Close() error`

---

## 服务端主逻辑（`server/server.go`）

- `InitFromCsv()` — 从 JSON 文件加载任务并启动
- `DealBridgeTask()` — 监听 Bridge 通道，动态增删任务
- `AddTask(t *Tunnel)` — 根据 Tunnel.Mode 启动对应代理服务器
- `StopServer(id)` / `StartTask(id)` / `DelTunnelAndHostByClientId()`
- `GetDashboardData()` — 系统信息（CPU/内存/流量/在线客户端数）
- `RunList sync.Map` — 运行中的服务，key=Tunnel.Id

---

## 客户端（`client/`）

| 文件 | 说明 |
|------|------|
| `client.go` | `TRPClient` 结构，连接/重连逻辑，支持独立 logger |
| `control.go` | 控制信道处理，任务下发，配置文件连接模式 |
| `local.go` | 本地端口监听，转发到 mux 隧道 |
| `health.go` | 健康检查（HTTP/TCP），后端高可用 |
| `register.go` | 注册模式 |

`TRPClient` 支持 `SetLogger()` 自定义日志（用于 SDK/GUI 多实例）

---

## Web 管理（`web/`）

- **框架**: Beego，端口默认 8081（conf 中为 8080）
- **路由**: `web/routers/router.go` — AutoRouter 模式
- **控制器**:
  - `LoginController` — 登录/注销，支持子用户
  - `ClientController` — 客户端 CRUD，含黑白名单/流量/限速
  - `IndexController` — Dashboard + 各协议隧道列表
  - `AuthController` — API 鉴权、IP 白名单授权
  - `GlobalController` — 全局配置修改

---

## 通信协议常量（`lib/common/const.go`）

```
WORK_MAIN/CHAN/CONFIG/REGISTER/SECRET/FILE
WORK_P2P_VISITOR/PROVIDER/CONNECT/SUCCESS/END
CONN_TCP / CONN_UDP
NEW_TASK / NEW_CONF / NEW_HOST
RES_MSG / RES_CLOSE
```

---

## 配置文件速查

### nps.conf（服务端）
```ini
bridge_port=8024          # NPC连接端口
bridge_type=tcp           # tcp|kcp
http_proxy_port=80
https_proxy_port=443
web_port=8081
web_username=admin
web_password=123
public_vkey=123           # 公共密钥（无需创建客户端）
auth_key=123              # API认证密钥
p2p_port=6000
flow_store_interval=1     # 流量持久化间隔(分钟)
allow_flow_limit=true
allow_rate_limit=true
allow_tunnel_num_limit=true
allow_connection_num_limit=true
```

### npc.conf（客户端）
```ini
[common]
server_addr=127.0.0.1:8024
conn_type=tcp             # tcp|kcp
vkey=123
crypt=true
compress=true
tls_enable=true
disconnect_timeout=60

[tcp]
mode=tcp
target_addr=127.0.0.1:8080
server_port=10000

[socks5]
mode=socks5
server_port=19009

[http]
mode=httpProxy
server_port=19004

[udp]
mode=udp
server_port=12253

[web]
host=c.o.com
target_addr=127.0.0.1:8083
```

---

## 安全机制

- **黑名单**: 全局 `global.BlackIpList`，每个 Client 可独立配置 `BlackIpList`
- **IP白名单**: Client 级别 `IpWhite/IpWhiteList/IpWhitePass`，`/auth/ipwhiteauth` 接口授权
- **流量限制**: `Flow.FlowLimit`（MB），超出断连
- **连接数限制**: `Client.MaxConn`
- **隧道数限制**: `Client.MaxTunnelNum`
- **带宽限速**: `Client.RateLimit`（KB/s），令牌桶 `rate.Rate`
- **认证**: Basic Auth、API `auth_key`（AES-128加密），Web session

---

## 关键调用链

```
NPC 启动
  → TRPClient.Start()
  → control.go: 建立 signal 信道 + nps_mux.Mux 隧道
  → 接收服务端 NEW_TASK 指令 → local.go 本地监听

NPS 启动
  → Bridge.StartTunnel() 监听 bridge_port
  → cliProcess() 验证 vkey → 注册 Client
  → DealBridgeTask() goroutine 处理任务变更
  → InitFromCsv() 加载历史任务

HTTP请求处理
  → httpServer.handleHost()
  → file.GetDb().GetInfoByHost(host) 匹配域名
  → BaseServer.DealClient() → bridge.SendLinkInfo()
  → nps_mux 新建逻辑连接 → NPC 本地转发
```

---

## 常见修改位置速查

| 需求 | 文件 |
|------|------|
| 新增代理协议 | `server/proxy/` 新建文件，实现 Service 接口；`server/server.go` AddTask() 添加 case |
| 修改认证逻辑 | `bridge/bridge.go` cliProcess()，`web/controllers/auth.go` |
| 修改黑名单逻辑 | `server/proxy/base.go` IsGlobalBlackIp()，`lib/common/util.go` IsBlackIp() |
| 修改流量统计 | `lib/file/obj.go` Flow struct，`server/proxy/base.go` FlowAdd() |
| 修改 Web API | `web/controllers/`，`web/routers/router.go` |
| 修改健康检查 | `client/health.go` |
| 修改多路复用 | `lib/nps_mux/mux.go` |
| 修改客户端数据结构 | `lib/file/obj.go` Client struct + `lib/file/db.go` 相关CRUD |
| 修改持久化 | `lib/file/file.go` JsonDb |
| 修改版本号 | `lib/version/version.go` |
