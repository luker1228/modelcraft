# Proposal: Add Enum GraphQL Typed Errors

## Change ID
`add-enum-graphql-typed-errors`

## Status
Draft

## Summary

Implement typed error handling for enum GraphQL operations following the established pattern from `optimize-project-graphql-errors` and `implement-cluster-graphql-typed-errors`. This change applies structured error handling to all enum mutations and queries, enabling clients to programmatically handle specific error scenarios like enum conflicts, validation errors, and reference constraints.

## Motivation

### Current State

The enum GraphQL schema (`api/graph/schema/enum.graphql`) currently lacks typed error handling:
- Mutations return simple types (`EnumDefinition!`, `Boolean!`) that throw generic GraphQL errors on failure
- No structured error responses for clients
- Cannot distinguish between "enum not found", "enum already exists", "invalid input", or "cannot delete referenced enum"
- Inconsistent with project and cluster APIs which already have typed errors

Current resolver implementation (`internal/interfaces/graphql/enum.resolvers.go`):
- Uses generic error handling via `bizerrors.WithGraphqlErrorHandler`
- Returns `nil` or generic errors to clients
- No error adapter layer
- No typed error conversion

### Business Need

Enum operations need typed error handling for:

1. **Conflict Detection** - Client needs to know if enum name already exists in the project
2. **Validation Feedback** - Invalid enum options, name format, or code values need specific error messages
3. **Reference Safety** - Cannot delete enums that are referenced by model fields; clients need clear error with field names
4. **Project Validation** - Enum operations must validate project exists (project-scoped resource)
5. **API Consistency** - Match the error handling pattern used by project and cluster APIs

### Why This Change

- **Type Safety**: Clients can handle specific error types programmatically
- **Better UX**: Provide context-specific error messages and suggestions
- **Consistent API**: All GraphQL resources (Project, Cluster, Enum) follow the same pattern
- **Backward Compatible**: Existing queries continue to work; error field is additive

## Goals

1. **Implement Typed Errors** - Add error types and unions to enum.graphql schema
2. **Create Error Adapter** - Convert bizerrors to GraphQL error types following established pattern
3. **Update Resolvers** - Modify enum resolvers to return typed errors via payload types
4. **Maintain Compatibility** - Ensure existing clients continue to work without changes
5. **Comprehensive Testing** - Unit tests for error adapter, integration tests for end-to-end validation

## Non-Goals

1. **Change Business Logic** - Error conditions remain the same, only representation changes
2. **Add New Error Types** - Use existing bizerrors definitions (EnumNotFound, etc.)
3. **Breaking Changes** - Must maintain backward compatibility with existing clients
4. **Performance Optimization** - Focus on correctness, not performance improvements

## User Impact

### Before (Current State)

```graphql
mutation {
  createEnum(input: {
    projectId: "my-project"
    name: "Status"
    title: "Order Status"
    options: [...]
  }) # Returns EnumDefinition! or throws generic error
}
```

Error response:
```json
{
  "errors": [
    {
      "message": "enum validation failed: Resource conflict: Enum already exists: Status",
      "path": ["createEnum"]
    }
  ],
  "data": { "createEnum": null }
}
```

### After (With Typed Errors)

```graphql
mutation {
  createEnum(input: {...}) {
    enum {
      id
      name
      title
    }
    error {
      __typename
      ... on Error {
        message
      }
      ... on ProjectNotFound {
        message  # "Project not found: invalid-project"
      }
      ... on EnumAlreadyExists {
        message    # "Enum already exists: Status"
        suggestion # "Please use a different enum name"
      }
      ... on InvalidEnumInput {
        message    # "Invalid enum options: duplicate code 'active'"
        suggestion # "Enum option codes must be unique"
      }
    }
  }
}
```

Error response:
```json
{
  "data": {
    "createEnum": {
      "enum": null,
      "error": {
        "__typename": "EnumAlreadyExists",
        "message": "Enum already exists: Status",
        "suggestion": "Please use a different enum name within this project"
      }
    }
  }
}
```

