# query-api Specification Delta

## NEW Requirements

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
      "reqId": "01930c8a-9abc-7890-abcd-ef1234567890"
    }
  }
}
```

#### Scenario: findMany returns wrapped result with items field

- **WHEN** executing a findMany query
- **THEN** the response SHALL be wrapped in a result object
- **AND** the result SHALL contain an `items` field with an array of matching records (never null)
- **AND** the `items` array MAY be empty if no records match
- **AND** the result SHALL contain a `timeCost` field with query execution time
- **AND** the result SHALL contain a `reqId` field with a unique request ID

**Example**:
```graphql
query {
  findMany(where: {age: {gte: 18}}, take: 10) {
    items {
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
    "findMany": {
      "items": [
        {"id": "1", "name": "User 1", "age": 25},
        {"id": "2", "name": "User 2", "age": 30}
      ],
      "timeCost": 23,
      "reqId": "01930c8a-def0-7890-abcd-ef1234567890"
    }
  }
}
```

#### Scenario: findMany returns empty items array when no matches

- **WHEN** executing a findMany query that matches no records
- **THEN** the response SHALL contain a result object
- **AND** the `items` field SHALL be an empty array (not null)
- **AND** metadata fields (`timeCost`, `reqId`) SHALL still be present

**Example Response**:
```json
{
  "data": {
    "findMany": {
      "items": [],
      "timeCost": 10,
      "reqId": "01930c8a-1111-7890-abcd-ef1234567890"
    }
  }
}
```

### Requirement: Semantic Field Naming for Data Payload

Query operations SHALL use semantically appropriate field names to distinguish singular vs plural results.

#### Scenario: Singular operations use 'item' field

- **WHEN** a query operation returns a single record (findUnique, findFirst)
- **THEN** the data payload field SHALL be named `item`
- **AND** the field type SHALL be the model type (nullable)
- **AND** this provides clear semantic indication of singular result

**GraphQL Type Example**:
```graphql
type UserFindUniqueResult {
  item: User              # Singular, nullable
  timeCost: Int!
  reqId: String!
}
```

#### Scenario: Plural operations use 'items' field

- **WHEN** a query operation returns multiple records (findMany)
- **THEN** the data payload field SHALL be named `items`
- **AND** the field type SHALL be a non-null array of non-null model types
- **AND** this provides clear semantic indication of plural result

**GraphQL Type Example**:
```graphql
type UserFindManyResult {
  items: [User!]!         # Plural, non-null array
  timeCost: Int!
  reqId: String!
}
```

#### Scenario: Naming consistency across operations

- **WHEN** developers use query operations
- **THEN** field naming SHALL be consistent and predictable
- **AND** `item` SHALL always indicate singular result
- **AND** `items` SHALL always indicate plural result
- **AND** naming follows natural English singular/plural forms

### Requirement: Query Execution Time Metadata

All query operations SHALL include execution time metadata in the response.

#### Scenario: timeCost field measures query execution time

- **WHEN** any query operation executes
- **THEN** the system SHALL measure the time from query start to result return
- **AND** the `timeCost` field SHALL contain this duration in milliseconds
- **AND** the value SHALL be a non-null integer
- **AND** the value SHALL be >= 0

#### Scenario: timeCost reflects actual query execution

- **WHEN** calculating timeCost
- **THEN** the measurement SHALL include database query time
- **AND** the measurement SHALL include result processing time
- **AND** the measurement SHALL NOT include network transmission time
- **AND** the timestamp SHALL be recorded immediately before query execution
- **AND** the duration SHALL be calculated immediately after result retrieval

#### Scenario: timeCost is consistent across operations

- **WHEN** comparing timeCost across different query operations
- **THEN** the measurement methodology SHALL be identical
- **AND** values SHALL be comparable for performance analysis
- **AND** units SHALL always be milliseconds (not microseconds or seconds)

### Requirement: Request Tracking ID Metadata

All query operations SHALL include a unique request tracking ID in the response.

#### Scenario: reqId field contains unique identifier

- **WHEN** any query operation executes
- **THEN** the response SHALL include a `reqId` field
- **AND** the value SHALL be a unique identifier for this request
- **AND** the value SHALL be a non-null string
- **AND** the format SHALL be UUID version 7

#### Scenario: reqId enables request tracing

- **WHEN** a request is processed
- **THEN** the same reqId SHALL be used in all logs for this request
- **AND** the reqId SHALL be available in error responses
- **AND** the reqId SHALL enable correlation across distributed systems
- **AND** the reqId SHALL be unique across all requests (collision-resistant)

#### Scenario: reqId is generated early in request lifecycle

- **WHEN** a GraphQL request is received
- **THEN** the reqId SHALL be generated before query execution
- **AND** the reqId SHALL be injected into the request context
- **AND** all resolvers SHALL have access to the reqId via context
- **AND** the same reqId SHALL be used for all operations in a batched query

#### Scenario: UUID v7 format for reqId

- **WHEN** generating a reqId
- **THEN** the system SHALL use UUID version 7 format
- **AND** the UUID SHALL be sortable by creation time
- **AND** the UUID SHALL include timestamp component for debugging
- **AND** the UUID SHALL be generated using cryptographically secure random source

**Example reqId**: `01930c8a-1234-7890-abcd-ef1234567890`

### Requirement: GraphQL Result Type Generation

The GraphQL schema generator SHALL create result wrapper types for each model and query operation.

#### Scenario: Generate result types for each model

- **WHEN** generating GraphQL schema for a model
- **THEN** the system SHALL create three result types:
  - `{Model}FindUniqueResult`
  - `{Model}FindFirstResult`
  - `{Model}FindManyResult`
- **AND** each type SHALL be unique per model
- **AND** type names SHALL follow consistent naming pattern

**Example for User model**:
```graphql
type UserFindUniqueResult {
  item: User
  timeCost: Int!
  reqId: String!
}

type UserFindFirstResult {
  item: User
  timeCost: Int!
  reqId: String!
}

type UserFindManyResult {
  items: [User!]!
  timeCost: Int!
  reqId: String!
}
```

#### Scenario: Result types implement common interface

- **WHEN** defining result types
- **THEN** all result types SHALL implement a `QueryResultMeta` interface
- **AND** the interface SHALL define common metadata fields
- **AND** this enables generic handling of query results

**Interface definition**:
```graphql
interface QueryResultMeta {
  timeCost: Int!
  reqId: String!
}

type UserFindUniqueResult implements QueryResultMeta {
  item: User
  timeCost: Int!
  reqId: String!
}
```

#### Scenario: Field descriptions document metadata

- **WHEN** generating result type fields
- **THEN** each field SHALL include a description
- **AND** `item` description SHALL indicate "Single matching record (nullable)"
- **AND** `items` description SHALL indicate "Array of matching records"
- **AND** `timeCost` description SHALL indicate "Query execution time in milliseconds"
- **AND** `reqId` description SHALL indicate "Unique request tracking ID (UUID v7)"

#### Scenario: Query operation signatures updated

- **WHEN** generating Query type
- **THEN** operation signatures SHALL use result types instead of direct model types
- **AND** findUnique SHALL return `{Model}FindUniqueResult`
- **AND** findFirst SHALL return `{Model}FindFirstResult`
- **AND** findMany SHALL return `{Model}FindManyResult`

**Example Query type**:
```graphql
type Query {
  findUnique(where: UniqueWhereInput!): UserFindUniqueResult
  findFirst(where: WhereInput, orderBy: [OrderByInput!]): UserFindFirstResult
  findMany(where: WhereInput, orderBy: [OrderByInput!], take: Int, skip: Int): UserFindManyResult
}
```

### Requirement: Error Response Metadata

Error responses SHALL include metadata fields for debugging and tracing.

#### Scenario: Errors include reqId in extensions

- **WHEN** a query operation fails with an error
- **THEN** the GraphQL error response SHALL include a `reqId` field in extensions
- **AND** the reqId SHALL match the request tracking ID
- **AND** this enables correlation between error and logs

**Example error response**:
```json
{
  "data": {
    "findUnique": null
  },
  "errors": [
    {
      "message": "Record not found",
      "path": ["findUnique"],
      "extensions": {
        "code": "NOT_FOUND",
        "reqId": "01930c8a-1234-7890-abcd-ef1234567890",
        "timeCost": 5
      }
    }
  ]
}
```

#### Scenario: Errors include timeCost in extensions

- **WHEN** a query operation fails after partial execution
- **THEN** the error response SHALL include a `timeCost` field in extensions
- **AND** the timeCost SHALL reflect time until error occurred
- **AND** this helps diagnose timeout and performance issues

## ADDED Requirements

### Requirement: Request Context Metadata Package

The system SHALL provide a request context package for managing request-scoped metadata.

#### Scenario: RequestMetadata struct definition

- **WHEN** creating a new request context
- **THEN** the system SHALL define a `RequestMetadata` struct
- **AND** the struct SHALL contain a `ReqID` field (string, UUID v7)
- **AND** the struct SHALL contain a `StartTime` field (time.Time)
- **AND** the struct SHALL be stored in Go context

**Example**:
```go
type RequestMetadata struct {
    ReqID     string    // UUID v7
    StartTime time.Time // Request start timestamp
}
```

#### Scenario: Context injection function

- **WHEN** a new request is received
- **THEN** the system SHALL provide a `WithMetadata(ctx)` function
- **AND** the function SHALL generate a new RequestMetadata with unique reqID
- **AND** the function SHALL record current time as StartTime
- **AND** the function SHALL inject metadata into context and return new context

#### Scenario: Context retrieval function

- **WHEN** a resolver needs access to request metadata
- **THEN** the system SHALL provide a `GetMetadata(ctx)` function
- **AND** the function SHALL extract RequestMetadata from context
- **AND** the function SHALL return nil if metadata not present (defensive)

#### Scenario: TimeCost calculation helper

- **WHEN** calculating query execution time
- **THEN** the system SHALL provide a `CalculateTimeCost(ctx)` function
- **AND** the function SHALL retrieve metadata from context
- **AND** the function SHALL calculate duration from StartTime to now
- **AND** the function SHALL return duration in milliseconds as integer

### Requirement: GraphQL Middleware for Metadata Injection

The GraphQL handler SHALL inject request metadata before query execution.

#### Scenario: Middleware injects metadata at request start

- **WHEN** a GraphQL request is received
- **THEN** middleware SHALL generate a unique reqID (UUID v7)
- **AND** middleware SHALL record the current timestamp
- **AND** middleware SHALL create RequestMetadata and inject into context
- **AND** the enriched context SHALL be passed to all resolvers

#### Scenario: Metadata available to all resolvers

- **WHEN** any resolver executes
- **THEN** the resolver SHALL have access to RequestMetadata via context
- **AND** the resolver CAN retrieve reqId using GetMetadata(ctx)
- **AND** the resolver CAN calculate timeCost using CalculateTimeCost(ctx)
- **AND** the metadata SHALL be consistent across all operations in the same request

### Requirement: Resolver Response Wrapping

Query resolvers SHALL wrap results with data payload and metadata fields.

#### Scenario: executeFindUnique wraps result

- **WHEN** executeFindUnique completes a query
- **THEN** the resolver SHALL wrap the result in a map containing:
  - `item` field with the retrieved record (or null)
  - `timeCost` field with execution duration
  - `reqId` field from request context
- **AND** the resolver SHALL return the wrapped map

#### Scenario: executeFindFirst wraps result

- **WHEN** executeFindFirst completes a query
- **THEN** the resolver SHALL wrap the result in a map containing:
  - `item` field with the first matching record (or null)
  - `timeCost` field with execution duration
  - `reqId` field from request context
- **AND** the resolver SHALL return the wrapped map

#### Scenario: executeFindMany wraps result

- **WHEN** executeFindMany completes a query
- **THEN** the resolver SHALL wrap the result in a map containing:
  - `items` field with array of matching records (never null)
  - `timeCost` field with execution duration
  - `reqId` field from request context
- **AND** the resolver SHALL return the wrapped map
- **AND** if no records match, `items` SHALL be an empty array

### Requirement: Backward Compatibility Breaking Change

This change SHALL be documented as a breaking change.

#### Scenario: Breaking change documented

- **WHEN** this change is deployed
- **THEN** the breaking change SHALL be documented in changelog
- **AND** all affected operations SHALL be listed:
  - findUnique
  - findFirst
  - findMany
- **AND** each operation's response structure change SHALL be documented
- **AND** before/after examples SHALL be provided

## Future Extensibility

### Note: Optional totalCount field for findMany

The `UserFindManyResult` type MAY include an optional `totalCount` field in future versions:

```graphql
type UserFindManyResult {
  items: [User!]!
  totalCount: Int         # Optional: total matching records (not implemented yet)
  timeCost: Int!
  reqId: String!
}
```

This field is reserved for future pagination support and SHALL NOT be populated in the initial implementation.

### Note: Migration to Relay Connection pattern

If cursor-based pagination is required in the future, the result types CAN be extended to include:
- `edges` field with cursor information
- `pageInfo` field with pagination metadata

This extension SHALL be non-breaking (additive only).
