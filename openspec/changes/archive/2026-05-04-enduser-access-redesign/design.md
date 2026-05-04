# 设计文档：EndUser 访问授权模型重设计

**变更名**: `enduser-access-redesign`  
**状态**: 已确认  
**日期**: 2026-05-02

---

## 架构图

### 现状（v2 原始设计）

```
Org 层
  EndUser（账号）
      │
      ├── [Channel A] grantEndUserProjectAccess ──▶ EndUserProjectAccess
      │                                                  ├── endUserId
      │                                                  ├── projectSlug
      │                                                  └── permissionBundleId  ← 直接绑 Bundle
      │
      └── [Channel B] assignEndUserRole ──────────▶ EndUserRoleUser
                                                         ├── endUserId
                                                         └── roleId ──▶ Role ──▶ Bundles

  问题：两条路并行，语义重叠
```

### 目标（重设计后）

```
Org 层
  EndUser（账号）                          ← 只管账号生命周期
      │
      │ "这个用户有访问哪些 Project？"
      │ （只读展示，从 RoleUser 反查）
      ▼
      [只读] 已授权 Project 列表 ──────────▶ 跳转到 Project 权限页 →

Project 层
  assignEndUserRole ──────────────────────▶ EndUserRoleUser
      ├── endUserId                              └── roleId
      ├── projectSlug（by context）                    └── Role
      └── roleId                                            └── Bundles
                                                                   └── Permissions

  "授权" = 创建一条 RoleUser 记录，Role 同时决定能访问 + 能做什么
```

---

## 数据模型

### 废弃

```sql
-- 删除
DROP TABLE end_user_project_access;
```

### 保留（无变更）

```sql
-- end_user_role_users（保留，这是唯一的授权通道）
CREATE TABLE end_user_role_users (
  id            UUID PRIMARY KEY,
  org_name      TEXT NOT NULL,
  project_slug  TEXT NOT NULL,
  end_user_id   UUID NOT NULL REFERENCES end_user_users(id),
  role_id       UUID NOT NULL REFERENCES end_user_roles(id),
  assigned_by   TEXT NOT NULL,
  assigned_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## GraphQL Schema 变更

### 后端删除（`api/graph/project/schema/end_user_access.graphql`）

整个文件废弃。删除以下接口：
- `listProjectEndUserAccess` (Query)
- `grantEndUserProjectAccess` (Mutation)
- `updateEndUserProjectAccess` (Mutation)
- `revokeEndUserProjectAccess` (Mutation)

### 后端新增（追加到 `api/graph/project/schema/rbac.graphql`）

```graphql
# ── 新增：Project 视角的用户角色分配列表 ──────────────────────────

type ProjectEndUserRoleUser {
  endUser: EndUser!
  role: EndUserRole!
  assignedAt: Time!
}

