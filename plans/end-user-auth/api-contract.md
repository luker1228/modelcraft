# End-User Auth — API 协议设计

> **所属模块**：终端用户认证（End-User Auth）
> **父文档**：[00-end-user-auth.md](./00-end-user-auth.md)
> **优先级**：P0

---

## 协议分层

| 消费方 | 协议 | Endpoint | 接口数 |
|--------|------|----------|--------|
| **终端用户客户端** | OpenAPI (REST) | `/org/{orgName}/project/{projectSlug}/api/` | 5 个 |
| **开发者管理侧** | GraphQL (Project Schema) | `/graphql/org/{orgName}/project/{projectSlug}/` | 追加到现有 Project GraphQL |

---

## 一、终端用户侧 — OpenAPI

终端用户客户端可见的 5 个接口，路径以 `/org/{orgName}/project/{projectSlug}/api/` 为前缀，**不需要开发者 Token**，使用独立的 end-user JWT（`Authorization: Bearer`）鉴权。

### 1.1 接口总览

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/org/{orgName}/project/{projectSlug}/api/auth/register` | 终端用户自助注册 | 无 |
| `POST` | `/org/{orgName}/project/{projectSlug}/api/auth/login` | 终端用户登录 | 无 |
| `POST` | `/org/{orgName}/project/{projectSlug}/api/auth/logout` | 登出 | End-User JWT |
| `POST` | `/org/{orgName}/project/{projectSlug}/api/auth/refresh` | 刷新 access token | HttpOnly Cookie |
| `GET` | `/org/{orgName}/project/{projectSlug}/api/auth/me` | 获取当前用户信息 | End-User JWT |

### 1.2 路径参数（所有接口共用）

| 参数 | 类型 | 说明 |
|------|------|------|
| `orgName` | string | 组织名称 |
| `projectSlug` | string | 项目标识，用于路由到 `private_{projectSlug}` 库 |

### 1.3 错误格式（统一）

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "用户名或密码错误"
  }
}
```

### 1.4 错误码定义

| 错误码 | HTTP 状态码 | 场景 |
|--------|------------|------|
| `INVALID_CREDENTIALS` | 401 | 用户名/密码错误（含不存在，防枚举） |
| `INVALID_REFRESH_TOKEN` | 401 | refresh token 无效、过期或已撤销 |
| `ACCOUNT_DISABLED` | 403 | 账号已被禁用 |
| `CONFLICT` | 409 | 用户名已存在 |
| `PARAM_INVALID` | 400 | 请求参数格式校验失败 |
| `CLUSTER_NOT_CONFIGURED` | 503 | Project 未关联 DatabaseCluster，终端用户功能不可用 |
| `UNAUTHORIZED` | 401 | 缺少或无效的 end-user JWT |

---

### POST /api/auth/register

终端用户自助注册，任何人均可调用。

**请求体**

```json
{
  "username": "alice",
  "password": "Abc12345"
}
```

| 字段 | 类型 | 必填 | 校验规则 |
|------|------|------|----------|
| `username` | string | ✅ | 3–64 字符，`^[a-zA-Z0-9_-]+$` |
| `password` | string | ✅ | 至少 8 位，包含字母 + 数字 |

**成功响应 `200 OK`**

```json
{
  "access_token": "<jwt>",
  "expires_in": 3600
}
```

> 注册成功后直接登录，返回 access token，同时写 `end_user_refresh_token` HttpOnly Cookie。
> HTTP 200 而非 201：虽然资源被创建，但响应体是 session 凭证而非新建用户对象，200 语义更准确，与 login 接口保持一致，方便前端统一处理。

**错误响应**

| 状态码 | code | 场景 |
|--------|------|------|
| 400 | `PARAM_INVALID` | 用户名格式不合法 / 密码强度不足 |
| 409 | `CONFLICT` | 用户名已存在 |
| 503 | `CLUSTER_NOT_CONFIGURED` | Project 未关联 Cluster |

---

### POST /api/auth/login

**请求体**

```json
{
  "username": "alice",
  "password": "Abc12345"
}
```

**成功响应 `200 OK`**

```json
{
  "access_token": "<jwt>",
  "expires_in": 3600
}
```

> 同时写 `end_user_refresh_token` HttpOnly Cookie（7d）。

**错误响应**

| 状态码 | code | 场景 |
|--------|------|------|
| 401 | `INVALID_CREDENTIALS` | 凭证错误（用户名不存在或密码错误，统一返回） |
| 403 | `ACCOUNT_DISABLED` | 账号被禁用 |
| 503 | `CLUSTER_NOT_CONFIGURED` | Project 未关联 Cluster |

---

### POST /api/auth/logout

**请求头**：`Authorization: Bearer <end-user-access-token>`

**成功响应 `204 No Content`**

> 清除 `end_user_refresh_token` Cookie，revoke 对应 refresh token（best-effort）。

---

### POST /api/auth/refresh

从 HttpOnly Cookie 中读取 `end_user_refresh_token`，执行 token rotation，签发新 access token。

**请求体**：无（从 Cookie 自动读取）

