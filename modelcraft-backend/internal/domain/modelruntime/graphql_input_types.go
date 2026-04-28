package modelruntime

import (
	"fmt"
	"modelcraft/internal/domain/modeldesign"

	"github.com/graphql-go/graphql"
)

// inputTypeGenerator 编排完整的输入类型生态系统生成，包括where、orderBy、create、update等输入类型。
// 提供缓存机制避免重复生成相同模型的输入类型。
type inputTypeGenerator struct {
	fieldConditionManager *fieldConditionTypeManager
	whereInputBuilder     *whereInputBuilder
	whereInputMap         map[string]*graphql.InputObject
	uniqueWhereInputMap   map[string]*graphql.InputObject
	orderByInputMap       map[string]*graphql.InputObject
	createInputMap        map[string]*graphql.InputObject
	updateInputMap        map[string]*graphql.InputObject
}

// newInputTypeGenerator 创建新的输入类型生成器。
func newInputTypeGenerator() *inputTypeGenerator {
	fieldConditionManager := newFieldConditionTypeManager()
	whereInputBuilder := newWhereInputBuilder(fieldConditionManager)

	return &inputTypeGenerator{
		fieldConditionManager: fieldConditionManager,
		whereInputBuilder:     whereInputBuilder,
		whereInputMap:         make(map[string]*graphql.InputObject),
		uniqueWhereInputMap:   make(map[string]*graphql.InputObject),
		orderByInputMap:       make(map[string]*graphql.InputObject),
		createInputMap:        make(map[string]*graphql.InputObject),
		updateInputMap:        make(map[string]*graphql.InputObject),
	}
}

// generateWhereInputType 创建包含逻辑操作符和字段条件的WhereInput
func (g *inputTypeGenerator) generateWhereInputType(model *RuntimeModel) *graphql.InputObject {
	if _, isOk := g.whereInputMap[model.Name]; isOk {
		return g.whereInputMap[model.Name]
	}
	whereInput := g.whereInputBuilder.buildWhereInputType(model)
	g.whereInputMap[model.Name] = whereInput

	return whereInput
}

// generateUniqueWhereInputType 创建包含逻辑操作符和字段条件的WhereInput
func (g *inputTypeGenerator) generateUniqueWhereInputType(model *RuntimeModel) (*graphql.InputObject, error) {
	if _, isOk := g.uniqueWhereInputMap[model.Name]; isOk {
		return g.uniqueWhereInputMap[model.Name], nil
	}
	whereInput, err := g.whereInputBuilder.buildUniqueWhereInputType(model)
	if err != nil {
		return nil, err
	}

	g.uniqueWhereInputMap[model.Name] = whereInput
	return whereInput, err
}

// GenerateFindUniqueArgs 为findUnique查询生成参数配置
func (g *inputTypeGenerator) GenerateFindUniqueArgs(model *RuntimeModel) (graphql.FieldConfigArgument, error) {
	whereInput, err := g.generateUniqueWhereInputType(model)
	if err != nil {
		return nil, err
	}

	args := make(map[string]*graphql.ArgumentConfig)
	args["where"] = &graphql.ArgumentConfig{
		Type:         graphql.NewNonNull(whereInput),
		DefaultValue: nil,
		Description:  "FindUnique查询参数",
	}

	return args, nil
}

// GenerateFindFirstArgs 为findFirst查询生成参数配置
func (g *inputTypeGenerator) GenerateFindFirstArgs(model *RuntimeModel) graphql.FieldConfigArgument {
	whereInput := g.generateWhereInputType(model)
	orderByInput := g.generateOrderByInputType(model)

	args := make(map[string]*graphql.ArgumentConfig)
	args["where"] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "FindFirst查询参数",
	}
	args["orderBy"] = &graphql.ArgumentConfig{
		Type:         graphql.NewList(orderByInput),
		DefaultValue: nil,
		Description:  "排序参数",
	}

	return args
}

// GenerateFindManyArgs 为findMany查询生成参数配置，支持where、orderBy、take、skip等字段
func (g *inputTypeGenerator) GenerateFindManyArgs(model *RuntimeModel) graphql.FieldConfigArgument {
	whereInput := g.generateWhereInputType(model)
	orderByInput := g.generateOrderByInputType(model)

	args := make(map[string]*graphql.ArgumentConfig)
	args["where"] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "FindMany查询参数",
	}
	args["orderBy"] = &graphql.ArgumentConfig{
		Type:         graphql.NewList(orderByInput),
		DefaultValue: nil,
		Description:  "排序参数",
	}

	args["take"] = &graphql.ArgumentConfig{
		Type:         graphql.Int,
		DefaultValue: 10,
	}
	args["skip"] = &graphql.ArgumentConfig{
		Type:         graphql.Int,
		DefaultValue: 0,
	}

	return args
}

