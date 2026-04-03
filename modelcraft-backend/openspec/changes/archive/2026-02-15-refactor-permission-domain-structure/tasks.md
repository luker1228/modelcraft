## 1. Preparation

- [ ] 1.1 Review existing code references to `internal/domain/role` and `internal/domain/membership`
- [ ] 1.2 Audit all imports with: `rg "domain/(role|membership)" internal/ -l`
- [ ] 1.3 Identify all affected files and services
- [ ] 1.4 Review `implement-casbin-permissions` spec for compatibility
- [ ] 1.5 Create feature branch: `refactor/permission-domain-structure`

## 2. Extend Permission Domain Entities

### 2.1 Add UserRoleStatus Enum
- [ ] 2.1.1 Define `UserRoleStatus` type in `internal/domain/permission/user_role.go`
- [ ] 2.1.2 Add constants: `UserRoleStatusActive`, `UserRoleStatusSuspended`, `UserRoleStatusInvited`
- [ ] 2.1.3 Write unit test: `TestUserRoleStatus_Values` to verify enum values

### 2.2 Extend UserRole Struct
- [ ] 2.2.1 Add `Status UserRoleStatus` field to UserRole struct
- [ ] 2.2.2 Add `InvitedBy string` field
- [ ] 2.2.3 Add `InvitedAt *time.Time` field
- [ ] 2.2.4 Add `JoinedAt *time.Time` field
- [ ] 2.2.5 Add `UpdatedAt time.Time` field
- [ ] 2.2.6 Write unit test: `TestUserRole_StructFields` to verify new fields exist

### 2.3 Add NewInvitation Constructor
- [ ] 2.3.1 Implement `NewInvitation(userID, roleID, orgName, invitedBy)` constructor
- [ ] 2.3.2 Write unit test: `TestNewInvitation_Success` - verify status=invited, invited_by set
- [ ] 2.3.3 Write unit test: `TestNewInvitation_ValidatesInput` - verify empty userID rejected

### 2.4 Add AcceptInvitation Method
- [ ] 2.4.1 Implement `AcceptInvitation() error` method
- [ ] 2.4.2 Write unit test: `TestAcceptInvitation_Success` - verify status changes to active
- [ ] 2.4.3 Write unit test: `TestAcceptInvitation_OnlyInvitedStatus` - verify error on active status
- [ ] 2.4.4 Write unit test: `TestAcceptInvitation_SetsJoinedAt` - verify joined_at timestamp

### 2.5 Add ChangeRole Method
- [ ] 2.5.1 Implement `ChangeRole(newRoleID int) error` method
- [ ] 2.5.2 Write unit test: `TestChangeRole_Success` - verify role_id updated
- [ ] 2.5.3 Write unit test: `TestChangeRole_InvalidRoleID` - verify error on zero/negative ID
- [ ] 2.5.4 Write unit test: `TestChangeRole_UpdatesTimestamp` - verify updated_at changed

### 2.6 Add Suspend and Activate Methods
- [ ] 2.6.1 Implement `Suspend()` method
- [ ] 2.6.2 Implement `Activate()` method
- [ ] 2.6.3 Write unit test: `TestSuspend_ChangesStatus` - verify status=suspended
- [ ] 2.6.4 Write unit test: `TestActivate_ChangesStatus` - verify status=active
- [ ] 2.6.5 Write unit test: `TestSuspendActivate_UpdatesTimestamp` - verify updated_at

### 2.7 Add IsActive Method
- [ ] 2.7.1 Implement `IsActive() bool` method
- [ ] 2.7.2 Write unit test: `TestIsActive_ReturnsTrue` - when status=active
- [ ] 2.7.3 Write unit test: `TestIsActive_ReturnsFalse` - when status=suspended/invited

### 2.8 Update Validate Method
- [ ] 2.8.1 Update `Validate()` to check status field is valid
- [ ] 2.8.2 Write unit test: `TestValidate_InvalidStatus` - verify error on unknown status
- [ ] 2.8.3 Write unit test: `TestValidate_Success` - verify validation passes with new fields

## 3. Extend Permission Repository Interfaces

