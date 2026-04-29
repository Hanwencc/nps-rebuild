# NPS Web UI (Vue 3 + TypeScript)

阶段 0：脚手架 + 登录 + 客户端 CRUD。后续阶段会陆续接入隧道、域名、仪表盘和全局参数。

## 目录约定

- 前端源码位于 `web-ui/`
- 构建产物输出到 `web/webui/dist/`，由 Go `embed` 一并打包进单个二进制
- 后端 API 入口：`/api/v1/*`（详见 `web/api/`）
- SPA 挂载点：`/ui/`（Hash 路由），与旧版 Beego 模板共存

## 开发环境

```powershell
# 1. 安装依赖
cd web-ui
yarn install

# 2. 启动 Vite 开发服务器（自动代理 /api -> http://127.0.0.1:8081）
yarn dev
# 默认地址：http://localhost:5173

# 3. 另开一个终端，启动 nps 后端
cd ..
go run ./cmd/nps
```

之后浏览器访问 http://localhost:5173 即可，登录后会持有 Beego session cookie。

## 生产构建

```powershell
cd web-ui
yarn build      # 输出到 ../web/webui/dist
cd ..
go build ./cmd/nps
./nps           # 访问 http://127.0.0.1:8081/ui/
```

## 技术栈

| 模块 | 选型 |
| ---- | ---- |
| 框架 | Vue 3.5 + `<script setup>` |
| 语言 | TypeScript 5.6 |
| 构建 | Vite 6 |
| 状态 | Pinia 2.3 |
| 路由 | Vue Router 4.5（Hash 模式） |
| UI   | Naive UI 2.40 |
| 样式 | Tailwind CSS v4（`@tailwindcss/vite`） |
| 网络 | axios 1.7（统一 envelope 拦截） |
| i18n | vue-i18n 10（zh-CN / en） |

## 鉴权双轨

后端中间件（`web/api/middleware.go`）允许两种方式：

1. **Session（默认）** — 通过 `/api/v1/auth/login` 登录后写入 Beego session，`withCredentials: true` 携带 cookie。
2. **签名直连** — 通过查询参数或请求头 `X-Auth-Key` + `X-Auth-Timestamp`，签名规则 `md5(auth_key + timestamp)`，时钟漂移容忍 ±20 秒。便于第三方脚本调用。

## 目录结构

```
web-ui/
├─ src/
│  ├─ api/          # axios 实例 + 各资源类型化客户端
│  ├─ stores/       # Pinia (auth / prefs)
│  ├─ i18n/         # 多语言资源
│  ├─ router/       # Vue Router + 守卫
│  ├─ layouts/      # 主布局（侧边栏 + 顶栏）
│  ├─ views/        # 页面级组件
│  └─ styles/       # Tailwind + 全局样式
├─ index.html
├─ vite.config.ts
└─ tsconfig*.json
```

## 后续阶段路线（建议）

- **Phase 1**：隧道（TCP/UDP/SOCKS5/HTTP/Secret/P2P/File）+ 域名解析（hosts）
- **Phase 2**：仪表盘（实时流量、连接数曲线，echarts）+ 全局参数
- **Phase 3**：删除旧版 Beego 模板（`web/views/*` 与 `web/static/page/*`），完全前后端分离