// generateOrderByInputType 创建模型特定的OrderBy输入类型
func (g *inputTypeGenerator) generateOrderByInputType(model *RuntimeModel) *graphql.InputObject {
	if _, isOk := g.orderByInputMap[model.Name]; isOk {
		return g.orderByInputMap[model.Name]
	}

	// 验证模型名称不为空
	if model == nil || model.Name == "" {
		// 返回一个空的OrderBy输入
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   "EmptyOrderByInput",
			Fields: graphql.InputObjectConfigFieldMap{},
		})
	}

	sortOrderEnum := g.generateSortOrderEnumType()
	fields := graphql.InputObjectConfigFieldMap{}

	if model.Fields != nil {
		for _, field := range model.Fields {
			// 跳过空名称字段
			if field.Name == "" {
				continue
			}
			if isQueryableField(field) {
				fields[field.Name] = &graphql.InputObjectFieldConfig{
					Type: sortOrderEnum,
				}
			}
		}
	}

	orderByInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   fmt.Sprintf("%sOrderByInput", gqlTypeName(model.Name)),
		Fields: fields,
	})

	g.orderByInputMap[model.Name] = orderByInput

	return orderByInput
}

// generateSortOrderEnumType 返回排序顺序枚举类型（asc、desc）
func (g *inputTypeGenerator) generateSortOrderEnumType() *graphql.Enum {
	sortOrderEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "SortOrder",
		Values: graphql.EnumValueConfigMap{
			"asc": &graphql.EnumValueConfig{
				Value: "asc",
			},
			"desc": &graphql.EnumValueConfig{
				Value: "desc",
			},
		},
	})

	return sortOrderEnum
}

// generateCreateInputType 创建模型特定的创建输入类型
func (g *inputTypeGenerator) generateCreateInputType(model *RuntimeModel) *graphql.InputObject {
	if _, isOk := g.createInputMap[model.Name]; isOk {
		return g.createInputMap[model.Name]
	}

	// 验证模型名称不为空
	if model == nil || model.Name == "" {
		// 返回一个空的Create输入
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   "EmptyCreateInput",
			Fields: graphql.InputObjectConfigFieldMap{},
		})
	}

	fields := graphql.InputObjectConfigFieldMap{}
	if model.Fields != nil {
		for _, field := range model.Fields {
			// 跳过空名称字段
			if field.Name == "" {
				continue
			}
			graphqlType, err := getGraphqlTypeBy(field.Type.Format)
			if err != nil {
				continue // 跳过无法处理的字段类型
			}

			// 根据NonNull和Required字段决定GraphQL类型
			var nonNullType graphql.Input = graphqlType
			if field.Required {
				nonNullType = graphql.NewNonNull(graphqlType)
			}

			fields[field.Name] = &graphql.InputObjectFieldConfig{
				Type:         nonNullType,
				Description:  field.Title,
				DefaultValue: nil,
			}
		}
	}

	createInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   fmt.Sprintf("%sCreateInput", gqlTypeName(model.Name)),
		Fields: fields,
	})

	g.createInputMap[model.Name] = createInput
	return createInput
}

// generateUpdateInputType 创建模型特定的更新输入类型
func (g *inputTypeGenerator) generateUpdateInputType(model *RuntimeModel) *graphql.InputObject {
	if _, isOk := g.updateInputMap[model.Name]; isOk {
		return g.updateInputMap[model.Name]
	}

	// 验证模型名称不为空
	if model == nil || model.Name == "" {
		// 返回一个空的Update输入
		return graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   "EmptyUpdateInput",
			Fields: graphql.InputObjectConfigFieldMap{},
		})
	}

	fields := graphql.InputObjectConfigFieldMap{}
	if model.Fields != nil {
		for _, field := range model.Fields {
			// 跳过空名称字段
			if field.Name == "" {
				continue
			}
			graphqlType, err := getGraphqlTypeBy(field.Type.Format)
			if err != nil {
				continue // 跳过无法处理的字段类型
			}

			// 更新操作中所有字段都是可选的
			fields[field.Name] = &graphql.InputObjectFieldConfig{
				Type:         graphqlType,
				Description:  field.Title,
				DefaultValue: nil,
			}
		}
	}

	updateInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   fmt.Sprintf("%sUpdateInput", gqlTypeName(model.Name)),
		Fields: fields,
	})

	g.updateInputMap[model.Name] = updateInput
	return updateInput
}

// GenerateCreateOneArgs 为createOne变更生成参数配置
func (g *inputTypeGenerator) GenerateCreateOneArgs(model *RuntimeModel) (graphql.FieldConfigArgument, error) {
	createInput := g.generateCreateInputType(model)

	args := make(map[string]*graphql.ArgumentConfig)
	args["data"] = &graphql.ArgumentConfig{
		Type:         createInput,
		DefaultValue: nil,
		Description:  "创建操作的数据输入",
	}

	return args, nil
}

