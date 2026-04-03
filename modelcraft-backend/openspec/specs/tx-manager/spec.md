# tx-manager Specification

## Purpose
TBD - created by archiving change replace-gorm-with-sqlc. Update Purpose after archive.
## Requirements
### Requirement: TxManager 接口定义
项目 SHALL 在 `internal/infrastructure/repository/` 提供 `TxManager` 接口，用于管理数据库事务，接口签名为显式传递 `dbgen.Querier`。

#### Scenario: TxManager 接口签名
- **WHEN** 定义 TxManager 接口
- **THEN** 接口包含 `WithTx(ctx context.Context, fn func(ctx context.Context, q dbgen.Querier) error) error` 方法

### Requirement: TxManager 提交与回滚行为
`TxManager.WithTx` SHALL 在 `fn` 返回 `nil` 时提交事务，返回非 `nil` 错误时回滚事务。

#### Scenario: fn 成功时提交
- **WHEN** `fn` 执行完毕返回 `nil`
- **THEN** 事务提交，数据持久化，`WithTx` 返回 `nil`

#### Scenario: fn 失败时回滚
- **WHEN** `fn` 执行过程中返回非 `nil` 错误
- **THEN** 事务回滚，数据变更撤销，`WithTx` 返回该错误

#### Scenario: fn 内 panic 时回滚
- **WHEN** `fn` 执行过程中发生 panic
- **THEN** 事务回滚，panic 继续向上传播

### Requirement: Querier 显式传递，不通过 context 隐式携带
事务中的 `dbgen.Querier` SHALL 通过 `fn` 的参数显式传递，不存入 context。

#### Scenario: app service 在事务中使用 Querier
- **WHEN** app service 调用 `txManager.WithTx(ctx, func(ctx, q) error { ... })`
- **THEN** app service 使用传入的 `q`（已绑定事务）创建 Repository，不从 context 提取

### Requirement: TxManager 为后续 UnitOfWork 预留扩展点
`TxManager` 接口 SHALL 保持稳定，后续 UnitOfWork 改造通过封装 `TxManager` 实现，不修改其接口。

#### Scenario: UnitOfWork 封装 TxManager
- **WHEN** 后续引入 UnitOfWork
- **THEN** `UnitOfWorkFactory` 内部使用 `TxManager`，`TxManager` 接口本身无需修改

### Requirement: app service 移除 *sql.DB 字段
所有 application service SHALL 移除 `*sql.DB` 类型字段，改用 `TxManager` 接口字段。

#### Scenario: app service 不感知 sqlc
- **WHEN** app service 需要执行事务操作
- **THEN** 通过 `TxManager` 接口调用，app service 不导入 `sqlc` 包
