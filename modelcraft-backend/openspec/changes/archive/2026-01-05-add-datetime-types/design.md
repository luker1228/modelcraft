# Design Document: Complete DateTime, Date and Time Type Support

## Context

ModelCraft currently has partial support for date/time field types. The domain layer defines three format types (`FormatDateTime`, `FormatDate`, `FormatTime`), and these can be persisted to MySQL using appropriate column types (DATETIME, DATE, TIME). However, the GraphQL runtime layer has incomplete implementation:

- `FormatDate` and `FormatTime` are mapped to `graphql.String` instead of proper scalar types
- No custom GraphQL scalar definitions exist for Date and Time
- WHERE clause filtering doesn't properly handle Date/Time comparisons
- No validation exists for date/time formats or ranges

This creates type safety issues, prevents date/time queries, and reduces API usability.

### Current State

**Domain Layer (modeldesign):**
- ✅ `FormatDate`, `FormatTime`, `FormatDateTime` defined in field types
- ✅ MySQL type mapping exists (DATE, TIME, DATETIME columns)
- ❌ No format validation in field validator
- ❌ No range constraints (minDate, maxDate, minTime, maxTime) in ValidationConfig

**Runtime Layer (modelruntime):**
- ✅ `FormatDateTime` maps to `graphql.DateTime` (provided by graphql-go)
- ❌ `FormatDate` and `FormatTime` map to `graphql.String` (incorrect)
- ❌ No custom Date/Time scalar implementations
- ❌ WHERE clause filtering doesn't handle date/time fields properly

### Stakeholders

- **API Users**: Need type-safe date/time fields and the ability to query by date/time ranges
- **GraphQL Clients**: Need proper scalar types for type generation and validation
- **Backend Developers**: Need consistent validation and SQL generation for date/time operations

## Goals / Non-Goals

### Goals

1. **Type Safety**: Expose Date and Time as proper GraphQL scalar types (not String)
2. **Format Validation**: Validate date/time formats at the API layer before persistence
3. **Range Constraints**: Support min/max date/time validation for business rules
4. **Query Filtering**: Enable WHERE clause operations (gt, lt, gte, lte, between) on date/time fields
5. **Consistency**: Align all three date/time types (Date, Time, DateTime) in implementation approach

### Non-Goals

- **Database Migration**: No changes to MySQL column types or existing data
- **Timezone Conversion**: No automatic timezone handling beyond ISO 8601 parsing (database stores as-is)
- **Calendar Operations**: No business logic for date arithmetic, holidays, or locale-specific formatting
- **New Field Types**: No addition of other date/time types (e.g., Duration, Interval)
- **Client Libraries**: No automatic updates to client-side GraphQL code (clients must regenerate from schema)

## Decisions

### Decision 1: Custom GraphQL Scalar Implementations

**What**: Implement custom `Date` and `Time` GraphQL scalars in Go following the graphql-go pattern.

**Why**:
- GraphQL does not have built-in Date/Time scalars (only Int, Float, String, Boolean, ID)
- graphql-go provides `graphql.DateTime` but not Date or Time
- Custom scalars provide type safety and format validation at the GraphQL layer
- Follows existing pattern established by `graphql.DateTime`

**Alternatives Considered**:
1. **Keep using String type**: Rejected because it loses type safety and requires client-side validation
2. **Use DateTime for all three types**: Rejected because Date and Time have different semantics and storage requirements
3. **Use third-party scalar library**: Rejected to minimize dependencies and maintain control over format rules

**Implementation Location**: New file `internal/domain/modelruntime/graphql_scalars.go`

**Format Standards**:
- **Date**: ISO 8601 `YYYY-MM-DD` (e.g., "2024-01-15")
- **Time**: 24-hour `HH:MM:SS` (e.g., "14:30:00")
- **DateTime**: ISO 8601 with timezone (e.g., "2024-01-15T14:30:00Z")

### Decision 2: ValidationConfig Extension for Range Constraints

**What**: Extend `ValidationConfig` struct with four new optional fields: `MinDate`, `MaxDate`, `MinTime`, `MaxTime`.

**Why**:
- Common business requirement (e.g., "booking date must be in the future", "event time must be during business hours")
- Follows existing pattern (ValidationConfig already has `Minimum`/`Maximum` for numbers)
- Backward compatible (new fields are optional pointers, default to nil)
- Centralized validation logic in domain layer

