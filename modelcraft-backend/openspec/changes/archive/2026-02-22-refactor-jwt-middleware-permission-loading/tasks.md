# Implementation Tasks

## Phase 1: Extend Permission Loader (Foundation)

### Task 1.1: Extend PermissionLoader Interface
- [ ] Add new method `LoadUserPermissionsAndRoles(ctx, userID, orgName) ([]string, []string, error)` to `PermissionLoaderInterface` in `internal/app/auth/permission_loader.go`
- [ ] Return format: `(roles []string, permissions []string, error)`
- [ ] Roles are simple role names (e.g., `["owner", "editor"]`)

### Task 1.2: Implement Role Loading in PermissionLoader
- [ ] In `internal/app/auth/permission_loader.go`, implement `LoadUserPermissionsAndRoles()`
- [ ] Query user roles from `user_roles` table using existing `userRoleRepo.ListUserRoles(ctx, userID, orgName)`
- [ ] Extract role names from returned `UserRole` entities
- [ ] Deduplicate role names (use map, convert to slice)
- [ ] Return both roles and permissions in single call
- [ ] Add unit tests for role loading logic

### Task 1.3: Update PermissionCache to Support Roles
- [ ] Modify `PermissionCache.GetUserPermissions()` signature to return roles as well
- [ ] Update cache key to include roles data: `auth:{orgName}:{userId}:{version}`
- [ ] Update cache value structure to store both roles and permissions as JSON: `{"roles": [...], "permissions": [...]}`
- [ ] Update serialization/deserialization logic for new format
- [ ] Add unit tests for role caching

## Phase 2: Simplify JWT Middleware (Authentication Only)

### Task 2.1: Remove Permission Injection from Chi JWT Middleware
- [ ] In `internal/middleware/jwt_auth.go`, function `validateModelCraftJWT()`:
  - Remove lines 204-205 that set `ContextKeyRoles` and `ContextKeyPermissions`
  - Keep only: `userID`, `email`, `name`, `organization`
- [ ] In function `validateCasdoorJWT()`:
  - Remove lines 278-283 that conditionally set permissions/roles
  - Keep only: `userID`, `email`, `name`, `organization`
- [ ] Update log messages to reflect that only identity is extracted
- [ ] Add unit tests to verify permissions are NOT in context after JWT validation

### Task 2.2: Remove Permission Injection from Chi JWT Middleware
- [ ] In `internal/middleware/chi_jwt_auth.go`, function `validateModelCraftJWTChi()`:
  - Remove lines 168-169 that set `ContextKeyRoles` and `ContextKeyPermissions`
- [ ] In function `validateCasdoorJWTChi()`:
  - Remove lines 234-238 that conditionally set permissions/roles
- [ ] Update log messages consistently with Chi middleware
- [ ] Add unit tests for Chi middleware

### Task 2.3: Update JWT Middleware Documentation
- [ ] Update comments in `jwt_auth.go` to clarify it only handles authentication
- [ ] Add comment directing to `OrgContextMiddleware` for authorization
- [ ] Update CLAUDE.md authentication section if needed

## Phase 3: Enhance Org Context Middleware (Authorization)

### Task 3.1: Update OrgContextMiddleware to Load Roles
- [ ] In `internal/middleware/org_context.go`, function `OrgContextMiddleware()`:
  - Change call from `permLoader.LoadUserPermissions()` to `permLoader.LoadUserPermissionsAndRoles()`
  - Extract both `roles` and `permissions` from return values
  - Inject `roles` into context: `ctx = context.WithValue(ctx, ContextKeyRoles, roles)`
- [ ] Update log messages to include role count
- [ ] Add unit tests for role injection

### Task 3.2: Update OrgContextMiddlewareWithCache to Load Roles
- [ ] In `internal/middleware/org_context.go`, function `OrgContextMiddlewareWithCache()`:
  - Update to call new cache method that returns roles
  - Inject both `roles` and `permissions` into context
- [ ] Update error handling to reject request on query failure (already implemented)
- [ ] Add unit tests for cached role loading

