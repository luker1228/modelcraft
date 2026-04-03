# rest-api-management Specification

## Purpose
TBD - created by archiving change adopt-oapi-codegen. Update Purpose after archive.
## Requirements
### Requirement: REST API Scope
The system SHALL maintain REST API endpoints only for tenant-management APIs (Auth and Organization) and infrastructure webhooks. Business domain APIs (Projects, Models, Clusters, Enums) SHALL be managed exclusively through GraphQL.

#### Scenario: Auth and Org endpoints are REST-based
- **WHEN** an API consumer needs to authenticate or manage organizations
- **THEN** REST endpoints at `/api/auth/*` and `/api/org/*` are available via OpenAPI spec
- **AND** endpoints follow REST conventions (POST for mutations, GET for queries where applicable)

#### Scenario: Business endpoints use GraphQL only
- **WHEN** an API consumer needs to manage projects, models, clusters, or enums
- **THEN** clients MUST use GraphQL at `/org/modelcraft/design/graphql` for all business operations
- **AND** no REST endpoints SHALL exist for business domain entities

### Requirement: Server Implementation Scope
The HTTP server implementation SHALL only contain handlers for tenant-management and webhook endpoints. Business domain logic MUST be served through the GraphQL engine.

#### Scenario: Server struct contains only auth/org/webhook handlers
- **WHEN** the HTTP server is constructed
- **THEN** the Server struct MUST only reference AuthHandler, OrgHandler, and WebhookHandler
- **AND** the GinEngine handles GraphQL routing independently
- **AND** no design domain HTTP handlers (Project, Model, Cluster, Enum) SHALL be present

#### Scenario: Chi router serves only tenant management routes
- **WHEN** the Chi router is configured
- **THEN** oapi-codegen generated routes MUST only include auth, org, and webhook paths
- **AND** business domain routes are NOT registered in the Chi router
- **AND** GraphQL routes are delegated to the Chi engine via middleware

