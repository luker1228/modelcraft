## MODIFIED Requirements

### Requirement: 数据权限 item 变更自动创建快照
执行数据权限 item 的绑定、替换或移除成功后，系统 SHALL 在同一事务内为该权限包创建一条快照记录，记录当前完整的 data permission item 列表。
快照记录 SHALL 保存每个 item 的 `modelId`、`grantType`、`preset`、`customPermissionId` 与 `sortOrder`。

#### Scenario: 绑定 preset item 后自动生成快照
- **WHEN** 管理员为 bundle 绑定一个新的 PRESET item
- **THEN** 系统新增一条快照记录
- **THEN** 快照中的 item 列表包含该模型对应的 `preset` 绑定信息

#### Scenario: 替换 custom item 后自动生成快照
- **WHEN** 管理员将某模型在 bundle 下的 CUSTOM item 替换为新的 CUSTOM item
- **THEN** 系统新增一条快照记录
- **THEN** 快照中仅保留替换后的唯一 item

#### Scenario: 修改权限包名称不触发快照
- **WHEN** 管理员仅修改权限包名称或描述
- **THEN** 快照表记录数量不变
- **THEN** 当前版本号不变

### Requirement: 查询权限包时可获取 item 级历史快照列表
`EndUserPermissionBundle` GraphQL 类型 SHALL 返回最近快照列表，并且每条快照的条目 SHALL 以 data permission item 结构表达，而不是以 `permissionId` 列表表达。

#### Scenario: 查询含 preset item 的快照
- **WHEN** 某历史版本中的模型绑定为 PRESET item
- **THEN** 对应快照条目返回 `modelId`、`grantType = PRESET`、`preset` 与 `sortOrder`

#### Scenario: 查询含 custom item 的快照
- **WHEN** 某历史版本中的模型绑定为 CUSTOM item
- **THEN** 对应快照条目返回 `modelId`、`grantType = CUSTOM`、`customPermissionId` 与 `sortOrder`
