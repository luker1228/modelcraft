## ADDED Requirements

### Requirement: sqlc 配置与查询文件组织
项目 SHALL 在根目录维护 `sqlc.yaml` 配置文件，`db/queries/` 目录下按领域拆分命名查询文件（`.sql`），生成代码输出到 `internal/infrastructure/dbgen/`。

#### Scenario: sqlc 配置正确生成代码
- **WHEN** 运行 `sqlc generate`
- **THEN** `internal/infrastructure/dbgen/` 目录下生成 `querier.go`（接口）、`db.go`（Queries struct）、`models.go`（PO struct）及各查询实现文件，编译无错误

#### Scenario: 生成代码不手动编辑
- **WHEN** 需要修改查询逻辑
- **THEN** 开发者修改 `db/queries/*.sql` 并重新运行 `sqlc generate`，不直接编辑 `internal/infrastructure/dbgen/` 下的文件

### Requirement: Querier 接口覆盖所有数据访问操作
sqlc 生成的 `Querier` 接口 SHALL 覆盖所有 Repository 所需的数据库操作，每个操作对应一条命名 SQL。

#### Scenario: 命名查询类型映射
- **WHEN** 定义命名查询
- **THEN** 单行查询使用 `:one`，多行查询使用 `:many`，写操作使用 `:exec` 或 `:execresult`，统计查询使用 `:one` 返回 `int64`

### Requirement: 可选过滤条件使用 nullable trick
动态可选过滤条件 SHALL 通过 nullable trick 实现，不引入 SQL builder 依赖。

#### Scenario: 空值表示不过滤
- **WHEN** 调用带可选过滤参数的查询，某参数传 `nil`
- **THEN** 该过滤条件不生效，SQL 等价于不含该 WHERE 子句

#### Scenario: 非空值正常过滤
- **WHEN** 调用带可选过滤参数的查询，某参数传非 `nil` 值
- **THEN** 该过滤条件生效，结果集按该条件过滤

### Requirement: Nullable 字段使用指针类型
sqlc 生成的 PO struct SHALL 使用 `*string`、`*time.Time` 等指针类型表示可空字段，不使用 `sql.NullString`。

#### Scenario: 可空字段为 NULL 时
- **WHEN** 数据库字段值为 NULL
- **THEN** 对应 Go 字段值为 `nil` 指针

#### Scenario: 可空字段有值时
- **WHEN** 数据库字段值非 NULL
- **THEN** 对应 Go 字段为非 nil 指针，解引用得到实际值

### Requirement: JSON 字段在 repository 层转换为具体 struct
数据库 JSON 类型字段 SHALL 在 repository 层的 `toDomain()`/`toRow()` 函数中进行 `json.Marshal`/`json.Unmarshal` 转换，sqlc 层使用 `json.RawMessage` 持有原始数据。

#### Scenario: 读取 JSON 字段
- **WHEN** repository 的 `toDomain()` 函数处理含 JSON 字段的 PO
- **THEN** 调用 `json.Unmarshal` 将 `json.RawMessage` 转换为对应 domain struct，转换失败返回错误

#### Scenario: 写入 JSON 字段
- **WHEN** repository 的 `toRow()` 函数处理含 JSON 字段的 domain 对象
- **THEN** 调用 `json.Marshal` 将 domain struct 序列化为 `json.RawMessage`

### Requirement: toDomain 和 toRow 为可单元测试的纯函数
PO 与 domain 对象之间的转换 SHALL 实现为纯函数（无副作用、无外部依赖），可独立单元测试。

#### Scenario: toDomain 单元测试
- **WHEN** 给定一个填充了所有字段的 PO struct
- **THEN** `toDomain()` 返回对应的 domain 对象，字段值一一映射，无需 mock 任何依赖

#### Scenario: toRow 单元测试
- **WHEN** 给定一个完整的 domain 对象
- **THEN** `toRow()` 返回对应的 sqlc params struct，字段值一一映射
