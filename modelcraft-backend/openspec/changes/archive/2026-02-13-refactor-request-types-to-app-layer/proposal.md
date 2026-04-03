# Change: Refactor Request Types from Interface Layer to App Layer (Command Pattern)

## Why

当前 `CreateClusterRequest` 等请求类型定义在 `internal/interfaces/http/requests/` 包中，但 App 层（`internal/app/`）直接依赖了这些 interface 层的类型。这违反了 DDD 分层架构的依赖方向原则：**内层不应依赖外层**。

同时，项目中存在两种不一致的模式：
- Cluster/Model 域使用 Request 对象，但定义在 interface 层
- Enum/Project 域直接传递多个散参数，缺乏类型安全

## What Changes

1. **引入 App 层 Command 类型**：在 `internal/app/{context}/` 中定义 Command 结构体（如 `CreateClusterCommand`），替代当前 interface 层的 Request 类型
2. **Interface 层只负责转换**：GraphQL resolver / HTTP handler 将外部输入转换为 App 层的 Command，再调用 App Service
3. **统一所有域的模式**：Enum 和 Project 域也引入 Command 对象，消除散参数风格
4. **删除 `internal/interfaces/http/requests/` 中的业务 Request 类型**（保留纯 HTTP 绑定用途的类型如有需要）

## Impact

- Affected specs: `cluster-management`, `project-management`, `enum-error-handling`
- Affected code:
  - `internal/interfaces/http/requests/*.go` — Request 类型将迁移或删除
  - `internal/app/cluster/cluster_app.go` — 改用 Command 类型
  - `internal/app/modeldesign/model_app.go` — 改用 Command 类型
  - `internal/app/modeldesign/enum_service.go` — 引入 Command 类型替代散参数
  - `internal/app/project/project_service.go` — 引入 Command 类型替代散参数
  - `internal/interfaces/graphql/*.resolvers.go` — 更新转换逻辑
  - `internal/interfaces/graphql/adapter/*.go` — 更新 mapper 目标类型
