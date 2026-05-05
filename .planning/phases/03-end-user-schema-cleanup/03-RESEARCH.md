# 阶段 3：end_user Schema 清理 - Research

**研究日期**：2026-05-05  
**领域**：GraphQL Schema 清理 / gqlgen 代码生成 / 路由注册  
**置信度**：HIGH（基于代码库直接读取验证）

---

## 摘要

end_user GraphQL schema 是一套独立的"影子 schema"，原本为端用户提供项目/用户/模型的只读访问。经过阶段 1 完成后，端用户 token 已统一为 ES256 + mc-platform issuer，本阶段的工作是**彻底删除**这套 schema 的所有组成部分。

主要发现：
1. end_user schema 由 6 个 `.graphql` 文件组成，对应 9 条 query
2. 生成代码位于 `internal/interfaces/graphql/enduser/generated/`，手写 resolver 共 11 个文件
3. 路由注册在后端 `routes.go`（`SetupEndUserGraphQLRoutesOnChi`）及 gateway `main.go`（2 条路由）
4. gqlgen 配置在 `gqlgen.end_user.yml`，justfile 的 `generate-gql` recipe 会调用它
5. **meta_user 路由（`/graphql/runtime/org/{orgName}/meta/user`）同属 end_user schema 注册，也需一并删除**
6. gateway 中存在已标注为 Deprecated 的 `EndUserClaims`、`VerifyEndUserAccessToken`、`endUserJWTSecret` 字段，阶段 3 是清理它们的时机

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SCHEMA-01 | 删除 `api/graph/end_user/` 目录及所有关联的 handler / resolver / 路由注册 | 已完整清查所有相关文件路径 |
| SCHEMA-02 | 核查原 end_user schema 中存活的 query，确认已迁移至 org/project schema 或可直接删除 | 6 条 query 逐条处置方式已核实 |
</phase_requirements>

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| end_user GraphQL schema 删除 | Backend (Go) | — | schema 文件、生成代码、resolver 全在后端 |
| gqlgen 配置删除 | Backend (Go) | — | `gqlgen.end_user.yml` 是后端生成配置 |
| 后端路由注销 | Backend (Go) `routes.go` | — | `SetupEndUserGraphQLRoutesOnChi` 注册在 chi |
| Gateway 路由删除 | Gateway (Go) `main.go` | Gateway `proxy/handler.go` | 2 条 `/graphql/end-user/*` 路由在 gateway |
| Gateway 废弃代码清理 | Gateway (Go) `auth/service.go` | `config/config.go` | Deprecated HMAC 字段在阶段 3 标记清理 |

---

## 1. end_user Schema 相关文件完整清单

### 1.1 GraphQL Schema 文件（6 个）

```
api/graph/end_user/schema/
├── base.graphql          # 基础类型 + Error 接口 + Query root
├── catalog.graphql       # modelDatabaseCatalog, modelCatalog 查询
├── meta_user.graphql     # me, findOne, findMany（Tuser 系统用户查询）
├── model.graphql         # model(id) 查询
├── project.graphql       # projects 查询
└── user.graphql          # user(id), users(input) 查询
```

### 1.2 gqlgen 配置文件（1 个）

```
gqlgen.end_user.yml       # 告知 gqlgen 使用哪个 schema、生成到哪个路径
```

配置内容：
- schema source: `api/graph/end_user/schema/*.graphql`
- 生成目标: `internal/interfaces/graphql/enduser/generated/`
- resolver 目录: `internal/interfaces/graphql/enduser/`
- 包名: `endusergraphql`

### 1.3 生成代码目录（全部由 gqlgen 自动生成，可安全删除）

```
internal/interfaces/graphql/enduser/generated/
├── generated.go          # 自动生成的 ExecutableSchema、Resolvers 接口等
└── model_gen.go          # 自动生成的 Go 类型（Tuser, RuntimeUser, ModelLite 等）
```

