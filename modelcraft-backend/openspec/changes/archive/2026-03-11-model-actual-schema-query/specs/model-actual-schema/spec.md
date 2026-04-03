## ADDED Requirements

### Requirement: model query supports withActualSchema parameter
`model(id)` query SHALL accept an optional `withActualSchema: Boolean` parameter (default `false`).
When `withActualSchema` is `false` or omitted, the response SHALL be identical to current behavior — `Model.dbTable` and all `Field.dbColumn` SHALL be null.
When `withActualSchema` is `true`, `Model.dbTable` SHALL reflect the actual table status, and each non-virtual `Field` SHALL have `dbColumn` populated from the actual database.

#### Scenario: Query without withActualSchema returns normal response
- **WHEN** client calls `model(projectSlug, id)` without `withActualSchema`
- **THEN** `Model.dbTable` is null and all `Field.dbColumn` are null

#### Scenario: Query with withActualSchema=false returns normal response
- **WHEN** client calls `model(projectSlug, id, withActualSchema: false)`
- **THEN** `Model.dbTable` is null and all `Field.dbColumn` are null

#### Scenario: Query with withActualSchema=true returns db column info
- **WHEN** client calls `model(projectSlug, id, withActualSchema: true)` and the table exists
- **THEN** `Model.dbTable` is `TABLE_EXISTS` and each non-virtual field has `dbColumn` populated

---

### Requirement: Model.dbTable reflects actual table status
When `withActualSchema` is `true`, `Model.dbTable` SHALL be set to one of:
- `TABLE_EXISTS` — table exists and columns were queried successfully
- `TABLE_MISSING` — connected to the database but table does not exist
- `CLUSTER_UNREACHABLE` — could not connect to the cluster or database

When `dbTable` is `TABLE_MISSING` or `CLUSTER_UNREACHABLE`, all `Field.dbColumn` SHALL be null and no error SHALL be returned — the status itself communicates the failure.

#### Scenario: Table exists sets dbTable to TABLE_EXISTS
- **WHEN** `withActualSchema` is `true` and the model's table exists in the actual database
- **THEN** `Model.dbTable` is `TABLE_EXISTS`

#### Scenario: Table missing sets dbTable to TABLE_MISSING
- **WHEN** `withActualSchema` is `true` and the model's table does not exist in the actual database
- **THEN** `Model.dbTable` is `TABLE_MISSING` and all `Field.dbColumn` are null

#### Scenario: Cluster unreachable sets dbTable to CLUSTER_UNREACHABLE
- **WHEN** `withActualSchema` is `true` and the cluster or database cannot be reached
- **THEN** `Model.dbTable` is `CLUSTER_UNREACHABLE` and all `Field.dbColumn` are null

---

### Requirement: DbColumnInfo contains column type, constraints, and foreign key
When `withActualSchema` is `true` and `Model.dbTable` is `TABLE_EXISTS`, each non-virtual `Field.dbColumn` SHALL contain:
- `columnType: String!` — the actual MySQL column type (e.g., `VARCHAR`, `BIGINT`, `TEXT`)
- `constraints: [ActualConstraintType!]!` — active constraints: `UNIQUE` and/or `NOT_NULL`
- `foreignKey: ActualForeignKey` — the FK constraint on this column, or `null` if none
- `conflicts: [FieldConflict!]!` — mismatches between design-time definition and actual DB state

`ActualConstraintType` SHALL be an enum with values `UNIQUE` and `NOT_NULL`.

`ActualForeignKey` SHALL contain:
- `referencedTable: String!` — the table being referenced
- `referencedColumn: String!` — the column being referenced
- `constraintName: String!` — the FK constraint name (enables grouping composite FK columns)

#### Scenario: Field with UNIQUE index shows UNIQUE in constraints
- **WHEN** a column has a UNIQUE index in the actual database
- **THEN** `Field.dbColumn.constraints` contains `UNIQUE`

#### Scenario: Field with NOT NULL column shows NOT_NULL in constraints
- **WHEN** a column is NOT NULL in the actual database
- **THEN** `Field.dbColumn.constraints` contains `NOT_NULL`

#### Scenario: Field with nullable column and no unique index returns empty constraints
- **WHEN** a column has no UNIQUE index and is nullable
- **THEN** `Field.dbColumn.constraints` is an empty array

#### Scenario: Field with foreign key returns foreignKey info
- **WHEN** a column has a FOREIGN KEY constraint in the actual database
- **THEN** `Field.dbColumn.foreignKey` is non-null with `referencedTable`, `referencedColumn`, and `constraintName`

#### Scenario: Field without foreign key returns null foreignKey
- **WHEN** a column has no FOREIGN KEY constraint
- **THEN** `Field.dbColumn.foreignKey` is null

---

### Requirement: Virtual fields are skipped in dbColumn
Fields with `format == ENUM_LABEL` are virtual fields with no corresponding database column. Their `dbColumn` SHALL always be null regardless of `withActualSchema`.

#### Scenario: Virtual field always has null dbColumn
- **WHEN** `withActualSchema` is `true` and a field has `format == ENUM_LABEL`
- **THEN** `Field.dbColumn` is null

---

### Requirement: conflicts reflect mismatches between design and actual DB
When `withActualSchema` is `true` and `Model.dbTable` is `TABLE_EXISTS`, the backend SHALL compute `Field.dbColumn.conflicts` by comparing the design-time field definition against the actual column state.

`FieldConflict` SHALL contain:
- `aspect: FieldConflictAspect!` — the type of mismatch (`UNIQUE_MISMATCH` or `NOT_NULL_MISMATCH`)
- `expected: String!` — the design-time value (e.g., `"true"`)
- `actual: String!` — the actual DB value (e.g., `"false"`)

A `UNIQUE_MISMATCH` conflict SHALL be reported when `Field.isUnique` does not match whether `UNIQUE` is present in `Field.dbColumn.constraints`.
A `NOT_NULL_MISMATCH` conflict SHALL be reported when `Field.nonNull` does not match whether `NOT_NULL` is present in `Field.dbColumn.constraints`.

#### Scenario: isUnique=true but no UNIQUE constraint in DB produces conflict
- **WHEN** `Field.isUnique` is `true` but the actual column has no UNIQUE index
- **THEN** `Field.dbColumn.conflicts` contains a `UNIQUE_MISMATCH` entry with `expected="true"` and `actual="false"`

#### Scenario: isUnique=false but UNIQUE constraint exists in DB produces conflict
- **WHEN** `Field.isUnique` is `false` but the actual column has a UNIQUE index
- **THEN** `Field.dbColumn.conflicts` contains a `UNIQUE_MISMATCH` entry with `expected="false"` and `actual="true"`

#### Scenario: nonNull=true but column is nullable in DB produces conflict
- **WHEN** `Field.nonNull` is `true` but the actual column is nullable
- **THEN** `Field.dbColumn.conflicts` contains a `NOT_NULL_MISMATCH` entry with `expected="true"` and `actual="false"`

#### Scenario: nonNull=false but column is NOT NULL in DB produces conflict
- **WHEN** `Field.nonNull` is `false` but the actual column is NOT NULL
- **THEN** `Field.dbColumn.conflicts` contains a `NOT_NULL_MISMATCH` entry with `expected="false"` and `actual="true"`

#### Scenario: No mismatch returns empty conflicts
- **WHEN** `Field.isUnique` and `Field.nonNull` both match the actual column constraints
- **THEN** `Field.dbColumn.conflicts` is an empty array
