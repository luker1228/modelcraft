## Context

`model(id)` query 目前只返回设计时定义的字段信息，前端无法得知字段在实际数据库中的真实列类型、UNIQUE/NOT NULL 约束以及 FOREIGN KEY 约束。

**Field 的两种用途：**
- 设计态：用于 schema 管理、部署、UI 展示
- 运行态：用于生成 GraphQL schema，执行查询

运行态只有一处依赖真实 DB schema：`IsUnique`（影响唯一性验证）。其他字段（类型、NonNull 等）运行态从设计时定义读取，不查真实 DB。因此"查询真实 schema"是设计态的低频诊断操作。

**现有代码基础：**
- `MySQLSchemaComparisonService.introspectTable()` 能查列信息和主键（`comparison_service.go`）
- `ClusterConnectionManager` 管理到客户数据库的连接
- `SchemaIntrospector`（`introspector.go`）有独立实现，但当前未查 FK 和 UNIQUE

**缺口：**
- **新增 infrastructure 能力**：在 `internal/infrastructure/database/ddl/actual_schema_service.go` 新建独立实现，提供两个原子查询：
  - `queryColumns()`：查 `INFORMATION_SCHEMA.COLUMNS`，通过 `COLUMN_KEY='UNI'` 直接判断单列 UNIQUE，不需要额外查 `INFORMATION_SCHEMA.STATISTICS`
  - `queryForeignKeys()`：查 `INFORMATION_SCHEMA.KEY_COLUMN_USAGE JOIN TABLE_CONSTRAINTS`，按列名索引 FK 信息
- 不修改、不扩展 `comparison_service.go` 及其 `introspectTable()`，两套实现完全隔离
- `Field` 类型没有 `dbColumn` 字段
- `Model` 类型没有 `dbTable` 字段

## Goals / Non-Goals

**Goals:**
- `model(id)` 支持通过 GraphQL 字段选择触发真实 schema 查询（选择 `actual` 或 `actualSchemaStatus` 字段时触发）
- `Model` 新增 `dbTable`，表达表级查询状态（TABLE_EXISTS / TABLE_MISSING / CLUSTER_UNREACHABLE）
- `Field` 新增 `dbColumn: DbColumnInfo`，包含：实际列类型、UNIQUE/NOT_NULL 约束、FOREIGN KEY、与设计时定义的冲突
- 虚拟字段（`ENUM_LABEL`）直接跳过，`dbColumn` 为 null
- `RELATION` 字段（逻辑外键）的 `actual.foreignKey` 展示真实 DB FK 约束（与 `RelationConfig` 逻辑外键无关）
- 向后兼容：不 select `actual` / `actualSchemaStatus` 时行为与现在完全一致

**Non-Goals:**
- `models` 列表 query 不支持
- `modelByName` 暂不支持
- 不做 schema 修复（repair usecase 已有）
- 不支持 TYPE_MISMATCH 冲突检测（暂不纳入）
- 不支持 CHECK 约束（MySQL 8.0+，暂不纳入）
- 表不存在时不报错，通过 `actualSchemaStatus` 表达

## Decisions

### 决策 1：`actual` 嵌入 `Field`，`actualSchemaStatus` 挂在 `Model`

**选择：**
```graphql
type Model {
  # ...现有字段不变
  actualSchemaStatus: ActualSchemaStatus  # null = 未查询
}

type Field {
  # ...现有字段不变
  actual: FieldActual                     # null = 虚拟字段 或 未查询
}
```

**备选方案（已否决）：** `Model.actualSchema: ModelActualSchema`，内含独立的 `fields: [FieldActualSchema!]!` 数组。

**否决理由：**
- 独立数组需要前端按 `fieldName` 字符串做 join，GraphQL object graph 本身就是为了消灭这种模式
- 无法直接表达"该字段在 DB 里缺失"与"虚拟字段跳过"的区别，前端仍需回查 `model.fields`
- `FieldActualSchema.fieldName: String!` 是退化的字符串引用，结构脆弱

**选择理由：**
- 零 join，前端每个 field 自带实际状态，直接渲染
- 虚拟字段 `actual == null` + `format == ENUM_LABEL` 语义明确，无歧义
- 字段在 DB 里缺失 → `actual == null` + `actualSchemaStatus == TABLE_EXISTS`，可精确表达

### 决策 2：`actualSchemaStatus` 为可空，null 表示未查询

**选择：** `actualSchemaStatus: ActualSchemaStatus`（可空），不引入 `NOT_QUERIED` 枚举值。

**理由：** `NOT_QUERIED` 本质上是"你没有在 query 里 select 这个字段"，用数据值表达查询层面的状态是职责混乱。GraphQL 的 null 已经表达了"未返回"的语义，不需要重复。

**前端判断逻辑：**
```
dbTable == null             → 未查询，不渲染实际列区域
dbTable == TABLE_EXISTS     → 查到了，逐字段渲染 dbColumn
dbTable == TABLE_MISSING    → 表不存在，全部字段标灰 + 提示
dbTable == CLUSTER_UNREACHABLE → 集群连不上，提示
```

### 决策 3：`foreignKey` 与 `constraints` 平行

**选择：**
```graphql
type FieldActual {
  columnType:  String!
  constraints: [ActualConstraintType!]!  # UNIQUE / NOT_NULL
  foreignKey:  ActualForeignKey          # 独立字段
  conflicts:   [FieldConflict!]!
}
```

**理由：** FK 的语义（引用哪张表、哪列、约束名）与 UNIQUE/NOT_NULL 布尔型约束有本质区别。`nullable` 字段去掉，由 `NOT_NULL` 在 `constraints` 里表达，避免两个字段表达同一状态产生歧义。

