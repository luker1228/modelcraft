package permission

import "context"

// RoleRepository defines the interface for role persistence operations.
type RoleRepository interface {
	// CreateRole creates a new role in the database
	CreateRole(ctx context.Context, role *Role) error

	// GetRoleByID retrieves a role by its ID
	GetRoleByID(ctx context.Context, id int) (*Role, error)

	// GetRoleByNameAndOrg retrieves a role by name and org_name
	GetRoleByNameAndOrg(ctx context.Context, name, orgName string) (*Role, error)

	// ListRolesByOrg lists all roles for a specific organization
	// includeSystem determines whether to include system roles in the result
	ListRolesByOrg(ctx context.Context, orgName string, includeSystem bool) ([]*Role, error)

	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, role *Role) error

	// DeleteRole deletes a role by ID
	// This will cascade delete related user_roles and role_permissions via foreign keys
	DeleteRole(ctx context.Context, id int) error
}

// PermissionRepository defines the interface for permission persistence operations.
type PermissionRepository interface {
	// AddPermission adds a permission to a role
	AddPermission(ctx context.Context, roleID int, orgName string, perm *Permission) error

	// RemovePermission removes a permission from a role
	RemovePermission(ctx context.Context, roleID int, obj, act string) error

	// ListPermissionsByRole lists all permissions for a specific role
	ListPermissionsByRole(ctx context.Context, roleID int) ([]*Permission, error)

	// ListPermissionsByRoleAndOrg lists permissions for a role filtered by org_name
	ListPermissionsByRoleAndOrg(ctx context.Context, roleID int, orgName string) ([]*Permission, error)

	// DeletePermissionsByRole deletes all permissions for a role (used during role deletion)
	DeletePermissionsByRole(ctx context.Context, roleID int) error
}

// UserRoleRepository defines the interface for user-role binding operations.
type UserRoleRepository interface {
	// AssignRole assigns a role to a user in a specific organization
	AssignRole(ctx context.Context, userRole *UserRole) error

	// RevokeRole revokes a role from a user in a specific organization
	RevokeRole(ctx context.Context, userID string, roleID int, orgName string) error

	// ListUserRoles lists all role assignments for a user in a specific organization
	ListUserRoles(ctx context.Context, userID, orgName string) ([]*UserRole, error)

	// ListRoleUsers lists all users who have a specific role in an organization
	ListRoleUsers(ctx context.Context, roleID int, orgName string) ([]*UserRole, error)

	// GetUserRole retrieves a specific user-role binding
	GetUserRole(ctx context.Context, userID string, roleID int, orgName string) (*UserRole, error)

	// DeleteUserRolesByRole deletes all user-role bindings for a role (used during role deletion)
	DeleteUserRolesByRole(ctx context.Context, roleID int) error
}
