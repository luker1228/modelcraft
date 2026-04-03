# Spec: JWT Authentication Middleware

## Overview

This specification defines the JWT authentication middleware that secures ModelCraft's runtime GraphQL API. The middleware validates JWT tokens using the appropriate AuthProvider for each project and injects user identity into the request context.

## Context

The middleware sits between the HTTP router and GraphQL handlers, intercepting all `/graphql/*` requests to verify authentication. It delegates JWT verification to the ProviderRegistry, which selects the correct provider based on the project ID in the route.

## ADDED Requirements

### Requirement: JWT Validation Middleware

The system MUST provide a Chi middleware function that validates JWT tokens for GraphQL requests.

#### Scenario: Middleware extracts project ID from route

**Given** a GraphQL request to `/graphql/ecommerce/prod-cluster/products/Order`
**When** the middleware processes the request
**Then** the middleware extracts `projectId` = "ecommerce" from route parameter `:projectId`

#### Scenario: Middleware extracts Bearer token from header

**Given** a request with header `Authorization: Bearer eyJhbGciOiJSUzI1NiIs...`
**When** the middleware processes the request
**Then** the middleware extracts token string = "eyJhbGciOiJSUzI1NiIs..."

#### Scenario: Middleware rejects missing Authorization header

**Given** a request without `Authorization` header
**When** the middleware processes the request
**Then** the middleware responds with:
- HTTP status 401 Unauthorized
- JSON body: `{"error": "missing authorization header"}`
**And** the request does not reach GraphQL handler

#### Scenario: Middleware rejects malformed Authorization header

**Given** a request with header `Authorization: InvalidFormat token123`
**When** the middleware processes the request
**Then** the middleware responds with:
- HTTP status 401 Unauthorized
- JSON body: `{"error": "authorization header must be Bearer token"}`

#### Scenario: Middleware rejects empty token

**Given** a request with header `Authorization: Bearer `
**When** the middleware processes the request
**Then** the middleware responds with:
- HTTP status 401 Unauthorized
- JSON body: `{"error": "missing token"}`

---

### Requirement: Provider-Based Token Validation

The middleware MUST use the ProviderRegistry to validate tokens according to each project's authentication configuration.

#### Scenario: Middleware gets provider from registry

**Given** a request for projectId = "ecommerce"
**When** the middleware processes the request
**Then** the middleware calls `registry.GetProvider("ecommerce")`
**And** uses the returned provider for validation

#### Scenario: Middleware rejects request when provider is nil

**Given** a request for projectId = "unknown-project"
**And** the registry returns `nil` provider (project has no auth config and no default)
**When** the middleware processes the request
**Then** the middleware responds with:
- HTTP status 500 Internal Server Error
- JSON body: `{"error": "authentication provider not configured for project"}`
**And** logs error at ERROR level: "No auth provider for project: unknown-project"

#### Scenario: Middleware validates JWT signature with provider

**Given** a request with a valid JWT token
**And** the registry returns a CasdoorProvider
**When** the middleware processes the request
**Then** the middleware:
1. Calls `provider.GetPublicKey(ctx)` to get verification key
2. Calls `jwt.Parse(token, keyFunc)` with the public key
3. Validates that `token.Method.Alg()` matches `provider.GetSigningMethod()`
4. Extracts `jwt.MapClaims` from parsed token
5. Calls `provider.ParseClaims(mapClaims)` to get UnifiedClaims

#### Scenario: Middleware rejects token with invalid signature

**Given** a request with a JWT token signed with wrong key
**When** the middleware processes the request
**Then** `jwt.Parse()` returns error
**And** the middleware responds with:
- HTTP status 401 Unauthorized
- JSON body: `{"error": "invalid token signature"}`

#### Scenario: Middleware rejects token with wrong signing method

**Given** a request with a JWT token signed with HS256
**And** the provider expects RS256
**When** the middleware processes the request
**Then** the middleware responds with:
- HTTP status 401 Unauthorized
- JSON body: `{"error": "unexpected signing method: HS256"}`

#### Scenario: Middleware rejects expired token

**Given** a request with a JWT token where `exp` claim is in the past
**When** the middleware processes the request
**Then** `jwt.Parse()` validates expiry and returns error
**And** the middleware responds with:
- HTTP status 401 Unauthorized
- JSON body: `{"error": "token has expired"}`

---

