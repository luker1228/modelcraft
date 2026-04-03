package generators

import (
	"fmt"
	"modelcraft/pkg/schema/core"
	"strings"

	"github.com/graphql-go/graphql"
)

// GraphQLSchemaGenerator GraphQL Schema生成器
type GraphQLSchemaGenerator struct {
	typeFactory *GraphQLTypeFactory
}

// NewGraphQLSchemaGenerator 创建新的GraphQL Schema生成器
func NewGraphQLSchemaGenerator() *GraphQLSchemaGenerator {
	return &GraphQLSchemaGenerator{
		typeFactory: NewGraphQLTypeFactory(),
	}
}

// GenerateGraphQLSchema 生成GraphQL Schema
func (g *GraphQLSchemaGenerator) GenerateGraphQLSchema(model *core.ModelDefinition) (*core.GraphQLSchemaResult, error) {
	if !model.IsValid() {
		return nil, fmt.Errorf("invalid modeldesign definition: %s", model.Name)
	}

	// 1. 构建GraphQL字段
	graphqlFields := graphql.Fields{}

	for _, field := range model.Fields {
		graphqlType := g.typeFactory.CreateFieldType(field)
		graphqlFields[field.Key] = &graphql.Field{
			Type:        graphqlType,
			Description: field.Description,
			Resolve:     g.typeFactory.CreateResolver(field),
		}
	}

	// 2. 创建对象类型
	objectType := graphql.NewObject(graphql.ObjectConfig{
		Name:        model.Name,
		Description: model.Description,
		Fields:      graphqlFields,
	})

	// 3. 构建查询和变更
	queryFields := g.buildQueryFields(objectType, model)
	mutationFields := g.buildMutationFields(objectType, model)

	// 4. 创建Schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: queryFields}),
		Mutation: graphql.NewObject(graphql.ObjectConfig{Name: "Mutation", Fields: mutationFields}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL modeldesign: %w", err)
	}

	// 5. 生成SDL
	sdl := g.generateSDL(&schema, model)

	return &core.GraphQLSchemaResult{
		Schema: &schema,
		SDL:    sdl,
	}, nil
}

// buildQueryFields 构建查询字段
func (g *GraphQLSchemaGenerator) buildQueryFields(
	objectType *graphql.Object,
	model *core.ModelDefinition,
) graphql.Fields {
	return graphql.Fields{
		"get": &graphql.Field{
			Type: objectType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: g.createGetResolver(model),
		},
		"list": &graphql.Field{
			Type: graphql.NewList(objectType),
			Args: graphql.FieldConfigArgument{
				"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
				"offset": &graphql.ArgumentConfig{Type: graphql.Int},
			},
			Resolve: g.createListResolver(model),
		},
	}
}

// buildMutationFields 构建变更字段
func (g *GraphQLSchemaGenerator) buildMutationFields(
	objectType *graphql.Object,
	model *core.ModelDefinition,
) graphql.Fields {
	// 为 update 操作创建包装类型
	updateResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Update" + model.Name + ResultTypeSuffix,
		Fields: graphql.Fields{
			FieldSuccess: &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			FieldUpdatedObj: &graphql.Field{
				Type: objectType,
			},
		},
	})

	return graphql.Fields{
		"create" + model.Name: &graphql.Field{
			Type:    objectType,
			Args:    g.buildInputArgs(model),
			Resolve: g.createMutationResolver(model, "create"),
		},
		"update" + model.Name: &graphql.Field{
			Type:    updateResultType,
			Args:    g.buildUpdateArgs(model),
			Resolve: g.createMutationResolver(model, "update"),
		},
		"delete" + model.Name: &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				FieldID: &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: g.createMutationResolver(model, "delete"),
		},
	}
}

// buildInputArgs 构建输入参数
func (g *GraphQLSchemaGenerator) buildInputArgs(model *core.ModelDefinition) graphql.FieldConfigArgument {
	args := graphql.FieldConfigArgument{}

	for _, field := range model.Fields {
		fieldType := g.typeFactory.CreateFieldType(field)
		args[field.Key] = &graphql.ArgumentConfig{
			Type:        fieldType,
			Description: field.Description,
		}
	}

	return args
}

