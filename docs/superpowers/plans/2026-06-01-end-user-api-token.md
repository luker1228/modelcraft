# End-User API Token (PAT) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 End-User 实现 Personal Access Token (PAT) 功能，包含后端 DB/领域/应用/中间件/GraphQL 层和前端 Token 管理页面。

**Architecture:** PAT 在后端 middleware 层识别 `mc_pat_` 前缀并验证，不改动 APISIX。Token hash 存入系统 DB，明文仅创建时返回一次。GraphQL API 挂在已有 Org-scoped endpoint，前端新增侧边栏导航项和 Token 管理页面。

**Tech Stack:** Go (chi, sqlc, gqlgen), MySQL, Next.js, Apollo Client, Tailwind CSS, shadcn/ui

---

## 文件结构总览

**后端新建：**
- `db/schema/mysql/17_end_user_api_tokens.sql` — DB migration
- `internal/domain/enduser/api_token.go` — 领域实体 + Repository 接口
- `internal/infrastructure/repository/sql_api_token_repository.go` — Repository 实现（直接 SQL，不用 sqlc）
- `internal/app/enduser/api_token_service.go` — 应用服务（创建/列出/撤销）
- `internal/middleware/chi_pat_auth.go` — PAT 验证 middleware
- `api/graph/org/schema/end_user_api_token.graphql` — GraphQL schema
- `internal/interfaces/graphql/org/end_user_api_token.resolvers.go` — Resolver 实现

**后端修改：**
- `internal/domain/enduser/end_user_repository.go` — 新增 APITokenRepository 接口
- `internal/interfaces/graphql/org/resolver.go` — 新增 APITokenService 字段
- `internal/interfaces/http/routes.go` — resolver 注入 APITokenService
- `internal/interfaces/http/chi_setup.go` — 挂载 PAT middleware

**前端新建：**
- `src/app/end-user/[orgName]/dashboard/tokens/page.tsx` — Token 管理页面
- `src/app/end-user/[orgName]/dashboard/tokens/_components/TokenTable.tsx` — Token 列表表格
- `src/app/end-user/[orgName]/dashboard/tokens/_components/CreateTokenDialog.tsx` — 新建 Token 弹窗
- `src/app/end-user/[orgName]/dashboard/tokens/_components/TokenRevealDialog.tsx` — 一次性展示弹窗

**前端修改：**
- `src/web/components/features/layout/EndUserAppLayout.tsx` — 新增 `'tokens'` activePage + 侧边栏导航项
- `src/api-client/end-user/graphql-docs.ts` — 新增 3 个 GraphQL 文档

---

## Task 1: DB Migration

**Files:**
- Create: `modelcraft-backend/db/schema/mysql/17_end_user_api_tokens.sql`

- [ ] **Step 1: 创建 migration 文件**

```sql
-- =============================================================================
-- End-User API Tokens (PAT)
-- 说明：
-- - 存储 EndUser 创建的 Personal Access Token（明文不落库，存 SHA-256 hash）
-- - token_hash UNIQUE 索引用于 O(1) 验证
-- - 软删除：deleted_at + delete_token 联合避让，支持撤销后同名重建
-- =============================================================================

CREATE TABLE IF NOT EXISTS `end_user_api_tokens` (
  `id`            VARCHAR(36)   NOT NULL COMMENT '唯一标识符 (UUID v7)',
  `org_name`      VARCHAR(255)  NOT NULL COMMENT '所属组织',
  `end_user_id`   VARCHAR(36)   NOT NULL COMMENT '创建者 EndUser ID',
  `name`          VARCHAR(255)  NOT NULL COMMENT '用户自定义名称',
  `token_hash`    VARCHAR(64)   NOT NULL COMMENT 'SHA-256(plaintext) hex，用于验证',
  `expires_at`    DATETIME      NULL     COMMENT 'NULL 表示永不过期',
  `last_used_at`  DATETIME      NULL     COMMENT '最近使用时间，异步更新',
  `created_at`    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted_at`    DATETIME      NULL     COMMENT '软删除时间，NULL 表示活跃',
  `delete_token`  VARCHAR(36)   NOT NULL DEFAULT '' COMMENT '软删除唯一标记，活跃时为空字符串',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_token_hash` (`token_hash`),
  UNIQUE KEY `uq_user_token_name` (`org_name`, `end_user_id`, `name`, `delete_token`),
  KEY `idx_user_tokens` (`org_name`, `end_user_id`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
  COMMENT='EndUser Personal Access Token 注册表（PAT）';
```

- [ ] **Step 2: 应用 migration**

```bash
cd modelcraft-backend
just db-migrate
```

Expected: migration 成功，`end_user_api_tokens` 表存在。

- [ ] **Step 3: 验证表结构**

```bash
just db-shell
# 在 MySQL shell 中：
DESCRIBE end_user_api_tokens;
```

Expected: 10 列，含 `token_hash` UNIQUE KEY。

- [ ] **Step 4: Commit**

```bash
git add db/schema/mysql/17_end_user_api_tokens.sql
git commit -m "feat(db): add end_user_api_tokens table for PAT support"
```

---

## Task 2: 领域层 — APIToken 实体与 Repository 接口

**Files:**
- Create: `modelcraft-backend/internal/domain/enduser/api_token.go`
- Modify: `modelcraft-backend/internal/domain/enduser/end_user_repository.go`

- [ ] **Step 1: 创建领域实体文件**

```go
// internal/domain/enduser/api_token.go
package enduser

import "time"

// APIToken 代表一个 EndUser 创建的 Personal Access Token。
// 明文 token 仅在创建时存在于内存，不落库，库中只存 SHA-256 hash。
type APIToken struct {
	ID          string
	OrgName     string
	EndUserID   string
	Name        string
	TokenHash   string     // SHA-256(plaintext) hex
	ExpiresAt   *time.Time // nil = 永不过期
	LastUsedAt  *time.Time // nil = 从未使用
	CreatedAt   time.Time
	DeletedAt   *time.Time
	DeleteToken string
}

// IsValid 检查 token 是否有效（未被撤销 & 未过期）。
func (t *APIToken) IsValid() bool {
	if t.DeletedAt != nil {
		return false
	}
	if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
		return false
	}
	return true
}
```

