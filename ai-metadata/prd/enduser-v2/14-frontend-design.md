# EndUser v2 前端设计

> 本文档描述 EndUser v2 改造所需的前端页面、路由、组件和 BFF 变更。

---

## 变更总览

| 类别 | v1（当前） | v2（目标） |
|------|-----------|-----------|
| **登录入口** | `/u/[orgName]/[projectSlug]/login` | `/u/[orgName]/login` |
| **EndUser 管理** | Project 下 `/end-users` | Org 下 `/end-users` |
| **Project 访问控制** | 无独立页面（在管理页混合） | Project 下独立的 `/end-user-access` |
| **JWT 签发** | 登录直接携带 `projectSlug` | 登录后选择 Project，再签发带 `projectSlug` 的 JWT |
| **数据访问入口** | 同 v1 | 不变：`/u/[orgName]/[projectSlug]/data` |

---

## 路由变更

### 新增路由

| 路由 | 文件路径 | 说明 |
|------|---------|------|
| `/u/[orgName]/login` | `src/app/u/[orgName]/login/page.tsx` | 新的 Org 级 EndUser 登录页 |
| `/u/[orgName]/select-project` | `src/app/u/[orgName]/select-project/page.tsx` | 登录后选择 Project 的中间页 |
| `/org/[orgName]/end-users` | `src/app/org/[orgName]/end-users/page.tsx` | Org 级 EndUser 管理页（从 Project 迁移） |
| `/org/[orgName]/project/[projectSlug]/end-user-access` | `src/app/org/[orgName]/project/[projectSlug]/end-user-access/page.tsx` | Project 级 EndUser 访问控制页 |

### 废弃/调整路由

| v1 路由 | 处理方式 | 说明 |
|---------|---------|------|
| `/u/[orgName]/[projectSlug]/login` | **重定向** → `/u/[orgName]/login` | 保留重定向兼容旧链接 |
| `/org/[orgName]/project/[projectSlug]/end-users` | **迁移** → Org 级页面 | 原 Project 级管理页废弃 |

---

## 页面详细设计

### 1. Org 级 EndUser 登录页

**路由**：`/u/[orgName]/login`

**功能**：
- EndUser 通过 Org 统一入口登录，不再绑定到特定 Project
- 输入：用户名 + 密码
- 登录成功 → 跳转到 `/u/[orgName]/select-project`（若有多个 Project 访问权）
- 登录成功 → 直接跳转到 `/u/[orgName]/[projectSlug]/data`（若只有一个 Project 访问权）
- 无 Project 访问权 → 页面内显示错误："您暂无项目访问权限，请联系管理员授权"（不重定向）

**文件**：
- `src/app/u/[orgName]/login/page.tsx` — 服务端组件，渲染 `EndUserOrgLoginCard`
- `src/app/u/[orgName]/login/layout.tsx` — 公开布局（无需 EndUser auth）
- `src/web/components/features/end-user-auth/EndUserOrgLoginCard.tsx` — 登录表单组件

**BFF 接口**：
- `POST /api/bff/org/[orgName]/end-user/auth/login`
  - 请求：`{ username: string, password: string }`
  - 响应（成功）：`{ projects: Array<{ slug: string, name: string }> }` + 设置临时 session cookie
  - 响应（无 Project 权限）：`{ error: { code: "NO_PROJECT_ACCESS", message: "..." } }`

---

### 2. Project 选择页

**路由**：`/u/[orgName]/select-project`

**功能**：
- 显示 EndUser 有权访问的 Project 列表
- 每个 Project 展示：名称、描述（可选）
- 点击 Project → 调用 `select-project` 接口签发带 `projectSlug` 的 JWT → 跳转到 `/u/[orgName]/[projectSlug]/data`

**说明**：
- 该页面仅在用户有 **多个** Project 访问权时出现
- 若只有一个 Project，登录时直接签发 JWT 并重定向，不经过此页

**文件**：
- `src/app/u/[orgName]/select-project/page.tsx`
- `src/web/components/features/end-user-auth/EndUserProjectSelector.tsx`

**BFF 接口**：
- `POST /api/bff/org/[orgName]/end-user/auth/select-project`
  - 请求：`{ projectSlug: string }` + 临时 session cookie（验证 identity）
  - 响应：签发正式 access token + refresh token cookie（path 绑定到 `/u/[orgName]/[projectSlug]/`）

---

### 3. Org 级 EndUser 管理页

**路由**：`/org/[orgName]/end-users`