// buildUpdateArgs 构建更新参数
func (g *GraphQLSchemaGenerator) buildUpdateArgs(model *core.ModelDefinition) graphql.FieldConfigArgument {
	args := graphql.FieldConfigArgument{
		FieldID: &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
		FieldReturnUpdatedObj: &graphql.ArgumentConfig{
			Type:        graphql.Boolean,
			Description: "是否返回完整的更新后对象，默认 false",
		},
	}

	for _, field := range model.Fields {
		// 更新时字段都是可选的
		// 注意：ID 作为 update 的定位参数单独存在，这里不要再重复生成一遍
		if field.Key == FieldID {
			continue
		}

		baseType := field.Type.GetGraphQLType()
		args[field.Key] = &graphql.ArgumentConfig{
			Type:        baseType, // 不使用NonNull
			Description: field.Description,
		}
	}

	return args
}

// createGetResolver 创建获取解析器
func (g *GraphQLSchemaGenerator) createGetResolver(model *core.ModelDefinition) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		idValue, exists := p.Args["id"]
		if !exists {
			return nil, fmt.Errorf("missing id argument")
		}
		id, ok := idValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid id argument type")
		}

		// 这里应该调用实际的数据访问层
		// 现在返回模拟数据
		result := map[string]interface{}{
			"id": id,
		}

		// 为每个字段设置默认值
		for _, field := range model.Fields {
			if field.Default != nil {
				result[field.Key] = field.Default
			} else {
				result[field.Key] = field.Type.GetDefaultValue()
			}
		}

		return result, nil
	}
}

// createListResolver 创建列表解析器
func (g *GraphQLSchemaGenerator) createListResolver(model *core.ModelDefinition) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		limit := 10
		offset := 0

		if l, ok := p.Args["limit"].(int); ok && l > 0 {
			limit = l
		}
		if o, ok := p.Args["offset"].(int); ok && o >= 0 {
			offset = o
		}

		// 这里应该调用实际的数据访问层
		// 现在返回模拟数据
		var results []interface{}
		for i := offset; i < offset+limit; i++ {
			item := map[string]interface{}{
				"id": fmt.Sprintf("%d", i+1),
			}

			// 为每个字段设置默认值
			for _, field := range model.Fields {
				if field.Default != nil {
					item[field.Key] = field.Default
				} else {
					item[field.Key] = field.Type.GetDefaultValue()
				}
			}

			results = append(results, item)
		}

		return results, nil
	}
}

// createMutationResolver 创建变更解析器
func (g *GraphQLSchemaGenerator) createMutationResolver(
	model *core.ModelDefinition,
	operation string,
) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		switch operation {
		case "create":
			return g.handleCreate(p, model)
		case "update":
			return g.handleUpdate(p, model)
		case "delete":
			return g.handleDelete(p, model)
		default:
			return nil, fmt.Errorf("unsupported operation: %s", operation)
		}
	}
}

// handleCreate 处理创建操作
func (g *GraphQLSchemaGenerator) handleCreate(
	p graphql.ResolveParams,
	model *core.ModelDefinition,
) (interface{}, error) {
	result := map[string]interface{}{
		"id": "new-id", // 这里应该生成实际的ID
	}

	// 从参数中获取字段值
	for _, field := range model.Fields {
		if value, exists := p.Args[field.Key]; exists {
			converted, err := g.typeFactory.validateAndConvert(value, field)
			if err != nil {
				return nil, fmt.Errorf("invalid value for field %s: %w", field.Key, err)
			}
			result[field.Key] = converted
		} else if field.Required {
			return nil, fmt.Errorf("required field %s is missing", field.Key)
		} else if field.Default != nil {
			result[field.Key] = field.Default
		}
	}

	return result, nil
}

