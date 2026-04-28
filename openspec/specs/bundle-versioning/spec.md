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
