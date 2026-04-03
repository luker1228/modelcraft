# Change: Implement Casbin-based Operation-Level Permissions with GraphQL Directives

## Why

ModelCraft currently lacks fine-grained operation-level authorization control. Users need a declarative, AI-friendly permission system that supports:
- Multi-tenant role-based access control (RBAC)
- System-wide predefined roles (owner, admin, editor, viewer) and tenant-specific custom roles
- GraphQL operation-level permission checks using declarative directives

The current JWT-based permission system embeds permissions directly in tokens, which is inflexible and difficult to manage dynamically. This change introduces Casbin as the policy engine with GraphQL directives as the enforcement layer.

## What Changes

- **Add Casbin integration** for policy-based access control using RBAC model
- **Add GraphQL `@hasPermission` directive** for declarative operation-level authorization
- **Add database tables** for roles, user-role bindings, and role permissions (multi-tenant aware)
- **Add permission management APIs** for custom role CRUD, user role assignment, and permission queries
- **Add system roles with hardcoded permissions** (owner, admin, editor, viewer) that cannot be modified
- **Integrate with existing JWT authentication** by extracting user_id and org_name from context

### Key Design Decisions

1. **Simplified Casbin Model**: Three-element model `(sub, obj, act)` without domain field, since multi-tenancy is handled at application layer via org_name
2. **Hybrid Permission Storage**: System role permissions hardcoded in Go code; custom role permissions stored in database
3. **Role Classification**: System roles (global, immutable) vs. tenant custom roles (org_name scoped, mutable)
4. **Permission Format**: `resource:action` format (e.g., `project:create`, `model:read`) consistent with existing middleware
5. **Query Optimization**: role_permissions table includes redundant org_name field for faster lookups

### Breaking Changes

None. This is additive functionality. Existing JWT permission checks remain functional during migration phase.

## Impact

### Affected Specs
- **NEW**: `casbin-auth` - Casbin enforcer integration and role-permission model
- **NEW**: `graphql-directive` - @hasPermission directive implementation
- **NEW**: `permission-management` - Role and permission management APIs
- **MODIFIED**: `jwt-middleware` - Enhanced to support Casbin integration alongside existing permission checks

### Affected Code
- `internal/infrastructure/auth/` - Add Casbin enforcer, system role definitions
- `internal/domain/permission/` - Add role, permission domain entities
- `internal/app/permission/` - Add role and permission management services
- `internal/interfaces/graphql/` - Add directive implementation and permission helpers
- `api/graph/schema/` - Add permission directive definition and permission management schema
- `db/schema/mysql/` - Add roles, user_roles, role_permissions tables
- `gqlgen.yml` - Register @hasPermission directive

### Database Changes
- **NEW TABLE**: `roles` - Role definitions (system + custom)
- **NEW TABLE**: `user_roles` - User-to-role bindings (multi-tenant)
- **NEW TABLE**: `role_permissions` - Custom role permissions (system roles hardcoded)

### Migration Strategy
1. Deploy database schema changes
2. Initialize system roles in database
3. Implement Casbin integration and directive
4. Gradually migrate operations to use @hasPermission directive
5. Optionally deprecate old JWT permission checks after full migration

## Dependencies

- **Casbin Go SDK**: `github.com/casbin/casbin/v2` (policy engine)
- **gqlgen**: Already used, will register custom directive

## Timeline Estimate

- **Phase 1** (Database + Casbin Core): 2-3 days
- **Phase 2** (GraphQL Directive): 1-2 days
- **Phase 3** (Permission Management APIs): 2-3 days
- **Phase 4** (Testing + Documentation): 1-2 days

**Total**: ~6-10 days
