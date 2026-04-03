## ADDED Requirements

### Requirement: Role Management Data Model

The system SHALL provide a data model for managing roles with fields: id, name, description, is_system, org_name, created_at, updated_at. System roles are globally immutable; custom roles are tenant-scoped and mutable.

#### Scenario: Role table schema

- **GIVEN** the database schema is initialized
- **WHEN** the roles table is created
- **THEN** the table has columns: id (PK), name (VARCHAR), description (TEXT), is_system (BOOLEAN), org_name (VARCHAR, nullable), created_at (TIMESTAMP), updated_at (TIMESTAMP)
- **AND** a unique constraint exists on (name, org_name)
- **AND** system roles have org_name=NULL and is_system=true

#### Scenario: System roles initialization

- **GIVEN** the database migration is executed
- **WHEN** the migration inserts system roles
- **THEN** four roles are inserted: "owner", "admin", "editor", "viewer"
- **AND** all have is_system=true and org_name=NULL
- **AND** each has a descriptive description field

#### Scenario: Custom role record

- **GIVEN** a tenant "org1" creates a custom role "data-analyst"
- **WHEN** the role is stored in the database
- **THEN** a record is created with name="data-analyst", org_name="org1", is_system=false
- **AND** the role is isolated to organization "org1"

### Requirement: User Role Assignment Data Model

The system SHALL provide a data model for binding users to roles with fields: id, user_id, role_id, org_name, created_at. User-role bindings are tenant-scoped and support cascade deletion.

#### Scenario: User_roles table schema

- **GIVEN** the database schema is initialized
- **WHEN** the user_roles table is created
- **THEN** the table has columns: id (PK), user_id (VARCHAR, FK to users.id), role_id (INT, FK to roles.id), org_name (VARCHAR), created_at (TIMESTAMP)
- **AND** a unique constraint exists on (user_id, role_id, org_name)
- **AND** foreign key on role_id references roles.id with ON DELETE CASCADE
- **AND** foreign key on user_id references users.id with ON DELETE CASCADE

#### Scenario: User assigned to role in tenant

- **GIVEN** user "user123" is assigned role "admin" (role_id=2) in organization "org1"
- **WHEN** the assignment is stored
- **THEN** a record is created with user_id="user123", role_id=2, org_name="org1"
- **AND** the same user can be assigned different roles in different organizations

#### Scenario: Cascade deletion when role is deleted

- **GIVEN** a custom role "data-analyst" is assigned to 5 users in "org1"
- **WHEN** the role is deleted
- **THEN** all 5 user_roles records are automatically deleted via CASCADE
- **AND** users lose the associated permissions

### Requirement: Role Permission Data Model

The system SHALL provide a data model for storing custom role permissions with fields: id, role_id, org_name, obj, act, created_at. System role permissions are hardcoded in Go code and not stored in the database.

#### Scenario: Role_permissions table schema

- **GIVEN** the database schema is initialized
- **WHEN** the role_permissions table is created
- **THEN** the table has columns: id (PK), role_id (INT, FK to roles.id), org_name (VARCHAR), obj (VARCHAR), act (VARCHAR), created_at (TIMESTAMP)
- **AND** a unique constraint exists on (role_id, obj, act)
- **AND** an index exists on (org_name, role_id) for query optimization
- **AND** foreign key on role_id references roles.id with ON DELETE CASCADE

#### Scenario: Custom role permission record

- **GIVEN** a custom role "data-analyst" (role_id=10) in "org1" has permission "model:read"
- **WHEN** the permission is stored
- **THEN** a record is created with role_id=10, org_name="org1", obj="model", act="read"
- **AND** the redundant org_name field improves query performance

#### Scenario: System role permissions not stored

- **GIVEN** system role "admin" has permissions defined in code
- **WHEN** querying role_permissions table for role_id=2 (admin)
- **THEN** no records are found
- **AND** permissions are loaded from SystemRolePermissions map in Go code

### Requirement: Create Custom Role API

