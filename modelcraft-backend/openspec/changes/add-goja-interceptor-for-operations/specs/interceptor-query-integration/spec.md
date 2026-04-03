# interceptor-query-integration Specification

## Purpose
Define how JavaScript interceptors integrate with GraphQL query operations (findFirst, findMany, findUnique, aggregate, count) to modify WHERE clauses and implement row-level security and data filtering.

## Why
Query interceptors enable row-level security by injecting tenant filters, department restrictions, and access control rules into WHERE clauses dynamically. This allows multi-tenant applications and ABAC systems to enforce data isolation without modifying application code.

## Related Specifications
- **interceptor-core**: Defines core interceptor execution engine
- **query-api**: Defines base query operation semantics

## ADDED Requirements

### Requirement: Interceptor Integration for findUnique Operation

The findUnique operation SHALL support interceptor execution to modify the WHERE clause before database query execution.

#### Scenario: Interceptor modifies WHERE clause in findUnique

- **WHEN** a findUnique query is executed and an interceptor is configured
- **THEN** the interceptor SHALL receive `input.where` containing the unique identifier condition
- **AND** the interceptor can add additional conditions to the WHERE clause
- **AND** the modified WHERE clause SHALL be passed to the repository FindUnique method
- **AND** the query SHALL only return a record if it matches ALL conditions (original + injected)

**Example**:

**Original Query**:
```graphql
query {
  findUnique(where: { id: "123" }) {
    item { id, name, email }
  }
}
```

**Interceptor Script** (tenant filtering):
```javascript
function intercept() {
    // Add tenant filter
    input.where.tenantId = { equals: user.metadata.tenantId };
    return input;
}

intercept();
```

**Effective WHERE Clause**:
```javascript
{
  id: "123",
  tenantId: { equals: "tenant_456" }
}
```

**SQL Generated**:
```sql
SELECT * FROM users WHERE id = '123' AND tenant_id = 'tenant_456' LIMIT 1;
```

#### Scenario: Interceptor denies findUnique based on permissions

- **WHEN** an interceptor determines the user lacks permission to view the specific record
- **THEN** the interceptor SHALL throw an error
- **AND** the FindUnique repository method SHALL NOT be called
- **AND** a GraphQL error SHALL be returned to the client
- **AND** the error message SHALL indicate permission denial

**Example Script**:
```javascript
function intercept() {
    // Only allow users to view their own records
    if (input.where.id !== user.id && !hasPermission(user, "admin")) {
        throw new Error("Access denied: can only view your own user record");
    }
    return input;
}

intercept();
```

#### Scenario: findUnique returns null if interceptor filters out record

- **WHEN** an interceptor adds conditions that exclude the target record
- **THEN** the repository FindUnique method SHALL return no results
- **AND** the GraphQL response SHALL contain `item: null`
- **AND** this is semantically correct (record doesn't exist in user's accessible scope)

**Example**:
```
User requests: { id: "123" }
Interceptor adds: { department: "engineering" }
Database has: User 123 in "sales" department
Result: item = null (user 123 not visible to engineering users)
```

---

### Requirement: Interceptor Integration for findFirst Operation

The findFirst operation SHALL support interceptor execution to modify the WHERE clause and control which record is returned.

#### Scenario: Interceptor injects filters into findFirst WHERE clause

- **WHEN** a findFirst query is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.where` containing the original filter conditions
- **AND** the interceptor can add, modify, or remove conditions
- **AND** the modified WHERE clause SHALL be passed to the repository FindFirst method
- **AND** the database query SHALL return the first record matching the modified conditions

**Example**:

**Original Query**:
```graphql
query {
  findFirst(where: { status: { equals: "active" } }) {
    item { id, name, status }
  }
}
```

**Interceptor Script** (department filtering):
```javascript
function intercept() {
    // Restrict to user's department
    input.where.department = { equals: user.metadata.department };
    return input;
}

intercept();
```

**Effective WHERE Clause**:
```javascript
{
  status: { equals: "active" },
  department: { equals: "engineering" }
}
```

#### Scenario: Interceptor modifies orderBy to enforce access patterns

- **WHEN** an interceptor needs to control result ordering for security reasons
- **THEN** the interceptor can modify `input.orderBy`
- **AND** the modified orderBy SHALL be applied to the SQL query
- **AND** this enables enforcing "most recent accessible" or priority-based access patterns

**Example Script**:
```javascript
function intercept() {
    // Always return most recently updated record first
    input.orderBy = { updatedAt: "desc" };
    return input;
}

intercept();
```

---

### Requirement: Interceptor Integration for findMany Operation

The findMany operation SHALL support interceptor execution to modify the WHERE clause, enforce pagination limits, and implement row-level security filtering.

#### Scenario: Interceptor injects row-level security filters

- **WHEN** a findMany query is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.where` containing the original filter
- **AND** the interceptor can inject conditions to restrict visible rows
- **AND** the modified WHERE clause SHALL be applied to all records
- **AND** only records matching the modified conditions SHALL be returned

