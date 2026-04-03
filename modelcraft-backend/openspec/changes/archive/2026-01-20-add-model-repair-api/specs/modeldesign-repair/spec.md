## ADDED Requirements
### Requirement: Schema Drift Detection

The system SHALL provide a domain service that compares ModelCraft model definitions against actual database schema to detect discrepancies.

#### Scenario: Detect missing table

- **WHEN** a model exists in the platform database with model_id, project_id, cluster_name, database_name, and model_name
- **AND** the corresponding table does not exist in the customer database at {cluster_name}.{database_name}.{model_name}
- **THEN** the detection service reports a `TABLE_MISSING` issue
- **AND** the issue includes the expected table name full path

#### Scenario: Detect missing fields

- **WHEN** a model's table exists in the customer database
- **AND** the model defines 5 fields: ["id", "name", "email", "status", "created_at"]
- **AND** the actual table only has 3 columns: ["id", "name", "created_at"]
- **THEN** the detection service reports a `FIELD_MISSING` issue for "email" and "status"
- **AND** the issue includes each missing field's expected type and constraints

#### Scenario: Detect field type mismatch

- **WHEN** a model's table exists in the customer database
- **AND** the model defines field "email" as type STRING with MaxLength=100
- **AND** the actual table column "email" is type VARCHAR(50)
- **THEN** the detection service reports a `FIELD_TYPE_MISMATCH` issue
- **AND** the issue includes both expected and actual type information

#### Scenario: Detect constraint mismatch

- **WHEN** a model defines field "name" as Required=true (NOT NULL)
- **AND** the actual table column "name" is nullable
- **THEN** the detection service reports a `FIELD_CONSTRAINT_MISMATCH` issue
- **AND** the issue specifies which constraint is mismatched

#### Scenario: Report no issues for healthy model

- **WHEN** a model's table exists with all defined columns matching types and constraints
- **AND** no system fields are missing
- **THEN** the detection service returns success with no issues
- **AND** the health status is marked as `HEALTHY`

#### Scenario: Ignore extra columns in dry run and additive modes

- **WHEN** a model defines 3 fields
- **AND** the actual table has 3 matching columns plus 2 extra columns
- **THEN** the detection service does NOT report issues for the extra columns in dry run and additive modes
- **AND** extra columns are only reported in issues when full sync mode is requested with deleteExtraFields=true

### Requirement: Model Repair Modes

The system SHALL support three distinct repair modes with progressively invasive actions.

#### Scenario: Dry run mode - detect only

- **WHEN** `repairModel` is called with `mode: DRY_RUN`
- **AND** the model has 2 missing columns and 1 type mismatch
- **THEN** the system detects all issues without making any database changes
- **AND** `changesApplied` is false
- **AND** all detected issues are returned in the result
- **AND** `executedDDL` array is empty

#### Scenario: Additive mode - add missing resources only

- **WHEN** `repairModel` is called with `mode: ADDITIVE`
- **AND** the table is missing
- **AND** the model defines 4 fields
- **THEN** the system executes CREATE TABLE DDL to recreate the table
- **AND** `changesApplied` is true
- **AND** `executedDDL` contains the CREATE TABLE statement
- **AND** extra columns (if any) are not removed

#### Scenario: Additive mode - add missing fields only

- **WHEN** `repairModel` is called with `mode: ADDITIVE`
- **AND** the table exists but is missing 2 fields: "phone" and "address"
- **THEN** the system executes ALTER TABLE ADD COLUMN for missing fields
- **AND** both fields are added in a single batch ALTER TABLE statement
- **AND** the ALTER uses `ALGORITHM=INPLACE, LOCK=NONE` for non-blocking operation
- **AND** existing columns are not modified

#### Scenario: Full sync mode - delete extra columns

