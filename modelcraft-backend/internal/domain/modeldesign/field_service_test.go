package modeldesign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFieldService_ValidateDuplicates_Success 测试字段重复验证成功
func TestFieldService_ValidateDuplicates_Success(t *testing.T) {
	service := NewFieldService()

	fields := []FieldDefinition{
		{
			Name:  "field1",
			Title: "字段1",
			Type:  GetFieldTypeByFormat(FormatString),
		},
		{
			Name:  "field2",
			Title: "字段2",
			Type:  GetFieldTypeByFormat(FormatInteger),
		},
	}

	err := service.ValidateDuplicates(fields)
	assert.NoError(t, err)
}

// TestFieldService_ValidateDuplicates_EmptyFields 测试空字段列表
func TestFieldService_ValidateDuplicates_EmptyFields(t *testing.T) {
	service := NewFieldService()

	err := service.ValidateDuplicates([]FieldDefinition{})
	assert.NoError(t, err)
}

// TestFieldService_ValidateDuplicates_DuplicateNames 测试重复字段名
func TestFieldService_ValidateDuplicates_DuplicateNames(t *testing.T) {
	service := NewFieldService()

	fields := []FieldDefinition{
		{
			Name:  "field1",
			Title: "字段1",
			Type:  GetFieldTypeByFormat(FormatString),
		},
		{
			Name:  "field1", // 重复的字段名
			Title: "字段2",
			Type:  GetFieldTypeByFormat(FormatInteger),
		},
	}

	err := service.ValidateDuplicates(fields)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field1'")
}

// TestFieldService_ValidateDuplicates_EmptyName 测试空字段名
func TestFieldService_ValidateDuplicates_EmptyName(t *testing.T) {
	service := NewFieldService()

	fields := []FieldDefinition{
		{
			Name:  "", // 空的字段名
			Title: "字段1",
			Type:  GetFieldTypeByFormat(FormatString),
		},
	}

	err := service.ValidateDuplicates(fields)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Name cant be blank")
}

// TestFieldService_GetNamesFromFields 测试从字段列表获取名称
func TestFieldService_GetNamesFromFields(t *testing.T) {
	service := NewFieldService()

	fields := []*FieldDefinition{
		{Name: "field1"},
		{Name: "field2"},
		{Name: "field3"},
	}

	names := service.GetNamesFromFields(fields)
	assert.Equal(t, []string{"field1", "field2", "field3"}, names)
}

// TestFieldService_GetNamesFromFields_Empty 测试空字段列表
func TestFieldService_GetNamesFromFields_Empty(t *testing.T) {
	service := NewFieldService()

	names := service.GetNamesFromFields([]*FieldDefinition{})
	assert.Empty(t, names)
}

// TestFieldService_ValidateAddFieldsNotExist_Success 测试添加字段验证成功
func TestFieldService_ValidateAddFieldsNotExist_Success(t *testing.T) {
	service := NewFieldService()

	existingFields := []*FieldDefinition{
		{Name: "existing1"},
		{Name: "existing2"},
	}

	err := service.ValidateAddFieldsNotExist(existingFields, "newField1", "newField2")
	assert.NoError(t, err)
}

// TestFieldService_ValidateAddFieldsNotExist_Duplicate 测试添加重复字段
func TestFieldService_ValidateAddFieldsNotExist_Duplicate(t *testing.T) {
	service := NewFieldService()

	existingFields := []*FieldDefinition{
		{Name: "existing1"},
		{Name: "existing2"},
	}

	err := service.ValidateAddFieldsNotExist(existingFields, "newField1", "existing1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "字段名称 'existing1' 已存在")
}

// TestFieldService_GetSupportedFieldFormats 测试获取支持的字段格式
func TestFieldService_GetSupportedFieldFormats(t *testing.T) {
	service := NewFieldService()

	formats := service.GetSupportedFieldFormats()
	assert.NotEmpty(t, formats)

	// 验证包含一些基本格式
	assert.Contains(t, formats, FormatString)
	assert.Contains(t, formats, FormatInteger)
	assert.Contains(t, formats, FormatBoolean)
}
