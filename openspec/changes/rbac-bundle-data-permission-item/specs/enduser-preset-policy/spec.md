## MODIFIED Requirements

### Requirement: 管理员可按模型选择预设模板进行授权
系统 SHALL 保留 `READ_WRITE_ALL`、`READ_ALL`、`READ_WRITE_OWNER`、`READ_ALL_WRITE_OWNER` 这些预设模板定义。
预设模板本身 SHALL 不落库为 `permission` 实体。
管理员在授权预设时，系统 SHALL 要求显式选择目标模型，并在 bundle item 中记录 `modelId + preset`。

#### Scenario: 绑定 READ_ALL 预设时必须指定模型
- **WHEN** 管理员为 bundle 绑定 `READ_ALL` 预设
- **THEN** 请求必须包含目标 `modelId`
- **THEN** 系统记录一个 `grantType = PRESET` 的 bundle item，而不是创建新的 permission 实体

#### Scenario: 查看模型可选预设时无需预先落库
- **WHEN** 管理员打开某模型的数据权限配置界面
- **THEN** 系统可以直接返回该模型可用的 preset 模板列表
- **THEN** 这些模板不要求事先存在于 `end_user_data_permissions` 表中

### Requirement: 依赖 END_USER_REF 字段的预设在绑定时校验模型上下文
当绑定 `READ_WRITE_OWNER` 或 `READ_ALL_WRITE_OWNER` 预设时，系统 SHALL 校验目标模型是否存在 `END_USER_REF` 类型字段。
字段缺失时，系统 SHALL 返回 `PresetRequiresOwnerField` 错误，不写入或替换任何 bundle item。

#### Scenario: 模型缺少 owner 字段时绑定 OWNER 预设失败
- **WHEN** 模型不存在 END_USER_REF 类型字段
- **WHEN** 管理员尝试为该模型绑定 `READ_WRITE_OWNER` 预设
- **THEN** 系统返回 `PresetRequiresOwnerField` 错误
- **THEN** 原有 bundle item SHALL 保持不变

#### Scenario: 模型具备 owner 字段时成功绑定 OWNER 预设
- **WHEN** 模型存在 END_USER_REF 类型字段
- **WHEN** 管理员为该模型绑定 `READ_ALL_WRITE_OWNER` 预设
- **THEN** 系统成功写入或替换对应的 PRESET item
- **THEN** 鉴权阶段按该模型上下文展开 owner 相关 row policy

### Requirement: 自定义权限可与预设在语义上重合
系统 SHALL 允许管理员创建与某个 preset 展开结果语义相同的 custom permission。
系统 MUST NOT 因为策略等价而阻止 custom permission 的创建或绑定。

#### Scenario: 自定义策略与 READ_ALL 等价时仍可保存
- **WHEN** 管理员创建一个 row/column policy 与 `READ_ALL` 语义完全一致的 custom permission
- **THEN** 系统允许该 custom permission 保存成功
- **THEN** 后续 bundle 仍可选择绑定该 custom permission，而不是强制改用 preset
