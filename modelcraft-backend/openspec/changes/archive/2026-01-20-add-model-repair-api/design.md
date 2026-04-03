## Design: Model Repair API

## Context

### Problem Statement
Models in ModelCraft can become out of sync with their underlying database resources through various scenarios:
1. Failed partial deployments (model saved in platform DB but table creation failed)
2. Manual database operations (tables dropped, columns altered by users)
3. External schema management tools modifying tables
4. Cluster/database migration or reconfiguration

Currently, no mechanism exists to detect or repair this drift. Users must either delete and recreate models (inefficient and loses runtime state) or manually execute SQL (error-prone, bypasses platform management).

### Solution Overview
Add a new GraphQL mutation `repairModel` that:
1. Detects schema drift by comparing platform DB model definition with actual customer DB table structure
2. Provides three modes: dry-run (detect only), additive (add missing), full sync (add missing, delete extra)
3. Executes necessary DDL operations to bring the database in sync with the model definition
4. Returns comprehensive results showing issues, operations, and state comparison

### Stakeholders
- **Platform Users**: Need to fix broken models without manual SQL
- **Operations Team**: Need visibility into model schema health
- **Runtime Service**: Relies on models having correct underlying table structure

## Goals / Non-Goals

### Goals
1. Detect schema drift between model definition and actual database table
2. Repair missing tables and columns with configurable modes
3. Provide detailed audit trail of detected issues and executed operations
4. Reuse existing infrastructure components (SchemaIntrospector, DDLConverter, DeploymentImpl)
5. Support project-scoped models with proper isolation

### Non-Goals
1. **Data migration**: The repair API ensures schema correctness but does not migrate or validate existing data
2. **Schema versioning**: No historical tracking of schema changes during repair
3. **Conflict resolution**: Type/constraint mismatches are reported but not automatically resolved (requires manual intervention)
4. **Bulk repair**: Only single model repair per invocation (can be called multiple times for batch operations)

## Decisions

### Decision 1: Three Repair Modes

**What**: Support `DRY_RUN`, `ADDITIVE`, and `FULL_SYNC` modes.

**Why**:
- **DRY_RUN**: Allows users to preview changes before applying them, critical for production environments
- **ADDITIVE**: Safest repair mode for production, adds missing resources without deleting anything
- **FULL_SYNC**: For environments where extra columns should be removed to maintain strict consistency

**Alternatives Considered**:
1. Two modes (DETECT, REPAIR): Too coarse-grained, doesn't allow preview before apply
2. Single mode with flags (repairWithDelete=true/false): Mixes concerns, harder to validate
3. Five modes including type fix and constraint fix: Too complex, type/constraint mismatches require manual intervention

### Decision 2: Type Mismatches are Reported, Not Fixed

**What**: When field types don't match, the repair reports the issue but does not execute ALTER TABLE MODIFY COLUMN.

**Why**:
- **Data safety**: Automatically modifying column types can cause data loss or corruption
- **User control**: Users may have valid reasons for intentional type differences
- **Complexity**: Handling all type conversion cases correctly is complex and requires user validation

**Alternatives Considered**:
1. Auto-fix safe conversions (e.g., VARCHAR(50) → VARCHAR(100)): Adds complexity, still risks issues with character sets
2. Auto-fix with confirmation prompt: Not possible in GraphQL API design
3. Separate "fixTypes" mutation: Possible future enhancement

### Decision 3: Reuse Existing Infrastructure Components

**What**: Build repair logic by composing existing SchemaIntrospector, DDLConverter, TypeMapper, and DeploymentImpl.

**Why**:
- **Consistency**: Ensures repair uses same logic as model creation/deployment
- **Maintainability**: Avoids duplicate code; changes to DDL generation automatically apply to repair
- **Efficiency**: Components already optimized and tested

**Components Used**:
- `SchemaIntrospector`: Query `INFORMATION_SCHEMA.COLUMNS` for actual table structure
- `DDLConverter.GenerateCreateTableDDL()`: Generate CREATE TABLE statements
- `MySQLDDLBuilder.BuildAddColumns()`: Generate ADD COLUMN statements
- `TypeMapper`: Bidirectional type conversion for comparison
- `DeploymentImpl.ExecDDL()`: Execute DDL with error handling

### Decision 4: SchemaIssue Type System

**What**: Define clear issue types: `TABLE_MISSING`, `FIELD_MISSING`, `FIELD_TYPE_MISMATCH`, `FIELD_CONSTRAINT_MISMATCH`, `DATABASE_MISSING`, `CLUSTER_NOT_FOUND`.

**Why**:
- **Clear categorization**: Makes it easy for clients to understand the problem
- **Extensibility**: Easy to add new issue types without breaking API
- **Actionable**: Each issue type maps to a specific repair action or requires manual intervention

**Alternatives Considered**:
1. Generic "error" field with message: Too opaque, harder for programmatic handling
2. Severity levels (warning, error, critical): Adds complexity without clear benefit
3. Detailed diff object: Overkill for GraphQL payload

### Decision 5: Safety Checks for Field Deletion

**What**: Prevent deletion of system fields (id, createdAt, updatedAt) and fields referenced by relations.

**Why**:
- **System integrity**: System fields are required for runtime operation
- **Data consistency**: Removing foreign key fields breaks relations

