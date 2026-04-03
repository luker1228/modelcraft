# Design: Aggregate Query Implementation

## Context

ModelCraft provides a GraphQL runtime for dynamically generated models, currently supporting basic CRUD operations and query operations (findUnique, findFirst, findMany). Users need to perform statistical analysis on their data without writing custom SQL queries.

Prisma ORM provides a well-designed aggregate API that is intuitive and type-safe. This design follows Prisma's aggregate pattern while adapting to ModelCraft's dynamic schema generation approach.

**Constraints:**
- Must work with dynamically generated GraphQL schemas (no static schema files)
- Must support MySQL aggregate functions through goqu query builder
- Must validate field types at schema generation time (e.g., AVG only on numeric fields)
- Must maintain consistency with existing query operation patterns

**Stakeholders:**
- End users who need data analytics capabilities
- Developers building dashboards and reports on ModelCraft data
- DBA users who want to avoid writing raw SQL queries

## Goals / Non-Goals

**Goals:**
- Provide standard SQL aggregation operations (COUNT, AVG, SUM, MIN, MAX)
- Support conditional aggregation with WHERE filters
- Handle null values correctly (count non-null, exclude from avg/sum)
- Generate type-safe GraphQL schema based on field types
- Maintain consistent error handling and validation patterns

**Non-Goals:**
- GROUP BY aggregation (deferred to future enhancement)
- HAVING clause support (deferred)
- Custom aggregate functions or expressions
- Aggregation across relations (complex joins)
- Real-time streaming aggregates

## Decisions

### 1. API Design: Prisma-Compatible Structure

**Decision:** Use Prisma's aggregate input/output structure with `_count`, `_avg`, `_sum`, `_min`, `_max` prefixes.

**Rationale:**
- Prisma's API is battle-tested and widely understood
- Clear namespacing avoids conflicts with model field names
- Nested structure `{ _count: { _all: true, field: true } }` is explicit and type-safe
- Reduces learning curve for developers familiar with Prisma

**Example:**
```graphql
query {
  aggregate {
    _count { _all }
    _avg { amount }
    _sum { amount }
    _min { amount }
    _max { amount }
  }
}
```

**Alternatives considered:**
- Flat structure `{ countAll, avgAmount }` - rejected due to naming collision risk
- SQL-like `{ count: "*", avg: ["amount"] }` - rejected as less intuitive

### 2. Field Type Validation

**Decision:** Validate aggregate operations against field types at schema generation time, not runtime.

**Rationale:**
- GraphQL schema should enforce type safety
- Early failure provides better developer experience
- Prevents invalid queries from being constructed
- Consistent with GraphQL best practices

**Implementation:**
- Only generate `_avg`, `_sum`, `_min`, `_max` input fields for numeric types (int, float, decimal)
- COUNT works on all field types
- `_all` special field for COUNT(*) is always available

**Alternatives considered:**
- Runtime validation - rejected as too late in the request cycle
- Allow all types and coerce - rejected as potentially misleading

### 3. NULL Handling Strategy

**Decision:** Follow SQL standard behavior for NULL values.

**Rationale:**
- COUNT(field) counts non-null values only
- COUNT(*) counts all rows including nulls
- AVG/SUM/MIN/MAX ignore null values
- This matches database behavior and user expectations

**Behavior:**
- `_count: { _all: true }` → COUNT(*) - counts all rows
- `_count: { field: true }` → COUNT(field) - counts non-null values only
- `_avg: { field: true }` → AVG(field) - excludes nulls, returns null if no non-null values
- `_sum: { field: true }` → SUM(field) - excludes nulls, returns null if no non-null values
- Empty result set: `{ _count: { _all: 0 }, _avg: { field: null } }`
- All nulls result: `{ _count: { _all: N, field: 0 }, _avg: { field: null } }`