### 3.1 Add CreateInvitation Method
- [ ] 3.1.1 Add `CreateInvitation(ctx, invitation *UserRole) error` to UserRoleRepository interface
- [ ] 3.1.2 Implement in `sql_user_role_repository.go`
- [ ] 3.1.3 Write unit test: `TestCreateInvitation_Success` - mock saves invitation to DB
- [ ] 3.1.4 Write unit test: `TestCreateInvitation_DatabaseError` - verify error handling

### 3.2 Add GetInvitation Method
- [ ] 3.2.1 Add `GetInvitation(ctx, userID, orgName string) (*UserRole, error)` to interface
- [ ] 3.2.2 Implement query: WHERE user_id=? AND org_name=? AND status='invited'
- [ ] 3.2.3 Write unit test: `TestGetInvitation_Found` - mock returns invitation
- [ ] 3.2.4 Write unit test: `TestGetInvitation_NotFound` - verify returns nil, no error

### 3.3 Add AcceptInvitation Method
- [ ] 3.3.1 Add `AcceptInvitation(ctx, userID, orgName string) error` to interface
- [ ] 3.3.2 Implement update: SET status='active', joined_at=NOW()
- [ ] 3.3.3 Write unit test: `TestAcceptInvitation_Success` - mock updates record
- [ ] 3.3.4 Write unit test: `TestAcceptInvitation_NotFound` - verify error when no invitation

### 3.4 Add UpdateMembershipStatus Method
- [ ] 3.4.1 Add `UpdateMembershipStatus(ctx, userID, orgName, status) error` to interface
- [ ] 3.4.2 Implement update: SET status=?, updated_at=NOW()
- [ ] 3.4.3 Write unit test: `TestUpdateMembershipStatus_Success` - mock updates status
- [ ] 3.4.4 Write unit test: `TestUpdateMembershipStatus_InvalidStatus` - verify validation

### 3.5 Add GetMembershipByID Method
- [ ] 3.5.1 Add `GetMembershipByID(ctx, id int) (*UserRole, error)` to interface
- [ ] 3.5.2 Implement query: WHERE id=?
- [ ] 3.5.3 Write unit test: `TestGetMembershipByID_Found` - mock returns record
- [ ] 3.5.4 Write unit test: `TestGetMembershipByID_NotFound` - verify returns nil

### 3.6 Add ListOrgMembers Method
- [ ] 3.6.1 Add `ListOrgMembers(ctx, orgName string) ([]*UserRole, error)` to interface
- [ ] 3.6.2 Implement query: WHERE org_name=? ORDER BY created_at
- [ ] 3.6.3 Write unit test: `TestListOrgMembers_Success` - mock returns list
- [ ] 3.6.4 Write unit test: `TestListOrgMembers_Empty` - verify returns empty slice

### 3.7 Add ListUserMemberships Method
- [ ] 3.7.1 Add `ListUserMemberships(ctx, userID string) ([]*UserRole, error)` to interface
- [ ] 3.7.2 Implement query: WHERE user_id=? ORDER BY org_name
- [ ] 3.7.3 Write unit test: `TestListUserMemberships_Success` - mock returns list
- [ ] 3.7.4 Write unit test: `TestListUserMemberships_MultipleOrgs` - verify cross-org query

### 3.8 Add DeleteMembership Method
- [ ] 3.8.1 Add `DeleteMembership(ctx, id int) error` to interface
- [ ] 3.8.2 Implement delete: WHERE id=?
- [ ] 3.8.3 Write unit test: `TestDeleteMembership_Success` - mock deletes record
- [ ] 3.8.4 Write unit test: `TestDeleteMembership_NotFound` - verify no error on missing record

### 3.9 Add DeleteUserRolesByOrg Method
- [ ] 3.9.1 Add `DeleteUserRolesByOrg(ctx, orgName string) error` to interface
- [ ] 3.9.2 Implement delete: WHERE org_name=?
- [ ] 3.9.3 Write unit test: `TestDeleteUserRolesByOrg_Success` - mock deletes all org records
- [ ] 3.9.4 Write unit test: `TestDeleteUserRolesByOrg_EmptyOrg` - verify handles no records

## 4. Update Application Services

