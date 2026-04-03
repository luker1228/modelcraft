# Spec: Project Authentication Configuration

## Overview

This specification defines the data model, repository, and HTTP API for managing per-project authentication configurations. Each project can configure its own authentication provider (Casdoor, Keycloak, OIDC), enabling multi-tenant scenarios where different organizations use different identity systems.

## Context

ModelCraft's project-based architecture allows logical isolation of resources. Authentication configuration follows the same pattern: each project can independently configure how users authenticate when accessing its runtime GraphQL API. This enables SaaS scenarios where Project A uses enterprise LDAP (via Casdoor) while Project B uses social login (via generic OIDC).

## ADDED Requirements

### Requirement: Database Schema

The system MUST persist project authentication configurations in the platform database.

#### Scenario: Schema defines project_auth_configs table

**Given** the platform database initialization
**When** the migration script `04_auth.sql` is executed
**Then** the table `project_auth_configs` is created with columns:
- `id` BIGINT PRIMARY KEY AUTO_INCREMENT
- `project_id` VARCHAR(64) NOT NULL UNIQUE
- `provider` VARCHAR(32) NOT NULL (enum: 'casdoor', 'keycloak', 'oidc')
- `enabled` BOOLEAN DEFAULT true
- `config` JSON NOT NULL (provider-specific configuration)
- `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
- `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

#### Scenario: Schema enforces unique project constraint

**Given** a project with ID "ecommerce" already has an auth config
**When** attempting to insert another auth config for "ecommerce"
**Then** the database rejects the insert with unique constraint violation
**And** only one auth config per project is allowed

#### Scenario: Schema cascades deletion with project

**Given** a project with ID "ecommerce" has an auth config
**When** the project "ecommerce" is deleted from projects table
**Then** the auth config is automatically deleted (FOREIGN KEY ON DELETE CASCADE)

#### Scenario: Schema indexes project_id for fast lookups

**Given** the project_auth_configs table
**When** querying by project_id
**Then** an index on `project_id` column ensures O(log n) lookup performance

---

### Requirement: Domain Model

The system MUST define a domain entity for project authentication configuration.

#### Scenario: ProjectAuthConfig entity structure

