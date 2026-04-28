## Why

当前 `applyEndUserPresetPolicy` 采用“先删除模型下全部 PRESET 再重建”的实现，语义更像重置而非同步。该方式会导致权限点 ID 抖动，并可能通过 FK CASCADE 造成权限包关联静默缩减；同时也不利于后续新增内置预设时的平滑升级。

## What Changes

- 将 `applyEndUserPresetPolicy` 语义调整为“按模型做预设集合差异同步（reconcile）”，不再做全量删除。
- 引入预设集合 diff 执行策略：`toCreate`、`toUpdate`、`toDelete`，并在事务内提交。
- 对已存在预设优先原地更新（保持 permission_id 稳定），仅对废弃或不再适配模型的预设执行删除。
- 支持“预设可计算”的虚拟查看：根据内置预设目录与模型格式计算可用预设，不要求查看即落库。
- 新增“权限包绑定时自动 ensure 预设权限”能力：绑定模型策略时，若对应预设权限不存在则自动创建，存在则复用。

## Capabilities

### New Capabilities
- `enduser-bundle-preset-ensure`: 在权限包绑定模型预设时，后端自动确保对应 PRESET 权限点存在（幂等复用/创建）。

### Modified Capabilities
- `enduser-preset-policy`: `applyEndUserPresetPolicy` 从“删旧建新”改为“差异同步模型预设集合”，并明确 ID 稳定性与删除边界。

## Impact

- GraphQL：`api/graph/project/schema/rbac.graphql`（apply 语义与绑定入口相关输入/返回可能调整）
- App：`internal/app/rbac/permission_app.go`（reconcile 主流程）、`internal/app/rbac/bundle_app.go`（绑定时 ensure）
- Repository/SQL：`db/queries/rbac/permission.sql`、`internal/domain/rbac/repository.go`、`internal/infrastructure/repository/sql_end_user_permission_repository.go`
- 测试：`internal/app/rbac/*`、`internal/domain/rbac/*` 的 apply/reconcile/幂等/并发与回归用例
