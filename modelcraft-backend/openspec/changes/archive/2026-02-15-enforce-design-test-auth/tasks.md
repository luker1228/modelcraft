# Implementation Tasks

## Task Status Overview

| # | Task | Status | Priority | Effort | Dependencies |
|---|------|--------|----------|--------|--------------|
| 1 | Override auth_token fixture | ✅ DONE | High | 15 min | None |
| 2 | Fix Cluster test queries | ✅ DONE | High | 20 min | None |
| 3 | Fix Enum test queries | ✅ DONE | High | 20 min | None |
| 4 | Fix cleanup fixtures | ✅ DONE | High | 15 min | Task 2, 3 |
| 5 | Verify Model tests | ⚠️ NEEDS_WORK | Medium | 15 min | None |
| 6 | Test without credentials | ⏳ TODO | High | 5 min | Task 1 |
| 7 | Test full suite | ⏳ TODO | High | 10 min | Task 1-5 |
| 8 | Test individual modules | ⏳ TODO | Medium | 15 min | Task 1-5 |
| 9 | Update documentation | ⏳ TODO | Low | 10 min | Task 7 |

**Status Legend:**
- ⏳ TODO - Not started
- 🚧 IN PROGRESS - Currently working on
- ✅ DONE - Completed
- ⚠️ NEEDS_WORK - Issues found, needs attention
- ❌ BLOCKED - Waiting on dependencies

**Total Progress:** 4/9 tasks completed (44%)

---

## Task List

### Task 1: Override `auth_token` fixture in Design conftest
**Priority**: High
**Estimated effort**: 15 minutes
**Dependencies**: None

Add overridden `auth_token` fixture to `tests/design/conftest.py` that:
- Checks if `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD` are configured
- Raises `RuntimeError` with clear message if missing
- Acquires JWT token from Casdoor if credentials present
- Returns JWT token for use in `graphql_client` fixture

**Files to modify:**
- `tests/design/conftest.py`

**Implementation details:**
- Use `@pytest.fixture(scope="session")` to match root fixture scope
- Import `get_test_access_token` from `common.auth`
- Error message: "Authentication required: CASDOOR_TEST_USERNAME and CASDOOR_TEST_PASSWORD must be configured in .env file"

**Validation:**
- Run design tests without credentials → Should fail with clear error
- Run design tests with credentials → Should pass fixture setup
- Verify error message matches expected format

---

### Task 2: Fix Cluster test GraphQL queries
**Priority**: High
**Estimated effort**: 20 minutes
**Dependencies**: None

Remove `orgName` parameter from all cluster-related GraphQL queries and mutations in `tests/design/cluster/test_cluster_graphql.py`.

**Queries/Mutations to fix:**
1. `GET_CLUSTER`: Change from `databaseCluster(orgName: $orgName, projectName: $projectName, id: $id)`
   → `databaseCluster(projectName: $projectName, id: $id)`

2. `LIST_CLUSTERS`: Change from `databaseClusters(orgName: $orgName, projectName: $projectName, input: $input)`
   → `databaseClusters(projectName: $projectName, input: $input)`

3. `UPDATE_CLUSTER`: Change from `updateDatabaseCluster(orgName: $orgName, projectName: $projectName, id: $id, input: $input)`
   → `updateDatabaseCluster(projectName: $projectName, id: $id, input: $input)`

4. `DELETE_CLUSTER`: Change from `deleteDatabaseCluster(orgName: $orgName, projectName: $projectName, id: $id)`
   → `deleteDatabaseCluster(projectName: $projectName, id: $id)`

**Test function calls to update:**
- Remove `"orgName": "built-in"` from all `variable_values` dicts
- Keep only `"projectName": "default"` and other required parameters

**Tracking tuples to fix:**
- Change `created_clusters.append((cluster["orgName"], cluster["projectName"], cluster["name"]))`
- To `created_clusters.append((cluster["projectName"], cluster["id"]))`

**Validation:**
- Run cluster tests → Should execute without GraphQL validation errors
- Verify clusters are created and retrieved successfully

