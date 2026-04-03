# Tasks: Optimize Integration Test User Setup

## Task Breakdown

### 1. Enhance Test User Setup Module
**Status**: Pending
**Owner**: Unassigned
**Estimated Effort**: 1-2 hours

**Description**: Enhance existing `tests/common/test_user_setup.py` to support standalone execution and improve error handling.

**Acceptance Criteria**:
- [ ] Review and enhance existing `tests/common/test_user_setup.py`
- [ ] Ensure `execute_test_user_setup(db_config)` function is idempotent
  - INSERT IGNORE for users and user_organizations
  - Returns user data dict: `{id, external_id, name, org_name, role_name, status}`
- [ ] Ensure `cleanup_test_user(db_config, user_id)` handles cascading deletions
- [ ] Add clear error messages for common failures:
  - Database connection issues
  - Missing organization or role
- [ ] Enhance `__main__` block for standalone CLI execution
- [ ] Add proper logging with emoji indicators (✅, 🧹, ❌)
- [ ] Update docstrings with usage examples

**Dependencies**: None

**Testing**:
- Manual execution: `python tests/common/test_user_setup.py`
- Test with fresh database
- Test with existing user (idempotency)
- Test cleanup function

---

### 2. Add Taskfile Commands
**Status**: Pending
**Owner**: Unassigned
**Estimated Effort**: 1 hour

**Description**: Add Taskfile commands to root and tests/ Taskfile for manual test user setup.

**Acceptance Criteria**:
- [ ] Add `test-user-setup` task to root `Taskfile.yml`:
  - Description: "Create test-integration user with owner role"
  - Executes: `python tests/common/test_user_setup.py`
  - Uses database config from `.env` file
  - Shows clear success/failure messages
- [ ] Add `test-user-cleanup` task to root `Taskfile.yml`:
  - Description: "Remove test-integration user"
  - Executes cleanup function from test_user_setup.py
  - Accepts optional USER_ID parameter
- [ ] Add corresponding tasks to `tests/Taskfile.yml`
- [ ] Update `task --list` help text with new commands
- [ ] Test commands in different environments (local, docker)

**Dependencies**: Task 1 (Enhanced Test User Setup Module)

**Testing**:
- Run `task test-user-setup` with fresh database
- Run `task test-user-setup` with existing user (should be idempotent)
- Run `task test-user-cleanup` to verify cleanup
- Check `task --list` output

---

### 3. Create Claude Code Skill
**Status**: Pending
**Owner**: Unassigned
**Estimated Effort**: 30 minutes

**Description**: Create a Claude Code skill for easy access to test user setup via `/test-user-setup` command.

**Acceptance Criteria**:
- [ ] Create `.claudecode/skills/test-user-setup.md` skill definition
- [ ] Skill executes `task test-user-setup`
- [ ] Provides clear feedback:
  - Success message with user details
  - Error message with troubleshooting steps
  - Suggests running `task deploy-local` if database not ready
- [ ] Includes description and usage examples
- [ ] Follows Claude Code skill format conventions
- [ ] Add cleanup skill: `.claudecode/skills/test-user-cleanup.md`

**Dependencies**: Task 2 (Taskfile Commands)

**Testing**:
- Test `/test-user-setup` command in Claude Code CLI
- Verify skill appears in skill list
- Test error handling (database not running, etc.)

---

### 4. Add Test User Provisioning Fixture
**Status**: Pending
**Owner**: Unassigned
**Estimated Effort**: 1 hour

**Description**: Create session-scoped pytest fixture to provision test user with owner role.

**Acceptance Criteria**:
- [ ] Add `test_user_with_owner_role` fixture to `tests/conftest.py`
- [ ] Fixture uses existing `execute_test_user_setup()` function
- [ ] Fixture is session-scoped (runs once per test session)
- [ ] Fixture returns user data dict: `{id, external_id, name, org_name, role_name}`
- [ ] Add cleanup in fixture finalizer (teardown)
- [ ] Make cleanup configurable via `KEEP_TEST_USER` env var (default: false)
- [ ] Log fixture actions with clear messages (✅ Created test user, 🧹 Cleaned up test user)
- [ ] Handle fixture failures gracefully with clear error messages

**Dependencies**: Task 1 (Enhanced Test User Setup Module)

**Testing**:
- Run fixture in isolation: `pytest tests/conftest.py --setup-show`
- Verify user created in database
- Verify user has owner role in user_organizations
- Verify cleanup removes user
- Test `KEEP_TEST_USER=true` flag

---