- **WHEN** `repairModel` is called with `mode: FULL_SYNC` and `deleteExtraFields: true`
- **AND** the model has 3 fields defined
- **AND** the actual table has those 3 fields plus 2 extra columns
- **THEN** the system executes ALTER TABLE DROP COLUMN for extra columns
- **AND** system fields (id, createdAt, updatedAt) are never dropped
- **AND** the result includes `extraFieldsRemoved: 2` with their names

#### Scenario: Full sync mode with deleteExtraFields=false (additive)

- **WHEN** `repairModel` is called with `mode: FULL_SYNC` and `deleteExtraFields: false`
- **THEN** the behavior is identical to ADDITIVE mode
- **AND** no extra columns are removed

#### Scenario: Full sync mode prevents deletion of system fields

- **WHEN** `repairModel` is called with `mode: FULL_SYNC` and `deleteExtraFields: true`
- **AND** the actual table has extra columns that are system fields (id, createdAt, updatedAt)
- **THEN** system fields are never dropped
- **AND** the operation succeeds without attempting to drop system fields

#### Scenario: Additive mode with field type mismatch

- **WHEN** `repairModel` is called with `mode: ADDITIVE`
- **AND** the detection reveals a field type mismatch (expected VARCHAR(100), actual VARCHAR(50))
- **THEN** the mismatch is reported in issues
- **AND** no ALTER TABLE MODIFY COLUMN is executed
- **AND** the result includes the type mismatch in detected issues

### Requirement: Repair Result Payload

The system SHALL return comprehensive repair results including detected issues, executed operations, and state comparison.

#### Scenario: Return detected issues list

- **WHEN** `repairModel` returns a result
- **AND** detection found 3 issues: 1 TABLE_MISSING, 2 FIELD_MISSING
- **THEN** the result includes `detectedIssues` array with 3 items
- **AND** each issue has `type` (TABLE_MISSING, FIELD_MISSING, etc.)
- **AND** each issue has `description` with readable message
- **AND** each issue has `details` object with specific information

#### Scenario: Return executed DDL statements

- **WHEN** `repairModel` is called with mode that applies changes
- **AND** one CREATE TABLE and one ALTER TABLE were executed
- **THEN** the result includes `executedDDL` array with 2 items
- **AND** each DDL item includes the raw SQL statement
- **AND** if dry run, the array is empty

#### Scenario: Include before/after state comparison

- **WHEN** `repairModel` is called with dry run mode
- **AND** before state had 2 missing fields
- **THEN** the result includes `beforeState` showing issues present before operation
- **AND** the result includes `afterState` showing expected state after repair would be applied
- **AND** `healthStatusBefore` and `healthStatusAfter` indicate health state

#### Scenario: Return health status

- **WHEN** `repairModel` detects no issues
- **THEN** `healthStatus` is "HEALTHY"
- **WHEN** `repairModel` detects minor issues (missing fields only)
- **THEN** `healthStatus` is "NEEDS_REPAIR"
- **WHEN** `repairModel` detects major issues (missing table)
- **THEN** `healthStatus` is "BROKEN"

### Requirement: GraphQL API Definitions

The system SHALL define GraphQL types and mutations for model repair operations.

#### Scenario: Define RepairMode enum

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `RepairMode` enum with values: `DRY_RUN`, `ADDITIVE`, `FULL_SYNC`

#### Scenario: Define RepairModelInput

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `RepairModelInput` with fields:
  - `projectId: ID!` - Project identifier
  - `modelId: ID!` - Model identifier
  - `mode: RepairMode!` - Repair mode (DRY_RUN, ADDITIVE, FULL_SYNC)
  - `deleteExtraFields: Boolean` - Optional flag for full sync mode to delete extra fields (default: false)

#### Scenario: Define SchemaIssue type

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `SchemaIssue` type with fields:
  - `type: String!` - Issue type (TABLE_MISSING, FIELD_MISSING, FIELD_TYPE_MISMATCH, etc.)
  - `description: String!` - Human-readable description
  - `tableName: String` - Affected table name
  - `fieldName: String` - Affected field name (if applicable)
  - `details: JSON` - Additional details like expected vs actual values

