## Why

权限包详情页（`/roles/bundles/:id`）当前使用硬编码 mock 数据渲染版本历史、可添加权限点列表、数据库下拉、已配置权限等核心区域，无法反映真实系统状态，配置操作也无法持久化。后端 RBAC GraphQL API 已具备完整实现（查询/变更权限包、权限点、版本快照），前端具备完整的 GraphQL client 定义，现在需要打通两端完成真实能力。

## What Changes

- **前端：移除 4 处 mock 常量**（`MOCK_VERSIONS`、`MOCK_EXTRA_PERMISSIONS`、`MOCK_DATABASES`、`MOCK_BUNDLED`），替换为真实 GraphQL 数据查询
- **前端：版本历史面板** 改为调用 `GetBundleVersionHistory` query，展示真实快照列表，「还原到此版本」调用 `RollbackBundleToVersion` mutation
- **前端：权限点列表** 改为调用 `ListProjectEndUserPermissions` query 获取所有可用权限点，已配置项通过 `GetEndUserPermissionBundle` 中的 `permissions` 字段区分
- **前端：「添加策略」Dialog 中的数据库下拉** 改为调用真实集群/数据库列表接口（`ListProjectDatabaseClusters` 或等效接口）
- **前端：添加/移除权限点** 调用 `AddEndUserPermissionToBundle` / `RemoveEndUserPermissionFromBundle` mutation，操作后刷新权限包数据
- **后端：无需新增接口**，现有 GraphQL schema 已覆盖所有需求；确认 `GetBundleVersionHistory` resolver 已正确实现快照查询

## Capabilities

### New Capabilities

- `bundle-detail-real-data`：权限包详情页（含版本历史、权限点列表、添加策略对话框）基于真实 API 数据渲染，支持真实 CRUD 操作

### Modified Capabilities

- `bundle-versioning`：前端消费已实现的版本快照 API（只涉及前端集成，spec 层需求不变）

## Impact

- **前端文件**：`src/app/.../roles/bundles/[bundleId]/page.tsx`（主要改动）、`src/api-client/rbac/graphql-docs.ts`（确认 query 已定义或补充）
- **后端**：预期无需改动；如发现 `GetBundleVersionHistory` 返回为空，需排查 resolver 实现
- **依赖**：前端现有 Apollo Client 配置、`useBundleManage` hook 可复用
