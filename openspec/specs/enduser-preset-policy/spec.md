## ADDED Requirements

### Requirement: 管理员可对模型应用预设权限策略

系统 SHALL 提供 `applyEndUserPresetPolicy` mutation，允许管理员为指定模型一键应用预设权限策略。
支持的预设类型：`READ_WRITE_ALL`、`READ_ALL`、`READ_WRITE_OWNER`、`READ_ALL_WRITE_OWNER`。
应用预设时，系统 SHALL 删除该模型下所有 `preset IS NOT NULL` 的权限点，再创建新的预设权限点。
`preset = null`（手动创建）的权限点 SHALL 不受任何影响。

#### Scenario: 应用 READ_WRITE_ALL 预设
- **WHEN** 管理员对某模型调用 `applyEndUserPresetPolicy(preset: READ_WRITE_ALL)`
- **THEN** 系统删除该模型下所有 preset 来源的旧权限点
- **THEN** 系统创建 4 个新权限点：SELECT ALL、INSERT ALL、UPDATE ALL、DELETE ALL，均标记 `preset = READ_WRITE_ALL`
- **THEN** 返回该模型下所有权限点（包含新建的预设权限点和原有自定义权限点）

#### Scenario: 应用 READ_ALL 预设
- **WHEN** 管理员对某模型调用 `applyEndUserPresetPolicy(preset: READ_ALL)`
- **THEN** 系统删除该模型下所有 preset 来源的旧权限点
- **THEN** 系统创建 1 个新权限点：SELECT ALL，标记 `preset = READ_ALL`
- **THEN** 返回该模型下所有权限点

#### Scenario: 保留自定义权限点
- **WHEN** 模型已有 1 个手动创建的权限点（preset = null）和 2 个预设权限点（preset != null）
- **WHEN** 管理员应用任意预设
- **THEN** 手动创建的权限点 SHALL 保持不变
- **THEN** 旧的预设权限点被全部替换为新预设展开的权限点

### Requirement: 依赖 END_USER_REF 字段的预设在字段缺失时返回结构化错误

当应用 `READ_WRITE_OWNER` 或 `READ_ALL_WRITE_OWNER` 预设时，系统 SHALL 校验模型是否存在 `END_USER_REF` 类型字段（owner 字段）。
字段缺失时，系统 SHALL 返回 `PresetRequiresOwnerField` 错误，不创建任何权限点。

#### Scenario: 模型缺少 owner 字段时应用 OWNER 预设
- **WHEN** 模型不存在 END_USER_REF 类型字段
- **WHEN** 管理员调用 `applyEndUserPresetPolicy(preset: READ_WRITE_OWNER)`
- **THEN** 系统返回 `PresetRequiresOwnerField` 错误，包含 preset 名称和建议操作
- **THEN** 模型权限点 SHALL 保持不变（不执行删除或创建）

#### Scenario: 模型有 owner 字段时成功应用 OWNER 预设
- **WHEN** 模型存在 END_USER_REF 类型字段
- **WHEN** 管理员调用 `applyEndUserPresetPolicy(preset: READ_WRITE_OWNER)`
- **THEN** 系统删除旧预设权限点
- **THEN** 系统创建 4 个新权限点：SELECT SELF、INSERT SELF、UPDATE SELF、DELETE SELF

### Requirement: 预设权限点携带 preset 字段标识来源

`EndUserPermission` 类型 SHALL 包含 `preset` 字段（nullable），用于区分权限点来源。
- `preset = null` 表示手动创建的自定义权限点
- `preset != null` 表示由预设策略创建的权限点，值为预设类型名称

#### Scenario: 查询权限点返回 preset 字段
- **WHEN** 查询模型下的权限点列表
- **THEN** 每个权限点 SHALL 包含 `preset` 字段
- **THEN** 手动创建的权限点 `preset` 为 null
- **THEN** 预设创建的权限点 `preset` 为对应的枚举值（如 `READ_WRITE_ALL`）
