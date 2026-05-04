## ADDED Requirements

### Requirement: 每个 Project 必须内置受保护的 project_admin 角色
系统 SHALL 为每个 project 内置受保护的 `project_admin` 角色。该角色 MUST 作为 project 内最高功能权限模板存在，并默认拥有全部 project 页面访问与管理能力。

#### Scenario: 新建 Project 时生成内置管理员角色
- **WHEN** 系统创建一个新的 project
- **THEN** 系统 SHALL 自动创建或确保存在受保护的 `project_admin` 角色

#### Scenario: project_admin 可查看 Project 下所有页面
- **WHEN** 某主体在 project 内被授予 `project_admin`
- **THEN** 系统 SHALL 允许其查看该 project 下所有页面入口

### Requirement: Project 授权必须区分功能权限与数据权限
系统 SHALL 将 project 级授权拆分为两层：

- 功能权限：控制页面与功能入口可见性
- 数据权限：控制数据页内 model 可见性、CRUD 动作与行级范围

系统 MUST NOT 仅依赖单一角色名同时表达两层语义。

#### Scenario: 仅授予数据页入口
- **WHEN** 某 project 成员被授予 `data.access` 但未被授予模型设计相关功能权限
- **THEN** 系统 SHALL 允许其进入数据页
- **THEN** 系统 MUST NOT 允许其进入模型设计页

#### Scenario: 数据页内动作仍受数据权限约束
- **WHEN** 某 project 成员具备 `data.access`
- **THEN** 系统 SHALL 继续根据其数据权限决定可访问的 model、可执行的 CRUD 与可见行范围

### Requirement: 企业侧主体负责授予 project_admin，project_admin 负责管理项目成员权限
系统 SHALL 允许企业侧主体配置某 project 的 `project_admin`。被授予 `project_admin` 的主体 SHALL 能管理该 project 下普通成员的功能权限与数据权限。

#### Scenario: 企业侧主体授予 Project 管理员
- **WHEN** 企业侧主体在某 project 中授予某成员 `project_admin`
- **THEN** 系统 SHALL 记录该成员在该 project 下的管理员身份

#### Scenario: project_admin 管理普通成员权限
- **WHEN** 某主体已在 project 下拥有 `project_admin`
- **THEN** 系统 SHALL 允许其配置普通成员的功能权限与数据权限
