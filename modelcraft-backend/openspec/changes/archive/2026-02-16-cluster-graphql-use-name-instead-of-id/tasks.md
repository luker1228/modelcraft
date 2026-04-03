# Tasks: Use Name Instead of ID in Cluster GraphQL API

## Overview

This document outlines the detailed implementation tasks for migrating cluster GraphQL API from ID-based to name-based identifiers.

## Prerequisites

- [x] Understand current cluster schema and resolver implementation
- [x] Verify name uniqueness is enforced in database
- [x] Confirm name immutability requirement
- [ ] Review and approve proposal
- [ ] Create database migration if index changes needed

## Task Breakdown

### Phase 1: Schema Updates

#### Task 1.1: Update cluster.graphql Schema
**File:** `api/graph/schema/cluster.graphql`

**Changes:**
1. Update Query definitions:
   ```graphql
   # Change from:
   databaseCluster(projectName: String!, id: ID!): GetClusterPayload! @hasPermission(action: "cluster:read")
   
   # Change to:
   databaseCluster(projectName: String!, name: String!): GetClusterPayload! @hasPermission(action: "cluster:read")
   ```

2. Remove redundant query:
   ```graphql
   # Remove this line:
   databaseClusterByName(projectName: String!, name: String!): GetClusterPayload! @hasPermission(action: "cluster:read")
   ```

3. Update Mutation definitions:
   ```graphql
   # Change from:
   updateDatabaseCluster(projectName: String!, id: ID!, input: UpdateDatabaseClusterInput!): UpdateClusterPayload! @hasPermission(action: "cluster:update")
   deleteDatabaseCluster(projectName: String!, id: ID!): DeleteClusterPayload! @hasPermission(action: "cluster:delete")
   
   # Change to:
   updateDatabaseCluster(projectName: String!, name: String!, input: UpdateDatabaseClusterInput!): UpdateClusterPayload! @hasPermission(action: "cluster:update")
   deleteDatabaseCluster(projectName: String!, name: String!): DeleteClusterPayload! @hasPermission(action: "cluster:delete")
   ```

4. Update TestDatabaseConnectionInput:
   ```graphql
   # Change from:
   input TestDatabaseConnectionInput {
     projectName: String
     id: ID
     connectionInfo: DatabaseConnectionInput
   }
   
   # Change to:
   input TestDatabaseConnectionInput {
     projectName: String
     name: String
     connectionInfo: DatabaseConnectionInput
   }
   ```

5. Keep DatabaseCluster type unchanged:
   ```graphql
   type DatabaseCluster implements Node {
     id: ID!  # Keep this
     projectName: String!
     name: String!  # This becomes the primary API identifier
     # ... rest unchanged
   }
   ```

**Acceptance Criteria:**
- [ ] Schema compiles without errors
- [ ] GraphQL introspection shows updated query/mutation signatures
- [ ] `databaseClusterByName` is removed from schema

**Estimated Time:** 30 minutes

---

#### Task 1.2: Regenerate GraphQL Code
**Command:** `make generate-gql` or equivalent code generation command

**Steps:**
1. Run code generation tool (gqlgen)
2. Review generated resolver signatures
3. Verify no compilation errors
4. Commit generated files

**Files Modified:**
- `api/graph/generated/generated.go` (or equivalent generated file)
- `api/graph/model/models_gen.go` (or equivalent)

**Acceptance Criteria:**
- [ ] Code generation completes successfully
- [ ] Generated Go code compiles
- [ ] Resolver interfaces updated with `name string` parameter instead of `id string`

**Estimated Time:** 15 minutes

---

### Phase 2: Resolver Implementation

#### Task 2.1: Update Cluster Query Resolver
**File:** `api/graph/resolver/cluster.go`

**Changes:**

1. Update `DatabaseCluster` resolver method:
   ```go
   // Change signature from:
   func (r *queryResolver) DatabaseCluster(ctx context.Context, projectName string, id string) (*model.GetClusterPayload, error) {
   
   // Change to:
   func (r *queryResolver) DatabaseCluster(ctx context.Context, projectName string, name string) (*model.GetClusterPayload, error) {
       // Update implementation to call service layer with name
       cluster, err := r.clusterService.GetByName(ctx, projectName, name)
       // ... error handling
   }
   ```

2. Remove `DatabaseClusterByName` resolver method:
   ```go
   // Delete this method entirely:
   // func (r *queryResolver) DatabaseClusterByName(ctx context.Context, projectName string, name string) (*model.GetClusterPayload, error)
   ```

**Acceptance Criteria:**
- [ ] Resolver method signature matches generated interface
- [ ] Implementation calls correct service method
- [ ] Error handling is appropriate (ClusterNotFound, etc.)
- [ ] Authorization checks remain in place

**Estimated Time:** 30 minutes

---