**Given** the domain layer
**When** defining the ProjectAuthConfig entity
**Then** the entity has fields:
```go
type ProjectAuthConfig struct {
    ID         int64
    ProjectID  string
    Provider   ProviderType  // enum: Casdoor, Keycloak, OIDC
    Enabled    bool
    Config     map[string]interface{}  // JSON
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

#### Scenario: Provider type enum values

**Given** the ProviderType enum
**When** defining supported providers
**Then** the enum includes:
- `ProviderCasdoor` = "casdoor"
- `ProviderKeycloak` = "keycloak"
- `ProviderOIDC` = "oidc"

#### Scenario: Config field stores provider-specific JSON

**Given** a ProjectAuthConfig for Casdoor
**When** setting the Config field
**Then** the config contains:
```json
{
  "endpoint": "https://casdoor.example.com",
  "client_id": "abc123",
  "client_secret": "secret",
  "organization": "tenant1",
  "application": "modelcraft",
  "certificate": "-----BEGIN CERTIFICATE-----..."
}
```

---

### Requirement: Repository Interface

The system MUST define a repository interface for data access.

#### Scenario: Repository interface methods

**Given** the domain layer repository interface
**When** defining ProjectAuthConfigRepository
**Then** the interface declares methods:
```go
type ProjectAuthConfigRepository interface {
    GetByProjectID(ctx context.Context, projectID string) (*ProjectAuthConfig, error)
    Create(ctx context.Context, config *ProjectAuthConfig) error
    Update(ctx context.Context, config *ProjectAuthConfig) error
    Delete(ctx context.Context, projectID string) error
}
```

#### Scenario: GetByProjectID returns nil for non-existent project

**Given** a project "unknown" has no auth config
**When** calling `repo.GetByProjectID(ctx, "unknown")`
**Then** the method returns `(nil, nil)` (not an error, absence is a valid state)

#### Scenario: Create returns error for duplicate project_id

**Given** a project "ecommerce" already has an auth config
**When** calling `repo.Create(ctx, newConfig)` with projectID = "ecommerce"
**Then** the method returns error with type CONFLICT
**And** the error message is "auth config already exists for project: ecommerce"

#### Scenario: Update returns error for non-existent project

**Given** a project "unknown" has no auth config
**When** calling `repo.Update(ctx, config)` with projectID = "unknown"
**Then** the method returns error with type NOT_FOUND
**And** the error message is "auth config not found for project: unknown"

#### Scenario: Delete is idempotent

**Given** a project "test" has no auth config
**When** calling `repo.Delete(ctx, "test")`
**Then** the method returns `nil` (no error)
**And** subsequent deletes also return `nil`

---

### Requirement: Configuration Validation

The system MUST validate configuration before persisting.

#### Scenario: Validate Casdoor configuration completeness

**Given** a ProjectAuthConfig with provider = Casdoor
**When** calling `ValidateCasdoorConfig(config.Config)`
**Then** the validator checks for required fields:
- `endpoint` (non-empty URL)
- `client_id` (non-empty string)
- `client_secret` (non-empty string)
- `organization` (non-empty string)
- `certificate` (valid PEM format)
**And** returns error if any required field is missing or invalid

#### Scenario: Validate certificate PEM format

**Given** a Casdoor config with certificate field
**When** calling `ValidateCasdoorConfig(config.Config)`
**Then** the validator uses `pem.Decode()` to check PEM format
**And** returns error if PEM decoding fails
**And** error message is "certificate is not valid PEM format"

#### Scenario: Validate endpoint URL format

**Given** a Casdoor config with endpoint = "invalid-url"
**When** calling `ValidateCasdoorConfig(config.Config)`
**Then** the validator uses `url.Parse()` to check URL format
**And** returns error if URL parsing fails
**And** error message is "endpoint must be a valid URL"

#### Scenario: Validate OIDC configuration

**Given** a ProjectAuthConfig with provider = OIDC
**When** calling `ValidateOIDCConfig(config.Config)`
**Then** the validator checks for required fields:
- `issuer` (non-empty URL)
- `jwks_uri` (non-empty URL)
- `client_id` (non-empty string)
**And** returns error if any required field is missing

---

### Requirement: Application Service

The system MUST provide an application service for managing auth configurations.

#### Scenario: Service creates or updates configuration

**Given** an AuthConfigService initialized with repository
**When** calling `service.CreateOrUpdateConfig(ctx, projectID, input)`
**Then** the service:
1. Validates the input configuration
2. Checks if config exists for project
3. If exists, calls `repo.Update()`
4. If not exists, calls `repo.Create()`
5. Invalidates provider cache in registry
6. Returns the saved config

#### Scenario: Service validates project existence

**Given** an AuthConfigService with access to ProjectRepository
**When** calling `service.CreateOrUpdateConfig(ctx, "unknown", input)`
**Then** the service first checks if project "unknown" exists
**And** if project doesn't exist, returns error type NOT_FOUND.PROJECT
**And** error message is "project not found: unknown"

#### Scenario: Service masks sensitive fields in responses

**Given** an AuthConfigService retrieving a Casdoor config
**When** calling `service.GetConfig(ctx, projectID)`
**Then** the returned DTO has sensitive fields masked:
- `client_secret` → "***"
- `certificate` → "-----BEGIN CERTIFICATE-----\n***\n-----END CERTIFICATE-----"
**And** endpoint, client_id, organization remain visible

#### Scenario: Service invalidates provider cache on update

**Given** a ProviderRegistry with cached provider for project "ecommerce"
**When** calling `service.CreateOrUpdateConfig(ctx, "ecommerce", newInput)`
**Then** the service calls `registry.InvalidateCache("ecommerce")`
**And** next GraphQL request for "ecommerce" loads fresh provider from new config

#### Scenario: Service invalidates cache on delete

**Given** a ProviderRegistry with cached provider for project "test"
**When** calling `service.DeleteConfig(ctx, "test")`
**Then** the service calls `registry.InvalidateCache("test")`
**And** next GraphQL request for "test" falls back to default provider

---

### Requirement: HTTP API

The system MUST expose REST endpoints for managing authentication configurations.

#### Scenario: Create or update auth config via POST

**Given** an HTTP handler for auth config
**When** client sends:
```http
POST /api/projects/ecommerce/auth-config
Content-Type: application/json
Authorization: Bearer {admin-token}

