# permission-management Specification Delta

## MODIFIED Requirements

### Requirement: Permission Loader Returns Roles and Permissions

The permission loader MUST support querying both user roles and permissions in a single database call for performance optimization.

#### Scenario: Load roles and permissions together

**Given** user "user-uuid-123" in organization "modelcraft" has:
- Role: "owner" (role_id=1)
- Role: "editor" (role_id=2)
- Owner role has permissions: ["*:*"]
- Editor role has permissions: ["model:read", "model:write"]

**When** calling `LoadUserPermissionsAndRoles(ctx, "user-uuid-123", "modelcraft")`

**Then** the method returns:
- `roles` = ["owner", "editor"] (deduplicated role names)
- `permissions` = ["*:*", "model:read", "model:write"] (deduplicated permissions)
- `error` = nil

**And** only ONE database query is made to `user_roles` table

**And** role details are fetched for each user role assignment

#### Scenario: User with no roles returns empty arrays

**Given** user "user-uuid-456" has no role assignments in organization "acme-corp"

**When** calling `LoadUserPermissionsAndRoles(ctx, "user-uuid-456", "acme-corp")`

**Then** the method returns:
- `roles` = [] (empty array)
- `permissions` = [] (empty array)
- `error` = nil

**And** logs: "User has no roles in organization: user_id=user-uuid-456, org=acme-corp"

#### Scenario: Deduplicate roles from multiple assignments

**Given** user "user-uuid-789" in organization "startup" has:
- Role assignment 1: role_id=2 (editor)
- Role assignment 2: role_id=2 (editor) - duplicate due to data issue

**When** calling `LoadUserPermissionsAndRoles(ctx, "user-uuid-789", "startup")`

**Then** the method returns:
- `roles` = ["editor"] (deduplicated, appears once)
- `permissions` = ["model:read", "model:write"] (from editor role)
- `error` = nil

#### Scenario: System role permissions loaded from hardcoded definitions

**Given** user has role "owner" (system role, role_id=1) in organization "org1"

**When** calling `LoadUserPermissionsAndRoles(ctx, userID, "org1")`

**Then** the loader calls `auth.GetSystemRolePermissions("owner")`

**And** returns hardcoded owner permissions (e.g., ["*:*"])

**And** does NOT query the `role_permissions` table for system roles

#### Scenario: Custom role permissions loaded from database

**Given** user has role "data-analyst" (custom role, role_id=10) in organization "org1"

**When** calling `LoadUserPermissionsAndRoles(ctx, userID, "org1")`

**Then** the loader queries `role_permissions` table for role_id=10

**And** returns permissions from database

---

## ADDED Requirements

### Requirement: Permission Cache Supports Roles

The permission cache MUST store and retrieve both roles and permissions as a single atomic unit to ensure consistency.

#### Scenario: Cache stores roles and permissions together

**Given** `LoadUserPermissionsAndRoles()` returns:
- `roles` = ["owner", "editor"]
- `permissions` = ["*:*", "model:read", "model:write"]

**When** `PermissionCache.GetUserPermissionsAndRoles()` caches the result

**Then** the cache key is: `auth:modelcraft:user-uuid-123:1` (includes version)

**And** the cache value is JSON:
```json
{
  "roles": ["owner", "editor"],
  "permissions": ["*:*", "model:read", "model:write"]
}
```

**And** the cache TTL is 5 minutes (300 seconds)

#### Scenario: Cache hit returns both roles and permissions

**Given** Redis cache contains:
- Key: `auth:org1:user-456:2`
- Value: `{"roles": ["viewer"], "permissions": ["model:read"]}`

**When** calling `GetUserPermissionsAndRoles(ctx, "user-456", "org1")`

**And** current version is 2 (matches cache key)

**Then** the method returns:
- `roles` = ["viewer"]
- `permissions` = ["model:read"]
- `error` = nil

**And** logs: "Cache hit: key=auth:org1:user-456:2, roles=1, permissions=1"

**And** NO database query is made

#### Scenario: Cache miss loads from database

**Given** Redis cache does NOT contain key `auth:org2:user-789:1`

**When** calling `GetUserPermissionsAndRoles(ctx, "user-789", "org2")`

**Then** the cache calls `permLoader.LoadUserPermissionsAndRoles(ctx, "user-789", "org2")`

**And** returns the result from database

**And** asynchronously writes to cache (non-blocking)

**And** logs: "Cache miss: key=auth:org2:user-789:1, loaded from DB: roles=2, permissions=5"

#### Scenario: Version mismatch triggers cache miss

**Given** Redis cache contains:
- Key: `auth:org1:user-123:1` (version 1)
- Value: `{"roles": ["editor"], "permissions": ["model:read"]}`

**And** permission version manager returns version 2 (permissions were updated)

**When** calling `GetUserPermissionsAndRoles(ctx, "user-123", "org1")`

**Then** the cache looks for key `auth:org1:user-123:2` (version 2)

**And** key is NOT found (version mismatch)

**And** triggers cache miss (loads from DB)

**And** stores in new cache key with version 2

**Rationale**: Version-based invalidation ensures stale permissions are never served

---

## Implementation Notes

### New Interface Method

```go
type PermissionLoaderInterface interface {
    // Existing method (kept for backward compatibility)
    LoadUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)

    // New method (returns both roles and permissions)
    LoadUserPermissionsAndRoles(ctx context.Context, userID, orgName string) (roles []string, permissions []string, error)
}
```

### Cache Interface Extension

```go
type PermissionCacheInterface interface {
    // Existing method (kept for backward compatibility)
    GetUserPermissions(ctx context.Context, userID, orgName string) ([]string, error)

    // New method (returns both roles and permissions)
    GetUserPermissionsAndRoles(ctx context.Context, userID, orgName string) (roles []string, permissions []string, error)
}
```

### Backward Compatibility

- Old `LoadUserPermissions()` method remains available
- Old `GetUserPermissions()` method remains available
- Existing code continues to work during migration
- New code should use `*AndRoles()` variants
