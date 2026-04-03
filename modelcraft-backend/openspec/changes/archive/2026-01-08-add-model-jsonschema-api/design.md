# Design: JSON Schema Export for Models

## Context

ModelCraft stores model field definitions with rich type information, validation rules, and metadata. External systems often expect JSON Schema format for form generation, validation, and documentation. We need a standard way to export ModelCraft models as JSON Schema Draft 7.

**Key Constraints:**
- JSON Schema Draft 7 is the widely supported version
- ModelCraft has field types and validation rules that don't have direct JSON Schema equivalents
- Custom metadata (relations, storage hints) must not conflict with standard JSON Schema keywords

**Stakeholders:**
- Frontend developers using form generators (React Hook Form, Formik, etc.)
- API consumers integrating with external validation libraries
- Documentation tooling (Swagger UI, Redoc, etc.)

## Goals / Non-Goals

**Goals:**
- Export complete model field definitions as valid JSON Schema Draft 7
- Map all ModelCraft field types to appropriate JSON Schema types and formats
- Include validation rules as JSON Schema validation keywords
- Preserve ModelCraft-specific metadata using `x-` custom properties
- Support enum fields with full option lists

**Non-Goals:**
- Support for JSON Schema Draft 2019-09 or 2020-12 (use Draft 7 for maximum compatibility)
- Bidirectional conversion (JSON Schema → ModelCraft model)
- Schema validation or linting (assume caller validates if needed)
- Generating example data or default values (can be added later)

## Decisions

### Decision 1: Use GraphQL API

**What:** Provide JSON Schema export via GraphQL query `modelJsonSchema(id: ID!): ModelJsonSchema`

**Why:**
- Consistent with existing model query patterns (`model`, `modelByName`)
- Allows clients to request schema as part of larger queries
- Natural fit for design-time operations (models are design-time entities)

**Alternatives Considered:**
- HTTP REST endpoint (`GET /api/models/:id/json-schema`): More direct but introduces inconsistency with GraphQL-first API design
- Runtime API: JSON Schema is primarily for design/documentation, not runtime execution

### Decision 2: Return JSON Schema as String

**What:** The `ModelJsonSchema` type contains a `schema: String!` field with JSON-encoded schema

**Why:**
- JSON Schema is inherently a JSON document, not a GraphQL type
- Allows arbitrary JSON Schema extensions without GraphQL schema changes
- Standard practice for returning JSON documents in GraphQL (see GitHub GraphQL API)

**Example:**
```graphql
type ModelJsonSchema {
  modelId: ID!
  modelName: String!
  schema: String!  # JSON-encoded JSON Schema
}
```

**Alternatives Considered:**
- Structured GraphQL type mirroring JSON Schema: Too rigid, requires GraphQL schema changes for JSON Schema spec updates
- Custom scalar type `JSONSchema`: Adds complexity without clear benefit

### Decision 3: Map Field Types Using JSON Schema `type` and `format`

**Mapping Table:**

| ModelCraft Type | JSON Schema Type | JSON Schema Format | Notes |
|----------------|------------------|-------------------|-------|
| STRING | string | - | |
| UUID | string | uuid | |
| DATE | string | date | ISO 8601 YYYY-MM-DD |
| DATETIME | string | date-time | ISO 8601 with timezone |
| TIME | string | time | HH:MM:SS |
| NUMBER | number | - | |
| INTEGER | integer | - | JSON Schema integer type |
| DECIMAL | number | - | Use custom `x-precision` and `x-scale` |
| BOOLEAN | boolean | - | |
| ENUM | string | - | Use `enum` keyword with option keys |
| ENUM_ARRAY | array | - | `items: { type: string, enum: [...] }` |
| RELATION | object | - | Use `x-relation` for metadata |

**Why:** Follows JSON Schema Draft 7 specification closely while preserving ModelCraft semantics

### Decision 4: Use `x-` Prefix for ModelCraft-Specific Properties

**Custom Properties:**
- `x-relation`: Relation configuration (for RELATION fields)
- `x-storageHint`: Database storage optimization hint
- `x-displayOrder`: Field display order in UI
- `x-isPrimary`: Whether field is primary key
- `x-isUnique`: Whether field has unique constraint
- `x-modelId`: Source model ID, don't show this property ?
    - why? id don't like name, it has random properties, but name is human readable and don't change when export/import
