# 后端认证方案设计（goth + private_{projectSlug}）

> **所属模块**：终端用户认证（End-User Auth）
> **父文档**：[00-end-user-auth.md](./00-end-user-auth.md)
> **优先级**：P0

---

## 用户故事

> 作为平台，我需要一套独立于开发者认证的轻量终端用户认证系统，使用 goth 在 Go Backend 实现，数据存储在 `private_{projectSlug}` 数据库中，通过 Project 关联的 DatabaseCluster 动态路由连接。

---

## 整体架构

```
终端用户
    │  POST /api/bff/end-user/auth/login
    ▼
BFF (Next.js) — 轻量代理，不处理业务逻辑
    │  内网 POST Go_Backend/internal/end-user/auth/login
    ▼
Go Backend (goth)
    ├── 查 mc_meta: Project(orgName, projectSlug) → ClusterID → DatabaseCluster
    ├── 通过 Cluster 连接信息连到对应 MySQL 实例
    ├── USE private_{projectSlug}
    ├── 验证 users 表（bcrypt）
    ├── 检查 is_forbidden
    ├── 生成 refresh token → 存 accounts 表
    └── 返回 { userId, refreshToken, expiresAt }
    ▼
BFF
    ├── 自签 end-user access token（jose, 1h）
    │     { role: "end_user", sub: userId, orgName, projectSlug }
    ├── 写 HttpOnly Cookie: end_user_refresh_token（7d）
    └── 返回 { accessToken, expiresIn: 3600 }
```

**与开发者认证的对称关系：**

| | 开发者 | 终端用户 |
|---|---|---|
| 凭证验证 | Go Backend（mc_meta） | Go Backend（private_{projectSlug}） |
| Refresh Token 存储 | mc_meta | private_{projectSlug}.accounts |
| BFF Cookie Key | `refresh_token` | `end_user_refresh_token` |
| Access Token payload | `{ user_id }` | `{ role: "end_user", sub, orgName, projectSlug }` |
| BFF 签名工具 | `jwt-utils.ts` | `jwt-utils.ts`（复用） |

---

## 数据库设计（private_{projectSlug}）

### 命名规则

```
数据库名：private_{projectSlug}
示例：private_crm、private_erp
```

> **重要**：`private_{projectSlug}` 库与 Project 关联的 DatabaseCluster **在同一 MySQL 实例上**，通过 Cluster 的连接信息路由，不依赖独立配置。

### 连接路由机制

```
请求：orgName=acme, projectSlug=crm
    │
    ▼ Go Backend
1. 查 mc_meta.projects WHERE org_name='acme' AND slug='crm' → cluster_id
2. 查 mc_meta.database_clusters WHERE id=cluster_id
   → { host, port, username, password.Decrypt() }
3. 建立连接：mysql://user:pass@host:port/
4. USE private_crm
5. 操作 users / accounts
```

**前置条件**：Project 必须已关联 DatabaseCluster，否则终端用户功能不可用（返回 503）。

### 库初始化时机

`private_{projectSlug}` 库在以下时机自动创建并迁移：
- 开发者首次在该 Project 下创建终端用户账号时
- Go Backend 检测库不存在时，自动执行建库 + 建表 DDL

### `users` 表 — 终端用户身份

