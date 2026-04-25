# EndUser 身份系统 v2 — GraphQL API 设计

---

## 变更总览

| 操作 | 接口 | v1 所在 Schema | v2 所在 Schema |
|------|------|--------------|--------------|
| 查询 EndUser 列表 | `listEndUsers` | Project GraphQL | **Org GraphQL** |
| 创建 EndUser | `createEndUser` | Project GraphQL | **Org GraphQL** |
| 更新 EndUser 状态 | `updateEndUserStatus` | Project GraphQL | **Org GraphQL** |
| 删除 EndUser | `deleteEndUser` | Project GraphQL | **Org GraphQL** |
| **新增** 授权 EndUser 访问 Project | `grantEndUserProjectAccess` | 无 | **Project GraphQL** |
| **新增** 撤销 EndUser 项目访问 | `revokeEndUserProjectAccess` | 无 | **Project GraphQL** |
| **新增** 查询 Project 的访问用户列表 | `listProjectEndUserAccess` | 无 | **Project GraphQL** |
| **新增** 登录后获取可访问 Project 列表 | `listAccessibleProjects` | 无 | **End-User Runtime** |
| 初始化私有库 | `initPrivateDB` | Project GraphQL | 废弃（v2 不再需要 per-project 私有库） |

---

## Org GraphQL — EndUser 账号管理

> 文件: `api/graph/org/schema/end_user.graphql`（新文件）

```graphql
# ============================================
# EndUser Error Types
# ============================================

type EndUserNotFound implements Error {
  message: String!
}

type EndUserAlreadyExists implements Error {
  message: String!
}

type EndUserPasswordTooWeak implements Error {
  message: String!
  suggestion: String
}

union CreateEndUserError = EndUserAlreadyExists | EndUserPasswordTooWeak | InvalidInput
union UpdateEndUserError = EndUserNotFound | InvalidInput
union DeleteEndUserError = EndUserNotFound
union ListEndUsersError = InvalidInput

# ============================================
# EndUser Types
# ============================================

type EndUser implements Node {
  id: ID!
  username: String!
  isForbidden: Boolean!
  createdBy: String
  createdAt: Time!
  updatedAt: Time!
}

type EndUserConnection {
  nodes: [EndUser!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

# ============================================
# Payload Types
# ============================================

type CreateEndUserPayload {
  endUser: EndUser
  error: CreateEndUserError
}

type UpdateEndUserStatusPayload {
  endUser: EndUser
  error: UpdateEndUserError
}

type DeleteEndUserPayload {
  success: Boolean!
  error: DeleteEndUserError
}

type ListEndUsersPayload {
  connection: EndUserConnection
  error: ListEndUsersError
}

# ============================================
# Input Types
# ============================================

input CreateEndUserInput {
  username: String!   # 3–64 chars, ^[a-zA-Z0-9_-]+$
  password: String!   # 至少 8 位，含字母 + 数字
}

input UpdateEndUserStatusInput {
  userId: ID!
  isForbidden: Boolean!
}

input DeleteEndUserInput {
  userId: ID!
}

input ListEndUsersInput {
  search: String  # username 模糊搜索
  first: Int      # 分页大小，默认 20
  after: String   # 游标
}

# ============================================
# Query & Mutation（挂载到 Org GraphQL）
# ============================================

extend type Query {
  listEndUsers(input: ListEndUsersInput): ListEndUsersPayload!
    @hasPermission(action: "end-user:read")
}

extend type Mutation {
  createEndUser(input: CreateEndUserInput!): CreateEndUserPayload!
    @hasPermission(action: "end-user:create")

  updateEndUserStatus(input: UpdateEndUserStatusInput!): UpdateEndUserStatusPayload!
    @hasPermission(action: "end-user:update")

  deleteEndUser(input: DeleteEndUserInput!): DeleteEndUserPayload!
    @hasPermission(action: "end-user:delete")
}
```

---

## Project GraphQL — 访问控制管理

> 文件: `api/graph/project/schema/end_user_access.graphql`（新文件）

