# Auth 模块 API 协议设计

> 基于 [auth.md](./auth.md)、[auth-login.md](./auth-login.md)、[auth-register.md](./auth-register.md) PRD 设计。

## 一、协议分类决策

按项目规则，Auth 操作走 **OpenAPI REST**（`/api/auth/`），不走 GraphQL。

---

## 二、端点变更总览

| 端点 | 变更类型 | 说明 |
|------|----------|------|
| `POST /api/auth/register` | **新增** | 手机号 + 密码注册 |
| `POST /api/auth/login` | **改造** | Casdoor OAuth code 交换 → 手机号 + 密码直接认证 |
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
      summary: Register a new user with phone number and password
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
          description: Invalid input (phone format, password too short)
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/AuthInvalidInputError"
        "409":
          description: Phone number already registered
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/PhoneAlreadyRegisteredError"
        "500":
          description: Server error
          content:
            application/json:
              schema:
                $ref: "common.yaml#/schemas/SystemError"

  /api/auth/login:
    post:
      operationId: login
      summary: Login with phone number and password
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
          description: Invalid input (phone format)
          content:
            application/json:
              schema:
                $ref: "auth.yaml#/schemas/AuthInvalidInputError"
        "401":
          description: Authentication failed (phone not found or wrong password)
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

  PhoneAlreadyRegisteredError:
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
            description: "e.g. 'phone number already registered: 138****1234'"

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
              - "incorrect password"
              - "refresh token not found"
              - "token reuse detected"
              - "refresh token expired"
