## Context

ModelCraft requires an operation-level authorization system that:
1. Supports multi-tenant RBAC with system roles and custom roles
2. Provides AI-friendly declarative permission syntax (GraphQL directives)
3. Integrates seamlessly with existing JWT authentication architecture
4. Allows dynamic permission management without code changes

### Current State

- JWT tokens contain permissions as string arrays (e.g., `["project:create", "model:read"]`)
- Middleware functions like `RequirePermission()` check permissions imperatively
- No support for custom roles or dynamic permission assignment
- No GraphQL-level permission enforcement

### Constraints

- Must maintain backward compatibility with existing JWT authentication
- Must support multi-tenancy via `org_name` field
- Must optimize for query performance (role_permissions table includes redundant org_name)
- System roles must be immutable and protected from modification/deletion

## Goals / Non-Goals

### Goals

- ✅ Implement Casbin RBAC model with user-role-permission hierarchy
- ✅ Add GraphQL `@hasPermission` directive for declarative operation authorization
- ✅ Create database schema for roles, user_roles, and role_permissions (multi-tenant aware)
- ✅ Hardcode system role permissions in Go code for performance and immutability
- ✅ Support tenant-scoped custom role creation with dynamic permissions
- ✅ Provide APIs for role management (CRUD) and user role assignment
- ✅ Integrate with existing JWT middleware to extract user identity

### Non-Goals

- ❌ Field-level permissions (may be added later, but out of scope for this change)
- ❌ Resource-level ownership checks (e.g., "user can only delete own posts") - this is application-level logic
- ❌ Migrate existing JWT permission checks immediately - coexistence is acceptable
- ❌ Casdoor integration for role synchronization (handled separately)
- ❌ Audit logging for permission changes (future enhancement)

## Decisions

### Decision 1: Casbin Model Design

**Choice**: Use simplified three-element model `(sub, obj, act)` without domain field.

**Rationale**:
- ModelCraft already handles multi-tenancy at application layer via `org_name` in context
- No need to duplicate tenant isolation in Casbin model
- Simplifies Casbin configuration and querying
- Application code filters roles and permissions by `org_name` before Casbin enforcement

**Model Configuration**:
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

### Decision 2: Hybrid Permission Storage

**Choice**: System roles have hardcoded permissions; custom roles store permissions in database.

**Rationale**:
- System roles (owner, admin, editor, viewer) have fixed, well-defined permissions
- Hardcoding improves performance (no DB query for system roles)
- Prevents accidental modification of critical permissions
- Custom roles need flexibility for tenant-specific needs

**Implementation**:
```go
var SystemRolePermissions = map[string][]Permission{
    "owner":  {{Obj: "*", Act: "*"}},  // Wildcard: all permissions
    "admin":  {{Obj: "project", Act: "*"}, {Obj: "model", Act: "*"}, ...},
    "editor": {{Obj: "project", Act: "read"}, {Obj: "model", Act: "create"}, ...},
    "viewer": {{Obj: "project", Act: "read"}, {Obj: "model", Act: "read"}, ...},
}
```

### Decision 3: Database Schema Design

**Choice**: Separate tables for roles, user_roles, and role_permissions with foreign key constraints.

**Rationale**:
- Clear separation of concerns (role definitions vs bindings vs permissions)
- Supports cascade deletion when roles are removed
- Enables efficient querying by org_name + role_name
- Redundant org_name in role_permissions improves query performance

**Tables**:
1. **roles**: Role metadata (name, description, is_system, org_name)
2. **user_roles**: User-to-role bindings (user_id, role_id, org_name)
3. **role_permissions**: Custom role permissions (role_id, org_name, obj, act)

### Decision 4: Permission Format

**Choice**: Use `resource:action` format (e.g., `project:create`, `model:delete`).

**Rationale**:
- Consistent with existing `CheckPermission()` function in middleware
- Intuitive and readable for AI assistants
- Supports wildcard matching (`project:*` matches all project operations)
- Aligns with industry standards (similar to AWS IAM, Kubernetes RBAC)

### Decision 5: GraphQL Directive Implementation

**Choice**: Implement `@hasPermission(action: String!)` directive that extracts user context and calls Casbin.

**Rationale**:
- Declarative syntax is AI-friendly and easy to maintain
- Colocates authorization logic with schema definitions
- Leverages gqlgen's built-in directive support
- Minimal boilerplate in resolver code

**Schema Example**:
```graphql
directive @hasPermission(action: String!) on FIELD_DEFINITION

type Mutation {
    createProject(input: ProjectInput!): ProjectPayload!
        @hasPermission(action: "project:create")
}
```

## Alternatives Considered

### Alternative 1: OPA (Open Policy Agent) instead of Casbin