**Example**:

**Original Query**:
```graphql
query {
  findMany(where: { status: { equals: "published" } }) {
    items { id, title, authorId }
  }
}
```

**Interceptor Script** (author-only access):
```javascript
function intercept() {
    // Non-admin users can only see their own posts
    if (!hasPermission(user, "admin")) {
        input.where.authorId = { equals: user.id };
    }
    return input;
}

intercept();
```

**Effective WHERE Clause** (non-admin user):
```javascript
{
  status: { equals: "published" },
  authorId: { equals: "user_123" }
}
```

#### Scenario: Interceptor enforces pagination limits

- **WHEN** an interceptor needs to prevent large result sets
- **THEN** the interceptor can modify `input.take` to limit results
- **AND** the interceptor can set a maximum `take` value (e.g., 100 records)
- **AND** if the original `take` exceeds the limit, it SHALL be reduced
- **AND** this prevents performance issues from unbounded queries

**Example Script**:
```javascript
function intercept() {
    const MAX_RESULTS = 100;

    // Enforce maximum result limit
    if (!input.take || input.take > MAX_RESULTS) {
        input.take = MAX_RESULTS;
    }

    return input;
}

intercept();
```

#### Scenario: Interceptor injects complex AND/OR conditions

- **WHEN** an interceptor needs to enforce multi-condition filtering
- **THEN** the interceptor can modify the WHERE clause to use AND/OR/NOT operators
- **AND** the modified conditions SHALL follow the query-api specification for logical operators
- **AND** complex nested conditions are supported

**Example Script** (multi-tenancy with shared resources):
```javascript
function intercept() {
    // User can see resources owned by their tenant OR marked as public
    input.where = {
        OR: [
            { tenantId: { equals: user.metadata.tenantId } },
            { visibility: { equals: "public" } }
        ]
    };

    return input;
}

intercept();
```

**SQL Generated**:
```sql
SELECT * FROM resources
WHERE (tenant_id = 'tenant_456' OR visibility = 'public');
```

#### Scenario: Interceptor combines with existing WHERE clause

- **WHEN** a findMany query has both original WHERE conditions and interceptor-injected conditions
- **THEN** the conditions SHALL be combined using AND logic (both must be true)
- **AND** the original WHERE clause SHALL NOT be discarded
- **AND** the interceptor can use AND arrays to append conditions

**Example Script**:
```javascript
function intercept() {
    // Preserve original conditions and add tenant filter
    if (!input.where.AND) {
        input.where = { AND: [input.where] };
    }

    input.where.AND.push({
        tenantId: { equals: user.metadata.tenantId }
    });

    return input;
}

intercept();
```

**Original WHERE**:
```javascript
{ status: { equals: "active" } }
```

**After Interceptor**:
```javascript
{
  AND: [
    { status: { equals: "active" } },
    { tenantId: { equals: "tenant_456" } }
  ]
}
```

---

### Requirement: Interceptor Integration for aggregate Operation

The aggregate operation SHALL support interceptor execution to modify the WHERE clause before aggregation calculations.

#### Scenario: Interceptor filters data before aggregation

- **WHEN** an aggregate query is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.where` containing the original filter
- **AND** the interceptor can add conditions to restrict which records are aggregated
- **AND** aggregation functions (COUNT, AVG, SUM, MIN, MAX) SHALL operate only on filtered data
- **AND** row-level security filters SHALL be applied before aggregation

**Example**:

**Original Query**:
```graphql
query {
  aggregate(where: { status: { equals: "completed" } }) {
    _count { _all }
    _sum { amount }
    _avg { amount }
  }
}
```

**Interceptor Script** (department-scoped aggregation):
```javascript
function intercept() {
    // Only aggregate data from user's department
    input.where.department = { equals: user.metadata.department };
    return input;
}

intercept();
```

**Effective WHERE Clause**:
```javascript
{
  status: { equals: "completed" },
  department: { equals: "engineering" }
}
```

**SQL Generated**:
```sql
SELECT
  COUNT(*) as _all,
  SUM(amount) as amount_sum,
  AVG(amount) as amount_avg
FROM transactions
WHERE status = 'completed' AND department = 'engineering';
```

#### Scenario: Empty result set returns zero aggregations

- **WHEN** an interceptor adds conditions that filter out all records
- **THEN** the aggregate query SHALL return zero/null values
- **AND** `_count._all` SHALL be 0
- **AND** `_sum`, `_avg`, `_min`, `_max` SHALL be null (no data to aggregate)
- **AND** this is semantically correct (no accessible data in scope)

---

### Requirement: Interceptor Integration for count Operation

The count operation SHALL support interceptor execution to modify the WHERE clause before counting records.

#### Scenario: Interceptor filters records before counting

- **WHEN** a count query is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.where` containing the original filter
- **AND** the interceptor can add conditions to restrict which records are counted
- **AND** the COUNT operation SHALL only count records matching the modified WHERE clause

**Example**:

**Original Query**:
```graphql
query {
  count(where: { status: { equals: "active" } })
}
```

