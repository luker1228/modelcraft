package modelruntime

import (
	"modelcraft/pkg/bizerrors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/spf13/cast"
)

// Selection 选择的字段
type Selection struct {
	FieldNames map[string]bool
}

type RawSQLFilter struct {
	SQL    string
	Params []any
}

// FindUniqueInput 查找唯一记录的输入参数
type FindUniqueInput struct {
	TableName  string
	Selection  *Selection
	Where      map[string]any
	RawFilters []RawSQLFilter
}

func newFindUniqueInput(tableName string, param graphql.ResolveParams) (*FindUniqueInput, error) {
	whereMap, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}
	return &FindUniqueInput{
		TableName: tableName,
		Where:     whereMap,
	}, nil
}

func getWhere(param map[string]any) (map[string]any, error) {
	w, ok := param[FieldWhere]
	if !ok {
		return map[string]any{}, nil
	}
	whereMap, ok := w.(map[string]any)
	if !ok {
		return nil, bizerrors.Errorf("where must be map[string]any")
	}

	return whereMap, nil
}

// FindManyInput 查找多个记录的输入参数，支持过滤、分页
type FindManyInput struct {
	TableName  string
	Selection  *Selection
	Where      map[string]any
	RawFilters []RawSQLFilter
	OrderBy    []OrderBy
	Limit      uint
	// ExplicitLimit indicates the caller explicitly passed a take value (including 0).
	// When false and Limit==0, the query uses the default limit.
	// When true and Limit==0, LIMIT 0 is applied (returns empty result set).
	ExplicitLimit bool
	Offset        uint
}

type OrderBy struct {
	Field     string
	Direction string
}

// FindManyInInput 通过 IN 条件批量查找关联记录的输入参数，用于解决 N+1 问题。
// 等价于：SELECT * FROM TableName WHERE ReferenceKey IN (Values...)
type FindManyInInput struct {
	// TableName 目标表名
	TableName string
	// ReferenceKey 目标表中被 IN 匹配的字段名（如 "id"）
	ReferenceKey string
	// Values 需要匹配的值列表（去重后传入）
	Values []any
}

// ListByCursorInput holds parameters for cursor-based keyset pagination.
type ListByCursorInput struct {
	TableName           string
	Selection           *Selection
	Where               map[string]any // extra WHERE (RLS etc.)
	RawFilters          []RawSQLFilter
	SortField           string      // required: field to sort by
	SortDirection       string      // "asc" or "desc"
	InsertionOrderField string      // optional: monotonically increasing field name
	After               *CursorData // nil = first page
	Limit               uint
}

func newFindManyInput(tableName string, param graphql.ResolveParams) (*FindManyInput, error) {
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}
	takeRaw, takeProvided := param.Args[FieldTake]
	if !takeProvided {
		takeRaw = 10
	}
	take, err := cast.ToUintE(takeRaw)
	if err != nil {
		return nil, bizerrors.Errorf("take must be an integer, val = %v type = %T", takeRaw, takeRaw)
	}
	skipRaw, ok := param.Args[FieldSkip]
	if !ok {
		skipRaw = 0
	}
	skip, err := cast.ToUintE(skipRaw)
	if err != nil {
		return nil, bizerrors.Errorf("skip must be an integer, val = %v type = %T", skipRaw, skipRaw)
	}
	orderBy, err := getOrderBy(param.Args)
	if err != nil {
		return nil, err
	}
	return &FindManyInput{
		TableName:     tableName,
		Where:         where,
		OrderBy:       orderBy,
		Limit:         take,
		ExplicitLimit: takeProvided,
		Offset:        skip,
	}, nil
}

func getOrderBy(param map[string]any) ([]OrderBy, error) {
	raw, ok := param[FieldOrderBy]
	if !ok || raw == nil {
		return nil, nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, bizerrors.Errorf("orderBy must be []any")
	}

	result := make([]OrderBy, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			return nil, bizerrors.Errorf("orderBy item must be map[string]any")
		}
		for field, directionVal := range entry {
			direction, err := cast.ToStringE(directionVal)
			if err != nil {
				return nil, bizerrors.Errorf(
					"orderBy direction must be string, val = %v type = %T",
					directionVal, directionVal,
				)
			}
			if direction != OrderByAsc && direction != OrderByDesc {
				return nil, bizerrors.Errorf(
					"orderBy direction must be %q or %q, got %q",
					OrderByAsc, OrderByDesc, direction,
				)
			}
			result = append(result, OrderBy{
				Field:     field,
				Direction: direction,
			})
		}
	}

	return result, nil
}

// FindFirstInput 查找第一个匹配记录的输入参数
type FindFirstInput struct {
	TableName  string
	Selection  *Selection
	Where      map[string]any
	RawFilters []RawSQLFilter
}

func newFindFirstInput(tableName string, param graphql.ResolveParams) (*FindFirstInput, error) {
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}
	return &FindFirstInput{
		TableName: tableName,
		Where:     where,
	}, nil
}

