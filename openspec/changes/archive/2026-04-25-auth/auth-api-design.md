# Auth 模块 API 协议设计

> 基于 [auth.md](./auth.md)、[auth-login.md](./auth-login.md)、[auth-register.md](./auth-register.md) PRD 设计。

## 一、协议分类决策

按项目规则，Auth 操作走 **OpenAPI REST**（`/api/auth/`），不走 GraphQL。

---

## 二、端点变更总览

| 端点 | 变更类型 | 说明 |
|------|----------|------|
| `POST /api/auth/register` | **新增** | 手机号 + 用户名 + 密码注册 |
| `POST /api/auth/login` | **改造** | Casdoor OAuth code 交换 → 手机号/用户名 + 密码直接认证 |
| `GET /api/auth/login-url` | **删除** | Casdoor OAuth 跳转 URL，不再需要 |
| `POST /api/auth/refresh` | **保留** | Refresh token rotation，逻辑不变 |
| `POST /api/auth/logout` | **保留** | 吊销 refresh token，逻辑不变 |

---

## 三、OpenAPI Schema 设计

### 3.1 Paths

```yaml
paths:
  /api/auth/register:
    post:
      operationId: register
      summary: Register a new user with phone number, userName and password
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
          description: Registration successful
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/RegisterResponse"
        "400":
          description: Invalid input (phone format, password too short, userName format)
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/AuthInvalidInputError"
        "409":
          description: Phone number or userName already registered
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/UserConflictError"
        "500":
          description: Server error
          content:
            application/json:
              schema:
                $ref: "common.yaml#/schemas/SystemError"

  /api/auth/login:
    post:
      operationId: login
      summary: Login with phone number or userName and password
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "auth.yaml#/schemas/LoginRequest"
      responses:
        "200":
          description: Login successful
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/LoginResponse"
        "400":
          description: Invalid input (phone format, missing fields)
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/AuthInvalidInputError"
        "401":
          description: Authentication failed (user not found or wrong password)
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/AuthenticationError"
        "500":
          description: Server error
          content:
            application/json:
              schema:
                $ref: "common.yaml#/schemas/SystemError"

  /api/auth/logout:
    post:
      operationId: logout
      summary: Logout (revoke refresh token)
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "auth.yaml#/schemas/RefreshTokenRequest"
      responses:
        "204":
          description: Logout successful

  /api/auth/refresh:
    post:
      operationId: refreshToken
      summary: Refresh tokens using refresh token (rotation)
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "auth.yaml#/schemas/RefreshTokenRequest"
      responses:
        "200":
          description: Token refresh successful
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/RefreshResponse"
        "401":
          description: Invalid, expired, or reused refresh token
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/AuthenticationError"
        "500":
          description: Server error
          content:
            application/json:
              schema:
                $ref: "common.yaml#/schemas/SystemError"
```

### 3.2 Error Schemas

```yaml
schemas:
  AuthInvalidInputError:
    type: object
    required: [requestId, error]
    properties:
      requestId:
        type: string
      error:
        type: object
        required: [code, message]
        properties:
          code:
            type: string
            enum: [PARAM_INVALID.AUTH]
          message:
            type: string
            description: |
              Possible messages:
              - "invalid phone number format: must be 11-digit mainland China number"
              - "password must be at least 8 characters"
              - "userName must be 3-32 chars, letters/digits/underscore/hyphen, not starting with digit"
              - "userName is a reserved word"

  UserConflictError:
    type: object
    required: [requestId, error]
    properties:
      requestId:
        type: string
      error:
        type: object
        required: [code, message]
        properties:
          code:
            type: string
            enum: [CONFLICT.USER]
          message:
            type: string
            description: |
              Possible messages:
              - "phone number already registered: 138****1234"
              - "userName already taken: alice"

  AuthenticationError:
    type: object
    required: [requestId, error]
    properties:
      requestId:
        type: string
      error:
        type: object
        required: [code, message]
        properties:
          code:
            type: string
            enum: [AUTHENTICATION_FAILED]
          message:
            type: string
            description: |
              Possible messages:
              - "phone number not found"
              - "userName not found"
              - "incorrect password"
              - "refresh token not found"
              - "token reuse detected"
              - "refresh token expired"
```