### 5. Update Integration Tests to Use New Fixture
**Status**: Pending
**Owner**: Unassigned
**Estimated Effort**: 1-2 hours

**Description**: Modify integration tests to depend on the new user provisioning fixture.

**Acceptance Criteria**:
- [ ] Add `test_user_with_owner_role` fixture parameter to integration test classes
- [ ] Verify `auth_token` fixture uses the same test user credentials
- [ ] Remove any assumptions about pre-existing test users
- [ ] Update `tests/runtime/integration/test_modelcraft_client.py`
- [ ] Update any other integration tests that depend on test user
- [ ] Ensure tests pass with clean database (no manual setup)
- [ ] Ensure tests pass with existing test user (idempotency)

**Dependencies**: Task 4 (Test User Provisioning Fixture)

**Testing**:
- Run integration tests with fresh database: `task deploy-local && task auto-test`
- Run integration tests with existing user: `task auto-test` (twice)
- Verify no permission errors
- Verify no "user not found" errors

---

### 6. Update Test Documentation
**Status**: Pending
**Owner**: Unassigned
**Estimated Effort**: 30 minutes

**Description**: Update test documentation to reflect the new automated user setup and manual commands.

**Acceptance Criteria**:
- [ ] Update `tests/README.md` with:
  - New fixture usage
  - Taskfile commands (`task test-user-setup`, `task test-user-cleanup`)
  - Claude Code skill usage (`/test-user-setup`)
- [ ] Update `CLAUDE.md` Test-Driven Development section
- [ ] Document `KEEP_TEST_USER` environment variable
- [ ] Add troubleshooting section for user provisioning failures
- [ ] Mark manual SQL scripts (`07_test_user_permissions.sql`, `08_fix_test_user_owner_role.sql`) as deprecated
- [ ] Add examples for different use cases:
  - Automated (pytest fixture)
  - Manual (Taskfile command)
  - Quick (Claude Code skill)

**Dependencies**: Tasks 1-5 (Implementation complete)

**Testing**:
- Review documentation for clarity
- Verify links and references are correct
- Ask a team member to follow the documentation

---

### 7. Validate CI/CD Compatibility
**Status**: Pending
**Owner**: Unassigned
**Estimated Effort**: 30 minutes

**Description**: Ensure the change works in CI/CD pipelines without manual intervention.

**Acceptance Criteria**:
- [ ] Run full test suite in Docker environment: `task full-auto-test-docker`
- [ ] Verify no manual SQL scripts needed in CI/CD setup
- [ ] Check test execution time (should be <100ms overhead for user provisioning)
- [ ] Verify cleanup happens even if tests fail
- [ ] Document any CI/CD configuration changes needed
- [ ] Verify Taskfile commands work in CI/CD environment

**Dependencies**: Tasks 1-5 (Implementation complete)

**Testing**:
- Simulate CI/CD run locally using Docker
- Check logs for fixture execution messages
- Verify test results
- Check that `task test-user-setup` works in Docker

---

## Task Sequencing

Tasks should be executed with the following dependencies:

```
1. Enhance Test User Setup Module
   ↓
   ├─→ 2. Add Taskfile Commands
   │    ↓
   │    3. Create Claude Code Skill
   │
   └─→ 4. Add Test User Provisioning Fixture
        ↓
        5. Update Integration Tests
        ↓
        6. Update Test Documentation
        7. Validate CI/CD Compatibility (can run in parallel with 6)
```

**Parallelizable work**:
- Tasks 2 & 4 can start in parallel after Task 1 completes
- Tasks 6 & 7 can run in parallel after Task 5 completes

---

## Rollback Plan

If issues arise:

1. **Immediate rollback**: Remove fixture from `conftest.py`, revert integration test changes
2. **Use manual setup**: Re-enable manual SQL script execution or use `task test-user-setup`
3. **Partial adoption**: Keep Taskfile commands and skill, but make fixture optional

---

## Verification Steps

After all tasks complete:

1. ✅ Manual command test: `task test-user-setup` creates user successfully
2. ✅ Skill test: `/test-user-setup` command works in Claude Code
3. ✅ Fresh environment test: `task deploy-local && task auto-test`
4. ✅ Idempotency test: `task auto-test` (run twice)
5. ✅ Docker environment test: `task full-auto-test-docker`
6. ✅ Cleanup test: Verify test user removed after session
7. ✅ Performance test: Check user provisioning adds <100ms overhead
8. ✅ Documentation test: New developer follows updated docs successfully
9. ✅ CI/CD test: Pipeline runs without manual pre-steps
