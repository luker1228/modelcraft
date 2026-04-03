# Tasks: Implement Cluster GraphQL Typed Errors

## Overview

Small, focused change to implement typed error handling for cluster GraphQL operations. All tasks are sequential and build on each other.

## Task List

### 1. Create cluster error adapter (Go)
**Deliverable:** `internal/interfaces/graphql/adapter/cluster_error_adapter.go`
- [x] Create `ClusterErrorAdapter` struct with logger
- [x] Implement `ConvertToGetClusterError()` - maps `ClusterNotFound`, `ProjectNotFound`
- [x] Implement `ConvertToCreateClusterError()` - maps `ClusterAlreadyExists`, `InvalidClusterInput`, `DatabaseConnectionFailed`, `ProjectNotFound`
- [x] Implement `ConvertToUpdateClusterError()` - maps `ClusterNotFound`, `InvalidClusterInput`, `DatabaseConnectionFailed`, `ProjectNotFound`
- [x] Implement `ConvertToDeleteClusterError()` - maps `ClusterNotFound`, `ProjectNotFound`
- [x] Implement `ConvertToTestConnectionError()` - maps `ClusterNotFound`, `DatabaseConnectionFailed`, `ProjectNotFound`
- [x] Follow pattern from `project_error_adapter.go`
- [x] Map `bizerrors` codes: `ClusterAlreadyExists`, `ClusterNotFound`, `ProjectNotFound`, `ParamInvalid`, `DatabaseConnectionFailed`
- [x] Add helpful suggestions for validation errors
- [x] **IMPORTANT:** `ProjectNotFound` should be checked first in all operations (project validation precedes other checks)

**Validation:** Code compiles, follows existing adapter pattern

**Dependencies:** None

---

### 2. Add unit tests for error adapter
**Deliverable:** `internal/interfaces/graphql/adapter/cluster_error_adapter_test.go`
- [x] Test `ConvertToGetClusterError()` with `ClusterNotFound` error
- [x] Test `ConvertToGetClusterError()` with `ProjectNotFound` error
- [x] Test `ConvertToCreateClusterError()` with all error types (including `ProjectNotFound`)
- [x] Test `ConvertToUpdateClusterError()` with all error types (including `ProjectNotFound`)
- [x] Test `ConvertToDeleteClusterError()` with `ClusterNotFound` and `ProjectNotFound` errors
- [x] Test `ConvertToTestConnectionError()` with connection failures and `ProjectNotFound`
- [x] Test nil error handling (should return nil)
- [x] Test unknown error codes (should log warning and return safe default)
- [x] Verify message and suggestion fields are populated correctly
- [x] **Test error priority:** Verify `ProjectNotFound` is handled correctly in all converters

**Validation:** `go test ./internal/interfaces/graphql/adapter/...` passes

**Dependencies:** Task 1

---

### 3. Update cluster query resolvers to use error adapter
**Deliverable:** Updated `internal/interfaces/graphql/cluster.resolvers.go` (queries)
- [x] Update `DatabaseCluster()` query resolver
  - Convert errors using error adapter
  - Populate `error` field in `GetClusterPayload`
  - Return `nil` cluster data on error
- [x] Update `DatabaseClusters()` query resolver (if needed for error handling)
- [x] Update `DatabaseClusterByName()` query resolver
  - Convert errors using error adapter
  - Populate `error` field in `GetClusterPayload`
  - Return `nil` cluster data on error
- [x] Remove old generic error handling pattern

**Validation:** Code compiles, queries return typed errors

**Dependencies:** Task 1, 2

---

### 4. Update cluster mutation resolvers to use error adapter
**Deliverable:** Updated `internal/interfaces/graphql/cluster.resolvers.go` (mutations)
- [x] Update `CreateDatabaseCluster()` mutation
  - Convert errors using error adapter
  - Populate `error` field in `CreateClusterPayload`
  - Return `nil` cluster data on error
- [x] Update `UpdateDatabaseCluster()` mutation
  - Convert errors using error adapter
  - Populate `error` field in `UpdateClusterPayload`
  - Return `nil` cluster data on error
- [x] Update `DeleteDatabaseCluster()` mutation
  - Convert errors using error adapter
  - Populate `error` field in `DeleteClusterPayload`
  - Set `success: false` on error
- [x] Update `TestDatabaseConnection()` mutation
  - Convert errors using error adapter
  - Populate `error` field in `TestConnectionPayload`
  - Set `success: false` on error
- [x] Remove old generic error handling pattern

**Validation:** Code compiles, mutations return typed errors

**Dependencies:** Task 3

---

### 5. Run GraphQL code generation
**Deliverable:** Regenerated GraphQL code
- [x] Run `task generate-gql` to regenerate GraphQL code
- [x] Verify no schema changes needed (types already defined)
- [x] Verify generated code compiles
- [x] Review generated types for error unions

**Validation:** `task generate-gql` succeeds, code compiles

**Dependencies:** Task 4

---

### 6. Update Python integration tests for typed errors
**Deliverable:** Updated `tests/design/test_cluster_graphql.py`
- [x] Update test queries/mutations to request `error` field with `__typename`
- [x] Update `test_database_cluster_query_by_id_not_found()` - verify `ClusterNotFound` error type
- [x] Update `test_delete_database_cluster_not_found()` - verify typed error response
- [x] Update `test_database_cluster_connection_test_not_found()` - verify typed error response
- [x] Add new test: `test_create_cluster_duplicate_name()` - verify `ClusterAlreadyExists` error
- [x] Add new test: `test_create_cluster_invalid_input()` - verify `InvalidClusterInput` error
- [x] Add new test: `test_connection_test_failure()` - verify `DatabaseConnectionFailed` error (if testable)
- [x] **Add new test: `test_create_cluster_nonexistent_project()`** - verify `ProjectNotFound` error
- [x] **Add new test: `test_update_cluster_nonexistent_project()`** - verify `ProjectNotFound` error
- [x] **Add new test: `test_delete_cluster_nonexistent_project()`** - verify `ProjectNotFound` error
- [x] Update error assertions to check `error.__typename` and `error.message`
- [x] Ensure backward compatibility - existing test queries still work

**Validation:** `pytest tests/design/test_cluster_graphql.py -v` passes

**Dependencies:** Task 5

---

### 7. Run full test suite
**Deliverable:** All tests passing
- [x] Run `task test-unit` - all Go unit tests pass
- [x] Run `pytest tests/design/test_cluster_graphql.py -v` - all integration tests pass
- [x] Run `task fmt` - code formatting correct
- [x] Run `task lint` - no linter errors
- [x] Run `task vet` - no vet errors

**Validation:** All checks pass

**Dependencies:** Task 6

---

## Task Summary

- **Total tasks:** 7
- **Estimated complexity:** Small (follows established pattern)
- **Parallel work:** None - all tasks are sequential
- **Critical path:** Tasks 1-7 must be done in order
- **Testing strategy:** Unit tests for adapter, integration tests for resolvers

## Notes

- Follow TDD: Write tests before implementation where possible
- Use `project_error_adapter.go` as reference implementation
- Maintain backward compatibility - don't break existing queries
- Add logging for unknown error codes
- Include helpful suggestions in validation errors