**功能**：
- 列表展示 Org 下所有 EndUser（不再按 Project 过滤）
- 搜索（按用户名）
- 创建 EndUser（仅 username + password，不绑定 Project）
- 禁用/启用 EndUser
- 删除 EndUser
- **新增**：点击用户可查看该用户有权访问的 Project 列表

**文件**：
- `src/app/org/[orgName]/end-users/page.tsx`
- `src/web/components/features/end-users/EndUsersManagementTable.tsx`
- `src/web/components/features/end-users/CreateEndUserDialog.tsx`
- `src/web/components/features/end-users/EndUserProjectAccessDrawer.tsx` — 查看/编辑用户 Project 访问权

**使用的 GraphQL**：
- `ListOrgEndUsers`（来自 Org Schema）
- `CreateOrgEndUser`
- `UpdateOrgEndUserStatus`
- `DeleteOrgEndUser`

---

### 4. Project 级 EndUser 访问控制页

**路由**：`/org/[orgName]/project/[projectSlug]/end-user-access`

**功能**：
- 展示有权访问当前 Project 的 EndUser 列表
- 授权新 EndUser（从 Org 用户池搜索选择）
- 修改 EndUser 的 PermissionBundle
- 移除 EndUser 的 Project 访问权
- **不再提供**：创建/删除 EndUser 账号（账号管理移到 Org 级）

**文件**：
- `src/app/org/[orgName]/project/[projectSlug]/end-user-access/page.tsx`
- `src/web/components/features/end-user-access/EndUserAccessTable.tsx`
- `src/web/components/features/end-user-access/GrantEndUserAccessDialog.tsx` — 从 Org 用户池选人授权
- `src/web/components/features/end-user-access/EditPermissionBundleDialog.tsx`

**使用的 GraphQL**：
- `ListEndUserProjectAccesses`（来自 Project Schema）
- `GrantEndUserProjectAccess`
- `UpdateEndUserProjectAccess`
- `RevokeEndUserProjectAccess`
- `ListOrgEndUsers`（用于搜索 Org 用户池，来自 Org Schema）

---

## BFF 变更

### 新增 BFF 路由

| 方法 | 路由 | 文件 | 说明 |
|------|------|------|------|
| `POST` | `/api/bff/org/[orgName]/end-user/auth/login` | `src/app/api/bff/org/[orgName]/end-user/auth/login/route.ts` | Org 级登录，返回 Project 列表 |
| `POST` | `/api/bff/org/[orgName]/end-user/auth/select-project` | `src/app/api/bff/org/[orgName]/end-user/auth/select-project/route.ts` | 选择 Project，签发最终 JWT |
| `POST` | `/api/bff/org/[orgName]/end-user/auth/logout` | `src/app/api/bff/org/[orgName]/end-user/auth/logout/route.ts` | 登出（清除 Cookie） |
| `GET` | `/api/bff/org/[orgName]/end-user/auth/me` | `src/app/api/bff/org/[orgName]/end-user/auth/me/route.ts` | 获取当前用户信息 |
| `POST` | `/api/bff/org/[orgName]/end-user/auth/refresh` | `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts` | 刷新 Token（Cookie-based） |

### 废弃 BFF 路由（向后兼容期内保留，返回 410 Gone 或重定向）

| 废弃路由 | 替代路由 |
|---------|---------|
| `POST /api/bff/end-user/auth/login` | `POST /api/bff/org/[orgName]/end-user/auth/login` |
| `POST /api/bff/end-user/auth/register` | Org 级自注册接口（后续规划） |
| `POST /api/bff/end-user/auth/refresh` | `POST /api/bff/org/[orgName]/end-user/auth/refresh` |
| `POST /api/bff/end-user/auth/logout` | `POST /api/bff/org/[orgName]/end-user/auth/logout` |
| `GET /api/bff/end-user/auth/me` | `GET /api/bff/org/[orgName]/end-user/auth/me` |

### 登录流程 BFF 实现细节

```
POST /api/bff/org/[orgName]/end-user/auth/login
  │
  ├── 1. 验证参数（username, password 非空）
  ├── 2. 调用 Go Backend → 验证账号密码
  │       ↓ 验证通过，返回 userId + refresh_token
  ├── 3. 查询该用户的 Project 访问列表
  │       ↓
  │   ┌── 有 1 个 Project 访问权：
  │   │    - 签发 JWT（含 projectSlug）
  │   │    - 设置 refresh token Cookie（path: /u/[orgName]/[projectSlug]/）
  │   │    - 响应：{ singleProject: true, projectSlug, accessToken }
  │   │
  │   ├── 有 N > 1 个 Project 访问权：
  │   │    - 签发临时 "pending" session Cookie（仅含 userId，不含 projectSlug）
  │   │    - 响应：{ singleProject: false, projects: [...] }
  │   │
  │   └── 无 Project 访问权：
  │        - 响应：{ error: { code: "NO_PROJECT_ACCESS", message: "您暂无项目访问权限，请联系管理员授权" } }
  │
  └── 前端根据响应决定：跳转 data 页 / 跳转 select-project 页 / 显示错误
```

