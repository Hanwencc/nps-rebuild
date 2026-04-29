# NPS API Token 接口手册

> 适用版本：迁移到 Vue 3 + RESTful 之后的 NPS。所有 `/api/v1/*` 接口均使用统一 JSON 信封：
>
> ```json
> { "code": 0, "message": "ok", "data": <any> }
> ```
>
> `code == 0` 表示成功；非零时 `message` 即错误原因，HTTP 状态码也会同步。

---

## 1. 认证方式

`/api/v1/*` 下除登录探活外的全部接口，都同时支持以下两种凭据；任选其一即可：

| 方式 | 用途 | 携带方式 |
|------|------|----------|
| **API Token**（推荐，机器对机器） | 长期凭证，可作用域限定 / 可吊销 / 可轮换 | 见下文 |
| Beego Session Cookie | 浏览器 SPA 登录后默认使用 | 浏览器自动管理 |

> **md5(auth_key+timestamp) 已彻底移除**，旧脚本必须迁移到 API Token。

### 1.1 携带 API Token

任选其一：

**方式 A — 双自定义头（推荐）**

```
X-Api-Key:    <KeyId>
X-Api-Secret: <Secret>
```

**方式 B — Authorization 头（以下 4 种写法均可）**

```
Authorization: Bearer <KeyId>.<Secret>
Authorization: bearer <KeyId>.<Secret>
Authorization: <KeyId>.<Secret>          # 不带 scheme 前缀
Authorization: <KeyId> <Secret>          # 空格分隔
```

> 即在 Postman 等工具里，如果 Authorization 直接填 `k_xxx.yyyy` 也能通过，无需手动加 `Bearer ` 前缀。

`<KeyId>` 形如 `k_0fd9f1d0d5c7dc78`，`<Secret>` 是 64 位十六进制串，**仅在创建 / 轮换接口的响应里返回一次**，服务端只保存 bcrypt 哈希。

### 1.2 失败响应

| HTTP | code | 触发场景 |
|------|------|----------|
| 401 | 4010 | 未携带凭据、Secret 错误、Token 已禁用 / 过期、来源 IP 不在白名单、请求路径或方法不在作用域内 |
| 403 | 4030 | 已认证但当前账号 / Token 无管理员权限（目前 Token 默认 `isAdmin=true`，主要影响子用户 session） |

---

## 2. Token 作用域字段

每个 Token 创建时可设置以下作用域，任意一项不满足都会被拒绝：

| 字段 | 含义 | 留空表示 |
|------|------|----------|
| `allowedPathPrefix` | 仅允许访问以该前缀开头的路径，例如 `/api/v1/clients` | 不限 |
| `allowedMethods` | 仅允许这些 HTTP 方法（不区分大小写），例如 `["GET","POST"]` | 不限 |
| `allowIps` | 来源 IP 白名单，支持精确 IP 或 CIDR，例如 `["10.0.0.0/8","192.168.1.5"]` | 不限 |
| `expiresAt` | Unix 秒级过期时间 | `0` = 永不过期 |
| `disabled` | 是否禁用 | `false` |

---

## 3. 支持 X-Api-Key 的接口清单

下列所有接口都同时接受 **API Token** 与 **Session Cookie**。除非另注，统一返回 JSON 信封。

### 3.1 仪表盘

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/dashboard/summary` | 服务器/客户端/隧道/域名/流量/系统负载汇总 |

### 3.2 客户端 (Clients)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/clients`                | 列表（query: `start`,`length`,`search`,`sort`,`order`） |
| POST   | `/api/v1/clients`                | 新建客户端 |
| GET    | `/api/v1/clients/:id`            | 详情 |
| PUT    | `/api/v1/clients/:id`            | 更新 |
| DELETE | `/api/v1/clients/:id`            | 删除 |
| POST   | `/api/v1/clients/:id/status`     | 启用 / 停用 |
| GET    | `/api/v1/clients/:id/quickinfo`  | 客户端连接/启动信息 |

### 3.3 隧道 (Tunnels)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/tunnels`            | 列表（query: `mode`,`clientId`,`start`,`length`,`search`） |
| POST   | `/api/v1/tunnels`            | 新建隧道 |
| GET    | `/api/v1/tunnels/:id`        | 详情 |
| PUT    | `/api/v1/tunnels/:id`        | 更新 |
| DELETE | `/api/v1/tunnels/:id`        | 删除 |
| POST   | `/api/v1/tunnels/:id/start`  | 启动 |
| POST   | `/api/v1/tunnels/:id/stop`   | 停止 |
| POST   | `/api/v1/tunnels/:id/copy`   | 复制一份隧道 |

### 3.4 域名解析 (Hosts)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/hosts`             | 列表 |
| POST   | `/api/v1/hosts`             | 新建 |
| GET    | `/api/v1/hosts/:id`         | 详情 |
| PUT    | `/api/v1/hosts/:id`         | 更新 |
| DELETE | `/api/v1/hosts/:id`         | 删除 |
| POST   | `/api/v1/hosts/:id/status`  | 启用 / 停用 |

