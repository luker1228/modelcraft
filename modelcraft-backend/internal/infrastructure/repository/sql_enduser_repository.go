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

const endUserStatusActive = "active"

// SqlEndUserRepository 是 enduser.EndUserRepository 的 MySQL 实现。
// 统一用户体系后，end-user 数据存储在统一的 users + user_orgs 表中。
// Project 级角色访问控制通过 project_role_users + project_roles 表管理。
type endUserDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SqlEndUserRepository struct {
	db      endUserDBTX
	orgName string
}

// NewSqlEndUserRepository creates a SqlEndUserRepository.
func NewSqlEndUserRepository(db endUserDBTX, orgName, _ string) enduser.EndUserRepository {
	return &SqlEndUserRepository{db: db, orgName: orgName}
}

// Save creates a new end-user in users + user_orgs (is_admin=false).
func (r *SqlEndUserRepository) Save(ctx context.Context, user *enduser.EndUser) error {
	orgName := user.OrgName
	if orgName == "" {
		orgName = r.orgName
	}

	const insertUser = `
		INSERT INTO users (id, name, phone, password_hash, deleted_at, delete_token, created_at, updated_at)
		VALUES (?, ?, '', ?, 0, 0, NOW(3), NOW(3))
	`
	if _, err := r.db.ExecContext(ctx, insertUser, user.ID, user.Username, user.Password.Hash); err != nil {
		return sqlerr.WrapSQLError(err)
	}

	const insertUserOrg = `
		INSERT INTO user_orgs
		  (id, user_id, org_name, is_admin, status, deleted_at, delete_token, created_at, updated_at)
		VALUES (?, ?, ?, 0, 'active', 0, 0, NOW(3), NOW(3))
	`
	userOrgID := user.ID + "-org"
	if _, err := r.db.ExecContext(ctx, insertUserOrg, userOrgID, user.ID, orgName); err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}

// GetByID retrieves an end-user by ID under org scope.
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetByID(ctx context.Context, orgName, id string) (*enduser.EndUser, error) {
	if orgName == "" {
		orgName = r.orgName
	}
	const q = `
		SELECT u.id, u.name, u.password_hash, uo.status, u.created_at, u.updated_at, uo.org_name
		FROM users u
		JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
		WHERE u.id = ? AND u.deleted_at = 0
	`
	return scanEndUser(r.db.QueryRowContext(ctx, q, orgName, id))
}

// GetByUsername retrieves an end-user by username under org scope.
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetByUsername(ctx context.Context, orgName, username string) (*enduser.EndUser, error) {
	if orgName == "" {
		orgName = r.orgName
	}
	const q = `
		SELECT u.id, u.name, u.password_hash, uo.status, u.created_at, u.updated_at, uo.org_name
		FROM users u
		JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
		WHERE u.name = ? AND u.deleted_at = 0
	`
	return scanEndUser(r.db.QueryRowContext(ctx, q, orgName, username))
}

func scanEndUser(row *sql.Row) (*enduser.EndUser, error) {
	var (
		id           string
		username     string
		passwordHash string
		status       string
		createdAt    time.Time
		updatedAt    time.Time
		orgName      string
	)
	if err := row.Scan(&id, &username, &passwordHash, &status, &createdAt, &updatedAt, &orgName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //nolint:nilnil
		}
		return nil, sqlerr.WrapSQLError(err)
	}
	return &enduser.EndUser{
		ID:          id,
		OrgName:     orgName,
		Username:    username,
		Password:    enduser.NewHashedPasswordFromHash(passwordHash),
		IsForbidden: status != endUserStatusActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// UpdatePassword updates the password hash.
func (r *SqlEndUserRepository) UpdatePassword(
	ctx context.Context, orgName, id string, hashedPassword enduser.HashedPassword,
) error {
	const q = `UPDATE users SET password_hash = ?, updated_at = NOW(3) WHERE id = ? AND deleted_at = 0`
	result, err := r.db.ExecContext(ctx, q, hashedPassword.Hash, id)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
	}
	return nil
}

// UpdateStatus updates user status via user_orgs.status.
func (r *SqlEndUserRepository) UpdateStatus(ctx context.Context, orgName, id string, isForbidden bool) error {
	if orgName == "" {
		orgName = r.orgName
	}
	status := endUserStatusActive
	if isForbidden {
		status = "suspended"
	}
	const q = `
		UPDATE user_orgs
		SET status = ?, updated_at = NOW(3)
		WHERE user_id = ? AND org_name = ? AND deleted_at = 0
	`
	result, err := r.db.ExecContext(ctx, q, status, id, orgName)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
	}
	return nil
}

