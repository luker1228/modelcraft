# Unified User System — Plan 2c: EndUser Repository 새 테이블 적용

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sql_enduser_repository.go` 의 SQL 쿼리를 삭제된 `end_user_users`/`end_user_roles`/`end_user_role_users` 에서 새 테이블 `users`/`user_orgs`/`project_roles`/`project_role_users` 로 업데이트해 실제 동작하게 한다.

**Architecture:** `EndUser` 도메인 엔티티를 유지하되 `IsBuiltin`/`CreatedBy` 필드를 제거(새 `users` 테이블에 없음), `IsForbidden` 을 `user_orgs.status` 로 대체. SQL 쿼리를 직접 새 테이블에 맞게 재작성. `EndUserSessionRepository` (refresh token 저장)도 `end_user_accounts` 대신 통합 `refresh_tokens` 테이블로 전환.

**Tech Stack:** Go, MySQL, sqlc raw SQL

**현재 상태:**
- `sql_enduser_repository.go` — 실제 구현이지만 `end_user_users` 등 삭제된 테이블 참조
- `sql_enduser_session_repository.go` — `end_user_accounts` 테이블 참조 (삭제됨)
- `go build ./...` — ✅ 현재 컴파일은 되나, 런타임에서 실패

**새 테이블 필드 매핑:**

| EndUser 필드 | SQL 소스 | 비고 |
|---|---|---|
| `ID` | `users.id` | |
| `OrgName` | `user_orgs.org_name` | JOIN users ON user_orgs.user_id = users.id |
| `Username` | `users.name` | users.name = username |
| `Password.Hash` | `users.password_hash` | |
| `IsForbidden` | `user_orgs.status != 'active'` | status='suspended' → is_forbidden=true |
| `IsBuiltin` | **폐기** → 항상 false | |
| `CreatedBy` | **폐기** → 빈 문자열 | |

---

## 文件变更地图

| 操作 | 文件 |
|------|------|
| **修改** | `internal/domain/enduser/end_user.go` — 移除 `IsBuiltin`/`CreatedBy`，`IsForbidden` 통합 status |
| **修改** | `internal/domain/enduser/end_user_repository.go` — `GetBuiltinByOrg` 제거 |
| **修改** | `internal/infrastructure/repository/sql_enduser_repository.go` — SQL 全部改为新表 |
| **修改** | `internal/infrastructure/repository/sql_enduser_session_repository.go` — 改用 `refresh_tokens` 表 |
| **修改** | 引用 `IsBuiltin`/`CreatedBy`/`GetBuiltinByOrg` 的调用方 |

---

## Task 1: EndUser domain 엔티티 정리 — IsBuiltin/CreatedBy 제거

**Files:**
- Modify: `modelcraft-backend/internal/domain/enduser/end_user.go`
- Modify: `modelcraft-backend/internal/domain/enduser/end_user_repository.go`

- [ ] **Step 1: 读取当前 end_user.go**

```bash
cat modelcraft-backend/internal/domain/enduser/end_user.go
```

- [ ] **Step 2: 修改 EndUser struct — 移除 IsBuiltin 和 CreatedBy**

将：
```go
type EndUser struct {
    ID          string
    OrgName     string
    Username    string
    Password    HashedPassword
    IsForbidden bool
    IsBuiltin   bool
    CreatedBy   string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```
改为：
```go
type EndUser struct {
    ID          string
    OrgName     string
    Username    string
    Password    HashedPassword
    IsForbidden bool  // true when user_orgs.status = 'suspended'
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

- [ ] **Step 3: 更新 NewEndUser 构造函数 — 移除 createdBy 参数**

将 `NewEndUser(id, orgName, username, createdBy string, hashedPwd HashedPassword)` 改为：
```go
func NewEndUser(id, orgName, username string, hashedPwd HashedPassword) (*EndUser, error) {
```
（移除 `createdBy` 参数）

- [ ] **Step 4: 删除 NewBuiltinEndUser 构造函数和 BuiltinAdminUsername 常量**

（builtin 개념 폐기）

- [ ] **Step 5: 更新 end_user_repository.go — 移除 GetBuiltinByOrg**

将 `EndUserRepository` 接口中的 `GetBuiltinByOrg` 방법 제거.

- [ ] **Step 6: 检查所有调用方**

```bash
grep -rn "IsBuiltin\|CreatedBy\|GetBuiltinByOrg\|NewBuiltinEndUser\|BuiltinAdminUsername\|NewEndUser(" \
  modelcraft-backend/internal/ --include="*.go" | grep -v "end_user.go\|end_user_repository.go"
```

对每个调用方做相应修改（移除 createdBy 参数、移除 IsBuiltin 字段引用）。

- [ ] **Step 7: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/domain/enduser/... 2>&1
```

- [ ] **Step 8: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/domain/enduser/
git commit -m "domain: remove EndUser.IsBuiltin/CreatedBy, remove GetBuiltinByOrg"
```

---

## Task 2: EndUser SQL 쿼리를 새 테이블로 재작성

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go`

`end_user_users` → `users JOIN user_orgs`，`end_user_role_users` → `project_role_users`，`end_user_roles` → `project_roles`。

- [ ] **Step 1: 读取当前文件**

```bash
cat modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go
```

- [ ] **Step 2: 用 Write 工具完整替换文件**

```go
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"
)

// SqlEndUserRepository is the MySQL implementation of enduser.EndUserRepository.
// After unification, end-users are stored in the unified `users` table with
// org binding in `user_orgs`. Project access is controlled by `project_role_users`.
type endUserDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SqlEndUserRepository struct {
	db      endUserDBTX
	orgName string
}

// NewSqlEndUserRepository creates a SqlEndUserRepository.
func NewSqlEndUserRepository(db endUserDBTX, orgName, _ string) enduser.EndUserRepository {
	return &SqlEndUserRepository{
		db:      db,
		orgName: orgName,
	}
}

// Save creates a new end-user in users + user_orgs tables.
// For end-user registration, user_orgs.is_admin = false always.
func (r *SqlEndUserRepository) Save(ctx context.Context, user *enduser.EndUser) error {
	orgName := user.OrgName
	if orgName == "" {
		orgName = r.orgName
	}

	// 1. Insert into users
	const insertUser = `
		INSERT INTO users (id, name, phone, password_hash, deleted_at, delete_token, created_at, updated_at)
		VALUES (?, ?, '', ?, 0, 0, NOW(3), NOW(3))
	`
	if _, err := r.db.ExecContext(ctx, insertUser, user.ID, user.Username, user.Password.Hash); err != nil {
		return sqlerr.WrapSQLError(err)
	}

	// 2. Insert into user_orgs (is_admin=false for end-users, single org)
	userOrgID := user.ID + "-org" // deterministic ID: avoid extra UUID generation
	const insertUserOrg = `
		INSERT INTO user_orgs (id, user_id, org_name, is_admin, status, deleted_at, delete_token, created_at, updated_at)
		VALUES (?, ?, ?, 0, 'active', 0, 0, NOW(3), NOW(3))
	`
	if _, err := r.db.ExecContext(ctx, insertUserOrg, userOrgID, user.ID, orgName); err != nil {
		return sqlerr.WrapSQLError(err)
	}

	return nil
}

// GetByID retrieves an end-user by ID.
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetByID(ctx context.Context, orgName, id string) (*enduser.EndUser, error) {
	if orgName == "" {
		orgName = r.orgName
	}
	const query = `
		SELECT u.id, u.name, u.password_hash, uo.status, u.created_at, u.updated_at, uo.org_name
		FROM users u
		JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
		WHERE u.id = ? AND u.deleted_at = 0
	`
	row := r.db.QueryRowContext(ctx, query, orgName, id)
	return scanEndUser(row)
}

// GetByUsername retrieves an end-user by username (users.name field).
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetByUsername(ctx context.Context, orgName, username string) (*enduser.EndUser, error) {
	if orgName == "" {
		orgName = r.orgName
	}
	const query = `
		SELECT u.id, u.name, u.password_hash, uo.status, u.created_at, u.updated_at, uo.org_name
		FROM users u
		JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
		WHERE u.name = ? AND u.deleted_at = 0
	`
	row := r.db.QueryRowContext(ctx, query, orgName, username)
	return scanEndUser(row)
}

func scanEndUser(row *sql.Row) (*enduser.EndUser, error) {
	var (
		id          string
		username    string
		passwordHash string
		status      string
		createdAt   time.Time
		updatedAt   time.Time
		orgName     string
	)

	err := row.Scan(&id, &username, &passwordHash, &status, &createdAt, &updatedAt, &orgName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //nolint:nilnil // per contract: not found is expected in repo layer
		}
		return nil, sqlerr.WrapSQLError(err)
	}

	return &enduser.EndUser{
		ID:          id,
		OrgName:     orgName,
		Username:    username,
		Password:    enduser.NewHashedPasswordFromHash(passwordHash),
		IsForbidden: status != "active",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// UpdatePassword updates the password hash for a user.
func (r *SqlEndUserRepository) UpdatePassword(
	ctx context.Context,
	orgName, id string,
	hashedPassword enduser.HashedPassword,
) error {
	const query = `
		UPDATE users
		SET password_hash = ?, updated_at = NOW(3)
		WHERE id = ? AND deleted_at = 0
	`
	result, err := r.db.ExecContext(ctx, query, hashedPassword.Hash, id)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
	}
	return nil
}

// UpdateStatus updates user's is_forbidden field via user_orgs.status.
// isForbidden=true → status='suspended'; isForbidden=false → status='active'.
func (r *SqlEndUserRepository) UpdateStatus(ctx context.Context, orgName, id string, isForbidden bool) error {
	if orgName == "" {
		orgName = r.orgName
	}
	status := "active"
	if isForbidden {
		status = "suspended"
	}
	const query = `
		UPDATE user_orgs
		SET status = ?, updated_at = NOW(3)
		WHERE user_id = ? AND org_name = ? AND deleted_at = 0
	`
	result, err := r.db.ExecContext(ctx, query, status, id, orgName)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
	}
	return nil
}

