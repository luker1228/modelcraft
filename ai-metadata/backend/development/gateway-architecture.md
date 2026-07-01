# ModelCraft Gateway 架构说明

> ⚠️ **已迁移（2026-05-16）**：`modelcraft-gateway` 已被 APISIX 替代。
> 本文档保留作历史参考，当前架构见 `apisix/apisix.template.yaml`。

> 适用范围：`modelcraft-gateway/`。本文档用于同步网关当前实现（以代码现状为准）。

## 1. 服务定位

`modelcraft-gateway` 是 ModelCraft 的统一入口网关，核心职责：

1. **认证编排**：
   - Developer 通道：校验后端签发的 ES256 Access Token
   - EndUser 通道：校验 HMAC-SHA256 Access Token（`JWT_SECRET`）
2. **反向代理**：将请求转发到后端，并注入/透传后端信任头（如 `X-User-ID`、`X-Internal-Token`）
3. **会话安全**：托管 refreshToken 的 httpOnly Cookie（浏览器不可读）
4. **可观测性**：统一请求 ID、结构化日志、OpenTelemetry Trace 透传

关键入口：`modelcraft-gateway/cmd/gateway/main.go`

---

## 2. 当前模块结构

```text
modelcraft-gateway/
├── cmd/gateway/main.go                 # 启动入口、路由装配、中间件顺序
├── internal/
│   ├── auth/
│   │   ├── service.go                  # JWT 校验与 refresh cookie 管理
│   │   └── handler.go                  # /api/tenant/auth 与 /api/end-user/auth 代理编排
│   ├── proxy/
│   │   ├── handler.go                  # GraphQL 代理（org/project，tenant + end-user 共用）
│   │   └── rest.go                     # /api/user/* REST 代理
│   ├── middleware/
│   │   ├── request_id.go               # 生成 X-Request-Id，校验 X-Client-Request-Id
│   │   ├── logger.go                   # request_start/request_end 结构化日志
│   │   └── zap_ctx.go                  # context 注入/提取 logger
│   ├── config/config.go                # 环境变量配置加载
│   └── telemetry/tracer.go             # OTLP TracerProvider 初始化
├── Dockerfile                          # 多阶段构建，非 root 运行，healthcheck
├── Jenkinsfile                         # Jenkins CI-only 流水线（PR/main/tag）
└── justfile                            # 本地开发、测试、lint、构建命令
```

---

## 3. 路由与认证矩阵

路由定义见：`modelcraft-gateway/cmd/gateway/main.go`

| 路由前缀 | 认证方式 | 主要能力 |
|---|---|---|
| `/api/tenant/auth/*` | 无（公开） | Developer 登录/注册/刷新/登出代理 |
| `/graphql/org/{orgName}` | Bearer ES256 | Org 级 GraphQL 代理（tenant + end-user 共用） |
| `/graphql/org/{orgName}/project/{projectSlug}` | Bearer ES256 | Project 级 GraphQL 代理（tenant + end-user 共用） |
| `/end-user/graphql/*`（Open Data API） | **Bearer PAT 仅此** | runtime SQL 查询，携带 `X-MC-Auth-*` 声明 end-user 角色 |
| `/api/user/*` | Bearer ES256 | 受保护 REST 代理 |
| `/healthz` | 无 | 健康检查 |

> **Open Data API 认证说明**：`/end-user/graphql/*` 目前只支持 PAT（Personal Access Token）认证。
> Access Token / 浏览器 session 不可用于此路径。调用方必须持有有效 PAT，并通过 `X-MC-Auth-*` headers 声明代表哪个终端用户。

---

## 4. 核心请求链路

### 4.1 中间件顺序

中间件装配顺序（见 `main.go`）：

1. CORS
2. `RequestID`（生成内部 `X-Request-Id`，仅保留合法 `X-Client-Request-Id`）
3. OTel Chi Middleware
4. Recoverer
5. RequestLogger

该顺序保证日志和 span 可关联到同一个 request id。

### 4.2 转发头策略

GraphQL/REST 代理都会执行以下策略：

- 删除外部 `Authorization` 后再转发（后端不再直接信任客户端 token）
- 注入 `X-User-ID`（来自已验证 JWT claims）
- 当下游仍需继续消费 access token 时，透传 `X-Internal-Token`
- 透传 `X-Request-Id` 与 `X-Client-Request-Id`
- 注入 W3C `traceparent` / `tracestate`

