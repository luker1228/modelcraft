# modeldesign-schema-operations Specification

## Purpose
TBD - created by archiving change add-schema-based-model-operations. Update Purpose after archive.
## Requirements
### Requirement: JSON Schema Parsing Service

The system SHALL provide a domain service that parses JSON Schema Draft 7 documents into ModelCraft DataModel entities.

#### Scenario: Parse basic schema structure

- **WHEN** a valid JSON Schema Draft 7 document is provided
- **AND** it contains `$schema`, `type: "object"`, `title`, `description`, `properties`, and `required` fields
- **THEN** the parser creates a `DataModel` entity
- **AND** sets `Title` from schema `title`
- **AND** sets `Description` from schema `description`
- **AND** sets `ModelName` from `x-modelName` custom property
- **AND** sets `ClusterName` from `x-clusterName` custom property
- **AND** sets `DatabaseName` from `x-databaseName` custom property

#### Scenario: Parse field definitions

- **WHEN** a schema has a `properties` object with 3 field definitions
- **THEN** the parser creates 3 `FieldDefinition` entities
- **AND** each field's `Name` matches the property key
- **AND** each field's `Title` and `Description` are extracted from property metadata
- **AND** field types are mapped from JSON Schema types to ModelCraft formats

#### Scenario: Map required fields

- **WHEN** a schema has `required: ["id", "name"]`
- **THEN** fields named "id" and "name" have `Required = true`
- **AND** other fields have `Required = false`

#### Scenario: Handle nullable fields

- **WHEN** a field property has `nullable: true`
- **THEN** the corresponding `FieldDefinition` has `NonNull = false`
- **WHEN** a field property has no `nullable` keyword or `nullable: false`
- **THEN** the corresponding `FieldDefinition` has `NonNull = true`

#### Scenario: Reject invalid schema version

- **WHEN** a JSON document has `$schema` other than "http://json-schema.org/draft-07/schema#"
- **THEN** parsing fails with error indicating unsupported schema version

#### Scenario: Reject missing required metadata

- **WHEN** a schema lacks `x-modelName`, `x-clusterName`, or `x-databaseName` custom properties
- **THEN** parsing fails with error indicating which required metadata is missing

### Requirement: Type Mapping from JSON Schema

The system SHALL map JSON Schema types and formats to ModelCraft field types bidirectionally.

#### Scenario: Map string type

- **WHEN** a property has `type: "string"` with no format
- **THEN** the field has `Type.Format = STRING`

#### Scenario: Map UUID format

- **WHEN** a property has `type: "string"` and `format: "uuid"`
- **THEN** the field has `Type.Format = UUID`

#### Scenario: Map date format

- **WHEN** a property has `type: "string"` and `format: "date"`
- **THEN** the field has `Type.Format = DATE`

#### Scenario: Map date-time format

- **WHEN** a property has `type: "string"` and `format: "date-time"`
- **THEN** the field has `Type.Format = DATETIME`

#### Scenario: Map time format

- **WHEN** a property has `type: "string"` and `format: "time"`
- **THEN** the field has `Type.Format = TIME`

#### Scenario: Map number type

- **WHEN** a property has `type: "number"`
- **THEN** the field has `Type.Format = NUMBER`

#### Scenario: Map integer type

- **WHEN** a property has `type: "integer"`
- **THEN** the field has `Type.Format = INTEGER`

#### Scenario: Map decimal type with precision

- **WHEN** a property has `type: "number"`, `x-precision: 10`, and `x-scale: 2`
- **THEN** the field has `Type.Format = DECIMAL`
- **AND** `Validation.Precision = 10` and `Validation.Scale = 2`

#### Scenario: Map boolean type

- **WHEN** a property has `type: "boolean"`
- **THEN** the field has `Type.Format = BOOLEAN`

#### Scenario: Map enum type

- **WHEN** a property has `type: "string"` and `enum: ["active", "inactive"]`
- **THEN** the field has `Type.Format = ENUM`
- **AND** `Validation.EnumValues` contains ["active", "inactive"]

