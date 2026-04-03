package mapper

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/interfaces/http/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testModelID = "model-001"

// TestConvertFieldDTOToDomain_RelateFKID_Propagated 验证 RELATION 格式字段的 relateFkId 被正确传递到 domain 对象
// 这是修复 bug 的回归测试：ConvertFieldDTOToDomain 曾经漏掉 RelateFKID 赋值，
// 导致 RELATION 字段始终因 "RELATION format field must have relate_fk_id" 校验失败。
func TestConvertFieldDTOToDomain_RelateFKID_Propagated(t *testing.T) {
	fkID := "019d43b0-29f5-7220-b491-0d6f022c1e01"
	dto := &dtos.FieldDefinitionDTO{
		Name:       "gl",
		Title:      "gl",
		Format:     modeldesign.FormatRelation,
		RelateFKID: &fkID,
	}

	field, err := FieldMapper.ConvertFieldDTOToDomain(testModelID, nil, dto)

	require.NoError(t, err)
	require.NotNil(t, field.RelateFKID, "RelateFKID 应被赋值，不应为 nil")
	assert.Equal(t, fkID, *field.RelateFKID)
}

// TestConvertFieldDTOToDomain_RelateFKID_Nil_WhenNotSet 验证不传 relateFkId 时 domain 对象的 RelateFKID 为 nil
func TestConvertFieldDTOToDomain_RelateFKID_Nil_WhenNotSet(t *testing.T) {
	dto := &dtos.FieldDefinitionDTO{
		Name:       "name",
		Title:      "Name",
		Format:     modeldesign.FormatString,
		RelateFKID: nil,
	}

	field, err := FieldMapper.ConvertFieldDTOToDomain(testModelID, nil, dto)

	require.NoError(t, err)
	assert.Nil(t, field.RelateFKID)
}

// TestConvertFieldDTOToDomain_RelationFormat_PassesDomainValidation 验证带 relateFkId 的 RELATION 字段能通过 domain 层校验
// 端到端验证修复后整条链路（mapper → domain.Validate）不再报错
func TestConvertFieldDTOToDomain_RelationFormat_PassesDomainValidation(t *testing.T) {
	fkID := "019d43b0-29f5-7220-b491-0d6f022c1e01"
	dto := &dtos.FieldDefinitionDTO{
		Name:       "gl",
		Title:      "gl",
		Format:     modeldesign.FormatRelation,
		RelateFKID: &fkID,
	}

	field, err := FieldMapper.ConvertFieldDTOToDomain(testModelID, nil, dto)
	require.NoError(t, err)

	// 补全 domain 校验所需的必填字段
	field.ModelLocator = &modeldesign.ModelLocator{
		ProjectScope: project.ProjectScope{OrgName: "my-org", ProjectSlug: "my-project"},
		DatabaseName: "db",
		ModelName:    "order",
	}

	validationErr := field.Validate()
	assert.NoError(t, validationErr, "带 relateFkId 的 RELATION 字段不应报 'must have relate_fk_id' 错误")
}

// TestConvertFieldDTOToDomain_BasicFields 验证基础字段映射正确（name、title、format 等）
func TestConvertFieldDTOToDomain_BasicFields(t *testing.T) {
	dto := &dtos.FieldDefinitionDTO{
		Name:     "age",
		Title:    "Age",
		Format:   modeldesign.FormatInteger,
		NonNull:  true,
		Required: true,
		IsUnique: true,
		IsArray:  false,
	}

	field, err := FieldMapper.ConvertFieldDTOToDomain(testModelID, nil, dto)

	require.NoError(t, err)
	assert.Equal(t, testModelID, field.ModelID)
	assert.Equal(t, "age", field.Name)
	assert.Equal(t, "Age", field.Title)
	assert.Equal(t, modeldesign.FormatInteger, field.Type.Format)
	assert.True(t, field.NonNull)
	assert.True(t, field.Required)
	assert.True(t, field.IsUnique)
	assert.False(t, field.IsArray)
	assert.Nil(t, field.RelateFKID)
}

// TestConvertFieldDTOsToDomain_RelateFKID_Propagated 验证批量转换时 RelateFKID 同样被正确传递
func TestConvertFieldDTOsToDomain_RelateFKID_Propagated(t *testing.T) {
	fkID := "fk-batch-001"
	dtos := []*dtos.FieldDefinitionDTO{
		{
			Name:       "rel",
			Title:      "Rel",
			Format:     modeldesign.FormatRelation,
			RelateFKID: &fkID,
		},
		{
			Name:   "name",
			Title:  "Name",
			Format: modeldesign.FormatString,
		},
	}

	fields, err := FieldMapper.ConvertFieldDTOsToDomain(testModelID, nil, dtos)

	require.NoError(t, err)
	require.Len(t, fields, 2)
	require.NotNil(t, fields[0].RelateFKID)
	assert.Equal(t, fkID, *fields[0].RelateFKID)
	assert.Nil(t, fields[1].RelateFKID)
}
