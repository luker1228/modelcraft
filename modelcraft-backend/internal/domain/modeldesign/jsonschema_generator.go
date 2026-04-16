package modeldesign

import (
	"encoding/json"
	"strings"
)

const (
	relationDirectionReverse     = "reverse"
	relationDirectionNormal      = "normal"
	relationCardinalityOneToMany = "one-to-many"
	relationWidgetMultiReadonly  = "relation-multi-readonly"
	relationTypeOneToMany        = "ONE_TO_MANY"
	relationTypeManyToOne        = "MANY_TO_ONE"
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
	jsonBytes, err := json.Marshal(schema)
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
		// Custom ModelCraft metadata stays at root level for model-level info
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

	// 应用验证规则（标准 JSON Schema Draft 7 字段写入 fieldSchema，x-mc 字段写入 xmc）
	xmc := g.buildXMC(field, fieldSchema)

	// 添加自定义ModelCraft属性（写入 xmc）
	g.applyCustomProperties(fieldSchema, field, xmc)

	// 写入 x-mc 对象
	fieldSchema["x-mc"] = xmc

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
		// Precision and scale will be added in x-mc

	// Boolean type
	case FormatBoolean:
		schema["type"] = string(SchemaTypeBoolean)

	// Enum types
	case FormatEnum:
		schema["type"] = string(SchemaTypeString)
		g.applyEnumCodes(schema, field)
	case FormatEnumArray:
		schema["type"] = string(SchemaTypeArray)
		items := map[string]interface{}{
			"type": string(SchemaTypeString),
		}
		g.applyEnumCodes(items, field)
		schema["items"] = items

	// Relation type
	case FormatRelation:
		// REVERSE direction relation is one-to-many, so schema should expose an array.
		if g.isOneToManyRelation(field) {
			schema["type"] = string(SchemaTypeArray)
			schema["items"] = map[string]interface{}{
				"type": string(SchemaTypeObject),
			}
			return
		}
		schema["type"] = string(SchemaTypeObject)
	}
}

// applyEnumCodes 只写顶层 enum 码值列表（标准 JSON Schema Draft 7 字段）
func (g *JSONSchemaGenerator) applyEnumCodes(schema map[string]interface{}, field *FieldDefinition) {
	if field.Enum == nil {
		return
	}
	codes := make([]string, len(field.Enum.Options))
	for i, opt := range field.Enum.Options {
		codes[i] = opt.Code
	}
	schema["enum"] = codes
}

// buildEnumMetadata 构建枚举元数据（写入 x-mc.enum）
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

// buildXMC 构建 x-mc 对象，包含所有非 JSON Schema Draft 7 标准字段。
// fieldSchema 在调用前已经由 applyTypeAndFormat 写入了标准字段（type/format/enum/items），
// 此函数负责从 fieldSchema 和 field 中收集需要迁入 x-mc 的内容，同时将验证规则中的
// 标准字段（maxLength/minLength 等）写入 fieldSchema。
func (g *JSONSchemaGenerator) buildXMC(
	field *FieldDefinition,
	fieldSchema map[string]interface{},
) map[string]interface{} {
	xmc := map[string]interface{}{}

	// ── 始终存在的字段 ──────────────────────────────────────────────────────────
	xmc["isPrimary"] = field.IsPrimary
	xmc["isUnique"] = field.IsUnique
	xmc["displayOrder"] = field.DisplayOrder
	if field.Type != nil {
		// Preserve ModelCraft field format in extension metadata and avoid overloading
		// JSON Schema standard "format" keyword.
		xmc["format"] = string(field.Type.Format)
	}
	// nullable：NonNull=true 表示非空（不 nullable），NonNull=false 表示 nullable
	xmc["nullable"] = !field.NonNull

	// ── 有值时才写入的字段 ────────────────────────────────────────────────────────
	if field.StorageHint != nil {
		xmc["storageHint"] = *field.StorageHint
	}

	if field.RelateFKID != nil {
		xmc["relateFkId"] = *field.RelateFKID
		// relation 元数据：RELATION 字段同样从 Metadata 注入 relation 信息
		if field.Metadata != nil {
			if rel, ok := field.Metadata["x-relation"]; ok {
				xmc["relation"] = rel
			}
		}
	}

	if field.BelongsToFKID != nil {
		xmc["belongsToFkId"] = *field.BelongsToFKID
		// relation 元数据：仅在 Metadata 已由 App 层填充时输出
		if field.Metadata != nil {
			if rel, ok := field.Metadata["x-relation"]; ok {
				xmc["relation"] = rel
			}
		}
	}

	// ── 显式关系类型标注（前端无需再推断） ────────────────────────────────────────
	if relationType, direction, ok := g.detectRelationTypeAndDirection(field); ok {
		xmc["relationType"] = relationType
		xmc["relationDirection"] = direction
	}

	// ── 枚举元数据 ─────────────────────────────────────────────────────────────
	if field.Enum != nil {
		xmc["enum"] = g.buildEnumMetadata(field.Enum)
		if enumLabelFieldName, ok := g.resolveEnumLabelFieldNameFromMetadata(field); ok {
			xmc["enumLabelFieldName"] = enumLabelFieldName
		}
	}

	// ── 验证规则（标准字段写 fieldSchema，x-mc 字段写 xmc） ─────────────────────
	g.applyValidationRules(fieldSchema, xmc, field)

	// ── widget 决策（优先级严格按规范） ──────────────────────────────────────────
	if w := g.decideWidget(field, fieldSchema); w != "" {
		xmc["widget"] = w
	}

	return xmc
}