### 1.4 手写 Resolver 文件（11 个，需手动删除）

```
internal/interfaces/graphql/enduser/
├── base.resolvers.go          # queryResolver 基础结构
├── catalog.resolvers.go       # ModelDatabaseCatalog, ModelCatalog 实现
├── catalog_error_mapper.go    # catalog 错误映射助手
├── handler.go                 # EndUserGraphQLHandler, EndUserPlaygroundHandler
├── meta_user.resolvers.go     # Me, FindOne, FindMany 实现
├── meta_user_helpers.go       # meta_user 助手函数（mapMetaUserError, toTuser 等）
├── model.resolvers.go         # Model(id) 实现
├── project.resolvers.go       # Projects 实现
├── resolver.go                # Resolver 结构体定义（依赖 modeldesign, enduser app 服务）
├── user.resolvers.go          # User(id), Users(input) 实现
└── user_error_mapper.go       # user 错误映射助手
└── user_mapper.go             # user -> GraphQL 类型转换
```

> **注意**：`resolver.go` 中的 `Resolver` 结构体依赖 `*modeldesign.ModelDesignAppService`、`*enduser.EndUserManagementAppService`、`*enduser.MetaUserAppService`。这些**服务本身不需要删除**，只是删除 resolver 对它们的引用。

---

## 2. 路由注册清查

### 2.1 Backend（`internal/interfaces/http/routes.go`）

函数 `SetupEndUserGraphQLRoutesOnChi`（第 668-705 行）注册了 **2 条路由组**：

| 路由模式 | handler | 说明 |
|---------|---------|------|
| `/graphql/end-user/org/{orgName}/project/{projectSlug}` | `EndUserGraphQLHandler` + `EndUserPlaygroundHandler` | 主 end-user GraphQL 端点 |
| `/graphql/runtime/org/{orgName}/meta/user` | `EndUserGraphQLHandler`（metaUserResolver）+ `EndUserPlaygroundHandler` | meta/user 系统用户查询端点 |

此函数在 `chi_setup.go` 中被调用（第 119-121 行）：
```go
if cfg.DesignHandlers != nil {
    SetupEndUserGraphQLRoutesOnChi(r, cfg.DesignHandlers, cfg.Config)
}
```

**需删除**：
- `routes.go` 中整个 `SetupEndUserGraphQLRoutesOnChi` 函数
- `chi_setup.go` 中对应的调用块（第 118-121 行）
- `routes.go` 中 `import endusergraphql "modelcraft/internal/interfaces/graphql/enduser"` 导入语句
- `routes.go` 中 `DesignHandlers` struct 内如有仅用于 end-user GraphQL 的字段（暂未发现，可重新检查）

> **关键发现**：`chi_setup.go` 的架构注释（第 50 行）还包含 `/graphql/end-user/*` 的描述需更新。

**不需删除**（阶段 3 范围外）：
- `SetupEndUserRoutesOnChi`（`/internal/end-users`、`/internal/end-user/data`）：这些是 BFF 内部 HTTP 路由，不是 GraphQL schema 路由，本阶段不删
- `chi_setup.go` 中 `cfg.DesignHandlers.EndUserAuthHandler` 引用（OpenAPI 端点，保留）

### 2.2 Gateway（`modelcraft-gateway/cmd/gateway/main.go`）

`main.go` 第 111-113 行注册了 end-user GraphQL 的代理路由：

```go
// End-User GraphQL — end-user HMAC JWT required.
r.Post("/graphql/end-user/org/{orgName}/project/{projectSlug}", proxyHandler.EndUserGraphQLHandler)
r.Post("/graphql/end-user/org/{orgName}/project/{projectSlug}/", proxyHandler.EndUserGraphQLHandler)
```

使用的 handler：`EndUserGraphQLHandler`（`internal/proxy/handler.go` 第 116-146 行）

