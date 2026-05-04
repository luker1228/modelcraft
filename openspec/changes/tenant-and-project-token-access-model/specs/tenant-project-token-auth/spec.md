## ADDED Requirements

### Requirement: 系统必须定义租户级与项目访问级两类访问令牌
系统 SHALL 定义两类访问令牌：

- `tenant token`：用于租户/org 范围主体
- `project access token`：用于项目访问主体

两类令牌都 MUST 至少包含 `iss`、`sub`、`org_name`、`exp`、`iat` 与可判定作用域的声明；系统 MUST 能基于这些声明区分租户级与项目访问级主体。

#### Scenario: 签发租户级令牌
- **WHEN** 企业侧主体登录成功
- **THEN** 系统 SHALL 返回 `tenant token`
- **THEN** 该令牌 MUST 可被识别为租户级作用域

#### Scenario: 签发项目访问级令牌
- **WHEN** 项目访问主体登录成功
- **THEN** 系统 SHALL 返回 `project access token`
- **THEN** 该令牌 MUST 可被识别为项目访问作用域

### Requirement: 同一套 Project API 必须同时接受两类令牌
系统 SHALL 允许同一套 project 级 API 同时接受 `tenant token` 与 `project access token`，并在进入业务处理前统一解析为 `ProjectPrincipal`。

`ProjectPrincipal` MUST 至少包含：

- `tokenScope`
- `subjectType`
- `subjectId`
- `orgName`
- `projectSlug`
- project 级功能权限展开结果
- project 级数据权限展开来源

#### Scenario: 租户级令牌访问 Project API
- **WHEN** 客户端使用 `tenant token` 调用某 project API
- **THEN** 系统 SHALL 将该请求解析为对应 project 上下文下的 `ProjectPrincipal`
- **THEN** 后续权限判断 SHALL 基于 `ProjectPrincipal` 执行

#### Scenario: 项目访问级令牌访问 Project API
- **WHEN** 客户端使用 `project access token` 调用某 project API
- **THEN** 系统 SHALL 将该请求解析为对应 project 上下文下的 `ProjectPrincipal`
- **THEN** 后续权限判断 SHALL 基于 `ProjectPrincipal` 执行

### Requirement: 租户级 API 必须拒绝项目访问级令牌
系统 MUST 对租户/org 治理 API 与 project API 采用不同的作用域准入规则。`project access token` MUST NOT 访问租户/org 治理 API。

#### Scenario: 项目访问级令牌访问 Org API
- **WHEN** 客户端使用 `project access token` 调用 org 列表、org 创建、org 更新等租户级 API
- **THEN** 系统 MUST 拒绝该请求并返回未授权错误

### Requirement: Project 级最终权限必须以后端展开结果为准
系统 MUST NOT 依赖访问令牌中的完整 project 权限列表作为最终授权依据。project 内的功能权限与数据权限 SHALL 以后端基于角色分配、权限包、预设或其它服务端授权数据展开的结果为准。

#### Scenario: 登录后权限发生变化
- **WHEN** 某项目访问主体已持有有效 `project access token`，且服务端修改了其 project 权限
- **THEN** 后续 project API 请求 SHALL 以后端最新展开结果执行授权判断
