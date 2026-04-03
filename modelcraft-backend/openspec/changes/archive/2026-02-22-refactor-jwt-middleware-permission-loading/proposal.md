# Change: Refactor JWT Middleware Permission Loading

## Why

Currently, the JWT authentication middleware extracts roles and permissions from JWT claims and injects them directly into the request context. This approach has several issues:

1. **Stale Permissions**: JWT-embedded permissions remain valid until token expiration (typically 1 hour), causing permission changes to take effect only after users re-authenticate
2. **Inconsistent Authorization**: Different token types (ModelCraft JWT vs Casdoor JWT) have different permission formats, leading to inconsistent authorization behavior
3. **No Organization Context**: JWT middleware runs before organization context is established, but permissions are organization-scoped (require both `userID` and `orgName`)
4. **Mixed Responsibilities**: JWT middleware should focus on authentication (verifying user identity), not authorization (loading permissions)

The current architecture also has an `OrgContextMiddleware` that loads permissions from the database, but it's only used for specific routes. This creates two different permission loading paths, leading to confusion and maintenance burden.

## What Changes

### Architecture Shift

**Before**: JWT Middleware → Inject permissions from JWT claims → Request Handler
**After**: JWT Middleware → Inject userID only → Org Context Middleware → Load fresh permissions from DB (with Redis cache) → Request Handler

### Specific Changes

1. **JWT Middleware (`internal/middleware/jwt_auth.go` & `chi_jwt_auth.go`)**
   - Remove logic that injects `roles` and `permissions` into context
   - Only inject basic identity information: `userID`, `email`, `name`, `organization`
   - Both ModelCraft JWT and Casdoor JWT are handled uniformly

2. **Org Context Middleware (`internal/middleware/org_context.go`)**
   - Extract `orgName` from URL path (`/org/{orgName}/design/graphql`)
   - Extract `userID` from context (set by JWT middleware)
   - Use `PermissionCache.GetUserPermissionsAndRoles(ctx, userID, orgName)` to load fresh permissions with Redis caching
   - Inject `orgName`, `roles`, `permissions` into context
   - Reject request if query fails (return 403 or 500)

3. **Permission Cache (`internal/app/auth/permission_cache.go`)**
   - No changes to cache logic (already uses Redis with 5-minute TTL)

4. **Permission Loader (`internal/app/auth/permission_loader.go`)**
   - Extend `LoadUserPermissions()` to return both roles and permissions
   - Change signature: `LoadUserPermissionsAndRoles(ctx, userID, orgName) (roles []string, permissions []string, error)`
   - Query roles from `user_roles` table in the same database call

### URL Design

The new architecture uses organization-scoped URLs:

```
POST /org/{orgName}/design/graphql
```

Where:
- `orgName` is the organization identifier (extracted by Org Context Middleware)
- All design-related GraphQL operations are under this path

## Impact

### Affected Specs

- **`jwt-middleware`**: MODIFIED - Remove permission injection logic
- **`permission-management`**: MODIFIED - Extend to support role loading alongside permissions
- **New capability**: `org-context-middleware` - Formal specification for organization-scoped permission loading

### Affected Code

- `internal/middleware/jwt_auth.go` - Remove lines 204-205, 278-283
- `internal/middleware/chi_jwt_auth.go` - Remove lines 168-169, 234-238
- `internal/middleware/org_context.go` - Extend to query roles
- `internal/app/auth/permission_loader.go` - Add `LoadUserPermissionsAndRoles()` method
- `internal/app/auth/permission_cache.go` - Update to call new loader method

### Breaking Changes

**None for API consumers**. This is an internal refactoring:
- External API contracts remain unchanged
- GraphQL queries continue to work as before
- Permission checks still use the same `RequirePermission()` middleware

**Internal Breaking Changes**:
- Code that reads `roles`/`permissions` directly from context before Org Context Middleware will fail
- New routes MUST use Org Context Middleware to get permissions
- URL structure changes to `/org/{orgName}/design/graphql` (from previous `/graphql` or `/design/graphql`)

### Migration Path

1. Deploy new middleware with dual support (read permissions from JWT claims OR context)
2. Update route registration to use new URL structure with Org Context Middleware
3. Deploy frontend changes to use new URL format
4. Remove fallback logic after verification period (future change)

### Performance Impact

**Before** (JWT-embedded permissions):
- JWT validation: ~1-2ms
- Permission check: <1ms (read from context)
- **Total**: ~2ms per request

**After** (DB-loaded permissions with Redis cache):
- JWT validation: ~1-2ms
- Redis cache hit: <1ms
- Redis cache miss + DB query: ~10-50ms (5-minute TTL, infrequent)
- Permission check: <1ms (read from context)
- **Total**: ~3ms per request (cache hit), ~52ms per request (cache miss every 5 minutes)

**Trade-off**: Slight performance regression (<1ms on average) in exchange for always-fresh permissions and consistent authorization model.