// Delete soft-deletes the user from users + user_orgs.
func (r *SqlEndUserRepository) Delete(ctx context.Context, orgName, id string) error {
	if orgName == "" {
		orgName = r.orgName
	}
	const softDeleteOrg = `
		UPDATE user_orgs
		SET deleted_at   = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
		    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED),
		    updated_at   = NOW(3)
		WHERE user_id = ? AND org_name = ? AND deleted_at = 0
	`
	if _, err := r.db.ExecContext(ctx, softDeleteOrg, id, orgName); err != nil {
		return sqlerr.WrapSQLError(err)
	}
	const softDeleteUser = `
		UPDATE users
		SET deleted_at   = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
		    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED),
		    updated_at   = NOW(3)
		WHERE id = ? AND deleted_at = 0
	`
	result, err := r.db.ExecContext(ctx, softDeleteUser, id)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
	}
	return nil
}

// ListWithTotal 带游标分页列出用户。
func (r *SqlEndUserRepository) ListWithTotal(
	ctx context.Context,
	q enduser.ListEndUsersQuery,
) ([]*enduser.EndUser, int64, error) {
	first := q.First
	if first <= 0 {
		first = 20
	}
	if first > 100 {
		first = 100
	}
	orgName := q.OrgName
	if orgName == "" {
		orgName = r.orgName
	}

	countSQL := `
		SELECT COUNT(*)
		FROM users u
		JOIN user_orgs uo ON uo.user_id = u.id AND uo.org_name = ? AND uo.deleted_at = 0
		WHERE u.deleted_at = 0
	`
	countArgs := []interface{}{orgName}
	if q.Search != "" {
		countSQL += ` AND u.name LIKE CONCAT('%', ?, '%')`
		countArgs = append(countArgs, q.Search)
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}

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
	rows, err := r.db.QueryContext(ctx, listSQL, orgName, q.Search, q.Search, q.After, q.After, first)
	if err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}
	defer rows.Close()

	items := make([]*enduser.EndUser, 0, first)
	for rows.Next() {
		var (
			id, username, passwordHash, status, rowOrgName string
			createdAt, updatedAt                           time.Time
		)
		if err := rows.Scan(
			&id, &username, &passwordHash, &status, &createdAt, &updatedAt, &rowOrgName,
		); err != nil {
			return nil, 0, sqlerr.WrapSQLError(err)
		}
		items = append(items, &enduser.EndUser{
			ID:          id,
			OrgName:     rowOrgName,
			Username:    username,
			Password:    enduser.NewHashedPasswordFromHash(passwordHash),
			IsForbidden: status != endUserStatusActive,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}
	return items, total, nil
}

// ListAccessibleProjectsByRoleAssignment 通过 project_role_users + project_roles 查询用户可访问的项目。
func (r *SqlEndUserRepository) ListAccessibleProjectsByRoleAssignment(
	ctx context.Context, orgName, endUserID string,
) ([]enduser.AccessibleProject, error) {
	const q = `
		SELECT DISTINCT
		  r.project_slug,
		  COALESCE(p.title, r.project_slug)      AS project_title,
		  COALESCE(p.description, '')             AS project_description,
		  COALESCE(p.status, 'active')            AS project_status,
		  p.created_at,
		  p.updated_at
		FROM project_role_users ur
		JOIN project_roles r
		  ON r.id = ur.role_id AND r.org_name = ur.org_name
		LEFT JOIN projects p
		  ON p.org_name = r.org_name AND p.slug = r.project_slug AND p.deleted_at = 0
		WHERE ur.org_name = ?
		  AND ur.user_id  = ?
		  AND r.deleted_at = 0
		ORDER BY r.project_slug ASC
	`
	rows, err := r.db.QueryContext(ctx, q, orgName, endUserID)
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	projects := make([]enduser.AccessibleProject, 0)
	for rows.Next() {
		var p enduser.AccessibleProject
		var createdAt, updatedAt *time.Time
		if err := rows.Scan(
			&p.ProjectSlug, &p.ProjectTitle, &p.ProjectDescription,
			&p.ProjectStatus, &createdAt, &updatedAt,
		); err != nil {
			return nil, sqlerr.WrapSQLError(err)
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

// HasProjectAccessByRole 检查用户在指定 project 下是否有任意 Role 分配。
func (r *SqlEndUserRepository) HasProjectAccessByRole(
	ctx context.Context, orgName, endUserID, projectSlug string,
) (bool, error) {
	const q = `
		SELECT COUNT(1)
		FROM project_role_users ur
		JOIN project_roles r ON r.id = ur.role_id AND r.org_name = ur.org_name
		WHERE ur.org_name = ? AND ur.user_id = ? AND r.project_slug = ? AND r.deleted_at = 0
	`
	var count int64
	if err := r.db.QueryRowContext(ctx, q, orgName, endUserID, projectSlug).Scan(&count); err != nil {
		return false, sqlerr.WrapSQLError(err)
	}
	return count > 0, nil
}

// Compile-time interface check.
var _ enduser.EndUserRepository = (*SqlEndUserRepository)(nil)
