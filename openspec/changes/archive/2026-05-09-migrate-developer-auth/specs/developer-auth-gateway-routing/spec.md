## ADDED Requirements

### Requirement: 开发者请求必须使用 gateway-authenticated routing
系统 MUST 将所有带开发者认证的 design-time 流量经由 gateway 路由。Web 客户端 MUST 使用 `front -> bff -> gateway -> backend` 路径，CLI 客户端 MUST 使用 `cli -> gateway -> backend` 路径。

#### Scenario: Web design 请求先经过 BFF 再进入 gateway
- **WHEN** 浏览器客户端调用 design-time GraphQL 或受保护的 design-time REST API
- **THEN** 请求先发送到 frontend BFF
- **THEN** BFF 将请求转发给 gateway
- **THEN** gateway 将请求转发给 backend

#### Scenario: CLI design 请求直接走 gateway
- **WHEN** CLI 客户端调用 design-time GraphQL 或受保护的 design-time REST API
- **THEN** 请求直接发送到 gateway
- **THEN** gateway 将请求转发给 backend

### Requirement: Gateway 必须是唯一的 developer authentication authority
gateway SHALL 成为唯一负责 developer access token validation、issuer compatibility rule 应用，以及判定开发者请求是否已认证的服务。

#### Scenario: 浏览器 bearer token 由 gateway 校验
- **WHEN** 浏览器客户端为 design-time 请求发送 developer bearer token
- **THEN** gateway 在将请求转发到 backend 之前完成 token 校验

#### Scenario: CLI bearer token 由 gateway 校验
- **WHEN** CLI 客户端为 design-time 请求发送 developer bearer token
- **THEN** gateway 在将请求转发到 backend 之前完成 token 校验

### Requirement: Frontend BFF 必须保持 developer auth 的 transport adapter 角色
frontend BFF SHALL 代理 developer authentication 与 design-time 请求，但不得演变为独立的 developer authentication truth source。

#### Scenario: Frontend auth route 转发到 gateway
- **WHEN** 浏览器调用 frontend developer auth route，例如 login、refresh 或 logout
- **THEN** frontend BFF 将该请求转发到 gateway
- **THEN** gateway 处理对应的 developer authentication 行为