### 4.1 Update OrganizationService
- [ ] 4.1.1 Replace membership import with permission import
- [ ] 4.1.2 Change `membershipRepo` field to `userRoleRepo`
- [ ] 4.1.3 Update `OrgMember` struct to use `UserRole`
- [ ] 4.1.4 Update `ListMembers()` method implementation
- [ ] 4.1.5 Write unit test: `TestListMembers_UsesUserRoleRepo` - verify new repo called
- [ ] 4.1.6 Write unit test: `TestListMembers_ReturnsUserRoles` - verify return type

### 4.2 Update RoleService
- [ ] 4.2.1 Replace role domain import with permission import
- [ ] 4.2.2 Update all `role.Role` references to `permission.Role`
- [ ] 4.2.3 Update method signatures
- [ ] 4.2.4 Write unit test: `TestRoleService_UsesPermissionRole` - verify correct type
- [ ] 4.2.5 Write unit test: `TestRoleService_MethodSignatures` - verify updated signatures

### 4.3 Audit and Update Other Services
- [ ] 4.3.1 Search for all imports: `rg "domain/role|domain/membership" internal/app/ -l`
- [ ] 4.3.2 Update each file systematically
- [ ] 4.3.3 Run unit tests for each updated service
- [ ] 4.3.4 Write integration test: `TestServices_NoDeprecatedImports` - verify no old imports

## 5. Update Infrastructure Layer

### 5.1 Implement New Repository Methods
- [ ] 5.1.1 Implement `CreateInvitation` in sql_user_role_repository.go
- [ ] 5.1.2 Write integration test: `TestGormUserRoleRepo_CreateInvitation` - real DB test
- [ ] 5.1.3 Implement `AcceptInvitation` with transaction support
- [ ] 5.1.4 Write integration test: `TestGormUserRoleRepo_AcceptInvitation` - verify DB update
- [ ] 5.1.5 Implement `UpdateMembershipStatus`
- [ ] 5.1.6 Write integration test: `TestGormUserRoleRepo_UpdateStatus` - verify DB update
- [ ] 5.1.7 Implement `ListOrgMembers` with proper indexing
- [ ] 5.1.8 Write integration test: `TestGormUserRoleRepo_ListOrgMembers` - verify query
- [ ] 5.1.9 Implement `DeleteUserRolesByOrg` for cascade deletion
- [ ] 5.1.10 Write integration test: `TestGormUserRoleRepo_DeleteByOrg` - verify cascade

### 5.2 Remove Deprecated Repositories
- [ ] 5.2.1 Delete `sql_membership_repository.go` if exists
- [ ] 5.2.2 Delete old `sql_role_repository.go` if separate
- [ ] 5.2.3 Write test: `TestDeprecatedRepositories_DoNotExist` - verify files deleted

### 5.3 Update Repository Factory
- [ ] 5.3.1 Remove deprecated repository initialization
- [ ] 5.3.2 Update factory to only create permission repositories
- [ ] 5.3.3 Write unit test: `TestRepositoryFactory_NoDeprecatedRepos` - verify factory

## 6. Database Schema Migration

### 6.1 Create Migration File
- [ ] 6.1.1 Create `db/schema/mysql/05_extend_user_roles.sql`
- [ ] 6.1.2 Add ALTER TABLE statements for new columns
- [ ] 6.1.3 Add UPDATE statement for existing records
- [ ] 6.1.4 Add CREATE INDEX statements
- [ ] 6.1.5 Write migration test: `test_migration_05_extend_user_roles.py`

### 6.2 Test Migration
- [ ] 6.2.1 Run migration on clean test database
- [ ] 6.2.2 Verify all columns added: `SHOW COLUMNS FROM user_roles`
- [ ] 6.2.3 Verify indexes created: `SHOW INDEX FROM user_roles`
- [ ] 6.2.4 Write test: `test_existing_records_get_defaults` - verify backfill

### 6.3 Test Data Integrity
- [ ] 6.3.1 Insert test records before migration
- [ ] 6.3.2 Run migration
- [ ] 6.3.3 Verify existing data preserved
- [ ] 6.3.4 Write test: `test_migration_preserves_data`

### 6.4 Update Documentation
- [ ] 6.4.1 Update `db/README.md` with new migration
- [ ] 6.4.2 Document rollback procedure
- [ ] 6.4.3 Add schema diagram showing new fields

## 7. Remove Deprecated Domains

