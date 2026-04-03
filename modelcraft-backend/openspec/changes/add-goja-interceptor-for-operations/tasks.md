# Implementation Tasks

## Overview
This document outlines the implementation tasks for adding Goja-based interceptors to all GraphQL operations. Tasks are ordered to enable incremental delivery and parallel work where possible.

## Phase 1: Foundation (Core Infrastructure)

### Task 1.1: Interceptor Domain Model
**Description**: Define domain entities for interceptor configuration and execution context

**Deliverables**:
- `InterceptorRule` entity with fields: ID, ModelID, Operation, Script, Priority, Enabled, CreatedAt, UpdatedAt
- `InterceptorContext` value object containing user, resource, action, environment attributes
- `InterceptorResult` value object with modified input and decision (allow/deny)
- Repository interface: `InterceptorRuleRepository`

**Acceptance Criteria**:
- Domain model follows DDD patterns
- Unit tests for value object validation
- Repository interface defined

**Dependencies**: None

---

### Task 1.2: Goja Engine Wrapper
**Description**: Create reusable Goja runtime wrapper with VM pooling and safety controls

**Deliverables**:
- `pkg/scriptengine/goja_engine.go` with VM pool (sync.Pool)
- Timeout mechanism (100ms default, configurable)
- Sandboxing (disable eval, Function constructor)
- Error handling and stack trace parsing
- Helper functions injection (contains, hasPermission, etc.)

**Acceptance Criteria**:
- VM pool reuses instances correctly
- Timeout interrupts long-running scripts
- Memory usage is stable under load
- Unit tests cover timeout, errors, and context injection

**Dependencies**: None (can work in parallel with 1.1)

---

### Task 1.3: Database Schema for Interceptor Rules
**Description**: Create database table to store interceptor configurations

**Deliverables**:
- Atlas schema definition: `db/schema/interceptor_rules.sql`
- Migration file: `db/migrations/YYYYMMDD_add_interceptor_rules.sql`
- sqlc model: `internal/infrastructure/repository/interceptor_rule_model.go`

**Schema**:
```sql
CREATE TABLE interceptor_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_id BIGINT NOT NULL,
    operation VARCHAR(50) NOT NULL,  -- findFirst, findMany, createOne, etc.
    script TEXT NOT NULL,
    priority INT DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    timeout_ms INT DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_model_operation (model_id, operation, enabled)
);
```

**Acceptance Criteria**:
- Schema passes Atlas validation
- Migration can be applied and rolled back
- Indexes support efficient lookup by model_id + operation

**Dependencies**: Task 1.1 (domain model defined)

---

### Task 1.4: Interceptor Repository Implementation
**Description**: Implement sqlc-based repository for interceptor rules

**Deliverables**:
- `internal/infrastructure/repository/sql_interceptor_repo.go`
- Methods: `GetByModelAndOperation(modelID, operation)`, `GetEnabled()`, `Create()`, `Update()`, `Delete()`
- Result caching (optional, for performance)

**Acceptance Criteria**:
- Repository methods follow existing patterns
- Integration tests verify CRUD operations
- Efficient query performance (indexed lookups)

**Dependencies**: Task 1.3 (database schema exists)

---

## Phase 2: Interceptor Execution Engine

### Task 2.1: Interceptor Executor Service
**Description**: Create application service to load, compile, and execute interceptor scripts

**Deliverables**:
- `internal/app/interceptor/executor_service.go`
- Methods: `Execute(ctx, modelID, operation, input, context) -> (modifiedInput, error)`
- Script compilation caching (LRU cache of compiled programs)
- Metrics collection (execution time, error rate)

**Execution Flow**:
1. Load enabled interceptors for model + operation
2. Sort by priority
3. For each interceptor:
   - Get VM from pool
   - Set context (user, resource, action, input)
   - Run compiled script
   - Parse result (modified input + allow/deny)
   - Return error if denied
4. Return final modified input

**Acceptance Criteria**:
- Multiple interceptors execute in priority order
- Failed interceptor stops execution (strict mode)
- Compiled scripts are cached and reused
- Unit tests mock repository and Goja engine

**Dependencies**: Task 1.1, 1.2, 1.4

---

### Task 2.2: Context Builder
**Description**: Build structured context object from request metadata

**Deliverables**:
- `internal/app/interceptor/context_builder.go`
- Extracts user info from `requestcontext.GetMetadata()`
- Builds resource attributes from model metadata
- Adds environment context (timestamp, IP, etc.)

**Context Structure**:
```go
type InterceptorContext struct {
    User struct {
        ID          string
        Roles       []string
        Permissions []string
        Metadata    map[string]interface{}
    }
    Resource struct {
        ModelName   string
        Operation   string
        Fields      []string
    }
    Environment struct {
        Timestamp   time.Time
        RequestID   string
        IPAddress   string
    }
    Input map[string]interface{}  // Original input (WHERE, data, etc.)
}
```