**成功响应 `200 OK`**

```json
{
  "access_token": "<new-jwt>",
  "expires_in": 3600
}
```

> 同时 rotate Cookie：旧 refresh token revoked，新 refresh token 写入 Cookie。

**错误响应**

| 状态码 | code | 场景 |
|--------|------|------|
| 401 | `INVALID_REFRESH_TOKEN` | Cookie 不存在 / token 无效 / 已过期 / 已 revoked |

---

### GET /api/auth/me

**请求头**：`Authorization: Bearer <end-user-access-token>`

**成功响应 `200 OK`**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "username": "alice",
  "created_at": "2026-04-10T08:00:00Z"
}
```

**错误响应**

| 状态码 | code | 场景 |
|--------|------|------|
| 401 | `UNAUTHORIZED` | JWT 缺失或无效 |
| 403 | `ACCOUNT_DISABLED` | 账号已被禁用 |

---

## 二、开发者管理侧 — GraphQL（Project Schema）

追加到现有 Project GraphQL Schema（`api/graph/project/schema/`），新增 `end_user.graphql` 文件。

### 2.1 新增文件

```
api/graph/project/schema/
└── end_user.graphql    ← 新增
```

### 2.2 错误类型（遵循现有 Error Interface + Union 模式）

> **注意**：`Error` interface、`InvalidInput`、`ProjectNotFound` 已在其他 schema 文件定义（分别在 `cluster.graphql`、`field.graphql`、`cluster.graphql`），`end_user.graphql` 直接引用，不重复定义。

```graphql
# ============================================
# End User Error Types（仅定义本模块新增类型）
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

# Error Unions
# ClusterNotFound 复用现有类型（cluster.graphql），语义接近：Project 未配置可用 Cluster
union CreateEndUserError  = EndUserAlreadyExists | EndUserPasswordTooWeak | ClusterNotFound | InvalidInput | ProjectNotFound
union UpdateEndUserError  = EndUserNotFound | ClusterNotFound | InvalidInput | ProjectNotFound
union DeleteEndUserError  = EndUserNotFound | ClusterNotFound | ProjectNotFound
union ListEndUsersError   = ClusterNotFound | ProjectNotFound
```

### 2.3 类型定义

```graphql
# ============================================
# End User Types
# ============================================

type EndUser implements Node {
  id: ID!
  username: String!
  isForbidden: Boolean!
  createdBy: String!       # 创建者开发者 user_id
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
  connection: EndUserConnection   # nullable：error 存在时为 null
  error: ListEndUsersError
}

```graphql
# ============================================
# Input Types
# ============================================

input CreateEndUserInput {
  username: String!     # 3–64 字符，^[a-zA-Z0-9_-]+$
  password: String!     # 至少 8 位，含字母 + 数字
}

input UpdateEndUserStatusInput {
  userId: ID!
  isForbidden: Boolean!
}

input DeleteEndUserInput {
  userId: ID!
}

input ListEndUsersInput {
  search: String         # 用户名模糊搜索
  first: Int             # 分页：取前 N 条，默认 20
  after: String          # 游标
}
```

### 2.5 Query & Mutation（追加到现有 schema）

```graphql
# 追加到 type Query（extend Query）
extend type Query {
  listEndUsers(input: ListEndUsersInput): ListEndUsersPayload! @hasPermission(action: "end-user:read")
}

# 追加到 type Mutation（extend Mutation）
extend type Mutation {
  createEndUser(input: CreateEndUserInput!): CreateEndUserPayload! @hasPermission(action: "end-user:create")
  updateEndUserStatus(input: UpdateEndUserStatusInput!): UpdateEndUserStatusPayload! @hasPermission(action: "end-user:update")
  deleteEndUser(input: DeleteEndUserInput!): DeleteEndUserPayload! @hasPermission(action: "end-user:delete")
}
```

### 2.6 完整 end_user.graphql 文件

```graphql
# api/graph/project/schema/end_user.graphql
#
# 依赖（已在其他文件定义，直接引用）：
#   - interface Error          → cluster.graphql
#   - type InvalidInput        → field.graphql
#   - type ProjectNotFound     → cluster.graphql
#   - type ClusterNotFound     → cluster.graphql
#   - scalar Time              → base.graphql
#   - interface Node           → base.graphql
#   - type PageInfo            → base.graphql

# ============================================
# Error Types（仅新增，不重复定义已有类型）
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

# ClusterNotFound 复用现有类型，表示 Project 未配置可用 Cluster
union CreateEndUserError  = EndUserAlreadyExists | EndUserPasswordTooWeak | ClusterNotFound | InvalidInput | ProjectNotFound
union UpdateEndUserError  = EndUserNotFound | ClusterNotFound | InvalidInput | ProjectNotFound
union DeleteEndUserError  = EndUserNotFound | ClusterNotFound | ProjectNotFound
union ListEndUsersError   = ClusterNotFound | ProjectNotFound

# ============================================
# Types
# ============================================

type EndUser implements Node {
  id: ID!
  username: String!
  isForbidden: Boolean!
  createdBy: String!
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
  connection: EndUserConnection   # nullable：error 存在时为 null
  error: ListEndUsersError
}

# ============================================
# Inputs
# ============================================

input CreateEndUserInput {
  username: String!   # 3–64 字符，^[a-zA-Z0-9_-]+$
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
  search: String   # 用户名模糊搜索
  first: Int       # 取前 N 条，默认 20
  after: String    # 游标
}

# ============================================
# Query & Mutation
# ============================================

extend type Query {
  listEndUsers(input: ListEndUsersInput): ListEndUsersPayload! @hasPermission(action: "end-user:read")
}

extend type Mutation {
  createEndUser(input: CreateEndUserInput!): CreateEndUserPayload! @hasPermission(action: "end-user:create")
  updateEndUserStatus(input: UpdateEndUserStatusInput!): UpdateEndUserStatusPayload! @hasPermission(action: "end-user:update")
  deleteEndUser(input: DeleteEndUserInput!): DeleteEndUserPayload! @hasPermission(action: "end-user:delete")
}
```