// handleUpdate 处理更新操作
func (g *GraphQLSchemaGenerator) handleUpdate(
	p graphql.ResolveParams,
	model *core.ModelDefinition,
) (interface{}, error) {
	idValue, exists := p.Args[FieldID]
	if !exists {
		return nil, fmt.Errorf("missing %s argument", FieldID)
	}
	id, ok := idValue.(string)
	if !ok {
		return nil, fmt.Errorf("invalid %s argument type", FieldID)
	}
	returnUpdatedObj := false

	if returnVal, exists := p.Args[FieldReturnUpdatedObj]; exists {
		if returnBool, ok := returnVal.(bool); ok {
			returnUpdatedObj = returnBool
		}
	}

	updatedObj := map[string]interface{}{
		FieldID: id,
	}

	// 从参数中获取字段值
	for _, field := range model.Fields {
		if value, exists := p.Args[field.Key]; exists {
			converted, err := g.typeFactory.validateAndConvert(value, field)
			if err != nil {
				return nil, fmt.Errorf("invalid value for field %s: %w", field.Key, err)
			}
			updatedObj[field.Key] = converted
		}
	}

	// 返回包装结果
	result := map[string]interface{}{
		FieldSuccess: true,
	}

	if returnUpdatedObj {
		result[FieldUpdatedObj] = updatedObj
	}

	return result, nil
}

// handleDelete 处理删除操作
func (g *GraphQLSchemaGenerator) handleDelete(
	p graphql.ResolveParams,
	model *core.ModelDefinition,
) (interface{}, error) {
	// id := p.Args["id"].(string)

	// 这里应该调用实际的删除逻辑
	// 现在直接返回成功
	return true, nil
}

// generateSDL 生成Schema Definition Language
func (g *GraphQLSchemaGenerator) generateSDL(
	schema *graphql.Schema,
	model *core.ModelDefinition,
) string {
	var sdl strings.Builder

	// 生成类型定义
	sdl.WriteString(fmt.Sprintf("type %s {\n", model.Name))
	for _, field := range model.Fields {
		fieldType := g.getSDLType(field)
		sdl.WriteString(fmt.Sprintf("  %s: %s", field.Key, fieldType))
		if field.Description != "" {
			sdl.WriteString(fmt.Sprintf(" # %s", field.Description))
		}
		sdl.WriteString("\n")
	}
	sdl.WriteString("}\n\n")

	// 生成更新结果类型
	sdl.WriteString(fmt.Sprintf("type Update%s%s {\n", model.Name, ResultTypeSuffix))
	sdl.WriteString(fmt.Sprintf("  %s: Boolean!\n", FieldSuccess))
	sdl.WriteString(fmt.Sprintf("  %s: %s\n", FieldUpdatedObj, model.Name))
	sdl.WriteString("}\n\n")

	// 生成查询定义
	sdl.WriteString("type Query {\n")
	sdl.WriteString(fmt.Sprintf("  get%s(id: ID!): %s\n", model.Name, model.Name))
	sdl.WriteString(fmt.Sprintf("  list%s(limit: Int, offset: Int): [%s]\n", model.Name, model.Name))
	sdl.WriteString("}\n\n")

	// 生成变更定义
	sdl.WriteString("type Mutation {\n")
	sdl.WriteString(fmt.Sprintf("  create%s(", model.Name))

	inputFields := make([]string, 0, len(model.Fields))
	for _, field := range model.Fields {
		fieldType := g.getSDLType(field)
		inputFields = append(inputFields, fmt.Sprintf("%s: %s", field.Key, fieldType))
	}
	sdl.WriteString(strings.Join(inputFields, ", "))
	sdl.WriteString(fmt.Sprintf("): %s\n", model.Name))

	sdl.WriteString(fmt.Sprintf("  update%s(%s: ID!, %s: Boolean, ", model.Name, FieldID, FieldReturnUpdatedObj))
	updateFields := make([]string, 0, len(model.Fields))
	for _, field := range model.Fields {
		// ID 作为 update 的定位参数单独存在，这里不要再重复生成一遍
		if field.Key == FieldID {
			continue
		}
		baseType := field.Type.GetSDLType()
		updateFields = append(updateFields, fmt.Sprintf("%s: %s", field.Key, baseType))
	}
	sdl.WriteString(strings.Join(updateFields, ", "))
	sdl.WriteString(fmt.Sprintf("): Update%s%s\n", model.Name, ResultTypeSuffix))

	sdl.WriteString(fmt.Sprintf("  delete%s(id: ID!): Boolean\n", model.Name))
	sdl.WriteString("}\n")

	return sdl.String()
}

// getSDLType 获取SDL类型
func (g *GraphQLSchemaGenerator) getSDLType(field *core.FieldDefinition) string {
	baseType := field.Type.GetSDLType()
	if field.Required {
		return baseType + "!"
	}
	return baseType
}
