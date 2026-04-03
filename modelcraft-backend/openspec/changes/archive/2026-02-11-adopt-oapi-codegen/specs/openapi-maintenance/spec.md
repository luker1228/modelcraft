# OpenAPI Maintenance Strategy Specification

## ADDED Requirements

### Requirement: OpenAPI Spec Organization  
The OpenAPI specification files SHALL be organized by domain, with only Auth, Org, and Webhook domains maintaining YAML files. The root specification SHALL reference only tenant-management domains.

#### Scenario: Auth domain specification maintained
- **WHEN** authentication requirements change
- **THEN** updates are made to `api/openapi/auth.yaml`
- **AND** includes endpoints: POST /api/auth/token, GET /api/auth/login-url, POST /api/auth/logout, GET /api/auth/check-org

#### Scenario: Organization domain specification maintained
- **WHEN** organization management requirements change
- **THEN** updates are made to `api/openapi/org.yaml`
- **AND** includes endpoints: POST /api/org/create

#### Scenario: Root specification references only required domains
- **WHEN** the root `api/openapi/openapi-root.yaml` is created/updated
- **THEN** it MUST only include path references for auth, org, and webhook domains
- **AND** component schema references are pruned to only include referenced types
- **AND** business domain paths (projects, models, clusters, enums) MUST NOT be added

### Requirement: Code Generation Scope Restriction
The oapi-codegen tooling SHALL generate server handlers only for tenant-management domains (Auth, Org, Webhook), with no handlers generated for business domains.

#### Scenario: Generated code includes only tenant management handlers
- **WHEN** `task generate-oapi` is executed
- **THEN** the generated `internal/interfaces/http/generated/server.gen.go` contains ServerInterface with only:
  - Auth handler methods (ExchangeToken, GetLoginURL, Logout, CheckOrganization)
  - Org handler methods (CreateOrganization)
  - Webhook handler methods (HandleCasdoorWebhook)
- **AND** no Project, Model, Cluster, or Enum handler methods are generated

### Requirement: OpenAPI Spec Bundling
The OpenAPI specification bundling process SHALL only include tenant-management domains in the final bundled spec.

#### Scenario: Bundle operation produces reduced spec
- **WHEN** `task bundle-oapi` is executed
- **THEN** `api/openapi/openapi.yaml` contains only auth, org, and webhook paths
- **AND** component schemas reference only types used by included paths
- **AND** the bundled spec is valid and completeness is not compromised

### Requirement: Business Domain API Exclusion
Business domain APIs (Projects, Models, Clusters, Enums) SHALL NOT have OpenAPI specifications. These domains are served exclusively via GraphQL.

#### Scenario: No business domain OpenAPI files exist
- **WHEN** the OpenAPI specification directory is examined
- **THEN** no `project.yaml`, `model.yaml`, `cluster.yaml`, or `enum.yaml` files SHALL exist in `api/openapi/`
- **AND** all business domain CRUD operations MUST use the GraphQL endpoint at `/api/design/graphql`