**需删除**：
- gateway `main.go` 中这 2 条路由注册
- `internal/proxy/handler.go` 中 `EndUserGraphQLHandler` 函数（第 113-146 行）
- gateway `main.go` 注释中 "End-User GraphQL" 块
- gateway `main.go` 第 49 行 `cfg.EndUserJWTSecret` 传参（此参数传给了 `auth.NewService`）

**顺带清理（Deprecated 残留代码）**：
- `internal/auth/service.go`：`EndUserClaims` 结构体（第 142-146 行，标注 Deprecated）
- `internal/auth/service.go`：`VerifyEndUserAccessToken` 方法（第 151-182 行，标注 Deprecated）
- `internal/auth/service.go`：`endUserJWTSecret` 字段及相关方法（`ClearEndUserRefreshCookie`、`SetEndUserRefreshCookie`、`GetEndUserRefreshCookie` 是 cookie 管理，**不属于 schema 清理范围**，保留）
- `internal/config/config.go`：`EndUserJWTSecret` 字段（第 23 行，标注 Deprecated）
- `auth.NewService` 函数签名移除 `endUserSecret string` 参数

> **注意**：`SetEndUserRefreshCookie`、`GetEndUserRefreshCookie`、`ClearEndUserRefreshCookie` 这三个 cookie 方法依然被 `handler.go` 的 `EndUserLogin/Refresh/Logout` 用于管理端用户 refresh cookie，**不能删除**。

---

## 3. 6 条原 Query 当前状态与处置方式确认

| 原 end_user Query | 文件来源 | 当前状态（org/project schema） | 处置方式 |
|------------------|---------|-------------------------------|---------|
| `projects` | `project.graphql` | ✅ org schema 有 `projects(input)` query（`api/graph/org/schema/project.graphql` 第 243 行） | **直接删除**，前端改用 org schema 的 `projects` |
| `user(id)` | `user.graphql` | ⚠️ project schema 没有 `user(id)`，有 `listProjectEndUsers` | **需确认**：是否迁移到 project schema（plans/unified-token-system.md 说迁移到 project schema） |
| `users(input)` | `user.graphql` | ⚠️ 同上，project schema 有 `listProjectEndUsers` 但字段不同 | **需确认**：与 `listProjectEndUsers` 字段对齐后删除原 `users` |
| `findOne(where)` | `meta_user.graphql` | ❌ org/project schema 均无等效接口 | **meta_user 路由保留特殊处理**：`/graphql/runtime/org/{orgName}/meta/user` 路由依然需要这些查询；但与 *end_user* schema 解耦：可将 meta_user 迁至独立 schema 或直接保留在 project schema 的 end_user 部分 |
| `findMany(where)` | `meta_user.graphql` | ❌ 同上 | 同上 |
| `me` | `meta_user.graphql` | ❌ 同上 | 同上 |
| `modelDatabaseCatalog(input)` | `catalog.graphql` | ✅ project schema 没有直接等价物，但 `data_handler.go` 已有 REST `/internal/end-user/data/database-catalog` 提供同样功能 | **直接删除**（REST 路由保留） |
| `modelCatalog(input)` | `catalog.graphql` | ✅ 同上，REST `/internal/end-user/data/model-catalog` 已有 | **直接删除** |
| `model(id)` | `model.graphql` | ✅ project schema 有 `model(id, withActualSchema)` query（`api/graph/project/schema/model.graphql` 第 291 行） | **直接删除**，前端改用 project schema 的 `model(id)` |

### 重要发现：meta_user 路由的特殊性

`me`、`findOne`、`findMany` 三条查询来自 `meta_user.graphql`，服务于 `/graphql/runtime/org/{orgName}/meta/user` 路由。这条路由在 `plans/unified-token-system.md` 中**没有明确提到**保留还是删除。

**当前代码实现**：
- 注册在 `SetupEndUserGraphQLRoutesOnChi` 中（第 699-704 行）
- 使用 end_user schema 的 `endusergraphql.EndUserGraphQLHandler` 和 `metaUserResolver`
- resolver 中的 `me`/`findOne`/`findMany` 调用 `appEnduser.MetaUserAppService`

