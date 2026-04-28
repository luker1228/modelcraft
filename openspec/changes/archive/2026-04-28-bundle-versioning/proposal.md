## Why

权限包（`EndUserPermissionBundle`）当前不保留历史，管理员每次修改权限点列表后，变更立即生效且不可追溯。误操作无法回滚，也无法审计历史状态。在生产环境中，一次误删权限点可能导致终端用户无法访问关键数据，而且完全没有恢复手段。

## What Changes

- **新增** `end_user_permission_bundle_snapshots` 数据库表，存储权限包的历史快照
- **新增** `BundleSnapshot` 领域值对象及 Repository 接口方法
- **新增** `restoreEndUserPermissionBundle` GraphQL Mutation，支持一键回滚
- **修改** `addEndUserPermissionToBundle` / `removeEndUserPermissionFromBundle` Use Case，操作成功后自动写入快照
- **修改** `EndUserPermissionBundle` GraphQL 类型，新增 `currentVersion` 和 `snapshots` 字段
- 快照滚动保留：每个权限包最多保留最近 **5 个**历史版本，超出时自动删除最旧的

## Capabilities

### New Capabilities

- `bundle-versioning`：权限包版本快照——每次权限列表变更时自动创建快照，支持查看历史版本和一键回滚

### Modified Capabilities

（无现有 spec 需要变更）

## Impact

- **数据库**：新增 `end_user_permission_bundle_snapshots` 表，需 Atlas 迁移
- **GraphQL API**：`api/graph/project/schema/permission.graphql` 新增类型和 mutation
- **后端**：`internal/domain/enduserrbac/`、`internal/usecase/enduserrbac/`、`internal/infrastructure/postgres/enduserrbac/`、resolver 层均需修改或新增文件
- **鉴权热路径不受影响**：鉴权仍读取当前关联表，不查快照表
- **前端**：权限包详情页新增版本历史入口（Drawer/Sheet），后端就绪后跟进
