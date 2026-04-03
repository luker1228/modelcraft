package modeldesign

import (
	"encoding/json"
)

// JSONSchemaGenerator 生成JSON Schema的域服务
type JSONSchemaGenerator struct{}

// NewJSONSchemaGenerator 创建JSON Schema生成器实例
func NewJSONSchemaGenerator() *JSONSchemaGenerator {
	return &JSONSchemaGenerator{}
}

// GenerateSchema 从DataModel生成JSON Schema Draft 7
func (g *JSONSchemaGenerator) GenerateSchema(model *DataModel) (string, error) {
	schema := g.buildSchema(model)

	// 序列化为JSON字符串
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// buildSchema 构建JSON Schema对象
func (g *JSONSchemaGenerator) buildSchema(model *DataModel) map[string]interface{} {
	schema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"type":        string(SchemaTypeObject),
		"title":       model.Title,
		"description": model.Description,
		"properties":  g.buildProperties(model.Fields),
		"required":    g.buildRequiredList(model.Fields),
		// Custom ModelCraft metadata
		"x-modelName":    model.ModelName,
		"x-databaseName": model.DatabaseName,
	}

	return schema
}

// buildProperties 构建properties对象
func (g *JSONSchemaGenerator) buildProperties(fields []*FieldDefinition) map[string]interface{} {
	properties := make(map[string]interface{})

	for _, field := range fields {
		properties[field.Name] = g.buildFieldSchema(field)
	}

	return properties
}

// buildFieldSchema 为单个字段构建JSON Schema
func (g *JSONSchemaGenerator) buildFieldSchema(field *FieldDefinition) map[string]interface{} {
	fieldSchema := map[string]interface{}{
		"title":       field.Title,
		"description": field.Description,
	}

	// 基本类型和格式映射
	g.applyTypeAndFormat(fieldSchema, field)

	// 应用验证规则
	g.applyValidationRules(fieldSchema, field)

	// 处理nullable
	if !field.NonNull {
		fieldSchema["nullable"] = true
	}

	// 添加自定义ModelCraft属性
	g.applyCustomProperties(fieldSchema, field)

	return fieldSchema
}

// applyTypeAndFormat 应用JSON Schema type和format
func (g *JSONSchemaGenerator) applyTypeAndFormat(schema map[string]interface{}, field *FieldDefinition) {
	if field.Type == nil {
		return
	}

	switch field.Type.Format {
	// String types
	case FormatString:
		schema["type"] = string(SchemaTypeString)
	case FormatUUID:
		schema["type"] = string(SchemaTypeString)
		schema["format"] = "uuid"
	case FormatDate:
		schema["type"] = string(SchemaTypeString)
		schema["format"] = "date"
	case FormatDateTime:
		schema["type"] = string(SchemaTypeString)
		schema["format"] = "date-time"
	case FormatTime:
		schema["type"] = string(SchemaTypeString)
		schema["format"] = "time"

	// Number types
	case FormatNumber:
		schema["type"] = string(SchemaTypeNumber)
	case FormatInteger:
		schema["type"] = "integer"
	case FormatDecimal:
		schema["type"] = string(SchemaTypeNumber)
		// Precision and scale will be added as custom properties

	// Boolean type
	case FormatBoolean:
		schema["type"] = string(SchemaTypeBoolean)

	// Enum types
	case FormatEnum:
		schema["type"] = string(SchemaTypeString)
		g.applyEnumOptions(schema, field)
	case FormatEnumArray:
		schema["type"] = string(SchemaTypeArray)
		items := map[string]interface{}{
			"type": string(SchemaTypeString),
		}
		g.applyEnumOptions(items, field)
		schema["items"] = items

	// Relation type
	case FormatRelation:
		schema["type"] = string(SchemaTypeObject)
	}
}

