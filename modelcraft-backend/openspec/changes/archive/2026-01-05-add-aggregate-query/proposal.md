# Change: Add Aggregate Query to ModelRuntime

## Why

ModelCraft currently supports `findUnique`, `findFirst`, and `findMany` queries for data retrieval, but lacks aggregation capabilities for statistical analysis. Users need to perform aggregate operations like counting records, calculating averages, sums, minimums, and maximums on their data, similar to Prisma's aggregate API.

Without aggregate support, users must:
- Fetch all records and perform calculations client-side (inefficient for large datasets)
- Write custom SQL queries outside the GraphQL API (breaks the unified API approach)
- Use external tools for basic statistical operations

## What Changes

- Add new `aggregate` query operation to GraphQL runtime schema
- Implement aggregate input types supporting:
  - `_count`: Count records (total or specific non-null fields)
  - `_avg`: Calculate average values for numeric fields
  - `_sum`: Calculate sum of numeric field values
  - `_min`: Find minimum value in a field
  - `_max`: Find maximum value in a field
- Support `where` filter conditions for conditional aggregation
- Generate aggregate result types dynamically based on model schema
- Implement SQL aggregation query builder using goqu library
- Add aggregate repository method to ClientDatabaseRepository interface

This change follows the existing patterns for query operations (findUnique, findFirst, findMany) and reuses the where condition handling infrastructure.

## Impact

**Affected specs:**
- `modelruntime-aggregate` (NEW): Aggregate query capability specification

**Affected code:**
- `internal/domain/modelruntime/model_resolver.go`: Add aggregate field to root query
- `internal/domain/modelruntime/graphql_repository.go`: Add Aggregate method to interface
- `internal/domain/modelruntime/graphql_input.go`: Add AggregateInput struct and constructor
- `internal/domain/modelruntime/graphql_input_types.go`: Add aggregate input type generation
- `internal/domain/modelruntime/graphql_constants.go`: Add aggregate operation constants
- `internal/infrastructure/database/dml/client_db_repo_impl.go`: Implement Aggregate method
- `internal/infrastructure/database/dml/sql_mapper.go`: Add convertAggregateInputToSQL function

**Breaking changes:** None - this is a pure addition to the Query type

**Migration path:** N/A - no breaking changes
