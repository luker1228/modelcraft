package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/sqlerr"
	"strings"
	"time"
)

type endUserProjectAccessDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// SqlEndUserProjectAccessRepository is the MySQL implementation of EndUserProjectAccessRepository.
type SqlEndUserProjectAccessRepository struct {
	db          endUserProjectAccessDBTX
	orgName     string
	projectSlug string
}

// NewSqlEndUserProjectAccessRepository creates a SqlEndUserProjectAccessRepository.
func NewSqlEndUserProjectAccessRepository(
	db endUserProjectAccessDBTX,
	orgName, projectSlug string,
) enduser.EndUserProjectAccessRepository {
	return &SqlEndUserProjectAccessRepository{
		db:          db,
		orgName:     orgName,
		projectSlug: projectSlug,
	}
}

func (r *SqlEndUserProjectAccessRepository) Grant(ctx context.Context, access *enduser.EndUserProjectAccess) error {
	const query = `
		INSERT INTO end_user_project_access (
			id, end_user_id, org_name, project_slug, permission_bundle_id, granted_by, granted_at
		)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`

	orgName := access.OrgName
	if orgName == "" {
		orgName = r.orgName
	}
	projectSlug := access.ProjectSlug
	if projectSlug == "" {
		projectSlug = r.projectSlug
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		access.ID,
		access.EndUserID,
		orgName,
		projectSlug,
		access.PermissionBundleID,
		nullableString(access.GrantedBy),
	)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	return nil
}

func (r *SqlEndUserProjectAccessRepository) GetByID(
	ctx context.Context,
	orgName, projectSlug, accessID string,
) (*enduser.EndUserProjectAccess, error) {
	const query = `
		SELECT
			a.id,
			a.org_name,
			a.project_slug,
			a.end_user_id,
			a.permission_bundle_id,
			COALESCE(b.name, ''),
			COALESCE(a.granted_by, ''),
			a.granted_at,
			u.id,
			u.username,
			u.password,
			u.is_forbidden,
			u.created_by,
			u.created_at,
			u.updated_at
		FROM end_user_project_access a
		JOIN end_user_users u
		  ON u.org_name = a.org_name
		 AND u.id = a.end_user_id
		LEFT JOIN end_user_permission_bundles b
		  ON b.id = a.permission_bundle_id
		 AND b.org_name = a.org_name
		 AND b.project_slug = a.project_slug
		WHERE a.id = ?
		  AND a.org_name = ?
		  AND a.project_slug = ?
		LIMIT 1
	`

	if orgName == "" {
		orgName = r.orgName
	}
	if projectSlug == "" {
		projectSlug = r.projectSlug
	}

	row := r.db.QueryRowContext(ctx, query, accessID, orgName, projectSlug)
	return scanEndUserProjectAccess(row)
}

func (r *SqlEndUserProjectAccessRepository) UpdatePermissionBundle(
	ctx context.Context,
	orgName, projectSlug, accessID, permissionBundleID string,
) error {
	const query = `
		UPDATE end_user_project_access
		SET permission_bundle_id = ?
		WHERE id = ?
		  AND org_name = ?
		  AND project_slug = ?
	`

	if orgName == "" {
		orgName = r.orgName
	}
	if projectSlug == "" {
		projectSlug = r.projectSlug
	}

	result, err := r.db.ExecContext(ctx, query, permissionBundleID, accessID, orgName, projectSlug)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "project access not found: "+accessID)
	}

	return nil
}

func (r *SqlEndUserProjectAccessRepository) Revoke(
	ctx context.Context,
	orgName, projectSlug, accessID string,
) error {
	const query = `
		DELETE FROM end_user_project_access
		WHERE id = ?
		  AND org_name = ?
		  AND project_slug = ?
	`

	if orgName == "" {
		orgName = r.orgName
	}
	if projectSlug == "" {
		projectSlug = r.projectSlug
	}

	result, err := r.db.ExecContext(ctx, query, accessID, orgName, projectSlug)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "project access not found: "+accessID)
	}

	return nil
}

func (r *SqlEndUserProjectAccessRepository) RemoveByEndUserID(
	ctx context.Context,
	orgName, endUserID string,
) error {
	const query = `
		DELETE FROM end_user_project_access
		WHERE end_user_id = ?
		  AND org_name = ?
	`

	if orgName == "" {
		orgName = r.orgName
	}

	_, err := r.db.ExecContext(ctx, query, endUserID, orgName)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	return nil
}

