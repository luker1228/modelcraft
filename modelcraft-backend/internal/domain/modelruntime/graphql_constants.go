package modelruntime

// GraphQL查询和变更的参数字段名常量
const (
	// FieldWhere GraphQL where查询条件参数
	FieldWhere = "where"
	// FieldData GraphQL data数据输入参数
	FieldData = "data"
	// FieldID GraphQL id字段名
	FieldID = "id"
	// FieldLabel GraphQL __label 字段名（用于显示 displayField 的值）
	FieldLabel = "__label"
	// FieldReturnUpdatedObj 返回更新后的对象字段名
	FieldReturnUpdatedObj = "returnUpdatedObj"
	// FieldReturnCreatedObj 返回创建的对象字段名
	FieldReturnCreatedObj = "returnCreatedObj"
	// FieldReturnDeletedObj 返回删除的对象字段名
	FieldReturnDeletedObj = "returnDeletedObj"
	// FieldSkipDuplicates 跳过重复项参数字段名
	FieldSkipDuplicates = "skipDuplicates"

	// 返回结果的字段名
	// FieldSuccess 操作是否成功的结果字段名
	FieldSuccess = "success"
	// FieldUpdatedObj 更新后的对象结果字段名
	FieldUpdatedObj = "updatedObj"
	// FieldCreatedObj 创建的对象结果字段名
	FieldCreatedObj = "createdObj"
	// FieldDeletedObj 删除的对象结果字段名
	FieldDeletedObj = "deletedObj"
	// FieldCount 受影响的记录数结果字段名
	FieldCount = "count"
	// FieldIdList 返回的ID列表结果字段名
	FieldIdList = "idList"
	// FieldID2 备用id字段名
	FieldID2 = "id"

	// 分页相关字段名
	// FieldLimit 分页限制参数字段名
	FieldLimit = "limit"
	// FieldOffset 分页偏移量参数字段名
	FieldOffset = "offset"
	// FieldTake 获取的记录数参数字段名
	FieldTake = "take"
	// FieldSkip 跳过的记录数参数字段名
	FieldSkip = "skip"

	// OrderBy 相关字段名
	// FieldOrderBy 排序参数字段名
	FieldOrderBy = "orderBy"
	// FieldSelect count操作的选择字段参数名
	FieldSelect = "select"
	// FieldFieldsCount count操作字段级计数的结果字段名
	FieldFieldsCount = "fieldsCount"

	// Query response wrapper fields
	// FieldItem 查询结果的单个项字段名 (findUnique, findFirst)
	FieldItem = "item"
	// FieldItems 查询结果的多个项字段名 (findMany)
	FieldItems = "items"
	// FieldTimeCost 查询执行时间字段名 (毫秒)
	FieldTimeCost = "timeCost"
	// FieldReqId 请求追踪ID字段名 (UUID v7)
	FieldReqId = "reqId"
	// FieldTotalCount 总记录数字段名 (findMany, 可选)
	FieldTotalCount = "totalCount"

	// 排序方向
	// OrderByAsc 升序排序
	OrderByAsc = "asc"
	// OrderByDesc 降序排序
	OrderByDesc = "desc"

	// GraphQL 类型名后缀
	// ResultTypeSuffix 结果类型后缀
	ResultTypeSuffix = "Result"
	// InputTypeSuffix 输入类型后缀
	InputTypeSuffix = "Input"

	// 聚合操作相关字段名
	// Field_Count 聚合操作中的计数字段名
	Field_Count = "_count"
	// Field_Avg 聚合操作中的平均值字段名
	Field_Avg = "_avg"
	// Field_Sum 聚合操作中的求和字段名
	Field_Sum = "_sum"
	// Field_Min 聚合操作中的最小值字段名
	Field_Min = "_min"
	// Field_Max 聚合操作中的最大值字段名
	Field_Max = "_max"
	// Field_All 聚合计数中的_all特殊字段名
	Field_All = "_all"
)

// GraphQL 操作名称常量
const (
	// OperationFindUnique 查找唯一记录操作
	OperationFindUnique = "findUnique"
	// OperationFindFirst 查找第一个记录操作
	OperationFindFirst = "findFirst"
	// OperationFindMany 查找多个记录操作
	OperationFindMany = "findMany"
	// OperationCreate 创建单个记录操作
	OperationCreate = "create"
	// OperationUpdate 更新单个记录操作
	OperationUpdate = "update"
	// OperationDelete 删除单个记录操作
	OperationDelete = "delete"
	// OperationCreateMany 批量创建记录操作
	OperationCreateMany = "createMany"
	// OperationUpdateMany 批量更新记录操作
	OperationUpdateMany = "updateMany"
	// OperationDeleteMany 批量删除记录操作
	OperationDeleteMany = "deleteMany"
	// OperationAggregate 聚合查询操作
	OperationAggregate = "aggregate"
	// OperationCount 计数查询操作
	OperationCount = "count"
)

// GraphQL 顶级类型名常量
const (
	// TypeQuery GraphQL Query根类型
	TypeQuery = "Query"
	// TypeMutation GraphQL Mutation根类型
	TypeMutation = "Mutation"
)

// 默认值常量
const (
	// DefaultLimit 默认查询限制数
	DefaultLimit = 10
	// DefaultOffset 默认偏移量
	DefaultOffset = 0
)

// 批量操作限制常量
const (
	// MaxCreateManyBatchSize createMany操作的最大批量大小
	MaxCreateManyBatchSize = 1000
)