#### Scenario: Map enum array type

- **WHEN** a property has `type: "array"` and `items: { type: "string", enum: ["read", "write"] }`
- **THEN** the field has `Type.Format = ENUM_ARRAY`
- **AND** `Validation.EnumValues` contains ["read", "write"]

#### Scenario: Parse full enum definition

- **WHEN** a property has `x-enum` custom property with `name`, `title`, `description`, `options` array
- **THEN** the parser creates an `EnumDefinition` entity
- **AND** associates it with the field

### Requirement: Validation Rule Mapping

The system SHALL convert JSON Schema validation keywords to ModelCraft ValidationConfig.

#### Scenario: Map string length constraints

- **WHEN** a property has `minLength: 5` and `maxLength: 100`
- **THEN** the field's `Validation.MinLength = 5` and `Validation.MaxLength = 100`

#### Scenario: Map pattern constraint

- **WHEN** a property has `pattern: "^[A-Z][a-z]+$"`
- **THEN** the field's `Validation.Pattern = "^[A-Z][a-z]+$"`

#### Scenario: Map numeric constraints

- **WHEN** a property has `minimum: 0` and `maximum: 100`
- **THEN** the field's `Validation.Minimum = 0` and `Validation.Maximum = 100`

#### Scenario: Map array item constraints

- **WHEN** a property has `minItems: 1` and `maxItems: 5`
- **THEN** the field's `Validation.MinItems = 1` and `Validation.MaxItems = 5`

#### Scenario: Map date range constraints

- **WHEN** a property has `x-minDate: "2024-01-01"` and `x-maxDate: "2024-12-31"`
- **THEN** the field's `Validation.MinDate = "2024-01-01"` and `Validation.MaxDate = "2024-12-31"`

#### Scenario: Map time range constraints

- **WHEN** a property has `x-minTime: "09:00:00"` and `x-maxTime: "18:00:00"`
- **THEN** the field's `Validation.MinTime = "09:00:00"` and `Validation.MaxTime = "18:00:00"`

### Requirement: Custom Metadata Extraction

The system SHALL extract ModelCraft-specific metadata from custom `x-*` properties.

#### Scenario: Extract display order

- **WHEN** a property has `x-displayOrder: 10`
- **THEN** the field's `DisplayOrder = 10`

#### Scenario: Extract primary key flag

- **WHEN** a property has `x-isPrimary: true`
- **THEN** the field's `IsPrimary = true`

#### Scenario: Extract unique constraint flag

- **WHEN** a property has `x-isUnique: true`
- **THEN** the field's `IsUnique = true`

#### Scenario: Extract storage hint

- **WHEN** a property has `x-storageHint: "indexed"`
- **THEN** the field's `StorageHint = "indexed"`

#### Scenario: Skip relation fields

- **WHEN** a property has `x-relation` custom property
- **THEN** the parser skips this field (does not include in parsed model)
- **AND** logs a warning that relation fields are not supported in schema import

### Requirement: Create Model from Schema

The system SHALL provide a GraphQL mutation to create a new model from JSON Schema input.

#### Scenario: Create model with valid schema

- **WHEN** `createModelFromSchema` mutation is called with valid JSON Schema
- **AND** the cluster and database exist
- **AND** no model with the same name exists in that cluster/database
- **THEN** a new model is created with all fields from the schema
- **AND** system fields (id, createdAt, updatedAt) are added automatically
- **AND** the model is deployed to the customer database
- **AND** the mutation returns the created model with assigned ID

#### Scenario: Reject duplicate model name

- **WHEN** `createModelFromSchema` mutation is called
- **AND** a model with the same `ModelName`, `ClusterName`, and `DatabaseName` already exists
- **THEN** the mutation fails with `ModelAlreadyExists` error
- **AND** no model is created

#### Scenario: Reject invalid cluster

- **WHEN** `createModelFromSchema` mutation is called
- **AND** the `ClusterName` does not exist
- **THEN** the mutation fails with `ClusterNotFound` error

#### Scenario: Reject malformed JSON

