# Design: Refactor JWT Middleware Permission Loading

## Architecture Overview

### Current Architecture (Before)

```
┌─────────────────────────────────────────────────────────────┐
│  Client Request: POST /design/graphql                       │
│  Header: Authorization: Bearer <ModelCraft_JWT>             │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  JWT Authentication Middleware                               │
│  - Validate JWT signature                                    │
│  - Extract claims: userID, email, name, org, roles, perms   │
│  - Inject ALL into context                                   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  GraphQL Handler                                             │
│  - Read permissions from context                             │
│  - Execute query with authorization checks                   │
└─────────────────────────────────────────────────────────────┘

Problems:
❌ Permissions in JWT are stale (valid until token expiration)
❌ Permission changes don't take effect until re-authentication
❌ No organization context during JWT validation
❌ Mixed authentication and authorization concerns
```

### New Architecture (After)

```
┌─────────────────────────────────────────────────────────────┐
│  Client Request: POST /org/{orgName}/design/graphql         │
│  Header: Authorization: Bearer <JWT>                        │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  JWT Authentication Middleware (Authentication ONLY)         │
│  - Validate JWT signature                                    │
│  - Extract IDENTITY: userID, email, name, organization       │
│  - Inject identity into context                              │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Org Context Middleware (Authorization)                      │
│  - Extract orgName from URL path parameter                   │
│  - Extract userID from context                               │
│  - Load fresh roles + permissions from DB (Redis cache)      │
│  - Inject orgName, roles, permissions into context           │
│  - REJECT if query fails or user not in org                  │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  GraphQL Handler                                             │
│  - Read permissions from context                             │
│  - Execute query with authorization checks                   │
└─────────────────────────────────────────────────────────────┘

Benefits:
✅ Always-fresh permissions (<5 minute cache)
✅ Permission changes take effect immediately
✅ Clear separation: authentication vs authorization
✅ Organization-scoped permission loading
```

## Component Design

### 1. Permission Loader Enhancement

**File**: `internal/app/auth/permission_loader.go`

**New Method Signature**:
```go
type PermissionLoaderInterface interface {
    // Old method (kept for backward compatibility)
    LoadUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)

    // New method (returns both roles and permissions)
    LoadUserPermissionsAndRoles(ctx context.Context, userID, orgName string) (roles []string, permissions []string, error)
}
```

**Implementation Logic**:
```go
func (l *PermissionLoader) LoadUserPermissionsAndRoles(ctx context.Context, userID, orgName string) ([]string, []string, error) {
    // Step 1: Query user roles in organization (existing logic)
    userRoles, err := l.userRoleRepo.ListUserRoles(ctx, userID, orgName)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to query user roles: %w", err)
    }

    if len(userRoles) == 0 {
        // User has no roles in org -> empty permissions
        return []string{}, []string{}, nil
    }

    // Step 2: Extract role names and load permissions (in parallel)
    roleMap := make(map[string]bool)      // Deduplicate roles
    permissionMap := make(map[string]bool) // Deduplicate permissions

    for _, userRole := range userRoles {
        // Get role details
        role, err := l.roleRepo.GetRoleByID(ctx, userRole.RoleID)
        if err != nil || role == nil {
            continue // Skip this role
        }

        // Add role name
        roleMap[role.Name] = true

        // Load permissions (existing logic)
        var rolePerms []*domainPerm.Permission
        if role.IsSystemRole() {
            rolePerms = auth.GetSystemRolePermissions(role.Name)
        } else {
            rolePerms, _ = l.permRepo.ListPermissionsByRole(ctx, userRole.RoleID)
        }

        // Add permissions
        for _, perm := range rolePerms {
            permissionMap[perm.String()] = true
        }
    }

    // Step 3: Convert maps to slices
    roles := make([]string, 0, len(roleMap))
    for role := range roleMap {
        roles = append(roles, role)
    }

    permissions := make([]string, 0, len(permissionMap))
    for perm := range permissionMap {
        permissions = append(permissions, perm)
    }

    return roles, permissions, nil
}
```

**Design Rationale**:
- Single database query round-trip (query user_roles once, then iterate)
- Deduplication using maps (user might have multiple roles with overlapping permissions)
- Fails gracefully (skip invalid roles, continue with others)

### 2. Permission Cache Enhancement

**File**: `internal/app/auth/permission_cache.go`

**Updated Method Signature**:
```go
type PermissionCacheInterface interface {
    GetUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)

    // New method
    GetUserPermissionsAndRoles(ctx context.Context, userID, orgName string) (roles []string, permissions []string, error)
}
```

**Cache Value Structure**:
```json
{
  "roles": ["owner", "editor"],
  "permissions": ["model:read", "model:write", "cluster:manage"]
}
```

**Cache Key Format** (unchanged):
```
auth:{orgName}:{userID}:{version}
```

