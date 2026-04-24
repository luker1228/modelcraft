package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"modelcraft/internal/domain/rbac"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
)

// SqlEndUserPermissionRepository is the sqlc-based implementation of
// rbac.EndUserPermissionRepository.
type SqlEndUserPermissionRepository struct {
	q dbgen.Querier
}

// NewSqlEndUserPermissionRepository creates a new SqlEndUserPermissionRepository
// backed by the given sqlc Querier.
func NewSqlEndUserPermissionRepository(q dbgen.Querier) rbac.EndUserPermissionRepository {
	return &SqlEndUserPermissionRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// compile-time interface assertion.
var _ rbac.EndUserPermissionRepository = (*SqlEndUserPermissionRepository)(nil)

// =========================
// Helpers
// =========================

func toDomainPermission(row dbgen.EndUserPermission) *rbac.EndUserPermission {
	var description *string
	if row.Description.Valid {
		d := row.Description.String
		description = &d
	}

	var columnPolicy *rbac.ColumnPolicy
	if row.ColumnPolicy != nil {
		var cp rbac.ColumnPolicy
		if err := json.Unmarshal(*row.ColumnPolicy, &cp); err == nil {
			columnPolicy = &cp
		}
	}

	return &rbac.EndUserPermission{
		OrgName:      row.OrgName,
		ProjectSlug:  row.ProjectSlug,
		ID:           row.ID,
		ModelID:      row.ModelID,
		Name:         row.Name,
		Description:  description,
		Action:       rbac.Action(row.Action),
		ColumnPolicy: columnPolicy,
		RowScope:     rbac.RowScope(row.RowScope),
	}
}

func toDBColumnPolicy(cp *rbac.ColumnPolicy) *json.RawMessage {
	if cp == nil {
		return nil
	}
	b, err := json.Marshal(cp)
	if err != nil {
		return nil
	}
	var raw json.RawMessage = b
	return &raw
}

// =========================
// Permissions
// =========================

func (r *SqlEndUserPermissionRepository) CreatePermission(ctx context.Context, p *rbac.EndUserPermission) error {
	params := dbgen.CreateEndUserPermissionParams{
		ID:           p.ID,
		OrgName:      p.OrgName,
		ProjectSlug:  p.ProjectSlug,
		ModelID:      p.ModelID,
		Name:         p.Name,
		Description:  sqlerr.PtrToNullStr(p.Description),
		Action:       dbgen.EndUserPermissionsAction(p.Action),
		ColumnPolicy: toDBColumnPolicy(p.ColumnPolicy),
		RowScope:     dbgen.EndUserPermissionsRowScope(p.RowScope),
	}

	return sqlerr.WrapSQLError(r.q.CreateEndUserPermission(ctx, params))
}

func (r *SqlEndUserPermissionRepository) GetPermissionByID(
	ctx context.Context,
	orgName, id string,
) (*rbac.EndUserPermission, error) {
	row, err := r.q.GetEndUserPermissionByID(ctx, dbgen.GetEndUserPermissionByIDParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user permission not found: " + id)
		}
		return nil, err
	}

	return toDomainPermission(row), nil
}

func (r *SqlEndUserPermissionRepository) ListPermissionsByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbac.EndUserPermission, error) {
	rows, err := r.q.ListEndUserPermissionsByProject(ctx, dbgen.ListEndUserPermissionsByProjectParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	perms := make([]*rbac.EndUserPermission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, toDomainPermission(row))
	}
	return perms, nil
}

func (r *SqlEndUserPermissionRepository) ListPermissionsByModel(
	ctx context.Context,
	orgName, modelID string,
) ([]*rbac.EndUserPermission, error) {
	rows, err := r.q.ListEndUserPermissionsByModel(ctx, dbgen.ListEndUserPermissionsByModelParams{
		ModelID: modelID,
		OrgName: orgName,
	})
	if err != nil {
		return nil, err
	}

	perms := make([]*rbac.EndUserPermission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, toDomainPermission(row))
	}
	return perms, nil
}

