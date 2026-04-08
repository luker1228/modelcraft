# ModelCraft 认证系统（Auth）完整分析报告

## 一、PRD 核心需求分析

### 1.1 问题陈述
- **现状**：基于 Casdoor，但仅使用登录和注册两个功能
- **目标**：用轻量级自研方案替换 Casdoor，降低系统复杂度
- **目标用户**：国内开发者

### 1.2 核心功能需求

#### **注册流程**
- 输入手机号 + 密码（+ 确认密码）
- 校验：
  - 手机号格式（11位国内号码）
  - 手机号唯一性
  - 密码强度（最少8位）
- 创建账号

#### **登录流程**
- 输入手机号 + 密码
- 验证成功后签发 JWT Token（有效期7天）
- 错误提示：手机号不存在/密码错误

### 1.3 不做内容
- ✗ 短信验证码登录
- ✗ 第三方登录（微信、GitHub）
- ✗ 忘记密码/重置密码
- ✗ 邮箱注册
- ✗ 图形验证码
- ✗ 多端 Token 管理
- ✗ 数据迁移（无存量用户）

### 1.4 影响范围
- **后端**：移除 Casdoor SDK，实现 JWT 签发/校验
- **前端**：替换登录/注册页，调整 Token 存储
- **基础设施**：下线 Casdoor 服务

---

## 二、系统架构概览

### 2.1 认证流程 (Auth Flow)

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ 1. 携带 JWT/Refresh Token
       │ 2. Header: Authorization: Bearer <token>
       ▼
┌──────────────────────────┐
│ JWT Middleware           │ 
│ (chi_jwt_auth.go)        │
├──────────────────────────┤
│ - 解析 Token             │
│ - 验证签名               │
│ - 验证过期时间           │
│ - 提取 Claims            │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│ Casbin Enforcer          │
│ (RBAC 权限检查)          │
├──────────────────────────┤
│ - 加载用户权限           │
│ - 匹配资源+操作          │
│ - 决定允许/拒绝          │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│ 业务逻辑                 │
│ (GraphQL Resolver)       │
└──────────────────────────┘
```

### 2.2 令牌策略 (Dual Token Strategy)

**Access Token (JWT)**
- 格式：JWT
- 有效期：7 天
- 存储：无状态（签名验证）
- 包含内容：
  - UserID, ExternalID, Name, Email
  - Organization, Roles, Permissions
  - Memberships（限10条，超过标记 hasMoreMemberships）

**Refresh Token (Opaque)**
- 格式：随机字符串
- 有效期：7 天
- 存储：数据库（有状态）
- 用途：Token 轮换、盗用检测
- 盗用检测：同一 refresh token 被重用则记录安全审计日志

### 2.3 分层架构

```
┌────────────────────────────────────────┐
│ Interfaces Layer                       │
│ ├─ HTTP Handlers (auth/handler.go)    │
│ └─ GraphQL Resolvers                  │
├────────────────────────────────────────┤
│ Application Layer                      │
│ ├─ TokenService (token_service.go)    │
│ ├─ PermissionCache                    │
│ ├─ PermissionLoader                   │
│ └─ PermissionVersionManager           │
├────────────────────────────────────────┤
│ Domain Layer                           │
│ ├─ User (domain/user/)                │
│ ├─ Auth Claims (domain/auth/)         │
│ ├─ Role (domain/role/)                │
│ ├─ Permission (domain/permission/)    │
│ ├─ Membership (domain/membership/)    │
│ └─ Organization (domain/organization/)│
├────────────────────────────────────────┤
│ Infrastructure Layer                   │
│ ├─ CasdoorProvider (auth provider)    │
│ ├─ Casbin Enforcer (RBAC)             │
│ ├─ Repository Implementations         │
│ └─ Database/Redis                     │
└────────────────────────────────────────┘
```

---

## 三、核心域实体详解

### 3.1 用户域 (User Domain)

#### **User Entity** (`internal/domain/user/user.go`)
```go
type User struct {
    ID         string    // ModelCraft 内部 UUID
    ExternalID string    // 来自 JWT.sub（Casdoor 用户ID）
    Name       string    // 来自 Casdoor
    Phone      string    // 来自 Casdoor
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**关键设计**：
- ExternalID 是与外部 IdP 的绑定点
- 首次登录时自动创建本地 User 记录
- Phone 字段预留给手机号登录功能

#### **UserRepository Interface** (`internal/domain/user/repository.go`)
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByExternalID(ctx context.Context, externalID string) (*User, error)
    ExistsByExternalID(ctx context.Context, externalID string) (bool, error)
    FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error)
}
```

---

### 3.2 认证域 (Auth Domain)

#### **3.2.1 JWT Claims 结构**

**ModelCraftClaims** (`internal/domain/auth/modelcraft_claims.go`)
```go
type ModelCraftClaims struct {
    jwt.RegisteredClaims
    
    // 用户身份
    UserID     string  // ModelCraft UUID
    ExternalID string  // Casdoor 用户 ID
    Name       string
    Email      string
    
    // 组织
    Organization string
    
    // 授权信息
    Roles       []string  // ["owner", "editor"]
    Permissions []string  // ["model:read", "model:write"]
    
    // 成员关系（限10条）
    Memberships        []MembershipClaimInfo
    HasMoreMemberships bool
    
    Issuer string  // "modelcraft"
}

