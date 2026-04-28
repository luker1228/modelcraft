## Why

管理员在为模型配置终端用户权限时，需要逐条手动创建权限点（SELECT/INSERT/UPDATE/DELETE × ALL/SELF），操作繁琐且易出错。预设策略功能允许管理员一键应用标准权限组合，快速为模型建立典型访问控制基线，再按需微调自定义权限点。

## What Changes

- 新增后端 `applyEndUserPresetPolicy` GraphQL mutation，接受 `modelId + preset` 参数
- 扩展 `EndUserPermission` GraphQL 类型，增加 `preset` 字段（nullable，标识权限点来源）
- 数据库 `end_user_permissions` 表新增 `preset` 列（ENUM，nullable）
- 应用预设时，**仅删除**该模型下 `preset IS NOT NULL` 的旧权限点，`preset = null`（手动创建）的权限点保持不变
- 支持 4 种预设：`READ_WRITE_ALL`、`READ_ALL`、`READ_WRITE_OWNER`、`READ_ALL_WRITE_OWNER`
- 依赖 `END_USER_REF` 字段的预设（`*_OWNER`）在模型缺少 owner 字段时返回结构化错误 `PresetRequiresOwnerField`

## Capabilities

### New Capabilities

- `enduser-preset-policy`: 管理员对指定模型一键应用预设权限策略，替换已有预设权限点，保留自定义权限点

### Modified Capabilities

（无现有 spec 级行为变更）

## Impact

- **后端 GraphQL Schema**：`api/graph/project/schema/rbac.graphql`（已完成 schema 定义）
- **数据库**：`db/schema/mysql/14_rbac_preset.sql`（已完成迁移）
- **后端实现链路**：SQL 查询 → sqlc 生成 → Domain → Infrastructure → App → Adapter → Resolver
- **前端**：permissions tab 每个 ModelCard 顶部新增预设策略入口（UI 待实现）
