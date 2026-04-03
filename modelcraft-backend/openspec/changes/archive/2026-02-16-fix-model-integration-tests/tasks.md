# Tasks: Fix Model Integration Tests

## Phase 1: Fix GraphQL Query Definitions

- [ ] **Fix duplicate parameters in test_model_errors.py queries**
  - [ ] Fix `CREATE_CLUSTER` - remove duplicate `projectName` in selection (line 36-37)
  - [ ] Fix `CREATE_MODEL` - remove duplicate `projectName` in selection (line 52-53)
  - [ ] Fix `GET_MODEL` - remove duplicate `projectName` parameter (line 75) and selection (line 79-80)
  - [ ] Fix `GET_MODEL_BY_NAME` - remove duplicate `projectName` parameter (line 100)
  - [ ] Fix `UPDATE_MODEL` - remove duplicate `projectName` parameter and selection
  - [ ] Fix `DELETE_MODEL` - remove duplicate `projectName` parameter
  - Validation: GraphQL syntax is valid (no duplicate parameters)

- [ ] **Fix duplicate parameters in test_model_graphql.py queries**
  - [ ] Fix `GET_MODEL` - ensure single `projectName` parameter (line 85)
  - [ ] Fix `LIST_MODELS` - ensure single `projectName` parameter
  - [ ] Fix `DELETE_MODEL` - ensure single `projectName` parameter
  - Validation: All queries use correct parameter names

## Phase 2: Fix Test Fixtures

- [ ] **Fix test_cluster fixture in test_model_errors.py**
  - [ ] Add null check for cluster creation result (line 217-220)
  - [ ] Fix cleanup tracking - use `(projectName, name)` tuple (line 223)
  - [ ] Add error handling for cluster creation failures
  - Validation: Cluster is successfully created and tracked correctly

- [ ] **Fix test_model fixture in test_model_errors.py**
  - [ ] Add null check for model creation result (line 241-242)
  - [ ] Fix tracking tuple - remove duplicate projectName (line 243)
  - [ ] Change from `(projectName, projectName, name)` to `(projectName, id)`
  - [ ] Add error handling for model creation failures
  - [ ] Return correct values matching fixture contract
  - Validation: Model is created and tracked with correct identifiers

- [ ] **Fix default_project fixture usage**
  - [ ] Update fixture return value to include `projectName` field
  - [ ] Or update all test usages to use `name` instead of `projectName`
  - [ ] Ensure consistency across all test files
  - Validation: No KeyError on accessing default_project fields

## Phase 3: Fix Cleanup Logic

- [ ] **Fix cluster cleanup in conftest.py**
  - [ ] Verify `DELETE_CLUSTER` mutation signature (line 118-124)
  - [ ] Ensure cleanup uses `name` parameter, not `id`
  - [ ] Update fixture tracking to match cleanup requirements
  - Validation: Clusters are successfully deleted without errors

- [ ] **Fix model cleanup in conftest.py**
  - [ ] Verify `DELETE_MODEL` mutation signature (line 157-163)
  - [ ] Ensure cleanup uses correct identifier (id or name)
  - [ ] Match tracking tuple structure to cleanup needs
  - Validation: Models are successfully deleted without errors

## Phase 4: Fix Test Logic

- [ ] **Fix test parameter passing in test_model_errors.py**
  - [ ] Remove duplicate `projectName` from all test query executions
  - [ ] Update line 251-253 (test_get_model_not_found_error)
  - [ ] Update line 266-271 (test_get_model_by_name_not_found)
  - [ ] Update all UPDATE_MODEL test calls
  - [ ] Update all DELETE_MODEL test calls
  - Validation: All query executions have correct parameters

- [ ] **Fix test parameter passing in test_model_graphql.py**
  - [ ] Update line 296 (test_get_model_not_found) - fix default_project access
  - [ ] Update line 316 (test_list_models_empty_result) - fix default_project access
  - [ ] Ensure all queries use correct parameter names
  - Validation: No KeyError or missing parameter errors

## Phase 5: Validation and Testing

- [ ] **Run all model tests**
  - [ ] Execute: `pytest tests/design/model/ -v`
  - [ ] Verify all 23 tests pass
  - [ ] Verify no fixture errors (TypeError, KeyError)
  - [ ] Verify no cluster reference errors
  - Validation: Test output shows "23 passed"

- [ ] **Verify cleanup functionality**
  - [ ] Run tests and check cleanup messages
  - [ ] Verify no "Failed to cleanup" warnings
  - [ ] Verify database is clean after test runs
  - Validation: All resources cleaned up successfully

- [ ] **Run tests multiple times**
  - [ ] Execute tests 3 times consecutively
  - [ ] Verify consistent pass results
  - [ ] Verify no resource conflicts
  - Validation: Tests are idempotent and repeatable

## Implementation Notes

- **Test files to modify:**
  - `tests/design/model/test_model_errors.py` (primary fixes)
  - `tests/design/model/test_model_graphql.py` (parameter fixes)
  - `tests/design/conftest.py` (cleanup fixes if needed)

- **Key principles:**
  - GraphQL parameters must be unique within each query/mutation
  - Fixture tracking tuples must match cleanup mutation requirements
  - Always validate resource creation before using resources
  - Add null checks for all external API responses

- **Testing strategy:**
  - Fix GraphQL syntax first (enables queries to run)
  - Fix fixtures second (enables resource creation)
  - Fix cleanup third (enables test repeatability)
  - Validate end-to-end last (ensures complete functionality)