type MembershipClaimInfo struct {
    OrgName     string
    DisplayName string
    Role        string
    JoinedAt    int64  // Unix ms
}
```

**UserClaims** - 简化身份认证用（仅包含 UserID）

#### **3.2.2 Token 对象**

**RefreshToken** (`internal/domain/auth/refresh_token.go`)
```go
type RefreshToken struct {
    ID        string     // UUID
    UserID    string
    TokenHash string     // SHA-256 hash
    ExpiresAt time.Time
    CreatedAt time.Time
    RevokedAt *time.Time // 软删除标记
}

func (t *RefreshToken) IsValid() bool { ... }
func (t *RefreshToken) IsRevoked() bool { ... }
```

**APIKey** (`internal/domain/auth/api_key.go`)
```go
type APIKey struct {
    ID         string
    UserID     string
    Name       string
    KeyHash    string      // 不存储明文
    KeyPrefix  string      // 用于展示
    LastUsedAt *time.Time
    ExpiresAt  *time.Time
    CreatedAt  time.Time
    RevokedAt  *time.Time  // 软删除
}

const APIKeyMaxPerUser = 20  // 每个用户最多20个 API Key
```

#### **3.2.3 认证提供者**

**ProviderType** (`internal/domain/auth/project_auth_config.go`)
```go
type ProviderType string

