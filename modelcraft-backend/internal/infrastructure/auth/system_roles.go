package auth

import "modelcraft/internal/domain/permission"

// SystemRolePermissions defines hardcoded permissions for system roles.
// These permissions are loaded into Casbin enforcer at startup and cannot be modified.
//
// System Roles:
// - owner: Full access to all resources (wildcard permission)
// - admin: Full access to project, model, cluster, enum resources (no user management)
// - editor: Read and update access to resources (can create models/enums, cannot delete)
// - viewer: Read-only access to all resources
var SystemRolePermissions = map[string][]*permission.Permission{
	// Owner: Wildcard permission grants access to everything
	permission.RoleOwner: {
		{Obj: "*", Act: "*"},
	},

	// Admin: Full access to all resources
	permission.RoleAdmin: {
		{Obj: "*", Act: "*"},
	},

	// Editor: Full access except delete operations
	permission.RoleEditor: {
		{Obj: "*", Act: "create"},
		{Obj: "*", Act: "read"},
		{Obj: "*", Act: "update"},
	},

	// Viewer: Read-only access to all resources (wildcard object, read action)
	permission.RoleViewer: {
		{Obj: "*", Act: "read"},
	},

	// Guest: Read-only access + create temporary API keys (demo mode)
	permission.RoleGuest: {
		{Obj: "*", Act: "read"},
		{Obj: "apitoken", Act: "create"},
	},
}

// GetSystemRolePermissions returns permissions for a given system role name.
// Returns nil if the role is not a system role.
func GetSystemRolePermissions(roleName string) []*permission.Permission {
	return SystemRolePermissions[roleName]
}

// IsSystemRole checks if a role name is a system role.
func IsSystemRole(roleName string) bool {
	_, exists := SystemRolePermissions[roleName]
	return exists
}
