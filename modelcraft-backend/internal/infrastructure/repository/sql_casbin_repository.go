// Package repository provides sqlc-based implementations of domain repository interfaces.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/internal/domain/permission"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"

	bizerrors "modelcraft/pkg/bizerrors"
)

// CasbinRoleToDomain converts a dbgen.Role row to a domain permission.Role entity.
func CasbinRoleToDomain(row dbgen.Role) *permission.Role {
	var description string
	if row.Description.Valid {
		description = row.Description.String
	}

	return &permission.Role{
		ID:          int(row.ID),
		Name:        row.Name,
		Description: description,
		IsSystem:    row.IsSystem,
		OrgName:     row.OrgName,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// CasbinRolePermissionToDomain converts a dbgen.RolePermission row to a domain permission.Permission entity.
func CasbinRolePermissionToDomain(row dbgen.RolePermission) *permission.Permission {
	return &permission.Permission{
		Obj: row.Obj,
		Act: row.Act,
	}
}

// CasbinUserRoleToDomain converts a dbgen.UserRole row to a domain permission.UserRole entity.
func CasbinUserRoleToDomain(row dbgen.UserRole) *permission.UserRole {
	return &permission.UserRole{
		ID:        int(row.ID),
		UserID:    row.UserID,
		RoleID:    int(row.RoleID),
		OrgName:   row.OrgName,
		CreatedAt: row.CreatedAt,
	}
}

// SqlCasbinRoleRepository is the sqlc-based implementation of permission.RoleRepository.
type SqlCasbinRoleRepository struct {
	q dbgen.Querier
}

// NewSqlCasbinRoleRepository creates a new SqlCasbinRoleRepository backed by the given sqlc Querier.
func NewSqlCasbinRoleRepository(q dbgen.Querier) permission.RoleRepository {
	return &SqlCasbinRoleRepository{q: q}
}

// CreateRole persists a new role to the database and populates role.ID with the generated primary key.
func (r *SqlCasbinRoleRepository) CreateRole(ctx context.Context, role *permission.Role) error {
	params := dbgen.CreateRoleParams{
		Name:        role.Name,
		Description: sql.NullString{String: role.Description, Valid: role.Description != ""},
		IsSystem:    role.IsSystem,
		OrgName:     role.OrgName,
	}

	var result sql.Result

	if err := ExecWithErrorHandling(func() error {
		var e error
		result, e = r.q.CreateRole(ctx, params)
		return e
	}); err != nil {
		return bizerrors.Wrapf(err, "failed to create role: %s", role.Name)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return bizerrors.Wrapf(err, "failed to retrieve last insert id for role: %s", role.Name)
	}

	role.ID = int(id)

	return nil
}

// GetRoleByID retrieves a role by its integer ID.
// Returns nil, shared.NewNotFoundError when the role is not found.
func (r *SqlCasbinRoleRepository) GetRoleByID(ctx context.Context, id int) (*permission.Role, error) {
	var row dbgen.Role

	err := QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetRoleByID(ctx, int64(id))
		return e
	})
	if err != nil {
		if IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("role not found by id: " + fmt.Sprint(id))
		}
		return nil, bizerrors.Wrapf(err, "failed to get role by id: %d", id)
	}

	return CasbinRoleToDomain(row), nil
}

// GetRoleByNameAndOrg retrieves a role by its name within a specific organization.
// Returns nil, shared.NewNotFoundError when no matching role exists.
func (r *SqlCasbinRoleRepository) GetRoleByNameAndOrg(
	ctx context.Context, name, orgName string,
) (*permission.Role, error) {
	var row dbgen.Role

	err := QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetRoleByNameAndOrg(ctx, dbgen.GetRoleByNameAndOrgParams{
			Name:    name,
			OrgName: orgName,
		})
		return e
	})
	if err != nil {
		if IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("role not found: " + name + " in org " + orgName)
		}
		return nil, bizerrors.Wrapf(err, "failed to get role by name %s and org %s", name, orgName)
	}

	return CasbinRoleToDomain(row), nil
}

// ListRolesByOrg returns all roles for the given organization.
// When includeSystem is true, system roles (org_name='__SYSTEM__') are included alongside org-specific roles.
func (r *SqlCasbinRoleRepository) ListRolesByOrg(
	ctx context.Context, orgName string, includeSystem bool,
) ([]*permission.Role, error) {
	var rows []dbgen.Role

	if err := QueryWithSQLErrorHandling(func() error {
		var e error
		if includeSystem {
			rows, e = r.q.ListRolesByOrgIncludeSystem(ctx, orgName)
		} else {
			rows, e = r.q.ListRolesByOrg(ctx, orgName)
		}
		return e
	}); err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list roles by org: %s", orgName)
	}

	roles := make([]*permission.Role, len(rows))
	for i, row := range rows {
		roles[i] = CasbinRoleToDomain(row)
	}

	return roles, nil
}