```sql
CREATE TABLE IF NOT EXISTS users (
    id           VARCHAR(36)  NOT NULL PRIMARY KEY,  -- UUID
    username     VARCHAR(64)  NOT NULL,
    password     VARCHAR(255) NOT NULL,              -- bcrypt hash, cost=12
    is_forbidden TINYINT(1)   NOT NULL DEFAULT 0,    -- 0=active, 1=disabled
    created_by   VARCHAR(36)  NOT NULL,              -- 创建者开发者 user_id（mc_meta）
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### `accounts` 表 — 会话 Token

```sql
CREATE TABLE IF NOT EXISTS accounts (
    id                 VARCHAR(36)  NOT NULL PRIMARY KEY,  -- UUID
    user_id            VARCHAR(36)  NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,              -- sha256(token)，查询索引
    expires_at         DATETIME     NOT NULL,
    revoked            TINYINT(1)   NOT NULL DEFAULT 0,
    created_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    UNIQUE KEY uq_token_hash (refresh_token_hash),
    CONSTRAINT fk_accounts_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**设计说明：**

| 字段 | 理由 |
|------|------|
| 库名无 org_name / project_slug 字段 | 已通过库名隔离，表内无需重复 |
| `refresh_token_hash` | opaque token 不明文存储，sha256 后查询 |
| `created_by` in users | 记录哪位开发者创建，供审计 |
| `IF NOT EXISTS` DDL | 支持首次访问自动初始化 |

---

## Go Backend 接口

所有接口路径以 `/internal/` 开头，仅 BFF 内网可访问，通过 `X-Internal-Token` Header 鉴权。

### POST /internal/end-user/auth/login

**请求体**：
```json
{ "orgName": "acme", "projectSlug": "crm", "username": "alice", "password": "Abc12345" }
```

**实现流程**：
```
1. 通过 orgName + projectSlug 路由到 private_crm 连接
   → Project 未关联 Cluster → 503 SERVICE_UNAVAILABLE
2. 确保 private_crm 库和表已初始化（auto-migrate）
3. 查 users WHERE username=? → 未找到 → 401 INVALID_CREDENTIALS
4. bcrypt.CompareHashAndPassword → 不匹配 → 401 INVALID_CREDENTIALS
5. is_forbidden=1 → 403 ACCOUNT_DISABLED
6. 生成 opaque refresh token（crypto/rand 32 bytes, base64url）
7. INSERT accounts { user_id, sha256(token), expires_at: +7d }
8. 返回 { userId, refreshToken, expiresAt }
```

**错误响应**：

| 状态码 | code | 场景 |
|--------|------|------|
| 401 | `INVALID_CREDENTIALS` | 凭证错误（含不存在） |
| 403 | `ACCOUNT_DISABLED` | 账号被禁用 |
| 503 | `CLUSTER_NOT_CONFIGURED` | Project 未关联 Cluster |

---

### POST /internal/end-user/auth/register

**请求体**：
```json
{ "orgName": "acme", "projectSlug": "crm", "username": "alice", "password": "Abc12345" }
```

**实现流程**：
```
1. 通过 orgName + projectSlug 路由到 private_crm 连接
   → Project 未关联 Cluster → 503 CLUSTER_NOT_CONFIGURED
2. 确保 private_crm 库和表已初始化（auto-migrate）
3. 校验用户名格式 ^[a-zA-Z0-9_-]{3,64}$，密码强度（min 8, 含字母+数字）
   → 不符合 → 400 PARAM_INVALID
4. bcrypt.GenerateFromPassword(password, cost=12)
5. INSERT users（唯一索引冲突 → 409 CONFLICT）
6. 生成 opaque refresh token（crypto/rand 32 bytes, base64url）
7. INSERT accounts { user_id, sha256(token), expires_at: +7d }
8. 返回 { userId, refreshToken, expiresAt }
```

**错误响应**：

| 状态码 | code | 场景 |
|--------|------|------|
| 400 | `PARAM_INVALID` | 用户名/密码格式不合规 |
| 409 | `CONFLICT` | 用户名已存在 |
| 503 | `CLUSTER_NOT_CONFIGURED` | Project 未关联 Cluster |

> **注意**：注册成功后行为与 login 相同（返回 refreshToken），BFF 自动签发 access token 并写 Cookie，终端用户注册即登录。

---

### GET /internal/end-user/auth/me

通过 `X-End-User-Id`、`X-Org-Name`、`X-Project-Slug` Header 传递已验证的用户上下文（BFF 验证 JWT 后注入），无需 refreshToken。

**BFF 调用方式**：
```
GET /internal/end-user/auth/me
Headers:
  X-Internal-Token: <secret>
  X-End-User-Id: <userId>
  X-Org-Name: acme
  X-Project-Slug: crm
```

**实现流程**：
```
1. 读取 X-End-User-Id / X-Org-Name / X-Project-Slug Header
2. 通过 orgName + projectSlug 路由到 private_crm 连接
3. 查 users WHERE id=? → 未找到 → 404
4. is_forbidden=1 → 403 ACCOUNT_DISABLED
5. 返回 { id, username, createdAt }
```

**成功响应（200）**：
```json
{ "id": "550e8400-...", "username": "alice", "createdAt": "2026-04-10T08:00:00Z" }
```

**错误响应**：

| 状态码 | code | 场景 |
|--------|------|------|
| 403 | `ACCOUNT_DISABLED` | 账号已被禁用（用于数据管理页右上角展示时校验） |
| 404 | `NOT_FOUND` | 用户不存在（异常情况） |
| 503 | `CLUSTER_NOT_CONFIGURED` | Project 未关联 Cluster |

---

### POST /internal/end-user/auth/refresh

```
1. sha256(refreshToken) → 查 accounts
2. 未找到 / revoked / 已过期 → 401
3. 旧 account 标记 revoked=1
4. 生成新 token，INSERT 新 account 记录（token rotation）
5. 返回 { userId, refreshToken: newToken, expiresAt }
```

---

### POST /internal/end-user/auth/logout

```
1. sha256(refreshToken) → 查 accounts
2. 标记 revoked=1
3. 返回 204
```

---

### POST Org GraphQL end-user management API — 创建终端用户

```json
{ "orgName": "acme", "projectSlug": "crm", "username": "alice", "password": "Abc12345", "createdBy": "dev-uuid" }
```

```
1. 路由到 private_crm，auto-migrate
2. 校验用户名格式 ^[a-zA-Z0-9_-]{3,64}$
3. bcrypt.GenerateFromPassword(password, cost=12)
4. INSERT users（唯一索引冲突 → 409 CONFLICT）
5. 返回 201 { id, username, createdAt }
```

### GET Org GraphQL listEndUsers query

返回分页列表：`{ total, items: [{ id, username, isForbidden, createdAt }] }`

### PATCH Org GraphQL mutation updateEndUserStatus

```json
{ "isForbidden": true }
```

更新 `users.is_forbidden`。禁用后 access token 1h 内自然过期（MVP 可接受）。

### DELETE Org GraphQL mutation deleteEndUser

物理删除 `users` 记录，同时 revoke 所有关联 `accounts`。

---

## BFF 层（轻量代理，对称复用现有模式）

**新增文件**（完全对称开发者 auth）：

| 新文件 | 对应现有文件 |
|--------|------------|
| `src/bff/end-user/end-user-go-client.ts` | `bff/auth/go-auth-client.ts` |
| `src/bff/end-user/end-user-cookie-utils.ts` | `bff/auth/cookie-utils.ts`（key 改为 `end_user_refresh_token`，path 绑定 Project） |
| `src/bff/end-user/end-user-jwt-utils.ts` | `bff/auth/jwt-utils.ts`（issuer 改为 `modelcraft-end-user`） |
| `src/bff/end-user/end-user-auth-client.ts` | `bff/auth/auth-client.ts`（client-side 工具函数） |
| `src/app/api/bff/end-user/auth/login/route.ts` | `app/api/bff/auth/login/route.ts` |
| `src/app/api/bff/end-user/auth/register/route.ts` | `app/api/bff/auth/register/route.ts` |
| `src/app/api/bff/end-user/auth/refresh/route.ts` | `app/api/bff/auth/refresh/route.ts` |
| `src/app/api/bff/end-user/auth/logout/route.ts` | `app/api/bff/auth/logout/route.ts` |
| `src/app/api/bff/end-user/auth/me/route.ts` | —（新增，用于获取终端用户信息） |

---

## middleware 扩展

```typescript
const END_USER_COOKIE        = 'end_user_refresh_token'
const END_USER_DATA_RE       = /^\/org\/[^/]+\/project\/[^/]+\/data/
const END_USER_LOGIN_RE      = /^\/org\/[^/]+\/project\/[^/]+\/user\/login/

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  if (pathname.startsWith('/api/')) return NextResponse.next()

  // 终端用户登录页 → 放行
  if (END_USER_LOGIN_RE.test(pathname)) return NextResponse.next()

  // 终端用户数据页 → 检查 end_user_refresh_token
  if (END_USER_DATA_RE.test(pathname)) {
    if (!request.cookies.has(END_USER_COOKIE)) {
      const base = pathname.match(/^(\/org\/[^/]+\/project\/[^/]+)/)?.[1] ?? ''
      const url  = new URL(`${base}/user/login`, request.url)
      url.searchParams.set('redirect', pathname)
      return NextResponse.redirect(url)
    }
    return NextResponse.next()
  }

  // 开发者路由 → 现有逻辑不变
  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) return NextResponse.next()
  if (!request.cookies.has('refresh_token')) {
    const url = new URL('/login', request.url)
    url.searchParams.set('returnUrl', pathname)
    return NextResponse.redirect(url)
  }
  return NextResponse.next()
}
```

---

## 验收标准

| # | 场景 | 预期结果 |
|---|------|----------|
| AC-1 | Project 未关联 Cluster 时尝试登录 | 503 CLUSTER_NOT_CONFIGURED |
| AC-2 | 首次创建用户时 private_crm 库不存在 | 自动建库建表，201 成功 |
| AC-3 | 正确凭证登录 | 200，accessToken，Set-Cookie: end_user_refresh_token |
| AC-4 | 错误密码登录 | 401 INVALID_CREDENTIALS |
| AC-5 | 被禁用账号登录 | 403 ACCOUNT_DISABLED |
| AC-6 | 无 cookie 访问 /data | middleware redirect 到 /user/login |
| AC-7 | 持有开发者 cookie 访问 /data | middleware redirect 到 /user/login |
| AC-8 | refresh token rotation | 旧 token revoked，新 token 有效 |
| AC-9 | logout 后旧 token 再用 | 401，token 已 revoked |
| AC-10 | 两个 Project 同名用户 | 分属不同库（private_crm / private_erp），完全隔离 |
| AC-11 | 删除用户后该用户所有 accounts revoked | 旧 refresh token 失效 |
