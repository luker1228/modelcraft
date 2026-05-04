# Spec: 统一 Token 体系 & Workspace 入口重构

> **状态**: 待实现  
> **优先级**: 高  
> **本期范围**: 后端 Token 统一 + 删除 end_user GraphQL schema + 前端 Workspace 入口  
> **下期范围**: Service Key（应用级）、功能权限 RBAC（admin vs 数据人员 tab 差异）

---

## Objective

### 问题

当前系统存在两套并行的 Token/认证体系：

| 体系 | Issuer | 认证方式 | GraphQL 端点 |
|------|--------|----------|-------------|
| 平台管理员（developer） | `mc-developer` | ES256 JWT → gateway 注入 X-User-ID | `/graphql/org/` + `/graphql/org/project/` |
| 端用户（end-user） | `mc-enduser` | HMAC JWT → X-Internal-Token | `/graphql/end-user/org/.../project/...` |

两套体系导致：
- 认证中间件逻辑分叉，维护成本高
- `api/graph/end_user/` 是一个只读的"影子 schema"，与 project schema 重叠但不统一
- 端用户的 CRUD 数据接口走不到 runtime GraphQL，体验割裂

### 目标

1. **Token 统一**：废弃 `mc-enduser` issuer，所有用户使用同一种 JWT 格式，靠 `scope` claim 区分权限范围
2. **Schema 统一**：删除 `api/graph/end_user/` 及对应 handler/resolver/路由，端用户复用 org/project schema
3. **Workspace 入口**：端用户登录后进入 project 列表 → 选择 project → 进入 workspace（复用 runtime GraphQL 做数据 CRUD）
4. **登录入口保持分离**：平台管理员走 `/api/auth/login`，端用户走 `/api/end-user/{orgSlug}/auth/login`，token 格式统一但入口隔离

### 用户故事

- **平台管理员**：登录 → 拿 Org Token → 进入 org dashboard 管理项目/成员
- **端用户**：访问 `3000/end-user/{orgSlug}/login` → 登录 → 看到可访问的 project 列表 → 点击进入 → 拿 Project Token → 进入 workspace → 做模型数据 CRUD

---

## Tech Stack

- **语言**: Go 1.22+
- **HTTP 框架**: chi
- **GraphQL**: gqlgen
- **JWT**: ES256（ECDSA P-256），现有 `jwt_signer.go`
- **前端**: Next.js，BFF 架构，GraphQL Codegen

---

## Commands

```bash
# 构建
just build

# 运行（开发）
just run

# 代码检查（pre-commit hook 自动触发）
just lint

# 自动修复 lint 问题
just lint-fix

# 生成 GraphQL 代码（修改 .graphql 文件后必须运行）
just generate-gql

# 生成 OpenAPI 代码（修改 .yaml 文件后必须运行）
just generate-oapi
```

---

## 核心设计

### Token 分层

```
登录
  └─► Org Token
        scope: "org"
        orgName: "acme"
        TTL: 1h（refresh token 自动续期）
        ├─► 可访问 /graphql/org/{orgName}/* （org 管理接口）
        └─► exchange()
              └─► Project Token
                    scope: "project"
                    orgName: "acme"
                    TTL: 1h（同 Org Token，scope downscoping 不降时效）
                    └─► 可访问 /graphql/org/{orgName}/project/{任意slug}/*
                             + /graphql/runtime/org/{orgName}/project/{任意slug}/*
                        具体哪个 project 可操作 → 由 RBAC membership 决定
```

**关键语义**：Project Token 的 scope 区分的是 **API 类型**（org 管理 vs project 操作），不绑定具体某个 project。Token 本身不携带 projectSlug。

### JWT Claims 结构（统一）

```go
type PlatformClaims struct {
    UserID  string `json:"user_id"`
    OrgName string `json:"org_name"`
    Scope   string `json:"scope"` // "org" | "project" | "service_key"（预留）
    jwt.RegisteredClaims
}
```

**Issuer 统一为 `mc-platform`**（废弃 `mc-developer` 和 `mc-enduser`）

### 中间件路由逻辑

```
请求进入
  ├─ Header: X-Internal-Token  → chi_internal_token 中间件（BFF 内部调用，保留不动）
  ├─ Header: Authorization: Bearer <JWT>
  │    └─ 验证签名 + issuer == "mc-platform"
  │         ├─ scope == "org"     → 允许访问 /graphql/org/{orgName}/* 路由组
  │         │                       拒绝访问 /graphql/org/{orgName}/project/* 路由组
  │         ├─ scope == "project" → 允许访问 /graphql/org/{orgName}/project/* 路由组
  │         │                       拒绝访问 /graphql/org/{orgName}/ org 管理路由组
  │         └─ scope 不匹配路由   → 403 Forbidden
  └─ 无 token → 401 Unauthorized
```

