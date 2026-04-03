# modelruntime-aggregate Specification

## Purpose
TBD - created by archiving change add-aggregate-query. Update Purpose after archive.
## Requirements
### Requirement: Aggregate Query Operation

The GraphQL runtime SHALL provide an `aggregate` query operation that allows users to perform statistical aggregations on model data including count, average, sum, minimum, and maximum calculations.

#### Scenario: Count all records

- **WHEN** a user queries `aggregate { _count { _all } }`
- **THEN** the system returns the total count of records in the table
- **AND** the result structure is `{ _count: { _all: <integer> } }`

#### Scenario: Count non-null field values

- **WHEN** a user queries `aggregate { _count { fieldName } }` for a specific field
- **THEN** the system returns the count of non-null values for that field
- **AND** null values are excluded from the count
- **AND** the result structure is `{ _count: { fieldName: <integer> } }`

#### Scenario: Multiple count operations

- **WHEN** a user queries `aggregate { _count { _all, field1, field2 } }`
- **THEN** the system returns counts for all requested targets in a single query
- **AND** the result structure is `{ _count: { _all: <int>, field1: <int>, field2: <int> } }`

#### Scenario: Calculate average of numeric field

- **WHEN** a user queries `aggregate { _avg { numericField } }` where numericField is int/float type
- **THEN** the system calculates the average value excluding null values
- **AND** the result structure is `{ _avg: { numericField: <number> } }` when there are non-null values
- **AND** returns `{ _avg: { numericField: null } }` when there are no rows OR all values are null
- **AND** this allows distinguishing true zero average from no data

#### Scenario: Calculate sum of numeric field

- **WHEN** a user queries `aggregate { _sum { numericField } }`
- **THEN** the system calculates the sum of all non-null values
- **AND** the result structure is `{ _sum: { numericField: <number> } }` when there are non-null values
- **AND** returns `{ _sum: { numericField: null } }` when there are no rows OR all values are null
- **AND** a true sum of zero (from zero values or negative/positive canceling) is returned as 0, not null

#### Scenario: Find minimum and maximum values

- **WHEN** a user queries `aggregate { _min { field }, _max { field } }`
- **THEN** the system returns the minimum and maximum values for the field
- **AND** the result structure is `{ _min: { field: <value> }, _max: { field: <value> } }`
- **AND** works on numeric, string, and date fields
- **AND** null values are excluded from min/max calculation

#### Scenario: Combined aggregations

- **WHEN** a user queries multiple aggregate operations together
- **THEN** the system executes all aggregations in a single database query
- **AND** returns all results in a single response object
- **AND** example: `{ _count: { _all }, _avg: { amount }, _sum: { amount }, _min: { amount }, _max: { amount } }`

#### Scenario: Aggregate with WHERE filter

- **WHEN** a user queries `aggregate(where: { field: { operator: value } }) { _count { _all } }`
- **THEN** the system applies the WHERE filter before aggregation
- **AND** only records matching the filter are included in calculations
- **AND** supports all existing WHERE operators (eq, ne, gt, lt, in, like, AND, OR, etc.)

#### Scenario: Empty result set

- **WHEN** an aggregate query matches zero records (due to WHERE filter or empty table)
- **THEN** the system returns `{ _count: { _all: 0 } }` for count operations
- **AND** other aggregates return null (e.g., `{ _avg: { field: null } }`)
- **AND** this allows differentiating between true aggregate value of zero and no data
- **AND** no error is thrown

#### Scenario: All field values are null

- **WHEN** aggregate operations are performed on a field where all values are null
- **THEN** `_count: { field: true }` returns the count of non-null values (0 in this case)
- **AND** `_count: { _all: true }` returns the total number of rows
- **AND** `_avg`, `_sum`, `_min`, `_max` return null for that field
- **AND** this distinguishes between no rows vs rows with null values

### Requirement: Type Safety for Aggregate Operations

The GraphQL schema generator SHALL enforce type constraints on aggregate operations at schema generation time.

#### Scenario: Numeric aggregates only on numeric fields

- **WHEN** generating the aggregate input schema
- **THEN** `_avg`, `_sum`, `_min`, `_max` input fields are ONLY generated for numeric field types (int, float, decimal)
- **AND** attempting to use numeric aggregates on string/boolean/date fields is prevented by schema validation
- **AND** the GraphQL schema does not include invalid field options

#### Scenario: Count supports all field types

- **WHEN** generating the aggregate count input schema
- **THEN** all model fields are available for count operations regardless of type
- **AND** `_all` special field is always available for COUNT(*)

