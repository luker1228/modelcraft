## Context

ModelCraft's current domain structure evolved organically, resulting in duplicate entities and unclear boundaries:

1. **Duplicate Role Definitions**: `internal/domain/role/role.go` and `internal/domain/permission/role.go` define separate `Role` structs with incompatible field types (string UUID vs int ID, `[]string` permissions vs proper value objects).

2. **Overlapping Membership Models**: `Membership` and `UserRole` both represent user-organization-role bindings but use inconsistent identifiers (OrgID vs OrgName, string RoleID vs int RoleID), causing confusion and potential data inconsistency.

3. **Scattered Responsibilities**: Permission-related concepts (roles, permissions, user-role bindings, membership status) are split across four domains (`role`, `permission`, `membership`, `user`), violating Domain-Driven Design's high cohesion principle.

This refactoring consolidates related concepts following DDD best practices and the existing `implement-casbin-permissions` change.

## Goals / Non-Goals

### Goals

- Eliminate duplicate `Role` entity definitions
- Consolidate `Membership` and `UserRole` into a single entity
- Establish clear domain boundaries following DDD principles:
  - **Identity Domain** (`user`): User identity information
  - **Auth Domain** (`auth`): Authentication mechanisms (JWT, providers)
  - **Permission Domain** (`permission`): Authorization, roles, permissions, and membership
  - **Organization Domain** (`organization`): Tenant containers
- Standardize on `OrgName` (string) for organization references
- Maintain backward compatibility during migration

### Non-Goals

- Changing the permission checking logic or Casbin integration
- Modifying the GraphQL API surface
- Changing database schema structure (only additive changes)
- Introducing new external dependencies

## Decisions

### Decision 1: Use `permission.Role` as Single Source of Truth

**Rationale**: The `permission.Role` entity from the `implement-casbin-permissions` change is more mature, uses proper integer IDs for database relations, and is already integrated with Casbin. The old `role.Role` uses string UUIDs and `[]string` for permissions, which violates DDD principles.

**Alternatives Considered**:
- Keep both and create a mapping layer: Adds complexity and maintenance burden
- Use old `role.Role`: Inferior design, incompatible with Casbin integration

### Decision 2: Merge Membership into UserRole

**Rationale**: Both entities represent the same business concept (user membership in an organization with a specific role). Keeping them separate creates:
- Data duplication and synchronization issues
- Confusion about which entity to use
- Inconsistent identifier types (OrgID vs OrgName)

The `permission` domain is the natural home since membership status directly relates to authorization (invited users cannot access resources until they accept and become active).

**Alternatives Considered**:
- Keep separate domains: Creates artificial boundaries and duplication
- Create a third "membership" domain: Adds unnecessary complexity
- Merge into `organization` domain: Violates single responsibility (organization management vs access control)

### Decision 3: Standardize on OrgName (String)

**Rationale**:
- The existing `permission.UserRole` and Casbin integration use `OrgName`
- Organization names are unique identifiers and more URL-friendly
- Avoids UUID lookups in hot paths (permission checks)
- Aligns with RESTful API design (e.g., `/orgs/{orgName}/...`)

**Trade-off**: Renaming organizations requires cascading updates to `user_roles` table (acceptable since renames are rare).

### Decision 4: Extend UserRole with Membership Fields

**Implementation**:
```go
// Extended UserRole entity
type UserRole struct {
    ID        int            `json:"id"`
    UserID    string         `json:"user_id"`
    RoleID    int            `json:"role_id"`
    OrgName   string         `json:"org_name"`

    // Membership management fields (NEW)
    Status    UserRoleStatus `json:"status"`     // active/suspended/invited
    InvitedBy string         `json:"invited_by"` // Inviter's user ID
    InvitedAt *time.Time     `json:"invited_at"`
    JoinedAt  *time.Time     `json:"joined_at"`

    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}
```

**Rationale**: Additive changes maintain backward compatibility. Existing records default to `Status: "active"` with `JoinedAt: CreatedAt`.

## Risks / Trade-offs

