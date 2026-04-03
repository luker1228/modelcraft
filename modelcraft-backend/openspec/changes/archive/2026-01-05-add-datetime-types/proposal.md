# Change: Complete DateTime, Date and Time Type Support

## Why

ModelCraft currently has partial implementation of DateTime, Date, and Time field types. While these types are defined at the domain level (`FormatDateTime`, `FormatDate`, `FormatTime`) and can be persisted to MySQL, they lack proper GraphQL scalar type definitions and runtime query filtering support.

Current limitations:
- Date and Time fields are mapped to GraphQL `String` instead of proper scalar types, losing type safety
- No GraphQL scalar serialization/deserialization for Date and Time types
- WHERE clause filtering (gt, lt, gte, lte, between) does not work properly with Date/Time fields
- No format validation (ISO 8601 for Date/DateTime, HH:MM:SS for Time)
- No range validation support (min/max date/time constraints)

This prevents users from:
- Building applications that require date/time queries (e.g., "find orders after 2024-01-01")
- Validating date/time input formats at the API layer
- Enforcing business rules with date/time constraints
- Using type-safe GraphQL clients with proper Date/Time types

## What Changes

**GraphQL Schema Enhancements:**
- Define custom GraphQL scalar types: `Date`, `Time` (DateTime scalar already exists in graphql-go)
- Implement serialization/deserialization for Date and Time scalars
- Update runtime GraphQL type mapper to use proper scalar types instead of String

**Format Validation:**
- Add format validation for Date fields: ISO 8601 `YYYY-MM-DD` (e.g., "2024-01-15")
- Add format validation for DateTime fields: ISO 8601 with timezone (e.g., "2024-01-15T14:30:00Z")
- Add format validation for Time fields: `HH:MM:SS` 24-hour format (e.g., "14:30:00")

**Range Validation:**
- Extend `ValidationConfig` to support `minDate`, `maxDate` for Date/DateTime fields
- Extend `ValidationConfig` to support `minTime`, `maxTime` for Time fields
- Validate date/time ranges in field validation logic

**Query Filtering Support:**
- Add Date/Time field support to WHERE clause generator in `graphql_field_conditions.go`
- Support comparison operators: `eq`, `ne`, `gt`, `gte`, `lt`, `lte` for Date/DateTime/Time fields
- Support `between` operator for date/time range queries
- Ensure proper SQL parameter binding for date/time values

**Type Mapping Consistency:**
- Update `runtimemodel.go` to map `FormatDate` → custom `Date` scalar
- Update `runtimemodel.go` to map `FormatTime` → custom `Time` scalar
- Keep `FormatDateTime` → `graphql.DateTime` mapping (already correct)

## Impact

**Affected specs:**
- `modeldesign-field-types` (NEW): Field type definitions and validation rules
- `modelruntime-datetime` (NEW): GraphQL scalar types and runtime query support

**Affected code:**
- `internal/domain/modeldesign/field_definition.go`: Extend ValidationConfig with date/time range fields
- `internal/domain/modeldesign/field_validator.go`: Add date/time format and range validation
- `internal/domain/modelruntime/runtimemodel.go`: Update GraphQL type mapping for Date/Time
- `internal/domain/modelruntime/graphql_scalars.go`: NEW - Define custom Date and Time scalar types
- `internal/domain/modelruntime/graphql_field_conditions.go`: Add Date/Time support to WHERE conditions
- `api/graph/schema/base.graphql`: Add scalar Date and scalar Time declarations

**Breaking changes:**
- GraphQL schema change: Date and Time fields will change from `String` to `Date`/`Time` scalars
- Clients must update to handle proper Date/Time scalar types instead of strings
- **Migration**: Existing data remains compatible (MySQL DATE/TIME columns unchanged); only GraphQL schema representation changes

**Migration path:**
1. Deploy updated GraphQL schema with new scalar types
2. Update client applications to use Date/Time scalars instead of String
3. No database migration needed - MySQL column types remain the same