**如果整个 end_user schema 删除**，这条路由也要跟着删。`MetaUserAppService` 的功能（系统用户查询）未来若要保留，需要：
1. 在 project schema 的 `end_user.graphql` 中新增 `me`/`findOne`/`findMany`，**或**
2. 为 meta/user 创建独立的 schema + gqlgen 配置，**或**
3. 直接删除该端点（如果没有外部调用方依赖）

> **建议**（需用户确认）：SCHEMA-01 阶段 3 的目标是删除 end_user schema。meta/user 路由目前只通过 BFF 内部 token 访问（`internalTokenMW`），客户端不直接调用。建议直接删除，待阶段 4 前端需要时再通过 project schema 重新定义相关接口。

---

## 4. 变更范围（文件删除/修改列表）

### 4.1 删除文件（共 20 个）

```bash
# Schema 文件（6 个）
api/graph/end_user/schema/base.graphql
api/graph/end_user/schema/catalog.graphql
api/graph/end_user/schema/meta_user.graphql
api/graph/end_user/schema/model.graphql
api/graph/end_user/schema/project.graphql
api/graph/end_user/schema/user.graphql

# gqlgen 配置（1 个）
gqlgen.end_user.yml

# 自动生成代码（2 个）
internal/interfaces/graphql/enduser/generated/generated.go
internal/interfaces/graphql/enduser/generated/model_gen.go

# 手写 Resolver（11 个）
internal/interfaces/graphql/enduser/base.resolvers.go
internal/interfaces/graphql/enduser/catalog.resolvers.go
internal/interfaces/graphql/enduser/catalog_error_mapper.go
internal/interfaces/graphql/enduser/handler.go
internal/interfaces/graphql/enduser/meta_user.resolvers.go
internal/interfaces/graphql/enduser/meta_user_helpers.go
internal/interfaces/graphql/enduser/model.resolvers.go
internal/interfaces/graphql/enduser/project.resolvers.go
internal/interfaces/graphql/enduser/resolver.go
internal/interfaces/graphql/enduser/user.resolvers.go
internal/interfaces/graphql/enduser/user_error_mapper.go
internal/interfaces/graphql/enduser/user_mapper.go
```

> 注：`internal/interfaces/graphql/enduser/generated/` 目录随文件删除后可删除空目录；`internal/interfaces/graphql/enduser/` 目录整体删除。

### 4.2 修改文件（Backend，共 2 个）

**`internal/interfaces/http/routes.go`**
- 删除 `import endusergraphql "modelcraft/internal/interfaces/graphql/enduser"`（第 30 行）
- 删除 `SetupEndUserGraphQLRoutesOnChi` 整个函数（第 664-706 行）

**`internal/interfaces/http/chi_setup.go`**
- 删除 `SetupEndUserGraphQLRoutesOnChi` 调用块（第 116-121 行）
- 更新架构注释（第 50 行），删除 `/graphql/end-user/*` 描述

### 4.3 修改文件（Gateway，共 3 个）

**`modelcraft-gateway/cmd/gateway/main.go`**
- 删除路由注册（第 111-113 行）：2 条 `/graphql/end-user/*` 路由
- 删除 `cfg.EndUserJWTSecret` 参数传入（第 49 行），并更新 `auth.NewService` 调用
- 删除相关注释块

**`modelcraft-gateway/internal/proxy/handler.go`**
- 删除 `EndUserGraphQLHandler` 函数（第 113-146 行）

**`modelcraft-gateway/internal/auth/service.go`**（Deprecated 清理）
- 删除 `endUserJWTSecret []byte` 字段（第 28 行）
- 删除 `EndUserClaims` 结构体（第 142-146 行）
- 删除 `VerifyEndUserAccessToken` 方法（第 151-182 行）
- 更新 `NewService` 函数签名，移除 `endUserSecret string` 参数

