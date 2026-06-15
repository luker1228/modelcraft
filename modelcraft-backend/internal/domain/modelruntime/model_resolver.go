package modelruntime

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"modelcraft/pkg/requestcontext"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/spf13/cast"
)

type graphqlEnumConfig struct {
	enumType *graphql.Enum
}

// graphqlModelResolver GraphQL模型解析器实现，用于生成GraphQL Schema并处理查询和变更。
// 无状态的 Schema 构建器：不持有任何 context，所有状态通过参数传递。
// - Schema 构建阶段：ctx 作为参数透传（用于日志和 Repository 查询）
// - 请求执行阶段：请求级状态通过 graphqlRequestContext 从 p.Context 读取
type graphqlModelResolver struct {
	model              *RuntimeModel
	enumConfigMap      map[string]*graphqlEnumConfig
	inputTypeGenerator *inputTypeGenerator
	modelRepo          ModelRepository
	lfkRepo            modeldesign.LogicalForeignKeyRepository
}

func newGraphqlModelResolver(ctx context.Context, model *RuntimeModel,
	modelRepo ModelRepository, lfkRepo modeldesign.LogicalForeignKeyRepository,
) *graphqlModelResolver {
	return &graphqlModelResolver{
		model:              model,
		modelRepo:          modelRepo,
		lfkRepo:            lfkRepo,
		inputTypeGenerator: newInputTypeGenerator(),
		enumConfigMap:      make(map[string]*graphqlEnumConfig),
	}
}

func (m *graphqlModelResolver) newGraphqlSchema(ctx context.Context) (*graphql.Schema, error) {
	logger := logfacade.GetLogger(ctx)

	modelType, err := m.createModelType(ctx)
	if err != nil {
		logger.Error(ctx, "createModelType_fail", logfacade.Err(err))
		return nil, bizerrors.New("createModelType_fail")
	}
	rootQuery, err := m.createRootQuery(ctx, modelType)
	if err != nil {
		logger.Error(ctx, "createRootQuery_fail", logfacade.Err(err))
		return nil, bizerrors.New("createRootQuery_fail")
	}

	baseModelType, err := m.generateModelTypeSkipRelation(ctx, m.model)
	if err != nil {
		logger.Error(ctx, "generateModelTypeSkipRelation_fail", logfacade.Err(err))
		return nil, bizerrors.New("generateModelTypeSkipRelation_fail")
	}
	rootMutation, err := m.createRootMutation(baseModelType)
	if err != nil {
		logger.Error(ctx, "createRootMutation_fail", logfacade.Err(err))
		return nil, bizerrors.New("createRootMutation_fail")
	}
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})

	return &schema, err
}

func (m *graphqlModelResolver) executeFindUnique(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionSelect); err != nil {
		return nil, err
	}
	startTime := time.Now()

	input, err := newFindUniqueInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}
	result, err := rctx.ClientRepo.FindUnique(p.Context, input)
	if err != nil {
		if bizerrors.Is(err, sql.ErrNoRows) {
			result = nil
		} else {
			return nil, err
		}
	}

	// Get metadata from context
	metadata := requestcontext.GetMetadata(p.Context)
	reqId := ""
	if metadata != nil {
		reqId = metadata.ReqID
	}

	// Calculate time cost
	timeCost := int(time.Since(startTime).Milliseconds())

	// Wrap result with metadata
	return map[string]any{
		FieldItem:     result,
		FieldTimeCost: timeCost,
		FieldReqId:    reqId,
	}, nil
}

func (m *graphqlModelResolver) createFindUniqueField(modelType graphql.Type) (*graphql.Field, error) {
	args, err := m.inputTypeGenerator.GenerateFindUniqueArgs(m.model)
	if err != nil {
		return nil, err
	}

	// Create result wrapper type
	resultType := m.createFindUniqueResultType(modelType)

	return &graphql.Field{
		Type: resultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)
			result, err2 := m.executeFindUnique(p)
			if err2 != nil {
				logger.Error(p.Context, "find_unique_fail", logfacade.Err(err2))
			} else {
				logger.Infof(p.Context, "find_unqiue_result=%+v", result)
			}
			return result, err2
		},
	}, nil
}

func (m *graphqlModelResolver) createRootQuery(ctx context.Context, modelType graphql.Type) (*graphql.Object, error) {
	findUniqueFields, err := m.createFindUniqueField(modelType)
	if err != nil {
		return nil, err
	}
	findFirstFields, err := m.createFindFirstField(modelType)
	if err != nil {
		return nil, err
	}
	findManyFields, err := m.createFindManyField(modelType)
	if err != nil {
		return nil, err
	}
	aggregateField, err := m.createAggregateField(ctx)
	if err != nil {
		return nil, err
	}
	countField, err := m.createCountField(ctx)
	if err != nil {
		return nil, err
	}
	listByCursorField, err := m.createListByCursorField(modelType)
	if err != nil {
		return nil, err
	}
	listByPageField, err := m.createListByPageField(modelType)
	if err != nil {
		return nil, err
	}
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			OperationFindUnique:   findUniqueFields,
			OperationFindFirst:    findFirstFields,
			OperationFindMany:     findManyFields,
			OperationAggregate:    aggregateField,
			OperationCount:        countField,
			OperationListByCursor: listByCursorField,
			OperationListByPage:   listByPageField,
		},
	})
	return rootQuery, nil
}

func (m *graphqlModelResolver) executeFindFirst(p graphql.ResolveParams) (any, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionSelect); err != nil {
		return nil, err
	}
	startTime := time.Now()

	input, err := newFindFirstInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}
	result, err := rctx.ClientRepo.FindFirst(p.Context, input)
	if err != nil {
		if bizerrors.Is(err, sql.ErrNoRows) {
			result = nil
		} else {
			return nil, err
		}
	}

	// Get metadata from context
	metadata := requestcontext.GetMetadata(p.Context)
	reqId := ""
	if metadata != nil {
		reqId = metadata.ReqID
	}

	// Calculate time cost
	timeCost := int(time.Since(startTime).Milliseconds())

	// Wrap result with metadata
	return map[string]any{
		FieldItem:     result,
		FieldTimeCost: timeCost,
		FieldReqId:    reqId,
	}, nil
}

func (m *graphqlModelResolver) createFindFirstField(modelType graphql.Type) (*graphql.Field, error) {
	args := m.inputTypeGenerator.GenerateFindFirstArgs(m.model)

	// Create result wrapper type
	resultType := m.createFindFirstResultType(modelType)

	return &graphql.Field{
		Type: resultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := m.executeFindFirst(p)
			if err != nil {
				logfacade.GetLogger(p.Context).Error(p.Context, "executeFindFirst fail", logfacade.Err(err))
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) executeFindMany(p graphql.ResolveParams) (map[string]any, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionSelect); err != nil {
		return nil, err
	}
	startTime := time.Now()

	input, err := newFindManyInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}
	result, err := rctx.ClientRepo.FindMany(p.Context, input)
	if err != nil {
		if bizerrors.Is(err, sql.ErrNoRows) {
			result = []map[string]any{}
		} else {
			return nil, err
		}
	}

	// Count total matching rows (same WHERE conditions, no LIMIT/OFFSET).
	var totalCount *int
	countResult, countErr := rctx.ClientRepo.Count(p.Context, &CountInput{
		TableName:  input.TableName,
		Where:      input.Where,
		RawFilters: input.RawFilters,
	})
	if countErr == nil {
		if v, ok := countResult[FieldCount]; ok {
			if n, castErr := cast.ToIntE(v); castErr == nil {
				totalCount = &n
			}
		}
	}

	// Get metadata from context
	metadata := requestcontext.GetMetadata(p.Context)
	reqId := ""
	if metadata != nil {
		reqId = metadata.ReqID
	}

	// Calculate time cost
	timeCost := int(time.Since(startTime).Milliseconds())

	return map[string]any{
		FieldItems:      result,
		FieldTotalCount: totalCount,
		FieldTimeCost:   timeCost,
		FieldReqId:      reqId,
	}, nil
}

