# Capability: Query API

## ADDED Requirements

### Requirement: Prisma-Style Operator Naming
The query condition system SHALL use Prisma-compatible operator naming conventions for logical and comparison operators.

#### Scenario: Logical operators use uppercase naming
- **WHEN** constructing AND conditions
- **THEN** the operator key SHALL be "AND" (uppercase) instead of "_and"

#### Scenario: Logical operators use uppercase OR naming
- **WHEN** constructing OR conditions
- **THEN** the operator key SHALL be "OR" (uppercase) instead of "_or"

#### Scenario: Field operators maintain Prisma naming
- **WHEN** using field comparison operators
- **THEN** operators SHALL use Prisma conventions: equals, not, in, notIn, lt, lte, gt, gte, contains, startsWith, endsWith, mode

### Requirement: Reserved Keyword Validation
The system SHALL maintain a comprehensive list of reserved keywords that cannot be used as field names to prevent conflicts with query operators.

#### Scenario: Reserved keyword list includes all operators
- **WHEN** the system initializes
- **THEN** reserved keywords SHALL include: AND, OR, NOT, equals, not, in, notIn, lt, lte, gt, gte, contains, startsWith, endsWith, mode

#### Scenario: Field name validation rejects reserved keywords
- **WHEN** creating or updating a field with a name matching a reserved keyword (case-insensitive)
- **THEN** the system SHALL reject the operation with a validation error
- **AND** the error message SHALL specify the conflicting keyword and suggest alternatives

#### Scenario: Reserved keyword check is case-insensitive
- **WHEN** validating field names against reserved keywords
- **THEN** the comparison SHALL be case-insensitive
- **AND** names like "and", "And", "AND" SHALL all be rejected

#### Scenario: Existing models with reserved keyword fields fail validation
- **WHEN** loading a model with field names matching reserved keywords
- **THEN** the system SHALL log a validation warning or error
- **AND** provide migration guidance to rename the conflicting fields

### Requirement: Operator Constant Definitions
The query package SHALL export well-documented constants for all supported operators aligned with Prisma naming.

#### Scenario: Constants use descriptive names
- **WHEN** developers import the query package
- **THEN** constants SHALL be available: LogicalOperatorAND, LogicalOperatorOR, FieldEquals, FieldNot, FieldIn, FieldNotIn, FieldLt, FieldLte, FieldGt, FieldGte, FieldContains, FieldStartsWith, FieldEndsWith, FieldMode

#### Scenario: Constant values match Prisma conventions
- **WHEN** using operator constants in conditions
- **THEN** their string values SHALL match Prisma operator names exactly

### Requirement: GraphQL Schema Generation with Prisma Naming
GraphQL input types for field conditions SHALL use Prisma-style operator names in generated schema definitions.

#### Scenario: String field input type uses Prisma operators
- **WHEN** generating StringFieldInput type
- **THEN** fields SHALL be named: equals, not, in, notIn, contains, startsWith, endsWith, mode

#### Scenario: Integer field input type uses Prisma operators
- **WHEN** generating IntFieldInput type
- **THEN** fields SHALL be named: equals, not, in, notIn, lt, lte, gt, gte

#### Scenario: Where input type uses uppercase logical operators
- **WHEN** generating model WhereInput types
- **THEN** logical operator fields SHALL be named: AND, OR (uppercase)

## MODIFIED Requirements

None - this is a new capability defining the query operator naming conventions and reserved keyword system.

## REMOVED Requirements

None - this change adds new requirements without removing existing functionality.
