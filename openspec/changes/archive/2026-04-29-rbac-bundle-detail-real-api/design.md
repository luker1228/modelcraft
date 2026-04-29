## Context

权限包详情页（`/roles/bundles/:id`）已有完整的 UI 骨架，但存在 4 处硬编码 mock 数据：
1. **`MOCK_VERSIONS`**：版本历史面板使用静态数组，还原操作调用 `setTimeout` 模拟
2. **`MOCK_BUNDLED`**：已配置权限点列表在 bundle.permissions 为空时回退到 mock
3. **`MOCK_EXTRA_PERMISSIONS`**：「添加策略」Dialog 的可选权限点在 API 返回为空时回退到 mock
4. **`MOCK_DATABASES`**：Dialog 顶部数据库下拉使用静态列表

后端现状：
- `EndUserPermissionBundle` 类型已内嵌 `currentVersion: Int!` 和 `snapshots: [EndUserPermissionBundleSnapshot!]!`，快照数据通过 `GET_END_USER_BUNDLE` 随权限包一起返回，无需单独 query
- `restoreEndUserPermissionBundle` mutation 已在 schema 定义，后端实现已完成
- 权限点数据由 `GET_END_USER_PERMISSIONS` 提供，`useBundleManage` 已在调用，结果存于 `allPermissions`
- 数据库集群只有 `databaseCluster`（单条）query，权限点模型不含 clusterId 字段，数据库维度筛选属于未来功能

## Goals / Non-Goals

**Goals:**
- 移除 `MOCK_VERSIONS`：版本历史面板改用 `bundle.snapshots` + `bundle.currentVersion` 渲染
- 添加 `restoreEndUserPermissionBundle` mutation 到 `graphql-docs.ts`，`VersionHistoryPanel` 的「还原」按钮调用真实接口
- 移除 `MOCK_BUNDLED`：已配置权限点直接使用 `bundle.permissions`，不再 fallback
- 移除 `MOCK_EXTRA_PERMISSIONS`：Dialog 中的可选权限点使用 `allPermissions`，空时显示空状态
- 移除 `MOCK_DATABASES` 和 `MODEL_DB_MAP`：Dialog 顶部数据库下拉区域改为纯 UI 隐藏（该功能依赖未来后端扩展）
- 扩展 `GET_END_USER_BUNDLE` query 的 selection set，补充 `currentVersion`、`snapshots`（含 `version`、`createdAt`、`createdBy`、`restoredFrom`、`permissions { sortOrder permissionId permission { id displayName } }`）
- 在 `useBundleManage` 中暴露 `restoreBundle` 方法

**Non-Goals:**
- 数据库维度筛选功能（权限点模型不含 clusterId，后端无对应接口）
- 版本快照的 `changeSummary` 文字描述（后端 `EndUserPermissionBundleSnapshot` 无此字段，改为展示权限点数量变化）
- 分页加载权限点（当前数量级不需要）
- 权限点的 `modelDisplayName` 字段（后端 `EndUserPermission` 无此字段，改用 `modelId` 展示）

## Decisions

### Decision 1：快照数据通过 `GET_END_USER_BUNDLE` 内嵌返回，不新增独立 query

后端 `EndUserPermissionBundle` 已内嵌 `snapshots` 字段，随权限包详情一起返回。独立 query 会增加网络请求和缓存复杂度，内嵌方式成本最低，满足当前「最多 5 个快照」的数量限制。

**备选：** 新增 `getBundleVersionHistory(bundleId: ID!)` query — 被否决，schema 未定义，无必要。

### Decision 2：`VersionHistoryPanel` 不再显示 `changeSummary`，改为展示权限点数量

后端快照无 `changeSummary` 字段。用「共 N 个权限点」替代，语义清晰且不需要后端改动。

**备选：** 后端新增 changeSummary 字段 — 工作量大，当前不值得。

### Decision 3：数据库下拉改为「敬请期待」占位或直接隐藏，不保留 mock

保留 mock 会误导用户认为数据库筛选已上线。直接移除下拉，Dialog 布局退化为「左侧 model 列表 + 右侧权限选择」两列，等后端扩展 clusterId 后再还原。

### Decision 4：`MOCK_BUNDLED` 直接移除，不保留 fallback

`bundle.permissions` 为空表示权限包真实为空，显示空状态更准确，保留 mock 会掩盖真实状态。

### Decision 5：`graphql-docs.ts` 补充 `RESTORE_END_USER_BUNDLE` mutation

`restoreEndUserPermissionBundle` 在后端已实现，前端只需添加 gql 文档并在 `useBundleManage` 中暴露 `restoreBundle(version: Int)` 方法。

## Risks / Trade-offs

- **`changeSummary` 缺失影响可读性** → 用权限点数量和时间戳替代，可读性略降但准确
- **`modelDisplayName` 缺失** → 用 `modelId` 展示（英文标识），后续可由前端维护 model 名称映射
- **`createdBy` 为 `String` 非 User 对象** → 直接渲染字符串，无法链接到用户详情页，可接受

## Migration Plan

1. 扩展 `GET_END_USER_BUNDLE` query selection set（在 graphql-docs.ts）
2. 添加 `RESTORE_END_USER_BUNDLE` mutation（在 graphql-docs.ts）
3. 扩展 `useBundleManage` 暴露 `restoreBundle` 方法
4. 修改 `page.tsx`：移除 4 个 mock 常量，`VersionHistoryPanel` 改用 `bundle.snapshots`，「还原」调用 `restoreBundle`，权限点列表移除 fallback，`AddStrategyDialog` 移除数据库下拉

无数据库迁移，无 API 破坏性变更，可随时回滚（删除改动即可）。
