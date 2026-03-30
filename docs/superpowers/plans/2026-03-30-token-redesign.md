# Token 体系重构实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有无状态双 Token 升级为现代化 SaaS 标准方案：Web 走 BFF + httpOnly Cookie，CLI 走 API Key 直连 Go Backend。

**Architecture:** Go Backend 负责所有数据存储和 token 生成；Next.js BFF 负责 Casdoor OAuth2 交互和 Cookie 生命周期；两端共享 JWT_SECRET 签发/验签 Access Token。

**Tech Stack:** Go (chi, sqlc, HMAC-SHA256), Next.js API Routes (TypeScript), MySQL

**Spec:** `docs/superpowers/specs/2026-03-30-token-design.md`

---

## 实现顺序

```
Part A: Go Backend（先做，BFF 依赖它）
  Phase A1 → 数据库 Schema
  Phase A2 → Domain + Repository 层
  Phase A3 → Application 层（auth 端点逻辑）
  Phase A4 → Interfaces 层（HTTP handler）
  Phase A5 → 中间件改造（移除 Casdoor 兼容层，新增 API Key 路径）
  Phase A6 → API Key 领域（GraphQL CRUD）
  Phase A7 → 数据清理 goroutine

Part B: Next.js BFF（Go Backend 就绪后做）
  Phase B1 → Go 内部 auth client
  Phase B2 → BFF auth 端点（/bff/auth/*）
  Phase B3 → 前端 token 存储迁移（localStorage → 内存 + httpOnly Cookie）
  Phase B4 → Auth Provider 适配
```

---

## Chunk 1: Go Backend — 数据库与基础层

### 每个阶段的验收标准

每个 Phase 完成后必须通过以下验收：
- 相关单元测试全部通过（`just test-unit`）
- lint 无报错（`just lint`）
- 验收命令明确列出

---

### Phase A1：数据库 Schema

**涉及文件：**
- 新建：`modelcraft-backend/db/schema/mysql/08_refresh_tokens.sql`
- 新建：`modelcraft-backend/db/schema/mysql/09_api_keys.sql`
- 新建：`modelcraft-backend/db/schema/mysql/10_security_audit_logs.sql`

- [ ] **A1-1：新建 refresh_tokens schema 文件**

```sql
-- modelcraft-backend/db/schema/mysql/08_refresh_tokens.sql
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          VARCHAR(36)  NOT NULL,
    user_id     VARCHAR(36)  NOT NULL,
    token_hash  VARCHAR(64)  NOT NULL COMMENT 'SHA256 hash，不存明文',
    expires_at  DATETIME     NOT NULL,
    created_at  DATETIME     NOT NULL,
    revoked_at  DATETIME     NULL     COMMENT 'NULL = 有效',
    PRIMARY KEY (id),
    INDEX idx_token_hash (token_hash),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **A1-2：新建 api_keys schema 文件**

```sql
-- modelcraft-backend/db/schema/mysql/09_api_keys.sql
CREATE TABLE IF NOT EXISTS api_keys (
    id           VARCHAR(36)  NOT NULL,
    user_id      VARCHAR(36)  NOT NULL,
    name         VARCHAR(100) NOT NULL COMMENT '用户命名，如 GitHub Actions',
    key_hash     VARCHAR(64)  NOT NULL COMMENT 'SHA256 hash，不存明文',
    key_prefix   VARCHAR(10)  NOT NULL COMMENT '完整 key 前 10 位，如 mc_a1b2c3d4',
    last_used_at DATETIME     NULL     COMMENT '防抖：距上次 > 1 分钟才更新',
    expires_at   DATETIME     NULL     COMMENT 'NULL = 永不过期',
    created_at   DATETIME     NOT NULL,
    revoked_at   DATETIME     NULL     COMMENT 'NULL = 有效',
    PRIMARY KEY (id),
    INDEX idx_key_hash (key_hash),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **A1-3：新建 security_audit_logs schema 文件**

```sql
-- modelcraft-backend/db/schema/mysql/10_security_audit_logs.sql
CREATE TABLE IF NOT EXISTS security_audit_logs (
    id         VARCHAR(36) NOT NULL,
    user_id    VARCHAR(36) NOT NULL,
    event      VARCHAR(50) NOT NULL COMMENT '如 REUSE_DETECTED',
    detail     JSON        NULL     COMMENT 'token_id, ip 等上下文',
    created_at DATETIME    NOT NULL,
    PRIMARY KEY (id),
    INDEX idx_user_id_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **A1-4：应用 schema 到本地数据库**

```bash
cd modelcraft-backend
just db up
just db status
```

预期输出：显示三个新表已创建，无报错。

- [ ] **A1-5：commit**

```bash
cd modelcraft-backend
git add db/schema/mysql/08_refresh_tokens.sql \
        db/schema/mysql/09_api_keys.sql \
        db/schema/mysql/10_security_audit_logs.sql
git commit -m "feat(db): add refresh_tokens, api_keys, security_audit_logs tables"
```

**✅ Phase A1 验收：**
```bash
just db status   # 确认三张新表存在
```

---

### Phase A2：sqlc 查询 + Domain 层

**涉及文件：**
- 新建：`modelcraft-backend/db/queries/refresh_tokens.sql`
- 新建：`modelcraft-backend/db/queries/api_keys.sql`
- 新建：`modelcraft-backend/db/queries/security_audit_logs.sql`
- 新建：`modelcraft-backend/internal/domain/auth/refresh_token.go`
- 新建：`modelcraft-backend/internal/domain/auth/api_key.go`
- 修改：`modelcraft-backend/internal/domain/auth/user_claims.go`（简化 claims，只保留 userId）

- [ ] **A2-1：新建 refresh_tokens sqlc 查询**

```sql
-- modelcraft-backend/db/queries/refresh_tokens.sql

-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
VALUES (?, ?, ?, ?, NOW());

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
FROM refresh_tokens
WHERE token_hash = ?
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = ?;

-- name: RevokeAllRefreshTokensByUserID :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = ? AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE (expires_at < DATE_SUB(NOW(), INTERVAL 30 DAY))
   OR (revoked_at IS NOT NULL AND revoked_at < DATE_SUB(NOW(), INTERVAL 30 DAY));
```

- [ ] **A2-2：新建 api_keys sqlc 查询**

```sql
-- modelcraft-backend/db/queries/api_keys.sql

-- name: InsertAPIKey :exec
INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, NOW());

-- name: GetAPIKeyByHash :one
SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at, revoked_at
FROM api_keys
WHERE key_hash = ?
LIMIT 1;

-- name: ListAPIKeysByUserID :many
SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at, revoked_at
FROM api_keys
WHERE user_id = ? AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: CountActiveAPIKeysByUserID :one
SELECT COUNT(*) FROM api_keys
WHERE user_id = ? AND revoked_at IS NULL;

-- name: RevokeAPIKey :exec
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = ? AND user_id = ?;

-- name: UpdateAPIKey :exec
UPDATE api_keys
SET name = ?, expires_at = ?
WHERE id = ? AND user_id = ?;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = ?;

-- name: GetAPIKeyByID :one
SELECT id, user_id, name, key_hash, key_prefix, last_used_at, expires_at, created_at, revoked_at
FROM api_keys
WHERE id = ?
LIMIT 1;

-- name: DeleteRevokedAPIKeys :exec
DELETE FROM api_keys
WHERE revoked_at IS NOT NULL AND revoked_at < DATE_SUB(NOW(), INTERVAL 90 DAY);
```

- [ ] **A2-3：新建 security_audit_logs sqlc 查询**

```sql
-- modelcraft-backend/db/queries/security_audit_logs.sql

-- name: InsertSecurityAuditLog :exec
INSERT INTO security_audit_logs (id, user_id, event, detail, created_at)
VALUES (?, ?, ?, ?, NOW());
```

- [ ] **A2-4：运行 sqlc 生成代码**

```bash
cd modelcraft-backend
just generate-sqlc
```

预期：`internal/infrastructure/dbgen/` 下生成新的类型安全查询函数，无报错。

- [ ] **A2-5：新建 RefreshToken domain 实体**

```go
// modelcraft-backend/internal/domain/auth/refresh_token.go
package auth

import "time"

// RefreshToken 代表一个有效的刷新令牌记录
type RefreshToken struct {
    ID        string
    UserID    string
    TokenHash string    // SHA256(明文 token)
    ExpiresAt time.Time
    CreatedAt time.Time
    RevokedAt *time.Time // nil = 有效
}