- **WHEN** `createModelFromSchema` mutation is called with invalid JSON
- **THEN** the mutation fails with `ParamInvalid` error
- **AND** the error message indicates JSON parsing failure

#### Scenario: Reject invalid schema structure

- **WHEN** `createModelFromSchema` mutation is called
- **AND** the schema is missing required fields (e.g., `title`, `x-modelName`)
- **THEN** the mutation fails with `ParamInvalid` error
- **AND** the error message indicates which required fields are missing

### Requirement: Synchronize Model Schema

The system SHALL provide a GraphQL mutation to synchronize an existing model with an updated JSON Schema, with optional destructive mode.

#### Scenario: Add missing fields during sync (default mode)

- **WHEN** `syncModelSchema` mutation is called with model ID and updated schema
- **AND** `deleteExtraFields` is not provided or is `false`
- **AND** the schema contains 2 new fields not present in the existing model
- **AND** the schema contains 3 fields that already exist in the model
- **THEN** the 2 new fields are added to the model
- **AND** the 2 new fields are deployed to the customer database
- **AND** the 3 existing fields are not modified
- **AND** the mutation returns `fieldsAdded: 2` and `fieldsSkipped: ["field1", "field2", "field3"]`

#### Scenario: Skip fields that already exist

- **WHEN** `syncModelSchema` mutation is called
- **AND** all fields in the schema already exist in the model (by name)
- **THEN** no fields are added
- **AND** the mutation returns `fieldsAdded: 0` and a list of all skipped field names

#### Scenario: Keep extra fields not in schema (default mode)

- **WHEN** `syncModelSchema` mutation is called
- **AND** `deleteExtraFields` is not provided or is `false`
- **AND** the existing model has 5 fields
- **AND** the schema only defines 3 of those fields
- **THEN** no fields are removed from the model
- **AND** all 5 fields remain in the model
- **AND** the sync is non-destructive
- **AND** the mutation returns `fieldsDeleted: 0`

#### Scenario: Delete extra fields in destructive mode

- **WHEN** `syncModelSchema` mutation is called with `deleteExtraFields: true`
- **AND** the existing model has 5 fields: ["id", "name", "age", "email", "phone"]
- **AND** the schema only defines 3 of those fields: ["id", "name", "age"]
- **THEN** fields "email" and "phone" are removed from the model
- **AND** the fields are removed from the customer database (DROP COLUMN)
- **AND** the mutation returns `fieldsDeleted: 2` and `deletedFields: ["email", "phone"]`

#### Scenario: Prevent deletion of system fields

- **WHEN** `syncModelSchema` mutation is called with `deleteExtraFields: true`
- **AND** the schema does not include system fields (id, createdAt, updatedAt)
- **THEN** system fields are NOT deleted
- **AND** the mutation succeeds
- **AND** system fields are excluded from `deletedFields` list

#### Scenario: Prevent deletion of fields with dependencies

- **WHEN** `syncModelSchema` mutation is called with `deleteExtraFields: true`
- **AND** the model has a field "userId" that is referenced by a relation (has `ParentRelationID` set)
- **AND** the schema does not include "userId"
- **THEN** the mutation fails with `OperationDenied` error
- **AND** the error message states "Cannot delete field 'userId' because it has dependencies (referenced by relations)"
- **AND** no fields are deleted

#### Scenario: Combined add and delete in destructive mode

- **WHEN** `syncModelSchema` mutation is called with `deleteExtraFields: true`
- **AND** the schema contains 2 new fields not in the model
- **AND** the schema omits 1 existing field (non-system, non-dependent)
- **AND** the schema contains 3 fields that already exist
- **THEN** the 2 new fields are added
- **AND** the 1 omitted field is deleted
- **AND** the mutation returns `fieldsAdded: 2`, `fieldsDeleted: 1`, `fieldsSkipped: [...]`, `deletedFields: [...]`

#### Scenario: Reject cluster/database mismatch