**Interceptor Script**:
```javascript
function intercept() {
    // Only count records in user's region
    input.where.region = { equals: user.metadata.region };
    return input;
}

intercept();
```

**SQL Generated**:
```sql
SELECT COUNT(*) as count
FROM users
WHERE status = 'active' AND region = 'us-west';
```

#### Scenario: Count with field-level selection respects interceptor filters

- **WHEN** a count query uses `select` parameter for field-level counts
- **THEN** the interceptor-modified WHERE clause SHALL apply to all field counts
- **AND** `_all`, field-specific counts SHALL operate on the filtered result set

**Example**:

**Original Query**:
```graphql
query {
  count(select: { _all: true, email: true, phoneNumber: true })
}
```

**Interceptor Script**:
```javascript
function intercept() {
    // Only count active users
    if (!input.where) {
        input.where = {};
    }
    input.where.status = { equals: "active" };
    return input;
}

intercept();
```

**Result**:
```json
{
  "fieldsCount": {
    "_all": 150,       // 150 active users
    "email": 145,      // 145 active users have email
    "phoneNumber": 120 // 120 active users have phone number
  }
}
```

---

### Requirement: Query Operation Error Handling

The system SHALL handle interceptor errors consistently across all query operations.

#### Scenario: Interceptor error aborts query execution

- **WHEN** an interceptor throws an error during a query operation
- **THEN** the repository query method SHALL NOT be called
- **AND** the operation SHALL return immediately with an error
- **AND** the GraphQL error response SHALL include the interceptor error message
- **AND** the response SHALL follow the standard error format

**Error Response Example**:
```json
{
  "errors": [
    {
      "message": "Interceptor error: Access denied - insufficient permissions",
      "extensions": {
        "code": "INTERCEPTOR_DENIED",
        "modelName": "User",
        "operation": "findMany",
        "interceptorId": 123
      }
    }
  ],
  "data": {
    "findMany": null
  }
}
```

#### Scenario: Timeout error in query interceptor

- **WHEN** an interceptor exceeds the timeout during a query operation
- **THEN** the operation SHALL be aborted
- **AND** the error message SHALL indicate timeout
- **AND** the client SHALL receive a GraphQL error
- **AND** the timeout SHALL be logged for investigation

---

### Requirement: Performance Impact on Query Operations

The system SHALL minimize performance overhead of interceptors on query operations.

#### Scenario: Interceptor overhead is measured and bounded

- **WHEN** a query operation executes with an interceptor
- **THEN** the interceptor execution time SHALL be measured separately from database query time
- **AND** total overhead (VM acquisition + script execution + result parsing) SHALL be < 5ms for typical interceptors
- **AND** if overhead exceeds 10ms, a performance warning SHALL be logged
- **AND** slow interceptors SHALL be identified for optimization

**Performance Breakdown**:
```
Total Query Time: 25ms
├── Interceptor Execution: 3ms (12%)
│   ├── VM Pool Acquisition: 0.1ms
│   ├── Context Setup: 0.5ms
│   ├── Script Execution: 2ms
│   └── Result Parsing: 0.4ms
└── Database Query: 22ms (88%)
```

#### Scenario: Query operations without interceptors have zero overhead

- **WHEN** a query operation is executed on a model with NO interceptors configured
- **THEN** the interceptor executor SHALL NOT be invoked
- **AND** performance SHALL be identical to pre-interceptor implementation
- **AND** no VM pool operations SHALL occur
- **AND** this ensures backward compatibility and no performance regression

---

### Requirement: Audit Logging for Query Operations

The system SHALL log interceptor actions on query operations for audit and compliance purposes.

#### Scenario: Successful interceptor modifications are logged

- **WHEN** an interceptor modifies the WHERE clause in a query operation
- **THEN** an audit log entry SHALL be created (optional, configurable)
- **AND** the log SHALL include:
  - User ID and request ID
  - Model name and operation
  - Original WHERE clause (before interception)
  - Modified WHERE clause (after interception)
  - Interceptor ID and script version
- **AND** this enables compliance auditing and security forensics

**Audit Log Example**:
```json
{
  "timestamp": "2025-01-06T10:30:00Z",
  "level": "info",
  "type": "interceptor_execution",
  "requestId": "01930c8a-1234-7890-abcd-ef1234567890",
  "userId": "user_123",
  "operation": {
    "modelName": "User",
    "operation": "findMany"
  },
  "interceptor": {
    "id": 456,
    "name": "tenant-filter",
    "version": "v1.2"
  },
  "changes": {
    "originalWhere": { "status": { "equals": "active" } },
    "modifiedWhere": {
      "status": { "equals": "active" },
      "tenantId": { "equals": "tenant_456" }
    }
  },
  "executionTimeMs": 2
}
```

#### Scenario: Denied queries are logged with reason

- **WHEN** an interceptor denies a query operation
- **THEN** an audit log entry SHALL be created
- **AND** the log SHALL include the denial reason
- **AND** this enables security monitoring and threat detection
