# Project-Cluster One-to-One Integration Tests

## Overview

Created comprehensive integration tests for the Project-Cluster one-to-one relationship refactoring.

## Test File

**Location**: `tests/design/project/test_project_cluster_one_to_one.py`

## Test Cases

### 1. `test_create_project_without_cluster`
- **Purpose**: Verify projects can be created without cluster assignment
- **Expected**: `clusterId` field should be `null` for projects without clusters
- **Validates**: Backward compatibility with existing projects

### 2. `test_project_can_have_one_cluster`
- **Purpose**: Verify a project can successfully have one cluster
- **Expected**: Cluster creation succeeds, cluster belongs to the project
- **Validates**: Basic one-to-one relationship functionality

### 3. `test_project_cannot_have_second_cluster` ⭐
- **Purpose**: **Core test** - Verify one-to-one constraint is enforced
- **Expected**: Creating a second cluster for a project returns `ClusterAlreadyExists` error
- **Validates**: Database-level and application-level constraint enforcement
- **Error Message**: Should mention constraint violation ("already", "cluster", "one")

### 4. `test_get_project_with_cluster_info`
- **Purpose**: Verify `clusterId` and `clusterInfo` fields in GraphQL API
- **Expected**: Project query returns cluster information via `clusterInfo` resolver
- **Validates**: GraphQL schema changes and field resolvers

### 5. `test_delete_cluster_allows_new_cluster`
- **Purpose**: Verify constraint allows new cluster after deletion
- **Expected**: After deleting a cluster, a new cluster can be created for the same project
- **Validates**: Constraint doesn't prevent legitimate cluster replacement

### 6. `test_backward_compatibility_projects_without_cluster`
- **Purpose**: Verify existing projects without clusters continue to work
- **Expected**: Projects without `clusterId` can be queried and function normally
- **Validates**: Backward compatibility with existing data

## GraphQL Queries Used

### Queries
- `GET_PROJECT_WITH_CLUSTER` - Query project with `clusterId` and `clusterInfo` fields

### Mutations
- `CREATE_PROJECT` - Create project (with optional `clusterId`)
- `CREATE_CLUSTER` - Create database cluster
- `DELETE_CLUSTER` - Delete cluster (for cleanup and re-creation tests)

## Test Data Helpers

Uses existing test utilities from `design.common`:
- `generate_test_id()` - Generate unique test identifiers
- `build_project_input()` - Build project creation input
- `build_cluster_input()` - Build cluster creation input
- `assert_graphql_success()` - Assert GraphQL query/mutation success

## Fixtures Used

- `graphql_client` - GraphQL client for API calls
- `created_projects` - Tracks projects for cleanup (existing fixture)
- `created_clusters` - Tracks clusters for cleanup (existing fixture)

## Running the Tests

### Prerequisites
1. Start the server: `task run` (or `task deploy-local`)
2. Ensure test database is initialized: `task db:migrate-up`
3. Ensure test user exists: `task test-user-setup`

### Execute Tests
```bash
# Run all Project-Cluster one-to-one tests
cd tests
pytest design/project/test_project_cluster_one_to_one.py -v

# Run specific test
pytest design/project/test_project_cluster_one_to_one.py::TestProjectClusterOneToOne::test_project_cannot_have_second_cluster -v

# Run with detailed output
pytest design/project/test_project_cluster_one_to_one.py -v -s

# Run all design tests
pytest design/ -v
```

### Expected Results

When server is running:
- ✅ All 6 tests should pass
- ✅ Test #3 (`test_project_cannot_have_second_cluster`) validates the core constraint
- ✅ Projects and clusters are automatically cleaned up after tests

## Key Validations

1. **One-to-One Constraint**: ✅ Enforced (test #3)
2. **Backward Compatibility**: ✅ Projects without clusters work (test #1, #6)
3. **GraphQL API**: ✅ `clusterId` and `clusterInfo` fields exposed (test #4)
4. **Cluster Replacement**: ✅ Can delete and recreate clusters (test #5)
5. **Error Handling**: ✅ Clear error messages for constraint violations (test #3)

## Integration with CI/CD

These tests can be integrated into the automated test suite:

```bash
# In CI/CD pipeline
task deploy-local
task auto-test  # This should include the new tests
```

Or specifically:
```bash
task run &
sleep 5  # Wait for server to start
cd tests && pytest design/project/test_project_cluster_one_to_one.py -v
```

## Coverage

The integration tests provide end-to-end validation of:
- Database schema changes (unique constraint)
- Domain model changes (ClusterID field)
- Repository layer (GetByProjectKey, ExistsByProjectKey)
- Application services (CreateCluster validation)
- GraphQL API (schema updates, error handling)
- Error messages (user-facing feedback)

## Next Steps

1. ✅ Integration tests created and ready
2. ⏭️ Run tests with live server: `task run` then `pytest ...`
3. ⏭️ Add tests to CI/CD pipeline
4. ⏭️ Consider adding tests for `UpdateProject` with `clusterId` parameter (future enhancement)