func (g *JSONSchemaGenerator) resolveEnumLabelFieldNameFromMetadata(field *FieldDefinition) (string, bool) {
	if field == nil || !field.IsEnumField() {
		return "", false
	}

	cfg, err := field.parseEnumDisplayFromMetadata()
	if err != nil || cfg == nil {
		return "", false
	}
	if cfg.Enabled != nil && !*cfg.Enabled {
		return "", false
	}

	if field.IsEnumArrayField() {
		if cfg.LabelsFieldName == nil {
			return "", false
		}
		if labelFieldName := strings.TrimSpace(*cfg.LabelsFieldName); labelFieldName != "" {
			return labelFieldName, true
		}
		return "", false
	}

	if cfg.LabelFieldName == nil {
		return "", false
	}
	if labelFieldName := strings.TrimSpace(*cfg.LabelFieldName); labelFieldName != "" {
		return labelFieldName, true
	}
	return "", false
}

// decideWidget 根据字段属性决定 widget 值，返回空字符串表示不写 widget 键。
// 优先级：BelongsToFKID > storageHint=TEXT > FormatEnum/FormatEnumArray > FormatDate > FormatDateTime > FormatTime
func (g *JSONSchemaGenerator) decideWidget(field *FieldDefinition, fieldSchema map[string]interface{}) string {
	// 0. one-to-many RELATION 字段（只读）→ 专用只读多选组件
	if field.Type != nil && field.Type.Format == FormatRelation && g.isOneToManyRelation(field) {
		return relationWidgetMultiReadonly
	}

	// 1. belongs-to 外键列 → 关联选择器
	if field.BelongsToFKID != nil {
		return "relation-selector"
	}

	// 2. storageHint == "TEXT" → 多行文本框
	if field.StorageHint != nil && *field.StorageHint == "TEXT" {
		return "textarea"
	}

	if field.Type == nil {
		return ""
	}

	switch field.Type.Format {
	// 3. 枚举（单选/多选） → 枚举下拉
	case FormatEnum, FormatEnumArray:
		return "enum-select"
	// 4. 日期
	case FormatDate:
		return "date"
	// 5. 日期时间
	case FormatDateTime:
		return "datetime-local"
	// 6. 时间
	case FormatTime:
		return "time"
	}

	return ""
}

// applyValidationRules 将验证规则分流写入：
//   - 标准 JSON Schema Draft 7 字段（maxLength/minLength/pattern/maximum/minimum/maxItems/minItems）→ fieldSchema
//   - 非标准验证字段（minDate/maxDate/minTime/maxTime/precision/scale/validateRule）→ xmc
func (g *JSONSchemaGenerator) applyValidationRules(
	schema map[string]interface{},
	xmc map[string]interface{},
	field *FieldDefinition,
) {
	if field.Validation == nil {
		return
	}

	v := field.Validation

	// ── 标准 JSON Schema Draft 7 验证字段 → schema ────────────────────────────
	if v.MaxLength != nil {
		schema["maxLength"] = *v.MaxLength
	}
	if v.MinLength != nil {
		schema["minLength"] = *v.MinLength
	}
	if v.Pattern != nil {
		schema["pattern"] = *v.Pattern
	}
	if v.Maximum != nil {
		schema["maximum"] = *v.Maximum
	}
	if v.Minimum != nil {
		schema["minimum"] = *v.Minimum
	}
	if v.MaxItems != nil {
		schema["maxItems"] = *v.MaxItems
	}
	if v.MinItems != nil {
		schema["minItems"] = *v.MinItems
	}

	// ── 非标准验证字段 → xmc ──────────────────────────────────────────────────
	if v.MinDate != nil {
		xmc["minDate"] = *v.MinDate
	}
	if v.MaxDate != nil {
		xmc["maxDate"] = *v.MaxDate
	}
	if v.MinTime != nil {
		xmc["minTime"] = *v.MinTime
	}
	if v.MaxTime != nil {
		xmc["maxTime"] = *v.MaxTime
	}
	if v.Precision != nil {
		xmc["precision"] = *v.Precision
	}
	if v.Scale != nil {
		xmc["scale"] = *v.Scale
	}
	if v.Rule != "" {
		xmc["validateRule"] = string(v.Rule)
	}
}