### 3.5 全局参数

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/global` | 读取 IP 黑名单 / 服务器外部 URL 等 |
| PUT | `/api/v1/global` | 写入 |

### 3.6 API Token 自管理

> 用 Token A 调用以下接口可以管理 Token B/C/...；当然 Token 自身也能被另一个 Token 删除，请妥善保管。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET    | `/api/v1/tokens`              | 列出所有 Token（不含明文 secret） |
| POST   | `/api/v1/tokens`              | 创建 Token，**响应中包含一次性 secret** |
| GET    | `/api/v1/tokens/:id`          | 详情 |
| PUT    | `/api/v1/tokens/:id`          | 更新作用域 / 启用状态 / 备注 / 过期 |
| DELETE | `/api/v1/tokens/:id`          | 删除 |
| POST   | `/api/v1/tokens/:id/rotate`   | 重新签发 secret，旧 secret 立即失效，**响应中包含新的一次性 secret** |

### 3.7 不需要鉴权的接口（无需带 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/api/v1/auth/health` | 健康检查 |
| POST | `/api/v1/auth/login`  | 用户名 / 密码登录（写 session） |
| GET  | `/api/v1/auth/me`     | 当前 session 用户信息（无 session 时返回 401，前端探活专用） |
| *    | `/api/v1/auth/logout` | 注销 session |
| *    | `/auth/ipwhiteauth`   | NPC 客户端 IP 白名单自助登记（vkey + pass） |

---

## 4. Token 管理接口的请求 / 响应示例

### 4.1 创建 Token

`POST /api/v1/tokens`

```json
{
  "remark": "ci-deploy",
  "allowedPathPrefix": "/api/v1/clients",
  "allowedMethods": ["GET", "POST"],
  "allowIps": ["10.0.0.0/8"],
  "expiresAt": 1830000000,
  "disabled": false
}
```

响应（**`secret` 只出现这一次**）：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "secret": "3ad4fc2281591355eed6e802545f6c2fa395da3a1813f14e1e1c040c41ec297f",
    "token": {
      "id": 1,
      "keyId": "k_0fd9f1d0d5c7dc78",
      "remark": "ci-deploy",
      "allowedPathPrefix": "/api/v1/clients",
      "allowedMethods": ["GET", "POST"],
      "allowIps": ["10.0.0.0/8"],
      "expiresAt": 1830000000,
      "createdAt": 1777451738,
      "lastUsedAt": 0,
      "lastUsedIp": "",
      "disabled": false
    }
  }
}
```

### 4.2 列出 Token

`GET /api/v1/tokens` → `data` 是 `token` 对象数组（不含 secret）。

### 4.3 更新 Token

`PUT /api/v1/tokens/:id`，请求体字段全部可选，未提供保持原值：

```json
{
  "remark": "ci-deploy v2",
  "allowedMethods": ["GET"],
  "disabled": true
}
```

### 4.4 轮换 Token

`POST /api/v1/tokens/:id/rotate`，请求体可为空对象 `{}`，响应同 4.1（含新的 `secret`）。

### 4.5 删除 Token

`DELETE /api/v1/tokens/:id` → `{ "code": 0, "message": "deleted" }`

---

## 5. 调用示例

### 5.1 curl

```bash
# 列出客户端
curl -H "X-Api-Key: k_0fd9f1d0d5c7dc78" \
     -H "X-Api-Secret: 3ad4fc...297f" \
     'http://nps.example.com:8081/api/v1/clients?start=0&length=20'

# 用 Bearer 创建一个隧道
curl -X POST \
  -H "Authorization: Bearer k_0fd9f1d0d5c7dc78.3ad4fc...297f" \
  -H "Content-Type: application/json" \
  -d '{"clientId":1,"mode":"tcp","port":12345,"target":"127.0.0.1:80"}' \
  http://nps.example.com:8081/api/v1/tunnels
```

### 5.2 PowerShell

```powershell
$h = @{ 'X-Api-Key' = 'k_0fd9f1d0d5c7dc78'; 'X-Api-Secret' = '3ad4fc...297f' }
Invoke-RestMethod 'http://nps.example.com:8081/api/v1/dashboard/summary' -Headers $h
```

### 5.3 Node.js (axios)

```ts
import axios from 'axios'
const http = axios.create({
  baseURL: 'http://nps.example.com:8081/api/v1',
  headers: {
    'X-Api-Key': process.env.NPS_KEY_ID!,
    'X-Api-Secret': process.env.NPS_SECRET!,
  },
})
const { data } = await http.get('/clients', { params: { start: 0, length: 20 } })
```

### 5.4 Go

```go
req, _ := http.NewRequest("GET", "http://nps.example.com:8081/api/v1/clients?start=0&length=20", nil)
req.Header.Set("X-Api-Key", os.Getenv("NPS_KEY_ID"))
req.Header.Set("X-Api-Secret", os.Getenv("NPS_SECRET"))
resp, err := http.DefaultClient.Do(req)
```

---

## 6. 安全建议

1. **最小权限**：每个集成方一把独立的 Token，按需收紧 `allowedPathPrefix` / `allowedMethods` / `allowIps`。
2. **设置过期**：CI / 临时脚本一律设置 `expiresAt`。
3. **轮换**：怀疑泄露立即 `rotate`；旧 secret 在下一次 token 验证时即失效。
4. **存储**：Secret 只存在你的密钥管理系统里（KMS / Vault / GitHub Actions Secrets），切勿入库。
5. **审计**：列表中的 `lastUsedAt` / `lastUsedIp` 字段每次调用 1 秒去抖刷新，可用于异常检测。
6. **传输**：生产环境务必通过 HTTPS（前置 nginx / Caddy）暴露 `/api/v1/*`。

---

## 7. 持久化与文件位置

- Token 元数据保存在 `conf/api_tokens.json`（与 `clients.json` 同目录）。
- 文件格式：每条 token JSON + `\n` + `CONN_DATA_SEQ` 分隔，符合现有 `JsonDb` 习惯。
- `SecretHash` 使用 bcrypt（cost = 默认），即使 `api_tokens.json` 泄露也无法直接拿到原始 secret。