---

## 三、BFF 路由映射

### 终端用户侧（BFF Route Handlers → Go Backend `/internal/`）

| BFF 公开路由 | Go 内网接口 | 说明 |
|-------------|------------|------|
| `POST /api/auth/register` | `POST /internal/end-user/auth/register` | BFF 签发 JWT + 写 Cookie |
| `POST /api/auth/login` | `POST /internal/end-user/auth/login` | BFF 签发 JWT + 写 Cookie |
| `POST /api/auth/logout` | `POST /internal/end-user/auth/logout` | BFF 清 Cookie + revoke token |
| `POST /api/auth/refresh` | `POST /internal/end-user/auth/refresh` | BFF 读 Cookie → token rotation → 重签 JWT |
| `GET /api/auth/me` | `GET /internal/end-user/auth/me` | BFF 验证 JWT → 解析 `userId`/`orgName`/`projectSlug` → 通过 Header 传给 Go |

> `me` 接口的内网调用 Header：
> ```
> X-Internal-Token: <shared-secret>
> X-End-User-Id:    <userId from JWT.sub>
> X-Org-Name:       <orgName from JWT>
> X-Project-Slug:   <projectSlug from JWT>
> ```
> Go Backend 通过 `X-Org-Name` + `X-Project-Slug` 路由到 `private_{projectSlug}` 库，通过 `X-End-User-Id` 查询用户记录。JWT 验证由 BFF 完成，Go Backend 不重复验证。

### 开发者侧（GraphQL → Go Resolver → `/internal/`）

| GraphQL Operation | Go Resolver → 内网接口 |
|-------------------|----------------------|
| `listEndUsers` | `GET /internal/end-users` |
| `createEndUser` | `POST /internal/end-users` |
| `updateEndUserStatus` | `PATCH /internal/end-users/{userId}/status` |
| `deleteEndUser` | `DELETE /internal/end-users/{userId}` |

---

## 四、Go Backend 内网接口（BFF 专用，`X-Internal-Token` 鉴权）

### 认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/internal/end-user/auth/register` | 注册：bcrypt 存密码，生成 refresh token |
| `POST` | `/internal/end-user/auth/login` | 登录：验证凭证，生成 refresh token |
| `POST` | `/internal/end-user/auth/logout` | 登出：revoke refresh token（幂等） |
| `POST` | `/internal/end-user/auth/refresh` | Token rotation |
| `GET` | `/internal/end-user/auth/me` | 查询当前用户信息（by userId from JWT） |

### 用户管理接口（供 GraphQL Resolver 调用）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/internal/end-users` | 创建用户（开发者创建，跳过自动登录） |
| `GET` | `/internal/end-users` | 分页列表 |
| `PATCH` | `/internal/end-users/{userId}/status` | 禁用/启用 |
| `DELETE` | `/internal/end-users/{userId}` | 物理删除 + revoke sessions |

> 所有内网接口 Body 均需携带 `org_name` + `project_slug` 用于 DB 路由。
> 鉴权：`X-Internal-Token: <shared-secret>` Header。

---

## 五、设计决策说明

### register vs createEndUser 的区别

| | `POST /api/auth/register`（OpenAPI） | `createEndUser`（GraphQL） |
|---|---|---|
| 调用方 | 终端用户客户端（自助） | 开发者管理侧 |
| 密码处理 | 用户自行设置 | 开发者设置初始密码 |
| 登录状态 | 注册后**自动登录**（返回 access token） | 不登录，仅创建账号 |
| 鉴权要求 | 无需 token | 需要开发者 JWT |

### Cookie path 限定

```
end_user_refresh_token Cookie
  path = /org/{orgName}/project/{projectSlug}
```

Cookie 路径绑定到具体 Project，防止不同 Project 的 refresh token 互相读取。

### `me` 接口的实现

Go Backend `/internal/end-user/auth/me` 接收 BFF 解析 JWT 后传入的 `userId`（通过 `X-End-User-Id` Header），查询 `private_{projectSlug}.users` 返回用户信息。JWT 验证由 BFF 完成，Go Backend 不重复验证。
