# query-api Specification

## Purpose
Query API provides Prisma-style query condition builders with support for logical operators (AND, OR, NOT) and field-level comparison operators. The API maintains clear semantic distinction between field-level operators (lowercase: not, in, equals) and logical operators (uppercase: AND, OR, NOT).
## Requirements

### Requirement: Query Operation Response Structure

All query operations (findUnique, findFirst, findMany) SHALL return wrapped result objects containing data payload and operational metadata fields.

#### Scenario: findUnique returns wrapped result with item field

- **WHEN** executing a findUnique query
- **THEN** the response SHALL be wrapped in a result object
- **AND** the result SHALL contain an `item` field with the matching record (nullable)
- **AND** the result SHALL contain a `timeCost` field with query execution time in milliseconds
- **AND** the result SHALL contain a `reqId` field with a unique request tracking ID

**Example**:
```graphql
query {
  findUnique(where: {id: "123"}) {
    item {
      id
      name
    }
    timeCost
    reqId
  }
}
```

**Response**:
```json
{
  "data": {
    "findUnique": {
      "item": {
        "id": "123",
        "name": "John Doe"
      },
      "timeCost": 15,
      "reqId": "01930c8a-1234-7890-abcd-ef1234567890"
    }
  }
}
```

#### Scenario: findUnique returns null item when record not found

- **WHEN** executing a findUnique query that matches no records
- **THEN** the response SHALL contain a result object
- **AND** the `item` field SHALL be null
- **AND** metadata fields (`timeCost`, `reqId`) SHALL still be present

**Example Response**:
```json
{
  "data": {
    "findUnique": {
      "item": null,
      "timeCost": 8,
      "reqId": "01930c8a-5678-7890-abcd-ef1234567890"
    }
  }
}
```

#### Scenario: findFirst returns wrapped result with item field

- **WHEN** executing a findFirst query
- **THEN** the response SHALL be wrapped in a result object
- **AND** the result SHALL contain an `item` field with the first matching record (nullable)
- **AND** the result SHALL contain a `timeCost` field with query execution time
- **AND** the result SHALL contain a `reqId` field with a unique request ID

**Example**:
```graphql
query {
  findFirst(where: {age: {gte: 18}}) {
    item {
      id
      name
      age
    }
    timeCost
    reqId
  }
}
```

**Response**:
```json
{
  "data": {
    "findFirst": {
      "item": {
        "id": "456",
        "name": "Jane Smith",
        "age": 25
      },
      "timeCost": 12,
      "reqId": "01930c8a-9876-7890-abcd-ef1234567890"
    }
  }
}
```

#### Scenario: findMany returns wrapped result with items field

- **WHEN** executing a findMany query
- **THEN** the response SHALL be wrapped in a result object
- **AND** the result SHALL contain an `items` field with an array of matching records (never null, can be empty)
- **AND** the result SHALL contain a `totalCount` field (reserved for future pagination support, currently null)
- **AND** the result SHALL contain a `timeCost` field with query execution time
- **AND** the result SHALL contain a `reqId` field with a unique request ID

**Example**:
```graphql
query {
  findMany(where: {status: {equals: "active"}}) {
    items {
      id
      name
      status
    }
    totalCount
    timeCost
    reqId
  }
}
```

**Response**:
```json
{
  "data": {
    "findMany": {
      "items": [
        {"id": "1", "name": "User 1", "status": "active"},
        {"id": "2", "name": "User 2", "status": "active"}
      ],
      "totalCount": null,
      "timeCost": 18,
      "reqId": "01930c8a-1111-7890-abcd-ef1234567890"
    }
  }
}
```

#### Scenario: Response metadata fields

- **WHEN** any query operation executes
- **THEN** the `timeCost` field SHALL contain query execution duration in milliseconds
- **AND** the `reqId` field SHALL contain a UUID v7 formatted request tracking ID
- **AND** the `reqId` SHALL be unique per request
- **AND** the `reqId` SHALL be sortable by time (UUID v7 property)

