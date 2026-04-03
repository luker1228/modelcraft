## 1. Application Layer — Syncer Service

- [x] 1.1 Write unit test for `SystemRolePermissionsSyncer.Sync()`: verify it calls delete-then-insert for each system role, and that the method is idempotent (running twice does not error)
- [x] 1.2 Implement `SystemRolePermissionsSyncer` in `internal/app/permission/system_role_syncer.go`
  - Depends on `PermissionRepository` (already exists)
  - Depends on `RoleRepository.GetSystemRoleByName` (already exists in `internal/domain/role/repository.go`)
  - For each system role name in `auth.SystemRolePermissions`:
    1. Fetch the system role record from DB by name
    2. Delete all existing `role_permissions` rows for that `role_id`
    3. Bulk-insert current permissions from `auth.SystemRolePermissions`
  - Log a summary line at INFO level: `"System role permissions synced: role=%s, count=%d"`
  - Return first error encountered; caller decides whether to abort startup

## 2. Bootstrap Integration

- [x] 2.1 Wire `SystemRolePermissionsSyncer` into the application startup sequence (after DB migration, before serving requests)
- [ ] 2.2 Write integration test (or startup smoke test) that confirms `role_permissions` rows exist for all four system roles after boot

## 3. Spec Update

- [ ] 3.1 Update `permission-management` spec delta to reflect new snapshot semantics (MODIFIED requirement)

## 4. Validation

- [x] 4.1 Run `task test-unit` — all existing permission tests pass
- [ ] 4.2 Run `task auto-test` — integration tests pass
- [ ] 4.3 Manually query `SELECT * FROM role_permissions WHERE role_id IN (SELECT id FROM roles WHERE is_system=1)` after startup and confirm rows exist
- [ ] 4.4 Manually edit a system role permission row in DB, restart the server, re-query — confirm the row is reset to the code definition
