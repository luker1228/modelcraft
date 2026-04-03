package core

import "github.com/graphql-go/graphql"

// FieldType 字段类型枚举
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeInteger FieldType = "integer"
	FieldTypeFloat   FieldType = "float"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeID      FieldType = "id"
)

// String 实现Stringer接口
func (ft FieldType) String() string {
	return string(ft)
}

// IsValid 验证字段类型是否有效
func (ft FieldType) IsValid() bool {
	switch ft {
	case FieldTypeString, FieldTypeInteger, FieldTypeFloat, FieldTypeBoolean, FieldTypeID:
		return true
	default:
		return false
	}
}

// GetJSONType 获取对应的JSON Schema类型
func (ft FieldType) GetJSONType() string {
	switch ft {
	case FieldTypeString, FieldTypeID:
		return string(FieldTypeString)
	case FieldTypeInteger:
		return string(FieldTypeInteger)
	case FieldTypeFloat:
		return "number"
	case FieldTypeBoolean:
		return string(FieldTypeBoolean)
	default:
		return string(FieldTypeString)
	}
}

// GetGraphQLType 获取对应的GraphQL类型
func (ft FieldType) GetGraphQLType() graphql.Type {
	switch ft {
	case FieldTypeString:
		return graphql.String
	case FieldTypeInteger:
		return graphql.Int
	case FieldTypeFloat:
		return graphql.Float
	case FieldTypeBoolean:
		return graphql.Boolean
	case FieldTypeID:
		return graphql.ID
	default:
		return graphql.String
	}
}

// GetSDLType 获取对应的GraphQL SDL类型名称
func (ft FieldType) GetSDLType() string {
	switch ft {
	case FieldTypeString:
		return "String"
	case FieldTypeInteger:
		return "Int"
	case FieldTypeFloat:
		return "Float"
	case FieldTypeBoolean:
		return "Boolean"
	case FieldTypeID:
		return "ID"
	default:
		return "String"
	}
}

// GetDefaultValue 获取类型的默认值
func (ft FieldType) GetDefaultValue() interface{} {
	switch ft {
	case FieldTypeString, FieldTypeID:
		return ""
	case FieldTypeInteger:
		return 0
	case FieldTypeFloat:
		return 0.0
	case FieldTypeBoolean:
		return false
	default:
		return nil
	}
}

// AllFieldTypes 返回所有支持的字段类型
func AllFieldTypes() []FieldType {
	return []FieldType{
		FieldTypeString,
		FieldTypeInteger,
		FieldTypeFloat,
		FieldTypeBoolean,
		FieldTypeID,
	}
}

// ParseFieldType 从字符串解析FieldType
func ParseFieldType(s string) (FieldType, error) {
	ft := FieldType(s)
	if !ft.IsValid() {
		return "", &InvalidFieldTypeError{Type: s}
	}
	return ft, nil
}

// MustParseFieldType 从字符串解析FieldType，失败时panic
func MustParseFieldType(s string) FieldType {
	ft, err := ParseFieldType(s)
	if err != nil {
		panic(err)
	}
	return ft
}

// InvalidFieldTypeError 无效字段类型错误
type InvalidFieldTypeError struct {
	Type string
}

func (e *InvalidFieldTypeError) Error() string {
	return "invalid field type: " + e.Type
}