#### Scenario: Semantic field naming

- **WHEN** operations return singular results (findUnique, findFirst)
- **THEN** the data field SHALL be named `item` (singular)
- **AND** the field type SHALL be nullable

**WHEN** operations return plural results (findMany)
- **THEN** the data field SHALL be named `items` (plural)
- **AND** the field type SHALL be a non-nullable array
- **AND** empty results SHALL return an empty array, not null

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

### Requirement: Count Operation for Record Counting

The GraphQL runtime SHALL provide a `count` query operation that allows users to count records matching optional filter conditions, with support for both simple record counts and field-level non-null value counts.

#### Scenario: Count all records without filter

- **WHEN** a user queries `count` without arguments
- **THEN** the system returns the total count of records in the table
- **AND** the result structure is `{ count: <integer> }`

**Example:**
```graphql
query {
  count
}
```
Returns: `{ "count": 42 }`

#### Scenario: Count records with WHERE filter

- **WHEN** a user queries `count(where: { conditions })`
- **THEN** the system applies the WHERE filter before counting
- **AND** returns only the count of records matching the filter
- **AND** the result structure is `{ count: <integer> }`
- **AND** supports all existing WHERE operators (equals, not, in, lt, gt, AND, OR, NOT, etc.)

**Example:**
```graphql
query {
  count(where: { status: { equals: "active" } })
}
```
Returns: `{ "count": 18 }`

#### Scenario: Count with complex WHERE conditions

- **WHEN** a user queries `count` with complex WHERE conditions (AND, OR, NOT)
- **THEN** the system evaluates the full logical expression before counting
- **AND** returns the count of records matching the complex filter

**Example:**
```graphql
query {
  count(where: {
    AND: [
      { status: { equals: "active" } },
      { NOT: { role: { in: ["guest", "banned"] } } }
    ]
  })
}
```
Returns: `{ "count": 12 }`

#### Scenario: Count on empty table or no matches

- **WHEN** a count query matches zero records (empty table or WHERE filter matches nothing)
- **THEN** the system returns `{ count: 0 }`
- **AND** no error is thrown

#### Scenario: Field-level count with select parameter

- **WHEN** a user queries `count(select: { _all: true, field1: true, ... })`
- **THEN** the system returns counts for all requested fields wrapped in `fieldsCount`
- **AND** `_all` represents COUNT(*) - total record count
- **AND** field-specific counts represent COUNT(field) - non-null value counts
- **AND** the result structure is `{ fieldsCount: { _all: <int>, field1: <int>, ... } }`
- **AND** NULL values are excluded from field-specific counts

**Example:**
```graphql
query {
  count(select: {
    _all: true,
    name: true,
    email: true
  })
}
```
Returns: `{ "fieldsCount": { "_all": 30, "name": 28, "email": 25 } }`

This means:
- 30 total records exist
- 28 records have non-null `name` values (2 nulls)
- 25 records have non-null `email` values (5 nulls)

#### Scenario: Combined WHERE filter and field-level counts

- **WHEN** a user queries `count(where: {...}, select: {...})`
- **THEN** the system first applies the WHERE filter
- **AND** then counts records and field values within the filtered result set
- **AND** the result structure includes counts wrapped in `fieldsCount`

**Example:**
```graphql
query {
  count(
    where: { status: { equals: "active" } }
    select: { _all: true, name: true }
  )
}
```
Returns: `{ "fieldsCount": { "_all": 18, "name": 16 } }`

This means:
- 18 active records exist
- 16 of those active records have non-null `name` values

#### Scenario: Select without _all field

- **WHEN** a user queries `count(select: { field1: true, field2: true })` without `_all`
- **THEN** the system returns only the requested field counts wrapped in `fieldsCount`
- **AND** `_all` is NOT included in the result

**Example:**
```graphql
query {
  count(select: { name: true, email: true })
}
```
Returns: `{ "fieldsCount": { "name": 28, "email": 25 } }`

