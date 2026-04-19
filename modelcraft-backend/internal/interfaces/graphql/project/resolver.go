package projectgraphql

import (
	"modelcraft/internal/app/cluster"
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

	// RLS (Row Level Security)
	RLSPolicyAppService *rls.ModelRLSPolicyAppService
	AuthSchemaAppService *rls.AuthSchemaAppService
}