// CreateOneInput 创建单个记录的输入参数
type CreateOneInput struct {
	CreatedObj bool
	TableName  string
	Id         string
	Data       map[string]any
	RawFilters []RawSQLFilter
}

func newCreateOneInput(tableName string, param graphql.ResolveParams) (*CreateOneInput, error) {
	data, ok := param.Args[FieldData].(map[string]any)
	if !ok {
		return nil, bizerrors.Errorf("data must be map[string]any")
	}
	// 检查是否选择了 createdObj 字段
	createdObj := hasSelectedField(param, FieldCreatedObj)
	return &CreateOneInput{
		CreatedObj: createdObj,
		TableName:  tableName,
		Data:       data,
	}, nil
}

// UpdateOneInput 更新单个记录的输入参数
type UpdateOneInput struct {
	TableName  string
	UpdatedObj bool
	Where      map[string]any
	Data       map[string]any
	RawFilters []RawSQLFilter
}

func newUpdateOneInput(tableName string, param graphql.ResolveParams) (*UpdateOneInput, error) {
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}
	data, ok := param.Args[FieldData].(map[string]any)
	if !ok {
		return nil, bizerrors.Errorf("data must be map[string]any")
	}

	// 检查是否选择了 updatedObj 字段
	updatedObj := hasSelectedField(param, FieldUpdatedObj)

	return &UpdateOneInput{
		TableName:  tableName,
		UpdatedObj: updatedObj,
		Where:      where,
		Data:       data,
	}, nil
}

// DeleteOneInput 删除单个记录的输入参数
type DeleteOneInput struct {
	DeletedObj bool
	TableName  string
	Where      map[string]any
	RawFilters []RawSQLFilter
}

func newDeleteOneInput(tableName string, param graphql.ResolveParams) (*DeleteOneInput, error) {
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}

	// 检查是否选择了 updatedObj 字段
	deletedObj := hasSelectedField(param, FieldDeletedObj)
	return &DeleteOneInput{
		DeletedObj: deletedObj,
		TableName:  tableName,
		Where:      where,
	}, nil
}

// CreateManyInput 批量创建记录的输入参数
type CreateManyInput struct {
	TableName      string
	Data           []map[string]any
	SkipDuplicates bool
	ReturnIdList   bool
}

func newCreateManyInput(tableName string, param graphql.ResolveParams) (*CreateManyInput, error) {
	dataValue, ok := param.Args[FieldData].([]any)
	if !ok {
		return nil, bizerrors.Errorf("data must be []any")
	}

	// 验证数据数组长度
	if len(dataValue) == 0 {
		return nil, bizerrors.Errorf("data cannot be empty")
	}
	if len(dataValue) > MaxCreateManyBatchSize {
		return nil, bizerrors.Errorf(
			"createMany batch size exceeds limit: %d > %d",
			len(dataValue),
			MaxCreateManyBatchSize,
		)
	}

	var data []map[string]any
	for _, item := range dataValue {
		// 直接转换为 map
		if m, ok := cast.ToStringMapE(item); ok == nil {
			data = append(data, m)
		}
	}

	// 获取 skipDuplicates 参数，默认为 false
	skipDuplicates := false
	if skip, ok := param.Args[FieldSkipDuplicates].(bool); ok {
		skipDuplicates = skip
	}

	// 检查是否选择了 idList 字段
	returnIdList := hasSelectedField(param, FieldIdList)

	return &CreateManyInput{
		TableName:      tableName,
		Data:           data,
		SkipDuplicates: skipDuplicates,
		ReturnIdList:   returnIdList,
	}, nil
}

// UpdateManyInput 批量更新记录的输入参数
type UpdateManyInput struct {
	TableName  string
	Where      map[string]any
	Data       map[string]any
	Take       uint
	RawFilters []RawSQLFilter
}

func newUpdateManyInput(tableName string, param graphql.ResolveParams) (*UpdateManyInput, error) {
	// 获取 take 参数，必需
	takeVal, ok := param.Args[FieldTake]
	if !ok {
		return nil, bizerrors.Errorf("take is required")
	}

	take, err := cast.ToUintE(takeVal)
	if err != nil {
		return nil, bizerrors.Errorf("take must be an integer, val = %v type = %T", takeVal, takeVal)
	}

	// 验证 take 范围
	if take < 1 || take > MaxCreateManyBatchSize {
		return nil, bizerrors.Errorf("take must be between 1 and %d, got %d", MaxCreateManyBatchSize, take)
	}

	// 获取 data 参数，必需
	data, ok := param.Args[FieldData].(map[string]any)
	if !ok {
		return nil, bizerrors.Errorf("data must be map[string]any")
	}

	// 获取 where 参数，可为 nil
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}

	return &UpdateManyInput{
		TableName: tableName,
		Where:     where,
		Data:      data,
		Take:      take,
	}, nil
}

