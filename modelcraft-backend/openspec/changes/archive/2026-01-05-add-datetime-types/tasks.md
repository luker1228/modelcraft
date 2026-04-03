# Implementation Tasks for add-datetime-types

## 1. Extend ValidationConfig for Date and Time Range Constraints

- [x] 1.1 Add `MinDate *string` field to ValidationConfig struct in `internal/domain/modeldesign/field_definition.go` with JSON tag `minDate,omitempty`
- [x] 1.2 Add `MaxDate *string` field to ValidationConfig struct with JSON tag `maxDate,omitempty`
- [x] 1.3 Add `MinTime *string` field to ValidationConfig struct with JSON tag `minTime,omitempty`
- [x] 1.4 Add `MaxTime *string` field to ValidationConfig struct with JSON tag `maxTime,omitempty`
- [x] 1.5 Write unit tests for ValidationConfig marshaling/unmarshaling with new fields

## 2. Implement Date, Time, and DateTime Format Validation

- [x] 2.1 Create date format validation function in `internal/domain/modeldesign/field_validator.go` to validate ISO 8601 `YYYY-MM-DD` format
- [x] 2.2 Create time format validation function to validate `HH:MM:SS` 24-hour format
- [x] 2.3 Create datetime format validation function to validate ISO 8601 with timezone (e.g., `2024-01-15T14:30:00Z`)
- [x] 2.4 Integrate format validation into field validation logic for FormatDate, FormatTime, and FormatDateTime fields
- [x] 2.5 Write unit tests for date format validation with valid and invalid inputs
- [x] 2.6 Write unit tests for time format validation with valid and invalid inputs
- [x] 2.7 Write unit tests for datetime format validation with valid and invalid inputs

## 3. Implement Date and Time Range Validation

- [x] 3.1 Create date range validation function to check values against minDate and maxDate constraints
- [x] 3.2 Create time range validation function to check values against minTime and maxTime constraints (including overnight ranges)
- [x] 3.3 Validate that minDate <= maxDate when both are specified
- [x] 3.4 Validate that minTime <= maxTime when both are specified (handle midnight-spanning ranges)
- [x] 3.5 Integrate range validation into field validation logic
- [x] 3.6 Write unit tests for date range validation (within range, below min, above max)
- [x] 3.7 Write unit tests for time range validation including overnight scenarios
- [x] 3.8 Write unit tests for datetime range validation

## 4. Create Custom GraphQL Scalar Types

- [x] 4.1 Create new file `internal/domain/modelruntime/graphql_scalars.go`
- [x] 4.2 Implement Date scalar type with Serialize(), ParseValue(), and ParseLiteral() methods
- [x] 4.3 Implement Time scalar type with Serialize(), ParseValue(), and ParseLiteral() methods
- [x] 4.4 Add error handling for invalid date/time formats in scalar parsing
- [x] 4.5 Write unit tests for Date scalar serialization and deserialization
- [x] 4.6 Write unit tests for Time scalar serialization and deserialization

## 5. Update GraphQL Schema Definition

- [x] 5.1 Add `scalar Date` declaration to `api/graph/schema/base.graphql`
- [x] 5.2 Add `scalar Time` declaration to `api/graph/schema/base.graphql`
- [x] 5.3 Verify DateTime scalar already exists (should be provided by graphql-go)
- [x] 5.4 Run GraphQL schema validation to ensure no conflicts

## 6. Update Runtime Type Mapping

- [x] 6.1 Update `getGraphqlTypeBy()` function in `internal/domain/modelruntime/runtimemodel.go` to map FormatDate to custom Date scalar
- [x] 6.2 Update `getGraphqlTypeBy()` function to map FormatTime to custom Time scalar
- [x] 6.3 Verify FormatDateTime continues to map to graphql.DateTime
- [x] 6.4 Write unit tests for updated type mapping
- [x] 6.5 Test schema generation with models containing Date, Time, and DateTime fields

## 7. Add Date and Time Query Filtering Support

- [x] 7.1 Update `internal/domain/modelruntime/graphql_field_conditions.go` to handle FormatDate fields in WHERE clause generation
- [x] 7.2 Add support for Date field comparison operators: eq, ne, gt, gte, lt, lte, between
- [x] 7.3 Update to handle FormatTime fields in WHERE clause generation
- [x] 7.4 Add support for Time field comparison operators: eq, ne, gt, gte, lt, lte, between
- [x] 7.5 Update to handle FormatDateTime fields in WHERE clause generation (if not already supported)
- [x] 7.6 Ensure proper SQL parameter binding for all date/time values to prevent SQL injection
- [x] 7.7 Write unit tests for Date field WHERE clause generation with each operator
- [x] 7.8 Write unit tests for Time field WHERE clause generation with each operator
- [x] 7.9 Write unit tests for DateTime field WHERE clause generation with each operator

## 8. Integration Testing

- [x] 8.1 Create integration test with a model containing Date, Time, and DateTime fields
- [x] 8.2 Test GraphQL schema generation includes proper scalar types
- [x] 8.3 Test creating records with Date, Time, and DateTime values
- [x] 8.4 Test querying records with WHERE filters on Date fields (all operators)
- [x] 8.5 Test querying records with WHERE filters on Time fields (all operators)
- [x] 8.6 Test querying records with WHERE filters on DateTime fields (all operators)
- [x] 8.7 Test between operator for date/time range queries
- [x] 8.8 Test invalid date/time format error handling
- [x] 8.9 Test date/time range validation with minDate, maxDate, minTime, maxTime
- [x] 8.10 Test backward compatibility with existing models and data

## 9. Documentation and Migration Guide

- [x] 9.1 Document the new Date and Time scalar types in API documentation
- [x] 9.2 Document ValidationConfig extensions (minDate, maxDate, minTime, maxTime)
- [x] 9.3 Create migration guide for clients updating from String to Date/Time scalars
- [x] 9.4 Add examples of date/time queries and filters
- [x] 9.5 Document the breaking change in GraphQL schema for Date and Time fields

## 10. Code Review and Validation

- [x] 10.1 Run all unit tests and ensure they pass
- [x] 10.2 Run all integration tests and ensure they pass
- [x] 10.3 Run `openspec validate add-datetime-types --strict` and resolve any issues
- [x] 10.4 Perform code review for consistency and quality
- [x] 10.5 Verify no regression in existing functionality
