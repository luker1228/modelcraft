package modelruntime

// gqlTypeName 将模型名称转换为合法的 GraphQL 类型名前缀。
//
// 背景：GraphQL 规范要求类型名必须以字母或下划线开头（/^[_A-Za-z]/）。
// ModelCraft 支持从外部数据库导入表结构（createdVia: IMPORTED），
// 这类表名可能以数字开头（如 "123orders"），而自建模型在创建时会
// 通过 validateModelDisplayName 校验拒绝此类名称。
// 若直接用原始表名拼接类型名（如 "123ordersWhereInput"），
// graphql-go 在注册 schema 时会因类型名不合法而 panic。
//
// 修复策略：统一在模型名前加 "T" 前缀，使任意模型名都能生成合法类型名。
// 例如：
//
//	"User"      -> "TUser"      -> "TUserWhereInput"
//	"123orders" -> "T123orders" -> "T123ordersWhereInput"
//
// 注意：前端 runtime-query-builder 中所有类型名也必须经过同样的转换，
// 以保证前后端生成的类型名一致。
func gqlTypeName(modelName string) string {
	return "T" + modelName
}
