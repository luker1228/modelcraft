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
	// MetaUserService for runtime meta/user route (me/findOne/findMany).
	// orgName is always injected from middleware; not exposed via GraphQL schema.
	MetaUserService *enduser.MetaUserAppService
}