### 3.3 Request Schemas

```yaml
  RegisterRequest:
    type: object
    required: [phone, userName, password]
    properties:
      phone:
        type: string
        description: "11-digit mainland China phone number"
        pattern: "^1[3-9]\\d{9}$"
        example: "13800138000"
      userName:
        type: string
        description: |
          Global unique username (immutable after registration).
          Rules: letters/digits/underscore/hyphen, not starting with digit, length 3–32,
          must not be a reserved word.
        pattern: "^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$"
        minLength: 3
        maxLength: 32
        example: "alice"
      password:
        type: string
        description: "Password (minimum 8 characters)"
        minLength: 8
        example: "mypassword123"

  LoginRequest:
    type: object
    required: [identifier, identifierType, password]
    properties:
      identifier:
        type: string
        description: "Login identifier — phone number or userName, determined by identifierType"
        example: "13800138000"
      identifierType:
        type: string
        enum: [PHONE, USERNAME]
        description: |
          - PHONE: identifier is an 11-digit mainland China phone number
          - USERNAME: identifier is the user's userName
        example: "PHONE"
      password:
        type: string
        description: "User password"
        example: "mypassword123"

  RefreshTokenRequest:
    type: object
    required: [refreshToken]
    properties:
      refreshToken:
        type: string
        description: Refresh token obtained during login
```

### 3.4 Response Schemas

```yaml
  RegisterResponse:
    allOf:
      - $ref: "common.yaml#/schemas/BaseResponse"
      - type: object
        required: [userId, orgName]
        properties:
          userId:
            type: string
            description: "Created user UUID"
          orgName:
            type: string
            description: "Auto-created initial Org name (serves as the Org's unique identifier)"
            example: "alice-org-a3f2"

  LoginResponse:
    allOf:
      - $ref: "common.yaml#/schemas/BaseResponse"
      - type: object
        required: [userId, userName, orgName, refreshToken, expiresAt]
        properties:
          userId:
            type: string
            description: ModelCraft internal user UUID
          userName:
            type: string
            description: User's unique username
            example: "alice"
          orgName:
            type: string
            description: "User's initial Org name (serves as the Org's unique identifier)"
            example: "alice-org-a3f2"
          refreshToken:
            type: string
            description: Opaque refresh token (store securely on client)
          expiresAt:
            type: string
            format: date-time
            description: Refresh token expiration time (7 days from now)

  RefreshResponse:
    allOf:
      - $ref: "common.yaml#/schemas/BaseResponse"
      - type: object
        required: [userId, refreshToken, expiresAt]
        properties:
          userId:
            type: string
            description: ModelCraft internal user UUID
          refreshToken:
            type: string
            description: New opaque refresh token (replaces the old one)
          expiresAt:
            type: string
            format: date-time
            description: New refresh token expiration time
```

### 3.5 openapi-root.yaml 变更

```yaml
paths:
  /api/auth/register:
    $ref: "auth.yaml#/paths/~1api~1auth~1register"
  /api/auth/login:
    $ref: "auth.yaml#/paths/~1api~1auth~1login"
  # /api/auth/login-url 移除（Casdoor 不再需要）
  /api/auth/logout:
    $ref: "auth.yaml#/paths/~1api~1auth~1logout"
  /api/auth/refresh:
    $ref: "auth.yaml#/paths/~1api~1auth~1refresh"

components:
  schemas:
    AuthInvalidInputError:
      $ref: "auth.yaml#/schemas/AuthInvalidInputError"
    UserConflictError:
      $ref: "auth.yaml#/schemas/UserConflictError"
    AuthenticationError:
      $ref: "auth.yaml#/schemas/AuthenticationError"
    RegisterRequest:
      $ref: "auth.yaml#/schemas/RegisterRequest"
    RegisterResponse:
      $ref: "auth.yaml#/schemas/RegisterResponse"
    LoginRequest:
      $ref: "auth.yaml#/schemas/LoginRequest"
    LoginResponse:
      $ref: "auth.yaml#/schemas/LoginResponse"
    RefreshTokenRequest:
      $ref: "auth.yaml#/schemas/RefreshTokenRequest"
    RefreshResponse:
      $ref: "auth.yaml#/schemas/RefreshResponse"
```

