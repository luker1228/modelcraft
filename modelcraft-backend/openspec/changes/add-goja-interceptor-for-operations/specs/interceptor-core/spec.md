# interceptor-core Specification

## Purpose
Define core interceptor functionality including script execution engine, context management, and rule configuration for ModelCraft's dynamic authorization and validation system using Goja JavaScript engine.

## Why
Interceptors enable dynamic authorization and validation without code redeployment. Using JavaScript via Goja provides AI-friendly rule authoring, mature ecosystem support, and sufficient performance for authorization decisions (<5ms typical overhead).

## ADDED Requirements

### Requirement: Interceptor Rule Configuration

The system SHALL provide a configuration mechanism for defining JavaScript-based interceptor rules that execute before GraphQL operations.

#### Scenario: Create interceptor rule for specific model and operation

- **WHEN** an administrator creates an interceptor rule
- **THEN** the rule SHALL be associated with a specific model ID
- **AND** the rule SHALL specify which operation(s) it applies to (findFirst, findMany, findUnique, createOne, updateOne, deleteOne, createMany, updateMany, deleteMany, aggregate, count)
- **AND** the rule SHALL contain valid JavaScript code
- **AND** the rule SHALL have a priority value for execution ordering (higher priority executes first)
- **AND** the rule SHALL have an enabled/disabled flag
- **AND** the rule SHALL have a configurable timeout in milliseconds (default: 100ms)

**Example Configuration**:
```json
{
  "modelId": 123,
  "operation": "findMany",
  "script": "function intercept() { input.where.tenantId = user.tenantId; return input; }",
  "priority": 10,
  "enabled": true,
  "timeoutMs": 100
}
```


#### Scenario: Disabled interceptors are skipped

- **WHEN** an interceptor rule has `enabled: false`
- **THEN** the interceptor SHALL NOT execute
- **AND** the rule SHALL remain in the database for future re-enabling
- **AND** no performance overhead is incurred for disabled interceptors

---

### Requirement: JavaScript Execution Engine

The system SHALL use Goja to execute JavaScript interceptor scripts with safety controls, performance optimizations, and resource limits.

#### Scenario: VM pool reuses runtime instances

- **WHEN** an interceptor needs to execute
- **THEN** the system SHALL acquire a Goja Runtime from a sync.Pool
- **AND** after execution completes, the Runtime SHALL be returned to the pool
- **AND** pool reuse rate SHALL exceed 90% under normal load
- **AND** memory usage SHALL remain stable (no leaks from VM reuse)

**Implementation Pattern**:
```go
var vmPool = sync.Pool{
    New: func() interface{} {
        vm := goja.New()
        // Disable dangerous globals
        vm.Set("eval", goja.Undefined())
        vm.Set("Function", goja.Undefined())
        return vm
    },
}

func executeInterceptor(script string, ctx *InterceptorContext) (*InterceptorResult, error) {
    vm := vmPool.Get().(*goja.Runtime)
    defer vmPool.Put(vm)
    // ... execution logic
}
```

#### Scenario: Script compilation is cached

- **WHEN** an interceptor script is first loaded
- **THEN** the script SHALL be compiled into a goja.Program
- **AND** the compiled Program SHALL be cached in memory
- **AND** subsequent executions SHALL reuse the compiled Program
- **AND** cache invalidation SHALL occur when script content changes
- **AND** cache hit rate SHALL exceed 95% for stable scripts

**Performance Impact**:
- First execution: ~5-10ms (parsing + compilation + execution)
- Cached execution: ~0.5-2ms (execution only)

#### Scenario: Timeout prevents infinite loops

- **WHEN** an interceptor script executes longer than the configured timeout
- **THEN** the Goja Runtime SHALL be interrupted
- **AND** an error SHALL be returned to the caller
- **AND** the operation SHALL fail (strict mode)
- **AND** the error message SHALL indicate timeout and script location

