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
//
// Note:
// - It operates on end_user_users in mc_meta.
// - Tenant isolation is enforced by org_name.
// - Query/Save methods follow plan contract:
//   - not found -> (nil, nil)
//   - update/delete must check RowsAffected.
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

// Save creates a new end-user.
func (r *SqlEndUserRepository) Save(ctx context.Context, user *enduser.EndUser) error {
	const query = `
		INSERT INTO end_user_users (
			id, org_name, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	orgName := user.OrgName
	if orgName == "" {
		orgName = r.orgName
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		orgName,
		user.Username,
		user.Password.Hash,
		boolToTinyInt(user.IsForbidden),
		boolToTinyInt(user.IsBuiltin),
		nullableCreatedBy(user.CreatedBy),
	)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	return nil
}

// GetByID retrieves an end-user by ID.
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetByID(ctx context.Context, orgName, id string) (*enduser.EndUser, error) {
	const query = `
		SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
		FROM end_user_users
		WHERE id = ? AND org_name = ? AND deleted_at = 0
	`

	if orgName == "" {
		orgName = r.orgName
	}

	row := r.db.QueryRowContext(ctx, query, id, orgName)
	return scanEndUser(row, orgName)
}

// GetByUsername retrieves an end-user by username.
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetByUsername(ctx context.Context, orgName, username string) (*enduser.EndUser, error) {
	const query = `
		SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
		FROM end_user_users
		WHERE username = ? AND org_name = ? AND deleted_at = 0
	`

	if orgName == "" {
		orgName = r.orgName
	}

	row := r.db.QueryRowContext(ctx, query, username, orgName)
	return scanEndUser(row, orgName)
}

func scanEndUser(row *sql.Row, orgName string) (*enduser.EndUser, error) {
	var (
		id          string
		username    string
		password    string
		isForbidden int
		isBuiltin   int
		createdBy   sql.NullString
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(
		&id,
		&username,
		&password,
		&isForbidden,
		&isBuiltin,
		&createdBy,
		&createdAt,
		&updatedAt,
	)
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
		Password:    enduser.NewHashedPasswordFromHash(password),
		IsForbidden: isForbidden == 1,
		IsBuiltin:   isBuiltin == 1,
		CreatedBy:   createdBy.String,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func nullableCreatedBy(createdBy string) any {
	if createdBy == "" {
		return nil
	}
	return createdBy
}

// UpdateStatus updates user's is_forbidden field.
// Returns NO_ROWS_AFFECTED when user does not exist.
func (r *SqlEndUserRepository) UpdateStatus(ctx context.Context, orgName, id string, isForbidden bool) error {
	const query = `
		UPDATE end_user_users
		SET is_forbidden = ?, updated_at = NOW()
		WHERE id = ? AND org_name = ? AND deleted_at = 0
	`

	if orgName == "" {
		orgName = r.orgName
	}

	result, err := r.db.ExecContext(ctx, query, boolToTinyInt(isForbidden), id, orgName)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, fmt.Sprintf("end user not found: %s", id))
	}

	return nil
}

// Delete soft-deletes an end-user.
// Returns NO_ROWS_AFFECTED when user does not exist.
func (r *SqlEndUserRepository) Delete(ctx context.Context, orgName, id string) error {
	const query = `
		UPDATE end_user_users
		SET deleted_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
		    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED),
		    updated_at = NOW()
		WHERE id = ? AND org_name = ? AND deleted_at = 0
	`

	if orgName == "" {
		orgName = r.orgName
	}

	result, err := r.db.ExecContext(ctx, query, id, orgName)
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

	// Total count (without cursor, with optional search)
	countSQL := `SELECT COUNT(*) FROM end_user_users WHERE org_name = ? AND deleted_at = 0`
	countArgs := make([]interface{}, 0, 2)
	countArgs = append(countArgs, query.OrgName)
	if query.Search != "" {
		countSQL += ` AND username LIKE CONCAT('%', ?, '%')`
		countArgs = append(countArgs, query.Search)
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}

	// List with search + cursor
	listSQL := `
		SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
		FROM end_user_users
		WHERE org_name = ?
		  AND deleted_at = 0
		  AND (? = '' OR username LIKE CONCAT('%', ?, '%'))
		  AND (? = '' OR id > ?)
		ORDER BY id ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(
		ctx,
		listSQL,
		query.OrgName,
		query.Search,
		query.Search,
		query.After,
		query.After,
		first,
	)
	if err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}
	defer rows.Close()

	items := make([]*enduser.EndUser, 0, first)
	for rows.Next() {
		var (
			id          string
			username    string
			password    string
			isForbidden int
			isBuiltin   int
			createdBy   sql.NullString
			createdAt   time.Time
			updatedAt   time.Time
		)

		if scanErr := rows.Scan(
			&id,
			&username,
			&password,
			&isForbidden,
			&isBuiltin,
			&createdBy,
			&createdAt,
			&updatedAt,
		); scanErr != nil {
			return nil, 0, sqlerr.WrapSQLError(scanErr)
		}

		items = append(items, &enduser.EndUser{
			ID:          id,
			OrgName:     query.OrgName,
			Username:    username,
			Password:    enduser.NewHashedPasswordFromHash(password),
			IsForbidden: isForbidden == 1,
			IsBuiltin:   isBuiltin == 1,
			CreatedBy:   createdBy.String,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}

	return items, total, nil
}

func boolToTinyInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// ListAccessibleProjectsByRoleAssignment 通过 end_user_role_users JOIN end_user_roles
// 查询用户在该 Org 下可访问的 Project 列表（替代旧的 end_user_project_access 路径）。
func (r *SqlEndUserRepository) ListAccessibleProjectsByRoleAssignment(
	ctx context.Context,
	orgName, endUserID string,
) ([]enduser.AccessibleProject, error) {
	const query = `
		SELECT DISTINCT ur.role_id, r.project_slug, COALESCE(p.title, r.project_slug) AS project_title
		FROM end_user_role_users ur
		JOIN end_user_roles r
		  ON r.id = ur.role_id
		 AND r.org_name = ur.org_name
		LEFT JOIN projects p
		  ON p.org_name = r.org_name
		 AND p.slug = r.project_slug
		 AND p.deleted_at = 0
		WHERE ur.org_name = ?
		  AND ur.user_id = ?
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
		var roleID, projectSlug, projectTitle string
		if scanErr := rows.Scan(&roleID, &projectSlug, &projectTitle); scanErr != nil {
			return nil, sqlerr.WrapSQLError(scanErr)
		}
		if _, ok := seen[projectSlug]; !ok {
			seen[projectSlug] = struct{}{}
			projects = append(projects, enduser.AccessibleProject{
				ProjectSlug:  projectSlug,
				ProjectTitle: projectTitle,
			})
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
		FROM end_user_role_users ur
		JOIN end_user_roles r
		  ON r.id = ur.role_id
		 AND r.org_name = ur.org_name
		WHERE ur.org_name = ?
		  AND ur.user_id = ?
		  AND r.project_slug = ?
		  AND r.deleted_at = 0
	`

	var count int64
	if err := r.db.QueryRowContext(ctx, query, orgName, endUserID, projectSlug).Scan(&count); err != nil {
		return false, sqlerr.WrapSQLError(err)
	}
	return count > 0, nil
}

// GetBuiltinByOrg retrieves the builtin admin EndUser for an org.
// Returns (nil, nil) when not found.
func (r *SqlEndUserRepository) GetBuiltinByOrg(ctx context.Context, orgName string) (*enduser.EndUser, error) {
	const query = `
		SELECT id, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at
		FROM end_user_users
		WHERE org_name = ? AND is_builtin = 1
		LIMIT 1
	`
	if orgName == "" {
		orgName = r.orgName
	}
	row := r.db.QueryRowContext(ctx, query, orgName)
	return scanEndUser(row, orgName)
}

// Compile-time interface satisfaction check.
var _ enduser.EndUserRepository = (*SqlEndUserRepository)(nil)
