# Specification: Dual-Token Authentication

## Overview

This specification defines the dual-token authentication flow for ModelCraft, separating external identity verification (Casdoor) from internal authorization (ModelCraft). The system exchanges an OAuth authorization code for a ModelCraft-signed JWT containing user identity and authorization data (roles and permissions).

## ADDED Requirements

### Requirement: Token Exchange with Enhanced Response

The system SHALL provide a token exchange endpoint that MUST accept an OAuth authorization code and return a ModelCraft-signed JWT along with complete user identity and authorization information.

**Acceptance Criteria**:
- Endpoint accepts POST request with authorization code
- Endpoint exchanges code with Casdoor for Casdoor JWT
- Endpoint validates Casdoor JWT and extracts user identity
- Endpoint queries ModelCraft database for user roles and permissions
- Endpoint generates ModelCraft JWT with user identity + authorization claims
- Endpoint returns enhanced response including accessToken, user info, organization, roles, and permissions
- Response enables clients to cache authorization data locally without additional queries

#### Scenario: Successful token exchange with roles and permissions

**Given**:
- A valid OAuth authorization code from Casdoor
- User exists in ModelCraft database with external_id matching Casdoor user
- User has 2 roles: "owner" (id=1) and "viewer" (id=2) in organization "modelcraft"
- Owner role has permissions: ["model:read", "model:write", "model:delete", "cluster:manage"]
- Viewer role has permissions: ["model:read"]

**When**:
- Client sends POST /api/auth/token with {"code": "valid-auth-code"}