---

### Task 3: Fix Enum test GraphQL queries
**Priority**: High
**Estimated effort**: 20 minutes
**Dependencies**: None

Remove `orgName` parameter from all enum-related GraphQL queries and mutations in `tests/design/enum/test_enum_graphql.py`.

**Queries/Mutations to fix:**
1. `GET_ENUM`: Change from `enum(orgName: $orgName, projectName: $projectName, name: $name)`
   → `enum(projectName: $projectName, name: $name)`

2. `LIST_ENUMS`: Change from `enums(orgName: $orgName, projectName: $projectName)`
   → `enums(projectName: $projectName)`

3. `UPDATE_ENUM` (if exists): Remove `orgName` parameter

4. `DELETE_ENUM` (if exists): Remove `orgName` parameter

**Test function calls to update:**
- Remove `"orgName": "built-in"` from all `variable_values` dicts
- Keep only `"projectName": "default"` and other required parameters

**Tracking tuples to fix:**
- Change `created_enums.append((enum["orgName"], enum["projectName"], enum["name"]))`
- To `created_enums.append((enum["projectName"], enum["name"]))`

**Validation:**
- Run enum tests → Should execute without GraphQL validation errors
- Verify enums are created and retrieved successfully

---

### Task 4: Fix cleanup fixtures in conftest
**Priority**: High
**Estimated effort**: 15 minutes
**Dependencies**: Task 2, Task 3

Update cleanup fixtures in `tests/design/conftest.py` to match the corrected GraphQL schema and tracking tuples.

**Fixtures to fix:**

1. **`created_clusters` fixture:**
   - Update docstring: "Clusters are stored as tuples of `(projectName, id)`"
   - Fix DELETE mutation: Remove `orgName` parameter
   - Update loop: `for project_name, cluster_id in clusters:`
   - Update execute call: Remove `"orgName"` from variable_values

2. **`created_enums` fixture:**
   - Update docstring: "Enums are stored as tuples of `(projectName, name)`"
   - Fix DELETE mutation: Remove `orgName` parameter (if present)
   - Update loop: `for project_name, enum_name in enums:`
   - Update execute call: Remove `"orgName"` from variable_values

3. **Verify `created_models` fixture:**
   - Check if it uses `orgName` - if so, remove it
   - Likely already correct (uses `projectName` and `id`)

**Validation:**
- Run tests and check cleanup logs
- Verify resources are successfully deleted after tests
- No "Failed to cleanup" warnings for schema mismatch

---

### Task 5: Verify and fix Model tests
**Priority**: Medium
**Estimated effort**: 15 minutes
**Dependencies**: None

Verify `tests/design/model/test_model_graphql.py` uses correct GraphQL schema parameters.

**Check queries:**
- `GET_MODEL`: Should be `model(projectName: $projectName, id: $id)` or `modelByName(...)`
- `LIST_MODELS`: Should be `models(input: $input)` with projectName in ModelQueryInput
- `UPDATE_MODEL`: Should be `updateModel(projectName: $projectName, id: $id, input: $input)`
- `DELETE_MODEL`: Should be `deleteModel(projectName: $projectName, id: $id)`

**If orgName is present:**
- Remove from all queries/mutations
- Remove from all `variable_values` dicts
- Update tracking tuples if needed

**Validation:**
- Run model tests → Should pass without schema errors
- Verify model CRUD operations work correctly

---

### Task 6: Test without credentials
**Priority**: High
**Estimated effort**: 5 minutes
**Dependencies**: Task 1

Verify that design tests fail correctly when credentials are missing.

**Test steps:**
1. Comment out `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD` in `.env`
2. Run: `pytest tests/design/project/test_project_crud.py -v`
3. Verify test session fails immediately with expected error message
4. Verify error contains: "Authentication required: CASDOOR_TEST_USERNAME and CASDOOR_TEST_PASSWORD must be configured in .env file"

**Expected outcome:**
- Tests fail during fixture setup (not during test execution)
- Clear error message printed to console
- No tests are executed

