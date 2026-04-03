# Design: Query Response Wrapper with Metadata

## Overview

This design adds wrapper types to all query operations (findUnique, findFirst, findMany) to separate data payload from operational metadata. Uses semantic field naming: `item` (singular) and `items` (plural).

## Architecture

### Type Hierarchy

```
QueryResultMeta (interface)
├── FindUniqueResult (implements QueryResultMeta)
├── FindFirstResult (implements QueryResultMeta)
└── FindManyResult (implements QueryResultMeta)
```

### GraphQL Schema

```graphql
# Common metadata interface
interface QueryResultMeta {
  timeCost: Int!          # Query execution time in milliseconds
  reqId: String!          # Request tracking ID (UUID v7)
}

# For findUnique
type UserFindUniqueResult implements QueryResultMeta {
  item: User              # Single object (nullable - may not exist)
  timeCost: Int!
  reqId: String!
}

# For findFirst
type UserFindFirstResult implements QueryResultMeta {
  item: User              # Single object (nullable - may not exist)
  timeCost: Int!
  reqId: String!
}

# For findMany
type UserFindManyResult implements QueryResultMeta {
  items: [User!]!         # Array (never null, but can be empty)
  totalCount: Int         # Total matching records (optional for now)
  timeCost: Int!
  reqId: String!
}

# Updated Query type
type Query {
  findUnique(where: UniqueWhereInput!): UserFindUniqueResult
  findFirst(where: WhereInput, orderBy: [OrderByInput!]): UserFindFirstResult
  findMany(where: WhereInput, orderBy: [OrderByInput!], take: Int, skip: Int): UserFindManyResult
  # ... other operations
}
```

## Implementation Strategy

### 1. Context-Level Metadata Tracking

Add middleware to track request metadata:

```go
// pkg/requestcontext/metadata.go
type RequestMetadata struct {
    ReqID     string
    StartTime time.Time
}

func WithMetadata(ctx context.Context) context.Context {
    metadata := &RequestMetadata{
        ReqID:     generateReqID(), // UUID v7
        StartTime: time.Now(),
    }
    return context.WithValue(ctx, metadataKey, metadata)
}

func GetMetadata(ctx context.Context) *RequestMetadata {
    return ctx.Value(metadataKey).(*RequestMetadata)
}

func CalculateTimeCost(ctx context.Context) int {
    metadata := GetMetadata(ctx)
    return int(time.Since(metadata.StartTime).Milliseconds())
}
```

### 2. Resolver Modifications

#### Before (current):
```go
func (m *graphqlModelResolver) executeFindMany(p graphql.ResolveParams) ([]map[string]any, error) {
    // ... query logic
    return result, nil
}

func (m *graphqlModelResolver) createFindManyField(modelType graphql.Type) (*graphql.Field, error) {
    return &graphql.Field{
        Type: graphql.NewList(graphql.NewNonNull(modelType)),
        // ...
    }, nil
}
```

#### After (proposed):
```go
func (m *graphqlModelResolver) executeFindMany(p graphql.ResolveParams) (map[string]any, error) {
    startTime := time.Now()
    
    // Execute query
    items, err := m.clientRepo.FindMany(m.ctx, input)
    if err != nil {
        return nil, err
    }
    
    // Calculate metadata
    metadata := requestcontext.GetMetadata(m.ctx)
    timeCost := int(time.Since(startTime).Milliseconds())
    
    // Wrap response
    return map[string]any{
        "items":    items,
        "timeCost": timeCost,
        "reqId":    metadata.ReqID,
    }, nil
}

func (m *graphqlModelResolver) createFindManyField(modelType graphql.Type) (*graphql.Field, error) {
    resultType := m.createFindManyResultType(modelType)
    return &graphql.Field{
        Type: resultType, // Now returns wrapper type
        // ...
    }, nil
}

func (m *graphqlModelResolver) createFindManyResultType(modelType graphql.Type) *graphql.Object {
    return graphql.NewObject(graphql.ObjectConfig{
        Name: m.model.Name + "FindManyResult",
        Fields: graphql.Fields{
            "items": &graphql.Field{
                Type:        graphql.NewList(graphql.NewNonNull(modelType)),
                Description: "Array of matching items",
            },
            "totalCount": &graphql.Field{
                Type:        graphql.Int,
                Description: "Total number of matching records (optional)",
            },
            "timeCost": &graphql.Field{
                Type:        graphql.NewNonNull(graphql.Int),
                Description: "Query execution time in milliseconds",
            },
            "reqId": &graphql.Field{
                Type:        graphql.NewNonNull(graphql.String),
                Description: "Request tracking ID",
            },
        },
        Description: "Result wrapper for findMany query",
    })
}
```

### 3. Request ID Generation

Use UUID v7 for sortable, time-based request IDs:

```go
// pkg/bizutils/uuid.go (already exists)
func GenerateRequestID() (string, error) {
    return GenerateUUIDV7() // Already implemented
}
```

### 4. Type Generation Pattern

Each model generates three result types:
- `{Model}FindUniqueResult`
- `{Model}FindFirstResult`
- `{Model}FindManyResult`

Pattern is consistent and predictable for code generation.

## Data Flow

### Request Lifecycle

