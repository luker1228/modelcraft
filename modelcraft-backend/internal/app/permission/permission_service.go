package permission

import (
	"context"
	"modelcraft/internal/domain/permission"
	"modelcraft/internal/infrastructure/auth"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"

	"github.com/casbin/casbin/v2"
)

// PermissionService provides business logic for permission management
type PermissionService struct {
	roleRepo permission.RoleRepository
	permRepo permission.PermissionRepository
	enforcer *casbin.Enforcer
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	roleRepo permission.RoleRepository,
	permRepo permission.PermissionRepository,
) *PermissionService {
	enforcer, _ := auth.GetEnforcer() // 忽略错误，使用全局 enforcer
	return &PermissionService{
		roleRepo: roleRepo,
		permRepo: permRepo,
		enforcer: enforcer,
	}
}

// AddPermissionToRole adds a permission to a custom role
func (s *PermissionService) AddPermissionToRole(ctx context.Context, roleID int, obj, act string) error {
	logger := logfacade.GetLogger(ctx)

	// Validate input
	perm := permission.NewPermission(obj, act)
	if err := perm.Validate(); err != nil {
		return err
	}

	// Get role
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.NotFound,
			"role not found: id=%d",
			roleID,
		)
	}

	// System role protection - cannot modify system role permissions
	if !role.CanModify() {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.OperationDenied,
			"cannot add permissions to system role: %s",
			role.Name,
		)
	}

	// Add to database
	if err := s.permRepo.AddPermission(ctx, roleID, role.OrgName, perm); err != nil {
		return err
	}

	// Add to Casbin enforcer (for custom roles)
	if s.enforcer != nil {
		_, err := s.enforcer.AddPolicy(role.Name, obj, act)
		if err != nil {
			logger.Errorf(
				ctx, err, "Failed to add Casbin policy for role %s",
				role.Name,
			)
			// Don't fail the operation if Casbin sync fails
		}
	}

	logger.Infof(
		ctx, "Added permission to role: role_id=%d, role_name=%s, "+
			"permission=%s:%s",
		roleID,
		role.Name,
		obj,
		act,
	)

	return nil
}

// RemovePermissionFromRole removes a permission from a custom role
func (s *PermissionService) RemovePermissionFromRole(ctx context.Context, roleID int, obj, act string) error {
	logger := logfacade.GetLogger(ctx)

	// Get role
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.NotFound,
			"role not found: id=%d",
			roleID,
		)
	}

	// System role protection
	if !role.CanModify() {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.OperationDenied,
			"cannot remove permissions from system role: %s",
			role.Name,
		)
	}

	// Remove from database
	if err := s.permRepo.RemovePermission(ctx, roleID, obj, act); err != nil {
		return err
	}

	// Remove from Casbin enforcer
	if s.enforcer != nil {
		_, err := s.enforcer.RemovePolicy(role.Name, obj, act)
		if err != nil {
			logger.Errorf(
				ctx, err, "Failed to remove Casbin policy for role %s",
				role.Name,
			)
			// Don't fail the operation if Casbin sync fails
		}
	}

	logger.Infof(
		ctx, "Removed permission from role: role_id=%d, role_name=%s, "+
			"permission=%s:%s",
		roleID,
		role.Name,
		obj,
		act,
	)

	return nil
}

// ListRolePermissions lists all permissions for a role (merges hardcoded + database)
func (s *PermissionService) ListRolePermissions(ctx context.Context, roleID int) ([]*permission.Permission, error) {
	// Get role
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.NotFound,
			"role not found: id=%d",
			roleID,
		)
	}

	// If system role, return hardcoded permissions
	if role.IsSystemRole() {
		systemPerms := auth.GetSystemRolePermissions(role.Name)
		if systemPerms != nil {
			return systemPerms, nil
		}
	}

	// For custom roles, load from database
	return s.permRepo.ListPermissionsByRole(ctx, roleID)
}
