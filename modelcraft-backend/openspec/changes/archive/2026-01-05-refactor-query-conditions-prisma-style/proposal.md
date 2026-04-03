# Change: Refactor Query Conditions to Prisma-Style Naming with Reserved Keywords

## Why

The current query condition implementation uses custom operator naming (`_and`, `_or`, `equals`, `not`, etc.) that differs from industry-standard Prisma ORM conventions. Additionally, there's no enforcement preventing field names from conflicting with operator keywords, which can cause parsing ambiguity. Aligning with Prisma's naming and adding reserved keyword validation improves developer familiarity, prevents naming conflicts, and makes the API more robust.

## What Changes

- **BREAKING**: Rename logical operators from `_and`/`_or` to uppercase `AND`/`OR` to match Prisma conventions
- **BREAKING**: Update field condition operator constants to align with Prisma naming
- **NEW**: Introduce reserved keyword validation - all operator keywords (AND, OR, NOT, equals, in, contains, etc.) become reserved and cannot be used as field names
- **NEW**: Add validation at model design time to reject field names that match reserved keywords
- Maintain the existing builder pattern API (`Field('name').Eq(value)`) while standardizing operator names
- Update GraphQL schema generation to use Prisma-style naming in generated types
- Provide comprehensive reserved keyword list and validation errors with helpful messages

## Impact

- **Affected specs**: query-api (new capability spec), modeldesign-field-validation (enhancement)
- **Affected code**:
  - `internal/domain/query/field_conditions.go` - Update operator constants, add reserved keyword list
  - `internal/domain/modeldesign/field_service.go` - Add field name validation against reserved keywords
  - `internal/domain/modelruntime/graphql_field_conditions.go` - Update GraphQL type field names
  - `internal/infrastructure/database/dml/query_parser.go` - Update query parsing logic
  - Test files that use query conditions
- **Migration path**: 
  - Existing code using `_and`/`_or` will need updates to `AND`/`OR`
  - Existing models with field names matching reserved keywords must be renamed
  - Provide migration guide with keyword list and renaming examples
- **Breaking change**: Yes - operator naming changes and reserved keyword enforcement require code and schema updates
