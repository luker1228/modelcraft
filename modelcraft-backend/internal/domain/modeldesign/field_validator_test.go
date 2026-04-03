package modeldesign

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldValidator_ValidateUniqueConstraint(t *testing.T) {
	validator := NewFieldValidator()

	tests := []struct {
		name    string
		field   *FieldDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "unique_string_with_max_length",
			field: &FieldDefinition{
				Type:       GetFieldTypeByFormat(FormatString),
				IsUnique:   true,
				Validation: &ValidationConfig{MaxLength: ptr(100)},
			},
			wantErr: false,
		},
		{
			name: "unique_string_without_max_length",
			field: &FieldDefinition{
				Type:     GetFieldTypeByFormat(FormatString),
				IsUnique: true,
			},
			wantErr: true,
			errMsg:  "isUnique字段必须设置maxLength",
		},
		{
			name: "unique_uuid_no_max_length_required",
			field: &FieldDefinition{
				Type:     GetFieldTypeByFormat(FormatUUID),
				IsUnique: true,
			},
			wantErr: false, // UUIDV7 has fixed length and natural ordering, no maxLength required
		},
		{
			name: "unique_integer",
			field: &FieldDefinition{
				Type:     GetFieldTypeByFormat(FormatInteger),
				IsUnique: true,
			},
			wantErr: false,
		},
		{
			name: "non_unique_string",
			field: &FieldDefinition{
				Type:     GetFieldTypeByFormat(FormatString),
				IsUnique: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUniqueConstraint(tt.field)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateStorageHint(t *testing.T) {
	validator := NewFieldValidator()

	tests := []struct {
		name    string
		field   *FieldDefinition
		wantErr bool
		errMsg  string
	}{
		// --- 原有 VARCHAR 裸类型校验 ---
		{
			name: "varchar_with_valid_max_length",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("VARCHAR"),
				Validation:  &ValidationConfig{MaxLength: ptr(100)},
			},
			wantErr: false,
		},
		{
			name: "varchar_without_max_length",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("VARCHAR"),
			},
			wantErr: true,
			errMsg:  "storageHint='VARCHAR'必须设置maxLength",
		},
		{
			name: "varchar_with_max_length_exceeding_255",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("VARCHAR"),
				Validation:  &ValidationConfig{MaxLength: ptr(500)},
			},
			wantErr: true,
			errMsg:  "storageHint='VARCHAR'但maxLength>255",
		},
		{
			name: "varchar_with_max_length_255",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("VARCHAR"),
				Validation:  &ValidationConfig{MaxLength: ptr(255)},
			},
			wantErr: false,
		},
		{
			name: "mediumtext_no_validation_needed",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("MEDIUMTEXT"),
			},
			wantErr: false,
		},
		{
			name: "no_storage_hint",
			field: &FieldDefinition{
				Type: GetFieldTypeByFormat(FormatString),
			},
			wantErr: false,
		},
		{
			name: "empty_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr(""),
			},
			wantErr: false,
		},
		// --- storageHint 与 format 冲突：字符串 format 使用数值/日期类型 ---
		{
			name: "string_format_with_int_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("INT"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		{
			name: "string_format_with_datetime_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("DATETIME"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		{
			name: "uuid_format_with_int_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatUUID),
				StorageHint: ptr("INT"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		// --- storageHint 与 format 冲突：数值 format 使用字符串/日期类型 ---
		{
			name: "integer_format_with_varchar_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatInteger),
				StorageHint: ptr("VARCHAR(64)"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		{
			name: "number_format_with_text_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatNumber),
				StorageHint: ptr("TEXT"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		{
			name: "decimal_format_with_datetime_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatDecimal),
				StorageHint: ptr("DATETIME"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		{
			name: "boolean_format_with_varchar_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatBoolean),
				StorageHint: ptr("VARCHAR(10)"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		// --- storageHint 与 format 冲突：日期时间 format 使用字符串/数值类型 ---
		{
			name: "datetime_format_with_varchar_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatDateTime),
				StorageHint: ptr("VARCHAR(32)"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		{
			name: "date_format_with_int_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatDate),
				StorageHint: ptr("INT"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		{
			name: "time_format_with_bigint_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatTime),
				StorageHint: ptr("BIGINT"),
			},
			wantErr: true,
			errMsg:  "不兼容",
		},
		// --- ENUM/ENUM_ARRAY 禁止 storageHint ---
		{
			name: "enum_format_with_any_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatEnum),
				StorageHint: ptr("VARCHAR(64)"),
			},
			wantErr: true,
			errMsg:  "不支持自定义storageHint",
		},
		{
			name: "enum_array_format_with_any_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatEnumArray),
				StorageHint: ptr("JSON"),
			},
			wantErr: true,
			errMsg:  "不支持自定义storageHint",
		},
		// --- 兼容：数值 format 使用正确类型 ---
		{
			name: "integer_format_with_bigint_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatInteger),
				StorageHint: ptr("BIGINT"),
			},
			wantErr: false,
		},
		{
			name: "decimal_format_with_custom_precision",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatDecimal),
				StorageHint: ptr("DECIMAL(18,4)"),
			},
			wantErr: false,
		},
		{
			name: "number_format_with_double_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatNumber),
				StorageHint: ptr("DOUBLE"),
			},
			wantErr: false,
		},
		// --- 兼容：日期时间 format 使用正确类型 ---
		{
			name: "datetime_format_with_timestamp_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatDateTime),
				StorageHint: ptr("TIMESTAMP"),
			},
			wantErr: false,
		},
		// --- 兼容：VARCHAR(64) 带长度参数 ---
		{
			name: "string_format_with_varchar64_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("VARCHAR(64)"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStorageHint(tt.field)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateAll(t *testing.T) {
	validator := NewFieldValidator()

	tests := []struct {
		name    string
		field   *FieldDefinition
		wantErr bool
	}{
		{
			name: "valid_field_all_checks_pass",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				IsUnique:    true,
				StorageHint: ptr("VARCHAR"),
				Validation:  &ValidationConfig{MaxLength: ptr(100)},
			},
			wantErr: false,
		},
		{
			name: "invalid_unique_constraint",
			field: &FieldDefinition{
				Type:     GetFieldTypeByFormat(FormatString),
				IsUnique: true,
			},
			wantErr: true,
		},
		{
			name: "invalid_storage_hint",
			field: &FieldDefinition{
				Type:        GetFieldTypeByFormat(FormatString),
				StorageHint: ptr("VARCHAR"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAll(tt.field)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFieldValidator_ValidateIsArray(t *testing.T) {
	validator := NewFieldValidator()

	tests := []struct {
		name    string
		field   *FieldDefinition
		wantErr bool
		errMsg  string
	}{
		// --- format=ENUM, isArray=false: valid ---
		{
			name: "enum_format_isArray_false_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: false,
			},
			wantErr: false,
		},
		// --- format=ENUM, isArray=true, no enum associated: valid ---
		{
			name: "enum_format_isArray_true_no_enum_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: true,
				Enum:    nil, // no enum definition associated
			},
			wantErr: false,
		},
		// --- format=ENUM, isArray=true, enum.IsMultiSelect=true: valid ---
		{
			name: "enum_format_isArray_true_enum_multiSelect_true_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: true,
				Enum: &EnumDefinition{
					IsMultiSelect: true,
				},
			},
			wantErr: false,
		},
		// --- format=ENUM, isArray=true, enum.IsMultiSelect=false: invalid ---
		{
			name: "enum_format_isArray_true_enum_multiSelect_false_invalid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: true,
				Enum: &EnumDefinition{
					IsMultiSelect: false,
				},
			},
			wantErr: true,
			errMsg:  "enum does not support multi-select",
		},
		// --- format!=ENUM, isArray=true: invalid ---
		{
			name: "string_format_isArray_true_invalid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatString),
				IsArray: true,
			},
			wantErr: true,
			errMsg:  "enum array not allowed for non-enum format",
		},
		{
			name: "integer_format_isArray_true_invalid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatInteger),
				IsArray: true,
			},
			wantErr: true,
			errMsg:  "enum array not allowed for non-enum format",
		},
		{
			name: "uuid_format_isArray_true_invalid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatUUID),
				IsArray: true,
			},
			wantErr: true,
			errMsg:  "enum array not allowed for non-enum format",
		},
		{
			name: "datetime_format_isArray_true_invalid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatDateTime),
				IsArray: true,
			},
			wantErr: true,
			errMsg:  "enum array not allowed for non-enum format",
		},
		// --- format!=ENUM, isArray=false: valid ---
		{
			name: "string_format_isArray_false_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatString),
				IsArray: false,
			},
			wantErr: false,
		},
		{
			name: "integer_format_isArray_false_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatInteger),
				IsArray: false,
			},
			wantErr: false,
		},
		{
			name: "boolean_format_isArray_false_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatBoolean),
				IsArray: false,
			},
			wantErr: false,
		},
		// --- ENUM_ARRAY format (legacy) should not be affected by isArray flag ---
		{
			name: "enum_array_format_isArray_false_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnumArray),
				IsArray: false,
			},
			wantErr: false,
		},
		{
			name: "enum_array_format_isArray_true_valid",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnumArray),
				IsArray: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateIsArray(tt.field)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
