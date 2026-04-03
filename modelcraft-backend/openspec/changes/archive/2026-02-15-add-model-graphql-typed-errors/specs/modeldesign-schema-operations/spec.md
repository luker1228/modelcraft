## ADDED Requirements

### Requirement: Typed Error Interface for Model Operations

Model GraphQL operations MUST return structured, typed errors that implement the `Error` interface, enabling clients to programmatically distinguish between different error scenarios.

**Rationale**: Consistent with Project, Cluster, and Enum error handling patterns. Enables type-safe error handling in GraphQL clients.

**Related**: `project-management` spec, `cluster-management` spec, `enum-error-handling` spec

#### Scenario: Get model with non-existent ID returns ModelNotFound error

- **WHEN** client queries `model` with a model ID that does not exist
- **THEN** the response MUST include `error` field with `__typename: "ModelNotFound"`
- **AND** the error message is "Model not found: <model-id>"
- **AND** the `model` field is `null`

#### Scenario: Get model with non-existent project returns ProjectNotFound error

- **WHEN** client queries `model` with a project ID that does not exist
- **THEN** the response MUST include `error` field with `__typename: "ProjectNotFound"`
- **AND** the error message is "Project not found: <project-id>"
- **AND** the `model` field is `null`

#### Scenario: Create model with duplicate name returns ModelAlreadyExists error

- **WHEN** client creates a model with a name that already exists in the same project
- **THEN** the response MUST include `error` field with `__typename: "ModelAlreadyExists"`
- **AND** the error message is "Model already exists: <model-name>"
- **AND** the error includes suggestion "Please use a different model name within this project"
- **AND** the `model` field is `null`

#### Scenario: Create model with invalid input returns InvalidModelInput error

- **WHEN** client creates a model with invalid input (e.g., empty name, invalid cluster/database combination)
- **THEN** the response MUST include `error` field with `__typename: "InvalidModelInput"`
- **AND** the error message describes the validation failure
- **AND** the error includes suggestion for resolution
- **AND** the `model` field is `null`

#### Scenario: Update model with non-existent ID returns ModelNotFound error

- **WHEN** client updates a model with an ID that does not exist
- **THEN** the response MUST include `error` field with `__typename: "ModelNotFound"`
- **AND** the error message is "Model not found: <model-id>"
- **AND** the `model` field is `null`
- **AND** the `success` field is `false`

#### Scenario: Delete model with non-existent ID returns ModelNotFound error

- **WHEN** client deletes a model with an ID that does not exist
- **THEN** the response MUST include `error` field with `__typename: "ModelNotFound"`
- **AND** the error message is "Model not found: <model-id>"
- **AND** the `success` field is `false`

#### Scenario: Delete deployed model returns CannotDeleteDeployedModel error

- **WHEN** client attempts to delete a deployed model
- **THEN** the response MUST include `error` field with `__typename: "CannotDeleteDeployedModel"`
- **AND** the error message is "Cannot delete deployed model: <model-name>"
- **AND** the error includes suggestion "Please deploy a new version or delete the deployment first"
- **AND** the `success` field is `false`

### Requirement: Error Type Definitions

GraphQL schema MUST define specific error types for Model operations that implement the `Error` interface with required `message` field and optional `suggestion` field.

**Rationale**: Provides clear contracts for error responses. Enables GraphQL schema validation and client code generation.

**Related**: GraphQL Error Interface defined in `project.graphql`

#### Scenario: Schema defines ModelNotFound error type

- **WHEN** introspecting the `ModelNotFound` type
- **THEN** it MUST implement the `Error` interface
- **AND** it MUST have a required `message: String!` field
- **AND** it has no other required fields

#### Scenario: Schema defines ModelAlreadyExists error type

- **WHEN** introspecting the `ModelAlreadyExists` type
- **THEN** it MUST implement the `Error` interface
- **AND** it MUST have a required `message: String!` field
- **AND** it MUST have an optional `suggestion: String` field

#### Scenario: Schema defines InvalidModelInput error type

- **WHEN** introspecting the `InvalidModelInput` type
- **THEN** it MUST implement the `Error` interface
- **AND** it MUST have a required `message: String!` field
- **AND** it MUST have an optional `suggestion: String` field

#### Scenario: Schema defines CannotDeleteDeployedModel error type

- **WHEN** introspecting the `CannotDeleteDeployedModel` type
- **THEN** it MUST implement the `Error` interface
- **AND** it MUST have a required `message: String!` field
- **AND** it MUST have an optional `suggestion: String` field

### Requirement: Error Unions per Operation

Each model mutation and query MUST define a union type containing all possible error types for that operation.

**Rationale**: Type-safe error handling. Clients can exhaustively handle all error cases. Self-documenting API.

**Related**: Project, Cluster, and Enum error union patterns

#### Scenario: GetModelError union defined

- **WHEN** introspecting the `GetModelError` union type
- **THEN** it MUST include `ModelNotFound` type
- **AND** it MUST include `ProjectNotFound` type
- **AND** it MUST include no other types

#### Scenario: CreateModelError union defined

- **WHEN** introspecting the `CreateModelError` union type
- **THEN** it MUST include `ModelAlreadyExists` type
- **AND** it MUST include `InvalidModelInput` type
- **AND** it MUST include `ProjectNotFound` type
- **AND** it MUST include no other types

#### Scenario: UpdateModelError union defined

