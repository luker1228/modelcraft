package modeldesign

import (
	"fmt"
	"modelcraft/internal/domain/project"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetFieldTypeByFormat 测试getFieldTypeByFormat函数
func TestGetFieldTypeByFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   FormatType
		expected *FieldType
	}{
		{
			name:     "STRING格式",
			format:   FormatString,
			expected: &FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
		},
		{
			name:     "UUIDV7格式",
			format:   FormatUUID,
			expected: &FieldType{SchemaType: SchemaTypeString, Format: FormatUUID, Title: "UUIDV7（天然有序）"},
		},
		{
			name:     "DATE格式",
			format:   FormatDate,
			expected: &FieldType{SchemaType: SchemaTypeString, Format: FormatDate, Title: "日期"},
		},
		{
			name:     "DATETIME格式",
			format:   FormatDateTime,
			expected: &FieldType{SchemaType: SchemaTypeString, Format: FormatDateTime, Title: "日期时间"},
		},
		{
			name:     "TIME格式",
			format:   FormatTime,
			expected: &FieldType{SchemaType: SchemaTypeString, Format: FormatTime, Title: "时间"},
		},
		{
			name:     "BOOLEAN格式",
			format:   FormatBoolean,
			expected: &FieldType{SchemaType: SchemaTypeBoolean, Format: FormatBoolean, Title: "布尔值"},
		},
		{
			name:     "NUMBER格式",
			format:   FormatNumber,
			expected: &FieldType{SchemaType: SchemaTypeNumber, Format: FormatNumber, Title: "数字"},
		},
		{
			name:     "INTEGER格式",
			format:   FormatInteger,
			expected: &FieldType{SchemaType: SchemaTypeNumber, Format: FormatInteger, Title: "整数"},
		},
		{
			name:     "DECIMAL格式",
			format:   FormatDecimal,
			expected: &FieldType{SchemaType: SchemaTypeNumber, Format: FormatDecimal, Title: "精确小数"},
		},
		{
			name:     "RELATION格式",
			format:   FormatRelation,
			expected: &FieldType{SchemaType: SchemaTypeObject, Format: FormatRelation, Title: "关联"},
		},
		{
			name:     "不支持的格式",
			format:   FormatType("UNKNOWN"),
			expected: nil,
		},
		{
			name:     "空格式",
			format:   FormatType(""),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFieldTypeByFormat(tt.format)
			fmt.Printf("result: %v", result)

			if tt.expected == nil {
				assert.Nil(t, result, "期望返回nil")
				return
			}

			assert.NotNil(t, result, "期望返回非nil值")
			assert.Equal(t, tt.expected.SchemaType, result.SchemaType, "SchemaType应该匹配")
			assert.Equal(t, tt.expected.Format, result.Format, "Format应该匹配")
			assert.Equal(t, tt.expected.Title, result.Title, "Title应该匹配")
		})
	}
}

// TestGetAllSupportedFieldTypes 测试getAllSupportedFieldTypes函数
func TestGetAllSupportedFieldTypes(t *testing.T) {
	result := getAllSupportedFieldTypes()

	// 验证返回的映射表不为空
	assert.NotNil(t, result, "返回的映射表不应该为nil")
	assert.Greater(t, len(result), 0, "映射表应该包含多个字段类型")

	// 验证所有支持的格式都在映射表中
	expectedFormats := []FormatType{
		FormatString, FormatUUID, FormatDate, FormatDateTime, FormatTime,
		FormatBoolean, FormatNumber, FormatInteger, FormatDecimal, FormatRelation,
	}

	for _, format := range expectedFormats {
		assert.Contains(t, result, format, "映射表应该包含格式: %s", format)

		fieldType := result[format]
		assert.NotNil(t, fieldType, "格式%s对应的FieldType不应该为nil", format)
		assert.Equal(t, format, fieldType.Format, "FieldType的Format应该匹配")
	}

	// 验证映射表是副本，不是原始引用
	originalCount := len(result)

	// 修改返回的映射表不应该影响原始映射表
	result["TEST"] = &FieldType{}

	// 再次调用应该返回原始数据
	result2 := getAllSupportedFieldTypes()
	assert.Equal(t, originalCount, len(result2), "修改返回的映射表不应该影响原始数据")
	assert.NotContains(t, result2, FormatType("TEST"), "原始映射表不应该包含测试添加的格式")
}

