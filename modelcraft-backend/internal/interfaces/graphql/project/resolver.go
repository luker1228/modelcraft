package projectgraphql

import (
	"modelcraft/internal/app/cluster"
	appmodeldatabase "modelcraft/internal/app/modeldatabase"
	"modelcraft/internal/app/modeldesign"
	"modelcraft/internal/app/permission"
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

	// RLS (Row Level Security) — expression validation + dry-run
	RLSExprValidateService *rls.RLSExprValidateService

	// RLS Policy V2
	PolicyCRUDService *rls.DataPolicyService

	// Database management
	ModelDatabaseAppService     *appmodeldatabase.ModelDatabaseAppService
	ModelDatabaseSyncAppService *appmodeldatabase.ModelDatabaseSyncAppService

	// Model sync (syncModelsFromDB)
	SyncModelsAppService *appmodeldatabase.SyncModelsAppService
}
