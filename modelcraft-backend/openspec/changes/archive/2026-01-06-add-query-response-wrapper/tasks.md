# Tasks: Add Query Response Wrapper with Metadata

## Prerequisites
- [ ] Proposal approved
- [ ] Design reviewed
- [ ] Breaking change announcement prepared

## Phase 1: Foundation (Request Metadata)

### Task 1.1: Create request metadata package
- [ ] Create `pkg/requestcontext/metadata.go`
- [ ] Define `RequestMetadata` struct with `ReqID` and `StartTime` fields
- [ ] Implement `WithMetadata(ctx)` to inject metadata into context
- [ ] Implement `GetMetadata(ctx)` to retrieve metadata from context
- [ ] Implement `CalculateTimeCost(ctx)` helper function
- [ ] Add unit tests for metadata package (100% coverage)

### Task 1.2: Add request ID generation
- [ ] Verify `pkg/bizutils/GenerateUUIDV7()` exists and works
- [ ] Create `GenerateRequestID()` wrapper function in bizutils (if needed)
- [ ] Add unit tests for request ID uniqueness (generate 10k IDs, verify no collisions)

### Task 1.3: Add GraphQL middleware for metadata injection
- [ ] Identify GraphQL handler entry point (likely in `internal/app/...`)
- [ ] Add middleware to inject RequestMetadata before query execution
- [ ] Ensure context propagates to all resolvers
- [ ] Add integration test to verify metadata availability in resolvers

## Phase 2: GraphQL Type Generation

### Task 2.1: Create QueryResultMeta interface
- [ ] Add `QueryResultMeta` interface definition in `graphql_constants.go`
- [ ] Define common fields: `timeCost: Int!`, `reqId: String!`
- [ ] Document interface purpose in comments

### Task 2.2: Create result type generators
- [ ] Add `createFindUniqueResultType(modelType)` method to `graphqlModelResolver`
  - Returns `{Model}FindUniqueResult` with `item: {Model}`, `timeCost`, `reqId`
- [ ] Add `createFindFirstResultType(modelType)` method
  - Returns `{Model}FindFirstResult` with `item: {Model}`, `timeCost`, `reqId`
- [ ] Add `createFindManyResultType(modelType)` method
  - Returns `{Model}FindManyResult` with `items: [{Model}!]!`, `totalCount`, `timeCost`, `reqId`
- [ ] Ensure all result types implement `QueryResultMeta` interface
- [ ] Add field descriptions to all generated fields

### Task 2.3: Update constant definitions
- [ ] Add `FieldItem = "item"` constant in `graphql_constants.go`
- [ ] Add `FieldItems = "items"` constant
- [ ] Add `FieldTimeCost = "timeCost"` constant
- [ ] Add `FieldReqId = "reqId"` constant
- [ ] Add `FieldTotalCount = "totalCount"` constant (for future use)

## Phase 3: Resolver Modifications

### Task 3.1: Modify findUnique resolver
- [ ] Update `executeFindUnique()` to return `map[string]any` instead of direct object
- [ ] Record operation start time at beginning of execution
- [ ] Wrap result with `item`, `timeCost`, `reqId` fields
- [ ] Handle null item case (when record not found)
- [ ] Update `createFindUniqueField()` to use new result type
- [ ] Update return type from direct model to wrapper type
- [ ] Ensure resolver uses context metadata

### Task 3.2: Modify findFirst resolver
- [ ] Update `executeFindFirst()` to return `map[string]any` instead of direct object
- [ ] Record operation start time at beginning of execution
- [ ] Wrap result with `item`, `timeCost`, `reqId` fields
- [ ] Handle null item case (when no records match)
- [ ] Update `createFindFirstField()` to use new result type
- [ ] Update return type from direct model to wrapper type
- [ ] Ensure resolver uses context metadata

### Task 3.3: Modify findMany resolver
- [ ] Update `executeFindMany()` to return `map[string]any` instead of `[]map[string]any`
- [ ] Record operation start time at beginning of execution
- [ ] Wrap result with `items`, `timeCost`, `reqId` fields
- [ ] Handle empty items array case (when no records match)
- [ ] Update `createFindManyField()` to use new result type
- [ ] Update return type from direct array to wrapper type
- [ ] Add `totalCount` field as nullable Int (set to nil for now)
- [ ] Ensure resolver uses context metadata

### Task 3.4: Update error handling
- [ ] Modify error responses to include `reqId` in extensions
- [ ] Add `timeCost` to error extensions
- [ ] Update `handleErr()` function to extract metadata from context
- [ ] Test error response format with integration tests

## Phase 4: Testing

### Task 4.1: Unit tests for type generation
- [ ] Test `createFindUniqueResultType()` generates correct schema
- [ ] Test `createFindFirstResultType()` generates correct schema
- [ ] Test `createFindManyResultType()` generates correct schema
- [ ] Verify all fields have correct types (nullable vs non-null)
- [ ] Verify field descriptions are present

