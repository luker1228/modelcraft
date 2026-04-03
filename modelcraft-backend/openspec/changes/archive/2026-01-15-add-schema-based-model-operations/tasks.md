# Implementation Tasks

## Domain Layer

- [x] Create `JSONSchemaParser` domain service in `internal/domain/modeldesign/`
  - [x] Implement `ParseSchema(schemaJSON string) (*DataModel, error)` method
  - [x] Map JSON Schema `type` to ModelCraft `FieldType.Format`
  - [x] Map JSON Schema `format` to field formats (uuid, date, date-time, time)
  - [x] Parse `required` array to set `FieldDefinition.Required`
  - [x] Parse `nullable` to set `FieldDefinition.NonNull`
  - [x] Map validation keywords (minLength, maxLength, pattern, minimum, maximum, minItems, maxItems)
  - [x] Map custom `x-*` properties (x-displayOrder, x-isPrimary, x-isUnique, x-storageHint)
  - [x] Parse enum options from `enum` and `x-enum` properties
  - [x] Handle nested enum definitions for ENUM and ENUM_ARRAY types
  - [x] Parse decimal precision/scale from `x-precision` and `x-scale`
  - [x] Parse date/time constraints from `x-minDate`, `x-maxDate`, `x-minTime`, `x-maxTime`
  - [x] Extract model metadata from schema root (title, description, x-modelName, x-clusterName, x-databaseName)
  - [x] Validate schema structure (must be Draft 7, must have required fields)
  - [x] Return clear errors for unsupported features (e.g., RELATION fields, $ref, allOf)

- [x] Add unit tests for `JSONSchemaParser`
  - [x] Test basic type mapping (string, integer, number, boolean)
  - [x] Test format mapping (uuid, date, date-time, time)
  - [x] Test validation rule mapping
  - [x] Test enum field parsing (single and multi-select)
  - [x] Test custom property extraction
  - [x] Test error cases (invalid JSON, missing required metadata, unsupported features)
  - [x] Test round-trip consistency (generate → parse → generate yields equivalent output)

## Application Layer

- [x] Add `CreateModelFromSchema` method to `ModelDesignAppService`
  - [x] Accept JSON Schema string as input
  - [x] Parse schema using `JSONSchemaParser`
  - [x] Validate cluster exists
  - [x] Check for model name conflicts
  - [x] Add system fields to parsed model
  - [x] Delegate to existing `transactionDeployModel` for creation
  - [x] Return created model with ID

- [x] Add `SyncModelSchema` method to `ModelDesignAppService`
  - [x] Accept model ID, JSON Schema string, and `deleteExtraFields` boolean as input
  - [x] Fetch existing model by ID (with fields)
  - [x] Parse schema using `JSONSchemaParser`
  - [x] Validate cluster and database match existing model
  - [x] Compare parsed fields with existing fields by name
  - [x] Identify fields in schema that don't exist in model (to add)
  - [x] Delegate to existing `AddFieldSync` for adding missing fields
  - [x] If `deleteExtraFields` is true:
    - [x] Identify fields in model that are not in schema (candidates for deletion)
    - [x] Filter out system fields (id, createdAt, updatedAt) from deletion candidates
    - [x] Validate no field has dependencies (check `ParentRelationID`)
    - [x] Delegate to existing `RemoveFieldSync` for each field to delete
  - [x] Return sync summary (fields added, fields skipped, fields deleted)

- [x] Add integration tests
  - [x] Test create model from valid schema
  - [x] Test sync adds missing fields (default mode)
  - [x] Test sync skips existing fields
  - [x] Test sync with `deleteExtraFields: false` keeps extra fields
  - [x] Test sync with `deleteExtraFields: true` removes extra fields
  - [x] Test sync prevents deletion of system fields
  - [x] Test sync prevents deletion of fields with dependencies
  - [x] Test combined add and delete in one sync operation
  - [x] Test error handling (invalid schema, cluster mismatch, etc.)

## Interface Layer - GraphQL

- [x] Update GraphQL schema in `api/graph/schema/model.graphql`
  - [x] Add `CreateModelFromSchemaInput` with `schema: String!`, `clusterName: String!`, `databaseName: String!`
  - [x] Add `CreateModelFromSchemaPayload` with `model: Model`
  - [x] Add `SyncModelSchemaInput` with `id: ID!`, `schema: String!`, `deleteExtraFields: Boolean`
  - [x] Add `SyncModelSchemaPayload` with `model: Model`, `fieldsAdded: Int!`, `fieldsSkipped: [String!]!`, `fieldsDeleted: Int!`, `deletedFields: [String!]!`
  - [x] Add mutations to Mutation type

- [x] Regenerate GraphQL code with `make generate-gql`

- [x] Implement resolvers in `internal/interfaces/graphql/model.resolvers.go`
  - [x] Implement `createModelFromSchema` mutation resolver
  - [x] Implement `syncModelSchema` mutation resolver
  - [x] Convert GraphQL inputs to application service parameters
  - [x] Handle errors and return appropriate GraphQL errors

## Request/Response DTOs

- [x] Add `CreateModelFromSchemaRequest` in `internal/interfaces/http/requests/`
  - [x] Fields: `Schema string`, `ClusterName string`, `DatabaseName string`
  - [x] Validation tags

- [x] Add `SyncModelSchemaRequest` in `internal/interfaces/http/requests/`
  - [x] Fields: `ModelID string`, `Schema string`, `DeleteExtraFields bool`
  - [x] Validation tags

- [x] Add response types if needed (may reuse existing types)

## Documentation

- [x] Update `docs/02-design/` with schema-based model operations documentation
- [x] Add examples of JSON Schema input format
- [x] Document limitations (no relations, additive sync only)
- [x] Add to API documentation

## Testing and Validation

- [x] Run all Go unit tests: `make test`
- [x] Run integration tests with pytest
- [x] Test round-trip workflow: export → modify → import
- [x] Test error scenarios with invalid schemas
- [x] Verify field type mappings are bidirectional and consistent

## Deployment Checklist

- [x] All tests passing
- [x] Code formatted: `make fmt`
- [x] Vet checks pass: `make vet`
- [x] GraphQL schema regenerated
- [x] Documentation updated
- [x] Manual testing completed on dev environment