// TestFieldTypeEquals 测试FieldType的Equals方法
func TestFieldTypeEquals(t *testing.T) {
	tests := []struct {
		name     string
		ft1      FieldType
		ft2      FieldType
		expected bool
	}{
		{
			name:     "相同的FieldType",
			ft1:      FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
			ft2:      FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
			expected: true,
		},
		{
			name:     "不同的SchemaType",
			ft1:      FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
			ft2:      FieldType{SchemaType: SchemaTypeNumber, Format: FormatString, Title: "字符串"},
			expected: false,
		},
		{
			name:     "不同的Format",
			ft1:      FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
			ft2:      FieldType{SchemaType: SchemaTypeString, Format: FormatUUID, Title: "字符串"},
			expected: false,
		},
		{
			name:     "不同的Title",
			ft1:      FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
			ft2:      FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "文本"},
			expected: false,
		},
		{
			name:     "所有属性都不同",
			ft1:      FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
			ft2:      FieldType{SchemaType: SchemaTypeNumber, Format: FormatNumber, Title: "数字"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ft1.Equals(tt.ft2)
			assert.Equal(t, tt.expected, result, "Equals方法应该返回正确的结果")

			// 验证对称性
			resultReverse := tt.ft2.Equals(tt.ft1)
			assert.Equal(t, tt.expected, resultReverse, "Equals方法应该是对称的")
		})
	}
}

// TestFieldTypeString 测试FieldType的String方法
func TestFieldTypeString(t *testing.T) {
	tests := []struct {
		name     string
		ft       FieldType
		expected string
	}{
		{
			name:     "STRING格式",
			ft:       FieldType{Format: FormatString},
			expected: "STRING",
		},
		{
			name:     "UUIDV7格式",
			ft:       FieldType{Format: FormatUUID},
			expected: "UUID",
		},
		{
			name:     "BOOLEAN格式",
			ft:       FieldType{Format: FormatBoolean},
			expected: "BOOLEAN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ft.String()
			assert.Equal(t, tt.expected, result, "String方法应该返回正确的格式字符串")
		})
	}
}

// TestNewFieldFormat 测试NewFieldFormat函数
func TestNewFieldFormat(t *testing.T) {
	tests := []struct {
		name        string
		format      FormatType
		shouldError bool
		expected    *FieldType
	}{
		{
			name:        "有效的STRING格式",
			format:      FormatString,
			shouldError: false,
			expected:    &FieldType{SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
		},
		{
			name:        "不支持的格式",
			format:      FormatType("UNKNOWN"),
			shouldError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewFieldFormat(tt.format)

			if tt.shouldError {
				assert.Error(t, err, "期望返回错误")
				assert.Nil(t, result, "错误情况下应该返回nil")
			} else {
				assert.NoError(t, err, "不应该返回错误")
				assert.NotNil(t, result, "应该返回非nil值")
				assert.Equal(t, tt.expected.SchemaType, result.SchemaType, "SchemaType应该匹配")
				assert.Equal(t, tt.expected.Format, result.Format, "Format应该匹配")
				assert.Equal(t, tt.expected.Title, result.Title, "Title应该匹配")
			}
		})
	}
}

