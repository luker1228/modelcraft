# Proposal: Refactor Test-Deployment Separation

## Summary

Refactor the automated test workflow to completely separate testing from deployment. This enables testing against any deployed environment (local development or Docker) with configurable test data cleanup, addressing the current tight coupling between `make automate-test` and deployment setup.

## Why

### Problem
The current `make automate-test` command tightly couples testing with deployment (database initialization, server startup), making it inefficient and inflexible:
- Developers must restart services and reinitialize databases for every test run, even for minor code changes
- Cannot test against already-running environments (local dev or Docker)
- No ability to preserve test data for debugging failed tests
- Core API tests and cleanup are mixed, lacking separation of concerns

### Impact
This separation enables:
- **Faster iteration**: Deploy once, test multiple times (5x faster workflow)
- **Flexible testing**: Target local, Docker, or remote environments easily
- **Better debugging**: Preserve test data by disabling cleanup
- **Clear workflow**: Explicit separation between deployment and testing phases

## Problem Statement

### Current Issues

1. **Test-Deployment Coupling**: The current `make automate-test` command tightly couples testing with deployment setup (database initialization, server startup), making it impossible to test already-running environments.

2. **Environment Inflexibility**: Cannot easily switch test targets between local development, Docker, or other deployed environments without modifying test code or scripts.

3. **Mixed Concerns**: Test execution includes both core API testing and data cleanup, but these concerns are not clearly separated or controllable.

4. **Workflow Inefficiency**: Users must restart services and reinitialize databases even when testing minor changes, wasting time and resources.

### User Story

As a developer, I want to:
- Deploy my application once (local or Docker)
- Run tests multiple times against the deployed environment
- Control whether to preserve or cleanup test data after tests
- Clearly separate core API tests from resource cleanup operations

**Current Workflow** (Inefficient):
```bash
# Every test run requires full setup
make automate-test  # Sets up DB + starts server + runs tests + cleanup
```

**Desired Workflow** (Efficient):
```bash
# Deploy once
make deploy-local   # or make deploy-docker

# Test multiple times
make auto-test ENV=local              # Test with data cleanup
make auto-test ENV=local CLEANUP=no   # Test without cleanup
make auto-test ENV=docker             # Test Docker environment
```

## Proposed Solution

### Architecture Overview

Implement a **configuration-driven test framework** with three core components:

```
┌─────────────────────────────────────────────────────┐
│           Deployment Layer (Independent)             │
│  make deploy-local / make deploy-docker              │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│         Test Configuration Layer                     │
│  Unified .env file for deployment and testing        │
│  - MODELCRAFT_BASE_URL                               │
│  - DB_* credentials (shared for deploy & test)       │
│  - CLEANUP_ENABLED=true/false                        │
│  tests/config.py reads from same .env file           │
└─────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────┐
│         Test Execution Layer                         │
│  1. Core API Tests (test_*.py)                       │
│  2. Resource Cleanup (cleanup_test_data.py)          │
└─────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Separation of Concerns**:
   - **Deployment**: Independent commands for local/Docker deployment
   - **Testing**: Environment-agnostic test runner with config selection
   - **Cleanup**: Optional, controlled by configuration parameters

2. **Configuration-Driven**:
   - Unified `.env` file shared between deployment and testing
   - Test configuration reads from same `.env` file (via tests/config.py)
   - Runtime parameters override config (e.g., `CLEANUP=no`)
   - No duplicate configuration files to maintain

3. **Two-Phase Testing**:
   - **Phase 1**: Core API tests (can include delete operations with verification)
   - **Phase 2**: Resource cleanup (optional, controlled by `CLEANUP_ENABLED`)

### Component Design

#### 1. Unified Environment Configuration

**Design Decision**: Use a single `.env` file for both deployment and testing, with additional test-specific variables:

```bash
# .env (unified for deployment and testing)
# Server Configuration
PORT=8080
GIN_MODE=debug

# Database Configuration (shared by deployment and tests)
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=Root@SecurePass123#
DB_DATABASE=modelcraft  # For Docker: modelcraft, For local: modelcraft_test

# Application URL (used by tests)
MODELCRAFT_BASE_URL=http://localhost:8080

