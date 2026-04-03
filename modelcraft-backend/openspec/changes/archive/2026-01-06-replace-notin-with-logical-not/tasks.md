# Implementation Tasks

## Overview
This document outlines the implementation tasks for replacing the `notIn` operator with the logical `NOT` operator.

## Task Breakdown

### Phase 1: Add Logical NOT Operator Support

#### Task 1.1: Add LogicalOperatorNOT constant
**What**: Define the LogicalOperatorNOT constant in the query package
**Where**: `internal/domain/query/field_conditions.go`
**Details**:
- Add `LogicalOperatorNOT = "NOT"` constant alongside LogicalOperatorAND and LogicalOperatorOR
- Update package documentation to mention NOT operator
**Validation**: Constant is exported and accessible

#### Task 1.2: Update IsLogicalOperator function
**What**: Include NOT in logical operator checks
**Where**: `internal/domain/query/field_conditions.go`
**Details**:
- Modify `IsLogicalOperator()` to return true for "NOT"
**Validation**: `IsLogicalOperator("NOT")` returns true

#### Task 1.3: Add LogicalNot helper function
**What**: Create a convenience function for building NOT conditions
**Where**: `internal/domain/query/field_conditions.go`
**Details**:
- Add `func LogicalNot(condition Condition) Condition` function
- Return `Condition{LogicalOperatorNOT: condition}`
- Add documentation with examples
**Validation**: Function can be called and produces correct structure

#### Task 1.4: Update Condition.Not() method semantics
**What**: Clarify or update the Not() method to use logical NOT
**Where**: `internal/domain/query/field_conditions.go` (line 73-76)
**Details**:
- Review existing `Condition.Not()` method - it currently returns `Condition{FieldNot: c}`
- Update to use `LogicalOperatorNOT` instead: `Condition{LogicalOperatorNOT: c}`
- This makes it a logical negation operator
**Validation**: Tests pass showing logical negation behavior

#### Task 1.5: Add NOT operator tests
**What**: Create comprehensive tests for NOT operator
**Where**: `internal/domain/query/field_conditions_test.go`
**Details**:
- Test simple NOT conditions: `NOT: { field: { equals: value } }`
- Test complex NOT conditions: `NOT: { AND: [...] }`
- Test multiple NOT in AND/OR combinations
- Test NOT replacing notIn semantics
**Validation**: All tests pass with >90% code coverage

### Phase 2: Remove notIn Operator

#### Task 2.1: Remove FieldNotIn constant
**What**: Delete the FieldNotIn constant definition
**Where**: `internal/domain/query/field_conditions.go` (line 115)
**Details**:
- Remove `FieldNotIn = "notIn"` constant
- Remove associated comment
**Validation**: Code compiles without the constant

#### Task 2.2: Remove NotIn helper functions
**What**: Delete NotIn() and FieldBuilder.NotIn() functions
**Where**: `internal/domain/query/field_conditions.go`
**Details**:
- Remove `func NotIn(vals ...any) Condition` (lines 280-283)
- Remove `func (fb *FieldBuilder) NotIn(vals ...any) Condition` (lines 206-209)
**Validation**: No references to these functions remain

#### Task 2.3: Update IsComparisonOperator function
**What**: Remove notIn from comparison operator checks
**Where**: `internal/domain/query/field_conditions.go` (line 160)
**Details**:
- Remove `FieldNotIn` from the switch statement
**Validation**: `IsComparisonOperator("notIn")` returns false

#### Task 2.4: Remove notIn from reserved keywords
**What**: Remove "notIn" from the reserved keywords list
**Where**: `internal/domain/query/field_conditions.go` (line 402)
**Details**:
- Remove "notIn" from reservedKeywords array
**Validation**: `IsReservedKeyword("notIn")` returns false

#### Task 2.5: Update all tests removing notIn references
**What**: Update or remove tests that use notIn
**Where**: 
- `internal/domain/query/field_conditions_test.go` (lines 21, 90)
- `internal/domain/query/reserved_keywords_test.go` (lines 24, 63-65, 143)
**Details**:
- Remove test cases for FieldNotIn constant
- Remove test cases for NotIn() functions
- Update reserved keyword tests to not expect "notIn"
- Migrate any notIn usage tests to use NOT + in pattern
**Validation**: All tests pass

### Phase 3: Update Query Visitor and SQL Generation

#### Task 3.1: Add NOT operator handling in query visitor
**What**: Implement NOT operator translation to SQL
**Where**: `internal/infrastructure/database/dml/query_visitor.go`
**Details**:
- Add case for LogicalOperatorNOT in visitor
- Generate SQL NOT(...) wrapping the nested condition
- Handle both object and array forms of NOT
**Validation**: Integration tests with NOT conditions produce correct SQL