注意：`scope=project` 放行**所有** project 路由，RBAC 在 resolver 层决定用户能操作哪些具体 project。

### 登录入口

| 入口 | URL | 适用用户 | 返回 |
|------|-----|----------|------|
| 平台管理员 | `POST /api/auth/login` | org 成员 | Org Token + refresh token |
| 端用户 | `POST /api/end-user/{orgSlug}/auth/login` | 端用户（orgSlug 做租户隔离） | Org Token + refresh token |

两个入口返回**完全相同格式**的 token，区别仅在：
- 端用户入口 URL 携带 `orgSlug`，后端不需要用户填写 org（降低登录摩擦）
- 实际可访问的 project 范围由 RBAC membership 决定，不是 token 本身决定

### Token Exchange

```
POST /api/auth/exchange
Authorization: Bearer <Org Token>
Body: {}   ← 无需指定 projectSlug

Response: { "accessToken": "<Project Token>", "expiresAt": "..." }
```

后端验证：
1. Org Token 有效（issuer + 签名校验）
2. 签发同 orgName 的 scope=project Token
3. 不验证具体 project 权限（RBAC 在下游处理）

### 删除 end_user Schema 的处理

| 原 end_user Query | 处置方式 |
|------------------|---------|
| `projects` | 复用 org schema 已有 `projects` query |
| `user(id)`, `users(input)` | 迁移到 project schema，受 RBAC 控制可见性 |
| `findOne(where)`, `findMany(where)` | 迁移到 project schema（用户搜索场景） |
| `modelDatabaseCatalog(input)` | 确认与 project schema 现有接口重叠 → 直接删除 |
| `modelCatalog(input)` | 同上 → 直接删除 |
| `model(id)` | 复用 project schema 已有 `model(id)` query |

### 前端 Workspace 入口流程

```
用户访问 /end-user/{orgSlug}/login
  → POST /api/end-user/{orgSlug}/auth/login
  → 拿到 Org Token
  → 前端请求 projects 列表（复用 org GraphQL: query { projects }）
  → 展示可访问 project 列表（无权限的不显示）
  → 用户点击某个 project
  → POST /api/auth/exchange { projectSlug }
  → 拿到 Project Token
  → 跳转 /workspace/{orgSlug}/{projectSlug}
  → Workspace 页只显示：模型数据 CRUD tab（runtime GraphQL）
  → 不显示：org 管理菜单、模型结构编辑、枚举管理等设计时功能
```

---

## Project Structure（涉及改动的目录）

```
modelcraft-backend/
├── internal/domain/auth/
│   ├── issuer.go                    # 修改：废弃 IssuerEndUser，统一为 IssuerPlatform
│   ├── modelcraft_claims.go         # 修改：新增 Scope、ProjectSlug claims
│   └── jwt_signer.go                # 修改：IssueAccessToken 支持 scope 参数
├── internal/middleware/
│   ├── chi_jwt_auth.go              # 修改：统一验证逻辑，按 scope 路由
│   └── chi_internal_token.go        # 保留不动
├── internal/interfaces/http/
│   ├── routes.go                    # 修改：删除 end-user 路由组，新增 exchange 端点
│   └── handlers/
│       ├── auth/                    # 修改：exchange handler
│       └── enduser/                 # 修改：auth_handler.go 仅保留登录逻辑，废弃 select-project
├── internal/interfaces/graphql/
│   └── enduser/                     # 删除：整个目录
├── internal/app/enduser/
│   └── end_user_auth_service.go     # 修改：删除 SelectProject，保留 Login
├── api/graph/end_user/              # 删除：整个目录
├── api/openapi/end-user-auth.yaml   # 修改：删除 select-project 端点，保留 login/logout/refresh
└── api/openapi/openapi.yaml         # 修改：新增 exchange 端点

modelcraft-front/
├── app/end-user/[orgSlug]/login/    # 新增：端用户登录页
├── app/workspace/[orgSlug]/[projectSlug]/  # 新增：workspace 主页（CRUD tab）
└── bff/                             # 修改：添加 exchange API 调用、workspace BFF 路由
```

---

## Code Style