func (t *RefreshToken) IsValid() bool {
    return t.RevokedAt == nil && time.Now().Before(t.ExpiresAt)
}

func (t *RefreshToken) IsRevoked() bool {
    return t.RevokedAt != nil
}
```

- [ ] **A2-6：新建 APIKey domain 实体**

```go
// modelcraft-backend/internal/domain/auth/api_key.go
package auth

import "time"

const APIKeyMaxPerUser = 20

// APIKey 代表一个 CLI/CI 使用的 API Key 记录
type APIKey struct {
    ID         string
    UserID     string
    Name       string
    KeyHash    string     // SHA256(明文 key)
    KeyPrefix  string     // 完整 key 前 10 位，如 "mc_a1b2c3d4"
    LastUsedAt *time.Time
    ExpiresAt  *time.Time // nil = 永不过期
    CreatedAt  time.Time
    RevokedAt  *time.Time // nil = 有效
}

func (k *APIKey) IsValid() bool {
    if k.RevokedAt != nil {
        return false
    }
    if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
        return false
    }
    return true
}
```

- [ ] **A2-7：修改 UserClaims，只保留 userId**

打开 `internal/domain/auth/user_claims.go`，将 claims 简化为：

```go
// internal/domain/auth/user_claims.go
package auth

import "github.com/golang-jwt/jwt/v5"

// UserClaims 是 Access Token 的 JWT payload，只含身份信息。
// 权限信息由 Go 中间件实时从 Casbin 加载，不放入 token。
type UserClaims struct {
    jwt.RegisteredClaims
    UserID string `json:"user_id"`
}
```

- [ ] **A2-8：为 RefreshToken 和 APIKey 写单元测试**

```go
// modelcraft-backend/internal/domain/auth/refresh_token_test.go
package auth_test

import (
    "testing"
    "time"

    domainauth "modelcraft/internal/domain/auth"
)

func TestRefreshToken_IsValid(t *testing.T) {
    t.Run("valid token", func(t *testing.T) {
        token := &domainauth.RefreshToken{
            ExpiresAt: time.Now().Add(time.Hour),
            RevokedAt: nil,
        }
        if !token.IsValid() {
            t.Error("expected valid")
        }
    })

    t.Run("revoked token", func(t *testing.T) {
        now := time.Now()
        token := &domainauth.RefreshToken{
            ExpiresAt: time.Now().Add(time.Hour),
            RevokedAt: &now,
        }
        if token.IsValid() {
            t.Error("expected invalid")
        }
    })

    t.Run("expired token", func(t *testing.T) {
        token := &domainauth.RefreshToken{
            ExpiresAt: time.Now().Add(-time.Hour),
            RevokedAt: nil,
        }
        if token.IsValid() {
            t.Error("expected invalid")
        }
    })
}
```

- [ ] **A2-9：运行单元测试**

```bash
cd modelcraft-backend
just test-unit-pkg ./internal/domain/auth/...
```

预期：全部 PASS。

- [ ] **A2-10：commit**

```bash
cd modelcraft-backend
git add db/queries/ \
        internal/domain/auth/refresh_token.go \
        internal/domain/auth/api_key.go \
        internal/domain/auth/user_claims.go \
        internal/infrastructure/dbgen/
git commit -m "feat(auth): add refresh token and api key domain entities + sqlc queries; simplify JWT claims to userId only"
```

**✅ Phase A2 验收：**
```bash
just test-unit-pkg ./internal/domain/auth/...   # 全部 PASS
just lint                                        # 无报错
```

---

### Phase A3：Repository 层

**涉及文件：**
- 新建：`modelcraft-backend/internal/infrastructure/repository/sql_refresh_token_repository.go`
- 新建：`modelcraft-backend/internal/infrastructure/repository/sql_api_key_repository.go`
- 新建：`modelcraft-backend/internal/domain/auth/refresh_token_repository.go`（接口）
- 新建：`modelcraft-backend/internal/domain/auth/api_key_repository.go`（接口）
- 新建：`modelcraft-backend/internal/domain/auth/security_audit_log_repository.go`（接口）
- 新建：`modelcraft-backend/internal/infrastructure/repository/sql_security_audit_log_repository.go`

- [ ] **A3-1：定义 RefreshTokenRepository 接口**

```go
// modelcraft-backend/internal/domain/auth/refresh_token_repository.go
package auth

import "context"

// RefreshTokenRepository 定义刷新令牌的存储接口
type RefreshTokenRepository interface {
    // Save 保存新的 Refresh Token（hash 已在调用方计算）
    Save(ctx context.Context, token *RefreshToken) error
    // FindByHash 通过 hash 查找，未找到返回 (nil, nil)
    FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
    // Revoke 吊销指定 token
    Revoke(ctx context.Context, id string) error
    // RevokeAllByUserID 吊销该用户所有有效 token（盗用检测触发）
    RevokeAllByUserID(ctx context.Context, userID string) error
    // DeleteExpired 清理过期记录（由定时任务调用）
    DeleteExpired(ctx context.Context) error
}
```

- [ ] **A3-2：定义 APIKeyRepository 接口**

```go
// modelcraft-backend/internal/domain/auth/api_key_repository.go
package auth

import (
    "context"
    "time"
)

// APIKeyRepository 定义 API Key 的存储接口
type APIKeyRepository interface {
    // Save 保存新 API Key
    Save(ctx context.Context, key *APIKey) error
    // FindByHash 通过 hash 查找，未找到返回 (nil, nil)
    FindByHash(ctx context.Context, hash string) (*APIKey, error)
    // FindByID 通过 ID 查找，未找到返回 (nil, nil)
    FindByID(ctx context.Context, id string) (*APIKey, error)
    // ListByUserID 列出用户所有有效 key
    ListByUserID(ctx context.Context, userID string) ([]*APIKey, error)
    // CountActiveByUserID 统计有效 key 数量（用于限额检查）
    CountActiveByUserID(ctx context.Context, userID string) (int, error)
    // Revoke 吊销指定 key（幂等）
    Revoke(ctx context.Context, id string, userID string) error
    // Update 更新 key 名称或过期时间
    Update(ctx context.Context, id string, userID string, name string, expiresAt *time.Time) error
    // UpdateLastUsed 异步更新最后使用时间（防抖）
    UpdateLastUsed(ctx context.Context, id string) error
    // DeleteRevoked 清理已吊销记录（由定时任务调用）
    DeleteRevoked(ctx context.Context) error
}
```

- [ ] **A3-3：实现 SqlRefreshTokenRepository**

```go
// modelcraft-backend/internal/infrastructure/repository/sql_refresh_token_repository.go
package repository

import (
    "context"
    "time"
    domainauth "modelcraft/internal/domain/auth"
    "modelcraft/internal/domain/shared"
    "modelcraft/internal/infrastructure/dbgen"
)

type SqlRefreshTokenRepository struct {
    q dbgen.Querier
}

func NewSqlRefreshTokenRepository(q dbgen.Querier) domainauth.RefreshTokenRepository {
    return &SqlRefreshTokenRepository{q: q}
}

func (r *SqlRefreshTokenRepository) Save(ctx context.Context, token *domainauth.RefreshToken) error {
    return ExecWithErrorHandling(func() error {
        return r.q.InsertRefreshToken(ctx, dbgen.InsertRefreshTokenParams{
            ID:        token.ID,
            UserID:    token.UserID,
            TokenHash: token.TokenHash,
            ExpiresAt: token.ExpiresAt,
        })
    })
}

func (r *SqlRefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domainauth.RefreshToken, error) {
    var row dbgen.RefreshToken
    err := QueryWithSQLErrorHandling(func() error {
        var e error
        row, e = r.q.GetRefreshTokenByHash(ctx, hash)
        return e
    })
    if err != nil {
        if shared.IsNotFoundError(err) {
            return nil, nil
        }
        return nil, err
    }
    return toDomainRefreshToken(row), nil
}

func (r *SqlRefreshTokenRepository) Revoke(ctx context.Context, id string) error {
    return ExecWithErrorHandling(func() error {
        return r.q.RevokeRefreshToken(ctx, id)
    })
}

func (r *SqlRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
    return ExecWithErrorHandling(func() error {
        return r.q.RevokeAllRefreshTokensByUserID(ctx, userID)
    })
}

func (r *SqlRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
    return ExecWithErrorHandling(func() error {
        return r.q.DeleteExpiredRefreshTokens(ctx)
    })
}

