## MODIFIED Requirements

### Requirement: End-user runtime 用户查询 MUST 支持租户内受限发现
End-user runtime GraphQL 路由 MUST 支持租户内 user 查询，并 SHALL 提供当前用户访问、单用户唯一查询、受限用户列表查询能力。

#### Scenario: 查询当前用户资料
- **WHEN** 已认证 end-user 调用 `me`
- **THEN** 服务端返回该用户在当前租户范围内的资料

#### Scenario: 按唯一条件查询单个用户
- **WHEN** 客户端使用合法唯一条件调用 `findOne`
- **THEN** 服务端返回至多一条租户内用户记录，并附带 runtime 元信息

#### Scenario: 使用受限条件查询用户列表
- **WHEN** 客户端使用合法受限过滤与分页参数调用 `findMany`
- **THEN** 服务端返回租户内用户列表，并附带 runtime 元信息
