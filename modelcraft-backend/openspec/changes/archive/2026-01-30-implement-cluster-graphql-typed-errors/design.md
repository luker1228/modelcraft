# Design: Cluster GraphQL Typed Errors

## Architecture Overview

This change applies the typed error pattern from `optimize-project-graphql-errors` to cluster operations. No new architectural patterns are introduced.

## Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│  GraphQL Layer (Resolvers)                                  │
│  internal/interfaces/graphql/cluster.resolvers.go           │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Query/Mutation Resolvers                          │    │
│  │  - DatabaseCluster()                               │    │
│  │  - CreateDatabaseCluster()                         │    │
│  │  - UpdateDatabaseCluster()                         │    │
│  │  - DeleteDatabaseCluster()                         │    │
│  │  - TestDatabaseConnection()                        │    │
│  └────────────────────────────────────────────────────┘    │
│                         │                                    │
│                         │ calls                              │
│                         ▼                                    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  ClusterErrorAdapter                               │    │
│  │  (NEW)                                             │    │
│  │  internal/interfaces/graphql/adapter/              │    │
│  │  cluster_error_adapter.go                          │    │
│  │                                                     │    │
│  │  Methods:                                          │    │
│  │  - ConvertToGetClusterError()                      │    │
│  │  - ConvertToCreateClusterError()                   │    │
│  │  - ConvertToUpdateClusterError()                   │    │
│  │  - ConvertToDeleteClusterError()                   │    │
│  │  - ConvertToTestConnectionError()                  │    │
│  └────────────────────────────────────────────────────┘    │
│                         │                                    │
│                         │ converts                           │
│                         ▼                                    │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ from bizerrors
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Domain Layer                                               │
│  pkg/bizerrors/                                             │
│                                                              │
│  Error Definitions:                                         │
│  - ClusterAlreadyExists (CONFLICT.CLUSTER)                  │
│  - ClusterNotFound (NOT_FOUND.CLUSTER)                      │
│  - DatabaseConnectionFailed (OPERATION_FAILED.DB_CONNECTION)│
│  - ParamInvalid (PARAM_INVALID)                             │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ to GraphQL types
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  GraphQL Schema                                             │
│  api/graph/schema/cluster.graphql                           │
│                                                              │
│  Error Types (already defined):                             │
│  - ClusterAlreadyExists implements Error                    │
│  - ClusterNotFound implements Error                         │
│  - InvalidClusterInput implements Error                     │
│  - DatabaseConnectionFailed implements Error                │
│                                                              │
│  Error Unions:                                              │
│  - GetClusterError                                          │
│  - CreateClusterError                                       │
│  - UpdateClusterError                                       │
│  - DeleteClusterError                                       │
│  - TestConnectionError                                      │
└─────────────────────────────────────────────────────────────┘
```

## Error Flow

### Success Flow
```
Client Request
    ↓
GraphQL Resolver (e.g., CreateDatabaseCluster)
    ↓
Application Service (DatabaseClusterService.CreateCluster)
    ↓ (success)
Return Payload { cluster: {...}, error: nil }
    ↓
Client receives cluster data
```

### Error Flow
```
Client Request
    ↓
GraphQL Resolver (e.g., CreateDatabaseCluster)
    ↓
Application Service (DatabaseClusterService.CreateCluster)
    ↓ (error: bizerrors.ClusterAlreadyExists)
ClusterErrorAdapter.ConvertToCreateClusterError()
    ↓
Map to GraphQL ClusterAlreadyExists type
    ↓
Return Payload { cluster: nil, error: ClusterAlreadyExists{message, suggestion} }
    ↓
Client receives typed error
```

## Error Mapping Table

| bizerrors Code | GraphQL Error Type | Operations | Suggestion |
|----------------|-------------------|------------|------------|
| `CONFLICT.CLUSTER` | `ClusterAlreadyExists` | Create | "Please use a different cluster name within this project" |
| `NOT_FOUND.CLUSTER` | `ClusterNotFound` | Get, Update, Delete, TestConnection | None |
| `NOT_FOUND.PROJECT` | `ProjectNotFound` | All operations | None |
| `PARAM_INVALID` | `InvalidClusterInput` | Create, Update | Extracted from error detail (e.g., "Cluster name must be alphanumeric") |
| `OPERATION_FAILED.DB_CONNECTION` | `DatabaseConnectionFailed` | Create, Update, TestConnection | "Please verify host, port, username, and password are correct" |

## Key Design Decisions

### Decision 1: Reuse Existing Schema
**Choice:** Use the already-defined error types in `cluster.graphql`, no schema changes needed.

**Rationale:**
- Schema already defines all necessary error types and unions
- Zero schema migration risk
- No breaking changes for clients
- Only implementation layer changes required

### Decision 2: Follow Project Error Adapter Pattern
**Choice:** Create `ClusterErrorAdapter` following the exact pattern from `ProjectErrorAdapter`.

**Rationale:**
- Consistency across the codebase
- Proven pattern that works well
- Easy to understand and maintain
- Testable in isolation
- Clear separation of concerns

**Code Pattern:**
```go
type ClusterErrorAdapter struct {
    logger logfacade.Logger
}