#### Task 3.2: Remove notIn handling from query visitor
**What**: Remove any notIn-specific SQL generation code
**Where**: `internal/infrastructure/database/dml/query_visitor.go`
**Details**:
- Search for "notIn" or "NOT IN" SQL generation
- Remove or update to handle via NOT + in pattern
**Validation**: No notIn references remain in query visitor

### Phase 4: Update GraphQL Schema Generation

#### Task 4.1: Add NOT field to WhereInput types
**What**: Include NOT in generated GraphQL where input types
**Where**: GraphQL schema generation code (search for `AND` and `OR` generation)
**Details**:
- Add NOT field to WhereInput types
- NOT should accept single WhereInput (not array like AND/OR)
- Ensure recursive type reference works
**Validation**: Generated schema includes NOT field

#### Task 4.2: Remove notIn from field input types
**What**: Remove notIn fields from StringFieldInput, IntFieldInput, etc.
**Where**: GraphQL schema generation for field input types
**Details**:
- Remove notIn from all field-level input type generators
- Ensure no references to notIn remain in schema
**Validation**: Generated schema has no notIn fields

### Phase 5: Documentation and Migration Guide

#### Task 5.1: Update package documentation
**What**: Update query package godoc
**Where**: `internal/domain/query/field_conditions.go` (top of file)
**Details**:
- Update operator list to remove notIn
- Add NOT to logical operators section
- Add examples demonstrating NOT usage
- Clarify not vs NOT distinction
**Validation**: godoc renders correctly with new examples

#### Task 5.2: Update API guide documentation
**What**: Update user-facing documentation
**Where**: `docs/03-runtime/api-guide.md`
**Details**:
- Remove all notIn examples
- Add NOT operator examples
- Add migration section: "Replacing notIn with NOT"
- Add semantic clarity section explaining not vs NOT
**Validation**: Documentation is clear and includes all use cases

#### Task 5.3: Add migration examples
**What**: Create comprehensive migration examples
**Where**: `docs/03-runtime/api-guide.md` or new migration doc
**Details**:
- Show before/after for common notIn patterns
- Provide code snippets in both GraphQL and Go
- Include complex examples with AND/OR combinations
**Validation**: Examples are tested and accurate

### Phase 6: Integration Testing and Validation

#### Task 6.1: Create end-to-end NOT operator tests
**What**: Test NOT operator through full stack
**Where**: Integration test suite
**Details**:
- Test GraphQL queries with NOT operator
- Test NOT with various field operators (equals, contains, in, gt, etc.)
- Test complex nesting (NOT + AND + OR combinations)
- Verify SQL generation is correct
**Validation**: All integration tests pass

#### Task 6.2: Create migration validation tests
**What**: Verify notIn -> NOT migration patterns work
**Where**: Integration test suite
**Details**:
- Test that `NOT: { field: { in: [...] } }` produces same results as old notIn
- Compare query results before/after migration
**Validation**: Functional equivalence demonstrated

#### Task 6.3: Performance validation
**What**: Ensure NOT operator performs equivalently to notIn
**Where**: Performance test suite
**Details**:
- Benchmark NOT + in vs old notIn implementation
- Verify no significant performance regression
**Validation**: Performance is acceptable (<5% regression if any)

#### Task 6.4: Update all example code
**What**: Find and update any example code using notIn
**Where**: Throughout codebase and docs
**Details**:
- Search for "notIn" in all files
- Update examples to use NOT pattern
**Validation**: No notIn references remain except in migration docs

## Dependencies

- **Phase 1** must complete before Phase 2 (add NOT before removing notIn)
- **Phases 1-2** must complete before Phase 3 (query visitor needs constants)
- **Phase 3** must complete before Phase 6.1-6.2 (need working implementation for integration tests)
- **Phase 4** is parallel to Phase 3 (independent schema generation)
- **Phase 5** can happen in parallel with Phases 3-4
- **Phase 6.4** should be last (final cleanup)

## Parallelizable Work

- Phase 3 (Query Visitor) and Phase 4 (GraphQL Schema) can be done in parallel
- Phase 5 (Documentation) can be done in parallel with Phases 3-4
- Within Phase 1, tasks 1.1-1.3 are sequential, but 1.5 can happen after 1.1-1.4
- Within Phase 2, all tasks can be done in any order after Phase 1 completes

## Success Criteria

- ✓ All tests pass
- ✓ No compilation errors
- ✓ No references to "notIn" remain in code (except migration docs)
- ✓ NOT operator works in all contexts (simple, complex, nested)
- ✓ GraphQL schema includes NOT, excludes notIn
- ✓ Documentation is complete and clear
- ✓ Integration tests demonstrate functional equivalence
- ✓ Performance is acceptable
