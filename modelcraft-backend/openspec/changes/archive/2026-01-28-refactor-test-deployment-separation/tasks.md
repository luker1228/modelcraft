# Implementation Tasks

**Design Update**: Based on user feedback, this implementation uses a **unified `.env` file** approach instead of separate test environment config files. Tests read configuration from the same `.env` file used for deployment, with additional test-specific variables (e.g., `CLEANUP_ENABLED`, `MODELCRAFT_BASE_URL`). Test commands are renamed to `auto-test*` to distinguish from unit tests (`make test`).

## Phase 1: Environment Configuration System

### Task 1.1: Add Test-Specific Variables to .env Files
**Estimated Effort**: Small
**Dependencies**: None
**Validation**: .env files contain test configuration variables

- [x] Update `.env` with test-specific variables:
  - `MODELCRAFT_BASE_URL=http://localhost:8080`
  - `CLEANUP_ENABLED=true`
- [x] Update `.env.example` with test-specific variables and documentation
- [x] Update `.env.autotest.example` with test-specific variables
- [x] Update `.env.docker.example` with test-specific variables
- [x] Add comments explaining test-specific usage

**Verification**:
```bash
# Files contain test configuration
cat .env | grep -E "MODELCRAFT_BASE_URL|CLEANUP_ENABLED"
cat .env.example | grep -E "MODELCRAFT_BASE_URL|CLEANUP_ENABLED"
```

---

### Task 1.2: Update Python Config Module
**Estimated Effort**: Small
**Dependencies**: Task 1.1
**Validation**: Config module reads from unified .env file

- [x] Modify `tests/config.py` to read from project root `.env` file
- [x] Add support for reading `MODELCRAFT_BASE_URL` from .env
- [x] Add `CLEANUP_ENABLED` attribute to `TestConfig` class
- [x] Maintain backward compatibility with existing hardcoded defaults
- [x] Use python-dotenv library to load .env file

**Verification**:
```python
# Test config loading
from config import config
assert config.get_base_url() == 'http://localhost:8080'
assert config.CLEANUP_ENABLED in [True, False]
assert config.DB_HOST is not None
```

---

## Phase 2: Smart Test Runner

### Task 2.1: Create Health Check Module
**Estimated Effort**: Small
**Dependencies**: Task 1.2
**Validation**: Health check correctly detects service availability

- [x] Create `tests/health_check.py` module
- [x] Implement `check_service_health(base_url: str) -> bool` function
  - Try `GET {base_url}/health` with 5-second timeout
  - Return `True` if 200 OK, `False` otherwise
- [x] Implement `check_database_connectivity(db_config: dict) -> bool` function
  - Attempt MySQL connection with configured credentials
  - Return `True` if successful, `False` otherwise
- [x] Implement `validate_environment(env: str)` function
  - Check both service and database health
  - Print clear error messages with troubleshooting suggestions
  - Exit with code 1 if validation fails

**Verification**:
```bash
# With service running
python3 -c "from health_check import check_service_health; print(check_service_health('http://localhost:8080'))"
# Should print: True

# With service stopped
# Should print: False with helpful error message
```

---

### Task 2.2: Develop Smart Test Runner Script
**Estimated Effort**: Medium
**Dependencies**: Task 2.1
**Validation**: Script runs tests with correct environment and cleanup control

- [x] Create `tests/run_smart_tests.sh` shell script
- [x] Parse command-line arguments:
  - `ENV` (default: local) - environment name
  - `CLEANUP` (optional) - yes/no to override config
  - `--cleanup-only` flag - run only Phase 2
  - `--no-cleanup` flag - skip Phase 2
- [x] Load environment configuration from project root `.env`
- [x] Validate environment health using `health_check.py`
- [x] Activate Python virtual environment (create if needed)
- [x] Execute Phase 1: Core API tests via pytest
  - Pass environment variables to pytest
  - Generate test report with environment metadata
- [x] Execute Phase 2: Resource cleanup (if enabled)
  - Run `cleanup_test_data.py` with environment context
  - Log cleanup statistics
- [x] Print summary with environment info and test results

**Verification**:
```bash
# Test script with different configurations
./tests/run_smart_tests.sh local
./tests/run_smart_tests.sh docker --no-cleanup
./tests/run_smart_tests.sh local --cleanup-only
```

---

### Task 2.3: Enhance Cleanup Script for Two-Phase Execution
**Estimated Effort**: Small
**Dependencies**: Task 1.2
**Validation**: Cleanup script respects environment configuration

- [x] Modify `tests/cleanup_test_data.py` to use `config` module
- [x] Add command-line argument parsing for environment selection
- [x] Load appropriate environment configuration before cleanup
- [x] Add `--dry-run` flag to preview cleanup without executing
- [x] Improve cleanup statistics reporting with environment context