func (m *graphqlModelResolver) createFindManyField(modelType graphql.Type) (*graphql.Field, error) {
	args := m.inputTypeGenerator.GenerateFindManyArgs(m.model)

	// Create result wrapper type
	resultType := m.createFindManyResultType(modelType)

	return &graphql.Field{
		Type: resultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := m.executeFindMany(p)
			if err != nil {
				logfacade.GetLogger(p.Context).Error(p.Context, "executeFindMany fail", logfacade.Err(err))
				return nil, err
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) executeAggregate(p graphql.ResolveParams) (map[string]any, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionSelect); err != nil {
		return nil, err
	}
	input, err := newAggregateInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}
	result, err := rctx.ClientRepo.Aggregate(p.Context, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *graphqlModelResolver) createAggregateField(ctx context.Context) (*graphql.Field, error) {
	args := m.inputTypeGenerator.GenerateAggregateArgs(m.model)

	// 创建聚合结果类型
	aggregateResultType := m.createAggregateResultType(ctx)

	return &graphql.Field{
		Type: aggregateResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := m.executeAggregate(p)
			if err != nil {
				logfacade.GetLogger(p.Context).Error(p.Context, "executeAggregate fail", logfacade.Err(err))
				return nil, err
			}
			return result, nil
		},
	}, nil
}

func (m *graphqlModelResolver) createAggregateResultType(ctx context.Context) *graphql.Object {
	logger := logfacade.GetLogger(ctx)

	// 验证模型名称不为空
	if m.model == nil || m.model.Name == "" {
		logger.Error(ctx, "model name is empty when creating aggregate result type")
		// 返回一个空的聚合结果类型
		return graphql.NewObject(graphql.ObjectConfig{
			Name:        "EmptyAggregateResult",
			Fields:      graphql.Fields{},
			Description: "Empty aggregate result",
		})
	}

	// 创建 _count 结果类型
	countFields := graphql.Fields{}
	countFields[Field_All] = &graphql.Field{
		Type:        graphql.Int,
		Description: "总记录数",
	}
	for _, field := range m.model.Fields {
		// 跳过空名称字段
		if field.Name == "" {
			logger.Warnf(ctx, "skip field with empty name in model %s", m.model.Name)
			continue
		}
		if isQueryableField(field) {
			countFields[field.Name] = &graphql.Field{
				Type:        graphql.Int,
				Description: field.Title + "字段的非空值数量",
			}
		}
	}
	countResultType := graphql.NewObject(graphql.ObjectConfig{
		Name:        gqlTypeName(m.model.Name) + "CountAggregateResult",
		Fields:      countFields,
		Description: "计数聚合结果",
	})

	// 创建数值聚合结果类型（仅数值字段）
	numericFields := graphql.Fields{}
	for _, field := range m.model.Fields {
		// 跳过空名称字段
		if field.Name == "" {
			continue
		}
		if isNumericField(field) {
			graphqlType, err := getGraphqlTypeBy(field.Type.Format)
			if err != nil {
				continue
			}
			numericFields[field.Name] = &graphql.Field{
				Type:        graphqlType,
				Description: field.Title,
			}
		}
	}

	// 创建聚合结果类型
	aggregateFields := graphql.Fields{}
	aggregateFields[Field_Count] = &graphql.Field{
		Type:        countResultType,
		Description: "计数聚合结果",
	}
	if len(numericFields) > 0 {
		aggregateFields[Field_Avg] = &graphql.Field{
			Type: graphql.NewObject(graphql.ObjectConfig{
				Name:        gqlTypeName(m.model.Name) + "AvgAggregateResult",
				Fields:      numericFields,
				Description: "平均值聚合结果",
			}),
			Description: "平均值聚合结果",
		}
		aggregateFields[Field_Sum] = &graphql.Field{
			Type: graphql.NewObject(graphql.ObjectConfig{
				Name:        gqlTypeName(m.model.Name) + "SumAggregateResult",
				Fields:      numericFields,
				Description: "求和聚合结果",
			}),
			Description: "求和聚合结果",
		}
		aggregateFields[Field_Min] = &graphql.Field{
			Type: graphql.NewObject(graphql.ObjectConfig{
				Name:        gqlTypeName(m.model.Name) + "MinAggregateResult",
				Fields:      numericFields,
				Description: "最小值聚合结果",
			}),
			Description: "最小值聚合结果",
		}
		aggregateFields[Field_Max] = &graphql.Field{
			Type: graphql.NewObject(graphql.ObjectConfig{
				Name:        gqlTypeName(m.model.Name) + "MaxAggregateResult",
				Fields:      numericFields,
				Description: "最大值聚合结果",
			}),
			Description: "最大值聚合结果",
		}
	}

	aggregateResultType := graphql.NewObject(graphql.ObjectConfig{
		Name:        gqlTypeName(m.model.Name) + "AggregateResult",
		Fields:      aggregateFields,
		Description: "聚合查询结果",
	})

	return aggregateResultType
}

// executeCount 执行count查询操作
func (m *graphqlModelResolver) executeCount(p graphql.ResolveParams) (map[string]any, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionSelect); err != nil {
		return nil, err
	}
	input, err := newCountInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}

	// Conflict check: using `select` switches to per-field counting mode, which
	// populates `fieldsCount` but NOT `count`. Querying both in one call is
	// meaningless — `count` will always be null — so we reject it explicitly.
	//
	// Fix: use one of the two mutually-exclusive forms:
	//   (a) { count { count } }                          — simple total
	//   (b) { count(select: {_all: true}) { fieldsCount { _all } } } — per-field totals
	if len(input.Select) > 0 && hasSelectedField(p, FieldCount) {
		return nil, bizerrors.NewError(
			bizerrors.ParamInvalid,
			"'count' and 'select' are mutually exclusive: "+
				"use { count { count } } for a simple total, or "+
				"{ count(select: {_all: true}) { fieldsCount { _all } } } for per-field counts",
		)
	}

	result, err := rctx.ClientRepo.Count(p.Context, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *graphqlModelResolver) createCountField(ctx context.Context) (*graphql.Field, error) {
	args := m.inputTypeGenerator.GenerateCountArgs(m.model)

	// 创建count结果类型（动态返回类型，根据是否有select参数）
	countResultType := m.createCountResultType(ctx)

	return &graphql.Field{
		Type: countResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := m.executeCount(p)
			if err != nil {
				logfacade.GetLogger(p.Context).Error(p.Context, "executeCount fail", logfacade.Err(err))
				return nil, err
			}
			return result, nil
		},
	}, nil
}

// createCountResultType 创建count查询的结果类型
// 支持两种返回格式：
// 1. 简单计数：{ count: Int! }
// 2. 字段级计数：{ fieldsCount: { _all: Int, field1: Int, ... } }
func (m *graphqlModelResolver) createCountResultType(ctx context.Context) *graphql.Object {
	logger := logfacade.GetLogger(ctx)

	// 验证模型名称不为空
	if m.model == nil || m.model.Name == "" {
		logger.Error(ctx, "model name is empty when creating count result type")
		// 返回一个空的计数结果类型
		return graphql.NewObject(graphql.ObjectConfig{
			Name:        "EmptyCountResult",
			Fields:      graphql.Fields{},
			Description: "Empty count result",
		})
	}

	// 创建字段级计数类型（包含 _all 和所有可查询字段）
	fieldsCountFields := graphql.Fields{}
	fieldsCountFields[Field_All] = &graphql.Field{
		Type:        graphql.Int,
		Description: "总记录数 COUNT(*)",
	}
	if m.model.Fields != nil {
		for _, field := range m.model.Fields {
			// 跳过空名称字段
			if field.Name == "" {
				logger.Warnf(ctx, "skip field with empty name in model %s", m.model.Name)
				continue
			}
			if isQueryableField(field) {
				fieldsCountFields[field.Name] = &graphql.Field{
					Type:        graphql.Int,
					Description: fmt.Sprintf("%s字段的非空值数量", field.Title),
				}
			}
		}
	}
	fieldsCountType := graphql.NewObject(graphql.ObjectConfig{
		Name:        m.model.Name + "FieldsCount",
		Fields:      fieldsCountFields,
		Description: "字段级计数结果",
	})

	// 创建count结果类型
	countResultFields := graphql.Fields{}
	countResultFields[FieldCount] = &graphql.Field{
		Type:        graphql.Int,
		Description: "简单计数结果（不使用select时）",
	}
	countResultFields[FieldFieldsCount] = &graphql.Field{
		Type:        fieldsCountType,
		Description: "字段级计数结果（使用select时）",
	}

	countResultType := graphql.NewObject(graphql.ObjectConfig{
		Name:        gqlTypeName(m.model.Name) + "CountResult",
		Fields:      countResultFields,
		Description: "Count查询结果",
	})

	return countResultType
}

