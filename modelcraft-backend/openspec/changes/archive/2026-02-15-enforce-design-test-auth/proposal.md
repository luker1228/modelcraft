# Proposal: Enforce Design Test Authentication and Fix GraphQL Schema Mismatches

## Problem Statement

Currently, Design-time integration tests in `tests/design/` have two critical issues:

### Issue 1: Optional Authentication
Tests allow running without authentication when `CASDOOR_TEST_USERNAME` or `CASDOOR_TEST_PASSWORD` are not configured. The root `conftest.py` returns `None` for the `auth_token` fixture when credentials are missing, printing warnings but allowing tests to continue.

This permissive behavior is problematic because:
1. Design API requires authentication in production environments
2. Tests may pass without auth locally but fail in CI/CD pipelines
3. Inconsistent test behavior based on configuration availability
4. Tests in `tests/design/project/` work correctly because they naturally fail when auth is required, but other tests may not

### Issue 2: GraphQL Schema Parameter Mismatches
Many tests use incorrect GraphQL query parameters that don't match the actual schema:

**Actual Schema (from `api/graph/schema/*.graphql`):**
- `project(name: String!)` - No orgName parameter
- `databaseCluster(projectName: String!, id: ID!)` - No orgName parameter
- `enum(projectName: String!, name: String!)` - No orgName parameter
- `deleteDatabaseCluster(projectName: String!, id: ID!)` - No orgName parameter

**Current Test Queries (INCORRECT):**
- Cluster tests: `databaseCluster(orgName: $orgName, projectName: $projectName, id: $id)`
- Enum tests: `enum(orgName: $orgName, projectName: $projectName, name: $name)`
- Cleanup fixtures: `deleteDatabaseCluster(orgName: $orgName, projectName: $projectName, id: $id)`

This causes tests to fail with GraphQL validation errors or unexpected behavior.

## Proposed Solution

### Solution 1: Enforce Mandatory Authentication

Make authentication **mandatory** for all tests in `tests/design/` by overriding the `auth_token` fixture in `tests/design/conftest.py` to raise an error when credentials are not configured.

**Key Changes:**
1. **Override `auth_token` fixture** in `tests/design/conftest.py`:
   - Check if `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD` are configured
   - Raise `RuntimeError` with clear message if credentials are missing
   - Acquire and return JWT token if credentials are present

2. **Error message**:
   ```
   Authentication required: CASDOOR_TEST_USERNAME and CASDOOR_TEST_PASSWORD must be configured in .env file
   ```

3. **Scope**: Only affects `tests/design/` directory
   - Runtime tests (`tests/runtime/`) continue to work without authentication
   - Runtime uses its own `runtime_graphql_client` fixture without auth dependency

### Solution 2: Fix GraphQL Schema Mismatches

Remove incorrect `orgName` parameters from all test queries and mutations to match the actual GraphQL schema.

**Key Changes:**

1. **Fix Cluster Tests** (`tests/design/cluster/test_cluster_graphql.py`):
   - `GET_CLUSTER`: Remove `orgName` parameter, keep only `projectName` and `id`
   - `LIST_CLUSTERS`: Remove `orgName` parameter, keep only `projectName`
   - `UPDATE_CLUSTER`: Remove `orgName` parameter
   - `DELETE_CLUSTER`: Remove `orgName` parameter
   - Update all test function calls to match

2. **Fix Enum Tests** (`tests/design/enum/test_enum_graphql.py`):
   - `GET_ENUM`: Remove `orgName` parameter, keep only `projectName` and `name`
   - `LIST_ENUMS`: Remove `orgName` parameter, keep only `projectName`
   - `UPDATE_ENUM`: Remove `orgName` parameter (if exists)
   - `DELETE_ENUM`: Remove `orgName` parameter (if exists)
   - Update all test function calls to match

3. **Fix Model Tests** (`tests/design/model/test_model_graphql.py`):
   - Verify all queries match schema (likely already correct)
   - Remove `orgName` if present