// UpdateRole updates the description of an existing role identified by role.ID.
// Only the description field is persisted; name, is_system, and org_name are immutable via SQL.
func (r *SqlCasbinRoleRepository) UpdateRole(ctx context.Context, role *permission.Role) error {
	params := dbgen.UpdateRoleParams{
		Description: sql.NullString{String: role.Description, Valid: role.Description != ""},
		ID:          int64(role.ID),
	}

	return ExecWithErrorHandling(func() error {
		return r.q.UpdateRole(ctx, params)
	})
}

// DeleteRole deletes a role by its integer ID.
// Related user_roles and role_permissions are cascade-deleted by the database foreign key constraints.
func (r *SqlCasbinRoleRepository) DeleteRole(ctx context.Context, id int) error {
	return ExecWithErrorHandling(func() error {
		return r.q.DeleteRole(ctx, int64(id))
	})
}

// SqlCasbinPermissionRepository is the sqlc-based implementation of permission.PermissionRepository.
type SqlCasbinPermissionRepository struct {
	q dbgen.Querier
}

// NewSqlCasbinPermissionRepository creates a new SqlCasbinPermissionRepository backed by the given sqlc Querier.
func NewSqlCasbinPermissionRepository(q dbgen.Querier) permission.PermissionRepository {
	return &SqlCasbinPermissionRepository{q: q}
}

// AddPermission adds a permission entry for the given role within the specified organization.
func (r *SqlCasbinPermissionRepository) AddPermission(
	ctx context.Context, roleID int, orgName string, perm *permission.Permission,
) error {
	params := dbgen.CreatePermissionParams{
		RoleID:  int64(roleID),
		OrgName: orgName,
		Obj:     perm.Obj,
		Act:     perm.Act,
	}

	return ExecWithErrorHandling(func() error {
		return r.q.CreatePermission(ctx, params)
	})
}

// RemovePermission deletes the specific permission entry identified by roleID, obj, and act.
func (r *SqlCasbinPermissionRepository) RemovePermission(ctx context.Context, roleID int, obj, act string) error {
	params := dbgen.DeletePermissionParams{
		RoleID: int64(roleID),
		Obj:    obj,
		Act:    act,
	}

	return ExecWithErrorHandling(func() error {
		return r.q.DeletePermission(ctx, params)
	})
}

// ListPermissionsByRole returns all permissions assigned to the given role, across all organizations.
func (r *SqlCasbinPermissionRepository) ListPermissionsByRole(
	ctx context.Context, roleID int,
) ([]*permission.Permission, error) {
	var rows []dbgen.RolePermission

	if err := QueryWithSQLErrorHandling(func() error {
		var e error
		rows, e = r.q.ListPermissionsByRole(ctx, int64(roleID))
		return e
	}); err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list permissions by role: %d", roleID)
	}

	perms := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		perms[i] = CasbinRolePermissionToDomain(row)
	}

	return perms, nil
}

// ListPermissionsByRoleAndOrg returns all permissions assigned to the given role within a specific organization.
func (r *SqlCasbinPermissionRepository) ListPermissionsByRoleAndOrg(
	ctx context.Context, roleID int, orgName string,
) ([]*permission.Permission, error) {
	var rows []dbgen.RolePermission

	if err := QueryWithSQLErrorHandling(func() error {
		var e error
		rows, e = r.q.ListPermissionsByRoleAndOrg(ctx, dbgen.ListPermissionsByRoleAndOrgParams{
			RoleID:  int64(roleID),
			OrgName: orgName,
		})
		return e
	}); err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list permissions by role %d and org %s", roleID, orgName)
	}

	perms := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		perms[i] = CasbinRolePermissionToDomain(row)
	}

	return perms, nil
}

// DeletePermissionsByRole removes all permission entries for the given role.
// This is typically called as part of a role deletion workflow.
func (r *SqlCasbinPermissionRepository) DeletePermissionsByRole(ctx context.Context, roleID int) error {
	return ExecWithErrorHandling(func() error {
		return r.q.DeletePermissionsByRole(ctx, int64(roleID))
	})
}

// SqlCasbinUserRoleRepository is the sqlc-based implementation of permission.UserRoleRepository.
type SqlCasbinUserRoleRepository struct {
	q dbgen.Querier
}

// NewSqlCasbinUserRoleRepository creates a new SqlCasbinUserRoleRepository backed by the given sqlc Querier.
func NewSqlCasbinUserRoleRepository(q dbgen.Querier) permission.UserRoleRepository {
	return &SqlCasbinUserRoleRepository{q: q}
}

