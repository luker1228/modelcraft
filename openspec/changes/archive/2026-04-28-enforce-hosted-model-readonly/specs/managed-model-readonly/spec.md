## ADDED Requirements

### Requirement: 模型类型必须由入口流程确定
系统 MUST 在模型生命周期入口确定模型类型并禁止调用方绕过：创建模型 SHALL 固定为 `SELF_BUILT`，导入模型 SHALL 固定为 `MANAGED`。

#### Scenario: 创建模型自动标记为自建
- **WHEN** 用户通过创建流程新建模型
- **THEN** 系统将模型类型保存为 `SELF_BUILT`

#### Scenario: 导入模型自动标记为托管
- **WHEN** 用户通过导入流程注册外部模型
- **THEN** 系统将模型类型保存为 `MANAGED`

### Requirement: 托管模型禁止执行 DDL 变更
系统 MUST 拒绝针对 `MANAGED` 模型的任何 DDL 相关操作。

#### Scenario: 托管模型执行 DDL 被拒绝
- **WHEN** 调用方尝试对 `MANAGED` 模型执行 DDL 变更
- **THEN** 系统拒绝请求并返回 `MANAGED_MODEL_READONLY` 类错误

### Requirement: 托管模型禁止字段增删改
系统 MUST 拒绝针对 `MANAGED` 模型的字段新增、字段删除、字段定义修改。

#### Scenario: 托管模型新增字段被拒绝
- **WHEN** 调用方尝试为 `MANAGED` 模型新增字段
- **THEN** 系统拒绝请求并返回只读限制错误

#### Scenario: 托管模型修改字段被拒绝
- **WHEN** 调用方尝试修改 `MANAGED` 模型字段定义
- **THEN** 系统拒绝请求并返回只读限制错误

#### Scenario: 托管模型删除字段被拒绝
- **WHEN** 调用方尝试删除 `MANAGED` 模型字段
- **THEN** 系统拒绝请求并返回只读限制错误

### Requirement: 托管模型仅允许数据读取
系统 MUST 仅允许对 `MANAGED` 模型执行读取类数据操作，并拒绝新增、更新、删除。

#### Scenario: 托管模型查询数据成功
- **WHEN** 调用方查询 `MANAGED` 模型数据
- **THEN** 系统返回查询结果

#### Scenario: 托管模型写入数据被拒绝
- **WHEN** 调用方尝试新增、更新或删除 `MANAGED` 模型数据
- **THEN** 系统拒绝请求并返回只读限制错误

### Requirement: 自建模型不受托管只读策略影响
系统 MUST 仅对 `MANAGED` 模型应用只读策略，`SELF_BUILT` 模型保留现有可编辑能力。

#### Scenario: 自建模型字段变更可执行
- **WHEN** 调用方对 `SELF_BUILT` 模型执行字段变更
- **THEN** 系统按既有规则继续处理，不因托管策略被拒绝