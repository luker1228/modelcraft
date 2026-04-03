## ADDED Requirements

### Requirement: Environment-Based Test Configuration
**ID**: test-automation.env-config
**Category**: Configuration
**Priority**: High

The test framework SHALL support environment-specific configuration files that define target deployment environments without coupling test execution to deployment setup.

#### Scenario: Switch Test Target from Local to Docker Environment

**Given** the application is deployed in Docker Compose with MySQL on port 6033
**And** a local development environment also exists on port 8080
**When** the developer runs `make test ENV=docker`
**Then** the test framework loads `tests/envs/docker.env` configuration
**And** connects to `http://localhost:8080` with Docker database credentials
**And** executes all test cases against the Docker environment
**And** does not modify or restart the running Docker containers

#### Scenario: Configure Test Without Deployment

**Given** ModelCraft is already running on `http://localhost:8080`
**When** the developer runs `make test ENV=local`
**Then** the test framework validates service health at the configured URL
**And** proceeds with test execution without initializing or modifying deployment
**And** reports clear errors if the service is not reachable

---

### Requirement: Two-Phase Test Execution
**ID**: test-automation.two-phase
**Category**: Test Execution
**Priority**: High

The test framework SHALL separate test execution into two distinct phases: core API testing (Phase 1) and resource cleanup (Phase 2), with independent control over each phase.

#### Scenario: Run Core Tests with Delete Operations

**Given** the test environment is configured with `CLEANUP_ENABLED=false`
**When** Phase 1 core tests execute
**Then** tests create projects, clusters, models, and enums
**And** tests can perform delete operations on test resources
**And** tests verify the results of delete operations (e.g., resource not found)
**And** other test data remains in the database after Phase 1 completes
**And** Phase 2 cleanup is skipped

#### Scenario: Run Full Test Suite with Automatic Cleanup

**Given** the test environment is configured with `CLEANUP_ENABLED=true`
**When** Phase 1 core tests complete successfully
**Then** Phase 2 resource cleanup executes automatically
**And** cleanup removes all test projects except 'default'
**And** cleanup removes test clusters, models, and enums from 'default' project
**And** cleanup reports the count of cleaned resources
**And** the test database is left in a clean state

#### Scenario: Run Cleanup Independently

**Given** previous test runs left test data in the database
**When** the developer runs `make test-cleanup-only ENV=local`
**Then** only Phase 2 cleanup logic executes
**And** Phase 1 core tests are skipped
**And** all test resources are removed from the database

---

### Requirement: Parameterized Cleanup Control
**ID**: test-automation.cleanup-control
**Category**: Test Execution
**Priority**: Medium

The test framework SHALL allow runtime control of cleanup behavior through command-line parameters, overriding configuration file defaults.

#### Scenario: Override Cleanup Configuration at Runtime

**Given** `tests/envs/local.env` has `CLEANUP_ENABLED=true`
**When** the developer runs `make test ENV=local CLEANUP=no`
**Then** the runtime parameter overrides the config file setting
**And** Phase 1 core tests execute normally
**And** Phase 2 cleanup is skipped
**And** test data is preserved in the database for inspection

#### Scenario: Enable Cleanup via Runtime Parameter

**Given** `tests/envs/docker.env` has `CLEANUP_ENABLED=false`
**When** the developer runs `make test ENV=docker CLEANUP=yes`
**Then** the runtime parameter overrides the config file setting
**And** Phase 2 cleanup executes after Phase 1 completes
**And** all test resources are removed

---

### Requirement: Deployment-Independent Test Execution
**ID**: test-automation.deploy-independent
**Category**: Workflow
**Priority**: High

The test framework SHALL enable repeated test execution against a deployed environment without requiring redeployment between test runs.

#### Scenario: Multiple Test Runs Without Redeployment

**Given** the Docker environment is deployed via `make deploy-docker`
**When** the developer runs `make test ENV=docker` 5 times consecutively
**Then** each test run executes against the same deployed environment
**And** no database re-initialization occurs between runs
**And** no container restarts occur between runs
**And** each test run completes successfully with independent test data
**And** total execution time is less than 50% of deploying 5 times

#### Scenario: Test After Code Changes

**Given** ModelCraft is running in local development mode
**And** initial tests pass with `make test ENV=local CLEANUP=no`
**When** the developer modifies a GraphQL resolver
**And** rebuilds the application with `make build && make run`
**And** runs `make test ENV=local` again
**Then** tests execute against the updated code
**And** no database re-initialization is required
**And** test results reflect the code changes

---

### Requirement: Environment Health Validation
**ID**: test-automation.health-check
**Category**: Reliability
**Priority**: Medium

The test framework SHALL validate that the target environment is healthy and reachable before executing tests, providing clear error messages if validation fails.

#### Scenario: Detect Unreachable Service

**Given** no ModelCraft service is running on `http://localhost:8080`
**When** the developer runs `make test ENV=local`
**Then** the test framework performs a health check to `/health` endpoint
**And** detects that the service is unreachable
**And** displays error message: "ModelCraft service is not running at http://localhost:8080. Please deploy first with: make deploy-local"
**And** exits without executing any tests
**And** does not create or modify any test data

#### Scenario: Validate Database Connectivity

**Given** ModelCraft service is running but database credentials in `tests/envs/local.env` are incorrect
**When** the developer runs `make test ENV=local`
**Then** the test framework validates database connectivity using configured credentials
**And** detects database connection failure
**And** displays error message with database host, port, and authentication status
**And** exits without executing tests

---

### Requirement: Test Report with Environment Context
**ID**: test-automation.contextual-reports
**Category**: Reporting
**Priority**: Low

The test framework SHALL generate test reports that include environment configuration context to aid debugging and auditing.

#### Scenario: Generate Report with Environment Metadata

**Given** tests are executed with `make test ENV=docker`
**When** the test suite completes
**Then** the generated HTML report includes a metadata section
**And** the metadata shows "Target Environment: docker"
**And** the metadata shows "Service URL: http://localhost:8080"
**And** the metadata shows "Database: localhost:6033/modelcraft"
**And** the metadata shows "Cleanup Enabled: true"
**And** the metadata shows test execution timestamp and duration

---

### Requirement: Backward Compatibility with Existing Workflow
**ID**: test-automation.backward-compat
**Category**: Compatibility
**Priority**: High

The test framework SHALL maintain backward compatibility with the existing `make automate-test` command for users who prefer the integrated workflow.

#### Scenario: Existing Command Still Works

**Given** a developer is familiar with `make automate-test`
**When** the developer runs `make automate-test`
**Then** the command executes the legacy integrated workflow
**And** sets up the test database via `setup_test_db.sh`
**And** runs pytest against local environment
**And** cleans up test data automatically
**And** produces the same test report format as before

#### Scenario: Legacy Command Uses New Infrastructure

**Given** the new test framework is implemented
**When** a developer runs `make automate-test`
**Then** internally, the command uses the new test infrastructure
**And** temporarily loads the `tests/envs/local.env` configuration
**And** executes both Phase 1 and Phase 2 with cleanup enabled
**And** the user experience remains unchanged from the previous version