The system SHALL provide a GraphQL mutation to create custom roles with validation to prevent system role name collisions and enforce tenant isolation.

#### Scenario: Create custom role success

- **GIVEN** an admin user in organization "org1"
- **WHEN** the user calls `createRole(input: {name: "data-analyst", description: "Read-only data access", orgName: "org1"})`
- **THEN** a role record is created with is_system=false
- **AND** the mutation returns the created role with ID
- **AND** the role is only visible within "org1"

#### Scenario: Prevent system role name collision

- **GIVEN** an admin user attempts to create a custom role
- **WHEN** the role name is "owner", "admin", "editor", or "viewer"
- **THEN** the mutation returns error "SystemRoleCannotBeCreated"
- **AND** the error message is "Cannot create custom role with system role name: {name}"
- **AND** no role record is created

#### Scenario: Role name uniqueness within tenant

- **GIVEN** organization "org1" already has a custom role "data-analyst"
- **WHEN** admin creates another role with name "data-analyst" in "org1"
- **THEN** the mutation returns error "RoleAlreadyExists"
- **AND** the error message is "Role 'data-analyst' already exists in organization 'org1'"

#### Scenario: Same role name in different tenants

- **GIVEN** organization "org1" has a custom role "data-analyst"
- **WHEN** admin in organization "org2" creates a role with name "data-analyst"
- **THEN** the role is created successfully in "org2"
- **AND** both roles coexist with different org_name values

### Requirement: Update Custom Role API

The system SHALL provide a GraphQL mutation to update custom role properties (name, description) with protection against modifying system roles.

#### Scenario: Update custom role success

- **GIVEN** a custom role "data-analyst" (id=10) exists in "org1"
- **WHEN** admin calls `updateRole(id: 10, input: {description: "Updated description"})`
- **THEN** the role description is updated
- **AND** the mutation returns the updated role
- **AND** updated_at timestamp is refreshed

#### Scenario: Prevent system role modification

- **GIVEN** system role "admin" (id=2, is_system=true)
- **WHEN** admin calls `updateRole(id: 2, input: {name: "super-admin"})`
- **THEN** the mutation returns error "SystemRoleCannotBeModified"
- **AND** the error message is "System role 'admin' cannot be modified"
- **AND** no changes are made to the database

#### Scenario: Rename custom role with uniqueness check

- **GIVEN** custom role "data-analyst" exists in "org1"
- **AND** custom role "researcher" also exists in "org1"
- **WHEN** admin renames "data-analyst" to "researcher"
- **THEN** the mutation returns error "RoleAlreadyExists"
- **AND** the rename operation is aborted

### Requirement: Delete Custom Role API

The system SHALL provide a GraphQL mutation to delete custom roles with cascade cleanup of user_roles and role_permissions, and protection against deleting system roles.

#### Scenario: Delete custom role success

- **GIVEN** a custom role "data-analyst" (id=10) exists in "org1"
- **AND** the role is assigned to 3 users
- **AND** the role has 5 permissions
- **WHEN** admin calls `deleteRole(id: 10)`
- **THEN** the role record is deleted
- **AND** all 3 user_roles records are deleted via CASCADE
- **AND** all 5 role_permissions records are deleted via CASCADE
- **AND** the mutation returns success message

#### Scenario: Prevent system role deletion

- **GIVEN** system role "admin" (id=2, is_system=true)
- **WHEN** admin calls `deleteRole(id: 2)`
- **THEN** the mutation returns error "SystemRoleCannotBeDeleted"
- **AND** the error message is "System role 'admin' cannot be deleted"
- **AND** no changes are made to the database

### Requirement: List Roles API

The system SHALL provide a GraphQL query to list roles with optional filtering by organization and system/custom role type.

#### Scenario: List all roles including system roles

- **GIVEN** organization "org1" has 2 custom roles
- **WHEN** admin calls `roles(orgName: "org1", includeSystem: true)`
- **THEN** the query returns 6 roles (4 system + 2 custom)
- **AND** system roles have is_system=true and org_name=null
- **AND** custom roles have is_system=false and org_name="org1"

