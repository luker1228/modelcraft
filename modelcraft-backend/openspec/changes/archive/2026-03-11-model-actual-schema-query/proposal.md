## Why

前端模型详情页需要展示该模型在实际数据库中的真实状态（列类型、外键、约束），以便用户判断设计时定义与实际 DB 是否一致，目前 `model(id)` query 只返回设计时定义，无法满足这一需求。

## What Changes

- `model(id)` query 新增可选参数 `withActualSchema: Boolean`（默认 `false`）
- `Model` 类型新增 `dbTable: DbTableStatus`，表达表级查询状态（TABLE_EXISTS / TABLE_MISSING / CLUSTER_UNREACHABLE）
- `Field` 类型新增 `dbColumn: DbColumnInfo`，仅当 `withActualSchema=true` 且表存在时填充；虚拟字段（ENUM_LABEL）始终为 null
- 新增 GraphQL 类型：`DbColumnInfo`、`ActualConstraintType`（enum）、`ActualForeignKey`、`FieldConflict`、`FieldConflictAspect`（enum）、`DbTableStatus`（enum）
- 后端计算设计时定义与实际 DB 的冲突（`UNIQUE_MISMATCH` / `NOT_NULL_MISMATCH`），直接填入 `dbColumn.conflicts`
- 新增 infrastructure 原子查询能力：`queryColumns()`（含 `COLUMN_KEY` 判断 UNIQUE）和 `queryForeignKeys()`，独立于现有 `introspectTable()` 实现

## Capabilities

### New Capabilities

- `model-actual-schema`: 通过 `model(id, withActualSchema: true)` 查询模型字段的实际 DB schema，包含列类型、UNIQUE/NOT_NULL 约束、FOREIGN KEY，以及与设计时定义的冲突

### Modified Capabilities

## Impact

- `api/graph/schema/model.graphql`：修改 `model` query 签名，新增 `DbTableStatus`、`DbColumnInfo` 等相关类型；`Model` 新增 `dbTable` 字段
- `api/graph/schema/field.graphql`：`Field` 类型新增 `dbColumn` 字段
- `internal/domain/modeldesign/`：新增 `ActualSchemaService` 接口及相关 domain 类型（`DbTableStatus`、`DbColumnInfo`、`ActualForeignKey`、`FieldConflict`、`ActualSchemaResult`）
- `internal/infrastructure/database/ddl/actual_schema_service.go`：新建文件，实现 `ActualSchemaServiceImpl`
- `internal/app/modeldesign/`：新增 `ActualSchemaQueryUseCase`
- `internal/interfaces/graphql/model.resolvers.go`：`Model()` resolver 处理 `withActualSchema` 参数
- `internal/interfaces/graphql/adapter/model_mapper.go`：`ConvertToGraphQLModel` 增加 `actualResult` 参数，填充 `dbTable` 和 `dbColumn`
- 需运行 `task generate-gql` 重新生成 GraphQL 代码