**Critical distinction:**
- `_avg: { age: null }` can mean EITHER no rows exist OR all age values are null
- Combined with `_count: { _all }`, you can differentiate these cases:
  - `_count._all: 0` + `_avg.age: null` → no rows
  - `_count._all: 9` + `_count.age: 0` + `_avg.age: null` → 9 rows, all nulls
  - `_count._all: 9` + `_count.age: 5` + `_avg.age: 25` → 9 rows, 5 non-null
- This allows differentiating true aggregate values (including zero) from absence of data

### 4. SQL Generation Strategy

**Decision:** Use goqu query builder with aggregate functions, reusing existing WHERE clause handling.

**Rationale:**
- goqu already used throughout the codebase
- Provides SQL injection protection via prepared statements
- `convertWhereToExpression()` handles complex conditions
- Maintains consistency with findMany, updateMany patterns

**Example SQL generation:**
```go
// Input: { _count: { _all: true }, _avg: { amount: true } }
// Output: SELECT COUNT(*) as _count__all, AVG(amount) as _avg_amount FROM orders WHERE ...
```

**Alternatives considered:**
- Raw SQL building - rejected due to injection risk
- sqlc aggregation - rejected as we use goqu for other queries
- Separate query per aggregate - rejected as inefficient

### 5. Result Structure Mapping

**Decision:** Map flat SQL result columns to nested GraphQL structure using naming convention.

**Rationale:**
- SQL returns flat result: `{ _count__all: 10, _avg_amount: 50.5 }`
- GraphQL expects nested: `{ _count: { _all: 10 }, _avg: { amount: 50.5 } }`
- Use `__` for nested _all, `_` for field separator
- Parser can reconstruct nested structure deterministically

**Implementation:**
```go
// SQL column names:
// _count__all → _count._all
// _count_field → _count.field
// _avg_field → _avg.field
// _sum_field → _sum.field
```

## Risks / Trade-offs

### Risk: Performance on Large Tables

**Risk:** Aggregate queries on large tables without indexes may be slow.

**Mitigation:**
- Document index requirements for aggregate performance
- Consider adding LIMIT validation (reject aggregates without WHERE on huge tables)
- Future: Add query cost estimation and warnings

**Trade-off:** Simplicity vs safety - we prioritize usability but document best practices.

### Risk: Memory Usage with Many Aggregate Fields

**Risk:** Selecting many aggregates (e.g., AVG/SUM/MIN/MAX on 50 fields) could be slow.

**Mitigation:**
- Single SQL query is still more efficient than multiple queries
- Database handles aggregation efficiently
- No artificial limits on number of aggregates

**Trade-off:** Flexibility vs performance - we trust users to select what they need.

### Trade-off: No GROUP BY Support

**Decision:** Deferred GROUP BY to future enhancement.

**Rationale:**
- GROUP BY adds significant complexity (multiple result rows)
- Requires new result type (array of aggregate results)
- 80% of use cases are simple aggregates without grouping
- Can be added later without breaking changes

**Future path:**
- Add optional `groupBy: [field1, field2]` parameter
- Return array of aggregate results instead of single object

## Migration Plan

**Deployment:**
1. Deploy code with new aggregate query field
2. Existing queries unaffected (backward compatible)
3. Update documentation and examples
4. Announce feature in release notes

**Rollback:**
- Remove aggregate field from Query type
- No data migration needed (read-only feature)
- No breaking changes to existing functionality

**Monitoring:**
- Track aggregate query usage and performance
- Monitor slow query logs for unindexed aggregates
- Collect feedback on missing aggregate operations

## Open Questions

1. **Should we limit the number of concurrent aggregate operations in a single query?**
   - Decision: No limit initially. Monitor and add if needed.

2. **Should we support aggregation on relation fields?**
   - Decision: No, deferred. Requires complex JOIN logic.

3. **Should we expose DISTINCT count?**
   - Decision: Deferred. Can add `_count: { field: { distinct: true } }` later.

4. **How to handle timezone for date/time MIN/MAX?**
   - Decision: Use database default timezone. Document behavior.
