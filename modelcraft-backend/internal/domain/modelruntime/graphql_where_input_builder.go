package modelruntime

import (
	"errors"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/query"

	"github.com/graphql-go/graphql"
)

// whereInputBuilder 构建递归的WhereInput类型，支持逻辑操作符和字段条件。
// 用于生成GraphQL查询中的where输入类型。
type whereInputBuilder struct {
	fieldConditionManager *fieldConditionTypeManager
}

// newWhereInputBuilder 创建新的WhereInputBuilder。
func newWhereInputBuilder(fieldConditionManager *fieldConditionTypeManager) *whereInputBuilder {
	return &whereInputBuilder{
		fieldConditionManager: fieldConditionManager,
	}
}

// buildWhereInputType 创建模型特定的WhereInput类型，支持递归
func (b *whereInputBuilder) buildWhereInputType(model *RuntimeModel) *graphql.InputObject {
	fields := graphql.InputObjectConfigFieldMap{}

	b.buildFieldConditionFields(fields, model)

	// 使用 gqlTypeName 确保类型名合法（见 graphql_type_name.go）
	whereInputName := "WhereInput"
	if model != nil && model.Name != "" {
		whereInputName = gqlTypeName(model.Name) + "WhereInput"
	}

	whereInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   whereInputName,
		Fields: fields,
	})

	b.buildLogicalOperatorFields(fields, whereInput)

	return whereInput
}

// buildUniqueWhereInputType 创建模型特定的UniqueWhereInput类型，支持递归
func (b *whereInputBuilder) buildUniqueWhereInputType(model *RuntimeModel) (*graphql.InputObject, error) {
	fields := graphql.InputObjectConfigFieldMap{}
	if model.getUniqueField() == nil {
		return nil, errors.New("model has no unique fields")
	}

	for _, field := range model.getUniqueField() {
		// 跳过空名称字段
		if field.Name == "" {
			continue
		}
		graphqlType, err := getGraphqlTypeBy(field.Type.Format)
		if err != nil {
			return nil, err
		}

		fields[field.Name] = &graphql.InputObjectFieldConfig{
			Type: graphqlType,
		}
	}

	// 使用 gqlTypeName 确保类型名合法（见 graphql_type_name.go）
	whereInputName := "UniqueWhereInput"
	if model != nil && model.Name != "" {
		whereInputName = gqlTypeName(model.Name) + "UniqueWhereInput"
	}

	whereInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   whereInputName,
		Fields: fields,
	})

	return whereInput, nil
}

func (b *whereInputBuilder) buildLogicalOperatorFields(
	fields graphql.InputObjectConfigFieldMap,
	whereInput *graphql.InputObject,
) {
	fields[query.LogicalOperatorAND] = &graphql.InputObjectFieldConfig{
		Type: graphql.NewList(graphql.NewNonNull(whereInput)),
	}

	fields[query.LogicalOperatorOR] = &graphql.InputObjectFieldConfig{
		Type: graphql.NewList(graphql.NewNonNull(whereInput)),
	}

	fields[query.LogicalOperatorNOT] = &graphql.InputObjectFieldConfig{
		Type: whereInput,
	}
}

func (b *whereInputBuilder) buildFieldConditionFields(fields graphql.InputObjectConfigFieldMap, model *RuntimeModel) {
	if model.Fields == nil {
		return
	}

	for _, field := range model.Fields {
		// 跳过空名称字段
		if field.Name == "" {
			continue
		}
		if !isQueryableField(field) {
			continue
		}

		conditionType := b.fieldConditionManager.getFieldConditionTypeByFormat(field.Type.Format)

		fields[field.Name] = &graphql.InputObjectFieldConfig{
			Type: conditionType,
		}
	}
}

func isQueryableField(field *modeldesign.FieldDefinition) bool {
	return field != nil
}