#### Scenario: Invalid field type error

- **WHEN** the schema includes a field that is not numeric
- **THEN** that field SHALL NOT appear in `_avg`, `_sum`, `_min`, `_max` input types
- **AND** only appears in `_count` input type

### Requirement: Aggregate Input Structure

The aggregate query SHALL accept an input object with the following structure.

#### Scenario: Aggregate input format

- **WHEN** calling the aggregate operation
- **THEN** the input accepts an optional `where` parameter for filtering
- **AND** accepts `_count` parameter as an object with field names set to true
- **AND** accepts `_avg` parameter as an object with numeric field names set to true
- **AND** accepts `_sum` parameter as an object with numeric field names set to true
- **AND** accepts `_min` parameter as an object with field names set to true
- **AND** accepts `_max` parameter as an object with field names set to true

#### Scenario: At least one aggregate required

- **WHEN** calling aggregate with no aggregate operations selected
- **THEN** the system returns a validation error
- **AND** error message indicates at least one aggregate operation must be specified

#### Scenario: Example query structure

- **WHEN** a valid aggregate query is submitted
- **THEN** it matches the following format:
```graphql
aggregate(
  where: { status: { eq: "active" } }
) {
  _count { _all, users }
  _avg { amount, quantity }
  _sum { amount }
  _min { createdAt }
  _max { createdAt }
}
```

### Requirement: SQL Generation for Aggregates

The system SHALL generate efficient SQL aggregate queries using the goqu query builder.

#### Scenario: COUNT(*) generation

- **WHEN** `_count: { _all: true }` is requested
- **THEN** SQL includes `COUNT(*) as _count__all`
- **AND** uses double underscore to separate nested _all field

#### Scenario: COUNT(field) generation

- **WHEN** `_count: { fieldName: true }` is requested
- **THEN** SQL includes `COUNT(fieldName) as _count_fieldName`
- **AND** uses single underscore separator for field names

#### Scenario: Multiple aggregate functions

- **WHEN** multiple aggregate operations are requested
- **THEN** SQL includes all aggregate functions in a single SELECT statement
- **AND** example: `SELECT COUNT(*) as _count__all, AVG(amount) as _avg_amount, SUM(amount) as _sum_amount, MIN(amount) as _min_amount, MAX(amount) as _max_amount FROM table`

#### Scenario: WHERE clause integration

- **WHEN** aggregate query includes WHERE conditions
- **THEN** SQL includes WHERE clause using existing `convertWhereToExpression()` function
- **AND** prepared statements are used to prevent SQL injection
- **AND** complex operators (AND, OR, nested conditions) are supported

#### Scenario: Result mapping to nested structure

- **WHEN** SQL result is returned with flat column names
- **THEN** the system maps columns back to nested GraphQL structure
- **AND** `_count__all` maps to `_count._all`
- **AND** `_avg_amount` maps to `_avg.amount`
- **AND** maintains type safety (integers for count, floats for avg/sum)

### Requirement: Error Handling

The aggregate operation SHALL handle errors consistently with existing query operations.

#### Scenario: Database connection error

- **WHEN** database connection fails during aggregate query
- **THEN** the system returns a GraphQL error with appropriate error code
- **AND** error is logged with context information
- **AND** follows existing error handling patterns from findMany/findFirst

#### Scenario: Invalid WHERE condition

- **WHEN** WHERE clause contains invalid operators or values
- **THEN** the system returns a validation error before executing SQL
- **AND** error message describes the validation failure

#### Scenario: Field does not exist

- **WHEN** aggregate query references a non-existent field
- **THEN** GraphQL schema validation prevents the query from being submitted
- **AND** client receives schema validation error

### Requirement: Performance and Limits

The aggregate operation SHALL execute efficiently and respect system limits.

#### Scenario: Single query execution

- **WHEN** multiple aggregate operations are requested in one query
- **THEN** the system executes exactly one SQL query
- **AND** database performs all aggregations in a single pass

#### Scenario: No artificial limits

- **WHEN** a user selects many aggregate fields (e.g., avg/sum/min/max on 20 fields)
- **THEN** the system processes all requested aggregates without arbitrary limits
- **AND** performance is limited only by database capabilities

#### Scenario: Large table handling

- **WHEN** aggregate query runs on a large table
- **THEN** the system executes the query without loading all rows into memory
- **AND** database engine handles aggregation efficiently using indexes when available
- **AND** query timeout follows existing timeout configuration