func toDomainRefreshToken(row dbgen.RefreshToken) *domainauth.RefreshToken {
    return &domainauth.RefreshToken{
        ID:        row.ID,
        UserID:    row.UserID,
        TokenHash: row.TokenHash,
        ExpiresAt: row.ExpiresAt,
        CreatedAt: row.CreatedAt,
        RevokedAt: NullTimeToPtr(row.RevokedAt),
    }
}
```

- [ ] **A3-4：实现 SqlAPIKeyRepository**

注意：
- `CountActiveByUserID` 的 sqlc 生成返回类型是 `int64`，接口是 `int`，需要显式转换：`return int(count), nil`
- nullable 字段（`LastUsedAt`、`ExpiresAt`、`RevokedAt`）全部用 `NullTimeToPtr` 处理

```go
// modelcraft-backend/internal/infrastructure/repository/sql_api_key_repository.go
package repository

import (
    "context"
    "time"
    domainauth "modelcraft/internal/domain/auth"
    "modelcraft/internal/domain/shared"
    "modelcraft/internal/infrastructure/dbgen"
)

type SqlAPIKeyRepository struct {
    q dbgen.Querier
}

func NewSqlAPIKeyRepository(q dbgen.Querier) domainauth.APIKeyRepository {
    return &SqlAPIKeyRepository{q: q}
}

func (r *SqlAPIKeyRepository) Save(ctx context.Context, key *domainauth.APIKey) error {
    return ExecWithErrorHandling(func() error {
        return r.q.InsertAPIKey(ctx, dbgen.InsertAPIKeyParams{
            ID:        key.ID,
            UserID:    key.UserID,
            Name:      key.Name,
            KeyHash:   key.KeyHash,
            KeyPrefix: key.KeyPrefix,
            ExpiresAt: PtrToNullTime(key.ExpiresAt),
        })
    })
}

func (r *SqlAPIKeyRepository) FindByHash(ctx context.Context, hash string) (*domainauth.APIKey, error) {
    var row dbgen.ApiKey
    err := QueryWithSQLErrorHandling(func() error {
        var e error
        row, e = r.q.GetAPIKeyByHash(ctx, hash)
        return e
    })
    if err != nil {
        if shared.IsNotFoundError(err) {
            return nil, nil
        }
        return nil, err
    }
    return toDomainAPIKey(row), nil
}

func (r *SqlAPIKeyRepository) FindByID(ctx context.Context, id string) (*domainauth.APIKey, error) {
    var row dbgen.ApiKey
    err := QueryWithSQLErrorHandling(func() error {
        var e error
        row, e = r.q.GetAPIKeyByID(ctx, id)
        return e
    })
    if err != nil {
        if shared.IsNotFoundError(err) {
            return nil, nil
        }
        return nil, err
    }
    return toDomainAPIKey(row), nil
}

func (r *SqlAPIKeyRepository) ListByUserID(ctx context.Context, userID string) ([]*domainauth.APIKey, error) {
    var rows []dbgen.ApiKey
    err := QueryWithSQLErrorHandling(func() error {
        var e error
        rows, e = r.q.ListAPIKeysByUserID(ctx, userID)
        return e
    })
    if err != nil {
        return nil, err
    }
    result := make([]*domainauth.APIKey, len(rows))
    for i, row := range rows {
        result[i] = toDomainAPIKey(row)
    }
    return result, nil
}

func (r *SqlAPIKeyRepository) CountActiveByUserID(ctx context.Context, userID string) (int, error) {
    var count int64
    err := QueryWithSQLErrorHandling(func() error {
        var e error
        count, e = r.q.CountActiveAPIKeysByUserID(ctx, userID)
        return e
    })
    return int(count), err  // sqlc 返回 int64，接口要求 int
}

func (r *SqlAPIKeyRepository) Revoke(ctx context.Context, id string, userID string) error {
    return ExecWithErrorHandling(func() error {
        return r.q.RevokeAPIKey(ctx, dbgen.RevokeAPIKeyParams{ID: id, UserID: userID})
    })
}

func (r *SqlAPIKeyRepository) Update(ctx context.Context, id string, userID string, name string, expiresAt *time.Time) error {
    return ExecWithErrorHandling(func() error {
        return r.q.UpdateAPIKey(ctx, dbgen.UpdateAPIKeyParams{
            ID:        id,
            UserID:    userID,
            Name:      name,
            ExpiresAt: PtrToNullTime(expiresAt),
        })
    })
}

func (r *SqlAPIKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
    return ExecWithErrorHandling(func() error {
        return r.q.UpdateAPIKeyLastUsed(ctx, id)
    })
}

func (r *SqlAPIKeyRepository) DeleteRevoked(ctx context.Context) error {
    return ExecWithErrorHandling(func() error {
        return r.q.DeleteRevokedAPIKeys(ctx)
    })
}

func toDomainAPIKey(row dbgen.ApiKey) *domainauth.APIKey {
    return &domainauth.APIKey{
        ID:         row.ID,
        UserID:     row.UserID,
        Name:       row.Name,
        KeyHash:    row.KeyHash,
        KeyPrefix:  row.KeyPrefix,
        LastUsedAt: NullTimeToPtr(row.LastUsedAt),
        ExpiresAt:  NullTimeToPtr(row.ExpiresAt),
        CreatedAt:  row.CreatedAt,
        RevokedAt:  NullTimeToPtr(row.RevokedAt),
    }
}
```

- [ ] **A3-5：定义 SecurityAuditLogRepository 接口**

```go
// modelcraft-backend/internal/domain/auth/security_audit_log_repository.go
package auth

import (
    "context"
    "time"
)

// SecurityAuditEvent 安全审计事件类型
type SecurityAuditEvent string

const EventReuseDetected SecurityAuditEvent = "REUSE_DETECTED"

// SecurityAuditLog 安全事件记录
type SecurityAuditLog struct {
    ID        string
    UserID    string
    Event     SecurityAuditEvent
    Detail    map[string]any // token_id, ip 等上下文
    CreatedAt time.Time
}

// SecurityAuditLogRepository 安全审计日志存储接口
type SecurityAuditLogRepository interface {
    Insert(ctx context.Context, log *SecurityAuditLog) error
}
```

- [ ] **A3-6：实现 SqlSecurityAuditLogRepository**

```go
// modelcraft-backend/internal/infrastructure/repository/sql_security_audit_log_repository.go
package repository

import (
    "context"
    "encoding/json"
    domainauth "modelcraft/internal/domain/auth"
    "modelcraft/internal/infrastructure/dbgen"
)

type SqlSecurityAuditLogRepository struct {
    q dbgen.Querier
}

func NewSqlSecurityAuditLogRepository(q dbgen.Querier) domainauth.SecurityAuditLogRepository {
    return &SqlSecurityAuditLogRepository{q: q}
}

func (r *SqlSecurityAuditLogRepository) Insert(ctx context.Context, log *domainauth.SecurityAuditLog) error {
    detail, _ := json.Marshal(log.Detail)
    return ExecWithErrorHandling(func() error {
        return r.q.InsertSecurityAuditLog(ctx, dbgen.InsertSecurityAuditLogParams{
            ID:     log.ID,
            UserID: log.UserID,
            Event:  string(log.Event),
            Detail: detail,
        })
    })
}
```

- [ ] **A3-7：lint 检查**

```bash
cd modelcraft-backend
just lint
```

预期：无报错。

- [ ] **A3-8：commit**

```bash
cd modelcraft-backend
git add internal/domain/auth/refresh_token_repository.go \
        internal/domain/auth/api_key_repository.go \
        internal/domain/auth/security_audit_log_repository.go \
        internal/infrastructure/repository/sql_refresh_token_repository.go \
        internal/infrastructure/repository/sql_api_key_repository.go \
        internal/infrastructure/repository/sql_security_audit_log_repository.go
git commit -m "feat(auth): add refresh token, api key, and security audit log repository interfaces and implementations"
```

**✅ Phase A3 验收：**
```bash
just lint                   # 无报错
just build                  # 编译通过（repository 层需要 DB，不运行单元测试）
```

---

## Chunk 2: Go Backend — Application 层与 HTTP Handlers

### Phase A4：Application 层

**涉及文件：**
- 修改：`modelcraft-backend/internal/app/auth/token_service.go`（重构登录、刷新、登出逻辑）
- 新建：`modelcraft-backend/internal/app/auth/commands.go`
- 新建：`modelcraft-backend/internal/app/auth/token_generator.go`（Refresh Token 生成工具）

- [ ] **A4-1：新建 token_generator.go**

```go
// modelcraft-backend/internal/app/auth/token_generator.go
package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)

