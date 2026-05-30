# 统一认证 & 管理/用户视图无缝切换

**日期**: 2026-05-30  
**状态**: 待实现  
**范围**: 后端登录接口合并 + 前端单一 auth store + 路由 guard

---

## 背景与目标

### 问题

同一个账号可以同时具有 admin 和 end-user 身份，但目前：

- 后端存在两套登录入口（`/api/auth/login` 和 `/api/end-user/auth/login`），逻辑高度重复
- 前端管理页和用户页各自维护独立的认证状态，切换视图需要重新登录
- 登录响应里返回 projects 列表，这个职责不属于登录接口

### 目标

1. 后端两套登录逻辑合并为同一实现，消除重复代码
2. 前端只维护一个 access token，管理视图和用户视图共用
3. 已登录用户切换视图 = 纯路由跳转，无需重新认证
4. 登录响应不再返回 projects 列表

### 前提条件（已确认）

- 数据库：admin 和 end-user 同在 `users` 表，`user_orgs.is_admin` 区分身份
- JWT：两套登录流已经签发同一结构的 `PlatformClaims`（`user_id`, `org_name`, `is_admin`），issuer 均为 `mc-platform`，签名均为 ES256
- 路由：前端管理路由 `/org/[orgName]/...` 与用户路由 `/end-user/[orgName]/...` 已独立隔离，保持不变

---

## 架构设计

### JWT 结构（不变）

```json
{
  "user_id": "uuid",
  "org_name": "my-org",
  "is_admin": true,
  "key": "mcuser",
  "iss": "mc-platform",
  "sub": "uuid",
  "aud": ["platform"],
  "iat": 1234567890,
  "exp": 1234571490
}
```

`is_admin` 是唯一的身份区分字段，由 Gateway 注入 `X-Is-Admin` header 供后端鉴权使用。

### 权限边界

| 用户类型 | 管理视图 `/org/...` | 用户视图 `/end-user/...` |
|---------|-------------------|------------------------|
| `is_admin=true` | ✅ 允许 | ✅ 允许 |
| `is_admin=false` | ❌ 拒绝（guard 拦截） | ✅ 允许 |

---

## 后端改动

### 1. 合并登录 Handler

两条路径保留，共用同一个 handler 函数：

```
POST /api/auth/login          → authHandler.HandleLogin()
POST /api/end-user/auth/login → authHandler.HandleLogin()  ← 指向同一方法
```

`HandleLogin` 的入参支持 `orgName`（可选），EndUser 登录时传入，tenant 登录时可省略。

### 2. 统一登录逻辑

将 `EndUserAuthAppService.LoginEndUser` 与 `TokenService.Login` 合并，统一流程：

```
1. 解析 identifier（phone/username）+ password
2. 查 users 表验证密码
3. 读 user_orgs.is_admin 确定身份
4. 签发 PlatformClaims JWT（is_admin 写入 token）
5. 生成 refresh token，存库
6. 返回 accessToken + expiresIn（不返回 projects）
```

合并后可删除 `EndUserAuthAppService` 中重复的密码验证、refresh token 生成逻辑。

### 3. 登录响应变更（破坏性）

**移除** `projects` 字段：

```json
// Before
{
  "accessToken": "...",
  "expiresIn": 3600,
  "userId": "...",
  "orgName": "...",
  "projects": [{"slug": "...", "title": "..."}]  ← 删除
}

// After
{
  "accessToken": "...",
  "expiresIn": 3600,
  "userId": "...",
  "orgName": "..."
}
```

前端登录后如需 project 列表，通过独立的 GraphQL query 获取。

### 4. Refresh Token 接口统一

同理，`/api/auth/refresh` 和 `/api/end-user/auth/refresh` 共用同一实现，响应也不再包含 projects。

---

## 前端改动

### 1. 单一 Auth Store

合并为一个 Zustand store，不区分"管理 token"和"用户 token"：

```typescript
interface AuthState {
  accessToken: string | null
  userId: string | null
  orgName: string | null
  isAdmin: boolean          // 从 JWT claims 解析
  expiresAt: number | null
}
```

`isAdmin` 从 JWT payload 解析，无需后端额外返回。

### 2. 路由 Guard

管理路由 `/org/[orgName]/...` 的 guard 逻辑：

```typescript
// middleware 或 layout 层
if (!authStore.isAdmin) {
  redirect(`/end-user/${orgName}/projects`)
}
```

用户路由 `/end-user/[orgName]/...` 只检查是否已登录，不检查 `isAdmin`。

### 3. 视图切换入口

在管理页顶部导航（`is_admin=true` 时显示）：

```
[切换到用户视图] → router.push(`/end-user/${orgName}/projects`)
```

在用户页顶部导航（`is_admin=true` 时显示）：

```
[切换到管理视图] → router.push(`/org/${orgName}/dashboard`)
```

切换时不发任何 API 请求，token 不变。

### 4. 登录后跳转逻辑

```typescript
onLoginSuccess(result) {
  authStore.setToken(result.accessToken)
  if (authStore.isAdmin) {
    router.push(`/org/${result.orgName}/dashboard`)
  } else {
    router.push(`/end-user/${result.orgName}/projects`)
  }
}
```

---

## 不在本次范围内

- 多 org 切换（用户属于多个 org 的场景）
- Token 过期自动续期的 UX 优化
- EndUser 自注册流程的改动
- RBAC 权限点的细化

---

## 验收标准

1. `is_admin=true` 的用户：
   - 通过任意一个登录接口都能成功登录
   - 登录后跳转到管理页
   - 管理页有"切换到用户视图"入口，点击后跳转到用户页，无需重新登录
   - 用户页有"切换到管理视图"入口，点击后跳转到管理页，无需重新登录

2. `is_admin=false` 的用户：
   - 登录后跳转到用户页
   - 没有"切换到管理视图"入口
   - 直接访问管理路由被 guard 拦截并重定向

3. 登录响应不包含 `projects` 字段

4. 后端只有一套密码验证 + refresh token 生成逻辑