func (r *SqlEndUserProjectAccessRepository) PermissionBundleExists(
	ctx context.Context,
	orgName, projectSlug, permissionBundleID string,
) (bool, error) {
	const query = `
		SELECT COUNT(1)
		FROM end_user_permission_bundles
		WHERE id = ?
		  AND org_name = ?
		  AND project_slug = ?
	`

	if orgName == "" {
		orgName = r.orgName
	}
	if projectSlug == "" {
		projectSlug = r.projectSlug
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, query, permissionBundleID, orgName, projectSlug).Scan(&count); err != nil {
		return false, sqlerr.WrapSQLError(err)
	}
	return count > 0, nil
}

func (r *SqlEndUserProjectAccessRepository) ListWithTotal(
	ctx context.Context,
	query enduser.ListEndUserProjectAccessQuery,
) ([]*enduser.EndUserProjectAccess, int64, error) {
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
	projectSlug := query.ProjectSlug
	if projectSlug == "" {
		projectSlug = r.projectSlug
	}

	countSQL := `
		SELECT COUNT(1)
		FROM end_user_project_access a
		JOIN end_user_users u
		  ON u.org_name = a.org_name
		 AND u.id = a.end_user_id
		WHERE a.org_name = ?
		  AND a.project_slug = ?
	`
	countArgs := make([]any, 0, 3)
	countArgs = append(countArgs, orgName, projectSlug)
	if strings.TrimSpace(query.Search) != "" {
		countSQL += ` AND u.username LIKE CONCAT('%', ?, '%')`
		countArgs = append(countArgs, query.Search)
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}

	listSQL := `
		SELECT
			a.id,
			a.org_name,
			a.project_slug,
			a.end_user_id,
			a.permission_bundle_id,
			COALESCE(b.name, ''),
			COALESCE(a.granted_by, ''),
			a.granted_at,
			u.id,
			u.username,
			u.password,
			u.is_forbidden,
			u.created_by,
			u.created_at,
			u.updated_at
		FROM end_user_project_access a
		JOIN end_user_users u
		  ON u.org_name = a.org_name
		 AND u.id = a.end_user_id
		LEFT JOIN end_user_permission_bundles b
		  ON b.id = a.permission_bundle_id
		 AND b.org_name = a.org_name
		 AND b.project_slug = a.project_slug
		WHERE a.org_name = ?
		  AND a.project_slug = ?
		  AND (? = '' OR u.username LIKE CONCAT('%', ?, '%'))
		  AND (? = '' OR a.id > ?)
		ORDER BY a.id ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(
		ctx,
		listSQL,
		orgName,
		projectSlug,
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

	items := make([]*enduser.EndUserProjectAccess, 0, first)
	for rows.Next() {
		item, scanErr := scanEndUserProjectAccessRows(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, sqlerr.WrapSQLError(err)
	}

	return items, total, nil
}

func scanEndUserProjectAccessRows(rows *sql.Rows) (*enduser.EndUserProjectAccess, error) {
	var (
		accessID             string
		orgName              string
		projectSlug          string
		endUserID            string
		permissionBundleID   string
		permissionBundleName string
		grantedBy            string
		grantedAt            time.Time
		userID               string
		username             string
		passwordHash         string
		isForbidden          int
		userCreatedBy        sql.NullString
		userCreatedAt        time.Time
		userUpdatedAt        time.Time
	)

	err := rows.Scan(
		&accessID,
		&orgName,
		&projectSlug,
		&endUserID,
		&permissionBundleID,
		&permissionBundleName,
		&grantedBy,
		&grantedAt,
		&userID,
		&username,
		&passwordHash,
		&isForbidden,
		&userCreatedBy,
		&userCreatedAt,
		&userUpdatedAt,
	)
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}

	createdBy := sqlerr.NullStrToPtr(userCreatedBy)

	return &enduser.EndUserProjectAccess{
		ID:                 accessID,
		OrgName:            orgName,
		ProjectSlug:        projectSlug,
		EndUserID:          endUserID,
		PermissionBundleID: permissionBundleID,
		PermissionName:     permissionBundleName,
		GrantedBy:          grantedBy,
		GrantedAt:          grantedAt,
		EndUser: &enduser.EndUser{
			ID:          userID,
			OrgName:     orgName,
			Username:    username,
			Password:    enduser.NewHashedPasswordFromHash(passwordHash),
			IsForbidden: isForbidden == 1,
			CreatedBy:   ptrToString(createdBy),
			CreatedAt:   userCreatedAt,
			UpdatedAt:   userUpdatedAt,
		},
	}, nil
}

func scanEndUserProjectAccess(row *sql.Row) (*enduser.EndUserProjectAccess, error) {
	var (
		accessID             string
		orgName              string
		projectSlug          string
		endUserID            string
		permissionBundleID   string
		permissionBundleName string
		grantedBy            string
		grantedAt            time.Time
		userID               string
		username             string
		passwordHash         string
		isForbidden          int
		userCreatedBy        sql.NullString
		userCreatedAt        time.Time
		userUpdatedAt        time.Time
	)

	err := row.Scan(
		&accessID,
		&orgName,
		&projectSlug,
		&endUserID,
		&permissionBundleID,
		&permissionBundleName,
		&grantedBy,
		&grantedAt,
		&userID,
		&username,
		&passwordHash,
		&isForbidden,
		&userCreatedBy,
		&userCreatedAt,
		&userUpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //nolint:nilnil // not found is expected in repository layer
		}
		return nil, sqlerr.WrapSQLError(err)
	}

	createdBy := sqlerr.NullStrToPtr(userCreatedBy)

	return &enduser.EndUserProjectAccess{
		ID:                 accessID,
		OrgName:            orgName,
		ProjectSlug:        projectSlug,
		EndUserID:          endUserID,
		PermissionBundleID: permissionBundleID,
		PermissionName:     permissionBundleName,
		GrantedBy:          grantedBy,
		GrantedAt:          grantedAt,
		EndUser: &enduser.EndUser{
			ID:          userID,
			OrgName:     orgName,
			Username:    username,
			Password:    enduser.NewHashedPasswordFromHash(passwordHash),
			IsForbidden: isForbidden == 1,
			CreatedBy:   ptrToString(createdBy),
			CreatedAt:   userCreatedAt,
			UpdatedAt:   userUpdatedAt,
		},
	}, nil
}

func nullableString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *SqlEndUserProjectAccessRepository) ListAccessibleProjectsByUserID(
	ctx context.Context,
	orgName, endUserID string,
) ([]enduser.AccessibleProject, error) {
	const query = `
		SELECT a.project_slug, COALESCE(p.title, a.project_slug) AS project_title
		FROM end_user_project_access a
		LEFT JOIN projects p
		  ON p.org_name = a.org_name
		 AND p.slug = a.project_slug
		WHERE a.org_name = ?
		  AND a.end_user_id = ?
		ORDER BY a.project_slug ASC
	`

	if orgName == "" {
		orgName = r.orgName
	}

	rows, err := r.db.QueryContext(ctx, query, orgName, endUserID)
	if err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}
	defer rows.Close()

	projects := make([]enduser.AccessibleProject, 0)
	for rows.Next() {
		var item enduser.AccessibleProject
		if scanErr := rows.Scan(&item.ProjectSlug, &item.ProjectTitle); scanErr != nil {
			return nil, sqlerr.WrapSQLError(scanErr)
		}
		projects = append(projects, item)
	}
	if err = rows.Err(); err != nil {
		return nil, sqlerr.WrapSQLError(err)
	}

	return projects, nil
}

func (r *SqlEndUserProjectAccessRepository) HasProjectAccess(
	ctx context.Context,
	orgName, endUserID, projectSlug string,
) (bool, error) {
	const query = `
		SELECT COUNT(1)
		FROM end_user_project_access
		WHERE org_name = ?
		  AND end_user_id = ?
		  AND project_slug = ?
	`

	if orgName == "" {
		orgName = r.orgName
	}
	if projectSlug == "" {
		projectSlug = r.projectSlug
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, query, orgName, endUserID, projectSlug).Scan(&count); err != nil {
		return false, sqlerr.WrapSQLError(err)
	}

	return count > 0, nil
}

// Compile-time interface satisfaction check.
var _ enduser.EndUserProjectAccessRepository = (*SqlEndUserProjectAccessRepository)(nil)