- [ ] **Step 2: 在 end_user_repository.go 末尾新增 APITokenRepository 接口**

在文件 `internal/domain/enduser/end_user_repository.go` 末尾追加（保留所有现有内容）：

```go
// APITokenRepository 定义 EndUser PAT 的持久化操作。
// 所有写操作都在系统 DB（非 tenant DB）上执行。
type APITokenRepository interface {
	// Save 插入新 token 记录（id 已由调用方生成）。
	Save(ctx context.Context, token *APIToken) error
	// FindByHash 通过 SHA-256 hash 查找活跃 token（未软删除）。
	// 未找到时返回 (nil, nil)。
	FindByHash(ctx context.Context, hash string) (*APIToken, error)
	// ListByUser 返回指定用户的全部活跃 token 列表（按 created_at DESC）。
	ListByUser(ctx context.Context, orgName, endUserID string) ([]*APIToken, error)
	// SoftDelete 软删除指定 token（设置 deleted_at + delete_token）。
	// 若 token 不属于该用户则返回 error。
	SoftDelete(ctx context.Context, id, orgName, endUserID string) error
	// UpdateLastUsed 异步更新 last_used_at，验证成功后调用。
	UpdateLastUsed(ctx context.Context, id string, at time.Time) error
}
```

- [ ] **Step 3: 编译验证**

```bash
cd modelcraft-backend
go build ./internal/domain/enduser/...
```

Expected: 编译成功，无错误。

- [ ] **Step 4: Commit**

```bash
git add internal/domain/enduser/api_token.go internal/domain/enduser/end_user_repository.go
git commit -m "feat(domain): add APIToken entity and APITokenRepository interface"
```

---

## Task 3: Infrastructure 层 — Repository 实现

**Files:**
- Create: `modelcraft-backend/internal/infrastructure/repository/sql_api_token_repository.go`

- [ ] **Step 1: 创建 Repository 实现**

```go
// internal/infrastructure/repository/sql_api_token_repository.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"modelcraft/internal/domain/enduser"
	"time"
)

// SqlAPITokenRepository 实现 enduser.APITokenRepository，操作系统 DB。
type SqlAPITokenRepository struct {
	db *sql.DB
}

// NewSqlAPITokenRepository 创建 APIToken Repository。
func NewSqlAPITokenRepository(db *sql.DB) *SqlAPITokenRepository {
	return &SqlAPITokenRepository{db: db}
}

func (r *SqlAPITokenRepository) Save(ctx context.Context, token *enduser.APIToken) error {
	query := `
		INSERT INTO end_user_api_tokens
		  (id, org_name, end_user_id, name, token_hash, expires_at, created_at, delete_token)
		VALUES (?, ?, ?, ?, ?, ?, ?, '')`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.OrgName,
		token.EndUserID,
		token.Name,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save api token: %w", err)
	}
	return nil
}

func (r *SqlAPITokenRepository) FindByHash(ctx context.Context, hash string) (*enduser.APIToken, error) {
	query := `
		SELECT id, org_name, end_user_id, name, token_hash,
		       expires_at, last_used_at, created_at, deleted_at, delete_token
		FROM end_user_api_tokens
		WHERE token_hash = ? AND deleted_at IS NULL
		LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, hash)
	return scanAPIToken(row)
}

func (r *SqlAPITokenRepository) ListByUser(ctx context.Context, orgName, endUserID string) ([]*enduser.APIToken, error) {
	query := `
		SELECT id, org_name, end_user_id, name, token_hash,
		       expires_at, last_used_at, created_at, deleted_at, delete_token
		FROM end_user_api_tokens
		WHERE org_name = ? AND end_user_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orgName, endUserID)
	if err != nil {
		return nil, fmt.Errorf("list api tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*enduser.APIToken
	for rows.Next() {
		token, err := scanAPIToken(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *SqlAPITokenRepository) SoftDelete(ctx context.Context, id, orgName, endUserID string) error {
	now := time.Now()
	query := `
		UPDATE end_user_api_tokens
		SET deleted_at = ?, delete_token = ?
		WHERE id = ? AND org_name = ? AND end_user_id = ? AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, now, id, id, orgName, endUserID)
	if err != nil {
		return fmt.Errorf("soft delete api token: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("api token not found or already deleted")
	}
	return nil
}

