package adapter

import (
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/internal/interfaces/http/dtos"
)

// FieldMapper is the singleton field mapper for project domain
var FieldMapper *fieldMapper

func init() {
	FieldMapper = &fieldMapper{}
}

type fieldMapper struct{}

func convertFormatType2Domain(format generated.FormatType) (modeldesign.FormatType, error) {
	switch format {
	case generated.FormatTypeString:
		return modeldesign.FormatString, nil
	case generated.FormatTypeInteger:
		return modeldesign.FormatInteger, nil
	case generated.FormatTypeNumber:
		return modeldesign.FormatNumber, nil
	case generated.FormatTypeBoolean:
		return modeldesign.FormatBoolean, nil
	case generated.FormatTypeDatetime:
		return modeldesign.FormatDateTime, nil
	case generated.FormatTypeDate:
		return modeldesign.FormatDate, nil
	case generated.FormatTypeTime:
		return modeldesign.FormatTime, nil
	case generated.FormatTypeUUID:
		return modeldesign.FormatUUID, nil
	case generated.FormatTypeDecimal:
		return modeldesign.FormatDecimal, nil
	case generated.FormatTypeRelation:
		return modeldesign.FormatRelation, nil
	case generated.FormatTypeEnum:
		return modeldesign.FormatEnum, nil
	case generated.FormatTypeEnumLabel:
		return modeldesign.FormatEnumLabel, nil
	default:
		return modeldesign.FormatString, fmt.Errorf("unknown format type: %s", format)
	}
}

// ConvertAddFieldInputToDTO converts GraphQL AddFieldInput to DTO
func (m *fieldMapper) ConvertAddFieldInputToDTO(input *generated.AddFieldInput) (*dtos.FieldDefinitionDTO, error) {
	format, err := convertFormatType2Domain(input.Format)
	if err != nil {
		return nil, err
	}
	dto := dtos.FieldDefinitionDTO{
		Name:   input.Name, // Use name as key
		Title:  input.Title,
		Format: format,
	}
	if input.NonNull == nil {
		dto.NonNull = false // Default to non-null
	} else {
		dto.NonNull = *input.NonNull
	}

	if input.Required == nil {
		dto.Required = false // Default to not required
	} else {
		dto.Required = *input.Required
	}

	if input.IsUnique == nil {
		dto.IsUnique = false // Default to not unique
	} else {
		dto.IsUnique = *input.IsUnique
	}

	if input.IsArray == nil {
		dto.IsArray = false // Default to single-select
	} else {
		dto.IsArray = *input.IsArray
	}

	if input.Description != nil {
		dto.Description = *input.Description
	}

	// model-enum 容错矩阵：按 format 仅保留有效参数
	switch format {
	case modeldesign.FormatEnum, modeldesign.FormatEnumArray:
		if input.RelateEnumName != nil {
			dto.RelateEnumName = input.RelateEnumName
		}
	case modeldesign.FormatEnumLabel:
		if input.EnumRelationID != nil {
			dto.EnumRelationID = input.EnumRelationID
		}
	}

	// Convert validate props
	if input.ValidationConfig != nil {
		dto.ValidationConfig = convertValidationInput2DTO(input.ValidationConfig)
	}

	// Convert relateFkId for RELATION-format fields
	if input.RelateFkID != nil {
		dto.RelateFKID = input.RelateFkID
	}

	if input.StorageHint != nil {
		dto.StorageHint = input.StorageHint
	}

	return &dto, nil
}

func convertValidationInput2DTO(input *generated.ValidationConfigInput) *dtos.ValidationConfigDTO {
	if input == nil {
		return nil
	}
	validation := &dtos.ValidationConfigDTO{}
	if input.MinLength != nil {
		minLength := int(*input.MinLength)
		validation.MinLength = &minLength
	}
	if input.MaxLength != nil {
		maxLength := int(*input.MaxLength)
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
	return validation
}

// ConvertValidationInputToDomain converts GraphQL ValidationConfigInput directly to domain ValidationConfig
func (m *fieldMapper) ConvertValidationInputToDomain(
	input *generated.ValidationConfigInput,
) *modeldesign.ValidationConfig {
	if input == nil {
		return nil
	}
	validation := &modeldesign.ValidationConfig{}
	if input.MinLength != nil {
		minLength := int(*input.MinLength)
		validation.MinLength = &minLength
	}
	if input.MaxLength != nil {
		maxLength := int(*input.MaxLength)
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
	return validation
}

// ConvertValidationDTO2Domain converts DTO ValidationConfig to domain ValidationConfig
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

// ConvertUpdateFieldInputToDTO converts GraphQL UpdateFieldInput to DTO
func (m *fieldMapper) ConvertUpdateFieldInputToDTO(
	fieldName string,
	input generated.UpdateFieldInput,
) dtos.FieldDefinitionDTO {
	dto := dtos.FieldDefinitionDTO{
		Name: fieldName,
	}

	if input.Title != nil {
		dto.Title = *input.Title
	}

	if input.Description != nil {
		dto.Description = *input.Description
	}

	// Convert validate props
	validation := convertValidationInput2DTO(input.ValidationConfig)
	if validation != nil {
		dto.ValidationConfig = validation
	}
	return dto
}
