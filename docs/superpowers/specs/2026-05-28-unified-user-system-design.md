# 统一用户体系设计

**日期**: 2026-05-28  
**状态**: 待实现

---

## 背景与目标

当前系统存在两套独立用户体系：

- **tenant 用户**（`users` 表）：管理员/开发者，全局唯一用户名，拥有平台 RBAC
- **end-user 用户**（`end_user_users` 表）：终端用户，Org 级唯一用户名，只有 Project 级数据角色

**目标**：合并为单一用户体系，简化认证逻辑，降低维护成本。

**核心约束**：
- 每个用户只能属于一个 Org
- 用户名全局唯一
- Project 级角色保留（不同项目可有不同权限）
- 普通用户不需要平台 RBAC

---

## 第一节：用户模型 & 数据库 Schema

### 表结构

```sql
-- 统一用户表（原 users 表扩展，原 end_user_users 废弃迁移）
users
  id             VARCHAR(36)     PRIMARY KEY
  username       VARCHAR(64)     NOT NULL UNIQUE  -- 全局唯一
  password_hash  VARCHAR(255)    NOT NULL
  deleted_at     BIGINT UNSIGNED NOT NULL DEFAULT 0
  delete_token   BIGINT UNSIGNED NOT NULL DEFAULT 0
  UNIQUE KEY uk_username (username, delete_token)

-- 用户 Org 绑定（原 user_organizations 演化）
user_orgs
  id         VARCHAR(36)  PRIMARY KEY
  user_id    VARCHAR(36)  NOT NULL  -- FK → users.id
  org_name   VARCHAR(36)  NOT NULL  -- FK → organizations.name
  is_admin   BOOLEAN      NOT NULL DEFAULT FALSE
  status     VARCHAR(20)  NOT NULL DEFAULT 'active'  -- active | suspended
  created_at DATETIME(3)
  updated_at DATETIME(3)
  deleted_at     BIGINT UNSIGNED NOT NULL DEFAULT 0
  delete_token   BIGINT UNSIGNED NOT NULL DEFAULT 0
  UNIQUE KEY uk_user_org (user_id, delete_token)  -- 每人只能属于一个 Org

-- Project 级数据角色（原 end_user_role_users，重命名）
project_role_users
  id           BIGINT       PRIMARY KEY AUTO_INCREMENT
  user_id      VARCHAR(36)  NOT NULL  -- FK → users.id
  org_name     VARCHAR(36)  NOT NULL
  project_slug VARCHAR(64)  NOT NULL
  role_name    VARCHAR(64)  NOT NULL  -- FK → end_user_roles.name
  created_at   DATETIME(3)
  UNIQUE KEY uk_user_project_role (user_id, org_name, project_slug, role_name)
```

### 角色体系保留情况

| 表 | 状态 | 适用用户 |
|---|---|---|
| `roles` / `user_roles` / `role_permissions` | **保留不变** | 仅 `is_admin=true` 的用户 |
| `end_user_roles` | **保留，重命名为 `project_roles`** | 所有用户（Project 级数据权限） |
| `project_role_users` | **新表**（原 `end_user_role_users` 迁移） | 所有用户 |

### 迁移策略

1. 原 `users` 表（tenant）直接保留，为每条记录补充 `user_orgs` 记录（`is_admin=true`）
2. 原 `end_user_users` 数据迁移到 `users` 表，`user_orgs.is_admin=false`
3. username 冲突时，`end_user_users.username` 加 `_{org_name}` 后缀，人工确认后修正
4. 原 `end_user_role_users` 迁移到 `project_role_users`，`user_id` 替换为新 `users.id`

---

## 第二节：认证 & Token

### 统一登录接口

两个前端入口共用同一个后端接口：

```
POST /api/auth/login
Body: { username, password }

Response: { token, refreshToken }
```

后端流程：
1. 全局查 `users` 表（by `username`）
2. 验证密码
3. 查 `user_orgs` 获取 `org_name` 和 `is_admin`
4. 签发统一 JWT

### Token（统一 ES256，废弃 HMAC-SHA256）

```json
{
  "sub": "<user_id>",
  "org_name": "<org_name>",
  "is_admin": true,
  "scope": "..."
}
```