func (r *SqlAPITokenRepository) UpdateLastUsed(ctx context.Context, id string, at time.Time) error {
	query := `UPDATE end_user_api_tokens SET last_used_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, at, id)
	return err
}

// scanAPIToken 从 sql.Row 或 sql.Rows 扫描一行数据。
type scanner interface {
	Scan(dest ...any) error
}

func scanAPIToken(s scanner) (*enduser.APIToken, error) {
	var t enduser.APIToken
	var expiresAt, lastUsedAt, deletedAt sql.NullTime
	err := s.Scan(
		&t.ID, &t.OrgName, &t.EndUserID, &t.Name, &t.TokenHash,
		&expiresAt, &lastUsedAt, &t.CreatedAt, &deletedAt, &t.DeleteToken,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan api token: %w", err)
	}
	if expiresAt.Valid {
		t.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		t.LastUsedAt = &lastUsedAt.Time
	}
	if deletedAt.Valid {
		t.DeletedAt = &deletedAt.Time
	}
	return &t, nil
}
```

- [ ] **Step 2: 验证接口合规性（编译检查）**

在文件末尾 `scanAPIToken` 函数前添加接口断言：

```go
// 编译时验证接口合规性
var _ enduser.APITokenRepository = (*SqlAPITokenRepository)(nil)
```

- [ ] **Step 3: 编译验证**

```bash
cd modelcraft-backend
go build ./internal/infrastructure/repository/...
```

Expected: 编译成功。

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/repository/sql_api_token_repository.go
git commit -m "feat(infra): implement SqlAPITokenRepository for PAT"
```

---

## Task 4: 应用层 — APITokenService

**Files:**
- Create: `modelcraft-backend/internal/app/enduser/api_token_service.go`

- [ ] **Step 1: 创建应用服务文件**

```go
// internal/app/enduser/api_token_service.go
package enduser

import (
	"context"
	"fmt"
	"modelcraft/internal/app/auth"
	"modelcraft/internal/domain/enduser"
	"modelcraft/pkg/bizutils"
	"time"
)

const maxTokensPerUser = 20

// APITokenService 处理 EndUser PAT 的创建、列出和撤销。
type APITokenService struct {
	repo enduser.APITokenRepository
}

// NewAPITokenService 创建 APITokenService。
func NewAPITokenService(repo enduser.APITokenRepository) *APITokenService {
	return &APITokenService{repo: repo}
}

// CreateAPITokenCommand 创建 PAT 的命令参数。
type CreateAPITokenCommand struct {
	OrgName     string
	EndUserID   string
	Name        string
	ExpiresAt   *time.Time // nil = 永不过期
}

// CreateAPITokenResult 创建成功后的返回值。
type CreateAPITokenResult struct {
	Token     *enduser.APIToken
	Plaintext string // 仅此一次返回，不落库
}

// CreateAPIToken 生成新的 PAT，返回明文 token（仅此一次）。
func (s *APITokenService) CreateAPIToken(ctx context.Context, cmd CreateAPITokenCommand) (*CreateAPITokenResult, error) {
	// 检查数量上限
	existing, err := s.repo.ListByUser(ctx, cmd.OrgName, cmd.EndUserID)
	if err != nil {
		return nil, fmt.Errorf("check token count: %w", err)
	}
	if len(existing) >= maxTokensPerUser {
		return nil, fmt.Errorf("token limit reached: max %d tokens per user", maxTokensPerUser)
	}

	// 生成随机 token（复用 auth 包的生成器）
	// GenerateRefreshToken 生成 32 字节 CSPRNG → 64 位 hex，同时返回 SHA-256 hash
	plaintext, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	// 添加 mc_pat_ 前缀（middleware 用于识别）
	fullPlaintext := "mc_pat_" + plaintext

	id, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("generate token id: %w", err)
	}

	token := &enduser.APIToken{
		ID:          id,
		OrgName:     cmd.OrgName,
		EndUserID:   cmd.EndUserID,
		Name:        cmd.Name,
		TokenHash:   hash, // SHA-256(plaintext without prefix)
		ExpiresAt:   cmd.ExpiresAt,
		CreatedAt:   time.Now(),
		DeleteToken: "",
	}
	if err := s.repo.Save(ctx, token); err != nil {
		return nil, fmt.Errorf("save token: %w", err)
	}

	return &CreateAPITokenResult{
		Token:     token,
		Plaintext: fullPlaintext,
	}, nil
}

// ListAPITokens 返回用户的全部活跃 token 列表。
func (s *APITokenService) ListAPITokens(ctx context.Context, orgName, endUserID string) ([]*enduser.APIToken, error) {
	return s.repo.ListByUser(ctx, orgName, endUserID)
}

// RevokeAPIToken 撤销（软删除）指定 token。
func (s *APITokenService) RevokeAPIToken(ctx context.Context, id, orgName, endUserID string) error {
	return s.repo.SoftDelete(ctx, id, orgName, endUserID)
}

// ValidateToken 验证 PAT 明文，成功返回 token 实体（含 OrgName/EndUserID）。
// 调用方应异步更新 last_used_at。
func (s *APITokenService) ValidateToken(ctx context.Context, plaintext string) (*enduser.APIToken, error) {
	// 去掉 mc_pat_ 前缀后计算 hash（与存储时一致）
	raw := plaintext[len("mc_pat_"):]
	hash := auth.HashToken(raw)
	token, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("find token: %w", err)
	}
	if token == nil || !token.IsValid() {
		return nil, fmt.Errorf("invalid or expired token")
	}
	return token, nil
}
```

- [ ] **Step 2: 编译验证**

```bash
cd modelcraft-backend
go build ./internal/app/enduser/...
```

Expected: 编译成功。

- [ ] **Step 3: Commit**

```bash
git add internal/app/enduser/api_token_service.go
git commit -m "feat(app): add APITokenService for PAT create/list/revoke/validate"
```

---

## Task 5: PAT 验证 Middleware

**Files:**
- Create: `modelcraft-backend/internal/middleware/chi_pat_auth.go`
- Modify: `modelcraft-backend/internal/interfaces/http/chi_setup.go`

- [ ] **Step 1: 创建 PAT middleware**

```go
// internal/middleware/chi_pat_auth.go
package middleware

import (
	"modelcraft/internal/app/enduser"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"
	"strings"
	"time"
)

const patPrefix = "mc_pat_"

