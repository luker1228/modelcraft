package generators

// GraphQL 字段名常量 (Schema Generator)
const (
	// 返回结果的字段名
	FieldSuccess    = "success"
	FieldUpdatedObj = "updatedObj"
	FieldCreatedObj = "createdObj"
	FieldDeletedObj = "deletedObj"
	FieldID         = "id"

	// 查询和变更的参数名
	FieldWhere            = "where"
	FieldData             = "data"
	FieldReturnUpdatedObj = "returnUpdatedObj"
	FieldReturnCreatedObj = "returnCreatedObj"
	FieldReturnDeletedObj = "returnDeletedObj"

	// 分页相关字段名
	FieldLimit  = "limit"
	FieldOffset = "offset"

	// GraphQL 类型名前缀
	ResultTypeSuffix = "Result"
)
