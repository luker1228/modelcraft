# modelruntime-datetime Specification

## Purpose
Defines GraphQL scalar type implementations for Date and Time types, runtime type mapping, and query filtering support for date/time fields in the ModelCraft runtime.

## ADDED Requirements

### Requirement: Date GraphQL Scalar Type

The system SHALL provide a custom GraphQL scalar type for Date fields.

#### Scenario: Define Date scalar in GraphQL schema

- **WHEN** the GraphQL schema is generated
- **THEN** a custom `Date` scalar type is declared in the schema
- **AND** the scalar is defined in `api/graph/schema/base.graphql` as `scalar Date`

#### Scenario: Serialize Date to string

- **WHEN** a Date value is returned from a GraphQL query
- **THEN** the Date scalar serializes the value to ISO 8601 format `YYYY-MM-DD`
- **AND** the serialized string is sent to the client

#### Scenario: Deserialize Date from string

- **WHEN** a Date value is received as input from a GraphQL mutation or query argument
- **THEN** the Date scalar parses the string in ISO 8601 format `YYYY-MM-DD`
- **AND** validates the format before accepting the value
- **AND** returns a validation error if the format is invalid

#### Scenario: Date scalar implementation location

- **WHEN** the Date scalar is implemented
- **THEN** it is defined in a new file `internal/domain/modelruntime/graphql_scalars.go`
- **AND** follows the graphql-go scalar implementation pattern
- **AND** implements `Serialize()`, `ParseValue()`, and `ParseLiteral()` methods

### Requirement: Time GraphQL Scalar Type

The system SHALL provide a custom GraphQL scalar type for Time fields.

#### Scenario: Define Time scalar in GraphQL schema

- **WHEN** the GraphQL schema is generated
- **THEN** a custom `Time` scalar type is declared in the schema
- **AND** the scalar is defined in `api/graph/schema/base.graphql` as `scalar Time`

#### Scenario: Serialize Time to string

- **WHEN** a Time value is returned from a GraphQL query
- **THEN** the Time scalar serializes the value to `HH:MM:SS` format
- **AND** the serialized string is sent to the client

#### Scenario: Deserialize Time from string

- **WHEN** a Time value is received as input from a GraphQL mutation or query argument
- **THEN** the Time scalar parses the string in `HH:MM:SS` format
- **AND** validates the format before accepting the value
- **AND** returns a validation error if the format is invalid

#### Scenario: Time scalar implementation location

- **WHEN** the Time scalar is implemented
- **THEN** it is defined in `internal/domain/modelruntime/graphql_scalars.go`
- **AND** follows the graphql-go scalar implementation pattern
- **AND** implements `Serialize()`, `ParseValue()`, and `ParseLiteral()` methods

### Requirement: Runtime Type Mapping for Date and Time

The system SHALL map Date and Time field formats to their corresponding GraphQL scalar types at runtime.

#### Scenario: Map FormatDate to Date scalar

- **WHEN** the runtime generates GraphQL types from field definitions
- **THEN** fields with `FormatDate` are mapped to the custom `Date` scalar
- **AND** NOT to `graphql.String`
- **AND** the mapping is updated in `internal/domain/modelruntime/runtimemodel.go` function `getGraphqlTypeBy()`

#### Scenario: Map FormatTime to Time scalar

- **WHEN** the runtime generates GraphQL types from field definitions
- **THEN** fields with `FormatTime` are mapped to the custom `Time` scalar
- **AND** NOT to `graphql.String`
- **AND** the mapping is updated in `internal/domain/modelruntime/runtimemodel.go` function `getGraphqlTypeBy()`

#### Scenario: Keep FormatDateTime mapping unchanged

- **WHEN** the runtime generates GraphQL types from field definitions
- **THEN** fields with `FormatDateTime` continue to map to `graphql.DateTime`
- **AND** existing DateTime behavior is not changed

#### Scenario: Type mapping consistency