### 7.1 Delete Role Domain
- [ ] 7.1.1 Run: `rm -rf internal/domain/role/`
- [ ] 7.1.2 Verify no imports remain: `rg "domain/role" internal/`
- [ ] 7.1.3 Write test: `TestRoleDomain_DoesNotExist` - verify directory deleted

### 7.2 Delete Membership Domain
- [ ] 7.2.1 Run: `rm -rf internal/domain/membership/`
- [ ] 7.2.2 Verify no imports remain: `rg "domain/membership" internal/`
- [ ] 7.2.3 Write test: `TestMembershipDomain_DoesNotExist` - verify directory deleted

### 7.3 Verify Clean State
- [ ] 7.3.1 Run: `task build` - ensure compilation succeeds
- [ ] 7.3.2 Run: `task test-unit` - ensure all tests pass
- [ ] 7.3.3 Run: `rg "domain/(role|membership)" internal/` - verify zero results

## 8. Update GraphQL Layer

### 8.1 Review Resolvers
- [ ] 8.1.1 Audit GraphQL resolvers: `rg "membership\\.Membership|role\\.Role" internal/interfaces/graphql/`
- [ ] 8.1.2 Update resolvers to use `permission.UserRole`
- [ ] 8.1.3 Write unit test: `TestResolvers_UsePermissionTypes` - verify correct types

### 8.2 Update Schema Types
- [ ] 8.2.1 Check if membership status exposed in GraphQL schema
- [ ] 8.2.2 Add new fields to schema if needed
- [ ] 8.2.3 Run: `task generate-gql` to regenerate code
- [ ] 8.2.4 Write integration test: `test_graphql_membership_fields.py` - verify schema

### 8.3 Test GraphQL Operations
- [ ] 8.3.1 Test query: listOrgMembers with new status field
- [ ] 8.3.2 Test mutation: acceptInvitation
- [ ] 8.3.3 Write integration test: `test_graphql_membership_operations.py`

## 9. Comprehensive Testing

### 9.1 Unit Tests
- [ ] 9.1.1 Run: `task test-unit` and verify all pass
- [ ] 9.1.2 Check coverage: `go test -cover ./internal/domain/permission/`
- [ ] 9.1.3 Ensure >80% coverage for modified code
- [ ] 9.1.4 Fix any failing tests

### 9.2 Domain Entity Tests
- [ ] 9.2.1 Test: `TestNewInvitation_AllScenarios` - multiple test cases
- [ ] 9.2.2 Test: `TestAcceptInvitation_AllScenarios`
- [ ] 9.2.3 Test: `TestSuspendActivate_Transitions`
- [ ] 9.2.4 Test: `TestChangeRole_Validation`

### 9.3 Repository Tests
- [ ] 9.3.1 Test: `TestUserRoleRepository_InvitationFlow` - create, get, accept
- [ ] 9.3.2 Test: `TestUserRoleRepository_StatusManagement`
- [ ] 9.3.3 Test: `TestUserRoleRepository_OrgOperations`
- [ ] 9.3.4 Test: `TestUserRoleRepository_CascadeDelete`

### 9.4 Integration Tests
- [ ] 9.4.1 Run: `task auto-test` for full integration suite
- [ ] 9.4.2 Test: `test_invitation_lifecycle.py` - end-to-end flow
- [ ] 9.4.3 Test: `test_membership_status_transitions.py`
- [ ] 9.4.4 Test: `test_multi_tenant_isolation.py` - verify org separation

### 9.5 Manual Testing
- [ ] 9.5.1 Manual test: Create invitation via GraphQL
- [ ] 9.5.2 Manual test: Accept invitation and verify status change
- [ ] 9.5.3 Manual test: Suspend member and verify access denied
- [ ] 9.5.4 Manual test: Change member role and verify permissions
- [ ] 9.5.5 Manual test: List org members and verify all fields present

## 10. Documentation

### 10.1 Update CLAUDE.md
- [ ] 10.1.1 Update "Core Domain Concepts" section
- [ ] 10.1.2 Update "Directory Structure" section
- [ ] 10.1.3 Add section: "Permission Domain - Unified Authorization and Membership"
- [ ] 10.1.4 Document UserRole dual purpose
- [ ] 10.1.5 Add code examples for common operations

