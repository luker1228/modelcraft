# permission-management Specification

## Purpose
TBD - created by archiving change implement-casbin-permissions. Update Purpose after archive.
## Requirements
### Requirement: Role Management Data Model

The system SHALL provide a data model for managing roles with fields: id, name, description, is_system, org_name, created_at, updated_at. System roles are globally defined and immutable; custom roles are tenant-scoped and mutable.

System roles use `org_name = '__SYSTEM__'` (not NULL) to satisfy MySQL's unique constraint on `(name, org_name)` — NULL values are not considered equal under MySQL's unique index semantics, which would allow duplicate system role names.

System roles are **globally shared across all organisations**: there are exactly 4 system role records in the `roles` table, and every organisation references them by `role_id` in `user_roles`. Per-organisation customisation of system role permissions is intentionally not supported.

#### Scenario: Role table schema

- **GIVEN** the database schema is initialized
- **WHEN** the roles table is created
- **THEN** the table has columns: id (PK), name (VARCHAR), description (TEXT), is_system (BOOLEAN), org_name (VARCHAR NOT NULL DEFAULT '__SYSTEM__'), created_at (TIMESTAMP), updated_at (TIMESTAMP)
- **AND** a unique constraint exists on (name, org_name)
- **AND** system roles have org_name='__SYSTEM__' and is_system=true

#### Scenario: System roles initialization

- **GIVEN** the database migration is executed
- **WHEN** the migration inserts system roles
- **THEN** four roles are inserted: "owner", "admin", "editor", "viewer"
- **AND** all have is_system=true and org_name='__SYSTEM__'
- **AND** each has a descriptive description field

#### Scenario: Custom role record

- **GIVEN** a tenant "org1" creates a custom role "data-analyst"
- **WHEN** the role is stored in the database
- **THEN** a record is created with name="data-analyst", org_name="org1", is_system=false
- **AND** the role is isolated to organization "org1"

### Requirement: User Role Assignment Data Model

The system SHALL provide a unified data model for binding users to roles AND managing membership lifecycle with fields: id, user_id, role_id, org_name, status, invited_by, invited_at, joined_at, created_at, updated_at. This replaces the previous separate `Membership` concept.

`user_roles.org_name` stores the **organisation the user belongs to** — it is always an actual organisation name (e.g. "org1") and is never `'__SYSTEM__'`. This field scopes the role assignment to a specific organisation, regardless of whether the role is a system role or a custom role.

Permission resolution for system roles follows this chain: `user_roles(user_id, org_name)` → `role_id` → `roles.name` → `SystemRolePermissions[name]` (in-memory map, no DB query).

#### Scenario: Extended user_roles table schema

- **GIVEN** the database schema is migrated
- **WHEN** the user_roles table structure is inspected
- **THEN** the table has all previous columns: id (PK), user_id (VARCHAR, FK to users.id), role_id (INT, FK to roles.id), org_name (VARCHAR), created_at (TIMESTAMP)
- **AND** new membership columns exist: status (VARCHAR DEFAULT 'active'), invited_by (VARCHAR nullable), invited_at (TIMESTAMP nullable), joined_at (TIMESTAMP nullable), updated_at (TIMESTAMP)
- **AND** a unique constraint exists on (user_id, role_id, org_name)
- **AND** foreign keys cascade on delete
- **AND** indexes exist on (status), (org_name, status)

#### Scenario: Active member record

- **GIVEN** user "user123" is directly assigned role "admin" (role_id=2) in organization "org1"
- **WHEN** the assignment is stored
- **THEN** a record is created with user_id="user123", role_id=2, org_name="org1", status="active"
- **AND** joined_at is set to the current timestamp
- **AND** invited_by, invited_at are NULL

#### Scenario: Invited member record

- **GIVEN** user "user456" is invited to organization "org1" with role "editor" by user "admin123"
- **WHEN** the invitation is created
- **THEN** a record is created with user_id="user456", role_id=3, org_name="org1", status="invited"
- **AND** invited_by="admin123"
- **AND** invited_at is set to the current timestamp
- **AND** joined_at is NULL

#### Scenario: Invitation acceptance

- **GIVEN** an invited user record exists with status="invited"
- **WHEN** the user accepts the invitation
- **THEN** status changes to "active"
- **AND** joined_at is set to the current timestamp
- **AND** invited_by and invited_at remain unchanged (audit trail)
- **AND** updated_at is updated

#### Scenario: Member suspension

- **GIVEN** an active user-role binding exists
- **WHEN** an admin suspends the member
- **THEN** status changes to "suspended"
- **AND** updated_at is updated
- **AND** the user loses access permissions immediately