**Alternatives Considered**:
1. **Custom validation rules only**: Rejected because min/max date is a common pattern deserving first-class support
2. **Separate DateValidationConfig struct**: Rejected to keep ValidationConfig unified and avoid type proliferation
3. **Store as time.Time instead of string**: Rejected to maintain consistency with JSON serialization and avoid parsing complexity in config

**Validation Behavior**:
- Format validation happens first (ensure valid ISO 8601 / HH:MM:SS)
- Range validation happens second (ensure within min/max if specified)
- MinDate/MaxDate apply to both Date and DateTime fields
- MinTime/MaxTime apply only to Time fields
- Overnight time ranges (minTime > maxTime) are supported (e.g., "22:00:00" to "06:00:00")

### Decision 3: WHERE Clause Filtering Implementation

**What**: Extend `internal/domain/modelruntime/graphql_field_conditions.go` to handle Date, Time, and DateTime fields in WHERE clause generation.

**Why**:
- Date/time queries are a primary use case for these field types
- Needs to integrate with existing WHERE clause infrastructure
- Must use SQL parameter binding (not string concatenation) to prevent SQL injection
- Should support all comparison operators for consistency with numeric fields

**Supported Operators**:
- `eq`: Exact match (e.g., `dateField: { eq: "2024-01-15" }`)
- `ne`: Not equal
- `gt`: Greater than (e.g., "find orders after 2024-01-01")
- `gte`: Greater than or equal
- `lt`: Less than (e.g., "find events before 2024-12-31")
- `lte`: Less than or equal
- `between`: Range query (e.g., `dateField: { between: ["2024-01-01", "2024-12-31"] }`)

**SQL Generation Approach**:
- Use goqu query builder's existing comparison methods
- Pass date/time values as SQL parameters (prepared statements)
- MySQL will handle date/time comparison using native DATE/TIME/DATETIME column types
- No need for special date parsing in SQL (database handles it)

**Alternatives Considered**:
1. **String-based comparison**: Rejected because MySQL native date comparison is more efficient and correct
2. **Convert all to Unix timestamps**: Rejected because it loses readability and timezone information
3. **Only support equality**: Rejected because range queries are essential for date/time use cases

### Decision 4: Breaking Change Handling

**What**: Treat the change from String to Date/Time scalars as a breaking GraphQL schema change.

**Why**:
- Existing clients expect String type for Date and Time fields
- Changing to scalar types will break GraphQL introspection and code generation
- Clients must update to handle proper scalar types
- However, underlying MySQL data and storage format is unchanged

**Migration Strategy**:
1. Deploy updated GraphQL schema with new scalar types
2. Notify clients of the breaking change
3. Clients regenerate GraphQL types from updated schema introspection
4. No database migration required (MySQL columns remain DATE/TIME)

**Impact**:
- ✅ No data loss or migration
- ✅ Improved type safety going forward
- ❌ Clients must update code (breaking change)

**Alternatives Considered**:
1. **API versioning (v1 vs v2)**: Rejected because it adds complexity and this is a foundational type fix
2. **Feature flag to toggle behavior**: Rejected because it complicates schema generation and is temporary
3. **Gradual rollout with both String and scalar types**: Rejected because it creates schema ambiguity

### Decision 5: Format Validation Approach

**What**: Validate date/time formats using Go's `time.Parse()` with strict format strings.

**Why**:
- Go standard library is reliable and well-tested
- `time.Parse()` with explicit format strings enforces exact format matching
- No need for regex or manual parsing (error-prone)
- Consistent with how DateTime is likely handled

**Format Strings**:
- Date: `"2006-01-02"` (ISO 8601 date)
- Time: `"15:04:05"` (24-hour time with seconds)
- DateTime: `time.RFC3339` (ISO 8601 with timezone)

**Error Messages**:
- Clearly indicate expected format (e.g., "expected format YYYY-MM-DD, got 'invalid-date'")
- Return validation error before attempting database operation
- Consistent with existing field validation error patterns

**Alternatives Considered**:
1. **Regex validation**: Rejected because it doesn't validate actual date validity (e.g., Feb 30)
2. **Database-level validation only**: Rejected because client should get clear error before SQL execution
3. **Flexible parsing (multiple formats)**: Rejected to maintain consistency and avoid ambiguity

## Risks / Trade-offs

### Risk 1: Breaking Change Impact

**Risk**: Existing GraphQL clients will break when Date/Time fields change from String to scalar types.

