# 变更提案：EndUser 访问授权模型重设计

**变更名**: `enduser-access-redesign`  
**状态**: 设计中  
**日期**: 2026-05-02  

---

## 背景与问题

当前设计（EndUser v2）在"访问控制"这件事上引入了 `EndUserProjectAccess` 这个独立实体，
表达 EndUser ↔ Project ↔ PermissionBundle 三元关系。

这个设计带来了以下问题：

1. **语义重叠**：Project 下已有完整的 Role → Bundle → Permission 体系，
   `EndUserProjectAccess` 再引入一条 `(endUserId, bundleId)` 的独立授权路径，
   形成了两条并行的权限赋予渠道（Channel 1: 直接 Bundle, Channel 2: Role → Bundles），
   逻辑冗余且前端难以呈现。

2. **"授权"概念不清晰**：管理员不知道该走哪条路。
   "分配 Role"和"给 Project 访问权"是两个独立操作，但它们都影响用户能做什么。

3. **散乱的入口**：
   - Org 侧用户详情页有「授权」按钮（走 `grantEndUserProjectAccess`）
   - Project 侧有访问控制页（走相同逻辑）
   - 两个入口做同一件事，没有明确分工

---

## 核心决策

> **废弃 `EndUserProjectAccess` 实体及其相关接口。**  
> **用 Role Assignment 作为唯一的授权通道。**

### 新的核心等式

```
授权 ≡ 给用户在某个 Project 下分配一个 Role
```

用户在 Project 下有 Role 分配 = 他能访问这个 Project，同时 Role 内的 Bundles 决定数据权限。

### 两层分工（清晰化）

```
Org 层（管人）
  └── /org/{orgName}/end-users       账号的生命周期：创建、禁用、删除

Project 层（管权）
  └── /project/{slug}/end-user-access    授权：分配 Role → 用户可访问 Project
                                         数据权限：Role 内的 Bundle 定义
```

---

## 影响范围

### 后端删除

| 实体 / 接口 | 类型 | 动作 |
|------------|------|------|
| `end_user_project_access` 表 | DB | DROP TABLE + migration |
| `EndUserProjectAccess` domain entity | Go | 删除 |
| `EndUserProjectAccessRepository` | Go | 删除 |
| `grantEndUserProjectAccess` | GraphQL Mutation | 删除 |
| `updateEndUserProjectAccess` | GraphQL Mutation | 删除 |
| `revokeEndUserProjectAccess` | GraphQL Mutation | 删除 |
| `listProjectEndUserAccess` | GraphQL Query | 删除 |

### 后端新增

| 接口 | 类型 | 说明 |
|------|------|------|
| `listProjectEndUserRoleUsers` | GraphQL Query | 新增 Project 视角的反向查询，列出该 Project 下所有有 Role 分配的用户（含 Role 信息） |

现有的 `endUserRoleAssignments(endUserId)` 只能按用户查，缺少 Project 侧视角。

```graphql
# 新增 Query
listProjectEndUserRoleUsers(input: ListProjectEndUserRoleUsersInput!): ListProjectEndUserRoleUsersPayload!

input ListProjectEndUserRoleUsersInput {
  search: String   # EndUser username 模糊搜索
  roleId: ID       # 可选，按 Role 过滤
  first: Int
  after: String
}

type ProjectEndUserRoleUser {
  endUser: EndUser!
  role: EndUserRole!
  assignedAt: Time!
}

type ListProjectEndUserRoleUsersPayload {
  connection: ProjectEndUserRoleUserConnection
  error: ListProjectEndUserRoleUsersError
}
```

### 前端变更

| 文件 | 变更 |
|------|------|
| `[userId]/page.tsx`（Org 用户详情页） | 「项目访问」区域：数据源从 `listProjectEndUserAccess` → `endUserRoleAssignments`；移除 Grant 弹窗；每行改为「前往管理→」跳转链接 |
| Project `/end-user-access` 页 | 列表数据源：`listProjectEndUserAccess` → 新的 `listProjectEndUserRoleUsers`；「添加用户」弹窗从 grant access → `assignEndUserRole` |
| `src/api-client/rbac/graphql-docs.ts` | 删除 `GRANT_PROJECT_END_USER_ACCESS`、`REVOKE_PROJECT_END_USER_ACCESS`、`LIST_PROJECT_END_USER_ACCESS` 等文档 |

---

## 不变的部分

- Org GraphQL：`listEndUsers` / `createEndUser` / `updateEndUserStatus` / `deleteEndUser` ✅
- Project RBAC：Role / Bundle / Permission 管理全部保留 ✅
- `assignEndUserRole` / `revokeEndUserRole` Mutation 保留 ✅
- `endUserRoleAssignments(endUserId)` Query 保留 ✅

---

## 实施顺序建议

1. 后端：新增 `listProjectEndUserRoleUsers` Query
2. 后端：废弃 `end_user_project_access` 表及相关接口（migration + code cleanup）
3. 前端同步 contract（`front-contract-pull`）
4. 前端：重写 Project `/end-user-access` 页面逻辑
5. 前端：更新 Org 用户详情页「项目访问」区域
