---
name: backend-develop
description: >
  ModelCraft Go 后端业务开发总指南。仅在“后端业务代码实现/修改”时触发：
  (1) 新增或修改 Domain、Application、Resolver、Repository 业务逻辑；
  (2) 新增业务能力需要打通接口到持久化链路（接口层→应用层→领域层→仓储层）；
  (3) 业务需求导致 Contract/SQL/实体字段变化，并需要同步后端实现；
  (4) 涉及多租户业务隔离（org 或 org+project）与参数传递链路设计。
  不要用于：纯调试排错、纯测试执行、纯部署运维、纯数据库迁移操作、前端任务、文档整理。
---

# ModelCraft Go 后端开发指南

## 触发边界（严格）

仅当任务是“后端业务开发实现”时使用本 skill：
- 需要编写或修改业务代码，而不是只读分析
- 变更发生在后端业务层链路（Interfaces/App/Domain/Infrastructure）
- 目标是交付业务功能或业务行为变更

以下场景不要触发本 skill（应使用对应专项 skill）：
- 后端报错定位、日志排查、线上问题修复（`backend-debug`）
- 仅执行测试（`bdd-test` 或 `integration-test`）
- 数据库连接、迁移、建表、`just db` 操作（`db-develop`）
- 部署、端口、环境变量、Docker（`deploy-info`）
- 前端开发、样式或交互任务（frontend 相关 skills）

## 分层架构速览

```
Interfaces (graphql/http)
      ↓
  App (用例编排)
      ↓
Domain (领域模型 + 仓储接口)   ←   Infrastructure (Repository + sqlc)
      ↓
    pkg/ (公共能力)
```

**铁律：上层依赖下层，下层不依赖上层。Domain 仅依赖 pkg/。**

