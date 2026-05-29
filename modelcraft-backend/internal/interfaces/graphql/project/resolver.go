package projectgraphql

import (
	"modelcraft/internal/app/cluster"
	appEnduser "modelcraft/internal/app/enduser"
	appmodeldatabase "modelcraft/internal/app/modeldatabase"
	"modelcraft/internal/app/modeldesign"
	"modelcraft/internal/app/permission"
	apprbac "modelcraft/internal/app/rbac"
	"modelcraft/internal/app/rls"
)

// Resolver is the GraphQL resolver for project domain
type Resolver struct {
	// Cluster operations
	ClusterAppService *cluster.DatabaseClusterAppService

	// Model design
	ModelDesignService       *modeldesign.ModelDesignAppService
	ReverseEngineerService   *modeldesign.ReverseEngineerAppService
	RepairModelUseCase       *modeldesign.RepairModelUseCase
	ActualSchemaQueryUseCase *modeldesign.ActualSchemaQueryUseCase
	GroupAppService          *modeldesign.ModelGroupAppService
	LogicalFKAppService      *modeldesign.LogicalFKAppService

	// Enum
	EnumAppService *modeldesign.EnumAppService

	// Permission (for @hasPermission directive)
	UserRoleService *permission.UserRoleService

	// Field selection checker
	FieldSelectionChecker *FieldSelectionChecker

	// RLS (Row Level Security)
	RLSPolicyAppService  *rls.ModelRLSPolicyAppService
	AuthSchemaAppService *rls.AuthSchemaAppService

	// End-User
	EndUserMgmtAppService *appEnduser.EndUserManagementAppService

	// RBAC (Data-Level Row & Column Permission)
	RBACPermissionSvc *apprbac.EndUserPermissionAppService
	RBACBundleSvc     *apprbac.EndUserBundleAppService
	RBACRoleSvc       *apprbac.EndUserRoleAppService
	RBACAuthzSvc      *apprbac.EndUserAuthzService

	// Database management
	ModelDatabaseAppService     *appmodeldatabase.ModelDatabaseAppService
	ModelDatabaseSyncAppService *appmodeldatabase.ModelDatabaseSyncAppService
}
