# Design Document: Goja Interceptor Architecture

## Overview
This document outlines the architectural design for integrating JavaScript-based interceptors using Goja into ModelCraft's GraphQL operations, enabling dynamic authorization, validation, and data transformation without code redeployment.

## Architectural Context

### System Boundaries
```
┌────────────────────────────────────────────────────────────┐
│                    GraphQL Runtime Layer                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          GraphQL Resolvers (model_resolver.go)       │  │
│  │  findFirst | findMany | createOne | updateOne | ...  │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                     │
│                       ▼                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Interceptor Executor (NEW)                   │  │
│  │  - Load interceptor rules                            │  │
│  │  - Execute JavaScript via Goja                       │  │
│  │  - Modify input (WHERE clause, data payload)         │  │
│  └────────────────────┬─────────────────────────────────┘  │
│                       │                                     │
│                       ▼                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │       Repository Layer (client_db_repo_impl.go)      │  │
│  │  SQL generation and database query execution         │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘
```

### Interceptor Components
```
┌─────────────────────────────────────────────────────────────┐
│                   Interceptor Subsystem                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┐     ┌──────────────────┐              │
│  │ InterceptorRule │────▶│ RuleRepository   │              │
│  │ (Domain Entity) │     │ (Infrastructure) │              │
│  └─────────────────┘     └──────────────────┘              │
│           │                                                  │
│           ▼                                                  │
│  ┌─────────────────────────────────┐                        │
│  │   ExecutorService               │                        │
│  │  - LoadRules(modelID, op)       │                        │
│  │  - Execute(ctx, input)          │                        │
│  │  - CacheCompiledScripts()       │                        │
│  └──────────┬──────────────────────┘                        │
│             │                                                │
│             ▼                                                │
│  ┌─────────────────────────────────┐                        │
│  │   GojaEngine (VM Pool)          │                        │
│  │  - sync.Pool of goja.Runtime    │                        │
│  │  - Timeout mechanism             │                        │
│  │  - Sandboxing (disable eval)    │                        │
│  │  - Helper functions injection   │                        │
│  └─────────────────────────────────┘                        │
│                                                              │
│  ┌─────────────────────────────────┐                        │
│  │   ContextBuilder                │                        │
│  │  - Build user context           │                        │
│  │  - Build resource context       │                        │
│  │  - Build environment context    │                        │
│  └─────────────────────────────────┘                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### Decision 1: Interceptor Execution Point (Before Repository Call)

**Problem**: Where in the request lifecycle should interceptors execute?

**Options Considered**:

1. **Before GraphQL Resolver** (HTTP middleware level)
   - ❌ Too early - GraphQL schema not parsed yet
   - ❌ Cannot access operation-specific input (WHERE, data)
   - ❌ Would require duplicating logic for each operation

2. **Inside Repository Layer** (database layer)
   - ❌ Too late - breaks separation of concerns
   - ❌ Repository should be a thin data access layer
   - ❌ Harder to test and maintain

3. **Between Resolver and Repository** (CHOSEN)
   - ✅ Has access to parsed GraphQL input
   - ✅ Can modify input before SQL generation
   - ✅ Clean separation: resolver → interceptor → repository
   - ✅ Easy to disable (skip interceptor call)

**Rationale**: Executing interceptors between the resolver and repository provides the optimal balance of access to operation context and clean architectural boundaries. The resolver parses GraphQL input into structured objects (FindManyInput, CreateOneInput), which are then modified by interceptors before being passed to the repository for SQL generation.

**Implementation Pattern**:
```go
func (m *graphqlModelResolver) executeFindMany(p graphql.ResolveParams) (map[string]any, error) {
    // 1. Parse GraphQL input
    input, err := newFindManyInput(m.model.Name, p)
    if err != nil {
        return nil, err
    }

    // 2. Execute interceptor (NEW)
    if m.interceptorExecutor != nil {
        modifiedInput, err := m.interceptorExecutor.Execute(
            m.ctx, m.model.ID, "findMany", input, buildContext(p),
        )
        if err != nil {
            return nil, err // Denied or error
        }
        input = modifiedInput.(*FindManyInput)
    }

    // 3. Call repository with modified input
    result, err := m.clientRepo.FindMany(m.ctx, input)
    // ...
}
```

---

### Decision 2: VM Pool Pattern (sync.Pool)

**Problem**: Creating a new Goja Runtime for every request is expensive (5-10ms overhead).

**Options Considered**:

1. **Create VM per request**
   - ❌ High overhead: ~5-10ms creation time
   - ❌ Memory churn from frequent allocation/GC
   - Simple to implement

2. **Single global VM with locking**
   - ❌ Severe bottleneck under concurrent load
   - ❌ Goja Runtime is NOT goroutine-safe
   - Simple to implement

3. **VM Pool (sync.Pool)** (CHOSEN)
   - ✅ Reuse Runtime instances: ~0.1ms acquisition
   - ✅ Goroutine-safe pooling
   - ✅ Automatic scaling with load
   - More complex to implement (reset state between uses)

**Rationale**: The `sync.Pool` pattern provides excellent performance characteristics by reusing Runtime instances while maintaining goroutine safety. The pool automatically scales with concurrent load and has minimal overhead.

**Implementation**:
```go
var vmPool = sync.Pool{
    New: func() interface{} {
        vm := goja.New()
        // Sandbox configuration
        vm.Set("eval", goja.Undefined())
        vm.Set("Function", goja.Undefined())
        return vm
    },
}