// GenerateRefreshToken 生成 32 字节 CSPRNG → 64 位 hex 字符串
func GenerateRefreshToken() (plaintext string, hash string, err error) {
    b := make([]byte, 32)
    if _, err = rand.Read(b); err != nil {
        return "", "", fmt.Errorf("generate refresh token: %w", err)
    }
    plaintext = hex.EncodeToString(b)
    sum := sha256.Sum256([]byte(plaintext))
    hash = hex.EncodeToString(sum[:])
    return plaintext, hash, nil
}

// base62Chars Base62 字符集（0-9, A-Z, a-z）
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateAPIKey 生成 API Key：mc_ 前缀 + 40 位 Base62（CSPRNG，约 238 bit 熵）
// 返回完整明文 key 和其 SHA256 hash
func GenerateAPIKey() (plaintext string, hash string, err error) {
    const length = 40
    b := make([]byte, length)
    if _, err = rand.Read(b); err != nil {
        return "", "", fmt.Errorf("generate api key: %w", err)
    }
    result := make([]byte, length)
    for i, v := range b {
        result[i] = base62Chars[int(v)%62]
    }
    plaintext = "mc_" + string(result)
    sum := sha256.Sum256([]byte(plaintext))
    hash = hex.EncodeToString(sum[:])
    return plaintext, hash, nil
}

// HashToken 对任意 token 字符串计算 SHA256 hash
func HashToken(token string) string {
    sum := sha256.Sum256([]byte(token))
    return hex.EncodeToString(sum[:])
}
```

- [ ] **A4-2：新建 commands.go**

```go
// modelcraft-backend/internal/app/auth/commands.go
package auth

import "time"

// LoginCommand BFF 登录时传入的用户信息（来自 Casdoor token）
type LoginCommand struct {
    ExternalID string
    Email      string
    Name       string
}

// LoginResult 登录成功后返回给 BFF
type LoginResult struct {
    UserID       string
    RefreshToken string    // 明文，BFF 存入 Cookie
    ExpiresAt    time.Time
}

// RefreshCommand BFF 刷新时传入
type RefreshCommand struct {
    RefreshToken string // 明文
}

// RefreshResult 刷新成功后返回给 BFF
type RefreshResult struct {
    UserID       string
    RefreshToken string    // 新明文 token
    ExpiresAt    time.Time
}

// LogoutCommand BFF 登出时传入
type LogoutCommand struct {
    RefreshToken string // 明文
}
```

- [ ] **A4-3：为 token_service.go 写单元测试（先写测试）**

```go
// modelcraft-backend/internal/app/auth/token_service_test.go
package auth_test

import (
    "context"
    "testing"
    "time"
    appauth "modelcraft/internal/app/auth"
    domainauth "modelcraft/internal/domain/auth"
)

// mockRefreshTokenRepo 用于测试
type mockRefreshTokenRepo struct {
    tokens map[string]*domainauth.RefreshToken
}

func (m *mockRefreshTokenRepo) Save(ctx context.Context, token *domainauth.RefreshToken) error {
    m.tokens[token.TokenHash] = token
    return nil
}

func (m *mockRefreshTokenRepo) FindByHash(ctx context.Context, hash string) (*domainauth.RefreshToken, error) {
    t, ok := m.tokens[hash]
    if !ok {
        return nil, nil
    }
    return t, nil
}

func (m *mockRefreshTokenRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
    for _, t := range m.tokens {
        if t.UserID == userID {
            now := time.Now()
            t.RevokedAt = &now
        }
    }
    return nil
}

func (m *mockRefreshTokenRepo) Revoke(ctx context.Context, id string) error {
    for _, t := range m.tokens {
        if t.ID == id {
            now := time.Now()
            t.RevokedAt = &now
        }
    }
    return nil
}

func (m *mockRefreshTokenRepo) DeleteExpired(_ context.Context) error { return nil }

// mockAuditLogRepo 用于测试审计日志
type mockAuditLogRepo struct {
    logs []*domainauth.SecurityAuditLog
}

func (m *mockAuditLogRepo) Insert(ctx context.Context, log *domainauth.SecurityAuditLog) error {
    m.logs = append(m.logs, log)
    return nil
}

func TestTokenService_Login(t *testing.T) {
    svc := appauth.NewTokenService(
        &mockRefreshTokenRepo{tokens: make(map[string]*domainauth.RefreshToken)},
        nil, // userRepo（使用现有实现）
        &mockAuditLogRepo{},
    )
    result, err := svc.Login(context.Background(), appauth.LoginCommand{
        ExternalID: "ext_123",
        Email:      "test@example.com",
        Name:       "Test User",
    })
    if err != nil {
        t.Fatalf("Login() error = %v", err)
    }
    if result.RefreshToken == "" {
        t.Error("Login() returned empty refresh token")
    }
    if result.ExpiresAt.Before(time.Now()) {
        t.Error("Login() returned already-expired token")
    }
}

func TestTokenService_Refresh_Rotation(t *testing.T) {
    repo := &mockRefreshTokenRepo{tokens: make(map[string]*domainauth.RefreshToken)}
    svc := appauth.NewTokenService(repo, nil, &mockAuditLogRepo{})

    loginResult, _ := svc.Login(context.Background(), appauth.LoginCommand{
        ExternalID: "ext_123", Email: "test@example.com", Name: "Test",
    })

    // 使用旧 token 做 refresh
    refreshResult, err := svc.Refresh(context.Background(), appauth.RefreshCommand{
        RefreshToken: loginResult.RefreshToken,
    })
    if err != nil {
        t.Fatalf("Refresh() error = %v", err)
    }
    if refreshResult.RefreshToken == loginResult.RefreshToken {
        t.Error("Refresh() should return a new token, not the same one")
    }

    // 旧 token 再次 refresh 应返回错误（reuse detection）
    _, err = svc.Refresh(context.Background(), appauth.RefreshCommand{
        RefreshToken: loginResult.RefreshToken,
    })
    if err == nil {
        t.Error("Refresh() with revoked token should return error")
    }
}

func TestTokenService_Refresh_ReuseDetection(t *testing.T) {
    repo := &mockRefreshTokenRepo{tokens: make(map[string]*domainauth.RefreshToken)}
    auditRepo := &mockAuditLogRepo{}
    svc := appauth.NewTokenService(repo, nil, auditRepo)

    loginResult, _ := svc.Login(context.Background(), appauth.LoginCommand{
        ExternalID: "ext_123", Email: "test@example.com", Name: "Test",
    })

    // 正常 refresh 一次，消耗原始 token
    svc.Refresh(context.Background(), appauth.RefreshCommand{RefreshToken: loginResult.RefreshToken})

    // 再次用原始（已 revoked）token refresh，触发盗用检测
    _, err := svc.Refresh(context.Background(), appauth.RefreshCommand{
        RefreshToken: loginResult.RefreshToken,
    })
    if err == nil {
        t.Fatal("expected error on revoked token reuse")
    }
    // 审计日志应有 REUSE_DETECTED 记录
    if len(auditRepo.logs) == 0 {
        t.Error("expected REUSE_DETECTED audit log entry")
    }
    if auditRepo.logs[0].Event != domainauth.EventReuseDetected {
        t.Errorf("expected event REUSE_DETECTED, got %s", auditRepo.logs[0].Event)
    }
}

