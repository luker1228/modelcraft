# Proposal: Add Goja Interceptor for Operations

## Change ID
`add-goja-interceptor-for-operations`

## Status
Draft

## Overview
Add JavaScript-based interceptor functionality using Goja to all GraphQL query and mutation operations (findFirst, findMany, findUnique, createOne, updateOne, deleteOne, etc.), enabling dynamic input modification for authorization, validation, and data transformation purposes.

## Motivation

### Business Need
ModelCraft requires flexible, dynamic authorization and validation capabilities that can be configured without recompiling or redeploying the application. Current limitations include:

1. **Static Authorization** - Authorization logic is hardcoded in Go, requiring deployment for rule changes
2. **Limited Validation** - Field-level validation is insufficient for complex cross-field business rules
3. **No Multi-tenancy Support** - Cannot easily inject tenant filters or row-level security
4. **AI-Generated Rules** - Need a scripting engine that AI can generate rules for easily

### Why Goja?
Based on research documented in `docs/04-auth/script-engine-research.md`:

- ✅ **JavaScript syntax** - AI-friendly, developer-friendly, widely understood
- ✅ **Pure Go implementation** - No cgo dependencies, cross-platform
- ✅ **High performance** - 6-7x faster than Otto, sufficient for ABAC decisions (<5ms)
- ✅ **Mature ecosystem** - 6,700+ GitHub stars, actively maintained
- ✅ **Dynamic state modification** - Can modify input arguments before execution

## Goals

1. **Enable Dynamic Authorization** - Allow JavaScript rules to modify or reject operations based on user context
2. **Support Input Transformation** - Modify query/mutation inputs (WHERE clauses, data payloads) before execution
3. **Provide Validation Capabilities** - Implement complex validation logic in JavaScript
4. **Maintain Performance** - Keep operation latency under acceptable thresholds (<10ms interceptor overhead)
5. **Ensure Security** - Sandbox JavaScript execution with timeouts and resource limits

## Non-Goals

1. **Replace existing Go-based validation** - Interceptors complement, not replace, core validation
2. **Support for response transformation** - Initial scope is input-only (output transformation is future work)
3. **Custom JavaScript libraries** - Only ES5.1 standard library available initially
4. **Database-level hooks** - Interceptors run at GraphQL resolver level, not database level

## User Impact

### Positive Impact
- **Developers** can implement complex authorization rules without Go code changes
- **AI systems** can generate and update rules dynamically
- **Multi-tenant applications** can inject tenant filters automatically
- **Compliance teams** can audit and modify rules without engineering involvement

### Migration Impact
- **Backward compatible** - Interceptors are optional, existing operations continue to work
- **No schema changes** - GraphQL schema remains unchanged
- **Configuration-based** - Interceptors are configured via model settings or external configuration

## Alternatives Considered

### Alternative 1: Casbin Rule Engine
**Rejected because:**
- DSL learning curve is steep
- Limited to authorization use cases
- Cannot modify input structures
- Less AI-friendly than JavaScript

### Alternative 2: Yaegi (Go Interpreter)
**Rejected because:**
- JavaScript is more AI-friendly than Go
- Go syntax has steeper learning curve for non-programmers
- Goja has better performance for simple condition evaluation

### Alternative 3: expr-lang
**Rejected because:**
- Expression-only, not full programming language
- Cannot handle complex logic (loops, functions)
- Limited state modification capabilities

## Open Questions

1. **Configuration Storage** - Where should interceptor scripts be stored?
   - Option A: In model metadata (design-time database)
   - Option B: Separate interceptor_rules table
   - Option C: File-based configuration
   - **Proposed**: Option B (separate table) for versioning and audit trail

2. **Execution Order** - If multiple interceptors are configured, what's the execution order?
   - **Proposed**: Sequential execution in priority order (configurable)

3. **Error Handling** - How should interceptor failures be handled?
   - Option A: Fail entire operation (strict mode)
   - Option B: Log error and continue (permissive mode)
   - **Proposed**: Configurable per interceptor with default=strict

4. **Context Injection** - What context should be available to interceptors?
   - User information (ID, roles, permissions)
   - Request metadata (IP, headers, timestamp)
   - Model metadata (table name, fields)
   - **Proposed**: All of the above via structured context object

5. **Testing Strategy** - How do users test interceptors before deployment?
   - **Proposed**: Provide a testing API endpoint that executes interceptor in dry-run mode

## Dependencies

- **Goja library** - `github.com/dop251/goja` (already researched, stable)
- **Interceptor configuration storage** - Requires new database table or model extension
- **Context management** - Extends existing `requestcontext` package

## Success Criteria

1. **Functional Requirements Met**
   - All query/mutation operations support interceptors
   - JavaScript can modify WHERE clauses, data payloads, and reject operations
   - Context (user, resource, environment) is available to scripts

2. **Performance Requirements Met**
   - Interceptor execution overhead < 10ms for typical rules
   - VM pool reuse keeps memory usage stable
   - No impact on operations without interceptors configured

3. **Security Requirements Met**
   - Script execution timeouts prevent infinite loops
   - Sandboxing prevents access to Go runtime internals
   - Input validation prevents script injection attacks

4. **Documentation Complete**
   - API documentation with examples
   - Migration guide for existing applications
   - Best practices for writing interceptor rules

## Timeline Considerations

This proposal does NOT include implementation timelines. Task breakdown and sequencing are documented in `tasks.md`.

## References

- Research: `docs/04-auth/script-engine-research.md`
- Goja quickstart: `docs/04-auth/goja-quick-start.md`
- Query API spec: `openspec/specs/query-api/spec.md`
- Existing GraphQL resolver: `internal/domain/modelruntime/model_resolver.go`
