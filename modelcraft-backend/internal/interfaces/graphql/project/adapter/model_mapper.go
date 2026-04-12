package adapter

import (
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"time"
)

// ModelMapper is the singleton model mapper for project domain
var ModelMapper = &modelMapper{}

type modelMapper struct{}

// ConvertToGraphQLModel converts domain DataModel to GraphQL Model without actual schema data.
func (m *modelMapper) ConvertToGraphQLModel(modelEntity *modeldesign.DataModel) (*generated.Model, error) {
	return m.ConvertToGraphQLModelWithActualSchema(modelEntity, nil)
}

// ConvertToGraphQLModelWithActualSchema converts domain DataModel to GraphQL Model.
// When actualResult is non-nil, dbTable and dbColumn fields are populated from the actual schema data.
func (m *modelMapper) ConvertToGraphQLModelWithActualSchema(
	modelEntity *modeldesign.DataModel,
	actualResult *modeldesign.ActualSchemaResult,
) (*generated.Model, error) {
	if modelEntity == nil {
		return nil, fmt.Errorf("model is nil")
	}
	graphqlFields := make([]*generated.Field, 0, len(modelEntity.Fields))

	for _, domainField := range modelEntity.Fields {
		graphqlField, err := m.convertFieldWithActualSchema(*domainField, actualResult)
		if err != nil {
			return nil, err
		}
		graphqlFields = append(graphqlFields, graphqlField)
	}

	gqlModel := &generated.Model{
		ID:           modelEntity.ID,
		ProjectSlug:  modelEntity.ProjectSlug,
		Name:         modelEntity.ModelName,
		Title:        modelEntity.Title,
		Description:  modelEntity.Description,
		DatabaseName: modelEntity.DatabaseName,
		StorageType:  modelEntity.StorageType,
		DisplayField: modelEntity.DisplayField,
		Fields:       graphqlFields,
		Group:        GroupMapper.GroupPlaceholder(modelEntity.GroupID),
		CreatedAt:    modelEntity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    modelEntity.UpdatedAt.Format(time.RFC3339),
	}

	if actualResult != nil {
		dbTable := convertDbTableStatus(actualResult.Status)
		gqlModel.DbTable = &dbTable
	}

	return gqlModel, nil
}

// ConvertFieldToGraphQLField converts domain FieldDefinitionDTO to GraphQL Field
func (m *modelMapper) ConvertFieldToGraphQLField(fieldDef modeldesign.FieldDefinition) (*generated.Field, error) {
	return m.convertFieldWithActualSchema(fieldDef, nil)
}

// convertFieldWithActualSchema converts a domain field to GraphQL, optionally filling dbColumn
// from the provided actual schema result.
func (m *modelMapper) convertFieldWithActualSchema(
	fieldDef modeldesign.FieldDefinition,
	actualResult *modeldesign.ActualSchemaResult,
) (*generated.Field, error) {
	var description *string
	if fieldDef.Description != "" {
		description = &fieldDef.Description
	}

	var validationConfig *generated.ValidationConfig
	if fieldDef.Validation != nil {
		validationConfig = convertValidationDomain2Graphql(fieldDef.Validation)
	}

	format := convertFormatTypeDomain2Model(fieldDef.Type.Format)
	if format == "" {
		return nil, fmt.Errorf("invalid format: %s", fieldDef.Type.Format)
	}
	schemaType := fieldDef.Type.SchemaType

	field := &generated.Field{
		Name:             fieldDef.Name,
		Title:            fieldDef.Title,
		Format:           format,
		SchemaType:       generated.SchemaType(schemaType),
		NonNull:          fieldDef.NonNull,
		Required:         fieldDef.Required,
		IsUnique:         fieldDef.IsUnique,
		IsPrimary:        fieldDef.IsPrimary,
		IsDeprecated:     fieldDef.IsDeprecated,
		IsArray:          fieldDef.IsArray,
		Description:      description,
		ValidationConfig: validationConfig,
		RelateFkID:       fieldDef.RelateFKID,
		BelongsToFkID:    fieldDef.BelongsToFKID,
		CreatedAt:        fieldDef.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        fieldDef.UpdatedAt.Format(time.RFC3339),
	}

	// Virtual fields (ENUM_LABEL) never have a db column.
	if actualResult != nil && actualResult.Status == modeldesign.DbTableExists && !fieldDef.IsEnumLabelField() {
		if colInfo, ok := actualResult.Fields[fieldDef.Name]; ok {
			field.DbColumn = convertDbColumnInfo(colInfo)
		}
	}

	return field, nil
}