### Benefits

1. **Type-Safe Error Handling** - Clients can distinguish error types programmatically
2. **Better Error Messages** - Context-specific messages with helpful suggestions
3. **Consistent API** - Matches project and cluster API patterns
4. **Project-Scoped Validation** - Proper `ProjectNotFound` errors when project doesn't exist
5. **Reference Awareness** - `CannotDeleteReferencedEnum` error includes field names that reference the enum
6. **Backward Compatible** - Existing clients continue to work; new error field is optional

## Alternatives Considered

### Alternative 1: Keep Current Generic Error Handling

**Rejected because:**
- Inconsistent with project and cluster APIs
- Poor client experience - cannot handle errors programmatically
- No structured error information
- Generic error messages are harder for users to understand

### Alternative 2: Use Error Extensions

**Rejected because:**
- GraphQL error extensions are not type-safe
- Clients cannot use GraphQL schema for error type validation
- Less discoverable than typed errors in schema
- Not consistent with established pattern in this project

### Alternative 3: Separate Error Queries

**Rejected because:**
- Requires additional round-trip to fetch error details
- Complicates client code
- Not aligned with GraphQL best practices (GitHub, Shopify patterns)

## Open Questions

### Q1: Should we add error codes to the bizerrors package?

**Current state**: Enum-related errors use generic codes (`CONFLICT`, `NOT_FOUND`, `PARAM_INVALID`)

**Options**:
- **Option A**: Add specific error codes (`EnumAlreadyExists`, `EnumNotFound`, `CannotDeleteReferencedEnum`)
- **Option B**: Keep generic codes, differentiate in error adapter based on context

**Proposed**: **Option A** - Add specific error codes to match project and cluster patterns. This provides:
- Better error logging and debugging
- Clearer error tracking in monitoring systems
- Consistent pattern across all domain resources

**Error codes to add**:
```go
// pkg/bizerrors/common_errors.go
var (
    EnumNotFound = ErrorDefinition{
        Code:      ErrorTypeNotFound + ".ENUM",
        EnMessage: "Enum not found: {0}",
        ZhMessage: "枚举不存在: {0}",
    }

    EnumAlreadyExists = ErrorDefinition{
        Code:      ErrorTypeConflict + ".ENUM",
        EnMessage: "Enum already exists: {0}",
        ZhMessage: "枚举已存在: {0}",
    }

    CannotDeleteReferencedEnum = ErrorDefinition{
        Code:      ErrorTypeOperationFailed + ".ENUM",
        EnMessage: "Cannot delete enum '{0}', it is referenced by fields: {1}",
        ZhMessage: "无法删除枚举 '{0}'，它被以下字段引用: {1}",
    }
)
```

### Q2: Should we add InvalidEnumInput error type or reuse ParamInvalid?

**Options**:
- **Option A**: Create `InvalidEnumInput` GraphQL type (specific to enum validation errors)
- **Option B**: Reuse generic `ParamInvalid` type

**Proposed**: **Option A** - Create `InvalidEnumInput` type for consistency with `InvalidProjectInput` and `InvalidClusterInput`. This:
- Makes errors domain-specific and easier to handle
- Allows adding enum-specific validation details in the future
- Follows the established pattern

### Q3: How should we handle "cannot delete" errors for referenced enums?

**Current behavior**: `DeleteEnum` checks if enum is referenced by fields and returns error with field names

**Options**:
- **Option A**: Create `CannotDeleteReferencedEnum` error type with `referencedBy: [String!]!` field
- **Option B**: Use generic `OperationDenied` with message containing field names
- **Option C**: Create `CannotDeleteReferencedEnum` with message containing field names (no separate field)

**Proposed**: **Option C** - Use message-first approach consistent with project/cluster patterns:
```graphql
type CannotDeleteReferencedEnum implements Error {
  message: String!     # "Cannot delete enum 'Status', it is referenced by fields: Order.status, Invoice.status"
  suggestion: String   # "Please remove the enum from these fields before deleting"
}
```

