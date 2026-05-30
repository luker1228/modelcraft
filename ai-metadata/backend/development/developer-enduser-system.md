# Developer / EndUser 双体系说明

> 适用范围：认证、网关代理、前端 BFF 联调。本文定义 ModelCraft 当前并行的两套用户体系。

## 1. 总览

ModelCraft 有两类用户，**共享同一张 `users` 表**，共享同一套 JWT 格式（统一 issuer `mc-platform`，ES256 签名），通过 `is_admin` claim 区分身份：

1. **Developer 体系（平台管理者，`is_admin=true`）**：负责 Org 管理、Project 创建、模型设计，拥有全量权限。
2. **EndUser 体系（端用户，`is_admin=false`）**：访问被授权的 Project 内的业务数据，不感知平台管理细节。

**同一个账号可以同时具有两种身份**（`user_orgs.is_admin` 决定）。两套体系使用**完全独立的前端路由**，但共享同一个 access token 和 auth store，切换视图仅需路由跳转，无需重新认证。

---

## 2. 用户能力边界

| 能力 | Developer（平台管理者） | EndUser（端用户） |
|------|------------------------|------------------|
| 创建/删除 Org | ✅ | ❌ |
| 邀请注册 / 管理账号 | ✅ | ❌ |
| 创建 / 删除 Project | ✅ | ❌ |
| 设计模型 / 字段 / 数据库结构 | ✅ | ❌ |
| **List 可访问的 Project** | ✅ | ✅ |
| 在 Project 内操作业务数据（CRUD） | ❌（他是设计者） | ✅ |
| 查看 / 管理自己的 API Key | ❌ | ✅ |
| 查看自己的 Profile | ❌ | ✅ |

---

## 3. 路由隔离与视图切换

两套体系使用**完全独立的前端路由树**，共享同一个 access token：

| 体系 | 前端路由前缀 | 登录后跳转 | 路由 Guard |
|------|------------|-----------|-----------|
| Developer | `/org/{orgName}/...` | `/org/{orgName}/dashboard` | 要求 `is_admin=true` |
| EndUser | `/end-user/{orgName}/...` | `/end-user/{orgName}/projects` | 只要求已登录 |

**视图切换**：`is_admin=true` 的用户可以在两个视图之间自由切换，点击"切换视图"按钮做路由跳转，不发任何 API 请求，token 不变。`is_admin=false` 的用户没有切换入口，强行访问管理路由会被 guard 重定向到用户视图。

**登录跳转逻辑**：登录成功后，前端解析 JWT 中的 `is_admin`，`true` 跳管理页，`false` 跳用户页。

---

## 4. 对照表

| 维度 | Developer 体系 | EndUser 体系 |
|------|---------------|--------------|
| **身份标记** | `user_orgs.is_admin = true` | `user_orgs.is_admin = false` |
| **前端路由前缀** | `/org/{orgName}/...` | `/end-user/{orgName}/...` |
| **登录接口** | `POST /api/auth/login` 或 `POST /api/end-user/auth/login`（共用同一实现） | 同左 |
| **JWT issuer** | `mc-platform` | `mc-platform`（相同） |
| **JWT is_admin** | `true` | `false` |
| **Access Token 签名** | ES256（ECDSA P-256） | ES256（相同） |
| **Auth Store** | 前端单一 Zustand store，共用 | 同左（共用） |
| **Org 级 GraphQL 路由** | `/graphql/org/{orgName}/` | `/graphql/org/{orgName}/`（共用） |
| **Project 级 GraphQL 路由** | `/graphql/org/{orgName}/project/{projectSlug}/` | `/graphql/org/{orgName}/project/{projectSlug}/`（共用） |
| **后端识别头（Gateway 注入）** | `X-User-ID` + `X-Is-Admin: true` | `X-User-ID` + `X-Is-Admin: false` |

---

## 5. GraphQL Schema 分层

Developer 和 EndUser **共用同一套 GraphQL Schema**（`org.graphql` + `project.graphql`），通过 JWT 中的 `is_admin` 字段（Gateway 注入为 `X-Is-Admin` header）在后端实现能力边界：

