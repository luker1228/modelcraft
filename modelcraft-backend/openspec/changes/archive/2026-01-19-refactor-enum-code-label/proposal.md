# Proposal: Refactor Enum Feature - Code/Label Rename and Relation Support

## Change ID
`refactor-enum-code-label`

## Summary
Refactor the enum feature to:
1. Rename `key`/`value` fields to `code`/`label` for enum options to use more precise industry-standard terminology
2. Add support for enum relation fields that allow querying full enum details (including label) alongside the stored code

## Motivation

### Why rename key/value to code/label?
- **Industry Standard**: Most database frameworks (Prisma, Sequelize, TypeORM, etc.) use `code`/`label` terminology for enums
- **Semantic Clarity**: `code` better represents the stored identifier (database value), while `label` clearly indicates the display value
- **API Consistency**: Aligns with common API patterns where `code` is the programmatic identifier and `label` is the human-readable text

### Why add enum relation support?
Currently, enum field data queries only return the stored code (key), requiring clients to lookup enum definitions separately to get display values. This proposal adds:
- A new field relation type for enum fields that returns full enum option details
- Backward compatibility - existing behavior is preserved
- Performance optimization - single query returns both code and label information

## Goals
1. Rename enum option fields from `key`/`value` to `code`/`label` across all layers
2. Provide backward compatibility for existing data and APIs
3. Add a new enum relation field type that returns `EnumOptionLabel` with code, label, and description
4. Maintain existing enum functionality without breaking changes in public APIs

## Non-Goals
- Changing the database schema for existing enum data (will use migration to handle)
- Modifying the core enum definition storage structure (only renaming fields in JSON payload)
- Changing the physical column type for enum fields (stores code as string)
- Adding i18n support for enum labels (out of scope for this change)

## Scope

### In Scope
1. **Domain Layer**: Update `EnumOption` struct to use `Code`/`Label` instead of `Key`/`Value`
2. **GraphQL Schema**: Update all related GraphQL types and inputs
3. **DTOs**: Update all DTOs that use enum options
4. **Mappers**: Update serialization/deserialization logic
5. **Tests**: Update all test files referencing enum options
6. **Runtime Query Support**: Add enum label relation field queries in runtime GraphQL
7. **Migration**: Create database migration to update stored JSON data

### Out of Scope
- Changing enum definition ID structure
- Modifying project-scoped enum behavior
- Changing enum association table structure
- Backward compatibility layer in API (clients will need to update to use new field names)

## Risks and Mitigations

### Risk 1: Breaking existing client code
**Mitigation**: This is a breaking change by design. We will:
- Document the migration path clearly
- Consider a deprecation period if needed
- Provide clear error messages for deprecated field names

### Risk 2: Data migration complexity
**Mitigation**:
- Write idempotent migration logic
- Test migration on copy of production data
- Support rollback plan

### Risk 3: Runtime query performance impact
**Mitigation**:
- The enum label relation is optional (only queried when requested)
- Implement efficient caching for enum definitions
- Use JOIN queries where appropriate

## Dependencies
- None - this is an internal refactoring

## Success Criteria
1. All enum option references use `code`/`label` terminology
2. Existing data is successfully migrated
3. Runtime queries can fetch enum labels alongside codes
4. All tests pass with updated terminology
5. GraphQL schema validation succeeds
6. Integration tests pass with enum label queries

## Implementation Phase
This proposal will be implemented in two sequential phases:

### Phase 1: Code/Label Rename
- Update domain models with new field names
- Update GraphQL schema and generated code
- Update DTOs and mappers
- Create data migration
- Update all tests

### Phase 2: Enum Relation Support
- Add enum relation field type to runtime schema generator
- Implement query logic for fetching enum labels
- Add integration tests for enum label queries
- Update documentation

## References
- Existing enum implementation: `internal/domain/modeldesign/enum_definition.go`
- Current GraphQL schema: `api/graph/schema/enum.graphql`
- Runtime schema generation: `internal/domain/modelruntime/graphqlschema_manager.go`
- Related spec: `openspec/specs/modeldesign-field-types/spec.md`
