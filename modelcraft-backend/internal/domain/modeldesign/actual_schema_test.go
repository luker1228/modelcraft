package modeldesign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActualSchemaResult_StatusValues(t *testing.T) {
	tests := []struct {
		name   string
		status DbTableStatus
	}{
		{"table exists", DbTableExists},
		{"table missing", DbTableMissing},
		{"cluster unreachable", DbTableClusterUnreachable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ActualSchemaResult{Status: tt.status}
			assert.Equal(t, tt.status, result.Status)
		})
	}
}

func TestActualSchemaResult_FieldsOnlyPopulatedWhenTableExists(t *testing.T) {
	t.Run("TABLE_EXISTS has fields", func(t *testing.T) {
		result := &ActualSchemaResult{
			Status: DbTableExists,
			Fields: map[string]*DbColumnInfo{
				"name": {ColumnType: "VARCHAR", Constraints: []ActualConstraintType{ActualConstraintNotNull}},
			},
		}
		assert.NotNil(t, result.Fields)
		assert.Contains(t, result.Fields, "name")
	})

	t.Run("TABLE_MISSING has nil fields", func(t *testing.T) {
		result := &ActualSchemaResult{Status: DbTableMissing}
		assert.Nil(t, result.Fields)
	})

	t.Run("CLUSTER_UNREACHABLE has nil fields", func(t *testing.T) {
		result := &ActualSchemaResult{Status: DbTableClusterUnreachable}
		assert.Nil(t, result.Fields)
	})
}

func TestDbColumnInfo_ConstraintCombinations(t *testing.T) {
	tests := []struct {
		name        string
		constraints []ActualConstraintType
		hasUnique   bool
		hasNotNull  bool
	}{
		{
			name:        "no constraints",
			constraints: []ActualConstraintType{},
			hasUnique:   false,
			hasNotNull:  false,
		},
		{
			name:        "UNIQUE only",
			constraints: []ActualConstraintType{ActualConstraintUnique},
			hasUnique:   true,
			hasNotNull:  false,
		},
		{
			name:        "NOT_NULL only",
			constraints: []ActualConstraintType{ActualConstraintNotNull},
			hasUnique:   false,
			hasNotNull:  true,
		},
		{
			name:        "UNIQUE and NOT_NULL",
			constraints: []ActualConstraintType{ActualConstraintUnique, ActualConstraintNotNull},
			hasUnique:   true,
			hasNotNull:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &DbColumnInfo{Constraints: tt.constraints}

			hasUnique := false
			hasNotNull := false
			for _, c := range col.Constraints {
				if c == ActualConstraintUnique {
					hasUnique = true
				}
				if c == ActualConstraintNotNull {
					hasNotNull = true
				}
			}
			assert.Equal(t, tt.hasUnique, hasUnique)
			assert.Equal(t, tt.hasNotNull, hasNotNull)
		})
	}
}

func TestDbColumnInfo_ForeignKey(t *testing.T) {
	t.Run("with foreign key", func(t *testing.T) {
		col := &DbColumnInfo{
			ColumnType:  "VARCHAR",
			Constraints: []ActualConstraintType{},
			ForeignKey: &ActualForeignKey{
				ReferencedTable:  "organization",
				ReferencedColumn: "id",
				ConstraintName:   "fk_user_org_id",
			},
			Conflicts: []FieldConflict{},
		}
		assert.NotNil(t, col.ForeignKey)
		assert.Equal(t, "organization", col.ForeignKey.ReferencedTable)
		assert.Equal(t, "id", col.ForeignKey.ReferencedColumn)
		assert.Equal(t, "fk_user_org_id", col.ForeignKey.ConstraintName)
	})

	t.Run("without foreign key", func(t *testing.T) {
		col := &DbColumnInfo{
			ColumnType:  "VARCHAR",
			Constraints: []ActualConstraintType{},
			ForeignKey:  nil,
			Conflicts:   []FieldConflict{},
		}
		assert.Nil(t, col.ForeignKey)
	})
}

func TestFieldConflict_AspectValues(t *testing.T) {
	t.Run("UNIQUE_MISMATCH conflict", func(t *testing.T) {
		conflict := FieldConflict{
			Aspect:   FieldConflictUniqueMismatch,
			Expected: "true",
			Actual:   "false",
		}
		assert.Equal(t, FieldConflictUniqueMismatch, conflict.Aspect)
		assert.Equal(t, "true", conflict.Expected)
		assert.Equal(t, "false", conflict.Actual)
	})

	t.Run("NOT_NULL_MISMATCH conflict", func(t *testing.T) {
		conflict := FieldConflict{
			Aspect:   FieldConflictNotNullMismatch,
			Expected: "false",
			Actual:   "true",
		}
		assert.Equal(t, FieldConflictNotNullMismatch, conflict.Aspect)
		assert.Equal(t, "false", conflict.Expected)
		assert.Equal(t, "true", conflict.Actual)
	})
}

func TestFieldDefinition_IsEnumLabelField_VirtualField(t *testing.T) {
	t.Run("ENUM_LABEL is virtual", func(t *testing.T) {
		field := &FieldDefinition{
			Type: &FieldType{Format: FormatEnumLabel},
		}
		assert.True(t, field.IsEnumLabelField())
	})

	t.Run("STRING is not virtual", func(t *testing.T) {
		field := &FieldDefinition{
			Type: &FieldType{Format: FormatString},
		}
		assert.False(t, field.IsEnumLabelField())
	})

	t.Run("RELATION is not virtual", func(t *testing.T) {
		field := &FieldDefinition{
			Type: &FieldType{Format: FormatRelation},
		}
		assert.False(t, field.IsEnumLabelField())
	})
}
