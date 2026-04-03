# Implementation Tasks: Implement Casdoor Runtime Authentication

## Task Organization

Tasks are organized by capability and sequenced to deliver incremental, testable progress. Each task should:
- Be completable in 2-4 hours
- Have clear validation criteria
- Produce user-visible or testable output

## Dependencies

```
Phase 1 (Foundation - Parallel)
├── [1.1] Database Schema
├── [1.2] Domain Models
└── [1.3] Configuration Setup

Phase 2 (Core Interfaces - Sequential)
├── [2.1] AuthProvider Interface (depends on 1.2)
├── [2.2] UnifiedClaims (depends on 1.2)
└── [2.3] ProviderRegistry (depends on 2.1, 2.2)

Phase 3 (Casdoor Implementation - Sequential)
├── [3.1] CasdoorProvider (depends on 2.1, 2.2)
└── [3.2] Casdoor Unit Tests (depends on 3.1)

Phase 4 (Middleware - Sequential)
├── [4.1] JWT Middleware (depends on 2.3, 3.1)
└── [4.2] Middleware Unit Tests (depends on 4.1)

Phase 5 (Config Management - Parallel with Phase 4)
├── [5.1] Auth Config Repository (depends on 1.1)
├── [5.2] Auth Config Application Service (depends on 5.1)
└── [5.3] Auth Config HTTP API (depends on 5.2)

Phase 6 (Integration)
├── [6.1] Wire Middleware to GraphQL Routes (depends on 4.1, 5.2)
├── [6.2] Integration Tests (depends on 6.1)
└── [6.3] Casdoor Deployment Guide (parallel)

Phase 7 (Documentation & Rollout)
├── [7.1] API Documentation
├── [7.2] Migration Guide
└── [7.3] Troubleshooting Guide
```

---

## Phase 1: Foundation (Parallel)

### [1.1] Database Schema for Auth Configuration

**Objective**: Create database table to store project-level auth configuration

**Acceptance Criteria**:
- [ ] SQL migration file created: `db/schema/mysql/04_auth.sql`
- [ ] Table `project_auth_configs` has columns: id, project_id, provider, enabled, config (JSON), created_at, updated_at
- [ ] Unique constraint on `project_id`
- [ ] Foreign key to `projects` table with CASCADE DELETE
- [ ] Index on `project_id`
- [ ] Migration runs successfully on fresh database
- [ ] Migration is idempotent (can run multiple times safely)

**Files**:
- `db/schema/mysql/04_auth.sql`

**Validation**:
```bash
# Apply migration
mysql -u root -p modelcraft < db/schema/mysql/04_auth.sql

# Verify table structure
mysql -u root -p modelcraft -e "DESCRIBE project_auth_configs;"

# Test idempotency
mysql -u root -p modelcraft < db/schema/mysql/04_auth.sql
```

**Spec**: `project-auth-config/spec.md` - Requirement: Database Schema

---

### [1.2] Domain Models for Authentication

**Objective**: Define core authentication domain entities and value objects

**Acceptance Criteria**:
- [ ] `UnifiedClaims` struct with fields: UserID, Username, Email, Organization, Roles, ExpiresAt, Raw
- [ ] `ProjectAuthConfig` entity with validation rules
- [ ] `AuthProvider` type enum: Casdoor, Keycloak, OIDC
- [ ] Config validation methods (e.g., ValidateCasdoorConfig)
- [ ] Unit tests for validation logic

**Files**:
- `internal/domain/auth/unified_claims.go`
- `internal/domain/auth/project_auth_config.go`
- `internal/domain/auth/provider_type.go`
- `internal/domain/auth/unified_claims_test.go`
- `internal/domain/auth/project_auth_config_test.go`

**Validation**:
```bash
# Run unit tests
go test ./internal/domain/auth/... -v

# Should test:
# - UnifiedClaims validation (required fields)
# - ProjectAuthConfig validation (JSON schema)
# - Invalid provider type rejection
```

**Spec**: `auth-provider-interface/spec.md` - Requirement: UnifiedClaims Structure

---

### [1.3] Configuration Setup

**Objective**: Add Casdoor default configuration to config.yaml and .env

