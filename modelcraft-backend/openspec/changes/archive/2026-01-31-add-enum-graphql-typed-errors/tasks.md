# Tasks: Add Enum GraphQL Typed Errors

## Overview

This document breaks down the implementation of typed error handling for enum GraphQL operations into sequential, verifiable tasks. Each task represents a small, testable unit of work that delivers visible progress.

## Task Checklist

### Phase 1: Preparation and Planning

- [ ] **Review existing error patterns**
  - Read `internal/interfaces/graphql/adapter/project_error_adapter.go`
  - Read `internal/interfaces/graphql/adapter/cluster_error_adapter.go`
  - Read `api/graph/schema/project.graphql` error definitions
  - Read `api/graph/schema/cluster.graphql` error definitions
  - **Validation**: Understand the established pattern and naming conventions

- [ ] **Analyze current enum implementation**
  - Read `api/graph/schema/enum.graphql` current schema
  - Read `internal/interfaces/graphql/enum.resolvers.go` resolver implementation
  - Read `internal/app/modeldesign/enum_service.go` service layer
  - Read `pkg/bizerrors/common_errors.go` error definitions
  - **Validation**: Identify all error scenarios that need typed errors

### Phase 2: Add Error Codes to bizerrors Package

- [ ] **Add enum-specific error codes**
  - Open `pkg/bizerrors/common_errors.go`
  - Add `EnumNotFound` error definition with code `NOT_FOUND.ENUM`
  - Add `EnumAlreadyExists` error definition with code `CONFLICT.ENUM`
  - Add `CannotDeleteReferencedEnum` error definition with code `OPERATION_DENIED.ENUM`
  - **Validation**: Run `task test-unit` to ensure no compilation errors

- [ ] **Update enum service to use specific error codes**
  - Open `internal/app/modeldesign/enum_service.go`
  - Update `CreateEnum` to check for duplicates and return `EnumAlreadyExists` error
  - Update `GetEnum` to return `EnumNotFound` error when not found
  - Update `UpdateEnum` to return `EnumNotFound` error when not found
  - Update `DeleteEnum` to return `CannotDeleteReferencedEnum` error when referenced
  - **Validation**: Run `task test-unit` to verify service tests still pass

### Phase 3: Update GraphQL Schema

- [ ] **Add error interface and types**
  - Open `api/graph/schema/enum.graphql`
  - Add error types (after imports, before queries):
    ```graphql
    # ============================================
    # Enum Error Types
    # ============================================

    # Enum-specific error types
    type EnumAlreadyExists implements Error {
      message: String!
      suggestion: String
    }

    type EnumNotFound implements Error {
      message: String!
    }

    type InvalidEnumInput implements Error {
      message: String!
      suggestion: String
    }

    type CannotDeleteReferencedEnum implements Error {
      message: String!
      suggestion: String
    }
    ```
  - **Validation**: Schema is syntactically correct (visual inspection)

- [ ] **Add error unions**
  - Add error unions after error types:
    ```graphql
    # Error unions for each mutation and query
    union GetEnumError = EnumNotFound | ProjectNotFound
    union CreateEnumError = EnumAlreadyExists | InvalidEnumInput | ProjectNotFound
    union UpdateEnumError = EnumNotFound | InvalidEnumInput | ProjectNotFound
    union DeleteEnumError = EnumNotFound | CannotDeleteReferencedEnum | ProjectNotFound
    ```
  - **Validation**: Verify all error types are defined before unions

- [ ] **Add payload types**
  - Add payload types after error unions:
    ```graphql
    # ============================================
    # Enum Payload Types
    # ============================================

    type GetEnumPayload {
      enum: EnumDefinition
      error: GetEnumError
    }

    type CreateEnumPayload {
      enum: EnumDefinition
      error: CreateEnumError
    }

    type UpdateEnumPayload {
      enum: EnumDefinition
      error: UpdateEnumError
    }

    type DeleteEnumPayload {
      success: Boolean!
      error: DeleteEnumError
    }
    ```
  - **Validation**: Verify payload types reference correct union types

