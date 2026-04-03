# Field Enum Format Refactor Design

**Date:** 2026-03-28
**Status:** Draft
**Author:** CodeBuddy

---

## 1. Overview

Refactor the field enum type system from a two-format approach (`ENUM` / `ENUM_ARRAY`) to a unified `FormatEnum` with an `isArray` boolean flag.

**Motivation:**
- Eliminates redundant `FormatEnumArray` constant and `ENUM_ARRAY` GraphQL enum value
- Consolidates logic that currently branches on `FormatEnum` vs `FormatEnumArray`
- Aligns field representation with `EnumDefinition.IsMultiSelect` as the single source of truth

---

## 2. Domain Model Changes

### 2.1 FieldDefinition

**File:** `internal/domain/modeldesign/field_definition.go`

**Add `IsArray` field:**
```go
type FieldDefinition struct {
    // ... existing fields ...

    Type     *FieldType // FormatType = ENUM (single source of truth for enum format)
    IsArray  bool      // true = multi-select, false = single-select
}
```

**Remove `FormatEnumArray` constant:**
```go
// REMOVE
FormatEnumArray FormatType = "ENUM_ARRAY"
```

**Update helper methods:**
```go
func (fd *FieldDefinition) IsEnumField() bool {
    return fd.Type.Format == FormatEnum
}

func (fd *FieldDefinition) IsEnumArrayField() bool {
    return fd.Type.Format == FormatEnum && fd.IsArray
}
```

### 2.2 EnumDefinition

**File:** `internal/domain/modeldesign/enum_definition.go`

**Add code length validation (max 64 chars) at enum option creation:**
- When `EnumOption.Code` is set, validate `len(code) <= 64`
- This enforces the `VARCHAR(64)` storage constraint at creation time

**`IsMultiSelect` remains the source of truth** for determining if an enum supports multi-select values.

---

## 3. GraphQL API Changes

### 3.1 FormatType Enum

**File:** `api/graph/project/schema/field.graphql`

**Remove:**
```graphql
# REMOVE from FormatType enum
ENUM_ARRAY  # 多选枚举
```

### 3.2 Field Type

**Add `isArray` to Field output type:**
```graphql
type Field {
  # ... existing fields ...
  isArray: Boolean!  # true for multi-select enum fields
}
```

### 3.3 AddFieldInput

**Add `isArray` to mutation input:**
```graphql
input AddFieldInput {
  # ... existing fields ...
  isArray: Boolean = false  # default false, true only valid when format=ENUM
}
```

---

## 4. Storage Mapping

### 4.1 TypeMapper

**File:** `internal/domain/modeldesign/type_mapper.go`

**Before:**
```go
case FormatEnum:
    return "VARCHAR(64)", nil
case FormatEnumArray:
    return string(ddlfactory.JSON), nil
```

**After:**
```go
case FormatEnum:
    if field.IsArray {
        return string(ddlfactory.JSON), nil  // multi-select: JSON array
    }
    return "VARCHAR(64)", nil  // single-select: enum code string
```

---

## 5. Validation Rules

### 5.1 isArray Validation

**File:** `internal/domain/modeldesign/field_validator.go`

| Condition | Result |
|-----------|--------|
| `format=ENUM, isArray=false` | Valid |
| `format=ENUM, isArray=true` | Valid (requires `enum.isMultiSelect=true`) |
| `format!=ENUM, isArray=true` | **Invalid** — error: `INVALID_PARAMETER: enum array not allowed for non-enum format` |
| `format!=ENUM, isArray=false` | Valid (isArray ignored) |

### 5.2 Enum Code Length Validation

**File:** `internal/domain/modeldesign/enum_definition.go`

- `EnumOption.Code` max length: **64 characters**
- Validated at enum creation and option add/update time
- Error: `INVALID_PARAMETER: enum option code exceeds 64 characters`
- Enforced via `validateCodeLength()` helper in domain model

### 5.3 isArray + Enum MultiSelect Consistency

When creating a field with `format=ENUM, isArray=true`:
- The referenced `enum.isMultiSelect` must be `true`
- Error: `INVALID_PARAMETER: enum does not support multi-select`

---

## 6. isArray Mutability

**`isArray` is immutable after field creation.**

`IsArray` is derived from `EnumDefinition.IsMultiSelect` at field creation time (see Section 7.3). It cannot be changed via `UpdateField`.

- `AddFieldInput.isArray` is optional; if omitted, defaults based on `enum.isMultiSelect`
- `UpdateFieldInput` does NOT include `isArray` — updating `isArray` is not allowed

---

## 7. Changes by File