### 决策 4：conflicts 由后端计算

**选择：** 后端在返回 `FieldActual` 时直接计算 `conflicts`，前端不自行 diff。

**理由：** 设计类型到 MySQL 类型的映射逻辑（`TypeMapper`）在后端已有，不应重复到前端。前端只需判断 `conflicts.length > 0` 即可决定是否展示警告。

**冲突类型（当前版本）：**
- `UNIQUE_MISMATCH`：`Field.IsUnique` 与实际 UNIQUE 约束不符（任意方向）
- `NOT_NULL_MISMATCH`：`Field.NonNull` 与实际 NOT NULL 不符（任意方向）

`expected` / `actual` 字段使用字符串值，如 `"true"` / `"false"`，便于前端直接展示文案。

### 决策 5：新增 `ActualSchemaService` domain 接口，实现在 infrastructure 层

**选择：** 在 `internal/domain/modeldesign/` 定义 `ActualSchemaService` 接口，在 `internal/infrastructure/database/ddl/` 实现，App 层新增 `ActualSchemaQueryUseCase`。

**理由：** 遵循 DDD 分层依赖规则。与现有 `SchemaComparisonService` 模式完全一致。

### 决策 6：通过 `withActualSchema` 参数控制是否查询真实 schema

**选择：** `model` query 新增 `withActualSchema: Boolean` 可选参数（默认 `false`），后端在 `Model()` resolver 中统一判断是否触发真实 schema 查询。

```graphql
model(projectSlug: String!, id: ID!, withActualSchema: Boolean): GetModelPayload!
```

**否决方案：** 通过 field resolver 懒加载（客户端 select `dbColumn` 或 `dbTable` 字段时自动触发）。

**否决理由：**
- gqlgen 的执行模型中，后端无法在 `Field.dbColumn` resolver 里可靠感知"整个 model 的真实 schema 已查询"，极易造成 N+1（每个 field 各触发一次 DB 查询）
- 规避 N+1 需要用 `FieldSelectionChecker` 在父 resolver 里预判子字段选择，本质上和加参数一样，但更隐晦、更难维护
- 强制要求前端同时 select `dbTable` 才能避免 N+1，是隐性约定，GraphQL schema 无法表达和强制

**选择理由：**
- 后端判断条件极简：`if withActualSchema { 查真实 schema }`，无歧义
- 与项目现有模式一致（`UpdateModel` 用 `FieldSelectionChecker` 判断是否回查 model，此处更直接）
- 调用方意图明确，`withActualSchema: true` 即表示"我知道这是重操作，我需要真实 schema"

**实现方式：**

```go
func (r *queryResolver) Model(ctx context.Context, projectSlug, id string,
    withActualSchema *bool) (*GetModelPayload, error) {

    modelEntity, err := r.ModelDesignService.GetModelByID(...)

    var actualResult *ActualSchemaResult  // nil 表示不查
    if withActualSchema != nil && *withActualSchema {
        actualResult, _ = r.ActualSchemaQueryUseCase.Query(ctx, modelEntity)
    }

    // mapper 根据 actualResult 是否为 nil 决定填充 dbTable 和 Field.dbColumn
    graphqlModel = adapter.ModelMapper.ConvertToGraphQLModel(modelEntity, actualResult)
}
```

`ActualSchemaResult` 结构：

```go
type ActualSchemaResult struct {
    Status DbTableStatus                  // TABLE_EXISTS / TABLE_MISSING / CLUSTER_UNREACHABLE
    Fields map[string]*DbColumnInfo       // fieldName → 列信息，仅 TABLE_EXISTS 时有值
}
```

`actualResult == nil` 时，mapper 将 `Model.dbTable` 和所有 `Field.dbColumn` 均置为 null，与现有行为完全一致。

## GraphQL Schema 变更

```graphql
# model query 签名变更
extend type Query {
  model(projectSlug: String!, id: ID!, withActualSchema: Boolean): GetModelPayload!
  # ...其他 query 不变
}

# Model 新增
type Model {
  # ...现有字段不变
  dbTable: DbTableStatus                  # null = 未查询
}

enum DbTableStatus {
  TABLE_EXISTS
  TABLE_MISSING
  CLUSTER_UNREACHABLE
}

# Field 新增
type Field {
  # ...现有字段不变
  dbColumn: DbColumnInfo                  # null = 虚拟字段 或 未查询
}

type DbColumnInfo {
  columnType:  String!
  constraints: [ActualConstraintType!]!
  foreignKey:  ActualForeignKey
  conflicts:   [FieldConflict!]!
}

enum ActualConstraintType {
  UNIQUE
  NOT_NULL
}

type ActualForeignKey {
  referencedTable:  String!
  referencedColumn: String!
  constraintName:   String!
}

type FieldConflict {
  aspect:   FieldConflictAspect!
  expected: String!
  actual:   String!
}

enum FieldConflictAspect {
  UNIQUE_MISMATCH
  NOT_NULL_MISMATCH
}
```

## Risks / Trade-offs

- **N+1 DB 查询风险** → `withActualSchema: true` 时在 `Model()` resolver 里一次性完成所有查询（1 次 DB 连接，3 条 SQL），结果通过 mapper 传递，不存在 N+1
- **`Field.dbColumn == null` 双重含义** → 虚拟字段（`format == ENUM_LABEL`）和未查询都是 null，前端通过 `dbTable` 是否为 null 区分这两种情况，无歧义
- **`Model` 类型新增 `dbTable` 字段** → 不 select 时不影响性能，99% 的查询不受影响；这是真实存在的类型污染，权衡后可接受
- **复合外键** → 单字段只展示该列参与的 FK，多列共享同一 `constraintName`，前端可按 `constraintName` 分组识别复合 FK
