## 1. GraphQL Schema 定义

- [x] 1.1 在 `api/graph/schema/model.graphql` 中新增 `DbTableStatus` enum（`TABLE_EXISTS`, `TABLE_MISSING`, `CLUSTER_UNREACHABLE`）
- [x] 1.2 在 `api/graph/schema/model.graphql` 中新增 `ActualConstraintType` enum（`UNIQUE`, `NOT_NULL`）
- [x] 1.3 在 `api/graph/schema/model.graphql` 中新增 `ActualForeignKey` 类型（`referencedTable`, `referencedColumn`, `constraintName`）
- [x] 1.4 在 `api/graph/schema/model.graphql` 中新增 `FieldConflictAspect` enum（`UNIQUE_MISMATCH`, `NOT_NULL_MISMATCH`）和 `FieldConflict` 类型（`aspect`, `expected`, `actual`）
- [x] 1.5 在 `api/graph/schema/model.graphql` 中新增 `DbColumnInfo` 类型（`columnType`, `constraints`, `foreignKey`, `conflicts`）
- [x] 1.6 在 `api/graph/schema/model.graphql` 的 `Model` 类型上新增 `dbTable: DbTableStatus` 字段
- [x] 1.7 在 `api/graph/schema/field.graphql` 的 `Field` 类型上新增 `dbColumn: DbColumnInfo` 字段
- [x] 1.8 修改 `model` query 签名，新增 `withActualSchema: Boolean` 可选参数
- [x] 1.9 运行 `task generate-gql` 重新生成 GraphQL 代码

## 2. Domain 层类型与接口定义

- [x] 2.1 在 `internal/domain/modeldesign/` 新增 `DbTableStatus` 类型（`TABLE_EXISTS`, `TABLE_MISSING`, `CLUSTER_UNREACHABLE`）
- [x] 2.2 在 `internal/domain/modeldesign/` 新增 `ActualForeignKey` 结构体（`ReferencedTable`, `ReferencedColumn`, `ConstraintName`）
- [x] 2.3 在 `internal/domain/modeldesign/` 新增 `FieldConflict` 结构体（`Aspect`, `Expected`, `Actual`）和 `FieldConflictAspect` 类型
- [x] 2.4 在 `internal/domain/modeldesign/` 新增 `DbColumnInfo` 结构体（`ColumnType`, `Constraints`, `ForeignKey`, `Conflicts`）
- [x] 2.5 在 `internal/domain/modeldesign/` 新增 `ActualSchemaResult` 结构体（`Status DbTableStatus`, `Fields map[string]*DbColumnInfo`）
- [x] 2.6 在 `internal/domain/modeldesign/` 新增 `ActualSchemaService` 接口，定义 `QueryActualSchema(ctx, db *sql.DB, databaseName, tableName string, fields []*FieldDefinition) (*ActualSchemaResult, error)`
- [x] 2.7 为 domain 层新增类型编写单元测试：验证 `DbColumnInfo` 的 `conflicts` 计算逻辑（`UNIQUE_MISMATCH` / `NOT_NULL_MISMATCH` 双向冲突）、虚拟字段判断、`ActualSchemaResult` 各状态的字段访问行为

## 3. Infrastructure 层实现

- [x] 3.1 在 `internal/infrastructure/database/ddl/` 新建 `actual_schema_service.go`，实现 `queryColumns(ctx, db, dbName, tableName)`：查询 `INFORMATION_SCHEMA.COLUMNS`，字段包含 `COLUMN_NAME`, `DATA_TYPE`, `IS_NULLABLE`, `COLUMN_KEY`（`COLUMN_KEY='UNI'` 表示该列有单列 UNIQUE 约束）
- [x] 3.2 在同文件实现 `queryForeignKeys(ctx, db, dbName, tableName)`：查询 `INFORMATION_SCHEMA.KEY_COLUMN_USAGE JOIN TABLE_CONSTRAINTS WHERE CONSTRAINT_TYPE='FOREIGN KEY'`，返回按列名索引的 `map[string]*ActualForeignKey`
- [x] 3.3 实现 `ActualSchemaServiceImpl.QueryActualSchema()`：调用 3.1、3.2，将结果组装为 `map[string]*DbColumnInfo`，并对比 `FieldDefinition.IsUnique` / `NonNull` 与实际约束计算 `conflicts`；表不存在（列查询返回空）时返回 `TABLE_MISSING` 状态
- [x] 3.4 为 3.1、3.2、3.3 编写单元测试（表驱动测试，覆盖 UNIQUE/NOT_NULL/FK 各场景、冲突双向计算、表不存在场景）

## 4. Application 层 Use Case

- [x] 4.1 在 `internal/app/modeldesign/` 新增 `ActualSchemaQueryUseCase` 结构体，持有 `ActualSchemaService` 和 `*repository.ClusterConnectionManager`
- [x] 4.2 实现 `Query(ctx, model *DataModel, orgName string) (*ActualSchemaResult, error)`：通过 `ClusterConnectionManager` 获取 DB 连接，调用 `ActualSchemaService`；连接失败时返回 `CLUSTER_UNREACHABLE` 状态，不返回 error
- [x] 4.3 为 `ActualSchemaQueryUseCase` 编写单元测试（mock `ActualSchemaService` 和 `ClusterConnectionManager`，覆盖 TABLE_EXISTS / TABLE_MISSING / CLUSTER_UNREACHABLE 三种路径）

## 5. Interfaces 层（Resolver + Mapper）

- [x] 5.1 修改 `internal/interfaces/graphql/model.resolvers.go` 的 `Model()` resolver：接收 `withActualSchema *bool` 参数，为 `true` 时调用 `ActualSchemaQueryUseCase.Query()`，结果传入 mapper
- [x] 5.2 修改 `internal/interfaces/graphql/adapter/model_mapper.go` 的 `ConvertToGraphQLModel` 签名，增加 `actualResult *ActualSchemaResult` 参数；`actualResult` 为 nil 时 `dbTable` 和所有 `dbColumn` 置 null
- [x] 5.3 在 mapper 的 `ConvertFieldToGraphQLField` 中：虚拟字段（`ENUM_LABEL`）的 `dbColumn` 置 null；其他字段从 `actualResult.Fields[field.Name]` 取值填充 `dbColumn`

## 6. 依赖注入

- [x] 6.1 在应用启动的依赖注入位置实例化 `ActualSchemaServiceImpl` 并注册
- [x] 6.2 实例化 `ActualSchemaQueryUseCase` 并注入到 query resolver

## 7. 集成测试

- [x] 7.1 在 `tests/design/model/` 新增集成测试：`withActualSchema=true` 且表存在时，`dbTable=TABLE_EXISTS`，各字段 `dbColumn.columnType`、`constraints`、`foreignKey` 正确返回
- [x] 7.2 新增集成测试：`withActualSchema=false`（或省略）时 `dbTable` 为 null，所有 `dbColumn` 为 null
- [x] 7.3 新增集成测试：设计时 `isUnique=true` 但 DB 无 UNIQUE 约束时，`dbColumn.conflicts` 包含 `UNIQUE_MISMATCH`
- [x] 7.4 新增集成测试：虚拟字段（`ENUM_LABEL`）的 `dbColumn` 始终为 null

## 8. 验收

- [x] 8.1 运行 `task test-unit` 确认单元测试全部通过
- [ ] 8.2 运行 `task auto-test` 确认集成测试全部通过
- [ ] 8.3 运行 `task check-all` 确认格式、lint、vet 全部通过
