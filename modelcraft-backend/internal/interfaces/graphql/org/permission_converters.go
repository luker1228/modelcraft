package orggraphql

import (
	"modelcraft/internal/domain/permission"
	"modelcraft/internal/interfaces/graphql/org/generated"
)

// convertRoleToGraphQL converts domain Role to GraphQL PermissionRole
func convertRoleToGraphQL(role *permission.Role) *generated.PermissionRole {
	if role == nil {
		return nil
	}

	return &generated.PermissionRole{
		ID:          int32(role.ID), //nolint:gosec // G115: role ID won't exceed int32 in practice
		Name:        role.Name,
		Description: strPtr(role.Description),
		IsSystem:    role.IsSystem,
		OrgName:     role.OrgName,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

// convertRolesToGraphQL converts slice of domain Roles to GraphQL PermissionRoles
func convertRolesToGraphQL(roles []*permission.Role) []*generated.PermissionRole {
	result := make([]*generated.PermissionRole, len(roles))
	for i, role := range roles {
		result[i] = convertRoleToGraphQL(role)
	}
	return result
}

// convertPermissionsToGraphQL converts domain Permissions to GraphQL PermissionDefs
func convertPermissionsToGraphQL(perms []*permission.Permission) []*generated.PermissionDef {
	result := make([]*generated.PermissionDef, len(perms))
	for i, perm := range perms {
		result[i] = &generated.PermissionDef{
			Obj: perm.Obj,
			Act: perm.Act,
		}
	}
	return result
}

// convertUserRolesToGraphQL converts domain UserRoles to GraphQL UserRoleAssignments
func convertUserRolesToGraphQL(userRoles []*permission.UserRole) []*generated.UserRoleAssignment {
	result := make([]*generated.UserRoleAssignment, len(userRoles))
	for i, ur := range userRoles {
		result[i] = &generated.UserRoleAssignment{
			ID:        int32(ur.ID), //nolint:gosec // G115: user role ID won't exceed int32 in practice
			UserID:    ur.UserID,
			RoleID:    int32(ur.RoleID), //nolint:gosec // G115: role ID won't exceed int32 in practice
			OrgName:   ur.OrgName,
			CreatedAt: ur.CreatedAt,
		}
	}
	return result
}

// strPtr converts string to pointer
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