// convertDbTableStatus maps domain DbTableStatus to generated DbTableStatus.
func convertDbTableStatus(status modeldesign.DbTableStatus) generated.DbTableStatus {
	switch status {
	case modeldesign.DbTableExists:
		return generated.DbTableStatusTableExists
	case modeldesign.DbTableMissing:
		return generated.DbTableStatusTableMissing
	default:
		return generated.DbTableStatusClusterUnreachable
	}
}

// convertDbColumnInfo maps domain DbColumnInfo to generated DbColumnInfo.
func convertDbColumnInfo(col *modeldesign.DbColumnInfo) *generated.DbColumnInfo {
	if col == nil {
		return nil
	}

	constraints := make([]generated.ActualConstraintType, 0, len(col.Constraints))
	for _, c := range col.Constraints {
		switch c {
		case modeldesign.ActualConstraintUnique:
			constraints = append(constraints, generated.ActualConstraintTypeUnique)
		case modeldesign.ActualConstraintNotNull:
			constraints = append(constraints, generated.ActualConstraintTypeNotNull)
		}
	}

	conflicts := make([]*generated.FieldConflict, 0, len(col.Conflicts))
	for _, fc := range col.Conflicts {
		aspect := convertFieldConflictAspect(fc.Aspect)
		conflicts = append(conflicts, &generated.FieldConflict{
			Aspect:   aspect,
			Expected: fc.Expected,
			Actual:   fc.Actual,
		})
	}

	return &generated.DbColumnInfo{
		ColumnType:   col.ColumnType,
		Unique:       col.Unique,
		NonNull:      col.NonNull,
		IsPrimaryKey: col.IsPrimaryKey,
		DefaultValue: col.DefaultValue,
		Constraints:  constraints,
		ForeignKey:   convertActualForeignKey(col.ForeignKey),
		Conflicts:    conflicts,
	}
}

// convertActualForeignKey maps domain ActualForeignKey to generated ActualForeignKey.
func convertActualForeignKey(fk *modeldesign.ActualForeignKey) *generated.ActualForeignKey {
	if fk == nil {
		return nil
	}
	return &generated.ActualForeignKey{
		ReferencedTable:  fk.ReferencedTable,
		ReferencedColumn: fk.ReferencedColumn,
		ConstraintName:   fk.ConstraintName,
	}
}

// convertFieldConflictAspect maps domain FieldConflictAspect to generated FieldConflictAspect.
func convertFieldConflictAspect(aspect modeldesign.FieldConflictAspect) generated.FieldConflictAspect {
	switch aspect {
	case modeldesign.FieldConflictPrimaryMismatch:
		return generated.FieldConflictAspectPrimaryMismatch
	case modeldesign.FieldConflictUniqueMismatch:
		return generated.FieldConflictAspectUniqueMismatch
	default:
		return generated.FieldConflictAspectNotNullMismatch
	}
}

func convertValidationDomain2Graphql(validation *modeldesign.ValidationConfig) *generated.ValidationConfig {
	return &generated.ValidationConfig{
		MinLength: validation.MinLength,
		MaxLength: validation.MaxLength,
		Pattern:   validation.Pattern,
		Minimum:   validation.Minimum,
		Maximum:   validation.Maximum,
	}
}

func convertFormatTypeDomain2Model(format modeldesign.FormatType) generated.FormatType {
	switch format {
	case modeldesign.FormatString:
		return generated.FormatTypeString
	case modeldesign.FormatInteger:
		return generated.FormatTypeInteger
	case modeldesign.FormatNumber:
		return generated.FormatTypeNumber
	case modeldesign.FormatBoolean:
		return generated.FormatTypeBoolean
	case modeldesign.FormatDateTime:
		return generated.FormatTypeDatetime
	case modeldesign.FormatDate:
		return generated.FormatTypeDate
	case modeldesign.FormatTime:
		return generated.FormatTypeTime
	case modeldesign.FormatUUID:
		return generated.FormatTypeUUID
	case modeldesign.FormatDecimal:
		return generated.FormatTypeDecimal
	case modeldesign.FormatRelation:
		return generated.FormatTypeRelation
	case modeldesign.FormatEnum:
		return generated.FormatTypeEnum
	case modeldesign.FormatEnumLabel:
		return generated.FormatTypeEnumLabel
	default:
		return ""
	}
}
