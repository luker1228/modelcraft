# Change: Adopt oapi-codegen for Auth/Org REST APIs, migrate business logic to GraphQL

## Why

Currently, the project maintains OpenAPI specifications for all domains (auth, org, webhook, projects, models, clusters, enums). However:

1. **Business API endpoints** (projects, models, clusters, enums) should be managed through GraphQL, not REST
2. Only **Auth and Org endpoints** are tenant-management APIs that justify REST API specification
3. Maintaining OpenAPI specs for business APIs creates duplication and synchronization overhead
4. The architecture already has GraphQL as the primary API layer for business logic
5. Restricting GET operations on business REST endpoints (if they exist) prevents query proliferation

## What Changes

- **Keep**: Auth (`auth.yaml`) and Org (`org.yaml`) OpenAPI specs fully maintained
- **Remove**: Project, Model, Cluster, Enum, Webhook OpenAPI specs (`project.yaml`, `model.yaml`, `cluster.yaml`, `enum.yaml`, `webhook.yaml`)
- **Disable**: All GET operations on non-auth/org endpoints in OpenAPI (business APIs use mutations only)
- **Update**: oapi-codegen configuration to only generate handlers for auth/org domains
- **Update**: Root OpenAPI spec to reference only auth/org paths
- **Migrate**: Design API routes to GraphQL (already done via gqlgen)
- **Remove**: HTTP handlers for business domains from the generated ServerInterface

## Impact

### Affected Specs
- `rest-api-management` - REST API scope definition

### Affected Code
- **OpenAPI files**: `api/openapi/openapi-root.yaml`, `api/openapi/project.yaml`, `api/openapi/model.yaml`, `api/openapi/cluster.yaml`, `api/openapi/enum.yaml`, `api/openapi/webhook.yaml`
- **Generated code**: `internal/interfaces/http/generated/server.gen.go` (regenerated with limited scope)
- **Server implementation**: `internal/interfaces/http/server.go` (remove design handler delegation)
- **Route registration**: `internal/interfaces/http/chi_setup.go` (remove design route mounting)
- **HTTP routing**: `internal/interfaces/http/routes.go` (remove design handlers)
- **Taskfile**: Verify `generate-oapi` and `bundle-oapi` still work correctly

### Breaking Changes
**BREAKING**: Any external REST API consumers using `/api/design/*` endpoints will need to migrate to GraphQL at `/api/design/graphql`