**Rejected because**:
- Rego language is complex and less AI-friendly
- Higher learning curve for developers
- Casbin is simpler for RBAC use cases
- Better Go ecosystem integration

### Alternative 2: Store all role permissions in database (including system roles)

**Rejected because**:
- Performance overhead for querying system role permissions
- Risk of accidental modification/deletion of critical permissions
- System roles are fixed by design, no need for database storage

### Alternative 3: Single table for roles and permissions (denormalized)

**Rejected because**:
- Causes data duplication when same role assigned to multiple users
- Difficult to update role permissions (must update all user records)
- Poor support for cascade deletion

### Alternative 4: Include domain field in Casbin model

**Rejected because**:
- Adds unnecessary complexity (multi-tenancy already handled in app layer)
- Requires passing org_name to every Casbin enforce call
- No benefit since we filter roles by org_name before Casbin check

## Risks / Trade-offs

### Risk 1: Performance Impact

**Risk**: Every GraphQL operation requires permission check, may impact latency.

**Mitigation**:
- Enable Casbin's built-in caching (`enforcer.EnableCache(true)`)
- Hardcode system role permissions to avoid DB queries
- Use database indexes on (org_name, role_id) in role_permissions table
- Monitor query performance and optimize as needed

### Risk 2: Migration Complexity

**Risk**: Existing JWT permissions and new Casbin permissions coexist, causing confusion.

**Mitigation**:
- Document clear migration path in tasks.md
- Add feature flag to toggle between old and new permission systems
- Provide migration tool to convert JWT permissions to Casbin roles
- Test both systems in parallel before deprecating old approach

### Risk 3: System Role Modification

**Risk**: Developers might accidentally modify system role permissions in code.

**Mitigation**:
- Add database constraint `is_system = true` with UNIQUE(name) for system roles
- Block DELETE and UPDATE operations on is_system=true roles via API validation
- Add unit tests to verify system role permissions remain unchanged
- Document system roles as immutable in code comments

### Risk 4: Data Inconsistency

**Risk**: User assigned to role in tenant A, but role_permissions lacks org_name filtering.

**Mitigation**:
- Enforce org_name consistency via foreign key constraints
- Validate org_name match between user_roles and role_permissions in application logic
- Add database index on (role_id, org_name) for fast validation
- Include org_name in all permission queries

## Migration Plan

### Phase 1: Database Schema (Non-Breaking)

1. Create roles, user_roles, role_permissions tables via migration
2. Insert 4 system roles (owner, admin, editor, viewer) with is_system=true
3. Add foreign key constraints and indexes
4. Deploy schema changes to staging environment
5. Verify schema integrity and rollback plan

### Phase 2: Casbin Integration (Non-Breaking)

1. Implement Casbin enforcer singleton with hardcoded system role permissions
2. Add permission domain entities (Role, Permission)
3. Add repository interfaces and sqlc implementations
4. Add unit tests for enforcer and repositories
5. Deploy to staging without enabling directive

### Phase 3: GraphQL Directive (Non-Breaking)

1. Implement `@hasPermission` directive in internal/interfaces/graphql/
2. Update gqlgen.yml to register directive
3. Add directive to selected GraphQL operations (pilot phase)
4. Test directive with different roles and permissions
5. Deploy to staging and validate functionality

### Phase 4: Permission Management APIs (Non-Breaking)

1. Add GraphQL schema for role and permission management
2. Implement resolvers for custom role CRUD
3. Implement resolvers for user role assignment
4. Add validation to prevent system role modification
5. Deploy APIs to staging

### Phase 5: Full Rollout

1. Add @hasPermission directive to all GraphQL operations
2. Provide migration tool to convert JWT permissions to Casbin roles
3. Update documentation with permission management guide
4. Deploy to production with monitoring
5. Optionally deprecate old JWT permission checks after validation

### Rollback Strategy

- **Phase 1-2**: Drop tables if issues found (no user-facing impact)
- **Phase 3-4**: Remove directive from schema, revert gqlgen.yml (directive not enforced)
- **Phase 5**: Disable directive via feature flag, fall back to JWT permissions

## Open Questions

### Q1: Should we support permission inheritance between roles?

**Status**: Deferred to future enhancement. Current implementation uses flat role-permission model.

### Q2: Should tenant admins be able to modify system role permissions for their tenant?

**Status**: No. System roles are globally immutable. Tenants can create custom roles with desired permissions.

### Q3: How to handle conflicting roles (e.g., user has both admin and viewer roles)?

**Status**: Permissive approach - user gets union of all role permissions. If any role grants permission, access is allowed.

### Q4: Should we audit permission changes (who granted/revoked permissions)?

**Status**: Out of scope for this change. Can be added as separate audit logging feature.

### Q5: Should we support time-based permissions (e.g., temporary admin access)?

**Status**: Out of scope. User_roles table can be extended in future to include expires_at field.
