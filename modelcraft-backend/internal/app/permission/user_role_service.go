package permission

import (
	"context"
	"modelcraft/internal/domain/permission"
	"modelcraft/internal/infrastructure/auth"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"

	"github.com/casbin/casbin/v2"
)

// UserRoleService provides business logic for user-role assignment
type UserRoleService struct {
	roleRepo       permission.RoleRepository
	userRoleRepo   permission.UserRoleRepository
	permRepo       permission.PermissionRepository
	versionManager PermissionVersionManager
	enforcer       *casbin.Enforcer
}

// PermissionVersionManager manages permission version numbers for cache invalidation
type PermissionVersionManager interface {
	GetVersion(ctx context.Context, orgName, userID string) (int64, error)
	IncrementVersion(ctx context.Context, orgName, userID string) (int64, error)
}

// NewUserRoleService creates a new user-role service
func NewUserRoleService(
	roleRepo permission.RoleRepository,
	userRoleRepo permission.UserRoleRepository,
	permRepo permission.PermissionRepository,
	versionManager PermissionVersionManager,
) *UserRoleService {
	enforcer, _ := auth.GetEnforcer() // 忽略错误，使用全局 enforcer
	return &UserRoleService{
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		permRepo:       permRepo,
		versionManager: versionManager,
		enforcer:       enforcer,
	}
}

// AssignRoleToUser assigns a role to a user in an organization
func (s *UserRoleService) AssignRoleToUser(ctx context.Context, userID string, roleID int, orgName string) error {
	logger := logfacade.GetLogger(ctx)

	// Validate input
	if userID == "" {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "user_id cannot be empty")
	}
	if roleID <= 0 {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "role_id must be positive")
	}
	if orgName == "" {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ParamInvalid, "org_name cannot be empty")
	}

	// Get role to validate it exists
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

	// For custom roles, validate org_name matches
	if !role.IsSystemRole() && role.OrgName != orgName {
		return bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.ParamInvalid,
			"role belongs to organization '%s', "+
				"cannot assign in organization '%s'",
			role.OrgName,
			orgName,
		)
	}

	// Check if assignment already exists
	existing, err := s.userRoleRepo.GetUserRole(ctx, userID, roleID, orgName)
	if err != nil {
		return err
	}
	if existing != nil {
		// Already assigned, no-op
		logger.Infof(
			ctx, "User role already assigned: user_id=%s, role_id=%d, org_name=%s",
			userID,
			roleID,
			orgName,
		)
		return nil
	}

	// Create user-role binding
	userRole := permission.NewUserRole(userID, roleID, orgName)
	if err := userRole.Validate(); err != nil {
		return err
	}

	// Persist to database
	if err := s.userRoleRepo.AssignRole(ctx, userRole); err != nil {
		return err
	}

	// Add to Casbin enforcer
	if s.enforcer != nil {
		err := auth.AddUserRole(s.enforcer, userID, role.Name)
		if err != nil {
			logger.Errorf(context.Background(), err, "Failed to add user role to Casbin enforcer")
			// Don't fail the operation if Casbin sync fails
		}
	}

	// Increment permission version to invalidate cache
	if s.versionManager != nil {
		_, err := s.versionManager.IncrementVersion(ctx, orgName, userID)
		if err != nil {
			// Log error but don't fail the assignment
			// Permission change succeeded, cache will expire naturally in 5min
			logger.Errorf(
				ctx, err, "Failed to increment permission version: userId=%s, orgName=%s",
				userID,
				orgName,
			)
		} else {
			logger.Infof(
				ctx, "Incremented permission version: userId=%s, orgName=%s",
				userID,
				orgName,
			)
		}
	}

	logger.Infof(
		ctx, "Assigned role to user: user_id=%s, role_id=%d, role_name=%s, "+
			"org_name=%s",
		userID,
		roleID,
		role.Name,
		orgName,
	)

	return nil
}

// RevokeRoleFromUser revokes a role from a user in an organization
func (s *UserRoleService) RevokeRoleFromUser(ctx context.Context, userID string, roleID int, orgName string) error {
	logger := logfacade.GetLogger(ctx)

	// Get role to get its name
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

	// Remove from database
	if err := s.userRoleRepo.RevokeRole(ctx, userID, roleID, orgName); err != nil {
		return err
	}

	// Remove from Casbin enforcer
	if s.enforcer != nil {
		err := auth.RemoveUserRole(s.enforcer, userID, role.Name)
		if err != nil {
			logger.Errorf(context.Background(), err, "Failed to remove user role from Casbin enforcer")
			// Don't fail the operation if Casbin sync fails
		}
	}

	// Increment permission version to invalidate cache
	if s.versionManager != nil {
		_, err := s.versionManager.IncrementVersion(ctx, orgName, userID)
		if err != nil {
			// Log error but don't fail the revocation
			// Permission change succeeded, cache will expire naturally in 5min
			logger.Errorf(
				ctx, err, "Failed to increment permission version: userId=%s, orgName=%s",
				userID,
				orgName,
			)
		} else {
			logger.Infof(
				ctx, "Incremented permission version: userId=%s, orgName=%s",
				userID,
				orgName,
			)
		}
	}

	logger.Infof(
		ctx, "Revoked role from user: user_id=%s, role_id=%d, role_name=%s, "+
			"org_name=%s",
		userID,
		roleID,
		role.Name,
		orgName,
	)

	return nil
}

