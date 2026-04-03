# org-context-middleware Specification

## Purpose

The organization context middleware establishes organization-scoped authorization by extracting the organization identifier from the URL path and loading fresh user roles and permissions from the database (with Redis caching). This middleware runs after JWT authentication and before request handlers.

## ADDED Requirements

### Requirement: Extract Organization from URL Path

The middleware MUST extract the organization name from the URL path parameter `{orgName}` and validate it is non-empty.

#### Scenario: Extract orgName from URL parameter

**Given** a request to URL `/org/modelcraft/design/graphql`

**And** the Chi router is configured with pattern `/org/{orgName}/design/graphql`

**When** the org context middleware processes the request

**Then** the middleware extracts `orgName` = "modelcraft" using `chi.URLParam(r, "orgName")`

**And** logs: "OrgContext: orgName=modelcraft"

#### Scenario: Reject request with missing orgName

**Given** a request to URL `/org//design/graphql` (empty orgName)

**Or** the route pattern does not include `{orgName}` parameter

**When** the org context middleware processes the request

**Then** the middleware responds with:
- HTTP status 400 Bad Request
- JSON body: `{"error": "organization name required in URL path", "code": "ORG_NAME_REQUIRED"}`

**And** logs: "Missing orgName in URL path"

**And** the request does NOT proceed to the handler

---

### Requirement: Extract User Identity from JWT Context

The middleware MUST retrieve the user ID from the request context (previously set by JWT authentication middleware) and validate it exists.

#### Scenario: Extract userID from context

**Given** JWT authentication middleware has injected `ContextKeyUserID` = "user-uuid-123" into context

**When** the org context middleware processes the request

**Then** the middleware extracts `userID` = "user-uuid-123" from context

**And** logs: "OrgContext: userId=user-uuid-123, orgName=modelcraft"

#### Scenario: Reject request with missing userID

**Given** the request context does NOT contain `ContextKeyUserID` (JWT auth was skipped or failed)

**When** the org context middleware processes the request

**Then** the middleware responds with:
- HTTP status 401 Unauthorized
- JSON body: `{"error": "user authentication required", "code": "AUTH_REQUIRED"}`

**And** logs: "Missing user_id in context (JWT auth failed?)"

**And** the request does NOT proceed to the handler

---

### Requirement: Load Fresh Roles and Permissions from Database

The middleware MUST query user roles and permissions for the specific organization using the permission cache (Redis + DB fallback).

#### Scenario: Load roles and permissions via cache

**Given** `userID` = "user-uuid-123" and `orgName` = "modelcraft"

**And** permission cache is configured with Redis client

**When** the org context middleware loads permissions

**Then** the middleware calls `permCache.GetUserPermissionsAndRoles(ctx, "user-uuid-123", "modelcraft")`

**And** receives:
- `roles` = ["owner", "editor"]
- `permissions` = ["*:*", "model:read", "model:write"]
- `error` = nil

**And** logs: "Authorization loaded: userId=user-uuid-123, orgName=modelcraft, roles=2, permissions=3"

#### Scenario: Reject request when permission query fails

**Given** the permission cache query returns an error (e.g., database connection failed)

**When** the org context middleware loads permissions

**Then** the middleware responds with:
- HTTP status 500 Internal Server Error
- JSON body: `{"error": "failed to load authorization data", "code": "PERMISSION_LOAD_FAILED"}`

**And** logs: "Failed to load roles and permissions: userId=user-uuid-123, orgName=modelcraft, error=<error details>"

**And** the request does NOT proceed to the handler

**Rationale**: Fail-closed security - if we can't verify permissions, deny access

#### Scenario: Reject user not in organization

**Given** user "user-uuid-456" has NO role assignments in organization "acme-corp"

**When** the org context middleware loads permissions

**Then** `GetUserPermissionsAndRoles()` returns:
- `roles` = [] (empty)
- `permissions` = [] (empty)
- `error` = nil

**And** the middleware responds with:
- HTTP status 403 Forbidden
- JSON body: `{"error": "user not authorized in this organization", "code": "USER_NOT_IN_ORG"}`

**And** logs: "User not authorized in organization: user_id=user-uuid-456, org_name=acme-corp"

**And** the request does NOT proceed to the handler