---

## 四、错误码体系

### 4.1 新增错误码

需添加到 `pkg/bizerrors/common_errors.go`：

```go
var (
    // AuthenticationFailed 登录失败（用户不存在、密码错误）
    // 错误码统一为 AUTHENTICATION_FAILED，通过 message 区分具体原因
    AuthenticationFailed = ErrorDefinition{
        Code:      "AUTHENTICATION_FAILED",
        EnMessage: "Authentication failed: {0}",
        ZhMessage: "认证失败: {0}",
    }

    // AuthParamInvalid 注册/登录参数校验失败（手机号格式、密码长度、userName 格式）
    AuthParamInvalid = ErrorDefinition{
        Code:      ErrorTypeParamInvalid + ".AUTH",
        EnMessage: "Invalid auth parameter: {0}",
        ZhMessage: "认证参数无效: {0}",
    }

    // UserConflict 注册时手机号或用户名已存在
    // 复用已有的 CONFLICT.USER 错误码
)
```

### 4.2 错误码映射表

| 错误码 | HTTP | 场景 |
|--------|------|------|
| `PARAM_INVALID.AUTH` | 400 | 手机号格式不合法（非11位）、密码少于8位、userName 格式不合法、userName 是保留字 |
| `CONFLICT.USER` | 409 | 注册时手机号已被注册、userName 已被占用 |
| `AUTHENTICATION_FAILED` | 401 | 登录时用户不存在、密码错误 |
| `UNAUTHORIZED` | 401 | Refresh token 无效/过期/被重用（已有） |
| `SYSTEM_ERROR` | 500 | 密码哈希失败、数据库异常等 |

### 4.3 安全考量

- 登录失败时，PRD 要求区分"手机号/用户名不存在"和"密码错误"两种提示。**code 统一为 `AUTHENTICATION_FAILED`**（不通过错误码区分），仅通过 message 携带具体原因。
- 注册时手机号或 userName 已存在均返回 409，PRD 明确要求前端区分展示。

---

## 五、userName 保留字黑名单

`UserName` 值对象校验时拒绝以下保留字（**前后端共同维护**，大小写不敏感）：

```
# 系统身份
admin, administrator, root, system, superuser

# 平台品牌
modelcraft, modelcraft-admin

# API 路径关键字
api, auth, login, logout, register, refresh, oauth, callback

# 资源名词（与 URL 路径冲突）
user, users, org, orgs, project, projects, model, models,
cluster, clusters, field, fields, group, groups, schema, schemas,
dashboard, settings, profile, me, self

# 通用保留
null, undefined, true, false, none, anonymous, guest,
public, private, static, assets, upload, uploads,
test, demo, example, sample,
support, help, info, about, contact, home, index

# 常见攻击向量
www, ftp, mail, smtp, pop3, imap
```

> **维护规则**：
> - 后端在 `internal/domain/user/username_reserved.go` 中维护 Go 版本（`map[string]struct{}`，存 lowercase）
> - 前端在 `src/lib/reserved-usernames.ts` 中维护 TypeScript 版本（`Set<string>`，存 lowercase）
> - 两端以 **后端为准**；新增保留字时同步更新两端。

---

## 六、领域模型变更

### 6.1 PhoneNumber 值对象（新增）

路径：`internal/domain/user/phone_number.go`