**Acceptance Criteria**:
- Context contains all necessary information
- Handles missing metadata gracefully
- Unit tests verify context structure

**Dependencies**: Task 2.1

---

## Phase 3: GraphQL Integration

### Task 3.1: Integrate Interceptor into Query Operations
**Description**: Add interceptor execution to findFirst, findMany, findUnique, aggregate, count

**Deliverables**:
- Modify `executeFindFirst`, `executeFindMany`, `executeFindUnique`, `executeAggregate`, `executeCount` in `model_resolver.go`
- Call interceptor executor before repository call
- Pass modified WHERE clause to repository

**Example Integration**:
```go
func (m *graphqlModelResolver) executeFindMany(p graphql.ResolveParams) (map[string]any, error) {
    input, err := newFindManyInput(m.model.Name, p)
    if err != nil {
        return nil, err
    }

    // NEW: Interceptor execution
    if m.interceptorExecutor != nil {
        modifiedInput, err := m.interceptorExecutor.Execute(
            m.ctx, m.model.ID, "findMany", input, buildContext(p),
        )
        if err != nil {
            return nil, err // Operation denied or interceptor error
        }
        input = modifiedInput.(*FindManyInput)
    }

    result, err := m.clientRepo.FindMany(m.ctx, input)
    // ... rest of implementation
}
```

**Acceptance Criteria**:
- All query operations call interceptor executor
- Original behavior preserved when no interceptor configured
- Errors from interceptor are properly propagated
- Integration tests verify interceptor modifies WHERE clause

**Dependencies**: Task 2.1, 2.2

---

### Task 3.2: Integrate Interceptor into Mutation Operations
**Description**: Add interceptor execution to createOne, createMany, updateOne, updateMany, deleteOne, deleteMany

**Deliverables**:
- Modify mutation executors in `model_resolver.go`
- Call interceptor before repository call
- Pass modified data payload or WHERE clause to repository

**Example Integration**:
```go
func (m *graphqlModelResolver) executeCreateOne(p graphql.ResolveParams) (interface{}, error) {
    input, err := newCreateOneInput(m.model.Name, p)
    if err != nil {
        return nil, err
    }

    // NEW: Interceptor execution
    if m.interceptorExecutor != nil {
        modifiedInput, err := m.interceptorExecutor.Execute(
            m.ctx, m.model.ID, "createOne", input, buildContext(p),
        )
        if err != nil {
            return nil, err
        }
        input = modifiedInput.(*CreateOneInput)
    }

    // ... UUID generation and rest of implementation
}
```

**Acceptance Criteria**:
- All mutation operations call interceptor executor
- Interceptor can modify data payload (createOne, updateOne)
- Interceptor can modify WHERE clause (updateMany, deleteMany)
- Interceptor can deny operations (return error)
- Integration tests verify data modification

**Dependencies**: Task 2.1, 2.2

---

## Phase 4: Testing and Validation

### Task 4.1: Integration Test Suite
**Description**: Comprehensive integration tests for interceptor functionality

**Deliverables**:
- Test interceptor modifying WHERE clause (tenant filtering)
- Test interceptor modifying data payload (default values, sanitization)
- Test interceptor denying operations (authorization failure)
- Test multiple interceptors in priority order
- Test timeout handling
- Test error scenarios (syntax errors, runtime errors)

**Test Scenarios**:
1. **Tenant Filtering**: Interceptor injects `tenantId: currentUser.tenantId` into WHERE
2. **Row-Level Security**: Interceptor denies access if user.department != resource.department
3. **Data Sanitization**: Interceptor removes sensitive fields from input
4. **Validation**: Interceptor rejects invalid email format
5. **Default Values**: Interceptor adds `createdBy: currentUser.id` to createOne data

**Acceptance Criteria**:
- All scenarios pass with expected behavior
- Tests run in CI pipeline
- Code coverage > 80% for interceptor components

**Dependencies**: Task 3.1, 3.2

---

### Task 4.2: Performance Benchmarking
**Description**: Measure interceptor execution overhead and optimize if needed

**Deliverables**:
- Benchmark tests: `internal/app/interceptor/executor_bench_test.go`
- Measure VM pool overhead, script compilation time, execution time
- Identify bottlenecks and optimize

**Performance Targets**:
- Simple interceptor (< 10 lines): < 1ms overhead
- Complex interceptor (50 lines): < 5ms overhead
- VM pool reuse: > 90% hit rate
- Compiled script cache: > 95% hit rate

**Acceptance Criteria**:
- Benchmarks run in CI
- Performance meets targets
- Optimization recommendations documented

**Dependencies**: Task 3.1, 3.2

---

## Phase 5: API and Documentation

### Task 5.1: Interceptor Management API
**Description**: REST API endpoints for CRUD operations on interceptor rules