**Example Timeout Handling**:
```go
done := make(chan struct{})
var result goja.Value
var err error

go func() {
    result, err = vm.RunProgram(compiledScript)
    close(done)
}()

select {
case <-done:
    return parseResult(result, err)
case <-time.After(timeoutMs * time.Millisecond):
    vm.Interrupt("execution timeout")
    return nil, ErrScriptTimeout
}
```

#### Scenario: Sandboxing prevents dangerous operations

- **WHEN** a Goja Runtime is created for interceptor execution
- **THEN** the `eval` function SHALL be disabled (set to undefined)
- **AND** the `Function` constructor SHALL be disabled
- **AND** access to Go runtime internals SHALL be prevented
- **AND** only explicitly provided context variables SHALL be accessible

**Security Restrictions**:
- ❌ No `eval("malicious code")`
- ❌ No `new Function("return process")` injection
- ❌ No access to filesystem, network, or syscalls
- ✅ Only context variables: `user`, `resource`, `environment`, `input`

#### Scenario: JavaScript errors include stack traces

- **WHEN** an interceptor script throws an error or exception
- **THEN** the error message SHALL include the JavaScript error message
- **AND** the error SHALL include a stack trace showing line numbers
- **AND** the stack trace SHALL map to the original script (not wrapped code)
- **AND** the error SHALL be logged with full context for debugging

**Example Error Output**:
```
InterceptorError: TypeError: Cannot read property 'id' of undefined
    at intercept (interceptor_script.js:5:23)
    at <anonymous>:1:1
Script: tenant-filter (Model: User, Operation: findMany)
```

---

### Requirement: Interceptor Context Injection

The system SHALL provide structured context information to JavaScript interceptors including user attributes, resource metadata, environment data, and operation input.

#### Scenario: User context contains identity and permissions

- **WHEN** an interceptor script executes
- **THEN** a `user` object SHALL be available in the JavaScript scope
- **AND** the user object SHALL contain `id`, `roles`, `permissions`, and `metadata` fields
- **AND** roles and permissions SHALL be arrays of strings
- **AND** metadata SHALL contain additional user attributes as key-value pairs

**User Context Structure**:
```javascript
{
  id: "user_123",
  roles: ["developer", "admin"],
  permissions: ["read:users", "write:users", "delete:users"],
  metadata: {
    tenantId: "tenant_456",
    department: "engineering",
    email: "user@example.com"
  }
}
```

#### Scenario: Resource context describes the target model

- **WHEN** an interceptor script executes
- **THEN** a `resource` object SHALL be available in the JavaScript scope
- **AND** the resource object SHALL contain `modelName`, `operation`, and `fields` information
- **AND** `modelName` identifies the target table/model
- **AND** `operation` specifies the GraphQL operation (findMany, createOne, etc.)
- **AND** `fields` lists the fields being queried or modified

**Resource Context Structure**:
```javascript
{
  modelName: "User",
  operation: "findMany",
  fields: ["id", "name", "email", "department"]
}
```

#### Scenario: Environment context provides request metadata

- **WHEN** an interceptor script executes
- **THEN** an `environment` object SHALL be available in the JavaScript scope
- **AND** the environment object SHALL contain `timestamp`, `requestId`, and `ipAddress`
- **AND** `timestamp` is the request start time (ISO 8601 format)
- **AND** `requestId` is the unique request tracking ID (UUID v7)
- **AND** `ipAddress` is the client IP address (if available)

**Environment Context Structure**:
```javascript
{
  timestamp: "2025-01-06T10:30:00Z",
  requestId: "01930c8a-1234-7890-abcd-ef1234567890",
  ipAddress: "192.168.1.100"
}
```

#### Scenario: Input context contains operation arguments

- **WHEN** an interceptor script executes
- **THEN** an `input` object SHALL be available in the JavaScript scope
- **AND** for query operations, `input.where` contains the WHERE clause filter
- **AND** for mutation operations, `input.data` contains the data payload
- **AND** for operations with pagination, `input.skip` and `input.take` are available
- **AND** the interceptor can modify any field in the input object

**Input Context Examples**:

**Query Operation (findMany)**:
```javascript
{
  where: {
    status: { equals: "active" },
    age: { gte: 18 }
  },
  skip: 0,
  take: 10,
  orderBy: { createdAt: "desc" }
}
```