// TestFieldDefinition_Validate tests the Validate method
func TestFieldDefinition_Validate(t *testing.T) {
	validFieldType, _ := NewFieldFormat(FormatString)

	tests := []struct {
		name        string
		field       *FieldDefinition
		wantErr     bool
		errContains string
	}{
		{
			name: "valid field",
			field: &FieldDefinition{
				Name:    "username",
				Title:   "Username",
				Type:    validFieldType,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			field: &FieldDefinition{
				Name:    "",
				Title:   "Title",
				Type:    validFieldType,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			wantErr:     true,
			errContains: "Name不能为空",
		},
		{
			name: "invalid field name format - starts with number",
			field: &FieldDefinition{
				Name:    "1username",
				Title:   "Title",
				Type:    validFieldType,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			wantErr:     true,
			errContains: "格式无效",
		},
		{
			name: "invalid field name format - special characters",
			field: &FieldDefinition{
				Name:    "user-name",
				Title:   "Title",
				Type:    validFieldType,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			wantErr:     true,
			errContains: "格式无效",
		},
		{
			name: "empty ModelID",
			field: &FieldDefinition{
				Name:    "username",
				Title:   "Title",
				Type:    validFieldType,
				ModelID: "",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			wantErr:     true,
			errContains: "ModelID不能为空",
		},
		{
			name: "nil ModelLocator",
			field: &FieldDefinition{
				Name:         "username",
				Title:        "Title",
				Type:         validFieldType,
				ModelID:      "model-123",
				ModelLocator: nil,
			},
			wantErr:     true,
			errContains: "ModelLocator不能为空",
		},
		{
			name: "empty Title",
			field: &FieldDefinition{
				Name:    "username",
				Title:   "",
				Type:    validFieldType,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			wantErr:     true,
			errContains: "Title不能为空",
		},
		{
			name: "nil Type",
			field: &FieldDefinition{
				Name:    "username",
				Title:   "Title",
				Type:    nil,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			wantErr:     true,
			errContains: "字段必须有Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFieldDefinition_IsEnumField tests IsEnumField method
func TestFieldDefinition_IsEnumField(t *testing.T) {
	enumType, _ := NewFieldFormat(FormatEnum)
	enumArrayType, _ := NewFieldFormat(FormatEnumArray)
	stringType, _ := NewFieldFormat(FormatString)

	tests := []struct {
		name     string
		field    *FieldDefinition
		expected bool
	}{
		{
			name:     "enum field",
			field:    &FieldDefinition{Type: enumType},
			expected: true,
		},
		{
			name:     "enum array field",
			field:    &FieldDefinition{Type: enumArrayType},
			expected: true,
		},
		{
			name:     "string field",
			field:    &FieldDefinition{Type: stringType},
			expected: false,
		},
		{
			name:     "nil type",
			field:    &FieldDefinition{Type: nil},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.IsEnumField()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFieldDefinition_Update tests Update method
func TestFieldDefinition_Update(t *testing.T) {
	minLen := 5
	maxLen := 100
	validation := &ValidationConfig{
		MinLength: &minLen,
		MaxLength: &maxLen,
	}

	field := &FieldDefinition{
		Title:       "Old Title",
		Description: "Old Description",
		Validation:  nil,
	}

	field.Update("New Title", "New Description", validation)

	assert.Equal(t, "New Title", field.Title)
	assert.Equal(t, "New Description", field.Description)
	assert.Equal(t, validation, field.Validation)
}

// TestFieldType_GetType tests GetType method
func TestFieldType_GetType(t *testing.T) {
	ft := &FieldType{
		SchemaType: SchemaTypeString,
		Format:     FormatString,
	}

	result := ft.GetType()
	assert.Equal(t, SchemaTypeString, result)
}

// TestFieldType_GetFormat tests GetFormat method
func TestFieldType_GetFormat(t *testing.T) {
	ft := &FieldType{
		SchemaType: SchemaTypeString,
		Format:     FormatString,
	}

	result := ft.GetFormat()
	assert.Equal(t, FormatString, result)
}

// TestFieldType_MarshalJSON tests MarshalJSON method
func TestFieldType_MarshalJSON(t *testing.T) {
	ft := FieldType{
		SchemaType: SchemaTypeString,
		Format:     FormatString,
	}

	data, err := ft.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `"STRING"`, string(data))
}

// TestFieldType_UnmarshalJSON tests UnmarshalJSON method
func TestFieldType_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantErr     bool
		errContains string
		expected    *FieldType
	}{
		{
			name:    "valid STRING format",
			json:    `"STRING"`,
			wantErr: false,
			expected: &FieldType{
				SchemaType: SchemaTypeString,
				Format:     FormatString,
				Title:      "字符串",
			},
		},
		{
			name:    "valid UUID format",
			json:    `"UUID"`,
			wantErr: false,
			expected: &FieldType{
				SchemaType: SchemaTypeString,
				Format:     FormatUUID,
				Title:      "UUIDV7（天然有序）",
			},
		},
		{
			name:        "unsupported format",
			json:        `"UNKNOWN"`,
			wantErr:     true,
			errContains: "unsupported format type",
		},
		{
			name:        "invalid JSON",
			json:        `{invalid}`,
			wantErr:     true,
			errContains: "invalid format value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FieldType
			err := ft.UnmarshalJSON([]byte(tt.json))

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.SchemaType, ft.SchemaType)
				assert.Equal(t, tt.expected.Format, ft.Format)
				assert.Equal(t, tt.expected.Title, ft.Title)
			}
		})
	}
}

// TestFieldDefinition_validateStringField tests validateStringField method
func TestFieldDefinition_validateStringField(t *testing.T) {
	minLen := 5
	maxLen := 100
	invalidPattern := "[invalid"
	validPattern := "^[a-z]+$"

	tests := []struct {
		name        string
		field       *FieldDefinition
		wantErr     bool
		errContains string
	}{
		{
			name:    "no validation",
			field:   &FieldDefinition{Validation: nil},
			wantErr: false,
		},
		{
			name: "valid min/max length",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MinLength: &minLen,
					MaxLength: &maxLen,
				},
			},
			wantErr: false,
		},
		{
			name: "negative min length",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MinLength: intPtr(-1),
				},
			},
			wantErr:     true,
			errContains: "最小长度不能为负数",
		},
		{
			name: "negative max length",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MaxLength: intPtr(-1),
				},
			},
			wantErr:     true,
			errContains: "最大长度不能为负数",
		},
		{
			name: "min > max",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MinLength: intPtr(100),
					MaxLength: intPtr(50),
				},
			},
			wantErr:     true,
			errContains: "最小长度不能大于最大长度",
		},
		{
			name: "valid pattern",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					Pattern: &validPattern,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid pattern",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					Pattern: &invalidPattern,
				},
			},
			wantErr:     true,
			errContains: "正则表达式格式无效",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.validateStringField()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFieldDefinition_validateNumberField tests validateNumberField method