4. **Fix Cleanup Fixtures** (`tests/design/conftest.py`):
   - `created_clusters` fixture: Update `DELETE_CLUSTER` mutation to remove `orgName`
   - `created_models` fixture: Verify `DELETE_MODEL` mutation (likely already correct)
   - `created_enums` fixture: Update `DELETE_ENUM` mutation to remove `orgName`
   - Update tuple unpacking in cleanup loops to match stored data structure

5. **Fix Test Data Tracking**:
   - Cluster tracking: Change from `(orgName, projectName, name)` to `(projectName, id)`
   - Enum tracking: Change from `(orgName, projectName, name)` to `(projectName, name)`
   - Update all `created_*.append()` calls in tests

### Why This Approach

**For Authentication (Solution 1):**
- ✅ **Override `auth_token` fixture** (chosen): Clean, leverages existing fixture dependency chain
- ❌ Pytest hooks: More complex, requires marker checking
- ❌ New authenticated client fixture: Requires updating all test signatures
- ❌ Session-scoped validation: Less granular, harder to debug

**Benefits:**
- Minimal code changes (single fixture override)
- Fails fast with clear error message
- Maintains isolation between Design and Runtime tests
- No impact on existing test code that already depends on `auth_token`

**For Schema Fixes (Solution 2):**
- ✅ **Direct parameter removal**: Aligns tests with actual GraphQL schema
- ✅ **Schema as source of truth**: Tests must match the API contract
- ✅ **Fixes permission errors**: Incorrect parameters cause GraphQL validation failures

**Benefits:**
- Tests will execute correctly with proper authentication
- Eliminates GraphQL validation errors
- Makes tests maintainable and accurate
- Aligns with actual API behavior

## Impact Analysis

### Files Modified
1. `tests/design/conftest.py`:
   - Add overridden `auth_token` fixture
   - Fix `created_clusters` fixture DELETE mutation (remove orgName)
   - Fix `created_enums` fixture DELETE mutation (remove orgName)
   - Update tuple unpacking in cleanup loops

2. `tests/design/cluster/test_cluster_graphql.py`:
   - Remove `orgName` from all GraphQL queries/mutations
   - Update all `graphql_client.execute()` calls
   - Fix `created_clusters.append()` calls to use `(projectName, id)`

3. `tests/design/enum/test_enum_graphql.py`:
   - Remove `orgName` from all GraphQL queries/mutations
   - Update all `graphql_client.execute()` calls
   - Fix `created_enums.append()` calls to use `(projectName, name)`

4. `tests/design/model/test_model_graphql.py`:
   - Verify and fix any `orgName` usage (if present)

### Behavior Changes

**Authentication:**
- **Before**: Design tests run without auth (may pass incorrectly)
- **After**: Design tests fail immediately if auth credentials not configured

**GraphQL Queries:**
- **Before**: Tests use incorrect parameters (`orgName`), fail with validation errors
- **After**: Tests use correct parameters matching schema, execute successfully

### Breaking Changes
- **Design tests**: Now require `CASDOOR_TEST_USERNAME` and `CASDOOR_TEST_PASSWORD` in `.env`
- **Test code**: Parameter changes are internal, no user-facing API changes
- **Runtime tests**: No impact (use separate fixture without auth)

## Implementation Plan

See `tasks.md` for detailed implementation steps.

## Validation

### Success Criteria
1. Design tests fail with clear error when credentials missing
2. Design tests pass when credentials are properly configured with owner role
3. Error message clearly indicates missing `.env` configuration
4. All cluster tests execute without GraphQL validation errors
5. All enum tests execute without GraphQL validation errors
6. All model tests execute without GraphQL validation errors
7. Cleanup fixtures successfully delete test resources

### Test Plan
1. Run design tests without credentials → Should fail with clear error
2. Run design tests with credentials → Should pass
3. Verify error message contains expected text
4. Run cluster tests → Should execute and pass (not fail with GraphQL errors)
5. Run enum tests → Should execute and pass
6. Run model tests → Should execute and pass
7. Verify cleanup deletes all test resources correctly

## Related Specifications

- `python-testing-guidelines`: Test configuration and fixture patterns
- `casbin-auth`: Authentication and authorization system
- `casdoor-provider`: Casdoor OAuth2 integration
