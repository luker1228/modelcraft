## Requirements

### Requirement: Bundle shall bind data permission by model-scoped item
系统 SHALL 以 data permission item 作为 bundle 内数据权限的正式绑定单元。每个 item 必须绑定一个 `modelId`，并且只能以两种来源之一存在：`PRESET` 或 `CUSTOM`。

#### Scenario: 绑定 preset item
- **WHEN** 管理员向 bundle 绑定一个 preset 数据权限
- **THEN** 请求必须同时提供 `bundleId`、`modelId` 和 `preset`
- **THEN** 系统创建或替换一个 `grantType = PRESET` 的 data permission item
- **THEN** item 不引用任何 custom permission 实体

#### Scenario: 绑定 custom item
- **WHEN** 管理员向 bundle 绑定一个自定义数据权限
- **THEN** 请求必须同时提供 `bundleId`、`modelId` 和 `customPermissionId`
- **THEN** 系统创建或替换一个 `grantType = CUSTOM` 的 data permission item
- **THEN** item 不保存任何 preset 值

### Requirement: A bundle shall contain at most one data permission item per model
系统 SHALL 保证同一 bundle 下同一 `modelId` 最多只有一个 data permission item。

#### Scenario: 同模型重复绑定 preset 时替换旧 item
- **WHEN** bundle B 已存在 model M 的 PRESET item
- **WHEN** 管理员再次为 bundle B 的 model M 绑定另一个 preset
- **THEN** 系统 SHALL 替换原 item，而不是新增第二条 item
- **THEN** bundle B 在 model M 下仍然只有一个 item

#### Scenario: 预设与自定义互斥
- **WHEN** bundle B 已存在 model M 的 CUSTOM item
- **WHEN** 管理员为同一个 bundle B 和 model M 绑定 preset
- **THEN** 系统 SHALL 用新的 PRESET item 替换旧的 CUSTOM item
- **THEN** bundle B 在 model M 下不存在并存的 preset/custom 两条配置

### Requirement: Bundle detail shall expose item-centric data permission view
系统 SHALL 在 bundle 详情查询中返回 item 视角的数据权限列表，而不是直接返回 permission 实体列表。

#### Scenario: 查看 preset item
- **WHEN** bundle 中某模型绑定的是 preset item
- **THEN** 返回条目中包含 `modelId`、`grantType = PRESET`、`preset` 以及最终策略摘要
- **THEN** 返回条目中不包含 `customPermissionId`

#### Scenario: 查看 custom item
- **WHEN** bundle 中某模型绑定的是 custom item
- **THEN** 返回条目中包含 `modelId`、`grantType = CUSTOM`、`customPermission` 摘要以及最终策略摘要
- **THEN** 返回条目中不包含 `preset`
