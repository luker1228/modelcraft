# Change: Add JSON Schema Export API for Models

## Why

Users need a standard way to export model field definitions as JSON Schema format for integration with external tools, form generators, validation libraries, and documentation systems. JSON Schema is a widely adopted standard that enables seamless interoperability between ModelCraft and third-party systems.

## What Changes

- Add a new GraphQL query `modelJsonSchema` that accepts a model ID and returns the model's field definitions converted to JSON Schema Draft 7 format
- Implement a JSON Schema generator in the domain layer that converts `FieldDefinition` entities to JSON Schema properties
- Map ModelCraft field types (STRING, NUMBER, BOOLEAN, DATE, DATETIME, etc.) to corresponding JSON Schema types and formats
- Include validation rules from `ValidationConfig` as JSON Schema validation keywords (minLength, maxLength, pattern, minimum, maximum, etc.)
- Represent enum fields with their full option lists in the JSON Schema
- Use `x-` prefixed custom properties for ModelCraft-specific metadata (relation info, storage hints, etc.) to avoid conflicts with standard JSON Schema keywords

## Impact

- **Affected specs**: Creates new capability `modeldesign-jsonschema-export`
- **Affected code**:
  - New domain service: `internal/domain/modeldesign/jsonschema_generator.go`
  - New GraphQL type and resolver: `api/graph/schema/model.graphql`, `internal/interfaces/graphql/model.resolvers.go`
  - Related entities: `internal/domain/modeldesign/model.go`, `internal/domain/modeldesign/field_definition.go`