```graphql
# ============================================
# EndUser Project Access 相关类型
# ============================================

type EndUserProjectAccessNotFound implements Error {
  message: String!
}

type EndUserProjectAccessAlreadyExists implements Error {
  message: String!
}

union GrantEndUserAccessError = EndUserNotFound | EndUserProjectAccessAlreadyExists | InvalidInput
union RevokeEndUserAccessError = EndUserProjectAccessNotFound
union ListProjectEndUserAccessError = InvalidInput

# ============================================
# Types
# ============================================

type EndUserProjectAccess {
  id: ID!
  endUser: EndUser!           # 关联的 Org 级 EndUser
  permissionBundleId: ID!
  permissionBundleName: String!
  grantedBy: String!
  grantedAt: Time!
}

type EndUserProjectAccessConnection {
  nodes: [EndUserProjectAccess!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

# ============================================
# Payload Types
# ============================================

type GrantEndUserAccessPayload {
  access: EndUserProjectAccess
  error: GrantEndUserAccessError
}

type RevokeEndUserAccessPayload {
  success: Boolean!
  error: RevokeEndUserAccessError
}

type ListProjectEndUserAccessPayload {
  connection: EndUserProjectAccessConnection
  error: ListProjectEndUserAccessError
}

# ============================================
# Input Types
# ============================================

input GrantEndUserAccessInput {
  endUserId: ID!            # Org 级 EndUser ID
  permissionBundleId: ID!   # 授予的 PermissionBundle
}

input RevokeEndUserAccessInput {
  accessId: ID!             # EndUserProjectAccess.id
}

input ListProjectEndUserAccessInput {
  search: String  # EndUser username 模糊搜索
  first: Int
  after: String
}

# ============================================
# Query & Mutation（挂载到 Project GraphQL）
# ============================================

extend type Query {
  listProjectEndUserAccess(input: ListProjectEndUserAccessInput): ListProjectEndUserAccessPayload!
    @hasPermission(action: "end-user-access:read")
}

extend type Mutation {
  grantEndUserAccess(input: GrantEndUserAccessInput!): GrantEndUserAccessPayload!
    @hasPermission(action: "end-user-access:grant")

  revokeEndUserAccess(input: RevokeEndUserAccessInput!): RevokeEndUserAccessPayload!
    @hasPermission(action: "end-user-access:revoke")
}
```

---

## End-User Runtime API — 登录与 Project 选择

> 文件: `api/graph/end_user/schema/auth.graphql`（新增）

```graphql
# EndUser 登录后获取可访问的 Project 列表（用于 Project 选择页）
type AccessibleProject {
  slug: String!
  title: String!
}

type ListAccessibleProjectsPayload {
  projects: [AccessibleProject!]!
}

type EndUserNoProjectAccess implements Error {
  message: String!  # "您暂无项目访问权限，请联系管理员授权"
}

union EndUserLoginError = InvalidInput | EndUserNoProjectAccess

# 登录结果：返回可访问 Project 列表（由 BFF 调用，返回给前端展示选择）
# 实际 JWT 签发在 BFF 层（EndUser 选择 Project 后，BFF 调用 issueEndUserToken）
```

---

## BFF 路由变更（REST）

| 端点 | v1 | v2 |
|------|----|----|
| EndUser 登录 | `POST /api/bff/end-user/auth/login` (project 路由下) | `POST /api/bff/org/{orgName}/end-user/auth/login` |
| 刷新 Token | `POST /api/bff/end-user/auth/refresh` (project 路由下) | `POST /api/bff/org/{orgName}/end-user/auth/refresh` |
| **新增** 选择 Project 并签发 JWT | 无 | `POST /api/bff/org/{orgName}/end-user/auth/select-project` |
| 登出 | `POST /api/bff/end-user/auth/logout` (project 路由下) | `POST /api/bff/org/{orgName}/end-user/auth/logout` |

### `POST /api/bff/org/{orgName}/end-user/auth/login`

```typescript
// Request
{ username: string; password: string }

// Response 200 — 登录成功，返回可访问 Project 列表
{ projects: Array<{ slug: string; title: string }> }

// Response 401 — 账号或密码错误
{ error: "INVALID_CREDENTIALS" }

// Response 403 — 账号被禁用
{ error: "ACCOUNT_FORBIDDEN" }

// Response 403 — 无 Project 访问权限
{ error: "NO_PROJECT_ACCESS"; message: "您暂无项目访问权限，请联系管理员授权" }
```

### `POST /api/bff/org/{orgName}/end-user/auth/select-project`

```typescript
// Request（附带登录后的临时 session cookie）
{ projectSlug: string }

// Response 200 — 签发带 projectSlug 的 EndUser JWT
{ accessToken: string }  // 同时写 HttpOnly Cookie: end_user_token

// Response 403 — 无该 Project 访问权限
{ error: "PROJECT_ACCESS_DENIED" }
```

---

## 废弃接口

| 接口 | 废弃原因 |
|------|---------|
| `initPrivateDB` (Project GraphQL) | v2 账号统一存 mc_meta，不再需要 per-project 私有库初始化 |
| `listEndUsers` (Project GraphQL) | 迁移到 Org GraphQL |
| `createEndUser` (Project GraphQL) | 迁移到 Org GraphQL |
| `updateEndUserStatus` (Project GraphQL) | 迁移到 Org GraphQL |
| `deleteEndUser` (Project GraphQL) | 迁移到 Org GraphQL |
