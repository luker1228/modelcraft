# 架构分层规范

> 本文档描述 ModelCraft 的 DDD 分层架构、各层职责边界、依赖规则和实际目录映射。
> `.codebuddy/rules/code-style/layered-dependency.md` 是本文档依赖规则的 lint 执行入口，规则详情以本文档为准。

## 分层总览

```
┌─────────────────────────────────────────────────────┐
│              Interfaces 接口层                       │
│   internal/interfaces/  (graphql / http / runtime)  │
├─────────────────────────────────────────────────────┤
│              Application 应用层                      │
│              internal/app/                          │
├──────────────────────────┬──────────────────────────┤
│     Domain 领域层         │   Infrastructure 基础设施层│
│   internal/domain/       │   internal/infrastructure/│
├──────────────────────────┴──────────────────────────┤
│              Shared Kernel                          │
│              pkg/                                   │
└─────────────────────────────────────────────────────┘
```

**依赖方向（单向，不可逆）**：

```
Interfaces → Application → Domain
                       ↘ Infrastructure → Domain
                                      ↘ pkg/
```

任何层都可以依赖 `pkg/`。Domain 层**不依赖任何 internal 层**。

---

## 各层职责

### Domain 层 (`internal/domain/`)

**唯一的核心**。包含业务实体、领域规则和仓储接口定义。**不依赖任何其他 internal 层**。

```
internal/domain/
├── modeldesign/   # 设计时域：DataModel、FieldDefinition、EnumDefinition
├── modelruntime/  # 运行时域：QueryExecutor
├── project/       # 项目隔离域
├── cluster/       # 数据库集群域
├── organization/  # 组织域
├── auth/          # 认证域
├── shared/        # 跨域共享：RepositoryError、sentinel errors
└── ...
```

**领域层包含什么**：
- 业务实体（`DataModel`、`FieldDefinition` 等）
- 仓储接口（`ModelRepository`、`ClusterRepository` 等，接口定义在使用方）
- 领域服务（`FieldValidator`、`ComparisonService` 等）
- 值对象（`ProjectScope`、`ModelLocator`、`FieldType` 等）

→ 仓储接口示例：`internal/domain/modeldesign/model_repository.go`
→ 实体定义示例：`internal/domain/modeldesign/model.go`
→ 值对象示例：`internal/domain/project/project_scope.go`

**✅ 正确**：Domain 层只导入 `pkg/`
```go
// internal/domain/modeldesign/model.go
import (
    "modelcraft/pkg/bizerrors"  // ✅ 只依赖 pkg/
)
```

**❌ 错误**：Domain 层导入 infrastructure 或 app
```go
import (
    "modelcraft/internal/infrastructure/repository"  // ❌ 禁止
    "modelcraft/internal/app/modeldesign"            // ❌ 禁止
)
```

---

### Infrastructure 层 (`internal/infrastructure/`)

