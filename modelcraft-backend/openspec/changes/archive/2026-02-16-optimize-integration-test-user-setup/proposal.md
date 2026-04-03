# Proposal: Optimize Integration Test User Setup

## Change ID
`optimize-integration-test-user-setup`

## Problem Statement

Currently, integration tests rely on manually pre-created test users with owner role assignments in the database. The existing approach has several issues:

1. **Manual Database Setup Required**: Test users must be created and assigned owner roles manually using SQL scripts (`07_test_user_permissions.sql`, `08_fix_test_user_owner_role.sql`)
2. **Race Condition Risk**: User creation happens dynamically on first login (from Casdoor JWT), but role assignment scripts assume the user already exists
3. **Environment Fragility**: Tests fail if the user doesn't exist or lacks proper roles, requiring manual intervention
4. **Poor Developer Experience**: New developers must run multiple SQL scripts and understand the timing dependencies before tests can run

## Proposed Solution

Introduce an **automated test user provisioning system** with three integration points:

1. **Automated pytest fixture** for integration tests
2. **Taskfile command** for manual/CI setup
3. **Claude Code skill** for easy access

### Key Changes

1. **New pytest fixture** (`test_user_with_owner_role`) in `tests/conftest.py`:
   - Creates user record in `users` table if not exists
   - Assigns owner role via `user_organizations` table
   - Returns user ID for test use
   - Cleanup after test session

2. **Database helper module** (`tests/common/test_user_setup.py`):
   - Reusable functions for user and role management
   - Idempotent insert operations with proper error handling
   - CLI support for standalone execution

3. **Taskfile commands** for manual setup:
   - `task test-user-setup`: Create test-integration user with owner role
   - `task test-user-cleanup`: Remove test-integration user
   - Both commands available in root and tests/ Taskfile

4. **Claude Code skill** (`.claudecode/skills/test-user-setup.md`):
   - Quick access via `/test-user-setup` command
   - Executes `task test-user-setup` with proper error handling
   - Provides clear feedback and troubleshooting guidance

5. **Update integration tests** to use the new fixture:
   - Replace manual user assumptions with fixture dependency
   - Ensure authentication uses the provisioned test user

6. **Document the change** in test documentation

## Benefits

- ✅ **Zero Manual Setup**: Tests work out-of-the-box after environment deployment
- ✅ **Deterministic**: Eliminates race conditions between user creation and role assignment
- ✅ **Better Isolation**: Each test session has a clean, known user state
- ✅ **Improved CI/CD**: Automated tests run reliably without pre-deployment scripts
- ✅ **Developer Friendly**: New developers can run tests immediately after environment setup

## Non-Goals

- Replacing Casdoor authentication mechanism
- Changing existing user permission schema
- Modifying runtime authentication/authorization logic
- Supporting multiple test users (single owner user is sufficient)

## Affected Components

- **Test Infrastructure**: `tests/conftest.py`, `tests/design/conftest.py`, `tests/runtime/conftest.py`
- **Test Utilities**: Enhanced `tests/common/test_user_setup.py` module (already exists)
- **Integration Tests**: `tests/runtime/integration/test_modelcraft_client.py` and related
- **Build Automation**: `Taskfile.yml` (root and tests/)
- **Claude Code Skills**: New `.claudecode/skills/test-user-setup.md`
- **Documentation**: Test workflow and setup guides

## Migration Path

### Phase 1: Enhance Existing Test User Setup (Non-Breaking)
- Enhance existing `tests/common/test_user_setup.py` for standalone execution
- Add Taskfile commands (`test-user-setup`, `test-user-cleanup`)
- Add Claude Code skill for easy access
- Existing tests continue working with manual setup

### Phase 2: Add Pytest Fixture Integration
- Add `test_user_with_owner_role` fixture to root `conftest.py`
- Fixture uses existing `test_user_setup.py` functions
- Tests can opt-in to use fixture

### Phase 3: Update Integration Tests
- Modify integration tests to use the new fixture
- Verify tests pass with both manual and automatic setup

### Phase 4: Deprecate Manual Scripts (Optional)
- Mark `07_test_user_permissions.sql` and `08_fix_test_user_owner_role.sql` as deprecated
- Update documentation to reflect the new approach

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Database connection failures during fixture setup | Add proper error handling with clear messages; fixture fails fast before tests run |
| User ID conflicts with existing data | Use deterministic UUID based on test config; check existence before creation |
| Cleanup failures leaving test data | Use try-finally blocks; provide manual cleanup command as fallback |
| Performance impact of per-session DB operations | Session-scoped fixture runs once; minimal overhead (<100ms) |

## Success Criteria

1. ✅ Taskfile command `task test-user-setup` creates test-integration user with owner role
2. ✅ Claude Code skill `/test-user-setup` provides easy access to user setup
3. ✅ Integration tests run successfully without manual SQL scripts (via pytest fixture)
4. ✅ New developers can run `task auto-test` immediately after `task deploy-local`
5. ✅ Test user is created with owner role before tests execute
6. ✅ Test user is cleaned up after test session completes
7. ✅ No test failures related to missing users or insufficient permissions
8. ✅ CI/CD pipeline runs tests without pre-deployment steps
9. ✅ Manual setup option remains available via `task test-user-setup`

## Open Questions

1. **User ID Strategy**: Should we use a fixed UUID for the test user or generate dynamically?
   - **Recommendation**: Use fixed UUID based on `CASDOOR_TEST_USERNAME` to ensure idempotency

2. **Cleanup Scope**: Should we clean up the test user after each test session?
   - **Recommendation**: Yes, but make it configurable via environment variable for debugging

3. **Organization Assignment**: Which organization should the test user belong to?
   - **Recommendation**: Use existing "modelcraft" organization (already in schema)

4. **Multiple Environments**: How to handle different test environments (local, docker)?
   - **Recommendation**: Use same fixture with environment-specific database config from `.env` files

## Related Work

- Existing auth fixture: `tests/conftest.py::auth_token` (obtains JWT from Casdoor)
- Existing DB config: `tests/common/config.py::TestConfig.get_db_config()`
- Manual user setup scripts: `db/schema/mysql/07_test_user_permissions.sql`, `08_fix_test_user_owner_role.sql`
- Project setup fixture: `tests/design/conftest.py::default_project`

## Timeline Estimate

- **Complexity**: Low-Medium (pure test infrastructure, no production code changes)
- **Effort**: ~4-6 hours
  - Design and implement db_helper.py: 1-2h
  - Add pytest fixture: 1h
  - Update integration tests: 1-2h
  - Testing and documentation: 1h

## Approval Required From

- Test infrastructure maintainer
- Integration test owners
- DevOps (for CI/CD impact validation)
