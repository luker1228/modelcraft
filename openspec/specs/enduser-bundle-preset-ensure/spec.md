## ADDED Requirements

### Requirement: 权限包绑定模型预设时自动确保 PRESET 权限点存在

系统 SHALL 提供"绑定模型预设到权限包"的后端能力。
在绑定时，系统 MUST 执行 `EnsurePresetPermission(modelId, preset)`：
- 若对应 PRESET 权限点已存在，则复用其 `permission_id`
- 若不存在，则按当前模型上下文创建后再绑定

该能力 MUST 幂等：相同 `bundleId + modelId + preset` 的重复请求不产生重复 PRESET 权限点，也不产生重复绑定。

#### Scenario: 预设已存在时复用并绑定
- **WHEN** 权限包请求绑定某模型的某个预设
- **WHEN** 该模型下对应 PRESET 权限点已存在
- **THEN** 系统复用已有 `permission_id` 完成绑定
- **THEN** 不创建新的 PRESET 权限点记录

#### Scenario: 预设不存在时自动创建并绑定
- **WHEN** 权限包请求绑定某模型的某个预设
- **WHEN** 该模型下不存在对应 PRESET 权限点
- **THEN** 系统先创建对应 PRESET 权限点
- **THEN** 使用新 `permission_id` 完成绑定

#### Scenario: 重复绑定请求幂等
- **WHEN** 对同一 `bundleId + modelId + preset` 重复提交绑定请求
- **THEN** 系统最多保留一条 PRESET 权限点记录
- **THEN** 权限包中最多保留一条对应绑定关系

### Requirement: 绑定 OWNER 预设时校验模型 owner 字段

当绑定请求指定 `READ_WRITE_OWNER` 或 `READ_ALL_WRITE_OWNER` 时，系统 MUST 校验模型存在 `END_USER_REF` 字段。
字段缺失时系统 SHALL 返回 `PresetRequiresOwnerField` 错误，并拒绝创建/绑定。

#### Scenario: OWNER 预设绑定校验失败
- **WHEN** 模型不存在 `END_USER_REF` 字段
- **WHEN** 权限包请求绑定 `READ_WRITE_OWNER` 预设
- **THEN** 系统返回 `PresetRequiresOwnerField` 错误
- **THEN** 不创建 PRESET 权限点且不写入绑定关系

#### Scenario: OWNER 预设绑定校验通过
- **WHEN** 模型存在 `END_USER_REF` 字段
- **WHEN** 权限包请求绑定 `READ_ALL_WRITE_OWNER` 预设
- **THEN** 系统成功 ensure 对应 PRESET 权限点
- **THEN** 系统成功写入权限包绑定关系

### Requirement: 模型预设可虚拟查看

系统 SHALL 提供模型预设"可用集合"查询能力，返回结果由内置预设目录与模型格式计算得到。
该查询 SHALL 为只读，不以"是否已落库"作为可见前提。

#### Scenario: 无任何 PRESET 落库时仍可查看可用预设
- **WHEN** 模型尚未创建任何 PRESET 权限点
- **WHEN** 管理员查询模型预设可用列表
- **THEN** 系统返回按模型格式计算出的可用预设集合
- **THEN** 查询过程不写入权限点或绑定关系