**`modelcraft-gateway/internal/config/config.go`**（Deprecated 清理）
- 删除 `EndUserJWTSecret string` 字段（第 23 行）
- 删除对应 `getEnv("JWT_SECRET", "")` 赋值（第 51 行）

### 4.4 不需修改（确认保留）

| 文件/组件 | 原因 |
|----------|------|
| `api/openapi/end-user-auth.yaml` | 端用户 REST 认证路由（login/register/refresh/logout），与 GraphQL schema 无关，保留 |
| `internal/interfaces/http/handlers/enduser/` 目录 | HTTP handler（auth、management、data），独立于 GraphQL schema，保留 |
| `internal/interfaces/http/routes.go:SetupEndUserRoutesOnChi` | `/internal/end-users` 管理路由，BFF 内部用，保留 |
| `internal/app/enduser/` 全部 app 服务 | 业务逻辑层，继续被 HTTP handler 和其他地方引用，保留 |
| `internal/interfaces/graphql/org/end_user.resolvers.go` | org schema 的 end_user 管理 resolver，与 end_user schema 无关，保留 |
| `internal/interfaces/graphql/project/end_user.resolvers.go` | project schema 的 end_user 管理 resolver，保留 |
| gateway `auth/handler.go` 中 `EndUserLogin/Refresh/Logout/Me` | 端用户认证 REST 代理，保留 |
| gateway `auth/service.go` 中 cookie 相关方法 | `SetEndUserRefreshCookie`、`GetEndUserRefreshCookie`、`ClearEndUserRefreshCookie`，保留 |

---

## 5. 清理顺序建议（防止编译断链）

### Wave 1：先断开外部引用

1. **修改 `routes.go`**：删除 `endusergraphql` import 和 `SetupEndUserGraphQLRoutesOnChi` 函数
2. **修改 `chi_setup.go`**：删除对 `SetupEndUserGraphQLRoutesOnChi` 的调用
3. **验证**：`go build ./internal/interfaces/http/...` 必须通过

> 先删引用方，再删被引用的包，避免 import cycle 或悬空引用。

### Wave 2：删除 enduser GraphQL 包

4. **删除整个 `internal/interfaces/graphql/enduser/` 目录**（包含 `generated/` 子目录）
5. **删除 `api/graph/end_user/` 目录**（schema 文件）
6. **删除 `gqlgen.end_user.yml`**
7. **验证**：`go build ./...` 必须通过（重点是 backend 整体编译）

### Wave 3：更新 justfile 和 gateway

8. **修改 justfile** 的 `generate-gql` recipe：删除 `go run ... --config gqlgen.end_user.yml` 那行
9. **修改 gateway `main.go`**：删除 2 条路由注册
10. **修改 gateway `proxy/handler.go`**：删除 `EndUserGraphQLHandler` 函数
11. **修改 gateway `auth/service.go`**：删除 Deprecated 字段和方法
12. **修改 gateway `config/config.go`**：删除 `EndUserJWTSecret` 字段
13. **验证**：`cd modelcraft-gateway && go build ./...` 必须通过

### Wave 4：验证收口

14. 运行 `just build`（backend）
15. 检查 `just lint` 通过
16. 确认 `/graphql/end-user/*` 路由返回 404（gateway 层面不再注册）

---

## 6. 风险点（删除后可能的编译错误需修复）

### 风险 1：`routes.go` 的 `DesignHandlers` 字段残留引用

`DesignHandlers` struct 中有若干 End-User 相关字段（`EndUserAuthAppService`、`EndUserMgmtAppService`、`EndUserMgmtHandler`、`EndUserDataHandler` 等），这些字段被 `SetupEndUserRoutesOnChi`（管理路由）和 `SetupProjectGraphQLRoutesOnChi`（project schema 引用 `EndUserMgmtAppService`）所使用。**删除 end_user GraphQL schema 不需要删除这些字段**，但需注意 `SetupEndUserGraphQLRoutesOnChi` 独有使用的 `MetaUserService` 创建代码（第 676-680 行）需要随函数一并删除：

