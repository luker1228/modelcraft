## ADDED Requirements

### Requirement: EndUser 登录成功后必须一次性签发令牌
系统 MUST 在 EndUser 凭据校验通过后直接签发 JWT，且 MUST 不要求通过第二次“选项目”接口再次签发令牌。

#### Scenario: 登录成功即返回令牌
- **WHEN** EndUser 提交正确用户名与密码，且存在至少一个可访问项目
- **THEN** 系统 MUST 在登录响应中返回 JWT

### Requirement: 系统必须返回 EndUser 有权限访问的全部项目
系统 SHALL 在登录成功响应中返回该 EndUser 在当前组织下有权限访问的完整项目列表。

#### Scenario: 返回完整可访问项目集合
- **WHEN** EndUser 在组织下被授权访问多个项目
- **THEN** 登录响应 MUST 包含该 EndUser 可访问的全部项目，而非单个项目

### Requirement: EndUser 必须可在已授权项目中自由选择
系统 MUST 支持 EndUser 在已授权项目范围内自由选择当前操作项目，且 MUST 不要求重新登录或再次签发令牌。

#### Scenario: 切换到另一个已授权项目
- **WHEN** EndUser 已登录并选择另一个同样已授权的项目
- **THEN** 系统 MUST 允许该项目下的业务访问（在权限校验通过前提下）

### Requirement: 无项目访问权限时必须拒绝登录
系统 MUST 在 EndUser 无任何项目访问关系时返回显式错误，不得签发 JWT。

#### Scenario: 账号认证通过但无项目访问
- **WHEN** EndUser 凭据校验通过且可访问项目列表为空
- **THEN** 系统 MUST 返回 `NO_PROJECT_ACCESS` 错误并终止登录流程