**Verification**:
```bash
# Dry-run cleanup
python3 tests/cleanup_test_data.py --env local --dry-run
# Should show what would be cleaned without actually deleting

# Actual cleanup
python3 tests/cleanup_test_data.py --env docker
# Should clean test resources from Docker database
```

---

## Phase 3: Makefile Integration

### Task 3.1: Create Deployment Commands
**Estimated Effort**: Small
**Dependencies**: None
**Validation**: Deployment commands work independently

- [x] Add `deploy-local` target to Makefile
  - Uses `docker-compose.local.yml` for third-party services (MySQL, Redis)
  - Runs application natively: `go run cmd/server/main.go -env .env &`
  - Wait for health check to pass
  - Print success message with service URLs
- [x] Create `docker-compose.local.yml` for local development
  - Contains only MySQL and Redis (no application container)
  - Uses standard ports: MySQL@3306, Redis@6379
  - Separate volumes and network to avoid conflicts with full Docker stack
- [x] Add `deploy-docker` target to Makefile (already exists)
  - Start Docker Compose: `docker compose up -d`
  - Wait for all services to be healthy
  - Print success message with service URLs
- [x] Add `deploy-stop` target to Makefile (updated)
  - Stop local server (if running): `pkill -f "go run cmd/server/main.go"`
  - Stop Docker Compose full stack: `docker compose down`
  - Stop Docker Compose local services: `docker compose -f docker-compose.local.yml down`
  - Print confirmation message

**Implementation Notes**:
- Created `docker-compose.local.yml` with only third-party services for local development
- This approach provides better development experience: native app execution + containerized dependencies
- Separate container names (`*-local` suffix) and volumes to avoid conflicts
- Application runs natively for fast iteration and easy debugging

**Verification**:
```bash
# Deploy and stop local environment
make deploy-local
curl http://localhost:8080/health
make deploy-stop

# Deploy and stop Docker environment
make deploy-docker
curl http://localhost:8080/health
make deploy-stop
```

---

### Task 3.2: Create Test Commands
**Estimated Effort**: Small
**Dependencies**: Task 2.2, Task 3.1
**Validation**: Test commands work with deployed environments

**Design Note**: Test commands renamed to `auto-test*` to distinguish from unit tests (`make test`).

- [x] Add `auto-test` target to Makefile (default env: local)
  - Usage: `make auto-test`
  - Calls: `tests/run_smart_tests.sh local`
- [x] Add `auto-test-env` target with ENV parameter
  - Usage: `make auto-test-env ENV=docker`
  - Calls: `tests/run_smart_tests.sh $(ENV)`
- [x] Add `auto-test-no-cleanup` target
  - Usage: `make auto-test-no-cleanup ENV=local`
  - Calls: `tests/run_smart_tests.sh $(ENV) --no-cleanup`
- [x] Add `auto-test-cleanup-only` target
  - Usage: `make auto-test-cleanup-only ENV=docker`
  - Calls: `tests/run_smart_tests.sh $(ENV) --cleanup-only`
- [x] Add CLEANUP parameter support
  - Usage: `make auto-test ENV=local CLEANUP=no`
  - Passes `CLEANUP` variable to script

**Verification**:
```bash
# Test different combinations
make deploy-local
make auto-test                           # Default: local with cleanup
make auto-test ENV=local CLEANUP=no      # Local without cleanup
make auto-test-cleanup-only ENV=local    # Cleanup only

make deploy-docker
make auto-test ENV=docker                # Docker with cleanup
```

---

### Task 3.3: Add Convenience Workflow Commands
**Estimated Effort**: Small
**Dependencies**: Task 3.1, Task 3.2
**Validation**: Convenience commands execute full workflow

- [x] Add `full-auto-test-local` target
  - Executes: `make deploy-stop && make deploy-local && make auto-test ENV=local`
  - Provides one-command full workflow for local testing
- [x] Add `full-auto-test-docker` target
  - Executes: `make deploy-stop && make deploy-docker && make auto-test ENV=docker`
  - Provides one-command full workflow for Docker testing
- [x] Keep existing `automate-test` target unchanged for backward compatibility
  - Current behavior: Runs setup_test_db.sh + pytest
  - No changes needed to this command

**Verification**:
```bash
# Full workflow commands
make full-auto-test-local    # Deploy local + test + cleanup
make full-auto-test-docker   # Deploy Docker + test + cleanup

# Backward compatibility
make automate-test           # Should work as before (original implementation)
```

---

## Phase 4: Documentation and Validation

