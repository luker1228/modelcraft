## Context

当前系统已经存在两类真实主体，但命名和路由仍然主要围绕 `developer / end-user` 展开：

- 企业侧管理主体通过 `/login` 进入，默认落在 org/project 管理空间。
- 项目访问主体通过 `/end-user/{orgName}/login` 进入，当前仍绑定独立 end-user 页面流与独立 end-user GraphQL schema。
- backend 同时维护 developer runtime route、project GraphQL route、end-user GraphQL route、meta/user route，多条链路的鉴权合同和 principal 注入方式并不统一。
- `users` 更接近控制面主体，`end_user_users` 更接近项目访问主体；两者虽然都叫 user，但并不承载同一种作用域语义。

随着 project 内模型设计、权限管理、数据 CRUD、`EndUserRef`/`meta/user` 查询继续收敛到同一套 project 能力，继续沿用“独立 end-user schema + 独立 end-user shell”会持续复制路由、token 语义、权限判断和前端容器。

## Goals / Non-Goals

**Goals:**

- 明确定义两类 token：`tenant token` 与 `project access token`。
- 让同一套 project API 同时支持两类 token，通过统一 `ProjectPrincipal` 完成鉴权。
- 定义统一 `/org/{orgName}/workspace` 容器下的两种访问模式：
  - tenant mode
  - project-access mode
- 将 project 内授权拆为两层：
  - 功能权限
  - 数据权限
- 为每个 project 提供受保护的内置 `project_admin` 角色。
- 明确 `users` 与 `end_user_users` 不合并，分别承担控制面主体与项目访问主体职责。

**Non-Goals:**

- 本次不把控制面主体再细分为完整的 org-level 角色体系。
- 本次不统一 `users` 与 `end_user_users` 的物理表结构。
- 本次不一次性删除所有旧路由；允许存在过渡期 facade。
- 本次不设计完整 API key 生命周期，只要求模型可兼容后续扩展。

## Decisions

### 1. Token 按作用域命名，而不是按“控制面/数据面”命名

**Decision**

系统定义两类访问令牌：

- `tenant token`
  - 由企业侧登录签发
  - 代表租户/org 范围主体
  - 可访问 org API 与 project API
- `project access token`
  - 由项目访问登录签发
  - 代表项目访问主体
  - 不可访问 tenant 治理 API
  - 可访问被授权 project 的 project API

**Why**

- project API 同时包含模型设计、权限管理、catalog、数据 CRUD，不是纯“数据面 API”。
- 作用域才是最稳定的区分维度：租户级能力 vs 项目级能力。

**Alternatives considered**

- `control token / data token`
  - 放弃原因：project 内同时存在管理与数据能力，`data token` 命名过窄。
- `developer token / end-user token`
  - 放弃原因：把入口术语错误延伸为最终权限语义。

### 2. 同一套 project schema 是长期唯一的 project 契约

**Decision**

长期方向上，project 级 GraphQL 能力以 `api/graph/project/schema/*` 为唯一契约；不再继续扩展独立 `api/graph/end_user/schema` 作为 project 级平行 schema。

项目访问登录所需的 project 列表、catalog、数据 CRUD、`meta/user`、project 级功能权限控制，最终都应收敛到统一 project schema 与 project 路由族。

**Why**

- project 能力本质上是同一组资源：project、model、role、bundle、catalog、runtime data。
- 为不同 token 维持两套平行 schema，会持续放大字段漂移、resolver 重复与鉴权分歧。

**Alternatives considered**

- 保留 `project schema` 和 `end_user schema` 长期并行
  - 放弃原因：同一 project 能力会重复维护两份契约和两套 token 解释。

### 3. 用统一 `ProjectPrincipal` 承接双 token 鉴权

**Decision**

所有 project API 在进入 resolver/app service 前，统一解析为 `ProjectPrincipal`：

```text
ProjectPrincipal
  tokenScope: tenant | project_access
  subjectType: control_user | project_user
  subjectID: string
  orgName: string
  projectSlug: string
  functionPermissions: [...]
  dataGrantSource: role_assignments / bundles / presets
  isProjectAdmin: bool
```

其中：

- `tenant token` 进入 project API 时，需按 org/project 上下文展开成 project principal。
- `project access token` 进入 project API 时，直接按 project assignment 与授权结果展开成 project principal。

**Why**

- resolver 不应再判断原始 token 家族，而应只判断 project principal。
- 这能把路由 facade 与业务权限解释彻底分层。

**Alternatives considered**

- 为两种 token 分别维护独立 middleware + 独立 resolver 入口
  - 放弃原因：会继续固化“两个世界”的实现。

### 4. 统一 workspace 路由，但区分访问模式

**Decision**