func (a *ClusterErrorAdapter) ConvertToCreateClusterError(err *bizerrors.BusinessError) generated.CreateClusterError {
    if err == nil {
        return nil
    }

    switch err.Info().GetCode() {
    case bizerrors.ClusterAlreadyExists.GetCode():
        suggestion := "Please use a different cluster name within this project"
        return &generated.ClusterAlreadyExists{
            Message:    err.Msg(),
            Suggestion: &suggestion,
        }
    case bizerrors.ParamInvalid.GetCode():
        // Map to InvalidClusterInput
    case bizerrors.DatabaseConnectionFailed.GetCode():
        // Map to DatabaseConnectionFailed
    case bizerrors.ProjectNotFound.GetCode():
        // Map to ProjectNotFound (project validation)
    default:
        a.logger.Errorf("Unknown error code for CreateCluster: %s", err.Info().GetCode())
        return &generated.InvalidClusterInput{Message: err.Msg()}
    }
}
```

### Decision 3: Handle Unknown Error Codes Gracefully
**Choice:** Log warning and return a safe default error type rather than panic.

**Rationale:**
- Production resilience - unknown errors don't crash the server
- Observability - logged warnings help identify missing mappings
- Client safety - always returns a valid error type
- Follows pattern from ProjectErrorAdapter

**Default error types by operation:**
- Create/Update: Return `InvalidClusterInput` (safe generic validation error)
- Get/Delete: Return `ClusterNotFound` (safe generic not found error)
- TestConnection: Return `DatabaseConnectionFailed` (safe generic connection error)

### Decision 4: Add Context-Aware Suggestions
**Choice:** Provide helpful suggestions for common error scenarios.

**Rationale:**
- Better developer experience
- Reduces support burden
- Follows GraphQL best practices
- Already established pattern in project errors

**Examples:**
- `ClusterAlreadyExists`: "Please use a different cluster name within this project"
- `InvalidClusterInput`: Extract from error detail (e.g., "Cluster name must contain only alphanumeric characters and hyphens")
- `DatabaseConnectionFailed`: "Please verify host, port, username, and password are correct"

### Decision 5: Maintain Backward Compatibility
**Choice:** Keep existing resolver behavior, only add typed error handling.

**Rationale:**
- No breaking changes for existing clients
- Payloads already have `error` field defined in schema
- Resolvers currently return generic errors via `bizerrors.WithGraphqlErrorHandler`
- New implementation populates `error` field instead of throwing

**Migration:**
```go
// OLD (current)
func (r *mutationResolver) CreateDatabaseCluster(...) (*generated.DatabaseCluster, error) {
    graphqlErr := bizerrors.WithGraphqlErrorHandler(ctx, func() error {
        // ... business logic
        return err // Generic GraphQL error thrown
    })
    return payload, graphqlErr
}

// NEW (with typed errors)
func (r *mutationResolver) CreateDatabaseCluster(...) (*generated.CreateClusterPayload, error) {
    adapter := adapter.NewClusterErrorAdapter()

    // Call business logic
    result, bizErr := r.DatabaseClusterService.CreateCluster(...)

    if bizErr != nil {
        return &generated.CreateClusterPayload{
            Cluster: nil,
            Error:   adapter.ConvertToCreateClusterError(bizErr),
        }, nil // No GraphQL error thrown, error in payload
    }

    return &generated.CreateClusterPayload{
        Cluster: result,
        Error:   nil,
    }, nil
}
```

## Testing Strategy

### Unit Tests (`cluster_error_adapter_test.go`)
- Test each error mapping function
- Test with all error codes in the mapping table
- Test nil error handling
- Test unknown error codes
- Test message and suggestion population
- Test that logger is called for unknown codes

### Integration Tests (`test_cluster_graphql.py`)
- Test queries returning typed errors
- Test mutations returning typed errors
- Test error `__typename` is correct
- Test error `message` field is populated
- Test error `suggestion` field when present
- Test backward compatibility - existing queries work
- Test success cases - empty error field

## Risks & Mitigations

### Risk 1: Schema Already Defines Payload Types
**Risk:** Current schema may define `DatabaseCluster` return type instead of `CreateClusterPayload`.

**Likelihood:** HIGH (based on current resolver signatures)

**Impact:** Medium - requires schema change

**Mitigation:**
- Review schema carefully during implementation
- If needed, update schema to use payload types (backward compatible - adds fields)
- Verify with `task generate-gql`
- Check `generated.go` for payload types

### Risk 2: Missing Error Codes in bizerrors
**Risk:** Some error scenarios may not have corresponding bizerrors definitions.

**Likelihood:** LOW (bizerrors already has cluster error codes)

**Impact:** Low - can add new codes if needed

**Mitigation:**
- Review `pkg/bizerrors/common_errors.go` to confirm all codes exist
- Add missing codes if identified during implementation

### Risk 3: Integration Test Failures
**Risk:** Existing tests may not expect typed errors.

**Likelihood:** MEDIUM

**Impact:** Medium - requires test updates

**Mitigation:**
- Review `test_cluster_graphql.py` before implementation
- Update test assertions to check for typed errors
- Ensure backward compatibility in test queries

## Open Questions

### Q1: Should we return payload types or direct types?
**Current observation:** Resolvers return `*generated.DatabaseCluster` directly, not `*generated.CreateClusterPayload`.

**Decision needed:** During implementation, check if schema needs updating to use payload types.

**Resolution approach:** Check schema and generated types, update if needed.

### Q2: Are there additional cluster error scenarios not covered?
**Current mapping:** Based on existing bizerrors codes.

**Decision needed:** During implementation, check application service for other error returns.

**Resolution approach:** Review `internal/app/cluster/cluster_app.go` for all error scenarios.

### Q3: Should suggestions be internationalized?
**Current approach:** English suggestions following project error pattern.

**Decision needed:** Future enhancement, not in scope for this change.

**Resolution approach:** Out of scope - suggestions are developer-facing (GraphQL API).
