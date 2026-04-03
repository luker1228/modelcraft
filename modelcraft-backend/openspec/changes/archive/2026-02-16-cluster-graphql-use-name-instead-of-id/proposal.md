# Proposal: Use Name Instead of ID in Cluster GraphQL API

## Why

This change improves developer experience by using human-readable cluster names as the primary identifier in GraphQL operations, replacing opaque IDs. Names are more intuitive, align with how developers think about resources, and simplify workflows in CLI tools, Infrastructure-as-Code, and documentation. This addresses the fundamental usability issue where developers must perform a two-step process (lookup ID, then operate) when names are the natural way to reference clusters.

## Problem Statement

Currently, the GraphQL API for database cluster operations uses `id: ID!` as the primary identifier for queries and mutations. This approach has several drawbacks:

### Issue 1: Poor User Experience
When developers interact with the API (via GraphQL playground, CLI, or SDK), they need to:
1. First query to get the cluster ID (which is often a UUID or numeric value)
2. Then use that ID in subsequent operations

This two-step process is cumbersome, especially when developers naturally think of clusters by their human-readable names (e.g., "prod-mysql", "dev-postgres").

### Issue 2: Inconsistent API Design
The API already has dual identifiers:
- **Current Primary**: `databaseCluster(projectName: String!, id: ID!)`
- **Alternative Query**: `databaseClusterByName(projectName: String!, name: String!)`

Having two separate query endpoints for the same resource suggests an inconsistency in the API design philosophy. The `name` is already unique within a project context and serves as a natural identifier.

### Issue 3: API Usability Gap
Real-world scenarios demonstrate the preference for name-based operations:
- Infrastructure-as-Code (IaC): Config files use cluster names, not IDs
- CLI commands: `modelcraft cluster get prod-mysql` is more intuitive than `modelcraft cluster get abc-123-def`
- Documentation: Examples reference clusters by name for readability
- Debugging: Logs and error messages reference cluster names, requiring ID lookup is an extra cognitive step

### Issue 4: Current Schema Structure
**Existing Queries:**
```graphql
databaseCluster(projectName: String!, id: ID!): GetClusterPayload!
databaseClusterByName(projectName: String!, name: String!): GetClusterPayload!
databaseClusters(projectName: String!, input: DatabaseClusterQueryInput): DatabaseClusterConnection!
```

**Existing Mutations:**
```graphql
createDatabaseCluster(input: CreateDatabaseClusterInput!): CreateClusterPayload!
updateDatabaseCluster(projectName: String!, id: ID!, input: UpdateDatabaseClusterInput!): UpdateClusterPayload!
deleteDatabaseCluster(projectName: String!, id: ID!): DeleteClusterPayload!
testDatabaseConnection(input: TestDatabaseConnectionInput!): TestConnectionPayload!
```

**Observation:**
- All mutations (update, delete) require `id: ID!`
- Query operations have both ID-based and name-based variants
- The `name` field is already unique and immutable within a project

## Proposed Solution

### Primary Change: Replace ID with Name in Mutations

Modify all cluster mutations and queries to use `name: String!` as the primary identifier instead of `id: ID!`.

## What Changes

This proposal introduces breaking changes to the cluster GraphQL API by replacing ID-based operations with name-based operations:

**Schema Changes:**
- Query `databaseCluster` parameter changes from `id: ID!` to `name: String!`
- Query `databaseClusterByName` is removed (redundant after change)
- Mutation `updateDatabaseCluster` parameter changes from `id: ID!` to `name: String!`
- Mutation `deleteDatabaseCluster` parameter changes from `id: ID!` to `name: String!`
- Input type `TestDatabaseConnectionInput` field changes from `id: ID` to `name: String`

**Implementation Changes:**
- GraphQL resolvers updated to accept and use `name` parameter
- Application service layer adds `GetClusterByName` method
- Repository lookups change from ID-based to name-based (with orgName and projectName context)
- All tests updated to use name-based operations

**Breaking Changes:**
- This is a **BREAKING CHANGE** - all API clients must migrate from ID-based to name-based operations
- Clients storing cluster IDs should migrate to storing cluster names
- The `id` field remains in the `DatabaseCluster` type for Node interface compliance and internal use

**Key Changes:**