统一使用 `/org/{orgName}/workspace` 作为工作区容器，但根据 tokenScope 进入不同访问模式：

- tenant mode
  - 有全局 sidebar
  - 默认 org/project 管理导航
- project-access mode
  - 无全局 sidebar
  - 先展示可访问 project 列表
  - 进入 project 后只显示被功能权限允许的页面

项目访问登录后不再进入独立 end-user 页面树。

**Why**

- 复用同一 workspace 容器可降低前端维护成本。
- 访问模式裁剪比单独再维护一套 end-user 产品壳更稳定。

**Alternatives considered**

- 继续维护独立 end-user shell
  - 放弃原因：UI 壳子、路由、GraphQL client、权限判断都会继续重复。

### 5. Project 授权拆为功能权限与数据权限两层

**Decision**

project 内权限模型分两层：

- 功能权限
  - 控制页面与入口可见性
  - 例如：`project.view`、`project.member.manage`、`model.design`、`project.role.manage`、`data.access`
- 数据权限
  - 控制数据页内部能力
  - 例如：可访问哪些 model、哪些 CRUD、哪些 row scope、哪些 preset/custom grant

每个 project 内置一个受保护的 `project_admin` 角色，默认拥有全部 project 页面能力和完整的 project 管理能力。非 `project_admin` 成员则通过功能权限和数据权限组合出可见能力。

**Why**

- “只能看数据页”是功能权限的结果，不需要再引入单独“数据人员”身份。
- 数据页里的 CRUD/RLS/model 可见性则属于另一层数据权限，不应和页面导航混在一起。

**Alternatives considered**

- 只有一套 project 角色，不拆功能权限与数据权限
  - 放弃原因：页面可见性和数据 CRUD 会互相污染，难以精确控制。
- 单独维护一套 project 管理人员体系
  - 放弃原因：会与 project 成员体系重复。

### 6. 控制面主体与项目访问主体保持分表

**Decision**

继续保留：

- `users` / `user_organizations`：控制面主体
- `end_user_users`：项目访问主体

本次不合并两者，不让 `END_USER_REF` 直接指向控制面 `users`。

**Why**

- `users` 更接近 org membership 语义。
- `end_user_users` 更接近项目访问与数据所有权语义。
- 如果直接合并，`owner`、`me`、project 访问范围、org/project 作用域都会混乱。

**Alternatives considered**

- 直接把 `end_user_users` 并入 `users`
  - 放弃原因：控制面主体与项目访问主体不是同一种 principal。

## Risks / Trade-offs

- [Risk] 旧 `end_user` schema 与新统一 project schema 在迁移期并存。 → Mitigation：先统一 principal 与 capability，再逐步收敛路由和页面。
- [Risk] `tenant token` 进入 project API 时，若直接默认全 project 权限，可能弱化显式 project 授权边界。 → Mitigation：在 principal 展开阶段显式标记 `isProjectAdmin` / functionPermissions，而不是在 resolver 内隐式放行。
- [Risk] project-access mode 共用同一 workspace，前端 layout 若实现粗糙，可能泄露不该出现的导航。 → Mitigation：把 mode 作为一等状态注入 layout，而不是散落条件渲染。
- [Risk] 功能权限与数据权限分层后，现有 role/bundle 模型需要补功能权限来源。 → Mitigation：为 `project_admin` 先提供内置模板，普通成员按增量方式引入功能权限映射。

## Migration Plan

1. 在 spec 层固定 token 家族、workspace mode、project principal、功能权限/数据权限分层。
2. 后端新增统一 project auth middleware，使 project 路由能接受两类 token 并注入 `ProjectPrincipal`。
3. 将项目访问登录响应升级为 `project access token + accessible projects + workspace entry metadata`。
4. 前端收敛到统一 workspace 路由，为 tenant mode 与 project-access mode 增加独立 layout 分支。
5. 为 project 引入受保护的内置 `project_admin` 角色，并补 project 成员功能权限映射。
6. 逐步把现有 end-user project 级查询与页面迁移到统一 project schema / workspace。

回滚策略：

- 若统一 schema 迁移面过大，可暂时保留旧 end-user facade，但内部仍必须通过统一 `ProjectPrincipal` 解释权限，不回退认证和授权语义模型。

## Open Questions

- `tenant token` 是否在当前阶段永远代表“该 org 下全部权限”，还是后续需要为 tenant 主体补精细 org 角色。
- `project_admin` 的功能权限是否完全硬编码为内置模板，还是后续允许在 UI 中显式查看/审计其权限集合。
- `project access token` 的 claims 是否需要携带可访问 project 摘要，还是登录后完全靠查询接口返回。
- `meta/user` 最终是否挂在 org-scoped project schema 扩展下，还是挂在统一 runtime namespace 下但复用 project principal。