- [ ] **Update query and mutation signatures**
  - Update query `enum` to return `GetEnumPayload!`
  - Update mutation `createEnum` to return `CreateEnumPayload!`
  - Update mutation `updateEnum` to return `UpdateEnumPayload!`
  - Update mutation `deleteEnum` to return `DeleteEnumPayload!`
  - Keep `enums` (list) and `enumReferences` queries unchanged (they don't have error scenarios)
  - **Validation**: All queries and mutations reference defined payload types

- [ ] **Generate GraphQL code**
  - Run `task generate-gql`
  - **Validation**: Code generation succeeds without errors

### Phase 4: Create Error Adapter

- [ ] **Create enum error adapter**
  - Create file `internal/interfaces/graphql/adapter/enum_error_adapter.go`
  - Copy structure from `cluster_error_adapter.go` as template
  - Implement `ConvertToGetEnumError` method
  - Implement `ConvertToCreateEnumError` method
  - Implement `ConvertToUpdateEnumError` method
  - Implement `ConvertToDeleteEnumError` method
  - Add appropriate suggestions for each error type
  - **Validation**: Code compiles without errors

- [ ] **Create error adapter unit tests**
  - Create file `internal/interfaces/graphql/adapter/enum_error_adapter_test.go`
  - Copy test structure from `cluster_error_adapter_test.go` as template
  - Write test for `ConvertToGetEnumError`:
    - Test `EnumNotFound` conversion
    - Test `ProjectNotFound` conversion
    - Test unknown error code fallback
  - Write test for `ConvertToCreateEnumError`:
    - Test `EnumAlreadyExists` conversion (with suggestion)
    - Test `InvalidEnumInput` conversion
    - Test `ProjectNotFound` conversion
    - Test unknown error code fallback
  - Write test for `ConvertToUpdateEnumError`:
    - Test `EnumNotFound` conversion
    - Test `InvalidEnumInput` conversion
    - Test `ProjectNotFound` conversion
    - Test unknown error code fallback
  - Write test for `ConvertToDeleteEnumError`:
    - Test `EnumNotFound` conversion
    - Test `CannotDeleteReferencedEnum` conversion (with suggestion)
    - Test `ProjectNotFound` conversion
    - Test unknown error code fallback
  - **Validation**: Run `task test-unit` - all adapter tests pass

### Phase 5: Update Resolvers

- [ ] **Update GetEnum (query) resolver**
  - Open `internal/interfaces/graphql/enum.resolvers.go`
  - Change return type from `*generated.EnumDefinition` to `*generated.GetEnumPayload`
  - Remove `WithGraphqlErrorHandler` wrapper
  - Add error adapter conversion logic
  - Handle `nil` enum (not found case)
  - Cast bizerror to BusinessError for adapter
  - **Validation**: Code compiles without errors

- [ ] **Update CreateEnum (mutation) resolver**
  - Change return type from `*generated.EnumDefinition` to `*generated.CreateEnumPayload`
  - Remove `WithGraphqlErrorHandler` wrapper
  - Add error adapter conversion logic
  - Handle validation errors with `InvalidEnumInput`
  - Handle duplicate enum errors with `EnumAlreadyExists`
  - Cast bizerror to BusinessError for adapter
  - **Validation**: Code compiles without errors

- [ ] **Update UpdateEnum (mutation) resolver**
  - Change return type from `*generated.EnumDefinition` to `*generated.UpdateEnumPayload`
  - Remove `WithGraphqlErrorHandler` wrapper
  - Add error adapter conversion logic
  - Handle not found errors
  - Handle validation errors
  - Cast bizerror to BusinessError for adapter
  - **Validation**: Code compiles without errors

- [ ] **Update DeleteEnum (mutation) resolver**
  - Change return type from `bool` to `*generated.DeleteEnumPayload`
  - Remove `WithGraphqlErrorHandler` wrapper
  - Add error adapter conversion logic
  - Handle not found errors
  - Handle reference constraint errors with `CannotDeleteReferencedEnum`
  - Cast bizerror to BusinessError for adapter
  - **Validation**: Code compiles without errors

- [ ] **Initialize error adapter in resolver**
  - Add `enumErrorAdapter` field to resolver struct (if needed)
  - Initialize adapter in resolver constructor
  - Use adapter in all enum resolvers
  - **Validation**: Run `task build` to ensure everything compiles

### Phase 6: Testing

- [ ] **Run unit tests**
  - Run `task test-unit`
  - **Validation**: All unit tests pass, including new error adapter tests

- [ ] **Update integration tests for typed errors**
  - Find enum integration tests in `tests/automated/`
  - Update tests to check for error field in response payloads
  - Add test for `EnumAlreadyExists` error scenario
  - Add test for `EnumNotFound` error scenario (query)
  - Add test for `InvalidEnumInput` error scenario
  - Add test for `CannotDeleteReferencedEnum` error scenario
  - Add test for `ProjectNotFound` error scenario
  - **Validation**: Run `task auto-test` - all integration tests pass

- [ ] **Manual testing in GraphQL Playground**
  - Start server with `task run`
  - Open GraphQL Playground at http://localhost:8080/playground
  - Test successful enum creation
  - Test duplicate enum creation (expect `EnumAlreadyExists` error)
  - Test querying non-existent enum (expect `EnumNotFound` error)
  - Test updating non-existent enum (expect `EnumNotFound` error)
  - Test deleting referenced enum (expect `CannotDeleteReferencedEnum` error)
  - Test operations with invalid project ID (expect `ProjectNotFound` error)
  - **Validation**: All error scenarios return proper typed errors with helpful messages

### Phase 7: Documentation and Cleanup

- [ ] **Update schema documentation**
  - Add inline comments to error types explaining when they occur
  - Add examples to error type suggestions
  - **Validation**: Schema is well-documented

- [ ] **Code cleanup**
  - Remove any unused imports
  - Ensure consistent error message formatting
  - Verify all logging statements use proper format
  - **Validation**: Run `task lint` - no new warnings

- [ ] **Final validation**
  - Run `task check-all` (format + lint + vet + test)
  - Run `task auto-test` for full integration test suite
  - **Validation**: All checks pass

## Task Dependencies

### Sequential Tasks (must be done in order)

1. Phase 1 tasks (Preparation) → Phase 2 tasks (Error codes)
2. Phase 2 tasks → Phase 3 tasks (Schema)
3. Phase 3 tasks → Phase 4 tasks (Error adapter)
4. Phase 4 tasks → Phase 5 tasks (Resolvers)
5. Phase 5 tasks → Phase 6 tasks (Testing)
6. Phase 6 tasks → Phase 7 tasks (Documentation)

### Parallel Tasks (can be done simultaneously)

Within Phase 4:
- Error adapter implementation and test file creation can be started in parallel

Within Phase 5:
- All resolver updates can be done in parallel after error adapter is complete

Within Phase 6:
- Unit test updates and integration test updates can be done in parallel

## Validation Checklist

After completing all tasks, verify:

- [ ] `task generate-gql` succeeds
- [ ] `task build` succeeds
- [ ] `task test-unit` passes (100% of tests)
- [ ] `task auto-test` passes (100% of tests)
- [ ] `task lint` reports no new warnings
- [ ] `task fmt-check` reports no formatting issues
- [ ] Manual GraphQL Playground testing shows proper error responses
- [ ] All error scenarios return typed errors (not generic GraphQL errors)
- [ ] Error messages are clear and helpful
- [ ] Suggestions guide users to resolution
- [ ] No breaking changes (existing queries still work)

## Success Metrics

- **Code Coverage**: Error adapter unit tests achieve >90% coverage
- **Integration Tests**: All enum error scenarios covered by integration tests
- **Performance**: Error adapter adds <1ms overhead
- **Backward Compatibility**: All existing enum tests pass without modification
- **User Experience**: Error messages are clear and actionable

## Notes

- Follow TDD principles: Write tests before implementing resolver changes
- Use existing project and cluster error adapters as reference implementations
- Keep error messages user-friendly and include context (project ID, enum name)
- All error messages should follow the format: "Action failed: resource_name"
- Suggestions should guide users to resolution steps
- Maintain consistency with CLAUDE.md GraphQL error handling guidelines
