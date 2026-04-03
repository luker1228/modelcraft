package modelruntime

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/query"
	"sync"

	"github.com/graphql-go/graphql"
)

// fieldConditionTypeManager 字段条件类型管理器，用于生成GraphQL字段条件输入类型。
// 提供不同数据类型的字段条件类型（字符串、整数、数字、布尔值）。
type fieldConditionTypeManager struct {
	cache map[string]*graphql.InputObject
	mu    sync.RWMutex
}

// newFieldConditionTypeManager 创建新的字段条件类型管理器。
func newFieldConditionTypeManager() *fieldConditionTypeManager {
	return &fieldConditionTypeManager{
		cache: make(map[string]*graphql.InputObject),
	}
}

// getStringFieldConditionType 返回字符串字段条件输入类型
func (m *fieldConditionTypeManager) getStringFieldConditionType() *graphql.InputObject {
	return m.getOrCreateType("StringFieldInput", func() *graphql.InputObject {
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name: "StringFieldInput",
			Fields: graphql.InputObjectConfigFieldMap{
				query.FieldEquals: &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				query.FieldNot: &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				query.FieldIn: &graphql.InputObjectFieldConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
				},
				query.FieldContains: &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				query.FieldStartsWith: &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				query.FieldEndsWith: &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				query.FieldMode: &graphql.InputObjectFieldConfig{
					Type: m.getQueryModeEnum(),
				},
			},
		})
	})
}

// getIntFieldConditionType 返回整数字段条件输入类型
func (m *fieldConditionTypeManager) getIntFieldConditionType() *graphql.InputObject {
	return m.getOrCreateType("IntFieldInput", func() *graphql.InputObject {
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name: "IntFieldInput",
			Fields: graphql.InputObjectConfigFieldMap{
				query.FieldEquals: &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
				query.FieldNot: &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
				query.FieldIn: &graphql.InputObjectFieldConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.Int)),
				},
				query.FieldLt: &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
				query.FieldLte: &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
				query.FieldGt: &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
				query.FieldGte: &graphql.InputObjectFieldConfig{
					Type: graphql.Int,
				},
			},
		})
	})
}

// getNumberFieldConditionType 返回数字字段条件输入类型（浮点数）
func (m *fieldConditionTypeManager) getNumberFieldConditionType() *graphql.InputObject {
	return m.getOrCreateType("NumberFieldInput", func() *graphql.InputObject {
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name: "NumberFieldInput",
			Fields: graphql.InputObjectConfigFieldMap{
				query.FieldEquals: &graphql.InputObjectFieldConfig{
					Type: graphql.Float,
				},
				query.FieldNot: &graphql.InputObjectFieldConfig{
					Type: graphql.Float,
				},
				query.FieldIn: &graphql.InputObjectFieldConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.Float)),
				},
				query.FieldLt: &graphql.InputObjectFieldConfig{
					Type: graphql.Float,
				},
				query.FieldLte: &graphql.InputObjectFieldConfig{
					Type: graphql.Float,
				},
				query.FieldGt: &graphql.InputObjectFieldConfig{
					Type: graphql.Float,
				},
				query.FieldGte: &graphql.InputObjectFieldConfig{
					Type: graphql.Float,
				},
			},
		})
	})
}

// getBooleanFieldConditionType 返回布尔字段条件输入类型
func (m *fieldConditionTypeManager) getBooleanFieldConditionType() *graphql.InputObject {
	return m.getOrCreateType("BooleanFieldInput", func() *graphql.InputObject {
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name: "BooleanFieldInput",
			Fields: graphql.InputObjectConfigFieldMap{
				query.FieldEquals: &graphql.InputObjectFieldConfig{
					Type: graphql.Boolean,
				},
				query.FieldNot: &graphql.InputObjectFieldConfig{
					Type: graphql.Boolean,
				},
			},
		})
	})
}

// getDateFieldConditionType 返回日期字段条件输入类型
func (m *fieldConditionTypeManager) getDateFieldConditionType() *graphql.InputObject {
	return m.getOrCreateType("DateFieldInput", func() *graphql.InputObject {
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name: "DateFieldInput",
			Fields: graphql.InputObjectConfigFieldMap{
				query.FieldEquals: &graphql.InputObjectFieldConfig{
					Type: GraphQLDate,
				},
				query.FieldNot: &graphql.InputObjectFieldConfig{
					Type: GraphQLDate,
				},
				query.FieldIn: &graphql.InputObjectFieldConfig{
					Type: graphql.NewList(graphql.NewNonNull(GraphQLDate)),
				},
				query.FieldLt: &graphql.InputObjectFieldConfig{
					Type: GraphQLDate,
				},
				query.FieldLte: &graphql.InputObjectFieldConfig{
					Type: GraphQLDate,
				},
				query.FieldGt: &graphql.InputObjectFieldConfig{
					Type: GraphQLDate,
				},
				query.FieldGte: &graphql.InputObjectFieldConfig{
					Type: GraphQLDate,
				},
			},
		})
	})
}