补充约束：

- **后端经 Gateway 调用其他下游服务接口时，必须透传 `X-Internal-Token`。**
- `X-Internal-Token` 承载的是当前请求链路中的 access token，不是固定 shared secret。
- 当前已知主要消费者是 `modelcraft-agent`；后续若新增下游服务，默认沿用同一透传规则。

实现位置：
- GraphQL 代理：`modelcraft-gateway/internal/proxy/handler.go`
- REST 代理：`modelcraft-gateway/internal/proxy/rest.go`

---

## 5. Token 与 Cookie 约定

实现位置：`modelcraft-gateway/internal/auth/service.go`

- Access Token：ES256（网关使用 `JWT_PUBLIC_KEY` 校验）
- Refresh Cookie：`mc_refresh_token`
- Cookie 属性：`HttpOnly + SameSite=Strict`

> 备注：代码中 `Secure=false`，仅适用于本地开发；生产需在 HTTPS 下开启 Secure。

---

## 6. 配置项（环境变量）

配置加载：`modelcraft-gateway/internal/config/config.go`

| 变量 | 必填 | 默认值 | 说明 |
|---|---|---|---|
| `GATEWAY_PORT` | 否 | `8090` | 网关监听端口 |
| `BACKEND_URL` | 否 | `http://localhost:8080` | 后端地址 |
| `JWT_PUBLIC_KEY` | 建议是 | 空 | ES256 公钥 |
| `FRONTEND_URL` | 否 | `http://localhost:3000` | CORS 白名单来源 |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | 否 | 空 | OTLP gRPC 上报端点（空则关闭导出） |
| `LOG_OUTPUT_PATH` | 否 | 空 | 日志输出文件（空则 stderr） |

---

## 7. 部署与 CI 现状

- Docker：多阶段构建 + 非 root 用户运行 + `/healthz` 健康检查
  - 文件：`modelcraft-gateway/Dockerfile`
- CI：Jenkins pipeline，当前按 **CI-only** 方式覆盖依赖校验、lint、test、二进制构建；
  在 `main/tag` 分支构建镜像，镜像仓库参数仍为占位变量。
  - 文件：`modelcraft-gateway/Jenkinsfile`

---

## 8. 与下游系统的边界约定

1. **前端（强制）**：所有浏览器侧与前端服务侧业务请求必须先到 Gateway，再由 Gateway 转发到 Backend；禁止任何直连 Backend 的 GraphQL/REST 调用。
2. **后端**：通过 `X-User-ID` 识别调用方，网关负责外层 token 校验；当后端继续经 Gateway 调用下游服务时，必须透传 `X-Internal-Token`。
   - **Developer 认证（design-time）**：Gateway 是唯一的 developer JWT 验签者。Gateway 验证 Bearer token 后删除 Authorization 头，并注入 `X-User-ID`；Backend 只信任该 header，不再直接校验 developer bearer token。
   - **Open Data API 认证（runtime）**：仅支持 PAT。Gateway 验证 PAT，并注入 `X-User-ID`；若下游链路仍需 access token，可通过 `X-Internal-Token` 继续透传。当前已知主要下游消费者是 `modelcraft-agent`。`X-MC-Auth-Userid-Str/Int`（RLS 上下文中的业务 end-user ID）由调用方在请求头中自行传入，Gateway 透传不修改。
3. **CLI**：CLI 必须走 `cli -> gateway -> backend` 路径，不得直接访问 Backend design-time 端点。
4. **可观测性**：后端日志应保留 `X-Request-Id` 与 `traceparent`，保证跨服务串联排障。

---

## 9. 本次元信息更新要点

本次同步后，ai-metadata 明确记录以下当前架构事实：

- 网关已是**双认证通道**（Developer + EndUser）
- 已支持 **EndUser GraphQL 专用路由**
- 认证与代理职责边界清晰（网关校验 token，后端信任注入头）
- 请求级 tracing / request-id / client-request-id 已形成统一链路
- Gateway 的 CI 路径为 Jenkins CI-only（PR/main/tag 触发）
- Developer / EndUser 身份体系全览：`./user-vs-end-user.md`
