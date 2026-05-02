## Requirements

### Requirement: 权限包详情页展示真实的数据权限 item 列表
权限包详情页 SHALL 直接渲染 bundle 返回的 data permission item 数组，不得把 preset 和 custom 都伪装成统一 permission 实体后再展示。

#### Scenario: 展示 preset item
- **WHEN** bundle 中某条数据权限来源于 preset
- **THEN** 页面展示该模型、preset 名称和策略摘要
- **THEN** 页面不展示不存在的 custom permission 标识

#### Scenario: 展示 custom item
- **WHEN** bundle 中某条数据权限来源于 custom permission
- **THEN** 页面展示该模型、自定义权限名称和策略摘要
- **THEN** 页面不展示 preset 名称

### Requirement: 添加数据权限 Dialog shall distinguish preset and custom binding paths
添加数据权限的 Dialog SHALL 明确区分两条真实写路径：绑定 preset item、绑定 custom item。

#### Scenario: 选择 preset 绑定路径
- **WHEN** 用户在 Dialog 中选择"预设模板"模式
- **THEN** 界面要求先选择模型，再展示该模型可选 preset 模板
- **THEN** 提交时调用 preset item 绑定接口

#### Scenario: 选择 custom 绑定路径
- **WHEN** 用户在 Dialog 中选择"自定义策略"模式
- **THEN** 界面要求先选择模型，再展示该模型下可复用的 custom permission 列表
- **THEN** 提交时调用 custom item 绑定接口

### Requirement: 同模型重复配置时前端应体现替换语义
前端 SHALL 在用户为同一 bundle/model 再次提交数据权限配置时提示这是替换现有 item，而不是叠加新增。

#### Scenario: 模型已有 item 时展示替换提示
- **WHEN** 当前 bundle 已存在 model M 的数据权限 item
- **WHEN** 用户尝试重新为 model M 配置 preset 或 custom
- **THEN** 界面提示"将替换该模型当前配置"
- **THEN** 成功后列表中 model M 仍只显示一条 item
