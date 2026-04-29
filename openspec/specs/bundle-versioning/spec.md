## ADDED Requirements

### Requirement: 权限点变更自动创建快照

执行 `addEndUserPermissionToBundle` 或 `removeEndUserPermissionFromBundle` 成功后，系统 SHALL 在同一事务内自动为该权限包创建一条快照记录，记录当前完整的权限点 ID 列表及 sortOrder。

#### Scenario: 添加权限点后自动生成快照
- **WHEN** 管理员调用 `addEndUserPermissionToBundle`，权限包 P 原有 2 个权限点
- **THEN** `end_user_permission_bundle_snapshots` 中新增一条记录，version = 当前最大版本 + 1，`permissions` JSONB 包含 3 个权限点 ID

#### Scenario: 移除权限点后自动生成快照
- **WHEN** 管理员调用 `removeEndUserPermissionFromBundle`，权限包 P 原有 3 个权限点
- **THEN** 快照表中新增一条记录，`permissions` JSONB 包含 2 个权限点 ID

#### Scenario: 修改名称不触发快照
- **WHEN** 管理员调用 `updateEndUserPermissionBundle` 仅修改权限包名称
- **THEN** 快照表记录数量不变，`currentVersion` 不变

---

### Requirement: 快照滚动保留最多 5 个历史版本

每个权限包在快照表中 SHALL 最多保留最近 5 个历史版本。超出时，系统 SHALL 在同一写入事务内删除 `version` 最小的快照记录。

#### Scenario: 超出 5 个版本时删除最旧版本
- **WHEN** 权限包 P 已有 5 条快照（v1-v5），再次修改权限点生成 v6
- **THEN** v1 快照被删除，快照表中只保留 v2-v6（共 5 条）

#### Scenario: 未超出上限时不删除
- **WHEN** 权限包 P 当前有 3 条快照（v1-v3），修改权限点生成 v4
- **THEN** 快照表保留 v1-v4（共 4 条），无删除

---

### Requirement: 查询权限包时可获取历史快照列表

`EndUserPermissionBundle` GraphQL 类型 SHALL 新增 `currentVersion: Int!` 字段（返回当前最大版本号，若无快照则返回 0）和 `snapshots: [EndUserPermissionBundleSnapshot!]!` 字段（返回最近 ≤5 条快照，按 version DESC 排列）。

#### Scenario: 查询有历史快照的权限包
- **WHEN** 客户端查询 `EndUserPermissionBundle.snapshots`，权限包 P 有 3 条快照
- **THEN** 返回 3 条 `EndUserPermissionBundleSnapshot`，按 version 降序排列

#### Scenario: 查询无历史快照的权限包
- **WHEN** 客户端查询 `EndUserPermissionBundle.currentVersion`，权限包 P 从未修改过权限点
- **THEN** 返回 `currentVersion = 0`，`snapshots = []`

#### Scenario: 快照中包含已删除权限点时字段为 null
- **WHEN** 快照 v2 中记录了权限点 A 的 ID，但权限点 A 已被删除
- **THEN** 快照 entry 中 `permission` 字段为 null，`permissionId` 仍返回原始 ID

---

### Requirement: 支持一键回滚到历史快照

系统 SHALL 提供 `restoreEndUserPermissionBundle(input: RestoreEndUserPermissionBundleInput!)` mutation。执行成功后，权限包的当前权限列表恢复为目标快照的状态，并生成新快照记录（`restored_from` 指向目标版本号）。

#### Scenario: 成功回滚到历史版本
- **WHEN** 管理员调用 `restoreEndUserPermissionBundle`，目标版本 v2 包含权限 A、B
- **THEN** 权限包当前权限变为 A、B，新增快照 v_new（`restored_from = 2`），`newVersion` 字段返回 v_new

#### Scenario: 回滚后鉴权反映新权限列表
- **WHEN** 完成回滚后，终端用户请求鉴权
- **THEN** 鉴权查询结果与回滚后的权限列表一致（不受快照表影响）

#### Scenario: 目标版本不存在时返回错误
- **WHEN** 管理员调用 `restoreEndUserPermissionBundle`，指定的 `targetVersion` 在快照表中不存在
- **THEN** 返回 `EndUserPermissionBundleSnapshotNotFound` 错误，权限列表不变

#### Scenario: 回滚生成的新快照参与滚动保留计数
- **WHEN** 权限包已有 5 条快照，执行回滚生成 v6
- **THEN** v1 被删除，保留 v2-v6（回滚版本不享有豁免）

---

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