- **WHEN** `syncModelSchema` mutation is called
- **AND** the schema's `x-clusterName` or `x-databaseName` differs from the existing model
- **THEN** the mutation fails with `ParamInvalid` error
- **AND** the error indicates cluster/database mismatch

#### Scenario: Reject non-existent model

- **WHEN** `syncModelSchema` mutation is called with a non-existent model ID
- **THEN** the mutation fails with `ModelNotFound` error

#### Scenario: Reject conflicting field types

- **WHEN** `syncModelSchema` mutation is called
- **AND** the schema defines a field with the same name as an existing field
- **BUT** the type is different (e.g., existing is STRING, schema is INTEGER)
- **THEN** the mutation fails with `ParamInvalid` error
- **AND** the error indicates the conflicting field name and types

### Requirement: GraphQL API Definitions

The system SHALL define GraphQL types and mutations for schema-based model operations.

#### Scenario: Define CreateModelFromSchemaInput

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `CreateModelFromSchemaInput` with fields:
  - `schema: String!` - JSON Schema Draft 7 document as string
  - `clusterName: String!` - Target cluster name
  - `databaseName: String!` - Target database name

#### Scenario: Define CreateModelFromSchemaPayload

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `CreateModelFromSchemaPayload` with field:
  - `model: Model` - The created model

#### Scenario: Define SyncModelSchemaInput

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `SyncModelSchemaInput` with fields:
  - `id: ID!` - Existing model ID
  - `schema: String!` - Updated JSON Schema Draft 7 document
  - `deleteExtraFields: Boolean` - Optional flag to enable destructive sync (default: false)

#### Scenario: Define SyncModelSchemaPayload

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `SyncModelSchemaPayload` with fields:
  - `model: Model` - The updated model
  - `fieldsAdded: Int!` - Number of fields added
  - `fieldsSkipped: [String!]!` - Names of fields that already existed
  - `fieldsDeleted: Int!` - Number of fields deleted (0 if `deleteExtraFields` is false)
  - `deletedFields: [String!]!` - Names of fields that were deleted

#### Scenario: Define mutations

- **WHEN** the GraphQL schema is defined
- **THEN** the Mutation type includes:
  - `createModelFromSchema(input: CreateModelFromSchemaInput!): CreateModelFromSchemaPayload!`
  - `syncModelSchema(input: SyncModelSchemaInput!): SyncModelSchemaPayload!`

### Requirement: Round-Trip Consistency

The system SHALL ensure bidirectional consistency between JSON Schema export and import operations.

#### Scenario: Export-import produces equivalent model

- **WHEN** a model with 5 fields is exported via `modelJsonSchema` query
- **AND** the exported JSON Schema is used to create a new model via `createModelFromSchema`
- **AND** the new model is exported again via `modelJsonSchema`
- **THEN** the second export produces a JSON Schema equivalent to the first export
- **AND** all field types, validation rules, and metadata are preserved

#### Scenario: Import-export-import produces equivalent model

- **WHEN** a model is created via `createModelFromSchema` with a given JSON Schema
- **AND** the model is exported via `modelJsonSchema`
- **AND** the exported schema is used to create another model
- **THEN** both models have identical field definitions (names, types, validation, metadata)

### Requirement: Error Handling

The system SHALL provide clear, actionable error messages for schema parsing and validation failures.

#### Scenario: Report missing required property

- **WHEN** a schema lacks the `x-modelName` property
- **THEN** the error message states "Required metadata 'x-modelName' is missing from schema"

#### Scenario: Report unsupported type

- **WHEN** a schema property has `type: "null"` or other unsupported type
- **THEN** the error message states "Unsupported JSON Schema type 'null' for field 'fieldName'"

#### Scenario: Report malformed enum

- **WHEN** a schema has `type: "string"` and `enum: []` (empty array)
- **THEN** the error message states "Enum field 'fieldName' must have at least one option"

#### Scenario: Report type conflict during sync

- **WHEN** syncing a model
- **AND** a field name exists with different type (STRING in model, INTEGER in schema)
- **THEN** the error message states "Field 'fieldName' type conflict: existing type STRING, schema type INTEGER"

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