**Mutation Operation (createOne)**:
```javascript
{
  data: {
    name: "John Doe",
    email: "john@example.com",
    role: "developer"
  }
}
```

#### Scenario: Helper functions are pre-loaded

- **WHEN** an interceptor script executes
- **THEN** helper functions SHALL be available in the global scope
- **AND** `contains(array, item)` checks if array contains item
- **AND** `hasPermission(user, permission)` checks if user has specific permission
- **AND** `isInDepartment(user, departments)` checks if user is in one of the departments
- **AND** custom helper functions can be added to the engine configuration

**Helper Function Examples**:
```javascript
// Check permission
if (!hasPermission(user, "delete:users")) {
    throw new Error("Permission denied");
}

// Check department
if (!isInDepartment(user, ["engineering", "admin"])) {
    throw new Error("Department restriction");
}

// Check array membership
if (!contains(user.roles, "admin")) {
    throw new Error("Admin role required");
}
```

---

### Requirement: Interceptor Result Handling

The system SHALL process interceptor script results to modify operation input or deny execution based on script output.

#### Scenario: Interceptor returns modified input

- **WHEN** an interceptor script completes successfully
- **THEN** the script SHALL return the `input` object (potentially modified)
- **AND** the modified input SHALL replace the original operation input
- **AND** the modified input SHALL be passed to the next interceptor (if any) or to the repository
- **AND** modified WHERE clauses SHALL be validated before SQL generation
- **AND** modified data payloads SHALL be validated against field types

**Example Script**:
```javascript
function intercept() {
    // Inject tenant filter
    if (!input.where) {
        input.where = {};
    }
    input.where.tenantId = { equals: user.metadata.tenantId };

    // Return modified input
    return input;
}

intercept();
```

#### Scenario: Interceptor denies operation by throwing error

- **WHEN** an interceptor script throws an error or exception
- **THEN** the operation SHALL be aborted immediately
- **AND** subsequent interceptors SHALL NOT execute
- **AND** the repository operation SHALL NOT be called
- **AND** a GraphQL error SHALL be returned to the client
- **AND** the error message SHALL include the interceptor's error message

**Example Script**:
```javascript
function intercept() {
    // Check permission
    if (!hasPermission(user, "read:sensitive_data")) {
        throw new Error("Access denied: insufficient permissions");
    }

    // Check time window
    var hour = new Date(environment.timestamp).getHours();
    if (hour < 9 || hour > 18) {
        throw new Error("Access denied: outside business hours");
    }

    return input;
}

intercept();
```

#### Scenario: Interceptor returns explicit denial object

- **WHEN** an interceptor script returns an object with `{ allowed: false, reason: "..." }`
- **THEN** the operation SHALL be denied
- **AND** the `reason` field SHALL be included in the error message
- **AND** this pattern is alternative to throwing errors (same effect)

**Example Script**:
```javascript
function intercept() {
    if (user.metadata.department !== resource.metadata.department) {
        return {
            allowed: false,
            reason: "Cross-department access denied"
        };
    }

    return input;
}

intercept();
```

---

### Requirement: Error Handling and Logging

The system SHALL handle interceptor failures gracefully and provide detailed logging for debugging and auditing.

#### Scenario: Syntax errors in scripts are detected at load time

- **WHEN** an interceptor rule is created with invalid JavaScript syntax
- **THEN** the script compilation SHALL fail immediately
- **AND** an error SHALL be returned to the caller
- **AND** the invalid rule SHALL NOT be saved to the database
- **AND** the error message SHALL indicate the syntax error location

**Example Validation**:
```go
func ValidateInterceptorScript(script string) error {
    _, err := goja.Compile("validation", script, true)
    if err != nil {
        return fmt.Errorf("syntax error: %w", err)
    }
    return nil
}
```

#### Scenario: Runtime errors abort operation and log details

- **WHEN** an interceptor script throws a runtime error during execution
- **THEN** the operation SHALL be aborted (strict mode)
- **AND** an error log entry SHALL be created with:
  - Request ID
  - User ID
  - Model name and operation
  - Script ID and version
  - Error message and stack trace
  - Execution time before failure
