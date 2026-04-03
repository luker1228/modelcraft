# Implementation Tasks for add-model-graphql-typed-errors

## 1. GraphQL Schema Changes

- [x] 1.1 Add error type definitions to `api/graph/schema/model.graphql`:
  - [x] ModelAlreadyExists with message (required) and suggestion (optional)
  - [x] ModelNotFound with message (required)
  - [x] InvalidModelInput with message (required) and suggestion (optional)
  - [x] CannotDeleteDeployedModel with message (required) and suggestion (optional)

- [x] 1.2 Add error union types to `api/graph/schema/model.graphql`:
  - [x] GetModelError = ModelNotFound | InvalidModelInput | ProjectNotFound
  - [x] CreateModelError = ModelAlreadyExists | InvalidModelInput | ProjectNotFound
  - [x] UpdateModelError = ModelNotFound | InvalidModelInput | ProjectNotFound
  - [x] DeleteModelError = ModelNotFound | CannotDeleteDeployedModel | ProjectNotFound

- [x] 1.3 Update payload types in `api/graph/schema/model.graphql`:
  - [x] GetModelPayload - change return type from direct Model to GetModelPayload with model and error fields
  - [x] CreateModelPayload - add optional error field
  - [x] UpdateModelPayload - add optional error field
  - [x] DeleteModelPayload - add optional error field

- [x] 1.4 Update mutations and queries in `api/graph/schema/model.graphql`:
  - [x] Update `model` query to return GetModelPayload
  - [x] Update `modelByName` query to return GetModelPayload
  - [x] Update `createModel` mutation to return CreateModelPayload
  - [x] Update `updateModel` mutation to return UpdateModelPayload
  - [x] Update `deleteModel` mutation to return DeleteModelPayload

- [x] 1.5 Run `task generate-gql` to regenerate GraphQL code

## 2. Error Adapter Implementation

- [x] 2.1 Create `internal/interfaces/graphql/adapter/model_error_adapter.go`:
  - [x] ConvertToGetError function
  - [x] ConvertToCreateError function
  - [x] ConvertToUpdateError function
  - [x] ConvertToDeleteError function

- [x] 2.2 Implement error conversion logic:
  - [x] Map NOT_FOUND.MODEL to ModelNotFound
  - [x] Map CONFLICT.MODEL to ModelAlreadyExists
  - [x] Map PARAM_INVALID to InvalidModelInput
  - [x] Map OPERATION_DENIED to CannotDeleteDeployedModel (for deployed model scenarios)
  - [x] Map NOT_FOUND.PROJECT to ProjectNotFound
  - [x] Handle unknown errors safely

## 3. Resolver Implementation Updates

- [x] 3.1 Update `Model` query resolver:
  - [x] Change return type to *generated.GetModelPayload
  - [x] Use ModelErrorAdapter.ConvertToGetError for errors
  - [x] Return model in data field or error in error field

- [x] 3.2 Update `ModelByName` query resolver:
  - [x] Change return type to *generated.GetModelPayload
  - [x] Use ModelErrorAdapter.ConvertToGetError for errors
  - [x] Return model in data field or error in error field

- [x] 3.3 Update `CreateModel` mutation resolver:
  - [x] Use ModelErrorAdapter.ConvertToCreateError for errors
  - [x] Return model in data field or error in error field

- [x] 3.4 Update `UpdateModel` mutation resolver:
  - [x] Use ModelErrorAdapter.ConvertToUpdateError for errors
  - [x] Return model in data field or error in error field

- [x] 3.5 Update `DeleteModel` mutation resolver:
  - [x] Use ModelErrorAdapter.ConvertToDeleteError for errors
  - [x] Return success=false with error in error field on failure

## 4. Testing

- [x] 4.1 Add unit tests for `model_error_adapter.go`:
  - [x] Test ModelNotFound conversion
  - [x] Test ModelAlreadyExists conversion
  - [x] Test InvalidModelInput conversion
  - [x] Test CannotDeleteDeployedModel conversion
  - [x] Test ProjectNotFound conversion
  - [x] Test unknown error handling

- [x] 4.2 Add integration tests for GraphQL errors:
  - [x] Test model query with non-existent ID returns ModelNotFound
  - [x] Test model query with non-existent project returns ProjectNotFound
  - [x] Test modelByName query with non-existent project returns ProjectNotFound
  - [x] Test modelByName query with non-existent model returns ModelNotFound
  - [x] Test create model with duplicate name returns ModelAlreadyExists
  - [x] Test create model with non-existent project returns ProjectNotFound
  - [x] Test update model with non-existent ID returns ModelNotFound
  - [x] Test update model with non-existent project returns ProjectNotFound
  - [x] Test delete model with non-existent ID returns ModelNotFound
  - [x] Test delete model with non-existent project returns ProjectNotFound
  - [x] Test successful operations return no error
  - [x] Test error message quality (descriptive, helpful suggestions)

- [x] 4.3 Verify backward compatibility:
  - [x] Note: Backward compatibility not required per user request ("don't need backward compatibility")

## 5. Documentation

- [x] 5.1 Update API documentation with error examples:
  - [x] Note: Not required per user request ("5.1 don't need do")

## 6. Code Quality

- [ ] 6.1 Run `task fmt` to format code
- [ ] 6.2 Run `task lint` to check for issues
- [ ] 6.3 Run `task vet` to verify code correctness
- [x] 6.4 Run `task build` to ensure successful compilation