// applyEnumOptions 应用枚举选项
func (g *JSONSchemaGenerator) applyEnumOptions(schema map[string]interface{}, field *FieldDefinition) {
	// 从ValidationConfig.EnumValues或Enum获取枚举选项
	if field.Validation != nil && len(field.Validation.EnumValues) > 0 {
		// 使用简单枚举值
		schema["enum"] = field.Validation.EnumValues
	} else if field.Enum != nil {
		// 使用完整枚举定义，提取codes作为enum值
		codes := make([]string, len(field.Enum.Options))
		for i, opt := range field.Enum.Options {
			codes[i] = opt.Code
		}
		schema["enum"] = codes

		// 添加完整枚举信息到x-enum
		schema["x-enum"] = g.buildEnumMetadata(field.Enum)
	}
}

// buildEnumMetadata 构建枚举元数据
func (g *JSONSchemaGenerator) buildEnumMetadata(enum *EnumDefinition) map[string]interface{} {
	options := make([]map[string]interface{}, len(enum.Options))
	for i, opt := range enum.Options {
		options[i] = map[string]interface{}{
			"code":        opt.Code,
			"label":       opt.Label,
			"description": opt.Description,
		}
	}

	return map[string]interface{}{
		"name":          enum.Name,
		"displayName":   enum.DisplayName,
		"description":   enum.Description,
		"isMultiSelect": enum.IsMultiSelect,
		"options":       options,
	}
}

// applyValidationRules 应用验证规则
func (g *JSONSchemaGenerator) applyValidationRules(schema map[string]interface{}, field *FieldDefinition) {
	if field.Validation == nil {
		return
	}

	v := field.Validation

	// String validation
	if v.MaxLength != nil {
		schema["maxLength"] = *v.MaxLength
	}
	if v.MinLength != nil {
		schema["minLength"] = *v.MinLength
	}
	if v.Pattern != nil {
		schema["pattern"] = *v.Pattern
	}

	// Number validation
	if v.Maximum != nil {
		schema["maximum"] = *v.Maximum
	}
	if v.Minimum != nil {
		schema["minimum"] = *v.Minimum
	}

	// Array validation
	if v.MaxItems != nil {
		schema["maxItems"] = *v.MaxItems
	}
	if v.MinItems != nil {
		schema["minItems"] = *v.MinItems
	}

	// Date/Time validation (custom properties)
	if v.MinDate != nil {
		schema["x-minDate"] = *v.MinDate
	}
	if v.MaxDate != nil {
		schema["x-maxDate"] = *v.MaxDate
	}
	if v.MinTime != nil {
		schema["x-minTime"] = *v.MinTime
	}
	if v.MaxTime != nil {
		schema["x-maxTime"] = *v.MaxTime
	}

	// Decimal precision/scale (custom properties)
	if v.Precision != nil {
		schema["x-precision"] = *v.Precision
	}
	if v.Scale != nil {
		schema["x-scale"] = *v.Scale
	}

	// Validation rule type
	if v.Rule != "" {
		schema["x-validateRule"] = string(v.Rule)
	}
}

// applyCustomProperties 添加ModelCraft自定义属性
func (g *JSONSchemaGenerator) applyCustomProperties(schema map[string]interface{}, field *FieldDefinition) {
	// Field metadata
	if field.StorageHint != nil {
		schema["x-storageHint"] = *field.StorageHint
	}
	schema["x-displayOrder"] = field.DisplayOrder
	schema["x-isPrimary"] = field.IsPrimary
	schema["x-isUnique"] = field.IsUnique

	// FK metadata
	if field.RelateFKID != nil {
		schema["x-relateFkId"] = *field.RelateFKID
	}
	if field.BelongsToFKID != nil {
		schema["x-belongsToFkId"] = *field.BelongsToFKID
	}

	// Mark readOnly for fields that cannot be edited by the user:
	//   - Primary key fields are generated by the database
	//   - RELATION fields are derived and displayed read-only in tables
	if field.IsPrimary || (field.Type != nil && field.Type.Format == FormatRelation) {
		schema["readOnly"] = true
	}
}

// buildRequiredList 构建required字段列表
func (g *JSONSchemaGenerator) buildRequiredList(fields []*FieldDefinition) []string {
	required := []string{}

	for _, field := range fields {
		if field.Required {
			required = append(required, field.Name)
		}
	}

	return required
}