```go
package user

import (
    "fmt"
    "regexp"
)

// PhoneNumber 值对象 — 中国大陆 11 位手机号
type PhoneNumber struct {
    Value string
}

var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

// NewPhoneNumber 创建并校验手机号
func NewPhoneNumber(value string) (PhoneNumber, error) {
    if !phoneRegexp.MatchString(value) {
        return PhoneNumber{}, fmt.Errorf(
            "invalid phone number format: must be 11-digit mainland China number",
        )
    }
    return PhoneNumber{Value: value}, nil
}

// Masked 返回脱敏手机号 (138****1234)
func (p PhoneNumber) Masked() string {
    if len(p.Value) != 11 {
        return p.Value
    }
    return p.Value[:3] + "****" + p.Value[7:]
}
```

### 6.2 UserName 值对象（新增）

路径：`internal/domain/user/username.go`

```go
package user

import (
    "fmt"
    "regexp"
    "strings"
)

// UserName 值对象 — 全局唯一，注册后不可修改
type UserName struct {
    Value string
}

var userNameRegexp = regexp.MustCompile(`^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$`)

// NewUserName 创建并校验用户名
func NewUserName(value string) (UserName, error) {
    if !userNameRegexp.MatchString(value) {
        return UserName{}, fmt.Errorf(
            "userName must be 3-32 chars, letters/digits/underscore/hyphen, not starting with digit",
        )
    }
    if isReservedUserName(strings.ToLower(value)) {
        return UserName{}, fmt.Errorf("userName is a reserved word: %s", value)
    }
    return UserName{Value: value}, nil
}
```

### 6.3 User 实体改造

路径：`internal/domain/user/user.go`

```go
type User struct {
    ID           string
    Phone        PhoneNumber
    UserName     UserName     // 全局唯一，注册后不可修改
    DisplayName  string       // 默认等于 userName
    PasswordHash string       // bcrypt 哈希
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// NewUser 注册时创建新用户
func NewUser(id string, phone PhoneNumber, userName UserName, passwordHash string) (*User, error) {
    now := time.Now()
    user := &User{
        ID:           id,
        Phone:        phone,
        UserName:     userName,
        DisplayName:  userName.Value, // 默认等于 userName
        PasswordHash: passwordHash,
        CreatedAt:    now,
        UpdatedAt:    now,
    }
    if err := user.Validate(); err != nil {
        return nil, err
    }
    return user, nil
}
```

### 6.4 密码强度校验（领域规则）

路径：`internal/domain/auth/password.go`

```go
package auth

import "fmt"

// ValidatePasswordStrength 密码强度校验
func ValidatePasswordStrength(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }
    return nil
}

// PasswordHasher 密码哈希接口（Domain 层定义，Infrastructure 层实现）
type PasswordHasher interface {
    Hash(password string) (string, error)
    Verify(password, hash string) error
}
```

### 6.5 UserRepository 接口变更

路径：`internal/domain/user/repository.go`

```go
type UserRepository interface {
    // Create 创建用户
    Create(ctx context.Context, user *User) error

    // GetByID 根据内部 UUID 获取用户
    GetByID(ctx context.Context, id string) (*User, error)

    // GetByPhone 根据手机号获取用户（PHONE 登录时使用）
    // 不存在时返回 (nil, NotFoundError)
    GetByPhone(ctx context.Context, phone string) (*User, error)

    // GetByUserName 根据用户名获取用户（USERNAME 登录时使用）
    // 不存在时返回 (nil, NotFoundError)
    GetByUserName(ctx context.Context, userName string) (*User, error)

    // ExistsByPhone 检查手机号是否已注册
    ExistsByPhone(ctx context.Context, phone string) (bool, error)

    // ExistsByUserName 检查用户名是否已占用
    ExistsByUserName(ctx context.Context, userName string) (bool, error)
}
```

### 6.6 Application 层 Commands

