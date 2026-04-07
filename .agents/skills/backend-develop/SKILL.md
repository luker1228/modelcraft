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

根据你正在做的事情选择：

| 场景 | 参考文档 |
|------|----------|
| 分层架构 / 依赖规则 / 各层职责 | [ai-metadata/backend/development/architecture.md](../../../../ai-metadata/backend/development/architecture.md) |
| 代码风格 / 命名 / 事务模式 / 协程 | [ai-metadata/backend/development/code-style.md](../../../../ai-metadata/backend/development/code-style.md) |
| 错误处理体系 / 各层职责 | [ai-metadata/backend/development/error-handling.md](../../../../ai-metadata/backend/development/error-handling.md) |
| Repository 层开发规范 | [ai-metadata/backend/development/repo-develop.md](../../../../ai-metadata/backend/development/repo-develop.md) |
| 日志规范（logfacade） | [ai-metadata/backend/development/logging.md](../../../../ai-metadata/backend/development/logging.md) |
| sqlc 自定义类型 | [ai-metadata/backend/development/sqlc-custom-types.md](../../../../ai-metadata/backend/development/sqlc-custom-types.md) |
| API Contract 同步（GraphQL/OpenAPI） | [ai-metadata/backend/development/contract-sync.md](../../../../ai-metadata/backend/development/contract-sync.md) |
| 领域模型设计 | [ai-metadata/backend/design/](../../../../ai-metadata/backend/design/) |
| 构建 / 测试 / just 命令 | [ai-metadata/backend/tools/justfile-guide.md](../../../../ai-metadata/backend/tools/justfile-guide.md) |
| 调试 / 本地运行 | [ai-metadata/backend/testing/debugging-workflow.md](../../../../ai-metadata/backend/testing/debugging-workflow.md) |

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

### Schema-First 代码生成命令

| 修改的文件 | 生成命令 |
|-----------|---------|
| `api/graph/schema/*.graphql` | `task generate-gql` |
| `api/openapi/*.yaml`（各模块） | `task generate-oapi` |
| `db/queries/*.sql` 或 `db/schema/mysql/*.sql` | `task generate-sqlc` |