实现 Domain 层定义的仓储接口，处理所有外部依赖（数据库、第三方服务）。**只依赖 Domain 层和 pkg/**。

```
internal/infrastructure/
├── repository/    # sqlc 仓储实现（实现 domain 层的接口）
│   ├── sql_modeldesign_repository.go
│   ├── sql_database_cluster_repository.go
│   ├── sql_error_analyzer.go   # DB 错误分类
│   └── error_helper.go
├── database/
│   ├── ddl/       # DDL 执行（部署模型到客户 DB）
│   └── dml/       # DML 执行（运行时查询）
├── dbgen/         # sqlc 生成的 DB 访问代码
└── auth/          # Casbin、AuthProvider 客户端
```

**Infrastructure 层规则**：
- 返回 `error`（`shared.RepositoryError`），**不返回** `*BusinessError`
- `RecordNotFound` → 返回 `(nil, nil)`，不转换为业务错误
- DB 错误分类用 `sql_error_analyzer.go`，不自行解析错误字符串

→ 错误处理详见 `ai-metadata/2-development/error-handling.md`

---

### Application 层 (`internal/app/`)

编排领域对象和仓储，实现业务用例。负责事务管理。**不直接调用 DB，通过仓储接口操作数据**。

```
internal/app/
├── modeldesign/   # 模型设计用例（创建/更新/删除模型、字段、枚举）
├── cluster/       # 集群管理用例
├── project/       # 项目管理用例
├── organization/  # 组织管理用例
├── auth/          # Token、权限加载用例
├── role/          # 角色管理用例
└── modelruntime/  # 运行时 GraphQL 查询用例
```

**Application 层规范**：
- 入参用 `Command` 对象（`commands.go`），**不直接接收 GraphQL 生成类型**
- 将 Repository `error` 转换为 `*BusinessError` 后向上传递
- `nil` 结果在此层赋予业务语义（转换为 `NOT_FOUND.*`）
- 事务由此层控制，不在 Repository 内开启事务

→ Command 定义示例：`internal/app/modeldesign/commands.go`

---

### Interfaces 层 (`internal/interfaces/`)

处理 HTTP/GraphQL 请求，调用 Application 层，格式化响应。**不直接调用仓储或领域服务**。

```
internal/interfaces/
├── graphql/              # 设计时 GraphQL（gqlgen）
│   ├── *.resolvers.go   # 各域解析器
│   ├── adapter/         # 实体 → GraphQL DTO 映射；BusinessError → GraphQL 错误
│   └── generated/       # gqlgen 自动生成，不手动编辑
├── http/                 # REST API（oapi-codegen）— 仅限 Auth/Org/Webhook
│   ├── server.go        # ServerInterface 实现
│   ├── handlers/        # 域处理器（org、webhook）
│   └── generated/       # oapi-codegen 生成，不手动编辑
└── runtime/              # 运行时 GraphQL（动态 Schema）
    └── handler.go
```

**Interfaces 层规范**：
- Resolver 从 `ctx` 提取 `orgName`（用 `ctxutils`），作为显式参数传给 Application 层
- 错误转换前必须记录 `logfacade.Stack(err)`（唯一允许打堆栈的地方）
- 将 `*BusinessError` 通过 `adapter/*_error_adapter.go` 转为 GraphQL 联合错误类型
- Mutation payload 中错误字段可为 nil，数据字段可为 nil，**不混用**

---

## 三条 API 通道

| 通道 | 路径 | Schema 来源 | 用途 |
|------|------|-------------|------|
| 设计时 GraphQL | `/org/modelcraft/design/graphql` | `api/graph/schema/*.graphql`（静态） | 模型/字段/枚举/项目/集群 CRUD |
| REST (OpenAPI) | `/api/auth/*` `/api/org/*` `/api/webhook/*` | `api/openapi/*.yaml` | 认证、组织管理、Webhook |
| 运行时 GraphQL | `/graphql/:projectId/:clusterName/:database/:modelName` | 运行时从模型定义动态生成 | 用户数据查询/变更 |

**业务域功能（Project/Model/Cluster/Enum）只走设计时 GraphQL，禁止添加到 REST API。**

---

## Shared Kernel (`pkg/`)

所有层均可依赖，**不可反向依赖任何 internal 层**。

```
pkg/
├── bizerrors/     # 业务错误（ErrorDefinition、BusinessError）
├── logfacade/     # 日志抽象（Logger 接口、字段常量）
├── bizutils/      # 通用工具（ID 生成、JSON、goroutine 安全启动）
├── ctxutils/      # Context 工具（提取 orgName、requestId）
├── schema/        # 运行时 Schema 生成（GraphQL/JSON Schema）
├── config/        # 配置加载
└── crypto/        # 加密工具
```

---

## GraphQL Schema 组织

```
api/graph/schema/
├── types.graphql    # 基础类型（Model、Field、Enum、Project、Cluster）
├── inputs.graphql   # Mutation 输入类型
├── errors.graphql   # 业务错误联合类型
└── schema.graphql   # 根 Query/Mutation
```

修改 `.graphql` 文件后运行 `task generate-gql` 更新生成代码。
**禁止运行 `task regenerate-gql`**（会删除 resolver 实现）。

---

## 依赖规则

各层允许和禁止的导入关系如下表，违反即视为架构违规：

| 层 | 可依赖 | 禁止依赖 |
|----|--------|----------|
| Interfaces | Application、同层 | Infrastructure、Domain（直接） |
| Application | Domain、Infrastructure | Interfaces |
| Infrastructure | Domain、pkg/ | Application、Interfaces |
| Domain | pkg/ | 所有 internal 层 |
| pkg/ | 无 | 所有 internal 层 |

### ✅ 正确示例

```go
// Interfaces → Application（正确）
// internal/interfaces/graphql/model.resolvers.go
import (
    "modelcraft/internal/app/modeldesign"           // ✅ Interfaces → App
    "modelcraft/internal/interfaces/graphql/adapter" // ✅ 同层
)

// Application → Domain + Infrastructure（正确）
// internal/app/modeldesign/model_app.go
import (
    "modelcraft/internal/domain/modeldesign"         // ✅ App → Domain
    "modelcraft/internal/infrastructure/repository"  // ✅ App → Infrastructure
)

// Infrastructure → Domain（正确）
// internal/infrastructure/repository/sql_modeldesign_repository.go
import (
    "modelcraft/internal/domain/modeldesign"  // ✅ Infrastructure → Domain
)

// Domain → pkg/（正确）
// internal/domain/modeldesign/model.go
import (
    "modelcraft/pkg/bizerrors"  // ✅ Domain → pkg/ only
)
```

### 值对象复用模式

值对象可通过 **Go 结构体嵌入** 在多个实体间复用，避免字段重复：

```go
// internal/domain/project/project_scope.go
type ProjectScope struct {
    OrgName     string
    ProjectSlug string
}

// internal/domain/modeldesign/model.go
type ModelLocator struct {
    project.ProjectScope  // ← 嵌入值对象
    DatabaseName string
    ModelName    string
}

// 嵌入后 ProjectScope 的字段被提升
locator.OrgName      // ✅ 可直接访问（字段提升）
locator.ProjectSlug  // ✅ 可直接访问（字段提升）
locator.DatabaseName // ✅ ModelLocator 的直接字段
locator.ModelName    // ✅ ModelLocator 的直接字段

// Validate 方法也被继承
err := locator.Validate() // ✅ 调用 ProjectScope.Validate() 再验证其他字段
```

**其他嵌入 ProjectScope 的实体**：
- `EnumDefinition` — 表示项目内的枚举类型
- `ModelGroup` — 表示项目内的模型分组
- `FieldEnumAssociation` — 关联字段到枚举，确保同项目范围

---

## 错误处理

## ❌ 错误示例

```go
// Domain 依赖 Infrastructure（禁止）
// internal/domain/modeldesign/model.go
import (
    "modelcraft/internal/infrastructure/repository"  // ❌ Domain → Infrastructure
)

// Infrastructure 依赖 Application（禁止）
// internal/infrastructure/repository/sql_modeldesign_repository.go
import (
    "modelcraft/internal/app/modeldesign"  // ❌ Infrastructure → App（逆向）
)

// Application 依赖 Interfaces（禁止）
// internal/app/modeldesign/model_app.go
import (
    "modelcraft/internal/interfaces/graphql/generated"  // ❌ App → Interfaces（逆向）
)

// Interfaces 直接依赖 Infrastructure（禁止）
// internal/interfaces/graphql/model.resolvers.go
import (
    "modelcraft/internal/infrastructure/repository"  // ❌ 跳层
)
```

> lint 执行入口：`.codebuddy/rules/code-style/layered-dependency.md`

---

## 参考索引

| 主题 | 文件 |
|------|------|
| 仓储接口定义示例 | `internal/domain/modeldesign/model_repository.go` |
| Application Command 定义 | `internal/app/modeldesign/commands.go` |
| GraphQL Resolver 示例 | `internal/interfaces/graphql/model.resolvers.go` |
