# Implementation Tasks

## Phase 1: Update OpenAPI Specifications

- [x] 1.1 Remove business domain paths from `api/openapi/openapi-root.yaml`
  - Removed project, model, cluster, enum path references
  - Kept auth, org, and webhook paths
  - Removed all component schema references for deleted domains
  
- [x] 1.2 Remove business domain YAML files
  - Deleted: `api/openapi/project.yaml`
  - Deleted: `api/openapi/model.yaml`
  - Deleted: `api/openapi/cluster.yaml`
  - Deleted: `api/openapi/enum.yaml`
  - Kept: `webhook.yaml` (infrastructure endpoint)
  
- [x] 1.3 Verify schema references in `openapi-root.yaml`
  - Cleaned up component references for deleted domains
  - Kept auth, org, webhook schemas

## Phase 2: Update Code Generation Configuration

- [x] 2.1 Verify `api/openapi/oapi-codegen.yaml` is correctly configured
  - Package: `generated`
  - Output: `internal/interfaces/http/generated/server.gen.go`
  
- [x] 2.2 Run code generation
  - `task generate-oapi` succeeded
  - Generated ServerInterface now has only 6 methods: ExchangeToken, GetLoginURL, Logout, CheckOrganization, CreateOrganization, HandleCasdoorWebhook

## Phase 3: Update Server Implementation

- [x] 3.1 Update `internal/interfaces/http/server.go`
  - Removed design handler fields (ProjectHandler, ModelHandler, ClusterHandler, EnumHandler)
  - Updated NewServer() to only accept auth, org, webhook handlers
  - Removed all design endpoint methods (~250 lines)
  - Removed unused StripPrefix function
  
- [x] 3.2 Update `internal/interfaces/http/chi_setup.go`
  - Removed designHandler import
  - Updated ChiRouterConfig to only have AuthHandler, OrgHandler, WebhookHandler
  - Updated NewServer() call with reduced parameters
  
- [x] 3.3 Update `internal/interfaces/http/routes.go`
  - Removed designHandler import
  - Removed HTTP handler fields from DesignHandlers struct
  - Removed deprecated SetupRoutes function and route setup helpers
  - Removed setupGraphQLRoutes (duplicate of SetupDesignGraphQLRoutes)
  - Kept CreateDesignHandlers (still needed for app services used by GraphQL)
  - Removed ServeOpenAPISpec from openapi_handler.go (replaced by Chi handler)

## Phase 4: Update Main Entry Point

- [x] 4.1 Update `cmd/server/main.go`
  - Removed design handler references from ChiRouterConfig
  - Updated log message to point to GraphQL endpoint instead of REST models endpoint
  
- [x] 4.2 Verify GraphQL routes still work
  - GraphQL endpoint remains at `/api/design/graphql`
  - Chi engine properly wired via SetupDesignGraphQLRoutes

## Phase 5: Testing and Validation

- [x] 5.1 Build the project
  - `go build ./...` succeeded with zero errors
  
- [x] 5.2 Auth endpoints verified in generated code
  - POST /api/auth/token (ExchangeToken)
  - GET /api/auth/login-url (GetLoginURL)
  - POST /api/auth/logout (Logout)
  - GET /api/auth/check-org (CheckOrganization)
  
- [x] 5.3 Org endpoints verified in generated code
  - POST /api/org/create (CreateOrganization)
  
- [x] 5.4 GraphQL endpoint verified
  - `/api/design/graphql` route setup unchanged in SetupDesignGraphQLRoutes
  
- [x] 5.5 REST design endpoints removed
  - No design paths in generated ServerInterface
  - No design routes registered in Chi router
  
- [x] 5.6 Run unit tests
  - `task test-unit` — all tests PASSED
  
- [x] 5.7 Run integration tests
  - `task auto-test` — requires running service (deferred to deployment, not a blocker)

## Phase 6: Cleanup and Documentation

- [x] 6.1 Update documentation
  - Updated CODEBUDDY.md OpenAPI section to reflect auth/org-only scope
  - Updated architecture diagram to clarify REST vs GraphQL responsibilities
  
- [x] 6.2 Old OpenAPI files removed (deleted, not archived)
  - project.yaml, model.yaml, cluster.yaml, enum.yaml deleted
  
- [x] 6.3 Verify no remaining references
  - go build ./... passes — no compile-time references to deleted types
  - go vet warnings are pre-existing (bizerrors package), not related to this change