- **废弃** end-user 的 HMAC-SHA256 token
- **统一** 使用 ES256，Gateway 统一验证
- Gateway 注入 header：`X-User-ID`、`X-Org-Name`、`X-Is-Admin`

### 注册流程（仅管理员）

```
POST /api/auth/register
Body: { username, password, orgName }
```

后端原子操作：
1. 创建 `users` 记录
2. 创建 `organizations` 记录（`name = orgName`）
3. 创建 `user_orgs` 记录（`is_admin=true`, `status=active`）

普通用户无自助注册，由管理员在管理后台创建。

---

## 第三节：前端路由 & 双壳子

### 路由结构

```
/tenant/login                    ← 管理员登录 + 注册（独立页面）
/end-user/login                  ← 统一登录入口（独立页面，管理员也可用）

/org/[orgName]/...               ← 管理后台壳子（is_admin=true 守卫）
/end-user/[orgName]/...          ← 用户端壳子（已登录即可）
```

### 两套壳子边界

| | `/tenant/*` 管理后台 | `/end-user/*` 用户端 |
|---|---|---|
| **目标用户** | 管理员 | 所有登录用户（含管理员） |
| **认证守卫** | `is_admin=true`，否则 403 | 只检查已登录 |
| **导航内容** | 项目管理、模型设计、RBAC 配置、用户管理 | 数据工作台、项目列表 |
| **共享** | Design System | Design System |

### 登录后跳转逻辑

```
/tenant/login 登录成功
  └─→ 直接跳转 /org/[orgName]/projects

/end-user/login 登录成功
  ├─ is_admin=true  → 跳转 /org/[orgName]/projects（管理后台）
  └─ is_admin=false → 跳转 /end-user/[orgName]/workspace
```

---

## 第四节：GraphQL API 变更

### Org Schema

**新增 / 修改：**

```graphql
# 用户管理（管理员操作）
createUser(username: String!, password: String!, isAdmin: Boolean!): CreateUserResult!
updateUserStatus(userId: ID!, status: UserStatus!): UpdateUserStatusResult!
deleteUser(userId: ID!): DeleteUserResult!
resetUserPassword(userId: ID!, newPassword: String!): ResetUserPasswordResult!

# 用户列表
users(where: UserWhereInput, after: String, first: Int): UserConnection!

# 当前用户可访问的项目（原 endUserProjects）
myProjects: [Project!]!
```

**废弃：**
- `createEndUser` → `createUser`
- `updateEndUserStatus` → `updateUserStatus`
- `deleteEndUser` → `deleteUser`
- `resetEndUserPassword` → `resetUserPassword`
- `endUserProjects` → `myProjects`

### Project Schema

无接口签名变更，仅底层 `userId` 指向统一 `users` 表：

```graphql
# 不变，userId 现在指向统一 users 表
assignProjectRole(userId: ID!, projectSlug: String!, roleName: String!): ...
revokeProjectRole(userId: ID!, projectSlug: String!, roleName: String!): ...
```

---

## 授权模型

### 默认：所有接口仅 admin 可调用

Org Schema 的所有 mutation/query，默认只有 `is_admin=true` 的用户可以调用。

**例外**：显式标记 `allowEndUser` 的接口才允许普通用户访问，例如：
- `myProjects` — 普通用户查自己的项目列表
- `me` — 查当前用户信息

授权检查在 **resolver 层**做（不在中间件层），从 context 中读取 `IsAdmin`，不通过直接返回 Unauthorized 错误。

### Gateway Header 注入

Gateway 从统一 ES256 token claims 中读取并注入以下 header，Backend 无条件信任：

```
X-User-ID:   <user_id>
X-Org-Name:  <org_name>
X-Is-Admin:  true | false
```

### accounts 表

私有 DB 的 `accounts` 表（存 refresh token）`user_id` FK 改为指向统一 `users` 表。破坏性更新，直接清表重建，不保留旧数据。

---

## 已决策事项

- **builtin end-user**：直接废弃，不迁移。本质是 tenant-user 在 end-user 侧的镜像实体，合并后无需保留。
- **refresh token cookie**：原 `mc_refresh_token` + `mc_enduser_refresh_token` 合并为一个 cookie。
- **数据迁移**：破坏性更新，不保留旧数据，不写迁移脚本。
