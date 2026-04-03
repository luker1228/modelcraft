# Proposal: Replace notIn with Logical NOT Operator

## Change ID
`replace-notin-with-logical-not`

## Type
Breaking Change

## Status
Proposed

## Context

The current query API includes a `notIn` field-level operator that allows checking if a field value is not in a list:
```graphql
{ status: { notIn: ["draft", "archived"] } }
```

This creates inconsistency with Prisma's design philosophy where:
- **Logical operators** (uppercase): `AND`, `OR`, `NOT` - operate on entire condition blocks
- **Field operators** (camelCase): `equals`, `not`, `in`, `contains`, etc. - operate on single field values

The confusion between `not` (field-level "not equals") and `NOT` (logical negation) is particularly problematic.

## Problem Statement

1. **Semantic confusion**: `notIn` is a field-level operator that semantically performs logical negation, blurring the boundary between field and logical operators
2. **Missing logical NOT**: There is currently no way to negate arbitrary complex conditions (e.g., `NOT: { AND: [...] }`)
3. **Inconsistent naming**: Having `notIn` but no `notContains`, `notStartsWith`, etc. is inconsistent
4. **Prisma divergence**: Prisma uses logical `NOT` for all negation needs, making `notIn` unnecessary

## Proposed Solution

### Remove `notIn` operator
Remove the `notIn` field-level operator entirely from:
- Query builder API (`FieldNotIn` constant and `NotIn()` functions)
- GraphQL schema generation (remove `notIn` from input types)
- Reserved keywords list
- Documentation

### Add logical `NOT` operator
Implement a logical `NOT` operator that:
- Uses uppercase `NOT` as the operator key (consistent with `AND`, `OR`)
- Accepts a single condition object or array of conditions to negate
- Supports negating any valid condition structure

### Migration path
Users replace `notIn` usage with equivalent `NOT` + `in` combinations:

**Before:**
```graphql
{ status: { notIn: ["draft", "archived"] } }
```

**After:**
```graphql
{ NOT: { status: { in: ["draft", "archived"] } } }
```

## Examples

### Example 1: Get posts where title does not contain "SQL"
```graphql
{
  NOT: {
    title: { contains: "SQL" }
  }
}
```

### Example 2: Get posts where status is not draft or archived (replacing notIn)
```graphql
{
  NOT: {
    status: { in: ["draft", "archived"] }
  }
}
```

### Example 3: Complex negation
```graphql
{
  NOT: {
    AND: [
      { published: { equals: true } },
      { views: { gt: 1000 } }
    ]
  }
}
```

### Example 4: Multiple NOT conditions
```graphql
{
  AND: [
    { NOT: { status: { equals: "archived" } } },
    { NOT: { title: { contains: "draft" } } }
  ]
}
```

## Semantic Clarity

This change establishes clear semantic boundaries:

| Operator | Level | Purpose | Example |
|----------|-------|---------|---------|
| `not` (lowercase) | Field | Field not equals value | `{ status: { not: "draft" } }` |
| `NOT` (uppercase) | Logical | Negate entire condition(s) | `{ NOT: { status: { equals: "draft" } } }` |
| `in` | Field | Field in value list | `{ status: { in: ["published"] } }` |
| `AND` | Logical | All conditions must match | `{ AND: [...] }` |
| `OR` | Logical | Any condition must match | `{ OR: [...] }` |

## Impact Assessment

### Breaking Changes
1. Existing queries using `notIn` will fail
2. GraphQL schema changes - `notIn` field removed from input types
3. Query builder API changes - `FieldNotIn`, `NotIn()`, and `fb.NotIn()` removed

### Migration Required
- Update all queries using `notIn` to use `NOT: { field: { in: [...] } }`
- Update code using `NotIn()` functions
- Update GraphQL client code expecting `notIn` in schema

### Benefits
1. **Semantic clarity**: Clear distinction between field operators and logical operators
2. **Prisma alignment**: Matches Prisma's proven design patterns
3. **More expressive**: Can now negate arbitrary complex conditions
4. **Consistency**: All negation follows same logical pattern

## Alternative Approaches Considered

### Alternative 1: Keep both `notIn` and add `NOT`
**Rejected**: Adds redundancy and maintains semantic confusion

### Alternative 2: Add `NOT` only, deprecate `notIn` gradually
**Rejected**: Prolonged migration period maintains confusion

### Alternative 3: Keep `notIn`, don't add logical `NOT`
**Rejected**: Limits expressiveness and maintains Prisma divergence

## Dependencies
- None: This is a self-contained change to the query-api spec

## Related Capabilities
- `query-api`: The primary spec being modified

## Open Questions
None - approach has been clarified through user feedback

## Success Criteria
1. `notIn` operator completely removed from codebase
2. Logical `NOT` operator implemented and tested
3. All existing tests updated to use new syntax
4. Documentation updated with clear examples
5. No `notIn` references remain in code or docs
