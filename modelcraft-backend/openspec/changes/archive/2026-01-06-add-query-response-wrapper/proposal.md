# Proposal: Add Query Response Wrapper with Metadata Fields

## Change ID
`add-query-response-wrapper`

## Summary
Add response wrapper objects to all query operations (findUnique, findFirst, findMany) to separate data payload from operational metadata fields (timeCost, reqId). Use semantic field naming: `item` for singular results, `items` for plural results.

## Problem Statement

Currently, ModelCraft GraphQL query operations return data directly:

```graphql
query {
  findUnique(where: {id: "123"}) {
    id
    name
  }
}
```

**Limitations**:
1. No place to add operational metadata (request timing, tracking IDs)
2. Cannot add pagination metadata (totalCount, hasNextPage) without breaking changes
3. Inconsistent with mutation operations which already return wrapped results
4. Difficult to add cross-cutting concerns (caching, performance monitoring)

## Proposed Solution

Wrap all query responses with result objects containing:
1. **Data field**: The actual query results
   - `item` for singular operations (findUnique, findFirst)
   - `items` for plural operations (findMany)
2. **Metadata fields**: timeCost, reqId, and future extensibility

### Design Rationale

**Why `item`/`items` instead of `data`/`data`?**
- ✅ Semantic clarity: `item` (singular) vs `items` (plural) is self-documenting
- ✅ Linguistic consistency: Same word root, just singular/plural
- ✅ Type-safe: Client code can distinguish singular vs array at field name level
- ✅ Avoids ambiguity: `data` doesn't indicate cardinality
- ✅ Simple and intuitive: No GraphQL jargon (like `nodes`)

**Why not `nodes`/`edges`?**
- While `nodes` is the GraphQL Relay standard, `items` is simpler and more universal
- We can migrate to Connection pattern later if needed without breaking changes
- `items` is more accessible to developers from REST backgrounds

## Impact Analysis

### Breaking Changes
⚠️ **This is a breaking change affecting all query operations.**

**Before**:
```graphql
query {
  findMany { id name }
}
# Returns: [{ id, name }, ...]
```

**After**:
```graphql
query {
  findMany {
    items { id name }
    timeCost
    reqId
  }
}
# Returns: { items: [...], timeCost: 15, reqId: "..." }
```

**Note**: Migration strategy and backward compatibility will be handled separately.

## Capabilities Affected

- `query-api` - Modified (all query operations)

## Alternatives Considered

### Alternative 1: Use `data` for all operations
**Rejected**: Semantically ambiguous (data is singular or plural?)

### Alternative 2: Use `nodes` (GraphQL standard)
**Rejected**: More complex than needed; can migrate later if pagination becomes critical

### Alternative 3: Add metadata at GraphQL extensions level
**Rejected**: Extensions are not queryable in GraphQL; cannot select metadata fields

### Alternative 4: Keep direct return, add separate metadata query
**Rejected**: Requires two queries; no correlation between data and metadata

## Benefits

1. **Operational Observability**: Track request performance and tracing
2. **Future-Proof**: Room for pagination metadata without breaking changes
3. **Consistency**: Aligns query operations with mutation operations
4. **Semantic Clarity**: Clear distinction between singular and plural results
5. **Extensibility**: Easy to add new metadata fields (serverTime, cacheHit, etc.)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking change impact | Breaking change acknowledged; migration handled separately |
| Client code complexity | Wrapper adds minimal nesting; offset by clarity |
| Performance overhead | Minimal; metadata is computed once per request |

## Dependencies

None. This is a self-contained change to query operations.

## Timeline Estimate

- Proposal review: 1-2 days
- Implementation: 3-5 days
- Testing: 2 days
- Documentation: 1 day

**Total: ~7-10 days**

## Success Criteria

1. All query operations return wrapped results with metadata
2. `timeCost` accurately reflects query execution time
3. `reqId` is unique per request and traceable in logs
4. Comprehensive tests for all operations
5. Documentation updated with new response structure

## Open Questions

1. What additional metadata fields should we plan for (serverTime, apiVersion)?
2. Should we add optional `totalCount` to findMany immediately or defer?

## References

- Related specs: `openspec/specs/query-api/spec.md`
- Mutation operations already use wrapper pattern (create, update, delete)
- Industry research: `WRAPPER_FIELD_ANALYSIS.md`, `METADATA_WRAPPER_PROPOSAL.md`
