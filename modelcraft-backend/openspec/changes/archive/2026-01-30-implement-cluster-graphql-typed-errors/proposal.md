# Proposal: Implement Cluster GraphQL Typed Errors

## Summary

Implement typed error handling for database cluster GraphQL operations following the established pattern from `optimize-project-graphql-errors`. This change applies the same structured error approach to all cluster mutations and queries, enabling clients to programmatically handle specific error scenarios like cluster conflicts, connection failures, and validation errors.

## Motivation

The cluster GraphQL schema (`api/graph/schema/cluster.graphql`) already defines error types but they are not fully implemented:
- `ClusterAlreadyExists`, `ClusterNotFound`, `InvalidClusterInput`, `DatabaseConnectionFailed`
- Error unions: `GetClusterError`, `CreateClusterError`, `UpdateClusterError`, `DeleteClusterError`, `TestConnectionError`
- Payload types with `error` fields

**Important:** Clusters are project-scoped resources, so all cluster operations must validate that the project exists first. The schema needs to include `ProjectNotFound` in all error unions.

However, the current resolver implementation (`internal/interfaces/graphql/cluster.resolvers.go`) uses the old error handling pattern:
- Returns generic GraphQL errors via `bizerrors.WithGraphqlErrorHandler`
- No typed error conversion
- No error adapter layer
- No structured error responses
- Missing `ProjectNotFound` error handling

This creates inconsistency with the project API and fails to leverage the well-defined error schema.

## User Impact

**Before (current state):**
```graphql
mutation {
  createDatabaseCluster(input: {...}) {
    cluster { id name }
    # Error field exists in schema but always returns null
    # Actual errors thrown as generic GraphQL errors
  }
}
```

**After (with typed errors):**
```graphql
mutation {
  createDatabaseCluster(input: {...}) {
    cluster { id name }
    error {
      __typename
      ... on Error {
        message
      }
      ... on ProjectNotFound {
        message  # "Project not found: invalid-project-id"
      }
      ... on ClusterAlreadyExists {
        message
        suggestion
      }
      ... on DatabaseConnectionFailed {
        message
        suggestion
      }
    }
  }
}
```

**Benefits:**
1. **Type-safe error handling**: Clients can distinguish between different error types programmatically
2. **Better error messages**: Structured errors with context-specific suggestions
3. **Consistent API**: Matches the project API error handling pattern
4. **Project-scoped validation**: Proper `ProjectNotFound` errors when project doesn't exist
5. **Backward compatible**: Existing queries continue to work
6. **Testable**: Clear error contracts can be validated in tests

## Success Criteria

1. All cluster mutations return typed errors via the `error` field in their payloads
2. Error adapter converts `bizerrors` codes to appropriate GraphQL error types
3. Unit tests validate error adapter conversions
4. Integration tests (`test_cluster_graphql.py`) validate end-to-end error responses
5. All existing tests continue to pass
6. GraphQL schema validation passes (`task generate-gql`)

## Related Work

- **Depends on:** `optimize-project-graphql-errors` (completed) - establishes the error handling pattern
- **Enables:** Future error handling improvements for Model and Enum operations
- **Related specs:** TBD in `specs/cluster-management/spec.md`

## Implementation Phases

This change is small enough to implement in a single phase:

### Phase 1: Implementation (Single PR)
1. **Update GraphQL schema** - Add `ProjectNotFound` to all cluster error unions
2. Create cluster error adapter following project error adapter pattern
3. Update cluster resolvers to use error adapter
4. Add unit tests for error adapter (including `ProjectNotFound` scenarios)
5. Update integration tests to validate typed errors (including project validation)
6. Validate with `task generate-gql` and `task test-unit`

Schema changes required: Add `ProjectNotFound` to error unions (backward compatible).