- **WHEN** introspecting the `UpdateModelError` union type
- **THEN** it MUST include `ModelNotFound` type
- **AND** it MUST include `InvalidModelInput` type
- **AND** it MUST include `ProjectNotFound` type
- **AND** it MUST include no other types

#### Scenario: DeleteModelError union defined

- **WHEN** introspecting the `DeleteModelError` union type
- **THEN** it MUST include `ModelNotFound` type
- **AND** it MUST include `CannotDeleteDeployedModel` type
- **AND** it MUST include `ProjectNotFound` type
- **AND** it MUST include no other types

### Requirement: Payload Types with Error Field

Each model mutation and query MUST return a payload type containing an optional `error` field of the corresponding error union type.

**Rationale**: Separates success and error states. Maintains backward compatibility by keeping nullable data fields.

**Related**: Project, Cluster, and Enum payload patterns

#### Scenario: GetModelPayload type defined

- **WHEN** introspecting the `GetModelPayload` type
- **THEN** it MUST have an optional `model: Model` field
- **AND** it MUST have an optional `error: GetModelError` field
- **AND** at most one of `model` or `error` is non-null in responses

#### Scenario: CreateModelPayload type defined

- **WHEN** introspecting the `CreateModelPayload` type
- **THEN** it MUST have an optional `model: Model` field
- **AND** it MUST have an optional `error: CreateModelError` field
- **AND** at most one of `model` or `error` is non-null in responses

#### Scenario: UpdateModelPayload type defined

- **WHEN** introspecting the `UpdateModelPayload` type
- **THEN** it MUST have a required `success: Boolean!` field
- **AND** it MUST have an optional `model: Model` field
- **AND** it MUST have an optional `error: UpdateModelError` field
- **AND** when `error` is non-null, `success` is `false`

#### Scenario: DeleteModelPayload type defined

- **WHEN** introspecting the `DeleteModelPayload` type
- **THEN** it MUST have a required `success: Boolean!` field
- **AND** it MUST have an optional `error: DeleteModelError` field
- **AND** when `error` is non-null, `success` is `false`

### Requirement: Business Error to GraphQL Error Conversion

The error adapter MUST convert domain-level business errors (bizerrors) to corresponding GraphQL error types based on error codes.

**Rationale**: Separates domain logic from GraphQL presentation. Centralizes error conversion logic for maintainability.

**Related**: `ClusterErrorAdapter`, `ProjectErrorAdapter`, `EnumErrorAdapter` implementations

#### Scenario: Convert ModelNotFound bizerror

- **WHEN** a `bizerrors.BusinessError` with code `NOT_FOUND.MODEL`
- **WHEN** the error adapter converts it for any model operation
- **THEN** the result MUST be `ModelNotFound` GraphQL type
- **AND** the message field MUST contain the original error message

#### Scenario: Convert ModelAlreadyExists bizerror

- **WHEN** a `bizerrors.BusinessError` with code `CONFLICT.MODEL`
- **WHEN** the error adapter converts it for CreateModel operation
- **THEN** the result MUST be `ModelAlreadyExists` GraphQL type
- **AND** the message field MUST contain the original error message
- **AND** the suggestion field MUST be "Please use a different model name within this project"

#### Scenario: Convert ParamInvalid bizerror for model validation

- **WHEN** a `bizerrors.BusinessError` with code `PARAM_INVALID`
- **WHEN** the error adapter converts it for any model operation
- **THEN** the result MUST be `InvalidModelInput` GraphQL type
- **AND** the message field MUST contain the validation error details
- **AND** the suggestion field MAY contain guidance if available

#### Scenario: Convert OperationDenied bizerror for deployed model

- **WHEN** a `bizerrors.BusinessError` with code containing `OPERATION_DENIED`
- **AND** the operation is deleting a deployed model
- **WHEN** the error adapter converts it for DeleteModel operation
- **THEN** the result MUST be `CannotDeleteDeployedModel` GraphQL type
- **AND** the message field MUST describe the restriction

#### Scenario: Convert ProjectNotFound bizerror

- **WHEN** a `bizerrors.BusinessError` with code `NOT_FOUND.PROJECT`
- **WHEN** the error adapter converts it for any model operation
- **THEN** the result MUST be `ProjectNotFound` GraphQL type
- **AND** the message field MUST contain the project ID

### Requirement: Backward Compatibility

Existing model GraphQL clients MUST continue to function without modifications when typed errors are introduced.

**Rationale**: Non-breaking change. Gradual migration path for clients.

**Related**: Project, Cluster, and Enum typed error rollouts

#### Scenario: Existing client ignores error field

- **WHEN** an existing client queries model operations
- **AND** the client does not query the `error` field
- **WHEN** a model operation fails
- **THEN** the client MUST receive `null` in the data field (existing behavior)
- **AND** the operation MUST NOT throw a GraphQL error at the top level
- **AND** the client's error handling logic MUST continue to work

#### Scenario: List and connection operations remain unchanged

- **WHEN** the `models` query returns a ModelConnection
- **AND** other model-related queries that return lists or connections are executed
- **THEN** they MUST return arrays directly (not wrapped in payload types)
- **AND** they MUST continue to work as before (no breaking changes)

#### Scenario: ReverseEngineerModel operation remains unchanged

- **WHEN** the `reverseEngineerModel` mutation is executed
- **THEN** it MUST continue to return the existing ReverseEngineerModelPayload type
- **AND** it MUST NOT include an error field change in this scope
