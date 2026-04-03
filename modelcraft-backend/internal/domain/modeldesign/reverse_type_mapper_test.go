package modeldesign

import (
	"modelcraft/internal/domain/project"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseTypeMapper_MapColumnToFieldType(t *testing.T) {
	mapper := NewReverseTypeMapper()

	tests := []struct {
		name            string
		col             TableColumn
		expectedFormat  FormatType
		expectedSupport bool
		expectedSkip    string
	}{
		// Integer types
		{
			name:            "TINYINT(1) should map to BOOLEAN",
			col:             TableColumn{DataType: "tinyint", Length: 1},
			expectedFormat:  FormatBoolean,
			expectedSupport: true,
		},
		{
			name:            "TINYINT should map to INTEGER",
			col:             TableColumn{DataType: "tinyint", Length: 4},
			expectedFormat:  FormatInteger,
			expectedSupport: true,
		},
		{
			name:            "SMALLINT should map to INTEGER",
			col:             TableColumn{DataType: "smallint"},
			expectedFormat:  FormatInteger,
			expectedSupport: true,
		},
		{
			name:            "INT should map to INTEGER",
			col:             TableColumn{DataType: "int"},
			expectedFormat:  FormatInteger,
			expectedSupport: true,
		},
		{
			name:            "BIGINT should map to INTEGER",
			col:             TableColumn{DataType: "bigint"},
			expectedFormat:  FormatInteger,
			expectedSupport: true,
		},

		// Float types
		{
			name:            "FLOAT should map to NUMBER",
			col:             TableColumn{DataType: "float"},
			expectedFormat:  FormatNumber,
			expectedSupport: true,
		},
		{
			name:            "DOUBLE should map to NUMBER",
			col:             TableColumn{DataType: "double"},
			expectedFormat:  FormatNumber,
			expectedSupport: true,
		},

		// Decimal type
		{
			name:            "DECIMAL should map to DECIMAL with precision",
			col:             TableColumn{DataType: "decimal", Precision: 10, Scale: 2},
			expectedFormat:  FormatDecimal,
			expectedSupport: true,
		},

		// String types
		{
			name:            "CHAR should map to STRING with maxLength",
			col:             TableColumn{DataType: "char", Length: 10},
			expectedFormat:  FormatString,
			expectedSupport: true,
		},
		{
			name:            "VARCHAR should map to STRING with maxLength",
			col:             TableColumn{DataType: "varchar", Length: 255},
			expectedFormat:  FormatString,
			expectedSupport: true,
		},
		{
			name:            "TEXT should map to STRING",
			col:             TableColumn{DataType: "text"},
			expectedFormat:  FormatString,
			expectedSupport: true,
		},

		// Date/Time types
		{
			name:            "DATE should map to DATE",
			col:             TableColumn{DataType: "date"},
			expectedFormat:  FormatDate,
			expectedSupport: true,
		},
		{
			name:            "DATETIME should map to DATETIME",
			col:             TableColumn{DataType: "datetime"},
			expectedFormat:  FormatDateTime,
			expectedSupport: true,
		},
		{
			name:            "TIMESTAMP should map to DATETIME",
			col:             TableColumn{DataType: "timestamp"},
			expectedFormat:  FormatDateTime,
			expectedSupport: true,
		},
		{
			name:            "TIME should map to TIME",
			col:             TableColumn{DataType: "time"},
			expectedFormat:  FormatTime,
			expectedSupport: true,
		},

		// JSON type
		{
			name:            "JSON should map to STRING",
			col:             TableColumn{DataType: "json"},
			expectedFormat:  FormatString,
			expectedSupport: true,
		},

		// Unsupported types
		{
			name:            "ENUM should be unsupported",
			col:             TableColumn{DataType: "enum"},
			expectedSupport: false,
			expectedSkip:    "enum/set",
		},
		{
			name:            "BLOB should be unsupported",
			col:             TableColumn{DataType: "blob"},
			expectedSupport: false,
			expectedSkip:    "binary/blob",
		},
		{
			name:            "GEOMETRY should be unsupported",
			col:             TableColumn{DataType: "geometry"},
			expectedSupport: false,
			expectedSkip:    "spatial type",
		},
		{
			name:            "BIT should be unsupported",
			col:             TableColumn{DataType: "bit"},
			expectedSupport: false,
			expectedSkip:    "BIT",
		},
		{
			name:            "Unknown type should be unsupported",
			col:             TableColumn{DataType: "unknown_type"},
			expectedSupport: false,
			expectedSkip:    "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapColumnToFieldType(tt.col)
			assert.Equal(t, tt.expectedSupport, result.Supported)
			if tt.expectedSupport {
				assert.Equal(t, tt.expectedFormat, result.Format)
			} else {
				assert.Contains(t, result.SkipReason, tt.expectedSkip)
			}
		})
	}
}

func TestReverseTypeMapper_MapColumnToFieldType_Validation(t *testing.T) {
	mapper := NewReverseTypeMapper()

	t.Run("DECIMAL should have precision and scale", func(t *testing.T) {
		result := mapper.MapColumnToFieldType(TableColumn{
			DataType:  "decimal",
			Precision: 10,
			Scale:     2,
		})
		assert.True(t, result.Supported)
		assert.NotNil(t, result.Validation)
		assert.Equal(t, 10, *result.Validation.Precision)
		assert.Equal(t, 2, *result.Validation.Scale)
	})

	t.Run("VARCHAR should have maxLength", func(t *testing.T) {
		result := mapper.MapColumnToFieldType(TableColumn{
			DataType: "varchar",
			Length:   255,
		})
		assert.True(t, result.Supported)
		assert.NotNil(t, result.Validation)
		assert.Equal(t, 255, *result.Validation.MaxLength)
	})

	t.Run("VARCHAR without length should have nil maxLength", func(t *testing.T) {
		result := mapper.MapColumnToFieldType(TableColumn{
			DataType: "varchar",
			Length:   0,
		})
		assert.True(t, result.Supported)
		assert.NotNil(t, result.Validation)
		assert.Nil(t, result.Validation.MaxLength)
	})
}

func TestReverseTypeMapper_BuildFieldDefinition(t *testing.T) {
	mapper := NewReverseTypeMapper()
	modelLocator := &ModelLocator{ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test-project"}}

	t.Run("should build field definition for supported type", func(t *testing.T) {
		col := TableColumn{
			Name:     "user_name",
			DataType: "varchar",
			Length:   100,
			Nullable: false,
			Comment:  "User name field",
		}
		fieldDef, result := mapper.BuildFieldDefinition(col, "model-123", modelLocator)

		assert.True(t, result.Supported)
		assert.NotNil(t, fieldDef)
		assert.Equal(t, "model-123", fieldDef.ModelID)
		assert.Equal(t, "user_name", fieldDef.Name)
		assert.Equal(t, "User Name", fieldDef.Title)
		assert.Equal(t, "User name field", fieldDef.Description)
		assert.True(t, fieldDef.NonNull)
	})

	t.Run("should return nil for unsupported type", func(t *testing.T) {
		col := TableColumn{
			Name:     "data",
			DataType: "blob",
		}
		fieldDef, result := mapper.BuildFieldDefinition(col, "model-123", modelLocator)

		assert.False(t, result.Supported)
		assert.Nil(t, fieldDef)
	})
}

func TestFormatFieldTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"snake_case", "user_name", "User Name"},
		{"camelCase", "userName", "UserName"},
		{"single word", "name", "Name"},
		{"multiple underscores", "first_middle_last", "First Middle Last"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFieldTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