func executeInterceptor(script *goja.Program, ctx *InterceptorContext) (*InterceptorResult, error) {
    // Acquire VM from pool
    vm := vmPool.Get().(*goja.Runtime)
    defer vmPool.Put(vm) // Return to pool when done

    // Set context for this execution
    vm.Set("user", ctx.User)
    vm.Set("resource", ctx.Resource)
    vm.Set("environment", ctx.Environment)
    vm.Set("input", ctx.Input)

    // Execute pre-compiled script
    result, err := vm.RunProgram(script)
    if err != nil {
        return nil, err
    }

    return parseResult(vm, result)
}
```

**Performance Impact**:
- First request (cold): ~5ms (VM creation + script compilation)
- Subsequent requests (warm): ~0.5-2ms (VM pool + cached script)
- Pool hit rate: >90% under normal load

---

### Decision 3: Script Compilation Caching

**Problem**: Parsing and compiling JavaScript on every request adds ~3-5ms overhead.

**Options Considered**:

1. **Parse and compile on every execution**
   - ❌ ~3-5ms overhead per request
   - ❌ Wasteful for stable scripts
   - Simple to implement

2. **Cache compiled programs in memory** (CHOSEN)
   - ✅ Compile once, execute many times
   - ✅ ~0.5ms execution time (no parsing)
   - ✅ LRU cache with automatic eviction
   - Requires cache invalidation on script updates

3. **Pre-compile at startup only**
   - ❌ Cannot support dynamic script updates
   - ❌ Requires restart for rule changes
   - Defeats purpose of dynamic rules

**Rationale**: Caching compiled programs provides a 5-10x performance improvement for repeated executions while still supporting dynamic rule updates. The cache is invalidated when scripts are modified, ensuring correctness.

**Implementation**:
```go
type ScriptCache struct {
    mu    sync.RWMutex
    cache map[string]*CachedProgram
}

type CachedProgram struct {
    Program   *goja.Program
    Hash      string    // SHA256 of script content
    CompiledAt time.Time
}

