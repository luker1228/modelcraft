# APISIX Gateway 迁移设计

**日期：** 2026-05-16  
**状态：** 已确认，待实施  
**范围：** 用 APISIX 完全替代 `modelcraft-gateway`

---

## 背景

现有 `modelcraft-gateway` 是一个自维护的 Go 服务，核心职责：
1. JWT 验签（ES256 Developer + HMAC-SHA256 EndUser）
2. 注入后端可信 Header（`X-User-ID`）
3. 路由分发
4. RequestID / 结构化日志 / OpenTelemetry Trace
5. Refresh Token 的 httpOnly Cookie 管理

职责 1-4 是标准网关能力，APISIX 原生插件均可覆盖。职责 5（Cookie 管理）属于认证响应的一部分，迁移到后端负责更合理。未来如需自定义逻辑，APISIX 插件生态（Lua）可以满足，无需自维护网关代码。

---

## 架构概览

```
Browser / CLI
      │
      ▼
   APISIX :9080
      │
      ├── /api/tenant/auth/*          → Backend（no-auth，透传）
      ├── /api/end-user/auth/*        → Backend（no-auth，透传）
      ├── /api/cli/end-user/auth/*    → Backend（no-auth，透传）
      ├── /graphql/*                  → Backend（jwt-auth 验签 → X-User-ID 注入）
      ├── /api/user/*                 → Backend（jwt-auth 验签 → X-User-ID 注入）
      └── /healthz                    → APISIX 直接响应 200
```

### 受保护路由插件栈

| 顺序 | 插件 | 作用 |
|------|------|------|
| 1 | `cors` | CORS，AllowedOrigins 白名单 |
| 2 | `request-id` | 生成 `X-Request-Id` |
| 3 | `opentelemetry` | W3C `traceparent` 透传 + OTLP 上报 |
| 4 | `jwt-auth` | ES256 验签，失败返回 401；**`hide_credentials: false`**（保留 Authorization header） |
| 5 | `serverless-pre-function` | Lua：从 JWT claims 提 `user_id` → 注入 `X-User-ID` |

公开路由（`/auth/*`）只挂 `cors` + `request-id` + `opentelemetry`，不走 `jwt-auth`。

`Authorization` header 原样透传到后端（`jwt-auth` 插件 `hide_credentials` 默认为 `false`，无需额外配置）。

---

## JWT 配置

- **统一算法：ES256**，Developer 和 EndUser token 使用同一对 EC 密钥
- **一个 APISIX Consumer：** `mc-user`，持有 EC 公钥
- **Token payload 新增字段：** `key: "mc-user"`（APISIX `jwt-auth` 插件要求）
- **保留 `aud` 字段：** `tenant` / `end_user`，供后端区分身份

---

## Docker Compose 集成

在现有 `docker-compose.yml` 中新增两个服务：

```yaml
services:
  etcd:
    image: bitnami/etcd:3.5
    environment:
      ALLOW_NONE_AUTHENTICATION: "yes"
    volumes:
      - etcd_data:/bitnami/etcd

  apisix:
    image: apache/apisix:3.9.0-debian
    ports:
      - "9080:9080"   # HTTP 入口（替代当前 gateway :8090）
      - "9180:9180"   # Admin API（本地调试用，生产环境不暴露此端口）
    volumes:
      - ./apisix/config.yaml:/usr/local/apisix/conf/config.yaml:ro
      - ./apisix/apisix.yaml:/usr/local/apisix/conf/apisix.yaml:ro
    depends_on:
      - etcd
      - backend

volumes:
  etcd_data:
```

**配置文件布局：**

```
apisix/
├── config.yaml       # APISIX 启动配置（etcd 地址、插件列表、OTel 端点）
└── apisix.yaml       # 声明式路由 + 插件配置（standalone 模式）
```

使用 **standalone 模式**（`apisix.yaml` 声明式配置），理由：
- 路由配置纳入 git 管理，和代码一起审查
- 本地开发启动即可用，无需额外 Admin API 初始化调用

---

## 后端需要改动的地方

### 1. JWT 签发统一为 ES256
- EndUser token 从 HMAC-SHA256 改为 ES256，使用与 Developer 相同的密钥对
- Token payload 新增 `key: "mc-user"` 字段

### 2. Cookie 管理移入后端
当前 Cookie 管理在 Gateway，迁移后后端自己负责：

| 端点 | 后端新增职责 |
|------|------------|
| `POST /auth/login` | 响应时 `Set-Cookie`（`mc_refresh_token`） |
| `POST /end-user/auth/login` | 响应时 `Set-Cookie`（`mc_enduser_refresh_token`） |
| `POST /auth/refresh` | 自己读请求 Cookie，换 token 后再 `Set-Cookie` |
| `POST /auth/logout` | 自己清除 Cookie（过期 Set-Cookie） |

Cookie 属性（`HttpOnly`、`SameSite=Strict`、`Secure`）通过环境变量配置化。

### 3. 后端信任 `X-User-ID` header（现状已满足）
APISIX 验签后原样透传 `Authorization` + 注入 `X-User-ID`，后端通过 `X-User-ID` 识别用户。根据现有架构文档，后端已经是这个模式，无需额外改动。

---

## 迁移顺序

```
Step 1  后端：Cookie 管理移入后端，EndUser token 改 ES256，token payload 加 key 字段
Step 2  APISIX：编写 apisix/config.yaml 和 apisix/apisix.yaml，Docker Compose 集成，本地联调验证
Step 3  切流：前端/CLI 请求目标从 :8090 改为 :9080，回归测试
Step 4  下线 modelcraft-gateway（删除 Docker Compose 中的 gateway 服务，代码归档或删除）
```

---

## 可以删掉的东西

迁移完成后，以下代码可以删除：

- `modelcraft-gateway/` 整个目录
- 或至少：
  - `internal/auth/service.go` 中的 Cookie 读写逻辑
  - `internal/proxy/handler.go` 中的 header 注入逻辑
  - `internal/auth/handler.go` 中的 Cookie 装饰逻辑

---

## 边界约定（不变）

1. 前端/CLI 所有请求必须经过 APISIX，禁止直连 Backend
2. Backend 通过 `X-User-ID` 识别调用方，APISIX 负责外层 token 验签
3. Backend 日志保留 `X-Request-Id` 与 `traceparent`，保证跨服务串联排障