#### Task 2.2: Update Cluster Mutation Resolvers
**File:** `api/graph/resolver/cluster.go`

**Changes:**

1. Update `UpdateDatabaseCluster` resolver:
   ```go
   // Change signature from:
   func (r *mutationResolver) UpdateDatabaseCluster(ctx context.Context, projectName string, id string, input model.UpdateDatabaseClusterInput) (*model.UpdateClusterPayload, error) {
   
   // Change to:
   func (r *mutationResolver) UpdateDatabaseCluster(ctx context.Context, projectName string, name string, input model.UpdateDatabaseClusterInput) (*model.UpdateClusterPayload, error) {
       // Update implementation to call service layer with name
       cluster, err := r.clusterService.UpdateByName(ctx, projectName, name, input)
       // ... error handling
   }
   ```

2. Update `DeleteDatabaseCluster` resolver:
   ```go
   // Change signature from:
   func (r *mutationResolver) DeleteDatabaseCluster(ctx context.Context, projectName string, id string) (*model.DeleteClusterPayload, error) {
   
   // Change to:
   func (r *mutationResolver) DeleteDatabaseCluster(ctx context.Context, projectName string, name string) (*model.DeleteClusterPayload, error) {
       // Update implementation to call service layer with name
       err := r.clusterService.DeleteByName(ctx, projectName, name)
       // ... error handling
   }
   ```

3. Update `TestDatabaseConnection` resolver (if it uses cluster lookup):
   ```go
   func (r *mutationResolver) TestDatabaseConnection(ctx context.Context, input model.TestDatabaseConnectionInput) (*model.TestConnectionPayload, error) {
       // Change logic from:
       // if input.ID != nil { cluster = getByID(...) }
       
       // To:
       // if input.Name != nil { cluster = getByName(...) }
   }
   ```

**Acceptance Criteria:**
- [ ] All resolver signatures match generated interfaces
- [ ] Service layer calls updated to use name
- [ ] Error handling preserved
- [ ] Authorization directives still apply

**Estimated Time:** 45 minutes

---

### Phase 3: Service Layer Updates

#### Task 3.1: Add/Update Service Methods for Name-Based Lookup
**File:** `internal/service/cluster.go` (or equivalent)

**Changes:**

1. Update or add `GetByName` method:
   ```go
   func (s *ClusterService) GetByName(ctx context.Context, projectName, name string) (*Cluster, error) {
       // Query database: WHERE project_name = ? AND name = ?
       // Return cluster or ClusterNotFound error
   }
   ```

2. Update or add `UpdateByName` method:
   ```go
   func (s *ClusterService) UpdateByName(ctx context.Context, projectName, name string, input UpdateClusterInput) (*Cluster, error) {
       // Find cluster by name
       // Validate input
       // Update cluster
       // Return updated cluster
   }
   ```

3. Update or add `DeleteByName` method:
   ```go
   func (s *ClusterService) DeleteByName(ctx context.Context, projectName, name string) error {
       // Find cluster by name
       // Soft delete or hard delete
       // Return error if not found
   }
   ```

4. If these methods already exist (due to `databaseClusterByName` query), refactor ID-based methods to use them internally

**Acceptance Criteria:**
- [ ] Methods efficiently query by (projectName, name)
- [ ] Proper error types returned (ClusterNotFound, etc.)
- [ ] Transactions handled correctly
- [ ] Business logic validation maintained

**Estimated Time:** 1 hour

---

#### Task 3.2: Update Repository/DAO Layer
**File:** `internal/repository/cluster.go` or `internal/dao/cluster.go`

**Changes:**

1. Ensure `FindByName` query exists:
   ```go
   func (r *ClusterRepository) FindByName(ctx context.Context, projectName, name string) (*ClusterEntity, error) {
       // SQL: SELECT * FROM clusters WHERE project_name = ? AND name = ? AND deleted_at IS NULL
   }
   ```

2. Verify database index exists:
   ```sql
   -- Ensure this index exists (or add migration):
   CREATE INDEX idx_clusters_project_name_name ON clusters(project_name, name);
   ```

3. Update delete/update queries if they currently use ID:
   ```go
   func (r *ClusterRepository) DeleteByName(ctx context.Context, projectName, name string) error {
       // SQL: UPDATE clusters SET deleted_at = NOW() WHERE project_name = ? AND name = ?
   }
   ```

**Acceptance Criteria:**
- [ ] Database queries use name-based lookups
- [ ] Indexes are in place for performance
- [ ] Transactions handled correctly
- [ ] No SQL injection vulnerabilities

**Estimated Time:** 1 hour

---

### Phase 4: Database Migration (if needed)

#### Task 4.1: Create Database Index Migration
**File:** Create new migration file (e.g., `migrations/000XXX_add_cluster_name_index.sql`)