```
1. HTTP Request → GraphQL Handler
   ↓
2. Middleware: Inject RequestMetadata into context
   - Generate reqId (UUID v7)
   - Record startTime
   ↓
3. GraphQL Resolver: executeFindMany
   - Record operation startTime
   - Execute query
   - Calculate timeCost
   - Wrap result with metadata
   ↓
4. Response: Return wrapped result
   {
     "items": [...],
     "timeCost": 15,
     "reqId": "01930c8a-..."
   }
```

## Performance Considerations

### Overhead Analysis

| Component | Overhead | Mitigation |
|-----------|----------|------------|
| UUID generation | ~1µs | Negligible; done once per request |
| Time calculation | ~100ns | Negligible; simple subtraction |
| Wrapper object | ~0 bytes | Go compiler optimizes map[string]any |
| GraphQL serialization | +2 fields | Minimal JSON overhead (~50 bytes) |

**Total overhead**: < 1ms, < 100 bytes per request

### Memory Impact

- `reqId`: 36 bytes (string)
- `timeCost`: 8 bytes (int64)
- Wrapper map: ~48 bytes (Go map overhead)

**Total**: ~92 bytes per query response

### Caching Implications

Response wrapper enables future caching strategies:
- Cache key can include reqId for debugging
- timeCost can drive cache TTL decisions
- Metadata doesn't affect cache validity

## Extensibility

### Future Metadata Fields

Easy to add without breaking changes:

```graphql
type UserFindManyResult {
  items: [User!]!
  totalCount: Int
  timeCost: Int!
  reqId: String!
  
  # Future additions (non-breaking)
  serverTime: DateTime    # Server timestamp
  apiVersion: String      # API version used
  cacheHit: Boolean       # Whether result was cached
  queryComplexity: Int    # GraphQL query complexity score
}
```

### Migration to Relay Connections

If cursor pagination is needed later, easy to migrate:

```graphql
type UserFindManyResult {
  items: [User!]!         # Keep for simple use cases
  edges: [UserEdge!]      # Add for cursor pagination
  pageInfo: PageInfo      # Add for pagination
  totalCount: Int
  timeCost: Int!
  reqId: String!
}
```

## Error Handling

### Error Response Structure

Errors follow standard GraphQL error format (unchanged):

```json
{
  "data": {
    "findMany": null
  },
  "errors": [
    {
      "message": "Database connection failed",
      "path": ["findMany"],
      "extensions": {
        "code": "DATABASE_ERROR",
        "reqId": "01930c8a-...",
        "timeCost": 5003
      }
    }
  ]
}
```

**Note**: Include reqId and timeCost in error extensions for debugging.

## Testing Strategy

### Unit Tests

1. **Metadata generation**
   - Test reqId uniqueness
   - Test timeCost accuracy
   - Test context propagation

2. **Wrapper construction**
   - Test item field population
   - Test items field population
   - Test metadata field inclusion

3. **Type generation**
   - Test result type creation for each operation
   - Test field definitions match schema

### Integration Tests

1. **End-to-end queries**
   - Test findUnique returns wrapped result
   - Test findFirst returns wrapped result
   - Test findMany returns wrapped result
   - Test metadata values are reasonable

2. **Error scenarios**
   - Test errors still include metadata
   - Test null item handling
   - Test empty items handling

### Performance Tests

1. **Overhead measurement**
   - Benchmark query execution with/without wrapper
   - Verify overhead < 1ms
   - Memory profiling

## Monitoring and Observability

### Metrics to Track

1. **Query Performance**
   - Average timeCost per operation type
   - P50, P95, P99 timeCost percentiles
   - Slow query identification (timeCost > threshold)

2. **Request Tracking**
   - reqId correlation across logs
   - Request volume by operation
   - Error rate by operation

3. **Client Adoption**
   - Usage of new wrapper fields
   - Queries still using old format (if compatibility layer exists)

### Logging Example

```go
logger.Infof("query_executed",
    "operation", "findMany",
    "model", m.model.Name,
    "reqId", metadata.ReqID,
    "timeCost", timeCost,
    "resultCount", len(items),
)
```

## Decision Log

| Decision | Rationale | Alternatives Considered |
|----------|-----------|-------------------------|
| Use `item`/`items` naming | Semantic clarity, linguistic consistency | `data`/`data`, `data`/`nodes`, `record`/`records` |
| Include timeCost and reqId | Essential for observability | serverTime, apiVersion (deferred) |
| UUID v7 for reqId | Sortable, collision-resistant | UUID v4, timestamp-based |
| Nullable item for singular | Match existing null semantics | Throw error on not found (breaking) |
| Non-null items array | Empty array for no matches | Nullable array (less ergonomic) |
| Interface for metadata | Type safety, consistency | No interface (code duplication) |

## Dependencies

### Internal
- `pkg/bizutils` - UUID v7 generation (already exists)
- `pkg/logfacade` - Logging (already exists)
- `internal/domain/modelruntime` - Resolver modifications

### External
- `graphql-go/graphql` - GraphQL type system (already used)

## Timeline

| Phase | Duration | Description |
|-------|----------|-------------|
| Design review | 1-2 days | Approve architecture and naming |
| Implementation | 3-5 days | Modify resolvers, add metadata tracking |
| Testing | 2 days | Unit, integration, performance tests |
| Documentation | 1 day | API docs, changelog |
| Deployment | 1 day | Release |

**Total**: ~7-10 days