// getTimeFieldConditionType 返回时间字段条件输入类型
func (m *fieldConditionTypeManager) getTimeFieldConditionType() *graphql.InputObject {
	return m.getOrCreateType("TimeFieldInput", func() *graphql.InputObject {
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name: "TimeFieldInput",
			Fields: graphql.InputObjectConfigFieldMap{
				query.FieldEquals: &graphql.InputObjectFieldConfig{
					Type: GraphQLTime,
				},
				query.FieldNot: &graphql.InputObjectFieldConfig{
					Type: GraphQLTime,
				},
				query.FieldIn: &graphql.InputObjectFieldConfig{
					Type: graphql.NewList(graphql.NewNonNull(GraphQLTime)),
				},
				query.FieldLt: &graphql.InputObjectFieldConfig{
					Type: GraphQLTime,
				},
				query.FieldLte: &graphql.InputObjectFieldConfig{
					Type: GraphQLTime,
				},
				query.FieldGt: &graphql.InputObjectFieldConfig{
					Type: GraphQLTime,
				},
				query.FieldGte: &graphql.InputObjectFieldConfig{
					Type: GraphQLTime,
				},
			},
		})
	})
}

// getDateTimeFieldConditionType 返回日期时间字段条件输入类型
func (m *fieldConditionTypeManager) getDateTimeFieldConditionType() *graphql.InputObject {
	return m.getOrCreateType("DateTimeFieldInput", func() *graphql.InputObject {
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name: "DateTimeFieldInput",
			Fields: graphql.InputObjectConfigFieldMap{
				query.FieldEquals: &graphql.InputObjectFieldConfig{
					Type: graphql.DateTime,
				},
				query.FieldNot: &graphql.InputObjectFieldConfig{
					Type: graphql.DateTime,
				},
				query.FieldIn: &graphql.InputObjectFieldConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.DateTime)),
				},
				query.FieldLt: &graphql.InputObjectFieldConfig{
					Type: graphql.DateTime,
				},
				query.FieldLte: &graphql.InputObjectFieldConfig{
					Type: graphql.DateTime,
				},
				query.FieldGt: &graphql.InputObjectFieldConfig{
					Type: graphql.DateTime,
				},
				query.FieldGte: &graphql.InputObjectFieldConfig{
					Type: graphql.DateTime,
				},
			},
		})
	})
}

// getFieldConditionTypeByFormat 根据格式返回相应的字段条件类型
func (m *fieldConditionTypeManager) getFieldConditionTypeByFormat(format modeldesign.FormatType) graphql.Type {
	switch format {
	case modeldesign.FormatString, modeldesign.FormatUUID:
		return m.getStringFieldConditionType()
	case modeldesign.FormatDate:
		return m.getDateFieldConditionType()
	case modeldesign.FormatTime:
		return m.getTimeFieldConditionType()
	case modeldesign.FormatDateTime:
		return m.getDateTimeFieldConditionType()
	case modeldesign.FormatInteger:
		return m.getIntFieldConditionType()
	case modeldesign.FormatNumber, modeldesign.FormatDecimal:
		return m.getNumberFieldConditionType()
	case modeldesign.FormatBoolean:
		return m.getBooleanFieldConditionType()
	default:
		return m.getStringFieldConditionType()
	}
}

func (m *fieldConditionTypeManager) getOrCreateType(
	name string,
	creator func() *graphql.InputObject,
) *graphql.InputObject {
	m.mu.RLock()
	if cached, exists := m.cache[name]; exists {
		m.mu.RUnlock()
		return cached
	}
	m.mu.RUnlock()

	newType := creator()

	m.mu.Lock()
	m.cache[name] = newType
	m.mu.Unlock()

	return newType
}

func (m *fieldConditionTypeManager) getQueryModeEnum() *graphql.Enum {
	return graphql.NewEnum(graphql.EnumConfig{
		Name: query.QueryModeName,
		Values: graphql.EnumValueConfigMap{
			query.QueryModeDefault: &graphql.EnumValueConfig{
				Value: query.QueryModeDefault,
			},
			query.QueryModeInsensitive: &graphql.EnumValueConfig{
				Value: query.QueryModeInsensitive,
			},
		},
	})
}