// ChiPATAuthMiddleware 识别 Authorization: Bearer mc_pat_xxx 请求，
// 验证 token 有效性后将 EndUser 身份注入 context，与 JWT 路径等价。
// 若 Authorization 头不以 mc_pat_ 开头，直接调用 next 不拦截。
func ChiPATAuthMiddleware(svc *enduser.APITokenService, logger logfacade.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			if !strings.HasPrefix(bearer, "Bearer "+patPrefix) {
				// 不是 PAT，交给下一个处理器（JWT 路径）
				next.ServeHTTP(w, r)
				return
			}

			plaintext := strings.TrimPrefix(bearer, "Bearer ")
			token, err := svc.ValidateToken(r.Context(), plaintext)
			if err != nil {
				logger.Warnf(r.Context(), "PAT validation failed: %v", err)
				writeJSONError(w, http.StatusUnauthorized, "invalid or expired token", "UNAUTHENTICATED")
				return
			}

			// 异步更新 last_used_at，不阻塞请求
			go func() {
				if updateErr := svc.UpdateLastUsedAt(r.Context(), token.ID, time.Now()); updateErr != nil {
					logger.Warnf(r.Context(), "update last_used_at failed: %v", updateErr)
				}
			}()

			// 注入 EndUser 身份到 context（与 JWT 路径相同的 key）
			ctx := ctxutils.SetUserID(r.Context(), token.EndUserID)
			ctx = ctxutils.SetOrgName(ctx, token.OrgName)
			ctx = ctxutils.SetUserType(ctx, "end_user")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
```

- [ ] **Step 2: 在 APITokenService 新增 UpdateLastUsedAt 方法**

在 `internal/app/enduser/api_token_service.go` 末尾追加：

```go
// UpdateLastUsedAt 更新 token 最后使用时间（通常异步调用）。
func (s *APITokenService) UpdateLastUsedAt(ctx context.Context, id string, at time.Time) error {
	return s.repo.UpdateLastUsed(ctx, id, at)
}
```

- [ ] **Step 3: 在 chi_setup.go 中挂载 PAT middleware**

在 `internal/interfaces/http/chi_setup.go` 的 `ChiRouterConfig` struct 中新增字段：

```go
// Modify: internal/interfaces/http/chi_setup.go

// ChiRouterConfig holds all dependencies needed to set up the Chi router.
type ChiRouterConfig struct {
	// ... 现有字段保持不变 ...

	// PAT authentication service (nil = PAT auth disabled)
	APITokenService *appEnduser.APITokenService  // 新增
}
```

在 `SetupChiRouter` 函数内，在 CLI Auth Routes 注册之后、OpenAPI Routes 之前，插入：

```go
// ============================================================
// PAT Auth Middleware (applies before OpenAPI & GraphQL routes)
// ============================================================
if cfg.APITokenService != nil {
    r.Use(middleware.ChiPATAuthMiddleware(cfg.APITokenService, cfg.Logger))
}
```

- [ ] **Step 4: 新增 import**

在 `chi_setup.go` 顶部 import 块新增（如未存在）：

```go
appEnduser "modelcraft/internal/app/enduser"
```

- [ ] **Step 5: 编译验证**

```bash
cd modelcraft-backend
go build ./internal/middleware/... && go build ./internal/interfaces/http/...
```

Expected: 编译成功（routes.go 尚未注入，会有编译错误后在 Task 6 修复）。

- [ ] **Step 6: Commit**

```bash
git add internal/middleware/chi_pat_auth.go \
        internal/app/enduser/api_token_service.go \
        internal/interfaces/http/chi_setup.go
git commit -m "feat(middleware): add ChiPATAuthMiddleware for Bearer mc_pat_ tokens"
```

---

## Task 6: GraphQL Schema + Code Generation

**Files:**
- Create: `modelcraft-backend/api/graph/org/schema/end_user_api_token.graphql`
- Modify: `modelcraft-backend/internal/interfaces/graphql/org/resolver.go`
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`

- [ ] **Step 1: 创建 GraphQL schema 文件**

```graphql
# api/graph/org/schema/end_user_api_token.graphql
#
# EndUser Personal Access Token (PAT) 管理
# Endpoint: /end-user/graphql/org/{orgName}/
# 认证：EndUser Bearer JWT（PAT 本身不能管理 PAT）

# ============================================================
# Error Types
# ============================================================

type APITokenNameConflict implements Error {
  message: String!
}

type APITokenLimitReached implements Error {
  message: String!
  limit:   Int!
}

type APITokenNotFound implements Error {
  message: String!
}

union CreateAPITokenError = APITokenNameConflict | APITokenLimitReached | InvalidInput
union RevokeAPITokenError = APITokenNotFound | InvalidInput

# ============================================================
# Data Types
# ============================================================

type EndUserAPIToken {
  id:          ID!
  name:        String!
  createdAt:   Time!
  expiresAt:   Time        # null = 永不过期
  lastUsedAt:  Time        # null = 从未使用
}

type CreateAPITokenPayload {
  token:     EndUserAPIToken
  plaintext: String         # 仅创建时返回一次明文；失败时为 null
  error:     CreateAPITokenError
}

type RevokeAPITokenPayload {
  success: Boolean
  error:   RevokeAPITokenError
}

# ============================================================
# Queries & Mutations
# ============================================================

extend type Query {
  # 列出当前 EndUser 的全部活跃 PAT
  endUserAPITokens: [EndUserAPIToken!]!
}

extend type Mutation {
  # 创建新 PAT，返回明文（仅此一次）
  createEndUserAPIToken(
    name:      String!
    expiresAt: Time    # null = 永不过期
  ): CreateAPITokenPayload!

  # 撤销（软删除）指定 PAT
  revokeEndUserAPIToken(id: ID!): RevokeAPITokenPayload!
}
```

- [ ] **Step 2: 运行代码生成**

```bash
cd modelcraft-backend
just generate-gql
```

Expected: 生成 `internal/interfaces/graphql/org/end_user_api_token.resolvers.go`（含空 resolver stubs）及更新 `generated/generated.go`。

- [ ] **Step 3: 在 resolver.go 新增 APITokenService 字段**

```go
// internal/interfaces/graphql/org/resolver.go
// 在 Resolver struct 末尾新增：

// EndUser PAT management
APITokenService *appEnduser.APITokenService  // 新增
```

- [ ] **Step 4: 在 routes.go 注入 APITokenService 到 resolver 和 ChiRouterConfig**

在 `SetupEndUserOrgGraphQLRoutesOnChi` 函数中，`orgResolver` 初始化块新增字段：

```go
APITokenService: handlers.EndUserAPITokenService,  // 新增
```

在 `DesignHandlers` struct（`internal/interfaces/http/routes.go` 顶部）新增字段：

```go
EndUserAPITokenService *appEnduser.APITokenService  // 新增
```

在 `NewDesignHandlers` 函数中（约 367 行附近），在 `endUserAppService` 创建后新增：

```go
apiTokenRepo := repository.NewSqlAPITokenRepository(repoFactory.SqlDB)
apiTokenService := appEnduser.NewAPITokenService(apiTokenRepo)
```

在 return 的 `DesignHandlers{}` 中新增：

```go
EndUserAPITokenService: apiTokenService,
```

在 `SetupChiRouter` 调用处（`main.go` 或 `server.go`），为 `ChiRouterConfig` 注入：

```go
APITokenService: handlers.EndUserAPITokenService,
```

- [ ] **Step 5: 编译验证**

```bash
cd modelcraft-backend
go build ./...
```

Expected: 编译成功（resolver stubs 尚未实现，但 gqlgen 生成的 stub 是合法 Go 代码）。

- [ ] **Step 6: Commit**

```bash
git add api/graph/org/schema/end_user_api_token.graphql \
        internal/interfaces/graphql/org/generated/ \
        internal/interfaces/graphql/org/end_user_api_token.resolvers.go \
        internal/interfaces/graphql/org/resolver.go \
        internal/interfaces/http/routes.go
git commit -m "feat(graphql): add EndUser API Token schema and generate resolvers"
```

---

## Task 7: GraphQL Resolver 实现

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/org/end_user_api_token.resolvers.go`

- [ ] **Step 1: 实现 resolver 方法**

将生成的 stub 替换为以下完整实现：

```go
// internal/interfaces/graphql/org/end_user_api_token.resolvers.go
package orggraphql

import (
	"context"
	"fmt"
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/ctxutils"
	"time"
)

// EndUserAPITokens returns all active PATs for the current end-user.
func (r *queryResolver) EndUserAPITokens(ctx context.Context) ([]*generated.EndUserAPIToken, error) {
	svc := r.APITokenService
	if svc == nil {
		return nil, fmt.Errorf("api token service not initialized")
	}

	endUserID, err := ctxutils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user id: %w", err)
	}
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get org name: %w", err)
	}

	tokens, err := svc.ListAPITokens(ctx, orgName, endUserID)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}

	result := make([]*generated.EndUserAPIToken, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, toGQLAPIToken(t))
	}
	return result, nil
}

