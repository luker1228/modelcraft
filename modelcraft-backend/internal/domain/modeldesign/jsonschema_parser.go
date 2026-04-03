package modeldesign

import (
	"context"
	"encoding/json"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"time"

	"github.com/spf13/cast"
)

// JSONSchemaParser 解析JSON Schema的域服务
type JSONSchemaParser struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewJSONSchemaParser 创建JSON Schema解析器实例
func NewJSONSchemaParser(ctx context.Context) *JSONSchemaParser {
	return &JSONSchemaParser{
		ctx:    ctx,
		logger: logfacade.GetLogger(ctx),
	}
}

// ParseSchema 从JSON Schema Draft 7解析为DataModel
func (p *JSONSchemaParser) ParseSchema(schemaJSON string) (*DataModel, error) {
	return p.ParseSchemaWithModelInfo(schemaJSON, "", "")
}

// ParseSchemaWithModelInfo 从JSON Schema Draft 7解析为DataModel，支持传入模型信息
func (p *JSONSchemaParser) ParseSchemaWithModelInfo(
	schemaJSON, modelName, databaseName string,
) (*DataModel, error) {
	// 解析JSON
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schemaMap); err != nil {
		return nil, bizerrors.Wrapf(err, "Failed to parse JSON Schema: invalid JSON")
	}

	// 验证schema版本
	if err := p.validateSchemaVersion(schemaMap); err != nil {
		return nil, err
	}

	// 提取模型元数据
	model, err := p.extractModelMetadataWithInfo(schemaMap, modelName, databaseName)
	if err != nil {
		return nil, err
	}

	// 解析字段定义
	fields, err := p.parseFields(schemaMap, model.GetModelLocator())
	if err != nil {
		return nil, err
	}

	model.Fields = fields

	return model, nil
}

// validateSchemaVersion 验证schema版本
func (p *JSONSchemaParser) validateSchemaVersion(schemaMap map[string]interface{}) error {
	rawSchema, ok := schemaMap["$schema"]
	if !ok || rawSchema == nil {
		return bizerrors.New("Missing or invalid '$schema' field")
	}
	schema, err := cast.ToStringE(rawSchema)
	if err != nil || schema == "" {
		return bizerrors.New("Missing or invalid '$schema' field")
	}
	if schema != "http://json-schema.org/draft-07/schema#" {
		return bizerrors.Errorf("Unsupported schema version: %s (only Draft 7 is supported)", schema)
	}
	return nil
}

// extractModelMetadataWithInfo 提取模型元数据，支持从外部参数获取模型信息
func (p *JSONSchemaParser) extractModelMetadataWithInfo(
	schemaMap map[string]interface{},
	modelName, databaseName string,
) (*DataModel, error) {
	// 验证type
	schemaType, err := cast.ToStringE(schemaMap["type"])
	if err != nil {
		return nil, bizerrors.New("Schema must have 'type' property as string")
	}
	if schemaType != "object" {
		return nil, bizerrors.Errorf("Schema type must be 'object', got '%s'", schemaType)
	}

	// 提取必需的元数据
	title := cast.ToString(schemaMap["title"])
	if title == "" {
		return nil, bizerrors.New("Required metadata 'title' is missing or not a string")
	}

	// 优先使用外部传入的模型信息，如果外部没有提供，则尝试从schema中获取
	if modelName == "" {
		modelName = cast.ToString(schemaMap["x-modelName"])
	}
	if databaseName == "" {
		databaseName = cast.ToString(schemaMap["x-databaseName"])
	}

	// 提取可选元数据
	description := cast.ToString(schemaMap["description"])

	// 生成模型ID
	id, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.Wrapf(err, "Failed to generate model ID")
	}

	now := time.Now()
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID: id,
			ModelLocator: ModelLocator{
				ModelName:    modelName,
				DatabaseName: databaseName,
			},
			Title:       title,
			Description: description,
			StorageType: "table", // 默认为table
			Version:     1,
			Status:      "draft",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Fields: []*FieldDefinition{},
	}

	return model, nil
}

