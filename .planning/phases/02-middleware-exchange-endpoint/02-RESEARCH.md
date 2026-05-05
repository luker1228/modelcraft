# 阶段 2：中间件 & Exchange 端点 — 研究

**研究时间**：2026-05-05  
**领域**：Go / chi HTTP 框架 / JWT ES256 / auth 中间件 / OpenAPI 代码生成  
**整体置信度**：HIGH（所有关键结论均经代码实证）

---

## 总结

阶段 1 已完成所有 JWT 格式统一工作：`PlatformClaims`（含 `scope` 字段）已定义，所有登录路径均产出 `mc-platform` issuer + `scope=org` 的 ES256 JWT。阶段 2 在此基础上新增两件事：

1. **`POST /api/auth/exchange`**：接收 Org Token，验证后签发同 orgName 的 `scope=project` Token，不携带 projectSlug。
2. **Scope 强制校验中间件**：`scope=org` 的 Token 无法访问 `/graphql/org/{orgName}/project/*`，`scope=project` 的 Token 无法访问 org 管理路由（`/graphql/org/{orgName}/` 非 project 路径），违者返回 403。

**核心发现**：backend 目前的 `ChiJWTAuthMiddleware` **不验证 JWT**，仅读取 gateway 注入的 `X-User-ID` Header。要做 scope 校验，backend 层需要直接解析 Bearer JWT 并提取 `scope` 字段，这与 exchange 端点共享同一个 JWT 解析逻辑（`JWTSigner.ParsePlatformClaims`，目前尚不存在，需新建）。

**主要建议**：在 `domain/auth` 层新增 `ParsePlatformClaims(token string) (*PlatformClaims, error)` 方法，Exchange Handler 和 Scope 中间件均复用此方法。Exchange 通过 OpenAPI spec + `just generate-oapi` 纳入生成流程，Scope 中间件以 chi Route 分组中间件方式注入。

---

<phase_requirements>
## 阶段需求

| ID | 描述 | 研究支撑 |
|----|------|---------|
| TOKEN-04 | `POST /api/auth/exchange` — 凭 Org Token 换取 Project Token（`scope=project`，不携带 projectSlug，TTL 1h） | JWTSigner 已有 `IssueAccessToken`，只需新增 JWT 解析方法即可实现 Exchange |
| TOKEN-05 | Scope 强制校验 — `scope=org` 无法访问 project 路由；`scope=project` 无法访问 org 管理路由 | routes.go 中 org/project GraphQL 路由已是独立 chi.Route 分组，可在 r.Use() 处直接插入 Scope 中间件 |
</phase_requirements>

---

## 架构责任映射