// createField 创建GraphQL字段，支持普通字段、关系字段和虚拟字段
func (r *graphqlModelResolver) createField(
	ctx context.Context,
	maxDepth int,
	field *RuntimeField,
	relateObjMaps map[string]*graphql.Object,
) (*graphql.Field, error) {
	graphqlField := &graphql.Field{
		Name:        field.Name,
		Description: field.Title,
	}

	if field.IsEnumField() {
		return r.createEnumField(field, graphqlField)
	}

	if field.IsRelationField() {
		return r.createRelationField(ctx, maxDepth, field, relateObjMaps, graphqlField)
	}

	return r.createScalarField(ctx, field, graphqlField)
}

// createScalarField 创建标量字段
func (r *graphqlModelResolver) createScalarField(ctx context.Context, field *RuntimeField, graphqlField *graphql.Field,
) (*graphql.Field, error) {
	graphqlType, err := getGraphqlTypeBy(field.Type.Format)
	if err != nil {
		logfacade.GetLogger(ctx).Errorf(ctx, "failed to get GraphQL type for field %s: %w", field.Name, err)
		return nil, bizerrors.Errorf("failed to get GraphQL type for field %s format %s", field.Name, field.Type.Format)
	}
	graphqlField.Type = graphqlType
	return graphqlField, nil
}

// createEnumField 创建枚举字段
func (r *graphqlModelResolver) createEnumField(field *RuntimeField, graphqlField *graphql.Field,
) (*graphql.Field, error) {
	enumConfig, err := r.getEnumConfig(field)
	if err != nil {
		return nil, err
	}
	graphqlField.Type = enumConfig.enumType
	// 设置解析器，验证数据库返回的字符串值是否为有效的枚举值
	graphqlField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
		logger := logfacade.GetLogger(p.Context)

		// 获取父对象记录
		record, ok := p.Source.(map[string]any)
		if !ok {
			logger.Warnf(p.Context, "invalid source type for enum field %s", field.Name)
			return nil, nil
		}
		logger.Infof(p.Context, "process enum fields: %s, record=%s",
			field.Name, bizutils.MarshalToStringIgnoreErr(record))

		// 获取字段值
		value, exists := record[field.Name]
		if !exists || value == nil {
			return nil, nil
		}

		// 转换为字符串
		strValue, err := cast.ToStringE(value)
		if err != nil {
			logger.Warnf(p.Context, "enum field %s has invalid type: %T", field.Name, value)
			return nil, bizerrors.Errorf("undefined enum: value %v is not a valid string", value)
		}

		// 检查值是否在枚举选项中
		enumDefinition := field.Enum
		if enumDefinition == nil {
			logger.Warnf(p.Context, "enum field %s has no enum definition", field.Name)
			return nil, bizerrors.Errorf("undefined enum: enum definition not found for field %s", field.Name)
		}

		// 检查枚举选项中是否包含该值
		if !enumDefinition.HasOptionCode(strValue) {
			logger.Warnf(p.Context, "enum field %s: value %s is not a valid option for enum %s", field.Name, strValue,
				enumDefinition.Name)
			return nil,
				bizerrors.Errorf("undefined enum: %s is not a valid value for enum %s", strValue, enumDefinition.Name)
		}

		logger.Infof(p.Context, "enum field %s: value %s is valid for enum %s",
			field.Name, strValue, enumDefinition.Name)
		// 返回有效的枚举值（GraphQL 会自动解析为枚举类型）
		return strValue, nil
	}
	return graphqlField, nil
}

// getEnumConfig 获取枚举的GraphQL类型配置
// 使用类型缓存确保每次返回相同的类型
func (r *graphqlModelResolver) getEnumConfig(field *RuntimeField) (*graphqlEnumConfig, error) {
	// 检查是否已定义该类型
	if enumCfg, ok := r.enumConfigMap[field.EnumName]; ok {
		return enumCfg, nil
	}
	enumDefinition := field.Enum
	// 验证枚举名称不为空
	if enumDefinition.Name == "" {
		return nil, bizerrors.Errorf("enum field %s has empty enum name", field.Name)
	}

	enumValueCfgMap := graphql.EnumValueConfigMap{}
	for _, opt := range enumDefinition.Options {
		// 验证枚举值名称不为空
		if opt.Code == "" {
			return nil, bizerrors.Errorf("enum %s has option with empty code", enumDefinition.Name)
		}
		enumValueCfgMap[opt.Code] = &graphql.EnumValueConfig{
			Value:       opt.Code, // 使用实际的枚举值（字符串本身），而不是graphql.String类型
			Description: opt.Label,
		}
	}
	// 定义枚举类型
	enumType := graphql.NewEnum(graphql.EnumConfig{
		Name:        enumDefinition.Name,
		Description: enumDefinition.Description,
		Values:      enumValueCfgMap,
	})

	enumModel := &graphqlEnumConfig{
		enumType: enumType,
	}
	r.enumConfigMap[field.EnumName] = enumModel
	return enumModel, nil
}

func normalizeEnumArrayCodes(sourceValue any) ([]string, bool) {
	if codes, ok := sourceValue.([]string); ok {
		return codes, true
	}

	strSlice, ok := sourceValue.([]interface{})
	if !ok {
		return nil, false
	}

	codes := make([]string, 0, len(strSlice))
	for _, v := range strSlice {
		if s, ok := v.(string); ok {
			codes = append(codes, s)
		}
	}
	return codes, true
}

func normalizeEnumCode(sourceValue any) (string, bool) {
	code, ok := sourceValue.(string)
	if !ok || code == "" {
		return "", false
	}
	return code, true
}

func (r *graphqlModelResolver) injectAutoEnumLabelFields(
	ctx context.Context,
	model *RuntimeModel,
	graphqlFields graphql.Fields,
) {
	logger := logfacade.GetLogger(ctx)

	for _, field := range model.Fields {
		if field == nil || field.Name == "" || !field.IsEnumField() {
			continue
		}

		autoFieldName, autoField, enabled := r.newAutoEnumLabelField(field)
		if !enabled {
			continue
		}
		if _, exists := graphqlFields[autoFieldName]; exists {
			logger.Warnf(
				ctx,
				"skip auto enum label field %s for source field %s: field already exists",
				autoFieldName,
				field.Name,
			)
			continue
		}
		graphqlFields[autoFieldName] = autoField
	}
}

func (r *graphqlModelResolver) newAutoEnumLabelField(sourceField *RuntimeField) (string, *graphql.Field, bool) {
	autoFieldName, enabled := sourceField.ResolveEnumDisplayFieldName()
	if !enabled {
		return "", nil, false
	}

	autoFieldType := graphql.Output(graphql.String)
	if sourceField.IsEnumArrayField() {
		autoFieldType = graphql.NewList(graphql.NewNonNull(graphql.String))
	}

	return autoFieldName, &graphql.Field{
		Name:        autoFieldName,
		Type:        autoFieldType,
		Description: fmt.Sprintf("Auto generated labels for enum field %s", sourceField.Name),
		Resolve:     r.createAutoEnumLabelResolver(sourceField, autoFieldName),
	}, true
}

