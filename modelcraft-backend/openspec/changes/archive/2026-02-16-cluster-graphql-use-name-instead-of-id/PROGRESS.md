# Cluster GraphQL API: Name-Based Migration Progress

## Status: 🔄 In Progress (Core Implementation Complete)

**Last Updated**: 2026-02-15

## Overview

Migration of cluster GraphQL API from ID-based to name-based operations.

**Proposal**: `cluster-graphql-use-name-instead-of-id`

## Completed Tasks ✅

### Phase 1: Schema Updates ✅ (Complete)
- ✅ **Task 1.1**: Updated `cluster.graphql` schema
  - Changed `databaseCluster` query parameter from `id: ID!` to `name: String!`
  - Removed `databaseClusterByName` query (redundant)
  - Updated `updateDatabaseCluster` mutation to use `name: String!`
  - Updated `deleteDatabaseCluster` mutation to use `name: String!`
  - Updated `TestDatabaseConnectionInput` to use `name: String`
  - Commit: `75d6dba`

- ✅ **Task 1.2**: Regenerated GraphQL code
  - Ran `task generate-gql`
  - All resolver interfaces updated
  - Code compiles without errors
  - Commit: `75d6dba`

### Phase 2: Resolver Implementation ✅ (Complete)
- ✅ **Task 2.1**: Updated resolver implementations
  - Modified `DatabaseCluster` query resolver to use `name` parameter
  - Modified `UpdateDatabaseCluster` mutation resolver to use `name`
  - Modified `DeleteDatabaseCluster` mutation resolver to use `name`
  - Modified `TestDatabaseConnection` to support `name` parameter
  - Removed `DatabaseClusterByName` resolver (redundant)
  - All resolvers properly handle name-based lookups
  - Commit: `75d6dba`

### Phase 3: Service Layer Updates ✅ (Complete)
- ✅ **Task 3.1**: Added `GetClusterByName` method
  - Implemented in application service layer
  - Queries by `(orgName, projectName, name)` composite key
  - Proper error handling for not found cases
  - Commit: `75d6dba`

- ✅ **Task 3.2**: Repository layer already supported name-based queries
  - Database has correct index: `idx_cluster_name` on `(org_name, project_name, name)`
  - No additional database migration needed

### Phase 4: Database ✅ (Complete)
- ✅ **Task 4.1**: Verified database indexes
  - Existing `idx_cluster_name` index on `(org_name, project_name, name)` is optimal
  - No new migration needed
  - Schema: `db/schema/mysql/02_database_cluster.sql`

### Phase 5: Test Updates ✅ (Complete)
- ✅ **Task 5.1**: Updated GraphQL integration tests
  - Updated all GraphQL query/mutation definitions to use `name`
  - Fixed `test_get_cluster_not_found` to use name
  - Fixed `test_list_clusters_for_project` tracking
  - Fixed `test_update_cluster_success` to use name
  - Fixed `test_delete_cluster_success` to use name
  - File: `tests/design/cluster/test_cluster_graphql.py`
  - Commit: `43473ce`

- ✅ **Task 5.2**: Updated test fixtures
  - Updated `created_clusters` fixture to track `(projectName, name)` tuples
  - Updated cleanup DELETE_CLUSTER mutation to use name
  - File: `tests/design/conftest.py`
  - Commit: `43473ce`

## Remaining Tasks 📋

### Phase 6: Documentation Updates 📝 (Pending)

#### Task 6.1: Update API Documentation
- [ ] Update API documentation examples
- [ ] Add migration guide for API consumers
- [ ] Document breaking change in changelog
- [ ] Update API reference

**Estimated Time**: 1 hour

#### Task 6.2: Update OpenSpec Documentation
- [ ] Create or update `openspec/specs/cluster-api.md`
- [ ] Document name-based identifier rationale
- [ ] Provide comprehensive examples

**Estimated Time**: 45 minutes

### Phase 7: Cleanup and Finalization 🧹 (Pending)

#### Task 7.1: Remove Unused Code
- [ ] Review and remove any unused ID-based methods
- [ ] Clean up dead code paths
- [ ] Remove obsolete comments