#### Scenario: List only custom roles for tenant

- **GIVEN** organization "org1" has 2 custom roles
- **WHEN** admin calls `roles(orgName: "org1", includeSystem: false)`
- **THEN** the query returns 2 roles (only custom roles)
- **AND** system roles are excluded

#### Scenario: Tenant isolation in role listing

- **GIVEN** organization "org1" has custom roles ["data-analyst", "researcher"]
- **AND** organization "org2" has custom roles ["engineer"]
- **WHEN** admin in "org1" calls `roles(orgName: "org1")`
- **THEN** only roles from "org1" and system roles are returned
- **AND** "engineer" role from "org2" is not visible

### Requirement: Add Permission to Role API

The system SHALL provide a GraphQL mutation to add permissions to custom roles with validation to prevent modifying system role permissions.

#### Scenario: Add permission to custom role

- **GIVEN** a custom role "data-analyst" (id=10) in "org1"
- **WHEN** admin calls `addPermissionToRole(roleId: 10, obj: "model", act: "read")`
- **THEN** a record is inserted into role_permissions with role_id=10, obj="model", act="read", org_name="org1"
- **AND** the mutation returns success
- **AND** users with this role can now perform "model:read" action

#### Scenario: Prevent duplicate permission

- **GIVEN** role "data-analyst" already has permission "model:read"
- **WHEN** admin calls `addPermissionToRole(roleId: 10, obj: "model", act: "read")`
- **THEN** the mutation returns error "PermissionAlreadyExists"
- **AND** no duplicate record is created

#### Scenario: Prevent adding permission to system role

- **GIVEN** system role "admin" (id=2, is_system=true)
- **WHEN** admin calls `addPermissionToRole(roleId: 2, obj: "project", act: "create")`
- **THEN** the mutation returns error "SystemRoleCannotBeModified"
- **AND** no permission record is created

### Requirement: Remove Permission from Role API

The system SHALL provide a GraphQL mutation to remove permissions from custom roles with protection against modifying system role permissions.

#### Scenario: Remove permission from custom role

- **GIVEN** custom role "data-analyst" has permission "model:read"
- **WHEN** admin calls `removePermissionFromRole(roleId: 10, obj: "model", act: "read")`
- **THEN** the permission record is deleted from role_permissions
- **AND** the mutation returns success
- **AND** users with this role can no longer perform "model:read" action

#### Scenario: Remove non-existent permission

- **GIVEN** custom role "data-analyst" does not have permission "model:delete"
- **WHEN** admin calls `removePermissionFromRole(roleId: 10, obj: "model", act: "delete")`
- **THEN** the mutation returns error "PermissionNotFound"
- **AND** no changes are made to the database

#### Scenario: Prevent removing permission from system role

- **GIVEN** system role "admin" (id=2, is_system=true)
- **WHEN** admin calls `removePermissionFromRole(roleId: 2, obj: "project", act: "create")`
- **THEN** the mutation returns error "SystemRoleCannotBeModified"
- **AND** no changes are made (system role permissions are hardcoded)

### Requirement: Assign Role to User API

The system SHALL provide a GraphQL mutation to assign roles to users with tenant-scoped bindings and duplicate prevention.

#### Scenario: Assign role to user success

- **GIVEN** user "user123" exists in organization "org1"
- **AND** custom role "data-analyst" (id=10) exists in "org1"
- **WHEN** admin calls `assignRoleToUser(userId: "user123", roleId: 10, orgName: "org1")`
- **THEN** a record is created in user_roles with user_id="user123", role_id=10, org_name="org1"
- **AND** the mutation returns success
- **AND** user "user123" gains permissions associated with "data-analyst" role

#### Scenario: Prevent duplicate role assignment

