# test-user-provisioning Specification

## Purpose
TBD - created by archiving change optimize-integration-test-user-setup. Update Purpose after archive.
## Requirements
### Requirement: Automated Test User Creation

The test infrastructure SHALL automatically create a test user with owner role before integration tests execute.

**Rationale**: Eliminates manual SQL script execution and ensures deterministic test environment setup.

#### Scenario: Test user does not exist in database

**Given**:
- Integration test suite is about to run
- Test user (identified by `CASDOOR_TEST_USERNAME`) does not exist in `users` table
- "modelcraft" organization exists in `organizations` table
- "owner" role exists in `roles` table

**When**:
- Test session starts
- `test_user_with_owner_role` fixture is invoked

**Then**:
- User record is created in `users` table with:
  - `id`: UUID derived from `CASDOOR_TEST_USERNAME` (deterministic)
  - `external_id`: Value from `CASDOOR_TEST_USERNAME` environment variable
  - `name`: "Test Integration User"
  - `phone`: Empty string
- User-organization association is created in `user_organizations` table with:
  - `user_id`: The created user's ID
  - `org_id`: ID of "modelcraft" organization
  - `role_id`: ID of "owner" role
  - `status`: "active"
- Fixture returns user data dict containing: `{id, external_id, name, org_name, role_name}`
- Log message confirms: "✅ Created test user with owner role: <external_id>"

#### Scenario: Test user already exists in database

**Given**:
- Integration test suite is about to run
- Test user already exists in `users` table (from previous test run or manual setup)
- User is already assigned to "modelcraft" organization with "owner" role

**When**:
- Test session starts
- `test_user_with_owner_role` fixture is invoked

**Then**:
- No duplicate user record is created (idempotent operation)
- No duplicate user-organization association is created
- Fixture returns existing user data dict
- Log message confirms: "✅ Test user already exists: <external_id>"

---

### Requirement: Database Helper Functions

The test infrastructure SHALL provide reusable database helper functions for user and role management.

**Rationale**: Encapsulates database operations and ensures consistency across test fixtures.

#### Scenario: Ensure user exists (new user)

**Given**:
- Database connection is available
- User with given `user_id` does not exist

**When**:
- `ensure_user_exists(db_config, user_id, external_id, name)` is called

**Then**:
- User record is inserted into `users` table
- Function returns user data dict: `{id, external_id, name, created_at}`
- No exception is raised

#### Scenario: Ensure user exists (existing user)

**Given**:
- Database connection is available
- User with given `user_id` already exists

**When**:
- `ensure_user_exists(db_config, user_id, external_id, name)` is called

**Then**:
- No duplicate user record is created
- Function returns existing user data dict
- No exception is raised (idempotent)

#### Scenario: Ensure user has role assignment

**Given**:
- Database connection is available
- User exists in `users` table
- Organization "modelcraft" exists
- Role "owner" exists
- User is not yet assigned to the organization

**When**:
- `ensure_user_has_role(db_config, user_id, org_name="modelcraft", role_name="owner")` is called

**Then**:
- `role_id` is looked up from `roles` table by `role_name`
- `org_id` is looked up from `organizations` table by `org_name`
- User-organization association record is created in `user_organizations` table
- Function returns association data dict: `{user_id, org_id, role_id, status}`
- No exception is raised

#### Scenario: Ensure user has role (already assigned)

**Given**:
- User is already assigned to organization with specified role

**When**:
- `ensure_user_has_role(db_config, user_id, org_name, role_name)` is called

**Then**:
- No duplicate association record is created
- Function returns existing association data
- No exception is raised (idempotent)

---

### Requirement: Test User Cleanup

The test infrastructure SHALL clean up test users after test session completes.

**Rationale**: Ensures clean state between test runs and prevents test data accumulation.

#### Scenario: Cleanup test user after successful tests

**Given**:
- Test session has completed successfully
- Test user was created by `test_user_with_owner_role` fixture
- `KEEP_TEST_USER` environment variable is not set (or set to "false")

**When**:
- Fixture teardown/finalizer is invoked

**Then**:
- User-organization associations are deleted from `user_organizations` table for the test user
- User record is deleted from `users` table
- Log message confirms: "🧹 Cleaned up test user: <external_id>"
- No exception is raised even if cleanup fails (logged as warning)

#### Scenario: Preserve test user for debugging

**Given**:
- Test session has completed (may have failed)
- Test user was created by `test_user_with_owner_role` fixture
- `KEEP_TEST_USER` environment variable is set to "true"

**When**:
- Fixture teardown/finalizer is invoked

**Then**:
- Test user is NOT deleted
- User-organization associations are NOT deleted
- Log message confirms: "ℹ️ Keeping test user for debugging (KEEP_TEST_USER=true)"

#### Scenario: Handle cleanup failures gracefully

**Given**:
- Test session has completed
- Database connection fails during cleanup

**When**:
- Fixture teardown attempts to clean up test user

**Then**:
- Exception is caught and logged as warning
- Cleanup failure does NOT cause test session to fail
- Log message warns: "⚠️ Failed to cleanup test user: <error_message>"

---