#### Scenario: Select parameter with invalid field

- **WHEN** GraphQL schema is generated
- **THEN** only valid model fields are available in the select input type
- **AND** attempting to select non-existent fields is prevented by schema validation

### Requirement: Count Input Structure

The count operation SHALL accept an input object with optional `where` and `select` parameters.

#### Scenario: Count input format

- **WHEN** calling the count operation
- **THEN** the input accepts an optional `where` parameter for filtering (same type as findMany)
- **AND** accepts an optional `select` parameter as an object with field names set to true
- **AND** `select` includes `_all` field for COUNT(*)
- **AND** `select` includes all model fields for COUNT(field)
- **AND** both parameters are optional (can be omitted)

#### Scenario: Count with no parameters

- **WHEN** calling `count` without any parameters
- **THEN** the system counts all records without filtering
- **AND** returns simple count format `{ count: N }`

#### Scenario: Select parameter triggers field-level count format

- **WHEN** `select` parameter is provided (even if `where` is not)
- **THEN** the result format changes to field-level structure `{ fieldsCount: { _all?, field1?, ... } }`
- **AND** no `count` field is returned (different format)

#### Scenario: Empty select parameter

- **WHEN** calling `count(select: {})`  (empty select object)
- **THEN** the system returns a validation error
- **AND** error message indicates at least one field must be selected

### Requirement: Count Result Structure Design

The count operation SHALL return different result structures based on whether the `select` parameter is used, following Prisma's dual format pattern.

#### Scenario: Simple count result format

- **WHEN** `select` parameter is NOT provided
- **THEN** the result type is `{ count: Int! }`
- **AND** the count represents total matching records (COUNT(*))
- **AND** this is the default format for simplicity

**GraphQL Type:**
```graphql
type CountResult {
  count: Int!
}
```

#### Scenario: Field-level count result format

- **WHEN** `select` parameter IS provided
- **THEN** the result type includes a `fieldsCount` field containing selected field counts
- **AND** `_all` field (if selected) represents COUNT(*)
- **AND** other fields represent COUNT(field) - non-null counts

**GraphQL Type (dynamically generated per model):**
```graphql
type CountResult {
  fieldsCount: UserFieldsCount
}

type UserFieldsCount {
  _all: Int
  name: Int
  email: Int
  # ... other selected fields
}
```

#### Scenario: Type safety for result structure

- **WHEN** GraphQL schema is generated for count operation
- **THEN** the resolver determines return type based on select parameter presence
- **AND** simple count uses scalar-wrapped result type
- **AND** field-level count uses object type with selected fields

### Requirement: SQL Generation for Count Queries

The system SHALL generate efficient SQL COUNT queries using the goqu query builder.

#### Scenario: Simple COUNT(*) generation

- **WHEN** count is called without `select` parameter
- **THEN** SQL includes `SELECT COUNT(*) as count FROM table`
- **AND** result is mapped to `{ count: N }` structure

#### Scenario: COUNT(*) with WHERE clause

- **WHEN** count is called with `where` parameter but no `select`
- **THEN** SQL includes `SELECT COUNT(*) as count FROM table WHERE <conditions>`
- **AND** WHERE conditions use existing `convertWhereToExpression()` function
- **AND** prepared statements prevent SQL injection

#### Scenario: Field-level COUNT generation

- **WHEN** count is called with `select` parameter
- **THEN** SQL includes `SELECT COUNT(*) as _all, COUNT(field1) as field1, ... FROM table`
- **AND** only selected fields are included in SQL
- **AND** uses column aliases matching result structure

**Example SQL:**
```sql
SELECT COUNT(*) as _all, COUNT(name) as name, COUNT(email) as email FROM users;
```

#### Scenario: Field-level COUNT with WHERE clause

- **WHEN** count is called with both `where` and `select` parameters
- **THEN** SQL includes WHERE clause before counting
- **AND** all COUNT functions operate on the filtered result set