**Mitigation**:
- Clearly document the breaking change in release notes
- Provide migration guide for client applications
- Consider deprecation period or communication plan
- Test with representative client applications before release

**Trade-off**: Short-term breaking change for long-term type safety and correctness.

### Risk 2: Timezone Handling Complexity

**Risk**: DateTime fields with timezone information may cause confusion (stored as-is in MySQL DATETIME, no timezone conversion).

**Mitigation**:
- Document timezone handling behavior clearly
- Validate that DateTime values include timezone (ISO 8601 requires it)
- Consider future enhancement for timezone normalization (non-goal for now)

**Trade-off**: Simple implementation now, may need timezone handling enhancements later.

### Risk 3: Performance Impact of Validation

**Risk**: Format and range validation on every field value could add latency.

**Mitigation**:
- Use efficient `time.Parse()` (stdlib is optimized)
- Validation only runs on create/update operations (not queries)
- Consider caching parsed time values if needed (premature optimization for now)

**Trade-off**: Slight performance cost for much better data integrity and error messages.

### Risk 4: Overnight Time Ranges

**Risk**: Time ranges that span midnight (e.g., 22:00:00 to 06:00:00) are complex to validate and may be confusing.

**Mitigation**:
- Document overnight range behavior clearly
- Provide examples in API documentation
- Write comprehensive unit tests for edge cases
- Consider whether this use case is actually needed (could be deferred)

**Trade-off**: Added complexity vs. supporting a legitimate use case (e.g., "night shift hours").

## Migration Plan

### Phase 1: Domain Layer Validation (Non-Breaking)
1. Add `MinDate`, `MaxDate`, `MinTime`, `MaxTime` to `ValidationConfig`
2. Implement format validation for Date, Time, DateTime in field validator
3. Implement range validation
4. Write unit tests
5. **Deploy**: Non-breaking (new fields are optional, unused if not specified)

### Phase 2: GraphQL Scalar Types (Breaking Change)
1. Create `internal/domain/modelruntime/graphql_scalars.go` with Date and Time scalars
2. Update `getGraphqlTypeBy()` to map FormatDate → Date, FormatTime → Time
3. Add `scalar Date` and `scalar Time` to `api/graph/schema/base.graphql`
4. Write unit tests for scalar serialization/deserialization
5. Write integration tests with schema generation
6. **Deploy**: Breaking change - notify clients, provide migration guide

### Phase 3: Query Filtering Support
1. Extend `graphql_field_conditions.go` to handle Date/Time fields in WHERE clauses
2. Support all comparison operators (eq, ne, gt, gte, lt, lte, between)
3. Write unit tests for SQL generation
4. Write integration tests for query filtering
5. **Deploy**: Enhancement - no breaking change (new functionality)

### Rollback Plan
- **Phase 1**: Can be rolled back without data loss (validation config is additive)
- **Phase 2**: Rollback requires reverting type mapping (clients would need to regenerate schemas again)
- **Phase 3**: Can be rolled back without data loss (filtering is additive functionality)

### Testing Strategy
- **Unit Tests**: Validation logic, scalar serialization, SQL generation
- **Integration Tests**: End-to-end with models containing Date/Time fields, query filtering
- **Manual Testing**: Test with representative client applications (if available)
- **Backward Compatibility**: Test that existing models without Date/Time fields are unaffected

## Open Questions

1. **Should we support flexible date parsing (e.g., "2024/01/15" or "01-15-2024")?**
   - **Current Decision**: No, strict ISO 8601 only for consistency
   - **Rationale**: Avoids ambiguity and follows industry standards

2. **Should overnight time ranges (minTime > maxTime) be supported?**
   - **Current Decision**: Yes, implement in validation logic
   - **Rationale**: Legitimate use case for shift work, events spanning midnight

3. **Should we add timezone normalization for DateTime fields?**
   - **Current Decision**: No, store as-is (non-goal for this change)
   - **Rationale**: Adds complexity, can be added later if needed

4. **Should we provide a deprecation period for the String → Scalar breaking change?**
   - **Current Decision**: TBD - depends on client base and release process
   - **Recommendation**: If possible, announce in advance and provide migration guide

5. **Should we support date arithmetic in GraphQL queries (e.g., "30 days from now")?**
   - **Current Decision**: No, out of scope
   - **Rationale**: Complex feature requiring time zone awareness and calendar logic