func (r *graphqlModelResolver) createAutoEnumLabelResolver(
	sourceField *RuntimeField, autoFieldName string,
) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		logger := logfacade.GetLogger(p.Context)

		if sourceField.IsEnumArrayField() {
			return r.resolveAutoEnumArrayLabels(p, sourceField, autoFieldName, logger)
		}
		return r.resolveAutoEnumLabel(p, sourceField, autoFieldName, logger)
	}
}

func (r *graphqlModelResolver) resolveAutoEnumLabel(
	p graphql.ResolveParams,
	sourceField *RuntimeField,
	autoFieldName string,
	logger logfacade.Logger,
) (string, error) {
	record, ok := p.Source.(map[string]any)
	if !ok {
		logger.Warnf(p.Context, "invalid source type for auto enum label field %s", autoFieldName)
		return "", nil
	}

	sourceValue, exists := record[sourceField.Name]
	if !exists || sourceValue == nil {
		return "", nil
	}

	if sourceField.Enum == nil {
		logger.Warnf(p.Context, "source field %s has no enum definition", sourceField.Name)
		return "", nil
	}

	code, ok := normalizeEnumCode(sourceValue)
	if !ok {
		logger.Warnf(p.Context, "enum field %s has invalid type for auto label: %T", sourceField.Name, sourceValue)
		return "", nil
	}

	opt, err := sourceField.Enum.GetOptionByCode(code)
	if err != nil {
		logger.Warnf(p.Context, "enum option not found for auto label: code=%s, enum=%s", code, sourceField.Enum.Name)
		return "", nil
	}
	return opt.Label, nil
}

func (r *graphqlModelResolver) resolveAutoEnumArrayLabels(
	p graphql.ResolveParams,
	sourceField *RuntimeField,
	autoFieldName string,
	logger logfacade.Logger,
) ([]string, error) {
	record, ok := p.Source.(map[string]any)
	if !ok {
		logger.Warnf(p.Context, "invalid source type for auto enum labels field %s", autoFieldName)
		return []string{}, nil
	}

	sourceValue, exists := record[sourceField.Name]
	if !exists || sourceValue == nil {
		return []string{}, nil
	}

	if sourceField.Enum == nil {
		logger.Warnf(p.Context, "source field %s has no enum definition", sourceField.Name)
		return []string{}, nil
	}

	codes, ok := normalizeEnumArrayCodes(sourceValue)
	if !ok {
		logger.Warnf(
			p.Context,
			"enum array field %s has invalid type for auto labels: %T",
			sourceField.Name,
			sourceValue,
		)
		return []string{}, nil
	}

	labels := make([]string, 0, len(codes))
	for _, code := range codes {
		opt, err := sourceField.Enum.GetOptionByCode(code)
		if err != nil {
			logger.Warnf(
				p.Context,
				"enum option not found for auto labels: code=%s, enum=%s",
				code,
				sourceField.Enum.Name,
			)
			return []string{}, nil
		}
		labels = append(labels, opt.Label)
	}
	return labels, nil
}

// createRelationField 创建关系字段（使用 LogicalForeignKey）
func (r *graphqlModelResolver) createRelationField(
	ctx context.Context,
	maxDepth int,
	field *RuntimeField,
	relateObjMaps map[string]*graphql.Object,
	graphqlField *graphql.Field,
) (*graphql.Field, error) {
	if field.RelateFKID == nil {
		return nil, bizerrors.Errorf("RELATION field %s has no relate_fk_id", field.Name)
	}

	// 通过 relate_fk_id 查询 LogicalForeignKey（normal 方向：source=FK列, target=被引用列）
	lf, err := r.lfkRepo.GetByID(ctx, *field.RelateFKID)
	if err != nil {
		return nil, bizerrors.Errorf("failed to get logical foreign key for field %s: %w", field.Name, err)
	}

	// 获取被引用模型
	refModel, err := r.modelRepo.GetByID(ctx, lf.RefModelID)
	if err != nil {
		return nil, bizerrors.Errorf("failed to get reference model %s: %w", lf.RefModelID, err)
	}

	// 获取或创建被引用模型的 GraphQL 对象类型（以模型 Name 为 key 缓存）
	referenceObj, exists := relateObjMaps[refModel.Name]
	if !exists {
		referenceObj, err = r.generateModelType(ctx, maxDepth-1, refModel, relateObjMaps)
		if err != nil {
			return nil, bizerrors.Errorf("failed to generate model type for %s: %w", refModel.Name, err)
		}
		relateObjMaps[refModel.Name] = referenceObj
	}

	// RELATION 字段：根据 FK 方向选择多对一或一对多关系处理
	if lf.IsReverse() {
		return r.createOneToManyFieldFromFK(lf, referenceObj, graphqlField), nil
	}

	return r.createManyToOneFieldFromFK(lf, referenceObj, graphqlField), nil
}

// createManyToOneFieldFromFK 创建多对一关系字段（基于 LogicalForeignKey）
func (r *graphqlModelResolver) createManyToOneFieldFromFK(lf *modeldesign.LogicalForeignKey,
	referenceObj *graphql.Object, graphqlField *graphql.Field,
) *graphql.Field {
	graphqlField.Type = referenceObj
	graphqlField.Resolve = r.createManyToOneResolverFromFK(lf, lf.RefModelName)
	return graphqlField
}

// createOneToManyFieldFromFK 创建一对多关系字段（基于 LogicalForeignKey 的 reverse 方向）
// reverse 方向表示当前模型是被引用方，一个当前记录对应多个关联记录
func (r *graphqlModelResolver) createOneToManyFieldFromFK(lf *modeldesign.LogicalForeignKey,
	referenceObj *graphql.Object, graphqlField *graphql.Field,
) *graphql.Field {
	graphqlField.Type = graphql.NewList(graphql.NewNonNull(referenceObj))
	graphqlField.Resolve = r.createOneToManyResolverFromFK(lf, referenceObj.Name())
	return graphqlField
}

// createOneToManyResolverFromFK 创建一对多关系解析器（基于 LogicalForeignKey 的 reverse 方向）
// 对于 reverse 方向的 LFK：
//   - lf.SourceFields = 当前模型（User）用于查询的列，比如 ["id"]
//   - lf.TargetFields = 关联模型（Order）中要匹配的列，比如 ["user_id"]
//
// 查询逻辑：将当前记录的 SourceFields 值作为 WHERE 条件中 TargetFields 的值
// 支持复合外键：多个 SourceFields/TargetFields 全部加入 WHERE 条件（AND 连接）
func (r *graphqlModelResolver) createOneToManyResolverFromFK(
	lf *modeldesign.LogicalForeignKey,
	refModelName string,
) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		rctx, _ := getGraphqlRequestContext(p.Context)
		logger := logfacade.GetLogger(p.Context)

		// 获取父对象记录
		record, ok := p.Source.(map[string]any)
		if !ok {
			logger.Warn(p.Context, "invalid source type for one-to-many relation")
			return []map[string]any{}, nil
		}

		logger.Infof(p.Context, "resolving one-to-many relation: record=%+v", record)

		// 从当前记录中提取 SourceFields 的所有值
		sourceValues := make([]any, 0, len(lf.SourceFields))
		for _, sourceField := range lf.SourceFields {
			value, exists := record[sourceField]
			if !exists || value == nil {
				// 如果任意一个 SourceField 值为 nil，返回空数组
				logger.Infof(p.Context,
					"one-to-many relation: source field %s is nil or missing, returning empty array",
					sourceField)
				return []map[string]any{}, nil
			}
			sourceValues = append(sourceValues, value)
		}

		// 防御性检查：SourceFields 与 TargetFields 数量必须一致
		if len(lf.TargetFields) != len(sourceValues) {
			logger.Warnf(p.Context, "one-to-many relation: FK field count mismatch: source=%d, target=%d",
				len(sourceValues), len(lf.TargetFields))
			return []map[string]any{}, nil
		}

		// 构建 WHERE 条件 - zip(TargetFields, sourceValues)，多个条件 AND 连接
		whereMap := make(map[string]any)
		for i, targetField := range lf.TargetFields {
			if i < len(sourceValues) {
				whereMap[targetField] = sourceValues[i]
			}
		}

		logger.Infof(p.Context, "querying one-to-many relation: TableName=%s, WHERE=%+v",
			lf.RefModelName, whereMap)

		// 调用 FindMany 查询关联记录
		results, err := rctx.ClientRepo.FindMany(p.Context, &FindManyInput{
			TableName: lf.RefModelName,
			Where:     whereMap,
		})
		if err != nil {
			logger.Errorf(p.Context, "failed to query one-to-many relation: %v", err)
			return nil, err
		}

		if results == nil {
			return []map[string]any{}, nil
		}

		return results, nil
	}
}

