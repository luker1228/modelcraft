# 双 UI 架构 — 后端 API 契约

> 本文档是 `ai-metadata/prd/dual-ui-architecture/` 的后端契约产物。
> 覆盖：EndUser 认证（OpenAPI REST）+ Workspace Project 列表（GraphQL Org Schema）。
>
> **参考设计原则**：见 `.agents/agents/backend-api.md`

---

## 变更范围

### 1. OpenAPI — auth.yaml（Tenant Auth 路径迁移）

**现状：** 租户认证接口路径为 `/api/auth/*`，与 end-user 的 `/api/end-user/auth/*` 命名不对称，且无法区分未来的 platform 体系。

**新设计：** 所有租户认证接口路径统一迁移至 `/api/tenant/auth/*`，保持与 end-user 体系对称。

| 旧路径 | 新路径 |
|--------|--------|
| `POST /api/auth/login` | `POST /api/tenant/auth/login` |
| `POST /api/auth/register` | `POST /api/tenant/auth/register` |
| `POST /api/auth/refresh` | `POST /api/tenant/auth/refresh` |
| `POST /api/auth/logout` | `POST /api/tenant/auth/logout` |

**Request / Response 结构不变**，仅路径变更。

**同步需要更新的地方：**

1. `api/openapi/auth.yaml` — 4 个 path 键重命名
2. `chi_setup.go` `conditionalAuthMiddleware` publicPaths 白名单路径同步更新
3. 前端 BFF 代理路由 `/api/bff/auth/*` → 目标地址从 `/api/auth/*` 改为 `/api/tenant/auth/*`
4. 前端 `middleware.ts` — `DEV_PUBLIC_PATHS` 对应的公开路径更新（`/tenant/login`）

---

### 2. OpenAPI — end-user-auth.yaml

**现状与问题：**
- 现有 `/api/end-user/auth/login` 接受可选 `projectSlug`，返回 Project Token（旧的两阶段设计）
- 现有 `/api/end-user/auth/select-project` 做 exchange，签发含 `projectSlug` 的 Project Token
- 与新设计不符：登录后统一跳 Workspace，不做 exchange，不存在 Project Token

**新设计：**
- 登录接口删除 `projectSlug` 参数
- 登录成功响应中的 `projects` 字段仍然保留（供前端 Workspace 页面渲染 Project 列表，减少一次额外查询）
- Token 统一为 `aud=end-user`，不区分 Org Token / Project Token
- **删除** `/api/end-user/auth/select-project` 端点

---

### 变更后的 end-user-auth.yaml 契约（差异说明）

#### `EndUserLoginRequest` — 变更

```yaml
EndUserLoginRequest:
  type: object
  required:
    - orgName
    - username
    - password
  properties:
    orgName:
      type: string
      description: Organization slug（Org 级账号池）
      example: acme
    username:
      type: string
      example: alice
    password:
      type: string
      example: s3cr3t
  # REMOVED: projectSlug（不再在登录时绑定 Project）
```

#### `EndUserLoginResponse` — 变更

```yaml
EndUserLoginResponse:
  allOf:
    - $ref: "common.yaml#/schemas/BaseResponse"
    - type: object
      required:
        - accessToken
        - expiresIn
        - aud
      properties:
        accessToken:
          type: string
          description: |
            JWT access token，aud=end-user。
            Payload 包含：sub（endUserId）、org_name、aud="end-user"、iss="mc-platform"
        expiresIn:
          type: integer
          description: Access token TTL（秒）
        aud:
          type: string
          enum: [end-user]
          description: 标识 Token 类型，固定为 end-user
        projects:
          type: array
          description: |
            该 EndUser 有权访问的 Project 列表（按 end_user_role_users 聚合去重）。
            列表为空时表示无可访问 Project，前端在 Workspace 展示空状态。
          items:
            $ref: "end-user-auth.yaml#/schemas/EndUserAccessibleProject"
  # Refresh token 以 httpOnly cookie 写入，不在 body 中返回
```

#### `EndUserAccessibleProject` — 不变

```yaml
EndUserAccessibleProject:
  type: object
  required:
    - slug
    - title
  properties:
    slug:
      type: string
      example: sales-system
    title:
      type: string
      example: 销售管理系统
    description:
      type: string
      description: Project 简介（可选，供 Workspace 卡片展示）
```

#### `/api/end-user/auth/select-project` — **删除**

此端点与新设计不符（不存在 Project Token exchange），从 OpenAPI 规范中删除。

#### `/api/end-user/auth/refresh` — 不变

刷新 Token 只更新 `aud=end-user` Token，不引入 Project 维度。

#### `/api/end-user/auth/me` — 不变

返回 EndUser 身份信息，用于前端顶部栏展示用户名。

