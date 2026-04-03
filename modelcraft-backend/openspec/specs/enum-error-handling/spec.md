# enum-error-handling Specification

## Purpose
TBD - created by archiving change add-enum-graphql-typed-errors. Update Purpose after archive.
## Requirements
### Requirement: Typed Error Interface for Enum Operations

Enum GraphQL operations MUST return structured, typed errors that implement the `Error` interface, enabling clients to programmatically distinguish between different error scenarios.

**Rationale**: Consistent with project and cluster APIs. Enables type-safe error handling in GraphQL clients.

**Related**: `project-management` spec, `cluster-management` spec

#### Scenario: Client queries non-existent enum

**Given** a project exists with ID "ecommerce"
**And** no enum named "Status" exists in the project
**When** client queries enum with projectName "ecommerce" and name "Status"
**Then** the response includes `error` field with `__typename: "EnumNotFound"`
**And** the error message is "Enum not found: Status"
**And** the `enum` field is `null`

#### Scenario: Client creates enum with duplicate name

**Given** a project exists with ID "ecommerce"
**And** an enum named "Status" already exists in the project
**When** client creates a new enum with projectName "ecommerce" and name "Status"
**Then** the response includes `error` field with `__typename: "EnumAlreadyExists"`
**And** the error message is "Enum already exists: Status"
**And** the error includes suggestion "Please use a different enum name within this project"
**And** the `enum` field is `null`

#### Scenario: Client creates enum with invalid project ID

**Given** no project exists with ID "invalid-project"
**When** client creates an enum with projectName "invalid-project"
**Then** the response includes `error` field with `__typename: "ProjectNotFound"`
**And** the error message is "Project not found: invalid-project"
**And** the `enum` field is `null`

#### Scenario: Client creates enum with invalid options

**Given** a project exists with ID "ecommerce"
**When** client creates an enum with duplicate option codes
**Then** the response includes `error` field with `__typename: "InvalidEnumInput"`
**And** the error message describes the validation failure (e.g., "Invalid enum options: duplicate code 'active'")
**And** the error includes suggestion for resolution
**And** the `enum` field is `null`

#### Scenario: Client deletes enum that is referenced by fields

**Given** a project exists with ID "ecommerce"
**And** an enum named "Status" exists in the project
**And** the enum is referenced by fields "Order.status" and "Invoice.status"
**When** client deletes the enum with projectName "ecommerce" and name "Status"
**Then** the response includes `error` field with `__typename: "CannotDeleteReferencedEnum"`
**And** the error message is "Cannot delete enum 'Status', it is referenced by fields: Order.status, Invoice.status"
**And** the error includes suggestion "Please remove the enum from these fields before deleting"
**And** the `success` field is `false`

### Requirement: Error Type Definitions

Enum GraphQL schema MUST define specific error types that implement the `Error` interface with required `message` field and optional `suggestion` field.

**Rationale**: Provides clear contracts for error responses. Enables GraphQL schema validation and client code generation.

**Related**: GraphQL Error Interface defined in `project.graphql`

#### Scenario: Schema defines EnumNotFound error type

**Given** the GraphQL schema is loaded
**When** introspecting the `EnumNotFound` type
**Then** it implements the `Error` interface
**And** it has a required `message: String!` field
**And** it has no other required fields

#### Scenario: Schema defines EnumAlreadyExists error type

**Given** the GraphQL schema is loaded
**When** introspecting the `EnumAlreadyExists` type
**Then** it implements the `Error` interface
**And** it has a required `message: String!` field
**And** it has an optional `suggestion: String` field

#### Scenario: Schema defines InvalidEnumInput error type

**Given** the GraphQL schema is loaded
**When** introspecting the `InvalidEnumInput` type
**Then** it implements the `Error` interface
**And** it has a required `message: String!` field
**And** it has an optional `suggestion: String` field

#### Scenario: Schema defines CannotDeleteReferencedEnum error type

**Given** the GraphQL schema is loaded
**When** introspecting the `CannotDeleteReferencedEnum` type
**Then** it implements the `Error` interface
**And** it has a required `message: String!` field
**And** it has an optional `suggestion: String` field

### Requirement: Error Unions per Operation

Each enum mutation and query MUST define a union type containing all possible error types for that operation.

**Rationale**: Type-safe error handling. Clients can exhaustively handle all error cases. Self-documenting API.

**Related**: Project and cluster error union patterns

#### Scenario: GetEnumError union defined

**Given** the GraphQL schema is loaded
**When** introspecting the `GetEnumError` union
**Then** it includes `EnumNotFound` type
**And** it includes `ProjectNotFound` type
**And** it includes no other types

#### Scenario: CreateEnumError union defined

**Given** the GraphQL schema is loaded
**When** introspecting the `CreateEnumError` union
**Then** it includes `EnumAlreadyExists` type
**And** it includes `InvalidEnumInput` type
**And** it includes `ProjectNotFound` type
**And** it includes no other types

#### Scenario: UpdateEnumError union defined

