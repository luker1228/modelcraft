## ADDED Requirements

### Requirement: GraphQL 资源不存在错误统一建模
系统 MUST 在 Org、Project、EndUser 三套 GraphQL Schema 中使用统一错误类型 `ResourceNotFound` 表达“资源不存在”语义，不再暴露语义等价的多个 `*NotFound` 类型。

#### Scenario: 查询不存在模型
- **WHEN** 调用方请求不存在的模型资源
- **THEN** 返回 payload.error，且错误类型为 `ResourceNotFound`

#### Scenario: 查询不存在用户资料
- **WHEN** 调用方请求不存在的用户资料
- **THEN** 返回 payload.error，且错误类型为 `ResourceNotFound`

### Requirement: ResourceNotFound 必须提供资源类别
系统 MUST 在 `ResourceNotFound` 中提供 `resourceType` 字段，并使用枚举值标识具体资源类别。

#### Scenario: 模型不存在映射
- **WHEN** 后端业务错误码为 `NOT_FOUND.MODEL`
- **THEN** GraphQL 返回 `ResourceNotFound.resourceType = MODEL`

#### Scenario: 项目不存在映射
- **WHEN** 后端业务错误码为 `NOT_FOUND.PROJECT`
- **THEN** GraphQL 返回 `ResourceNotFound.resourceType = PROJECT`

### Requirement: 错误 Union 必须收敛 NotFound 成员
系统 SHALL 将各 Query/Mutation 错误 union 中表示资源不存在的成员统一为 `ResourceNotFound`，不得继续混用多个 `*NotFound` 成员。

#### Scenario: GetModelError 收敛
- **WHEN** 定义 `GetModelError` union
- **THEN** union 中的 not-found 成员仅允许 `ResourceNotFound`

#### Scenario: RBAC 相关 union 收敛
- **WHEN** 定义 RBAC 相关 mutation 错误 union
- **THEN** 各 union 的 not-found 成员仅允许 `ResourceNotFound`

### Requirement: 细粒度错误码必须继续可观测
系统 MUST 保留 GraphQL `extensions.code` 的细粒度业务错误码（例如 `NOT_FOUND.MODEL`、`NOT_FOUND.PROJECT`），以支持定位与审计。

#### Scenario: 返回统一类型且保留细粒度 code
- **WHEN** 发生 `NOT_FOUND.CLUSTER`
- **THEN** payload.error 类型为 `ResourceNotFound`，且 `extensions.code = NOT_FOUND.CLUSTER`

### Requirement: 调用方必须基于统一类型进行断言
系统 SHALL 要求调用方（前端文档与 BDD）使用 `ResourceNotFound` 与 `resourceType` 进行断言，不再依赖具体 `XxxNotFound` 的 `__typename`。

#### Scenario: 前端 GraphQL 文档断言
- **WHEN** 编写 GraphQL 错误片段
- **THEN** 仅使用 `... on ResourceNotFound { message resourceType }`

#### Scenario: BDD 错误类型断言
- **WHEN** 编写资源不存在场景断言
- **THEN** 断言 `ResourceNotFound` 并校验 `resourceType`