**Then**:
- System exchanges code with Casdoor successfully
- System queries user roles from database (2 queries: user_roles + role_permissions per role)
- System generates ModelCraft JWT signed with JWT_SECRET
- System returns 200 OK with response:
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 3600,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "externalId": "casdoor-user-id",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "organization": {
    "name": "modelcraft"
  },
  "roles": [
    {"id": 1, "name": "owner", "displayName": "Owner"},
    {"id": 2, "name": "viewer", "displayName": "Viewer"}
  ],
  "permissions": [
    "model:read", "model:write", "model:delete", "cluster:manage"
  ]
}
```
- JWT payload contains claims:
```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "external_id": "casdoor-user-id",
  "name": "John Doe",
  "email": "john@example.com",
  "organization": "modelcraft",
  "roles": ["owner", "viewer"],
  "permissions": ["model:read", "model:write", "model:delete", "cluster:manage"],
  "iss": "modelcraft",
  "exp": 1234567890,
  "iat": 1234564290
}
```
- Permissions are deduplicated (no duplicates even though both roles have "model:read")

#### Scenario: Token exchange fails when user not found in ModelCraft database

**Given**:
- A valid OAuth authorization code from Casdoor
- Casdoor JWT contains user with external_id "new-user-id"
- No user exists in ModelCraft database with external_id "new-user-id"

**When**:
- Client sends POST /api/auth/token with {"code": "valid-auth-code"}

**Then**:
- System exchanges code with Casdoor successfully
- System queries users table by external_id = "new-user-id"
- Query returns no results
- System returns 404 Not Found with error:
```json
{
  "error": "user not found in ModelCraft database",
  "code": "NOT_FOUND.USER",
  "requestId": "..."
}
```
- No JWT is generated
- Client should direct user to organization setup flow

#### Scenario: Token exchange fails with invalid authorization code

**Given**:
- An invalid or expired OAuth authorization code

**When**:
- Client sends POST /api/auth/token with {"code": "invalid-code"}

**Then**:
- System attempts to exchange code with Casdoor
- Casdoor returns error (invalid_grant or similar)
- System returns 400 Bad Request with error:
```json
{
  "error": "invalid_request",
  "error_description": "Authorization code is invalid or expired",
  "code": "AUTH_CASDOOR_ERROR",
  "requestId": "..."
}
```

---

### Requirement: ModelCraft JWT Validation

The system SHALL validate ModelCraft-signed JWTs in authentication middleware and MUST inject claims into request context for use by authorization logic.

**Acceptance Criteria**:
- Middleware detects JWT issuer by parsing "iss" claim
- Middleware validates ModelCraft JWT signature using JWT_SECRET
- Middleware verifies JWT expiration timestamp
- Middleware extracts user identity and authorization claims
- Middleware injects claims into request context (user_id, email, name, organization, roles, permissions)
- Middleware returns 401 Unauthorized for invalid or expired tokens

#### Scenario: Valid ModelCraft JWT is accepted and claims are injected

**Given**:
- A valid ModelCraft JWT signed with correct secret
- JWT contains claims: user_id, name, email, organization, roles, permissions
- JWT is not expired (exp > current_time)
- JWT issuer is "modelcraft"

**When**:
- Client sends GraphQL request with Authorization header: "Bearer <modelcraft-jwt>"

**Then**:
- Middleware extracts Bearer token from Authorization header
- Middleware parses JWT to read "iss" claim → detects "modelcraft"
- Middleware validates JWT signature using JWT_SECRET from config
- Middleware verifies exp claim > current timestamp
- Middleware extracts claims and injects into Chi context:
  - c.Set("user_id", "550e8400-...")
  - c.Set("email", "john@example.com")
  - c.Set("name", "John Doe")
  - c.Set("organization", "modelcraft")
  - c.Set("permissions", []string{"model:read", "model:write", ...})
  - c.Set("roles", []string{"owner"})
- Middleware logs successful authentication: "JWT validated successfully for user: 550e8400-... (email: john@example.com, org: modelcraft)"
- Request proceeds to GraphQL handler

#### Scenario: Expired ModelCraft JWT is rejected

**Given**:
- A ModelCraft JWT with valid signature
- JWT exp claim is in the past (expired 1 hour ago)

**When**:
- Client sends request with Authorization header: "Bearer <expired-jwt>"

**Then**:
- Middleware extracts and parses JWT
- Middleware validates signature successfully
- Middleware checks exp claim → detects expiration
- Middleware returns 401 Unauthorized with error:
```json
{
  "error": "token has expired",
  "code": "AUTH_TOKEN_EXPIRED"
}
```
- Request is aborted, does not reach GraphQL handler

#### Scenario: JWT with invalid signature is rejected

**Given**:
- A JWT claiming to be from "modelcraft"
- JWT signature does not match (tampered or signed with wrong key)

**When**:
- Client sends request with Authorization header: "Bearer <invalid-jwt>"

**Then**:
- Middleware attempts to validate JWT signature
- Signature verification fails
- Middleware returns 401 Unauthorized with error:
```json
{
  "error": "invalid token signature",
  "code": "AUTH_INVALID_TOKEN"
}
```

---

### Requirement: Backward Compatibility with Casdoor JWT

The system SHALL support both Casdoor JWT and ModelCraft JWT during a migration period to enable gradual client migration without service disruption.

**Acceptance Criteria**:
- Middleware detects token type by parsing "iss" claim without validation
- Middleware validates Casdoor JWT using Casdoor public key if issuer is not "modelcraft"
- Middleware validates ModelCraft JWT using JWT_SECRET if issuer is "modelcraft"
- Both token types inject compatible claims into request context
- Configuration flags control which token types are accepted
- Logs distinguish between token types for monitoring

#### Scenario: Casdoor JWT is accepted during migration period

**Given**:
- Configuration flag `auth.design.accept_casdoor_jwt = true`
- A valid Casdoor JWT signed with Casdoor's private key
- JWT issuer claim is NOT "modelcraft" (e.g., "casdoor" or empty)

**When**:
- Client sends request with Authorization header: "Bearer <casdoor-jwt>"

**Then**:
- Middleware extracts Bearer token
- Middleware parses JWT (without validation) to read "iss" claim
- Middleware detects issuer is not "modelcraft"
- Middleware validates JWT using Casdoor public key (RSA)
- Middleware extracts Casdoor claims (sub, name, email, owner)
- Middleware queries database for user roles and permissions (fallback behavior)
- Middleware injects claims into context
- Log message includes: "JWT validated successfully (type: casdoor)"
- Request proceeds normally

#### Scenario: Casdoor JWT is rejected after migration

**Given**:
- Configuration flag `auth.design.accept_casdoor_jwt = false`
- A valid Casdoor JWT

**When**:
- Client sends request with Authorization header: "Bearer <casdoor-jwt>"

**Then**:
- Middleware detects issuer is not "modelcraft"
- Middleware checks configuration → accept_casdoor_jwt = false
- Middleware returns 401 Unauthorized with error:
```json
{
  "error": "Casdoor JWT is no longer supported, please use ModelCraft JWT",
  "code": "AUTH_UNSUPPORTED_TOKEN_TYPE"
}
```

---

### Requirement: Permission Check Using JWT Claims

The system SHALL perform permission checks using permissions from JWT claims when available, eliminating database queries for authorization decisions.

**Acceptance Criteria**:
- Permission middleware reads permissions from request context first
- If permissions exist in context, use them directly without database query
- If permissions not in context, fallback to database query (backward compatibility)
- Permission check logic supports both string format ("model:read") and structured format
- GraphQL permission directives use same logic
- Logs indicate whether permissions came from context or database

#### Scenario: Permission check succeeds using JWT claims (no database query)

**Given**:
- Request context contains permissions: ["model:read", "model:write", "model:delete"]
- GraphQL mutation requires permission "model:write"
- ModelCraft JWT middleware has already injected permissions into context

**When**:
- GraphQL resolver for createModel mutation is called
- Permission directive checks for "model:write"

**Then**:
- Permission middleware retrieves permissions from context via c.Get("permissions")
- Permissions found in context → no database query executed
- Middleware checks if "model:write" exists in permissions array
- Check succeeds (permission found)
- Log message: "Permission check allowed: user=..., org=..., action=model:write (source: jwt)"
- Resolver proceeds to execute mutation

#### Scenario: Permission check fails using JWT claims

**Given**:
- Request context contains permissions: ["model:read"]
- GraphQL mutation requires permission "model:delete"

**When**:
- GraphQL resolver for deleteModel mutation is called

**Then**:
- Permission middleware retrieves permissions from context
- Middleware checks if "model:delete" exists in permissions array
- Check fails (permission not found)
- Log message: "Permission check denied: user=..., org=..., action=model:delete (source: jwt)"
- Middleware returns 403 Forbidden:
```json
{
  "error": "insufficient permissions",
  "code": "OPERATION_DENIED.PERMISSION",
  "required": "model:delete"
}
```
- Resolver is not executed

#### Scenario: Permission check falls back to database when JWT claims missing

**Given**:
- Request context does NOT contain permissions (e.g., old Casdoor JWT without permissions)
- GraphQL query requires permission "model:read"

**When**:
- GraphQL resolver is called

**Then**:
- Permission middleware attempts to retrieve permissions from context
- Permissions not found in context
- Middleware falls back to database query:
  1. Queries user_roles table for user's roles in organization
  2. Queries role_permissions table for permissions of each role
- Middleware checks if "model:read" exists in queried permissions
- Check succeeds or fails based on database data
- Log message includes: "Permission check ... (source: database)"

---

### Requirement: Token Service for JWT Generation

The system SHALL provide a TokenService that MUST generate ModelCraft JWTs with user identity and authorization claims by querying user roles and permissions from the database.

**Acceptance Criteria**:
- TokenService accepts Casdoor claims as input
- TokenService queries user by external_id from users table
- TokenService queries user roles from user_roles table
- TokenService queries permissions for each role from role_permissions table
- TokenService aggregates and deduplicates permissions
- TokenService generates JWT signed with JWT_SECRET from configuration
- TokenService sets appropriate claims (sub, user_id, external_id, name, email, organization, roles, permissions)
- TokenService sets expiration based on configuration (default: 1 hour)
- TokenService sets issuer to "modelcraft"

#### Scenario: Token service generates JWT with aggregated permissions from multiple roles

**Given**:
- Casdoor claims: external_id="casdoor-123", name="Jane Doe", email="jane@example.com", owner="modelcraft"
- User exists in database: id="uuid-456", external_id="casdoor-123"
- User has 2 roles in "modelcraft" organization:
  - Role 1 (id=1, name="editor"): permissions ["model:read", "model:write"]
  - Role 2 (id=3, name="viewer"): permissions ["model:read", "cluster:read"]
- JWT_SECRET is "test-secret-key"
- JWT expiration is 3600 seconds

**When**:
- TokenService.IssueToken(ctx, casdoorClaims) is called

**Then**:
- Service queries users table: SELECT * FROM users WHERE external_id = "casdoor-123"
- Query returns user with id="uuid-456"
- Service queries user_roles table: SELECT * FROM user_roles WHERE user_id="uuid-456" AND org_name="modelcraft"
- Query returns 2 rows (role_id=1, role_id=3)
- Service queries role_permissions for role_id=1 → returns ["model:read", "model:write"]
- Service queries role_permissions for role_id=3 → returns ["model:read", "cluster:read"]
- Service aggregates permissions: ["model:read", "model:write", "cluster:read"]
- Service deduplicates: ["model:read", "model:write", "cluster:read"] (no duplicates)
- Service builds ModelCraftClaims:
  - sub: "uuid-456"
  - user_id: "uuid-456"
  - external_id: "casdoor-123"
  - name: "Jane Doe"
  - email: "jane@example.com"
  - organization: "modelcraft"
  - roles: ["editor", "viewer"]
  - permissions: ["model:read", "model:write", "cluster:read"]
  - iss: "modelcraft"
  - exp: current_time + 3600
  - iat: current_time
- Service signs JWT using HMAC-SHA256 with JWT_SECRET
- Service returns EnhancedTokenResponse with accessToken + user + organization + roles + permissions

#### Scenario: Token service returns error when user not found

**Given**:
- Casdoor claims: external_id="unknown-user"
- No user exists in database with external_id="unknown-user"

**When**:
- TokenService.IssueToken(ctx, casdoorClaims) is called

**Then**:
- Service queries users table: SELECT * FROM users WHERE external_id = "unknown-user"
- Query returns no rows
- Service returns error: bizerrors.New(bizerrors.NotFound, "user not found in ModelCraft database")
- No JWT is generated

---

### Requirement: Configuration for JWT Settings

The system SHALL support configuration of JWT signing secret, token expiration, and issuer through environment variables and config files.

**Acceptance Criteria**:
- JWT_SECRET environment variable is required (application fails to start if not set)
- JWT expiration is configurable via `jwt.expiration` in config.yaml (default: 3600 seconds)
- JWT issuer is configurable via `jwt.issuer` in config.yaml (default: "modelcraft")
- Configuration is validated at application startup
- Logs indicate JWT configuration on startup

#### Scenario: Application starts successfully with valid JWT configuration

**Given**:
- Environment variable JWT_SECRET is set to "my-secure-secret-key-32-bytes-long!!"
- config.yaml contains:
```yaml
jwt:
  expiration: 7200
  issuer: "modelcraft"
