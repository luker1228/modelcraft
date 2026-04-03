# Implementation Tasks

## 1. Domain Layer - JSON Schema Generator

- [x] 1.1 Create `JSONSchemaGenerator` service in `internal/domain/modeldesign/jsonschema_generator.go`
- [x] 1.2 Implement field type mapping (STRING → string, NUMBER → number/integer, BOOLEAN → boolean, etc.)
- [x] 1.3 Implement format mapping (DATE → date, DATETIME → date-time, UUID → uuid, etc.)
- [x] 1.4 Implement validation rule conversion (ValidationConfig → JSON Schema validation keywords)
- [x] 1.5 Implement enum field handling (include enum options as JSON Schema enum keyword)
- [x] 1.6 Add custom `x-` properties for ModelCraft-specific metadata (x-relation, x-storageHint, x-displayOrder, etc.)
- [x] 1.7 Handle required and nullable field attributes (required array, nullable keyword)

## 2. GraphQL API Layer

- [x] 2.1 Add `modelJsonSchema` query to `api/graph/schema/model.graphql`
- [x] 2.2 Define `ModelJsonSchema` output type with `schema` field (JSON string or structured object)
- [x] 2.3 Implement resolver in `internal/interfaces/graphql/model.resolvers.go`
- [x] 2.4 Wire up domain service dependency in resolver

## 3. Testing

- [x] 3.1 Add unit tests for `JSONSchemaGenerator` covering all field types
- [x] 3.2 Add unit tests for validation rule conversion
- [x] 3.3 Add unit tests for enum field handling
- [x] 3.4 Add integration test for GraphQL query execution
- [x] 3.5 Test with models containing various field types and configurations

## 4. Documentation

- [x] 4.1 Update GraphQL API documentation with `modelJsonSchema` query example
- [x] 4.2 Document JSON Schema mapping rules (field types, validation, custom properties)
- [x] 4.3 Add example JSON Schema output for reference