const (
    ProviderCasdoor   ProviderType = "casdoor"   // 已实现
    ProviderKeycloak  ProviderType = "keycloak"  // 预留
    ProviderOIDC      ProviderType = "oidc"      // 预留
)
```

**ProjectAuthConfig**
```go
type ProjectAuthConfig struct {
    ID          int64
    OrgName     string
    ProjectSlug string
    Provider    ProviderType
    Enabled     bool
    Config      map[string]interface{}
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

#### **3.2.4 安全审计**

**SecurityAuditLog** (`internal/domain/auth/security_audit_log_repository.go`)
```go
type SecurityAuditEvent string

const EventReuseDetected SecurityAuditEvent = "REUSE_DETECTED"

type SecurityAuditLog struct {
    ID        string
    UserID    string
    Event     SecurityAuditEvent
    Detail    map[string]any  // 灵活存储事件详情
    CreatedAt time.Time
}
```

---

### 3.3 权限域 (Permission Domain)

#### **3.3.1 Permission 值对象** (`internal/domain/permission/permission.go`)

```go
type Permission struct {
    Obj string  // 资源对象：project, model, *, 等
    Act string  // 操作：create, read, update, delete, *
}

// 权限格式："resource:action"
// 示例：
//   "project:create"
//   "model:*"      (通配符：model 下所有操作)
//   "*:*"          (超级权限)
```

**Permission 匹配逻辑**：
1. `*:*` 匹配一切
2. `resource:*` 匹配该资源下所有操作
3. `resource:action` 精确匹配

#### **3.3.2 Role 实体** (`internal/domain/permission/role.go`)

```go
type Role struct {
    ID          int       // 数据库自增 ID
    Name        string    // 角色名称
    Description string
    IsSystem    bool      // 系统角色标记
    OrgName     string    // '__SYSTEM__' for 系统角色, orgName for 自定义
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

const (
    SystemOrgName = "__SYSTEM__"
    RoleOwner     = "owner"
    RoleAdmin     = "admin"
    RoleEditor    = "editor"
    RoleViewer    = "viewer"
)
```

**系统角色权限** (`internal/infrastructure/auth/system_roles.go`)：
- **Owner**: `*:*` (全量访问)
- **Admin**: `*:*` (全量访问)
- **Editor**: `*:create`, `*:read`, `*:update` (无删除权限)
- **Viewer**: `*:read` (只读)

#### **3.3.3 UserRole 实体** (`internal/domain/permission/user_role.go`)

```go
type UserRole struct {
    ID        int       // 数据库自增 ID
    UserID    string    // 用户 UUID
    RoleID    int       // Role ID (外键)
    OrgName   string    // 组织名称
    CreatedAt time.Time
}
```

**关键特性**：
- 用户可在不同组织拥有不同角色
- 一个用户在一个组织内只能有一个角色
- 通过三元组 (UserID, RoleID, OrgName) 唯一确定

#### **3.3.4 权限仓储**

```go
// RoleRepository - 角色管理
type RoleRepository interface {
    CreateRole(ctx context.Context, role *Role) error
    GetRoleByID(ctx context.Context, id int) (*Role, error)
    GetRoleByNameAndOrg(ctx context.Context, name, orgName string) (*Role, error)
    ListRolesByOrg(ctx context.Context, orgName string, includeSystem bool) ([]*Role, error)
    UpdateRole(ctx context.Context, role *Role) error
    DeleteRole(ctx context.Context, id int) error
}

// PermissionRepository - 权限管理
type PermissionRepository interface {
    AddPermission(ctx context.Context, roleID int, orgName string, perm *Permission) error
    RemovePermission(ctx context.Context, roleID int, obj, act string) error
    ListPermissionsByRole(ctx context.Context, roleID int) ([]*Permission, error)
    DeletePermissionsByRole(ctx context.Context, roleID int) error
}

// UserRoleRepository - 用户-角色绑定
type UserRoleRepository interface {
    AssignRole(ctx context.Context, userRole *UserRole) error
    RevokeRole(ctx context.Context, userID string, roleID int, orgName string) error
    ListUserRoles(ctx context.Context, userID, orgName string) ([]*UserRole, error)
    ListRoleUsers(ctx context.Context, roleID int, orgName string) ([]*UserRole, error)
    GetUserRole(ctx context.Context, userID string, roleID int, orgName string) (*UserRole, error)
}
```

---

### 3.4 组织与成员 (Organization & Membership)

#### **3.4.1 Organization 实体** (`internal/domain/organization/organization.go`)

```go
type OrgStatus string

const (
    OrgStatusActive    OrgStatus = "active"
    OrgStatusSuspended OrgStatus = "suspended"
    OrgStatusDeleted   OrgStatus = "deleted"
)

type Organization struct {
    Name        string       // 组织唯一标识符（主键）
    DisplayName string       // UI 显示名称
    OwnerID     string       // 组织所有者用户 ID
    Status      OrgStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**名称规则**：
- 2-64 字符
- 小写字母、数字、下划线、连字符
- 必须以字母开头

#### **3.4.2 Membership 实体** (`internal/domain/membership/membership.go`)

```go
type MembershipStatus string

const (
    MembershipStatusActive    MembershipStatus = "active"
    MembershipStatusSuspended MembershipStatus = "suspended"
    MembershipStatusInvited   MembershipStatus = "invited"
)

type Membership struct {
    ID        string            // UUID
    UserID    string
    OrgName   string            // 组织名称
    Status    MembershipStatus
    InvitedBy string            // 邀请人 ID（可空）
    InvitedAt *time.Time        // 邀请时间（可空）
    JoinedAt  *time.Time        // 加入时间（可空）
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**成员生命周期**：
1. NewMembership() → active (直接加入)
2. NewInvitation() → invited (等待接受)
3. AcceptInvitation() → active (接受邀请)
4. Suspend() → suspended (挂起成员)

#### **3.4.3 关联关系**

```
User ──────── Membership ────┐
                              ├──→ Organization
                              │
         UserRole ────────────┤
            │                 │
            ▼                 │
          Role ──→ Permission │
```

---

## 四、应用层服务

### 4.1 TokenService (`internal/app/auth/token_service.go`)

**职责**：处理登录、Token 刷新、登出

```go
type TokenService struct {
    refreshTokenRepo domainauth.RefreshTokenRepository
    userRepo         domainUser.UserRepository
    auditLogRepo     domainauth.SecurityAuditLogRepository
    refreshTTL       time.Duration
}

// 主要方法：
func (s *TokenService) Login(ctx context.Context, cmd LoginCommand) (*LoginResult, error)
func (s *TokenService) Refresh(ctx context.Context, cmd RefreshCommand) (*RefreshResult, error)
func (s *TokenService) Logout(ctx context.Context, cmd LogoutCommand) error
```

**Login 流程**：
1. 根据 ExternalID 查找或创建 User
2. 生成 opaque refresh token
3. 计算 token hash，存入数据库
4. 返回明文 token（仅一次）

**Refresh 流程**：
1. 使用旧 token hash 查询数据库
2. 检查是否被吊销
3. 盗用检测：同一 token 被重用则记录审计日志
4. 生成新 token，吊销旧 token
5. 返回新 token

### 4.2 PermissionCache (`internal/app/auth/permission_cache.go`)

**职责**：缓存用户权限，支持版本管理

```go
type PermissionCache struct {
    redis          *redis.Client
    permLoader     PermissionLoaderInterface
    versionManager *PermissionVersionManager
    cacheTTL       time.Duration
}

func (pc *PermissionCache) GetUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)
func (pc *PermissionCache) GetUserPermissionsAndRoles(ctx context.Context, userID, orgName string) (RolePermissions, error)
```

**缓存策略**：
- Redis 缓存，TTL = 5 分钟
- 支持版本号失效机制
- 版本更新时自动清除旧缓存

### 4.3 PermissionLoader (`internal/app/auth/permission_loader.go`)

**职责**：从数据库加载权限

```go
type PermissionLoader struct {
    roleRepo       permission.RoleRepository
    userRoleRepo   permission.UserRoleRepository
    permRepo       permission.PermissionRepository
}

func (pl *PermissionLoader) LoadUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)
```

### 4.4 APIKeyService (`internal/app/auth/api_key_service.go`)

**职责**：API Key 管理（生成、撤销、更新）

**特性**：
- 每个用户最多 20 个 API Key
- 支持过期时间设置
- 支持软删除（吊销）

---

## 五、基础设施层

### 5.1 CasdoorProvider (`internal/infrastructure/auth/casdoor_provider.go`)

```go
type CasdoorProvider struct {
    endpoint     string
    clientID     string
    clientSecret string
    organization string
    application  string
    certificate  string
    publicKey    *rsa.PublicKey  // 缓存的公钥
}

func (p *CasdoorProvider) GetPublicKey(ctx context.Context) (interface{}, error)
func (p *CasdoorProvider) GetSigningMethod() string  // "RS256"
func (p *CasdoorProvider) Type() string             // "casdoor"
```

**工作原理**：
- 从 X.509 证书中提取 RSA 公钥
- 缓存公钥以减少解析开销
- 用于 JWT 签名验证

### 5.2 Casbin Enforcer (`internal/infrastructure/auth/casbin_enforcer.go`)

**职责**：RBAC 权限决策引擎

```go
type CasbinEnforcer struct {
    enforcer *casbin.Enforcer
}

func (ce *CasbinEnforcer) Enforce(sub, obj, act string) bool
```

**工作流**：
1. Casbin Model 定义 RBAC 模型
2. 从数据库加载策略规则
3. 检查主体(subject)对客体(object)的操作(action)权限

### 5.3 SystemRolePermissions (`internal/infrastructure/auth/system_roles.go`)

```go
var SystemRolePermissions = map[string][]*permission.Permission{
    permission.RoleOwner: {
        {Obj: "*", Act: "*"},
    },
    permission.RoleAdmin: {
        {Obj: "*", Act: "*"},
    },
    permission.RoleEditor: {
        {Obj: "*", Act: "create"},
        {Obj: "*", Act: "read"},
        {Obj: "*", Act: "update"},
    },
    permission.RoleViewer: {
        {Obj: "*", Act: "read"},
    },
}
```

---

## 六、接口层

### 6.1 HTTP Handlers (`internal/interfaces/http/handlers/auth/handler.go`)

```go
type Handler struct {
    casdoorURL   string
    clientID     string
    clientSecret string
    redirectURI  string
    tokenService *appAuth.TokenService
    logger       logfacade.Logger
}

// 主要端点：
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request)      // POST /api/auth/login
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request)    // POST /api/auth/refresh
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request)     // POST /api/auth/logout
func (h *Handler) GetLoginURL(w http.ResponseWriter, r *http.Request)      // GET /api/auth/login-url
```

**Request/Response 格式**：

```json
// POST /api/auth/login
Request:
{
    "externalId": "casdoor_user_id",
    "email": "user@example.com",
    "name": "User Name"
}

Response:
{
    "userId": "uuid-v7",
    "refreshToken": "opaque_token_string",
    "expiresAt": "2026-04-14T22:30:00Z"
}

// POST /api/auth/refresh
Request:
{
    "refreshToken": "opaque_token_string"
}

Response: (同上)

// POST /api/auth/logout
Request:
{
    "refreshToken": "opaque_token_string"
}

// GET /api/auth/login-url?state=<state_value>
Response:
{
    "loginUrl": "https://casdoor.com/login/oauth/authorize?..."
}
```

### 6.2 GraphQL Schema (`internal/interfaces/graph/org/schema/`)

#### **user_management.graphql**

```graphql
# 类型定义
type CurrentUser {
    id: String!
    externalID: String!
    email: String!
    name: String!
    organization: Organization
    role: Role
    permissions: [String!]!
}

type Organization {
    id: ID!
    name: String!
    displayName: String
    ownerID: String!
    status: OrganizationStatus!
    createdAt: String!
    updatedAt: String!
}

type Role {
    id: ID!
    name: String!
    description: String
    permissions: [String!]!
    isSystem: Boolean!
    createdAt: String!
    updatedAt: String!
}

type OrganizationMember {
    id: ID!
    userID: String!
    userName: String!
    orgID: String!
    role: Role!
    status: MembershipStatus!
    joinedAt: String
    createdAt: String!
}

# 查询
extend type Query {
    me: CurrentUser!
    myOrganizations: [Organization!]!
    organizationMembers: [OrganizationMember!]! @hasPermission(action: "user:list")
    roles: [Role!]! @hasPermission(action: "role:read")
}

# 变更
extend type Mutation {
    updateOrganization(input: UpdateOrganizationInput!): UpdateOrganizationPayload! @hasPermission(action: "organization:update")
    createRole(input: CreateRoleInput!): CreateRolePayload! @hasPermission(action: "role:manage")
    deleteRole(id: ID!): DeleteRolePayload! @hasPermission(action: "role:manage")
}
```

#### **permission.graphql**

```graphql
type PermissionRole {
    id: Int!
    name: String!
    description: String
    isSystem: Boolean!
    orgName: String!
    createdAt: Time!
    updatedAt: Time!
}

type PermissionDef {
    obj: String!
    act: String!
}

type UserRoleAssignment {
    id: Int!
    userId: String!
    roleId: Int!
    orgName: String!
    createdAt: Time!
}

# 查询
extend type Query {
    permissionRoles(orgName: String!, includeSystem: Boolean): [PermissionRole!]!
    permissionRole(id: Int!): PermissionRole
    userRoleAssignments(userId: String!, orgName: String!): [UserRoleAssignment!]!
    rolePermissionsList(roleId: Int!): [PermissionDef!]!
}

# 变更
extend type Mutation {
    createCustomRole(input: CreateCustomRoleInput!): CreateCustomRolePayload!
    updatePermissionRole(roleId: Int!, input: UpdateRoleInput!): UpdatePermissionRolePayload!
    deletePermissionRole(roleId: Int!): DeletePermissionRolePayload!
    addPermissionToRole(roleId: Int!, obj: String!, act: String!): AddRolePermissionPayload!
    removePermissionFromRole(roleId: Int!, obj: String!, act: String!): RemoveRolePermissionPayload!
    assignRoleToUser(userId: String!, roleId: Int!, orgName: String!): AssignRolePayload!
    revokeRoleFromUser(userId: String!, roleId: Int!, orgName: String!): RevokeRolePayload!
}
```

#### **base.graphql**

```graphql
scalar Int64
scalar Date
scalar Time

directive @hasPermission(action: String!) on FIELD_DEFINITION

interface Node {
    id: ID!
}

type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: String
    endCursor: String
}

type Query {
    hello: String!
    ping: String!
    node(id: ID!): Node
}

type Mutation {
    pong: String!
}
```

---

## 七、业务流程

### 7.1 用户注册流程

```
用户输入手机号 + 密码
        ↓
前端格式校验
├─ 手机号格式（11位）
├─ 密码长度（≥8位）
└─ 密码确认匹配
        ↓
POST /auth/register
        ↓
后端唯一性校验（手机号是否已注册）
        ↓
成功：bcrypt 哈希密码 → 存入数据库 → 跳转登录页
失败：提示"该手机号已注册"
```

### 7.2 用户登录流程

```
用户输入手机号 + 密码
        ↓
POST /auth/login { externalId, email, name }
        ↓
后端查询或创建 User（通过 ExternalID）
        ↓
生成 opaque Refresh Token
├─ UUID 作为 token ID
├─ SHA-256 hash 存储
└─ TTL = 7 天
        ↓
签发 JWT Access Token
├─ 有效期 7 天
├─ Claims 包含 roles, permissions, memberships
└─ 签名算法 HS256/RS256
        ↓
返回 { userId, refreshToken, expiresAt }
        ↓
前端存储 Token
├─ Access Token → localStorage / memory
└─ Refresh Token → secure cookie / localStorage
        ↓
跳转到 Workspace 页
```

### 7.3 Token 刷新流程

```
前端: Access Token 即将过期
        ↓
POST /auth/refresh { refreshToken }
        ↓
后端查询 Refresh Token 记录
        ↓
检查是否被吊销
        ↓
检查是否过期
        ↓
盗用检测：同一 token 被重用
├─ 是 → 记录审计日志 "REUSE_DETECTED" → 吊销用户所有 refresh token
└─ 否 → 继续
        ↓
生成新 Refresh Token
        ↓
吊销旧 Refresh Token（逻辑删除）
        ↓
签发新 Access Token
        ↓
返回 { userId, newRefreshToken, expiresAt }
```

### 7.4 用户登出流程

```
前端: 用户点击登出按钮
        ↓
POST /auth/logout { refreshToken }
        ↓
后端吊销 Refresh Token
├─ 标记 RevokedAt 时间戳
└─ 后续查询时返回 nil（逻辑删除）
        ↓
前端清除本地 Token
├─ 清除 localStorage / cookies
└─ 重置状态
        ↓
重定向到登录页
```

### 7.5 权限检查流程 (RBAC)

```
请求携带 Authorization Header
        ↓
JWT Middleware 解析 Token
        ↓
从 Claims 提取：
├─ UserID
├─ Organization (OrgName)
├─ Roles
└─ Permissions
        ↓
GraphQL Directive @hasPermission(action: "resource:action")
        ↓
检查权限：
1. 加载用户权限（从缓存或数据库）
2. 匹配权限规则
   ├─ "*:*" 匹配一切
   ├─ "resource:*" 匹配该资源下所有操作
   └─ "resource:action" 精确匹配
        ↓
允许 → 执行 Resolver
拒绝 → 返回 403 Forbidden
```

---

## 八、文件清单

### 8.1 Domain Layer
```
internal/domain/
├── auth/
│   ├── api_key.go                      // API Key 值对象
│   ├── api_key_repository.go           // API Key 仓储接口
│   ├── config_validator.go             // 认证配置验证器
│   ├── errors.go                       // 错误定义
│   ├── modelcraft_claims.go            // JWT Claims（完整）
│   ├── project_auth_config.go          // 项目认证配置
│   ├── provider.go                     // 认证提供者接口
│   ├── refresh_token.go                // Refresh Token 值对象
│   ├── refresh_token_repository.go     // Refresh Token 仓储接口
│   ├── refresh_token_test.go
│   ├── security_audit_log_repository.go// 安全审计日志接口
│   └── user_claims.go                  // JWT Claims（简化）
├── user/
│   ├── user.go                         // User 实体
│   ├── user_test.go
│   └── repository.go                   // User 仓储接口
├── role/
│   ├── role.go                         // Role 实体
│   ├── role_test.go
│   └── repository.go                   // Role 仓储接口
├── permission/
│   ├── permission.go                   // Permission 值对象
│   ├── permission_test.go
│   ├── role.go                         // Role 实体（权限管理版）
│   ├── role_test.go
│   ├── user_role.go                    // UserRole 实体
│   ├── user_role_test.go
│   └── repository.go                   // 权限/角色/用户-角色仓储接口
├── membership/
│   ├── membership.go                   // Membership 实体
│   ├── membership_test.go
│   └── repository.go                   // Membership 仓储接口
└── organization/
    ├── organization.go                 // Organization 实体
    ├── organization_test.go
    └── repository.go                   // Organization 仓储接口
```

### 8.2 Application Layer
```
internal/app/auth/
├── token_service.go                    // Token 生成/刷新/登出
├── token_service_test.go
├── api_key_service.go                  // API Key 管理
├── api_key_service_test.go
├── cleanup_service.go                  // Token 清理服务
├── commands.go                         // 命令定义
├── permission_cache.go                 // 权限缓存
├── permission_cache_test.go
├── permission_loader.go                // 权限加载器
├── permission_loader_test.go
├── permission_version_manager.go       // 权限版本管理
├── permission_version_manager_test.go
├── token_generator.go                  // Token 生成器
└── (其他服务)
```

### 8.3 Infrastructure Layer
```
internal/infrastructure/auth/
├── casdoor_provider.go                 // Casdoor 认证提供者
├── casbin_enforcer.go                  // Casbin RBAC 引擎
├── casbin_enforcer_test.go
├── casbin_model.conf                   // Casbin 模型定义
└── system_roles.go                     // 系统角色权限配置
```

### 8.4 Interfaces Layer
```
internal/interfaces/
├── http/
│   └── handlers/
│       └── auth/
│           └── handler.go              // HTTP 认证处理器
├── middleware/
│   └── chi_jwt_auth.go                 // JWT 中间件
└── graph/org/schema/
    ├── user_management.graphql         // 用户/组织/角色 GraphQL Schema
    ├── permission.graphql              // 权限管理 GraphQL Schema
    └── base.graphql                    // 基础类型定义
```

### 8.5 设计文档
```
ai-metadata/
├── prd/auth/
│   ├── auth.md                         // 认证需求文档（本文）
│   └── process-flow.md                 // 业务流程图
├── backend/design/domain-model/
│   ├── 1-auth.md                       // 认证域设计
│   └── 4-rbac.md                       // RBAC 设计
└── diagrams/
    └── cluster/
        └── process-flow.md             // 集群流程图
```

---

## 九、业务对象总结

### 9.1 核心对象

| 对象 | 描述 | 关键字段 |
|------|------|--------|
| **User** | 用户实体 | ID, ExternalID, Name, Phone, CreatedAt |
| **Organization** | 组织实体 | Name, DisplayName, OwnerID, Status |
| **Membership** | 用户-组织关联 | ID, UserID, OrgName, Status, JoinedAt |
| **Role** | 角色定义 | ID, Name, OrgName, IsSystem, Permissions |
| **Permission** | 权限 | Obj, Act (resource:action 格式) |
| **UserRole** | 用户-角色绑定 | UserID, RoleID, OrgName |
| **RefreshToken** | 刷新令牌 | ID, UserID, TokenHash, ExpiresAt, RevokedAt |
| **APIKey** | API 密钥 | ID, UserID, KeyHash, ExpiresAt |
| **ModelCraftClaims** | JWT 声明 | UserID, Roles, Permissions, Memberships |
| **ProjectAuthConfig** | 项目认证配置 | OrgName, ProjectSlug, Provider, Config |

### 9.2 关键流程

| 流程 | 触发 | 主要步骤 |
|------|------|--------|
| **注册** | 用户提交表单 | 校验 → 唯一性检查 → 创建 User |
| **登录** | 用户提交凭证 | 查询 User → 生成 Token → 返回 |
| **Token 刷新** | Access Token 即将过期 | 查询 Refresh Token → 盗用检测 → 生成新 Token |
| **登出** | 用户点击登出 | 吊销 Refresh Token → 清除本地 Token |
| **权限检查** | 请求受保护资源 | 解析 JWT → 提取权限 → RBAC 验证 |
| **角色分配** | 组织管理员操作 | 创建 UserRole → 加载权限 → 缓存清除 |

---

## 十、关键设计决策

### 10.1 Dual Token 策略

**为什么需要两个 Token？**

1. **Access Token (JWT)**
   - 优点：无状态，减少数据库查询
   - 缺点：无法立即吊销
   - 用途：常规请求认证

2. **Refresh Token (Opaque)**
   - 优点：有状态，可控制，支持吊销
   - 缺点：需要数据库查询
   - 用途：Token 轮换、盗用检测

### 10.2 权限缓存三阶段

```
Phase 1: 直接从数据库加载
         (PermissionLoader)
              ↓
Phase 2: Redis 缓存 + TTL
         (PermissionCache)
              ↓
Phase 3: 版本号失效机制
         (PermissionVersionManager)
```

### 10.3 系统角色不可修改

```
系统角色（Owner, Admin, Editor, Viewer）
├─ OrgName = '__SYSTEM__'
├─ IsSystem = true
└─ 权限硬编码，不可修改
   （保证行为一致性）
```

### 10.4 ExternalID 绑定策略

```
外部 IdP (Casdoor)
    │
    ├─ 用户 JWT
    │  └─ sub = casdoor_user_id
    │
    ▼
ModelCraft User
    └─ ExternalID = casdoor_user_id
       （建立联系，支持 IdP 切换）
```

---

## 十一、扩展点

### 11.1 支持多认证提供者

当前架构设计支持：
- ✓ Casdoor (实现)
- ○ Keycloak (预留)
- ○ OIDC (预留)

只需实现 `AuthProvider` 接口和相应 `ConfigValidator`

### 11.2 权限模型升级

当前 RBAC 可升级到：
- **ABAC** (Attribute-Based Access Control) 
  - 基于用户/资源属性的动态权限
- **ReBAC** (Relationship-Based Access Control)
  - 基于关系图的权限（Zanzibar 模型）

### 11.3 OAuth2 授权码流程

当前支持的流程：
- ✓ Resource Owner Password Grant (内部使用)
- ○ Authorization Code Flow (预留)
- ○ Client Credentials Flow (API Key 替代)

---

## 十二、测试清单

### 12.1 单元测试

- [x] User 实体校验
- [x] Role 实体校验
- [x] Permission 匹配逻辑
- [x] Membership 状态转移
- [x] JWT Claims 验证
- [x] Refresh Token 有效性检查
- [x] API Key 有效性检查

### 12.2 集成测试

- [x] TokenService.Login
- [x] TokenService.Refresh (包括盗用检测)
- [x] TokenService.Logout
- [x] PermissionCache 缓存命中/失误
- [x] Casbin Enforcer 权限决策

### 12.3 端到端测试

- [ ] 完整注册流程
- [ ] 完整登录流程
- [ ] Token 轮换流程
- [ ] 权限检查流程
- [ ] 组织成员管理流程

---

## 十三、部署清单

### 13.1 数据库迁移

```sql
-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY,
    external_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 组织表
CREATE TABLE organizations (
    name VARCHAR(64) PRIMARY KEY,
    display_name VARCHAR(255),
    owner_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 成员表
CREATE TABLE memberships (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    org_name VARCHAR(64) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    invited_by UUID,
    invited_at TIMESTAMP,
    joined_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (org_name) REFERENCES organizations(name),
    UNIQUE(user_id, org_name)
);

-- Refresh Token 表
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (user_id, revoked_at)
);

-- 角色表
CREATE TABLE roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    org_name VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (org_name) REFERENCES organizations(name),
    UNIQUE(name, org_name)
);

-- 权限表
CREATE TABLE role_permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    role_id INT NOT NULL,
    obj VARCHAR(64) NOT NULL,
    act VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    UNIQUE(role_id, obj, act)
);

-- 用户-角色表
CREATE TABLE user_roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id UUID NOT NULL,
    role_id INT NOT NULL,
    org_name VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (org_name) REFERENCES organizations(name),
    UNIQUE(user_id, role_id, org_name)
);

-- API Key 表
CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(10) NOT NULL,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (user_id, revoked_at)
);

-- 安全审计日志表
CREATE TABLE security_audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    event VARCHAR(64) NOT NULL,
    detail JSON,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (user_id, created_at)
);
```

### 13.2 环境变量

```bash
# JWT 配置
JWT_SECRET_KEY=<random-32-byte-string>
JWT_ISSUER=modelcraft
JWT_ACCESS_TOKEN_TTL=7d
JWT_REFRESH_TOKEN_TTL=7d

# Casdoor 配置
CASDOOR_ENDPOINT=https://casdoor.example.com
CASDOOR_CLIENT_ID=<client_id>
CASDOOR_CLIENT_SECRET=<client_secret>
CASDOOR_ORGANIZATION=<org_name>
CASDOOR_APPLICATION=<app_name>
CASDOOR_CERTIFICATE=<x509_certificate>

# Redis 配置
REDIS_URL=redis://localhost:6379/0
PERMISSION_CACHE_TTL=300s

# 数据库配置
DATABASE_URL=postgres://user:pass@localhost/modelcraft
```

### 13.3 初始化脚本

```bash
# 创建系统角色
INSERT INTO roles (name, description, is_system, org_name) VALUES
('owner', 'Organization Owner', TRUE, '__SYSTEM__'),
('admin', 'Organization Admin', TRUE, '__SYSTEM__'),
('editor', 'Organization Editor', TRUE, '__SYSTEM__'),
('viewer', 'Organization Viewer', TRUE, '__SYSTEM__');

# 初始化权限
-- 由 SystemRolePermissions 在代码中加载
```