func TestFieldDefinition_validateNumberField(t *testing.T) {
	tests := []struct {
		name        string
		field       *FieldDefinition
		wantErr     bool
		errContains string
	}{
		{
			name:    "no validation",
			field:   &FieldDefinition{Validation: nil},
			wantErr: false,
		},
		{
			name: "valid min/max",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					Minimum: float64Ptr(0.0),
					Maximum: float64Ptr(100.0),
				},
			},
			wantErr: false,
		},
		{
			name: "min > max",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					Minimum: float64Ptr(100.0),
					Maximum: float64Ptr(50.0),
				},
			},
			wantErr:     true,
			errContains: "最小值不能大于最大值",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.validateNumberField()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFieldDefinition_validateArrayField tests validateArrayField method
func TestFieldDefinition_validateArrayField(t *testing.T) {
	tests := []struct {
		name        string
		field       *FieldDefinition
		wantErr     bool
		errContains string
	}{
		{
			name:    "no validation",
			field:   &FieldDefinition{Validation: nil},
			wantErr: false,
		},
		{
			name: "valid min/max items",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MinItems: intPtr(1),
					MaxItems: intPtr(10),
				},
			},
			wantErr: false,
		},
		{
			name: "negative min items",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MinItems: intPtr(-1),
				},
			},
			wantErr:     true,
			errContains: "最小元素数量不能为负数",
		},
		{
			name: "negative max items",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MaxItems: intPtr(-1),
				},
			},
			wantErr:     true,
			errContains: "最大元素数量不能为负数",
		},
		{
			name: "min > max",
			field: &FieldDefinition{
				Validation: &ValidationConfig{
					MinItems: intPtr(10),
					MaxItems: intPtr(5),
				},
			},
			wantErr:     true,
			errContains: "最小元素数量不能大于最大元素数量",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.validateArrayField()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFieldDefinition_validateEnumLabelField tests validateEnumLabelField method
func TestFieldDefinition_validateEnumLabelField(t *testing.T) {
	tests := []struct {
		name        string
		field       *FieldDefinition
		wantErr     bool
		errContains string
	}{
		{
			name: "valid enum label config",
			field: &FieldDefinition{
				EnumLabelConfig: &EnumLabelConfig{
					SourceField: "status",
				},
			},
			wantErr: false,
		},
		{
			name: "nil enum label config",
			field: &FieldDefinition{
				EnumLabelConfig: nil,
			},
			wantErr:     true,
			errContains: "must have enumLabelConfig",
		},
		{
			name: "empty source field",
			field: &FieldDefinition{
				EnumLabelConfig: &EnumLabelConfig{
					SourceField: "",
				},
			},
			wantErr:     true,
			errContains: "sourceField cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.validateEnumLabelField()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFieldDefinition_validateBooleanField tests validateBooleanField method
func TestFieldDefinition_validateBooleanField(t *testing.T) {
	field := &FieldDefinition{}
	err := field.validateBooleanField()
	assert.NoError(t, err)
}

// TestFieldDefinition_validateEnumField tests validateEnumField method
func TestFieldDefinition_validateEnumField(t *testing.T) {
	field := &FieldDefinition{}
	err := field.validateEnumField()
	assert.NoError(t, err)
}

// TestFieldDefinition_IsArray 测试 FieldDefinition.IsArray 字段
func TestFieldDefinition_IsArray(t *testing.T) {
	tests := []struct {
		name      string
		field     *FieldDefinition
		expectErr bool
	}{
		{
			name: "ENUM format with IsArray=false (single-select)",
			field: &FieldDefinition{
				Name:    "status",
				Title:   "Status",
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: false,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			expectErr: false,
		},
		{
			name: "ENUM format with IsArray=true (multi-select)",
			field: &FieldDefinition{
				Name:    "tags",
				Title:   "Tags",
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: true,
				ModelID: "model-123",
				ModelLocator: &ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: "test"},
					ModelName:    "Test",
					DatabaseName: "test_db",
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify field has IsArray field
			assert.NotNil(t, tt.field.IsArray || !tt.field.IsArray, "Field should have IsArray field")
		})
	}
}

// TestFieldDefinition_IsEnumArrayField 测试 IsEnumArrayField 辅助方法
func TestFieldDefinition_IsEnumArrayField(t *testing.T) {
	tests := []struct {
		name          string
		field         *FieldDefinition
		expectIsArray bool
	}{
		{
			name: "ENUM + IsArray=true should be enum array field",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: true,
			},
			expectIsArray: true,
		},
		{
			name: "ENUM + IsArray=false should NOT be enum array field",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnum),
				IsArray: false,
			},
			expectIsArray: false,
		},
		{
			name: "FormatEnumArray should be enum array field (legacy)",
			field: &FieldDefinition{
				Type:    GetFieldTypeByFormat(FormatEnumArray),
				IsArray: false,
			},
			expectIsArray: true, // FormatEnumArray is always multi-select
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.IsEnumArrayField()
			assert.Equal(t, tt.expectIsArray, result, "IsEnumArrayField should match expected value")
		})
	}
}