### Requirement: Claims Validation and Context Injection

After parsing, the middleware MUST validate claims and inject user identity into request context.

#### Scenario: Middleware validates parsed claims

**Given** a request with a valid JWT
**And** the provider returns UnifiedClaims
**When** the middleware processes the request
**Then** the middleware calls `claims.Validate()`
**And** if validation fails, responds with 401 Unauthorized
**And** if validation succeeds, proceeds to context injection

#### Scenario: Middleware injects user into context

**Given** a request with valid JWT and validated claims
**When** the middleware processes the request
**Then** the middleware calls `c.Set("user", claims)`
**And** calls `c.Next()` to pass control to GraphQL handler
**And** the handler can retrieve claims using `GetUserFromContext(c)`

#### Scenario: Helper function retrieves user from context

**Given** a gin.Context with user set by middleware
**When** the GraphQL handler calls `GetUserFromContext(c)`
**Then** the function returns `(*UnifiedClaims, bool)` tuple
**And** the boolean indicates whether user exists (true if found)

---

### Requirement: Route Integration

The middleware MUST be selectively applied only to runtime GraphQL endpoints.

#### Scenario: Middleware applied to GraphQL routes

**Given** the server is initializing routes
**When** the system sets up `/graphql/*` routes
**Then** the middleware is applied to all routes matching `/graphql/:projectId/*`
**And** the middleware is NOT applied to `/api/design/*` routes
**And** the middleware is NOT applied to `/health` endpoint

#### Scenario: Design-time API remains unauthenticated

**Given** a request to `/api/design/models/test-model`
**When** the request is processed
**Then** the JWT middleware is not invoked
**And** the request reaches the handler without authentication

#### Scenario: Health check remains unauthenticated

**Given** a request to `/health`
**When** the request is processed
**Then** the JWT middleware is not invoked
**And** the request returns 200 OK without authentication

---

### Requirement: Error Handling and Logging

The middleware MUST provide clear error messages and structured logging.

#### Scenario: Log successful authentication

**Given** a request with valid JWT
**When** the middleware successfully validates the token
**Then** the system logs at DEBUG level:
- "JWT authentication successful"
- User ID, Organization, Project ID

#### Scenario: Log authentication failures

**Given** a request with invalid JWT
**When** the middleware rejects the token
**Then** the system logs at WARN level:
- "JWT authentication failed"
- Error reason (e.g., "invalid signature", "expired token")
- Project ID
- Request path

#### Scenario: Log provider errors

**Given** a request where provider.GetPublicKey() fails
**When** the middleware encounters the error
**Then** the system logs at ERROR level:
- "Failed to get public key from provider"
- Provider type
- Project ID
- Error details
**And** responds with 500 Internal Server Error

#### Scenario: Return clear error messages to client

**Given** any authentication failure
**When** the middleware responds to the client
**Then** the error message must:
- Not expose internal implementation details (no stack traces)
- Be actionable (e.g., "token has expired" suggests refresh)
- Use consistent JSON format: `{"error": "message"}`

---

### Requirement: Multi-Tenant Isolation

The middleware MUST support future multi-tenant access control validation.

#### Scenario: Store organization claim for future validation

**Given** a request with valid JWT containing Organization = "tenant1"
**When** the middleware injects claims into context
**Then** the GraphQL handler can access `claims.Organization`
**And** future authorization logic can validate that the user's organization matches the project's tenant

#### Scenario: Log organization mismatch (future feature)

**Note**: This scenario documents future behavior (not implemented in Phase 1)

**Given** a user from Organization "tenant1" accessing a project owned by "tenant2"
**When** authorization validation is implemented
**Then** the system should log at WARN level:
- "Organization mismatch: user from tenant1 attempting to access tenant2 project"
**And** return 403 Forbidden

---

### Requirement: Performance Considerations

The middleware MUST minimize latency impact on GraphQL requests.

#### Scenario: Cache provider lookups

**Given** multiple concurrent requests for the same project
**When** the middleware calls `registry.GetProvider(projectId)`
**Then** the registry returns cached provider (no database query)
**And** provider's public key is also cached (no repeated certificate parsing)

#### Scenario: JWT verification performance

**Given** a single GraphQL request with JWT validation
**When** measured with benchmarking
**Then** JWT verification overhead must be < 1ms on average
**And** RS256 signature verification is the dominant cost (~0.1-0.5ms)

