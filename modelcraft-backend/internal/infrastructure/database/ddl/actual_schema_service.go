// Package ddl provides DDL-level database operations including schema introspection and actual schema querying.
package ddl

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizerrors"
	"strings"
)

// Column key constants from INFORMATION_SCHEMA.
const (
	columnKeyPrimary = "PRI"
	columnKeyUnique  = "UNI"
)

// actualColumnRow represents a row from the INFORMATION_SCHEMA.COLUMNS query.
type actualColumnRow struct {
	ColumnName         string
	DataType           string
	CharacterMaxLength sql.NullInt64
	IsNullable         string
	ColumnKey          string
	ColumnDefault      sql.NullString
}

// actualForeignKeyRow represents a row from the foreign key query.
type actualForeignKeyRow struct {
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
	ConstraintName   string
}

// ActualSchemaServiceImpl implements modeldesign.ActualSchemaService using raw SQL against INFORMATION_SCHEMA.
type ActualSchemaServiceImpl struct{}

// NewActualSchemaService creates a new ActualSchemaServiceImpl.
func NewActualSchemaService() modeldesign.ActualSchemaService {
	return &ActualSchemaServiceImpl{}
}

// QueryActualSchema queries the actual database schema for the given table, computes conflicts
// against the provided field definitions, and returns an ActualSchemaResult.
// Returns TABLE_MISSING status (not an error) when the table does not exist.
func (s *ActualSchemaServiceImpl) QueryActualSchema(
	ctx context.Context,
	db *sql.DB,
	databaseName string,
	tableName string,
	fields []*modeldesign.FieldDefinition,
) (*modeldesign.ActualSchemaResult, error) {
	columns, err := s.queryColumns(ctx, db, databaseName, tableName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "query columns for %s.%s", databaseName, tableName)
	}

	// Table does not exist when INFORMATION_SCHEMA returns no columns.
	if len(columns) == 0 {
		return &modeldesign.ActualSchemaResult{Status: modeldesign.DbTableMissing}, nil
	}

	foreignKeys, err := s.queryForeignKeys(ctx, db, databaseName, tableName)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "query foreign keys for %s.%s", databaseName, tableName)
	}

	// Build field definition index for conflict detection.
	fieldIndex := make(map[string]*modeldesign.FieldDefinition, len(fields))
	for _, f := range fields {
		fieldIndex[f.Name] = f
	}

	// Assemble DbColumnInfo for each actual column.
	columnInfoMap := make(map[string]*modeldesign.DbColumnInfo, len(columns))
	for _, col := range columns {
		constraints := s.buildConstraints(col)
		conflicts := s.computeConflicts(col, fieldIndex[col.ColumnName])

		// Extract boolean flags from INFORMATION_SCHEMA data
		isPrimaryKey := col.ColumnKey == columnKeyPrimary
		unique := col.ColumnKey == columnKeyUnique
		nonNull := col.IsNullable == "NO"

		// Extract default value
		var defaultValue *string
		if col.ColumnDefault.Valid {
			defaultValue = &col.ColumnDefault.String
		}

		// Build column length if applicable
		columnType := strings.ToUpper(col.DataType)
		var columnLength *int64
		if col.CharacterMaxLength.Valid {
			columnLength = &col.CharacterMaxLength.Int64
		}

		colInfo := &modeldesign.DbColumnInfo{
			ColumnType:   columnType,
			ColumnLength: columnLength,
			Unique:       unique,
			NonNull:      nonNull,
			IsPrimaryKey: isPrimaryKey,
			DefaultValue: defaultValue,
			Constraints:  constraints,
			ForeignKey:   foreignKeys[col.ColumnName],
			Conflicts:    conflicts,
		}
		columnInfoMap[col.ColumnName] = colInfo
	}

	return &modeldesign.ActualSchemaResult{
		Status: modeldesign.DbTableExists,
		Fields: columnInfoMap,
	}, nil
}

