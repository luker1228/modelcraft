## Why

当前 runtime 能力仍主要按 `developer / end-user` 两套入口与心智建模，但实际业务语义已经更接近三层身份：组织级、项目级、用户级。随着数据面能力持续扩展，以及未来为用户级主体引入 API key，这种按入口分裂的模型会放大重复接口、重复鉴权和重复权限判断，难以稳定演进。

## What Changes

- 引入统一的**三层身份 runtime 访问模型**，用 `组织级 / 项目级 / 用户级` 替代 `developer / end-user` 作为核心权限语义。
- 将 runtime 访问能力明确划分为 **control plane** 与 **data plane**，避免把模型管理能力与数据访问能力混在同一套身份命名下。
- 为 data plane 定义统一 principal 模型，显式区分主体层级、作用范围、凭证类型与访问模式。
- 定义三种 data plane 访问模式：用户本人访问、管理员扮演某个用户访问、管理员全权限访问。
- 统一 catalog、model schema、runtime data、`meta/user` 等数据面能力的授权解释，要求“能被用户级访问的能力，在满足上下文条件时也能被上层管理员访问”。
- 约束用户级数据访问的最小可见面：在查看具体数据前，必须先有能力查看其可访问范围内的 database / model catalog。
- 为后续引入用户级 API key 与管理员 API key 预留统一接入模型，避免再按“入口系统”复制一套 runtime 接口。

## Capabilities

### New Capabilities
- `runtime-data-plane-access`: 定义组织级、项目级、用户级三层身份在 data plane 下的统一 principal、访问模式与能力边界。

### Modified Capabilities
- `enduser-access-model`: 将现有 end-user project 访问语义扩展为用户级 data access 语义，明确 catalog 可见性、项目作用域与上层管理员的兼容访问规则。

## Impact

- 影响后端 runtime 鉴权与 principal 解析：`/graphql/org/.../db/.../model/...`、`/graphql/end-user/...`、`/graphql/runtime/.../meta/user` 等数据面入口都需要挂靠统一 access model。
- 影响前端 workspace 与 data-plane 查询接入：现有 `developer / end-user` 命名需要逐步收敛到组织级 / 项目级 / 用户级心智。
- 影响 `meta/user`、database catalog、model catalog、runtime record CRUD 的权限解释与后续路由设计。
- 影响未来 API key 方案：用户级与管理员级 API key 将复用同一 principal 模型，而不是各自复制一套 runtime 接口体系。