func TestTokenService_Logout(t *testing.T) {
    repo := &mockRefreshTokenRepo{tokens: make(map[string]*domainauth.RefreshToken)}
    svc := appauth.NewTokenService(repo, nil, &mockAuditLogRepo{})

    loginResult, _ := svc.Login(context.Background(), appauth.LoginCommand{
        ExternalID: "ext_123", Email: "test@example.com", Name: "Test",
    })

    err := svc.Logout(context.Background(), appauth.LogoutCommand{
        RefreshToken: loginResult.RefreshToken,
    })
    if err != nil {
        t.Fatalf("Logout() error = %v", err)
    }

    // Logout 后再 Refresh 应失败
    _, err = svc.Refresh(context.Background(), appauth.RefreshCommand{
        RefreshToken: loginResult.RefreshToken,
    })
    if err == nil {
        t.Error("Refresh() after Logout() should return error")
    }
}
```

- [ ] **A4-4：运行测试确认失败**

```bash
cd modelcraft-backend
just test-unit-pkg ./internal/app/auth/...
```

预期：编译失败或 FAIL（因为 token_service.go 还未实现新接口）。

- [ ] **A4-5：重构 token_service.go 实现新接口**

核心职责：
1. `Login(ctx, cmd LoginCommand) (*LoginResult, error)` — 查/创 user，生成 Refresh Token 写 DB
2. `Refresh(ctx, cmd RefreshCommand) (*RefreshResult, error)` — 验证 → 盗用检测 → 轮换
3. `Logout(ctx, cmd LogoutCommand) error` — 吊销指定 token
4. 移除所有 Casdoor 相关逻辑（Casdoor 已移至 BFF）

注意：`Login` 中查/创 user 的逻辑保留（通过 externalId 查 user，不存在则创建）。

`Refresh()` 的骨架结构（必须严格实现此分支逻辑）：

```go
func (s *TokenService) Refresh(ctx context.Context, cmd RefreshCommand) (*RefreshResult, error) {
    // 1. 计算 hash，查 DB
    hash := HashToken(cmd.RefreshToken)
    token, err := s.refreshTokenRepo.FindByHash(ctx, hash)
    if err != nil {
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }

    // 2. token 不存在 → 401
    if token == nil {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthUnauthorized, "refresh token not found")
    }

    // 3. 已 revoked → 盗用检测：吊销该用户所有 token + 写审计日志 + 401
    if token.RevokedAt != nil {
        // 全设备强制下线
        _ = s.refreshTokenRepo.RevokeAllByUserID(ctx, token.UserID)
        // 写审计日志
        _ = s.auditLogRepo.Insert(ctx, &domainauth.SecurityAuditLog{
            ID:     bizutils.NewID(),
            UserID: token.UserID,
            Event:  domainauth.EventReuseDetected,
            Detail: map[string]any{"token_id": token.ID},
        })
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthUnauthorized, "token reuse detected")
    }

    // 4. 已过期 → 401
    if time.Now().After(token.ExpiresAt) {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthUnauthorized, "refresh token expired")
    }

    // 5. 正常轮换：revoke 旧 token，生成新 token
    if err := s.refreshTokenRepo.Revoke(ctx, token.ID); err != nil {
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }
    plaintext, newHash, err := GenerateRefreshToken()
    if err != nil {
        return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate new refresh token")
    }
    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    newToken := &domainauth.RefreshToken{
        ID:        bizutils.NewID(),
        UserID:    token.UserID,
        TokenHash: newHash,
        ExpiresAt: expiresAt,
        CreatedAt: time.Now(),
    }
    if err := s.refreshTokenRepo.Save(ctx, newToken); err != nil {
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }
    return &RefreshResult{
        UserID:       token.UserID,
        RefreshToken: plaintext,
        ExpiresAt:    expiresAt,
    }, nil
}
```

- [ ] **A4-6：运行测试确认通过**

```bash
cd modelcraft-backend
just test-unit-pkg ./internal/app/auth/...
```

预期：全部 PASS。

- [ ] **A4-7：commit**

```bash
cd modelcraft-backend
git add internal/app/auth/
git commit -m "feat(auth): refactor token service - stateful refresh token, remove Casdoor dependency"
```

**✅ Phase A4 验收：**
```bash
just test-unit-pkg ./internal/app/auth/...   # 全部 PASS
just lint                                     # 无报错
```

---

### Phase A5：HTTP Handler 层

**涉及文件：**
- 修改：`modelcraft-backend/internal/interfaces/http/handlers/auth/token_handler.go`
- 修改：`modelcraft-backend/internal/interfaces/http/handlers/auth/handler.go`
- 修改：`modelcraft-backend/api/openapi/auth.yaml`（更新端点定义）
- 修改：`modelcraft-backend/internal/middleware/chi_jwt_auth.go`（移除 Casdoor 兼容层）
- 新建：`modelcraft-backend/internal/middleware/api_key_verifier.go`（APIKeyVerifier 接口）

- [ ] **A5-1：更新 auth.yaml 定义新端点**

在 `api/openapi/auth.yaml` 中：
- 将 `POST /api/auth/token` 重命名为 `POST /api/auth/login`，更新请求/响应 schema
- 更新 `POST /api/auth/refresh` 请求 body（传 refreshToken 明文）
- 新增 `POST /api/auth/logout`

- [ ] **A5-2：重新生成 OpenAPI 代码**

```bash
cd modelcraft-backend
just generate-oapi
```

预期：`internal/interfaces/http/generated/` 更新，无报错。

- [ ] **A5-3：重构 token_handler.go**

实现三个新 handler：
- `HandleLogin` — 接收 `{ externalId, email, name }`，调用 `TokenService.Login`，返回 `{ userId, refreshToken, expiresAt }`
- `HandleRefresh` — 接收 `{ refreshToken }`，调用 `TokenService.Refresh`，返回新 token 对
- `HandleLogout` — 接收 `{ refreshToken }`，调用 `TokenService.Logout`，返回 204

- [ ] **A5-4：改造 chi_jwt_auth.go，移除 Casdoor 兼容层**

修改验证逻辑：
- 移除 `detectTokenIssuerChi()` 中的 Casdoor RSA 验证分支
- 只保留 ModelCraft HMAC-SHA256 验证
- 新增 API Key 验证分支（`mc_` 前缀识别）：

```go
func JWTAuthMiddleware(apiKeySvc APIKeyVerifier) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)
            if strings.HasPrefix(token, "mc_") {
                // API Key 路径：Phase A6 会注入真实实现；
                // 在 A6 完成前，此处必须返回 401，不能透传未认证请求。
                if apiKeySvc == nil {
                    http.Error(w, "Unauthorized", http.StatusUnauthorized)
                    return
                }
                userID, err := apiKeySvc.VerifyAPIKey(r.Context(), token)
                if err != nil || userID == "" {
                    http.Error(w, "Unauthorized", http.StatusUnauthorized)
                    return
                }
                ctx := ctxutils.SetUserID(r.Context(), userID)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
            // JWT 路径
            claims, err := validateModelCraftJWT(token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            ctx := ctxutils.SetUserID(r.Context(), claims.UserID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

`APIKeyVerifier` 在 A5 阶段以 `nil` 注入（返回 401），在 A6 阶段替换为真实实现：

```go
// internal/middleware/api_key_verifier.go
package middleware

import "context"

// APIKeyVerifier API Key 验证接口；由 A6 阶段的 APIKeyService 实现
type APIKeyVerifier interface {
    VerifyAPIKey(ctx context.Context, rawKey string) (userID string, err error)
}
```

- [ ] **A5-5：运行完整 lint + 编译检查**

```bash
cd modelcraft-backend
just lint
just build
```

预期：编译通过，lint 无报错。

- [ ] **A5-6：集成测试：手动验证新端点**

启动服务：
```bash
just run force=true
```

测试 login 端点（用测试 user 数据）：
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"externalId":"test_user","email":"test@example.com","name":"Test User"}'
```

预期：返回 `{ userId, refreshToken, expiresAt }`。

测试 refresh 端点：
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refreshToken":"<上一步返回的 token>"}'
```

预期：返回新 token 对，旧 token 已失效。

- [ ] **A5-7：commit**

```bash
cd modelcraft-backend
git add internal/interfaces/http/handlers/auth/ \
        internal/middleware/chi_jwt_auth.go \
        internal/middleware/api_key_verifier.go \
        api/openapi/auth.yaml \
        internal/interfaces/http/generated/
git commit -m "feat(auth): new login/refresh/logout handlers, remove Casdoor JWT compat layer"
```

**✅ Phase A5 验收：**
```bash
just build                         # 编译通过
just lint                          # 无报错
# 手动测试三个端点均返回预期响应
curl -X POST http://localhost:8080/api/auth/login   -d '...'  # 200
curl -X POST http://localhost:8080/api/auth/refresh -d '...'  # 200 新 token
curl -X POST http://localhost:8080/api/auth/logout  -d '...'  # 204
# 验证 revoked token 再次 refresh 返回 401
curl -X POST http://localhost:8080/api/auth/refresh -d '{"refreshToken":"<已用过的旧token>"}' # 401
```

---

## Chunk 3: Go Backend — API Key 领域

### Phase A6：API Key GraphQL

**涉及文件：**
- 修改：`modelcraft-backend/api/graph/org/schema/user_management.graphql`（新增 API Key 相关 type/query/mutation）
- 新建：`modelcraft-backend/api/graph/org/schema/api_key.graphql`
- 新建：`modelcraft-backend/internal/app/auth/api_key_service.go`
- 修改：`modelcraft-backend/internal/interfaces/graphql/org/*.resolvers.go`（新增 resolver）
- 修改：`modelcraft-backend/internal/interfaces/graphql/org/adapter/`（新增 error adapter）
- 修改：`modelcraft-backend/internal/middleware/chi_jwt_auth.go`（补全 API Key 验证逻辑）

- [ ] **A6-1：新建 api_key.graphql schema**

```graphql
# modelcraft-backend/api/graph/org/schema/api_key.graphql

extend type Query {
  apiKeys: [ApiKey!]! @hasPermission(action: "apikey:read")
}

extend type Mutation {
  createApiKey(input: CreateApiKeyInput!): CreateApiKeyPayload!
    @hasPermission(action: "apikey:create")
  revokeApiKey(id: ID!): RevokeApiKeyPayload!
    @hasPermission(action: "apikey:delete")
  updateApiKey(id: ID!, input: UpdateApiKeyInput!): UpdateApiKeyPayload!
    @hasPermission(action: "apikey:update")
}

type ApiKey {
  id:          ID!
  name:        String!
  keyPrefix:   String!
  lastUsedAt:  Time
  expiresAt:   Time
  revokedAt:   Time
  createdAt:   Time!
}

type CreateApiKeyResult {
  id:        ID!
  name:      String!
  key:       String!
  keyPrefix: String!
  createdAt: Time!
}

input CreateApiKeyInput {
  name:      String!
  expiresAt: Time
}

input UpdateApiKeyInput {
  name:      String
  expiresAt: Time
}

type CreateApiKeyPayload {
  result: CreateApiKeyResult
  error:  CreateApiKeyError
}

type RevokeApiKeyPayload {
  apiKey: ApiKey
  error:  RevokeApiKeyError
}

type UpdateApiKeyPayload {
  apiKey: ApiKey
  error:  UpdateApiKeyError
}
```

- [ ] **A6-2：在 errors.graphql 中新增 API Key 错误类型**

新建（或追加到已有的）`api/graph/org/schema/errors.graphql`。若该文件已存在，在末尾追加；若不存在，创建该文件：

```graphql
# api/graph/org/schema/errors.graphql
type ApiKeyLimitExceeded {
  message: String!
}

type ApiKeyNotFound {
  message: String!
}

union CreateApiKeyError = ApiKeyLimitExceeded | InvalidInput
union RevokeApiKeyError = ApiKeyNotFound
union UpdateApiKeyError = ApiKeyNotFound | InvalidInput
```

**注意：** `InvalidInput` 须在 org schema 中已定义（查看现有的 `base.graphql` 或 `errors.graphql`）。若尚不存在，在 `errors.graphql` 中同时添加：

```graphql
type InvalidInput {
  message: String!
}
```

- [ ] **A6-3：生成 GraphQL 代码**

```bash
cd modelcraft-backend
just generate-gql
```

预期：`internal/interfaces/graphql/org/generated/` 更新，新增 API Key 相关类型，无报错。

- [ ] **A6-4：新建 api_key_service.go（Application 层）**

```go
// modelcraft-backend/internal/app/auth/api_key_service.go
package auth

import (
    "context"
    domainauth "modelcraft/internal/domain/auth"
)

// APIKeyService 处理 API Key CRUD 用例
type APIKeyService struct {
    apiKeyRepo domainauth.APIKeyRepository
}

// ListAPIKeys 返回当前用户所有有效 key（只查未吊销的）
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID string) ([]*domainauth.APIKey, error)

// CreateAPIKey 创建新 key，返回包含明文 key 的结果（明文只返回一次）
func (s *APIKeyService) CreateAPIKey(ctx context.Context, cmd CreateAPIKeyCommand) (*CreateAPIKeyResult, error)
// 内部逻辑：检查限额（≤20）→ 生成 key（mc_ + 40位 Base62）→ 存 hash → 返回明文

// RevokeAPIKey 吊销指定 key（幂等）
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id string, userID string) (*domainauth.APIKey, error)

// UpdateAPIKey 更新 key 名称或过期时间
func (s *APIKeyService) UpdateAPIKey(ctx context.Context, cmd UpdateAPIKeyCommand) (*domainauth.APIKey, error)
```

- [ ] **A6-5：为 CreateAPIKey 写单元测试**

重点测试：
- 创建成功，明文 key 以 `mc_` 开头
- 创建成功，DB 中只存 hash（验证明文 ≠ hash）
- 超过 20 个限额时返回 `ApiKeyLimitExceeded` 错误
- 吊销后的 key 不计入限额，可继续创建

```bash
cd modelcraft-backend
just test-unit-pkg ./internal/app/auth/...
```

预期：全部 PASS。

- [ ] **A6-6：实现 GraphQL resolver**

在 `internal/interfaces/graphql/org/` 对应 resolver 文件中实现四个方法：
- `APIKeys` — 调用 `APIKeyService.ListAPIKeys`
- `CreateApiKey` — 调用 `APIKeyService.CreateAPIKey`
- `RevokeApiKey` — 调用 `APIKeyService.RevokeAPIKey`
- `UpdateApiKey` — 调用 `APIKeyService.UpdateAPIKey`

遵循现有 resolver 模式（错误用 `logfacade.Stack(err)` + adapter 转换）。

- [ ] **A6-7：补全中间件 API Key 验证逻辑**

在 `chi_jwt_auth.go` 的 `mc_` 分支中实现完整逻辑：

```go
if strings.HasPrefix(token, "mc_") {
    hash := authapp.HashToken(token)
    key, err := apiKeyRepo.FindByHash(ctx, hash)
    if err != nil || key == nil || !key.IsValid() {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // 异步防抖更新 last_used_at
    bizutils.GoWithCtx(ctx, func(ctx context.Context) {
        // 检查距上次更新是否超过 1 分钟
        if key.LastUsedAt == nil || time.Since(*key.LastUsedAt) > time.Minute {
            _ = apiKeyRepo.UpdateLastUsed(ctx, key.ID)
        }
    })
    ctx = ctxutils.SetUserID(r.Context(), key.UserID)
    next.ServeHTTP(w, r.WithContext(ctx))
    return
}
```

- [ ] **A6-8：集成测试：验证 API Key 完整流程**

```bash
# 启动服务
just run force=true

# 1. 登录获取 access token（用 /api/auth/login）
# 2. 用 access token 调用 GraphQL createApiKey mutation
# 3. 用返回的 api key 直接调用受保护端点
curl http://localhost:8080/graphql/org/test-org/ \
  -H "Authorization: Bearer mc_<api_key>" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ apiKeys { id name keyPrefix } }"}'
```

预期：返回该用户的 API Key 列表。

- [ ] **A6-9：commit**

```bash
cd modelcraft-backend
git add api/graph/org/schema/ \
        internal/app/auth/api_key_service.go \
        internal/interfaces/graphql/org/ \
        internal/middleware/chi_jwt_auth.go
git commit -m "feat(apikey): add API Key CRUD GraphQL + middleware verification"
```

**✅ Phase A6 验收：**
```bash
just test-unit-pkg ./internal/app/auth/...   # 全部 PASS
just lint                                     # 无报错
just build                                    # 编译通过
# GraphQL createApiKey → revokeApiKey → updateApiKey 均返回预期结果
# mc_ 前缀 key 可通过中间件验证
# 无效/过期/吊销 key 返回 401
```

---

### Phase A7：数据清理 Goroutine

**涉及文件：**
- 新建：`modelcraft-backend/internal/app/auth/cleanup_service.go`
- 修改：`modelcraft-backend/main.go` 或启动入口（注册定时任务）

- [ ] **A7-1：新建 cleanup_service.go**

```go
// modelcraft-backend/internal/app/auth/cleanup_service.go
package auth

import (
    "context"
    "time"
    domainauth "modelcraft/internal/domain/auth"
    "modelcraft/pkg/bizutils"
    "modelcraft/pkg/logfacade"
)

// CleanupService 负责定期清理过期的 token 和 key 记录
type CleanupService struct {
    refreshTokenRepo domainauth.RefreshTokenRepository
    apiKeyRepo       domainauth.APIKeyRepository
    logger           logfacade.Logger
}

// Start 启动后台清理 goroutine，每 24 小时执行一次
func (s *CleanupService) Start(ctx context.Context) {
    bizutils.GoWithCtx(ctx, func(ctx context.Context) {
        ticker := time.NewTicker(24 * time.Hour)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                s.runCleanup(ctx)
            case <-ctx.Done():
                return
            }
        }
    })
}

func (s *CleanupService) runCleanup(ctx context.Context) {
    if err := s.refreshTokenRepo.DeleteExpired(ctx); err != nil {
        s.logger.Errorf("cleanup refresh tokens failed: %v", err)
    }
    if err := s.apiKeyRepo.DeleteRevoked(ctx); err != nil {
        s.logger.Errorf("cleanup api keys failed: %v", err)
    }
}
```

- [ ] **A7-2：在应用启动时注册 CleanupService**

在 `main.go` 或依赖注入入口中调用 `cleanupService.Start(ctx)`。

- [ ] **A7-3：commit**

```bash
cd modelcraft-backend
git add internal/app/auth/cleanup_service.go main.go
git commit -m "feat(auth): add daily cleanup goroutine for expired tokens and revoked keys"
```

**✅ Phase A7 验收（Go Backend 全部完成）：**
```bash
just test-unit          # 全部 PASS
just lint               # 无报错
just build              # 编译通过
just run force=true     # 服务正常启动，无启动报错
```

---

## Chunk 4: Next.js BFF

### Phase B1：Go 内部 Auth Client

**涉及文件：**
- 新建：`modelcraft-front/src/bff/auth/go-auth-client.ts`（封装对 Go Backend 内部端点的调用）

- [ ] **B1-1：新建 go-auth-client.ts**

```typescript
// modelcraft-front/src/bff/auth/go-auth-client.ts

const GO_BACKEND_INTERNAL_URL = process.env.GO_BACKEND_INTERNAL_URL ?? 'http://localhost:8080'

export interface LoginResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

export interface RefreshResult {
  userId: string
  refreshToken: string
  expiresAt: string
}

// callGoLogin 调用 Go Backend /api/auth/login
// 传入从 Casdoor token 提取的用户信息
export async function callGoLogin(params: {
  externalId: string
  email: string
  name: string
}): Promise<LoginResult> {
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params),
  })
  if (!res.ok) throw new Error(`Go login failed: ${res.status}`)
  return res.json()
}

// callGoRefresh 调用 Go Backend /api/auth/refresh，执行 token 轮换
export async function callGoRefresh(refreshToken: string): Promise<RefreshResult> {
  const res = await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  })
  if (res.status === 401) throw new TokenReuseError('Refresh token revoked or reused')
  if (!res.ok) throw new Error(`Go refresh failed: ${res.status}`)
  return res.json()
}

// callGoLogout 调用 Go Backend /api/auth/logout，吊销指定 token
export async function callGoLogout(refreshToken: string): Promise<void> {
  await fetch(`${GO_BACKEND_INTERNAL_URL}/api/auth/logout`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  })
  // 即使失败也继续（Cookie 会被清除）
}

export class TokenReuseError extends Error {}
```

- [ ] **B1-2：新增环境变量**

在 `modelcraft-front/.env.dev.example` 中新增：
```
GO_BACKEND_INTERNAL_URL=http://localhost:8080
```

- [ ] **B1-3：commit**

```bash
cd modelcraft-front
git add src/bff/auth/go-auth-client.ts .env.dev.example
git commit -m "feat(bff): add Go Backend internal auth client"
```

**✅ Phase B1 验收：**
```bash
cd modelcraft-front
npx tsc --noEmit   # TypeScript 类型检查通过
```

---

### Phase B2：BFF Auth 端点

**涉及文件：**
- 新建：`modelcraft-front/src/app/api/bff/auth/token/route.ts`
- 新建：`modelcraft-front/src/app/api/bff/auth/refresh/route.ts`
- 新建：`modelcraft-front/src/app/api/bff/auth/logout/route.ts`
- 新建：`modelcraft-front/src/bff/auth/jwt-utils.ts`（BFF 签发 Access Token）
- 新建：`modelcraft-front/src/bff/auth/cookie-utils.ts`（httpOnly Cookie 管理）

- [ ] **B2-1：新建 jwt-utils.ts（BFF 签发 Access Token）**

```typescript
// modelcraft-front/src/bff/auth/jwt-utils.ts
import { SignJWT, jwtVerify } from 'jose'

const JWT_SECRET = new TextEncoder().encode(
  process.env.JWT_SECRET ?? (() => { throw new Error('JWT_SECRET is required') })()
)
const ISSUER = 'modelcraft'
const EXPIRY = '1h'

export async function signAccessToken(userId: string): Promise<string> {
  return new SignJWT({ user_id: userId })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuer(ISSUER)
    .setIssuedAt()
    .setExpirationTime(EXPIRY)
    .sign(JWT_SECRET)
}

export async function verifyAccessToken(token: string): Promise<{ userId: string }> {
  const { payload } = await jwtVerify(token, JWT_SECRET, { issuer: ISSUER })
  return { userId: payload['user_id'] as string }
}
```

- [ ] **B2-2：新建 cookie-utils.ts（httpOnly Cookie 管理）**

```typescript
// modelcraft-front/src/bff/auth/cookie-utils.ts
import { cookies } from 'next/headers'

const COOKIE_NAME = 'refresh_token'
const COOKIE_MAX_AGE = 7 * 24 * 60 * 60  // 7 天（秒）

export function setRefreshTokenCookie(token: string): void {
  cookies().set(COOKIE_NAME, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    maxAge: COOKIE_MAX_AGE,
    path: '/api/bff/auth',
  })
}

export function getRefreshTokenFromCookie(): string | undefined {
  return cookies().get(COOKIE_NAME)?.value
}

export function clearRefreshTokenCookie(): void {
  cookies().delete(COOKIE_NAME)
}
```

- [ ] **B2-3：新建 /bff/auth/token/route.ts（登录端点）**

```typescript
// modelcraft-front/src/app/api/bff/auth/token/route.ts
import { NextRequest, NextResponse } from 'next/server'
import { getCasdoorToken, extractUserInfoFromToken } from '@/bff/auth/casdoor'
import { callGoLogin } from '@/bff/auth/go-auth-client'
import { signAccessToken } from '@/bff/auth/jwt-utils'
import { setRefreshTokenCookie } from '@/bff/auth/cookie-utils'

export async function POST(req: NextRequest) {
  const { code, redirectUri } = await req.json()

  // 1. 用 code 向 Casdoor 换取 token
  const casdoorToken = await getCasdoorToken(code, redirectUri)
  const { externalId, email, name } = extractUserInfoFromToken(casdoorToken)

  // 2. 调用 Go Backend login（user sync + 生成 Refresh Token）
  const { userId, refreshToken } = await callGoLogin({ externalId, email, name })

  // 3. BFF 签发 Access Token
  const accessToken = await signAccessToken(userId)

  // 4. 设置 httpOnly Cookie
  setRefreshTokenCookie(refreshToken)

  return NextResponse.json({ accessToken, expiresIn: 3600 })
}
```

- [ ] **B2-4：新建 /bff/auth/refresh/route.ts（刷新端点）**

```typescript
// modelcraft-front/src/app/api/bff/auth/refresh/route.ts
import { NextRequest, NextResponse } from 'next/server'
import { callGoRefresh, TokenReuseError } from '@/bff/auth/go-auth-client'
import { signAccessToken } from '@/bff/auth/jwt-utils'
import { getRefreshTokenFromCookie, setRefreshTokenCookie, clearRefreshTokenCookie } from '@/bff/auth/cookie-utils'

export async function POST(req: NextRequest) {
  const refreshToken = getRefreshTokenFromCookie()
  if (!refreshToken) {
    return NextResponse.json({ error: 'No refresh token' }, { status: 401 })
  }

  try {
    const { userId, refreshToken: newRefreshToken } = await callGoRefresh(refreshToken)

    // 更新 httpOnly Cookie（新 Refresh Token）
    setRefreshTokenCookie(newRefreshToken)

    // 签发新 Access Token
    const accessToken = await signAccessToken(userId)
    return NextResponse.json({ accessToken, expiresIn: 3600 })
  } catch (err) {
    // Token 盗用或已过期：清除 Cookie，让前端跳转登录页
    clearRefreshTokenCookie()
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }
}
```

- [ ] **B2-5：新建 /bff/auth/logout/route.ts（登出端点）**

```typescript
// modelcraft-front/src/app/api/bff/auth/logout/route.ts
import { NextResponse } from 'next/server'
import { callGoLogout } from '@/bff/auth/go-auth-client'
import { getRefreshTokenFromCookie, clearRefreshTokenCookie } from '@/bff/auth/cookie-utils'

export async function POST() {
  const refreshToken = getRefreshTokenFromCookie()

  // 无论 Go 端是否成功，都清除 Cookie
  if (refreshToken) {
    await callGoLogout(refreshToken)
  }
  clearRefreshTokenCookie()

  return new NextResponse(null, { status: 204 })
}
```

- [ ] **B2-6：TypeScript 类型检查**

```bash
cd modelcraft-front
npx tsc --noEmit
```

预期：无类型错误。

- [ ] **B2-7：commit**

```bash
cd modelcraft-front
git add src/app/api/bff/auth/ src/bff/auth/jwt-utils.ts src/bff/auth/cookie-utils.ts
git commit -m "feat(bff): add /bff/auth/token, /refresh, /logout endpoints with httpOnly Cookie"
```

**✅ Phase B2 验收：**

启动 Go Backend + Next.js BFF，手动测试完整登录流程：

```bash
# 在 modelcraft-front 目录
npm run dev

# 1. 访问 Casdoor 登录页，完成登录获取 code
# 2. 调用 BFF token 端点
curl -X POST http://localhost:3000/api/bff/auth/token \
  -H "Content-Type: application/json" \
  -d '{"code":"xxx","redirectUri":"http://localhost:3000/auth/callback"}'
# 预期：返回 { accessToken } + Set-Cookie: refresh_token=...; HttpOnly

# 3. 调用 refresh 端点（自动携带 Cookie）
curl -X POST http://localhost:3000/api/bff/auth/refresh \
  --cookie "refresh_token=<上一步的 token>"
# 预期：返回新 accessToken，Cookie 更新

# 4. 调用 logout 端点
curl -X POST http://localhost:3000/api/bff/auth/logout \
  --cookie "refresh_token=<token>"
# 预期：204，Cookie 清除

# 5. 用已 logout 的 token 再次 refresh
curl -X POST http://localhost:3000/api/bff/auth/refresh \
  --cookie "refresh_token=<已吊销的 token>"
# 预期：401
```

---

### Phase B3：前端 Token 存储迁移

**涉及文件：**
- 修改：`modelcraft-front/src/bff/auth/casdoor.ts`（移除 localStorage token 存储，改为调用 BFF 端点）
- 新建：`modelcraft-front/src/web/stores/auth-store.ts`（内存 Access Token 存储）
- 修改：`modelcraft-front/src/app/auth/callback/page.tsx`（适配新流程）

- [ ] **B3-1：新建 auth-store.ts（内存 Access Token 存储）**

```typescript
// modelcraft-front/src/web/stores/auth-store.ts
import { create } from 'zustand'

interface AuthState {
  accessToken: string | null
  expiresAt: number | null   // Unix timestamp（毫秒）
  setAccessToken: (token: string, expiresIn: number) => void
  clearAccessToken: () => void
  isTokenExpired: () => boolean
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  expiresAt: null,

  setAccessToken: (token: string, expiresIn: number) => {
    set({
      accessToken: token,
      expiresAt: Date.now() + expiresIn * 1000,
    })
  },

  clearAccessToken: () => set({ accessToken: null, expiresAt: null }),

  isTokenExpired: () => {
    const { expiresAt } = get()
    if (!expiresAt) return true
    // 提前 5 分钟视为过期，触发刷新
    return Date.now() > expiresAt - 5 * 60 * 1000
  },
}))
```

- [ ] **B3-2：修改 auth callback 页面，适配新 BFF 流程**

`src/app/auth/callback/page.tsx` 中：
- 将原来直接调用后端 `/api/auth/token` 改为调用 BFF `/api/bff/auth/token`
- 将返回的 `accessToken` 存入 `useAuthStore`（内存），不再写 localStorage
- 移除所有 `localStorage.setItem('auth_token', ...)` 和 `localStorage.setItem('auth_refresh_token', ...)`

- [ ] **B3-3：修改 casdoor.ts 的 refresh 逻辑**

将原来的 `refreshAccessToken()` 函数：
- 请求地址改为 `/api/bff/auth/refresh`（不带 body，Cookie 自动携带）
- 成功后调用 `useAuthStore.setAccessToken()`
- 移除 localStorage 读写

- [ ] **B3-4：全局搜索确认无残留 localStorage token 存储**

```bash
cd modelcraft-front
grep -r "auth_token\|auth_refresh_token" src/ --include="*.ts" --include="*.tsx"
```

预期：只应在旧注释或迁移说明中出现，不应有 `setItem` 调用。

- [ ] **B3-5：TypeScript 类型检查 + lint**

```bash
cd modelcraft-front
npx tsc --noEmit
npm run lint
```

预期：无错误。

- [ ] **B3-6：commit**

```bash
cd modelcraft-front
git add src/web/stores/auth-store.ts \
        src/app/auth/callback/page.tsx \
        src/bff/auth/casdoor.ts
git commit -m "feat(bff): migrate token storage - localStorage to memory + httpOnly Cookie"
```

**✅ Phase B3 验收：**
```bash
npm run lint        # 无报错
npx tsc --noEmit    # 无类型错误
# 打开浏览器 DevTools：
# - Application > localStorage 中不再有 auth_token / auth_refresh_token
# - Application > Cookies 中有 refresh_token（标记 HttpOnly）
# - 刷新页面后仍保持登录（BFF 自动续期）
```

---

### Phase B4：Auth Provider 适配与端到端验证

**涉及文件：**
- 修改：`modelcraft-front/src/web/components/auth-provider.tsx`（适配新 token 存储方式）
- 修改：`modelcraft-front/src/bff/apollo/clients.ts`（从 store 获取 Access Token）

- [ ] **B4-1：修改 auth-provider.tsx**

- 将 token 存在性检查从 `localStorage.getItem('auth_token')` 改为 `useAuthStore.getState().accessToken`
- 定时刷新逻辑改为：检查 `isTokenExpired()` → 调用 `/api/bff/auth/refresh` → 更新 store
- 登出逻辑改为：调用 `/api/bff/auth/logout` → `clearAccessToken()`

- [ ] **B4-2：修改 apollo clients.ts，从 store 获取 token**

将 Apollo auth link 中的 token 来源：
```typescript
// 之前
const token = localStorage.getItem('auth_token')

// 之后
const token = useAuthStore.getState().accessToken
```

- [ ] **B4-3：端到端验证完整登录流程**

```
1. 清空所有 Cookie 和 localStorage
2. 访问 http://localhost:3000/（受保护路由）
3. 自动跳转登录页
4. 点击登录 → 跳转 Casdoor → 完成登录 → 跳转回 callback
5. 验证：localStorage 中无 auth_token
6. 验证：Cookie 中有 refresh_token（HttpOnly）
7. 进入业务页面，正常使用 GraphQL 查询
8. 等待 1h（或手动清空内存 token） → 验证自动续期
9. 点击登出 → 验证 Cookie 清除 → 验证跳转登录页
```

- [ ] **B4-4：最终 lint + 类型检查**

```bash
cd modelcraft-front
npm run lint
npx tsc --noEmit
```

预期：全部通过。

- [ ] **B4-5：commit**

```bash
cd modelcraft-front
git add src/web/components/auth-provider.tsx \
        src/bff/apollo/clients.ts
git commit -m "feat(bff): adapt auth provider and apollo client to memory token store"
```

**✅ Phase B4（BFF 全部完成）最终验收：**

| 验收项 | 验证方式 | 预期结果 |
|--------|---------|---------|
| 登录后 localStorage 无 token | DevTools > Application > Storage | 无 auth_token / auth_refresh_token |
| Refresh Token 以 httpOnly Cookie 存储 | DevTools > Application > Cookies | refresh_token 标记 HttpOnly |
| Access Token 刷新透明 | 手动等待或清空内存 | 页面不跳转，自动续期 |
| 登出清除 Cookie | DevTools > Cookies | refresh_token 消失 |
| 已吊销 token 无法刷新 | 用旧 Cookie 手动 refresh | 401，跳转登录 |
| CLI API Key 正常工作 | `curl -H "Authorization: Bearer mc_xxx"` | 返回正常业务数据 |
| GraphQL API Key 管理正常 | 在 UI 创建/吊销/编辑 key | 正确响应 |

---

## 根项目收尾

- [ ] **在根项目提交子模块引用更新**

```bash
cd /path/to/root
git add modelcraft-backend modelcraft-front
git commit -m "chore: update submodules - token redesign (stateful refresh token + API Key + BFF httpOnly Cookie)"
```