// AssignRole persists a new user-role binding and populates userRole.ID with the generated primary key.
func (r *SqlCasbinUserRoleRepository) AssignRole(ctx context.Context, userRole *permission.UserRole) error {
	params := dbgen.CreateUserRoleParams{
		UserID:  userRole.UserID,
		RoleID:  int64(userRole.RoleID),
		OrgName: userRole.OrgName,
	}

	var result sql.Result

	if err := ExecWithErrorHandling(func() error {
		var e error
		result, e = r.q.CreateUserRole(ctx, params)
		return e
	}); err != nil {
		return bizerrors.Wrapf(err,
			"failed to assign role %d to user %s in org %s", userRole.RoleID, userRole.UserID, userRole.OrgName,
		)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return bizerrors.Wrapf(err, "failed to retrieve last insert id for user_role assignment")
	}

	userRole.ID = int(id)

	return nil
}

// RevokeRole removes the user-role binding identified by userID, roleID, and orgName.
func (r *SqlCasbinUserRoleRepository) RevokeRole(ctx context.Context, userID string, roleID int, orgName string) error {
	params := dbgen.DeleteUserRoleParams{
		UserID:  userID,
		RoleID:  int64(roleID),
		OrgName: orgName,
	}

	return ExecWithErrorHandling(func() error {
		return r.q.DeleteUserRole(ctx, params)
	})
}

// ListUserRoles returns all role bindings for the given user within the specified organization.
func (r *SqlCasbinUserRoleRepository) ListUserRoles(
	ctx context.Context, userID, orgName string,
) ([]*permission.UserRole, error) {
	var rows []dbgen.UserRole

	if err := QueryWithSQLErrorHandling(func() error {
		var e error
		rows, e = r.q.ListUserRoles(ctx, dbgen.ListUserRolesParams{
			UserID:  userID,
			OrgName: orgName,
		})
		return e
	}); err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list user roles for user %s in org %s", userID, orgName)
	}

	userRoles := make([]*permission.UserRole, len(rows))
	for i, row := range rows {
		userRoles[i] = CasbinUserRoleToDomain(row)
	}

	return userRoles, nil
}

// ListRoleUsers returns all user-role bindings for the given role within the specified organization.
func (r *SqlCasbinUserRoleRepository) ListRoleUsers(
	ctx context.Context, roleID int, orgName string,
) ([]*permission.UserRole, error) {
	var rows []dbgen.UserRole

	if err := QueryWithSQLErrorHandling(func() error {
		var e error
		rows, e = r.q.ListRoleUsers(ctx, dbgen.ListRoleUsersParams{
			RoleID:  int64(roleID),
			OrgName: orgName,
		})
		return e
	}); err != nil {
		return nil, bizerrors.Wrapf(err, "failed to list users for role %d in org %s", roleID, orgName)
	}

	userRoles := make([]*permission.UserRole, len(rows))
	for i, row := range rows {
		userRoles[i] = CasbinUserRoleToDomain(row)
	}

	return userRoles, nil
}

// GetUserRole retrieves the specific user-role binding identified by userID, roleID, and orgName.
// Returns ErrRecordNotFound when no matching binding exists.
func (r *SqlCasbinUserRoleRepository) GetUserRole(
	ctx context.Context, userID string, roleID int, orgName string,
) (*permission.UserRole, error) {
	var row dbgen.UserRole

	err := QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetUserRole(ctx, dbgen.GetUserRoleParams{
			UserID:  userID,
			RoleID:  int64(roleID),
			OrgName: orgName,
		})
		return e
	})
	if err != nil {
		if IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("user role not found")
		}
		return nil, bizerrors.Wrapf(err,
			"failed to get user role for user %s, role %d, org %s", userID, roleID, orgName,
		)
	}

	return CasbinUserRoleToDomain(row), nil
}

// DeleteUserRolesByRole removes all user-role bindings for the given role.
// This is typically called as part of a role deletion workflow.
func (r *SqlCasbinUserRoleRepository) DeleteUserRolesByRole(ctx context.Context, roleID int) error {
	return ExecWithErrorHandling(func() error {
		return r.q.DeleteUserRolesByRole(ctx, int64(roleID))
	})
}

// Compile-time interface satisfaction checks.
var (
	_ permission.RoleRepository       = (*SqlCasbinRoleRepository)(nil)
	_ permission.PermissionRepository = (*SqlCasbinPermissionRepository)(nil)
	_ permission.UserRoleRepository   = (*SqlCasbinUserRoleRepository)(nil)
)
