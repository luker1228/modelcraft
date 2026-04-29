## ADDED Requirements

### Requirement: 前端版本历史面板消费真实快照数据

版本历史面板 SHALL 通过 `GET_END_USER_BUNDLE` query 的 `snapshots` 字段获取快照列表，通过 `currentVersion` 字段标识当前版本，不得使用任何硬编码快照数据。

#### Scenario: 权限包有快照记录时展示历史列表
- **WHEN** `bundle.snapshots` 包含至少一个快照
- **THEN** 面板按 version DESC 排列展示快照列表，每条显示版本号、时间、操作人、权限点数量；当前版本（`bundle.currentVersion`）标注「当前」标记

#### Scenario: 权限包无快照时显示空提示
- **WHEN** `bundle.snapshots` 为空数组
- **THEN** 面板显示「暂无历史版本」的空状态，不显示任何 mock 快照

---

### Requirement: 前端「还原」按钮调用真实 restoreEndUserPermissionBundle mutation

版本历史面板的「还原到此版本」按钮 SHALL 调用 `restoreEndUserPermissionBundle` mutation（通过 `useBundleManage` 暴露的 `restoreBundle(version: Int)` 方法），操作成功后刷新权限包数据（refetch `GET_END_USER_BUNDLE`），不得使用 `setTimeout` 模拟。

#### Scenario: 用户点击还原按钮后成功回滚
- **WHEN** 用户点击某历史版本的「还原到此版本」按钮
- **THEN** 按钮进入 loading 态，mutation 调用 `restoreEndUserPermissionBundle(input: { bundleId, version })`，成功后 toast 提示「已还原到版本 vN」，权限包数据刷新

#### Scenario: 还原操作失败时显示错误
- **WHEN** `restoreEndUserPermissionBundle` 返回 error
- **THEN** toast 显示错误信息，权限包数据不变

---

### Requirement: GET_END_USER_BUNDLE query 包含快照字段

`graphql-docs.ts` 中的 `GET_END_USER_BUNDLE` query SHALL 在 selection set 中包含 `currentVersion` 和 `snapshots`（含 `version`、`createdAt`、`createdBy`、`restoredFrom`、`permissions { sortOrder permissionId }`）。

#### Scenario: query 返回带快照的权限包数据
- **WHEN** 调用 `GET_END_USER_BUNDLE($id: ID!)`
- **THEN** 响应包含 `currentVersion: Int` 和 `snapshots: [{version, createdAt, createdBy, restoredFrom, permissions}]` 字段
