package auth

import (
	"context"
	"fmt"
	"modelcraft/internal/infrastructure/auth"

	domainPerm "modelcraft/internal/domain/permission"
)

// RolePermissionInfo contains permissions for a single role
type RolePermissionInfo struct {
	Permissions []string `json:"permissions"`
}

// RolePermissions maps role names to their permissions
// Example: {"owner": {"permissions": ["*:*"]}, "editor": {"permissions": ["model:read", "model:write"]}}
type RolePermissions map[string]*RolePermissionInfo

// PermissionLoaderInterface defines the interface for loading user permissions.
// This enables mocking in tests.
type PermissionLoaderInterface interface {
	// LoadUserPermissions loads only permissions (legacy method, kept for backward compatibility)
	LoadUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)

	// LoadUserPermissionsAndRoles loads both roles and their permissions in a map structure
	// Returns: map[roleName]*RolePermissionInfo where each role contains its permissions
	LoadUserPermissionsAndRoles(ctx context.Context, userID, orgName string) (RolePermissions, error)
}

// PermissionLoader loads user permissions from the database.
// Phase 1: Direct DB queries (no caching)
// Phase 2: Redis caching will be added
type PermissionLoader struct {
	userRoleRepo domainPerm.UserRoleRepository
	roleRepo     domainPerm.RoleRepository
	permRepo     domainPerm.PermissionRepository
}

// Ensure PermissionLoader implements PermissionLoaderInterface
var _ PermissionLoaderInterface = (*PermissionLoader)(nil)

// NewPermissionLoader creates a new permission loader
func NewPermissionLoader(
	userRoleRepo domainPerm.UserRoleRepository,
	roleRepo domainPerm.RoleRepository,
	permRepo domainPerm.PermissionRepository,
) *PermissionLoader {
	return &PermissionLoader{
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		permRepo:     permRepo,
	}
}

// LoadUserPermissions loads all permissions for a user in a specific organization.
// Returns a list of permission strings in format "resource:action" (e.g., "model:read", "cluster:manage").
// Returns an empty array if the user has no roles in the organization.
//
// Phase 1 Implementation:
// 1. Query user roles in organization
// 2. For each role, load permissions
// 3. Deduplicate and return permission strings
//
// Performance Target: < 50ms (direct DB query)
func (l *PermissionLoader) LoadUserPermissions(ctx context.Context, userID, orgName string) ([]string, error) {
	// Step 1: Query user roles in organization
	userRoles, err := l.userRoleRepo.ListUserRoles(ctx, userID, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// If user has no roles, return empty permissions
	if len(userRoles) == 0 {
		return []string{}, nil
	}

	// Step 2: Load permissions for each role
	permissionMap := make(map[string]bool) // Deduplicate permissions

	for _, userRole := range userRoles {
		// Get role details
		role, err := l.roleRepo.GetRoleByID(ctx, userRole.RoleID)
		if err != nil {
			continue // Skip this role but continue with others
		}

		if role == nil {
			continue
		}

		// Load permissions based on role type
		var rolePerms []*domainPerm.Permission

		// If system role, get hardcoded permissions
		if role.IsSystemRole() {
			rolePerms = auth.GetSystemRolePermissions(role.Name)
		} else {
			// Custom role - load from database
			rolePerms, err = l.permRepo.ListPermissionsByRole(ctx, userRole.RoleID)
			if err != nil {
				continue // Skip this role but continue with others
			}
		}

		// Add permissions to map (deduplicate)
		for _, perm := range rolePerms {
			permStr := perm.String() // Format: "resource:action"
			permissionMap[permStr] = true
		}
	}

	// Step 3: Convert map to slice
	permissions := make([]string, 0, len(permissionMap))
	for perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// LoadUserPermissionsAndRoles loads both roles and their permissions in a map structure.
// Returns: map[roleName]*RolePermissionInfo where each role contains its permissions
// Returns an empty map if the user has no roles in the organization.
//
// This method does NOT deduplicate roles - if a user has multiple role assignments with the same role,
// each role appears once in the map with combined permissions.
//
// Performance Target: < 50ms (direct DB query)
func (l *PermissionLoader) LoadUserPermissionsAndRoles(
	ctx context.Context,
	userID string,
	orgName string,
) (RolePermissions, error) {
	// Step 1: Query user roles in organization
	userRoles, err := l.userRoleRepo.ListUserRoles(ctx, userID, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// If user has no roles, return empty map
	if len(userRoles) == 0 {
		return RolePermissions{}, nil
	}

	// Step 2: Build role permissions map
	rolePerms := make(RolePermissions)

	for _, userRole := range userRoles {
		// Get role details
		role, err := l.roleRepo.GetRoleByID(ctx, userRole.RoleID)
		if err != nil {
			continue // Skip this role but continue with others
		}

		if role == nil {
			continue
		}

		// Load permissions based on role type
		var perms []*domainPerm.Permission

		// If system role, get hardcoded permissions
		if role.IsSystemRole() {
			perms = auth.GetSystemRolePermissions(role.Name)
		} else {
			// Custom role - load from database
			perms, err = l.permRepo.ListPermissionsByRole(ctx, userRole.RoleID)
			if err != nil {
				continue // Skip this role but continue with others
			}
		}

		// Convert permissions to strings
		permStrings := make([]string, 0, len(perms))
		for _, perm := range perms {
			permStrings = append(permStrings, perm.String()) // Format: "resource:action"
		}

		// Store role permissions in map (if role exists, permissions are overwritten)
		rolePerms[role.Name] = &RolePermissionInfo{
			Permissions: permStrings,
		}
	}

	return rolePerms, nil
}