### 10.2 Update Domain Documentation
- [ ] 10.2.1 Create/update `docs/01-common/domain-models.md`
- [ ] 10.2.2 Remove Membership entity documentation
- [ ] 10.2.3 Update UserRole documentation with new fields
- [ ] 10.2.4 Add domain architecture diagram
- [ ] 10.2.5 Document domain boundaries and responsibilities

### 10.3 Create Migration Guide
- [ ] 10.3.1 Create `docs/migrations/permission-domain-refactor.md`
- [ ] 10.3.2 Document breaking changes
- [ ] 10.3.3 Provide code migration examples (before/after)
- [ ] 10.3.4 Document database migration steps
- [ ] 10.3.5 Add troubleshooting section

### 10.4 Update Inline Documentation
- [ ] 10.4.1 Add package comment to `internal/domain/permission/` package
- [ ] 10.4.2 Update `UserRole` struct comment
- [ ] 10.4.3 Add godoc examples for new methods
- [ ] 10.4.4 Update repository interface comments

## 11. Code Review and Approval

### 11.1 Pre-Review Checklist
- [ ] 11.1.1 Run: `task fmt` to format all code
- [ ] 11.1.2 Run: `task lint` to check linting issues
- [ ] 11.1.3 Run: `task vet` to check for issues
- [ ] 11.1.4 Run: `task test-unit` - ensure all tests pass
- [ ] 11.1.5 Run: `task auto-test` - ensure integration tests pass

### 11.2 Self-Review
- [ ] 11.2.1 Review all changed files in diff
- [ ] 11.2.2 Check for commented-out code
- [ ] 11.2.3 Verify no debug print statements
- [ ] 11.2.4 Check for proper error handling
- [ ] 11.2.5 Verify test coverage adequate

### 11.3 Create Pull Request
- [ ] 11.3.1 Create PR with descriptive title
- [ ] 11.3.2 Write detailed PR description with:
  - Summary of changes
  - Migration steps
  - Testing performed
  - Screenshots (if applicable)
- [ ] 11.3.3 Link related issues/specs
- [ ] 11.3.4 Add appropriate labels

### 11.4 Code Review Process
- [ ] 11.4.1 Request review from team lead
- [ ] 11.4.2 Address all review comments
- [ ] 11.4.3 Re-run tests after changes
- [ ] 11.4.4 Get final approval

## 12. Deployment

### 12.1 Staging Deployment
- [ ] 12.1.1 Merge PR to main branch
- [ ] 12.1.2 Deploy to staging environment
- [ ] 12.1.3 Run database migration on staging
- [ ] 12.1.4 Verify staging application starts successfully
- [ ] 12.1.5 Check staging logs for errors

### 12.2 Staging Verification
- [ ] 12.2.1 Test invitation flow in staging
- [ ] 12.2.2 Test membership status changes
- [ ] 12.2.3 Verify GraphQL queries work
- [ ] 12.2.4 Check database for correct schema
- [ ] 12.2.5 Run smoke tests on staging

### 12.3 Production Deployment
- [ ] 12.3.1 Create deployment plan
- [ ] 12.3.2 Schedule maintenance window (if needed)
- [ ] 12.3.3 Backup production database
- [ ] 12.3.4 Deploy to production
- [ ] 12.3.5 Run database migration on production
- [ ] 12.3.6 Verify production health checks

### 12.4 Post-Deployment Monitoring
- [ ] 12.4.1 Monitor application logs for errors
- [ ] 12.4.2 Monitor database performance
- [ ] 12.4.3 Check error rates in monitoring dashboard
- [ ] 12.4.4 Verify key user flows working
- [ ] 12.4.5 Monitor for 24 hours post-deployment

## 13. Cleanup and Archive

### 13.1 Archive Change
- [ ] 13.1.1 Run: `openspec archive refactor-permission-domain-structure --yes`
- [ ] 13.1.2 Verify specs updated correctly
- [ ] 13.1.3 Run: `openspec validate --all --strict`

### 13.2 Project Management
- [ ] 13.2.1 Close related GitHub issues
- [ ] 13.2.2 Update project board
- [ ] 13.2.3 Update CHANGELOG.md

### 13.3 Team Communication
- [ ] 13.3.1 Announce completion to team
- [ ] 13.3.2 Share migration guide with team
- [ ] 13.3.3 Schedule knowledge sharing session (if needed)