```
POST /api/bff/org/[orgName]/end-user/auth/select-project
  │
  ├── 1. 读取 pending session Cookie（验证 userId）
  ├── 2. 验证 projectSlug 是否在用户可访问列表中
  ├── 3. 签发正式 JWT（含 orgName, projectSlug, userId）
  ├── 4. 清除 pending Cookie
  ├── 5. 设置正式 refresh token Cookie（path: /u/[orgName]/[projectSlug]/）
  └── 6. 响应：{ accessToken, projectSlug }
```

### Cookie 变更

| Cookie | v1 Path | v2 Path | 说明 |
|--------|---------|---------|------|
| `end_user_refresh_token` | `/u/[orgName]/[projectSlug]/` | `/u/[orgName]/[projectSlug]/` | 正式 refresh token，不变 |
| `end_user_pending_session` | 无 | `/u/[orgName]/` | 新增：登录后等待选 Project 的临时 session |

---

## 类型变更

### `src/types/end-user-auth.ts`

```typescript
// 新增：登录响应（v2）
export type EndUserLoginResponseV2 =
  | {
      singleProject: true
      projectSlug: string
      accessToken: string
    }
  | {
      singleProject: false
      projects: Array<{ slug: string; name: string }>
    }
  | {
      error: { code: 'NO_PROJECT_ACCESS'; message: string }
    }

// 新增：EndUser 可访问的 Project
export interface EndUserAccessibleProject {
  slug: string
  name: string
  permissionBundleId: string | null
}

// 保留（不变）
export interface EndUserJWTPayload {
  sub: string // userId
  org_name: string
  project_slug: string // 正式 JWT 仍包含 projectSlug
  role: 'end_user'
  iss: 'mc-enduser'
}
```

---

## 导航变更

### Org 级侧边栏

在 Org 管理导航中新增"终端用户"入口（对应 `/org/[orgName]/end-users`）：

```
Org 级导航:
├── 项目管理
├── 成员管理
├── 角色与权限
├── 终端用户        ← 新增（原来在 Project 下）
└── 设置
```

### Project 级侧边栏

将原来的"终端用户"改名为"访问控制"，功能聚焦：

```
Project 级导航:
├── 数据模型
├── 数据库集群
├── 访问控制        ← 原"终端用户"，路由改为 /end-user-access
├── RBAC 权限
└── 项目设置
```

---

## GraphQL Codegen 更新

v2 引入 Org Schema 的新 EndUser 相关 query/mutation，需要重新运行 codegen：

```bash
cd modelcraft-front

# 1. 确保后端已 subtree push 最新 contract
# （前端禁止直接改 contract/，通过 front-contract-pull skill 同步）
front-contract-pull

# 2. 运行 codegen
npm run codegen
```

新生成的 hooks 将包含：
- `useListOrgEndUsersQuery`
- `useCreateOrgEndUserMutation`
- `useUpdateOrgEndUserStatusMutation`
- `useDeleteOrgEndUserMutation`
- `useListEndUserProjectAccessesQuery`
- `useGrantEndUserProjectAccessMutation`
- `useUpdateEndUserProjectAccessMutation`
- `useRevokeEndUserProjectAccessMutation`

---

## 迁移兼容性说明

### 旧登录 URL 兼容

保留 `/u/[orgName]/[projectSlug]/login` 路由，添加 Next.js redirect：

```typescript
// src/app/u/[orgName]/[projectSlug]/login/page.tsx
import { redirect } from 'next/navigation'

export default function LegacyLoginPage({ params }) {
  redirect(`/u/${params.orgName}/login`)
}
```

这确保旧链接（如书签、邮件链接）不会 404。

### 现有 Cookie 失效处理

v2 的 Cookie path 结构不变（`/u/[orgName]/[projectSlug]/`），但 pending session Cookie 是新增的。现有的 refresh token Cookie 在 v2 仍然有效，不需要强制重新登录。

---

## 参考

- [12-graphql-api-design.md](./12-graphql-api-design.md) — GraphQL API 接口定义
- [13-database-schema.md](./13-database-schema.md) — DB Schema 变更
- [前端架构总览](../../front/development/architecture.md) — 目录规范
- [API Client 设计规范](../../front/development/api-client-design.md) — API Client 层规范
