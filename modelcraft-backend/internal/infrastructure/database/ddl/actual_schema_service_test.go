package ddl

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/modeldesign"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActualSchemaServiceImpl_QueryActualSchema_TableExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	colCols := []string{
		"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT",
	}
	colRows := sqlmock.NewRows(colCols).
		AddRow("id", "bigint", sql.NullInt64{}, "NO", "PRI", nil).
		AddRow("email", "varchar", sql.NullInt64{Int64: 256, Valid: true}, "NO", "UNI", nil).
		AddRow("name", "varchar", sql.NullInt64{}, "YES", "", nil)

	fkRows := sqlmock.NewRows([]string{
		"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME",
	})

	mock.ExpectQuery("INFORMATION_SCHEMA.COLUMNS").
		WithArgs("testdb", "users").
		WillReturnRows(colRows)
	mock.ExpectQuery("INFORMATION_SCHEMA.KEY_COLUMN_USAGE").
		WithArgs("testdb", "users").
		WillReturnRows(fkRows)

	svc := NewActualSchemaService()
	fields := []*modeldesign.FieldDefinition{
		{Name: "id", IsUnique: false, NonNull: true},
		{Name: "email", IsUnique: true, NonNull: true},
		{Name: "name", IsUnique: false, NonNull: false},
	}

	result, err := svc.QueryActualSchema(context.Background(), db, "testdb", "users", fields)
	require.NoError(t, err)
	assert.Equal(t, modeldesign.DbTableExists, result.Status)
	assert.Len(t, result.Fields, 3)

	// email: UNI + NOT NULL, no conflicts
	emailCol := result.Fields["email"]
	require.NotNil(t, emailCol)
	assert.Equal(t, "VARCHAR", emailCol.ColumnType)
	require.NotNil(t, emailCol.ColumnLength)
	assert.Equal(t, int64(256), *emailCol.ColumnLength)
	assert.Contains(t, emailCol.Constraints, modeldesign.ActualConstraintUnique)
	assert.Contains(t, emailCol.Constraints, modeldesign.ActualConstraintNotNull)
	assert.Empty(t, emailCol.Conflicts)

	// name: nullable, no unique
	nameCol := result.Fields["name"]
	require.NotNil(t, nameCol)
	assert.Empty(t, nameCol.Constraints)
	assert.Empty(t, nameCol.Conflicts)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActualSchemaServiceImpl_QueryActualSchema_TableMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	colCols := []string{
		"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT",
	}
	mock.ExpectQuery("INFORMATION_SCHEMA.COLUMNS").
		WithArgs("testdb", "missing_table").
		WillReturnRows(sqlmock.NewRows(colCols))

	svc := NewActualSchemaService()
	result, err := svc.QueryActualSchema(context.Background(), db, "testdb", "missing_table", nil)

	require.NoError(t, err)
	assert.Equal(t, modeldesign.DbTableMissing, result.Status)
	assert.Nil(t, result.Fields)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActualSchemaServiceImpl_QueryActualSchema_WithForeignKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	colCols := []string{
		"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT",
	}
	colRows := sqlmock.NewRows(colCols).
		AddRow("org_id", "bigint", sql.NullInt64{}, "NO", "", nil)

	fkRows := sqlmock.NewRows([]string{
		"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME",
	}).AddRow("org_id", "organization", "id", "fk_users_org_id")

	mock.ExpectQuery("INFORMATION_SCHEMA.COLUMNS").
		WithArgs("testdb", "users").
		WillReturnRows(colRows)
	mock.ExpectQuery("INFORMATION_SCHEMA.KEY_COLUMN_USAGE").
		WithArgs("testdb", "users").
		WillReturnRows(fkRows)

	svc := NewActualSchemaService()
	result, err := svc.QueryActualSchema(context.Background(), db, "testdb", "users", nil)

	require.NoError(t, err)
	assert.Equal(t, modeldesign.DbTableExists, result.Status)

	colInfo := result.Fields["org_id"]
	require.NotNil(t, colInfo)
	require.NotNil(t, colInfo.ForeignKey)
	assert.Equal(t, "organization", colInfo.ForeignKey.ReferencedTable)
	assert.Equal(t, "id", colInfo.ForeignKey.ReferencedColumn)
	assert.Equal(t, "fk_users_org_id", colInfo.ForeignKey.ConstraintName)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActualSchemaServiceImpl_QueryActualSchema_UniqueMismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	// Design says isUnique=true, but actual COLUMN_KEY is '' (not UNI)
	colCols := []string{
		"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT",
	}
	colRows := sqlmock.NewRows(colCols).
		AddRow("email", "varchar", sql.NullInt64{}, "YES", "", nil)

	mock.ExpectQuery("INFORMATION_SCHEMA.COLUMNS").
		WithArgs("testdb", "users").
		WillReturnRows(colRows)
	mock.ExpectQuery("INFORMATION_SCHEMA.KEY_COLUMN_USAGE").
		WithArgs("testdb", "users").
		WillReturnRows(sqlmock.NewRows([]string{
			"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME",
		}))

	svc := NewActualSchemaService()
	fields := []*modeldesign.FieldDefinition{
		{Name: "email", IsUnique: true, NonNull: false},
	}

	result, err := svc.QueryActualSchema(context.Background(), db, "testdb", "users", fields)
	require.NoError(t, err)

	colInfo := result.Fields["email"]
	require.NotNil(t, colInfo)
	require.Len(t, colInfo.Conflicts, 1)
	assert.Equal(t, modeldesign.FieldConflictUniqueMismatch, colInfo.Conflicts[0].Aspect)
	assert.Equal(t, "true", colInfo.Conflicts[0].Expected)
	assert.Equal(t, "false", colInfo.Conflicts[0].Actual)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActualSchemaServiceImpl_QueryActualSchema_NotNullMismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	// Design says nonNull=false, but actual IS_NULLABLE='NO'
	colCols := []string{
		"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT",
	}
	colRows := sqlmock.NewRows(colCols).
		AddRow("name", "varchar", sql.NullInt64{}, "NO", "", nil)

	mock.ExpectQuery("INFORMATION_SCHEMA.COLUMNS").
		WithArgs("testdb", "users").
		WillReturnRows(colRows)
	mock.ExpectQuery("INFORMATION_SCHEMA.KEY_COLUMN_USAGE").
		WithArgs("testdb", "users").
		WillReturnRows(sqlmock.NewRows([]string{
			"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME",
		}))

	svc := NewActualSchemaService()
	fields := []*modeldesign.FieldDefinition{
		{Name: "name", IsUnique: false, NonNull: false},
	}

	result, err := svc.QueryActualSchema(context.Background(), db, "testdb", "users", fields)
	require.NoError(t, err)

	colInfo := result.Fields["name"]
	require.Len(t, colInfo.Conflicts, 1)
	assert.Equal(t, modeldesign.FieldConflictNotNullMismatch, colInfo.Conflicts[0].Aspect)
	assert.Equal(t, "false", colInfo.Conflicts[0].Expected)
	assert.Equal(t, "true", colInfo.Conflicts[0].Actual)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActualSchemaServiceImpl_QueryActualSchema_BothConflicts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	// Design: isUnique=false, nonNull=true; Actual: UNI, IS_NULLABLE='YES'
	colCols := []string{
		"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT",
	}
	colRows := sqlmock.NewRows(colCols).
		AddRow("code", "varchar", sql.NullInt64{}, "YES", "UNI", nil)

	mock.ExpectQuery("INFORMATION_SCHEMA.COLUMNS").
		WithArgs("testdb", "items").
		WillReturnRows(colRows)
	mock.ExpectQuery("INFORMATION_SCHEMA.KEY_COLUMN_USAGE").
		WithArgs("testdb", "items").
		WillReturnRows(sqlmock.NewRows([]string{
			"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME",
		}))

	svc := NewActualSchemaService()
	fields := []*modeldesign.FieldDefinition{
		{Name: "code", IsUnique: false, NonNull: true},
	}

	result, err := svc.QueryActualSchema(context.Background(), db, "testdb", "items", fields)
	require.NoError(t, err)

	colInfo := result.Fields["code"]
	require.Len(t, colInfo.Conflicts, 2)

	aspects := make(map[modeldesign.FieldConflictAspect]bool)
	for _, c := range colInfo.Conflicts {
		aspects[c.Aspect] = true
	}
	assert.True(t, aspects[modeldesign.FieldConflictUniqueMismatch])
	assert.True(t, aspects[modeldesign.FieldConflictNotNullMismatch])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActualSchemaServiceImpl_QueryActualSchema_NoConflictsWhenFieldNotInDesign(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	// Column exists in DB but not in design-time fields slice
	colCols := []string{
		"COLUMN_NAME", "DATA_TYPE", "CHARACTER_MAXIMUM_LENGTH", "IS_NULLABLE", "COLUMN_KEY", "COLUMN_DEFAULT",
	}
	colRows := sqlmock.NewRows(colCols).
		AddRow("extra_col", "varchar", sql.NullInt64{}, "NO", "UNI", nil)

	mock.ExpectQuery("INFORMATION_SCHEMA.COLUMNS").
		WithArgs("testdb", "users").
		WillReturnRows(colRows)
	mock.ExpectQuery("INFORMATION_SCHEMA.KEY_COLUMN_USAGE").
		WithArgs("testdb", "users").
		WillReturnRows(sqlmock.NewRows([]string{
			"COLUMN_NAME", "REFERENCED_TABLE_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME",
		}))

	svc := NewActualSchemaService()
	// Pass empty fields list — no design-time definition for extra_col
	result, err := svc.QueryActualSchema(context.Background(), db, "testdb", "users", nil)
	require.NoError(t, err)

	colInfo := result.Fields["extra_col"]
	require.NotNil(t, colInfo)
	assert.Empty(t, colInfo.Conflicts)

	require.NoError(t, mock.ExpectationsWereMet())
}
