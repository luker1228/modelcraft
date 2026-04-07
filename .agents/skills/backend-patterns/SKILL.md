---
name: backend-patterns
description: ModelCraft Go 后端架构模式、API 设计、数据访问与服务端最佳实践。
---

# ModelCraft Go 后端模式

面向本项目（Go + Chi + sqlc + gqlgen）的后端架构模式与落地规范。

## 何时启用

- 设计/实现 REST 或 GraphQL API
- 新增应用层、领域层、基础设施层
- 优化数据库访问与事务
- 增加中间件、鉴权、限流
- 统一错误处理与日志

## 架构模式（DDD 分层）

```
Interfaces Layer     HTTP handlers / GraphQL resolvers
Application Layer    Use case orchestration
Domain Layer         Entities / Domain services
Infrastructure       Repositories / DB / External services
```

- **依赖规则**：内层不依赖外层，**都依赖抽象**
- **领域优先**：业务逻辑必须在 Domain
- **应用层负责事务边界**

### 各层职责

- **Interfaces 层**：请求解析与校验、DTO 转换、错误映射、响应组装（对外 API）
- **Application 层**：用例编排、事务边界、跨领域协调、前置验证
- **Domain 层**：实体与领域服务、业务规则、Repository 接口定义
- **Infrastructure 层**：数据持久化、外部服务调用、Repository 实现、模型转换

### 依赖规则（不得依赖实现）

- **Domain 层**：仅依赖语言标准库与领域抽象，**禁止依赖 Infrastructure/Interfaces 实现**
- **Application 层**：只依赖 Domain 抽象（Repository 接口、领域服务），**禁止依赖具体实现**
- **Interfaces 层**：可依赖 Application 暴露的用例与 DTO，**禁止依赖 Infrastructure 实现**
- **Infrastructure 层**：依赖 Domain 接口并提供实现，**不反向依赖 Application/Interfaces**

> 结论：**依赖只能指向抽象（接口/契约），不能指向实现**。

## API 设计（Schema-First）

### 总原则

Schema-First 的核心：**写 Schema，生成代码，禁止手改生成文件**。

两层由 Schema 锁定，Agent 在中间两层自由发挥：

```
┌─────────────────────────────┐
│       Interfaces Layer      │  ← gqlgen + oapi-codegen 生成
├─────────────────────────────┤
│      Application Layer      │  手写，业务逻辑编排
├─────────────────────────────┤
│        Domain Layer         │  手写，核心领域模型
├─────────────────────────────┤
│    Infrastructure Layer     │  ← sqlc 生成
└─────────────────────────────┘
```

每次修改 Schema 后，运行对应的生成命令（详见 `taskfile` skill）。

---

### GraphQL（gqlgen）— Interfaces Layer

**文件位置**：`api/graph/schema/*.graphql`

**工作流**：
1. 修改 `.graphql` schema 文件
2. 运行 `task generate-gql`
3. 实现生成的 resolver 接口

**关键约束**：
- ⛔ **禁止** `task clean-gql` 与 `task regenerate-gql`（有代码丢失风险）
- ⛔ **禁止**手改 `internal/interfaces/graphql/generated/` 下的生成文件
- ✅ 只改 `api/graph/schema/*.graphql`，然后 `task generate-gql`

**GraphQL 错误处理模式**（Schema 设计）：

```graphql
interface Error { message: String! }

type ResourceNotFound implements Error { message: String! }

union CreateResourceError = ResourceAlreadyExists | InvalidResourceInput

type CreateResourcePayload {
  resource: Resource          # Nullable
  error: CreateResourceError  # Optional
}
```

错误适配器：`internal/interfaces/graphql/adapter/*_error_adapter.go`

---

### OpenAPI（oapi-codegen）— Interfaces Layer

**文件位置**：`api/openapi/*.yaml`（模块文件，如 `auth.yaml`、`user.yaml`）

**工作流**：
1. 修改对应模块的 `api/openapi/*.yaml`
2. 运行 `task generate-oapi`
3. 实现生成的 `ServerInterface`

**关键约束**：
- ⛔ **禁止**直接编辑 `api/openapi/openapi.yaml`（聚合文件，由工具生成）
- ⛔ **禁止**手改 `internal/interfaces/http/generated/` 下的生成文件
- ✅ 只改各模块 `api/openapi/*.yaml`，然后 `task generate-oapi`

**示例**：

```yaml
# api/openapi/auth.yaml
/api/auth/login-url:
  get:
    operationId: getLoginURL
    parameters:
      - name: state
        in: query
        schema:
          type: string
```

生成后只需实现接口：

```go
// 只需实现这个接口，参数类型由生成代码保证
func (h *Handler) GetLoginURL(w http.ResponseWriter, r *http.Request, params GetLoginURLParams) {
    // ...
}
```

---

### sqlc — Infrastructure Layer