// parseFields 解析字段定义
func (p *JSONSchemaParser) parseFields(
	schemaMap map[string]interface{},
	modelLocator *ModelLocator,
) ([]*FieldDefinition, error) {
	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok || len(properties) == 0 {
		return nil, bizerrors.New("Schema must have 'properties' with at least one field")
	}

	// 获取required列表
	requiredFields := p.extractRequiredFields(schemaMap)

	fields := make([]*FieldDefinition, 0, len(properties))
	for fieldName, fieldSchema := range properties {
		fieldMap, ok := fieldSchema.(map[string]interface{})
		if !ok {
			p.logger.Warnf(p.ctx, "Skipping field '%s': invalid field schema", fieldName)
			continue
		}

		// 跳过relation字段
		if _, hasRelation := fieldMap["x-relation"]; hasRelation {
			p.logger.Warnf(p.ctx, "Skipping field '%s': relation fields are not supported in schema import", fieldName)
			continue
		}

		field, err := p.parseField(fieldName, fieldMap, modelLocator, requiredFields)
		if err != nil {
			return nil, bizerrors.Wrapf(err, "Failed to parse field '%s'", fieldName)
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// extractRequiredFields 提取required字段列表
func (p *JSONSchemaParser) extractRequiredFields(schemaMap map[string]interface{}) map[string]bool {
	requiredList, ok := schemaMap["required"].([]interface{})
	if !ok {
		return map[string]bool{}
	}

	requiredMap := make(map[string]bool, len(requiredList))
	for _, field := range requiredList {
		requiredMap[cast.ToString(field)] = true
	}
	return requiredMap
}

// parseField 解析单个字段
func (p *JSONSchemaParser) parseField(
	fieldName string,
	fieldMap map[string]interface{},
	modelLocator *ModelLocator,
	requiredFields map[string]bool,
) (*FieldDefinition, error) {
	// 生成字段ID (使用空字符串，将由repository设置)
	now := time.Now()

	// 提取基本信息
	title := cast.ToString(fieldMap["title"])
	if title == "" {
		title = fieldName // 默认使用字段名作为标题
	}
	description := cast.ToString(fieldMap["description"])

	// 解析类型
	fieldType, err := p.parseFieldType(fieldMap)
	if err != nil {
		return nil, err
	}

	// 解析验证规则
	validation := p.parseValidationConfig(fieldMap, fieldType)

	// 解析nullable
	nonNull := !cast.ToBool(fieldMap["nullable"])

	// 解析required
	required := requiredFields[fieldName]

	// 解析自定义属性
	displayOrder := cast.ToString(fieldMap["x-displayOrder"])
	isPrimary := cast.ToBool(fieldMap["x-isPrimary"])
	isUnique := cast.ToBool(fieldMap["x-isUnique"])

	var storageHint *string
	if hint := cast.ToString(fieldMap["x-storageHint"]); hint != "" {
		storageHint = &hint
	}

	field := &FieldDefinition{
		ModelID:      "", // 将由应用层设置
		Name:         fieldName,
		ModelLocator: modelLocator,
		Title:        title,
		Description:  description,
		Type:         fieldType,
		StorageHint:  storageHint,
		NonNull:      nonNull,
		Required:     required,
		IsUnique:     isUnique,
		IsPrimary:    isPrimary,
		Status:       FieldStatusInit,
		Validation:   validation,
		DisplayOrder: displayOrder,
		Metadata:     map[string]any{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return field, nil
}

// parseFieldType 解析字段类型
func (p *JSONSchemaParser) parseFieldType(fieldMap map[string]interface{}) (*FieldType, error) {
	schemaType, err := cast.ToStringE(fieldMap["type"])
	if err != nil {
		return nil, bizerrors.New("Field must have 'type' property")
	}

	format := cast.ToString(fieldMap["format"])

	var fieldFormat FormatType

	switch schemaType {
	case "string":
		fieldFormat = p.parseStringType(format, fieldMap)
	case "integer":
		fieldFormat = FormatInteger
	case "number":
		// 检查是否是decimal类型
		if _, hasPrecision := fieldMap["x-precision"]; hasPrecision {
			fieldFormat = FormatDecimal
		} else {
			fieldFormat = FormatNumber
		}
	case "boolean":
		fieldFormat = FormatBoolean
	case "array":
		// 检查items类型来确定是否是枚举数组
		if items, ok := fieldMap["items"].(map[string]interface{}); ok {
			if _, hasEnum := items["enum"]; hasEnum {
				fieldFormat = FormatEnumArray
			} else {
				return nil, bizerrors.Errorf("Unsupported array type (only enum arrays are supported)")
			}
		} else {
			return nil, bizerrors.Errorf("Array type must have 'items' definition")
		}
	default:
		return nil, bizerrors.Errorf("Unsupported JSON Schema type '%s' for field", schemaType)
	}

	// 使用GetFieldTypeByFormat获取完整的FieldType
	fieldType := GetFieldTypeByFormat(fieldFormat)
	if fieldType == nil {
		return nil, bizerrors.Errorf("Unsupported field format: %s", fieldFormat)
	}

	return fieldType, nil
}

// parseStringType 解析字符串类型的format
func (p *JSONSchemaParser) parseStringType(format string, fieldMap map[string]interface{}) FormatType {
	// 检查是否有enum
	if _, hasEnum := fieldMap["enum"]; hasEnum {
		return FormatEnum
	}

	switch format {
	case "uuid":
		return FormatUUID
	case "date":
		return FormatDate
	case "date-time":
		return FormatDateTime
	case "time":
		return FormatTime
	default:
		return FormatString
	}
}

// parseValidationConfig 解析验证配置
func (p *JSONSchemaParser) parseValidationConfig(
	fieldMap map[string]interface{}, fieldType *FieldType,
) *ValidationConfig {
	validation := &ValidationConfig{}

	// String validation
	if val := cast.ToInt(fieldMap["maxLength"]); val > 0 {
		validation.MaxLength = &val
	}
	if val := cast.ToInt(fieldMap["minLength"]); val > 0 {
		validation.MinLength = &val
	}
	if pattern := cast.ToString(fieldMap["pattern"]); pattern != "" {
		validation.Pattern = &pattern
	}

	// Number validation
	if _, hasMax := fieldMap["maximum"]; hasMax {
		val := cast.ToFloat64(fieldMap["maximum"])
		validation.Maximum = &val
	}
	if _, hasMin := fieldMap["minimum"]; hasMin {
		val := cast.ToFloat64(fieldMap["minimum"])
		validation.Minimum = &val
	}

	// Array validation
	if val := cast.ToInt(fieldMap["maxItems"]); val > 0 {
		validation.MaxItems = &val
	}
	if val := cast.ToInt(fieldMap["minItems"]); val > 0 {
		validation.MinItems = &val
	}

	// Date/Time validation (custom properties)
	if minDate := cast.ToString(fieldMap["x-minDate"]); minDate != "" {
		validation.MinDate = &minDate
	}
	if maxDate := cast.ToString(fieldMap["x-maxDate"]); maxDate != "" {
		validation.MaxDate = &maxDate
	}
	if minTime := cast.ToString(fieldMap["x-minTime"]); minTime != "" {
		validation.MinTime = &minTime
	}
	if maxTime := cast.ToString(fieldMap["x-maxTime"]); maxTime != "" {
		validation.MaxTime = &maxTime
	}

	// Decimal precision/scale (custom properties)
	if precision := cast.ToInt(fieldMap["x-precision"]); precision > 0 {
		validation.Precision = &precision
	}
	if scale := cast.ToInt(fieldMap["x-scale"]); scale > 0 {
		validation.Scale = &scale
	}

	// Enum values
	switch fieldType.Format {
	case FormatEnum:
		validation.EnumValues = p.extractEnumValuesFromField(fieldMap)
	case FormatEnumArray:
		validation.EnumValues = p.extractEnumValuesFromArrayItems(fieldMap)
	}

	// Validation rule type
	if rule := cast.ToString(fieldMap["x-validateRule"]); rule != "" {
		validation.Rule = ValidateRule(rule)
	}

	return validation
}

func (p *JSONSchemaParser) extractEnumValuesFromField(fieldMap map[string]interface{}) []string {
	enum, ok := fieldMap["enum"].([]interface{})
	if !ok {
		return nil
	}
	return p.extractEnumValues(enum)
}

func (p *JSONSchemaParser) extractEnumValuesFromArrayItems(fieldMap map[string]interface{}) []string {
	items, ok := fieldMap["items"].(map[string]interface{})
	if !ok {
		return nil
	}
	enum, ok := items["enum"].([]interface{})
	if !ok {
		return nil
	}
	return p.extractEnumValues(enum)
}

// extractEnumValues 提取枚举值
func (p *JSONSchemaParser) extractEnumValues(enumArray []interface{}) []string {
	values := make([]string, 0, len(enumArray))
	for _, val := range enumArray {
		if strVal := cast.ToString(val); strVal != "" {
			values = append(values, strVal)
		}
	}
	return values
}

// ParseSchemaWithLogger 使用指定logger解析schema
func (p *JSONSchemaParser) ParseSchemaWithLogger(schemaJSON string, logger logfacade.Logger) (*DataModel, error) {
	p.logger = logger
	return p.ParseSchema(schemaJSON)
}

// ParseSchemaWithLoggerAndModelInfo 使用指定logger解析schema，支持传入模型信息
func (p *JSONSchemaParser) ParseSchemaWithLoggerAndModelInfo(
	schemaJSON string,
	logger logfacade.Logger,
	modelName, databaseName string,
) (*DataModel, error) {
	p.logger = logger
	return p.ParseSchemaWithModelInfo(schemaJSON, modelName, databaseName)
}
