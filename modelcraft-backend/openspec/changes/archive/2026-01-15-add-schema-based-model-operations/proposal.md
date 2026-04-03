# Add Schema-Based Model Operations

## Why

ModelCraft currently supports exporting models as JSON Schema via the `modelJsonSchema` GraphQL query, but lacks the inverse operation to import or create models from JSON Schema. This asymmetry creates friction in several user workflows:

1. **No round-trip workflow**: Users can export models but cannot import them back after modification
2. **Manual field creation**: Users must construct models field-by-field through the API instead of providing a declarative schema
3. **Limited model migration**: Moving models between environments requires manual recreation
4. **No schema-first development**: Users cannot design models in external schema editors and import them

Adding schema-based model operations completes the import/export cycle and enables declarative model management.

## Overview

Enable model creation and synchronization using JSON Schema format as input. This allows users to define models declaratively using the same JSON Schema Draft 7 format that the system already exports, creating a symmetric import/export workflow.

## Problem Statement

Currently, ModelCraft supports:
- **Export**: Models can be exported to JSON Schema via the `modelJsonSchema` GraphQL query
- **Import from DB**: Models can be reverse-engineered from existing database tables via DDL or introspection

However, there is no direct way to:
- Create a model by providing a JSON Schema definition
- Synchronize an existing model's structure with an updated JSON Schema

This creates an asymmetry in the workflow and requires users to manually construct model definitions field-by-field through the existing API.

## Proposed Solution

Add two new GraphQL mutations:

1. **`createModelFromSchema`** - Create a new model from JSON Schema Draft 7
2. **`syncModelSchema`** - Synchronize an existing model with a JSON Schema
   - **Default behavior (safe)**: Add missing fields, keep extra fields not in schema
   - **Optional destructive mode**: When `deleteExtraFields: true`, remove fields not present in schema

These operations will:
- Parse JSON Schema Draft 7 documents (same format exported by `modelJsonSchema`)
- Map JSON Schema types and validation rules to ModelCraft field definitions
- Support both creation and incremental synchronization workflows
- Support optional destructive sync with explicit opt-in flag
- Preserve ModelCraft-specific metadata in `x-*` custom properties

## User Scenarios

### Scenario 1: Import/Export Round-Trip
A user exports a model as JSON Schema, modifies it (adds fields, changes validation), and imports it back to update the model.

### Scenario 2: Schema-First Development
A developer designs data models in JSON Schema format first (using any schema editor/tool), then creates the models in ModelCraft from those schemas.

### Scenario 3: Model Migration
A user needs to copy model structure from one environment to another by exporting as JSON Schema and importing in the target environment.

### Scenario 4: Incremental Schema Updates
A user maintains a "master" JSON Schema file for a model and periodically syncs it to add new fields without disrupting existing data or removing fields.

### Scenario 5: Complete Schema Replacement
A user needs to completely replace a model's structure to match a canonical schema definition, removing fields that are no longer needed. They use `syncModelSchema` with `deleteExtraFields: true` to perform a destructive sync.

## Scope

### In Scope
- Parse JSON Schema Draft 7 format
- Create new models from complete schema definitions
- Sync existing models with two modes:
  - **Additive mode (default)**: Add missing fields, keep extra fields
  - **Destructive mode (opt-in)**: Add missing fields, delete fields not in schema
- Map all supported field types (STRING, UUID, DATE, DATETIME, TIME, NUMBER, INTEGER, DECIMAL, BOOLEAN, ENUM, ENUM_ARRAY)
- Map validation rules (length, pattern, min/max, date/time ranges, array constraints)
- Handle required fields and nullable fields
- Process custom `x-*` properties (displayOrder, isPrimary, isUnique, storageHint, etc.)
- GraphQL API mutations for both operations

### Out of Scope
- Support for RELATION fields in schema (relations require referential integrity checks)
- Support for non-Draft 7 JSON Schema versions
- Automatic type coercion/migration for existing data
- REST API endpoints (GraphQL only in this change)
- Batch operations (one model at a time)

## Technical Approach

### Architecture
1. **Domain Layer**: `JSONSchemaParser` service to convert JSON Schema → `DataModel`
   - Mirrors the existing `JSONSchemaGenerator` (export direction)
   - Validates schema structure and required metadata
   - Maps types and validation rules bidirectionally

2. **Application Layer**: New use cases in `ModelDesignAppService`
   - `CreateModelFromSchema(ctx, schema)` - delegates to existing `CreateModelSync`
   - `SyncModelSchema(ctx, modelID, schema, deleteExtraFields)` - compares fields and:
     - Adds fields in schema but not in model
     - Optionally deletes fields in model but not in schema (when `deleteExtraFields=true`)

3. **Interface Layer**: New GraphQL mutations and resolvers

### Key Design Decisions
- **Reuse existing domain logic**: Schema parser produces standard `DataModel` entities, which flow through existing validation and deployment logic
- **Symmetric with export**: Use the exact same JSON Schema structure that `modelJsonSchema` produces
- **Safe by default**: Sync is additive by default; field deletion requires explicit opt-in via `deleteExtraFields: true`
- **Protection against accidental data loss**: Destructive sync requires explicit boolean flag
- **Validation-first**: Schema parsing fails fast with clear errors before any database changes

## Impact Analysis

### Dependencies
- Extends existing `modeldesign` domain (add parser alongside generator)
- Uses existing model creation and field addition infrastructure
- No changes to database schema or deployment logic

### Breaking Changes
None. This is additive functionality only.

### Migration Requirements
None. Existing models and workflows are unchanged.

### Testing Requirements
- Unit tests for JSON Schema parsing and type mapping
- Integration tests for create and sync operations
- Round-trip tests (export → modify → import)
- Error handling tests (invalid schemas, missing required fields, type mismatches)

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Schema parsing errors causing data loss | High | Parse and validate completely before any DB operations; use transactions |
| Type mapping mismatches breaking existing data | Medium | Only allow creation/addition of fields, never modification of existing field types |
| Accidental field deletion destroying data | High | Require explicit `deleteExtraFields: true` flag; document risks clearly; default to safe mode |
| Deleting fields with dependencies (relations) | High | Validate field dependencies before deletion; reject if field is referenced by relations |
| Performance issues with large schemas | Low | Schema parsing is lightweight; actual field creation uses existing optimized code |

## Success Criteria

- User can export a model via `modelJsonSchema`, modify the JSON, and create a new model from it
- User can sync a model by providing an updated schema, and only missing fields are added (default behavior)
- User can perform destructive sync with `deleteExtraFields: true` to remove fields not in schema
- System prevents deletion of fields with dependencies (relation fields)
- All field types and validation rules round-trip correctly (export → import → export yields equivalent schema)
- Clear error messages for invalid schemas or conflicting field definitions
- Integration tests verify end-to-end workflows including both sync modes