**文件位置**：
- 表结构：`db/schema/mysql/*.sql`
- 查询语句：`db/queries/*.sql`

**工作流**：
1. 修改表结构（`db/schema/mysql/`）或查询语句（`db/queries/`）
2. 运行 `task generate-sqlc`
3. 在 Repository 实现中调用生成的方法

**关键约束**：
- ⛔ **禁止**手改 `internal/infrastructure/dbgen/` 下的生成文件
- ✅ SQL 是唯一的真相来源，字段名/类型以 SQL 定义为准

**示例**：

```sql
-- db/queries/project.sql

-- name: GetProjectBySlug :one
SELECT * FROM projects
WHERE slug = ? AND org_name = ?
LIMIT 1;
```

生成后直接调用：

```go
// 类型安全，参数结构体由 sqlc 生成，编译期保障
project, err := q.GetProjectBySlug(ctx, dbgen.GetProjectBySlugParams{
    Slug:    slug,
    OrgName: orgName,
})
```

---

### 状态流转链路

一个请求的完整状态转移：

```
OpenAPI/GraphQL Schema
     ↓ (代码生成)
DTO（Interfaces 层，面向协议）
     ↓ (mapper 转换)
Domain 对象（Domain 层，面向业务）
     ↓ (mapper 转换)
DB Model（Infrastructure 层，面向存储）
     ↓ (sqlc 生成)
SQL Schema
```

每层对象形态不同，通过 `internal/interfaces/mapper/` 显式转换，禁止跨层共用同一对象。

## 应用层模式

### Use Case 组织

- 一个用例只做一件事
- 组合仓储与领域服务，不直接操作 DB

```go
// 应用层示例
func (s *Service) CreateProject(ctx context.Context, input *CreateProjectInput) (*Project, error) {
	if err := input.Validate(); err != nil {
		return nil, pkgerrors.Wrap(err)
	}
	return s.repo.Create(ctx, input)
}
```

## 仓储与数据访问

### Repository 规范

- 接口定义在 Domain 层
- sqlc 实现放在 Infrastructure
- 统一接受 `context.Context`

```go
type ModelRepository interface {
	Get(ctx context.Context, id string) (*modeldesign.Model, error)
	Create(ctx context.Context, model *modeldesign.Model) error
}
```

### SQL 安全

- 必须使用参数化查询
- 动态 `order by`/`group by` 必须白名单校验

## 错误处理模式

- **必须使用** `pkg/bizerrors`
- 业务错误使用 `bizerrors.New`/`Wrap`
- 数据不存在返回 `nil`（不是 error）

```go
if pkgerrors.Is(err, pkgerrors.NotFound) {
	return nil, nil
}
```

## 日志模式

- 使用 `pkg/logfacade`
- 关键分支必须打日志
- 结构化字段仅限错误或常量

```go
logger := logfacade.GetDefault()
logger.Infof("operation started, id=%s", id)
logger.Errorf("operation failed, err=%v", err)
```

## 事务管理（最佳实践）

### 核心原则

- 事务边界必须由 **Application Service** 管理
- **事务尽可能短**：只包含必须原子化的写操作
- **前置验证在事务外**：减少锁持有时间
- **事务中实例化 Repository**：传入 `tx`，避免 `WithTx` 模式

### 推荐模式（Repository 在事务内实例化）

```go
err := s.db.Transaction(func(tx *sql.DB) error {
	repo := repository.NewSqlcOrganizationRepository(tx)
	if err := repo.Create(ctx, entity); err != nil {
		return bizerrors.Wrap(err, "create organization failed")
	}
	return nil
})
```

### 禁止模式（WithTx 反复包裹）

```go
// ❌ 不推荐：每次操作都调用 WithTx
err := s.db.Transaction(func(tx *sql.DB) error {
	txRepo := s.repo.WithTx(tx).(organization.OrganizationRepository)
	return txRepo.Create(ctx, entity)
})
```

### 事务失败返回业务错误

```go
if err := s.db.Transaction(func(tx *sql.DB) error {
	// ...
	return bizerrors.Wrap(err, "transaction failed")
}); err != nil {
	return nil, err
}
```

## 安全与鉴权

- HTTP 中间件负责鉴权
- 解析 JWT 时不允许日志打印敏感信息

## 性能模式

- 预分配切片容量
- 热路径避免反射与过度分配
- 减少 N+1 查询，优先批量加载

## 测试策略

- 遵循 TDD：先写测试再实现
- **Domain 层单元测试覆盖率必须 ≥80%**
- **Infrastructure 层仅需必要的单元测试**（核心查询/事务边界/错误处理）
- **Interfaces 层 Adapter 转换必须有单元测试**（请求/响应 DTO 映射、错误映射）
- 集成测试使用 `tests/automated`

**记住**：模式要服务于复杂度，优先保证可维护性与可测试性。