**Acceptance Criteria**:
- [ ] `configs/config.yaml` has `auth` section with runtime authentication settings
- [ ] `.env.example` includes CASDOOR_* variables (CASDOOR_ENDPOINT, CASDOOR_CLIENT_ID, etc.)
- [ ] Config loading logic reads auth configuration from environment
- [ ] Validation: Config fails fast if JWT signing method is invalid

**Files**:
- `configs/config.yaml`
- `.env.example`
- `pkg/config/config.go` (if auth struct needs to be added)

**Validation**:
```bash
# Check config loads successfully
go run cmd/server/main.go --config configs/config.yaml --dry-run

# Verify environment variables override defaults
export CASDOOR_ENDPOINT="https://test.casdoor.com"
go run cmd/server/main.go --config configs/config.yaml --dry-run | grep "test.casdoor.com"
```

**Spec**: `project-auth-config/spec.md` - Requirement: Configuration Management

---

## Phase 2: Core Interfaces (Sequential)

### [2.1] AuthProvider Interface

**Objective**: Define the core abstraction for authentication providers

**Acceptance Criteria**:
- [ ] `AuthProvider` interface with methods: GetPublicKey, GetSigningMethod, ParseClaims, Type
- [ ] Interface documentation explains each method's contract
- [ ] Example mock implementation for testing

**Files**:
- `internal/domain/auth/provider.go`
- `internal/domain/auth/provider_mock.go` (for testing)

**Validation**:
```bash
# Compile check
go build ./internal/domain/auth/...

# Interface should be usable by mock
go test ./internal/domain/auth/... -v
```

**Spec**: `auth-provider-interface/spec.md` - Requirement: AuthProvider Interface Definition

---

### [2.2] UnifiedClaims Helper Functions

**Objective**: Add utility functions for working with UnifiedClaims

**Acceptance Criteria**:
- [ ] `ParseStandardClaims` function extracts exp, iss, sub from jwt.MapClaims
- [ ] `Validate` method checks required fields and expiry
- [ ] `IsExpired` helper method
- [ ] Unit tests cover edge cases (missing fields, expired tokens, invalid format)

**Files**:
- `internal/domain/auth/unified_claims.go` (extend from 1.2)
- `internal/domain/auth/unified_claims_test.go`

**Validation**:
```bash
go test ./internal/domain/auth/... -v -run TestUnifiedClaims
```

**Spec**: `auth-provider-interface/spec.md` - Requirement: Claims Validation

---

### [2.3] ProviderRegistry Implementation

**Objective**: Create registry for dynamic provider selection based on project

**Acceptance Criteria**:
- [ ] `ProviderRegistry` struct with provider cache (map[string]AuthProvider)
- [ ] `GetProvider(projectID)` method with caching
- [ ] `createProvider(config)` factory method
- [ ] Default provider fallback when project has no config
- [ ] Thread-safe cache access (use sync.RWMutex)
- [ ] Unit tests with mock repository

**Files**:
- `internal/app/auth/provider_registry.go`
- `internal/app/auth/provider_registry_test.go`

**Validation**:
```bash
go test ./internal/app/auth/... -v -run TestProviderRegistry

# Should test:
# - Cache hit/miss behavior
# - Default provider fallback
# - Concurrent access safety
```

**Spec**: `auth-provider-interface/spec.md` - Requirement: Provider Registry

---

## Phase 3: Casdoor Implementation (Sequential)

### [3.1] CasdoorProvider Implementation

**Objective**: Implement AuthProvider interface for Casdoor

**Acceptance Criteria**:
- [ ] `CasdoorProvider` struct with fields: endpoint, clientID, organization, certificate
- [ ] `GetPublicKey` parses PEM certificate and caches RSA public key
- [ ] `GetSigningMethod` returns "RS256"
- [ ] `ParseClaims` extracts Casdoor-specific fields (owner, name, sub, email)
- [ ] `Type` returns "casdoor"
- [ ] Error handling for invalid certificate format

**Files**:
- `internal/infrastructure/auth/casdoor_provider.go`
- `internal/infrastructure/auth/casdoor_config.go`

