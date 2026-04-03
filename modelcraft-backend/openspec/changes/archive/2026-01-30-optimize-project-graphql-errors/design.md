## Context

The current GraphQL API uses a simple error handling pattern where mutations return payload types with:
- Boolean `success` field
- Nullable data field (e.g., `project: Project`)

This approach has limitations:
1. **Poor error discrimination**: Clients can't distinguish between different error types programmatically
2. **Generic error messages**: Errors are buried in the generic GraphQL error response, not in the payload
3. **No structured retry logic**: Clients can't determine if an error is retryable
4. **Inconsistent with industry best practices**: GitHub, Shopify, and other GraphQL APIs use typed errors

The system already has a robust error classification system in `pkg/bizerrors/` with error codes like:
- `NOT_FOUND.PROJECT` - Project does not exist
- `CONFLICT.PROJECT` - Project ID already exists
- `PARAM_INVALID.PROJECT` - Invalid input parameters
- `OPERATION_DENIED.PROJECT` - Cannot perform operation (e.g., delete default project)

## Goals / Non-Goals

**Goals:**
- Provide structured, typed errors in GraphQL responses following industry best practices
- Enable clients to handle specific error cases programmatically
- Maintain backward compatibility with existing clients
- Map existing `bizerrors` error codes to GraphQL error types
- Improve developer experience with clear error handling patterns

**Non-Goals:**
- Changing the underlying error handling in `pkg/bizerrors/` (already well-designed)
- Modifying error handling for other resources (Model, Cluster, Enum) - can be done in future changes
- Changing HTTP status codes or REST API error handling
- Breaking existing client code

## Decisions

### Decision 1: Use Union Types for Mutation Errors

**Choice:** Each mutation returns a union type containing all possible error types for that operation

**Rationale:**
- Follows GraphQL best practices (GitHub, Shopify patterns)
- Type-safe error handling on client side
- Clear documentation of possible errors per operation
- GraphQL introspection reveals all possible error cases

**Example:**
```graphql
union CreateProjectError = ProjectAlreadyExists | InvalidProjectInput

type CreateProjectPayload {
  project: Project
  errors: [CreateProjectError!]!
}
```

**Alternatives considered:**
1. Single shared error type for all operations
   - ❌ Less type-safe, clients can't know which errors apply to which mutations
2. Error codes as strings
   - ❌ Not type-safe, loses benefits of GraphQL type system
3. Throwing exceptions only (no payload errors)
   - ❌ Violates GraphQL best practices, errors are expected business cases, not exceptions

### Decision 2: Maintain Backward Compatibility

**Choice:** Keep existing `success: Boolean!` and nullable data fields, add new `errors` field

**Rationale:**
- Zero breaking changes for existing clients
- Gradual migration path
- Clients can adopt typed errors incrementally

**Migration pattern:**
```graphql
# Old client (still works)
mutation {
  createProject(input: {...}) {
    success
    project { id }
  }
}

# New client (preferred)
mutation {
  createProject(input: {...}) {
    project { id }
    errors {
      __typename
      ... on Error {
        message
      }
      ... on ProjectAlreadyExists {
        message
        suggestion
      }
    }
  }
}
```

### Decision 3: Simplified Error Interface

**Choice:** Define `Error` interface with only `message` field, and optional `suggestion` field in specific error types

**Rationale:**
- Simpler, cleaner API - most errors only need a descriptive message
- Avoids over-engineering with excessive fields like `path`, `code`, etc.
- `suggestion` field can be added to specific error types when helpful (e.g., input validation)
- Follows principle of "start simple, add complexity only when needed"
- Easier for clients to consume - just check `message` and `suggestion`

**Structure:**
```graphql
interface Error {
  message: String!
}

type ProjectAlreadyExists implements Error {
  message: String!
  suggestion: String  # Optional - only when helpful
}

type InvalidProjectInput implements Error {
  message: String!
  suggestion: String  # e.g., "Project ID must be lowercase with hyphens"
}
```

