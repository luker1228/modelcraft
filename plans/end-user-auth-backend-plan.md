# 终端用户认证（End-User Auth）后端落地技术方案

> **文档目标**：仅输出技术方案，不含代码实现。
> **输入依据**：
> - `ai-metadata/prd/end-user-auth/00-end-user-auth.md`
> - `ai-metadata/prd/end-user-auth/03-backend-design.md`
> - `ai-metadata/prd/end-user-auth/api-contract.md`
> - `ai-metadata/prd/end-user-auth/end-user-auth-domain.puml`
> - `ai-metadata/backend/development/architecture.md`
> - `ai-metadata/backend/development/repo-develop.md`
> - `ai-metadata/backend/development/error-handling.md`
> - `ai-metadata/backend/design/domain-model/6-database-cluster.md`
> - `ai-metadata/backend/design/domain-model/3-project.md`
> - `ai-metadata/backend/common-mistakes.md`

---

## 目录

1. [领域层设计](#1-领域层设计)
2. [数据库连接路由机制](#2-数据库连接路由机制)
3. [应用层设计](#3-应用层设计)
4. [Infrastructure 层设计](#4-infrastructure-层设计)
5. [接口层设计](#5-接口层设计)
6. [安全设计](#6-安全设计)
7. [实现顺序](#7-实现顺序)
8. [验收口径](#8-验收口径)
9. [主要风险与回滚思路](#9-主要风险与回滚思路)

---

## 1. 领域层设计

### 1.1 目录结构

```
internal/domain/enduser/
├── end_user.go                    # EndUser 实体（聚合根）
├── end_user_session.go            # EndUserSession 实体（聚合根）
├── end_user_repository.go         # EndUserRepository 接口
├── end_user_session_repository.go # EndUserSessionRepository 接口
├── end_user_auth_service.go       # EndUserAuthService 领域服务接口
├── credential.go                  # Credential 值对象
├── hashed_password.go             # HashedPassword 值对象
└── errors.go                      # 领域层 sentinel errors（不依赖 bizerrors）
```

---

### 1.2 EndUser 实体

#### 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `ID` | `string` | UUID，主键 |
| `Username` | `string` | 用户名，3–64 字符，`^[a-zA-Z0-9_-]+$`，同 Project 内唯一 |
| `Password` | `HashedPassword` | bcrypt 哈希值对象 |
| `IsForbidden` | `bool` | 是否禁用，对应 `users.is_forbidden` |
| `CreatedBy` | `string` | 创建者开发者 user_id（mc_meta） |
| `CreatedAt` | `time.Time` | |
| `UpdatedAt` | `time.Time` | |

#### 行为方法

```go
// NewEndUser 工厂方法：校验格式、生成 UUID、设置初始状态
func NewEndUser(username, createdBy string, hashedPwd HashedPassword) (*EndUser, error)

// Enable 启用账号（DISABLED → ACTIVE）
func (u *EndUser) Enable()

// Disable 禁用账号（ACTIVE → DISABLED）
func (u *EndUser) Disable()

// IsActive 是否可登录（未被禁用）
func (u *EndUser) IsActive() bool

// VerifyPassword 校验明文密码（委托 HashedPassword）
func (u *EndUser) VerifyPassword(plainPassword string) bool
```

#### 不变量

- `Username` 创建后不可修改（写入后无 SetUsername 方法）
- `IsForbidden=true` 账号不可登录，在 `Authenticate` 领域服务中拦截
- `Username` 格式：`^[a-zA-Z0-9_-]{3,64}$`，由 `NewEndUser` 强制校验

---

### 1.3 HashedPassword 值对象

```go
type HashedPassword struct {
    Hash      string // bcrypt hash 字符串
    Algorithm string // 固定 "bcrypt"
}

// NewHashedPasswordFromPlain 从明文创建（调用 bcrypt，cost=12）
func NewHashedPasswordFromPlain(plain string) (HashedPassword, error)

// NewHashedPasswordFromHash 从数据库已有 hash 恢复（不重新哈希）
func NewHashedPasswordFromHash(hash string) HashedPassword

// Verify 校验明文密码是否匹配（bcrypt.CompareHashAndPassword）
func (p HashedPassword) Verify(plain string) bool

// ValidateStrength 密码强度校验：至少 8 位，含字母 + 数字
func ValidatePasswordStrength(plain string) error
```

> **注意**：`HashedPassword` 仅在 Domain 层内使用。明文密码不在系统任何层持久化，仅在输入验证时临时存在。

---

### 1.4 Credential 值对象

```go
// Credential 输入凭证（仅用于登录/注册，不持久化）
type Credential struct {
    Username      string
    PlainPassword string
}

func NewCredential(username, plainPassword string) (Credential, error)
```

---

### 1.5 EndUserSession 实体（对应 accounts 表）

#### 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `ID` | `string` | UUID，主键 |
| `UserID` | `string` | 关联 users.id |
| `RefreshTokenHash` | `string` | sha256(opaqueToken)，索引查询，不存明文 |
| `ExpiresAt` | `time.Time` | 过期时间（创建时 +7d） |
| `Revoked` | `bool` | 是否已撤销，对应 `accounts.revoked` |
| `CreatedAt` | `time.Time` | |

#### 行为方法

```go
// NewEndUserSession 工厂方法：生成 UUID、计算过期时间
// opaqueToken 为调用方生成的 base64url 原始 token，此处只存 hash
func NewEndUserSession(userID, refreshTokenHash string, ttl time.Duration) *EndUserSession

// Invalidate 撤销会话（登出 / token rotation）
func (s *EndUserSession) Invalidate()

// IsExpired 是否已过期
func (s *EndUserSession) IsExpired() bool

// IsValid 是否可用（未撤销 + 未过期）
func (s *EndUserSession) IsValid() bool
```

> 与现有 `internal/domain/auth/RefreshToken` 结构对齐（`Revoked` 用 `bool` 而非 `*time.Time`，因为 private 库不需要审计时间戳）。

---

### 1.6 EndUserRepository 接口

```go
// internal/domain/enduser/end_user_repository.go

type EndUserRepository interface {
    // Save 创建新终端用户（INSERT，唯一冲突返回 shared.ErrTypeDuplicatedKey）
    Save(ctx context.Context, user *EndUser) error

    // GetByID 根据 ID 查询用户（不存在时返回 (nil, nil)）
    GetByID(ctx context.Context, id string) (*EndUser, error)

    // GetByUsername 根据用户名查询（不存在时返回 (nil, nil)）
    GetByUsername(ctx context.Context, username string) (*EndUser, error)

    // UpdateStatus 更新 is_forbidden 字段（检查 RowsAffected，0 行返回 NotFoundError）
    UpdateStatus(ctx context.Context, id string, isForbidden bool) error

    // Delete 物理删除用户记录（检查 RowsAffected，0 行返回 NotFoundError）
    Delete(ctx context.Context, id string) error

    // ListWithTotal 分页查询，支持用户名模糊搜索
    // cursor-based pagination：after 为上一页最后 id
    // 返回 (users, totalCount, error)
    ListWithTotal(ctx context.Context, query ListEndUsersQuery) ([]*EndUser, int64, error)
}

// ListEndUsersQuery 列表查询参数
type ListEndUsersQuery struct {
    Search   string // 用户名模糊搜索（可选）
    First    int    // 每页条数，默认 20，最大 100
    After    string // cursor（上一页最后一个 ID，可选）
}
```

> **注意**：`EndUserRepository` 的实现对 `private_{projectSlug}` 库操作，无需 org_name / project_slug 字段（已通过库名隔离）。

---

### 1.7 EndUserSessionRepository 接口

```go
// internal/domain/enduser/end_user_session_repository.go

type EndUserSessionRepository interface {
    // Save 创建新会话记录
    Save(ctx context.Context, session *EndUserSession) error

    // GetByTokenHash 根据 sha256(token) 查找会话（不存在时返回 (nil, nil)）
    GetByTokenHash(ctx context.Context, tokenHash string) (*EndUserSession, error)

    // RevokeByID 将指定 session 标记为 revoked=1
    RevokeByID(ctx context.Context, id string) error

    // RevokeAllByUserID 撤销指定用户的所有活跃会话（DELETE 用户时调用）
    RevokeAllByUserID(ctx context.Context, userID string) error
}
```

---

### 1.8 EndUserAuthService 领域服务接口

```go
// internal/domain/enduser/end_user_auth_service.go

// EndUserAuthService 定义终端用户认证的领域服务接口
// 实现位于 Application 层（不在 Domain 层实现，因为需要依赖 Repository）
type EndUserAuthService interface {
    // Authenticate 验证凭证，返回 (user, error)
    // 错误场景：凭证错误 / 账号禁用
    Authenticate(ctx context.Context, cred Credential) (*EndUser, error)
}
```

> 领域服务接口定义在 Domain 层，不提供实现，实现逻辑在 Application 层的 Use Case 中内联。

---

## 2. 数据库连接路由机制

### 2.1 整体路由链路

```
请求参数: orgName + projectSlug
    │
    ▼ ClusterConnectionManager.GetConnection(ctx, orgName, projectSlug)
    │   ├── 查 mc_meta.projects WHERE org_name=? AND slug=?
    │   │       → cluster_id（Project 未配置 Cluster → 返回 503）
    │   └── 查 mc_meta.database_clusters WHERE id=cluster_id
    │           → { host, port, username, password.Decrypt() }
    │
    ▼ PrivateDBManager.GetOrCreateDB(ctx, conn, projectSlug)
    │   ├── 构建 DSN: user:pass@tcp(host:port)/?timeout=Xs&parseTime=true
    │   ├── USE / CREATE DATABASE private_{projectSlug}
    │   └── 返回 *sql.DB（已切换到 private_{projectSlug}）
    │
    ▼ EndUserRepository(q) / EndUserSessionRepository(q)
        └── 操作 users / accounts 表
```

### 2.2 PrivateDBManager — 新组件

**位置**：`internal/infrastructure/repository/private_db_manager.go`

```
PrivateDBManager
├── connections  sync.Map    // key: "orgName.projectSlug" → *sql.DB
├── clusterMgr   *ClusterConnectionManager  // 复用现有集群连接管理
└── migrator     *PrivateMigrator           // DDL 自动初始化
```

**核心方法**：

```go
// GetOrInit 获取（或初始化）private_{projectSlug} 的数据库连接
// 1. 若缓存中已有有效连接，直接返回
// 2. 通过 ClusterConnectionManager 获取 cluster 连接信息
// 3. 建立到 private_{projectSlug} 库的连接
// 4. 首次连接时执行 auto-migrate（CREATE DATABASE IF NOT EXISTS + 建表）
// 5. 缓存连接，返回 *sql.DB
func (m *PrivateDBManager) GetOrInit(ctx context.Context, orgName, projectSlug string) (*sql.DB, error)
```

**与 ClusterConnectionManager 的区别**：

| 维度 | ClusterConnectionManager | PrivateDBManager |
|------|--------------------------|------------------|
| 目的 | 获取 cluster 主库连接（无 database 切换） | 获取 `private_{projectSlug}` 连接 |
| 缓存 key | clusterID | `orgName.projectSlug` |
| 初始化 | 无 DDL | 自动执行建库建表 |
| 依赖 | DatabaseClusterRepository | ClusterConnectionManager |

### 2.3 PrivateMigrator — DDL 自动初始化

**位置**：`internal/infrastructure/database/private/migrator.go`

```
PrivateMigrator
└── Migrate(ctx context.Context, db *sql.DB, projectSlug string) error
    ├── CREATE DATABASE IF NOT EXISTS `private_{projectSlug}`
    │       DEFAULT CHARACTER SET utf8mb4
    ├── USE `private_{projectSlug}`
    ├── CREATE TABLE IF NOT EXISTS users (...)   -- DDL 见 PRD
    └── CREATE TABLE IF NOT EXISTS accounts (...) -- DDL 见 PRD
```

**DDL 幂等性**：所有 DDL 使用 `IF NOT EXISTS`，可重入执行。

**触发时机**：

- 开发者创建第一个终端用户时（`CreateEndUser` 用例）
- 后续每次 `PrivateDBManager.GetOrInit` 时，首次建立连接时执行 migrate
- Migrate 只在连接首次建立时执行一次，通过一个 `sync.Map` 记录已 migrate 的 `projectSlug`

### 2.4 连接池管理策略

- **缓存粒度**：per `orgName.projectSlug`（不是 per clusterID），因为同一 cluster 下可能有多个 Project 的 private 库
- **连接池配置**：复用 `config.DatabaseConfig`（MaxOpenConns / MaxIdleConns / ConnMaxLifetime）
- **健康检查**：参照现有 `ClusterConnectionManager` 的 Ping 检测模式，连接失效时重建
- **不做全局 sync**：`PrivateDBManager` 不做类似 ClusterManager 的定时全量 sync，按需获取即可

### 2.5 Project 未配置 Cluster 的处理

```
PrivateDBManager.GetOrInit
    └── ClusterConnectionManager.GetConnection → 返回 NOT_FOUND 错误
        └── App 层转换为 bizerrors.ClusterNotConfigured（503）
```

需新增错误码：`CLUSTER_NOT_CONFIGURED.END_USER`，HTTP 503。

---

## 3. 应用层设计

### 3.1 目录结构

```
internal/app/enduser/
├── commands.go               # 所有 Command / Query 结构体
├── end_user_app_service.go   # 用户管理用例（CreateEndUser、ListEndUsers 等）
└── end_user_auth_service.go  # 认证用例（Register、Login、Logout、Refresh、Me）
```

### 3.2 Commands & Queries

```go
// commands.go

// --- 认证命令 ---

type RegisterCommand struct {
    OrgName     string
    ProjectSlug string
    Username    string
    Password    string
}
type RegisterResult struct {
    UserID       string
    RefreshToken string // opaque token 明文（仅此一次）
    ExpiresAt    time.Time
}

type LoginCommand struct {
    OrgName     string
    ProjectSlug string
    Username    string
    Password    string
}
type LoginResult = RegisterResult // 完全相同结构

type LogoutCommand struct {
    OrgName      string
    ProjectSlug  string
    RefreshToken string // opaque token 明文
}

type RefreshCommand struct {
    OrgName      string
    ProjectSlug  string
    RefreshToken string // opaque token 明文
}
type RefreshResult = RegisterResult

type GetMeCommand struct {
    OrgName     string
    ProjectSlug string
    UserID      string // 由 BFF 从 JWT 解析后通过 Header 传入
}

// --- 用户管理命令 ---

type CreateEndUserCommand struct {
    OrgName     string
    ProjectSlug string
    Username    string
    Password    string
    CreatedBy   string // 开发者 user_id（mc_meta）
}

type ListEndUsersCommand struct {
    OrgName     string
    ProjectSlug string
    Search      string
    First       int    // 默认 20
    After       string // cursor
}
type ListEndUsersResult struct {
    Items      []*enduser.EndUser
    TotalCount int64
    HasNextPage bool
    EndCursor  string
}

type UpdateEndUserStatusCommand struct {
    OrgName     string
    ProjectSlug string
    UserID      string
    IsForbidden bool
}

type DeleteEndUserCommand struct {
    OrgName     string
    ProjectSlug string
    UserID      string
}
```

---

### 3.3 EndUserAuthAppService — 认证用例

#### RegisterEndUser

**函数签名**：

```go
func (s *EndUserAuthAppService) RegisterEndUser(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error)
```

**核心步骤**：

1. 通过 `PrivateDBManager.GetOrInit(ctx, orgName, projectSlug)` 获取 DB 连接（含 auto-migrate）
2. 构建 `EndUserRepository` 和 `EndUserSessionRepository`（传入上一步的 `*sql.DB`）
3. `ValidatePasswordStrength(cmd.Password)` → 强度不足返回 `PARAM_INVALID`
4. `ValidateUsername(cmd.Username)` → 格式不合法返回 `PARAM_INVALID`
5. `NewHashedPasswordFromPlain(cmd.Password, cost=12)` → bcrypt hash
6. `NewEndUser(username, "", hashedPwd)` → 注册时 CreatedBy 为空（自助注册）
7. `endUserRepo.Save(ctx, user)` → 唯一冲突 → `CONFLICT.END_USER`
8. 生成 opaque refresh token：`crypto/rand` 32 bytes → base64url
9. `sha256(token)` → `tokenHash`
10. `NewEndUserSession(user.ID, tokenHash, 7*24h)` → `sessionRepo.Save`
11. 返回 `RegisterResult{ UserID, RefreshToken(明文), ExpiresAt }`

**错误映射**：

| 场景 | 错误码 |
|------|--------|
| username 格式不合法 / 密码强度不足 | `PARAM_INVALID.END_USER` → HTTP 400 |
| 用户名已存在 | `CONFLICT.END_USER` → HTTP 409 |
| Project 未关联 Cluster | `CLUSTER_NOT_CONFIGURED.END_USER` → HTTP 503 |

---

#### LoginEndUser

**函数签名**：

```go
func (s *EndUserAuthAppService) LoginEndUser(ctx context.Context, cmd LoginCommand) (*LoginResult, error)
```

**核心步骤**：

1. `PrivateDBManager.GetOrInit` → DB 连接（若 private 库不存在，登录本就不可能成功，auto-migrate 后表为空，后续步骤会返回 NOT_FOUND）
2. `endUserRepo.GetByUsername(ctx, username)` → `(nil, nil)` → **返回 `INVALID_CREDENTIALS`（不暴露用户是否存在）**
3. `user.VerifyPassword(cmd.Password)` → `false` → `INVALID_CREDENTIALS`
4. `user.IsActive()` → `false` → `ACCOUNT_DISABLED`（403）
5. 生成 opaque token → sha256 → `sessionRepo.Save`
6. 返回 `LoginResult`

> **防枚举**：步骤 2 和步骤 3 统一返回相同错误码 `INVALID_CREDENTIALS`，不区分"用户不存在"和"密码错误"。

---

#### LogoutEndUser

**函数签名**：

```go
func (s *EndUserAuthAppService) LogoutEndUser(ctx context.Context, cmd LogoutCommand) error
```

**核心步骤**：

1. `PrivateDBManager.GetOrInit` → DB 连接
2. `sha256(cmd.RefreshToken)` → `tokenHash`
3. `sessionRepo.GetByTokenHash(ctx, tokenHash)` → `nil` → **直接返回 nil（幂等，不报错）**
4. `sessionRepo.RevokeByID(ctx, session.ID)` → 标记 `revoked=1`
5. 返回 nil

---

#### RefreshEndUserToken

**函数签名**：

```go
func (s *EndUserAuthAppService) RefreshEndUserToken(ctx context.Context, cmd RefreshCommand) (*RefreshResult, error)
```

**核心步骤**：

1. `PrivateDBManager.GetOrInit` → DB 连接
2. `sha256(cmd.RefreshToken)` → `tokenHash`
3. `sessionRepo.GetByTokenHash(ctx, tokenHash)` → `nil` → `INVALID_REFRESH_TOKEN`
4. `session.IsValid()` → `false`（revoked 或过期） → `INVALID_REFRESH_TOKEN`
5. **Token Rotation**（在事务中）：
   a. `sessionRepo.RevokeByID(ctx, session.ID)` — 旧 token 撤销
   b. 生成新 opaque token → sha256
   c. `NewEndUserSession(session.UserID, newHash, 7*24h)` → `sessionRepo.Save`
6. 返回 `RefreshResult{ UserID: session.UserID, RefreshToken: newToken, ExpiresAt }`

> **事务**：步骤 5 的 Revoke + Insert 需在同一事务内，使用 `TxManager.WithTx`，防止 Revoke 成功但 Insert 失败导致用户被踢出。

---

#### GetEndUserMe

**函数签名**：

```go
func (s *EndUserAuthAppService) GetEndUserMe(ctx context.Context, cmd GetMeCommand) (*enduser.EndUser, error)
```

**核心步骤**：

1. `PrivateDBManager.GetOrInit` → DB 连接
2. `endUserRepo.GetByID(ctx, cmd.UserID)` → `nil` → `NOT_FOUND.END_USER`
3. `user.IsActive()` → `false` → `ACCOUNT_DISABLED`
4. 返回 user

> JWT 验证由 BFF 完成，Go Backend 不重复验证。UserID 通过 `X-End-User-Id` Header 传入。

---

### 3.4 EndUserManagementAppService — 用户管理用例

#### CreateEndUser（开发者创建）

**函数签名**：

```go
func (s *EndUserManagementAppService) CreateEndUser(ctx context.Context, cmd CreateEndUserCommand) (*enduser.EndUser, error)
```

**核心步骤**：

1. `PrivateDBManager.GetOrInit` → DB 连接（含 auto-migrate，这是 **第一次创建用户的入口**）
2. `ValidatePasswordStrength` + `ValidateUsername`
3. `NewHashedPasswordFromPlain(cmd.Password, cost=12)`
4. `NewEndUser(username, cmd.CreatedBy, hashedPwd)` → `createdBy` 设置为开发者 ID
5. `endUserRepo.Save` → 唯一冲突 → `CONFLICT.END_USER`
6. 返回已创建的 `EndUser`（**不自动登录**，与自助注册的区别）

---

#### ListEndUsers

**函数签名**：

```go
func (s *EndUserManagementAppService) ListEndUsers(ctx context.Context, cmd ListEndUsersCommand) (*ListEndUsersResult, error)
```

**核心步骤**：

1. `PrivateDBManager.GetOrInit` → DB 连接
2. `endUserRepo.ListWithTotal(ctx, query)` → `([]*EndUser, totalCount, error)`
3. 计算 cursor：取最后一个 user.ID 作为 endCursor，判断 `len(items) == first` 作为 `hasNextPage`
4. 返回 `ListEndUsersResult`

---

#### UpdateEndUserStatus

**函数签名**：

```go
func (s *EndUserManagementAppService) UpdateEndUserStatus(ctx context.Context, cmd UpdateEndUserStatusCommand) (*enduser.EndUser, error)
```

**核心步骤**：

1. `PrivateDBManager.GetOrInit` → DB 连接
2. `endUserRepo.GetByID(ctx, cmd.UserID)` → `nil` → `NOT_FOUND.END_USER`
3. 更新内存实体状态（`Enable()` 或 `Disable()`）
4. `endUserRepo.UpdateStatus(ctx, user.ID, cmd.IsForbidden)`
5. 返回更新后的 EndUser

---

#### DeleteEndUser

**函数签名**：

```go
func (s *EndUserManagementAppService) DeleteEndUser(ctx context.Context, cmd DeleteEndUserCommand) error
```

**核心步骤**（需事务）：

1. `PrivateDBManager.GetOrInit` → DB 连接
2. `endUserRepo.GetByID` → `nil` → `NOT_FOUND.END_USER`（前置检查，提供明确错误）
3. **在事务中**：
   a. `sessionRepo.RevokeAllByUserID(ctx, userID)` — 撤销所有 sessions
   b. `endUserRepo.Delete(ctx, userID)` — 物理删除用户
4. 返回 nil

> 事务保证 sessions 被撤销和用户被删除的原子性，防止孤儿 session。

---

## 4. Infrastructure 层设计

### 4.1 目录结构

```
internal/infrastructure/
├── database/
│   └── private/
│       └── migrator.go              # PrivateMigrator（建库建表 DDL）
└── repository/
    ├── private_db_manager.go        # PrivateDBManager（连接管理）
    ├── sql_enduser_repository.go    # EndUserRepository MySQL 实现
    └── sql_enduser_session_repository.go  # EndUserSessionRepository MySQL 实现
```

### 4.2 PrivateMigrator DDL

```sql
-- 建库（幂等）
CREATE DATABASE IF NOT EXISTS `private_{projectSlug}`
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `private_{projectSlug}`;

-- users 表（幂等）
CREATE TABLE IF NOT EXISTS users (
    id           VARCHAR(36)  NOT NULL PRIMARY KEY,
    username     VARCHAR(64)  NOT NULL,
    password     VARCHAR(255) NOT NULL,
    is_forbidden TINYINT(1)   NOT NULL DEFAULT 0,
    created_by   VARCHAR(36)  NOT NULL,
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- accounts 表（幂等）
CREATE TABLE IF NOT EXISTS accounts (
    id                 VARCHAR(36)  NOT NULL PRIMARY KEY,
    user_id            VARCHAR(36)  NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,
    expires_at         DATETIME     NOT NULL,
    revoked            TINYINT(1)   NOT NULL DEFAULT 0,
    created_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    UNIQUE KEY uq_token_hash (refresh_token_hash),
    CONSTRAINT fk_accounts_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

### 4.3 SQL 查询设计（sqlc 风格）

#### EndUserRepository — SQL 语句

```sql
-- name: GetEndUserByID :one
SELECT id, username, password, is_forbidden, created_by, created_at, updated_at
FROM users
WHERE id = ?;

-- name: GetEndUserByUsername :one
SELECT id, username, password, is_forbidden, created_by, created_at, updated_at
FROM users
WHERE username = ?;

-- name: InsertEndUser :exec
INSERT INTO users (id, username, password, is_forbidden, created_by, created_at, updated_at)
VALUES (?, ?, ?, 0, ?, NOW(), NOW());

-- name: UpdateEndUserIsForbidden :execrows
UPDATE users
SET is_forbidden = ?, updated_at = NOW()
WHERE id = ?;

-- name: DeleteEndUser :execrows
DELETE FROM users WHERE id = ?;

-- name: ListEndUsers :many
-- 游标分页：after 传上一页最后一个 id
SELECT id, username, is_forbidden, created_by, created_at, updated_at
FROM users
WHERE (? = '' OR username LIKE CONCAT('%', ?, '%'))
  AND (? = '' OR id > ?)        -- cursor: after
ORDER BY id ASC
LIMIT ?;

-- name: CountEndUsers :one
SELECT COUNT(*) FROM users
WHERE (? = '' OR username LIKE CONCAT('%', ?, '%'));
```

#### EndUserSessionRepository — SQL 语句

```sql
-- name: InsertEndUserSession :exec
INSERT INTO accounts (id, user_id, refresh_token_hash, expires_at, revoked, created_at)
VALUES (?, ?, ?, ?, 0, NOW());

-- name: GetEndUserSessionByTokenHash :one
SELECT id, user_id, refresh_token_hash, expires_at, revoked, created_at
FROM accounts
WHERE refresh_token_hash = ?;

-- name: RevokeEndUserSessionByID :execrows
UPDATE accounts SET revoked = 1 WHERE id = ?;

-- name: RevokeAllEndUserSessionsByUserID :exec
UPDATE accounts SET revoked = 1 WHERE user_id = ?;
```

---

### 4.4 EndUserRepository 实现规范

**结构体**：

```go
type SqlEndUserRepository struct {
    q dbgen.Querier  // 接收 Querier 接口，支持事务
}

func NewSqlEndUserRepository(q dbgen.Querier) enduser.EndUserRepository {
    return &SqlEndUserRepository{q: q}
}

// 编译期接口满足检查（文件末尾）
var _ enduser.EndUserRepository = (*SqlEndUserRepository)(nil)
```

**错误处理规范**：

- `GetByID` / `GetByUsername`：使用**模式 B**（`(value, bool, error)`），因为不存在是 App 层判断的（App 层按需转换为 `NOT_FOUND.END_USER` 或 `INVALID_CREDENTIALS`）

  > 实际上，根据 repo-develop.md，对于"不存在时 App 层决定语义"的情况，应返回 `(nil, nil)` 的模式 A，交由 App 层检查 nil。  
  > **采用模式 A**（`(*EndUser, error)`）：不存在 → `(nil, nil)`，DB 错误 → `(nil, RepositoryError)`。

- `Save`：`sqlerr.ExecWithErrorHandling`，唯一冲突 → `shared.ErrTypeDuplicatedKey`，App 层转换为 `CONFLICT.END_USER`

- `UpdateStatus` / `Delete`：检查 `RowsAffected == 0` → 返回 `shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "...")`

> **关键注意（对齐 BM-20250415-0001 错题）**：`private_{projectSlug}` 库已经通过库名完全隔离，**表内字段不需要 org_name / project_slug 过滤**。连接建立时已经 `USE private_{projectSlug}`，不会跨租户污染。

---

### 4.5 PrivateDBManager 实现要点

```go
type PrivateDBManager struct {
    connections   sync.Map             // key: "orgName.projectSlug" → *sql.DB
    migrated      sync.Map             // key: "orgName.projectSlug" → bool（已初始化标志）
    clusterMgr    *ClusterConnectionManager
    migrator      *private.PrivateMigrator
    dbConfig      *config.DatabaseConfig
    logger        logfacade.Logger
}
```

**`GetOrInit` 流程**：

1. 尝试从 `connections` 取缓存连接，若存在则 Ping 检查
2. Ping 失败 → 清除缓存，重建连接
3. 缓存不存在 → 调用 `clusterMgr.GetConnection(ctx, orgName, projectSlug)` 获取 cluster 主连接信息
   - 若 cluster 不存在（`NotFoundError`）→ 转换为 `CLUSTER_NOT_CONFIGURED.END_USER`
4. 构建 private 库 DSN：`user:pass@tcp(host:port)/private_{projectSlug}?parseTime=true&charset=utf8mb4`
5. `sql.Open` + `db.Ping`
6. 若 `migrated` 中未记录，执行 `migrator.Migrate(ctx, db, projectSlug)`，写入 `migrated`
7. 配置连接池参数，存入 `connections`
8. 返回 `*sql.DB`

---

## 5. 接口层设计

### 5.1 HTTP Handler（`/internal/` 内网 REST 接口）

#### 路由注册

在 `internal/interfaces/http/routes.go` 中追加路由组：

```
/internal/end-user/auth/register   POST  → endUserAuthHandler.Register
/internal/end-user/auth/login      POST  → endUserAuthHandler.Login
/internal/end-user/auth/logout     POST  → endUserAuthHandler.Logout
/internal/end-user/auth/refresh    POST  → endUserAuthHandler.Refresh
/internal/end-user/auth/me         GET   → endUserAuthHandler.Me

/internal/end-users                POST  → endUserMgmtHandler.Create
/internal/end-users                GET   → endUserMgmtHandler.List
/internal/end-users/{userId}/status PATCH → endUserMgmtHandler.UpdateStatus
/internal/end-users/{userId}       DELETE → endUserMgmtHandler.Delete
```

所有 `/internal/` 路由共用 `InternalTokenMiddleware`（校验 `X-Internal-Token` Header）。

#### Handler 目录结构

```
internal/interfaces/http/handlers/
└── enduser/
    ├── auth_handler.go       # Register / Login / Logout / Refresh / Me
    ├── management_handler.go # Create / List / UpdateStatus / Delete
    └── dto.go                # 请求/响应 JSON 结构体
```

#### AuthHandler 结构体

```go
type AuthHandler struct {
    authService    *enduser.EndUserAuthAppService
    mgmtService    *enduser.EndUserManagementAppService
    logger         logfacade.Logger
}
```

**Handler 标准模式**（对齐现有 auth handlers）：

```
1. 解析 JSON 请求体 → 返回 400 PARAM_INVALID（格式错误）
2. 从 Header 提取 orgName / projectSlug / userId（me 接口）
3. 调用 App 层 Use Case
4. BusinessError → writeJSONError(w, bizErr)
5. 成功 → writeJSON(w, result)
```

**错误响应格式**（对齐 api-contract.md）：

```json
{ "error": { "code": "INVALID_CREDENTIALS", "message": "..." } }
```

**`me` 接口的 Header 提取**：

```
X-End-User-Id:    {userId}
X-Org-Name:       {orgName}
X-Project-Slug:   {projectSlug}
```

BFF 验证 JWT 后将这三个值写入 Header，Handler 直接读取，不再验证 JWT。

---

### 5.2 GraphQL Resolver（Project Schema）

#### Schema 文件

新增 `api/graph/project/schema/end_user.graphql`（内容见 api-contract.md §2.6）。

运行 `just generate-gql` 生成 Resolver 骨架，**禁止运行 `just clean-gql`**。

#### Resolver 目录结构

```
internal/interfaces/graphql/project/
├── end_user.resolvers.go     # 新增：EndUser GraphQL Resolver 实现
└── adapter/
    └── end_user_error_adapter.go  # 新增：BusinessError → GraphQL Union 映射
```

#### Resolver 结构体

GraphQL Resolver 依赖 Application 层服务，而不是直接调用内网 HTTP 接口（GraphQL Resolver 和 HTTP Handler 共享同一 App Service 实例，通过 DI 注入）：

```go
// end_user.resolvers.go
// 这些方法挂载在现有的 projectResolver 上（或独立的 endUserResolver）

func (r *mutationResolver) CreateEndUser(ctx context.Context, input generated.CreateEndUserInput) (*generated.CreateEndUserPayload, error)
func (r *mutationResolver) UpdateEndUserStatus(ctx context.Context, input generated.UpdateEndUserStatusInput) (*generated.UpdateEndUserStatusPayload, error)
func (r *mutationResolver) DeleteEndUser(ctx context.Context, input generated.DeleteEndUserInput) (*generated.DeleteEndUserPayload, error)
func (r *queryResolver) ListEndUsers(ctx context.Context, input *generated.ListEndUsersInput) (*generated.ListEndUsersPayload, error)
```

#### Resolver 从 Context 提取参数

GraphQL Project Schema 的所有 Resolver 通过 `ctxutils` 从请求 Context 提取 `orgName` 和 `projectSlug`：

```go
orgName := ctxutils.GetOrgName(ctx)       // 已有模式，复用
projectSlug := ctxutils.GetProjectSlug(ctx) // 已有模式，复用
```

#### 错误适配器（end_user_error_adapter.go）

```go
// CreateEndUserError → GraphQL Union 映射
func (a *EndUserErrorAdapter) ConvertToCreateError(bizErr *bizerrors.BusinessError) generated.CreateEndUserError
```

映射规则：

| BusinessError 码 | GraphQL Union 类型 |
|------------------|-------------------|
| `CONFLICT.END_USER` | `EndUserAlreadyExists` |
| `PARAM_INVALID.END_USER` (密码) | `EndUserPasswordTooWeak` |
| `PARAM_INVALID.END_USER` (其他) | `InvalidInput` |
| `CLUSTER_NOT_CONFIGURED.END_USER` | `ClusterNotFound` |
| `NOT_FOUND.PROJECT` | `ProjectNotFound` |
| `NOT_FOUND.END_USER` | `EndUserNotFound` |
| 其他 | 透传 `error`（系统异常） |

---

### 5.3 DI 组装（DesignHandlers 扩展）

在 `internal/interfaces/http/routes.go` 的 `DesignHandlers` 结构体追加：

```go
// End-User Services
EndUserAuthAppService   *enduser.EndUserAuthAppService
EndUserMgmtAppService   *enduser.EndUserManagementAppService
```

在路由初始化时追加 `PrivateDBManager` 的构建和两个 App Service 的初始化。

---

## 6. 安全设计

### 6.1 密码存储

- 算法：`bcrypt`，`cost=12`（生产推荐值，防爆破）
- 位置：`HashedPassword.NewHashedPasswordFromPlain`，仅在 Domain 层触发
- 明文密码生命周期：仅存在于 HTTP Handler 的请求解析到 Use Case 调用的短暂栈空间，不落 DB、不落日志

### 6.2 Refresh Token 生成与存储

```
生成：crypto/rand.Read(32 bytes) → base64.RawURLEncoding → opaqueToken（43 字符）
存储：sha256(opaqueToken) → hex string → accounts.refresh_token_hash
传输：opaqueToken 明文，仅在 Go Backend 响应体出现一次
```

**为什么用 sha256 存储**：防止数据库泄露时，攻击者无法直接使用 hash 冒充 token（sha256 不可逆推 opaqueToken）。

### 6.3 Token Rotation

- **触发时机**：`/refresh` 接口调用时
- **原子性保证**：旧 token Revoke + 新 token Insert 在同一事务（`TxManager.WithTx`）
- **撤销后旧 token**：`accounts.revoked=1`，下次 GetByTokenHash 后 `IsValid()` 返回 false → `INVALID_REFRESH_TOKEN`

### 6.4 禁用账号处理

- 禁用后已有 access token（1h）会在自然过期前仍有效（MVP 可接受）
- **登录时**：`LoginEndUser` 在 `VerifyPassword` 后立即检查 `IsActive()`，禁用账号无法获取新 token
- **`me` 接口**：`GetEndUserMe` 检查 `IsActive()`，返回 `ACCOUNT_DISABLED`，前端刷新页面会感知
- v2 增强（非本期）：可在 `refresh` 接口也检查 `is_forbidden`，提前终止会话

### 6.5 内网接口鉴权

所有 `/internal/` 接口使用 `InternalTokenMiddleware`：

```
X-Internal-Token: {shared-secret}
```

- shared-secret 通过环境变量 `INTERNAL_TOKEN` 配置，不硬编码
- 中间件：常量时间比较（`subtle.ConstantTimeCompare`），防止时序攻击
- 现有项目已有此中间件，直接复用

### 6.6 防枚举

- 用户名不存在 vs 密码错误：统一返回 `INVALID_CREDENTIALS`，不区分
- 登录失败不记录用户名到响应体

---

## 7. 实现顺序

### Phase 0：先决条件确认（1 天）

```
[ ] 确认 plans/end-user-auth-backend-plan.md 方案已评审
[ ] 确认新增错误码定义位置（pkg/bizerrors/common_errors.go）
[ ] 确认 PrivateDBManager 与现有 ClusterConnectionManager 的共存方式
```

---

### Phase 1：Domain 层（可独立，无依赖）

```
[ P1-1 ] internal/domain/enduser/end_user.go          — EndUser 实体 + HashedPassword
[ P1-2 ] internal/domain/enduser/end_user_session.go  — EndUserSession 实体
[ P1-3 ] internal/domain/enduser/credential.go        — Credential 值对象
[ P1-4 ] internal/domain/enduser/end_user_repository.go
[ P1-5 ] internal/domain/enduser/end_user_session_repository.go
[ P1-6 ] pkg/bizerrors/common_errors.go               — 追加错误码定义
```

**P1-1 ~ P1-6 完全并行**，无相互依赖。

---

### Phase 2：Infrastructure 层（依赖 Phase 1）

```
[ P2-1 ] internal/infrastructure/database/private/migrator.go  — DDL 自动初始化
[ P2-2 ] internal/infrastructure/repository/private_db_manager.go
         （依赖 P2-1 + ClusterConnectionManager）
[ P2-3 ] internal/infrastructure/repository/sql_enduser_repository.go
         （依赖 P1-4）
[ P2-4 ] internal/infrastructure/repository/sql_enduser_session_repository.go
         （依赖 P1-5）
```

**P2-1 和 P2-3/P2-4 并行**；P2-2 需等 P2-1 完成。

---

### Phase 3：Application 层（依赖 Phase 1 + Phase 2）

```
[ P3-1 ] internal/app/enduser/commands.go
[ P3-2 ] internal/app/enduser/end_user_auth_service.go
         — Register / Login / Logout / Refresh / Me
[ P3-3 ] internal/app/enduser/end_user_app_service.go
         — Create / List / UpdateStatus / Delete
```

P3-1 先完成，P3-2 和 P3-3 可并行。

---

### Phase 4：接口层（依赖 Phase 3）

```
[ P4-1 ] api/graph/project/schema/end_user.graphql — Schema 文件
[ P4-2 ] just generate-gql                         — 生成 GraphQL 代码
[ P4-3 ] internal/interfaces/http/handlers/enduser/auth_handler.go
[ P4-4 ] internal/interfaces/http/handlers/enduser/management_handler.go
[ P4-5 ] internal/interfaces/graphql/project/end_user.resolvers.go
         （依赖 P4-2）
[ P4-6 ] internal/interfaces/graphql/project/adapter/end_user_error_adapter.go
[ P4-7 ] internal/interfaces/http/routes.go        — 路由注册 + DI 组装
```

P4-1 先完成，P4-2 后运行；P4-3/P4-4/P4-6 可并行；P4-5 依赖 P4-2；P4-7 最后完成。

---

### Phase 5：测试（与 Phase 4 并行推进）

```
[ P5-1 ] Domain 单元测试：HashedPassword.Verify / EndUserSession.IsValid
[ P5-2 ] Repository 集成测试（可选，需 Docker MySQL）
[ P5-3 ] BDD 验收测试：对齐 03-backend-design.md AC-1 ~ AC-11
[ P5-4 ] HTTP Handler 测试：e2e 通过 curl / httptest
```

---

## 8. 验收口径

对齐 `03-backend-design.md` 和 `api-contract.md` 的 AC 条目：

| AC # | 测试场景 | 预期结果 | 验证接口 |
|------|----------|----------|----------|
| **AC-1** | Project 未关联 Cluster，调用任意 end-user 接口 | HTTP 503，`code: CLUSTER_NOT_CONFIGURED` | POST /internal/end-user/auth/login |
| **AC-2** | 首次在 Project 下创建终端用户（private 库不存在） | private_{slug} 库自动建库建表，201 成功 | POST /internal/end-users |
| **AC-3** | 正确凭证登录 | 200，返回 `{ userId, refreshToken, expiresAt }` | POST /internal/end-user/auth/login |
| **AC-4** | 错误密码登录 | 401，`code: INVALID_CREDENTIALS` | POST /internal/end-user/auth/login |
| **AC-5** | 被禁用账号登录 | 403，`code: ACCOUNT_DISABLED` | POST /internal/end-user/auth/login |
| **AC-6** | 正确 refresh token 调用 refresh | 200，新 token 有效；旧 token 返回 401 | POST /internal/end-user/auth/refresh |
| **AC-7** | 使用已 revoked 的旧 refresh token | 401，`code: INVALID_REFRESH_TOKEN` | POST /internal/end-user/auth/refresh |
| **AC-8** | logout 后旧 token 再用 | 401，`code: INVALID_REFRESH_TOKEN` | POST /internal/end-user/auth/refresh |
| **AC-9** | 删除用户后该用户所有 sessions revoked | 旧 refresh token → 401 | DELETE /internal/end-users/{userId} + POST refresh |
| **AC-10** | 两个 Project（crm, erp）各注册同名用户 alice | 各自独立，互不干扰 | 分别 POST /internal/end-user/auth/login |
| **AC-11** | 注册用户名已存在 | 409，`code: CONFLICT.END_USER` | POST /internal/end-users 或 POST register |
| **AC-12** | 密码弱（不含数字/字母/不足 8 位） | 400，`code: PARAM_INVALID.END_USER` | POST register |
| **AC-13** | 用户名格式不合法（含特殊字符） | 400，`code: PARAM_INVALID.END_USER` | POST register |
| **AC-14** | me 接口查询被禁用用户 | 403，`code: ACCOUNT_DISABLED` | GET /internal/end-user/auth/me |
| **AC-15** | GraphQL listEndUsers 返回分页列表 | 含 nodes / pageInfo / totalCount | GraphQL Project |
| **AC-16** | GraphQL createEndUser 不自动登录 | 201，endUser 对象，无 refreshToken | GraphQL Project |
| **AC-17** | GraphQL updateEndUserStatus 禁用账号 | endUser.isForbidden=true | GraphQL Project |
| **AC-18** | GraphQL deleteEndUser | success=true | GraphQL Project |
| **AC-19** | Token rotation 事务性：Revoke 成功但 Insert 失败时旧 token 仍有效 | 旧 token 不应被 revoke（事务回滚） | 模拟 DB 注入错误 |

---

## 9. 主要风险与回滚思路

### 风险 1：PrivateDBManager 连接池泄露

**描述**：`private_{projectSlug}` 的 `*sql.DB` 连接池被创建后，如果 Cluster 被删除或重新配置，缓存连接无法感知变化，可能持续使用失效连接。

**缓解措施**：
1. 复用 `ClusterConnectionManager` 的 Ping 健康检查模式，连接失效时自动重建
2. 设置 `ConnMaxLifetime`（如 30min），连接超期强制重建
3. 如 Cluster 配置变更，可通过 `PrivateDBManager.EvictCache(orgName, projectSlug)` 主动清除缓存（供 Cluster 更新事件调用，v2 增强）

**回滚**：不需要回滚数据，连接失败会以 `SYSTEM_ERROR` 返回，用户重试后会重建连接。

---

### 风险 2：PrivateMigrator 并发初始化竞态

**描述**：高并发场景下，多个请求同时首次访问同一 `private_{projectSlug}`，可能同时触发 `CREATE DATABASE` + 建表，导致冲突或重复执行。

**缓解措施**：
1. 使用 `sync.Mutex` 或 `singleflight.Group` 保证同一 `projectSlug` 的 Migrate 只执行一次
2. 所有 DDL 使用 `IF NOT EXISTS`，即使并发执行也不会报错（MySQL 的 DDL 是幂等的）
3. `migrated sync.Map` 记录已完成初始化的库，跳过重复执行

**回滚**：无需回滚，DDL 幂等，并发重复执行安全。

---

### 风险 3：bcrypt 成本过高导致登录响应慢

**描述**：`cost=12` 在现代服务器上约 250-400ms（合理），但如果部署环境 CPU 弱，可能超过前端期望的响应时间。

**缓解措施**：
1. 登录接口本身允许 500ms 响应时间（认证类接口可接受）
2. 用户管理（创建用户）是开发者操作，对响应时间不敏感
3. 若 MVP 阶段发现性能问题，可调整 `cost=10`（对安全影响可接受）

**回滚**：不涉及数据迁移，修改 cost 只影响新密码哈希，不影响旧记录验证。

---

### 风险 4：`private_{projectSlug}` DDL 与 mc_meta 使用同一 MySQL 实例

**描述**：`private_{projectSlug}` 库与 Project 关联的 DatabaseCluster 在**同一 MySQL 实例**上。如果开发者将 Cluster 配置成 mc_meta 所在的实例，`private_` 库会建在同一实例上，可能引发性能竞争。

**缓解措施**：
1. PRD 明确设计如此（终端用户数据存在与 Project 关联的 cluster 上），不做额外隔离
2. MVP 阶段记录此风险，v2 可考虑限制 private 库不能建在 mc_meta 实例上
3. 库名前缀 `private_` 明确与 mc_meta 的表名区分

**回滚**：无法回滚已建立的库，但可通过重新配置 Cluster 迁移（v2 功能）。

---

### 风险 5：GraphQL Resolver 上下文中 ProjectSlug 获取失败

**描述**：GraphQL Project Schema 的 Resolver 需要从 ctx 提取 `orgName` 和 `projectSlug`。若 `ctxutils.GetProjectSlug(ctx)` 返回空值（中间件未设置），会导致 App 层路由失败。

**缓解措施**：
1. 在 Resolver 开始时做防御性检查：若 `projectSlug == ""` 立即返回 `InvalidInput` 错误
2. 验证 Project GraphQL 的中间件确实已设置 projectSlug 到 Context（参照现有 Resolver 的实现模式）
3. App 层 `ValidateProjectScope` 也会校验 `orgName` 和 `projectSlug` 非空

**回滚**：不涉及数据，修复中间件配置即可。

---

## 附录：新增错误码定义

需在 `pkg/bizerrors/common_errors.go` 追加：

```go
// End-User Auth 错误码
EndUserInvalidCredentials  = ErrorDefinition{Code: "INVALID_CREDENTIALS",         Domain: "END_USER"}  // → 401
EndUserInvalidRefreshToken = ErrorDefinition{Code: "INVALID_REFRESH_TOKEN",        Domain: "END_USER"}  // → 401
EndUserAccountDisabled     = ErrorDefinition{Code: "ACCOUNT_DISABLED",             Domain: "END_USER"}  // → 403
EndUserConflict            = ErrorDefinition{Code: "CONFLICT",                     Domain: "END_USER"}  // → 409
EndUserParamInvalid        = ErrorDefinition{Code: "PARAM_INVALID",                Domain: "END_USER"}  // → 400
EndUserNotFound            = ErrorDefinition{Code: "NOT_FOUND",                    Domain: "END_USER"}  // → 404
EndUserClusterNotConfigured = ErrorDefinition{Code: "CLUSTER_NOT_CONFIGURED",      Domain: "END_USER"}  // → 503
```

> HTTP 状态码映射需在 `pkg/bizerrors/business_error.go` 的 `GetHTTPStatusCode()` 中追加 `CLUSTER_NOT_CONFIGURED` → 503 和 `ACCOUNT_DISABLED` → 403 的映射规则（若不存在）。

---

*文档版本：2026-04-16 v1.0*