type ProjectEndUserRoleUserConnection {
  nodes: [ProjectEndUserRoleUser!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ListProjectEndUserRoleUsersPayload {
  connection: ProjectEndUserRoleUserConnection
  error: ListProjectEndUserRoleUsersError
}

union ListProjectEndUserRoleUsersError = InvalidInput | ProjectNotFound

input ListProjectEndUserRoleUsersInput {
  search: String    # EndUser username 模糊搜索
  roleId: ID        # 可选，按 Role 过滤
  first: Int
  after: String
}

extend type Query {
  # 新增：列出当前 Project 下所有有角色分配的用户
  listProjectEndUserRoleUsers(
    input: ListProjectEndUserRoleUsersInput
  ): ListProjectEndUserRoleUsersPayload! @hasPermission(action: "rbac:read")
}
```

> `assignEndUserRole` / `revokeEndUserRole` Mutation 已存在，无需新增。

---

## 页面交互设计

### Org 侧：`/org/{orgName}/end-users/{userId}`（用户详情页）

**「项目访问」区域变更：**

```
┌─────────────────────────────────────────────────────────┐
│ 项目访问                                                 │
├──────────────────────┬────────────────┬─────────────────┤
│ 项目名               │ 角色           │ 授权时间        │
├──────────────────────┼────────────────┼─────────────────┤
│ 数据平台             │ 只读用户       │ 2026-04-01      │  前往管理 →
│ 销售分析             │ 编辑用户       │ 2026-04-15      │  前往管理 →
└──────────────────────┴────────────────┴─────────────────┘
  「前往管理 →」 → 跳转到 /org/{orgName}/project/{slug}/end-user-access
```

- **移除**：「授权」按钮和 Grant 弹窗
- **数据来源**：`endUserRoleAssignments(endUserId)` → 按 `projectSlug` 分组展示
- 用户在 Org 侧只能**看**，不能**改**

### Project 侧：`/org/{orgName}/project/{slug}/end-user-access`（访问控制页）

```
┌─────────────────────────────────────────────────────────────┐
│ 终端用户访问控制                              + 添加用户     │
├─────────────────┬────────────────┬──────────────┬──────────┤
│ 用户名          │ 角色           │ 授权时间     │ 操作     │
├─────────────────┼────────────────┼──────────────┼──────────┤
│ alice           │ 只读用户       │ 2026-04-01   │ 修改角色 撤销 │
│ bob             │ 编辑用户       │ 2026-04-15   │ 修改角色 撤销 │
└─────────────────┴────────────────┴──────────────┴──────────┘
```

**「添加用户」弹窗：**

```
┌─────────────────────────────────┐
│ 添加终端用户                     │
├─────────────────────────────────┤
│ 用户名  [下拉选择 Org 内用户]    │
│ 角色    [下拉选择 Role]          │
│         默认预选第一个可用 Role  │
└──────────────┬──────────────────┘
               │ 确认 → assignEndUserRole(endUserId, roleId)
```

> 「可用 Role」指 `isImplicit = false` 的 Role。若 Project 下尚无任何 Role，提示管理员先到 RBAC 页创建 Role 后再添加用户。
>
> **注意**：未被分配任何 Role 的用户，登录后看不到该 Project（因为 Project 访问列表由 `end_user_role_users` 反查得出）。

**「修改角色」**：先 `revokeEndUserRole`，再 `assignEndUserRole`（或后端提供 update 接口）

**「撤销」**：`revokeEndUserRole`

---

## 登录流程（不变）

EndUser 登录后，"能访问哪些 Project"通过查询 `endUserRoleAssignments` 反查有 Role 分配的 Project 列表，
不再依赖 `listAccessibleProjects` 走 `EndUserProjectAccess` 路径。

```
用户登录
  └──▶ 查询 end_user_role_users WHERE end_user_id = ?
  └──▶ 提取不重复的 project_slug 列表
  └──▶ 返回可访问 Project 列表（含 project_title）
```

---

## 后端实现要点

### `listProjectEndUserRoleUsers` Resolver

需要在 Project Scope 下：

1. 从 Context 获取 `projectSlug`
2. 查询 `end_user_role_users WHERE project_slug = ?`
3. JOIN `end_user_users`（用于搜索和返回用户信息）
4. 支持 `search`（username 模糊）和 `roleId` 过滤
5. 返回 cursor-based 分页

### `assignEndUserRole` 调整

现有接口 `AssignEndUserRoleInput { endUserId: ID!, roleId: ID! }` 已够用，
但需要确认：**在 Project Scope 下调用时，Resolver 是否已正确地从 Context 获取 `projectSlug`**。
如果当前实现依赖了 `EndUserProjectAccess` 做「用户是否在本 Project」的存在性检查，需要改为查 `end_user_role_users`。

### 错误类型调整

`EndUserNotFoundInProject` 这个错误类型目前语义是「用户没有 ProjectAccess 记录」。
废弃 ProjectAccess 后，改为「用户不存在于该 Org」（因为 EndUser 是 Org 级的，
Project 下只要是同 Org 用户就能被分配 Role）。

---

## 前端 API Client 变更

### 删除

```typescript
// src/api-client/rbac/graphql-docs.ts
- LIST_PROJECT_END_USER_ACCESS
- GRANT_PROJECT_END_USER_ACCESS
- UPDATE_PROJECT_END_USER_ACCESS
- REVOKE_PROJECT_END_USER_ACCESS
```

### 新增

```typescript
// src/api-client/rbac/graphql-docs.ts

export const LIST_PROJECT_END_USER_ROLE_USERS = gql`
  query ListProjectEndUserRoleUsers($input: ListProjectEndUserRoleUsersInput) {
    listProjectEndUserRoleUsers(input: $input) {
      connection {
        nodes {
          endUser {
            id
            username
            isForbidden
          }
          role {
            id
            name
            description
          }
          assignedAt
        }
        pageInfo {
          hasNextPage
          endCursor
        }
        totalCount
      }
      error {
        __typename
        ... on InvalidInput {
          message
        }
        ... on ProjectNotFound {
          message
        }
      }
    }
  }
`
```

---

## 已确认设计决策

1. **`assignEndUserRole` 中的「用户存在性检查」**  
   错误类型 `EndUserNotFoundInProject` 语义改为「EndUser 不属于当前 Org」。  
   校验逻辑：只检查「EndUser 是否属于同一个 Org」，只要是同 Org 用户即可被分配任意 Role，无需任何 Project 级预注册。

2. **同一用户可以有多个 Role**  
   后端不限制，UI 也支持多 Role 展示（每行一个 Role 分配记录）。

3. **Role 默认选取规则**  
   「添加用户」弹窗的 Role 下拉：默认预选第一个 `isImplicit = false` 的 Role。若当前 Project 下无任何 Role，提示管理员先在 RBAC 页创建 Role。无系统自动生成的默认 Role。