---

### Requirement: Feature Flag for Gradual Rollout

The middleware MUST support disabling authentication for specific projects during migration.

#### Scenario: Feature flag disables auth for specific project

**Note**: This is an optional feature for migration period

**Given** configuration flag `auth.runtime.optional_for_projects = ["default", "legacy"]`
**And** a request to projectId = "default"
**When** the middleware processes the request
**Then** if `Authorization` header is missing, allow the request through without authentication
**And** if `Authorization` header is present, still validate it
**And** log at INFO level: "Auth optional mode: allowing unauthenticated request for project default"

#### Scenario: Feature flag logs warning

**Given** authentication is disabled for a project via feature flag
**When** the server starts
**Then** the system logs at WARN level:
- "Authentication is optional for projects: [default, legacy]"
- "This is a migration aid and should be removed in production"

---

## Design Notes

**Middleware Order**: The JWT middleware should run AFTER logging middleware (to log all requests) but BEFORE GraphQL handler (to inject user context).

**Context Key**: Uses string key "user" to store claims in gin.Context. Alternative: Use typed context key to avoid collisions.

**Error Response Format**: Returns JSON errors consistently with other API endpoints. Status codes:
- 401 Unauthorized: Authentication failed (missing/invalid token)
- 500 Internal Server Error: Provider configuration error

**Token Extraction**: Only supports Bearer token format. Does not support query parameter tokens (security risk).

**No Token Refresh**: Middleware only validates tokens, does not refresh them. Clients must handle refresh by exchanging refresh token with Casdoor directly.

---

## Dependencies

**External Libraries**:
- `github.com/gin-gonic/gin` - Web framework
- `github.com/golang-jwt/jwt/v5` - JWT parsing

**Internal Dependencies**:
- `AuthProvider` interface and `ProviderRegistry`
- `UnifiedClaims` domain model
- Logging facade

---

## Testing Strategy

**Unit Tests**:
- Mock ProviderRegistry and AuthProvider
- Test all error scenarios (missing header, invalid format, expired token)
- Test successful validation flow
- Test context injection and retrieval
- Use httptest to simulate Chi requests

**Integration Tests**:
- Real CasdoorProvider with test certificate
- Generate real JWT tokens (signed with test key)
- Test full request-response cycle
- Test concurrent requests (race detector)
- Measure performance (benchmark)

**Example Unit Test Structure**:
```go
func TestJWTAuthMiddleware_MissingHeader(t *testing.T) {
    // Arrange
    registry := &MockProviderRegistry{}
    middleware := JWTAuthMiddleware(registry)
    router := gin.New()
    router.Use(middleware)
    router.GET("/graphql/:projectId", func(c *gin.Context) {
        c.Status(200)
    })

    // Act
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/graphql/test", nil)
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, 401, w.Code)
    assert.Contains(t, w.Body.String(), "missing authorization header")
}
```

---

## Security Considerations

**Token Storage**: Clients must store JWT tokens securely (httpOnly cookies or secure local storage). Middleware does not enforce storage method.

**HTTPS Requirement**: In production, HTTPS must be enforced (Bearer tokens in plain HTTP are vulnerable to interception). Server should reject HTTP requests or redirect to HTTPS.

**Token Expiry**: Middleware relies on `exp` claim validation by jwt.Parse(). Short-lived tokens (2 hours) reduce replay attack window.

**Logging Sensitive Data**: Token contents are NOT logged (only validation results). Avoid logging full tokens or claims with PII in production.

**Rate Limiting**: Middleware does not implement rate limiting. Consider adding rate limiter middleware before JWT auth to prevent brute force attacks.

---

## Open Questions

1. **Custom Claims Support**: Should middleware support provider-specific custom claims in UnifiedClaims.Raw?
   - Use case: Casdoor roles or permissions for future RBAC
   - Current: Raw map is available but unused

2. **Token Introspection**: Should middleware support OAuth 2.0 token introspection endpoint?
   - Use case: Immediate revocation without waiting for expiry
   - Complexity: Requires HTTP call to Casdoor on every request (performance concern)

3. **Authorization Logic**: Where should multi-tenant validation live?
   - Option A: In middleware (rejects before reaching GraphQL)
   - Option B: In GraphQL resolvers (more flexible, can return typed errors)
   - Recommendation: Option A for simplicity (Phase 2 enhancement)