// queryColumns queries INFORMATION_SCHEMA.COLUMNS for the given table.
// Returns an empty slice when the table does not exist.
func (s *ActualSchemaServiceImpl) queryColumns(
	ctx context.Context,
	db *sql.DB,
	databaseName, tableName string,
) ([]actualColumnRow, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.QueryContext(ctx, query, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("INFORMATION_SCHEMA.COLUMNS query failed: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var result []actualColumnRow
	for rows.Next() {
		var row actualColumnRow
		err := rows.Scan(
			&row.ColumnName,
			&row.DataType,
			&row.CharacterMaxLength,
			&row.IsNullable,
			&row.ColumnKey,
			&row.ColumnDefault,
		)
		if err != nil {
			return nil, fmt.Errorf("scan column row: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate column rows: %w", err)
	}
	return result, nil
}

// queryForeignKeys queries KEY_COLUMN_USAGE joined with TABLE_CONSTRAINTS to find FK constraints
// for the given table. Returns a map keyed by column name.
func (s *ActualSchemaServiceImpl) queryForeignKeys(
	ctx context.Context,
	db *sql.DB,
	databaseName, tableName string,
) (map[string]*modeldesign.ActualForeignKey, error) {
	query := `
		SELECT
			kcu.COLUMN_NAME,
			kcu.REFERENCED_TABLE_NAME,
			kcu.REFERENCED_COLUMN_NAME,
			kcu.CONSTRAINT_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			ON  tc.CONSTRAINT_NAME  = kcu.CONSTRAINT_NAME
			AND tc.TABLE_SCHEMA     = kcu.TABLE_SCHEMA
			AND tc.TABLE_NAME       = kcu.TABLE_NAME
		WHERE kcu.TABLE_SCHEMA = ?
		  AND kcu.TABLE_NAME   = ?
		  AND tc.CONSTRAINT_TYPE = 'FOREIGN KEY'
	`

	rows, err := db.QueryContext(ctx, query, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("INFORMATION_SCHEMA foreign key query failed: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	result := make(map[string]*modeldesign.ActualForeignKey)
	for rows.Next() {
		var row actualForeignKeyRow
		if err := rows.Scan(
			&row.ColumnName,
			&row.ReferencedTable,
			&row.ReferencedColumn,
			&row.ConstraintName,
		); err != nil {
			return nil, fmt.Errorf("scan foreign key row: %w", err)
		}
		result[row.ColumnName] = &modeldesign.ActualForeignKey{
			ReferencedTable:  row.ReferencedTable,
			ReferencedColumn: row.ReferencedColumn,
			ConstraintName:   row.ConstraintName,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate foreign key rows: %w", err)
	}
	return result, nil
}

// buildConstraints converts INFORMATION_SCHEMA column data into ActualConstraintType slice.
// COLUMN_KEY='UNI' indicates a single-column UNIQUE constraint.
// IS_NULLABLE='NO' indicates a NOT NULL constraint.
func (s *ActualSchemaServiceImpl) buildConstraints(col actualColumnRow) []modeldesign.ActualConstraintType {
	var constraints []modeldesign.ActualConstraintType
	if col.ColumnKey == columnKeyUnique {
		constraints = append(constraints, modeldesign.ActualConstraintUnique)
	}
	if col.IsNullable == "NO" {
		constraints = append(constraints, modeldesign.ActualConstraintNotNull)
	}
	return constraints
}

// computeConflicts compares the actual column state against the design-time field definition.
// Returns nil when field is nil (no design-time definition to compare against).
func (s *ActualSchemaServiceImpl) computeConflicts(
	col actualColumnRow,
	field *modeldesign.FieldDefinition,
) []modeldesign.FieldConflict {
	if field == nil {
		return []modeldesign.FieldConflict{}
	}

	actualPrimaryKey := col.ColumnKey == columnKeyPrimary
	actualUnique := col.ColumnKey == columnKeyUnique
	actualNotNull := col.IsNullable == "NO"

	var conflicts []modeldesign.FieldConflict

	if field.IsPrimary != actualPrimaryKey {
		conflicts = append(conflicts, modeldesign.FieldConflict{
			Aspect:   modeldesign.FieldConflictPrimaryMismatch,
			Expected: boolToString(field.IsPrimary),
			Actual:   boolToString(actualPrimaryKey),
		})
	}

	if field.IsUnique != actualUnique {
		conflicts = append(conflicts, modeldesign.FieldConflict{
			Aspect:   modeldesign.FieldConflictUniqueMismatch,
			Expected: boolToString(field.IsUnique),
			Actual:   boolToString(actualUnique),
		})
	}

	if field.NonNull != actualNotNull {
		conflicts = append(conflicts, modeldesign.FieldConflict{
			Aspect:   modeldesign.FieldConflictNotNullMismatch,
			Expected: boolToString(field.NonNull),
			Actual:   boolToString(actualNotNull),
		})
	}

	if conflicts == nil {
		return []modeldesign.FieldConflict{}
	}
	return conflicts
}

// boolToString converts a bool to its string representation.
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