```
Org Schema（共用）                      Project Schema（共用）
/graphql/org/{orgName}/                /graphql/org/{orgName}/project/{slug}/
  createProject（is_admin=true）          模型设计 / 字段管理（is_admin=true）
  deleteProject（is_admin=true）          数据库结构（is_admin=true）
  manageUsers（is_admin=true）            业务数据 CRUD（is_admin=false）
  listProject（两者均可，范围不同）         Runtime GraphQL（is_admin=false）
  me（两者均可）                           ...
  ...
```

EndUser 在共用端点上访问，后端根据 `X-Is-Admin` header 限制可操作的数据范围。

---

## 6. 请求链路

### 6.1 Developer 链路

```
Browser -> Front BFF(/api/auth/*) -> Gateway(/auth/*, /graphql/org/*) -> Backend
```

### 6.2 EndUser 链路

```
Browser -> Front BFF(/api/end-user/auth/*) 
        -> Gateway(/end-user/auth/*, /graphql/org/*)   ← 共用 GraphQL 端点
        -> Backend
```

### 6.3 登录流程（两体系共用同一后端实现）

```
POST /api/auth/login  或  POST /api/end-user/auth/login
  { identifier, identifierType, password, orgName? }

→ 200 {
    accessToken,   // PlatformClaims JWT，is_admin 已写入
    expiresIn,
    userId,
    orgName
  }
  refreshToken 存于 httpOnly cookie mc_refresh_token

前端解析 JWT is_admin：
  true  → router.push(`/org/${orgName}/dashboard`)
  false → router.push(`/end-user/${orgName}/projects`)
```

**注意**：登录响应不再返回 `projects` 列表，前端需要时单独调 GraphQL 查询。

---

## 7. 强制边界（必须遵守）

1. 前端（浏览器侧 + 前端服务侧）**必须先访问 Gateway，再转发到 Backend**。
2. 禁止前端任何业务请求直连 Backend（包括 GraphQL/REST）。
3. **Gateway 是唯一的 JWT 验签者**，使用 ES256（`JWT_PUBLIC_KEY`）验签 `mc-platform` issuer。
4. Backend design-time 端点只信任 Gateway 注入的 `X-User-ID` + `X-Is-Admin`，不接受直接 bearer token。
5. **CLI** 走 `cli -> gateway -> backend` 路径，不得直连 Backend。
6. **管理路由 `/org/...` 必须在前端 guard 层检查 `is_admin=true`**，`is_admin=false` 的用户直接重定向到用户视图。

---

## 8. 代码锚点（当前实现）

### Gateway
- 路由装配：`modelcraft-gateway/cmd/gateway/main.go`
- Token 校验（统一 mc-platform ES256）：`modelcraft-gateway/internal/auth/service.go`
- GraphQL 代理：`modelcraft-gateway/internal/proxy/handler.go`

### Frontend
- Auth Store（单一）：`modelcraft-front/src/store/authStore.ts`（或类似路径）
- 管理路由 Guard（检查 `is_admin`）：`modelcraft-front/src/app/org/[orgName]/layout.tsx`
- 登录页（管理端）：`modelcraft-front/src/app/login/`
- 登录页（用户端）：`modelcraft-front/src/app/end-user/[orgName]/login/`

### Backend
- 统一登录 Handler：`modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go`（`HandleLogin`）
- EndUser Auth Handler：`modelcraft-backend/internal/interfaces/http/handlers/enduser/auth_handler.go`（路由指向同一实现）
- JWT 签发：`modelcraft-backend/internal/domain/auth/jwt_signer.go`（`IssueAccessToken`）
- PlatformClaims 结构：`modelcraft-backend/internal/domain/auth/platform_claims.go`

---

## 9. 文档关系

- Gateway 架构细节：`./gateway-architecture.md`
- 部署联调检查项：`../deployment/README.md`
- EndUser v2 PRD：`../../prd/enduser-v2/`
- Token 统一里程碑路线图：`.planning/ROADMAP.md`
