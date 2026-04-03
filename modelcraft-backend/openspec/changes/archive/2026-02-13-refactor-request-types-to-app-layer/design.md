# Design: Request Types 归属层分析与重构方案

## Context

### 当前问题

在 DDD 分层架构中，依赖关系应该是：

```
Interface Layer → App Layer → Domain Layer
（外层依赖内层，内层不依赖外层）
```

但当前代码中，App 层反向依赖了 Interface 层：

```go
// internal/app/cluster/cluster_app.go
import "modelcraft/internal/interfaces/http/requests"  // ❌ App 层依赖 Interface 层

func (s *DatabaseClusterAppService) CreateCluster(ctx context.Context, req requests.CreateClusterRequest) (string, error) {
```

同时还依赖了 Interface 层的 DTO：

```go
import "modelcraft/internal/interfaces/http/dtos"  // ❌ App 层依赖 Interface 层
```

### 现状分析

| 域 | 当前模式 | 问题 |
|---|---|---|
| **Cluster** | `requests.CreateClusterRequest`（定义在 interface 层） | App 层反向依赖 Interface 层 |
| **Model** | `requests.CreateModelRequest`（定义在 interface 层） | App 层反向依赖 Interface 层 |
| **Enum** | 6 个散参数传递 | 缺乏类型安全，参数顺序易错 |
| **Project** | 5 个散参数传递 | 缺乏类型安全，参数顺序易错 |

## Goals / Non-Goals

### Goals
- 修正分层依赖方向：App 层不再 import Interface 层包
- 统一所有域的 App Service 入参模式：使用 Command 对象
- 保持向后兼容（外部 API 无变化）
- 保持代码简洁，不过度抽象

### Non-Goals
- 不引入 CQRS 框架或消息总线
- 不改变 Domain 层的实体设计
- 不修改 GraphQL schema 或对外 API 行为
- 不重构 Interface 层的 Adapter/Mapper 架构（但会更新其转换目标类型）

## Decisions

### Decision 1: Request 类型应定义在 App 层，命名为 Command

**选择**: 在 `internal/app/{context}/` 中定义 Command 结构体

**理由**:
- Command 是 App 层的"入口契约"，描述"要执行什么操作"
- 遵循依赖方向：Interface 层可以 import App 层的 Command 类型
- Command 命名符合 CQS（Command-Query Separation）惯例，语义清晰
- 避免与 GraphQL generated input 类型或 HTTP binding 类型混淆

**具体结构**:

```
internal/app/
├── cluster/
│   ├── commands.go           # CreateClusterCommand, UpdateClusterCommand
│   ├── cluster_app.go        # Service 方法接受 Command
│   └── cluster_app_test.go
├── modeldesign/
│   ├── commands.go           # CreateModelCommand, CreateEnumCommand, etc.
│   ├── model_app.go
│   └── enum_service.go
└── project/
    ├── commands.go           # CreateProjectCommand
    └── project_service.go
```

**Command 示例**:

```go
// internal/app/cluster/commands.go
package cluster

// CreateClusterCommand 创建数据库集群命令
type CreateClusterCommand struct {
    OrgName           string
    ProjectName       string
    Name              string
    Title             string
    Description       string
    Host              string
    Port              int
    Username          string
    Password          string
    ConnectionTimeout int
}

// UpdateClusterCommand 更新数据库集群命令
type UpdateClusterCommand struct {
    Name              *string
    Title             *string
    Description       *string
    Host              *string
    Port              *int
    Username          *string
    Password          *string
    ConnectionTimeout *int
    Status            *string
}
```

**注意**: Command 结构体不需要 JSON tag 或 binding tag，因为它不直接绑定 HTTP 请求。JSON 绑定是 Interface 层的职责。

### Decision 2: Interface 层负责转换

GraphQL Resolver 和 HTTP Handler 分别从各自的输入类型转换到 App 层 Command：

```
GraphQL Input (generated)  ──→  Command (app layer)  ──→  App Service
HTTP Request (binding)     ──→  Command (app layer)  ──→  App Service
```

现有的 `adapter/cluster_mapper.go` 中的 `ToCreateClusterRequest` 方法改名为 `ToCreateClusterCommand`，返回类型改为 `cluster.CreateClusterCommand`。

### Decision 3: 处理 DTO 依赖

当前 App 层还依赖了 `internal/interfaces/http/dtos.ConnectionInfo`。有两种处理方式：

- **方案 A**: 将 `ConnectionInfo` 移到 App 层或 Domain 层（如果它是业务概念）
- **方案 B**: 在 Command 中展平连接信息字段（当前 `CreateClusterCommand` 已展平了 host/port/username/password）

**选择方案 B**：Command 中已经展平了连接字段，只需在 `TestConnection` 等方法中也避免使用 interface 层的 DTO。可以在 Domain 层定义 `ConnectionInfo` 值对象，或在 App 层定义一个轻量 DTO。

### Alternatives Considered

**方案 A: 保持 Request 在 Interface 层，App 层使用散参数**
- 优点：简单，不需要新类型
- 缺点：参数多时（如 Cluster 有 10+ 字段）函数签名冗长且易错，Enum/Project 当前就是这个问题

**方案 B: 定义 shared DTO 包**
- 优点：所有层都能引用
- 缺点：引入横切依赖，shared 包容易膨胀为垃圾桶，违背 DDD 分层原则

**方案 C: 直接传 Domain Entity 给 App Service**
- 优点：减少类型数量
- 缺点：Domain Entity 有 ID、时间戳等生命周期字段，作为入参不合适；调用方不应构造完整实体

## Risks / Trade-offs

| 风险 | 缓解措施 |
|---|---|
| 重构范围大，可能引入 bug | 按域逐个迁移，每迁移一个域就运行完整测试 |
| Command 与 Request 看似重复 | Command 不含 HTTP 绑定信息，是独立的业务语义；Request 是传输协议关注点 |
| 临时存在两种模式 | 制定迁移顺序，一次性完成一个域的切换 |

## Migration Plan

采用**逐域迁移**策略，每个域独立完成：

1. **Cluster 域**（最复杂，先做作为模板）
   - 创建 `internal/app/cluster/commands.go`
   - 修改 `cluster_app.go` 使用 `CreateClusterCommand`/`UpdateClusterCommand`
   - 更新 `adapter/cluster_mapper.go` 转换目标
   - 更新 resolver
   - 运行测试验证

2. **Model 域**
   - 创建 `internal/app/modeldesign/commands.go`
   - 修改 `model_app.go`
   - 更新 resolver
   - 运行测试验证

3. **Enum 域**（当前是散参数，引入 Command）
   - 创建 Command 类型
   - 修改 `enum_service.go` 签名
   - 更新 resolver
   - 运行测试验证

4. **Project 域**（当前是散参数，引入 Command）
   - 创建 Command 类型
   - 修改 `project_service.go` 签名
   - 更新 resolver
   - 运行测试验证

5. **清理**
   - 移除或精简 `internal/interfaces/http/requests/` 中不再需要的类型
   - 确保 App 层不再 import `internal/interfaces/` 下任何包
   - 全量测试

## Open Questions

- `internal/interfaces/http/dtos/` 中的类型（如 `ConnectionInfo`）是否也应迁移到 App 层或 Domain 层？目前建议暂时保留，仅确保 App 层不直接引用。
