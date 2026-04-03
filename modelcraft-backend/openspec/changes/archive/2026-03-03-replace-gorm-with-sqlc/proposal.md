## Why

sqlc 的隐式行为（零值更新、软删除、魔法 hook）导致代码难以理解和调试，且 `*sql.DB` 类型泄漏到 application 层破坏了 DDD 分层约束。sqlc 通过从 SQL schema 直接生成类型安全的 Go 代码，消除运行时反射和 ORM 魔法，让数据访问层的行为完全可预期。

## What Changes

- **新增** `sqlc.yaml` 配置文件及 `db/queries/*.sql` 命名查询文件
- **新增** `internal/infrastructure/dbgen/` sqlc 生成的类型安全代码（`Querier` 接口 + `Queries` 实现）
- **新增** `TxManager`：显式传递 `dbgen.Querier` 的事务管理器，替换 `s.db.Transaction()` 模式
- **新增** `sql_error_analyzer`：基于 `database/sql` 的错误映射，替换 `gorm_error_analyzer`
- **新增** `sql_connection.go`：基于 `*sql.DB` 的连接管理，替换 `db_connection.go`
- **替换** 全部 Repository 实现：使用 sqlc 生成的 `Queries`，去除 `GormBaseRepository` 继承
- **删除** `base.go`、`sql_model.go`、`gorm_error_analyzer.go`、`db_connection.go`
- **删除** `go.mod` 中 `sqlc` 和 `go-sql-driver/mysql` 依赖

## Capabilities

### New Capabilities

- `sqlc-data-access`: sqlc 代码生成配置、命名查询文件、生成的 Querier 接口及 Queries 实现
- `tx-manager`: 基于显式 Querier 传递的事务管理器（方案 B），为后续 UnitOfWork 改造预留扩展点

### Modified Capabilities

- `database-migration`: 数据库连接从 `*sql.DB` 切换为 `*sql.DB`，驱动从 gorm mysql driver 切换为 `go-sql-driver/mysql`

## Impact

**代码变更：**
- `internal/infrastructure/repository/`：全部 repository 文件重写，删除 sqlc 相关基础设施
- `internal/app/`：事务调用方式变更（`s.db.Transaction` → `s.txManager.WithTx`），`*sql.DB` 字段替换为 `TxManager`
- `internal/infrastructure/dbgen/`：新增目录，sqlc 生成代码（不手动编辑）

**依赖变更：**
- 移除：`sqlc`、`go-sql-driver/mysql`
- 新增：`github.com/sqlc-dev/sqlc`（开发工具）、`github.com/go-sql-driver/mysql`

**不变：**
- domain 层接口（`ModelRepository` 等）完全不动
- `db/schema/mysql/*.sql`（Atlas 继续管理迁移）
- app service 业务逻辑
- QueryExecutor / 运行时动态 SQL
