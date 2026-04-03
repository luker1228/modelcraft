## 1. 基础设施准备

- [x] 1.1 安装 sqlc CLI 工具，添加 `github.com/go-sql-driver/mysql` 到 `go.mod`
- [x] 1.2 创建 `sqlc.yaml`：配置 engine=mysql、null_style=option、emit_interface=true、输出到 `internal/infrastructure/dbgen/`
- [x] 1.3 创建 `db/queries/` 目录，按领域拆分查询文件：`project.sql`、`model.sql`、`field.sql`、`relation.sql`、`enum.sql`、`cluster.sql`、`org.sql`、`user.sql`、`casbin.sql`
- [x] 1.4 将所有现有 Repository 的查询逻辑翻译为命名 SQL（`:one`/`:many`/`:exec`/`:execresult`），含 nullable trick 的可选过滤条件
- [x] 1.5 运行 `sqlc generate`，验证 `internal/infrastructure/dbgen/` 生成正确（`querier.go`、`models.go`、`db.go`）
- [x] 1.6 实现 `internal/infrastructure/repository/sql_connection.go`：基于 `sql.Open()` + `go-sql-driver/mysql`，返回 `*sql.DB`，配置连接池参数
- [x] 1.7 实现 `internal/infrastructure/repository/sql_error_analyzer.go`：移植 `gorm_error_analyzer` 逻辑，去掉 `sql.ErrNoRows` 依赖，改为检测 `sql.ErrNoRows`
- [x] 1.8 实现 `internal/infrastructure/repository/tx_manager.go`：`TxManager` 接口 + `SqlTxManager` 实现，显式传递 `dbgen.Querier`
- [x] 1.9 运行 `task build` 验证新基础设施编译通过

## 2. project_repository 迁移

- [x] 2.1 为 `ProjectPO` 的 `toDomain()`/`toRow()` 编写纯函数单元测试
- [x] 2.2 实现新 `project_repository.go`（基于 `dbgen.Queries`），删除 `ProjectPO` 中的 sqlc tags
- [x] 2.3 更新 `cmd/server/main.go` 中 project repository 的初始化，从 `*sql.DB` 切换为 `*sql.DB`
- [ ] 2.4 运行 `task auto-test` 验证 project 相关测试通过

## 3. modeldesign_repository 迁移

- [x] 3.1 为 `ModelMetaPO` 的 `toDomain()`/`toRow()` 编写纯函数单元测试（含 JSON 字段 marshal/unmarshal）
- [x] 3.2 实现新 `modeldesign_repository.go`（含 nullable trick 的 Query/Count 方法）
- [x] 3.3 为 `FieldDefinitionPO` 的 `toDomain()`/`toRow()` 编写纯函数单元测试
- [x] 3.4 实现新 `field_definition_repository.go`
- [x] 3.5 为 `ModelRelationPO` 的 `toDomain()`/`toRow()` 编写纯函数单元测试（含 JSON 字段 source_fields/target_fields）
- [x] 3.6 实现新 `model_relation_repository.go`（原 sql_model.go 中的 relation 部分）
- [x] 3.7 更新 `internal/app/modeldesign/model_app.go`：将 `s.db *sql.DB` 替换为 `s.txManager TxManager`，更新事务调用方式
- [ ] 3.8 运行 `task auto-test` 验证 modeldesign 相关测试通过

## 4. 其余 Repository 迁移

- [x] 4.1 迁移 `enum_definition_repository.go` + `field_enum_association_repository.go`，编写 toDomain 单元测试
- [x] 4.2 迁移 `model_group_repository.go`，编写 toDomain 单元测试
- [x] 4.3 迁移 `database_cluster_repository.go`，编写 toDomain 单元测试
- [x] 4.4 迁移 `modelruntime_repository.go`，编写 toDomain 单元测试
- [x] 4.5 迁移 `organization_repository.go` + `user_repository.go` + `membership_repository.go`，编写 toDomain 单元测试
- [x] 4.6 更新相关 app service（organization、project）：替换 `*sql.DB` 字段为 `TxManager`
- [ ] 4.7 运行 `task auto-test` 验证上述模块测试通过

## 5. casbin_repository 迁移

- [x] 5.1 确认 casbin adapter 与 sqlc/database/sql 的兼容性（open question 验证）
- [x] 5.2 迁移 `casbin_role_repository.go`、`casbin_permission_repository.go`、`casbin_user_role_repository.go`
- [ ] 5.3 运行 `task auto-test` 验证权限相关测试通过

## 6. 清理与最终验证

- [ ] 6.1 删除 `internal/infrastructure/repository/base.go`（GormBaseRepository）
- [ ] 6.2 删除 `internal/infrastructure/repository/sql_model.go`
- [ ] 6.3 删除 `internal/infrastructure/repository/gorm_error_analyzer.go`
- [ ] 6.4 删除 `internal/infrastructure/repository/db_connection.go`
- [ ] 6.5 从 `go.mod` 移除 `sqlc` 和 `go-sql-driver/mysql`，运行 `go mod tidy`
- [ ] 6.6 运行 `task check-all`（fmt + lint + vet + test）全量验证
- [ ] 6.7 运行 `task build-prod` 确认生产构建通过
- [ ] 6.8 运行 `task auto-test` 最终集成测试全量验证