- **WHEN** a model with Date, Time, and DateTime fields is loaded
- **THEN** all three types are properly mapped to their respective scalars
- **AND** GraphQL schema generation succeeds without errors

### Requirement: Date and Time Query Filtering Support

The system SHALL support WHERE clause filtering operations on Date and Time fields.

#### Scenario: Equality comparison for Date fields

- **WHEN** a GraphQL query includes `where: { dateField: { eq: "2024-01-15" } }`
- **THEN** the system generates SQL with proper date comparison
- **AND** only records with dateField equal to 2024-01-15 are returned

#### Scenario: Inequality comparison for Date fields

- **WHEN** a GraphQL query includes `where: { dateField: { ne: "2024-01-15" } }`
- **THEN** the system generates SQL excluding the specified date
- **AND** records with different dates are returned

#### Scenario: Greater than comparison for Date fields

- **WHEN** a GraphQL query includes `where: { dateField: { gt: "2024-01-15" } }`
- **THEN** the system generates SQL with `dateField > '2024-01-15'`
- **AND** only records after the specified date are returned

#### Scenario: Greater than or equal comparison for Date fields

- **WHEN** a GraphQL query includes `where: { dateField: { gte: "2024-01-15" } }`
- **THEN** the system generates SQL with `dateField >= '2024-01-15'`
- **AND** records on or after the specified date are returned

#### Scenario: Less than comparison for Date fields

- **WHEN** a GraphQL query includes `where: { dateField: { lt: "2024-01-15" } }`
- **THEN** the system generates SQL with `dateField < '2024-01-15'`
- **AND** only records before the specified date are returned

#### Scenario: Less than or equal comparison for Date fields

- **WHEN** a GraphQL query includes `where: { dateField: { lte: "2024-01-15" } }`
- **THEN** the system generates SQL with `dateField <= '2024-01-15'`
- **AND** records on or before the specified date are returned

#### Scenario: Between operator for Date range queries

- **WHEN** a GraphQL query includes `where: { dateField: { between: ["2024-01-01", "2024-12-31"] } }`
- **THEN** the system generates SQL with `dateField BETWEEN '2024-01-01' AND '2024-12-31'`
- **AND** only records within the date range are returned

#### Scenario: Time field comparison operators

- **WHEN** a GraphQL query includes WHERE conditions on Time fields (e.g., `timeField: { gt: "14:30:00" }`)
- **THEN** the system applies comparison operators (eq, ne, gt, gte, lt, lte, between) to Time fields
- **AND** generates proper SQL time comparisons
- **AND** returns correctly filtered results

#### Scenario: DateTime field comparison operators

- **WHEN** a GraphQL query includes WHERE conditions on DateTime fields (e.g., `createdAt: { gt: "2024-01-15T14:30:00Z" }`)
- **THEN** the system applies comparison operators to DateTime fields
- **AND** handles timezone-aware comparisons correctly
- **AND** returns correctly filtered results

#### Scenario: WHERE clause implementation location

- **WHEN** Date/Time filtering support is added
- **THEN** the implementation is in `internal/domain/modelruntime/graphql_field_conditions.go`
- **AND** extends the existing `convertWhereToExpression()` function or related helpers
- **AND** follows the pattern used for other field types (string, number, etc.)

### Requirement: SQL Parameter Binding for Date and Time

The system SHALL use prepared statements and proper SQL parameter binding for date/time values.

#### Scenario: Date parameter binding in WHERE clause

- **WHEN** a WHERE clause includes a date comparison
- **THEN** the date value is passed as a SQL parameter (not concatenated into the query string)
- **AND** SQL injection is prevented
- **AND** the goqu query builder handles parameter binding

#### Scenario: Time parameter binding in WHERE clause

- **WHEN** a WHERE clause includes a time comparison
- **THEN** the time value is passed as a SQL parameter
- **AND** SQL injection is prevented