// CreateEndUserAPIToken creates a new PAT and returns the plaintext once.
func (r *mutationResolver) CreateEndUserAPIToken(ctx context.Context, name string, expiresAt *time.Time) (*generated.CreateAPITokenPayload, error) {
	svc := r.APITokenService
	if svc == nil {
		return &generated.CreateAPITokenPayload{
			Error: &generated.InvalidInput{Message: "api token service not initialized"},
		}, nil
	}

	if name == "" {
		return &generated.CreateAPITokenPayload{
			Error: &generated.InvalidInput{Message: "name is required"},
		}, nil
	}

	endUserID, err := ctxutils.GetUserIDFromContext(ctx)
	if err != nil {
		return &generated.CreateAPITokenPayload{
			Error: &generated.InvalidInput{Message: "unauthenticated"},
		}, nil
	}
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return &generated.CreateAPITokenPayload{
			Error: &generated.InvalidInput{Message: "org context missing"},
		}, nil
	}

	result, err := svc.CreateAPIToken(ctx, appEnduser.CreateAPITokenCommand{
		OrgName:   orgName,
		EndUserID: endUserID,
		Name:      name,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		// 简化错误映射（生产可细化）
		return &generated.CreateAPITokenPayload{
			Error: &generated.InvalidInput{Message: err.Error()},
		}, nil
	}

	gqlToken := toGQLAPIToken(result.Token)
	plaintext := result.Plaintext
	return &generated.CreateAPITokenPayload{
		Token:     gqlToken,
		Plaintext: &plaintext,
	}, nil
}

// RevokeEndUserAPIToken soft-deletes a PAT by ID.
func (r *mutationResolver) RevokeEndUserAPIToken(ctx context.Context, id string) (*generated.RevokeAPITokenPayload, error) {
	svc := r.APITokenService
	if svc == nil {
		return &generated.RevokeAPITokenPayload{
			Error: &generated.InvalidInput{Message: "api token service not initialized"},
		}, nil
	}

	endUserID, err := ctxutils.GetUserIDFromContext(ctx)
	if err != nil {
		return &generated.RevokeAPITokenPayload{
			Error: &generated.InvalidInput{Message: "unauthenticated"},
		}, nil
	}
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return &generated.RevokeAPITokenPayload{
			Error: &generated.InvalidInput{Message: "org context missing"},
		}, nil
	}

	if err := svc.RevokeAPIToken(ctx, id, orgName, endUserID); err != nil {
		return &generated.RevokeAPITokenPayload{
			Error: &generated.APITokenNotFound{Message: err.Error()},
		}, nil
	}

	success := true
	return &generated.RevokeAPITokenPayload{Success: &success}, nil
}

// toGQLAPIToken 将领域对象转换为 GraphQL 类型。
func toGQLAPIToken(t *appEnduser.DomainAPIToken) *generated.EndUserAPIToken {
	return &generated.EndUserAPIToken{
		ID:         t.ID,
		Name:       t.Name,
		CreatedAt:  t.CreatedAt,
		ExpiresAt:  t.ExpiresAt,
		LastUsedAt: t.LastUsedAt,
	}
}
```

> **注意：** `toGQLAPIToken` 的参数类型以实际 import 为准。领域类型在 `internal/domain/enduser` 包，import alias 为 `domainEnduser`；如需要则调整。

- [ ] **Step 2: 编译验证**

```bash
cd modelcraft-backend
go build ./...
```

Expected: 编译成功。若有类型不匹配，按 gqlgen 生成的 `generated.EndUserAPIToken` 字段名调整。

- [ ] **Step 3: 手动测试（curl）**

启动后端服务后，用有效 EndUser JWT 调用：

```bash
curl -s -X POST http://localhost:8080/end-user/graphql/org/<orgName>/ \
  -H "Authorization: Bearer <enduser_jwt>" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ endUserAPITokens { id name createdAt } }"}'
```

Expected: 返回空数组 `{"data":{"endUserAPITokens":[]}}`。

- [ ] **Step 4: Commit**

```bash
git add internal/interfaces/graphql/org/end_user_api_token.resolvers.go
git commit -m "feat(resolver): implement EndUser API Token query and mutations"
```

---

## Task 8: 前端 GraphQL 文档

**Files:**
- Modify: `modelcraft-front/src/api-client/end-user/graphql-docs.ts`

- [ ] **Step 1: 在 graphql-docs.ts 末尾追加 PAT 相关文档**

```typescript
// === API Token (PAT) ===

