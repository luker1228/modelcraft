## ADDED Requirements

### Requirement: 系统必须为 develop 与 runtime 提供独立的 record workspace
系统 MUST 为开发者模型编辑入口与终端用户数据入口提供独立的 record workspace 实现，并 MUST 不再通过单一 workspace 上的 `workspaceMode` 分支承载两套场景。

#### Scenario: develop 入口挂载独立 workspace
- **WHEN** 用户访问 `/org/{orgName}/project/{projectSlug}/model-editor`
- **THEN** 系统 MUST 渲染 develop 专用 record workspace，而不是向共享 workspace 传入 `workspaceMode="develop"`

#### Scenario: runtime 入口挂载独立 workspace
- **WHEN** 用户访问 `/end-user/{orgName}/{projectSlug}/data`
- **THEN** 系统 MUST 渲染 runtime 专用 record workspace，而不是向共享 workspace 传入 `workspaceMode="end_user"`

### Requirement: develop workspace 必须独占结构维护能力
系统 MUST 仅在 develop workspace 暴露字段插入、字段废弃/删除、关系维护等模型结构相关能力，runtime workspace MUST 不显示且不得调用这些能力。

#### Scenario: develop workspace 显示结构维护控制
- **WHEN** 开发者在 develop workspace 打开模型数据工作区
- **THEN** 系统 MUST 提供插入字段、字段生命周期维护或关系维护中的相应控制

#### Scenario: runtime workspace 不暴露结构维护控制
- **WHEN** 终端用户在 runtime workspace 查看同一模型
- **THEN** 系统 MUST 不显示字段结构维护与关系维护控制，也 MUST 不发起对应请求

### Requirement: runtime workspace 必须拥有独立的运行时访问边界
系统 MUST 由 runtime workspace 自身决定运行时 record query/mutation 使用的 actor、token、endpoint 与 GraphQL 文档，并 MUST 不再依赖 develop/runtime 模式分支在共享业务组件中切换这些访问方式。

#### Scenario: runtime record 操作不依赖 workspaceMode
- **WHEN** runtime workspace 执行 record 查询、新增、编辑或删除
- **THEN** 系统 MUST 通过 runtime 专用访问边界完成调用，且实现中不再依赖 `workspaceMode` 判断访问路径

#### Scenario: 后续接入 impersonation 不影响 develop workspace
- **WHEN** 后续为 runtime workspace 引入新的 actor 类型（例如 impersonated end-user）
- **THEN** 系统 SHALL 仅需扩展 runtime 的访问边界，而不要求修改 develop workspace 的访问实现

### Requirement: 共享 record primitives 必须与身份解耦
系统 MUST 将共享层限制为与身份无关的 primitives；任何仍被多个 workspace 复用的 schema utility、模板或 widgets MUST 不直接创建 Apollo client、读取 auth store 或自行推导 develop/end-user 身份。

#### Scenario: 共享原语不直接创建 client
- **WHEN** 一个 schema utility、模板或 widget 被定义为 shared primitive
- **THEN** 该实现 MUST 不直接 import `createModelRuntimeClient`、`createEndUserScopedClient` 或 end-user / developer auth store

#### Scenario: 共享远程 widget 通过注入边界访问数据
- **WHEN** 一个被两个 workspace 共用的 widget 需要远程查询数据
- **THEN** 它 MUST 通过 workspace 注入的 access adapter 或等价边界完成访问，而不是在组件内部决定访问身份与端点
