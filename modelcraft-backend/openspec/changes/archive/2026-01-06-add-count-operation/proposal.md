# Change: Add Count Operation for Record Counting

## Why

Users need a simple, efficient way to count records without fetching full result sets. While the existing `aggregate` operation supports `_count`, it requires understanding the aggregate input structure and nested result format. A dedicated `count` operation provides a cleaner, more intuitive API for the common use case of "how many records match this filter?"

This follows the Prisma pattern where both `count()` and `aggregate()._count` coexist with clear separation of concerns:
- `count`: Simple record counting with optional WHERE filtering, returns `{count: N}`
- `aggregate._count`: Field-level null/non-null counting as part of statistical aggregations

## What Changes

- **Add** new `count` query operation to the runtime GraphQL schema
- **Add** `count` operation supporting optional `where` parameter for filtering
- **Add** `count` operation supporting optional `select` parameter for field-level counts
- **Add** `CountResult` type returning `{count: N}` for simple counts or `{fieldsCount: {_all: N, field1: N, ...}}` for field-level counts
- **Add** constant `OperationCount` to query constants
- **Add** SQL generation for COUNT queries with WHERE clause support
- **Update** Query API specification to document count operation semantics

**Breaking Changes**: None. This is a purely additive change.

## Impact

### Affected Specs
- `query-api`: Add count operation requirements

### Affected Code
- `internal/domain/modelruntime/graphql_constants.go`: Add OperationCount constant
- `internal/domain/modelruntime/graphql_input_types.go`: Add GenerateCountArgs method
- `internal/domain/modelruntime/model_resolver.go`: Add createCountField and executeCount methods
- `internal/domain/modelruntime/query_executor.go`: Add count SQL generation logic
- Runtime GraphQL schema generation: Add count field to Query type

### Performance Considerations
- COUNT queries are more efficient than findMany + length for large datasets
- Uses database COUNT(*) and COUNT(field) which are optimized operations
- WHERE filtering happens at database level for optimal performance
- No data transfer overhead beyond the count result

## Migration Path

N/A - This is a new feature with no breaking changes. Existing code continues to work unchanged.

## Industry Alignment

This design follows **Prisma's dual API pattern**:

**Prisma example:**
```typescript
// Simple count - returns number
const count = await prisma.user.count()
// Result: 42

// Count with WHERE filter
const activeCount = await prisma.user.count({
  where: { status: { equals: 'active' } }
})
// Result: 18

// Field-level counts with select
const fieldCounts = await prisma.user.count({
  select: {
    _all: true,
    name: true,
    email: true
  }
})
// Result: { fieldsCount: { _all: 30, name: 28, email: 25 } }

// Combined: filter + field counts
const filteredFieldCounts = await prisma.user.count({
  where: { status: { equals: 'active' } },
  select: {
    _all: true,
    name: true
  }
})
// Result: { fieldsCount: { _all: 18, name: 16 } }
```

**GraphQL Query Examples:**
```graphql
# Simple count
query {
  count
}
# Returns: { count: 42 }

# Count with filter
query {
  count(where: { status: { equals: "active" } })
}
# Returns: { count: 18 }

# Field-level counts
query {
  count(select: { _all: true, name: true, email: true })
}
# Returns: { fieldsCount: { _all: 30, name: 28, email: 25 } }

# Filter + field counts
query {
  count(
    where: { status: { equals: "active" } }
    select: { _all: true, name: true }
  )
}
# Returns: { fieldsCount: { _all: 18, name: 16 } }
```

**Other ORMs comparison:**
- **TypeORM**: `count({ where: {...} })` returns number
- **Sequelize**: `count({ where: {...} })` returns number
- **Django ORM**: `filter(...).count()` returns number
- **Entity Framework**: `Where(...).Count()` returns number

The dual API approach (dedicated `count` + aggregate's `_count`) provides both simplicity and power while maintaining semantic clarity.