```

**When**:
- Application starts

**Then**:
- Application loads JWT_SECRET from environment
- Application loads jwt.expiration from config → 7200 seconds
- Application loads jwt.issuer from config → "modelcraft"
- Application validates JWT_SECRET is not empty
- Application logs: "JWT configuration loaded: expiration=7200s, issuer=modelcraft"
- Application starts successfully

#### Scenario: Application fails to start when JWT_SECRET is missing

**Given**:
- Environment variable JWT_SECRET is not set
- config.yaml contains jwt configuration

**When**:
- Application starts

**Then**:
- Application attempts to load JWT_SECRET
- JWT_SECRET is empty or not found
- Application logs error: "JWT_SECRET environment variable is required but not set"
- Application exits with error code 1
- Application does not start

---

### Requirement: Integration Test Support for Dual-Token Flow

The system SHALL provide test utilities that enable integration tests to obtain and use ModelCraft JWT tokens for authenticated API calls.

**Acceptance Criteria**:
- Test utilities support exchanging test credentials for ModelCraft JWT
- Integration test fixtures automatically use ModelCraft JWT by default
- Tests can validate enhanced token response structure
- Tests can verify permission checks use JWT claims (no database queries)
- Backward compatibility with existing test structure maintained

#### Scenario: Integration tests obtain ModelCraft JWT via token exchange

**Given**:
- Integration test environment is running
- Test user exists in database with owner role
- Test credentials configured (CASDOOR_TEST_USERNAME, CASDOOR_TEST_PASSWORD)

**When**:
- Test fixture `auth_token` is requested
- Fixture calls `get_modelcraft_token(test_config)`

**Then**:
- Function obtains Casdoor JWT using password flow
- Function calls `/api/auth/token` to exchange for ModelCraft JWT
- Function parses response and extracts `accessToken`
- Function validates JWT has issuer "modelcraft"
- Function returns ModelCraft JWT as string
- Subsequent tests use ModelCraft JWT for API calls
- All authenticated API requests succeed

#### Scenario: Test validates enhanced token response structure

**Given**:
- Test has Casdoor JWT from password flow
- Test user has roles and permissions configured in database

**When**:
- Test calls POST /api/auth/token with Casdoor JWT
- Test receives response

**Then**:
- Response has status code 200
- Response body contains:
  - `accessToken` (string, non-empty)
  - `tokenType` (value: "Bearer")
  - `expiresIn` (integer)
  - `user` object with: id, externalId, name, email
  - `organization` object with: name
  - `roles` array with objects containing: id, name, displayName
  - `permissions` array of strings (e.g., ["model:read", "model:write"])
- Test parses JWT claims and verifies:
  - `iss` claim equals "modelcraft"
  - `organization` claim matches response organization.name
  - `permissions` claim matches response permissions array
- Test passes validation

#### Scenario: Test verifies permission checks use JWT claims (no DB queries)

**Given**:
- Test has valid ModelCraft JWT with permissions: ["model:read", "model:write"]
- Database connection monitoring is enabled

**When**:
- Test makes GraphQL mutation requiring "model:write" permission
- Test monitors database queries during request

**Then**:
- GraphQL mutation succeeds
- Database query log does NOT contain queries to:
  - `user_roles` table
  - `role_permissions` table
- Only business logic queries are executed (e.g., INSERT for createModel)
- Test verifies permission check used JWT claims (not database)
- Test passes

---

## Notes

### Security Considerations
- JWT_SECRET must be at least 32 bytes for HMAC-SHA256 signing
- JWT_SECRET must be different from Casdoor credentials
- JWT_SECRET should be rotated periodically (requires re-authentication of all users)
- Token lifetime (expiration) balances security (shorter = fresher permissions) and UX (longer = fewer re-auths)

### Performance Impact
- Token issuance adds 2-4 database queries per login (acceptable one-time cost)
- Permission checks reduce from ~10-50ms (database query) to <1ms (in-memory check)
- Expected overall improvement: >90% reduction in authorization latency

### Migration Strategy
- Phase 1: Deploy dual-token support (both Casdoor and ModelCraft JWTs accepted)
- Phase 2: Update clients to use ModelCraft JWT
- Phase 3: Monitor usage of Casdoor JWT (should decline to zero)
- Phase 4: Deprecate Casdoor JWT support after all clients migrated
