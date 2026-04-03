# Spec: Test User Provisioning

## Overview

This specification defines the automated test user provisioning capability for integration tests. It ensures a test user with owner role exists before integration tests run, eliminating manual database setup requirements.

## ADDED Requirements

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

## Implementation Notes

### Test User Setup Module Location
`tests/common/test_user_setup.py` (already exists, enhance for standalone execution)

### Key Functions

```python
def execute_test_user_setup(
    db_config: dict
) -> dict:
    """
    Create test user with owner role (idempotent).

    Reads CASDOOR_TEST_USERNAME from environment.
    Uses fixed UUID for deterministic behavior.

    Returns: {id, external_id, name, org_name, role_name, status}
    Raises: RuntimeError on failure (connection, missing org/role)
    """

def cleanup_test_user(
    db_config: dict,
    user_id: str = '487101d6-92bb-459e-b4f1-426255126d27'
) -> None:
    """
    Remove user and associated records.

    Raises: No exceptions (logs warnings on failure)
    """
```

### Taskfile Commands

**Root Taskfile.yml**:
```yaml
test-user-setup:
  desc: Create test-integration user with owner role
  cmds:
    - |
      echo "🚀 Setting up test user..."
      cd tests && python -m common.test_user_setup
      echo "✅ Test user setup complete"

test-user-cleanup:
  desc: Remove test-integration user
  cmds:
    - |
      echo "🧹 Cleaning up test user..."
      cd tests && python -c "from common.test_user_setup import cleanup_test_user, TestConfig; cleanup_test_user(TestConfig().get_db_config())"
      echo "✅ Cleanup complete"
```

**tests/Taskfile.yml**:
```yaml
test-user-setup:
  desc: Create test-integration user with owner role
  cmds:
    - python -m common.test_user_setup

test-user-cleanup:
  desc: Remove test-integration user
  cmds:
    - python -c "from common.test_user_setup import cleanup_test_user, TestConfig; cleanup_test_user(TestConfig().get_db_config())"
```

### Claude Code Skills

**.claudecode/skills/test-user-setup.md**:
```markdown
# Test User Setup Skill

Creates test-integration user with owner role for integration testing.

## Usage
/test-user-setup

## What it does
- Creates test user with fixed UUID (487101d6-92bb-459e-b4f1-426255126d27)
- Assigns owner role in modelcraft organization
- Idempotent: safe to run multiple times

## Requirements
- Database must be running (task deploy-local)
- Database migrations must be applied

## Error handling
If setup fails, check:
- Database is running: docker ps
- Database is accessible: task login-db
- Migrations are applied: task deploy-local
```

**.claudecode/skills/test-user-cleanup.md**:
```markdown
# Test User Cleanup Skill

Removes test-integration user and associated data.

## Usage
/test-user-cleanup

## What it does
- Deletes user_organizations associations
- Deletes user record
- Idempotent: safe to run even if user doesn't exist

## Requirements
- Database must be running
```

### Fixture Definition

```python
@pytest.fixture(scope="session")
def test_user_with_owner_role(test_config):
    """
    Session-scoped fixture: provision test user with owner role.

    Uses execute_test_user_setup() from tests/common/test_user_setup.py

    Returns: {id, external_id, name, org_name, role_name}
    """
    from common.test_user_setup import execute_test_user_setup, cleanup_test_user
    from common.config import TestConfig

    config = TestConfig()
    db_config = config.get_db_config()

    # Setup
    user_info = execute_test_user_setup(db_config)
    print(f"✅ Test user created: {user_info['external_id']}")

    yield user_info

    # Teardown
    keep_user = os.getenv('KEEP_TEST_USER', 'false').lower() == 'true'
    if not keep_user:
        cleanup_test_user(db_config, user_info['id'])
        print(f"🧹 Cleaned up test user: {user_info['external_id']}")
    else:
        print(f"ℹ️ Keeping test user for debugging (KEEP_TEST_USER=true)")
```

### UUID Generation Strategy

Use fixed UUID for test user:

```python
# In setup_test_user.sql and test_user_setup.py
test_user_id = '487101d6-92bb-459e-b4f1-426255126d27'
```

This ensures:
- Same user ID across test runs (idempotency)
- Consistency with existing SQL scripts
- No collisions with production UUIDs

Alternative (if dynamic generation needed):
```python
import uuid

external_id = os.getenv('CASDOOR_TEST_USERNAME', 'test-integration')
user_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"modelcraft-test-{external_id}"))
```

---

## Testing Requirements

1. **Unit Tests** for `test_user_setup.py`:
   - Test `execute_test_user_setup()` with fresh database
   - Test idempotency (run twice, verify no errors)
   - Test error handling (missing org, missing role, DB connection failure)
   - Test `cleanup_test_user()` function

2. **Taskfile Command Tests**:
   - Test `task test-user-setup` with fresh database
   - Test `task test-user-setup` with existing user (idempotent)
   - Test `task test-user-cleanup`
   - Verify error messages and exit codes

3. **Claude Code Skill Tests**:
   - Test `/test-user-setup` command
   - Test `/test-user-cleanup` command
   - Verify error handling and user feedback

4. **Integration Tests** for fixture:
   - Test fixture with clean database
   - Test fixture with existing user
   - Test cleanup behavior
   - Test `KEEP_TEST_USER` flag

5. **End-to-End Tests**:
   - Run full integration test suite with fresh database
   - Verify no permission errors
   - Verify user provisioning logs appear
   - Verify cleanup logs appear

---

## Dependencies

- **Database Schema**: Requires tables `users`, `roles`, `organizations`, `user_organizations` (already exist)
- **Test Configuration**: Requires `TestConfig.get_db_config()` (already exists in `tests/common/config.py`)
- **Test User Setup Module**: `tests/common/test_user_setup.py` (already exists)
- **Test User SQL Script**: `tests/setup_test_user.sql` (already exists)
- **Environment Variables**: Requires `CASDOOR_TEST_USERNAME`, `DB_*` variables (already required)
- **Taskfile**: go-task installed (already required for project)
- **Claude Code**: .claudecode directory structure (will be created)

---

## Non-Functional Requirements

- **Performance**: User provisioning should add <100ms overhead to test session startup
- **Idempotency**: Running provisioning multiple times must not cause errors or duplicates
- **Reliability**: Provisioning failures must prevent tests from running (fail fast)
- **Maintainability**: Database helper functions should be reusable for other test fixtures
- **Discoverability**: Taskfile commands and skills should be easy to find (`task --list`, skill autocomplete)
- **Error Messages**: All error messages should be clear and include troubleshooting steps
