# NPS 安全漏洞审计报告

> 审计时间：2026-04-24  
> 审计范围：全量源码（Go 1.24，模块 ehang.io/nps）  
> 标准参考：OWASP Top 10

---

## 漏洞总览

| # | 严重级别 | 漏洞名称 | 涉及文件 |
|---|----------|----------|----------|
| 1 | 🔴 严重 | 验证码校验缺少 `return`，可绕过暴力破解防护 | `web/controllers/login.go` |
| 2 | 🔴 严重 | TLS 客户端禁用证书验证，易遭中间人攻击 | `lib/crypt/tls.go`, `client/control.go` |
| 3 | 🔴 严重 | AES-CBC 使用固定 IV，加密强度严重削弱 | `lib/crypt/crypt.go` |
| 4 | 🔴 严重 | CORS 全局通配符暴露认证接口 | `web/controllers/auth.go` |
| 5 | 🟠 高危 | 密码明文存储与明文比较，存在时序侧信道 | `web/controllers/login.go` |
| 6 | 🟠 高危 | MD5 用于 VerifyKey/密码哈希验证 | `lib/file/db.go`, `lib/common/util.go` |
| 7 | 🟠 高危 | 攻击者可控 IP 参数注入白名单 | `web/controllers/auth.go` |
| 8 | 🟠 高危 | 文件服务器路径无限制，可泄露宿主机文件系统 | `client/local.go` |
| 9 | 🟠 高危 | 系统默认弱凭据 | `conf/nps.conf` |
| 10 | 🟡 中危 | API 认证使用 MD5+时间戳，20 秒重放窗口 | `web/controllers/base.go` |
| 11 | 🟡 中危 | 客户端版本检查已注释，允许旧版本接入 | `bridge/bridge.go` |
| 12 | 🟡 中危 | 安全敏感随机数使用 `math/rand` | `lib/crypt/crypt.go` |

---

## 详细分析

---

### 🔴 漏洞 1 — 验证码校验缺少 `return`（暴力破解绕过）

- **文件**: `web/controllers/login.go`
- **OWASP**: A07 认证失败
- **危害**: 验证码完全失效，攻击者可无限次暴力破解管理员密码

**问题代码**（`Verify` 方法）：
```go
if captchaOpen {
    if !cpt.VerifyReq(self.Ctx.Request) {
        self.Data["json"] = map[string]interface{}{...}
        self.ServeJSON()
        // ← 缺少 return！下方登录逻辑照常执行
    }
}
if self.doLogin(username, password, true) {
    ...
}
```

**修复方案**：
```go
if captchaOpen {
    if !cpt.VerifyReq(self.Ctx.Request) {
        self.Data["json"] = map[string]interface{}{
            "status": 0,
            "msg": "the verification code is wrong, please get it again and try again",
        }
        self.ServeJSON()
        return  // ← 添加 return
    }
}
```

---

### 🔴 漏洞 2 — TLS 客户端禁用证书验证（MITM 攻击）

- **文件**: `lib/crypt/tls.go:49`, `client/control.go:221`
- **OWASP**: A02 加密失败
- **危害**: NPC 客户端连接服务端时不验证服务端证书，攻击者可部署伪造服务端劫持全部隧道流量

**问题代码**：
```go
// lib/crypt/tls.go
func NewTlsClientConn(conn net.Conn) net.Conn {
    conf := &tls.Config{
        InsecureSkipVerify: true, // ← 跳过证书校验
    }
    return tls.Client(conn, conf)
}
```

**修复方案**：
- 在服务端颁发自签 CA，并将 CA 证书随安装包分发给客户端
- 客户端配置中加载 CA 证书进行验证：
```go
caCert, _ := os.ReadFile(caCertPath)
pool := x509.NewCertPool()
pool.AppendCertsFromPEM(caCert)
conf := &tls.Config{
    RootCAs:    pool,
    ServerName: serverHostname,
}
```
- 或使用证书指纹固定（Certificate Pinning）

---

### 🔴 漏洞 3 — AES-CBC 使用固定 IV（加密弱化）

- **文件**: `lib/crypt/crypt.go:21`
- **OWASP**: A02 加密失败
- **危害**: 相同明文始终产生相同密文，易遭已知明文攻击、块重放攻击

**问题代码**：
```go
func AesEncrypt(origData, key []byte) ([]byte, error) {
    block, _ := aes.NewCipher(key)
    blockSize := block.BlockSize()
    origData = PKCS5Padding(origData, blockSize)
    // IV 直接取 key 的前 16 字节，永远固定不变
    blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
    ...
}
```