### Risk 1: Breaking Existing Code

**Impact**: All code referencing `internal/domain/role` or `internal/domain/membership` will break.

**Mitigation**:
- Comprehensive grep audit: `rg "domain/role|domain/membership" internal/`
- Update all imports and type references before removing old packages
- Run full test suite after each migration phase
- Keep deprecated packages temporarily with deprecation comments

### Risk 2: Database Migration Complexity

**Impact**: Adding membership columns to `user_roles` table requires data migration if membership records exist.

**Mitigation**:
- Schema migration is additive (new columns with defaults)
- Existing `user_roles` records auto-populate: `status='active', joined_at=created_at`
- Separate migration script for any orphaned `membership` table data
- Test migration on staging environment first

### Risk 3: Coordination with Casbin Implementation

**Impact**: This refactoring depends on `implement-casbin-permissions` change being stable.

**Mitigation**:
- Wait for Casbin implementation to be deployed and tested
- Review Casbin change spec to ensure compatibility
- Coordinate with implementer to avoid conflicts

### Trade-off: Loss of Explicit Membership Domain

**Benefit**: Simpler architecture, fewer files, clearer responsibilities

**Cost**: "Membership" concept is now implicit in the `permission` domain. Documentation must clearly explain that `UserRole` handles both authorization AND membership lifecycle.

**Mitigation**: Add comprehensive documentation in `permission` domain package comments and OpenSpec.

## Migration Plan

### Phase 1: Extend Entities (Days 1-2)

1. Add membership fields to `permission.UserRole`
2. Add membership methods: `AcceptInvitation()`, `Suspend()`, `Activate()`, `ChangeRole()`
3. Extend `UserRoleRepository` interface with membership operations
4. Write unit tests for new functionality
5. **Checkpoint**: All tests pass, no breaking changes yet

### Phase 2: Update Application Services (Days 3-4)

1. Update `internal/app/organization/organization_service.go`:
   - Replace `membership.Membership` with `permission.UserRole`
   - Update `OrgMember` struct to use new entity
2. Update `internal/app/role/role_service.go`:
   - Replace `role.Role` with `permission.Role`
   - Update all method signatures
3. Audit and update all other services referencing old entities
4. **Checkpoint**: Application compiles, integration tests pass

### Phase 3: Database Schema Migration (Day 5)

1. Create migration file: `db/schema/mysql/05_extend_user_roles.sql`
2. Add columns to `user_roles`:
   ```sql
   ALTER TABLE user_roles
     ADD COLUMN status VARCHAR(20) DEFAULT 'active',
     ADD COLUMN invited_by VARCHAR(36),
     ADD COLUMN invited_at TIMESTAMP NULL,
     ADD COLUMN joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;
   ```
3. Run migration on test database
4. Verify data integrity
5. **Checkpoint**: Schema updated, existing data preserved

### Phase 4: Remove Deprecated Domains (Day 6)

1. Delete `internal/domain/role/` directory
2. Delete `internal/domain/membership/` directory
3. Remove infrastructure implementations:
   - `internal/infrastructure/repository/sql_role_repository.go` (old version)
   - `internal/infrastructure/repository/sql_membership_repository.go`
4. Run full test suite
5. **Checkpoint**: All tests pass, code compiles without deprecated packages

### Phase 5: Documentation and Cleanup (Day 7)

1. Update `docs/01-common/domain-models.md` with new structure
2. Update `CLAUDE.md` architecture section
3. Update inline documentation in domain packages
4. Create migration guide for external contributors
5. **Checkpoint**: Documentation complete, ready for review

### Rollback Plan

If issues arise, rollback by:
1. Revert code changes (Git)
2. Rollback database migration:
   ```sql
   ALTER TABLE user_roles DROP COLUMN status, DROP COLUMN invited_by, DROP COLUMN invited_at, DROP COLUMN joined_at, DROP COLUMN updated_at;
   ```
3. Restore deprecated domain packages from Git history

## Open Questions

None. The scope is well-defined and aligns with existing Casbin implementation.
