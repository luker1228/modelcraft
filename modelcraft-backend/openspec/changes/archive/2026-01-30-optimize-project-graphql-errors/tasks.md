## 1. Schema Definition

- [x] 1.1 Define `Error` interface in `project.graphql`
- [x] 1.2 Define specific error types (`ProjectAlreadyExists`, `ProjectNotFound`, `InvalidProjectInput`, `CannotDeleteDefaultProject`)
- [x] 1.3 Create union types for each mutation (`CreateProjectError`, `UpdateProjectError`, etc.)
- [x] 1.4 Update payload types to include typed `errors` field while maintaining existing fields

## 2. Error Adapter Implementation

- [x] 2.1 Create `internal/interfaces/graphql/adapter/project_error_adapter.go`
- [x] 2.2 Implement `ConvertToGraphQLError()` function to map `bizerrors.BusinessError` to GraphQL error types
- [x] 2.3 Add helper functions for each error type conversion
- [x] 2.4 Handle edge cases (nil errors, unknown error types)

## 3. Resolver Updates

- [x] 3.1 Update `CreateProject` resolver to populate `errors` field
- [x] 3.2 Update `UpdateProject` resolver to populate `errors` field
- [x] 3.3 Update `DeleteProject` resolver to populate `errors` field
- [x] 3.4 Update `ArchiveProject` resolver to populate `errors` field
- [x] 3.5 Update `ActivateProject` resolver to populate `errors` field
- [x] 3.6 Ensure backward compatibility (maintain `success` and nullable fields)

## 4. Business Error Definitions

- [x] 4.1 Verify `ProjectNotFound` exists in `pkg/bizerrors/common_errors.go`
- [x] 4.2 Add `ProjectAlreadyExists` (added as `CONFLICT.PROJECT`)
- [x] 4.3 Add `InvalidProjectInput` (added as `PARAM_INVALID.PROJECT`)
- [x] 4.4 Add `CannotDeleteDefaultProject` (added as `OPERATION_DENIED.PROJECT`)

## 5. Code Generation

- [x] 5.1 Run `task generate-gql` to regenerate GraphQL code
- [x] 5.2 Verify generated types in `internal/interfaces/graphql/generated/`
- [x] 5.3 Fix any compilation errors from new types

## 6. Testing

- [x] 6.1 Add unit tests for error adapter (all error type conversions)
- [x] 6.2 Verify existing project tests pass (61.2% coverage maintained)
- [x] 6.3 Verify backward compatibility with existing tests
- [x] 6.4 Manual verification via GraphQL code generation and compilation

## 7. Documentation

- [x] 7.1 GraphQL schema documentation comments already in schema file
- [x] 7.2 Error handling examples in `openspec/changes/optimize-project-graphql-errors/EXAMPLE_SCHEMA.md`
- [x] 7.3 Migration path documented in proposal.md and design.md

## 8. Validation

- [x] 8.1 Run `openspec validate optimize-project-graphql-errors --strict` - PASSED
- [x] 8.2 Project unit tests pass - PASSED (61.2% coverage)
- [x] 8.3 Error adapter unit tests pass - PASSED (all 10 test scenarios)
- [x] 8.4 Code builds successfully without errors