```go
// 第 676-680 行（删除 SetupEndUserGraphQLRoutesOnChi 时随之删除）
metaUserDBAdapter := appEnduser.NewPrivateDBManagerAdapter(handlers.PrivateDBManager)
metaUserResolver := &endusergraphql.Resolver{
    ModelDesignService: handlers.ModelAppService,
    EndUserMgmtService: handlers.EndUserMgmtAppService,
    MetaUserService:    appEnduser.NewMetaUserAppService(metaUserDBAdapter),
}
```

### 风险 2：gateway `auth.NewService` 签名变更

删除 `EndUserJWTSecret` 参数后，`auth.NewService` 调用方（`main.go`）需同步更新，否则编译失败。**清理顺序必须保证 service.go 和 main.go 同一 Wave 修改。**

### 风险 3：`SelectProject` app 服务方法残留

`EndUserAuthAppService.SelectProjectContext` 方法目前仍在（被 `server.go` 的 `EndUserSelectProject` 引用，via OpenAPI generated handler）。这与 schema 清理无关，不在本阶段删除范围。注意不要误删。

### 风险 4：gateway build 独立验证

gateway 是独立 Go module，需单独运行 `go build ./...`，backend 构建通过不代表 gateway 也通过。

### 风险 5：`graphify-out` 目录

`internal/interfaces/graphql/graphify-out/` 是 graphify 缓存目录，**不需要处理**，不影响编译。

---

## 7. 环境可用性

| 依赖 | 说明 | 可用 |
|------|------|------|
| `just generate-gql` | 需更新以移除 end_user 配置 | ✓（justfile 已在 backend） |
| `just build` | backend 构建验证 | ✓ |
| `go build ./...` | backend/gateway 均需验证 | ✓ |
| `just lint` | lint 检查 | ✓ |

---

## 8. 假设与待确认项

| # | 假设内容 | 风险 | 建议确认 |
|---|---------|------|---------|
| A1 | `meta/user` 路由（`/graphql/runtime/org/{orgName}/meta/user`）在阶段 3 可整体删除，阶段 4 前端不依赖 | 中：若前端已有代码调用该路由会破坏 | 确认前端是否调用了 `/graphql/runtime/org/.../meta/user` |
| A2 | `user(id)` 和 `users(input)` 的功能由 project schema 的 `listProjectEndUsers` 覆盖，前端已有迁移路径 | 低：两者字段略有不同 | plans/unified-token-system.md 明确说迁移到 project schema，可按此执行 |
| A3 | gateway 的 `EndUserRefreshCookieName` cookie 配置在阶段 3 保留（因为 end-user auth REST 路由还在） | 低 | 确认保留，本阶段不改 cookie 逻辑 |

---

## 原 Query 处置方式汇总

| 原 end_user Query | 处置 | 替代接口 | 可行性 |
|------------------|------|---------|--------|
| `projects` | 直接删除 | org schema `projects(input)` | ✅ 已有现成接口 |
| `model(id)` | 直接删除 | project schema `model(id, withActualSchema)` | ✅ 已有现成接口 |
| `modelDatabaseCatalog` | 直接删除 | REST `/internal/end-user/data/database-catalog` | ✅ REST 替代 |
| `modelCatalog` | 直接删除 | REST `/internal/end-user/data/model-catalog` | ✅ REST 替代 |
| `user(id)` | 迁移到 project schema | project schema 新增 `endUser(id)` 或复用 `listProjectEndUsers` | ⚠️ 需新增字段（SCHEMA-02） |
| `users(input)` | 迁移到 project schema | project schema `listProjectEndUsers(input)` | ⚠️ 字段差异需对齐（SCHEMA-02） |
| `me` | 暂时删除（随 meta/user 路由一起） | 待阶段 4 前端需要时补充 | ⚠️ 依赖阶段 4 设计 |
| `findOne(where)` | 同上 | 同上 | ⚠️ 同上 |
| `findMany(where)` | 同上 | 同上 | ⚠️ 同上 |