### Task 4.1: Update Test Documentation
**Estimated Effort**: Small
**Dependencies**: All previous tasks
**Validation**: Documentation is clear and comprehensive

- [x] Update `tests/README.md` with new workflow documentation
  - Explain test-deployment separation concept
  - Document all new Makefile commands with examples
  - Add troubleshooting section for common issues
- [x] Create `tests/WORKFLOW_GUIDE.md` with detailed usage scenarios (already exists)
  - Quick start guide
  - Environment configuration guide
  - Common workflows (local dev, Docker, cleanup control)
  - FAQ section
- [x] Update `CLAUDE.md` to reference new test workflow
  - Add section on test-deployment separation
  - Update testing commands in "Common Development Commands"
- [x] Add inline comments to `run_smart_tests.sh` and `health_check.py` (already exists)

**Verification**:
```bash
# Documentation exists and is up-to-date
cat tests/README.md
cat tests/WORKFLOW_GUIDE.md
cat CLAUDE.md | grep -A 10 "Testing"
```

---

### Task 4.2: End-to-End Testing and Validation
**Estimated Effort**: Medium
**Dependencies**: All previous tasks
**Validation**: All workflows work correctly

- [ ] Test local environment workflow
  - `make deploy-local` → verify service is running
  - `make auto-test ENV=local` → verify tests pass
  - `make auto-test ENV=local CLEANUP=no` → verify data persists
  - `make auto-test-cleanup-only ENV=local` → verify cleanup works
  - `make deploy-stop` → verify service stops
- [ ] Test Docker environment workflow
  - `make deploy-docker` → verify containers are healthy
  - `make auto-test ENV=docker` → verify tests pass
  - `make auto-test ENV=docker CLEANUP=no` → verify data persists
  - Verify database state using `docker exec` or phpMyAdmin
  - `make deploy-stop` → verify containers stop
- [ ] Test convenience workflows
  - `make full-auto-test-local` → verify complete flow
  - `make full-auto-test-docker` → verify complete flow
  - `make automate-test` → verify backward compatibility
- [ ] Test error scenarios
  - Run `make auto-test` without deployment → verify error message
  - Use invalid environment name → verify error handling
  - Use incorrect database credentials → verify error detection
- [ ] Test multiple consecutive test runs
  - Deploy once, run `make auto-test` 5 times → verify no redeployment

**Verification Checklist**:
```bash
# All commands should work correctly
✓ make deploy-local
✓ make deploy-docker
✓ make deploy-stop
✓ make auto-test
✓ make auto-test-env ENV=docker
✓ make auto-test-no-cleanup ENV=local
✓ make auto-test-cleanup-only ENV=local
✓ make full-auto-test-local
✓ make full-auto-test-docker
✓ make automate-test  # Original implementation, unchanged

# Error handling works
✓ make auto-test (without deployment) → clear error
✓ make auto-test ENV=invalid → clear error
✓ Invalid credentials → clear error

# Performance improvement
✓ 5 consecutive test runs < 50% of 5 deployments
```

---

### Task 4.3: Create Example Usage Guide
**Estimated Effort**: Small
**Dependencies**: Task 4.2
**Validation**: Examples are accurate and helpful

- [ ] Create `tests/EXAMPLES.md` with real-world scenarios
  - Scenario 1: Daily development workflow
  - Scenario 2: Debugging failed tests with data inspection
  - Scenario 3: Testing Docker before deployment
  - Scenario 4: CI/CD integration example
- [ ] Add example `.env` files with comments
- [ ] Create quick reference card (one-page cheat sheet)
- [ ] Add video/GIF demonstrations (optional)

**Verification**:
```bash
# Follow each example scenario and verify it works
cat tests/EXAMPLES.md
# Execute scenarios 1-4
```

---

## Task Summary

| Phase | Tasks | Effort | Parallelizable |
|-------|-------|--------|----------------|
| Phase 1: Configuration | 2 | Small | No (sequential) |
| Phase 2: Test Runner | 3 | Medium | Task 2.3 can be parallel with 2.2 |
| Phase 3: Makefile | 3 | Small | Task 3.1 can be parallel with Phase 2 |
| Phase 4: Documentation | 3 | Medium | After implementation complete |

**Total Estimated Effort**: 2-3 days for one developer

**Critical Path**: Task 1.1 → Task 1.2 → Task 2.1 → Task 2.2 → Task 3.2 → Task 4.2

**Parallelization Opportunities**:
- Task 3.1 (deployment commands) can be developed while Task 2.2 (test runner) is in progress
- Task 2.3 (cleanup enhancement) can be parallel with Task 2.2
- All documentation (Task 4.1, 4.3) can be written while testing (Task 4.2) is ongoing
