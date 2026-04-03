package modeldesign

import (
	"fmt"
	"modelcraft/pkg/ddlfactory"
	"strings"
)

// ReverseTypeMapper 反向类型映射器（MySQL -> Domain）
type ReverseTypeMapper struct{}

// NewReverseTypeMapper 创建反向类型映射器
func NewReverseTypeMapper() *ReverseTypeMapper {
	return &ReverseTypeMapper{}
}

// MappingResult 映射结果
type MappingResult struct {
	Format     FormatType        // Domain字段格式
	Supported  bool              // 是否支持该类型
	SkipReason string            // 如果不支持，跳过的原因
	Validation *ValidationConfig // 验证配置（包含长度、精度等）
}

// TableColumn 表列定义（Domain层结构，避免循环依赖）
type TableColumn struct {
	Name          string
	DataType      string
	Length        int64
	Precision     int
	Scale         int
	Nullable      bool
	DefaultValue  *string
	AutoIncrement bool
	Comment       string
}

// MapColumnToFieldType 将MySQL列定义映射到Domain字段类型
func (m *ReverseTypeMapper) MapColumnToFieldType(col TableColumn) MappingResult {
	dataType := strings.ToUpper(col.DataType)
	validation := &ValidationConfig{}

	// 按类型分类处理
	if result, ok := m.mapIntegerType(dataType, col); ok {
		return result
	}
	if result, ok := m.mapFloatType(dataType); ok {
		return result
	}
	if result, ok := m.mapDecimalType(dataType, col, validation); ok {
		return result
	}
	if result, ok := m.mapStringType(dataType, col, validation); ok {
		return result
	}
	if result, ok := m.mapDateTimeType(dataType); ok {
		return result
	}
	if result, ok := m.mapJSONType(dataType); ok {
		return result
	}

	return m.mapUnsupportedType(dataType)
}

// mapIntegerType 映射整数类型
func (m *ReverseTypeMapper) mapIntegerType(dataType string, col TableColumn) (MappingResult, bool) {
	if dataType == "TINYINT" && col.Length == 1 {
		return MappingResult{Format: FormatBoolean, Supported: true, Validation: &ValidationConfig{}}, true
	}

	integerTypes := map[string]bool{
		"TINYINT": true, "SMALLINT": true, "MEDIUMINT": true,
		"INT": true, "INTEGER": true, "BIGINT": true,
	}
	if integerTypes[dataType] {
		return MappingResult{Format: FormatInteger, Supported: true, Validation: &ValidationConfig{}}, true
	}
	return MappingResult{}, false
}

// mapFloatType 映射浮点类型
func (m *ReverseTypeMapper) mapFloatType(dataType string) (MappingResult, bool) {
	floatTypes := map[string]bool{"FLOAT": true, "DOUBLE": true, "REAL": true}
	if floatTypes[dataType] {
		return MappingResult{Format: FormatNumber, Supported: true, Validation: &ValidationConfig{}}, true
	}
	return MappingResult{}, false
}

// mapDecimalType 映射DECIMAL类型
func (m *ReverseTypeMapper) mapDecimalType(
	dataType string, col TableColumn, validation *ValidationConfig,
) (MappingResult, bool) {
	if dataType == "DECIMAL" || dataType == "NUMERIC" {
		if col.Precision > 0 {
			precision := col.Precision
			scale := col.Scale
			validation.Precision = &precision
			validation.Scale = &scale
		}
		return MappingResult{Format: FormatDecimal, Supported: true, Validation: validation}, true
	}
	return MappingResult{}, false
}

// mapStringType 映射字符串类型
func (m *ReverseTypeMapper) mapStringType(
	dataType string, col TableColumn, validation *ValidationConfig,
) (MappingResult, bool) {
	textTypes := map[string]bool{
		"TEXT": true, "TINYTEXT": true, "MEDIUMTEXT": true, "LONGTEXT": true,
	}

	switch {
	case dataType == "CHAR" || dataType == "VARCHAR":
		if col.Length > 0 {
			maxLen := int(col.Length)
			validation.MaxLength = &maxLen
		}
		return MappingResult{Format: FormatString, Supported: true, Validation: validation}, true
	case textTypes[dataType]:
		return MappingResult{Format: FormatString, Supported: true, Validation: validation}, true
	}
	return MappingResult{}, false
}