// Delete soft-deletes an end-user from both users and user_orgs.
func (r *SqlEndUserRepository) Delete(ctx context.Context, orgName, id string) error {
	if orgName == "" {
		orgName = r.orgName
	}

	// Soft-delete user_orgs first (FK child)
	const deleteOrg = `
		UPDATE user_orgs
		SET deleted_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
		    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED),
		    updated_at = NOW(3)
		WHERE user_id = ? AND org_name = ? AND deleted_at = 0
	`
	if _, err := r.db.ExecContext(ctx, deleteOrg, id, orgName); err != nil {
		return sqlerr.WrapSQLError(err)
	}

	// Soft-delete users
	const deleteUser = `
		UPDATE users
		SET deleted_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
		    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED),
		    updated_at = NOW(3)
		WHERE id = ? AND deleted_at = 0
	`
	result, err := r.db.ExecContext(ctx, deleteUser, id)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
	}
	return nil
}

// ListWithTotal retrieves users with cursor pagination and total count.
func (r *SqlEndUserRepository) ListWithTotal(
	ctx context.Context,
	query enduser.ListEndUsersQuery,
) ([]*enduser.EndUser, int64, error) {
	first := query.First
	if first <= 0 {
		first = 20
	}
	if first > 100 {
		first = 100
	}

	orgName := query.OrgName
	if orgName == "" {
		orgName = r.orgName
	}

	// Total count
	countSQL := `
		SELECT COUNT(*)
		FROM users u
		JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
		WHERE u.deleted_at = 0
	`
	countArgs := []interface{}{orgName}
	if query.Search != "" {
		countSQL += ` AND u.name LIKE CONCAT('%', ?, '%')`
		countArgs = append(countArgs, query.Search)
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}

	// List
	listSQL := `
		SELECT u.id, u.name, u.password_hash, uo.status, u.created_at, u.updated_at, uo.org_name
		FROM users u
		JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
		WHERE u.deleted_at = 0
		  AND (? = '' OR u.name LIKE CONCAT('%', ?, '%'))
		  AND (? = '' OR u.id > ?)
		ORDER BY u.id ASC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(
		ctx, listSQL,
		orgName,
		query.Search, query.Search,
		query.After, query.After,
		first,
	)
	if err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}
	defer rows.Close()

	items := make([]*enduser.EndUser, 0, first)
	for rows.Next() {
		var (
			id           string
			username     string
			passwordHash string
			status       string
			createdAt    time.Time
			updatedAt    time.Time
			rowOrgName   string
		)
		if scanErr := rows.Scan(&id, &username, &passwordHash, &status, &createdAt, &updatedAt, &rowOrgName); scanErr != nil {
			return nil, 0, sqlerr.WrapSQLError(scanErr)
		}
		items = append(items, &enduser.EndUser{
			ID:          id,
			OrgName:     rowOrgName,
			Username:    username,
			Password:    enduser.NewHashedPasswordFromHash(passwordHash),
			IsForbidden: status != "active",
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}
	return items, total, nil
}

// ListAccessibleProjectsByRoleAssignment 通过 project_role_users JOIN project_roles 查询用户可访问的项目。
func (r *SqlEndUserRepository) ListAccessibleProjectsByRoleAssignment(
	ctx context.Context,
	orgName, endUserID string,
) ([]enduser.AccessibleProject, error) {
	const query = `
		SELECT DISTINCT
		  r.project_slug,
		  COALESCE(p.title, r.project_slug) AS project_title,
		  COALESCE(p.description, '')        AS project_description,
		  COALESCE(p.status, 'active')       AS project_status,
		  p.created_at,
		  p.updated_at
		FROM project_role_users ur
		JOIN project_roles r
		  ON r.id = ur.role_id
		 AND r.org_name = ur.org_name
		LEFT JOIN projects p
		  ON p.org_name = r.org_name
		 AND p.slug = r.project_slug
		 AND p.deleted_at = 0
		WHERE ur.org_name = ?
		  AND ur.user_id  = ?
		  AND r.deleted_at = 0
		ORDER BY r.project_slug ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orgName, endUserID)
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	projects := make([]enduser.AccessibleProject, 0)
	for rows.Next() {
		var p enduser.AccessibleProject
		var createdAt, updatedAt *time.Time
		if scanErr := rows.Scan(
			&p.ProjectSlug,
			&p.ProjectTitle,
			&p.ProjectDescription,
			&p.ProjectStatus,
			&createdAt,
			&updatedAt,
		); scanErr != nil {
			return nil, sqlerr.WrapSQLError(scanErr)
		}
		if createdAt != nil {
			p.ProjectCreatedAt = *createdAt
		}
		if updatedAt != nil {
			p.ProjectUpdatedAt = *updatedAt
		}
		if _, ok := seen[p.ProjectSlug]; !ok {
			seen[p.ProjectSlug] = struct{}{}
			projects = append(projects, p)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	return projects, nil
}

// HasProjectAccessByRole 检查用户在指定 org+project 下是否有任意 Role 分配。
func (r *SqlEndUserRepository) HasProjectAccessByRole(
	ctx context.Context,
	orgName, endUserID, projectSlug string,
) (bool, error) {
	const query = `
		SELECT COUNT(1)
		FROM project_role_users ur
		JOIN project_roles r
		  ON r.id = ur.role_id
		 AND r.org_name = ur.org_name
		WHERE ur.org_name = ?
		  AND ur.user_id  = ?
		  AND r.project_slug = ?
		  AND r.deleted_at   = 0
	`
	var count int64
	if err := r.db.QueryRowContext(ctx, query, orgName, endUserID, projectSlug).Scan(&count); err != nil {
		return false, sqlerr.WrapSQLError(err)
	}
	return count > 0, nil
}

// Compile-time interface satisfaction check.
var _ enduser.EndUserRepository = (*SqlEndUserRepository)(nil)
```

- [ ] **Step 3: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1
```

- [ ] **Step 4: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go
git commit -m "repo: rewrite EndUser repository SQL for new unified tables (users/user_orgs/project_roles)"
```

---

## Task 3: EndUserSession Repository → refresh_tokens 테이블로 전환

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_enduser_session_repository.go`

`end_user_accounts` 테이블이 삭제됐다. EndUser refresh token을 통합 `refresh_tokens` 테이블에 저장한다.

- [ ] **Step 1: 读取当前文件**

```bash
cat modelcraft-backend/internal/infrastructure/repository/sql_enduser_session_repository.go
```

- [ ] **Step 2: 读取 EndUserSessionRepository 接口**

```bash
cat modelcraft-backend/internal/domain/enduser/end_user_session_repository.go
```

- [ ] **Step 3: 读取已有的 refresh_tokens 表 SQL 结构**

```bash
cat modelcraft-backend/db/schema/mysql/08_refresh_tokens.sql
grep -n "type SqlRefreshToken\|refresh_token" \
  modelcraft-backend/internal/infrastructure/repository/sql_refresh_token_repository.go 2>/dev/null | head -20
```

- [ ] **Step 4: 重写 session repository 使用 refresh_tokens 表**

`EndUserSession` 엔티티를 `RefreshToken`으로 매핑. `refresh_tokens` 테이블에 저장하되, EndUser session임을 `user_id`로 구분 (동일 `user_id`).

```go
package repository

import (
    "context"
    "database/sql"
    "time"

    "modelcraft/internal/domain/enduser"
    "modelcraft/internal/infrastructure/sqlerr"
)

// SqlEndUserSessionRepository stores end-user refresh tokens in the unified refresh_tokens table.
type SqlEndUserSessionRepository struct {
    db endUserDBTX
}

// NewSqlEndUserSessionRepository creates a SqlEndUserSessionRepository.
func NewSqlEndUserSessionRepository(db endUserDBTX, _, _ string) enduser.EndUserSessionRepository {
    return &SqlEndUserSessionRepository{db: db}
}
```

根据 `EndUserSessionRepository` 接口实现所有方法，使用 `refresh_tokens` 表（字段: id, user_id, token_hash, expires_at, revoked, created_at）。

- [ ] **Step 5: 编译检查**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/repository/... 2>&1
```

- [ ] **Step 6: Commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add modelcraft-backend/internal/infrastructure/repository/sql_enduser_session_repository.go
git commit -m "repo: EndUserSession repository delegates to unified refresh_tokens table"
```

---

## Task 4: 全局编译 + 调用方修复

- [ ] **Step 1: 全量编译，查看剩余错误**

```bash
cd modelcraft-backend && go build ./... 2>&1 | head -40
```

- [ ] **Step 2: 修复 IsBuiltin/CreatedBy 调用方**

```bash
grep -rn "IsBuiltin\|\.CreatedBy\|GetBuiltinByOrg\|NewBuiltinEndUser\|createdBy" \
  modelcraft-backend/internal/ --include="*.go" | grep -v "end_user.go\|end_user_repository.go\|sql_enduser"
```

逐一修复（移除字段引用、更新构造函数调用）。

- [ ] **Step 3: 修复 NewEndUser 调用方 (4-arg → 3-arg)**

```bash
grep -rn "NewEndUser(" modelcraft-backend/internal/ --include="*.go" | grep -v "end_user.go"
```

将所有 `NewEndUser(id, orgName, username, createdBy, pwd)` 改为 `NewEndUser(id, orgName, username, pwd)`。

- [ ] **Step 4: 编译通过**

```bash
cd modelcraft-backend && go build ./... 2>&1
```

预期：0 errors。

- [ ] **Step 5: 运行测试**

```bash
cd modelcraft-backend && go test ./internal/domain/enduser/... ./internal/app/enduser/... -v 2>&1 | tail -20
```

- [ ] **Step 6: 最终 commit**

```bash
cd /data/home/lukemxjia/modelcraft
git add -A
git commit -m "chore: Plan 2c complete — EndUser repository fully ported to unified tables"
```
