## MODIFIED Requirements

### Requirement: Atlas Go SDK Integration
The application SHALL use Atlas Go SDK for schema migration instead of CLI or external processes. Database connection is provided via `*sql.DB` (standard library), not `*sql.DB`.

#### Scenario: Using Atlas Go SDK
- **WHEN** the application loads migration functionality
- **THEN** it imports Atlas Go SDK packages directly
- **AND** does not rely on `os/exec` to call Atlas CLI
- **AND** uses Atlas's `schema` and `migrate` packages programmatically
- **AND** constructs migration plans in-memory before execution
- **AND** receives a `*sql.DB` connection, not a `*sql.DB`

## ADDED Requirements

### Requirement: 数据库连接基于 database/sql
应用程序 SHALL 使用 `*sql.DB`（标准库）作为数据库连接类型，不使用 `*sql.DB`。

#### Scenario: 启动时建立连接
- **WHEN** 应用程序启动初始化数据库连接
- **THEN** 使用 `go-sql-driver/mysql` 驱动通过 `sql.Open()` 建立连接
- **AND** 配置连接池参数（`MaxIdleConns`、`MaxOpenConns`、`ConnMaxLifetime`）
- **AND** 返回 `*sql.DB` 供 infrastructure 层使用

#### Scenario: sqlc 依赖完全移除
- **WHEN** 项目完成迁移
- **THEN** `go.mod` 中不包含 `sqlc` 和 `go-sql-driver/mysql`
- **AND** 项目中无任何文件导入 `sqlc`