| 能力 | 主责层 | 辅助层 | 理由 |
|------|--------|--------|------|
| Exchange 业务逻辑（验证 Org Token + 签发 Project Token） | `internal/interfaces/http/handlers/auth/` | `internal/domain/auth/` | Exchange 是无状态的 token downscoping，无需额外 App Service；直接在 handler 中调用 JWTSigner |
| JWT Bearer 解析（backend 层） | `internal/domain/auth/` 新增 ParsePlatformClaims | — | 职责对称性：sign 在 domain，parse 也应在 domain |
| Scope 路由校验中间件 | `internal/middleware/` 新增 `chi_scope_auth.go` | — | 遵循现有中间件组织规范 |
| OpenAPI spec 定义（exchange 端点） | `api/openapi/auth.yaml` | `api/openapi/openapi.yaml` 引用 | auth.yaml 已管理所有 /api/auth/* 端点 |
| 路由注册（exchange） | `chi_setup.go` publicPaths + `server.go` + `generated.ServerInterface` | — | 现有 OpenAPI 路由流程的一部分 |
| Gateway scope 验证（是否需要） | **不需要** | — | exchange 端点在 gateway 层透明转发（类似 `/api/auth/login`），backend 自行解析 JWT |

---

## 1. 当前路由结构分析

### 1.1 Backend GraphQL 路由注册方式

[VERIFIED: 代码读取 routes.go, chi_setup.go]

backend 的 GraphQL 路由在 `routes.go` 中以**三个独立 chi Route 分组**注册：

```
/graphql/org/{orgName}/                        → SetupOrgGraphQLRoutesOnChi
    r.Use(middleware.ChiJWTAuthMiddleware)      ← 当前仅校验 X-User-ID 或 X-Internal-Token
    r.Use(middleware.ChiGraphQLOrgMiddleware)
    r.Post("/", orggraphql.OrgGraphQLHandler)
    r.Get("/",  orggraphql.OrgPlaygroundHandler)

/graphql/org/{orgName}/project/{projectSlug}/  → SetupProjectGraphQLRoutesOnChi
    r.Use(middleware.ChiJWTAuthMiddleware)
    r.Use(middleware.ChiGraphQLOrgMiddleware)
    r.Use(middleware.ChiGraphQLProjectMiddleware)
    r.Post("/", projectgraphql.ProjectGraphQLHandler)
    r.Get("/",  projectgraphql.ProjectPlaygroundHandler)

/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}  → SetupRuntimeGraphQLRoutesOnChi
    r.With(runtimeMW)  ← runtimeMW 内部调用 ChiJWTAuthMiddleware + ChiGraphQLOrgMiddleware
```

**路由分组是独立的**，可以在每个分组内单独添加 scope 中间件，无需修改全局路由注册逻辑。

### 1.2 ChiJWTAuthMiddleware 当前行为

[VERIFIED: 代码读取 chi_jwt_auth.go]

```go
// 当前三条认证路径（优先级从高到低）：
// 1. SkipValidation = true → 直接放行（dev/test）
// 2. X-Internal-Token header 匹配 → 放行（BFF 内部调用）
// 3. X-User-ID header 非空 → 放行，注入 UserID 到 context
// 4. 以上都不满足 → 401
```

**重要结论**：`ChiJWTAuthMiddleware` **不解析 Bearer JWT**。对于 `/graphql/org/*` 路由，请求必须先经过 Gateway，Gateway 在验证 JWT 后：
- 删除 `Authorization` header
- 注入 `X-User-ID` header

因此 backend 目前从 context 获取的是 UserID，没有 scope 信息。

### 1.3 `scope` 信息如何到达 backend

当前设计中 scope 信息**不到达 backend**（除了 runtime 端点通过 `RuntimeAuthMiddleware` 直接解析 Bearer JWT）。

对于 scope 强制校验，有两种可行方案：

**方案 A**：Gateway 在转发时注入 `X-Token-Scope` header（类似 X-User-ID）  
**方案 B**：Backend 中间件直接解析 Bearer JWT 提取 scope

分析：
- Gateway 目前的 `auth.Claims` 结构只有 `UserID` 字段，没有 `Scope` 字段（`VerifyAccessToken` 返回 `*Claims` 不含 scope）
- Backend 的 `RuntimeAuthMiddleware` 已经证明"backend 直接解析 Bearer JWT"是可行且已有先例的
- Exchange 端点本身就需要 backend 直接解析 Bearer JWT（exchange 是公开端点，不经过 Gateway JWT 验证流程）

**推荐方案 B**：Backend Scope 中间件直接解析 Bearer JWT。理由：
1. Exchange 端点已需要此能力，复用逻辑
2. 避免修改 Gateway auth 结构（减少 scope）
3. 与 `RuntimeAuthMiddleware` 现有模式一致

### 1.4 Gateway 路由注册现状

[VERIFIED: gateway/cmd/gateway/main.go]

```go
// 现有路由（所有 graphql 路由经 proxyHandler 转发）：
r.Post("/graphql/org/{orgName}", proxyHandler.GraphQLOrgHandler)
r.Post("/graphql/org/{orgName}/", proxyHandler.GraphQLOrgHandler)
r.Post("/graphql/org/{orgName}/project/{projectSlug}", proxyHandler.GraphQLProjectHandler)
r.Post("/graphql/org/{orgName}/project/{projectSlug}/", proxyHandler.GraphQLProjectHandler)
r.Post("/graphql/end-user/org/{orgName}/project/{projectSlug}", proxyHandler.EndUserGraphQLHandler)
```

`/api/auth/exchange` 目前不存在，需要在 gateway 层和 backend 层同时添加。

---

## 2. Exchange 端点实现策略

### 2.1 端点定位

[VERIFIED: 代码读取 auth/handler.go, routes.go, chi_setup.go]

`POST /api/auth/exchange` 应跟随现有 auth 端点的注册模式：

```
api/openapi/auth.yaml      → 定义 path + schemas
api/openapi/openapi.yaml   → $ref 引用 auth.yaml 的新 path
just generate-oapi         → 生成 generated/server.gen.go（新增 ExchangeToken operationId）
internal/interfaces/http/server.go    → 实现 Server.ExchangeToken()
internal/interfaces/http/handlers/auth/handler.go  → 实现 Handler.HandleExchange()
internal/interfaces/http/chi_setup.go → publicPaths 中添加 /api/auth/exchange
```

### 2.2 Exchange Handler 逻辑

```go
// POST /api/auth/exchange
// 1. 从 Authorization header 提取 Bearer token
// 2. 调用 jwtSigner.ParsePlatformClaims(tokenStr) 解析并验证签名
// 3. 校验 issuer == "mc-platform"（Validate() 内部已检查）
// 4. 校验 scope == "org"（不能用 project token 换，返回 400/403）
// 5. 调用 jwtSigner.IssueAccessToken(claims.UserID, claims.OrgName, "project") 签发新 token
// 6. 返回 { "accessToken": "<Project Token>", "expiresAt": "..." }
```

**关键依赖**：`Handler` 已有 `tokenService *appAuth.TokenService`，但 exchange 不需要 tokenService（无数据库操作）；需要 `jwtSigner`。

**Handler 实现位置选择**：

| 选择 | 位置 | 评估 |
|------|------|------|
| 选择 A（推荐） | 扩展现有 `authHandlers.Handler`，新增 `HandleExchange` 方法 | Handler 已持有 tokenService，可通过 tokenService 的 jwtSigner 引用；或者让 Handler 持有 jwtSigner |
| 选择 B | 新建 `ExchangeHandler` | 过度拆分，仅一个方法 |

推荐选择 A。具体地：`authHandlers.Handler` 目前依赖 `tokenService`，但 exchange 需要 `jwtSigner` 做解析和签发。

**分析 TokenService 是否暴露 jwtSigner**：查看 `internal/app/auth/token_service.go`，tokenService 内部持有 jwtSigner 但不对外暴露。

两个做法：
1. 在 `TokenService` 中新增 `ExchangeToken(orgToken string) (string, error)` 方法（推荐，符合 App Service 封装原则）
2. 让 `authHandlers.Handler` 额外持有 `jwtSigner`，handler 直接调用

推荐做法 1（通过 TokenService 封装 Exchange 业务逻辑），理由：
- Exchange 虽然无数据库操作，但属于认证业务逻辑范畴
- 保持 handler 薄，业务逻辑在 app service 层
- 方便测试 mock

### 2.3 Domain 层 ParsePlatformClaims 方法

[VERIFIED: domain/auth/jwt_signer.go 中只有 IssueAccessToken，无 Parse 方法]

`JWTSigner` 目前只有签发能力（`IssueAccessToken`），没有解析/验证能力。需要新增：

```go
// ParsePlatformClaims 解析并验证 ES256 JWT，返回 PlatformClaims。
// 验证内容：签名（ES256）、issuer（mc-platform）、过期时间、scope 合法性。
func (s *JWTSigner) ParsePlatformClaims(tokenStr string) (*PlatformClaims, error)
```

实现参考：`auth_handler.go` 的 `parseEndUserJWT` 方法已有完整的 ES256 + PlatformClaims 解析逻辑，可直接搬到 `JWTSigner` 上：

```go
func (s *JWTSigner) ParsePlatformClaims(tokenStr string) (*PlatformClaims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &PlatformClaims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return &s.privateKey.PublicKey, nil  // 直接用内部持有的公钥，无需 PEM 转换
    })
    if err != nil || !token.Valid {
        return nil, fmt.Errorf("invalid token: %w", err)
    }
    claims, ok := token.Claims.(*PlatformClaims)
    if !ok {
        return nil, fmt.Errorf("unexpected claims type")
    }
    if err := claims.Validate(); err != nil {
        return nil, err
    }
    return claims, nil
}
```

**注意**：`JWTSigner` 持有 `privateKey *ecdsa.PrivateKey`，公钥通过 `&s.privateKey.PublicKey` 直接取得，比 `parseEndUserJWT` 的 PEM 路径更高效。

### 2.4 OpenAPI Spec 管理

[VERIFIED: api/openapi/auth.yaml, api/openapi/openapi.yaml]

在 `api/openapi/auth.yaml` 中新增 exchange path 和对应 schema：

```yaml
# 在 auth.yaml paths 部分添加：
/api/auth/exchange:
  post:
    operationId: exchangeToken
    summary: Exchange Org Token for Project Token
    tags: [Auth]
    security:
      - bearerAuth: []
    requestBody:
      required: false
      content:
        application/json:
          schema:
            type: object
    responses:
      "200":
        description: Exchange successful
        content:
          application/json:
            schema:
              $ref: "auth.yaml#/schemas/ExchangeResponse"
      "400":
        description: Invalid scope (not org token)
      "401":
        description: Invalid or expired token
      "500":
        description: Server error
```

然后在 `api/openapi/openapi.yaml` 中引用此 path（与 `/api/auth/login` 等同一处理方式），运行 `just generate-oapi` 生成代码。

### 2.5 gateway 层是否需要改动

**Exchange 端点的 gateway 路由注册**：exchange 类似 login，是公开端点（需要 Bearer token 自带验证，不经过 gateway JWT 校验），应在 gateway 的 `/api/end-user/auth` 或 `/auth` 路由组旁边单独注册透传：

```go
// gateway/cmd/gateway/main.go 新增：
r.Post("/api/auth/exchange", authHandler.Exchange)
```

`authHandler.Exchange` 实现：直接透传到 backend `/api/auth/exchange`，**不做 JWT 验证**（backend 自己验证），类似 `authHandler.EndUserMe`（透传 Authorization header）。

---

## 3. Scope 中间件实现策略

### 3.1 设计决策：Backend 层直接解析 JWT

[VERIFIED: 分析 chi_jwt_auth.go + runtime_auth_middleware.go + 方案对比]

Scope 中间件放在 **backend** 层（不在 gateway），直接从 `Authorization: Bearer <JWT>` 解析 scope。

理由：
1. gateway 目前的 `GraphQLOrgHandler` / `GraphQLProjectHandler` 会**删除 Authorization header**，替换为 `X-User-ID`；scope 信息丢失
2. 若要在 gateway 层做 scope 校验，需要修改 gateway auth.Claims 结构 + handler 逻辑，涉及 gateway 仓库变更
3. backend 层的 `RuntimeAuthMiddleware` 已有直接解析 Bearer JWT 的先例

**关键问题**：gateway 删除 Authorization header 后，backend 的 scope 中间件拿不到 Bearer token。

**解决方案**：修改 gateway 的 `director` 函数，在注入 `X-User-ID` 的同时**也透传 `X-Token-Scope` header**（从解析好的 claims 取），或者更简洁地：

> **方案 C（最终推荐）**：在 gateway 的 `GraphQLOrgHandler` 和 `GraphQLProjectHandler` 中，解析 JWT claims 后注入 `X-Token-Scope` header（值为 `"org"` 或 `"project"`），backend scope 中间件从此 header 读取 scope，无需再次验证签名。

这样 backend 的 scope 校验逻辑极其简单：

```go
// chi_scope_auth.go
func RequireScope(allowedScopes ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            scope := r.Header.Get("X-Token-Scope")
            if scope == "" {
                // fallback: SkipValidation 模式或 InternalToken 模式 → 放行
                next.ServeHTTP(w, r)
                return
            }
            for _, s := range allowedScopes {
                if scope == s {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            writeJSONError(w, http.StatusForbidden, "insufficient scope", "INSUFFICIENT_SCOPE")
        })
    }
}
```

**这需要修改 gateway 两个 handler**（`GraphQLOrgHandler` + `GraphQLProjectHandler`）注入 `X-Token-Scope`，但修改量极小且不影响现有认证逻辑。

**替代方案（方案 B 更纯粹）**：若不想动 gateway，可以在 backend scope 中间件中对 Bearer JWT 做二次解析。但需要处理：
1. SkipValidation 模式（无 JWT，直接放行）
2. X-Internal-Token 模式（无 JWT，直接放行）
3. X-User-ID 模式（gateway 已删除 Authorization，scope 信息丢失）

**结论**：方案 C（gateway 注入 X-Token-Scope）最简洁，方案 B 需要 backend 二次解析 JWT 但 gateway 已删除 Authorization header，两者都需要少量 gateway 改动。**最终推荐方案 C**，gateway 改动仅 2 行，backend scope 中间件无需 JWT 解析。

### 3.2 Scope 中间件注入位置

[VERIFIED: routes.go 路由分组结构]

```go
// SetupOrgGraphQLRoutesOnChi — 仅允许 scope=org
router.Route("/graphql/org/{orgName}", func(r chi.Router) {
    r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
    r.Use(middleware.RequireScope(domainAuth.TokenScopeOrg))  // ← 新增，拒绝 scope=project
    r.Use(middleware.ChiGraphQLOrgMiddleware())
    r.Post("/", orggraphql.OrgGraphQLHandler(orgResolver))
    r.Get("/", orggraphql.OrgPlaygroundHandler())
})

// SetupProjectGraphQLRoutesOnChi — 仅允许 scope=project
router.Route("/graphql/org/{orgName}/project/{projectSlug}", func(r chi.Router) {
    r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
    r.Use(middleware.RequireScope(domainAuth.TokenScopeProject))  // ← 新增，拒绝 scope=org
    r.Use(middleware.ChiGraphQLOrgMiddleware())
    r.Use(middleware.ChiGraphQLProjectMiddleware())
    r.Post("/", projectgraphql.ProjectGraphQLHandler(projectResolver))
    r.Get("/", projectgraphql.ProjectPlaygroundHandler())
})
```

**Runtime 路由**（`/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}`）：
- 当前通过 `RuntimeAuthMiddleware` 做自己的 JWT 解析（直接读 Bearer）
- 需同时允许 `scope=project`，且 RuntimeAuthMiddleware 已有 scope 注入（第 115 行：`scope, _ := claims["scope"].(string)`）
- Runtime 路由应添加 `RequireScope(TokenScopeProject)` 或在 `runtimeMW` 内部补充 scope 检查

**注意**：`/graphql/end-user/org/{orgName}/project/{projectSlug}` 路由使用 `X-Internal-Token`（BFF 调用），scope 中间件需要**在 X-Internal-Token 模式下自动放行**（与方案 C 的 fallback 逻辑一致）。

### 3.3 X-Internal-Token 路由的兼容性

[VERIFIED: chi_jwt_auth.go tryInternalTokenAuth]

`SetupEndUserGraphQLRoutesOnChi` 中的路由使用 `internalTokenMW`（独立中间件），与 `ChiJWTAuthMiddleware` 不同路由组，**不受 scope 中间件影响**，无需额外处理。

---

## 4. JWT 解析在 Backend 的处理方式

### 4.1 当前 Backend JWT 解析现状

[VERIFIED: chi_jwt_auth.go, runtime_auth_middleware.go, auth_handler.go]

backend 目前有两种 JWT 解析路径：

| 路径 | 位置 | 场景 |
|------|------|------|
| 不解析 JWT，信任 X-User-ID | `ChiJWTAuthMiddleware` | 设计时 GraphQL（org/project route） |
| 直接解析 Bearer JWT | `RuntimeAuthMiddleware.Middleware()` | Runtime GraphQL |
| 直接解析 Bearer JWT | `enduser.AuthHandler.parseEndUserJWT()` | GET /api/end-user/auth/me |

**Exchange 端点的 JWT 解析**：必须直接解析 Bearer JWT（因为 exchange 是公开端点，到达 backend 前 gateway 不做 JWT 校验）。

### 4.2 新增 ParsePlatformClaims 方法

在 `domain/auth/jwt_signer.go` 中新增 `ParsePlatformClaims`，统一解析逻辑：

```go
// ParsePlatformClaims 使用本 signer 的 ES256 公钥验证并解析 PlatformClaims JWT。
// 解析后会调用 claims.Validate() 做完整性校验（issuer、scope、expiry）。
func (s *JWTSigner) ParsePlatformClaims(tokenStr string) (*PlatformClaims, error)
```

此方法被以下组件复用：
- `app/auth/TokenService.ExchangeToken(orgToken string)` 
- 可选：scope 中间件（若采用方案 B 在 backend 层二次解析）

### 4.3 Exchange 中不需要额外 App Service 依赖

Exchange 只做：验证 token → 签发新 token。无数据库操作，无外部依赖。逻辑轻薄，可以：
- 放在 `TokenService` 中（推荐，保持业务逻辑封装）
- 或在 handler 中直接调用 `jwtSigner.ParsePlatformClaims` + `jwtSigner.IssueAccessToken`

---

## 5. 变更范围（文件列表）

### Backend（modelcraft-backend）

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `internal/domain/auth/jwt_signer.go` | 修改 | 新增 `ParsePlatformClaims(tokenStr string) (*PlatformClaims, error)` |
| `internal/domain/auth/jwt_signer_test.go` | 修改 | 新增 ParsePlatformClaims 单元测试 |
| `internal/app/auth/token_service.go` | 修改 | 新增 `ExchangeToken(orgToken string) (string, time.Time, error)` |
| `internal/app/auth/token_service_test.go` | 修改 | 新增 ExchangeToken 测试 |
| `internal/interfaces/http/handlers/auth/handler.go` | 修改 | 新增 `HandleExchange(w, r)` 方法 |
| `internal/middleware/chi_scope_auth.go` | **新建** | `RequireScope(allowedScopes ...string)` 中间件 |
| `internal/middleware/chi_scope_auth_test.go` | **新建** | RequireScope 单元测试 |
| `internal/interfaces/http/routes.go` | 修改 | 在 org/project GraphQL 路由组中添加 `r.Use(middleware.RequireScope(...))` |
| `internal/interfaces/http/server.go` | 修改 | 新增 `func (s *Server) ExchangeToken(w, r)` 委托方法 |
| `internal/interfaces/http/chi_setup.go` | 修改 | `publicPaths` 中添加 `/api/auth/exchange` |
| `api/openapi/auth.yaml` | 修改 | 新增 `/api/auth/exchange` path 定义 + `ExchangeResponse` schema |
| `api/openapi/openapi.yaml` | 修改 | 引用 `auth.yaml#/paths/~1api~1auth~1exchange` |

### Gateway（modelcraft-gateway）

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `internal/proxy/handler.go` | 修改 | `GraphQLOrgHandler` 和 `GraphQLProjectHandler` 解析 claims 后注入 `X-Token-Scope` header |
| `internal/auth/service.go` | 修改 | `Claims` 结构体新增 `Scope string` 字段（从 JWT 中解析） |
| `internal/auth/handler.go` | 修改 | 新增 `Exchange` 方法（透传 Bearer token 到 backend） |
| `cmd/gateway/main.go` | 修改 | 注册 `r.Post("/api/auth/exchange", authHandler.Exchange)` |

**合计：约 12 个文件**

---

## 6. 风险点与注意事项

### 风险 1：X-Token-Scope Header 被外部伪造

**描述**：如果外部请求可以直接带 `X-Token-Scope: org` header 绕过 scope 校验  
**现状**：backend 仅在内网（不直接对公网），通过 gateway 转发；gateway 在 `director` 函数中会**覆写**（`req.Header.Set`）X-Token-Scope，外部设置的同名 header 会被覆盖  
**风险等级**：LOW — 但仍需在 gateway 的 `director` 函数中先 `Del("X-Token-Scope")` 再 `Set`，防止直连 backend 的场景泄露

### 风险 2：SkipValidation 模式下 scope 校验失效

**描述**：dev/test 环境 `SkipValidation: true` 时，gateway 不验证 JWT 不会注入 X-Token-Scope，backend scope 中间件收不到 header  
**处理方式**：`RequireScope` 在 X-Token-Scope 为空时**放行**（与现有 SkipValidation 逻辑一致）；或者 scope 中间件也检查 `SkipValidation` 配置  
**推荐**：当 X-Token-Scope 为空（包括 X-Internal-Token 模式）时直接放行，保持与现有 auth 中间件的 bypass 语义一致

### 风险 3：Runtime 路由 scope 兼容

**描述**：`SetupRuntimeGraphQLRoutesOnChi` 中的 `runtimeMW` 直接解析 Bearer JWT（不经过 gateway 的 X-Token-Scope 注入路径），scope 信息已在 `RuntimeAuthMiddleware` 中注入到 context  
**处理方式**：Runtime 路由可以单独在 `runtimeMW` 内部检查 scope（从 context 的 `EndUserIdentity.Scope` 读取），而不是复用 `RequireScope`（后者读 X-Token-Scope header）  
**注意**：不要对 runtime 路由施加 X-Token-Scope 检查，因为 runtime 有自己的 Bearer 解析路径

### 风险 4：Exchange 端点的 Gateway 路由定位

**描述**：`/api/auth/exchange` 应该放在 gateway 的哪个路由分组？  
**现状**：gateway 的公开 auth 路由分组是 `/auth/`（映射到 `/api/auth/`）；而 exchange 路径是 `/api/auth/exchange`  
**当前 gateway 路由**：`/auth/login` → backend `/api/auth/login`，但 gateway 是 `/auth/`，不是 `/api/auth/`  
**关键发现**：gateway 的 `/auth/` 路由组使用了 `authHandler.Login` 等直接处理（带 cookie 管理），路径前缀**不含 `/api`**；但前端访问的是 `/api/auth/login`，说明前端访问的是 backend 直接暴露的 `/api/auth/*`，不经过 gateway 的 `/auth/` 路由

需要澄清：前端实际调用的路径是 `/api/auth/exchange` 还是 gateway 的某个路径？根据现有设计，exchange 应该跟随 `/api/auth/*` 的路由风格，直接在 gateway 注册透传（参考 `EndUserMe` 的实现方式）。

### 风险 5：scope 校验遗漏 GET Playground 端点

**描述**：`/graphql/org/{orgName}/` 路由组有 GET Playground（`r.Get("/",...)`），scope 中间件同样会应用到 GET 请求  
**影响**：开发者在浏览器访问 Playground 时需要有 `scope=org` token  
**处理方式**：scope 中间件只在 X-Token-Scope header 存在时才做校验（空时放行），dev 环境 SkipValidation=true 时不会有问题

### 风险 6：openapi.yaml 引用语法

[VERIFIED: auth.yaml 中现有路径定义格式]

现有 `openapi.yaml` 通过 `paths` 合并的方式引用各 domain yaml，需要确认 exchange path 是否以同样方式合并（工作流是 `auth.yaml` 定义 path，然后 openapi.yaml 用 `$ref` 或 inline 引用）。建议直接在 `auth.yaml` 的 `paths:` 下追加，与现有格式保持一致。

---

## 7. 代码示例

### 7.1 ParsePlatformClaims（domain/auth/jwt_signer.go）

```go
// ParsePlatformClaims 使用本 signer 的 ES256 公钥验证并解析 PlatformClaims JWT。
func (s *JWTSigner) ParsePlatformClaims(tokenStr string) (*PlatformClaims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &PlatformClaims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return &s.privateKey.PublicKey, nil
    })
    if err != nil || !token.Valid {
        return nil, fmt.Errorf("jwt_signer: parse claims: %w", err)
    }
    claims, ok := token.Claims.(*PlatformClaims)
    if !ok {
        return nil, errors.New("jwt_signer: unexpected claims type")
    }
    if err := claims.Validate(); err != nil {
        return nil, fmt.Errorf("jwt_signer: invalid claims: %w", err)
    }
    return claims, nil
}
```

### 7.2 ExchangeToken（app/auth/token_service.go）

```go
// ExchangeToken 用 Org Token 换取 Project Token。
// 不做 RBAC 校验；scope 降级逻辑：org → project。
func (s *TokenService) ExchangeToken(orgToken string) (string, time.Time, error) {
    claims, err := s.jwtSigner.ParsePlatformClaims(orgToken)
    if err != nil {
        return "", time.Time{}, fmt.Errorf("exchange: invalid org token: %w", err)
    }
    if claims.Scope != domainAuth.TokenScopeOrg {
        return "", time.Time{}, bizerrors.New(ErrInvalidScope, "token scope must be org for exchange")
    }
    projectToken, err := s.jwtSigner.IssueAccessToken(claims.UserID, claims.OrgName, domainAuth.TokenScopeProject)
    if err != nil {
        return "", time.Time{}, fmt.Errorf("exchange: issue project token: %w", err)
    }
    expiresAt := time.Now().Add(time.Duration(s.jwtSigner.TTLSeconds()) * time.Second)
    return projectToken, expiresAt, nil
}
```

### 7.3 RequireScope 中间件（middleware/chi_scope_auth.go）

```go
package middleware

import "net/http"

const headerTokenScope = "X-Token-Scope"

// RequireScope 返回一个 chi 中间件，校验请求的 X-Token-Scope header 是否在 allowedScopes 中。
// 若 header 为空（SkipValidation 或 X-Internal-Token 模式），直接放行。
func RequireScope(allowedScopes ...string) func(http.Handler) http.Handler {
    allowed := make(map[string]bool, len(allowedScopes))
    for _, s := range allowedScopes {
        allowed[s] = true
    }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            scope := r.Header.Get(headerTokenScope)
            if scope == "" {
                // header 为空：SkipValidation / InternalToken 模式，放行
                next.ServeHTTP(w, r)
                return
            }
            if !allowed[scope] {
                writeJSONError(w, http.StatusForbidden, "insufficient scope", "INSUFFICIENT_SCOPE")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### 7.4 Gateway X-Token-Scope 注入（gateway/proxy/handler.go）

```go
// 在 GraphQLOrgHandler 和 GraphQLProjectHandler 中，解析 claims 后注入 scope
func (h *Handler) GraphQLOrgHandler(w http.ResponseWriter, r *http.Request) {
    claims, ok := h.extractAndVerify(w, r)
    if !ok {
        return
    }
    ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
    // 新增：将 scope 存入 context，director 函数中注入 X-Token-Scope
    ctx = context.WithValue(ctx, tokenScopeContextKey, claims.Scope)
    h.reverseProxy.ServeHTTP(w, r.WithContext(ctx))
}

// 在 director 函数中注入：
if scope, ok := req.Context().Value(tokenScopeContextKey).(string); ok && scope != "" {
    req.Header.Del("X-Token-Scope")  // 防止外部伪造
    req.Header.Set("X-Token-Scope", scope)
}
```

**注意**：`auth.Claims` 目前只有 `UserID` 字段，需要新增 `Scope string` 字段，`VerifyAccessToken` 解析时从 JWT 中提取 `scope` claim。

---

## 8. 环境可用性

阶段 2 是纯代码/配置变更，无外部依赖。

| 依赖 | 状态 | 说明 |
|------|------|------|
| `just generate-oapi` | ✓ 可用 | 修改 auth.yaml 后需运行 |
| `just build` | ✓ 可用 | 常规构建验证 |
| `just lint` | ✓ 可用 | pre-commit 钩子强制 |
| JWT ES256（ecdsa/p256） | ✓ 已集成 | golang-jwt/jwt/v5 已在使用 |

---

## 9. 验证架构

### 测试框架
| 属性 | 值 |
|------|-----|
| 框架 | Go 标准 testing |
| 快速运行 | `cd modelcraft-backend && go test ./internal/domain/auth/... ./internal/app/auth/... ./internal/middleware/... -v` |
| 完整运行 | `just check-all` |

### 需求 → 测试映射

| 需求 ID | 行为 | 测试类型 | 测试命令 |
|---------|------|---------|---------|
| TOKEN-04 | ExchangeToken 成功返回 scope=project token | 单元 | `go test ./internal/app/auth/... -run TestExchangeToken` |
| TOKEN-04 | Exchange 用 project token 调用返回错误 | 单元 | `go test ./internal/app/auth/... -run TestExchangeToken_InvalidScope` |
| TOKEN-04 | Exchange 用过期 token 返回 401 | 单元 | `go test ./internal/app/auth/... -run TestExchangeToken_Expired` |
| TOKEN-05 | scope=org token 访问 project 路由返回 403 | 单元（中间件） | `go test ./internal/middleware/... -run TestRequireScope` |
| TOKEN-05 | scope=project token 访问 org 路由返回 403 | 单元（中间件） | `go test ./internal/middleware/... -run TestRequireScope` |
| TOKEN-05 | X-Internal-Token 请求绕过 scope 校验 | 单元（中间件） | `go test ./internal/middleware/... -run TestRequireScope_InternalToken` |

### Wave 0 缺口

- [ ] `internal/middleware/chi_scope_auth_test.go` — RequireScope 中间件测试（新文件）
- [ ] `internal/domain/auth/jwt_signer_test.go` — ParsePlatformClaims 测试用例（已有文件，追加）

---

## 开放问题

1. **Gateway exchange 路由归属** [RESOLVED]：gateway 的 `/auth/login` 路由对应 frontend 访问 `/auth/login` 还是 `/api/auth/login`？需要确认 exchange 端点在 gateway 的挂载路径，以匹配 frontend 实际调用路径。**解决方案**：PLAN-02 任务 3 要求执行者先读取 gateway main.go 确认实际路由前缀，并在 SUMMARY 中记录最终外部 URL（预期为 `/auth/exchange`，无 `/api` 前缀）。acceptance_criteria 使用 `grep -v '^//' | grep -c 'exchange'` 验证路由注册存在。

2. **Runtime 路由 scope 校验策略** [RESOLVED]：Runtime 端点走自己的 `RuntimeAuthMiddleware`（直接解析 Bearer），不走 gateway 注入的 X-Token-Scope 路径。**解决方案**：在 `RuntimeAuthMiddleware.Middleware()` 内部，在注入 `EndUserIdentity` 之前补充 `scope != TokenScopeProject` 校验，返回 403，见 PLAN-02 任务 3 步骤 3b。

3. **auth.Claims.Scope 扩展** [RESOLVED by Wave 1]：gateway 的 `auth.Claims` 结构体在 PLAN-01 中已扩展为含 `Scope string` 字段，`VerifyAccessToken` 解析后可取到 scope 供 X-Token-Scope 注入使用。

---

## 假设清单

| # | 假设内容 | 章节 | 错误风险 |
|---|---------|------|---------|
| A1 | Exchange 端点不需要 gateway 做 JWT 校验，backend 自行解析 | §2.5 | 如错误，需在 gateway 新增独立验证路径 |
| A2 | `/api/auth/exchange` 挂载路径与现有 `/api/auth/login` 保持一致（含 `/api` 前缀） | §6.风险4 | 如 gateway 路径前缀不含 `/api`，需调整 gateway 路由注册位置 |

---

## 来源

### 主要（HIGH 置信度）

- `modelcraft-backend/internal/interfaces/http/routes.go` — 路由分组完整结构
- `modelcraft-backend/internal/middleware/chi_jwt_auth.go` — 中间件实现
- `modelcraft-backend/internal/domain/auth/jwt_signer.go` — JWTSigner API
- `modelcraft-backend/internal/domain/auth/platform_claims.go` — PlatformClaims + 常量
- `modelcraft-backend/internal/interfaces/http/chi_setup.go` — 路由注册全貌
- `modelcraft-gateway/internal/proxy/handler.go` — gateway 转发逻辑
- `modelcraft-gateway/internal/auth/service.go` — gateway JWT 验证
- `modelcraft-gateway/cmd/gateway/main.go` — gateway 路由注册
- `.planning/phases/01-token-core-unified/01-01/02/03-SUMMARY.md` — 阶段 1 完成产物

### 次要（MEDIUM 置信度）

- `api/openapi/auth.yaml` — 现有 auth spec 格式（用于推断 exchange spec 格式）
- `internal/interfaces/http/middleware/runtime_auth_middleware.go` — 已有 Bearer 解析先例

---

## 元数据

**置信度分布**：
- 路由结构分析：HIGH（直接代码验证）
- Exchange 实现策略：HIGH（所有依赖已存在）
- Scope 中间件策略：HIGH（模式已有先例）
- JWT 解析策略：HIGH（ParsePlatformClaims 方法已有 parseEndUserJWT 原型）
- Gateway 改动范围：MEDIUM（X-Token-Scope 注入逻辑需对照 gateway auth.Claims 扩展）

**研究日期**：2026-05-05  
**有效期估计**：30 天（代码稳定期内）

---

## RESEARCH COMPLETE

**阶段**：2 — 中间件 & Exchange 端点  
**置信度**：HIGH

### 关键发现

1. **backend 当前不持有 JWT scope 信息**：`ChiJWTAuthMiddleware` 仅读取 `X-User-ID`（由 gateway 注入），scope 在 gateway 层已被丢弃。Scope 中间件必须通过 gateway 注入 `X-Token-Scope` header（方案 C），或通过 backend 二次解析 Bearer JWT（方案 B，但 gateway 已删除 Authorization header）。**推荐方案 C**。

2. **`JWTSigner` 缺少 Parse 方法**：现有 `JWTSigner` 只有 `IssueAccessToken`，需新增 `ParsePlatformClaims` 方法供 Exchange 和 scope 验证使用。实现参考 `auth_handler.go.parseEndUserJWT`，但可直接用 `&s.privateKey.PublicKey` 无需 PEM 转换。

3. **两个独立 chi 路由组已存在**：`/graphql/org/{orgName}` 和 `/graphql/org/{orgName}/project/{projectSlug}` 是完全独立的 chi Route 分组，在各自的 `r.Use()` 链中插入 `RequireScope` 中间件即可，无需重构路由结构。

4. **Exchange 端点走 OpenAPI 生成流程**：与 `/api/auth/login` 同一流程（auth.yaml → openapi.yaml → `just generate-oapi` → server.gen.go → server.go 委托）。

5. **Gateway Claims 结构需扩展**：`auth.Claims` 只有 `UserID`，`VerifyAccessToken` 需要同时提取 `scope` 供 X-Token-Scope 注入使用，需在 gateway 层扩展 Claims + VerifyAccessToken 返回值。

### 创建文件

`.planning/phases/02-middleware-exchange-endpoint/02-RESEARCH.md`

### 可开始规划

研究完成。规划者可以基于本研究创建 PLAN.md 文件，建议分 2 个 Wave：
- **Wave 1**：domain 层（ParsePlatformClaims）+ app 层（ExchangeToken）+ gateway Claims 扩展 + X-Token-Scope 注入
- **Wave 2**：接口层（OpenAPI spec + 生成 + handler + server.go 委托 + chi_setup.go 注册）+ scope 中间件 + routes.go 注入
