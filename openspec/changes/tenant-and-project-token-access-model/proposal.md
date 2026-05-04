## Why

当前系统把企业侧登录、项目访问登录、独立 end-user 页面流、独立 end-user GraphQL schema 混在一起讨论，导致身份语义、页面入口、project API 鉴权合同长期耦合。现在需要把认证与授权模型收敛成稳定骨架：企业侧主体持有租户级能力，项目访问主体持有 project 范围能力，二者共用同一套 project API 与 workspace 容器。

## What Changes

- 引入两类访问令牌：`tenant token` 与 `project access token`，分别承载租户级与项目级作用域。
- 在 backend 建立统一 `ProjectPrincipal` 鉴权模型，使同一套 project API 可以同时接受两类 token。
- **BREAKING** 停止将 project 级运行时能力继续建模为独立 `end_user` GraphQL schema 的长期方向，改为以 `project` schema 为唯一 project 级契约。
- 定义统一的 `/org/{orgName}/workspace` 工作区容器：
  - 企业侧登录进入 tenant mode；
  - 项目访问登录进入 project-access mode。
- **BREAKING** 项目访问登录不再跳转到独立 end-user 页面树，而是进入统一 workspace，并按功能权限裁剪导航与页面可见性。
- 在 project 内建立两层授权：
  - 功能权限：决定能否看到模型设计、角色权限、数据页等页面和入口；
  - 数据权限：决定进入数据页后可访问哪些 model、哪些 CRUD 动作、哪些行范围。
- 每个 project 内置受保护的 `project_admin` 角色；企业侧主体负责授予 project 管理员，project 管理员负责配置本 project 的成员功能权限与数据权限。
- 保留“项目访问由 role assignment 决定”的规则，但把登录产物、workspace 导航和 project API 鉴权统一到新的 token/principal 模型下。

## Capabilities

### New Capabilities
- `tenant-project-token-auth`: 定义租户级 token、项目访问级 token、统一 project principal，以及同一套 project API 的双 token 鉴权合同。
- `workspace-access-modes`: 定义统一 workspace 路由下的 tenant mode 与 project-access mode，包括默认落点、导航裁剪与 project 选择体验。
- `project-function-permissions`: 定义 project 内置 `project_admin` 角色、功能权限与数据权限的分层模型。

### Modified Capabilities
- `enduser-two-phase-auth`: 调整登录成功后的令牌语义与返回结构，使其返回项目访问级 token 与可访问项目列表，并支持统一 workspace 入口。
- `enduser-access-model`: 调整 project 访问授权与 project principal 的关系，使 role assignment 不仅决定“能否进入 project”，也成为 project 级功能权限与数据权限展开的基础。

## Impact

- Backend GraphQL 路由与中间件：`internal/interfaces/http/routes.go`、统一 auth middleware、project principal 注入链路。
- Backend auth domain：issuer、claims、token 签发/校验、project membership 与权限展开逻辑。
- Backend GraphQL contract：`api/graph/project/schema/*` 需要承载项目访问登录所需的 project 目录、catalog、数据 CRUD 与权限控制语义。
- Frontend 入口与 layout：`/login`、`/end-user/{orgName}/login`、`/org/{orgName}/workspace`、sidebar 与 project 选择页。
- Frontend API client：统一 project API client，按 token 类型进入不同 workspace mode。
- 现有独立 `end_user` schema、路由、页面树将进入收敛/迁移路径。
