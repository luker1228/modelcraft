package mapper

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/http/dtos"
	"time"
)

var FieldMapper *fieldMapper

func init() {
	FieldMapper = &fieldMapper{}
}

type fieldMapper struct{}

func (m *fieldMapper) ConvertValidationDTO2Domain(input *dtos.ValidationConfigDTO) *modeldesign.ValidationConfig {
	if input == nil {
		return nil
	}
	validation := &modeldesign.ValidationConfig{}
	if input.MinLength != nil {
		minLength := *input.MinLength
		validation.MinLength = &minLength
	}
	if input.MaxLength != nil {
		maxLength := *input.MaxLength
		validation.MaxLength = &maxLength
	}
	if input.Pattern != nil {
		validation.Pattern = input.Pattern
	}
	if input.Minimum != nil {
		minimum := float64(*input.Minimum)
		validation.Minimum = &minimum
	}
	if input.Maximum != nil {
		maximum := float64(*input.Maximum)
		validation.Maximum = &maximum
	}
	if len(input.EnumValues) > 0 {
		validation.EnumValues = input.EnumValues
	}
	return validation
}

func (m *fieldMapper) ConvertValidationDomain2DTO(validation *modeldesign.ValidationConfig) *dtos.ValidationConfigDTO {
	return &dtos.ValidationConfigDTO{
		MinLength:  validation.MinLength,
		MaxLength:  validation.MaxLength,
		Pattern:    validation.Pattern,
		Minimum:    validation.Minimum,
		Maximum:    validation.Maximum,
		EnumValues: validation.EnumValues,
	}
}

// ConvertFieldDTOToDomain 将FieldDefinitionDTO转换为域模型FieldDefinition
func (m *fieldMapper) ConvertFieldDTOToDomain(
	modelID string,
	modelLocator *modeldesign.ModelLocator,
	field *dtos.FieldDefinitionDTO,
) (*modeldesign.FieldDefinition, error) {
	props := m.ConvertValidationDTO2Domain(field.ValidationConfig)
	fieldType, err := modeldesign.NewFieldFormat(field.Format)
	if err != nil {
		return nil, err
	}

	f := &modeldesign.FieldDefinition{
		ModelID:      modelID,
		ModelLocator: modelLocator,
		Name:         field.Name,
		Title:        field.Title,
		Description:  field.Description,
		Type:         fieldType,
		StorageHint:  field.StorageHint,
		NonNull:      field.NonNull,
		Required:     field.Required,
		IsUnique:     field.IsUnique,
		IsPrimary:    false,
		IsArray:      field.IsArray,
		Validation:   props,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	applyEnumBindingByFormat(f, field)

	if field.RelateFKID != nil {
		f.RelateFKID = field.RelateFKID
	}

	return f, nil
}

// ConvertFieldDTOsToDomain 批量将FieldDefinitionDTO转换为域模型FieldDefinition
func (m *fieldMapper) ConvertFieldDTOsToDomain(
	modelID string,
	modelLocator *modeldesign.ModelLocator,
	fieldDtos []*dtos.FieldDefinitionDTO,
) ([]*modeldesign.FieldDefinition, error) {
	fields := make([]*modeldesign.FieldDefinition, 0, len(fieldDtos))
	for _, fieldDto := range fieldDtos {
		field, err := m.ConvertFieldDTOToDomain(modelID, modelLocator, fieldDto)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

// applyEnumBindingByFormat 应用 model-enum 参数容错矩阵。
// - ENUM:       仅使用 relateEnumName
// - 其他类型:    忽略 enum 绑定参数
func applyEnumBindingByFormat(target *modeldesign.FieldDefinition, input *dtos.FieldDefinitionDTO) {
	target.EnumName = ""

	switch input.Format {
	case modeldesign.FormatEnum, modeldesign.FormatEnumArray:
		if input.RelateEnumName != nil {
			target.EnumName = *input.RelateEnumName
		}
	}
}
