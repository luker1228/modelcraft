## 1. Implementation

### 1.1 Add Count Operation Constants
- [x] Add `OperationCount = "count"` constant to `internal/domain/modelruntime/graphql_constants.go`

### 1.2 Implement Count Input Type Generation
- [x] Add `GenerateCountArgs` method to `inputTypeGenerator` in `internal/domain/modelruntime/graphql_input_types.go`
- [x] Generate `where` parameter accepting `WhereInput` type (reuse existing)
- [x] Generate `select` parameter accepting count selection input with `_all` and all model fields

### 1.3 Implement Count Result Type
- [x] Add `createCountResultType` method to `graphqlModelResolver` in `internal/domain/modelruntime/model_resolver.go`
- [x] When `select` is NOT used: return type with single `count` field (Int!)
- [x] When `select` is used: return type with `fieldsCount` field containing nested object with `_all` (Int!) and selected fields (Int!)

### 1.4 Add Count GraphQL Field
- [x] Add `createCountField` method to `graphqlModelResolver`
- [x] Register count field in root Query type
- [x] Wire up count args and result type
- [x] Implement resolve function calling `executeCount`

### 1.5 Implement Count Execution Logic
- [x] Add `executeCount` method to `graphqlModelResolver`
- [x] Parse `where` parameter using existing `ParseWhereInput` function
- [x] Parse optional `select` parameter to determine which counts to return
- [x] Call query executor with parsed parameters

### 1.6 Implement Count SQL Generation
- [x] Add `convertCountInputToSQL` in `internal/infrastructure/database/dml/sql_mapper.go`
- [x] Generate `SELECT COUNT(*) as count` for simple count (no select)
- [x] Generate `SELECT COUNT(*) as _count__all, COUNT(field1) as _count_field1, ...` for field-level counts
- [x] Apply WHERE clause using existing `convertWhereToExpression` function
- [x] Execute query and map results to expected structure

### 1.7 Result Mapping
- [x] Map single count result to `{count: N}` structure
- [x] Map field-level counts to `{fieldsCount: {_all: N, field1: N, ...}}` structure
- [x] Handle empty result set (return 0 for counts)

## 2. Testing

### 2.1 Unit Tests
- [ ] Test count operation constant registration
- [ ] Test count input type generation (with/without select)
- [ ] Test count result type generation
- [ ] Test SQL query generation for simple count
- [ ] Test SQL query generation with WHERE filter
- [ ] Test SQL query generation with select (field-level counts)
- [ ] Test result mapping for both formats

### 2.2 Integration Tests
- [ ] Test simple count query via GraphQL
- [ ] Test count with WHERE filter
- [ ] Test count with select (field-level)
- [ ] Test count with WHERE + select combined
- [ ] Test count on empty table (returns 0)
- [ ] Test count with complex WHERE conditions (AND, OR, NOT)
- [ ] Test field counts with NULL values (COUNT(field) excludes nulls)

### 2.3 Performance Tests
- [ ] Verify COUNT query performance on large dataset
- [ ] Compare with findMany + array length approach
- [ ] Verify WHERE filter is pushed to database (EXPLAIN query)

## 3. Documentation

### 3.1 Spec Documentation
- [ ] Update `openspec/changes/add-count-operation/specs/query-api/spec.md` with requirements
- [ ] Document count operation semantics
- [ ] Document return structure for simple vs field-level counts
- [ ] Provide GraphQL query examples

### 3.2 Code Documentation
- [ ] Add godoc comments to all new functions
- [ ] Document count input structure
- [ ] Document count result structure
- [ ] Add examples in function comments

### 3.3 User Documentation
- [ ] Add count operation examples to test queries or documentation
- [ ] Document difference between count and aggregate._count
- [ ] Provide migration examples from aggregate._count to count (if applicable)

## 4. Validation

- [ ] Run `openspec validate add-count-operation --strict` and fix all issues
- [ ] Run Go tests: `make test`
- [ ] Run integration tests: `pytest automated/`
- [ ] Manual testing via GraphQL Playground
- [ ] Code review and approval

## Dependencies

- No external dependencies
- Reuses existing WHERE input types from query-api
- Builds on existing SQL query builder infrastructure

## Validation Criteria

- All tests pass
- OpenSpec validation passes with --strict
- Count queries execute efficiently (single database query)
- GraphQL schema generates correctly with count field
- Documentation is complete and accurate