- `x-modelName`: Source model name
- `x-fieldName`: Field name (redundant but useful for reference)
- `x-enum`: Full enum definition with keys, values, descriptions (for ENUM/ENUM_ARRAY fields)

**Why:**
- JSON Schema spec reserves `x-` prefix for custom extensions
- Prevents conflicts with future JSON Schema keywords
- Standard practice in OpenAPI/Swagger specifications

### Decision 5: Map ValidationConfig to Standard JSON Schema Keywords

**Mapping:**

| ValidationConfig | JSON Schema | Notes |
|-----------------|-------------|-------|
| MaxLength | maxLength | |
| MinLength | minLength | |
| Pattern | pattern | |
| Maximum | maximum | |
| Minimum | minimum | |
| MaxItems | maxItems | Array fields |
| MinItems | minItems | Array fields |
| MinDate, MaxDate | format + custom constraint | Use `x-minDate`, `x-maxDate` |
| MinTime, MaxTime | format + custom constraint | Use `x-minTime`, `x-maxTime` |
| Precision, Scale | custom constraint | Use `x-precision`, `x-scale` |
| EnumValues | enum | |

**Why:** Maximizes compatibility with standard JSON Schema validators while preserving all validation rules

### Decision 6: Handle Required and Nullable Separately

**Implementation:**
- `required` array at schema level: Lists field names where `FieldDefinition.Required == true`
- `nullable` keyword at field level: Set to `true` when `FieldDefinition.NonNull == false`

**Example:**
```json
{
  "type": "object",
  "required": ["id", "name"],
  "properties": {
    "id": { "type": "string", "format": "uuid" },
    "name": { "type": "string" },
    "description": { "type": "string", "nullable": true }
  }
}
```

**Why:** Follows JSON Schema best practices and OpenAPI 3.0 nullable convention

## Risks / Trade-offs

### Risk: JSON Schema Spec Evolution
**Description:** JSON Schema Draft 8 (2019-09) and later versions have breaking changes

**Mitigation:**
- Target Draft 7 for maximum compatibility (most widely supported)
- Add `$schema` field to indicate version explicitly
- Document upgrade path in code comments

### Risk: Enum Key vs Value Confusion
**Description:** ModelCraft enums store keys but display values; JSON Schema `enum` expects display values

**Mitigation:**
- Store enum option keys in JSON Schema `enum` keyword (matches database storage)
- Include full enum definition in `x-enum` custom property with both keys and values
- Document this design decision clearly

### Risk: Relation Field Complexity
**Description:** RELATION fields have complex metadata that doesn't map to standard JSON Schema

**Mitigation:**
- Use `type: object` as base type
- Store full relation config in `x-relation` custom property
- Consider adding `$ref` to related model schema in future iteration

### Trade-off: String vs Structured Return Type
**Decision:** Return JSON Schema as string instead of structured GraphQL type

**Pros:**
- Flexibility for JSON Schema spec changes
- Allows arbitrary extensions
- Standard practice

**Cons:**
- Clients must parse JSON string
- No GraphQL-level validation of schema structure

**Accepted:** Flexibility outweighs convenience; clients can use JSON Schema libraries

## Migration Plan

**Phase 1: Initial Implementation**
1. Implement core JSON Schema generator
2. Add GraphQL query
3. Test with existing models

**Phase 2: Iteration**
1. Gather feedback from early users
2. Refine mappings based on real-world usage
3. Add optional parameters (e.g., `includeMetadata: Boolean`)

**No Breaking Changes:** This is a new feature; no migration needed

**Rollback Plan:** Remove GraphQL query and domain service if unused after 1 month

## Open Questions

- **Q:** Should we support generating schemas for nested relation fields recursively?
  **A:** Not in initial version; add `x-relation` metadata only. Can add `$ref` support later.

- **Q:** Should we generate `examples` or `default` values?
  **A:** Not in initial version; can be added as optional feature later.

- **Q:** Should we support filtering fields (e.g., only required fields)?
  **A:** Not in initial version; clients can filter JSON Schema on their side.
