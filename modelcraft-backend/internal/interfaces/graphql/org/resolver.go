package orggraphql

import (
	"modelcraft/internal/app/apitoken"
	"modelcraft/internal/app/cluster"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/app/permission"
	appProfile "modelcraft/internal/app/profile"
	"modelcraft/internal/app/project"
	"modelcraft/internal/app/role"
	"modelcraft/internal/domain/user"
)

// Resolver is the GraphQL resolver for org domain
type Resolver struct {
	// Project CRUD
	ProjectAppService *project.ProjectAppService
	ClusterAppService *cluster.DatabaseClusterAppService

	// Organization
	OrganizationAppService *organization.OrganizationAppService
	ProfileAppService      *appProfile.AppService
	UserRepo               user.UserRepository

	// Permission (Casbin)
	RoleAppService    *role.RoleAppService
	RoleService       *permission.RoleService
	PermissionService *permission.PermissionService
	UserRoleService   *permission.UserRoleService

	// EndUser PAT management
	APITokenService *apitoken.APITokenService
}
