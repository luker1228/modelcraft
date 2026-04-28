package repository

import (
	"context"
	"encoding/json"
	"modelcraft/internal/domain/rbac"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"

	"github.com/google/uuid"
)

// SqlEndUserDataPermissionRepository is the sqlc-based implementation of
// rbac.EndUserPermissionRepository.
type SqlEndUserDataPermissionRepository struct {
	q dbgen.Querier
}

// NewSqlEndUserDataPermissionRepository creates a new SqlEndUserDataPermissionRepository
// backed by the given sqlc Querier.
func NewSqlEndUserDataPermissionRepository(q dbgen.Querier) rbac.EndUserPermissionRepository {
	return &SqlEndUserDataPermissionRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// compile-time interface assertion.
var _ rbac.EndUserPermissionRepository = (*SqlEndUserDataPermissionRepository)(nil)

// =========================
// Helpers
// =========================

func toDomainPermission(row dbgen.EndUserDataPermission) *rbac.EndUserPermission {
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

	var rowPolicy *rbac.RowPolicy
	if row.RowPolicy != nil {
		var rp rbac.RowPolicy
		if err := json.Unmarshal(*row.RowPolicy, &rp); err == nil {
			rp.Normalize()
			rowPolicy = &rp
		}
	}

	var preset *rbac.PermissionPreset
	if row.Preset.Valid {
		p := rbac.PermissionPreset(row.Preset.EndUserDataPermissionsPreset)
		preset = &p
	}

	var databaseName *string
	if row.DatabaseName.Valid {
		d := row.DatabaseName.String
		databaseName = &d
	}
	var modelName *string
	if row.ModelName.Valid {
		d := row.ModelName.String
		modelName = &d
	}

	return &rbac.EndUserPermission{
		OrgName:      row.OrgName,
		ProjectSlug:  row.ProjectSlug,
		ID:           row.ID,
		ModelID:      row.ModelID,
		DatabaseName: databaseName,
		ModelName:    modelName,
		Name:         row.Name,
		Description:  description,
		Type:         rbac.PermissionType(row.Type),
		ColumnPolicy: columnPolicy,
		RowPolicy:    rowPolicy,
		Preset:       preset,
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

func toDBRowPolicy(rp *rbac.RowPolicy) *json.RawMessage {
	if rp == nil {
		return nil
	}
	rp.Normalize()
	b, err := json.Marshal(rp)
	if err != nil {
		return nil
	}
	var raw json.RawMessage = b
	return &raw
}

func toDBPreset(preset *rbac.PermissionPreset) dbgen.NullEndUserDataPermissionsPreset {
	if preset == nil {
		return dbgen.NullEndUserDataPermissionsPreset{Valid: false}
	}
	return dbgen.NullEndUserDataPermissionsPreset{
		EndUserDataPermissionsPreset: dbgen.EndUserDataPermissionsPreset(*preset),
		Valid:                        true,
	}
}

// =========================
// Permissions
// =========================

func (r *SqlEndUserDataPermissionRepository) CreatePermission(ctx context.Context, p *rbac.EndUserPermission) error {
	params := dbgen.CreateEndUserPermissionParams{
		ID:           p.ID,
		OrgName:      p.OrgName,
		ProjectSlug:  p.ProjectSlug,
		DatabaseName: sqlerr.PtrToNullStr(p.DatabaseName),
		ModelName:    sqlerr.PtrToNullStr(p.ModelName),
		ModelID:      p.ModelID,
		Name:         p.Name,
		Description:  sqlerr.PtrToNullStr(p.Description),
		Type:         dbgen.EndUserDataPermissionsType(p.Type),
		ColumnPolicy: toDBColumnPolicy(p.ColumnPolicy),
		RowPolicy:    toDBRowPolicy(p.RowPolicy),
		Preset:       toDBPreset(p.Preset),
	}

	return r.q.CreateEndUserPermission(ctx, params)
}

func (r *SqlEndUserDataPermissionRepository) GetPermissionByID(
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

func (r *SqlEndUserDataPermissionRepository) ListPermissionsByProject(
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

func (r *SqlEndUserDataPermissionRepository) ListPermissionsByModel(
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

func (r *SqlEndUserDataPermissionRepository) ListPresetPermissionsByModel(
	ctx context.Context,
	orgName, modelID string,
) ([]*rbac.EndUserPermission, error) {
	rows, err := r.q.ListPresetPermissionsByModel(ctx, dbgen.ListPresetPermissionsByModelParams{
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

func (r *SqlEndUserDataPermissionRepository) GetPermissionByModelTypeName(
	ctx context.Context,
	orgName, modelID string,
	permissionType rbac.PermissionType,
	name string,
) (*rbac.EndUserPermission, error) {
	row, err := r.q.GetEndUserPermissionByModelTypeName(ctx, dbgen.GetEndUserPermissionByModelTypeNameParams{
		ModelID: modelID,
		OrgName: orgName,
		Type:    dbgen.EndUserDataPermissionsType(permissionType),
		Name:    name,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("end user permission not found")
		}
		return nil, sqlerr.WrapSQLError(err)
	}
	return toDomainPermission(row), nil
}

func (r *SqlEndUserDataPermissionRepository) UpdatePermission(ctx context.Context, p *rbac.EndUserPermission) error {
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

func (r *SqlEndUserDataPermissionRepository) DeletePermission(ctx context.Context, orgName, id string) error {
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

func (r *SqlEndUserDataPermissionRepository) UpdatePresetPermission(
	ctx context.Context,
	p *rbac.EndUserPermission,
) error {
	params := dbgen.UpdateEndUserPresetPermissionParams{
		Name:        p.Name,
		Description: sqlerr.PtrToNullStr(p.Description),
		RowPolicy:   toDBRowPolicy(p.RowPolicy),
		Preset:      toDBPreset(p.Preset),
		ID:          p.ID,
		OrgName:     p.OrgName,
	}

	result, err := r.q.UpdateEndUserPresetPermission(ctx, params)
	if err != nil {
		return sqlerr.WrapSQLError(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "preset permission not found: "+p.ID)
	}
	return nil
}

func (r *SqlEndUserDataPermissionRepository) IsPermissionReferencedByBundle(
	ctx context.Context,
	permissionID string,
) (bool, error) {
	referenced, err := r.q.IsPermissionReferencedByBundle(ctx, permissionID)
	if err != nil {
		return false, sqlerr.WrapSQLError(err)
	}
	return referenced, nil
}

func (r *SqlEndUserDataPermissionRepository) DeletePresetPermissionsByModel(
	ctx context.Context,
	orgName, modelID string,
) error {
	_, err := r.q.DeleteEndUserPermissionsByModelAndType(ctx, dbgen.DeleteEndUserPermissionsByModelAndTypeParams{
		ModelID: modelID,
		OrgName: orgName,
		Type:    dbgen.EndUserDataPermissionsTypePRESET,
	})
	if err != nil {
		return sqlerr.WrapSQLError(err)
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

func (r *SqlEndUserDataPermissionRepository) CreateBundle(ctx context.Context, b *rbac.EndUserPermissionBundle) error {
	params := dbgen.CreateEndUserBundleParams{
		ID:          b.ID,
		OrgName:     b.OrgName,
		ProjectSlug: b.ProjectSlug,
		Name:        b.Name,
		Description: sqlerr.PtrToNullStr(b.Description),
	}

	return sqlerr.WrapSQLError(r.q.CreateEndUserBundle(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) GetBundleByID(
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

func (r *SqlEndUserDataPermissionRepository) ListBundlesByProject(
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

func (r *SqlEndUserDataPermissionRepository) UpdateBundle(ctx context.Context, b *rbac.EndUserPermissionBundle) error {
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

func (r *SqlEndUserDataPermissionRepository) DeleteBundle(ctx context.Context, orgName, id string) error {
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

func (r *SqlEndUserDataPermissionRepository) AddPermissionToBundle(
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

func (r *SqlEndUserDataPermissionRepository) RemovePermissionFromBundle(
	ctx context.Context,
	bundleID, permissionID string,
) error {
	_, err := r.q.RemovePermissionFromBundle(ctx, dbgen.RemovePermissionFromBundleParams{
		BundleID:     bundleID,
		PermissionID: permissionID,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserDataPermissionRepository) ListPermissionsInBundle(
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
		ID:          row.ID,
		Name:        row.Name,
		Description: description,
		IsImplicit:  row.IsImplicit,
	}
}

func (r *SqlEndUserDataPermissionRepository) CreateRole(ctx context.Context, role *rbac.EndUserRole) error {
	params := dbgen.CreateEndUserRoleParams{
		ID:          role.ID,
		OrgName:     role.OrgName,
		Name:        role.Name,
		Description: sqlerr.PtrToNullStr(role.Description),
		IsImplicit:  role.IsImplicit,
	}

	return sqlerr.WrapSQLError(r.q.CreateEndUserRole(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) GetRoleByID(
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

func (r *SqlEndUserDataPermissionRepository) ListRolesByProject(
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

func (r *SqlEndUserDataPermissionRepository) UpdateRole(ctx context.Context, role *rbac.EndUserRole) error {
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

func (r *SqlEndUserDataPermissionRepository) DeleteRole(ctx context.Context, orgName, id string) error {
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

func (r *SqlEndUserDataPermissionRepository) AssignBundleToRole(
	ctx context.Context,
	orgName, projectSlug, roleID, bundleID string,
) error {
	params := dbgen.AssignBundleToRoleParams{
		ID:          uuid.NewString(),
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		RoleID:      roleID,
		BundleID:    bundleID,
	}

	return sqlerr.WrapSQLError(r.q.AssignBundleToRole(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) RevokeBundleFromRole(
	ctx context.Context,
	roleID, bundleID string,
) error {
	_, err := r.q.RevokeBundleFromRole(ctx, dbgen.RevokeBundleFromRoleParams{
		RoleID:   roleID,
		BundleID: bundleID,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserDataPermissionRepository) ListBundlesByRole(
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

func (r *SqlEndUserDataPermissionRepository) GrantBundleToUser(
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

func (r *SqlEndUserDataPermissionRepository) RevokeBundleFromUser(
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

func (r *SqlEndUserDataPermissionRepository) AssignRoleToUser(
	ctx context.Context,
	userID, orgName, projectSlug, roleID string,
) error {
	params := dbgen.AssignRoleToUserParams{
		ID:      uuid.NewString(),
		UserID:  userID,
		RoleID:  roleID,
		OrgName: orgName,
	}

	return sqlerr.WrapSQLError(r.q.AssignRoleToUser(ctx, params))
}

func (r *SqlEndUserDataPermissionRepository) RevokeRoleFromUser(
	ctx context.Context,
	userID, orgName, projectSlug, roleID string,
) error {
	_, err := r.q.RevokeRoleFromUser(ctx, dbgen.RevokeRoleFromUserParams{
		UserID:  userID,
		RoleID:  roleID,
		OrgName: orgName,
	})
	return sqlerr.WrapSQLError(err)
}

func (r *SqlEndUserDataPermissionRepository) GetBundleIDsByUserDirect(
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
	ids = append(ids, rows...)
	return ids, nil
}

func (r *SqlEndUserDataPermissionRepository) GetBundleIDsByUserExplicitRoles(
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
	ids = append(ids, rows...)
	return ids, nil
}

func (r *SqlEndUserDataPermissionRepository) GetBundleIDsByImplicitRoles(
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
	ids = append(ids, rows...)
	return ids, nil
}

func (r *SqlEndUserDataPermissionRepository) GetPermissionsByBundleIDs(
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
