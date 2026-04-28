## MODIFIED Requirements

### Requirement: 管理员可对模型应用预设权限策略

系统 SHALL 提供模型级 `applyEndUserPresetPolicy`（或等价 reconcile 接口），用于将模型的 PRESET 权限点同步到目标集合，而非全量删除后重建。
目标集合 SHALL 由内置预设目录与模型格式计算得到。

同步时系统 SHALL 执行差异计算：
- `toCreate = desired - existing`
- `toUpdate = desired ∩ existing`（当行策略有差异）
- `toDelete = existing - desired`

系统 SHALL 在单事务内完成一次 apply 的差异操作。

#### Scenario: 模型预设差异同步成功
- **WHEN** 管理员对模型触发一次 apply/reconcile
- **THEN** 系统按 `toCreate/toUpdate/toDelete` 执行差异同步
- **THEN** 系统不得执行“按模型全量删除 PRESET 再重建”的流程
- **THEN** 同步完成后返回模型下最新权限点集合

#### Scenario: 新增内置预设后自动补齐
- **WHEN** 后端新增一个内置预设，且该模型满足适配条件
- **WHEN** 管理员触发 apply/reconcile
- **THEN** 系统将该预设识别为 `toCreate` 并创建对应 PRESET 权限点
- **THEN** 已存在且未变化的预设权限点不得被删除重建

#### Scenario: 自定义权限点不受影响
- **WHEN** 模型同时存在 `type=CUSTOM` 与 `type=PRESET` 权限点
- **WHEN** 管理员触发 apply/reconcile
- **THEN** 系统仅处理 PRESET 差异
- **THEN** 所有 CUSTOM 权限点保持不变

### Requirement: 依赖 END_USER_REF 字段的预设在字段缺失时返回结构化错误

系统 SHALL 在“显式指定 OWNER 预设”的场景下校验模型是否存在 `END_USER_REF` 字段。
对于模型级 reconcile（未显式指定某个 OWNER 预设）场景，系统 SHALL 仅同步可适配预设集合，不将 OWNER 预设纳入 `desired`。

#### Scenario: 模型级 reconcile 遇到无 owner 字段
- **WHEN** 模型不存在 `END_USER_REF` 字段
- **WHEN** 管理员触发模型级 apply/reconcile
- **THEN** 系统仅同步非 OWNER 预设
- **THEN** 本次流程不因 OWNER 预设不适配而失败

#### Scenario: 显式应用 OWNER 预设且字段缺失
- **WHEN** 模型不存在 `END_USER_REF` 字段
- **WHEN** 调用方显式请求应用 `READ_WRITE_OWNER` 或 `READ_ALL_WRITE_OWNER`
- **THEN** 系统返回 `PresetRequiresOwnerField` 结构化错误
- **THEN** 不创建或更新对应 OWNER 预设权限点

### Requirement: 预设权限点携带 preset 字段标识来源

`EndUserPermission` 类型 SHALL 包含 `preset` 字段（nullable）用于标识来源。
- `preset = null` 表示手动创建权限点
- `preset != null` 表示 PRESET 来源权限点

#### Scenario: 查询权限点返回来源字段
- **WHEN** 查询模型权限点列表
- **THEN** 每个权限点均返回 `preset` 字段
- **THEN** PRESET 权限点返回对应枚举值，CUSTOM 返回 null

## ADDED Requirements

### Requirement: 预设权限点在差异同步中保持引用稳定

对 `toUpdate` 集合中的预设权限点，系统 MUST 原地更新并保持 `permission_id` 稳定。
对 `toDelete` 集合中的预设权限点，若仍被权限包引用，系统 MUST 以结构化错误阻断删除并回滚本次同步。

#### Scenario: 已存在预设策略变更时原地更新
- **WHEN** 模型中某预设权限点已存在，且新策略定义与现存 `row_policy` 不一致
- **THEN** 系统更新该记录内容而不更换 `permission_id`
- **THEN** 已绑定该权限点的权限包关联保持有效

#### Scenario: 待删除预设仍被权限包引用
- **WHEN** 某 PRESET 落入 `toDelete`
- **WHEN** 该 PRESET 权限点仍存在权限包引用
- **THEN** 系统返回包含引用信息的结构化错误
- **THEN** 本次 apply/reconcile 事务整体回滚
