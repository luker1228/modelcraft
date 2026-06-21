// Package permission provides application-layer services for permission management.
package permission

import (
	"context"
	"modelcraft/internal/domain/permission"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/auth"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// SystemRolePermissionsSyncer syncs system role permissions from the in-memory definition
// into the role_permissions table on every startup.
//
// The database snapshot is a read-only copy for display and audit purposes.
// Code remains the authoritative source; manual DB edits are overwritten on restart.
type SystemRolePermissionsSyncer struct {
	roleRepo permission.RoleRepository
	permRepo permission.PermissionRepository
}

// NewSystemRolePermissionsSyncer creates a new SystemRolePermissionsSyncer.
func NewSystemRolePermissionsSyncer(
	roleRepo permission.RoleRepository,
	permRepo permission.PermissionRepository,
) *SystemRolePermissionsSyncer {
	return &SystemRolePermissionsSyncer{
		roleRepo: roleRepo,
		permRepo: permRepo,
	}
}

// Sync performs a force-reset of system roles and their permissions in the database.
//
// For each system role it:
//  1. Fetches the role record from DB; creates it if it does not exist yet.
//  2. Deletes all existing role_permissions rows for that role_id.
//  3. Bulk-inserts the current permissions from the in-memory definition.
//
// The operation is idempotent. The first error encountered is returned and
// the caller is responsible for deciding whether to abort startup.
func (s *SystemRolePermissionsSyncer) Sync(ctx context.Context) error {
	logger := logfacade.GetLogger(ctx)

	for roleName, perms := range auth.SystemRolePermissions {
		// 1. Fetch the system role DB record; create it when it does not exist yet.
		role, err := s.roleRepo.GetRoleByNameAndOrg(ctx, roleName, permission.SystemOrgName)
		if err != nil {
			if !shared.IsNotFoundError(err) {
				logger.With(logfacade.Err(err)).Errorf(
					ctx, nil, "Failed to fetch system role during sync",
				)
				return bizerrors.Wrapf(err, "fetch system role: %s", roleName)
			}
			// Role does not exist yet; fall through to create it.
			role = nil
		}
		if role == nil {
			role = &permission.Role{
				Name:     roleName,
				IsSystem: true,
				OrgName:  permission.SystemOrgName,
			}
			if err := s.roleRepo.CreateRole(ctx, role); err != nil {
				logger.With(logfacade.Err(err)).Errorf(
					ctx, nil, "Failed to create system role during sync",
				)
				return bizerrors.Wrapf(err, "create system role: %s", roleName)
			}
			logger.Infof(ctx, "System role created: role=%s, id=%d", roleName, role.ID)
		}

		// 2. Delete all existing permissions for this role.
		if err := s.permRepo.DeletePermissionsByRole(ctx, role.ID); err != nil {
			logger.With(logfacade.Err(err)).Errorf(
				ctx, nil, "Failed to delete system role permissions during sync",
			)
			return bizerrors.Wrapf(err, "delete permissions for system role: %s", roleName)
		}

		// 3. Insert current permissions from in-memory definition.
		for _, perm := range perms {
			if err := s.permRepo.AddPermission(ctx, role.ID, permission.SystemOrgName, perm); err != nil {
				logger.With(logfacade.Err(err)).Errorf(
					ctx, nil, "Failed to insert system role permission during sync",
				)
				return bizerrors.Wrapf(
					err,
					"insert permission %s:%s for system role: %s",
					perm.Obj,
					perm.Act,
					roleName,
				)
			}
		}

		logger.Infof(
			ctx, "System role permissions synced: role=%s, count=%d",
			roleName,
			len(perms),
		)
	}

	return nil
}