路径：`internal/app/auth/commands.go`

```go
// RegisterCommand 注册命令
type RegisterCommand struct {
    Phone    string // 11 位手机号
    UserName string // 全局唯一用户名
    Password string // 明文密码（>= 8 位）
}

// RegisterResult 注册结果
type RegisterResult struct {
    UserID  string
    OrgName string // 自动创建的初始 Org 名称（即 Org 的唯一标识）
}

// LoginCommand 登录命令
type LoginCommand struct {
    Identifier     string         // 手机号 或 用户名
    IdentifierType IdentifierType // PHONE | USERNAME
    Password       string         // 明文密码
}

// IdentifierType 登录标识符类型
type IdentifierType string

const (
    IdentifierTypePhone    IdentifierType = "PHONE"
    IdentifierTypeUserName IdentifierType = "USERNAME"
)

// LoginResult 登录结果
type LoginResult struct {
    UserID       string
    UserName     string
    OrgName      string    // 用户归属 Org 的名称（即 Org 的唯一标识）
    RefreshToken string    // 明文 refresh token（仅此次返回）
    ExpiresAt    time.Time // Refresh token 过期时间
}
```

---

## 七、请求/响应示例

### 7.1 注册成功

```
POST /api/auth/register
Content-Type: application/json

{
  "phone": "13800138000",
  "userName": "alice",
  "password": "mypassword123"
}

→ 201 Created
{
  "requestId": "req_550e8400e29b41d4",
  "userId": "01904d3a-7b6c-7f00-8000-000000000001",
  "orgName": "alice-org-a3f2"
}
```

### 7.2 注册 — 手机号已存在

```
→ 409 Conflict
{
  "requestId": "req_550e8400e29b41d5",
  "error": {
    "code": "CONFLICT.USER",
    "message": "phone number already registered: 138****0000"
  }
}
```

### 7.3 注册 — 用户名已占用

```
→ 409 Conflict
{
  "requestId": "req_550e8400e29b41d5",
  "error": {
    "code": "CONFLICT.USER",
    "message": "userName already taken: alice"
  }
}
```

### 7.4 注册 — 参数校验失败

```
→ 400 Bad Request
{
  "requestId": "req_550e8400e29b41d6",
  "error": {
    "code": "PARAM_INVALID.AUTH",
    "message": "userName is a reserved word: admin"
  }
}
```

### 7.5 登录成功（手机号）

```
POST /api/auth/login
Content-Type: application/json

{
  "identifier": "13800138000",
  "identifierType": "PHONE",
  "password": "mypassword123"
}

→ 200 OK
{
  "requestId": "req_550e8400e29b41d7",
  "userId": "01904d3a-7b6c-7f00-8000-000000000001",
  "userName": "alice",
  "orgName": "alice-org-a3f2",
  "refreshToken": "a1b2c3d4e5f6...64chars",
  "expiresAt": "2026-04-16T10:30:00Z"
}
```

### 7.6 登录成功（用户名）

```
POST /api/auth/login
Content-Type: application/json

{
  "identifier": "alice",
  "identifierType": "USERNAME",
  "password": "mypassword123"
}

→ 200 OK
{
  "requestId": "req_550e8400e29b41d7",
  "userId": "01904d3a-7b6c-7f00-8000-000000000001",
  "userName": "alice",
  "orgName": "alice-org-a3f2",
  "refreshToken": "a1b2c3d4e5f6...64chars",
  "expiresAt": "2026-04-16T10:30:00Z"
}
```

### 7.7 登录 — 用户不存在

```
→ 401 Unauthorized
{
  "requestId": "req_550e8400e29b41d8",
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "phone number not found"
  }
}
```

### 7.8 登录 — 密码错误

```
→ 401 Unauthorized
{
  "requestId": "req_550e8400e29b41d9",
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "incorrect password"
  }
}
```

### 7.9 Token 刷新

