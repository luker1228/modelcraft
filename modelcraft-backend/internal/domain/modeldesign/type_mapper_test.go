package modeldesign

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ptr helper function for creating pointers
func ptr[T any](v T) *T {
	return &v
}

func TestMySQLTypeMapper_MapToMySQL(t *testing.T) {
	mapper := NewMySQLTypeMapper()

	tests := []struct {
		name    string
		field   *FieldDefinition
		want    string
		wantErr bool
	}{
		// String类型测试
		{
			name: "string_short_text",
			field: &FieldDefinition{
				Type:       GetFieldTypeByFormat(FormatString),
				Validation: &ValidationConfig{MaxLength: ptr(100)},
			},
			want: "VARCHAR(100)",
		},
		{
			name: "string_long_text",
			field: &FieldDefinition{
				Type:       GetFieldTypeByFormat(FormatString),
				Validation: &ValidationConfig{MaxLength: ptr(500)},
			},
			want: "TEXT",
		},
		{
			name: "string_no_max_length",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatString),
			},
			want: "TEXT",
		},
		{
			name: "string_max_length_255",
			field: &FieldDefinition{
				Type:       GetFieldTypeByFormat(FormatString),
				Validation: &ValidationConfig{MaxLength: ptr(255)},
			},
			want: "VARCHAR(255)",
		},
		{
			name: "string_max_length_256",
			field: &FieldDefinition{
				Type:       GetFieldTypeByFormat(FormatString),
				Validation: &ValidationConfig{MaxLength: ptr(256)},
			},
			want: "TEXT",
		},
		// UUIDV7类型测试（天然有序）
		{
			name: "uuid",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatUUID),
			},
			want: "CHAR(36)",
		},
		{
			name: "end_user_ref",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatEndUserRef),
			},
			want: "CHAR(36)",
		},
		// 整数类型测试
		{
			name: "integer",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatInteger),
			},
			want: "INT",
		},
		// Decimal类型测试
		{
			name: "decimal_with_precision",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatDecimal),
				Validation: &ValidationConfig{
					Precision: ptr(10),
					Scale:     ptr(2),
				},
			},
			want: "DECIMAL(10,2)",
		},
		{
			name: "decimal_without_precision",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatDecimal),
			},
			want: "DECIMAL(10,2)", // 默认精度
		},
		{
			name: "decimal_custom_precision",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatDecimal),
				Validation: &ValidationConfig{
					Precision: ptr(18),
					Scale:     ptr(4),
				},
			},
			want: "DECIMAL(18,4)",
		},
		// Boolean类型测试
		{
			name: "boolean",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatBoolean),
			},
			want: "TINYINT(1)",
		},
		// Number类型测试
		{
			name: "number",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatNumber),
			},
			want: "DOUBLE",
		},
		// 日期时间类型测试
		{
			name: "date",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatDate),
			},
			want: "DATE",
		},
		{
			name: "datetime",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatDateTime),
			},
			want: "DATETIME",
		},
		{
			name: "time",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatTime),
			},
			want: "TIME",
		},
		// StorageHint优先级测试
		{
			name: "storage_hint_mediumtext",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("MEDIUMTEXT"),
			},
			want: "MEDIUMTEXT",
		},
		{
			name: "storage_hint_longtext",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("LONGTEXT"),
			},
			want: "LONGTEXT",
		},
		{
			name: "storage_hint_overrides_auto_mapping",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("VARCHAR(191)"),
				Validation:  &ValidationConfig{MaxLength: ptr(500)},
			},
			want: "VARCHAR(191)",
		},
		// ENUM format tests with IsArray flag
		{
			name: "enum_single_select",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: false,
			},
			want: "VARCHAR(64)",
		},
		{
			name: "enum_multi_select",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: true,
			},
			want: "JSON",
		},
		{
			name: "enum_array_legacy",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatEnumArray),
			},
			want: "JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapper.MapToMySQL(tt.field)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMySQLTypeMapper_UnsupportedFormat(t *testing.T) {
	mapper := NewMySQLTypeMapper()

	// 创建一个不支持的格式类型
	field := &FieldDefinition{
		Type: GetFieldTypeByFormat(FormatType("unsupported")), // 不支持的格式
	}

	// 对于不支持的格式，getFieldTypeByFormat会返回nil
	if field.Type == nil {
		// 创建一个有效的FieldType但使用不支持的Format
		field.Type = &FieldType{
			SchemaType: SchemaTypeString,
			Format:     FormatType("unsupported"),
			Title:      "不支持的格式",
		}
	}

	_, err := mapper.MapToMySQL(field)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的格式类型")
}
