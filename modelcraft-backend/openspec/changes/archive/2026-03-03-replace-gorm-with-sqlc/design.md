## Context

当前数据访问层基于 sqlc v2，存在以下问题：
1. `*sql.DB` 类型从 infrastructure 层泄漏到 application 层（`ModelDesignAppService.db *sql.DB`），违反 DDD 分层约束
2. sqlc 隐式行为（struct 更新零值问题、自动软删除等）导致 bug 难以排查
3. 事务内部重建 Repository（`NewModelRepository(tx)`），代码冗余
4. Repository 接口 `WithTx(tx *sql.DB)` 暴露了 sqlc 具体类型，无法替换底层实现

现有 Schema 全部为手写 SQL（`db/schema/mysql/*.sql`），由 Atlas 管理迁移，与 sqlc 完全兼容。`go.mod` 中已存在 `github.com/jmoiron/sqlx v1.4.0`，说明团队对 SQL-first 方向有基础认知。

## Goals / Non-Goals

**Goals:**
- 用 sqlc 生成类型安全的数据访问代码，消除 sqlc 魔法行为
- 通过 `TxManager`（方案 B：显式传递 Querier）解除 `*sql.DB` 对 application 层的污染
- 为后续 UnitOfWork 改造预留接口扩展点
- domain 层接口零改动

**Non-Goals:**
- 不涉及 QueryExecutor / 运行时动态 SQL
- 不引入 UnitOfWork（后续独立变更）
- 不引入 testcontainers（集成测试由 `task auto-test` 覆盖）
- 不改变 Atlas 迁移管理方式

## Decisions

### D1：sqlc 作为代码生成工具

**选择 sqlc，不选 ent（Meta）**

ent 采用图模型 + 代码生成，需要重写 schema 定义层，迁移成本高。sqlc 直接读取已有的 `db/schema/mysql/*.sql`，零 schema 改写成本。

项目已有完整的手写 SQL schema，sqlc 是自然延伸，而非引入新范式。

### D2：动态可选过滤使用 nullable trick

**选择 nullable trick，不选 squirrel（SQL builder）**

```sql
-- 可选过滤：传 nil 表示不过滤
AND (? IS NULL OR name LIKE CONCAT('%', ?, '%'))
```

squirrel 引入新依赖且返回 `*sql.Rows` 需手动 scan，失去类型安全。nullable trick 保持 sqlc 的类型安全优势，SQL 可读性高，且 MySQL 优化器可正确处理此模式。

### D3：TxManager 方案 B（显式传递 Querier）

**不选方案 A（context 携带 tx）**

```go
type TxManager interface {
    WithTx(ctx context.Context, fn func(ctx context.Context, q dbgen.Querier) error) error
}
```

context 携带 tx 隐式性强，调试困难，违反"显式优于隐式"原则。显式传递 Querier 让依赖关系一目了然，且 TxManager 接口稳定，后续 UnitOfWork 只需在外层封装，不改动 TxManager 实现。

### D4：Nullable 字段使用指针类型

**选择 `*string`/`*time.Time`，不选 `sql.NullString`**

sqlc 配置 `null_style: "option"` 生成指针类型，与项目 Go 代码风格一致，避免 `.Valid`/`.String` 的双字段访问模式。

### D5：JSON 字段映射到具体 Go struct

**在 repository 层做 json.Marshal/Unmarshal，不用 `json.RawMessage` 透传**

JSON 字段（`validation`、`metadata`、`source_fields` 等）在 repository 层转换为具体 domain struct，保持 domain 层类型安全。这与现有 sqlc 自定义类型（`StringSlice`）的职责相同，只是实现更显式。

### D6：Repository 测试策略

**纯函数单元测试 + auto-test 集成测试，不 mock Querier**

mock sqlc 生成的 `Querier` 接口本质上是测试生成代码能否被调用，没有业务价值。`toDomain()`/`mapError()` 为纯函数，直接单元测试。SQL 查询正确性由 `task auto-test` 的真实数据库覆盖。

## Risks / Trade-offs

**[风险] MySQL sqlc 类型推断不如 PostgreSQL 成熟**
→ 缓解：对复杂 nullable 列使用 `sqlc:type` 注解手动覆盖类型推断

**[风险] JSON 字段需要手动 Marshal/Unmarshal，样板代码增加**
→ 缓解：每个 JSON 字段的转换逻辑集中在 repository 的 `toDomain()`/`toRow()` 函数，测试覆盖这两个纯函数即可

**[风险] 大量 Repository 同时迁移，集成测试风险集中**
→ 缓解：按模块顺序迁移（基础设施 → 核心 domain → 外围），每个模块迁移后运行 auto-test 验证

**[Trade-off] nullable trick 导致部分 SQL 参数传两次**
```sql
AND (? IS NULL OR name LIKE CONCAT('%', ?, '%'))
--   ↑ 用于 IS NULL 判断    ↑ 用于实际匹配
```
→ 接受：类型安全和无额外依赖的收益大于参数重复的代价

## Migration Plan

**分三阶段，每阶段独立可验证：**

**阶段 1：基础设施（不影响业务）**
1. 安装 sqlc，写 `sqlc.yaml`
2. 建 `db/queries/` 目录，翻译现有查询为命名 SQL
3. 生成 `internal/infrastructure/dbgen/`
4. 实现 `TxManager`、`sql_error_analyzer`、`sql_connection.go`
5. 运行 `task build` 验证编译

**阶段 2：逐模块替换 Repository**
按顺序：`project` → `modeldesign`（含 field/relation）→ `enum` → `cluster` → `org/user/membership` → `casbin`
每个模块：写 toDomain 纯函数测试 → 实现新 Repository → 运行 auto-test → 删除旧实现

**阶段 3：清理**
- 删除 sqlc 基础设施文件
- 从 `go.mod` 移除 sqlc 依赖
- 最终 `task check-all` 全量验证

**回滚策略：** 阶段 2 每个模块替换前保留旧文件（加 `_gorm.go` 后缀），验证通过后删除，任意节点可回滚到上一个稳定模块。

## Open Questions

- sqlc 对 `DATETIME(3)` 类型的推断是否需要 `sqlc:type` 注解覆盖为 `time.Time`？需在阶段 1 验证。
- `casbin_*_repository` 依赖 casbin adapter 接口，是否完全由 sqlc 接管还是保留部分 sqlc？需在阶段 2 进入 casbin 模块时确认。