**Validation**:
```bash
# Compile check
go build ./internal/infrastructure/auth/...

# Manual test with sample certificate
go run examples/casdoor_test/main.go
```

**Spec**: `casdoor-provider/spec.md` - Requirement: Casdoor Provider Implementation

---

### [3.2] Casdoor Provider Unit Tests

**Objective**: Comprehensive unit tests for CasdoorProvider

**Acceptance Criteria**:
- [ ] Test valid certificate parsing
- [ ] Test invalid certificate rejection
- [ ] Test claims parsing with complete Casdoor JWT payload
- [ ] Test claims parsing with missing optional fields
- [ ] Test public key caching (second call doesn't re-parse)
- [ ] Mock jwt.MapClaims for testing

**Files**:
- `internal/infrastructure/auth/casdoor_provider_test.go`
- `internal/infrastructure/auth/testdata/test_certificate.pem`

**Validation**:
```bash
go test ./internal/infrastructure/auth/... -v -cover

# Coverage should be >80%
go test ./internal/infrastructure/auth/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

**Spec**: `casdoor-provider/spec.md` - All scenarios

---

## Phase 4: Middleware (Sequential)

### [4.1] JWT Authentication Middleware

**Objective**: Generic middleware that uses ProviderRegistry to validate JWT

**Acceptance Criteria**:
- [ ] `JWTAuthMiddleware(registry)` returns gin.HandlerFunc
- [ ] Extracts projectID from route param (`:projectId`)
- [ ] Extracts Bearer token from Authorization header
- [ ] Gets provider from registry
- [ ] Validates JWT signature using provider's public key
- [ ] Validates signing method matches provider's expectation
- [ ] Parses claims using provider
- [ ] Stores UnifiedClaims in gin.Context with key "user"
- [ ] Returns 401 with clear error message for invalid token
- [ ] Returns 500 for provider configuration errors

**Files**:
- `internal/middleware/jwt_auth.go`
- `internal/middleware/context.go` (helper to get user from context)

**Validation**:
```bash
# Compile check
go build ./internal/middleware/...

# Will be validated in integration tests (4.2)
```

**Spec**: `jwt-middleware/spec.md` - Requirement: JWT Validation Middleware

---

### [4.2] Middleware Unit Tests

**Objective**: Unit tests for JWT middleware using mocks

**Acceptance Criteria**:
- [ ] Test successful JWT validation flow
- [ ] Test missing Authorization header (401)
- [ ] Test invalid Bearer token format (401)
- [ ] Test expired token (401)
- [ ] Test invalid signature (401)
- [ ] Test wrong signing method (401)
- [ ] Test project without auth config (fallback to default)
- [ ] Test provider configuration error (500)
- [ ] Mock ProviderRegistry and AuthProvider

**Files**:
- `internal/middleware/jwt_auth_test.go`

**Validation**:
```bash
go test ./internal/middleware/... -v -cover

# All scenarios should pass
go test ./internal/middleware/... -v -run TestJWTAuth
```

**Spec**: `jwt-middleware/spec.md` - All scenarios

---

## Phase 5: Config Management (Parallel with Phase 4)

### [5.1] Auth Config Repository

**Objective**: Data access layer for project auth configuration

**Acceptance Criteria**:
- [ ] `ProjectAuthConfigRepository` interface with methods: GetByProjectID, Create, Update, Delete
- [ ] sqlc implementation: `GormProjectAuthConfigRepository`
- [ ] sqlc model maps to `project_auth_configs` table
- [ ] JSON config field properly serialized/deserialized
- [ ] Error handling for not found, duplicate, constraint violations

**Files**:
- `internal/domain/auth/repository.go` (interface)
- `internal/infrastructure/repository/project_auth_config_repository.go` (impl)
- `internal/infrastructure/repository/project_auth_config_model.go` (sqlc model)

**Validation**:
```bash
# Integration test with real database
go test ./internal/infrastructure/repository/... -v -run TestProjectAuthConfig

# Should test CRUD operations
```

**Spec**: `project-auth-config/spec.md` - Requirement: Repository Interface

---

### [5.2] Auth Config Application Service

**Objective**: Business logic for managing auth configurations

**Acceptance Criteria**:
- [ ] `AuthConfigService` with methods: GetConfig, CreateOrUpdateConfig, DeleteConfig
- [ ] Validates JSON config against provider schema (e.g., Casdoor requires endpoint, clientID)
- [ ] Validates certificate format for Casdoor
- [ ] Invalidates ProviderRegistry cache on config update
- [ ] Returns typed errors (NOT_FOUND, PARAM_INVALID, CONFLICT)

**Files**:
- `internal/app/auth/auth_config_service.go`
- `internal/app/auth/auth_config_service_test.go`

**Validation**:
```bash
go test ./internal/app/auth/... -v -run TestAuthConfigService

# Should test:
# - Valid config creation
# - Invalid config rejection
# - Cache invalidation
```

**Spec**: `project-auth-config/spec.md` - Requirement: Config Validation

---

### [5.3] Auth Config HTTP API

**Objective**: REST endpoints for managing auth configuration

**Acceptance Criteria**:
- [ ] `POST /api/projects/:projectId/auth-config` - Create/update config
- [ ] `GET /api/projects/:projectId/auth-config` - Get config
- [ ] `DELETE /api/projects/:projectId/auth-config` - Delete config
- [ ] Request/response DTOs defined
- [ ] Input validation (e.g., provider must be valid enum)
- [ ] Returns appropriate HTTP status codes (200, 201, 400, 404, 500)
- [ ] Sensitive fields (clientSecret, certificate) masked in GET response

**Files**:
- `internal/interfaces/http/handlers/auth_config_handler.go`
- `internal/interfaces/http/requests/auth_config_request.go`
- `internal/interfaces/http/dtos/auth_config_dto.go`

**Validation**:
```bash
# Manual testing with curl
curl -X POST http://localhost:8080/api/projects/test-project/auth-config \
  -H "Content-Type: application/json" \
  -d '{"provider":"casdoor","config":{...}}'

# Integration tests (6.2 will cover this)
```

**Spec**: `project-auth-config/spec.md` - Requirement: HTTP API

---

## Phase 6: Integration (Sequential)

### [6.1] Wire Middleware to GraphQL Routes

**Objective**: Integrate JWT middleware into server startup

**Acceptance Criteria**:
- [ ] Initialize ProviderRegistry with auth config repository
- [ ] Initialize default Casdoor provider from config
- [ ] Apply JWTAuthMiddleware to `/graphql/*` routes
- [ ] Design-time routes (`/api/design/*`) remain unauthenticated
- [ ] Health check endpoint remains unauthenticated
- [ ] Feature flag: `auth.runtime.required` (default true, can disable for testing)
- [ ] Logs provider initialization status on startup

**Files**:
- `cmd/server/main.go`
- `cmd/server/auth.go` (new file for auth setup logic)

**Validation**:
```bash
# Start server with auth enabled
go run cmd/server/main.go

# Verify middleware is applied
curl http://localhost:8080/graphql/default/test -v
# Should return 401 Unauthorized

# Verify design API still works
curl http://localhost:8080/api/design/models
# Should return 200 OK
```

**Spec**: `jwt-middleware/spec.md` - Requirement: Route Integration

---

### [6.2] Integration Tests

**Objective**: End-to-end tests with real JWT tokens and Casdoor

**Acceptance Criteria**:
- [ ] Test fixture: Mock Casdoor JWT generator (using test RSA key pair)
- [ ] Test: Valid JWT allows GraphQL query execution
- [ ] Test: Invalid JWT rejected with 401
- [ ] Test: Expired JWT rejected with 401
- [ ] Test: Missing Authorization header rejected
- [ ] Test: Multi-tenant isolation (user can't access wrong project)
- [ ] Test: Config change invalidates cache (new provider picked up)
- [ ] Python pytest suite in `tests/automated/test_auth.py`

**Files**:
- `tests/automated/test_auth.py`
- `tests/automated/fixtures/jwt_generator.py`
- `tests/automated/fixtures/test_rsa_key.pem`

**Validation**:
```bash
# Run integration tests
task auto-test

# Or run specific auth tests
cd tests
pytest automated/test_auth.py -v
```

**Spec**: All specs - Integration scenarios

---

### [6.3] Casdoor Deployment Guide

**Objective**: Update documentation with deployment instructions

**Acceptance Criteria**:
- [ ] Docker Compose file for Casdoor (`docker-compose.casdoor.yml`)
- [ ] Configuration guide for creating Organization and Application
- [ ] Guide for obtaining certificate and client credentials
- [ ] Environment variable setup instructions
- [ ] Troubleshooting section (common errors and fixes)

**Files**:
- `docs/casdoor-setup.md`
- `docker-compose.casdoor.yml`
- `docs/runtime-auth-design.md` (update deployment section)

**Validation**:
```bash
# Start Casdoor
docker-compose -f docker-compose.casdoor.yml up -d

# Verify Casdoor is running
curl http://localhost:8000/health

# Follow setup guide to create application
```

**Spec**: N/A (documentation task)

---

## Phase 7: Documentation & Rollout (Parallel)

### [7.1] API Documentation

**Objective**: Document authentication requirements for API consumers

**Acceptance Criteria**:
- [ ] Update `docs/03-runtime/runtime-api.md` with authentication section
- [ ] Example: How to obtain JWT from Casdoor
- [ ] Example: How to use JWT in GraphQL requests
- [ ] Error codes and troubleshooting (401 Unauthorized scenarios)
- [ ] Multi-tenant access control explanation

**Files**:
- `docs/03-runtime/runtime-api.md`
- `docs/03-runtime/authentication.md` (new)

**Validation**:
- [ ] Documentation reviewed by team
- [ ] Examples tested manually

**Spec**: N/A (documentation task)

---

### [7.2] Migration Guide for Existing Clients

**Objective**: Help existing runtime API users migrate to authenticated endpoints

**Acceptance Criteria**:
- [ ] Guide explains breaking change
- [ ] Step-by-step migration instructions
- [ ] Code examples in multiple languages (curl, JavaScript, Python)
- [ ] Temporary workaround using feature flag
- [ ] Timeline for migration (grace period)

**Files**:
- `docs/migration/runtime-auth-migration.md`

**Validation**:
- [ ] Guide reviewed by team
- [ ] Examples tested

**Spec**: N/A (documentation task)

---

### [7.3] Troubleshooting Guide

**Objective**: Common issues and resolutions

**Acceptance Criteria**:
- [ ] Symptoms: 401 Unauthorized - Common causes and fixes
- [ ] Symptoms: 500 Internal Server Error - Provider config issues
- [ ] Symptoms: Token expired - How to refresh
- [ ] Symptoms: Wrong organization claim - Multi-tenant isolation explanation
- [ ] Debug tools: JWT decoder, certificate validator
- [ ] Logs to check for diagnostics

**Files**:
- `docs/troubleshooting/auth-issues.md`

**Validation**:
- [ ] Guide covers issues found during integration testing
- [ ] Team validates usefulness

**Spec**: N/A (documentation task)

---

## Summary

**Total Tasks**: 23
**Estimated Effort**: 7-10 days (as per proposal)

**Task Distribution**:
- Phase 1 (Foundation): 3 tasks - ~1 day (parallel)
- Phase 2 (Core Interfaces): 3 tasks - ~1 day (sequential)
- Phase 3 (Casdoor): 2 tasks - ~1 day (sequential)
- Phase 4 (Middleware): 2 tasks - ~1 day (sequential)
- Phase 5 (Config Management): 3 tasks - ~1.5 days (parallel with Phase 4)
- Phase 6 (Integration): 3 tasks - ~1.5 days (sequential)
- Phase 7 (Documentation): 3 tasks - ~1 day (parallel)
- Buffer: ~1 day for unexpected issues

**Critical Path**: Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 6

**Parallelizable Work**:
- Phase 1 tasks can all run in parallel
- Phase 5 can run parallel with Phase 4
- Phase 7 can start early (parallel with implementation)

**Validation Strategy**:
- Unit tests after each component (coverage >80%)
- Integration tests after Phase 6.1
- Manual testing with real Casdoor instance
- Pytest automation for regression suite
