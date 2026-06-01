package orggraphql

import (
	"modelcraft/internal/app/cluster"
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/app/permission"
	appProfile "modelcraft/internal/app/profile"
	"modelcraft/internal/app/project"
	"modelcraft/internal/app/rls"
	"modelcraft/internal/app/role"
	"modelcraft/internal/domain/user"
)

// Resolver is the GraphQL resolver for org domain
type Resolver struct {
	// Project CRUD
	ProjectAppService    *project.ProjectAppService
	ClusterAppService    *cluster.DatabaseClusterAppService
	AuthSchemaAppService *rls.AuthSchemaAppService

	// Organization
	OrganizationAppService *organization.OrganizationAppService
	ProfileAppService      *appProfile.AppService
	UserRepo               user.UserRepository

	// Permission (Casbin)
	RoleAppService    *role.RoleAppService
	RoleService       *permission.RoleService
	PermissionService *permission.PermissionService
	UserRoleService   *permission.UserRoleService

	// EndUser management
	EndUserMgmtAppService *appEnduser.EndUserManagementAppService
	MetaUserAppService    *appEnduser.MetaUserAppService

	// EndUser PAT management
	APITokenService *appEnduser.APITokenService
}
