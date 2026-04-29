## 1. 前端 GraphQL Client 扩展

- [x] 1.1 扩展 `GET_END_USER_BUNDLE` query：在 selection set 中补充 `currentVersion`、`snapshots { version createdAt createdBy restoredFrom permissions { sortOrder permissionId } }`
- [x] 1.2 新增 `RESTORE_END_USER_BUNDLE` mutation：调用 `restoreEndUserPermissionBundle(input: { bundleId, version })`，返回 `bundle { id currentVersion snapshots { ... } permissions { ... } }` 和 `error`

## 2. useBundleManage Hook 扩展

- [x] 2.1 在 `useBundleManage` 中导入并注册 `RESTORE_END_USER_BUNDLE` mutation（使用项目级 Apollo client，refetch `GET_END_USER_BUNDLE`）
- [x] 2.2 新增 `restoreBundle(version: number): Promise<MutationResult>` 方法并在返回值中暴露

## 3. VersionHistoryPanel 改为真实数据

- [x] 3.1 删除 `MOCK_VERSIONS` 常量和 `BundleVersion` interface
- [x] 3.2 修改 `VersionHistoryPanel` props：接收 `snapshots: EndUserPermissionBundleSnapshot[]` 和 `currentVersion: number`（替代原 `versions: BundleVersion[]` 和 `currentVersionNo`）
- [x] 3.3 更新面板内渲染逻辑：`changeSummary` 改为「共 N 个权限点」（取 `snapshot.permissions.length`），`createdBy` 直接展示字符串，`restoredFrom` 非空时显示「由 vX 还原」标注
- [x] 3.4 「还原」按钮调用 `restoreBundle(snapshot.version)` 替代 `setTimeout` mock，成功/失败分别 toast 提示
- [x] 3.5 `snapshots` 为空时显示「暂无历史版本」空状态

## 4. 权限包详情页主体清理

- [x] 4.1 删除 `MOCK_BUNDLED` 常量
- [x] 4.2 修改 `permissions` 派生逻辑：直接使用 `bundle?.permissions ?? []`，不做 mock fallback；`bundle.permissions` 为空时显示「暂无权限点」空状态（含「添加权限点」按钮）
- [x] 4.3 将 `bundle.currentVersion` 和 `bundle.snapshots` 传入 `VersionHistoryPanel`

## 5. AddStrategyDialog 清理

- [x] 5.1 删除 `MOCK_DATABASES`、`MOCK_EXTRA_PERMISSIONS`、`MODEL_DB_MAP` 常量
- [x] 5.2 移除 Dialog 顶部数据库 `Select` 控件及相关 `selectedDb` state
- [x] 5.3 移除基于 `selectedDb` 的权限点过滤逻辑，可选权限点直接使用 `allPermissions`（过滤掉已在 bundle 中的）
- [x] 5.4 `allPermissions` 为空或全已添加时显示「暂无可添加权限点」空状态
