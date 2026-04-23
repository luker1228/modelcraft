package endusergraphql

import (
	"modelcraft/internal/app/enduser"
	"modelcraft/internal/app/modeldesign"
)

// Resolver is the GraphQL resolver for end-user domain.
// It provides a strict subset of capabilities for runtime data access only.
type Resolver struct {
	// Model design service for catalog queries
	ModelDesignService *modeldesign.ModelDesignAppService
	// End-user management service for standard user queries.
	EndUserMgmtService *enduser.EndUserManagementAppService
}