**What we removed:**
- ❌ `path` field - adds complexity without clear value for most cases
- ❌ Error-specific context fields (like `existingProjectId`, `projectId`) - message is sufficient
- ❌ Complex error structures - keep it simple

### Decision 4: Map bizerrors to GraphQL Errors in Adapter Layer

**Choice:** Create dedicated adapter (`project_error_adapter.go`) to convert `bizerrors.BusinessError` to GraphQL error types

**Rationale:**
- Separation of concerns: resolvers focus on business logic, adapters handle mapping
- Testable: error conversion logic can be unit tested independently
- Reusable: pattern can be applied to Model, Cluster, Enum in future
- Maintains DDD architecture: GraphQL layer doesn't pollute domain layer

**Implementation pattern:**
```go
func ConvertToGraphQLError(bizErr *bizerrors.BusinessError) (generated.CreateProjectError, error) {
    switch bizErr.Info.GetCode() {
    case bizerrors.ProjectAlreadyExists.GetCode():
        return &generated.ProjectAlreadyExists{
            Message: bizErr.Msg(),
            Suggestion: "Please use a different project ID",
        }, nil
    case bizerrors.ParamInvalid.GetCode():
        return &generated.InvalidProjectInput{
            Message: bizErr.Msg(),
            Suggestion: extractSuggestionFromError(bizErr),
        }, nil
    // ... other cases
    }
}
```

## Risks / Trade-offs

### Risk 1: Schema Complexity
**Risk:** GraphQL schema becomes more verbose with many error types

**Mitigation:**
- Start with Project only (4 error types, 5 mutations)
- Document patterns clearly for future resource types
- Use code generation tools to reduce boilerplate

**Trade-off accepted:** Verbosity is worth the type safety and clarity

### Risk 2: Code Generation Issues
**Risk:** gqlgen may have issues with union types or generate unexpected code

**Mitigation:**
- Test with small example first
- Review generated code in PR
- gqlgen v0.17.83 (project version) has mature union support

**Likelihood:** Low - union types are well-supported in gqlgen

### Risk 3: Client Migration Burden
**Risk:** Clients may not adopt new error pattern if migration is complex

**Mitigation:**
- Keep old pattern working (backward compatible)
- Provide clear documentation with examples
- Show benefits (better error messages, retry logic)

**Trade-off accepted:** Optional migration is better than forcing breaking change

## Migration Plan

### Phase 1: Implementation (This Change)
1. Add new schema definitions to `project.graphql`
2. Implement error adapter layer
3. Update resolvers to populate `errors` field
4. Keep all existing fields functional
5. Run code generation (`task generate-gql`)
6. Add comprehensive tests

### Phase 2: Documentation & Rollout
1. Update API documentation with error handling examples
2. Announce new pattern in release notes
3. Monitor client adoption through logs/metrics
4. Gather feedback from API consumers

### Phase 3: Expand Pattern (Future Changes)
1. Apply same pattern to Model mutations
2. Apply to Cluster mutations
3. Apply to Enum mutations
4. Consider deprecation timeline for old pattern (6-12 months minimum)

### Rollback Plan
If issues are discovered post-deployment:
1. New `errors` field can be ignored by clients (non-breaking)
2. Old `success` and nullable fields still work
3. Can fix error adapter without schema changes
4. Worst case: remove `errors` field in next deployment (still backward compatible)

## Open Questions

### Q1: Should we include suggestion field in all error types?
**Current thinking:** Only include in errors where suggestions make sense (e.g., ProjectAlreadyExistsError could suggest using different ID)

**Decision needed:** Yes/No - will decide during implementation

### Q2: Should path field use JSON path format or custom format?
**Current thinking:** Use simple format like "input.id", "input.title" - matches GraphQL field paths

**Decision needed:** Finalize during implementation based on gqlgen capabilities

### Q3: Should we add error codes as enums?
**Current thinking:** No - error types are sufficient, adding codes would be redundant

**Decision needed:** Confirm during schema design

### Q4: Future: Should we add a RetryableError interface?
**Current thinking:** Not in this change, but consider for future (network errors vs validation errors)

**Decision needed:** Out of scope for this change