func (c *ScriptCache) Get(ruleID string, script string) (*goja.Program, error) {
    c.mu.RLock()
    cached, exists := c.cache[ruleID]
    c.mu.RUnlock()

    // Check if script content changed (cache invalidation)
    currentHash := sha256Hash(script)
    if exists && cached.Hash == currentHash {
        return cached.Program, nil // Cache hit
    }

    // Cache miss or stale - compile new program
    program, err := goja.Compile(ruleID, script, true)
    if err != nil {
        return nil, err
    }

    // Update cache
    c.mu.Lock()
    c.cache[ruleID] = &CachedProgram{
        Program:    program,
        Hash:       currentHash,
        CompiledAt: time.Now(),
    }
    c.mu.Unlock()

    return program, nil
}
```

**Cache Hit Rate Target**: >95% for stable production environments

---

### Decision 4: Timeout Mechanism (Interrupt-based)

**Problem**: Malicious or buggy scripts can run indefinitely (infinite loops).

**Options Considered**:

1. **No timeout**
   - ❌ Security risk: script can DoS the system
   - ❌ Resource exhaustion possible
   - Simple to implement

2. **Context-based timeout** (context.WithTimeout)
   - ❌ Doesn't interrupt JavaScript execution
   - ❌ Goroutine leaks if script never returns
   - Misleading API (looks like it works)

3. **Goja Interrupt mechanism** (CHOSEN)
   - ✅ Actually stops JavaScript execution
   - ✅ Goroutine cleanup guaranteed
   - ✅ Configurable timeout per interceptor
   - Requires goroutine for async execution

**Rationale**: Goja's `Interrupt()` method is the only reliable way to stop long-running scripts. Context timeouts don't work because JavaScript execution is blocking.

**Implementation**:
```go
func executeWithTimeout(vm *goja.Runtime, program *goja.Program, timeoutMs int) (goja.Value, error) {
    resultChan := make(chan goja.Value, 1)
    errChan := make(chan error, 1)

    // Execute script in goroutine
    go func() {
        result, err := vm.RunProgram(program)
        if err != nil {
            errChan <- err
        } else {
            resultChan <- result
        }
    }()

    // Wait for completion or timeout
    timeout := time.Duration(timeoutMs) * time.Millisecond
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errChan:
        return nil, err
    case <-time.After(timeout):
        vm.Interrupt("execution timeout") // Stop the script
        return nil, ErrScriptTimeout
    }
}
```

**Default Timeout**: 100ms (configurable per interceptor)

---

### Decision 5: Sandboxing Strategy (Disable Dangerous Globals)

**Problem**: JavaScript can potentially access sensitive data or perform dangerous operations.

**Options Considered**:

1. **No sandboxing**
   - ❌ Security risk: eval(), Function() can execute arbitrary code
   - ❌ Could access Go runtime internals
   - Simple to implement

2. **Custom isolated VM implementation**
   - ❌ Massive engineering effort
   - ❌ High maintenance burden
   - ❌ Likely to have security holes

3. **Disable dangerous globals** (CHOSEN)
   - ✅ Prevents eval() and Function() injection
   - ✅ Limits script to provided context only
   - ✅ Simple to implement and maintain
   - ⚠️ Not 100% secure, but good enough for trusted scripts

**Rationale**: Since interceptor scripts are administrator-configured (not user-provided), we can trust the source but still want to prevent accidental security issues. Disabling `eval()` and `Function()` prevents the most common attack vectors.

**Implementation**:
```go
func createSandboxedVM() *goja.Runtime {
    vm := goja.New()

    // Disable code injection vectors
    vm.Set("eval", goja.Undefined())
    vm.Set("Function", goja.Undefined())

    // Provide safe helper functions
    vm.Set("contains", jsContains)
    vm.Set("hasPermission", jsHasPermission)

    return vm
}
```

**Security Posture**:
- ✅ Prevents: eval() injection, Function() constructor abuse
- ✅ Isolates: No access to Go runtime, filesystem, network
- ⚠️ Trusts: Administrator-authored scripts are assumed non-malicious
- ⚠️ Doesn't prevent: Intentional DoS (covered by timeout), complex logic bugs

---

### Decision 6: Interceptor Configuration Storage (Separate Table)

**Problem**: Where should interceptor rules be stored?

**Options Considered**:

1. **Embedded in Model metadata (JSON column)**
   - ❌ Hard to query and filter
   - ❌ No versioning or audit trail
   - ❌ Large JSON blobs in model table
   - Simple to implement

2. **File-based configuration (YAML/JSON files)**
   - ❌ Requires filesystem access (not cloud-native)
   - ❌ No dynamic updates without restart
   - ❌ Harder to version and audit
   - Simple to implement

3. **Separate database table** (CHOSEN)
   - ✅ Efficient querying (indexed by model_id + operation)
   - ✅ Supports versioning and audit trail
   - ✅ Independent lifecycle from models
   - ✅ Easier to manage and monitor
   - More complex (additional table, repository)

**Rationale**: A dedicated `interceptor_rules` table provides the best balance of performance, manageability, and auditability. It follows the single responsibility principle and allows interceptors to be managed independently of model definitions.

**Schema Design**:
```sql
CREATE TABLE interceptor_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_id BIGINT NOT NULL,
    operation VARCHAR(50) NOT NULL,  -- findFirst, findMany, createOne, etc.
    script TEXT NOT NULL,
    priority INT DEFAULT 0,          -- Execution order (higher = first)
    enabled BOOLEAN DEFAULT TRUE,
    timeout_ms INT DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    description TEXT,

    INDEX idx_model_operation (model_id, operation, enabled),
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
);
```

**Query Pattern**:
```sql
-- Efficient lookup for enabled interceptors, sorted by priority
SELECT * FROM interceptor_rules
WHERE model_id = ? AND operation = ? AND enabled = true
ORDER BY priority DESC, id ASC;
```

---

### Decision 7: Error Handling Strategy (Strict Mode)

**Problem**: How should the system handle interceptor failures?

**Options Considered**:

1. **Permissive mode (log error, continue)**
   - ❌ Security risk: bypasses authorization checks
   - ❌ Data integrity risk: bypasses validation
   - ❌ Silent failures hard to debug
   - Simpler user experience

2. **Strict mode (abort on error)** (CHOSEN)
   - ✅ Fail-safe: never bypass security checks
   - ✅ Clear error messages to client
   - ✅ Forces script authors to handle errors properly
   - Requires better error messages and debugging tools

3. **Configurable mode (per interceptor)**
   - ⚠️ Adds complexity
   - ⚠️ Risk of misconfiguration (accidentally permissive)
   - Could be added later if needed

**Rationale**: For security-critical interceptors (authorization, validation), strict mode is the only safe choice. It's better to fail loudly than to silently bypass security checks.

**Error Flow**:
```go
func (m *graphqlModelResolver) executeFindMany(p graphql.ResolveParams) (map[string]any, error) {
    input, err := newFindManyInput(m.model.Name, p)
    if err != nil {
        return nil, err
    }

    // Interceptor execution in strict mode
    if m.interceptorExecutor != nil {
        modifiedInput, err := m.interceptorExecutor.Execute(...)
        if err != nil {
            // ABORT OPERATION - do not call repository
            return nil, fmt.Errorf("interceptor error: %w", err)
        }
        input = modifiedInput.(*FindManyInput)
    }

    // Only reaches here if interceptor succeeded or no interceptor configured
    result, err := m.clientRepo.FindMany(m.ctx, input)
    // ...
}
```

**Error Response**:
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

---

## Cross-Cutting Concerns

### Logging and Observability

**Metrics to Collect**:
```
# Execution time histogram
interceptor_execution_duration_ms{model, operation, interceptor_id}