// applyCustomProperties 在 fieldSchema 写入标准 JSON Schema 字段（readOnly），
// 不再写任何 x-* 顶层键。
func (g *JSONSchemaGenerator) applyCustomProperties(
	schema map[string]interface{},
	field *FieldDefinition,
	_ map[string]interface{}, // xmc 已在 buildXMC 中填充，此处保留签名一致性
) {
	// Mark readOnly for fields that cannot be edited by the user:
	//   - Primary key fields are generated by the database
	//   - RELATION fields are derived and displayed read-only in tables
	if field.IsPrimary || (field.Type != nil && field.Type.Format == FormatRelation) {
		schema["readOnly"] = true
	}
}

func (g *JSONSchemaGenerator) isOneToManyRelation(field *FieldDefinition) bool {
	if field == nil || field.Metadata == nil {
		return false
	}

	relRaw, ok := field.Metadata["x-relation"]
	if !ok || relRaw == nil {
		return false
	}

	switch rel := relRaw.(type) {
	case map[string]interface{}:
		if cardinality, ok := rel["cardinality"].(string); ok && cardinality == relationCardinalityOneToMany {
			return true
		}
		if direction, ok := rel["direction"].(string); ok && direction == relationDirectionReverse {
			return true
		}
	case map[string]string:
		if cardinality, ok := rel["cardinality"]; ok && cardinality == relationCardinalityOneToMany {
			return true
		}
		if direction, ok := rel["direction"]; ok && direction == relationDirectionReverse {
			return true
		}
	}

	return false
}

func (g *JSONSchemaGenerator) detectRelationTypeAndDirection(field *FieldDefinition) (relationType, direction string, ok bool) {
	if field == nil {
		return "", "", false
	}

	// 优先依据元数据（x-relation.direction / x-relation.cardinality）
	if field.Metadata != nil {
		if relRaw, exists := field.Metadata["x-relation"]; exists && relRaw != nil {
			switch rel := relRaw.(type) {
			case map[string]interface{}:
				if dir, hasDir := rel["direction"].(string); hasDir {
					switch dir {
					case relationDirectionReverse:
						return relationTypeOneToMany, relationDirectionReverse, true
					case relationDirectionNormal:
						return relationTypeManyToOne, relationDirectionNormal, true
					}
				}
				if cardinality, hasCardinality := rel["cardinality"].(string); hasCardinality {
					switch cardinality {
					case relationCardinalityOneToMany:
						return relationTypeOneToMany, relationDirectionReverse, true
					case "many-to-one":
						return relationTypeManyToOne, relationDirectionNormal, true
					}
				}
			case map[string]string:
				if dir, hasDir := rel["direction"]; hasDir {
					switch dir {
					case relationDirectionReverse:
						return relationTypeOneToMany, relationDirectionReverse, true
					case relationDirectionNormal:
						return relationTypeManyToOne, relationDirectionNormal, true
					}
				}
				if cardinality, hasCardinality := rel["cardinality"]; hasCardinality {
					switch cardinality {
					case relationCardinalityOneToMany:
						return relationTypeOneToMany, relationDirectionReverse, true
					case "many-to-one":
						return relationTypeManyToOne, relationDirectionNormal, true
					}
				}
			}
		}
	}

	// 元数据缺失时，按字段形态兜底推断
	if field.BelongsToFKID != nil {
		return relationTypeManyToOne, relationDirectionNormal, true
	}
	if field.RelateFKID != nil {
		return relationTypeOneToMany, relationDirectionReverse, true
	}

	return "", "", false
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