#### Scenario: Define RepairModelPayload

- **WHEN** the GraphQL schema is defined
- **THEN** it includes `RepairModelPayload` type with fields:
  - `model: Model` - The repaired model
  - `changesApplied: Boolean!` - Whether any changes were actually made
  - `detectedIssues: [SchemaIssue!]!` - List of all detected issues
  - `executedDDL: [String!]!` - List of DDL statements executed
  - `healthStatusBefore: String!` - Health status before repair
  - `healthStatusAfter: String!` - Health status after repair
  - `extraFieldsRemoved: [String!]!` - Names of extra fields removed (empty if none)
  - `fieldsAdded: [String!]!` - Names of fields added (empty if none)

#### Scenario: Define repairModel mutation

- **WHEN** the GraphQL schema is defined
- **THEN** the Mutation type includes:
  - `repairModel(input: RepairModelInput!): RepairModelPayload!`

### Requirement: Error Handling and Safety

The system SHALL provide robust error handling and safety measures for repair operations.

#### Scenario: Reject non-existent model

- **WHEN** `repairModel` is called with a model_id that does not exist
- **THEN** the mutation fails with `ModelNotFound` error
- **AND** no database connection is made

#### Scenario: Validate cluster exists

- **WHEN** `repairModel` is called
- **AND** the model's cluster_name does not exist in the system
- **THEN** the mutation fails with `ClusterNotFound` error
- **AND** no further processing occurs

#### Scenario: Warn on missing database

- **WHEN** `repairModel` is called
- **AND** the cluster exists but the database_name does not exist in that cluster
- **THEN** the mutation returns a `DATABASE_MISSING` issue in detected issues
- **AND** the issue indicates which database is missing
- **AND** no changes are applied even if mode is not DRY_RUN

#### Scenario: Prevent deletion of referenced fields

- **WHEN** `repairModel` is called with `mode: FULL_SYNC` and `deleteExtraFields: true`
- **AND** a field to be removed is referenced by a model relation
- **THEN** the mutation fails with `OperationDenied` error
- **AND** the error message indicates which field cannot be dropped due to dependencies

#### Scenario: Idempotent repair operation

- **WHEN** `repairModel` is called on a healthy model
- **THEN** no database operations are executed
- **AND** `changesApplied` is false
- **AND** `detectedIssues` is empty
- **AND** `healthStatusBefore` and `healthStatusAfter` are both "HEALTHY"

#### Scenario: Transaction rollback on failure

- **WHEN** `repairModel` is executing in non-dry-run mode
- **AND** a database error occurs during DDL execution (e.g., permission denied)
- **THEN** the mutation fails with appropriate error
- **AND** partial changes are not persisted if possible
- **AND** the error message includes the DDL that failed

### Requirement: Integration with Existing Services

The system SHALL reuse existing infrastructure components for schema operations.

#### Scenario: Use existing SchemaIntrospector

- **WHEN** detecting actual database schema
- **THEN** the repair service uses the existing `SchemaIntrospector` component
- **AND** it queries `INFORMATION_SCHEMA.COLUMNS` to get column definitions
- **AND** no duplicate introspection logic is created

#### Scenario: Use existing DDL Converter

- **WHEN** generating CREATE TABLE DDL for repair
- **THEN** the repair service uses the existing `DDLConverter` and `MySQLDDLBuilder`
- **AND** the generated DDL format matches existing deployment DDL
- **AND** table comments include model description

#### Scenario: Use existing TypeMapper

- **WHEN** comparing model field types to database column types
- **THEN** the repair service uses the existing `TypeMapper` for conversion
- **AND** type comparison logic uses the same mapping rules as model creation

#### Scenario: Use existing DeploymentImpl

- **WHEN** executing DDL in repair
- **THEN** the repair service uses the existing `DeploymentImpl` execution methods
- **AND** database connection management follows existing patterns
- **AND** error handling matches existing deployment logic
