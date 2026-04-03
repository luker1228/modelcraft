package permission

import (
	"context"
	"modelcraft/internal/domain/permission"
	"modelcraft/internal/infrastructure/auth"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"

	"github.com/casbin/casbin/v2"
)

// RoleService provides business logic for role management
type RoleService struct {
	roleRepo     permission.RoleRepository
	permRepo     permission.PermissionRepository
	userRoleRepo permission.UserRoleRepository
	enforcer     *casbin.Enforcer
}

// NewRoleService creates a new role service
func NewRoleService(
	roleRepo permission.RoleRepository,
	permRepo permission.PermissionRepository,
	userRoleRepo permission.UserRoleRepository,
) *RoleService {
	enforcer, _ := auth.GetEnforcer() // 忽略错误，使用全局 enforcer
	return &RoleService{
		roleRepo:     roleRepo,
		permRepo:     permRepo,
		userRoleRepo: userRoleRepo,
		enforcer:     enforcer,
	}
}

// CreateCustomRoleInput represents input for creating a custom role
type CreateCustomRoleInput struct {
	Name        string
	Description string
	OrgName     string
}

// CreateCustomRole creates a new custom role with validation
func (s *RoleService) CreateCustomRole(
	ctx context.Context,
	input *CreateCustomRoleInput,
) (*permission.Role, error) {
	logger := logfacade.GetLogger(ctx)

	// Validate input
	if input.Name == "" {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ParamInvalid,
			"role name cannot be empty",
		)
	}
	if input.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ParamInvalid,
			"org_name cannot be empty",
		)
	}

	// Prevent using system role names
	if permission.IsSystemRoleName(input.Name) {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.Conflict,
			"cannot create custom role with system role name: "+
				"%s",
			input.Name,
		)
	}

	// Prevent using __SYSTEM__ org_name
	if input.OrgName == permission.SystemOrgName {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ParamInvalid,
			"cannot use '__SYSTEM__' as org_name for custom role",
		)
	}

	// Check if role already exists
	existing, err := s.roleRepo.GetRoleByNameAndOrg(ctx, input.Name, input.OrgName)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.Conflict,
			"role with name '%s' already exists in organization '%s'",
			input.Name,
			input.OrgName,
		)
	}

	// Create role entity
	role := permission.NewRole(input.Name, input.Description, input.OrgName)

	// Validate entity
	if err := role.Validate(); err != nil {
		return nil, err
	}

	// Persist to database
	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	logger.Infof(
		ctx, "Created custom role: name=%s, org_name=%s, id=%d",
		role.Name,
		role.OrgName,
		role.ID,
	)

	return role, nil
}

// UpdateRoleInput represents input for updating a role
type UpdateRoleInput struct {
	Name        string
	Description string
}

// UpdateRole updates an existing role with system role protection
func (s *RoleService) UpdateRole(
	ctx context.Context,
	roleID int,
	input *UpdateRoleInput,
) (*permission.Role, error) {
	logger := logfacade.GetLogger(ctx)

	// Get existing role
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

	// System role protection
	if !role.CanModify() {
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.OperationDenied,
			"cannot modify system role: %s",
			role.Name,
		)
	}

	// Update fields
	if input.Name != "" {
		if err := s.validateRoleRename(ctx, role, input.Name); err != nil {
			return nil, err
		}
		role.Name = input.Name
	}

	if input.Description != "" {
		role.Description = input.Description
	}

	// Validate entity
	if err := role.Validate(); err != nil {
		return nil, err
	}

	// Persist changes
	if err := s.roleRepo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}

	logger.Infof(
		ctx, "Updated role: id=%d, name=%s, org_name=%s",
		role.ID,
		role.Name,
		role.OrgName,
	)

	return role, nil
}

func (s *RoleService) validateRoleRename(ctx context.Context, role *permission.Role, newName string) error {
	// Prevent using system role names
	if permission.IsSystemRoleName(newName) {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.Conflict,
			"cannot rename custom role to system role name: "+
				"%s",
			newName,
		)
	}

	if newName == role.Name {
		return nil
	}

	// Check name uniqueness within org
	existing, err := s.roleRepo.GetRoleByNameAndOrg(ctx, newName, role.OrgName)
	if err != nil {
		return err
	}
	if existing != nil {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.Conflict,
			"role with name '%s' already exists in organization '%s'",
			newName,
			role.OrgName,
		)
	}

	return nil
}

// DeleteRole deletes a role with system role protection and cascade cleanup
func (s *RoleService) DeleteRole(ctx context.Context, roleID int) error {
	logger := logfacade.GetLogger(ctx)

	// Get existing role
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
	if !role.CanDelete() {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.OperationDenied,
			"cannot delete system role: %s",
			role.Name,
		)
	}

	// Delete from database (cascade delete will handle user_roles and role_permissions via foreign keys)
	if err := s.roleRepo.DeleteRole(ctx, roleID); err != nil {
		return err
	}

	// Remove custom role permissions from Casbin enforcer
	if s.enforcer != nil {
		// Remove all policies for this role
		_, err := s.enforcer.RemoveFilteredPolicy(0, role.Name)
		if err != nil {
			logger.Errorf(
				ctx, "Failed to remove Casbin policies for role %s: %v",
				role.Name,
				err,
			)
			// Don't fail the operation if Casbin cleanup fails
		}
	}

	logger.Infof(
		ctx, "Deleted role: id=%d, name=%s, org_name=%s",
		role.ID,
		role.Name,
		role.OrgName,
	)

	return nil
}

// GetRole retrieves a role by ID
func (s *RoleService) GetRole(ctx context.Context, roleID int) (*permission.Role, error) {
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

	return role, nil
}

// ListRoles lists all roles for an organization
func (s *RoleService) ListRoles(
	ctx context.Context,
	orgName string,
	includeSystem bool,
) ([]*permission.Role, error) {
	return s.roleRepo.ListRolesByOrg(ctx, orgName, includeSystem)
}