// createManyToOneResolverFromFK 创建多对一关系解析器（基于 LogicalForeignKey）。
//
// 使用 dataloader 解决 N+1 问题：
//   - resolver 调用 loader.Load(ctx, fkValue) 返回一个 Thunk（函数闭包），不立即发 SQL
//   - graphql-go 广度优先执行：同一层所有字段的 Load() 先全部被调用（收集 key）
//   - 随后 graphql-go dethunk 时才逐一调用 Thunk，此时 dataloader 触发批量 IN 查询
//   - 最终发出一条 WHERE referenceKey IN (k1, k2, ...) SQL，结果按 key 分发
func (r *graphqlModelResolver) createManyToOneResolverFromFK(
	lf *modeldesign.LogicalForeignKey,
	refModelName string,
) graphql.FieldResolveFn {
	foreignKey := lf.SourceFields[0]
	referenceKey := lf.TargetFields[0]

	return func(p graphql.ResolveParams) (interface{}, error) {
		rctx, _ := getGraphqlRequestContext(p.Context)
		logger := logfacade.GetLogger(p.Context)

		record, ok := p.Source.(map[string]any)
		if !ok {
			logger.Warn(p.Context, "invalid source type for many-to-one relation")
			return nil, nil
		}

		foreignKeyValue, exists := record[foreignKey]
		if !exists || foreignKeyValue == nil {
			return nil, nil
		}

		fkStr, ok := toString(foreignKeyValue)
		if !ok {
			// 无法转为字符串 key，fallback 到单条查询
			result, err := rctx.ClientRepo.FindFirst(p.Context, &FindFirstInput{
				TableName: refModelName,
				Where:     map[string]any{referenceKey: foreignKeyValue},
			})
			if err != nil {
				if shared.IsNotFoundError(err) {
					return nil, nil
				}
				return nil, err
			}
			return result, nil
		}

		// 从请求级 context 取 loader（懒初始化，同一请求内复用）
		loader := rctx.getOrCreateLoader(refModelName, referenceKey)

		// Load 不立即发 SQL，只是将 key 加入 batch 队列并返回 Thunk。
		// graphql-go 将 resolver 返回的 func() interface{} 识别为 thunk，
		// 先收集同层所有字段的 thunk，再统一 dethunk——
		// 正是在 dethunk 阶段，dataloader 才触发批量 IN 查询。
		thunk := loader.Load(p.Context, fkStr)

		return func() (interface{}, error) {
			result, err := thunk()
			if err != nil {
				// 悬空外键或批量查询失败，按 LEFT JOIN 语义返回 nil
				logger.Warnf(p.Context,
					"many-to-one relation load failed (dangling FK?): table=%s key=%s val=%s err=%v",
					refModelName, referenceKey, fkStr, err)
				return nil, nil
			}
			// 显式判断 nil：dataloader 未找到记录时返回 map[string]any(nil)，
			// 若直接赋给 interface{} 会产生"有类型的 nil"，graphql-go 会将其当作
			// 非 nil 对象解析，导致返回 {"id": null} 而非 null。
			if result == nil {
				return nil, nil
			}
			return result, nil
		}, nil
	}
}

func (r *graphqlModelResolver) generateModelType(ctx context.Context, maxDepth int, model *RuntimeModel,
	relateObj map[string]*graphql.Object,
) (*graphql.Object, error) {
	graphqlfields := graphql.Fields{}
	logger := logfacade.GetLogger(ctx)

	// 验证模型名称不为空
	if model == nil {
		return nil, bizerrors.Errorf("model is nil")
	}
	if model.Name == "" {
		return nil, bizerrors.Errorf("model name is empty")
	}

	if len(model.Fields) == 0 {
		return nil, bizerrors.Errorf("model %s has no fields", model.Name)
	}
	logger.Infof(ctx, "model: %+v", model)
	for _, field := range model.Fields {
		// 验证字段名称不为空
		if field.Name == "" {
			return nil, bizerrors.Errorf("model %s has field with empty name", model.Name)
		}
		logger.Infof(ctx, "modelName=%s, field=%+v", model.Name, field)
		if field.IsRelationField() && maxDepth <= 0 {
			continue
		}
		graphqlfield, err := r.createField(ctx, maxDepth, field, relateObj)
		if err != nil {
			logfacade.GetLogger(ctx).Errorf(ctx, "create graphql field err %s", field.Name)
			return nil, err
		}
		if graphqlfield == nil {
			continue
		}
		graphqlfields[field.Name] = graphqlfield
	}
	r.injectAutoEnumLabelFields(ctx, model, graphqlfields)

	// 注入 _displayName 字段（始终返回 String!，根据当前模型的 displayField 解析）
	graphqlfields[FieldDisplayName] = r.createDisplayNameField(model.DisplayField)

	modelType := graphql.NewObject(graphql.ObjectConfig{
		Name:        gqlTypeName(model.Name) + "Query",
		Fields:      graphqlfields,
		Description: model.Description,
	})
	return modelType, nil
}

func (r *graphqlModelResolver) generateModelTypeSkipRelation(
	ctx context.Context, model *RuntimeModel,
) (*graphql.Object, error) {
	graphqlfields := graphql.Fields{}

	// 验证模型名称不为空
	if model == nil {
		return nil, bizerrors.Errorf("model is nil")
	}
	if model.Name == "" {
		return nil, bizerrors.Errorf("model name is empty")
	}

	for _, field := range model.Fields {
		// 验证字段名称不为空
		if field.Name == "" {
			return nil, bizerrors.Errorf("model %s has field with empty name", model.Name)
		}
		if field.IsRelationField() {
			continue
		}
		graphqlfield, err := r.createField(ctx, 0, field, map[string]*graphql.Object{})
		if err != nil {
			return nil, err
		}
		graphqlfields[field.Name] = graphqlfield
	}
	r.injectAutoEnumLabelFields(ctx, model, graphqlfields)
	modelType := graphql.NewObject(graphql.ObjectConfig{
		Name:        gqlTypeName(model.Name) + "Mutation",
		Fields:      graphqlfields,
		Description: model.Description,
	})
	return modelType, nil
}

func (r *graphqlModelResolver) createModelType(ctx context.Context) (*graphql.Object, error) {
	relationGraphqlObjMap := map[string]*graphql.Object{}
	maxDepth := 1
	obj, err := r.generateModelType(ctx, maxDepth, r.model, relationGraphqlObjMap)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// createFindUniqueResultType creates a wrapper result type for findUnique operation
func (r *graphqlModelResolver) createFindUniqueResultType(modelType graphql.Type) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(r.model.Name) + "FindUniqueResult",
		Fields: graphql.Fields{
			FieldItem: &graphql.Field{
				Type:        modelType,
				Description: "Single matching record (nullable if not found)",
			},
			FieldTimeCost: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Query execution time in milliseconds",
			},
			FieldReqId: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Unique request tracking ID (UUID v7)",
			},
		},
		Description: "Result wrapper for findUnique query with metadata",
	})
}

