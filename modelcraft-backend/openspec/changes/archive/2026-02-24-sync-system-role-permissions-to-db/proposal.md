# Change: Sync System Role Permissions to DB on Startup

## Why

System role permissions are currently hardcoded in Go (`system_roles.go`) and never written to the database. The `rolePermissionsList` GraphQL query works for display purposes by reading from memory, but the `role_permissions` table contains no records for system roles. This creates an inconsistency: management UIs that read the DB directly see empty permissions for system roles, and there is no audit trail of what permissions each system role had at any given deployment version.

The agreed design is: **code is the authoritative source for system role permissions; DB holds a read-only snapshot for display and audit**. On every startup the snapshot is forcefully reset to match the code — meaning manual DB edits to system role permissions have no effect at runtime.

## Data Model Clarification: How Users Are Associated with System Roles

System roles are **globally defined but organisation-scoped in assignment**:

```
roles table (system roles — global, shared across all orgs)
id | name    | is_system | org_name
1  | owner   | TRUE      | __SYSTEM__
2  | admin   | TRUE      | __SYSTEM__
3  | editor  | TRUE      | __SYSTEM__
4  | viewer  | TRUE      | __SYSTEM__

user_roles table (assignment — per org)
id | user_id | role_id | org_name   ← org_name here is the actual org, NOT __SYSTEM__
1  | alice   | 2       | org1       ← alice is admin in org1
2  | alice   | 4       | org2       ← alice is viewer in org2
3  | bob     | 1       | org1       ← bob is owner in org1
```

Key rules:
- There are exactly 4 system role records in `roles`, shared by all organisations (`org_name = '__SYSTEM__'`).
- `user_roles.role_id` points directly to one of these 4 records.
- `user_roles.org_name` stores the **organisation the user belongs to** — it is never `'__SYSTEM__'`.
- The same `role_id` (e.g. admin) can appear in `user_roles` for many different organisations; each row is independent.
- Permission checks resolve: `user_roles(user_id, org_name) → role_id → roles.name → SystemRolePermissions[name]` (in-memory, no DB query for system roles).
- System roles are globally defined; per-organisation customisation of system role permissions is intentionally not supported.

The `org_name = '__SYSTEM__'` sentinel (instead of NULL) is used in `roles` to satisfy MySQL's unique constraint on `(name, org_name)` — NULL values are not considered equal under MySQL's unique index semantics.

## What Changes

- On application startup, after DB migrations run, a new `SystemRolePermissionsSyncer` upserts all system role permissions into the `role_permissions` table and deletes any stale rows that no longer exist in code.
- The syncer runs as an idempotent, force-reset operation (delete-then-insert per role), not a merge.
- `ListRolePermissions` continues to serve system role permissions from the in-memory map (no behavioral change at runtime), keeping query performance unchanged.
- The `role_permissions` table gains a DB snapshot that management UIs can query uniformly for both system and custom roles.
- The existing spec requirement **"System role permissions not stored"** is updated to reflect the new snapshot semantics.
- The existing spec requirement **"Role Management Data Model"** is corrected: system roles use `org_name = '__SYSTEM__'`, not `org_name = NULL` as previously stated.

## Impact

- Affected specs: `permission-management`
- Affected code:
  - New: `internal/app/permission/system_role_syncer.go` — startup syncer service
  - Modified: application bootstrap / wire-up to call syncer after migrations
  - No changes to GraphQL schema, resolvers, or runtime permission-check path
