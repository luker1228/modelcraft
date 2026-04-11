# api-contract.md

## 1) 协议归属判断（Org GraphQL vs Project GraphQL）

### 结论
- **Profile 编辑、当前用户 User+Profile 联合查询**：归属 **Org GraphQL**（`/graphql/org/{orgName}/`）
- **注册（phone + userName + password）并自动创建 profile**：归属 **OpenAPI Auth REST**（`POST /api/auth/register`）

### 理由
1. `user/profile` 属于**组织级身份与账号域**，不是项目内模型设计能力，不应放在 Project GraphQL。  
2. 项目规则明确：**认证/会话类接口走 OpenAPI**，注册属于 auth 能力，不能放到 GraphQL。  
3. “GraphQL 优先”在该需求下体现为：**业务资料读写走 GraphQL**，但**注册链路保持 Auth REST**。

---

## 2) 最小闭环协议设计总览

| 能力 | 协议 | Endpoint | 说明 |
|---|---|---|---|
| 注册并自动建档 | OpenAPI REST | `POST /api/auth/register` | 注册成功后同事务创建 profile，写默认 nickname |
| 更新 profile | Org GraphQL | `mutation updateMyProfile` | 支持部分更新 `nickname/avatarUrl/bio` |
| 查询当前用户+profile | Org GraphQL | `query myUserProfile` | 一次查询返回 `User + Profile` |

---

## 3) GraphQL 设计（Org GraphQL）

> 建议新增文件：`api/graph/org/schema/profile.graphql`  
> 复用已存在的 `interface Error { message: String! }`（当前在 `org/schema/project.graphql` 中定义）。

```graphql
# ============================================
# Profile Error Types
# ============================================

type UserNotFound implements Error {
  message: String!
}

type ProfileNotFound implements Error {
  message: String!
}

type InvalidProfileInput implements Error {
  message: String!
  suggestion: String
}

union GetMyUserProfileError = UserNotFound | ProfileNotFound
union UpdateMyProfileError = ProfileNotFound | InvalidProfileInput

# ============================================
# Profile Domain Types
# ============================================

enum UserStatus {
  REGISTERED
  ACTIVE
  SUSPENDED
}

type Profile implements Node {
  id: ID!
  userId: ID!
  nickname: String!
  avatarUrl: String
  bio: String
  createdAt: String!
  updatedAt: String!
}

type User implements Node {
  id: ID!
  phone: String!
  userName: String!
  status: UserStatus!
  createdAt: String!
  updatedAt: String!
  profile: Profile!
}

# ============================================
# Payload Types
# ============================================

type GetMyUserProfilePayload {
  user: User
  error: GetMyUserProfileError
}

type UpdateMyProfilePayload {
  profile: Profile
  error: UpdateMyProfileError
}

# ============================================
# Input Types
# ============================================

input UpdateMyProfileInput {
  nickname: String
  avatarUrl: String
  bio: String
}

# ============================================
# Queries & Mutations
# ============================================

extend type Query {
  myUserProfile: GetMyUserProfilePayload! @hasPermission(action: "user:read")
}

extend type Mutation {
  updateMyProfile(input: UpdateMyProfileInput!): UpdateMyProfilePayload! @hasPermission(action: "profile:update")
}
```

### GraphQL 语义约束
- `myUserProfile` 成功时：`user != nil && error == nil`；失败时互斥。
- `updateMyProfile` 为 **PATCH 语义**（只更新传入字段）。
- `UpdateMyProfileInput` 至少一个字段非空，否则返回 `InvalidProfileInput`。
- 头像相关能力当前可使用 mock 数据，暂不引入头像专用错误类型。

---

## 4) OpenAPI 补充（注册链路，Auth 域）

> 变更文件：`api/openapi/auth.yaml`（在现有 `/api/auth/register` 上增强）  
> 注册仍 `security: []`，并明确“创建 user + profile 为原子操作”。

