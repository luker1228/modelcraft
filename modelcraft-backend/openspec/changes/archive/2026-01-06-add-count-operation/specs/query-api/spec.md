## ADDED Requirements

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