This approach:
- Keeps schema simple (no complex fields)
- Message contains all context needed
- Consistent with `CannotDeleteDefaultProject` pattern
- Suggestions guide users to resolution

## Dependencies

### Required Before Implementation

- None - Project and cluster typed errors are already implemented and provide the pattern to follow

### Enables Future Work

- Model field typed error handling (enum fields reference enums)
- Enum validation API improvements
- Better error documentation generation

### Related Changes

- **Depends on**: `optimize-project-graphql-errors` (completed) - establishes error pattern
- **Depends on**: `implement-cluster-graphql-typed-errors` (completed) - demonstrates project-scoped resource pattern
- **Enables**: Future model field error handling improvements

## Success Criteria

### Functional Requirements

1. All enum mutations return typed errors via `error` field in payload types
2. All enum queries return typed errors via `error` field in payload types
3. Error adapter correctly converts bizerrors codes to GraphQL error types
4. Project validation returns `ProjectNotFound` error for invalid project IDs
5. Duplicate enum names return `EnumAlreadyExists` error with suggestion
6. Invalid enum input returns `InvalidEnumInput` error with specific validation message
7. Delete referenced enum returns `CannotDeleteReferencedEnum` with field names in message

### Technical Requirements

1. GraphQL schema validates correctly with `task generate-gql`
2. Error adapter unit tests cover all error conversion paths
3. Integration tests validate end-to-end error responses
4. All existing enum tests continue to pass
5. No breaking changes to existing GraphQL API

### Performance Requirements

1. Error adapter adds negligible overhead (<1ms)
2. No impact on operations that succeed (no errors)
3. Memory usage remains stable

### Documentation Requirements

1. GraphQL schema is self-documenting with error types
2. Error adapter code has clear comments explaining conversion logic
3. Integration tests serve as usage examples

## Implementation Phases

This change is small enough to implement in a single phase.

### Phase 1: Implementation (Single PR)

**Schema Changes** (backward compatible):
1. Add error types: `EnumAlreadyExists`, `EnumNotFound`, `InvalidEnumInput`, `CannotDeleteReferencedEnum`
2. Add error unions: `GetEnumError`, `CreateEnumError`, `UpdateEnumError`, `DeleteEnumError`
3. Add payload types: `GetEnumPayload`, `CreateEnumPayload`, `UpdateEnumPayload`, `DeleteEnumPayload`
4. Update mutations to return payload types instead of simple types
5. Update queries to return payload types where appropriate

**Code Changes**:
1. Add enum error codes to `pkg/bizerrors/common_errors.go`
2. Create `internal/interfaces/graphql/adapter/enum_error_adapter.go`
3. Create unit tests `internal/interfaces/graphql/adapter/enum_error_adapter_test.go`
4. Update `internal/interfaces/graphql/enum.resolvers.go` to use error adapter
5. Update enum service to use specific error codes instead of generic ones
6. Update integration tests in `tests/automated/test_enum_*.py`

**Validation**:
1. Run `task generate-gql` to regenerate GraphQL code
2. Run `task test-unit` to validate unit tests
3. Run `task auto-test` to validate integration tests
4. Manual testing via GraphQL Playground

## Timeline Considerations

This proposal does NOT include implementation timelines. Task breakdown and sequencing are documented in `tasks.md`.

## References

- **Pattern Reference**: `openspec/changes/archive/2026-01-30-optimize-project-graphql-errors/`
- **Similar Implementation**: `openspec/changes/archive/2026-01-30-implement-cluster-graphql-typed-errors/`
- **Error Adapter Example**: `internal/interfaces/graphql/adapter/cluster_error_adapter.go`
- **Current Schema**: `api/graph/schema/enum.graphql`
- **Current Resolver**: `internal/interfaces/graphql/enum.resolvers.go`
- **Current Service**: `internal/app/modeldesign/enum_service.go`
- **Error Package**: `pkg/bizerrors/common_errors.go`
- **CLAUDE.md Guidelines**: GraphQL Error Handling Pattern section