**修复方案**：每次加密生成随机 IV，并将 IV 拼接到密文头部：
```go
func AesEncrypt(origData, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    origData = PKCS5Padding(origData, block.BlockSize())
    iv := make([]byte, block.BlockSize())
    if _, err = io.ReadFull(rand.Reader, iv); err != nil { // crypto/rand
        return nil, err
    }
    blockMode := cipher.NewCBCEncrypter(block, iv)
    crypted := make([]byte, len(origData))
    blockMode.CryptBlocks(crypted, origData)
    return append(iv, crypted...), nil // IV 前置
}

func AesDecrypt(crypted, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    bs := block.BlockSize()
    if len(crypted) < bs {
        return nil, errors.New("ciphertext too short")
    }
    iv := crypted[:bs]
    crypted = crypted[bs:]
    blockMode := cipher.NewCBCDecrypter(block, iv)
    origData := make([]byte, len(crypted))
    blockMode.CryptBlocks(origData, crypted)
    _, origData = PKCS5UnPadding(origData)
    return origData, nil
}
```

---

### 🔴 漏洞 4 — CORS 全局通配符暴露认证接口

- **文件**: `web/controllers/auth.go:48`
- **OWASP**: A05 安全配置错误
- **危害**: 攻击者构造恶意页面，诱导已登录管理员的浏览器向 `/auth/ipwhiteauth` 发起跨域请求，可将任意 IP 注入白名单

**问题代码**：
```go
func (s *AuthController) IpWhiteAuth() {
    s.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*") // ← 通配符
    ...
}
```

**修复方案**：
```go
// 从配置读取受信任源，或直接移除 CORS 头（该接口通常不需要跨域）
allowedOrigin := beego.AppConfig.String("cors_allowed_origin")
if allowedOrigin != "" {
    s.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
} else {
    // 不设置 CORS 头，拒绝跨域访问
}
```

---

### 🟠 漏洞 5 — 密码明文存储与明文比较

- **文件**: `web/controllers/login.go:78,99`, `conf/clients.json`
- **OWASP**: A02 加密失败
- **危害**: 配置文件或数据库泄露即可获取所有用户明文密码；`==` 字符串比较存在时序侧信道

**问题代码**：
```go
// 管理员密码明文对比
if password == beego.AppConfig.String("web_password") && username == beego.AppConfig.String("web_username")

// 子用户密码明文存储在 clients.json，明文对比
if !auth && v.WebPassword == password && v.WebUserName == username
```

**修复方案**：
- 密码存储改用 bcrypt（`golang.org/x/crypto/bcrypt`）
- 使用 `subtle.ConstantTimeCompare` 代替 `==` 比较

---

### 🟠 漏洞 6 — MD5 用于 VerifyKey/密码哈希

- **文件**: `lib/file/db.go:82,135,307`, `lib/common/util.go`
- **OWASP**: A02 加密失败
- **危害**: MD5 已被完全破解，可通过彩虹表/GPU 碰撞快速还原原始密钥

**问题代码**：
```go
if common.Getverifyval(v.VerifyKey) == vKey  // Getverifyval 内部 = MD5
if crypt.Md5(value.(*Tunnel).Password) == p  // 隧道密码 MD5
if crypt.Md5(v.VerifyKey) == vkey
```

**修复方案**：
- VerifyKey 验证改用 HMAC-SHA256 或 bcrypt
- 隧道 Password 用 bcrypt 存储和验证

---

### 🟠 漏洞 7 — 攻击者可控 IP 参数注入白名单

- **文件**: `web/controllers/auth.go:52`
- **OWASP**: A01 权限控制失效
- **危害**: 攻击者若知晓 `vkey` 和 `IpWhitePass`（均为明文），可将自己的 IP 写入任意客户端的 IP 白名单，绕过访问控制

**问题代码**：
```go
vkey := s.getEscapeString("vkey")
ip := s.getEscapeString("ip")        // ← 完全由请求方控制
password := s.getEscapeString("pass")
...
if !ipExists {
    c.IpWhiteList = append(c.IpWhiteList, ip) // ← 写入攻击者指定的 IP
}
```

**修复方案**：
- 强制使用服务端获取的请求来源 IP，完全忽略 `ip` 参数：
```go
ip := s.Ctx.Input.IP() // 不信任客户端传入的 ip 参数
```
- 同时对 `IpWhitePass` 进行哈希存储

---

### 🟠 漏洞 8 — 文件服务器路径无限制（宿主机文件泄露）

