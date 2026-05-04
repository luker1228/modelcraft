## ADDED Requirements

### Requirement: Backend design API 必须信任 gateway 提供的 developer identity
backend design-time API SHALL 通过 gateway 提供的 trusted identity contract 来认证 developer request，而不是直接校验 developer JWT。

#### Scenario: Backend 接受经过 gateway 认证的 developer request
- **WHEN** gateway 转发一个带有所需 trusted developer identity headers 的 design-time 请求
- **THEN** backend 将该请求视为已认证的 developer request 并接受处理

#### Scenario: Backend 拒绝缺少 trusted developer identity 的请求
- **WHEN** 一个 design-time 请求到达 backend 时缺少所需的 trusted developer identity contract
- **THEN** backend 以 unauthorized 拒绝该请求

### Requirement: Backend design API 必须停止依赖 direct developer bearer token verification
backend design-time API SHALL 不再把浏览器或 CLI 直接提供的 developer bearer token 当作受保护 design-time access 的主要认证机制。

#### Scenario: 仅有 direct developer bearer token 不再足以访问 backend
- **WHEN** 客户端直接向 backend 发送一个受保护的 design-time 请求，并且只携带 developer bearer token
- **THEN** backend 必须拒绝该请求，除非它同时带有 trusted gateway identity contract

### Requirement: Developer auth migration 可以删除 legacy backend JWT compatibility
该迁移 SHALL 允许在 gateway-based compatibility 建立完成后，以破坏性方式删除 legacy backend-side developer JWT compatibility，包括 direct legacy issuer handling。

#### Scenario: Legacy backend developer JWT path 被移除
- **WHEN** gateway compatibility rollout 完成
- **THEN** backend design-time API 不再保留 legacy direct developer JWT verification 行为