// createFindFirstResultType creates a wrapper result type for findFirst operation
func (r *graphqlModelResolver) createFindFirstResultType(modelType graphql.Type) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(r.model.Name) + "FindFirstResult",
		Fields: graphql.Fields{
			FieldItem: &graphql.Field{
				Type:        modelType,
				Description: "First matching record (nullable if not found)",
			},
			FieldTimeCost: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Query execution time in milliseconds",
			},
			FieldReqId: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Unique request tracking ID (UUID v7)",
			},
		},
		Description: "Result wrapper for findFirst query with metadata",
	})
}

// createFindManyResultType creates a wrapper result type for findMany operation
func (r *graphqlModelResolver) createFindManyResultType(modelType graphql.Type) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(r.model.Name) + "FindManyResult",
		Fields: graphql.Fields{
			FieldItems: &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(modelType)),
				Description: "Array of matching records",
			},
			FieldTotalCount: &graphql.Field{
				Type:        graphql.Int,
				Description: "Total number of matching records (optional, not implemented yet)",
			},
			FieldTimeCost: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Query execution time in milliseconds",
			},
			FieldReqId: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Unique request tracking ID (UUID v7)",
			},
		},
		Description: "Result wrapper for findMany query with metadata",
	})
}

// createListByCursorResultType creates the GraphQL result type for listByCursor.
func (r *graphqlModelResolver) createListByCursorResultType(modelType graphql.Type) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(r.model.Name) + "ListByCursorResult",
		Fields: graphql.Fields{
			FieldItems: &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(modelType)),
				Description: "Current cursor page records",
			},
			FieldNextCursor: &graphql.Field{
				Type:        graphql.String,
				Description: "Opaque cursor for the next page (null = last page)",
			},
			FieldHasNextPage: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Whether more results exist",
			},
			FieldTimeCost: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Query execution time in milliseconds",
			},
			FieldReqId: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Request tracking ID",
			},
		},
		Description: r.listByCursorFieldDescription(),
	})
}

func (r *graphqlModelResolver) createListByPageResultType(modelType graphql.Type) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(r.model.Name) + "ListByPageResult",
		Fields: graphql.Fields{
			FieldItems: &graphql.Field{
				Type:        graphql.NewList(graphql.NewNonNull(modelType)),
				Description: "Records in the current page",
			},
			FieldTotal: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Total number of matching records",
			},
			FieldPageIndex: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Current 1-based page number",
			},
			FieldPageSize: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Requested page size",
			},
			FieldTimeCost: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "Query execution time in milliseconds",
			},
			FieldReqId: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "Request tracking ID",
			},
		},
		Description: "Result wrapper for listByPage offset pagination. Use explicit orderBy for stable page traversal.",
	})
}

func (r *graphqlModelResolver) listByCursorFieldDescription() string {
	if r.model != nil && r.model.InsertionOrderField != nil && *r.model.InsertionOrderField != "" {
		return fmt.Sprintf(
			"Cursor pagination result. The caller must use sortField=%q,"+
				" which is the model insertionOrderField."+
				" Using any other field can cause duplicate or missing rows across pages.",
			*r.model.InsertionOrderField,
		)
	}
	return "Cursor pagination result. sortField should use a monotonic insertion-order field." +
		" Using another field can cause duplicate or missing rows across pages."
}

func (m *graphqlModelResolver) executeListByCursor(p graphql.ResolveParams) (map[string]any, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionSelect); err != nil {
		return nil, err
	}
	startTime := time.Now()

	// Parse arguments
	sortField, _ := p.Args[FieldSortField].(string)
	if sortField == "" {
		return nil, bizerrors.Errorf("sortField is required for listByCursor")
	}
	sortDirection, _ := p.Args[FieldSortDirection].(string)
	if sortDirection != OrderByAsc && sortDirection != OrderByDesc {
		sortDirection = OrderByAsc
	}
	limitRaw, _ := p.Args[FieldLimit].(int)
	if limitRaw <= 0 {
		limitRaw = 20
	}
	limit := uint(limitRaw)
	where, err := getWhere(p.Args)
	if err != nil {
		return nil, err
	}

	// Parse after cursor
	var after *CursorData
	if afterStr, ok := p.Args[FieldAfter].(string); ok && afterStr != "" {
		decoded, err := decodeCursor(afterStr)
		if err != nil {
			return nil, bizerrors.Errorf("invalid after cursor: %w", err)
		}
		after = &decoded
	}

	// Insertion-order field from model config
	insertionOrderField := ""
	if m.model.InsertionOrderField != nil {
		insertionOrderField = *m.model.InsertionOrderField
	}

	input := &ListByCursorInput{
		TableName:           m.model.Name,
		SortField:           sortField,
		SortDirection:       sortDirection,
		InsertionOrderField: insertionOrderField,
		After:               after,
		Limit:               limit,
		Where:               where,
	}

	// Inject RLS row filter

	// Fetch limit+1 rows to detect hasNextPage
	rows, err := rctx.ClientRepo.ListByCursor(p.Context, input)
	if err != nil {
		return nil, err
	}

	hasNextPage := len(rows) > int(limit)
	if hasNextPage {
		rows = rows[:limit]
	}

	// Build nextCursor from last record
	var nextCursorStr *string
	if hasNextPage && len(rows) > 0 {
		last := rows[len(rows)-1]
		sv := fmt.Sprintf("%v", last[sortField])
		cd := CursorData{SortField: sortField, SortValue: sv}
		if insertionOrderField != "" {
			cd.IOField = insertionOrderField
			cd.IOValue = fmt.Sprintf("%v", last[insertionOrderField])
		}
		encoded := encodeCursor(cd)
		nextCursorStr = &encoded
	}

	metadata := requestcontext.GetMetadata(p.Context)
	reqId := ""
	if metadata != nil {
		reqId = metadata.ReqID
	}

	return map[string]any{
		FieldItems:       rows,
		FieldNextCursor:  nextCursorStr,
		FieldHasNextPage: hasNextPage,
		FieldTimeCost:    int(time.Since(startTime).Milliseconds()),
		FieldReqId:       reqId,
	}, nil
}

func (m *graphqlModelResolver) executeListByPage(p graphql.ResolveParams) (map[string]any, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionSelect); err != nil {
		return nil, err
	}
	startTime := time.Now()

	pageIndexRaw, _ := p.Args[FieldPageIndex].(int)
	if pageIndexRaw <= 0 {
		pageIndexRaw = 1
	}
	pageSizeRaw, _ := p.Args[FieldPageSize].(int)
	if pageSizeRaw <= 0 {
		pageSizeRaw = 20
	}

	where, err := getWhere(p.Args)
	if err != nil {
		return nil, err
	}
	orderBy, err := getOrderBy(p.Args)
	if err != nil {
		return nil, err
	}
	if len(orderBy) == 0 {
		return nil, bizerrors.Errorf("orderBy is required for listByPage")
	}
	if len(orderBy) > 3 {
		return nil, bizerrors.Errorf("listByPage supports at most 3 orderBy fields")
	}

	findManyInput := &FindManyInput{
		TableName: m.model.Name,
		Where:     where,
		OrderBy:   orderBy,
		Limit:     uint(pageSizeRaw),
		Offset:    uint((pageIndexRaw - 1) * pageSizeRaw),
	}

	items, err := rctx.ClientRepo.FindMany(p.Context, findManyInput)
	if err != nil {
		if bizerrors.Is(err, sql.ErrNoRows) {
			items = []map[string]any{}
		} else {
			return nil, err
		}
	}

	countResult, err := rctx.ClientRepo.Count(p.Context, &CountInput{
		TableName:  findManyInput.TableName,
		Where:      findManyInput.Where,
		RawFilters: findManyInput.RawFilters,
	})
	if err != nil {
		return nil, err
	}
	total, err := cast.ToIntE(countResult[FieldCount])
	if err != nil {
		return nil, bizerrors.Errorf("invalid count result for listByPage: %w", err)
	}

	metadata := requestcontext.GetMetadata(p.Context)
	reqID := ""
	if metadata != nil {
		reqID = metadata.ReqID
	}

	return map[string]any{
		FieldItems:     items,
		FieldTotal:     total,
		FieldPageIndex: pageIndexRaw,
		FieldPageSize:  pageSizeRaw,
		FieldTimeCost:  int(time.Since(startTime).Milliseconds()),
		FieldReqId:     reqID,
	}, nil
}