# Success/error counts
interceptor_execution_count{model, operation, status="success|error|timeout"}

# Cache performance
interceptor_cache_hit_rate{type="compiled_script"}
interceptor_vm_pool_utilization

# Resource usage
interceptor_vm_pool_size
interceptor_active_executions
```

**Log Levels**:
- **ERROR**: Script syntax errors, runtime errors, timeouts
- **WARN**: High latency (>10ms), low cache hit rate
- **INFO**: Successful executions (optional, configurable)
- **DEBUG**: Full script context and results

---

### Testing Strategy

**Unit Tests**:
- Goja engine wrapper (VM pool, timeout, sandboxing)
- Context builder (user, resource, environment)
- Script cache (invalidation, hit rate)
- Executor service (rule loading, execution order)

**Integration Tests**:
- End-to-end interceptor execution for each operation
- Multi-interceptor priority ordering
- Error handling and denial scenarios
- Performance benchmarks

**Test Scenarios**:
1. **Tenant Filtering**: Inject `tenantId` into WHERE clause
2. **Row-Level Security**: Deny access based on user.department
3. **Data Validation**: Reject invalid email formats
4. **Audit Logging**: Inject `createdBy`, `updatedBy` fields
5. **Default Values**: Add timestamps, UUIDs automatically

---

## Performance Considerations

### Expected Overhead
| Scenario | Overhead | Breakdown |
|----------|----------|-----------|
| No interceptor | 0ms | - |
| Simple interceptor (cached, warm pool) | 0.5-2ms | 0.1ms VM + 0.5ms exec + 0.4ms parse |
| Complex interceptor (50 lines) | 2-5ms | 0.1ms VM + 3ms exec + 0.5ms parse |
| First execution (cold) | 5-10ms | 5ms compile + 2ms exec |

### Optimization Opportunities
1. **Pre-compile at startup**: Load and compile all active interceptors on server start
2. **Batch context building**: Reuse context objects for multiple interceptors
3. **Lazy loading**: Only load interceptors for models that are actually queried
4. **Async logging**: Don't block operation on audit log writes

---

## Security Considerations

### Threat Model

**Threats Mitigated**:
- ✅ Code injection (eval, Function disabled)
- ✅ Infinite loops (timeout mechanism)
- ✅ Resource exhaustion (VM pool limits)
- ✅ Unauthorized access (authorization interceptors)

**Remaining Risks**:
- ⚠️ Malicious admin: Can write intentionally harmful scripts
- ⚠️ Logic bugs: Complex scripts may have authorization holes
- ⚠️ Side channels: Timing attacks possible (but low risk)

**Mitigations**:
- **Admin trust**: Interceptors are administrator-configured, not user-provided
- **Code review**: Production interceptors should undergo security review
- **Testing**: Comprehensive test suite for authorization logic
- **Monitoring**: Alert on high error rates, timeouts, unusual patterns

---

## Future Enhancements (Out of Scope)

1. **Response Transformation Interceptors**
   - Modify query results after database execution
   - Use cases: field masking, data enrichment, computed fields

2. **Custom JavaScript Libraries**
   - Allow importing trusted npm packages (lodash, moment.js)
   - Requires module loader implementation

3. **Visual Rule Builder**
   - Web UI for building interceptors without JavaScript
   - Drag-and-drop condition builder for non-programmers

4. **Interceptor Debugging Tools**
   - Step-through debugger for scripts
   - Log viewer with execution trace

5. **Interceptor Testing Framework**
   - Dry-run API endpoint
   - Unit test framework for scripts
   - Coverage reporting

---

## References

- **Goja Documentation**: https://github.com/dop251/goja
- **Research Document**: `docs/04-auth/script-engine-research.md`
- **Goja Quickstart**: `docs/04-auth/goja-quick-start.md`
- **Query API Spec**: `openspec/specs/query-api/spec.md`