// GenerateUpdateOneArgs 为updateOne变更生成参数配置
func (g *inputTypeGenerator) GenerateUpdateOneArgs(model *RuntimeModel) (graphql.FieldConfigArgument, error) {
	whereInput, err := g.generateUniqueWhereInputType(model)
	if err != nil {
		return nil, err
	}

	updateInput := g.generateUpdateInputType(model)

	args := make(map[string]*graphql.ArgumentConfig)
	args[FieldWhere] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "更新操作的查询条件",
	}
	args[FieldData] = &graphql.ArgumentConfig{
		Type:         updateInput,
		DefaultValue: nil,
		Description:  "更新操作的数据输入",
	}

	return args, nil
}

// GenerateDeleteOneArgs 为deleteOne变更生成参数配置
func (g *inputTypeGenerator) GenerateDeleteOneArgs(model *RuntimeModel) (graphql.FieldConfigArgument, error) {
	whereInput, err := g.generateUniqueWhereInputType(model)
	if err != nil {
		return nil, err
	}

	args := make(map[string]*graphql.ArgumentConfig)
	args["where"] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "删除操作的查询条件",
	}

	return args, nil
}

// GenerateCreateManyArgs 为createMany变更生成参数配置
func (g *inputTypeGenerator) GenerateCreateManyArgs(model *RuntimeModel) (graphql.FieldConfigArgument, error) {
	createInput := g.generateCreateInputType(model)

	args := make(map[string]*graphql.ArgumentConfig)
	args["data"] = &graphql.ArgumentConfig{
		Type:         graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(createInput))),
		DefaultValue: nil,
		Description:  "批量创建操作的数据输入",
	}
	args["skipDuplicates"] = &graphql.ArgumentConfig{
		Type:         graphql.Boolean,
		DefaultValue: false,
		Description: "是否跳过唯一索引冲突，默认为false。" +
			"当为true时，如果某条记录违反唯一索引约束，该记录将被跳过，继续处理后续记录",
	}

	return args, nil
}

// GenerateUpdateManyArgs 为updateMany变更生成参数配置
func (g *inputTypeGenerator) GenerateUpdateManyArgs(model *RuntimeModel) (graphql.FieldConfigArgument, error) {
	whereInput := g.generateWhereInputType(model)
	updateInput := g.generateUpdateInputType(model)

	args := make(map[string]*graphql.ArgumentConfig)
	args[FieldWhere] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "批量更新操作的查询条件，可为null表示更新所有匹配的记录",
	}
	args[FieldData] = &graphql.ArgumentConfig{
		Type:         graphql.NewNonNull(updateInput),
		DefaultValue: nil,
		Description:  "批量更新操作的数据输入，必需",
	}
	args[FieldTake] = &graphql.ArgumentConfig{
		Type:         graphql.NewNonNull(graphql.Int),
		DefaultValue: nil,
		Description:  "批量更新的最大记录数，必需，范围1-1000",
	}

	return args, nil
}

// GenerateDeleteManyArgs 为deleteMany变更生成参数配置
func (g *inputTypeGenerator) GenerateDeleteManyArgs(model *RuntimeModel) (graphql.FieldConfigArgument, error) {
	whereInput := g.generateWhereInputType(model)

	args := make(map[string]*graphql.ArgumentConfig)
	args[FieldWhere] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "批量删除操作的查询条件，可为null表示删除所有匹配的记录",
	}
	args[FieldTake] = &graphql.ArgumentConfig{
		Type:         graphql.NewNonNull(graphql.Int),
		DefaultValue: nil,
		Description:  "批量删除的最大记录数，必需，范围1-1000",
	}

	return args, nil
}

// isNumericField 判断字段是否为数值类型（支持avg、sum、min、max操作）
func isNumericField(field *modeldesign.FieldDefinition) bool {
	return field.Type.Format == modeldesign.FormatNumber ||
		field.Type.Format == modeldesign.FormatInteger ||
		field.Type.Format == modeldesign.FormatDecimal
}