func (m *graphqlModelResolver) createListByCursorField(modelType graphql.Type) (*graphql.Field, error) {
	args := m.inputTypeGenerator.GenerateListByCursorArgs(m.model)
	resultType := m.createListByCursorResultType(modelType)

	return &graphql.Field{
		Type:        resultType,
		Args:        args,
		Description: m.listByCursorFieldDescription(),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := m.executeListByCursor(p)
			if err != nil {
				logfacade.GetLogger(p.Context).Error(p.Context, "executeListByCursor fail", logfacade.Err(err))
				return nil, err
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) createListByPageField(modelType graphql.Type) (*graphql.Field, error) {
	args := m.inputTypeGenerator.GenerateListByPageArgs(m.model)
	resultType := m.createListByPageResultType(modelType)
	desc := "Convenience page-based pagination." +
		" Internally this uses findMany + count, and is intended for simple table paging."

	return &graphql.Field{
		Type:        resultType,
		Args:        args,
		Description: desc,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			result, err := m.executeListByPage(p)
			if err != nil {
				logfacade.GetLogger(p.Context).Error(p.Context, "executeListByPage fail", logfacade.Err(err))
				return nil, err
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) createRootMutation(modelType graphql.Type) (*graphql.Object, error) {
	createOneField, err := m.createCreateOneField(modelType)
	if err != nil {
		return nil, err
	}
	updateOneField, err := m.createUpdateOneField(modelType)
	if err != nil {
		return nil, err
	}
	deleteOneField, err := m.createDeleteOneField(modelType)
	if err != nil {
		return nil, err
	}
	createManyField, err := m.createCreateManyField(modelType)
	if err != nil {
		return nil, err
	}
	updateManyField, err := m.createUpdateManyField(modelType)
	if err != nil {
		return nil, err
	}
	deleteManyField, err := m.createDeleteManyField(modelType)
	if err != nil {
		return nil, err
	}

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			OperationCreate:     createOneField,
			OperationUpdate:     updateOneField,
			OperationDelete:     deleteOneField,
			OperationCreateMany: createManyField,
			OperationUpdateMany: updateManyField,
			OperationDeleteMany: deleteManyField,
		},
	})
	return rootMutation, nil
}

func (m *graphqlModelResolver) executeCreateOne(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionInsert); err != nil {
		return nil, err
	}
	input, err := newCreateOneInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}

	for _, field := range m.model.Fields {
		if field.Type.Format == modeldesign.FormatUUID {
			// Only generate UUIDV7 if user didn't provide a value
			if _, exists := input.Data[field.Name]; !exists || input.Data[field.Name] == nil {
				uuidv7, err := bizutils.GenerateUUIDV7()
				if err != nil {
					return nil, err
				}
				input.Data[field.Name] = uuidv7
			}
		}
	}


	input.Id = cast.ToString(input.Data[FieldID])

	id, err := rctx.ClientRepo.CreateOne(p.Context, input)
	if err != nil {
		return nil, err
	}

	// 检查客户端是否选择了 createdObj 字段
	hasCreatedObjField := input.CreatedObj
	// 返回包装结果
	result := map[string]interface{}{
		FieldID: id,
	}

	if hasCreatedObjField {
		// 获取完整的创建对象
		createdObj, err := rctx.ClientRepo.FindUnique(p.Context, &FindUniqueInput{
			TableName: m.model.Name,
			Where: map[string]any{
				FieldID: id,
			},
		})
		if err == nil && createdObj != nil {
			result[FieldCreatedObj] = createdObj
		}
	}

	return result, nil
}

func (m *graphqlModelResolver) createCreateOneField(modelType graphql.Type) (*graphql.Field, error) {
	args, err := m.inputTypeGenerator.GenerateCreateOneArgs(m.model)
	if err != nil {
		return nil, err
	}

	// 创建 CreateOneResult 包装类型
	createResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(m.model.Name) + "Create" + ResultTypeSuffix,
		Fields: graphql.Fields{
			FieldID: &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			FieldCreatedObj: &graphql.Field{
				Type: modelType,
			},
		},
	})

	return &graphql.Field{
		Type: createResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)
			result, err := m.executeCreateOne(p)
			if err != nil {
				logger.Error(p.Context, "createOne_fail", logfacade.Err(err))
			} else {
				logger.Infof(p.Context, "createOne_result=%+v", result)
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) executeUpdateOne(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionUpdate); err != nil {
		return nil, err
	}
	input, err := newUpdateOneInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}


	// 返回包装结果
	result := map[string]interface{}{
		FieldSuccess: true,
	}
	updatedObj, err := rctx.ClientRepo.UpdateOne(p.Context, input)
	if err != nil {
		return nil, err
	}

	// 检查客户端是否选择了 updatedObj 字段
	hasUpdatedObjField := input.UpdatedObj

	if hasUpdatedObjField {
		result[FieldUpdatedObj] = updatedObj
	}

	return result, nil
}

func (m *graphqlModelResolver) createUpdateOneField(modelType graphql.Type) (*graphql.Field, error) {
	args, err := m.inputTypeGenerator.GenerateUpdateOneArgs(m.model)
	if err != nil {
		return nil, err
	}

	// 创建 UpdateOneResult 包装类型
	updateResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Update" + m.model.Name + ResultTypeSuffix,
		Fields: graphql.Fields{
			FieldSuccess: &graphql.Field{
				Type: graphql.Boolean,
			},
			FieldUpdatedObj: &graphql.Field{
				Type: modelType,
			},
		},
	})

	return &graphql.Field{
		Type: updateResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)
			result, err := m.executeUpdateOne(p)
			if err != nil {
				err = handleErr(p.Context, err)
			} else {
				logger.Infof(p.Context, "updateOne_result=%+v", result)
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) executeDeleteOne(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionDelete); err != nil {
		return nil, err
	}
	input, err := newDeleteOneInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}
	// 返回包装结果
	result := map[string]interface{}{
		FieldSuccess: true,
	}
	deleteResult, err := rctx.ClientRepo.DeleteOne(p.Context, input)
	if err != nil {
		return nil, err
	}

	if deleteResult == nil && input.DeletedObj {
		result[FieldSuccess] = false
		return result, nil
	}

	if input.DeletedObj {
		result[FieldDeletedObj] = deleteResult
	}
	return result, nil
}