**Implementation Logic**:
```go
func (c *PermissionCache) GetUserPermissionsAndRoles(ctx context.Context, userID, orgName string) ([]string, []string, error) {
    // Step 1: Get version
    version, err := c.versionManager.GetVersion(ctx, orgName, userID)
    if err != nil {
        version = 1 // Fallback
    }

    cacheKey := c.buildVersionedCacheKey(orgName, userID, version)

    // Step 2: Try cache first
    cached, err := c.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        // Cache hit - deserialize
        var data struct {
            Roles       []string `json:"roles"`
            Permissions []string `json:"permissions"`
        }
        if json.Unmarshal([]byte(cached), &data) == nil {
            return data.Roles, data.Permissions, nil
        }
    }

    // Step 3: Cache miss - load from DB
    roles, permissions, err := c.permLoader.LoadUserPermissionsAndRoles(ctx, userID, orgName)
    if err != nil {
        return nil, nil, err
    }

    // Step 4: Store in cache (async, best effort)
    go c.cacheRolesAndPermissions(cacheKey, roles, permissions)

    return roles, permissions, nil
}
```

**Design Rationale**:
- Backward compatible (keep old `GetUserPermissions()` method)
- Single cache entry for both roles and permissions (atomic update)
- Async cache write (don't block request if Redis is slow)

### 3. JWT Middleware Simplification

**Files**: `internal/middleware/jwt_auth.go`, `internal/middleware/chi_jwt_auth.go`

**Changes**:
```go
// BEFORE (validateModelCraftJWT)
c.Set(ContextKeyUserID, claims.UserID)
c.Set(ContextKeyEmail, claims.Email)
c.Set(ContextKeyName, claims.Name)
c.Set(ContextKeyOrganization, claims.Organization)
c.Set(ContextKeyRoles, claims.Roles)              // ❌ REMOVE
c.Set(ContextKeyPermissions, claims.Permissions)  // ❌ REMOVE

// AFTER
c.Set(ContextKeyUserID, claims.UserID)
c.Set(ContextKeyEmail, claims.Email)
c.Set(ContextKeyName, claims.Name)
c.Set(ContextKeyOrganization, claims.Organization)
// Roles and permissions will be loaded by OrgContextMiddleware
```

**Updated Log Messages**:
```go
// BEFORE
logger.Infof("ModelCraft JWT validated successfully for user: %s (email: %s, org: %s)",
    claims.UserID, claims.Email, claims.Organization)

// AFTER
logger.Infof("JWT authentication successful: userID=%s, email=%s, org=%s (authorization deferred to org context middleware)",
    claims.UserID, claims.Email, claims.Organization)
```

**Design Rationale**:
- Middleware focuses on single responsibility (authentication only)
- Reduces coupling between JWT structure and authorization logic
- Both token types (ModelCraft, Casdoor) handled uniformly

### 4. Org Context Middleware Enhancement

**File**: `internal/middleware/org_context.go`

**Updated Implementation**:
```go
func OrgContextMiddlewareWithCache(permCache auth.PermissionCacheInterface) func(http.Handler) http.Handler {
    logger := logfacade.GetDefault()

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Step 1: Extract orgName from URL path
            orgName := chi.URLParam(r, "orgName")
            if orgName == "" {
                logger.Errorf("Missing orgName in URL path")
                writeOrgContextError(w, http.StatusBadRequest,
                    "organization name required in URL path", "ORG_NAME_REQUIRED")
                return
            }

            // Step 2: Extract userID from JWT context
            ctx := r.Context()
            userID, ok := ctx.Value(ContextKeyUserID).(string)
            if !ok || userID == "" {
                logger.Errorf("Missing user_id in context (JWT auth failed?)")
                writeOrgContextError(w, http.StatusUnauthorized,
                    "user authentication required", "AUTH_REQUIRED")
                return
            }

            // Step 3: Load roles + permissions from cache/DB
            roles, permissions, err := permCache.GetUserPermissionsAndRoles(r.Context(), userID, orgName)
            if err != nil {
                logger.Errorf("Failed to load roles and permissions: userId=%s, orgName=%s, error=%v",
                    userID, orgName, err)
                writeOrgContextError(w, http.StatusInternalServerError,
                    "failed to load authorization data", "PERMISSION_LOAD_FAILED")
                return
            }

            // Step 4: Reject if user has no roles/permissions (not in org)
            if len(permissions) == 0 {
                logger.Warn("User not authorized in organization",
                    logfacade.String("user_id", userID),
                    logfacade.String("org_name", orgName))
                writeOrgContextError(w, http.StatusForbidden,
                    "user not authorized in this organization", "USER_NOT_IN_ORG")
                return
            }

            logger.Infof("Authorization loaded: userId=%s, orgName=%s, roles=%d, permissions=%d",
                userID, orgName, len(roles), len(permissions))

            // Step 5: Inject into context
            ctx = context.WithValue(ctx, ContextKeyOrgName, orgName)
            ctx = context.WithValue(ctx, ContextKeyRoles, roles)
            ctx = context.WithValue(ctx, ContextKeyPermissions, permissions)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**Design Rationale**:
- Middleware runs AFTER JWT auth (depends on userID in context)
- Organization context is established before authorization
- Query failure is treated as authorization failure (fail-closed security)
- Empty permissions = user not in org (explicit rejection)

## URL Structure Design

### Pattern
```
/org/{orgName}/design/graphql
```

### Examples
```
POST /org/modelcraft/design/graphql
POST /org/acme-corp/design/graphql
POST /org/my-startup/design/graphql
```

### Rationale
- **`/org/` prefix**: Makes it clear this is organization-scoped
- **`{orgName}` parameter**: Extracted by Chi router (`chi.URLParam(r, "orgName")`)
- **`/design/` subdomain**: Distinguishes from other domains (e.g., `/runtime/`, `/admin/`)
- **`/graphql` endpoint**: Standard GraphQL single endpoint convention

### Alternative Considered (Rejected)
```
❌ /design/{orgName}/graphql
```
Rejected because:
- Less clear that `orgName` is the primary scope
- Harder to add other organization-level routes (e.g., `/settings`, `/members`)
- Doesn't follow multi-tenant best practices (tenant ID should be early in path)

## Error Handling

### Scenarios

| Scenario | HTTP Status | Error Code | Error Message |
|----------|-------------|------------|---------------|
| Missing `orgName` in URL | 400 | `ORG_NAME_REQUIRED` | "organization name required in URL path" |
| Missing JWT token | 401 | `AUTH_MISSING_TOKEN` | "missing authorization token" |
| Invalid JWT token | 401 | `AUTH_INVALID_TOKEN` | "invalid token" |
| Missing `userID` in context | 401 | `AUTH_REQUIRED` | "user authentication required" |
| Permission query failed | 500 | `PERMISSION_LOAD_FAILED` | "failed to load authorization data" |
| User not in organization | 403 | `USER_NOT_IN_ORG` | "user not authorized in this organization" |

### Error Response Format
```json
{
  "error": "user not authorized in this organization",
  "code": "USER_NOT_IN_ORG"
}
```

## Performance Analysis

### Latency Breakdown

**Cache Hit Path** (~3ms total):
1. JWT validation: ~1-2ms
2. Redis cache read: <1ms
3. Context injection: <0.1ms

**Cache Miss Path** (~52ms total, once per 5 minutes):
1. JWT validation: ~1-2ms
2. Redis cache miss: ~1ms
3. Database query (roles + permissions): ~10-50ms
4. Redis cache write (async): 0ms (non-blocking)
5. Context injection: <0.1ms

### Cache Efficiency

**Assumptions**:
- Cache TTL: 5 minutes (300 seconds)
- Average user request rate: 10 req/min

**Cache Hit Rate**:
- First request: Cache miss (query DB)
- Next 49 requests (within 5 min): Cache hit (read Redis)
- Cache hit rate: 49/50 = **98%**

**Average Latency**:
- 98% × 3ms + 2% × 52ms = **3.94ms**
- Compared to JWT-embedded: ~2ms
- **Overhead: ~1.94ms per request** (acceptable trade-off for always-fresh permissions)

### Database Load

**Before** (JWT-embedded):
- 0 queries per request (permissions in JWT)

**After** (DB-loaded with Redis cache):
- Cache hit (98%): 0 queries
- Cache miss (2%): 1 query (roles + permissions in single call)
- **Average: 0.02 queries per request**

For 1000 req/min → 20 DB queries/min → Negligible database load

## Security Considerations

### Fail-Closed Design
- If permission query fails → Request is **REJECTED** (500 error)
- If user has no permissions → Request is **REJECTED** (403 error)
- Redis failure → Falls back to DB (no bypass)
- DB failure → Request fails (no default-allow)

### Permission Freshness
- Cache TTL: 5 minutes (configurable)
- Permission changes take effect within 5 minutes (vs. 1 hour with JWT)
- Trade-off: Performance vs. freshness (5 min is reasonable for most use cases)

### Version-Based Invalidation
- Version number incremented when permissions change
- Old cache entries automatically stale (version mismatch)
- No need for manual cache clearing
- Prevents permission escalation attacks

## Migration Strategy

### Phase 1: Deploy New Code (Week 1)
- Deploy backend with new middleware
- Keep old routes working (parallel deployment)
- Monitor error rates and latency

### Phase 2: Frontend Migration (Week 2-3)
- Update frontend to use new URL format (`/org/{orgName}/design/graphql`)
- Test thoroughly in staging
- Gradual rollout (10% → 50% → 100%)

### Phase 3: Deprecate Old Routes (Week 4)
- Return 410 Gone for old routes with migration instructions
- Monitor for stragglers
- Remove old code after 30 days

### Rollback Plan
If issues arise:
1. Revert route changes (use old URLs)
2. Re-enable permission injection in JWT middleware (temporary)
3. Investigate root cause
4. Fix and redeploy
