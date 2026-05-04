## ADDED Requirements

### Requirement: 系统必须提供统一的 Workspace 容器
系统 SHALL 以 `/org/{orgName}/workspace` 作为统一工作区容器，同时承载租户级访问与项目访问登录后的默认工作空间。

#### Scenario: 租户级主体进入 Workspace
- **WHEN** 企业侧主体登录成功
- **THEN** 系统 SHALL 将其导向 `/org/{orgName}/workspace`

#### Scenario: 项目访问主体进入 Workspace
- **WHEN** 项目访问主体登录成功
- **THEN** 系统 SHALL 将其导向 `/org/{orgName}/workspace`

### Requirement: Workspace 必须区分 tenant mode 与 project-access mode
统一 workspace 容器 SHALL 支持至少两种访问模式：

- `tenant mode`
- `project-access mode`

其中：

- `tenant mode` SHALL 显示 org/project 管理导航与全局 sidebar
- `project-access mode` MUST NOT 显示全局 sidebar

#### Scenario: 租户级主体看到完整导航
- **WHEN** 持有 `tenant token` 的主体进入 workspace
- **THEN** 系统 SHALL 以 `tenant mode` 渲染页面
- **THEN** 页面 SHALL 显示 org/project 管理导航与全局 sidebar

#### Scenario: 项目访问主体不显示全局 Sidebar
- **WHEN** 持有 `project access token` 的主体进入 workspace
- **THEN** 系统 SHALL 以 `project-access mode` 渲染页面
- **THEN** 页面 MUST NOT 显示全局 sidebar

### Requirement: Project-access mode 必须先展示可访问 Project 列表
系统 SHALL 在 `project-access mode` 下先展示该主体可访问的 project 列表，并仅允许进入已授权 project。

#### Scenario: 展示可访问 Project 列表
- **WHEN** 项目访问主体进入 workspace 且拥有多个可访问 project
- **THEN** 系统 SHALL 展示其可访问 project 列表
- **THEN** 列表中 MUST NOT 包含未授权 project

### Requirement: Project 页面可见性必须由功能权限控制
系统 SHALL 在进入某个 project 后根据该主体的功能权限决定页面与入口可见性。

#### Scenario: 仅具备数据访问功能权限
- **WHEN** 项目访问主体进入某个 project，且仅拥有 `data.access`
- **THEN** 系统 SHALL 仅展示数据页入口
- **THEN** 系统 MUST NOT 展示模型设计、角色权限管理等入口

#### Scenario: 具备模型设计功能权限
- **WHEN** 项目访问主体进入某个 project，且拥有 `model.design`
- **THEN** 系统 SHALL 展示模型设计页入口