**Implementation**:
- Check field name against system field list before adding to DROP COLUMN list
- Check field's `ParentRelationID` before adding to DROP COLUMN list
- Return `FIELD_HAS_DEPENDENCIES` error if deletion is blocked

### Decision 6: GraphQL Result Payload Structure

**What**: Return comprehensive payload with 8 fields including issues, DDL, before/after state.

**Why**:
- **Audit trail**: Users can see exactly what happened
- **Validation**: DRY_RUN can validate expected changes before applying
- **Debugging**: When things go wrong, full context is available

**Payload Structure**:
```graphql
type RepairModelPayload {
  model: Model                      # Updated model
  changesApplied: Boolean!          # Did anything change?
  detectedIssues: [SchemaIssue!]!   # All issues found
  executedDDL: [String!]!           # SQL that ran
  healthStatusBefore: String!       # Before repair
  healthStatusAfter: String!        # After repair
  extraFieldsRemoved: [String!]!    # Dropped fields
  fieldsAdded: [String!]!           # Added fields
}
```

## Architecture

### Service Layer Composition

```
GraphQL Resolver (model.resolvers.go)
    ↓
RepairModelUseCase (app/modeldesign/repair_app.go)
    ↓
SchemaComparisonService (domain/modeldesign/comparison_service.go)
    ↓
┌─────────────────────────────────┐
│ Existing Components              │
├─────────────────────────────────┤
│ - ModelRepository (load model)   │
│ - SchemaIntrospector (compare)   │
│ - DDLConverter (generate DDL)    │
│ - DeploymentImpl (execute DDL)   │
│ - TypeMapper (type conversion)   │
└─────────────────────────────────┘
```

### Data Flow

```
1. GraphQL Request: repairModel(modelId, mode, deleteExtraFields)
   ↓
2. Load model from platform DB (ModelRepository)
   ↓
3. Introspect actual table schema (SchemaIntrospector)
   ↓
4. Compare and detect issues (SchemaComparisonService)
   ↓
5. If mode == DRY_RUN: Return results without changes
   ↓
6. Generate repair DDL (DDLConverter, MySQLDDLBuilder)
   ↓
7. Execute DDL (DeploymentImpl)
   ↓
8. Build and return result payload
```

### Key Domain Entities

```go
// Issue types for schema drift detection
type SchemaIssueType string

const (
    TableMissing           SchemaIssueType = "TABLE_MISSING"
    FieldMissing           SchemaIssueType = "FIELD_MISSING"
    FieldTypeMismatch      SchemaIssueType = "FIELD_TYPE_MISMATCH"
    FieldConstraintMismatch SchemaIssueType = "FIELD_CONSTRAINT_MISMATCH"
    DatabaseMissing        SchemaIssueType = "DATABASE_MISSING"
    ClusterNotFound        SchemaIssueType = "CLUSTER_NOT_FOUND"
)

// Detected issue with details
type SchemaIssue struct {
    Type        SchemaIssueType
    Description string
    TableName   string
    FieldName   string  // Empty for TABLE_MISSING
    Details     map[string]interface{}
}

// Supported repair modes
type RepairMode string

const (
    DryRun   RepairMode = "DRY_RUN"
    Additive RepairMode = "ADDITIVE"
    FullSync RepairMode = "FULL_SYNC"
)

// Repair result with comprehensive information
type RepairResult struct {
    Model               *DataModel
    ChangesApplied      bool
    DetectedIssues      []SchemaIssue
    ExecutedDDL         []string
    HealthStatusBefore  string
    HealthStatusAfter   string
    ExtraFieldsRemoved  []string
    FieldsAdded         []string
}
```

## Risks / Trade-offs

### Risk 1: Race Conditions During Repair

**Issue**: While repair is running, another process might modify the same table.

**Mitigation**:
- Use database transactions where possible
- Document that repair should be run during maintenance windows in production
- Consider adding advisory locks in future versions

### Risk 2: Data Loss from Table Recreation

**Issue**: FULL_SYNC with table deletion drops all data.

**Mitigation**:
- Default to ADDITIVE mode to minimize data loss
- Add warning in documentation for FULL_SYNC with table repair
- Show which tables will be dropped in DRY_RUN before applying

### Risk 3: Permission Issues

**Issue**: Repair might fail due to insufficient database permissions.

**Mitigation**:
- Clear error messages indicating which permission is missing
- Document required database permissions for repair operations

### Trade-off 1: Complexity vs. Safety

**Decision**: Not auto-fixing type mismatches to prioritize data safety over convenience.

**Rationale**: Data integrity is more important than automation. Type changes need user validation.

### Trade-off 2: Granular Health Status

**Decision**: Three health levels (HEALTHY, NEEDS_REPAIR, BROKEN) rather than percentages.

**Rationale**: Simpler for clients to understand and act upon. More detail available in issues array.

## Migration Plan

### No Database Migration Required

- The repair API is implemented in application and domain layers
- No new tables or schema changes needed in platform DB
- Uses existing model, field_definitions tables

### Backwards Compatibility

- New mutation added without changing existing ones
- Existing model operations (create, update, delete, sync) are unaffected
- GraphQL schema extends existing types without breaking changes

### Rollback

- Remove repairModel mutation from GraphQL schema
- Delete repair service files from app and domain layers
- No rollback needed for data (no schema changes)

## Open Questions

None at this time. The design leverages existing, tested components and follows established patterns in the codebase.