1. **Update Query Signatures:**
   ```graphql
   # Before
   databaseCluster(projectName: String!, id: ID!): GetClusterPayload!
   
   # After
   databaseCluster(projectName: String!, name: String!): GetClusterPayload!
   ```

2. **Update Mutation Signatures:**
   ```graphql
   # Before
   updateDatabaseCluster(projectName: String!, id: ID!, input: UpdateDatabaseClusterInput!): UpdateClusterPayload!
   deleteDatabaseCluster(projectName: String!, id: ID!): DeleteClusterPayload!
   
   # After
   updateDatabaseCluster(projectName: String!, name: String!, input: UpdateDatabaseClusterInput!): UpdateClusterPayload!
   deleteDatabaseCluster(projectName: String!, name: String!): DeleteClusterPayload!
   ```

3. **Deprecate Redundant Query:**
   - Remove `databaseClusterByName` as it becomes redundant
   - The primary `databaseCluster` query now uses name

4. **Update Input Types:**
   ```graphql
   # TestDatabaseConnectionInput
   input TestDatabaseConnectionInput {
     projectName: String
     name: String # Changed from: id: ID
     connectionInfo: DatabaseConnectionInput
   }
   
   # ListDatabasesInput
   input ListDatabasesInput {
     projectName: String!
     name: String! # Changed from: name: String! (already uses name, keep as is)
     offset: Int
     limit: Int
     search: String
   }
   ```

5. **Keep ID Field in DatabaseCluster Type:**
   ```graphql
   type DatabaseCluster implements Node {
     id: ID!  # Keep for Node interface compliance and internal references
     projectName: String!
     name: String!  # Primary identifier for API operations
     title: String!
     # ... other fields
   }
   ```

### Why This Approach

**Advantages:**
- ✅ **Better UX**: Developers can reference clusters by memorable names
- ✅ **Consistent API**: Single, predictable pattern for all operations
- ✅ **Natural Semantics**: Aligns with how developers conceptualize resources
- ✅ **Simplified Queries**: No need to look up IDs before operations
- ✅ **IaC Friendly**: Config files can directly reference cluster names
- ✅ **Backward Compatible for ID**: Keep ID field in type for internal use and Node interface

**Trade-offs:**
- Name uniqueness must be enforced at the database level (already implemented)
- Name immutability must be maintained (or name changes require careful handling)
- Internal resolvers need to query by name instead of ID (straightforward change)

**Alternative Approaches Considered:**
- ❌ **Keep both ID and Name**: Maintains API bloat and inconsistency
- ❌ **Use composite keys**: Overly complex for simple resource identification
- ❌ **Add name parameter alongside ID**: Doesn't solve the primary UX issue

## Impact Analysis

### Files Modified

1. **GraphQL Schema** (`api/graph/schema/cluster.graphql`):
   - Update all query signatures to use `name: String!`
   - Update all mutation signatures to use `name: String!`
   - Remove `databaseClusterByName` query
   - Update input types

2. **Resolver Implementations**:
   - `api/graph/resolver/cluster.go`: Update resolver methods to accept and query by name
   - Internal service calls: Modify to fetch by name instead of ID

3. **Service Layer**:
   - `internal/service/cluster.go`: Add or update methods for name-based lookups
   - Database queries: Change from `WHERE id = ?` to `WHERE name = ? AND project_name = ?`

4. **Test Files**:
   - `tests/design/cluster/test_cluster_graphql.py`: Update all GraphQL queries/mutations to use `name`
   - `tests/design/conftest.py`: Update cleanup fixtures to use name
   - Update test data tracking from ID to name

5. **Generated Code**:
   - Run `make generate-gql` to regenerate resolver stubs and models

### Behavior Changes

**Before:**
```graphql
query {
  databaseCluster(projectName: "my-project", id: "123-abc-456") {
    cluster { name, title }
  }
}

mutation {
  updateDatabaseCluster(
    projectName: "my-project",
    id: "123-abc-456",
    input: { title: "New Title" }
  ) {
    cluster { id, name }
  }
}
```

**After:**
```graphql
query {
  databaseCluster(projectName: "my-project", name: "prod-mysql") {
    cluster { id, name, title }
  }
}

mutation {
  updateDatabaseCluster(
    projectName: "my-project",
    name: "prod-mysql",
    input: { title: "New Title" }
  ) {
    cluster { id, name }
  }
}
```