```go
// Token scope 常量
const (
    TokenScopeOrg        = "org"
    TokenScopeProject    = "project"
    TokenScopeServiceKey = "service_key" // 预留，本期不实现
)

// Claims 示例
claims := &PlatformClaims{
    UserID:      userID,
    OrgName:     orgName,
    Scope:       TokenScopeProject,
    ProjectSlug: projectSlug,
    RegisteredClaims: jwt.RegisteredClaims{
        Issuer:    IssuerPlatform,
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
    },
}

// 中间件 scope 校验示例
func RequireProjectScope(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims := ClaimsFromCtx(r.Context())
        if claims.Scope != TokenScopeProject {
            render.Status(r, http.StatusForbidden)
            render.JSON(w, r, ErrInsufficientScope)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## Testing Strategy

- **单元测试**：`internal/domain/auth/` 的 JWT sign/verify 逻辑，scope 校验中间件
- **集成测试**：`tests/design/` 中的 auth 流程（login → exchange → GraphQL query）
- **BDD 验收测试**：`tests-bdd/` 中补充以下场景：
  - 端用户登录 → 获取 project 列表 → exchange → 访问 workspace
  - Org Token 无法访问 project-scoped 端点（scope 限制验证）
  - Project Token 无法访问 org-scoped 端点（scope 限制验证）
  - 无效 token 返回 401

---

## Boundaries

**Always:**
- 修改 `.graphql` 文件后立即运行 `just generate-gql`
- 修改 `.yaml` 文件后立即运行 `just generate-oapi`
- 提交前运行 `just lint`，失败则 `just lint-fix`
- Token 验证失败返回标准错误格式（`errors` 数组 + `requestId`）

**Ask first:**
- 是否需要数据库迁移（refresh_tokens 表的 issuer 字段变更）
- 前端 BFF 路由是否需要同步修改（前后端协调）
- 是否需要兼容旧 token（渐进迁移 vs 强制切换）

**Never:**
- 直接编辑 `internal/interfaces/graphql/generated/`（自动生成目录）
- 运行 `just clean-gql`（会删除已实现的 resolver）
- 绕过 pre-commit hook（`git commit --no-verify`）
- 删除 `X-Internal-Token` 中间件（BFF 内部调用依赖）

---

## Success Criteria

- [ ] 端用户通过 `/api/end-user/{orgSlug}/auth/login` 登录，拿到 scope=org 的 JWT，issuer=mc-platform
- [ ] 平台管理员通过 `/api/auth/login` 登录，拿到同格式的 scope=org JWT
- [ ] 两种用户的 token 经过同一个中间件验证逻辑，无分叉
- [ ] `POST /api/auth/exchange { projectSlug }` 返回 scope=project 的 JWT，TTL 与 org token 相同（1h）
- [ ] scope=org 的 token 无法访问 `/graphql/org/{orgName}/project/{slug}/` 端点（返回 403）
- [ ] scope=project 的 token 无法访问 `/graphql/org/{orgName}/` 端点（返回 403）
- [ ] `api/graph/end_user/` 目录及对应 handler/resolver 已删除，`just build` 通过
- [ ] 前端 `/end-user/{orgSlug}/login` 页面可正常登录并展示 project 列表
- [ ] 点击 project 后完成 exchange，跳转进入 `/workspace/{orgSlug}/{projectSlug}`
- [ ] workspace 页展示模型数据 CRUD tab，复用现有 runtime GraphQL 端点
- [ ] `just lint` 通过，`just build` 通过，现有 BDD 测试不回归

---

## Open Questions

1. **refresh_tokens 表**：现有表是否有 `issuer` 字段区分 developer/enduser？迁移时是否需要使旧 refresh token 失效？
2. **前端 token 存储**：workspace 场景下 Project Token 存 cookie 还是 localStorage？httpOnly cookie 更安全但 exchange 流程需要服务端参与。
3. **兼容期**：是否需要一段时间内同时支持旧 `mc-enduser` issuer？还是直接硬切？
4. **前端路由保护**：`/workspace` 路由需要验证 Project Token scope，BFF 层如何实现？

---

## 本期不做（明确边界）

| 功能 | 原因 |
|------|------|
| Service Key（应用级 token） | 架构预留，`scope=service_key` 占位，下期实现 |
| 功能权限 RBAC | project admin vs 数据人员的 tab 差异，依赖功能权限体系，下期 |
| 行级数据隔离（RLS） | 独立子系统，与 token 统一无直接依赖 |
| end-user 自注册流程 | 现有 `/api/end-user/auth/register` 是否保留待评估 |