// GenerateAggregateArgs 为aggregate查询生成参数配置
func (g *inputTypeGenerator) GenerateAggregateArgs(model *RuntimeModel) graphql.FieldConfigArgument {
	whereInput := g.generateWhereInputType(model)

	// 验证模型名称不为空
	modelName := "Empty"
	if model != nil && model.Name != "" {
		modelName = model.Name
	}

	// 创建 _count 输入类型（支持所有字段 + _all）
	countFields := graphql.InputObjectConfigFieldMap{}
	countFields[Field_All] = &graphql.InputObjectFieldConfig{
		Type:        graphql.Boolean,
		Description: "计算总记录数",
	}
	if model != nil && model.Fields != nil {
		for _, field := range model.Fields {
			// 跳过空名称字段
			if field.Name == "" {
				continue
			}
			if isQueryableField(field) {
				countFields[field.Name] = &graphql.InputObjectFieldConfig{
					Type:        graphql.Boolean,
					Description: fmt.Sprintf("计算%s字段的非空值数量", field.Title),
				}
			}
		}
	}
	countInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        fmt.Sprintf("%sCountAggregateInput", gqlTypeName(modelName)),
		Fields:      countFields,
		Description: "计数聚合输入",
	})

	// 创建数值聚合输入类型（仅数值字段）
	numericFields := graphql.InputObjectConfigFieldMap{}
	if model.Fields != nil {
		for _, field := range model.Fields {
			// 跳过空名称字段
			if field.Name == "" {
				continue
			}
			if isNumericField(field) {
				numericFields[field.Name] = &graphql.InputObjectFieldConfig{
					Type:        graphql.Boolean,
					Description: field.Title,
				}
			}
		}
	}

	// 初始化参数配置
	args := make(map[string]*graphql.ArgumentConfig)
	args[FieldWhere] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "聚合查询的过滤条件",
	}
	args[Field_Count] = &graphql.ArgumentConfig{
		Type:         countInputType,
		DefaultValue: nil,
		Description:  "计数聚合",
	}

	// 仅当模型包含数值字段时，才创建数值聚合输入类型
	// 避免创建空字段的InputObject导致GraphQL schema验证失败
	if len(numericFields) > 0 {
		// 创建 _avg 输入类型
		avgInputType := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:        fmt.Sprintf("%sAvgAggregateInput", gqlTypeName(modelName)),
			Fields:      numericFields,
			Description: "平均值聚合输入（仅数值字段）",
		})

		// 创建 _sum 输入类型
		sumInputType := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:        fmt.Sprintf("%sSumAggregateInput", gqlTypeName(modelName)),
			Fields:      numericFields,
			Description: "求和聚合输入（仅数值字段）",
		})

		// 创建 _min 输入类型
		minInputType := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:        fmt.Sprintf("%sMinAggregateInput", gqlTypeName(modelName)),
			Fields:      numericFields,
			Description: "最小值聚合输入（仅数值字段）",
		})

		// 创建 _max 输入类型
		maxInputType := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:        fmt.Sprintf("%sMaxAggregateInput", gqlTypeName(modelName)),
			Fields:      numericFields,
			Description: "最大值聚合输入（仅数值字段）",
		})

		args[Field_Avg] = &graphql.ArgumentConfig{
			Type:         avgInputType,
			DefaultValue: nil,
			Description:  "平均值聚合",
		}
		args[Field_Sum] = &graphql.ArgumentConfig{
			Type:         sumInputType,
			DefaultValue: nil,
			Description:  "求和聚合",
		}
		args[Field_Min] = &graphql.ArgumentConfig{
			Type:         minInputType,
			DefaultValue: nil,
			Description:  "最小值聚合",
		}
		args[Field_Max] = &graphql.ArgumentConfig{
			Type:         maxInputType,
			DefaultValue: nil,
			Description:  "最大值聚合",
		}
	}

	return args
}

// GenerateCountArgs 为count查询生成参数配置
// count 支持 where 参数进行过滤，以及 select 参数用于字段级计数
func (g *inputTypeGenerator) GenerateCountArgs(model *RuntimeModel) graphql.FieldConfigArgument {
	whereInput := g.generateWhereInputType(model)

	// 创建 select 输入类型（支持所有字段 + _all）
	selectFields := graphql.InputObjectConfigFieldMap{}
	selectFields[Field_All] = &graphql.InputObjectFieldConfig{
		Type:        graphql.Boolean,
		Description: "计算总记录数 COUNT(*)",
	}
	if model.Fields != nil {
		for _, field := range model.Fields {
			if isQueryableField(field) {
				selectFields[field.Name] = &graphql.InputObjectFieldConfig{
					Type:        graphql.Boolean,
					Description: fmt.Sprintf("计算%s字段的非空值数量 COUNT(%s)", field.Title, field.Name),
				}
			}
		}
	}
	selectInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        fmt.Sprintf("%sCountSelectInput", gqlTypeName(model.Name)),
		Fields:      selectFields,
		Description: "Count操作的字段选择输入",
	})

	// 初始化参数配置
	args := make(map[string]*graphql.ArgumentConfig)
	args[FieldWhere] = &graphql.ArgumentConfig{
		Type:         whereInput,
		DefaultValue: nil,
		Description:  "Count查询的过滤条件",
	}
	args[FieldSelect] = &graphql.ArgumentConfig{
		Type:         selectInputType,
		DefaultValue: nil,
		Description:  "选择要计数的字段（默认返回总记录数）",
	}

	return args
}