### Requirement: Integration Test User Dependency

Integration tests SHALL declare dependency on the test user provisioning fixture.

**Rationale**: Ensures test user is provisioned before integration tests run and makes dependencies explicit.

#### Scenario: Integration test uses provisioned test user

**Given**:
- Integration test class/function is defined
- Test requires authenticated user with owner permissions

**When**:
- Test declares `test_user_with_owner_role` fixture parameter
- Test executes

**Then**:
- Fixture runs before test execution
- Test receives user data dict from fixture
- Test can use `auth_token` fixture which authenticates as the provisioned user
- Test completes without permission errors

---

### Requirement: Configuration from Environment

Test user provisioning SHALL read configuration from environment variables.

**Rationale**: Supports different test environments (local, docker, CI/CD) without code changes.

#### Scenario: Read test user credentials from environment

**Given**:
- `.env` file contains:
  - `CASDOOR_TEST_USERNAME=test-integration`
  - `DB_HOST=localhost`
  - `DB_USER=root`
  - `DB_PASSWORD=password`
  - `DB_NAME=modelcraft`

**When**:
- `test_user_with_owner_role` fixture is invoked

**Then**:
- User is created with `external_id="test-integration"`
- Database operations use connection parameters from environment
- No hard-coded values are used

#### Scenario: Optional cleanup configuration

**Given**:
- `.env` file contains `KEEP_TEST_USER=true`

**When**:
- Test session completes

**Then**:
- Test user is preserved (not cleaned up)

---

### Requirement: Error Handling and Logging

Test user provisioning SHALL provide clear error messages and logging.

**Rationale**: Aids debugging when provisioning fails and provides visibility into fixture operations.

#### Scenario: Database connection failure

**Given**:
- Database is not running or unreachable
- Test session attempts to start

**When**:
- `test_user_with_owner_role` fixture attempts to provision user

**Then**:
- Fixture fails immediately with clear error message
- Error message includes:
  - "Failed to connect to database"
  - Database host and port from configuration
  - Suggestion: "Ensure database is running and accessible"
- Test session is aborted (no tests run with invalid state)

#### Scenario: Missing organization or role

**Given**:
- "modelcraft" organization does not exist in database
- OR "owner" role does not exist in database

**When**:
- `ensure_user_has_role()` is called

**Then**:
- Function raises clear exception
- Error message includes:
  - "Organization 'modelcraft' not found" OR "Role 'owner' not found"
  - Suggestion: "Run database migrations (task deploy-local)"
- Test session is aborted

---

### Requirement: Taskfile Commands for Manual Setup

The build automation SHALL provide Taskfile commands for manual test user setup and cleanup.

**Rationale**: Supports manual setup workflows, CI/CD scripts, and provides quick access without running full test suite.

#### Scenario: Create test user via Taskfile command

**Given**:
- Database is running and accessible
- `.env` file contains database configuration
- User runs `task test-user-setup`

**When**:
- Command executes `python tests/common/test_user_setup.py`

**Then**:
- Test user is created with owner role (idempotent)
- Success message is displayed: "✅ Test user created successfully"
- User details are shown (ID, external ID, name, organization, role)
- Exit code is 0 on success

#### Scenario: Cleanup test user via Taskfile command

**Given**:
- Test user exists in database
- User runs `task test-user-cleanup`

**When**:
- Command executes cleanup function from test_user_setup.py

**Then**:
- Test user and associations are removed
- Success message is displayed: "🧹 Test user cleaned up"
- Exit code is 0 on success

#### Scenario: Handle setup errors gracefully

**Given**:
- Database is not running
- User runs `task test-user-setup`

**When**:
- Command attempts to connect to database

**Then**:
- Error message is displayed: "❌ Failed to connect to database"
- Suggestion is provided: "Ensure database is running (task deploy-local)"
- Exit code is non-zero
- Command fails without creating partial state

---

### Requirement: Claude Code Skill for Quick Access

The development environment SHALL provide a Claude Code skill for test user setup.

**Rationale**: Provides quick, discoverable access to test user setup for developers using Claude Code CLI.

#### Scenario: Setup test user via Claude Code skill

**Given**:
- Developer is using Claude Code CLI
- Database is running
- User executes `/test-user-setup` command

**When**:
- Skill executes `task test-user-setup`

**Then**:
- Test user is created with owner role
- Skill displays success message with user details
- Skill suggests next steps (e.g., "Run 'task auto-test' to execute tests")

#### Scenario: Handle setup errors in skill

**Given**:
- Database is not running
- User executes `/test-user-setup` command

**When**:
- Skill attempts to run `task test-user-setup`

**Then**:
- Skill displays error message
- Skill provides troubleshooting guidance:
  - "Database connection failed"
  - "Try: task deploy-local"
  - "Verify database is running: docker ps"
- Skill does not fail silently

#### Scenario: Cleanup test user via skill

**Given**:
- Test user exists
- User executes `/test-user-cleanup` command

**When**:
- Skill executes `task test-user-cleanup`

**Then**:
- Test user is removed
- Success confirmation is displayed
- Skill does not throw errors if user doesn't exist (idempotent)

---