- **GIVEN** user "user123" already has role "data-analyst" in "org1"
- **WHEN** admin calls `assignRoleToUser(userId: "user123", roleId: 10, orgName: "org1")`
- **THEN** the mutation returns error "UserRoleAlreadyExists"
- **AND** no duplicate record is created

#### Scenario: Same user different roles in different tenants

- **GIVEN** user "alice" has role "admin" in "org1"
- **WHEN** admin in "org2" assigns role "viewer" to "alice"
- **THEN** two user_roles records exist: ("alice", "admin", "org1") and ("alice", "viewer", "org2")
- **AND** alice has different permissions in each organization

### Requirement: Revoke Role from User API

The system SHALL provide a GraphQL mutation to revoke role assignments from users with tenant-scoped removal.

#### Scenario: Revoke role from user success

- **GIVEN** user "user123" has role "data-analyst" in "org1"
- **WHEN** admin calls `revokeRoleFromUser(userId: "user123", roleId: 10, orgName: "org1")`
- **THEN** the user_roles record is deleted
- **AND** the mutation returns success
- **AND** user "user123" loses permissions associated with "data-analyst" role

#### Scenario: Revoke non-existent role assignment

- **GIVEN** user "user123" does not have role "data-analyst" in "org1"
- **WHEN** admin calls `revokeRoleFromUser(userId: "user123", roleId: 10, orgName: "org1")`
- **THEN** the mutation returns error "UserRoleNotFound"
- **AND** no changes are made to the database

### Requirement: List User Roles API

The system SHALL provide a GraphQL query to list all roles assigned to a user within a specific organization.

#### Scenario: List user roles in tenant

- **GIVEN** user "alice" has roles ["admin", "data-analyst"] in "org1"
- **WHEN** admin calls `userRoles(userId: "alice", orgName: "org1")`
- **THEN** the query returns 2 roles with details (name, description, is_system)
- **AND** roles from other organizations are excluded

#### Scenario: User with no roles

- **GIVEN** user "bob" has no role assignments in "org1"
- **WHEN** admin calls `userRoles(userId: "bob", orgName: "org1")`
- **THEN** the query returns an empty array
- **AND** no error is raised

### Requirement: List Role Permissions API

The system SHALL provide a GraphQL query to list all permissions granted by a role, merging hardcoded system role permissions with database-stored custom role permissions.

#### Scenario: List system role permissions

- **GIVEN** role "admin" (id=2, is_system=true)
- **WHEN** admin calls `rolePermissions(roleId: 2)`
- **THEN** the query returns permissions from SystemRolePermissions["admin"] hardcoded map
- **AND** includes permissions like [{obj: "project", act: "*"}, {obj: "model", act: "*"}]

#### Scenario: List custom role permissions

- **GIVEN** custom role "data-analyst" (id=10) has permissions [{obj: "model", act: "read"}, {obj: "project", act: "read"}] in database
- **WHEN** admin calls `rolePermissions(roleId: 10)`
- **THEN** the query returns permissions from role_permissions table
- **AND** includes [{obj: "model", act: "read"}, {obj: "project", act: "read"}]

### Requirement: Permission Management Authorization

The system SHALL restrict role and permission management operations to users with admin or owner roles using the @hasPermission directive.

#### Scenario: Admin can manage roles

- **GIVEN** user "alice" has role "admin" in "org1"
- **WHEN** alice calls `createRole(input: {...})`
- **THEN** the @hasPermission(action: "role:manage") directive passes
- **AND** the operation succeeds

#### Scenario: Viewer cannot manage roles

- **GIVEN** user "bob" has role "viewer" in "org1"
- **WHEN** bob calls `createRole(input: {...})`
- **THEN** the @hasPermission(action: "role:manage") directive fails
- **AND** the mutation returns error "permission denied: requires 'role:manage' in organization 'org1'"

#### Scenario: Owner can manage roles globally

- **GIVEN** user "charlie" has role "owner" (wildcard permissions)
- **WHEN** charlie calls any role management mutation
- **THEN** all @hasPermission directives pass via wildcard permission
- **AND** operations succeed
