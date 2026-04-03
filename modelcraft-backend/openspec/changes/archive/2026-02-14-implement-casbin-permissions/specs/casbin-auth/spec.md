## ADDED Requirements

### Requirement: Casbin RBAC Model Configuration

The system SHALL use Casbin as the policy engine for role-based access control with a simplified three-element model (subject, object, action) without domain field, since multi-tenancy is handled at application layer.

#### Scenario: Casbin enforcer initialization

- **GIVEN** the application starts
- **WHEN** Casbin enforcer is initialized
- **THEN** the enforcer loads the RBAC model from `casbin_model.conf`
- **AND** the enforcer enables caching for performance optimization
- **AND** the enforcer loads system role permissions from hardcoded definitions

#### Scenario: System role permission enforcement

- **GIVEN** a user with role "admin" in organization "org1"
- **WHEN** the user attempts action "project:create"
- **THEN** the enforcer checks hardcoded system role permissions
- **AND** grants access if "admin" role includes "project:create" permission
- **AND** denies access otherwise

### Requirement: System Role Definitions

The system SHALL define four immutable system roles with hardcoded permissions: owner, admin, editor, and viewer. System roles cannot be modified or deleted.

#### Scenario: Owner role permissions

- **GIVEN** a user assigned the "owner" role
- **WHEN** the user attempts any action on any resource
- **THEN** access is granted via wildcard permission (obj: "*", act: "*")

#### Scenario: Admin role permissions

- **GIVEN** a user assigned the "admin" role
- **WHEN** the user attempts create, read, update, or delete operations on project, model, cluster, or enum resources
- **THEN** access is granted
- **AND** user management operations are denied (reserved for owner)

#### Scenario: Editor role permissions

- **GIVEN** a user assigned the "editor" role
- **WHEN** the user attempts read or update operations on project, model, cluster, or enum resources
- **THEN** access is granted
- **AND** create operations are allowed for model and enum
- **AND** delete operations are denied

#### Scenario: Viewer role permissions

- **GIVEN** a user assigned the "viewer" role
- **WHEN** the user attempts read operations on project, model, cluster, or enum resources
- **THEN** access is granted
- **AND** all write operations (create, update, delete) are denied

### Requirement: Custom Role Support

The system SHALL allow tenants to create custom roles with tenant-specific permissions. Custom role permissions are stored in the database and scoped to the tenant's organization.

#### Scenario: Create custom role for tenant

- **GIVEN** an admin user in organization "org1"
- **WHEN** the user creates a custom role "data-analyst" with permissions ["model:read", "project:read"]
- **THEN** a role record is created with org_name="org1" and is_system=false
- **AND** permissions are stored in role_permissions table with org_name="org1"

#### Scenario: Custom role permissions enforcement

- **GIVEN** a user assigned custom role "data-analyst" in organization "org1"
- **WHEN** the user attempts action "model:read"
- **THEN** the enforcer queries role_permissions table filtered by org_name="org1" and role_id
- **AND** grants access if permission exists
- **AND** denies access otherwise

#### Scenario: System role name collision prevention

- **GIVEN** an admin user attempts to create a custom role
- **WHEN** the role name is "owner", "admin", "editor", or "viewer"
- **THEN** the system returns error "Cannot create custom role with system role name"
- **AND** no role record is created

### Requirement: Multi-Tenant Role Assignment

The system SHALL support assigning users to different roles in different organizations. A user can have role "admin" in org1 and role "viewer" in org2.

#### Scenario: User with different roles in different tenants

- **GIVEN** user "alice" is assigned role "admin" in organization "org1"
- **AND** user "alice" is assigned role "viewer" in organization "org2"
- **WHEN** alice attempts action "project:create" in context of "org1"
- **THEN** access is granted based on "admin" role permissions
- **WHEN** alice attempts action "project:create" in context of "org2"
- **THEN** access is denied based on "viewer" role permissions

#### Scenario: User with multiple roles in same tenant

- **GIVEN** user "bob" is assigned roles "editor" and "viewer" in organization "org1"
- **WHEN** bob attempts action "model:update" in context of "org1"
- **THEN** the enforcer computes union of all role permissions
- **AND** grants access if any role grants the permission (permissive approach)

### Requirement: Permission Check Integration

The system SHALL integrate with JWT authentication middleware to extract user identity (user_id, org_name) and perform Casbin permission checks before allowing GraphQL operations.

#### Scenario: Permission check with valid authentication

- **GIVEN** a valid JWT token with user_id="user123" and org_name="org1"
- **AND** user is assigned role "admin" in "org1"
- **WHEN** the user invokes a GraphQL mutation with @hasPermission(action: "project:create")
- **THEN** the system extracts user_id and org_name from context
- **AND** queries user's roles in "org1" from user_roles table
- **AND** loads permissions for "admin" role from hardcoded definitions
- **AND** calls enforcer.Enforce(user_id, "project", "create")
- **AND** grants access if enforcer returns true

#### Scenario: Permission check with missing authentication

- **GIVEN** a GraphQL request without JWT token
- **WHEN** the user invokes a GraphQL mutation with @hasPermission(action: "project:create")
- **THEN** the system attempts to extract user_id from context
- **AND** returns error "permission denied: user not authenticated"
- **AND** the resolver is not executed

#### Scenario: Permission denied response

- **GIVEN** a user with role "viewer" in organization "org1"
- **WHEN** the user attempts action "project:delete"
- **THEN** the enforcer denies access
- **AND** the system returns GraphQL error "permission denied: requires 'project:delete' in organization 'org1'"

### Requirement: Performance Optimization

The system SHALL optimize permission check performance through caching and query optimization to minimize latency impact on GraphQL operations.

#### Scenario: Casbin cache enabled

- **GIVEN** Casbin enforcer is initialized
- **WHEN** the first permission check is performed for user "alice" with action "project:read"
- **THEN** the result is cached in memory
- **WHEN** a subsequent permission check is performed for the same user and action within the same request
- **THEN** the cached result is returned without re-querying the database

#### Scenario: System role permission hardcoded

- **GIVEN** a user with system role "admin"
- **WHEN** a permission check is performed
- **THEN** permissions are loaded from hardcoded SystemRolePermissions map
- **AND** no database query is executed for system role permissions
- **AND** response latency is minimized

#### Scenario: Database query optimization

- **GIVEN** a user with custom role in organization "org1"
- **WHEN** permissions are queried from role_permissions table
- **THEN** the query uses index on (org_name, role_id)
- **AND** query execution time is under 10ms for typical role permission sets
