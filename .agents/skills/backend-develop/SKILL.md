---
name: backend-develop
description: >
  ModelCraft Go 后端开发完整指南。覆盖 DDD 分层架构（pkg/interfaces/app/domain/infrastructure）、
  Schema-First API 设计（GraphQL + OpenAPI + sqlc）、错误处理体系、事务管理、日志规范、Repository
  模式与测试策略。当你在进行任何 Go 后端开发工作时都应触发此技能，包括：
  (1) 新增功能、设计接口或领域模型，
  (2) 实现 Repository / App / Resolver，
  (3) 设计错误处理或日志，
  (4) 修改 GraphQL schema / OpenAPI / SQL 查询，
  (5) 不确定代码该放哪一层，或如何遵守架构约束。
---

# ModelCraft Go 后端开发指南

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

根据你正在做的事情选择对应文档：

| 我要做… | 文档内容摘要 | 文档 |
|---------|-------------|------|
| 不确定代码该放哪层、某层能不能依赖另一层 | 完整分层图、各层职责边界、允许/禁止的 import 关系、目录映射、GraphQL 三条 API 通道 | [architecture.md](../../../ai-metadata/backend/development/architecture.md) |
| 在 `internal/domain/` 下定义 Repository 接口 | Repository 接口方法签名规范：ctx 必传、orgName 必传（含 FindByID）、projectSlug 的适用范围、正确/错误示例 | [domain-development.md](../../../ai-metadata/backend/development/domain-development.md) |
| 在 `internal/infrastructure/` 下实现 Repository | sqlerr wrapper 用法、RecordNotFound 两种模式（返回 error vs 返回 bool）、sql.Null\* 转换、事务 Querier 接收、编译期接口检查 | [repo-develop.md](../../../ai-metadata/backend/development/repo-develop.md) |
| 处理错误、不知道哪层该用什么错误类型 | bizerrors vs RepositoryError 的职责边界、各层错误转换代码示例、RecordNotFound 判断流程、日志+堆栈打印时机 | [error-handling.md](../../../ai-metadata/backend/development/error-handling.md) |
| 修改 `.graphql` / `.yaml` / `.sql` 文件后需要生成代码 | Schema-First 工作流：哪个文件改了跑哪条命令、禁止手改的目录、两套 GraphQL schema 的目录结构 | [contract-sync.md](../../../ai-metadata/backend/development/contract-sync.md) |
| 写日志、不确定该用哪个日志方法或字段 | logfacade 用法、字段常量、Stack() 只在 Interfaces 层用、禁止裸 log | [logging.md](../../../ai-metadata/backend/development/logging.md) |
| 处理 JSON 字段或自定义 sqlc 类型 | StringSlice / JSONMap 等自定义类型的 Scan+Value 实现模板、`db:"type:json"` 标签要求、常见 sqlc 坑（软删除、零值更新、N+1） | [sqlc-custom-types.md](../../../ai-metadata/backend/development/sqlc-custom-types.md) |
| 设计领域模型、实体/值对象/聚合边界 | 核心领域模型文档、DDD 原则、业务规则（最高优先级，实现必须以此为准） | [design/](../../../ai-metadata/backend/design/) |
| 运行 just 命令、不知道命令叫什么 | 所有 just 命令速查表（run/test/lint/db/generate/deploy 等），含参数说明 | [justfile-guide.md](../../../ai-metadata/backend/tools/justfile-guide.md) |
| 本地启动服务、查日志、数据库重置 | just run / just logs / just db reset 常用命令、request_id 追踪、常见问题快速修复 | [debugging-workflow.md](../../../ai-metadata/backend/testing/debugging-workflow.md) |

## 最高频规则（直接内联，无需查文档）

### 错误处理三层流转

```
Repository.Find()
    ├─ sql.ErrNoRows → (nil, nil) 或 (nil, NotFoundError)   # 由模式决定，见05
    ├─ 其他 DB 错误  → (nil, RepositoryError)               # 禁止 BusinessError
    └─ 成功          → (entity, nil)

App.UseCase()
    ├─ err != nil    → ConvertRepositoryError → BusinessError(SYSTEM_ERROR)
    ├─ entity == nil → NewErrorFromContext    → BusinessError(NOT_FOUND.XXX)
    └─ 成功          → return entity

Interfaces (Resolver)
    └─ BusinessError → adapter → GraphQL 联合错误
       （转换前必须 logfacade.Stack(err)，这是唯一打堆栈的地方）
```

### 绝对禁止

- `go func()` → 必须用 `bizutils.GoWithCtx`
- 标准库 `errors` → 用 `pkg/bizerrors`
- 裸 `log` → 用 `logfacade`
- `task regenerate-gql` → 会删除 resolver 实现，只用 `task generate-gql`
- 手改 `internal/infrastructure/dbgen/`、`internal/interfaces/graphql/generated/`、`internal/interfaces/http/generated/`
- 手改 `api/openapi/openapi.yaml`（聚合文件，由工具生成）
- Domain 层依赖 Infrastructure 或 App
- Repository 层返回 `*BusinessError`
- **Domain Repository 接口方法缺少 `ctx context.Context`**（第一个参数必须是 ctx）
- **`FindByID` / `GetByID` 等按 ID 查询方法缺少 `orgName`**（租户隔离漏洞，攻击者可跨租户读取）

### Domain Repository 接口方法签名规范

```go
// ✅ 正确：ctx 第一位，orgName 必传，项目域资源需 projectSlug
type EnumRepository interface {
    Create(ctx context.Context, enum *EnumDefinition) error          // orgName 通过实体携带
    FindByID(ctx context.Context, orgName, id string) (*EnumDefinition, error)  // 显式 orgName
    FindByName(ctx context.Context, orgName, projectSlug, name string) (*EnumDefinition, error)
    List(ctx context.Context, orgName, projectSlug string) ([]*EnumDefinition, error)
    Delete(ctx context.Context, orgName, projectSlug, name string) error
}

// ❌ 错误：缺 ctx + 缺 orgName → 租户隔离漏洞
FindByID(id string) (*EnumDefinition, error)
Create(enum *EnumDefinition) error
```

**orgName 规则**：
- 查询/删除/检查方法 → 显式传入 `orgName`，写入 SQL WHERE 条件
- `Create`/`Update` → `orgName` 由实体（`ProjectScope` 嵌入）携带，不重复传参
- `projectSlug` → 仅项目域资源（Model/Field/Enum）需要，Org/Cluster/Project 本身不需要

详见 [domain-development.md](../../../ai-metadata/backend/development/domain-development.md)

### Schema-First 代码生成命令

| 修改的文件 | 生成命令 |
|-----------|---------|
| `api/graph/schema/*.graphql` | `task generate-gql` |
| `api/openapi/*.yaml`（各模块） | `task generate-oapi` |
| `db/queries/*.sql` 或 `db/schema/mysql/*.sql` | `task generate-sqlc` |