```yaml
paths:
  /api/auth/register:
    post:
      operationId: register
      summary: Register user and initialize profile atomically
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "auth.yaml#/schemas/RegisterRequest"
      responses:
        "201":
          description: Registration successful (user + profile created)
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/RegisterResponse"
        "400":
          description: Invalid input
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/AuthInvalidInputError"
        "409":
          description: phone or userName already exists
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: "auth.yaml#/schemas/PhoneAlreadyRegisteredError"
                  - $ref: "auth.yaml#/schemas/UserNameAlreadyRegisteredError"
        "500":
          description: Profile init failed or internal error
          content:
            application/json:
              schema:
                $ref: "common.yaml#/schemas/SystemError"

schemas:
  RegisterRequest:
    type: object
    required:
      - phone
      - userName
      - password
    properties:
      phone:
        type: string
        pattern: "^1[3-9]\\d{9}$"
      userName:
        type: string
        minLength: 3
        maxLength: 32
        pattern: "^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$"
      password:
        type: string
        minLength: 8

  RegisterProfileSnapshot:
    type: object
    required:
      - id
      - userId
      - nickname
    properties:
      id:
        type: string
      userId:
        type: string
      nickname:
        type: string
        description: "default generated nickname, e.g. user_A1B2C3"
      avatarUrl:
        type: string
      bio:
        type: string

  RegisterResponse:
    allOf:
      - $ref: "common.yaml#/schemas/BaseResponse"
      - type: object
        required:
          - userId
          - orgName
          - profile
        properties:
          userId:
            type: string
          orgName:
            type: string
          profile:
            $ref: "auth.yaml#/schemas/RegisterProfileSnapshot"

  UserNameAlreadyRegisteredError:
    type: object
    required:
      - requestId
      - error
    properties:
      requestId:
        type: string
      error:
        type: object
        required:
          - code
          - message
        properties:
          code:
            type: string
            enum:
              - CONFLICT.USER
          message:
            type: string
```

---

## 5) 错误模型（Error Interface + Union + 错误码）

## GraphQL（类型化错误）
- `GetMyUserProfileError = UserNotFound | ProfileNotFound`
- `UpdateMyProfileError = ProfileNotFound | InvalidProfileInput`

## OpenAPI（HTTP 状态码）
- `400` → `PARAM_INVALID.AUTH`（注册参数非法）
- `409` → `CONFLICT.USER`（phone/userName 冲突）
- `500` → `SYSTEM_ERROR`（含 profile 初始化失败）

> 建议新增/对齐 bizerrors（供实现层映射）：
- `NOT_FOUND.PROFILE`
- `PARAM_INVALID.PROFILE`
- `CONFLICT.PROFILE`（后续若增加显式创建 profile 场景）

---

## 6) 数据约束说明（必须）

1. **`user.phone` 唯一**（全局唯一）  
2. **`user.userName` 唯一**（全局唯一）  
3. **`profile.userId` 唯一** + 外键关联 `user.id`，保证 1:1  
4. 注册成功条件：`user` 与 `profile` 必须同时落库（建议单事务）  
5. 默认昵称规则：`user_` + 6位随机大写字母数字（示例：`user_A1B2C3`）  
6. API 字段使用 camelCase：`avatarUrl`（DB 列可为 `avatar_url`）

---

## 7) 闭环调用示例（前端/BFF 视角）

1. `POST /api/auth/register`  
   入参：`phone + userName + password`  
   出参：`userId + orgName + profile(default nickname)`  

2. 登录成功后（已有认证态），调用 GraphQL：
```graphql
query MyUserProfile {
  myUserProfile {
    user {
      id
      phone
      userName
      status
      profile {
        id
        nickname
        avatarUrl
        bio
      }
    }
    error {
      __typename
      ... on UserNotFound { message }
      ... on ProfileNotFound { message }
    }
  }
}
```

3. 编辑资料：
```graphql
mutation UpdateMyProfile {
  updateMyProfile(input: {
    nickname: "new_nick",
    avatarUrl: "mock://avatar/default-2.png",
    bio: "hello"
  }) {
    profile {
      id
      nickname
      avatarUrl
      bio
      updatedAt
    }
    error {
      __typename
      ... on InvalidProfileInput { message suggestion }
      ... on ProfileNotFound { message }
    }
  }
}
```

---

## 8) 需用户确认项

1. **强一致策略**：是否明确要求“注册必须在单事务内创建 profile，失败即整体回滚”（建议：是）。  
2. **昵称与简介长度**：`nickname` / `bio` 的最大长度（建议：32 / 256）。  
3. **头像处理策略**：当前头像能力暂不细化协议约束，代码层可使用 mock 实现。  
4. **Profile 缺失策略（已确认）**：`myUserProfile` 遇到 profile 缺失时，返回 `ProfileNotFound`。  
5. **兼容策略（已确认）**：保留现有 `me` 查询，并新增字段化 payload 查询 `myUserProfile`。