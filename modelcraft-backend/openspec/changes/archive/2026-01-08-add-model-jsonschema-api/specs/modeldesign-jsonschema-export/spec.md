# modeldesign-jsonschema-export Specification

## ADDED Requirements

### Requirement: GraphQL Query for JSON Schema Export

The system SHALL provide a GraphQL query that returns a model's field definitions as JSON Schema Draft 7 format.

#### Scenario: Query JSON Schema for valid model

- **WHEN** a GraphQL query `modelJsonSchema(id: "model-123")` is executed
- **AND** the model exists
- **THEN** a `ModelJsonSchema` object is returned with `modelId`, `modelName`, and `schema` fields
- **AND** the `schema` field contains a valid JSON Schema Draft 7 document as a JSON-encoded string

#### Scenario: Query JSON Schema for non-existent model

- **WHEN** a GraphQL query `modelJsonSchema(id: "non-existent")` is executed
- **AND** the model does not exist
- **THEN** a null result is returned
- **AND** no error is thrown (following GraphQL conventions for missing resources)

#### Scenario: Query JSON Schema with invalid ID format

- **WHEN** a GraphQL query `modelJsonSchema(id: "")` is executed with an empty or invalid ID
- **THEN** a GraphQL validation error is returned
- **AND** the error indicates the ID is invalid

### Requirement: JSON Schema Structure

The system SHALL generate JSON Schema Draft 7 documents with standard structure for model field definitions.

#### Scenario: Schema contains required metadata

- **WHEN** generating JSON Schema for a model
- **THEN** the schema includes `$schema: "http://json-schema.org/draft-07/schema#"` field
- **AND** the schema includes `type: "object"` at the root level
- **AND** the schema includes `title` field set to the model title
- **AND** the schema includes `description` field set to the model description

#### Scenario: Schema contains properties for all fields

- **WHEN** generating JSON Schema for a model with 5 fields
- **THEN** the schema includes a `properties` object with 5 entries
- **AND** each entry key matches the field name
- **AND** each entry value is a JSON Schema property definition

#### Scenario: Schema includes required fields array

- **WHEN** generating JSON Schema for a model with 3 required fields and 2 optional fields
- **THEN** the schema includes a `required` array at the root level
- **AND** the `required` array contains exactly 3 field names (those with `Required == true`)
- **AND** optional fields (those with `Required == false`) are not in the `required` array

### Requirement: Field Type Mapping

The system SHALL map ModelCraft field types to corresponding JSON Schema types and formats.

#### Scenario: Map STRING field type

- **WHEN** a field has type `STRING`
- **THEN** the JSON Schema property has `type: "string"`
- **AND** no `format` field is set

#### Scenario: Map UUID field type

- **WHEN** a field has type `UUID`
- **THEN** the JSON Schema property has `type: "string"` and `format: "uuid"`

#### Scenario: Map DATE field type

- **WHEN** a field has type `DATE`
- **THEN** the JSON Schema property has `type: "string"` and `format: "date"`
- **AND** the description indicates the expected format is ISO 8601 YYYY-MM-DD

#### Scenario: Map DATETIME field type

- **WHEN** a field has type `DATETIME`
- **THEN** the JSON Schema property has `type: "string"` and `format: "date-time"`
- **AND** the description indicates the expected format is ISO 8601 with timezone

#### Scenario: Map TIME field type

- **WHEN** a field has type `TIME`
- **THEN** the JSON Schema property has `type: "string"` and `format: "time"`
- **AND** the description indicates the expected format is HH:MM:SS

#### Scenario: Map NUMBER field type

- **WHEN** a field has type `NUMBER`
- **THEN** the JSON Schema property has `type: "number"`

#### Scenario: Map INTEGER field type

- **WHEN** a field has type `INTEGER`
- **THEN** the JSON Schema property has `type: "integer"`

#### Scenario: Map DECIMAL field type

- **WHEN** a field has type `DECIMAL` with precision 10 and scale 2
- **THEN** the JSON Schema property has `type: "number"`
- **AND** the property includes `x-precision: 10` and `x-scale: 2` custom properties

#### Scenario: Map BOOLEAN field type

- **WHEN** a field has type `BOOLEAN`
- **THEN** the JSON Schema property has `type: "boolean"`

### Requirement: Validation Rule Mapping

The system SHALL convert ModelCraft ValidationConfig rules to JSON Schema validation keywords.

#### Scenario: Map string length constraints

- **WHEN** a STRING field has `ValidationConfig.MinLength = 5` and `ValidationConfig.MaxLength = 100`
- **THEN** the JSON Schema property includes `minLength: 5` and `maxLength: 100`

#### Scenario: Map pattern constraint

- **WHEN** a STRING field has `ValidationConfig.Pattern = "^[A-Z][a-z]+$"`
- **THEN** the JSON Schema property includes `pattern: "^[A-Z][a-z]+$"`

#### Scenario: Map numeric constraints

- **WHEN** a NUMBER field has `ValidationConfig.Minimum = 0` and `ValidationConfig.Maximum = 100`
- **THEN** the JSON Schema property includes `minimum: 0` and `maximum: 100`

#### Scenario: Map array item constraints

- **WHEN** an ENUM_ARRAY field has `ValidationConfig.MinItems = 1` and `ValidationConfig.MaxItems = 5`
- **THEN** the JSON Schema property includes `minItems: 1` and `maxItems: 5`

#### Scenario: Map date range constraints