func (r *SqlEndUserPermissionRepository) UpdatePermission(ctx context.Context, p *rbac.EndUserPermission) error {
	params := dbgen.UpdateEndUserPermissionParams{
		Name:         p.Name,
		Description:  sqlerr.PtrToNullStr(p.Description),
		ColumnPolicy: toDBColumnPolicy(p.ColumnPolicy),
		ID:           p.ID,
		OrgName:      p.OrgName,
	}

	result, err := r.q.UpdateEndUserPermission(ctx, params)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user permission not found: "+p.ID)
	}
	return nil
}

func (r *SqlEndUserPermissionRepository) DeletePermission(ctx context.Context, orgName, id string) error {
	result, err := r.q.DeleteEndUserPermission(ctx, dbgen.DeleteEndUserPermissionParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user permission not found: "+id)
	}
	return nil
}

// =========================
// Bundles
// =========================

func toDomainBundle(row dbgen.EndUserPermissionBundle) *rbac.EndUserPermissionBundle {
	var description *string
	if row.Description.Valid {
		d := row.Description.String
		description = &d
	}

	return &rbac.EndUserPermissionBundle{
		OrgName:     row.OrgName,
		ProjectSlug: row.ProjectSlug,
		ID:          row.ID,
		Name:        row.Name,
		Description: description,
	}
}

func (r *SqlEndUserPermissionRepository) CreateBundle(ctx context.Context, b *rbac.EndUserPermissionBundle) error {
	params := dbgen.CreateEndUserBundleParams{
		ID:          b.ID,
		OrgName:     b.OrgName,
		ProjectSlug: b.ProjectSlug,
		Name:        b.Name,
		Description: sqlerr.PtrToNullStr(b.Description),
	}

	return sqlerr.WrapSQLError(r.q.CreateEndUserBundle(ctx, params))
}

func (r *SqlEndUserPermissionRepository) GetBundleByID(
	ctx context.Context,
	orgName, id string,
) (*rbac.EndUserPermissionBundle, error) {
	row, err := r.q.GetEndUserBundleByID(ctx, dbgen.GetEndUserBundleByIDParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user bundle not found: " + id)
		}
		return nil, err
	}

	return toDomainBundle(row), nil
}

func (r *SqlEndUserPermissionRepository) ListBundlesByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbac.EndUserPermissionBundle, error) {
	rows, err := r.q.ListEndUserBundlesByProject(ctx, dbgen.ListEndUserBundlesByProjectParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	bundles := make([]*rbac.EndUserPermissionBundle, 0, len(rows))
	for _, row := range rows {
		bundles = append(bundles, toDomainBundle(row))
	}
	return bundles, nil
}

func (r *SqlEndUserPermissionRepository) UpdateBundle(ctx context.Context, b *rbac.EndUserPermissionBundle) error {
	params := dbgen.UpdateEndUserBundleParams{
		Name:        b.Name,
		Description: sqlerr.PtrToNullStr(b.Description),
		ID:          b.ID,
		OrgName:     b.OrgName,
	}

	result, err := r.q.UpdateEndUserBundle(ctx, params)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user bundle not found: "+b.ID)
	}
	return nil
}

func (r *SqlEndUserPermissionRepository) DeleteBundle(ctx context.Context, orgName, id string) error {
	result, err := r.q.DeleteEndUserBundle(ctx, dbgen.DeleteEndUserBundleParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user bundle not found: "+id)
	}
	return nil
}

func (r *SqlEndUserPermissionRepository) AddPermissionToBundle(
	ctx context.Context,
	bundleID, permissionID string,
	sortOrder int,
) error {
	params := dbgen.AddPermissionToBundleParams{
		ID:           uuid.NewString(),
		BundleID:     bundleID,
		PermissionID: permissionID,
		SortOrder:    int32(sortOrder),
	}

	return sqlerr.WrapSQLError(r.q.AddPermissionToBundle(ctx, params))
}