#### Scenario: DateTime parameter binding in WHERE clause

- **WHEN** a WHERE clause includes a datetime comparison
- **THEN** the datetime value is passed as a SQL parameter
- **AND** timezone information is preserved in the parameter
- **AND** SQL injection is prevented

#### Scenario: Between operator parameter binding

- **WHEN** a WHERE clause uses the between operator for date/time ranges
- **THEN** both boundary values are passed as SQL parameters
- **AND** SQL injection is prevented
- **AND** the SQL generated is `field BETWEEN ? AND ?` with parameters

### Requirement: Error Handling for Date and Time Operations

The system SHALL provide clear error messages for invalid date/time operations.

#### Scenario: Invalid date format in WHERE clause

- **WHEN** a WHERE clause contains an invalid date string (e.g., "invalid-date")
- **THEN** the system returns a GraphQL error
- **AND** the error message indicates the expected date format `YYYY-MM-DD`
- **AND** the query is not executed

#### Scenario: Invalid time format in WHERE clause

- **WHEN** a WHERE clause contains an invalid time string (e.g., "25:00:00")
- **THEN** the system returns a GraphQL error
- **AND** the error message indicates the expected time format `HH:MM:SS`
- **AND** the query is not executed

#### Scenario: Invalid datetime format in WHERE clause

- **WHEN** a WHERE clause contains an invalid datetime string
- **THEN** the system returns a GraphQL error
- **AND** the error message indicates the expected ISO 8601 format with timezone
- **AND** the query is not executed

#### Scenario: Database-level date/time errors

- **WHEN** a database error occurs during date/time comparison (e.g., incompatible types)
- **THEN** the system logs the error with context information
- **AND** returns a GraphQL error to the client
- **AND** follows existing error handling patterns

### Requirement: Type Safety in GraphQL Schema Generation

The system SHALL enforce type constraints for Date and Time fields in generated schemas.

#### Scenario: Date fields expose Date scalar in schema

- **WHEN** a model with a Date field is registered
- **THEN** the generated GraphQL schema exposes the field as `Date` type
- **AND** NOT as `String` type
- **AND** GraphQL clients can use proper Date scalar types

#### Scenario: Time fields expose Time scalar in schema

- **WHEN** a model with a Time field is registered
- **THEN** the generated GraphQL schema exposes the field as `Time` type
- **AND** NOT as `String` type
- **AND** GraphQL clients can use proper Time scalar types

#### Scenario: DateTime fields expose DateTime scalar in schema

- **WHEN** a model with a DateTime field is registered
- **THEN** the generated GraphQL schema continues to expose the field as `DateTime` type
- **AND** existing behavior is maintained

#### Scenario: Input types use correct scalars

- **WHEN** input types are generated for mutations and filters
- **THEN** Date fields use `Date` scalar in input types
- **AND** Time fields use `Time` scalar in input types
- **AND** DateTime fields use `DateTime` scalar in input types
- **AND** WHERE clause input types reflect the correct scalar types

### Requirement: Backward Compatibility and Migration

The system SHALL handle the transition from String to proper scalar types for Date and Time fields.

#### Scenario: Existing data remains compatible

- **WHEN** the GraphQL schema changes from String to Date/Time scalars
- **THEN** existing data in MySQL DATE and TIME columns is not affected
- **AND** no database migration is required
- **AND** data continues to be stored and retrieved correctly

#### Scenario: GraphQL clients must update

- **WHEN** the GraphQL schema is updated with new scalar types
- **THEN** GraphQL clients must be updated to handle Date/Time scalars instead of strings
- **AND** this is documented as a breaking change
- **AND** clients using introspection will automatically discover the new types

#### Scenario: API version compatibility

- **WHEN** deploying the updated schema
- **THEN** the change is treated as a breaking GraphQL schema change
- **AND** appropriate API versioning or migration communication is provided
- **AND** existing clients may experience type mismatches until updated