**Migration Up:**
```sql
-- Ensure composite index on (project_name, name) for efficient lookups
CREATE INDEX IF NOT EXISTS idx_clusters_project_name_name 
ON clusters(project_name, name) 
WHERE deleted_at IS NULL;

-- Optional: Analyze query performance
ANALYZE clusters;
```

**Migration Down:**
```sql
DROP INDEX IF EXISTS idx_clusters_project_name_name;
```

**Acceptance Criteria:**
- [ ] Migration runs successfully in dev environment
- [ ] Index improves query performance (verify with EXPLAIN)
- [ ] Migration is idempotent
- [ ] Rollback works correctly

**Estimated Time:** 30 minutes

---

### Phase 5: Test Updates

#### Task 5.1: Update GraphQL Integration Tests
**File:** `tests/design/cluster/test_cluster_graphql.py`

**Changes:**

1. Update GraphQL query definitions:
   ```python
   # Change from:
   GET_CLUSTER = """
   query GetCluster($projectName: String!, $id: ID!) {
     databaseCluster(projectName: $projectName, id: $id) {
       cluster { id, name, title }
       error { __typename message }
     }
   }
   """
   
   # Change to:
   GET_CLUSTER = """
   query GetCluster($projectName: String!, $name: String!) {
     databaseCluster(projectName: $projectName, name: $name) {
       cluster { id, name, title }
       error { __typename message }
     }
   }
   """
   ```

2. Update mutation definitions:
   ```python
   # Change UPDATE_CLUSTER mutation:
   # $id: ID! → $name: String!
   # updateDatabaseCluster(projectName: $projectName, id: $id, ...) 
   # → updateDatabaseCluster(projectName: $projectName, name: $name, ...)
   
   # Change DELETE_CLUSTER mutation similarly
   ```

3. Update all test function calls:
   ```python
   # Change from:
   result = graphql_client.execute(GET_CLUSTER, {
       "projectName": project_name,
       "id": cluster_id
   })
   
   # Change to:
   result = graphql_client.execute(GET_CLUSTER, {
       "projectName": project_name,
       "name": cluster_name
   })
   ```

4. Update test data tracking:
   ```python
   # Change from:
   created_clusters.append((project_name, cluster_id))
   
   # Change to:
   created_clusters.append((project_name, cluster_name))
   ```

5. Remove any tests specifically for `databaseClusterByName` query (or merge into main test)

**Acceptance Criteria:**
- [ ] All cluster GraphQL tests pass
- [ ] Tests cover name-based operations
- [ ] Tests verify error cases (cluster not found with invalid name)
- [ ] Test cleanup works correctly

**Estimated Time:** 1.5 hours

---

#### Task 5.2: Update Test Fixtures
**File:** `tests/design/conftest.py`

**Changes:**

1. Update `created_clusters` fixture cleanup:
   ```python
   # Change from:
   DELETE_CLUSTER = """
   mutation DeleteCluster($projectName: String!, $id: ID!) {
     deleteDatabaseCluster(projectName: $projectName, id: $id) {
       success
       error { __typename message }
     }
   }
   """
   
   # Change to:
   DELETE_CLUSTER = """
   mutation DeleteCluster($projectName: String!, $name: String!) {
     deleteDatabaseCluster(projectName: $projectName, name: $name) {
       success
       error { __typename message }
     }
   }
   """
   ```

2. Update cleanup loop:
   ```python
   # Change from:
   for project_name, cluster_id in created_clusters:
       result = client.execute(DELETE_CLUSTER, {
           "projectName": project_name,
           "id": cluster_id
       })
   
   # Change to:
   for project_name, cluster_name in created_clusters:
       result = client.execute(DELETE_CLUSTER, {
           "projectName": project_name,
           "name": cluster_name
       })
   ```

**Acceptance Criteria:**
- [ ] Fixture cleanup works with name-based deletion
- [ ] No resource leaks in test environment
- [ ] Cleanup errors are properly logged

**Estimated Time:** 30 minutes

---

#### Task 5.3: Add New Test Cases
**File:** `tests/design/cluster/test_cluster_graphql.py`

**New Tests:**

1. Test get cluster by name:
   ```python
   def test_get_cluster_by_name_success(graphql_client, auth_token, test_project, created_clusters):
       # Create cluster
       # Retrieve by name
       # Assert correct cluster returned
   ```

2. Test update cluster by name:
   ```python
   def test_update_cluster_by_name(graphql_client, auth_token, test_project, created_clusters):
       # Create cluster
       # Update via name
       # Verify update successful
   ```

3. Test delete cluster by name:
   ```python
   def test_delete_cluster_by_name(graphql_client, auth_token, test_project):
       # Create cluster
       # Delete via name
       # Verify deletion successful
   ```

4. Test error case - cluster not found:
   ```python
   def test_get_cluster_by_invalid_name(graphql_client, auth_token, test_project):
       # Query with non-existent name
       # Assert ClusterNotFound error returned
   ```