---

## 来源

### PRIMARY（HIGH 置信度，代码库直接读取）

- [VERIFIED: codebase] `api/graph/end_user/schema/*.graphql` — 6 个 schema 文件内容
- [VERIFIED: codebase] `gqlgen.end_user.yml` — gqlgen 配置文件
- [VERIFIED: codebase] `internal/interfaces/graphql/enduser/` — 14 个文件清单
- [VERIFIED: codebase] `internal/interfaces/http/routes.go` — `SetupEndUserGraphQLRoutesOnChi` 函数（第 664-706 行）
- [VERIFIED: codebase] `internal/interfaces/http/chi_setup.go` — 路由注册调用（第 116-121 行）
- [VERIFIED: codebase] `modelcraft-gateway/cmd/gateway/main.go` — 路由注册（第 111-113 行）
- [VERIFIED: codebase] `modelcraft-gateway/internal/proxy/handler.go` — `EndUserGraphQLHandler`（第 113-146 行）
- [VERIFIED: codebase] `modelcraft-gateway/internal/auth/service.go` — Deprecated `EndUserClaims`、`VerifyEndUserAccessToken`
- [VERIFIED: codebase] `modelcraft-gateway/internal/config/config.go` — Deprecated `EndUserJWTSecret`
- [VERIFIED: codebase] `api/graph/org/schema/project.graphql` — org schema `projects` query 已有
- [VERIFIED: codebase] `api/graph/project/schema/model.graphql` — project schema `model(id)` query 已有
- [VERIFIED: codebase] `justfile` — `generate-gql` recipe 调用了 `gqlgen.end_user.yml`
- [VERIFIED: codebase] `.planning/phases/01-token-core-unified/01-02-SUMMARY.md` — 阶段 1 Wave 2 完成内容

---

## RESEARCH COMPLETE

**阶段**：3 - end_user Schema 清理  
**置信度**：HIGH

### 核心发现

1. **删除范围已完整清查**：Schema 文件 6 个、gqlgen 配置 1 个、生成代码 2 个、手写 resolver 11 个，共 20 个文件需删除
2. **路由注册双层**：backend `routes.go` 中 `SetupEndUserGraphQLRoutesOnChi` + gateway `main.go` 中 2 条 `/graphql/end-user/*` 路由，均需删除
3. **meta/user 路由随之删除**：`/graphql/runtime/org/{orgName}/meta/user` 注册在同一函数中，阶段 3 一并删除，`me/findOne/findMany` 三条 query 暂不迁移
4. **gateway 有 Deprecated 残留代码**：`EndUserClaims`、`VerifyEndUserAccessToken`、`EndUserJWTSecret` 字段在 service.go 中已标注 Deprecated，阶段 3 顺带清理
5. **正确清理顺序**：先删 import/调用方（Wave 1）→ 再删包本身（Wave 2）→ 最后清理 gateway（Wave 3），防止编译断链

### 文件创建

`.planning/phases/03-end-user-schema-cleanup/03-RESEARCH.md`

### 置信度评估

| 领域 | 等级 | 原因 |
|------|------|------|
| 文件清单 | HIGH | 代码库直接读取验证 |
| 路由注册 | HIGH | 代码库直接读取验证 |
| Query 处置方式 | MEDIUM | 部分（me/findOne/findMany）未在 plans/unified-token-system.md 中明确 |

### 待规划时确认

- A1：前端是否当前调用了 `/graphql/runtime/org/.../meta/user` 端点（影响是否需要保留 meta/user 路由）
- A2：`user(id)` / `users(input)` 迁移方案：新增到 project schema 的 `end_user.graphql` 还是直接删除（取决于阶段 4 前端需求）