export const END_USER_API_TOKENS = gql`
  query EndUserAPITokens {
    endUserAPITokens {
      id
      name
      createdAt
      expiresAt
      lastUsedAt
    }
  }
`

export const CREATE_END_USER_API_TOKEN = gql`
  mutation CreateEndUserAPIToken($name: String!, $expiresAt: Time) {
    createEndUserAPIToken(name: $name, expiresAt: $expiresAt) {
      token {
        id
        name
        createdAt
        expiresAt
      }
      plaintext
      error {
        ... on InvalidInput {
          message
        }
        ... on APITokenNameConflict {
          message
        }
        ... on APITokenLimitReached {
          message
          limit
        }
      }
    }
  }
`

export const REVOKE_END_USER_API_TOKEN = gql`
  mutation RevokeEndUserAPIToken($id: ID!) {
    revokeEndUserAPIToken(id: $id) {
      success
      error {
        ... on APITokenNotFound {
          message
        }
        ... on InvalidInput {
          message
        }
      }
    }
  }
`
```

- [ ] **Step 2: 编译验证**

```bash
cd modelcraft-front
npx tsc --noEmit 2>&1 | grep graphql-docs
```

Expected: 无错误输出。

- [ ] **Step 3: Commit**

```bash
git add src/api-client/end-user/graphql-docs.ts
git commit -m "feat(api-client): add EndUser API Token GraphQL documents"
```

---

## Task 9: 前端 — 侧边栏导航项

**Files:**
- Modify: `modelcraft-front/src/web/components/features/layout/EndUserAppLayout.tsx`

- [ ] **Step 1: 更新 ActivePage 类型**

在 `EndUserAppLayout.tsx` 中，将：

```typescript
type ActivePage = 'projects' | 'cli'
```

改为：

```typescript
type ActivePage = 'projects' | 'cli' | 'tokens'
```

- [ ] **Step 2: 在导航 import 中加入 KeyRound 图标**

```typescript
import { ChevronRight, Terminal, PanelLeftClose, PanelLeftOpen, KeyRound } from 'lucide-react'
```

- [ ] **Step 3: 在侧边栏导航 CLI 按钮后新增 Tokens 按钮**

在 `EndUserAppLayoutInner` 的 `<nav>` 内，CLI 按钮之后追加：

```tsx
{/* API Token */}
<button
  type="button"
  onClick={() => router.push(`/end-user/${orgName}/dashboard/tokens`)}
  title={collapsed ? 'API Token' : undefined}
  className={cn(
    'flex w-full items-center rounded-md border-l-[3px] py-2 text-[13px] font-medium transition-colors duration-150',
    collapsed ? 'justify-center px-2' : 'gap-2.5 px-3',
    activePage === 'tokens'
      ? 'border-l-primary bg-primary/[0.08] text-primary'
      : 'border-l-transparent text-muted-foreground hover:bg-accent/50 hover:text-foreground'
  )}
>
  <KeyRound
    className={cn(
      'size-4 flex-shrink-0',
      activePage === 'tokens' ? 'opacity-100' : 'opacity-50'
    )}
    strokeWidth={1.5}
  />
  {!collapsed && <span className="flex-1 text-left">API Token</span>}
</button>
```

- [ ] **Step 4: Lint 验证**

```bash
cd modelcraft-front
npx eslint src/web/components/features/layout/EndUserAppLayout.tsx
```

Expected: `ESLint: No issues found`

- [ ] **Step 5: Commit**

```bash
git add src/web/components/features/layout/EndUserAppLayout.tsx
git commit -m "feat(layout): add API Token nav item to EndUserAppLayout sidebar"
```

---

## Task 10: 前端 — Token 管理页面

**Files:**
- Create: `modelcraft-front/src/app/end-user/[orgName]/dashboard/tokens/page.tsx`
- Create: `modelcraft-front/src/app/end-user/[orgName]/dashboard/tokens/_components/TokenTable.tsx`
- Create: `modelcraft-front/src/app/end-user/[orgName]/dashboard/tokens/_components/CreateTokenDialog.tsx`
- Create: `modelcraft-front/src/app/end-user/[orgName]/dashboard/tokens/_components/TokenRevealDialog.tsx`

- [ ] **Step 1: 创建 TokenRevealDialog（一次性展示弹窗）**

```tsx
// src/app/end-user/[orgName]/dashboard/tokens/_components/TokenRevealDialog.tsx
'use client'

import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'

interface TokenRevealDialogProps {
  plaintext: string
  onClose: () => void
}