# Test Configuration
CLEANUP_ENABLED=true  # Control test data cleanup
```

**Benefits**:
- Single source of truth for configuration
- Easier to maintain (no duplicate config files)
- Tests and deployment always use consistent configuration
- Environment-specific values handled by different .env files (.env for local, docker-compose env vars for Docker)

#### 2. Smart Test Runner Script

```bash
# tests/run_smart_tests.sh
#!/bin/bash
# Usage: ./run_smart_tests.sh [ENV] [OPTIONS]
# ENV: local (default) | docker
# OPTIONS: --no-cleanup | --cleanup-only

Features:
- Load environment-specific configuration
- Validate target service health
- Execute core API tests (Phase 1)
- Execute cleanup tests (Phase 2, optional)
- Generate test reports with environment context
```

#### 3. Makefile Integration

```makefile
# Deployment commands (independent)
deploy-local:       # Start local development environment
deploy-docker:      # Start Docker Compose environment
deploy-stop:        # Stop running environment

# Test commands (independent, work with any deployed environment)
auto-test:          # Run tests with default env (local)
auto-test-env:      # Run tests with specific env: make auto-test-env ENV=docker
auto-test-no-cleanup:    # Run tests without cleanup
auto-test-cleanup-only:  # Run cleanup only
```

## Benefits

### For Developers

1. **Faster Iteration**: Deploy once, test multiple times without redeployment
2. **Flexibility**: Test against local, Docker, or remote environments easily
3. **Debugging**: Preserve test data for inspection by disabling cleanup
4. **Clear Workflow**: Explicit separation between deployment and testing

### For Maintenance

1. **Modular Design**: Each component (deployment, config, tests, cleanup) is independent
2. **Simplified Configuration**: Single `.env` file for deployment and testing
3. **Clear Separation**: Core tests vs cleanup logic are distinct and controllable
4. **Reduced Duplication**: No need to maintain multiple environment-specific config files

### Future Extensibility (Out of Scope for This Change)

The architecture naturally supports future enhancements:
- CI/CD integration: Same test suite can be reused in automated pipelines
- Environment parity: Configuration-driven approach works for staging/production
- Parallel testing: Multiple test jobs can target the same deployed environment

## Trade-offs and Considerations

### Advantages
- ✅ Complete decoupling of deployment and testing
- ✅ Flexible environment targeting
- ✅ Controllable test data management
- ✅ Faster test iteration cycles

### Challenges
- ⚠️ Requires explicit deployment before testing (not automatic)
- ⚠️ Need to add test-specific variables to existing `.env` file
- ⚠️ Users must understand the two-step workflow (deploy → test)
- ⚠️ Different .env files for local vs Docker (but simpler than multiple test config files)

### Mitigations
- Provide convenience commands: `make full-test-local` (deploy + test)
- Clear documentation and examples
- Health check validation before test execution
- Helpful error messages if environment is not running

## Implementation Scope

### In Scope
1. Unified `.env` file configuration for deployment and testing
2. Python config module updated to read from `.env` file
3. Smart test runner script with health checks
4. Makefile commands renamed to `auto-test*` for automated testing
5. Two-phase test execution (core tests + optional cleanup)
6. Documentation and usage examples

### Out of Scope
1. CI/CD pipeline configuration (GitHub Actions) - future enhancement
2. Production environment testing - requires additional security considerations
3. Performance/load testing - different testing concern
4. Refactoring existing test cases - tests remain unchanged

## Success Criteria

1. **Deployment Independence**: Can deploy once and run tests 10 times without redeployment
2. **Environment Switching**: Can switch from local to Docker testing in under 10 seconds
3. **Cleanup Control**: Can run tests with and without cleanup via configuration
4. **Backward Compatibility**: Existing `make automate-test` still works for convenience
5. **Documentation**: Complete usage guide with examples for all scenarios

## Related Work

- Existing test infrastructure: `tests/automated/`, `tests/cleanup_test_data.py`
- Current deployment: Docker Compose, local Go server startup
- Configuration system: `config.yaml`, `.env` files
- OpenSpec specs: `deployment`, `docker`, `config-env-selection`

## Next Steps

1. Review and approve this proposal
2. Create detailed implementation tasks
3. Implement environment configuration system
4. Develop smart test runner script
5. Update Makefile and documentation
6. Validate with both local and Docker environments