func (r *SqlEndUserPermissionRepository) RemovePermissionFromBundle(
	ctx context.Context,
	bundleID, permissionID string,
) error {
	_, err := r.q.RemovePermissionFromBundle(ctx, dbgen.RemovePermissionFromBundleParams{
		BundleID:     bundleID,
		PermissionID: permissionID,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserPermissionRepository) ListPermissionsInBundle(
	ctx context.Context,
	bundleID string,
) ([]*rbac.EndUserPermission, error) {
	rows, err := r.q.ListPermissionsInBundle(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	perms := make([]*rbac.EndUserPermission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, toDomainPermission(row))
	}
	return perms, nil
}

// =========================
// Roles
// =========================

func toDomainRole(row dbgen.EndUserRole) *rbac.EndUserRole {
	var description *string
	if row.Description.Valid {
		d := row.Description.String
		description = &d
	}

	return &rbac.EndUserRole{
		OrgName:     row.OrgName,
		ProjectSlug: row.ProjectSlug,
		ID:          row.ID,
		Name:        row.Name,
		Description: description,
		IsImplicit:  row.IsImplicit,
	}
}

func (r *SqlEndUserPermissionRepository) CreateRole(ctx context.Context, role *rbac.EndUserRole) error {
	params := dbgen.CreateEndUserRoleParams{
		ID:          role.ID,
		OrgName:     role.OrgName,
		ProjectSlug: role.ProjectSlug,
		Name:        role.Name,
		Description: sqlerr.PtrToNullStr(role.Description),
		IsImplicit:  role.IsImplicit,
	}

	return sqlerr.WrapSQLError(r.q.CreateEndUserRole(ctx, params))
}

func (r *SqlEndUserPermissionRepository) GetRoleByID(
	ctx context.Context,
	orgName, id string,
) (*rbac.EndUserRole, error) {
	row, err := r.q.GetEndUserRoleByID(ctx, dbgen.GetEndUserRoleByIDParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user role not found: " + id)
		}
		return nil, err
	}

	return toDomainRole(row), nil
}

func (r *SqlEndUserPermissionRepository) ListRolesByProject(
	ctx context.Context,
	orgName, projectSlug string,
) ([]*rbac.EndUserRole, error) {
	rows, err := r.q.ListEndUserRolesByProject(ctx, dbgen.ListEndUserRolesByProjectParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	roles := make([]*rbac.EndUserRole, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, toDomainRole(row))
	}
	return roles, nil
}

func (r *SqlEndUserPermissionRepository) UpdateRole(ctx context.Context, role *rbac.EndUserRole) error {
	params := dbgen.UpdateEndUserRoleParams{
		Name:        role.Name,
		Description: sqlerr.PtrToNullStr(role.Description),
		ID:          role.ID,
		OrgName:     role.OrgName,
	}

	result, err := r.q.UpdateEndUserRole(ctx, params)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user role not found: "+role.ID)
	}
	return nil
}

func (r *SqlEndUserPermissionRepository) DeleteRole(ctx context.Context, orgName, id string) error {
	result, err := r.q.DeleteEndUserRole(ctx, dbgen.DeleteEndUserRoleParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "end user role not found: "+id)
	}
	return nil
}

func (r *SqlEndUserPermissionRepository) AssignBundleToRole(
	ctx context.Context,
	roleID, bundleID string,
) error {
	params := dbgen.AssignBundleToRoleParams{
		ID:       uuid.NewString(),
		RoleID:   roleID,
		BundleID: bundleID,
	}

	return sqlerr.WrapSQLError(r.q.AssignBundleToRole(ctx, params))
}