export function TokenRevealDialog({ plaintext, onClose }: TokenRevealDialogProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(plaintext)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Dialog open onOpenChange={() => {}}>
      <DialogContent
        className="sm:max-w-md"
        onPointerDownOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <DialogTitle>Token 已创建</DialogTitle>
          <DialogDescription className="text-amber-600">
            ⚠️ 请立即复制。此 Token 仅展示一次，关闭后无法再次查看。
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-2 rounded-md border bg-muted px-3 py-2">
          <code className="min-w-0 flex-1 break-all font-mono text-xs text-foreground">
            {plaintext}
          </code>
          <Button
            variant="ghost"
            size="icon"
            className="size-7 flex-shrink-0"
            onClick={() => void handleCopy()}
          >
            {copied ? (
              <Check className="size-4 text-green-500" />
            ) : (
              <Copy className="size-4" />
            )}
          </Button>
        </div>

        <DialogFooter>
          <Button onClick={onClose} disabled={!copied} className="w-full">
            {copied ? '我已复制，关闭' : '请先复制 Token'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: 创建 CreateTokenDialog（新建 Token 弹窗）**

```tsx
// src/app/end-user/[orgName]/dashboard/tokens/_components/CreateTokenDialog.tsx
'use client'

import { useState } from 'react'
import { useMutation } from '@apollo/client'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@web/components/ui/radio-group'
import { CREATE_END_USER_API_TOKEN } from '@api-client/end-user/graphql-docs'

type ExpiryOption = 'never' | '30d' | '90d'

interface CreateTokenDialogProps {
  onClose: () => void
  onCreated: (plaintext: string) => void
  onRefetch: () => void
}

export function CreateTokenDialog({ onClose, onCreated, onRefetch }: CreateTokenDialogProps) {
  const [name, setName] = useState('')
  const [expiry, setExpiry] = useState<ExpiryOption>('never')
  const [error, setError] = useState('')

  const [createToken, { loading }] = useMutation(CREATE_END_USER_API_TOKEN)

  const getExpiresAt = (): string | null => {
    if (expiry === 'never') return null
    const days = expiry === '30d' ? 30 : 90
    const date = new Date()
    date.setDate(date.getDate() + days)
    return date.toISOString()
  }

  const handleSubmit = async () => {
    if (!name.trim()) {
      setError('请输入 Token 名称')
      return
    }
    setError('')

    const { data } = await createToken({
      variables: { name: name.trim(), expiresAt: getExpiresAt() },
    })

    const payload = data?.createEndUserAPIToken
    if (payload?.error) {
      setError(payload.error.message ?? '创建失败')
      return
    }
    if (payload?.plaintext) {
      onRefetch()
      onCreated(payload.plaintext)
    }
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>新建 API Token</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="token-name">名称</Label>
            <Input
              id="token-name"
              placeholder="例如：my-cli、github-ci"
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoFocus
            />
            {error && <p className="text-xs text-destructive">{error}</p>}
          </div>

          <div className="space-y-1.5">
            <Label>过期时间</Label>
            <RadioGroup
              value={expiry}
              onValueChange={(v) => setExpiry(v as ExpiryOption)}
              className="space-y-1"
            >
              <div className="flex items-center gap-2">
                <RadioGroupItem value="never" id="exp-never" />
                <Label htmlFor="exp-never" className="cursor-pointer font-normal">永不过期</Label>
              </div>
              <div className="flex items-center gap-2">
                <RadioGroupItem value="30d" id="exp-30d" />
                <Label htmlFor="exp-30d" className="cursor-pointer font-normal">30 天</Label>
              </div>
              <div className="flex items-center gap-2">
                <RadioGroupItem value="90d" id="exp-90d" />
                <Label htmlFor="exp-90d" className="cursor-pointer font-normal">90 天</Label>
              </div>
            </RadioGroup>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={loading}>取消</Button>
          <Button onClick={() => void handleSubmit()} disabled={loading}>
            {loading ? '创建中...' : '创建'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 3: 创建 TokenTable（Token 列表表格）**

```tsx
// src/app/end-user/[orgName]/dashboard/tokens/_components/TokenTable.tsx
'use client'

import { useState } from 'react'
import { useMutation } from '@apollo/client'
import { Trash2 } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@web/components/ui/alert-dialog'
import { REVOKE_END_USER_API_TOKEN } from '@api-client/end-user/graphql-docs'

interface APIToken {
  id: string
  name: string
  createdAt: string
  expiresAt?: string | null
  lastUsedAt?: string | null
}

interface TokenTableProps {
  tokens: APIToken[]
  onRefetch: () => void
}

function formatDate(iso?: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
  })
}

function formatLastUsed(iso?: string | null): string {
  if (!iso) return '从未使用'
  const diff = Date.now() - new Date(iso).getTime()
  const hours = Math.floor(diff / 3600000)
  if (hours < 1) return '刚刚'
  if (hours < 24) return `${hours} 小时前`
  return `${Math.floor(hours / 24)} 天前`
}