**Given** the GraphQL schema is loaded
**When** introspecting the `UpdateEnumError` union
**Then** it includes `EnumNotFound` type
**And** it includes `InvalidEnumInput` type
**And** it includes `ProjectNotFound` type
**And** it includes no other types

#### Scenario: DeleteEnumError union defined

**Given** the GraphQL schema is loaded
**When** introspecting the `DeleteEnumError` union
**Then** it includes `EnumNotFound` type
**And** it includes `CannotDeleteReferencedEnum` type
**And** it includes `ProjectNotFound` type
**And** it includes no other types

### Requirement: Payload Types with Error Field

Each enum mutation and query MUST return a payload type containing an optional `error` field of the corresponding error union type.

**Rationale**: Separates success and error states. Maintains backward compatibility by keeping nullable data fields.

**Related**: Project and cluster payload patterns

#### Scenario: GetEnumPayload type defined

**Given** the GraphQL schema is loaded
**When** introspecting the `GetEnumPayload` type
**Then** it has an optional `enum: EnumDefinition` field
**And** it has an optional `error: GetEnumError` field
**And** at most one of `enum` or `error` is non-null in responses

#### Scenario: CreateEnumPayload type defined

**Given** the GraphQL schema is loaded
**When** introspecting the `CreateEnumPayload` type
**Then** it has an optional `enum: EnumDefinition` field
**And** it has an optional `error: CreateEnumError` field
**And** at most one of `enum` or `error` is non-null in responses

#### Scenario: UpdateEnumPayload type defined

**Given** the GraphQL schema is loaded
**When** introspecting the `UpdateEnumPayload` type
**Then** it has an optional `enum: EnumDefinition` field
**And** it has an optional `error: UpdateEnumError` field
**And** at most one of `enum` or `error` is non-null in responses

#### Scenario: DeleteEnumPayload type defined

**Given** the GraphQL schema is loaded
**When** introspecting the `DeleteEnumPayload` type
**Then** it has a required `success: Boolean!` field
**And** it has an optional `error: DeleteEnumError` field
**And** when `error` is non-null, `success` is `false`
**And** when `error` is null, `success` is `true`

### Requirement: Business Error to GraphQL Error Conversion

The error adapter MUST convert domain-level business errors (bizerrors) to corresponding GraphQL error types based on error codes.

**Rationale**: Separates domain logic from GraphQL presentation. Centralizes error conversion logic for maintainability.

**Related**: `ClusterErrorAdapter`, `ProjectErrorAdapter` implementations

#### Scenario: Convert EnumNotFound bizerror

**Given** a `bizerrors.BusinessError` with code `NOT_FOUND.ENUM`
**When** the error adapter converts it for GetEnum operation
**Then** the result is `EnumNotFound` GraphQL type
**And** the message field contains the original error message

#### Scenario: Convert EnumAlreadyExists bizerror

**Given** a `bizerrors.BusinessError` with code `CONFLICT.ENUM`
**When** the error adapter converts it for CreateEnum operation
**Then** the result is `EnumAlreadyExists` GraphQL type
**And** the message field contains the original error message
**And** the suggestion field is "Please use a different enum name within this project"

#### Scenario: Convert ParamInvalid bizerror for enum validation

**Given** a `bizerrors.BusinessError` with code `PARAM_INVALID`
**When** the error adapter converts it for CreateEnum operation
**Then** the result is `InvalidEnumInput` GraphQL type
**And** the message field contains the validation error details
**And** the suggestion field contains guidance if available

#### Scenario: Convert CannotDeleteReferencedEnum bizerror

**Given** a `bizerrors.BusinessError` with code `OPERATION_DENIED.ENUM`
**And** the error message includes referenced field names
**When** the error adapter converts it for DeleteEnum operation
**Then** the result is `CannotDeleteReferencedEnum` GraphQL type
**And** the message field includes the enum name and referenced field names
**And** the suggestion field is "Please remove the enum from these fields before deleting"

#### Scenario: Convert ProjectNotFound bizerror

**Given** a `bizerrors.BusinessError` with code `NOT_FOUND.PROJECT`
**When** the error adapter converts it for any enum operation
**Then** the result is `ProjectNotFound` GraphQL type
**And** the message field contains the project ID

### Requirement: Backward Compatibility

Existing enum GraphQL clients MUST continue to function without modifications when typed errors are introduced.

**Rationale**: Non-breaking change. Gradual migration path for clients.

**Related**: Project and cluster typed error rollouts

#### Scenario: Existing client ignores error field

**Given** an existing client that queries enum operations
**And** the client does not query the `error` field
**When** an enum operation fails
**Then** the client receives `null` in the data field (existing behavior)
**And** the operation does not throw a GraphQL error at the top level
**And** the client's error handling logic continues to work

#### Scenario: List operations remain unchanged

**Given** the `enums` query returns a list of enums
**And** the `enumReferences` query returns a list of field names
**When** these queries are executed
**Then** they return arrays directly (not wrapped in payload types)
**And** they continue to work as before (no breaking changes)