5. Test name uniqueness enforcement:
   ```python
   def test_create_duplicate_cluster_name(graphql_client, auth_token, test_project, created_clusters):
       # Create cluster with name "test-cluster"
       # Attempt to create another with same name
       # Assert ClusterAlreadyExists error
   ```

**Acceptance Criteria:**
- [ ] All new tests pass
- [ ] Tests cover happy path and error cases
- [ ] Tests verify name uniqueness
- [ ] Test coverage maintained or improved

**Estimated Time:** 1 hour

---

### Phase 6: Documentation Updates

#### Task 6.1: Update API Documentation
**Files:** 
- API documentation (if separate from schema)
- Example queries in docs
- README or developer guide

**Changes:**
1. Update all example queries to use `name` instead of `id`
2. Add migration guide for API consumers
3. Document breaking change in changelog
4. Update API reference documentation

**Acceptance Criteria:**
- [ ] All documentation examples use name-based API
- [ ] Migration guide is clear and complete
- [ ] Breaking change is prominently documented

**Estimated Time:** 1 hour

---

#### Task 6.2: Update OpenSpec Documentation
**Files:**
- `openspec/specs/cluster-api.md` (create if doesn't exist)
- `openspec/project.md` (update if needed)

**Changes:**
1. Document new API contract
2. Explain name-based identifier rationale
3. Provide examples
4. Link to migration guide

**Acceptance Criteria:**
- [ ] Spec clearly describes name-based API
- [ ] Examples are accurate and helpful
- [ ] Rationale is well-explained

**Estimated Time:** 45 minutes

---

### Phase 7: Cleanup and Finalization

#### Task 7.1: Remove Unused Code
**Files:** Various

**Changes:**
1. Remove any ID-based service methods if no longer needed
2. Remove `DatabaseClusterByName` resolver implementation
3. Clean up any dead code paths
4. Remove obsolete comments

**Acceptance Criteria:**
- [ ] No unused functions remain
- [ ] Code is clean and maintainable
- [ ] No compilation warnings

**Estimated Time:** 30 minutes

---

#### Task 7.2: Code Review and Testing
**Steps:**
1. Self-review all changes
2. Run full test suite: `make test`
3. Run GraphQL tests: `pytest tests/design/cluster/`
4. Manual testing via GraphQL playground
5. Performance testing for name-based queries
6. Submit PR for team review

**Acceptance Criteria:**
- [ ] All tests pass (unit, integration, GraphQL)
- [ ] Code review feedback addressed
- [ ] No performance regressions
- [ ] CI/CD pipeline passes

**Estimated Time:** 2 hours

---

#### Task 7.3: Create Migration Guide for Clients
**File:** Create `docs/migration/cluster-id-to-name.md`

**Content:**
1. Overview of change
2. Step-by-step migration instructions
3. Before/after code examples
4. Common pitfalls and solutions
5. FAQ section
6. Support contact information

**Acceptance Criteria:**
- [ ] Guide is comprehensive and clear
- [ ] Examples cover common use cases
- [ ] Published to appropriate documentation site

**Estimated Time:** 1.5 hours

---

## Summary

### Total Estimated Time
- Phase 1 (Schema): 45 minutes
- Phase 2 (Resolvers): 1 hour 15 minutes
- Phase 3 (Service Layer): 2 hours
- Phase 4 (Database): 30 minutes
- Phase 5 (Tests): 3 hours
- Phase 6 (Documentation): 1 hour 45 minutes
- Phase 7 (Cleanup): 4 hours

**Total: ~13 hours** (approximately 2 work days)

### Dependencies
- Phase 2 depends on Phase 1
- Phase 3 can be done in parallel with Phase 2
- Phase 4 should be done before Phase 7.2
- Phase 5 depends on Phases 2 & 3
- Phase 6 can be done in parallel with Phase 5
- Phase 7 depends on all previous phases

### Risk Mitigation
- **Risk:** Breaking existing clients
  - **Mitigation:** Provide deprecation period with parallel support (optional)
  - **Mitigation:** Clear communication and migration guide

- **Risk:** Performance degradation
  - **Mitigation:** Ensure proper database indexing
  - **Mitigation:** Performance testing before merge

- **Risk:** Data inconsistency
  - **Mitigation:** Verify name uniqueness constraints
  - **Mitigation:** Thorough integration testing

### Definition of Done
- [ ] All code changes implemented and reviewed
- [ ] All tests passing (unit, integration, GraphQL)
- [ ] Database migrations applied successfully
- [ ] Documentation updated and published
- [ ] Migration guide created and distributed
- [ ] Performance validated
- [ ] PR approved and merged
- [ ] Changes deployed to staging environment
- [ ] Smoke tests passed in staging
- [ ] Breaking change communicated to stakeholders
