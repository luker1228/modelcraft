# design-test-auth Specification

## Purpose
TBD - created by archiving change enforce-design-test-auth. Update Purpose after archive.
## Requirements
### Requirement: Mandatory Authentication for Design Tests

Design-time integration tests SHALL require valid authentication credentials to execute.

#### Scenario: Design tests fail when credentials missing
- **GIVEN** Design tests are executed in `tests/design/`
- **WHEN** `CASDOOR_TEST_USERNAME` or `CASDOOR_TEST_PASSWORD` are not configured
- **THEN** the test session SHALL fail during fixture setup
- **AND** SHALL raise `RuntimeError` with message "Authentication required: CASDOOR_TEST_USERNAME and CASDOOR_TEST_PASSWORD must be configured in .env file"
- **AND** no tests SHALL execute

**Rationale**: Design API requires authentication in production. Tests must validate authenticated behavior.

#### Scenario: Design tests execute when credentials present
- **GIVEN** Design tests are executed in `tests/design/`
- **WHEN** `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD` are properly configured
- **THEN** the test session SHALL acquire a JWT access token
- **AND** SHALL pass the token to all `graphql_client` fixtures
- **AND** tests SHALL execute normally with authentication

**Rationale**: Valid credentials enable proper end-to-end testing of authenticated APIs.

---

### Requirement: Design-Specific Auth Token Fixture

The `tests/design/conftest.py` SHALL override the root `auth_token` fixture to enforce mandatory authentication.

#### Scenario: Override session-scoped auth_token fixture
- **GIVEN** `tests/design/conftest.py` is loaded by pytest
- **WHEN** tests request the `auth_token` fixture
- **THEN** the Design-specific fixture SHALL take precedence over the root fixture
- **AND** SHALL have `scope="session"` to match root fixture scope
- **AND** SHALL validate credentials before attempting token acquisition

**Rationale**: Fixture override leverages pytest's built-in precedence rules without requiring test signature changes.

#### Scenario: Fail fast on missing credentials
- **GIVEN** the `auth_token` fixture is executed
- **WHEN** `test_config.CASDOOR_TEST_USERNAME` is empty or `None`
- **OR** `test_config.CASDOOR_TEST_PASSWORD` is empty or `None`
- **THEN** SHALL raise `RuntimeError` immediately
- **AND** SHALL NOT attempt to acquire token from Casdoor
- **AND** error message SHALL clearly indicate missing `.env` configuration

**Rationale**: Early validation provides clear feedback before network calls.

#### Scenario: Acquire token when credentials valid
- **GIVEN** the `auth_token` fixture is executed
- **WHEN** both `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD` are configured
- **THEN** SHALL call `get_test_access_token(test_config)`
- **AND** SHALL return the JWT access token on success
- **AND** SHALL log success message with token length
- **AND** SHALL propagate exceptions from token acquisition

**Rationale**: Reuses existing token acquisition logic from `common.auth`.

---

### Requirement: Clear Error Messages

Authentication errors SHALL provide actionable error messages that guide developers to fix configuration issues.

#### Scenario: Error message includes specific variables
- **GIVEN** authentication fails due to missing credentials
- **WHEN** the error is raised
- **THEN** error message SHALL mention `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD`
- **AND** SHALL mention `.env file` as the configuration source
- **AND** SHALL use format: "Authentication required: CASDOOR_TEST_USERNAME and CASDOOR_TEST_PASSWORD must be configured in .env file"

**Rationale**: Specific variable names help developers quickly identify and fix configuration issues.

#### Scenario: Token acquisition failure propagates
- **GIVEN** credentials are configured but token acquisition fails
- **WHEN** `get_test_access_token()` raises an exception
- **THEN** the exception SHALL propagate to pytest
- **AND** SHALL include original error details from `common.auth`
- **AND** SHALL indicate whether the failure is due to network, credentials, or Casdoor availability

**Rationale**: Detailed error messages help diagnose authentication system issues.

---