export function TokenTable({ tokens, onRefetch }: TokenTableProps) {
  const [revokeTarget, setRevokeTarget] = useState<APIToken | null>(null)
  const [revokeToken] = useMutation(REVOKE_END_USER_API_TOKEN)

  const handleRevoke = async () => {
    if (!revokeTarget) return
    await revokeToken({ variables: { id: revokeTarget.id } })
    setRevokeTarget(null)
    onRefetch()
  }

  if (tokens.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16 text-muted-foreground">
        <p className="text-sm">暂无 API Token</p>
        <p className="mt-1 text-xs">点击右上角「新建 Token」创建第一个</p>
      </div>
    )
  }

  return (
    <>
      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/40">
              <th className="px-4 py-2.5 text-left text-xs font-medium text-muted-foreground">名称</th>
              <th className="px-4 py-2.5 text-left text-xs font-medium text-muted-foreground">创建时间</th>
              <th className="px-4 py-2.5 text-left text-xs font-medium text-muted-foreground">过期时间</th>
              <th className="px-4 py-2.5 text-left text-xs font-medium text-muted-foreground">最后使用</th>
              <th className="px-4 py-2.5 text-right text-xs font-medium text-muted-foreground">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {tokens.map((token) => (
              <tr key={token.id} className="hover:bg-muted/20">
                <td className="px-4 py-3 font-medium text-foreground">{token.name}</td>
                <td className="px-4 py-3 text-muted-foreground">{formatDate(token.createdAt)}</td>
                <td className="px-4 py-3 text-muted-foreground">
                  {token.expiresAt ? formatDate(token.expiresAt) : '永不过期'}
                </td>
                <td className="px-4 py-3 text-muted-foreground">{formatLastUsed(token.lastUsedAt)}</td>
                <td className="px-4 py-3 text-right">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-7 text-muted-foreground hover:text-destructive"
                    onClick={() => setRevokeTarget(token)}
                  >
                    <Trash2 className="size-3.5" />
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <AlertDialog open={!!revokeTarget} onOpenChange={(open) => !open && setRevokeTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>撤销 Token</AlertDialogTitle>
            <AlertDialogDescription>
              确认撤销「{revokeTarget?.name}」？撤销后该 Token 立即失效，无法恢复。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={() => void handleRevoke()}
            >
              确认撤销
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
```

- [ ] **Step 4: 创建主页面 page.tsx**

```tsx
// src/app/end-user/[orgName]/dashboard/tokens/page.tsx
'use client'

import { useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import { useQuery, ApolloProvider } from '@apollo/client'
import { Plus } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { EndUserAppLayout } from '@web/components/features/layout/EndUserAppLayout'
import { createEndUserOrgScopedClient } from '@api-client/apollo/clients'
import { END_USER_API_TOKENS } from '@api-client/end-user/graphql-docs'
import { TokenTable } from './_components/TokenTable'
import { CreateTokenDialog } from './_components/CreateTokenDialog'
import { TokenRevealDialog } from './_components/TokenRevealDialog'

function TokensPageContent() {
  const params = useParams<{ orgName: string }>()
  const orgName = params.orgName
  const [showCreate, setShowCreate] = useState(false)
  const [revealPlaintext, setRevealPlaintext] = useState<string | null>(null)

  const { data, loading, refetch } = useQuery(END_USER_API_TOKENS)
  const tokens = data?.endUserAPITokens ?? []

  return (
    <EndUserAppLayout orgName={orgName} activePage="tokens">
      <div className="mx-auto max-w-4xl px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-lg font-semibold text-foreground">API Token</h1>
            <p className="mt-0.5 text-sm text-muted-foreground">
              创建 Personal Access Token 供 CLI 或外部程序使用，无需每次登录。
            </p>
          </div>
          <Button size="sm" onClick={() => setShowCreate(true)}>
            <Plus className="mr-1.5 size-4" />
            新建 Token
          </Button>
        </div>

        {loading ? (
          <div className="flex justify-center py-16">
            <div className="size-5 animate-spin rounded-full border-2 border-border border-t-foreground" />
          </div>
        ) : (
          <TokenTable tokens={tokens} onRefetch={() => void refetch()} />
        )}
      </div>

      {showCreate && (
        <CreateTokenDialog
          onClose={() => setShowCreate(false)}
          onCreated={(plaintext) => {
            setShowCreate(false)
            setRevealPlaintext(plaintext)
          }}
          onRefetch={() => void refetch()}
        />
      )}

      {revealPlaintext && (
        <TokenRevealDialog
          plaintext={revealPlaintext}
          onClose={() => setRevealPlaintext(null)}
        />
      )}
    </EndUserAppLayout>
  )
}

export default function TokensPage() {
  const params = useParams<{ orgName: string }>()
  const orgName = params?.orgName ?? ''
  const client = useMemo(() => createEndUserOrgScopedClient(orgName), [orgName])

  return (
    <ApolloProvider client={client}>
      <TokensPageContent />
    </ApolloProvider>
  )
}
```

- [ ] **Step 5: Lint 验证**

```bash
cd modelcraft-front
npx eslint \
  "src/app/end-user/[orgName]/dashboard/tokens/page.tsx" \
  "src/app/end-user/[orgName]/dashboard/tokens/_components/TokenTable.tsx" \
  "src/app/end-user/[orgName]/dashboard/tokens/_components/CreateTokenDialog.tsx" \
  "src/app/end-user/[orgName]/dashboard/tokens/_components/TokenRevealDialog.tsx"
```

Expected: `ESLint: No issues found`

- [ ] **Step 6: TypeScript 验证**

```bash
cd modelcraft-front
npx tsc --noEmit 2>&1 | grep "dashboard/tokens"
```

Expected: 无错误输出。

- [ ] **Step 7: Commit**

```bash
git add \
  src/app/end-user/\[orgName\]/dashboard/tokens/page.tsx \
  src/app/end-user/\[orgName\]/dashboard/tokens/_components/
git commit -m "feat(frontend): add API Token management page with create/revoke UI"
```

---

## Task 11: 端到端验证

- [ ] **Step 1: 启动后端服务**

```bash
cd deploy
docker-compose -f compose/docker-compose.local.yml up -d --build modelcraft-backend
```

- [ ] **Step 2: 验证 PAT 创建流程**

```bash
# 1. 获取 EndUser JWT（用已有账号登录）
TOKEN=$(curl -s -X POST http://localhost:8080/api/end-user/auth/login \
  -H "Content-Type: application/json" \
  -d '{"orgName":"<org>","username":"<user>","password":"<pass>"}' \
  | jq -r '.accessToken')

# 2. 创建 PAT
PAT_RESPONSE=$(curl -s -X POST http://localhost:8080/end-user/graphql/org/<org>/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query":"mutation { createEndUserAPIToken(name: \"test-token\") { token { id name } plaintext error { ... on InvalidInput { message } } } }"}')

echo $PAT_RESPONSE | jq .
```

Expected: 返回 `plaintext` 字段，格式为 `mc_pat_xxx`。

- [ ] **Step 3: 验证 PAT 可用于 API 调用**

```bash
PAT=$(echo $PAT_RESPONSE | jq -r '.data.createEndUserAPIToken.plaintext')

# 用 PAT 直接调用 GraphQL（无需 JWT）
curl -s -X POST http://localhost:8080/end-user/graphql/org/<org>/ \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ endUserAPITokens { id name createdAt } }"}'
```

Expected: 返回刚创建的 token 列表。

- [ ] **Step 4: 验证 Token 列表 & 撤销**

```bash
# 撤销 token（用 JWT 调用）
TOKEN_ID=$(echo $PAT_RESPONSE | jq -r '.data.createEndUserAPIToken.token.id')
curl -s -X POST http://localhost:8080/end-user/graphql/org/<org>/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"query\":\"mutation { revokeEndUserAPIToken(id: \\\"$TOKEN_ID\\\") { success } }\"}"
```

Expected: `{"data":{"revokeEndUserAPIToken":{"success":true}}}`

- [ ] **Step 5: 验证撤销后 PAT 失效**

```bash
curl -s -X POST http://localhost:8080/end-user/graphql/org/<org>/ \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ endUserAPITokens { id } }"}'
```

Expected: 返回 401 Unauthorized。

- [ ] **Step 6: 验证前端页面正常运行**

在浏览器访问：`http://localhost:3002/end-user/<orgName>/dashboard/tokens`

验证：
- 侧边栏「API Token」导航项高亮
- 可新建 token，弹窗展示明文（复制后才能关闭）
- Token 列表正常显示
- 撤销按钮触发确认弹窗，撤销后列表刷新

- [ ] **Step 7: 最终 Commit**

```bash
git add -A
git commit -m "feat: end-user API Token (PAT) — full implementation complete"
```