**Rationale**: Empty permissions means user is not a member of the organization

---

### Requirement: Inject Authorization Context

The middleware MUST inject organization name, roles, and permissions into the request context for downstream handlers and permission checks.

#### Scenario: Inject orgName, roles, permissions into context

**Given** successfully loaded:
- `orgName` = "modelcraft"
- `roles` = ["owner", "editor"]
- `permissions` = ["*:*", "model:read", "model:write"]

**When** the org context middleware completes authorization

**Then** the middleware injects into context:
- `ContextKeyOrgName` = "modelcraft"
- `ContextKeyRoles` = ["owner", "editor"]
- `ContextKeyPermissions` = ["*:*", "model:read", "model:write"]

**And** continues to the next handler with updated context

**And** the next handler can read permissions using `ctx.Value(ContextKeyPermissions).([]string)`

#### Scenario: Permission middleware reads injected permissions

**Given** org context middleware has injected permissions into context

**And** the request handler uses `RequirePermission("model:write")` middleware

**When** the permission middleware checks authorization

**Then** it reads `ContextKeyPermissions` from context

**And** finds "model:write" permission in the list

**And** allows the request to proceed

**Rationale**: Org context middleware and permission middleware work together seamlessly

---

### Requirement: Performance with Redis Caching

The middleware MUST leverage Redis caching to minimize database queries and maintain low latency.

#### Scenario: Cache hit provides sub-millisecond response

**Given** permission cache contains valid entry for user + org

**When** the org context middleware loads permissions

**Then** Redis cache returns the result in <1ms

**And** NO database query is made

**And** total middleware latency is <2ms

#### Scenario: Cache miss triggers database query

**Given** permission cache does NOT contain entry for user + org

**When** the org context middleware loads permissions

**Then** cache calls database to load permissions (~10-50ms)

**And** asynchronously stores result in Redis cache

**And** total middleware latency is ~10-52ms (first request only)

**And** subsequent requests for same user + org hit cache (<1ms)

#### Scenario: Cache TTL ensures freshness

**Given** permission cache has 5-minute TTL

**When** a permission is changed in the database

**Then** the change takes effect within 5 minutes (when cache expires)

**And** the next request loads fresh permissions from database

**And** cache is updated with new permissions

**Rationale**: Balance between performance and permission freshness

---

## Middleware Chain Order

The org context middleware MUST run in a specific order relative to other middlewares:

```
1. JWT Authentication Middleware (set userID in context)
   ↓
2. Org Context Middleware (set orgName, roles, permissions in context)
   ↓
3. Permission Middleware (optional, check specific permissions)
   ↓
4. Request Handler (execute business logic)
```

**Critical**: Org context middleware MUST run AFTER JWT authentication (depends on `userID` in context).

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": "<human-readable message>",
  "code": "<machine-readable error code>"
}
```

### Error Codes

| Code | Status | Meaning |
|------|--------|---------|
| `ORG_NAME_REQUIRED` | 400 | Missing or empty orgName in URL path |
| `AUTH_REQUIRED` | 401 | Missing userID in context (JWT auth not run) |
| `PERMISSION_LOAD_FAILED` | 500 | Database or cache query failed |
| `USER_NOT_IN_ORG` | 403 | User has no roles/permissions in organization |

---

## Configuration

### URL Pattern

The middleware expects routes configured with organization parameter:

```go
r.Route("/org/{orgName}", func(r chi.Router) {
    r.Use(ChiJWTAuthMiddleware(jwtConfig))        // Step 1: Authenticate
    r.Use(OrgContextMiddlewareWithCache(permCache)) // Step 2: Authorize

    r.Post("/design/graphql", designGraphQLHandler) // Step 3: Handle
})
```

### Dependency Injection

The middleware requires a `PermissionCacheInterface` instance:

```go
permCache := auth.NewPermissionCache(
    redisClient,
    permLoader,
    versionManager,
    5*time.Minute, // Cache TTL
    logger,
)

middleware := OrgContextMiddlewareWithCache(permCache)
```

---

## Related Specifications

- **`jwt-middleware`**: Provides user authentication and injects `userID` into context
- **`permission-management`**: Defines permission loading and caching logic
- **`dual-token-auth`**: Defines JWT token structure and validation rules