{
  "provider": "casdoor",
  "enabled": true,
  "config": {
    "endpoint": "https://casdoor.example.com",
    "client_id": "abc123",
    "client_secret": "secret",
    "organization": "tenant1",
    "certificate": "-----BEGIN CERTIFICATE-----..."
  }
}
```
**Then** the handler:
1. Extracts projectID = "ecommerce" from path
2. Parses and validates request body
3. Calls service.CreateOrUpdateConfig()
4. Returns 200 OK (update) or 201 Created (create) with masked config

#### Scenario: Get auth config via GET

**Given** an HTTP handler for auth config
**When** client sends:
```http
GET /api/projects/ecommerce/auth-config
Authorization: Bearer {admin-token}
```
**Then** the handler:
1. Extracts projectID from path
2. Calls service.GetConfig(projectID)
3. Returns 200 OK with masked config JSON
4. If config doesn't exist, returns 404 Not Found

#### Scenario: Delete auth config via DELETE

**Given** an HTTP handler for auth config
**When** client sends:
```http
DELETE /api/projects/ecommerce/auth-config
Authorization: Bearer {admin-token}
```
**Then** the handler:
1. Extracts projectID from path
2. Calls service.DeleteConfig(projectID)
3. Returns 204 No Content on success
4. Returns 404 Not Found if config doesn't exist

#### Scenario: Return validation errors with 400 status

**Given** a client sends invalid configuration (missing required field)
**When** the handler validates the request
**Then** the handler returns:
- HTTP status 400 Bad Request
- JSON body with error details:
```json
{
  "error": "validation failed",
  "details": {
    "field": "endpoint",
    "issue": "endpoint is required"
  }
}
```

#### Scenario: Return 404 for non-existent project

**Given** a client sends request for project "unknown"
**When** the service checks project existence
**Then** the handler returns:
- HTTP status 404 Not Found
- JSON body: `{"error": "project not found: unknown"}`

---

### Requirement: Security and Access Control

The system MUST protect auth configuration endpoints.

#### Scenario: Endpoints require admin authentication

**Note**: This scenario documents future behavior (Phase 2 - Design-Time API Auth)

**Given** an unauthenticated client
**When** client attempts to access `/api/projects/*/auth-config`
**Then** the system returns 401 Unauthorized
**And** error message is "authentication required"

#### Scenario: Endpoints require admin role

**Note**: This scenario documents future behavior (Phase 2 - Authorization/RBAC)

**Given** an authenticated user with role = "viewer"
**When** client attempts to POST/DELETE `/api/projects/*/auth-config`
**Then** the system returns 403 Forbidden
**And** error message is "admin privileges required"

#### Scenario: Audit log records config changes

**Note**: This scenario documents future behavior (Phase 3 - Audit Logging)

**Given** an admin user updates auth config for project "ecommerce"
**When** the update succeeds
**Then** the system logs audit entry:
- Action: "auth_config_updated"
- User: admin user ID
- Project: "ecommerce"
- Timestamp: current time
- Old/new config diff (with secrets masked)

---

### Requirement: Default Configuration

The system MUST support a default provider when projects have no specific config.

#### Scenario: Server loads default Casdoor provider from config.yaml

**Given** the server configuration file `configs/config.yaml` contains:
```yaml
auth:
  default_provider: "casdoor"
  casdoor:
    endpoint: "${CASDOOR_ENDPOINT}"
    client_id: "${CASDOOR_CLIENT_ID}"
    client_secret: "${CASDOOR_CLIENT_SECRET}"
    organization: "${CASDOOR_ORGANIZATION}"
    certificate: "${CASDOOR_CERTIFICATE}"
```
**When** the server initializes ProviderRegistry
**Then** the registry creates a default CasdoorProvider from these settings
**And** uses it as fallback for projects without auth config

#### Scenario: Registry returns default provider for unconfigured project

**Given** a project "test" has no entry in project_auth_configs table
**And** a default Casdoor provider is configured
**When** the middleware calls `registry.GetProvider("test")`
**Then** the registry returns the default CasdoorProvider
**And** logs at INFO level: "Using default provider for project: test"

#### Scenario: Registry returns nil if no default and no project config

**Given** a project "test" has no entry in project_auth_configs table
**And** no default provider is configured
**When** the middleware calls `registry.GetProvider("test")`
**Then** the registry returns `nil`
**And** the middleware rejects the request with 500 Internal Server Error

---

## Design Notes

**Configuration Storage**: The `config` JSON field allows flexible provider-specific settings without schema changes. Each provider type has its own expected JSON structure.

**Sensitive Data**: Client secrets and certificates are stored in the database. In production, consider:
- Encrypting the `config` JSON field at rest
- Using secrets management service (Vault, AWS Secrets Manager)
- Rotating secrets regularly

**Cache Invalidation**: When auth config is updated, the ProviderRegistry cache must be invalidated to pick up new settings. This is critical for multi-instance deployments (consider Redis pub/sub for cross-instance invalidation).

**Backward Compatibility**: Projects without auth config fall back to default provider, ensuring existing projects continue to work during migration.

---

## Dependencies

**Database**:
- `projects` table (foreign key reference)

**Internal Dependencies**:
- ProviderRegistry (for cache invalidation)
- Project domain model (for validation)

---

## Testing Strategy

**Unit Tests**:
- Repository CRUD operations with mock database
- Validation functions (config structure, PEM format, URLs)
- Service logic (cache invalidation, error handling)
- HTTP handler request/response mapping

**Integration Tests**:
- Full flow: POST config → cache invalidation → GET config
- Database constraints (unique project_id, foreign key cascade)
- Concurrent updates to same project (pessimistic locking)

**Example Integration Test**:
```bash
# Create project
POST /api/projects {"id": "test", "title": "Test"}

# Create auth config
POST /api/projects/test/auth-config {...}

# Verify config persisted
GET /api/projects/test/auth-config

# Verify provider loaded
# (Make authenticated GraphQL request, should use new provider)

# Delete project
DELETE /api/projects/test

# Verify auth config also deleted (cascade)
GET /api/projects/test/auth-config  # Returns 404
```

---

## Open Questions

1. **Configuration Encryption**: Should the `config` JSON be encrypted at rest?
   - Concern: Certificates and secrets in plain text
   - Solution: Use database-level encryption or application-level encryption with KEK

2. **Configuration History**: Should the system track config change history?
   - Use case: Audit trail, rollback to previous config
   - Complexity: Requires separate history table

3. **Configuration Validation Timing**: Validate on save or on use?
   - Current: Validate on save (fail fast)
   - Alternative: Validate on first use (more flexible but harder to debug)

4. **Multi-Instance Cache Invalidation**: How to invalidate cache across multiple server instances?
   - Current: Only local cache invalidation
   - Solution: Redis pub/sub or database triggers to notify other instances
