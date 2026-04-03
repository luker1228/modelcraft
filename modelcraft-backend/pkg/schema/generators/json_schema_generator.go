package generators

import (
	"encoding/json"
	"fmt"
	"modelcraft/pkg/schema/core"
)

// JSONSchemaBuilder JSON Schema构建器
type JSONSchemaBuilder struct {
	schema map[string]interface{}
}

// NewJSONSchemaBuilder 创建新的JSON Schema构建器
func NewJSONSchemaBuilder() *JSONSchemaBuilder {
	return &JSONSchemaBuilder{
		schema: make(map[string]interface{}),
	}
}

// SetBasicInfo 设置基本信息
func (b *JSONSchemaBuilder) SetBasicInfo(name, description string) *JSONSchemaBuilder {
	b.schema["$modeldesign"] = "http://json-modeldesign.org/draft-07/modeldesign#"
	b.schema["type"] = "object"
	b.schema["title"] = name
	b.schema["description"] = description
	b.schema["additionalProperties"] = false
	return b
}

// AddField 添加字段
func (b *JSONSchemaBuilder) AddField(field *core.FieldDefinition) *JSONSchemaBuilder {
	if b.schema["properties"] == nil {
		b.schema["properties"] = make(map[string]interface{})
	}

	properties, ok := b.schema["properties"].(map[string]interface{})
	if !ok {
		properties = make(map[string]interface{})
		b.schema["properties"] = properties
	}
	fieldSchema := b.buildFieldSchema(field)
	properties[field.Key] = fieldSchema

	return b
}

// SetRequiredFields 设置必填字段
func (b *JSONSchemaBuilder) SetRequiredFields(requiredFields []string) *JSONSchemaBuilder {
	if len(requiredFields) > 0 {
		b.schema["required"] = requiredFields
	}
	return b
}

// Build 构建Schema
func (b *JSONSchemaBuilder) Build() map[string]interface{} {
	return b.schema
}

// buildFieldSchema 构建字段Schema
func (b *JSONSchemaBuilder) buildFieldSchema(field *core.FieldDefinition) map[string]interface{} {
	fieldSchema := map[string]interface{}{
		"type":        field.Type.GetJSONType(),
		"description": field.Description,
	}

	// 添加默认值
	if field.Default != nil {
		fieldSchema["default"] = field.Default
	}

	// 添加验证规则
	if field.Validation != nil {
		b.addValidationRules(fieldSchema, field.Validation)
	}

	return fieldSchema
}

// addValidationRules 添加验证规则
func (b *JSONSchemaBuilder) addValidationRules(fieldSchema map[string]interface{}, validation *core.ValidationRules) {
	if validation.MinLength != nil {
		fieldSchema["minLength"] = *validation.MinLength
	}
	if validation.MaxLength != nil {
		fieldSchema["maxLength"] = *validation.MaxLength
	}
	if validation.Pattern != nil {
		fieldSchema["pattern"] = *validation.Pattern
	}
	if validation.Minimum != nil {
		fieldSchema["minimum"] = *validation.Minimum
	}
	if validation.Maximum != nil {
		fieldSchema["maximum"] = *validation.Maximum
	}
}

// JSONSchemaGenerator JSON Schema生成器
type JSONSchemaGenerator struct {
	builder *JSONSchemaBuilder
}

// NewJSONSchemaGenerator 创建新的JSON Schema生成器
func NewJSONSchemaGenerator() *JSONSchemaGenerator {
	return &JSONSchemaGenerator{
		builder: NewJSONSchemaBuilder(),
	}
}

// GenerateJSONSchema 生成JSON Schema
func (g *JSONSchemaGenerator) GenerateJSONSchema(model *core.ModelDefinition) (*core.JSONSchemaResult, error) {
	if !model.IsValid() {
		return nil, fmt.Errorf("invalid modeldesign definition: %s", model.Name)
	}

	// 1. 重置构建器
	g.builder = NewJSONSchemaBuilder()

	// 2. 设置基本信息
	g.builder.SetBasicInfo(model.Name, model.Description)

	// 3. 添加字段
	requiredFields := model.GetRequiredFields()
	for _, field := range model.Fields {
		g.builder.AddField(field)
	}

	// 4. 设置必填字段
	g.builder.SetRequiredFields(requiredFields)

	// 5. 构建Schema
	schema := g.builder.Build()

	// 6. 生成JSON字符串
	rawBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON modeldesign: %w", err)
	}

	return &core.JSONSchemaResult{
		Schema: schema,
		Raw:    string(rawBytes),
	}, nil
}