```

### 3.3 Request Schemas

```yaml
  RegisterRequest:
    type: object
    required: [phone, password]
    properties:
      phone:
        type: string
        description: "11-digit mainland China phone number"
        pattern: "^1[3-9]\\d{9}$"
        example: "13800138000"
      password:
        type: string
        description: "Password (minimum 8 characters)"
        minLength: 8
        example: "mypassword123"

  LoginRequest:
    type: object
    required: [phone, password]
    properties:
      phone:
        type: string
        description: "11-digit mainland China phone number"
        pattern: "^1[3-9]\\d{9}$"
        example: "13800138000"
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
        required: [userId]
        properties:
          userId:
            type: string
            description: "Created user UUID"

  LoginResponse:
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
    PhoneAlreadyRegisteredError:
      $ref: "auth.yaml#/schemas/PhoneAlreadyRegisteredError"
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
    // AuthenticationFailed 登录失败（手机号不存在、密码错误）
    // 错误码统一为 AUTHENTICATION_FAILED，通过 message 区分具体原因
    AuthenticationFailed = ErrorDefinition{
        Code:      "AUTHENTICATION_FAILED",
        EnMessage: "Authentication failed: {0}",
        ZhMessage: "认证失败: {0}",
    }

    // AuthParamInvalid 注册/登录参数校验失败（手机号格式、密码长度）
    AuthParamInvalid = ErrorDefinition{
        Code:      ErrorTypeParamInvalid + ".AUTH",
        EnMessage: "Invalid auth parameter: {0}",
        ZhMessage: "认证参数无效: {0}",
    }

    // PhoneAlreadyRegistered 注册时手机号已存在
    // 复用已有的 CONFLICT.USER 错误码
)
```

### 4.2 错误码映射表

| 错误码 | HTTP | 场景 |
|--------|------|------|
| `PARAM_INVALID.AUTH` | 400 | 手机号格式不合法（非11位）、密码少于8位 |
| `CONFLICT.USER` | 409 | 注册时手机号已被注册 |
| `AUTHENTICATION_FAILED` | 401 | 登录时手机号不存在、密码错误 |
| `UNAUTHORIZED` | 401 | Refresh token 无效/过期/被重用（已有） |
| `SYSTEM_ERROR` | 500 | 密码哈希失败、数据库异常等 |

### 4.3 安全考量

- 登录失败时，PRD 要求区分"手机号不存在"和"密码错误"两种提示。从安全角度通常建议统一返回"认证失败"（防枚举攻击），但 PRD 明确要求前端区分展示。因此 **code 统一为 `AUTHENTICATION_FAILED`**（不通过错误码区分），仅通过 message 携带具体原因。
- 注册时手机号已存在返回 409，PRD 明确要求提示用户"手机号已注册"。

---

## 五、领域模型变更

### 5.1 PhoneNumber 值对象（新增）

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

### 5.2 User 实体改造

路径：`internal/domain/user/user.go`

```go
type User struct {
    ID           string
    Phone        PhoneNumber // 手机号（唯一标识）
    PasswordHash string      // bcrypt 哈希
    Name         string      // 显示名称，默认为脱敏手机号
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// NewUser 注册时创建新用户
func NewUser(id string, phone PhoneNumber, passwordHash string) (*User, error) {
    now := time.Now()
    user := &User{
        ID:           id,
        Phone:        phone,
        PasswordHash: passwordHash,
        Name:         phone.Masked(),
        CreatedAt:    now,
        UpdatedAt:    now,
    }
    if err := user.Validate(); err != nil {
        return nil, err
    }
    return user, nil
}
```

### 5.3 密码强度校验（领域规则）

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

### 5.4 UserRepository 接口变更

路径：`internal/domain/user/repository.go`

```go
type UserRepository interface {
    // Create 创建用户
    Create(ctx context.Context, user *User) error

    // GetByID 根据内部 UUID 获取用户
    GetByID(ctx context.Context, id string) (*User, error)

    // GetByPhone 根据手机号获取用户（登录时使用）
    // 不存在时返回 (nil, NotFoundError)
    GetByPhone(ctx context.Context, phone string) (*User, error)

    // ExistsByPhone 检查手机号是否已注册
    ExistsByPhone(ctx context.Context, phone string) (bool, error)
}
```

### 5.5 Application 层 Commands

路径：`internal/app/auth/commands.go`

```go
// RegisterCommand 注册命令
type RegisterCommand struct {
    Phone    string // 11 位手机号
    Password string // 明文密码（>= 8 位）
}

// RegisterResult 注册结果
type RegisterResult struct {
    UserID string
}

// LoginCommand 登录命令（替换 Casdoor 版本）
type LoginCommand struct {
    Phone    string // 11 位手机号
    Password string // 明文密码
}

// LoginResult 登录结果
type LoginResult struct {
    UserID       string
    RefreshToken string    // 明文 refresh token（仅此次返回）
    ExpiresAt    time.Time // Refresh token 过期时间
}
```

---

## 六、数据库 Schema 变更

`users` 表改动：

```sql
-- 新增密码哈希字段
ALTER TABLE `users`
  ADD COLUMN `password_hash` VARCHAR(255) NOT NULL DEFAULT ''
  COMMENT 'bcrypt password hash'
  AFTER `phone`;

-- phone 列改为唯一索引
ALTER TABLE `users`
  ADD UNIQUE INDEX `uk_phone` (`phone`);

-- external_id 不再强制（过渡期保留，Casdoor 下线后移除）
ALTER TABLE `users`
  DROP INDEX `external_id`;
```

---

## 七、请求/响应示例

### 7.1 注册成功

```
POST /api/auth/register
Content-Type: application/json

{
  "phone": "13800138000",
  "password": "mypassword123"
}

→ 201 Created
{
  "requestId": "req_550e8400e29b41d4",
  "userId": "01904d3a-7b6c-7f00-8000-000000000001"
}
```

### 7.2 注册 — 手机号已存在

```
→ 409 Conflict
{
  "requestId": "req_550e8400e29b41d5",
  "error": {
    "code": "CONFLICT.USER",
    "message": "手机号已注册: 138****0000"
  }
}
```

### 7.3 注册 — 参数校验失败

```
→ 400 Bad Request
{
  "requestId": "req_550e8400e29b41d6",
  "error": {
    "code": "PARAM_INVALID.AUTH",
    "message": "密码长度不能少于 8 位"
  }
}
```

### 7.4 登录成功

```
POST /api/auth/login
Content-Type: application/json

{
  "phone": "13800138000",
  "password": "mypassword123"
}

→ 200 OK
{
  "requestId": "req_550e8400e29b41d7",
  "userId": "01904d3a-7b6c-7f00-8000-000000000001",
  "refreshToken": "a1b2c3d4e5f6...64chars",
  "expiresAt": "2026-04-15T10:30:00Z"
}
```

### 7.5 登录 — 手机号不存在

```
→ 401 Unauthorized
{
  "requestId": "req_550e8400e29b41d8",
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "手机号不存在"
  }
}
```

### 7.6 登录 — 密码错误

```
→ 401 Unauthorized
{
  "requestId": "req_550e8400e29b41d9",
  "error": {
    "code": "AUTHENTICATION_FAILED",
    "message": "密码错误"
  }
}
```

### 7.7 Token 刷新

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
  "expiresAt": "2026-04-15T10:30:00Z"
}
```

### 7.8 登出

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

### 8.2 密码哈希使用 bcrypt

bcrypt 是纯计算（无 IO），可在 Domain 层直接调用，不违反 Domain 纯净性原则。但为了可测试性，通过 `PasswordHasher` 接口在 Domain 层定义，Infrastructure 层实现 bcrypt。

### 8.3 Casdoor 过渡策略

- `external_id` 列暂时保留但去除唯一约束和 NOT NULL
- `GET /api/auth/login-url` 端点直接移除
- `POST /api/auth/login` 请求体从 `{code, state}` 改为 `{phone, password}`（破坏性变更，与 PRD 替换目标一致）
- 现有 refresh / logout / verify-token 逻辑完全复用

### 8.4 错误信息国际化

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
- [x] 密码强度校验为领域规则（`>= 8` 位）
- [x] User 实体不变量完整（ID + Phone + PasswordHash）
- [x] UserRepository 新增 `GetByPhone` / `ExistsByPhone`

**向后兼容**：
- [x] 移除 Casdoor 专属端点 `/api/auth/login-url`
- [x] `LoginRequest` 从 Casdoor OAuth 改为手机号+密码（破坏性变更，PRD 预期）
- [x] `refresh` / `logout` 保持不变
- [x] 现有 refresh token 体系完全复用
