# Implementation Tasks

## 1. Core Query Operator Refactoring

- [x] 1.1 Update operator constants in `internal/domain/query/field_conditions.go`
  - [x] Change `LogicalOperatorAnd = "_and"` to `LogicalOperatorAND = "AND"`
  - [x] Change `LogicalOperatorOr = "_or"` to `LogicalOperatorOR = "OR"`
  - [x] Verify all field operator constants match Prisma naming (equals, not, in, notIn, lt, lte, gt, gte, contains, startsWith, endsWith, mode)
  
- [x] 1.2 Update helper functions to use new constants
  - [x] Update `And()` function to use `LogicalOperatorAND`
  - [x] Update `Or()` function to use `LogicalOperatorOR`
  - [x] Update `IsLogicalOperator()` to check for "AND" and "OR"

- [x] 1.3 Add reserved keywords list and validation
  - [x] Create `GetReservedKeywords()` function returning []string of all operator names
  - [x] Create `IsReservedKeyword(name string)` function with case-insensitive check
  - [x] Document each reserved keyword with inline comments

## 2. Field Name Validation

- [x] 2.1 Add reserved keyword validation in field service
  - [x] Import query package in `internal/domain/modeldesign/field_service.go`
  - [x] Add validation check in field creation logic
  - [x] Add validation check in field update logic
  - [x] Return descriptive error with conflicting keyword and suggestions

- [x] 2.2 Create validation error types
  - [x] Define error messages with field name and conflicting keyword
  - [x] Include field name and conflicting keyword in error message
  - [x] Provide helpful migration guidance in error text

## 3. GraphQL Schema Generation Updates

- [x] 3.1 Update field condition type generation in `internal/domain/modelruntime/graphql_field_conditions.go`
  - [x] Ensure all `getStringFieldConditionType()` fields use Prisma naming
  - [x] Ensure all `getIntFieldConditionType()` fields use Prisma naming
  - [x] Ensure all `getNumberFieldConditionType()` fields use Prisma naming
  - [x] Ensure all date/time field condition types use Prisma naming
  - [x] Update `getBooleanFieldConditionType()` field names

- [x] 3.2 Update where input type generation
  - [x] Locate where input type generation code
  - [x] Change AND/OR field names to uppercase in generated types
  - [x] Verify generated schema matches Prisma conventions

## 4. Query Parser Updates

- [x] 4.1 Update query parsing logic in `internal/infrastructure/database/dml/query_parser.go`
  - [x] Update logical operator detection to check for "AND"/"OR"
  - [x] Update default logical operator to use new constant
  - [x] Update any hardcoded operator string checks

- [x] 4.2 Update query node processing
  - [x] Review `internal/infrastructure/database/dml/query_visitor.go`
  - [x] Update operator handling to use new constants
  - [x] Verify SQL generation uses correct logic

## 5. Testing and Migration

- [x] 5.1 Update existing tests
  - [x] Find all tests using `_and`/`_or` via `rg -n "_and|_or" --type go`
  - [x] Update test cases to use `AND`/`OR`
  - [x] Update test assertions for operator names

- [x] 5.2 Create reserved keyword validation tests
  - [x] Test field creation with reserved keyword names (should fail)
  - [x] Test case-insensitive keyword detection
  - [x] Test error messages provide helpful guidance
  - [x] Test legitimate field names that are similar but not keywords

- [x] 5.3 Migration documentation (covered in code documentation and error messages)
  - [x] Document all operator naming changes
  - [x] Provide error messages with suggestions
  - [x] List all reserved keywords in GetReservedKeywords()
  - [x] Provide field renaming strategies in error messages

## 6. Validation and Documentation

- [x] 6.1 Run comprehensive tests
  - [x] Run unit tests: `go test ./internal/domain/query/...`
  - [x] Run integration tests for query parsing
  - [x] Run field service tests
  - [x] Verify all tests pass

- [x] 6.2 Update code documentation
  - [x] Add package-level documentation explaining Prisma alignment
  - [x] Document reserved keywords in godoc comments
  - [x] Add usage examples in code comments

- [x] 6.3 Validate with OpenSpec
  - [x] Run `openspec validate refactor-query-conditions-prisma-style --strict`
  - [x] Resolve any validation errors
  - [x] Ensure all scenarios are covered by implementation