func (m *graphqlModelResolver) createDeleteOneField(modelType graphql.Type) (*graphql.Field, error) {
	args, err := m.inputTypeGenerator.GenerateDeleteOneArgs(m.model)
	if err != nil {
		return nil, err
	}

	deleteResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Delete" + m.model.Name + ResultTypeSuffix,
		Fields: graphql.Fields{
			FieldSuccess: &graphql.Field{
				Type: graphql.Boolean,
			},
			FieldDeletedObj: &graphql.Field{
				Type: modelType,
			},
		},
	})

	return &graphql.Field{
		Type: deleteResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)
			result, err := m.executeDeleteOne(p)
			if err != nil {
				logger.Error(p.Context, "deleteOne_fail", logfacade.Err(err))
			} else {
				logger.Infof(p.Context, "deleteOne_result=%+v", result)
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) executeCreateMany(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionInsert); err != nil {
		return nil, err
	}
	logger := logfacade.GetLogger(p.Context)
	logger.Infof(p.Context, "type=%T", p.Args[FieldData])
	input, err := newCreateManyInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}
	for _, dataItem := range input.Data {
		for _, field := range m.model.Fields {
			if field.Type.Format == modeldesign.FormatUUID {
				// Only generate UUIDV7 if user didn't provide a value
				if _, exists := dataItem[field.Name]; !exists || dataItem[field.Name] == nil {
					uuidv7, err := bizutils.GenerateUUIDV7()
					if err != nil {
						return nil, err
					}
					dataItem[field.Name] = uuidv7
				}
			}
		}
	}
	result, err := rctx.ClientRepo.CreateMany(p.Context, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *graphqlModelResolver) createCreateManyField(modelType graphql.Type) (*graphql.Field, error) {
	args, err := m.inputTypeGenerator.GenerateCreateManyArgs(m.model)
	if err != nil {
		return nil, err
	}

	// 创建 CreateManyResult 包装类型
	createManyResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(m.model.Name) + "CreateMany" + ResultTypeSuffix,
		Fields: graphql.Fields{
			FieldCount: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "成功创建的记录数",
			},
			FieldIdList: &graphql.Field{
				Type:        graphql.NewList(graphql.ID),
				Description: "创建的记录ID列表（仅在查询中选择时返回）",
			},
		},
	})

	return &graphql.Field{
		Type: createManyResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)
			result, err := m.executeCreateMany(p)
			if err != nil {
				logger.Error(p.Context, "createMany_fail", logfacade.Err(err))
			} else {
				logger.Infof(p.Context, "createMany_success")
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) executeUpdateMany(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionUpdate); err != nil {
		return nil, err
	}
	input, err := newUpdateManyInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}


	result, err := rctx.ClientRepo.UpdateMany(p.Context, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *graphqlModelResolver) createUpdateManyField(modelType graphql.Type) (*graphql.Field, error) {
	args, err := m.inputTypeGenerator.GenerateUpdateManyArgs(m.model)
	if err != nil {
		return nil, err
	}

	// 创建 UpdateManyResult 包装类型
	updateManyResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(m.model.Name) + "UpdateMany" + ResultTypeSuffix,
		Fields: graphql.Fields{
			FieldCount: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "成功更新的记录数",
			},
		},
	})

	return &graphql.Field{
		Type: updateManyResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)
			result, err := m.executeUpdateMany(p)
			if err != nil {
				logger.Error(p.Context, "updateMany_fail", logfacade.Err(err))
			} else {
				logger.Infof(p.Context, "updateMany_success")
			}
			return result, err
		},
	}, nil
}

func (m *graphqlModelResolver) executeDeleteMany(p graphql.ResolveParams) (interface{}, error) {
	rctx, _ := getGraphqlRequestContext(p.Context)
	if err := rctx.RLS.Permissions.CheckAction(ActionDelete); err != nil {
		return nil, err
	}
	input, err := newDeleteManyInput(m.model.Name, p)
	if err != nil {
		return nil, err
	}
	result, err := rctx.ClientRepo.DeleteMany(p.Context, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *graphqlModelResolver) createDeleteManyField(modelType graphql.Type) (*graphql.Field, error) {
	args, err := m.inputTypeGenerator.GenerateDeleteManyArgs(m.model)
	if err != nil {
		return nil, err
	}

	// 创建 DeleteManyResult 包装类型
	deleteManyResultType := graphql.NewObject(graphql.ObjectConfig{
		Name: gqlTypeName(m.model.Name) + "DeleteMany" + ResultTypeSuffix,
		Fields: graphql.Fields{
			FieldCount: &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "成功删除的记录数",
			},
		},
	})

	return &graphql.Field{
		Type: deleteManyResultType,
		Args: args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)
			result, err := m.executeDeleteMany(p)
			if err != nil {
				logger.Error(p.Context, "deleteMany_fail", logfacade.Err(err))
			} else {
				logger.Infof(p.Context, "deleteMany_success")
			}
			return result, err
		},
	}, nil
}

func handleErr(ctx context.Context, err error) error {
	logger := logfacade.GetLogger(ctx)
	logger.Error(ctx, "before_process_graphql_err", logfacade.Err(err))
	err = handleRepoErr(ctx, err)
	return gqlerrors.FormatError(err)
}

// handleRepoErr 分析repo错误并创建BusinessError
func handleRepoErr(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	var bizErr *bizerrors.BusinessError
	if bizerrors.As(err, &bizErr) {
		return bizErr
	}

	var repoErr *shared.RepositoryError
	if bizerrors.As(err, &repoErr) {
		switch repoErr.Type {
		case shared.ErrTypeConnection:
			return bizerrors.NewErrorFromContext(ctx, bizerrors.ConnectFailed, repoErr.Cause)
		case shared.ErrTypeTimeout:
			return bizerrors.NewErrorFromContext(ctx, bizerrors.TimeOut)
		case shared.ErrTypeConstraint:
			return bizerrors.NewErrorFromContext(ctx, bizerrors.DuplicateKey)
		case shared.ErrTypePermission:
			return bizerrors.NewErrorFromContext(ctx, bizerrors.DatabaseAccessDenied)

		}
		return bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError)
	}
	return err
}

// createDisplayNameField 创建 _displayName 字段
// _displayName 字段用于返回 displayField 指定字段的值（转为字符串）
// 如果 displayField 未配置或对应值不可用（null/空/对象/数组），返回空字符串 ""
func (r *graphqlModelResolver) createDisplayNameField(displayField *string) *graphql.Field {
	return &graphql.Field{
		Name:        FieldDisplayName,
		Type:        graphql.NewNonNull(graphql.String),
		Description: "Display name resolved from displayField configuration",
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			logger := logfacade.GetLogger(p.Context)

			// 获取父对象记录
			record, ok := p.Source.(map[string]any)
			if !ok {
				logger.Warnf(p.Context, "_displayName: invalid source type: %T", p.Source)
				return "", nil
			}

			// 检查是否配置了 displayField（按当前模型）
			if displayField == nil || *displayField == "" {
				return "", nil
			}

			displayFieldName := *displayField
			value, exists := record[displayFieldName]
			if exists && value != nil {
				return valueToString(value), nil
			}

			// 兜底查询：当上游只返回了部分列且缺少 displayField 时，按当前记录 id 回查 displayField。
			idValue, idExists := record[FieldID]
			if !idExists || idValue == nil {
				return "", nil
			}
			rctx, ok := getGraphqlRequestContext(p.Context)
			if !ok || rctx.ClientRepo == nil {
				return "", nil
			}

			row, err := rctx.ClientRepo.FindUnique(p.Context, &FindUniqueInput{
				TableName: r.model.Name,
				Selection: &Selection{
					FieldNames: map[string]bool{
						displayFieldName: true,
					},
				},
				Where: map[string]any{
					FieldID: idValue,
				},
			})
			if err != nil || row == nil {
				return "", nil
			}

			return valueToString(row[displayFieldName]), nil
		},
	}
}

// valueToString 将任意值转换为字符串，用于 _displayName 解析
// 对于非标量类型（对象、数组），返回空字符串
func valueToString(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return cast.ToString(val)
	case uint, uint8, uint16, uint32, uint64:
		return cast.ToString(val)
	case float32, float64:
		return cast.ToString(val)
	case []interface{}, map[string]interface{}:
		// 数组或对象，返回空字符串
		return ""
	default:
		// 尝试使用 cast.ToString
		result, err := cast.ToStringE(val)
		if err != nil {
			return ""
		}
		return result
	}
}
