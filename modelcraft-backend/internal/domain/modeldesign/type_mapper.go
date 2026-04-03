package modeldesign

import (
	"fmt"
	"modelcraft/pkg/ddlfactory"

	bizerrors "modelcraft/pkg/bizerrors"
)

// TypeMapper 类型映射器接口
type TypeMapper interface {
	MapToMySQL(field *FieldDefinition) (string, error)
}

// MySQLTypeMapper MySQL类型映射器
type MySQLTypeMapper struct{}

// NewMySQLTypeMapper 创建MySQL类型映射器
func NewMySQLTypeMapper() *MySQLTypeMapper {
	return &MySQLTypeMapper{}
}

// MapToMySQL 将FieldDefinition映射到MySQL类型
func (m *MySQLTypeMapper) MapToMySQL(field *FieldDefinition) (string, error) {
	// 检查field和field.Type是否为nil
	if field == nil {
		return "", bizerrors.Errorf("field cannot be nil")
	}

	if field.Type == nil {
		return "", bizerrors.Errorf("field.Type cannot be nil")
	}

	// 如果有storageHint，优先使用
	if field.StorageHint != nil && *field.StorageHint != "" {
		return *field.StorageHint, nil
	}

	// 根据Format自动映射
	switch field.Type.Format {
	case FormatString:
		return m.mapString(field)
	case FormatUUID:
		return "CHAR(36)", nil
	case FormatDate:
		return "DATE", nil
	case FormatDateTime:
		return "DATETIME", nil
	case FormatTime:
		return "TIME", nil
	case FormatNumber:
		return "DOUBLE", nil
	case FormatInteger:
		return "INT", nil
	case FormatDecimal:
		return m.mapDecimal(field)
	case FormatBoolean:
		return "TINYINT(1)", nil
	case FormatEnum:
		// ENUM format: check IsArray flag for storage type
		// - IsArray=false (single-select): VARCHAR(64)
		// - IsArray=true (multi-select): JSON
		if field.IsArray {
			return string(ddlfactory.JSON), nil
		}
		return "VARCHAR(64)", nil
	case FormatEnumArray:
		// Legacy FormatEnumArray (deprecated, will be removed)
		// Always multi-select, stored as JSON
		return string(ddlfactory.JSON), nil
	default:
		// 检查是否是Object或Array类型（通过SchemaType判断）
		switch field.Type.SchemaType {
		case SchemaTypeObject, SchemaTypeArray:
			return string(ddlfactory.JSON), nil
		}
		return "", bizerrors.Errorf("不支持的格式类型: %s", field.Type)
	}
}

// mapString 映射字符串类型
func (m *MySQLTypeMapper) mapString(field *FieldDefinition) (string, error) {
	if field.Validation != nil && field.Validation.MaxLength != nil {
		maxLen := *field.Validation.MaxLength
		if maxLen <= 255 {
			return fmt.Sprintf("VARCHAR(%d)", maxLen), nil
		}
	}
	// 没有maxLength或maxLength>255，使用TEXT
	return "TEXT", nil
}

// mapDecimal 映射decimal类型
func (m *MySQLTypeMapper) mapDecimal(field *FieldDefinition) (string, error) {
	if field.Validation == nil || field.Validation.Precision == nil || field.Validation.Scale == nil {
		// 默认精度
		return "DECIMAL(10,2)", nil
	}

	precision := *field.Validation.Precision
	scale := *field.Validation.Scale

	return fmt.Sprintf("DECIMAL(%d,%d)", precision, scale), nil
}