- **AND** the client SHALL receive a GraphQL error response

**Log Entry Example**:
```json
{
  "level": "error",
  "timestamp": "2025-01-06T10:30:15Z",
  "requestId": "01930c8a-1234-7890-abcd-ef1234567890",
  "userId": "user_123",
  "interceptor": {
    "id": 456,
    "modelName": "User",
    "operation": "findMany",
    "scriptHash": "a1b2c3d4"
  },
  "error": {
    "type": "TypeError",
    "message": "Cannot read property 'id' of undefined",
    "stack": "at intercept (script:5:23)",
    "executionTimeMs": 3
  }
}
```

#### Scenario: Timeout errors indicate slow scripts

- **WHEN** an interceptor script exceeds the configured timeout
- **THEN** a timeout error SHALL be logged
- **AND** the log SHALL include:
  - Script execution time (actual time spent)
  - Configured timeout value
  - Suggestion to optimize or increase timeout
- **AND** the client SHALL receive a timeout error message

**Timeout Error Message**:
```
Interceptor timeout: Script execution exceeded 100ms (Model: User, Operation: findMany)
Suggestion: Optimize script or increase timeout configuration
```

---

### Requirement: Performance Monitoring

The system SHALL collect and expose metrics for interceptor execution to enable performance monitoring and optimization.

#### Scenario: Execution time is measured per interceptor

- **WHEN** an interceptor executes
- **THEN** the execution time SHALL be measured in milliseconds
- **AND** execution time metrics SHALL be aggregated per model + operation
- **AND** metrics SHALL include: min, max, avg, p50, p95, p99 latencies
- **AND** metrics SHALL be exposed via monitoring endpoints (e.g., Prometheus)

**Metrics Schema**:
```
interceptor_execution_duration_ms{model="User", operation="findMany", interceptor_id="123"} histogram
interceptor_execution_count{model="User", operation="findMany", status="success|error|timeout"} counter
interceptor_cache_hit_rate{type="compiled_script"} gauge
interceptor_vm_pool_size gauge
```

#### Scenario: High latency triggers alerts

- **WHEN** interceptor execution time exceeds a threshold (e.g., 10ms)
- **THEN** a warning log entry SHALL be created
- **AND** the log SHALL include script ID and execution time
- **AND** if p95 latency exceeds threshold for 5 minutes, an alert SHALL be triggered
- **AND** alerts SHALL notify operations team for investigation

#### Scenario: Cache hit rates are tracked

- **WHEN** the system retrieves a compiled script
- **THEN** cache hit/miss events SHALL be recorded
- **AND** cache hit rate SHALL be calculated (hits / (hits + misses))
- **AND** low cache hit rates (< 90%) SHALL trigger optimization review
- **AND** cache size and eviction events SHALL be monitored

---

### Requirement: Repository Interface

The system SHALL define a repository interface for persisting and retrieving interceptor rules.

#### Scenario: Repository supports CRUD operations

- **WHEN** the system needs to manage interceptor rules
- **THEN** the repository interface SHALL provide:
  - `Create(rule *InterceptorRule) error`
  - `GetByID(id int64) (*InterceptorRule, error)`
  - `GetByModelAndOperation(modelID int64, operation string) ([]*InterceptorRule, error)`
  - `GetEnabled(modelID int64, operation string) ([]*InterceptorRule, error)`
  - `Update(rule *InterceptorRule) error`
  - `Delete(id int64) error`

#### Scenario: GetEnabled returns sorted interceptors

- **WHEN** `GetEnabled(modelID, operation)` is called
- **THEN** only rules with `enabled: true` SHALL be returned
- **AND** rules SHALL be sorted by priority in descending order (highest first)
- **AND** query results SHALL use database indexes for efficient lookup

**SQL Query Example**:
```sql
SELECT * FROM interceptor_rules
WHERE model_id = ? AND operation = ? AND enabled = true
ORDER BY priority DESC, id ASC;
```