### Breaking Changes

**This is a BREAKING CHANGE:**
- ⚠️ All existing API clients using ID-based operations will need to update
- ⚠️ Clients must switch from passing `id` to passing `name` parameters
- ⚠️ The `databaseClusterByName` query will be removed

**Migration Path:**
1. Clients should update their queries/mutations to use `name` instead of `id`
2. If clients store cluster IDs, they should migrate to storing cluster names
3. Temporary backward compatibility could be provided via deprecated fields (optional)

**Version Compatibility:**
- Suggest a deprecation period with both APIs available (optional grace period)
- Clear documentation and migration guide
- API version bump (if versioning is in place)

## Implementation Plan

See `tasks.md` for detailed implementation steps.

## Validation

### Success Criteria

1. **Schema Changes:**
   - All cluster queries accept `name: String!` instead of `id: ID!`
   - All cluster mutations accept `name: String!` instead of `id: ID!`
   - `databaseClusterByName` query is removed
   - GraphQL schema validation passes

2. **Resolver Implementation:**
   - Resolvers correctly look up clusters by name within project scope
   - Proper error handling for non-existent cluster names
   - Performance is acceptable (ensure indexed lookups on name + projectName)

3. **Test Coverage:**
   - All existing cluster tests updated and passing
   - Tests verify lookup by name works correctly
   - Tests verify error cases (cluster not found, duplicate names)
   - Integration tests confirm end-to-end functionality

4. **Code Generation:**
   - `make generate-gql` completes without errors
   - Generated resolver signatures match new schema
   - No compilation errors in Go code

### Test Plan

1. **Unit Tests:**
   - Test resolver methods with valid cluster names
   - Test resolver methods with invalid/non-existent names
   - Test name uniqueness enforcement

2. **Integration Tests:**
   - Create cluster and retrieve by name
   - Update cluster using name identifier
   - Delete cluster using name identifier
   - Verify error responses for missing clusters
   - Test concurrent operations on same cluster name

3. **GraphQL Tests:**
   - Update `tests/design/cluster/test_cluster_graphql.py`:
     - Change all `id` variables to `name`
     - Update query/mutation parameters
     - Verify responses
   - Update cleanup fixtures in `conftest.py`

4. **Performance Tests:**
   - Benchmark name-based lookups vs ID-based lookups
   - Ensure database indexes on (name, project_name) exist
   - Verify query performance is acceptable

5. **Backward Compatibility Tests** (if grace period implemented):
   - Verify deprecated ID-based queries still work with warnings
   - Verify new name-based queries work correctly
   - Confirm deprecation notices appear in GraphQL introspection

### Performance Considerations

- Ensure composite index on `(project_name, name)` exists in database
- Name-based lookups should be as fast as ID-based lookups with proper indexing
- Consider caching strategies if needed (likely not necessary for typical workloads)

## Related Specifications

- `cluster-api-spec`: Database cluster API design and architecture
- `graphql-conventions`: GraphQL API design guidelines and best practices
- `python-testing-guidelines`: Test configuration and fixture patterns
- `database-schema`: Database table definitions and indexes

## Migration Guide

For API consumers who need to migrate from ID-based to name-based operations:

### Step 1: Identify Current Usage
```graphql
# Old code
query GetCluster($id: ID!) {
  databaseCluster(projectName: "my-project", id: $id) {
    cluster { name, title }
  }
}
```

### Step 2: Update to Name-Based
```graphql
# New code
query GetCluster($name: String!) {
  databaseCluster(projectName: "my-project", name: $name) {
    cluster { id, name, title }
  }
}
```

### Step 3: Update Client Code
- Replace all `id` variables with `name` variables
- If storing cluster references, store `name` instead of `id`
- Update any internal mappings or caches

### Step 4: Test Thoroughly
- Verify all cluster operations work with name-based identifiers
- Test error handling for non-existent cluster names
- Ensure uniqueness constraints are respected

## Notes

- The `id` field remains in the `DatabaseCluster` type for Node interface compliance
- Internal systems can still use IDs for optimization if needed
- This change improves the API contract and user experience significantly
- Consider similar changes for other resources (models, enums) in future proposals