- **文件**: `client/local.go:72`
- **OWASP**: A01 权限控制失效 / A05 安全配置错误
- **危害**: 管理员（或被攻陷的管理员账号）将 `local_path` 设为 `/` 或 `C:\`，通过 file 模式隧道暴露 NPC 宿主机整个文件系统

**问题代码**：
```go
Handler: http.StripPrefix(t.StripPre, http.FileServer(http.Dir(t.LocalPath)))
// t.LocalPath 来自 Web 界面配置，无路径限制
```

**修复方案**：
- 服务端在保存任务前验证 `LocalPath`，禁止根目录及系统敏感路径
- NPC 客户端本地也应检查路径合法性：
```go
absPath, err := filepath.Abs(t.LocalPath)
if err != nil || absPath == "/" || absPath == "C:\\" {
    logs.Error("LocalPath is not allowed: %s", t.LocalPath)
    return
}
```

---

### 🟠 漏洞 9 — 系统默认弱凭据

- **文件**: `conf/nps.conf`
- **OWASP**: A07 认证失败
- **危害**: 攻击者可用默认账号直接登录任何未修改配置的 NPS 实例

**问题配置**：
```ini
web_username=admin
web_password=123
public_vkey=123
auth_key=123
auth_crypt_key=213
```

**修复方案**：
- 首次启动时检测是否为默认凭据，若是则强制跳转修改密码页面
- 或在安装时自动生成随机高强度密码并展示给用户

---

### 🟡 漏洞 10 — API 认证 MD5+时间戳，20 秒重放窗口

- **文件**: `web/controllers/base.go:43`
- **OWASP**: A07 认证失败
- **危害**: 20 秒内捕获的合法 API 请求可被原样重放；MD5 签名易被伪造

**问题代码**：
```go
// math.Abs(timeNowUnix - timestamp) <= 20 秒内有效
crypt.Md5(configKey + strconv.Itoa(timestamp)) == md5Key
```

**修复方案**：
- 改用 HMAC-SHA256 签名
- 引入一次性 nonce，服务端记录已用 nonce 防重放
- 或缩短时间窗口至 5 秒

---

### 🟡 漏洞 11 — 客户端版本检查已注释

- **文件**: `bridge/bridge.go:200`
- **危害**: 版本不匹配的旧版本客户端（可能含已知漏洞）可正常接入服务端

**问题代码**：
```go
if b, err := c.GetShortLenContent(); err != nil || string(b) != version.GetVersion() {
    // logs.Info("The client %s version does not match", c.Conn.RemoteAddr())
    // c.Close()   ← 已注释，不再拒绝
    // return
}
```

**修复方案**：取消注释或改为允许一定范围内的版本（已有 `GetVersion()` 返回最低兼容版本）。

---

### 🟡 漏洞 12 — 安全敏感场景使用 `math/rand`

- **文件**: `lib/crypt/crypt.go:75`
- **OWASP**: A02 加密失败
- **危害**: `math/rand` 是伪随机数生成器，以时间戳为种子，可被预测，用于生成 `auth_key` 等安全凭据存在风险

**问题代码**：
```go
func GetRandomString(l int) string {
    str := "0123456789abcdefghijklmnopqrstuvwxyz"
    r := rand.New(rand.NewSource(time.Now().UnixNano())) // math/rand
    ...
}
```

**修复方案**：
```go
import "crypto/rand"

func GetRandomString(l int) string {
    const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
    result := make([]byte, l)
    for i := range result {
        n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
        result[i] = chars[n.Int64()]
    }
    return string(result)
}
```

---

## 修复优先级

| 优先级 | 漏洞编号 | 建议完成时间 |
|--------|----------|-------------|
| P0（立即修复） | #1 验证码绕过 | 1 天内 |
| P0（立即修复） | #3 AES 固定 IV | 1 天内 |
| P0（立即修复） | #4 CORS 通配符 | 1 天内 |
| P1（尽快修复） | #2 TLS 不验证证书 | 1 周内 |
| P1（尽快修复） | #5 密码明文存储 | 1 周内 |
| P1（尽快修复） | #7 IP 白名单可伪造 | 1 周内 |
| P1（尽快修复） | #8 文件路径无限制 | 1 周内 |
| P1（尽快修复） | #9 默认弱凭据 | 1 周内 |
| P2（计划修复） | #6 MD5 哈希 | 1 个月内 |
| P2（计划修复） | #10 API 重放窗口 | 1 个月内 |
| P2（计划修复） | #11 版本检查注释 | 1 个月内 |
| P2（计划修复） | #12 math/rand | 1 个月内 |