---

### Task 7: Test with credentials - Full suite
**Priority**: High
**Estimated effort**: 10 minutes
**Dependencies**: Task 1, Task 2, Task 3, Task 4, Task 5

Verify all design tests pass correctly when credentials are configured.

**Test steps:**
1. Restore `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD` in `.env`
2. Run: `pytest tests/design/ -v`
3. Verify JWT token is acquired successfully
4. Verify all test modules execute and pass:
   - `tests/design/project/test_project_crud.py`
   - `tests/design/model/test_model_graphql.py`
   - `tests/design/model/test_model_errors.py`
   - `tests/design/cluster/test_cluster_graphql.py`
   - `tests/design/enum/test_enum_graphql.py`

**Expected outcome:**
- Token acquisition succeeds
- Console shows: "✅ Obtained test access token for Design tests (length=...)"
- All tests execute without GraphQL validation errors
- Tests pass (or fail for legitimate reasons, not schema errors)
- Cleanup fixtures successfully delete test resources

---

### Task 8: Test individual modules
**Priority**: Medium
**Estimated effort**: 15 minutes
**Dependencies**: Task 1, Task 2, Task 3, Task 4, Task 5

Test each module individually to verify fixes.

**Test commands:**
```bash
# Test cluster operations
pytest tests/design/cluster/test_cluster_graphql.py::TestClusterCRUD::test_create_cluster_success -v

# Test enum operations
pytest tests/design/enum/test_enum_graphql.py -v

# Test model operations
pytest tests/design/model/test_model_graphql.py -v

# Test project operations
pytest tests/design/project/test_project_crud.py -v
```

**Expected outcome:**
- Each module executes successfully
- No GraphQL schema validation errors
- Resources created and cleaned up properly
- Permissions are correctly enforced (owner role has access)

---

### Task 9: Update documentation (if needed)
**Priority**: Low
**Estimated effort**: 10 minutes
**Dependencies**: Task 7

Check if any documentation needs updating to reflect mandatory auth requirement.

**Files to check:**
- `tests/design/README.md` (if exists)
- `CLAUDE.md` testing section
- Root `README.md` testing instructions

**Update if needed:**
- Add note that Design tests require Casdoor credentials
- Document required `.env` variables
- Provide example `.env` configuration

---

## Task Summary

| Task | Priority | Dependencies | Validation |
|------|----------|--------------|------------|
| 1. Override auth_token fixture | High | None | Manual test |
| 2. Fix Cluster test queries | High | None | Run cluster tests |
| 3. Fix Enum test queries | High | None | Run enum tests |
| 4. Fix cleanup fixtures | High | Task 2, 3 | Check cleanup logs |
| 5. Verify Model tests | Medium | None | Run model tests |
| 6. Test without credentials | High | Task 1 | Fails with clear error |
| 7. Test with credentials - Full | High | Task 1-5 | All tests pass |
| 8. Test individual modules | Medium | Task 1-5 | Each module passes |
| 9. Update documentation | Low | Task 7 | Docs accurate |

## Estimated Total Effort
125 minutes (~2 hours)

## Parallelization Opportunities
- Task 2, 3, 5 can be done in parallel (schema fixes)
- Task 1 can be done independently
- Task 4 depends on Task 2 and 3 completing
- Task 6 depends only on Task 1
- Task 7 requires all fixes to complete
- Task 8 can validate during or after Task 7

## Implementation Order (Recommended)

**Phase 1: Auth Fix (15 min)**
1. Task 1: Override auth_token fixture
2. Task 6: Test without credentials

**Phase 2: Schema Fixes (55 min in parallel)**
- Task 2: Fix Cluster tests
- Task 3: Fix Enum tests
- Task 5: Verify Model tests

**Phase 3: Cleanup (15 min)**
4. Task 4: Fix cleanup fixtures

**Phase 4: Integration Testing (25 min)**
5. Task 7: Full test suite
6. Task 8: Individual module tests

**Phase 5: Documentation (10 min)**
7. Task 9: Update documentation