#### Scenario: Member role change

- **GIVEN** an active user-role binding exists with role_id=3 (editor)
- **WHEN** an admin changes the role to viewer (role_id=4)
- **THEN** role_id is updated to 4
- **AND** status remains "active"
- **AND** updated_at is updated
- **AND** permissions change immediately

#### Scenario: List organization members

- **GIVEN** organization "org1" has 5 user-role bindings with various statuses
- **WHEN** querying for all members of "org1"
- **THEN** all 5 records are returned
- **AND** records include status, invited_at, joined_at fields
- **AND** results can be filtered by status (e.g., only "active" members)

#### Scenario: Cascade deletion when role is deleted

- **GIVEN** role_id=5 has 3 user-role bindings
- **WHEN** role_id=5 is deleted
- **THEN** all 3 user-role bindings are automatically deleted via CASCADE
- **AND** no orphaned records remain

#### Scenario: Same system role assigned across multiple organisations

- **GIVEN** system role "admin" has role_id=2 in the roles table
- **WHEN** user "alice" is assigned admin in "org1" AND user "bob" is assigned admin in "org2"
- **THEN** two user_roles records exist: (alice, role_id=2, org_name="org1") and (bob, role_id=2, org_name="org2")
- **AND** both reference the same role_id=2 record
- **AND** org_name in both records is the actual organisation name, never '__SYSTEM__'
- **AND** alice's permissions in org1 are independent of bob's permissions in org2

#### Scenario: Same user holds different roles in different organisations

- **GIVEN** user "alice" is assigned "admin" in "org1" and "viewer" in "org2"
- **WHEN** the system resolves alice's permissions for org1
- **THEN** alice has admin permissions (from SystemRolePermissions["admin"])
- **AND** when resolving alice's permissions for org2, she has viewer permissions only
- **AND** the two assignments coexist without conflict due to the unique key on (user_id, role_id, org_name)

### Requirement: Role Permission Data Model

The system SHALL provide a data model for storing role permissions with fields: id, role_id, org_name, obj, act, created_at. Custom role permissions are stored and managed via CRUD operations. System role permissions are authoritative in Go code (`system_roles.go`) and are written to the database as a **read-only snapshot** on every application startup; the snapshot is forcefully reset (delete-then-insert) on each boot, so manual DB edits to system role permissions have no effect at runtime.

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

#### Scenario: System role permissions snapshot written on startup

- **GIVEN** the application starts up after DB migrations complete
- **WHEN** the system role permissions syncer runs
- **THEN** for each system role (owner, admin, editor, viewer), all existing `role_permissions` rows for that role_id are deleted
- **AND** the current permissions from `SystemRolePermissions` in Go code are inserted
- **AND** a log line is emitted: "System role permissions synced: role=<name>, count=<n>"
- **AND** the syncer is idempotent: running it twice produces the same final state without error

#### Scenario: System role permissions snapshot is read-only at runtime

- **GIVEN** a DB operator manually edits a `role_permissions` row for system role "admin"
- **WHEN** the next application startup occurs
- **THEN** the syncer resets the row back to the code-defined permissions
- **AND** runtime permission checks continue to use the in-memory `SystemRolePermissions` map (no behavioral change)

#### Scenario: System role permissions visible in DB for display

- **GIVEN** the application has started and the syncer has run
- **WHEN** querying `role_permissions` table for a system role's role_id
- **THEN** records are found matching the code-defined permissions
- **AND** the `rolePermissionsList` GraphQL query returns these permissions (served from in-memory map, not DB read)

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

### Requirement: UserRole Entity Membership Methods

The UserRole entity SHALL provide methods for managing membership lifecycle: AcceptInvitation(), ChangeRole(), Suspend(), Activate(), and IsActive(). These methods enforce business rules and state transitions.

#### Scenario: Accept invitation method

- **GIVEN** a UserRole entity with status="invited"
- **WHEN** AcceptInvitation() is called
- **THEN** status changes to "active"
- **AND** joined_at is set to current time
- **AND** method returns nil error

#### Scenario: Accept invitation on non-invited status fails

- **GIVEN** a UserRole entity with status="active"
- **WHEN** AcceptInvitation() is called
- **THEN** method returns error "can only accept invitation when status is 'invited'"
- **AND** status remains unchanged

#### Scenario: Change role method

- **GIVEN** a UserRole entity with role_id=2
- **WHEN** ChangeRole(4) is called
- **THEN** role_id changes to 4
- **AND** updated_at is set to current time
- **AND** method returns nil error

#### Scenario: Change role with invalid id fails

- **GIVEN** a UserRole entity
- **WHEN** ChangeRole(0) is called with invalid role_id
- **THEN** method returns error "role_id must be positive"
- **AND** role_id remains unchanged