// ListUserRoles lists all role assignments for a user in an organization
func (s *UserRoleService) ListUserRoles(ctx context.Context, userID, orgName string) ([]*permission.UserRole, error) {
	return s.userRoleRepo.ListUserRoles(ctx, userID, orgName)
}

// ListRoleUsers lists all users who have a specific role in an organization
func (s *UserRoleService) ListRoleUsers(
	ctx context.Context,
	roleID int,
	orgName string,
) ([]*permission.UserRole, error) {
	// Validate role exists
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

	return s.userRoleRepo.ListRoleUsers(ctx, roleID, orgName)
}

// CheckPermission checks if a user has permission to perform an action on a resource
// This is the core permission check logic used by GraphQL directives and middleware
func (s *UserRoleService) CheckPermission(ctx context.Context, userID, orgName, action string) (bool, error) {
	logger := logfacade.GetLogger(ctx)

	// Parse action into obj:act format
	perm, err := permission.NewPermissionFromString(action)
	if err != nil {
		return false, err
	}

	// Get user's roles in this organization
	userRoles, err := s.userRoleRepo.ListUserRoles(ctx, userID, orgName)
	if err != nil {
		return false, err
	}

	if len(userRoles) == 0 {
		logger.Infof(context.Background(), "Permission check denied: user %s has no roles in org %s", userID, orgName)
		return false, nil
	}

	// Load permissions for each role and check with Casbin
	for _, userRole := range userRoles {
		role, err := s.roleRepo.GetRoleByID(ctx, userRole.RoleID)
		if err != nil {
			logger.Errorf(context.Background(), err, "Failed to get role %d", userRole.RoleID)
			continue
		}
		if role == nil {
			continue
		}

		// CRITICAL: Add user-role mapping to Casbin enforcer
		s.ensureUserRoleMapping(userID, role.Name, logger)

		// For system roles, permissions are already loaded in enforcer at startup
		// For custom roles, we need to load permissions from database and sync to enforcer
		s.syncCustomRolePermissions(ctx, role, logger)
	}

	// Check permission with Casbin enforcer
	if s.enforcer != nil {
		allowed, err := auth.CheckPermission(s.enforcer, userID, perm.Obj, perm.Act)
		if err != nil {
			logger.Errorf(context.Background(), err, "Casbin permission check failed")
			return false, err
		}

		if allowed {
			logger.Infof(
				context.Background(),
				"Permission check allowed: user=%s, org=%s, action=%s",
				userID, orgName, action,
			)
		} else {
			logger.Infof(
				context.Background(),
				"Permission check denied: user=%s, org=%s, action=%s",
				userID, orgName, action,
			)
		}

		return allowed, nil
	}

	// Fallback: deny if enforcer is not available
	logger.Errorf(context.Background(), nil, "Casbin enforcer not available for permission check")
	return false, nil
}

func (s *UserRoleService) ensureUserRoleMapping(userID, roleName string, logger logfacade.Logger) {
	if s.enforcer == nil {
		return
	}

	if err := auth.AddUserRole(s.enforcer, userID, roleName); err != nil {
		logger.Errorf(context.Background(), err, "Failed to add user role mapping to enforcer")
		// Continue anyway - the mapping might already exist
	}
}

func (s *UserRoleService) syncCustomRolePermissions(
	ctx context.Context, role *permission.Role, logger logfacade.Logger,
) {
	if role.IsSystemRole() {
		return
	}

	perms, err := s.permRepo.ListPermissionsByRole(ctx, role.ID)
	if err != nil {
		logger.Errorf(context.Background(), err, "Failed to list permissions for role %d", role.ID)
		return
	}

	if s.enforcer == nil {
		return
	}

	for _, p := range perms {
		if _, err := s.enforcer.AddPolicy(role.Name, p.Obj, p.Act); err != nil {
			logger.Errorf(context.Background(), err, "Failed to sync permission to Casbin")
		}
	}
}