func (r *SqlEndUserPermissionRepository) RevokeBundleFromRole(
	ctx context.Context,
	roleID, bundleID string,
) error {
	_, err := r.q.RevokeBundleFromRole(ctx, dbgen.RevokeBundleFromRoleParams{
		RoleID:   roleID,
		BundleID: bundleID,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserPermissionRepository) ListBundlesByRole(
	ctx context.Context,
	roleID string,
) ([]*rbac.EndUserPermissionBundle, error) {
	rows, err := r.q.ListBundlesByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	bundles := make([]*rbac.EndUserPermissionBundle, 0, len(rows))
	for _, row := range rows {
		bundles = append(bundles, toDomainBundle(row))
	}
	return bundles, nil
}

// =========================
// User Grants & Authz Chain
// =========================

func (r *SqlEndUserPermissionRepository) GrantBundleToUser(
	ctx context.Context,
	userID, orgName, projectSlug, bundleID string,
) error {
	params := dbgen.GrantBundleToUserParams{
		ID:          uuid.NewString(),
		UserID:      userID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		BundleID:    bundleID,
	}

	return sqlerr.WrapSQLError(r.q.GrantBundleToUser(ctx, params))
}

func (r *SqlEndUserPermissionRepository) RevokeBundleFromUser(
	ctx context.Context,
	userID, orgName, projectSlug, bundleID string,
) error {
	_, err := r.q.RevokeBundleFromUser(ctx, dbgen.RevokeBundleFromUserParams{
		UserID:      userID,
		BundleID:    bundleID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserPermissionRepository) AssignRoleToUser(
	ctx context.Context,
	userID, orgName, projectSlug, roleID string,
) error {
	params := dbgen.AssignRoleToUserParams{
		ID:          uuid.NewString(),
		UserID:      userID,
		RoleID:      roleID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	}

	return sqlerr.WrapSQLError(r.q.AssignRoleToUser(ctx, params))
}

func (r *SqlEndUserPermissionRepository) RevokeRoleFromUser(
	ctx context.Context,
	userID, orgName, projectSlug, roleID string,
) error {
	_, err := r.q.RevokeRoleFromUser(ctx, dbgen.RevokeRoleFromUserParams{
		UserID:      userID,
		RoleID:      roleID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserPermissionRepository) GetBundleIDsByUserDirect(
	ctx context.Context,
	userID, orgName, projectSlug string,
) ([]string, error) {
	rows, err := r.q.GetBundleIDsByUserDirect(ctx, dbgen.GetBundleIDsByUserDirectParams{
		UserID:      userID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row)
	}
	return ids, nil
}

func (r *SqlEndUserPermissionRepository) GetBundleIDsByUserExplicitRoles(
	ctx context.Context,
	userID, orgName, projectSlug string,
) ([]string, error) {
	rows, err := r.q.GetBundleIDsByUserExplicitRoles(ctx, dbgen.GetBundleIDsByUserExplicitRolesParams{
		UserID:      userID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row)
	}
	return ids, nil
}

func (r *SqlEndUserPermissionRepository) GetBundleIDsByImplicitRoles(
	ctx context.Context,
	orgName, projectSlug string,
) ([]string, error) {
	rows, err := r.q.GetBundleIDsByImplicitRoles(ctx, dbgen.GetBundleIDsByImplicitRolesParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row)
	}
	return ids, nil
}

func (r *SqlEndUserPermissionRepository) GetPermissionsByBundleIDs(
	ctx context.Context,
	orgName string,
	bundleIDs []string,
) ([]*rbac.EndUserPermission, error) {
	if len(bundleIDs) == 0 {
		return []*rbac.EndUserPermission{}, nil
	}

	rows, err := r.q.GetPermissionsByBundleIDs(ctx, dbgen.GetPermissionsByBundleIDsParams{
		Bundleids: bundleIDs,
		OrgName:   orgName,
	})
	if err != nil {
		return nil, err
	}

	perms := make([]*rbac.EndUserPermission, 0, len(rows))
	for _, row := range rows {
		perms = append(perms, toDomainPermission(row))
	}
	return perms, nil
}