#### Scenario: Suspend method

- **GIVEN** a UserRole entity with status="active"
- **WHEN** Suspend() is called
- **THEN** status changes to "suspended"
- **AND** updated_at is updated

#### Scenario: Activate method

- **GIVEN** a UserRole entity with status="suspended"
- **WHEN** Activate() is called
- **THEN** status changes to "active"
- **AND** updated_at is updated

#### Scenario: IsActive check

- **GIVEN** a UserRole entity with status="active"
- **WHEN** IsActive() is called
- **THEN** method returns true

#### Scenario: IsActive check on suspended

- **GIVEN** a UserRole entity with status="suspended"
- **WHEN** IsActive() is called
- **THEN** method returns false

### Requirement: Extended UserRoleRepository Interface

The UserRoleRepository interface SHALL include methods for membership management operations in addition to role assignment: CreateInvitation, AcceptInvitation, UpdateMembershipStatus, ListOrgMembers, ListUserMemberships, DeleteMembership, DeleteUserRolesByOrg.

#### Scenario: CreateInvitation repository method

- **GIVEN** a UserRole entity with status="invited"
- **WHEN** CreateInvitation(ctx, invitation) is called
- **THEN** the invitation is persisted to the database
- **AND** invited_by, invited_at fields are stored
- **AND** method returns nil error

#### Scenario: AcceptInvitation repository method

- **GIVEN** an invited user-role record exists in database
- **WHEN** AcceptInvitation(ctx, userID, orgName) is called
- **THEN** the record's status is updated to "active"
- **AND** joined_at is set to current timestamp
- **AND** method returns nil error

#### Scenario: UpdateMembershipStatus repository method

- **GIVEN** an active user-role record exists
- **WHEN** UpdateMembershipStatus(ctx, userID, orgName, "suspended") is called
- **THEN** the record's status is updated to "suspended"
- **AND** updated_at is updated
- **AND** method returns nil error

#### Scenario: ListOrgMembers repository method

- **GIVEN** organization "org1" has 5 members with different statuses
- **WHEN** ListOrgMembers(ctx, "org1") is called
- **THEN** all 5 UserRole records are returned
- **AND** records include all membership fields (status, invited_at, joined_at)

#### Scenario: ListUserMemberships repository method

- **GIVEN** user "user123" has memberships in 3 organizations
- **WHEN** ListUserMemberships(ctx, "user123") is called
- **THEN** all 3 UserRole records are returned
- **AND** records span different org_name values

#### Scenario: DeleteMembership repository method

- **GIVEN** a user-role record with id=10 exists
- **WHEN** DeleteMembership(ctx, 10) is called
- **THEN** the record is deleted from the database
- **AND** method returns nil error

#### Scenario: DeleteUserRolesByOrg repository method

- **GIVEN** organization "org1" has 5 user-role bindings
- **WHEN** DeleteUserRolesByOrg(ctx, "org1") is called
- **THEN** all 5 records are deleted
- **AND** method returns nil error
- **AND** this supports organization deletion cascade

### Requirement: System Role Permissions Startup Syncer

The system SHALL provide a `SystemRolePermissionsSyncer` service in the application layer that synchronises the in-memory `SystemRolePermissions` definition to the `role_permissions` table on every application startup.

#### Scenario: Syncer runs successfully on first boot

- **GIVEN** the `role_permissions` table is empty
- **WHEN** `SystemRolePermissionsSyncer.Sync(ctx)` is called
- **THEN** rows are inserted for all permissions of all four system roles
- **AND** the method returns nil

#### Scenario: Syncer resets permissions on subsequent boots

- **GIVEN** `role_permissions` contains system role rows from a previous version (possibly different permissions)
- **WHEN** `SystemRolePermissionsSyncer.Sync(ctx)` is called
- **THEN** stale rows are deleted and current code-defined rows are inserted
- **AND** custom role permission rows are untouched
- **AND** the method returns nil

#### Scenario: Syncer skips role if system role not found in DB

- **GIVEN** a system role name exists in code but its record is missing from the `roles` table
- **WHEN** `SystemRolePermissionsSyncer.Sync(ctx)` is called
- **THEN** the syncer logs a warning and continues with the remaining roles
- **AND** the method returns nil (non-fatal)

#### Scenario: Syncer called during application bootstrap

- **GIVEN** the application is starting up
- **WHEN** DB migrations complete
- **THEN** `SystemRolePermissionsSyncer.Sync(ctx)` is invoked before the HTTP server begins accepting requests
- **AND** a startup error from the syncer aborts the boot sequence

