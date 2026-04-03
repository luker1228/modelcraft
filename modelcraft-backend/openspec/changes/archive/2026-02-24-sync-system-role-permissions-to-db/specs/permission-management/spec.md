## MODIFIED Requirements

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

## ADDED Requirements

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