```
POST /api/auth/refresh
Content-Type: application/json

{
  "refreshToken": "a1b2c3d4e5f6...64chars"
}

→ 200 OK
{
  "requestId": "req_550e8400e29b41da",
  "userId": "01904d3a-7b6c-7f00-8000-000000000001",
  "refreshToken": "new_token...64chars",
  "expiresAt": "2026-04-16T10:30:00Z"
}
```

### 7.10 登出

```
POST /api/auth/logout
Content-Type: application/json

{
  "refreshToken": "a1b2c3d4e5f6...64chars"
}

→ 204 No Content
```

---

## 八、关键设计决策

### 8.1 LoginResponse 不直接返回 JWT

与现有 token 体系保持一致：登录只返回 refresh token，客户端通过 `/api/auth/refresh` 获取 JWT。理由：
- 保持 refresh token rotation 机制
- JWT 短生命周期（~1小时），refresh token 长生命周期（7天）
- 前端已有基于 refresh 的 token 管理逻辑

### 8.2 Org 以 name 作为唯一标识（无暴露 ID）

Org 聚合根以 `name` 作为全局唯一标识，不对外暴露数据库 ID。设计意图：
- `orgName` 直接用于 URL 路径（`/graphql/org/{orgName}/`），无需额外映射
- 避免客户端持有内部 ID，符合 Org 作为"命名空间"的语义

### 8.3 初始 Org 命名规则

注册时 `RegistrationService` 自动创建初始 Org，`orgName` 格式为 `{userName}-org-{4位随机hex}`，例如 `alice-org-a3f2`。

### 8.4 密码哈希使用 bcrypt

bcrypt 是纯计算（无 IO），可在 Domain 层直接调用，不违反 Domain 纯净性原则。但为了可测试性，通过 `PasswordHasher` 接口在 Domain 层定义，Infrastructure 层实现 bcrypt。

### 8.5 Casdoor 过渡策略

- `external_id` 列暂时保留但去除唯一约束和 NOT NULL
- `GET /api/auth/login-url` 端点直接移除
- `POST /api/auth/login` 请求体从 `{code, state}` 改为 `{identifier, identifierType, password}`（破坏性变更，与 PRD 替换目标一致）
- 现有 refresh / logout / verify-token 逻辑完全复用

### 8.6 错误信息国际化

错误 message 字段支持中英文（通过 `bizerrors.NewErrorFromContext` 自动从 ctx 提取语言偏好），确保前端可直接展示。

---

## 九、设计检查清单

**OpenAPI**：
- [x] 成功响应通过 `allOf` 继承 `BaseResponse`
- [x] 错误 schema 包含 `requestId` + `error.code` + `error.message`
- [x] 错误码与 `bizerrors` 定义一致
- [x] HTTP 状态码正确对应（400/401/409/500）
- [x] auth 接口设置 `security: []`（无需认证）
- [x] 字段命名使用 `camelCase`

**领域模型**：
- [x] `PhoneNumber` 值对象封装手机号校验规则
- [x] `UserName` 值对象封装用户名校验规则（含保留字校验）
- [x] 密码强度校验为领域规则（`>= 8` 位）
- [x] User 实体不变量完整（ID + Phone + UserName + DisplayName + PasswordHash）
- [x] `displayName` 默认等于 `userName`（来自领域模型）
- [x] UserRepository 新增 `GetByUserName` / `ExistsByUserName`
- [x] LoginCommand 支持 `IdentifierType` 枚举（PHONE / USERNAME）
- [x] LoginResult 包含 `userName` 和 `orgName`
- [x] RegisterResult 包含 `orgName`（Org 的唯一标识）

**向后兼容**：
- [x] 移除 Casdoor 专属端点 `/api/auth/login-url`
- [x] `LoginRequest` 从 Casdoor OAuth 改为双标识符+密码（破坏性变更，PRD 预期）
- [x] `refresh` / `logout` 保持不变
- [x] 现有 refresh token 体系完全复用