// mapDateTimeType 映射日期时间类型
func (m *ReverseTypeMapper) mapDateTimeType(dataType string) (MappingResult, bool) {
	typeMapping := map[string]FormatType{
		"DATE":      FormatDate,
		"DATETIME":  FormatDateTime,
		"TIMESTAMP": FormatDateTime,
		"TIME":      FormatTime,
	}
	if format, ok := typeMapping[dataType]; ok {
		return MappingResult{Format: format, Supported: true, Validation: &ValidationConfig{}}, true
	}
	return MappingResult{}, false
}

// mapJSONType 映射JSON类型
func (m *ReverseTypeMapper) mapJSONType(dataType string) (MappingResult, bool) {
	if dataType == string(ddlfactory.JSON) {
		return MappingResult{Format: FormatString, Supported: true, Validation: &ValidationConfig{}}, true
	}
	return MappingResult{}, false
}

// mapUnsupportedType 映射不支持的类型
func (m *ReverseTypeMapper) mapUnsupportedType(dataType string) MappingResult {
	unsupportedTypes := map[string]string{
		"ENUM": "enum/set", "SET": "enum/set",
		"BINARY": "binary/blob", "VARBINARY": "binary/blob",
		"BLOB": "binary/blob", "TINYBLOB": "binary/blob",
		"MEDIUMBLOB": "binary/blob", "LONGBLOB": "binary/blob",
		"BIT": "BIT",
	}
	if category, ok := unsupportedTypes[dataType]; ok {
		return MappingResult{
			Supported: false, SkipReason: fmt.Sprintf("MySQL type %s (%s) is not supported", dataType, category),
		}
	}

	spatialTypes := map[string]bool{
		"GEOMETRY": true, "POINT": true, "LINESTRING": true, "POLYGON": true,
		"MULTIPOINT": true, "MULTILINESTRING": true, "MULTIPOLYGON": true, "GEOMETRYCOLLECTION": true,
	}
	if spatialTypes[dataType] {
		return MappingResult{
			Supported: false, SkipReason: fmt.Sprintf("MySQL type %s (spatial type) is not supported", dataType),
		}
	}

	return MappingResult{Supported: false, SkipReason: fmt.Sprintf("Unknown MySQL type: %s", dataType)}
}

// BuildFieldDefinition 从列定义构建字段定义
func (m *ReverseTypeMapper) BuildFieldDefinition(
	col TableColumn,
	modelID string,
	modelLocator *ModelLocator,
) (*FieldDefinition, *MappingResult) {
	// 映射类型
	mappingResult := m.MapColumnToFieldType(col)

	// 如果不支持该类型，返回 nil 和映射结果
	if !mappingResult.Supported {
		return nil, &mappingResult
	}

	// 创建字段类型
	fieldType := GetFieldTypeByFormat(mappingResult.Format)
	if fieldType == nil {
		mappingResult.Supported = false
		mappingResult.SkipReason = fmt.Sprintf("Failed to get field type for format: %s", mappingResult.Format)
		return nil, &mappingResult
	}

	// 构建字段定义
	fieldDef := &FieldDefinition{
		ModelID:      modelID,
		Name:         strings.ToLower(col.Name),
		ModelLocator: modelLocator,
		Title:        formatFieldTitle(col.Name),
		Description:  col.Comment,
		Type:         fieldType,
		NonNull:      !col.Nullable,
		Required:     false, // 默认非必填
		IsUnique:     false,
		IsPrimary:    false,
		Status:       FieldStatusInit,
		Validation:   mappingResult.Validation,
		DisplayOrder: "", // 默认空字符串，由应用层设置
	}

	return fieldDef, &mappingResult
}

// formatFieldTitle 格式化字段标题
// 例如: "user_name" -> "User Name", "userId" -> "UserId"
func formatFieldTitle(name string) string {
	// 如果包含下划线，按下划线分割并首字母大写
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for i, part := range parts {
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[0:1]) + strings.ToLower(part[1:])
			}
		}
		return strings.Join(parts, " ")
	}

	// 驼峰命名，直接首字母大写
	if len(name) > 0 {
		return strings.ToUpper(name[0:1]) + name[1:]
	}

	return name
}
