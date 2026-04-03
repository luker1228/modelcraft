package generators

import (
	"fmt"
	"modelcraft/pkg/schema/core"
	"strconv"

	"github.com/graphql-go/graphql"
)

// GraphQLTypeFactory GraphQL类型工厂
type GraphQLTypeFactory struct{}

// NewGraphQLTypeFactory 创建新的GraphQL类型工厂
func NewGraphQLTypeFactory() *GraphQLTypeFactory {
	return &GraphQLTypeFactory{}
}

// CreateFieldType 创建字段类型
func (f *GraphQLTypeFactory) CreateFieldType(field *core.FieldDefinition) graphql.Output {
	baseType := field.Type.GetGraphQLType()

	if field.Required {
		return graphql.NewNonNull(baseType)
	}

	return baseType
}

// CreateResolver 创建解析器
func (f *GraphQLTypeFactory) CreateResolver(field *core.FieldDefinition) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		if source, ok := p.Source.(map[string]interface{}); ok {
			value := source[field.Key]
			return f.validateAndConvert(value, field)
		}
		return field.Default, nil
	}
}

// validateAndConvert 验证和转换值
func (f *GraphQLTypeFactory) validateAndConvert(value interface{}, field *core.FieldDefinition) (interface{}, error) {
	// 基本类型转换和验证逻辑
	if value == nil {
		return field.Default, nil
	}

	// 根据字段类型进行转换
	switch field.Type {
	case core.FieldTypeInteger:
		return f.convertToInt(value)
	case core.FieldTypeFloat:
		return f.convertToFloat(value)
	case core.FieldTypeBoolean:
		return f.convertToBool(value)
	case core.FieldTypeString, core.FieldTypeID:
		return f.convertToString(value)
	default:
		return value, nil
	}
}

// convertToInt 转换为整数
func (f *GraphQLTypeFactory) convertToInt(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
		return nil, fmt.Errorf("cannot convert string '%s' to integer", v)
	default:
		return nil, fmt.Errorf("cannot convert %T to integer", value)
	}
}

// convertToFloat 转换为浮点数
func (f *GraphQLTypeFactory) convertToFloat(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
		return nil, fmt.Errorf("cannot convert string '%s' to float", v)
	default:
		return nil, fmt.Errorf("cannot convert %T to float", value)
	}
}

// convertToBool 转换为布尔值
func (f *GraphQLTypeFactory) convertToBool(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, nil
		}
		return nil, fmt.Errorf("cannot convert string '%s' to boolean", v)
	default:
		return nil, fmt.Errorf("cannot convert %T to boolean", value)
	}
}

// convertToString 转换为字符串
func (f *GraphQLTypeFactory) convertToString(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}