| Status | File | Change |
|--------|------|--------|
| [x] 已完成 | `internal/domain/modeldesign/field_definition.go` | Add `IsArray bool`; remove `FormatEnumArray` constant; update helpers |
| [ ] 未完成 | `internal/domain/modeldesign/enum_definition.go` | Add `EnumOption.Code` length validation (max 64 chars); add `validateCodeLength()` helper |
| [x] 已完成 | `api/graph/project/schema/field.graphql` | Remove `ENUM_ARRAY` from `FormatType`; add `isArray Boolean!` to `Field`; add `isArray Boolean = false` to `AddFieldInput` |
| [x] 已完成 | `internal/domain/modeldesign/type_mapper.go` | Consolidate ENUM/ENUM_ARRAY into `FormatEnum + IsArray` switch case |
| [x] 已完成 | `internal/domain/modeldesign/field_validator.go` | Add `isArray` validation rules (Section 5.1, 5.3) |
| [x] 已完成 | `internal/domain/modeldesign/field_service.go` | Derive `IsArray` from `EnumDefinition.IsMultiSelect` at creation (Section 9.3) |
| [x] 已完成 | `internal/interfaces/graphql/project/adapter/field_mapper.go` | Map `FieldDefinition.IsArray` → GraphQL `Field.isArray`; update `formatToGraphQL()` to remove `ENUM_ARRAY` |
| [x] 已完成 | `internal/interfaces/graphql/project/adapter/model_mapper.go` | Update `mapEnumFieldToGraphQL()` to include `isArray` in output |
| [x] 已完成 | `internal/domain/modeldesign/jsonschema_generator.go` | Output `isArray` in JSON Schema; remove `ENUM_ARRAY` format handling |
| [x] 已完成 | `internal/domain/modeldesign/jsonschema_parser.go` | Parse `isArray` from JSON Schema `x-is-array` extension |

---

## 8. Database Migration (Breaking Change)

**This is a breaking change.** Existing fields with `format=ENUM_ARRAY` must be migrated to `format=ENUM, isArray=true`.

Migration SQL (one-time):
```sql
-- Migrate ENUM_ARRAY fields to ENUM + isArray
UPDATE field_definition
SET    type_format = 'ENUM'
WHERE  type_format = 'ENUM_ARRAY';
```

**No backward compatibility** — old `ENUM_ARRAY` format value will be rejected after this change.

---

## 9. Derived / Helper Logic

### 9.1 IsEnumArrayField Helper

```go
// Returns true for multi-select enum fields
func (fd *FieldDefinition) IsEnumArrayField() bool {
    return fd.Type.Format == FormatEnum && fd.IsArray
}
```

### 9.2 IsEnumField (Unchanged behavior)

```go
// Returns true for any enum field (single or multi-select)
func (fd *FieldDefinition) IsEnumField() bool {
    return fd.Type.Format == FormatEnum
}
```

### 9.3 Field Creation Derivation

On `AddField` mutation with `format=ENUM`:
1. Look up `EnumDefinition` by `enumConfig.enumName`
2. Set `field.IsArray = enumDefinition.IsMultiSelect`
3. Validate: if `isArray=true` but `enum.IsMultiSelect=false`, reject with error

---

## 10. Impact Summary

**Removed:**
- `FormatEnumArray` constant
- `ENUM_ARRAY` GraphQL enum value
- Branching logic: `if format == FormatEnumArray {...} else if format == FormatEnum {...}`

**Added:**
- `IsArray bool` on `FieldDefinition`
- `isArray Boolean!` on GraphQL `Field` type
- `isArray Boolean = false` on GraphQL `AddFieldInput`
- `EnumOption.Code` max length validation (64 chars)
- Strict validation: `isArray=true` only valid with `format=ENUM`

**Preserved:**
- `EnumDefinition.IsMultiSelect` as the source of truth
- JSON Schema generation for enum fields
- All existing field validation rules (nonNull, required, unique, etc.)

---

## 11. Task Completion Checklist

- [x] Domain model: add `FieldDefinition.IsArray` and remove `FormatEnumArray`
- [x] Domain model: TypeMapper consolidate ENUM/ENUM_ARRAY logic to use IsArray flag
- [ ] Domain model: enforce `EnumOption.Code` length <= 64
- [x] GraphQL schema: remove `ENUM_ARRAY` and add `Field.isArray`
- [x] GraphQL input: add `AddFieldInput.isArray`
- [x] GraphQL adapter: map domain `IsArray` to GraphQL field output
- [x] GraphQL adapter: remove `FormatTypeEnumArray` from conversion logic
- [x] DTO layer: add `FieldDefinitionDTO.IsArray`
- [x] Mapper layer: support `IsArray` conversion DTO ↔ Domain
- [x] Validation: add `isArray` legality and enum multi-select consistency checks
- [x] Service layer: derive `field.IsArray` from `EnumDefinition.IsMultiSelect`
- [x] JSON Schema generator/parser: support `isArray` and remove `ENUM_ARRAY` handling
- [x] Data migration: convert historical `ENUM_ARRAY` to `ENUM`