// DeleteManyInput 批量删除记录的输入参数
type DeleteManyInput struct {
	TableName  string
	Where      map[string]any
	Take       uint
	RawFilters []RawSQLFilter
}

func newDeleteManyInput(tableName string, param graphql.ResolveParams) (*DeleteManyInput, error) {
	// 获取 take 参数，必需
	takeVal, ok := param.Args[FieldTake]
	if !ok {
		return nil, bizerrors.Errorf("take is required")
	}

	take, err := cast.ToUintE(takeVal)
	if err != nil {
		return nil, bizerrors.Errorf("take must be an integer, val = %v type = %T", takeVal, takeVal)
	}

	// 验证 take 范围
	if take < 1 || take > MaxCreateManyBatchSize {
		return nil, bizerrors.Errorf("take must be between 1 and %d, got %d", MaxCreateManyBatchSize, take)
	}

	// 获取 where 参数，可为 nil
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}

	return &DeleteManyInput{
		TableName: tableName,
		Where:     where,
		Take:      take,
	}, nil
}

// hasSelectedField 检查 GraphQL 查询中是否选择了指定字段
func hasSelectedField(p graphql.ResolveParams, fieldName string) bool {
	if p.Info.FieldASTs == nil {
		return false
	}

	for _, fieldAST := range p.Info.FieldASTs {
		if fieldAST.SelectionSet != nil {
			for _, selection := range fieldAST.SelectionSet.Selections {
				if field, ok := selection.(*ast.Field); ok {
					if field.Name != nil && field.Name.Value == fieldName {
						return true
					}
				}
			}
		}
	}
	return false
}

// AggregateInput 聚合查询的输入参数
type AggregateInput struct {
	TableName  string
	Where      map[string]any
	RawFilters []RawSQLFilter
	Count      map[string]bool // field name -> true, 或 "_all" -> true
	Avg        map[string]bool // field name -> true
	Sum        map[string]bool // field name -> true
	Min        map[string]bool // field name -> true
	Max        map[string]bool // field name -> true
}

func newAggregateInput(tableName string, param graphql.ResolveParams) (*AggregateInput, error) {
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}

	input := &AggregateInput{
		TableName: tableName,
		Where:     where,
		Count:     make(map[string]bool),
		Avg:       make(map[string]bool),
		Sum:       make(map[string]bool),
		Min:       make(map[string]bool),
		Max:       make(map[string]bool),
	}

	// 解析聚合字段
	aggregateFields := []struct {
		key    string
		target map[string]bool
	}{
		{Field_Count, input.Count},
		{Field_Avg, input.Avg},
		{Field_Sum, input.Sum},
		{Field_Min, input.Min},
		{Field_Max, input.Max},
	}

	for _, af := range aggregateFields {
		parseAggregateField(param.Args, af.key, af.target)
	}

	// 验证至少选择了一个聚合操作
	if !hasAnyAggregate(input) {
		return nil, bizerrors.Errorf("at least one aggregate operation must be specified")
	}

	return input, nil
}

// parseAggregateField parses a single aggregate field from args
func parseAggregateField(args map[string]any, key string, target map[string]bool) {
	arg, ok := args[key]
	if !ok {
		return
	}

	argMap, ok := arg.(map[string]any)
	if !ok {
		return
	}

	for field, val := range argMap {
		if boolVal, ok := val.(bool); ok && boolVal {
			target[field] = true
		}
	}
}

// hasAnyAggregate checks if any aggregate operation is specified
func hasAnyAggregate(input *AggregateInput) bool {
	return len(input.Count) > 0 || len(input.Avg) > 0 || len(input.Sum) > 0 ||
		len(input.Min) > 0 || len(input.Max) > 0
}

func parseSelectArg(selectArg any) (map[string]bool, error) {
	selectMap, ok := selectArg.(map[string]any)
	if !ok {
		return map[string]bool{}, nil
	}

	result := make(map[string]bool)
	for field, val := range selectMap {
		if boolVal, ok := val.(bool); ok && boolVal {
			result[field] = true
		}
	}
	if len(result) == 0 {
		return nil, bizerrors.Errorf("at least one field must be selected when using select parameter")
	}

	return result, nil
}

// CountInput count查询的输入参数
type CountInput struct {
	TableName  string
	Where      map[string]any
	RawFilters []RawSQLFilter
	Select     map[string]bool // field name -> true, 或 "_all" -> true（如果为nil表示简单计数）
}

func newCountInput(tableName string, param graphql.ResolveParams) (*CountInput, error) {
	where, err := getWhere(param.Args)
	if err != nil {
		return nil, err
	}

	input := &CountInput{
		TableName: tableName,
		Where:     where,
	}

	// 解析 select 参数（可选）
	if selectArg, ok := param.Args[FieldSelect]; ok {
		selectMap, err := parseSelectArg(selectArg)
		if err != nil {
			return nil, err
		}
		if selectMap != nil {
			input.Select = selectMap
		}
	}

	return input, nil
}