**Example SQL:**
```sql
SELECT COUNT(*) as _all, COUNT(name) as name
FROM users
WHERE status = 'active';
```

#### Scenario: Result mapping from SQL to response

- **WHEN** SQL result is returned
- **THEN** for simple count: map single row column to `{ count: <value> }`
- **AND** for field-level count: map columns to `{ fieldsCount: { _all: <value>, field1: <value>, ... } }`
- **AND** maintain type safety (integers for all counts)

### Requirement: Error Handling

The count operation SHALL handle errors consistently with existing query operations.

#### Scenario: Database connection error

- **WHEN** database connection fails during count query
- **THEN** the system returns a GraphQL error with appropriate error code
- **AND** error is logged with context information
- **AND** follows existing error handling patterns from findMany/findFirst

#### Scenario: Invalid WHERE condition

- **WHEN** WHERE clause contains invalid operators or values
- **THEN** the system returns a validation error before executing SQL
- **AND** error message describes the validation failure
- **AND** consistent with existing WHERE validation

#### Scenario: Field does not exist in select

- **WHEN** count query references a non-existent field in select
- **THEN** GraphQL schema validation prevents the query from being submitted
- **AND** client receives schema validation error at query time

### Requirement: Semantic Distinction from Aggregate Count

The query API SHALL maintain clear semantic distinction between `count` and `aggregate._count` operations.

#### Scenario: Count operation purpose

- **WHEN** users need to answer "how many records match this filter?"
- **THEN** they should use the `count` operation
- **AND** it provides the simplest API for record counting
- **AND** returns `{ count: N }` by default for ergonomics

#### Scenario: Aggregate count purpose

- **WHEN** users need statistical aggregations including counts alongside AVG/SUM/MIN/MAX
- **THEN** they should use the `aggregate` operation with `_count` selection
- **AND** it returns nested structure `{ _count: {...}, _avg: {...}, ... }`
- **AND** designed for multi-metric queries

#### Scenario: Functional equivalence for field counts

- **WHEN** comparing `count(select: {...})` and `aggregate { _count {...} }`
- **THEN** both return field-level counts with the same semantics
- **AND** COUNT(field) excludes NULL values in both cases
- **AND** `_all` represents COUNT(*) in both cases
- **AND** count operation has simpler input/output structure

**Equivalent queries:**
```graphql
# Using count
query {
  count(select: { _all: true, name: true })
}
# Returns: { fieldsCount: { _all: 30, name: 28 } }

# Using aggregate
query {
  aggregate {
    _count { _all, name }
  }
}
# Returns: { _count: { _all: 30, name: 28 } }
```

#### Scenario: Documentation clarifies when to use each

- **WHEN** developers read the query API documentation
- **THEN** the difference between `count` and `aggregate._count` SHALL be explicitly explained
- **AND** examples demonstrate both use cases
- **AND** guidance on choosing between them is provided:
  - Use `count` for simple "how many?" questions
  - Use `aggregate._count` when combining with other aggregations

### Requirement: Performance and Efficiency

The count operation SHALL execute efficiently without unnecessary overhead.

#### Scenario: Single query execution

- **WHEN** count is called with multiple fields in select
- **THEN** the system executes exactly one SQL query
- **AND** database performs all COUNT operations in a single pass

#### Scenario: No data transfer overhead

- **WHEN** count query executes on large dataset
- **THEN** only the count result is transferred from database
- **AND** no record data is loaded into application memory
- **AND** significantly more efficient than findMany + array length

#### Scenario: Database optimization utilization

- **WHEN** count query executes with WHERE clause
- **THEN** database engine can utilize indexes for filtering
- **AND** COUNT operation benefits from query optimizer
- **AND** execution plan is visible via EXPLAIN for verification

#### Scenario: Large table handling

- **WHEN** count query runs on table with millions of records
- **THEN** the system executes the query without artificial limits
- **AND** performance is limited only by database capabilities
- **AND** query timeout follows existing timeout configuration