- **WHEN** a DATE field has `ValidationConfig.MinDate = "2024-01-01"` and `ValidationConfig.MaxDate = "2024-12-31"`
- **THEN** the JSON Schema property includes `x-minDate: "2024-01-01"` and `x-maxDate: "2024-12-31"` custom properties
- **AND** the description indicates the date range constraint

#### Scenario: Map time range constraints

- **WHEN** a TIME field has `ValidationConfig.MinTime = "09:00:00"` and `ValidationConfig.MaxTime = "18:00:00"`
- **THEN** the JSON Schema property includes `x-minTime: "09:00:00"` and `x-maxTime: "18:00:00"` custom properties
- **AND** the description indicates the time range constraint

### Requirement: Enum Field Handling

The system SHALL include enum options in JSON Schema for ENUM and ENUM_ARRAY field types.

#### Scenario: Map single-select enum field

- **WHEN** a field has type `ENUM` with enum options `[{key: "active", value: "Active"}, {key: "inactive", value: "Inactive"}]`
- **THEN** the JSON Schema property has `type: "string"`
- **AND** the property includes `enum: ["active", "inactive"]` (using option keys)
- **AND** the property includes `x-enum` custom property with full enum definition including keys, values, and descriptions

#### Scenario: Map multi-select enum field

- **WHEN** a field has type `ENUM_ARRAY` with enum options `[{key: "read", value: "Read"}, {key: "write", value: "Write"}]`
- **THEN** the JSON Schema property has `type: "array"`
- **AND** the property includes `items: { type: "string", enum: ["read", "write"] }`
- **AND** the property includes `x-enum` custom property with full enum definition

#### Scenario: Enum field with enum definition reference

- **WHEN** a field has type `ENUM` and references an enum definition by name
- **THEN** the JSON Schema property includes the enum options from the referenced enum definition
- **AND** the `x-enum` custom property includes the enum definition name and full details

### Requirement: Nullable Field Handling

The system SHALL indicate nullable fields using the JSON Schema nullable keyword.

#### Scenario: Non-nullable field

- **WHEN** a field has `NonNull = true`
- **THEN** the JSON Schema property does not include a `nullable` keyword (defaults to non-nullable)

#### Scenario: Nullable field

- **WHEN** a field has `NonNull = false`
- **THEN** the JSON Schema property includes `nullable: true`

### Requirement: Custom Metadata Properties

The system SHALL include ModelCraft-specific metadata as custom `x-` prefixed properties to avoid conflicts with standard JSON Schema keywords.

#### Scenario: Include display order

- **WHEN** generating JSON Schema for a field with `DisplayOrder = 10`
- **THEN** the JSON Schema property includes `x-displayOrder: 10`

#### Scenario: Include primary key flag

- **WHEN** generating JSON Schema for a field with `IsPrimary = true`
- **THEN** the JSON Schema property includes `x-isPrimary: true`

#### Scenario: Include unique constraint flag

- **WHEN** generating JSON Schema for a field with `IsUnique = true`
- **THEN** the JSON Schema property includes `x-isUnique: true`

#### Scenario: Include storage hint

- **WHEN** generating JSON Schema for a field with `StorageHint = "indexed"`
- **THEN** the JSON Schema property includes `x-storageHint: "indexed"`

#### Scenario: Include relation metadata for RELATION fields

- **WHEN** generating JSON Schema for a field with type `RELATION`
- **THEN** the JSON Schema property has `type: "object"`
- **AND** the property includes `x-relation` custom property with relation configuration (reference model, relation type, etc.)
- **AND** the property includes `x-relationId` with the relation ID

#### Scenario: Include model metadata

- **WHEN** generating JSON Schema for any field
- **THEN** the JSON Schema property includes `x-modelId` with the source model ID
- **AND** the property includes `x-modelName` with the source model name
- **AND** the property includes `x-fieldName` with the field name

### Requirement: GraphQL Type Definitions

The system SHALL define GraphQL types for JSON Schema export in the GraphQL schema.

#### Scenario: Define ModelJsonSchema output type

- **WHEN** the GraphQL schema is defined
- **THEN** it includes a `ModelJsonSchema` type with the following fields:
  - `modelId: ID!` - The model's unique identifier
  - `modelName: String!` - The model's name
  - `schema: String!` - The JSON-encoded JSON Schema Draft 7 document

#### Scenario: Define modelJsonSchema query

- **WHEN** the GraphQL schema is defined
- **THEN** it includes a query `modelJsonSchema(id: ID!): ModelJsonSchema` in the Query type
- **AND** the query accepts a required `id` parameter of type ID
- **AND** the query returns a nullable `ModelJsonSchema` result

### Requirement: Error Handling

The system SHALL handle errors gracefully when generating JSON Schema.

#### Scenario: Handle model not found

- **WHEN** generating JSON Schema for a non-existent model ID
- **THEN** the query returns null (not an error)
- **AND** no internal error is logged

#### Scenario: Handle missing field type

- **WHEN** generating JSON Schema for a field without a type
- **THEN** an internal error is logged
- **AND** the field is skipped in the generated schema
- **AND** the overall schema generation continues for other fields

#### Scenario: Handle invalid validation config

- **WHEN** generating JSON Schema for a field with invalid validation configuration (e.g., minLength > maxLength)
- **THEN** an internal warning is logged
- **AND** the invalid constraint is omitted from the JSON Schema
- **AND** valid constraints are still included
