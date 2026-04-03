# Tasks: Reorganize Python Integration Tests

## Phase 1: Preparation and Scaffolding ✅

- [ ] **Task 1.1**: Create design directory structure
  - Create `tests/design/` subdirectories: `common/`, `project/`, `cluster/`, `model/`, `compatibility/`
  - Add `__init__.py` to each subdirectory
  - Create `tests/design/conftest.py` for shared Design fixtures
  - Create `tests/design/README.md` documenting Design-time tests
  - **Validation**: Directory structure exists and all `__init__.py` files are present

- [ ] **Task 1.2**: Create runtime directory structure
  - Create `tests/runtime/` subdirectories: `query/`, `integration/`
  - Add `__init__.py` to each subdirectory
  - Create `tests/runtime/conftest.py` for shared Runtime fixtures
  - Create `tests/runtime/README.md` documenting Runtime tests and dependencies on Design
  - **Validation**: Directory structure exists and all `__init__.py` files are present

- [ ] **Task 1.3**: Create common utilities directory
  - Create `tests/common/` directory with `__init__.py`
  - Move `config.py`, `health_check.py`, `cleanup_test_data.py` to `tests/common/`
  - **Validation**: Files moved, no import errors in existing tests

- [ ] **Task 1.4**: Create shared Design infrastructure
  - Create `tests/design/common/graphql_client.py` with client utilities
  - Create `tests/design/common/fixtures.py` with common test fixtures
  - Create `tests/design/common/test_data.py` with test data builders
  - Create `tests/design/common/assertions.py` with custom assertions
  - **Validation**: Import test utilities in a dummy test, verify no errors

## Phase 2: Migrate Design Tests by Domain 🔄

- [ ] **Task 2.1**: Migrate project domain tests to design/
  - Move `automated/test_project_crud.py` → `design/project/test_project_crud.py`
  - Move `automated/test_project_isolation.py` → `design/project/test_project_isolation.py`
  - Update import paths (replace `from config import config` with `from tests.common.config import config`)
  - **Validation**: Run `pytest design/project/ -v`, all tests pass

- [ ] **Task 2.2**: Keep and update existing cluster tests in design/
  - Review `design/test_cluster_graphql.py` (already in correct location)
  - Move to `design/cluster/test_cluster_graphql.py` for consistency
  - Update import paths if needed
  - **Validation**: Run `pytest design/cluster/ -v`, all tests pass

- [ ] **Task 2.3**: Migrate and consolidate model tests in design/
  - Keep `design/test_model_graphql.py` → move to `design/model/test_model_graphql.py`
  - Move `automated/test_model_jsonschema_export.py` → `design/model/test_jsonschema_export.py`
  - Move `automated/test_schema_based_operations.py` → `design/model/test_schema_operations.py`
  - Update import paths
  - **Validation**: Run `pytest design/model/ -v`, all tests pass

- [ ] **Task 2.4**: Migrate compatibility tests to design/
  - Move `automated/test_backward_compatibility.py` → `design/compatibility/test_backward_compatibility.py`
  - Update import paths
  - **Validation**: Run `pytest design/compatibility/ -v`, all tests pass

## Phase 3: Migrate Runtime Tests 🚀

- [ ] **Task 3.1**: Migrate runtime query tests
  - Keep `runtime/test_user_graphql.py` → move to `runtime/query/test_user_graphql.py` for consistency
  - Update imports to use `from tests.design.common import ...` for shared utilities
  - Update imports to use `from tests.common.config import config`
  - **Validation**: Run `pytest runtime/query/ -v`, all tests pass

- [ ] **Task 3.2**: Migrate runtime integration tests
  - Move `automated/test_modelcraft_client_test_graphql.py` → `runtime/integration/test_modelcraft_client.py`
  - Update imports to use Design utilities if needed
  - Document any Design test dependencies in docstring
  - **Validation**: Run `pytest runtime/integration/ -v`, all tests pass

## Phase 4: Extract and Refactor Shared Code 🧹

- [ ] **Task 4.1**: Extract common GraphQL client setup to design/common
  - Review all Design test files for repeated client initialization code
  - Extract to `design/common/graphql_client.py` with `create_design_graphql_client()` function
  - Update Design test files to use shared client
  - **Validation**: Run `pytest design/ -v`, all tests pass

- [ ] **Task 4.2**: Create shared Design fixtures in conftest.py
  - Extract module-scoped fixtures from individual Design test files
  - Add to `design/conftest.py` (e.g., `graphql_client`, `created_projects`)
  - Remove duplicate fixture definitions from test files
  - **Validation**: Run `pytest design/ -v`, fixtures work correctly

- [ ] **Task 4.3**: Create Runtime fixtures that can use Design utilities
  - Add Runtime-specific fixtures to `runtime/conftest.py`
  - Document how Runtime tests can import from `tests.design.common`
  - Add example showing Design tool usage in Runtime test
  - **Validation**: Run `pytest runtime/ -v`, fixtures work correctly

- [ ] **Task 4.4**: Update common utilities imports across all tests
  - Search for all `from config import config` and replace with `from tests.common.config import config`
  - Update any other utility imports to use `tests.common.*`
  - **Validation**: Run `pytest` (all tests), no import errors