### Task 4.2: Unit tests for resolver wrapping
- [ ] Test `executeFindUnique()` wraps result correctly
- [ ] Test null item handling in findUnique
- [ ] Test `executeFindFirst()` wraps result correctly
- [ ] Test null item handling in findFirst
- [ ] Test `executeFindMany()` wraps result correctly
- [ ] Test empty items array handling
- [ ] Verify timeCost is calculated correctly (within reasonable range)
- [ ] Verify reqId is included and is valid UUID v7

### Task 4.3: Integration tests
- [ ] Add test for findUnique query with metadata fields
  - Query: `findUnique { item { id }, timeCost, reqId }`
  - Verify response structure
  - Verify metadata values are reasonable
- [ ] Add test for findFirst query with metadata fields
  - Query: `findFirst { item { id }, timeCost, reqId }`
  - Verify response structure
- [ ] Add test for findMany query with metadata fields
  - Query: `findMany { items { id }, timeCost, reqId }`
  - Verify response structure
- [ ] Add test for null item scenarios (not found)
- [ ] Add test for empty items array (no matches)
- [ ] Add test for error responses include metadata

### Task 4.4: Performance tests
- [ ] Benchmark query execution before and after wrapper
- [ ] Verify overhead < 1ms per query
- [ ] Memory profiling to verify minimal overhead
- [ ] Document performance impact in test results

### Task 4.5: Update existing tests
- [ ] Update `tests/runtime/test_user_graphql.py` to use new response structure
  - Update all findUnique queries to select `item { ... }`
  - Update all findFirst queries to select `item { ... }`
  - Update all findMany queries to select `items { ... }`
- [ ] Update assertions to check wrapper fields
- [ ] Add assertions for timeCost and reqId presence
- [ ] Ensure all tests pass

## Phase 5: Documentation

### Task 5.1: Update API documentation
- [ ] Update `docs/03-runtime/api-guide.md`
  - Update all query examples to show new response structure
  - Add section explaining metadata fields
  - Document timeCost meaning (milliseconds)
  - Document reqId format (UUID v7)
- [ ] Update response examples with wrapper structure
- [ ] Add section on "Query Metadata" explaining timeCost and reqId

### Task 5.2: Update GraphQL schema documentation
- [ ] Document QueryResultMeta interface
- [ ] Document all result types
- [ ] Add inline comments explaining field semantics
- [ ] Update introspection query examples

### Task 5.3: Add changelog entry
- [ ] Create changelog entry for new version
- [ ] List breaking changes
- [ ] List new features (metadata fields)

## Phase 6: Deployment Preparation

### Task 6.1: Version bump
- [ ] Update version in relevant files
- [ ] Update API version constant (if exists)
- [ ] Tag release

## Phase 7: Validation

### Task 7.1: OpenSpec validation
- [ ] Run `openspec validate add-query-response-wrapper --strict`
- [ ] Resolve any validation errors
- [ ] Ensure all spec deltas are correctly formatted

### Task 7.2: Manual testing
- [ ] Test queries in GraphiQL playground
- [ ] Verify metadata fields populate correctly
- [ ] Test null item cases
- [ ] Test empty items cases
- [ ] Test error scenarios

### Task 7.3: Pre-deployment checklist
- [ ] All tests passing (unit, integration, performance)
- [ ] Documentation complete and reviewed
- [ ] Breaking change documented in changelog
- [ ] Rollback plan prepared

## Phase 8: Post-Deployment

### Task 8.1: Monitoring
- [ ] Set up alerts for query performance (timeCost metrics)
- [ ] Monitor reqId usage in logs
- [ ] Track error rates by operation

### Task 8.2: Archive change
- [ ] Run `openspec archive add-query-response-wrapper --yes`
- [ ] Update `specs/query-api/spec.md` with finalized requirements
- [ ] Verify archived change passes validation

## Task Summary

**Total tasks**: 48 (8 phases)
**Estimated time**: ~7-10 days

### Critical Path:
1. Phase 1 (Foundation) → Phase 2 (Types) → Phase 3 (Resolvers) → Phase 4 (Testing) → Phase 5 (Documentation) → Phase 6 (Deployment)

### Parallel work opportunities:
- Documentation (Phase 5) can start after Phase 3 is complete
- Performance tests (4.4) can run in parallel with integration tests (4.3)

## Validation Checklist

Before marking this change as complete:
- [ ] All tasks checked off
- [ ] All tests passing (100% pass rate)
- [ ] Documentation reviewed and approved
- [ ] Performance overhead verified < 1ms
- [ ] Breaking change properly documented
- [ ] OpenSpec validation passes with `--strict`