---

### 3. GraphQL — Project Schema 新增查询

**背景：**
Workspace 页面需要在用户进入后动态获取最新的可访问 Project 列表（登录时返回的列表有时效性，用户权限变更后需要刷新）。

**新增 Query：`listEndUserAccessibleProjects`**

归属：`api/graph/project/schema/project.graphql`（追加到现有文件，不新建 end_user.graphql）

```graphql
# ============================================
# EndUser Workspace — 可访问 Project 列表
# ============================================

# 用于 Workspace 主页展示 Project 卡片列表。
# 由 EndUser 身份调用，不需要租户端权限。
# 后端按 end_user_role_users 聚合去重，只返回当前 EndUser 有 Role 的 Project。

type EndUserAccessibleProject {
  slug:        String!
  title:       String!
  description: String    # 可选，Workspace 卡片展示用
}

type ListEndUserAccessibleProjectsPayload {
  projects: [EndUserAccessibleProject!]
  error:    ListEndUserAccessibleProjectsError
}

union ListEndUserAccessibleProjectsError = InvalidInput

extend type Query {
  # 不需要 @hasPermission（EndUser 自己查自己的 Project 列表，RBAC 在后端实现层校验）。
  listEndUserAccessibleProjects: ListEndUserAccessibleProjectsPayload!
}
```

**Endpoint：** `POST /graphql/org/{orgName}/project/{projectSlug}/`（Project GraphQL）

**Token 要求：** 沿用 Project GraphQL 现有鉴权，后端实现层通过 EndUser 身份（Token 中的 `user_id` + `org_name`）查询 `end_user_role_users` 聚合去重返回列表。

---

### 4. GraphQL — Runtime 数据操作（已有，需确认）

数据操作页（`/data`）使用 Runtime GraphQL，通过动态 Schema 对运行时数据进行 CRUD。

**Endpoint：** `POST /org/{orgName}/project/{projectSlug}/db/{dbSlug}/model/{modelSlug}`

此端点已存在，不需要新建契约。需要确认的是：
- EndUser 调用 Runtime GraphQL 时，后端校验 `aud=end-user` 且 `orgName` 匹配
- RBAC `end_user_role_users` 中有该 EndUser 对该 Project 的 Role，才允许访问

---

## 错误码变更

| 场景 | HTTP 状态 | error.code | 说明 |
|------|-----------|------------|------|
| 登录成功但无可访问 Project | 200 | — | 正常响应，`projects` 为空数组 |
| 账号已禁用 | 401 | `AUTHENTICATION_FAILED` | 复用现有错误码 |
| Org 不存在 | 404 | `NOT_FOUND.ORGANIZATION` | 复用现有错误码 |
| aud 不匹配（跨端访问）| 403 | `UNAUTHORIZED` | 中间件层返回，不到业务层 |

---

## 领域层接口（Application Layer）

### EndUserAuthAppService — 变更

```go
// LoginEndUser — 变更：移除 projectSlug 参数，始终返回 AccessibleProjects
type LoginEndUserCommand struct {
    OrgName  string
    Username string
    Password string
    // REMOVED: ProjectSlug string
}

type LoginEndUserResult struct {
    EndUserID         string
    AccessToken       string  // aud=end-user JWT
    ExpiresIn         int
    AccessibleProjects []AccessibleProject  // 聚合自 end_user_role_users
}

type AccessibleProject struct {
    Slug        string
    Title       string
    Description string
}
```

### EndUserAuthAppService — 删除

```go
// REMOVED: SelectProjectCommand / SelectProjectResult
// 不再做 Project Token exchange
```

---

## 契约完备性检查

**OpenAPI**
- [x] 登录接口去掉 `projectSlug`，响应含 `projects` 列表和 `aud` 字段
- [x] 删除 `/select-project` 端点
- [x] 其余接口（refresh、logout、me）不变
- [x] 错误码与 `bizerrors` 对应

**GraphQL**
- [x] 新增 `listEndUserAccessibleProjects`（Workspace 动态刷新用）归属 `project.graphql`，不新建独立文件
- [x] Error Union 遵循项目规范（实现 `Error` interface）
- [x] Payload 遵循 `data + error` 互斥模式
- [x] 无需 `@hasPermission`（EndUser 自查，后端实现层校验）

**待确认**
- [ ] `listEndUserAccessibleProjects` 是否需要支持分页（v1 Project 数量预计不多，先不分页）
- [ ] Runtime GraphQL 的 `aud=end-user` Token 校验是否需要中间件层变更
- [ ] `end-user-auth.yaml` 的路径前缀是否随新路由变化（现在是 `/api/end-user/auth/...`，URL 本身不变，只有前端 BFF 路由变化）
