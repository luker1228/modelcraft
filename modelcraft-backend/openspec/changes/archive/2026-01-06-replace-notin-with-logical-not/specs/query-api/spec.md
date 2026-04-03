# query-api Specification Delta

This document contains the specification changes for the `replace-notin-with-logical-not` proposal.

## ADDED Requirements

### Requirement: Logical NOT Operator Support
The query condition system SHALL support a logical NOT operator for negating condition blocks.

#### Scenario: NOT operator negates single condition
- **WHEN** constructing a query with `NOT: { field: { operator: value } }`
- **THEN** the system SHALL match records where the nested condition is false
- **AND** the operator key SHALL be "NOT" (uppercase)

**Example:**
```json
{
  "NOT": {
    "title": { "contains": "SQL" }
  }
}
```
Matches: All records where title does NOT contain "SQL"

#### Scenario: NOT operator negates complex conditions
- **WHEN** constructing a query with `NOT: { AND: [...] }` or `NOT: { OR: [...] }`
- **THEN** the system SHALL negate the entire nested logical expression
- **AND** support arbitrary nesting depth

**Example:**
```json
{
  "NOT": {
    "AND": [
      { "published": { "equals": true } },
      { "views": { "gt": 1000 } }
    ]
  }
}
```
Matches: Records where NOT (published = true AND views > 1000)

#### Scenario: NOT operator in combination with AND/OR
- **WHEN** using NOT within AND or OR conditions
- **THEN** the NOT SHALL be evaluated as part of the logical expression
- **AND** multiple NOT conditions SHALL be supported

**Example:**
```json
{
  "AND": [
    { "NOT": { "status": { "equals": "archived" } } },
    { "NOT": { "title": { "contains": "draft" } } }
  ]
}
```
Matches: Records where status != "archived" AND title does not contain "draft"

#### Scenario: Replace notIn with NOT + in combination
- **WHEN** users need to express "field value not in list" semantics
- **THEN** they SHALL use `NOT: { field: { in: [...] } }`
- **AND** this SHALL be functionally equivalent to the removed `notIn` operator

**Example:**
```json
{
  "NOT": {
    "status": { "in": ["draft", "archived"] }
  }
}
```
Matches: All records where status is not "draft" and not "archived"

### Requirement: Operator Constant for Logical NOT
The query package SHALL export a constant for the logical NOT operator.

#### Scenario: LogicalOperatorNOT constant is defined
- **WHEN** developers import the query package
- **THEN** the constant `LogicalOperatorNOT` SHALL be available
- **AND** its value SHALL be "NOT" (uppercase string)

#### Scenario: IsLogicalOperator includes NOT
- **WHEN** calling `IsLogicalOperator("NOT")`
- **THEN** the function SHALL return true
- **AND** NOT SHALL be treated equivalently to AND and OR in logical contexts

### Requirement: GraphQL Schema Generation with NOT Operator
GraphQL input types SHALL include the NOT logical operator in where input types.

#### Scenario: Where input type includes NOT field
- **WHEN** generating model WhereInput types
- **THEN** logical operator fields SHALL be named: AND, OR, NOT (uppercase)
- **AND** NOT field SHALL accept the same WhereInput type (recursive)

**Example schema:**
```graphql
input PostWhereInput {
  AND: [PostWhereInput!]
  OR: [PostWhereInput!]
  NOT: PostWhereInput
  
  title: StringFieldInput
  status: StringFieldInput
  views: IntFieldInput
}
```

### Requirement: Semantic Distinction Between not and NOT
The query API SHALL maintain clear semantic distinction between field-level and logical-level negation.

#### Scenario: Lowercase 'not' is field-level negation
- **WHEN** using `{ field: { not: value } }`
- **THEN** this SHALL mean "field does not equal value"
- **AND** it operates at the field comparison level

**Example:**
```json
{ "status": { "not": "draft" } }
```
Matches: Records where status != "draft"

#### Scenario: Uppercase 'NOT' is logical negation
- **WHEN** using `{ NOT: { conditions } }`
- **THEN** this SHALL mean "negate the entire condition block"
- **AND** it operates at the logical operator level

**Example:**
```json
{ "NOT": { "status": { "equals": "draft" } } }
```
Matches: Records where the condition (status = "draft") is false

#### Scenario: Documentation clarifies the distinction
- **WHEN** developers read the query API documentation
- **THEN** the semantic difference between `not` and `NOT` SHALL be explicitly explained
- **AND** examples SHALL demonstrate both use cases clearly
- **AND** migration guidance from `notIn` SHALL be provided

## MODIFIED Requirements

### Requirement: Prisma-Style Operator Naming
The query condition system SHALL use Prisma-compatible operator naming conventions for logical and comparison operators.

#### Scenario: Logical operators use uppercase NOT naming
- **WHEN** constructing NOT conditions
- **THEN** the operator key SHALL be "NOT" (uppercase)

#### Scenario: Field operators maintain Prisma naming
- **WHEN** using field comparison operators
- **THEN** operators SHALL use Prisma conventions: equals, not, in, lt, lte, gt, gte, contains, startsWith, endsWith, mode

### Requirement: Reserved Keyword Validation
The system SHALL maintain a comprehensive list of reserved keywords that cannot be used as field names to prevent conflicts with query operators.

#### Scenario: Reserved keyword list includes all operators
- **WHEN** the system initializes
- **THEN** reserved keywords SHALL include: AND, OR, NOT, equals, not, in, lt, lte, gt, gte, contains, startsWith, endsWith, mode

### Requirement: Operator Constant Definitions
The query package SHALL export well-documented constants for all supported operators aligned with Prisma naming.

#### Scenario: Constants use descriptive names
- **WHEN** developers import the query package
- **THEN** constants SHALL be available: LogicalOperatorAND, LogicalOperatorOR, LogicalOperatorNOT, FieldEquals, FieldNot, FieldIn, FieldLt, FieldLte, FieldGt, FieldGte, FieldContains, FieldStartsWith, FieldEndsWith, FieldMode

### Requirement: GraphQL Schema Generation with Prisma Naming
GraphQL input types for field conditions SHALL use Prisma-style operator names in generated schema definitions.

#### Scenario: String field input type uses Prisma operators
- **WHEN** generating StringFieldInput type
- **THEN** fields SHALL be named: equals, not, in, contains, startsWith, endsWith, mode

#### Scenario: Integer field input type uses Prisma operators
- **WHEN** generating IntFieldInput type
- **THEN** fields SHALL be named: equals, not, in, lt, lte, gt, gte

#### Scenario: Where input type uses uppercase logical operators
- **WHEN** generating model WhereInput types
- **THEN** logical operator fields SHALL be named: AND, OR, NOT (uppercase)
- **AND** NOT field SHALL accept single WhereInput (not array)
