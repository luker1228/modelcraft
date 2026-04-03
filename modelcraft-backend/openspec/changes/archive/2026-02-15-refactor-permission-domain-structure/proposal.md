# Change: Refactor Permission Domain Structure to Eliminate Redundancy

## Why

The current domain structure has significant redundancy and unclear boundaries:

1. **Duplicate Role Entities**: Two separate `Role` structs exist in `internal/domain/role/` and `internal/domain/permission/` with different field types and purposes, causing confusion about which to use.

2. **Overlapping Membership Concepts**: Both `Membership` (in `internal/domain/membership/`) and `UserRole` (in `internal/domain/permission/`) represent user-organization-role bindings, with inconsistent identifiers:
   - `Membership` uses `OrgID` (UUID) and `RoleID` (string UUID)
   - `UserRole` uses `OrgName` (string) and `RoleID` (int)

3. **Unclear Domain Boundaries**: The separation of `role`, `permission`, `membership`, and `user` domains creates artificial boundaries that don't align with Domain-Driven Design principles of high cohesion.

4. **Type Safety Issues**: The old `role.Role` uses `[]string` for permissions instead of the proper `Permission` value object, violating DDD principles.

This refactoring consolidates related concepts into cohesive domains following DDD best practices and eliminates data redundancy.

## What Changes

### Domain Consolidation

1. **Delete `internal/domain/role/`** - Consolidate into `permission` domain
2. **Delete `internal/domain/membership/`** - Merge functionality into `permission.UserRole`
3. **Keep `internal/domain/permission/`** as the unified permission and membership management domain
4. **Keep `internal/domain/user/`** as the identity domain
5. **Keep `internal/domain/auth/`** as the authentication domain
6. **Keep `internal/domain/organization/`** as the tenant domain

### Entity Changes

1. **Extend `permission.UserRole`** to include membership status fields:
   - Add `Status` field (active/suspended/invited)
   - Add `InvitedBy`, `InvitedAt`, `JoinedAt` fields
   - Add methods: `AcceptInvitation()`, `ChangeRole()`, `Suspend()`, `Activate()`

2. **Use `permission.Role` as the single source of truth**:
   - Remove `role.Role` struct entirely
   - Update all references to use `permission.Role`

3. **Standardize on `OrgName` (string)** throughout:
   - Remove `OrgID` references in favor of `OrgName`
   - Align with existing `permission` domain conventions

### Repository Changes

1. **Extend `UserRoleRepository`** with membership operations:
   - Add `CreateInvitation()`, `AcceptInvitation()`, `UpdateMembershipStatus()`
   - Add `ListOrgMembers()`, `ListUserMemberships()`, `DeleteMembership()`

2. **Remove `MembershipRepository`** interface

### Application Layer Changes

1. **Update services** to use consolidated entities:
   - `internal/app/organization/` - Use `permission.UserRole` instead of `membership.Membership`
   - `internal/app/permission/` - Already using correct entities
   - Remove references to `role.Role`

### Breaking Changes

**BREAKING**: This is a structural refactoring that changes domain boundaries. Migration strategy:
1. Deploy extended schema with backward-compatible changes
2. Migrate data from `membership` references to `user_roles` table
3. Update all code references
4. Remove deprecated domain packages

## Impact

### Affected Specs

- **MODIFIED**: `permission-management` - Extend UserRole entity with membership capabilities
- **NEW**: `domain-structure` - Document the consolidated domain architecture

### Affected Code

- **DELETE**: `internal/domain/role/` (2 files)
- **DELETE**: `internal/domain/membership/` (3 files)
- **MODIFY**: `internal/domain/permission/user_role.go` - Add membership fields and methods
- **MODIFY**: `internal/domain/permission/repository.go` - Extend UserRoleRepository interface
- **MODIFY**: `internal/app/organization/organization_service.go` - Use `permission.UserRole`
- **MODIFY**: `internal/app/role/role_service.go` - Update to use `permission.Role`
- **MODIFY**: All imports referencing `internal/domain/role` or `internal/domain/membership`

### Database Changes

**No schema changes required** - This is purely a code refactoring. The existing `user_roles` table already has an `id` primary key that can accommodate the additional fields via schema migration if needed.

### Migration Strategy

1. **Phase 1**: Extend `permission.UserRole` with membership fields (backward compatible)
2. **Phase 2**: Update application services to use consolidated entities
3. **Phase 3**: Run database migration to add membership columns to `user_roles` table
4. **Phase 4**: Migrate existing membership data (if any)
5. **Phase 5**: Remove deprecated domain packages
6. **Phase 6**: Update documentation and tests

## Dependencies

- Depends on `implement-casbin-permissions` change being completed
- No new external dependencies required

## Risks

- **Medium Risk**: Code references scattered across multiple files need careful refactoring
- **Mitigation**: Comprehensive test coverage, gradual migration, rollback plan
- **Low Risk**: Database schema changes are additive and backward compatible