| 层 | 可依赖 | 禁止依赖 |
|----|--------|----------|
| Interfaces | Application、同层 | Infrastructure、Domain（直接） |
| Application | Domain、Infrastructure | Interfaces |
| Infrastructure | Domain、pkg/ | Application、Interfaces |
| Domain | **仅 pkg/** | 所有 internal 层 |
| pkg/ | 无 | 所有 internal 层 |

## 快速导航

| 我要做… | 文档内容摘要 | 文档 |
|---------|-------------|------|
| 不确定代码该放哪层、某层能不能依赖另一层 | 分层图、职责边界、依赖规则、目录映射 | [architecture.md](../../../ai-metadata/backend/development/architecture.md) |
| 定义 Domain Repository 接口 | `ctx`/`orgName`/`projectSlug` 签名规范、scope 判定、反例 | [domain-development.md](../../../ai-metadata/backend/development/domain-development.md) |
| 处理 context / 显式参数传递 | Interface 提取、App/Domain 禁止隐式提取、全链路参数下传 | [context-handling.md](../../../ai-metadata/backend/development/context-handling.md) |
| 在 Repository 写 SQL 持久化 | sqlerr wrapper、NotFound 两种模式、事务 querier、SQL scope 条件 | [repo-develop.md](../../../ai-metadata/backend/development/repo-develop.md) |
| 处理 Application 层事务（WithTx） | WithTx 回调内必须使用 tx querier 构造 tx-scoped repo，禁止直接调 s.*Repo.* | [tx-querier-rules.md](../../../ai-metadata/backend/development/tx-querier-rules.md) |
| 处理错误边界 | RepositoryError vs BusinessError、跨层转换与日志打点 | [error-handling.md](../../../ai-metadata/backend/development/error-handling.md) |
| 做跨层类型转换（DTO/Entity/DB） | Null/指针语义、转换边界、常见转换坑与约束 | [type-conversion.md](../../../ai-metadata/backend/development/type-conversion.md) |
| 写日志、不确定字段 key 或 Stack() 使用时机 | logfacade 用法、字段常量、Stack() 边界与禁止项 | [logging.md](../../../ai-metadata/backend/development/logging.md) |
| 修改 Contract 后生成代码 | GraphQL/OpenAPI/sqlc 的生成命令与禁止手改目录 | [contract-sync.md](../../../ai-metadata/backend/development/contract-sync.md) |
| 运行 just 命令 | run/test/lint/db/generate 命令速查 | [justfile-guide.md](../../../ai-metadata/backend/tools/justfile-guide.md) |

## 多租户隔离模型（先判定 scope）

先看 API Contract 在 `modelcraft-backend/api/graph/` 的归属：

| Contract 目录 | scope | 必传参数 | SQL 最小过滤条件 |
|--------------|-------|----------|-------------------|
| `api/graph/org/schema/` | `org scoped` | `ctx`, `orgName`（写操作可由实体携带） | `WHERE org_name = ?` |
| `api/graph/project/schema/` | `org + project scoped` | `ctx`, `orgName`, `projectSlug`（写操作可由实体 `ProjectScope` 携带） | `WHERE org_name = ? AND project_slug = ?` |

**核心安全结论：**
- `FindByID` / `GetByID` 必须带 `orgName`，不能只靠 ID。
- 当 Contract 位于 `api/graph/project/schema/` 时，必须过滤 `org_name + project_slug`。

## ProjectScope 嵌入模式（推荐）

```go
type ModelLocator struct {
    project.ProjectScope        // 嵌入: OrgName + ProjectSlug
    DatabaseName         string
    ModelName            string
}
```

说明：
- 项目域实体/值对象优先嵌入 `ProjectScope`，统一承载 `OrgName + ProjectSlug`。
- `Create/Update` 通过实体携带 scope；`Find/Get/List/Delete/Exists` 仍要显式参数，避免调用侧遗漏隔离键。

## 参数必须从上一直传到下（禁止中途丢失）

租户参数链路必须贯穿：

```
Interfaces -> Application -> Domain Service/Repository Interface -> Infrastructure Repository -> SQL
```

### 标准链路

1. Interfaces 层
- 用 `ctxutils` 提取 `orgName` / `userID`
- project-scoped 请求额外获取 `projectSlug`
- 调用 App 时显式传参，不传“黑盒 context”

2. Application 层
- 方法签名显式接收 `orgName`、`projectSlug`、`userID`
- 继续向 Domain/Repository 显式下传
- 禁止在 App 层重新从 context 提取租户参数

3. Domain 层
- Repository 接口签名表达隔离边界（org 或 org+project）
- `Create/Update` 可通过实体中的 `ProjectScope` 携带 org/project
- `Find/Get/Delete/Exists/List` 必须显式参数，不能省略

4. Infrastructure 层
- 把 org/project 参数写入 sqlc params 或 SQL WHERE
- 不允许只按 ID 查询项目域资源
- 事务内用 tx querier 构造 tx-scoped repo，再传同样的租户参数

## 高频反例（看到就改）

- `FindByID(ctx, id)`：缺 `orgName`，可跨租户读取
- `List(ctx, orgName)`（项目域资源）：缺 `projectSlug`
- Resolver 提取了 `orgName`，App 签名未接收，导致下游丢参
- Repository SQL 只写 `WHERE id = ?`，没写 org/project 条件
- App/Domain 用 `ctx.Value()` 隐式取租户参数

## 绝对禁止

- `go func()`，必须用 `bizutils.GoWithCtx`
- 标准库 `errors`，统一用 `pkg/bizerrors`
- 裸 `log`，统一用 `logfacade`
- `task regenerate-gql`（会覆盖 resolver 实现），只用 `task generate-gql`
- 手改生成目录：
  - `internal/infrastructure/dbgen/`
  - `internal/interfaces/graphql/generated/`
  - `internal/interfaces/http/generated/`
- 手改 `api/openapi/openapi.yaml`（聚合文件）
- Domain 层依赖 Infrastructure / App
- Repository 层返回 `*BusinessError`
- 对 `dbgenwrap.NewSafeQuerier` 已包装调用再做 `sqlerr.WrapSQLError(...)`

## Schema-First 生成命令

| 修改的文件 | 生成命令 |
|-----------|---------|
| `api/graph/org/schema/*.graphql` 或 `api/graph/project/schema/*.graphql` | `task generate-gql` |
| `api/openapi/*.yaml`（各模块） | `task generate-oapi` |
| `db/queries/*.sql` 或 `db/schema/mysql/*.sql` | `task generate-sqlc` |

## 开发前 30 秒检查

- 先按 `api/graph/org|project` 判定资源 scope
- 检查方法签名是否显式包含所需参数
- 检查参数是否从 Resolver 一路传到 SQL
- 检查 SQL 是否按 scope 写了完整 WHERE 过滤
- 修改 contract 后是否执行对应 generate 命令
