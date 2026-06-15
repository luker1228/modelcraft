# 1. 登录与认证

> 代码位置：`internal/domain/auth/` `internal/domain/user/` `internal/domain/enduser/`

## 概述

ModelCraft 自建用户体系，支持两类用户角色（同一张 `users` 表）：

| 角色 | 注册方式 | 典型入口 | 说明 |
|------|---------|---------|------|
| **管理员（Admin）** | 自注册（手机号+密码） | `/api/tenant/auth/register` | 注册后自动创建个人 Org，拥有 Org 的完整管理权 |
| **普通用户（EndUser）** | 由管理员创建 | 管理员控制台 | 不能自注册；只能被管理员在 Org 内创建并分配角色 |

**核心原则：一个用户可以既是管理员又是普通用户。**  
`user_orgs.is_admin = 1` 代表该用户在当前 Org 中是管理员；普通用户在其有权访问的项目里使用数据。

---

## 统一用户模型

两类角色共享同一张 `users` 表，通过 `user_orgs` 的 `is_admin` 字段区分权限级别。

```
users
├── id             UUID，全局主键
├── name           用户名（管理员：3-32字符，须以字母/下划线/连字符开头；EndUser：3-64字符）
├── phone          手机号（仅管理员注册时使用；EndUser 为空）
├── password_hash  bcrypt 哈希
├── display_name   UI 显示名（可选）
├── created_at / updated_at / deleted_at / delete_token

user_orgs
├── user_id        → users.id
├── org_name       → organizations.name
├── is_admin       0 = 普通用户  1 = 管理员
├── status         active | suspended
└── 唯一约束：uk_user_orgs_user（user_id, delete_token）
            → 每个用户只属于一个 Org
```

---

## 认证流程

### 管理员登录（Tenant）

```
POST /api/tenant/auth/login
  Identifier（用户名 or 手机号）+ Password
        │
        ▼
  TokenService.Login()
        │  GetByUsernameGlobal / GetByPhoneGlobal
        │  users JOIN user_orgs → 取 OrgName + IsAdmin
        │  bcrypt.Verify
        │
        ▼
  IssueAccessToken(userID, orgName, isAdmin=true)
  + RefreshToken → httpOnly Cookie
```

### 普通用户登录（EndUser）

```
POST /api/end-user/auth/login
  OrgName（可选）+ Identifier + IdentifierType + Password
        │
        ▼
  TokenService.LoginEndUser()
        │  GetByUsernameGlobal / GetByPhoneGlobal（全局唯一，无 orgName scope）
        │  bcrypt.Verify
        │  检查 IsActive（是否被禁用）
        │
        ▼
  IssueAccessToken(userID, orgName, isAdmin=false)
  + RefreshToken → httpOnly Cookie
```

### CLI 登录（命令行工具）

```
POST /api/cli/end-user/auth/login
  与 EndUser 登录相同逻辑，但 RefreshToken 返回在 Body 中而非 Cookie
  响应额外包含 projects（用户可访问的项目列表）
```

### CLI PAT（Personal Access Token）认证

PAT 是 CLI 工具的免密认证方式，用于长期脚本或 Agent 场景。

```
管理员在控制台 / CLI 创建 PAT
  GraphQL: CreateEndUserAPIToken(name, expiresAt?)
        │
        ▼
  生成 mc_pat_<随机串>（明文仅展示一次）
  存储 SHA-256(raw) → end_user_api_tokens.token_hash
        │
        ▼
  CLI 使用 PAT 调用：
    Authorization: Bearer mc_pat_<token>
        │
        ▼
  ChiPATAuthMiddleware（不走 JWT 中间件）
    ├── 识别 Bearer mc_pat_ 前缀
    ├── ValidateToken() → FindByHash → 校验 IsValid()
    ├── 异步更新 last_used_at
    └── 注入 context：UserID, OrgName, UserType=end_user
        │
        ▼
  GET /api/tenant/auth/whoami
    ├── 返回 userId, orgName, isAdmin, projects
    └── CLI 凭此完成身份识别（无需输入账号密码）
```

**PAT 约束：**
- 每个 EndUser 最多 20 个活跃 PAT
- 支持命名（Name 在同用户 Org 内唯一）
- 支持过期时间（`expiresAt: null` = 永不过期）
- 软删除（Revoke）后可同名重建
- `token_hash` 唯一索引，O(1) 验证

---

## Token 设计

所有角色签发相同格式的 ES256 JWT（`PlatformClaims`）：

```go
PlatformClaims {
    user_id  string  // users.id
    org_name string  // 所属 Org
    is_admin bool    // 是否为 Org 管理员
    key      string  // APISIX Consumer key = "mcuser"
    aud      "platform"
    iss      "modelcraft"
    exp      // 1 小时
}
```

Refresh Token 为 opaque token，存储于 `refresh_tokens` 表，支持：
- 轮换（rotation）：每次 refresh 吊销旧 token，签发新 token
- 盗用检测（reuse detection）：已吊销 token 重放 → 撤销该用户所有 session

---

## 管理员视角 vs 普通用户视角

| 维度 | 管理员视角 | 普通用户视角 |
|------|-----------|------------|
| 路由前缀 | `/api/tenant/auth/*` 、GraphQL `/graphql/org/{orgName}/` | `/api/end-user/auth/*`、CLI `/api/cli/end-user/auth/*` |
| JWT `is_admin` | `true` | `false` |
| 可做的事 | 管理 Org/Project/Model/User；创建普通用户 | 在被授权的 Project 内读写数据 |
| 注册方式 | 自注册（手机号+密码） → 自动创建 Org | 由管理员创建，不可自注册 |

---

## 关键代码位置

| 组件 | 路径 |
|------|------|
| JWT Claims 定义 | `internal/domain/auth/platform_claims.go` |
| JWT 签发 / 验证 | `internal/domain/auth/jwt_signer.go` |
| RefreshToken 实体 | `internal/domain/auth/refresh_token.go` |
| 管理员用户实体 | `internal/domain/user/user.go` |
| 普通用户实体 | `internal/domain/enduser/end_user.go` |
| bcrypt 值对象 | `internal/domain/enduser/hashed_password.go` |
| PAT 实体 | `internal/domain/enduser/api_token.go` |
| Token 业务逻辑（统一） | `internal/app/auth/token_service.go` |
| EndUser 认证方法 | `internal/app/auth/token_service_enduser.go` |
| PAT CRUD 服务 | `internal/app/enduser/api_token_service.go` |
| PAT 中间件 | `internal/middleware/chi_pat_auth.go` |
| 管理员 handler | `internal/interfaces/http/handlers/auth/handler.go` |
| EndUser / CLI handler | `internal/interfaces/http/handlers/enduser/auth_handler.go` |
| PAT GraphQL resolver | `internal/interfaces/graphql/org/end_user_api_token.resolvers.go` |
| 数据库 Schema（用户） | `db/schema/mysql/06_users.sql` |
| 数据库 Schema（PAT） | `db/schema/mysql/17_end_user_api_tokens.sql` |