## Phase 5: Cleanup Old Structure 🧹

- [ ] **Task 5.1**: Remove automated directory
  - Verify all tests from `automated/` have been migrated
  - Remove `automated/` directory
  - Remove `automated/pytest.ini` (merge into root `pytest.ini`)
  - **Validation**: `automated/` directory no longer exists

- [ ] **Task 5.2**: Remove old utility scripts from root
  - Verify `config.py`, `health_check.py`, `cleanup_test_data.py` moved to `common/`
  - Remove old files from `tests/` root
  - **Validation**: Only migrated files exist in `common/`

- [ ] **Task 5.3**: Verify old design/runtime files are reorganized
  - Ensure `design/test_*.py` files moved to `design/{domain}/`
  - Ensure `runtime/test_*.py` files moved to `runtime/{category}/`
  - Remove any leftover files at the top level
  - **Validation**: Directory structure matches proposal

## Phase 6: Documentation and Configuration 📝

- [ ] **Task 6.1**: Create pytest.ini at tests root
  - Create/update `tests/pytest.ini` with test discovery paths
  - Configure testpaths to include both `design/` and `runtime/`
  - Add markers for design and runtime tests
  - **Validation**: `pytest --collect-only` discovers all tests correctly

- [ ] **Task 6.2**: Create root conftest.py
  - Create `tests/conftest.py` with session-scoped fixtures
  - Add configuration loading fixture
  - Add environment detection fixture
  - **Validation**: Fixtures available to both Design and Runtime tests

- [ ] **Task 6.3**: Update test documentation
  - Update `tests/README.md` with new Design/Runtime structure
  - Update `tests/design/README.md` explaining Design-time tests
  - Create `tests/runtime/README.md` explaining Runtime tests and Design dependencies
  - Add "How to Add New Tests" section with examples
  - **Validation**: Documentation accurately reflects new structure

- [ ] **Task 6.4**: Create migration guide for developers
  - Document the Design vs Runtime distinction
  - Document import path changes (old → new)
  - Provide examples of adding tests to Design and Runtime
  - Document how Runtime can import Design utilities
  - **Validation**: Guide is clear and complete

## Phase 7: CI/CD and Final Verification ✅

- [ ] **Task 7.1**: Update Taskfile/Makefile commands
  - Verify `task auto-test` still works (should run all tests)
  - Optionally add `task test-design` and `task test-runtime` commands
  - Update test paths in automation scripts
  - **Validation**: CI commands work without modification

- [ ] **Task 7.2**: Run full test suite
  - Execute `pytest` (discover all tests)
  - Execute `pytest design/ -v` (Design tests only)
  - Execute `pytest runtime/ -v` (Runtime tests only)
  - Verify test count matches pre-migration (9 test files)
  - **Validation**: All tests pass, no tests missing

- [ ] **Task 7.3**: Test deployment workflows
  - Run `task deploy-local && task auto-test`
  - Run `task deploy-docker && task auto-test ENV=docker`
  - Verify test discovery and execution work correctly
  - **Validation**: All deployment + test workflows work

- [ ] **Task 7.4**: Verify dependency from Runtime to Design
  - Create a simple test in `runtime/` that imports from `design/common/`
  - Verify import works correctly
  - Document this pattern in Runtime README
  - **Validation**: Runtime can successfully use Design utilities

- [ ] **Task 7.5**: Code review and cleanup
  - Remove any unused imports
  - Ensure all `__init__.py` files are present
  - Verify no leftover files in old locations
  - Check for consistent import styles
  - **Validation**: Code is clean and organized

## Dependencies

- **Task 1.3 → Task 2.x**: Common utilities must be moved before updating test imports
- **Task 1.4 → Task 4.x**: Design common infrastructure must exist before extracting shared code
- **Task 2.x → Task 4.x**: All Design tests must be migrated before refactoring shared code
- **Task 3.x → Task 4.3**: Runtime tests must be migrated before creating Runtime fixtures
- **Task 2.x, 3.x → Task 5.x**: All tests migrated before removing old structure
- **Task 5.x → Task 6.x**: Cleanup complete before documenting final structure
- **Task 6.x → Task 7.x**: Documentation complete before final verification

## Rollback Plan

If issues arise during migration:

1. **Git safety**: Use `git mv` for all file moves to preserve history
2. **Incremental commits**: Commit after each phase for easy rollback
3. **Test validation**: Run tests after each domain migration; rollback if failures occur
4. **Backup**: Keep old directory structure until Phase 7 verification complete

## Estimated Effort

- **Phase 1**: 30 minutes (scaffolding)
- **Phase 2**: 1 hour (Design test migration)
- **Phase 3**: 30 minutes (Runtime test migration)
- **Phase 4**: 1-1.5 hours (refactoring shared code)
- **Phase 5**: 15 minutes (cleanup)
- **Phase 6**: 1 hour (documentation)
- **Phase 7**: 30 minutes (final verification)

**Total**: ~4.5-5 hours for complete migration