**Deliverables**:
- `POST /api/v1/models/{modelId}/interceptors` - Create interceptor
- `GET /api/v1/models/{modelId}/interceptors` - List interceptors
- `PUT /api/v1/models/{modelId}/interceptors/{id}` - Update interceptor
- `DELETE /api/v1/models/{modelId}/interceptors/{id}` - Delete interceptor
- `POST /api/v1/models/{modelId}/interceptors/{id}/test` - Test interceptor (dry-run)

**Request/Response Examples**:
```json
POST /api/v1/models/123/interceptors
{
  "operation": "findMany",
  "script": "function intercept() { input.where.tenantId = user.tenantId; return input; }",
  "priority": 10,
  "enabled": true,
  "timeoutMs": 100
}
```

**Acceptance Criteria**:
- API endpoints follow RESTful conventions
- Input validation prevents invalid scripts
- Test endpoint returns interceptor result without executing operation
- API tests verify CRUD operations

**Dependencies**: Task 2.1, 1.4

---

### Task 5.2: Documentation
**Description**: User-facing documentation for interceptor feature

**Deliverables**:
- `docs/05-interceptor/README.md` - Overview and concepts
- `docs/05-interceptor/api.md` - API reference
- `docs/05-interceptor/examples.md` - Common use cases with code examples
- `docs/05-interceptor/best-practices.md` - Performance, security, testing guidelines

**Documentation Sections**:
1. **Introduction** - What are interceptors, why use them
2. **Configuration** - How to create and manage interceptor rules
3. **Context API** - Available variables (user, resource, input, environment)
4. **Examples**:
   - Tenant filtering (multi-tenancy)
   - Row-level security (ABAC)
   - Data validation
   - Audit logging
   - Default value injection
5. **Best Practices**:
   - Keep scripts simple and fast
   - Use compiled functions, not inline code
   - Test interceptors before enabling
   - Monitor performance metrics
6. **Troubleshooting** - Common errors and solutions

**Acceptance Criteria**:
- Documentation is clear and comprehensive
- Examples are executable and tested
- Best practices are actionable

**Dependencies**: Task 5.1, 4.1

---

## Optional Future Enhancements (Not in Scope)

### Response Transformation Interceptors
- Modify query results after database execution
- Use cases: field masking, data enrichment, computed fields

### Custom JavaScript Libraries
- Allow importing trusted npm packages
- Provide utility libraries (lodash, moment.js)

### Visual Rule Builder
- Web UI for building interceptor rules without JavaScript
- Drag-and-drop condition builder

### Interceptor Debugging Tools
- Step-through debugger for interceptor scripts
- Log viewer with script execution trace

---

## Task Dependencies Diagram

```
Phase 1: Foundation
├── 1.1 Domain Model ─────────┐
├── 1.2 Goja Engine ──────────┤
├── 1.3 DB Schema ────────────┤
└── 1.4 Repository ───────────┘
            │
            ▼
Phase 2: Execution Engine
├── 2.1 Executor Service ─────┐
└── 2.2 Context Builder ──────┘
            │
            ▼
Phase 3: GraphQL Integration
├── 3.1 Query Operations ─────┐
└── 3.2 Mutation Operations ──┘
            │
            ▼
Phase 4: Testing
├── 4.1 Integration Tests ────┐
└── 4.2 Performance Bench ────┘
            │
            ▼
Phase 5: API & Docs
├── 5.1 Management API ───────┐
└── 5.2 Documentation ─────────┘
```

---

## Parallelization Opportunities

**Can work in parallel**:
- Task 1.1 + 1.2 (domain model and Goja engine are independent)
- Task 3.1 + 3.2 (query and mutation integration after executor service is ready)
- Task 4.1 + 4.2 (tests and benchmarks after integration complete)
- Task 5.1 + 5.2 (API and docs after testing complete)

**Must be sequential**:
- Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5

---

## Validation Checklist

Before marking this change as complete, verify:

- [ ] All query operations (findFirst, findMany, findUnique, aggregate, count) support interceptors
- [ ] All mutation operations (createOne, createMany, updateOne, updateMany, deleteOne, deleteMany) support interceptors
- [ ] Interceptors can modify WHERE clauses
- [ ] Interceptors can modify data payloads
- [ ] Interceptors can deny operations (return error)
- [ ] Multiple interceptors execute in priority order
- [ ] Context includes user, resource, environment, and input
- [ ] VM pool reuses instances efficiently
- [ ] Script compilation is cached
- [ ] Timeout mechanism prevents infinite loops
- [ ] Sandboxing prevents access to Go internals
- [ ] Performance overhead < 10ms for typical interceptors
- [ ] Integration tests cover all scenarios
- [ ] API endpoints for CRUD operations work correctly
- [ ] Documentation is complete and accurate
- [ ] `openspec validate add-goja-interceptor-for-operations --strict` passes
