package modeldesign

import (
	"context"
	"database/sql"
)

// DbTableStatus represents the status of a model's actual database table.
type DbTableStatus string

const (
	// DbTableExists indicates the table exists and columns were queried successfully.
	DbTableExists DbTableStatus = "TABLE_EXISTS"
	// DbTableMissing indicates connected to the database but the table does not exist.
	DbTableMissing DbTableStatus = "TABLE_MISSING"
	// DbTableClusterUnreachable indicates the cluster or database could not be reached.
	DbTableClusterUnreachable DbTableStatus = "CLUSTER_UNREACHABLE"
)

// ActualConstraintType represents a constraint type found in the actual database column.
type ActualConstraintType string

const (
	// ActualConstraintUnique indicates a UNIQUE constraint on the column.
	ActualConstraintUnique ActualConstraintType = "UNIQUE"
	// ActualConstraintNotNull indicates a NOT NULL constraint on the column.
	ActualConstraintNotNull ActualConstraintType = "NOT_NULL"
)

// FieldConflictAspect represents the type of mismatch between design-time and actual DB.
type FieldConflictAspect string

const (
	// FieldConflictUniqueMismatch indicates Field.IsUnique does not match the actual UNIQUE constraint.
	FieldConflictUniqueMismatch FieldConflictAspect = "UNIQUE_MISMATCH"
	// FieldConflictNotNullMismatch indicates Field.NonNull does not match the actual NOT NULL constraint.
	FieldConflictNotNullMismatch FieldConflictAspect = "NOT_NULL_MISMATCH"
	// FieldConflictPrimaryMismatch indicates Field.IsPrimary does not match the actual PRIMARY KEY constraint.
	FieldConflictPrimaryMismatch FieldConflictAspect = "PRIMARY_MISMATCH"
)

// ActualForeignKey represents a foreign key constraint found in the actual database.
type ActualForeignKey struct {
	// ReferencedTable is the table being referenced.
	ReferencedTable string
	// ReferencedColumn is the column being referenced.
	ReferencedColumn string
	// ConstraintName is the FK constraint name, useful for identifying composite FK columns.
	ConstraintName string
}

// FieldConflict represents a mismatch between design-time field definition and actual DB state.
type FieldConflict struct {
	// Aspect is the type of mismatch.
	Aspect FieldConflictAspect
	// Expected is the design-time value (e.g., "true").
	Expected string
	// Actual is the actual DB value (e.g., "false").
	Actual string
}

// DbColumnInfo contains actual database column information for a field.
type DbColumnInfo struct {
	// ColumnType is the actual MySQL column type including length if applicable (e.g., "VARCHAR(256)", "BIGINT").
	ColumnType string
	// ColumnLength is the maximum character length for string types, or nil for other types.
	ColumnLength *int64
	// Unique indicates whether the column has a UNIQUE constraint.
	Unique bool
	// NonNull indicates whether the column has a NOT NULL constraint.
	NonNull bool
	// IsPrimaryKey indicates whether the column is a PRIMARY KEY.
	IsPrimaryKey bool
	// DefaultValue is the default value of the column, or nil if none.
	DefaultValue *string
	// Constraints lists active constraints on the column.
	Constraints []ActualConstraintType
	// ForeignKey is the FK constraint on this column, or nil if none.
	ForeignKey *ActualForeignKey
	// Conflicts lists mismatches between design-time definition and actual DB state.
	Conflicts []FieldConflict
}

// ActualSchemaResult holds the result of querying a model's actual database schema.
type ActualSchemaResult struct {
	// Status is the table-level query status.
	Status DbTableStatus
	// Fields maps field name to actual column info; only populated when Status is DbTableExists.
	Fields map[string]*DbColumnInfo
}

// ActualSchemaService queries the actual database schema for a model's table.
type ActualSchemaService interface {
	// QueryActualSchema queries the actual schema for the given table, comparing against
	// the provided field definitions to compute conflicts.
	// Returns TABLE_MISSING when the table does not exist (not an error).
	QueryActualSchema(
		ctx context.Context,
		db *sql.DB,
		databaseName string,
		tableName string,
		fields []*FieldDefinition,
	) (*ActualSchemaResult, error)
}