### Task 3.3: Verify Permission Middleware Compatibility
- [ ] Review `internal/middleware/permission.go` to ensure it still works correctly
- [ ] Verify `getPermissionsWithContext()` reads from context correctly
- [ ] Add integration tests: JWT Auth → Org Context → Permission Check

## Phase 4: Update Route Configuration

### Task 4.1: Update Chi Router Configuration
- [ ] In `cmd/server/main.go` (or router setup file):
  - Change route from `/design/graphql` to `/org/{orgName}/design/graphql`
  - Ensure middleware chain: JWT Auth → Org Context → GraphQL Handler
- [ ] Verify `chi.URLParam(r, "orgName")` correctly extracts organization name
- [ ] Add route registration tests

### Task 4.2: Update Chi Router Configuration (if used)
- [ ] If Chi routes exist, update to `/org/:orgName/design/graphql`
- [ ] Ensure middleware chain order is correct
- [ ] Verify parameter extraction with `c.Param("orgName")`

### Task 4.3: Add Backward Compatibility Routes (Optional)
- [ ] Consider adding redirect from old URLs to new URLs
- [ ] Or return 410 Gone with migration instructions
- [ ] Document URL migration in API changelog

## Phase 5: Testing & Validation

### Task 5.1: Unit Tests
- [ ] Test JWT middleware does NOT inject permissions
- [ ] Test Org Context middleware correctly loads roles + permissions
- [ ] Test PermissionLoader returns both roles and permissions
- [ ] Test PermissionCache stores and retrieves roles correctly

### Task 5.2: Integration Tests
- [ ] Test full flow: JWT Auth → Org Context → Permission Check
- [ ] Test with valid ModelCraft JWT (no embedded permissions)
- [ ] Test with valid Casdoor JWT (no embedded permissions)
- [ ] Test with invalid orgName (should reject)
- [ ] Test with user not in organization (should reject with 403)
- [ ] Test permission query failure (should reject with 500)

### Task 5.3: Performance Tests
- [ ] Measure Redis cache hit latency (target: <1ms)
- [ ] Measure Redis cache miss + DB query latency (target: <50ms)
- [ ] Verify cache TTL is 5 minutes
- [ ] Test concurrent requests with same user+org (cache sharing)

### Task 5.4: Manual Testing
- [ ] Test GraphQL queries with new URL format
- [ ] Verify permissions are always fresh (change role, verify immediate effect)
- [ ] Test with multiple organizations per user
- [ ] Verify error messages are user-friendly

## Phase 6: Documentation & Deployment

### Task 6.1: Update Documentation
- [ ] Update CLAUDE.md with new middleware architecture
- [ ] Update authentication.md with URL structure
- [ ] Add migration guide for API consumers
- [ ] Update OpenAPI/GraphQL schema with new paths

### Task 6.2: Deployment Checklist
- [ ] Verify all unit tests pass
- [ ] Verify all integration tests pass
- [ ] Review code changes with team
- [ ] Deploy to staging environment
- [ ] Run smoke tests on staging
- [ ] Monitor error logs for authentication issues
- [ ] Deploy to production
- [ ] Monitor performance metrics (latency, cache hit rate)

### Task 6.3: Post-Deployment Validation
- [ ] Verify Redis cache is being used (check cache hit rate)
- [ ] Verify permission changes take effect immediately (<5 minutes)
- [ ] Monitor authentication error rate
- [ ] Collect performance metrics for 1 week
- [ ] Adjust cache TTL if needed based on metrics

## Dependencies

- **Task 1.x must complete before Task 2.x**: Permission loader must support roles before middleware can use it
- **Task 2.x and Task 3.x can run in parallel**: JWT and Org Context changes are independent
- **Task 4.x depends on Task 2.x and Task 3.x**: Routes need both middlewares ready
- **Task 5.x depends on Task 4.x**: Testing requires full implementation
- **Task 6.x depends on Task 5.x**: Documentation and deployment after validation
