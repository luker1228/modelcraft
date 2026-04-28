## 1. 数据库迁移

- [x] 1.1 在 `db/schema/` 中新增 `end_user_permission_bundle_snapshots` 表定义（含 `UNIQUE (bundle_id, version)` 约束和 `version DESC` 索引）
- [x] 1.2 运行 `just db up` 应用迁移，确认表结构正确

## 2. sqlc 查询与代码生成

- [x] 2.1 在 `db/queries/` 中新增快照相关 sqlc 查询：`InsertBundleSnapshot`、`ListBundleSnapshots`、`DeleteOldBundleSnapshots`、`GetBundleCurrentVersion`
- [x] 2.2 运行 `just generate-sqlc` 生成 Go 查询代码，确认无编译错误

## 3. 领域层

- [x] 3.1 新增 `internal/domain/enduserrbac/bundle_snapshot.go`，定义 `BundleSnapshot` 值对象（`Version`、`Permissions []SnapshotPermissionEntry`、`CreatedAt`、`CreatedBy`、`RestoredFrom *int`）
- [x] 3.2 在 `internal/domain/enduserrbac/bundle.go` 的 `Bundle` 实体上新增 `Snapshots []BundleSnapshot` 字段
- [x] 3.3 在 `internal/domain/enduserrbac/repository.go` 的 Repository 接口中新增方法：`SaveBundleSnapshot`、`ListBundleSnapshots`、`DeleteOldBundleSnapshots`

## 4. Infrastructure 层

- [x] 4.1 在 `internal/infrastructure/postgres/enduserrbac/` 实现新增的三个 Repository 方法（基于 sqlc 生成代码）
- [x] 4.2 编写单元测试验证快照写入和滚动删除逻辑

## 5. Use Case 层：快照写入

- [x] 5.1 修改 `addEndUserPermissionToBundle` Use Case，在权限添加成功后于同一事务内调用 `SaveBundleSnapshot` + `DeleteOldBundleSnapshots`
- [x] 5.2 修改 `removeEndUserPermissionFromBundle` Use Case，同上逻辑

## 6. Use Case 层：回滚

- [x] 6.1 新增 `internal/usecase/enduserrbac/restore_bundle.go`，实现 `RestoreBundleUseCase`：查询目标快照 → 事务内替换当前权限点关联（DELETE + INSERT）→ 写新快照（`restored_from = targetVersion`）→ 触发滚动删除
- [x] 6.2 编写单元测试：回滚成功、目标版本不存在返回错误、回滚后快照参与滚动保留计数

## 7. GraphQL Schema 变更

- [x] 7.1 在 `api/graph/project/schema/permission.graphql` 新增 `EndUserPermissionBundleSnapshot`、`EndUserPermissionSnapshotEntry` 类型
- [x] 7.2 在 `EndUserPermissionBundle` 类型上新增 `currentVersion: Int!` 和 `snapshots` 字段
- [x] 7.3 新增 `RestoreEndUserPermissionBundleInput`、`RestoreEndUserPermissionBundlePayload`、`EndUserPermissionBundleSnapshotNotFound` 类型，以及 `restoreEndUserPermissionBundle` mutation
- [x] 7.4 运行 `just generate-gql` 生成 resolver 接口代码，确认无编译错误

## 8. GraphQL Resolver 实现

- [x] 8.1 实现 `EndUserPermissionBundle.currentVersion` 和 `snapshots` resolver（调用 `ListBundleSnapshots`，懒加载）
- [x] 8.2 实现 `EndUserPermissionBundleSnapshot.permissions` resolver（JOIN 当前权限点，已删除返回 null）
- [x] 8.3 实现 `restoreEndUserPermissionBundle` mutation resolver，调用 `RestoreBundleUseCase`，映射错误类型

## 9. BDD 验收测试

- [x] 9.1 新增 BDD feature 文件，覆盖以下场景：添加权限点后自动生成快照、超出 5 个版本时滚动删除、回滚到历史版本、修改名称不触发快照
- [x] 9.2 运行 BDD 测试套件，确认全部场景通过