**Estimated Time**: 30 minutes

#### Task 7.2: Code Review and Testing
- [ ] Run full test suite: `task test`
- [ ] Run GraphQL tests: `pytest tests/design/cluster/`
- [ ] Manual testing via GraphQL playground
- [ ] Performance testing
- [ ] Submit PR for review

**Estimated Time**: 2 hours

#### Task 7.3: Create Client Migration Guide
- [ ] Create `docs/migration/cluster-id-to-name.md`
- [ ] Before/after code examples
- [ ] Common pitfalls and solutions
- [ ] FAQ section

**Estimated Time**: 1.5 hours

## Git Commits

| Commit | Message | Date |
|--------|---------|------|
| `43473ce` | test: update cluster tests to use name instead of ID | 2026-02-15 |
| `63c6725` | docs(openspec): add What Changes section to cluster proposal | 2026-02-15 |
| `75d6dba` | feat(cluster): use name instead of ID in cluster GraphQL API | 2026-02-15 |

## Files Modified

### Core Implementation
1. ✅ `api/graph/schema/cluster.graphql` - Schema definition
2. ✅ `api/graph/resolver/cluster.go` - Resolver implementations
3. ✅ `internal/application/cluster_app.go` - Service layer (GetClusterByName)
4. ✅ `api/graph/generated/generated.go` - Auto-generated code

### Tests
5. ✅ `tests/design/cluster/test_cluster_graphql.py` - Integration tests
6. ✅ `tests/design/conftest.py` - Test fixtures

### Documentation
7. ✅ `openspec/changes/cluster-graphql-use-name-instead-of-id/proposal.md`
8. ✅ `openspec/changes/cluster-graphql-use-name-instead-of-id/tasks.md`
9. ✅ `openspec/changes/cluster-graphql-use-name-instead-of-id/PROGRESS.md` (this file)

## Breaking Changes ⚠️

This is a **BREAKING CHANGE** for API consumers:

### Query Changes
```graphql
# Before
query GetCluster($projectName: String!, $id: ID!) {
  databaseCluster(projectName: $projectName, id: $id) {
    cluster { id, name }
  }
}

# After
query GetCluster($projectName: String!, $name: String!) {
  databaseCluster(projectName: $projectName, name: $name) {
    cluster { id, name }
  }
}
```

### Mutation Changes
```graphql
# Before
mutation UpdateCluster($projectName: String!, $id: ID!, $input: UpdateDatabaseClusterInput!) {
  updateDatabaseCluster(projectName: $projectName, id: $id, input: $input) {
    cluster { id, name }
  }
}

# After
mutation UpdateCluster($projectName: String!, $name: String!, $input: UpdateDatabaseClusterInput!) {
  updateDatabaseCluster(projectName: $projectName, name: $name, input: $input) {
    cluster { id, name }
  }
}
```

## Verification Steps

### 1. Code Compilation
```bash
go build ./...
```
✅ Status: PASSED

### 2. Code Generation
```bash
task generate-gql
```
✅ Status: PASSED

### 3. Unit Tests (Pending)
```bash
go test ./internal/application/... -v
go test ./api/graph/resolver/... -v
```
⏳ Status: NOT RUN YET

### 4. Integration Tests (Pending)
```bash
pytest tests/design/cluster/test_cluster_graphql.py -v
```
⏳ Status: NOT RUN YET

## Next Steps

1. **Run Tests**: Execute Python integration tests to verify all changes work correctly
2. **Update Documentation**: Create comprehensive API migration guide
3. **Performance Testing**: Verify name-based queries perform well
4. **Code Review**: Submit PR for team review
5. **Client Communication**: Notify API consumers about breaking changes

## Notes

- Database already has optimal index: `idx_cluster_name` on `(org_name, project_name, name)`
- The `id` field remains in `DatabaseCluster` type for Node interface compliance
- Name uniqueness is enforced at database level with composite unique constraint
- All resolvers now consistently use name-based lookups

## Time Investment

- **Estimated Total**: ~13 hours
- **Time Spent**: ~5 hours
- **Remaining**: ~8 hours (mostly documentation and testing)
