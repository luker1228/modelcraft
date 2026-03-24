---
name: coding-standards
description: Go 语言通用编码规范、最佳实践与工程化指导。
globs: **/*.go
---

# Go 编码规范与最佳实践

面向 ModelCraft Go 后端的编码标准、架构分层、错误处理与可维护性指导。
Refer to @modelcraft-go/ai-metadata/2-development/ for detailed specifications.

## 代码质量原则

### 1. 可读性优先
- 代码是给人读的
- 命名清晰、结构清楚
- 适度注释解释"为什么"

### 2. KISS
- 先用最简单可行方案
- 避免过度设计与提前优化

### 3. DRY
- 提取复用逻辑
- 避免复制粘贴

### 4. YAGNI
- 不做未被证明需要的泛化

## 架构分层（DDD）

Refer to @modelcraft-go/ai-metadata/2-development/architecture.md for architecture details.

```
┌─────────────────────────────────────────┐
│    Interfaces 接口层 (graphql/http)      │  ← 请求解析、响应格式化
├─────────────────────────────────────────┤
│    Application 应用层 (internal/app/)    │  ← 用例编排、事务管理
├──────────────────┬──────────────────────┤
│  Domain 领域层   │  Infrastructure       │
│ (internal/domain)│ (internal/infrastructure) │
├──────────────────┴──────────────────────┤
│    Shared Kernel (pkg/)                 │
└─────────────────────────────────────────┘
```

### 依赖方向（单向，不可逆）

```
Interfaces → Application → Domain
                       ↘ Infrastructure → Domain
                                      ↘ pkg/
```

### 各层依赖规则

| 层 | 可依赖 | 禁止依赖 |
|----|--------|----------|
| Interfaces | Application、同层 | Infrastructure、Domain（直接） |
| Application | Domain、Infrastructure | Interfaces |
| Infrastructure | Domain、pkg/ | Application、Interfaces |
| Domain | **仅 pkg/** | 所有 internal 层 |
| pkg/ | 无 | 所有 internal 层 |

### 命名约定

- **包名**: 小写，单个单词，避免 stutter
- **接口**: 名词形式（如 `ModelRepository`、`QueryExecutor`）
- **领域服务**: 行为导向（如 `ModelService`）
- **应用服务**: 用例导向（如 `CreateModelUseCase`）
- **接收者命名**: `s`、`r`、`h`
- **常量用名词，错误用 `ErrXxx`，错误信息小写不带句号**

### 包命名

```go
// ✅ GOOD
package auth

type Service struct {}

// ❌ BAD
package authservice

type AuthService struct {}
```

## 错误处理（项目规范）

Refer to @modelcraft-go/ai-metadata/2-development/error-handling.md for error handling details.

### 错误包体系

项目有两套错误包，职责不同，**不可混用**：

| 包 | 路径 | 用途 |
|----|------|------|
| `bizerrors` | `pkg/bizerrors/` | 业务错误，跨层传递，暴露给客户端 |
| `shared.RepositoryError` | `internal/domain/shared/repository_error.go` | Repository 层技术错误，不暴露给客户端 |

**禁止直接使用标准库 `errors`**，所有通用错误包装使用 `pkg/bizerrors`。

### 错误码体系

```
ErrorType.DOMAIN
```

| ErrorType | HTTP 状态码 | 示例 |
|-----------|------------|------|
| `NOT_FOUND` | 404 | `NOT_FOUND.MODEL`、`NOT_FOUND.PROJECT` |
| `PARAM_INVALID` | 400 | `PARAM_INVALID.GROUP`、`PARAM_INVALID.FK` |
| `OPERATION_FAILED` | 403 | `OPERATION_FAILED.PROJECT`、`OPERATION_FAILED.ENUM` |
| `CONFLICT` | 409 | `CONFLICT.MODEL`、`CONFLICT.FIELD` |
| `SYSTEM_ERROR` | 500 | `SYSTEM_ERROR` |

### 创建错误

```go
import "modelcraft/pkg/bizerrors"

// 普通场景
bizerrors.NewError(bizerrors.ModelNotFound, modelID)

// 携带请求上下文（推荐，自动提取 requestId 和语言）
bizerrors.NewErrorFromContext(ctx, bizerrors.ClusterNotFound, projectSlug)

// 包装底层错误（保留错误链）
bizerrors.WrapError(err, bizerrors.SystemError, detail)
```

### 错误类型判断

```go
if bizErr.IsNotFoundError()     { ... }
if bizErr.IsConflictError()     { ... }
if bizErr.IsParamInvalidError() { ... }
```

### 各层错误职责

**Infrastructure 层（Repository）**：
- 返回 `shared.RepositoryError`，**不返回** `*BusinessError`
- RecordNotFound → 返回 `(nil, nil)`（模式 B）或 `shared.NewNotFoundError`（模式 A）
- 使用 `ExecWithErrorHandling` / `QueryWithSQLErrorHandling` 包装 DB 操作
- **不打印错误日志**

**Application 层**：
- 将 Repository error 转换为 `*BusinessError`
- nil 结果在此层赋予业务语义：`bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, id)`
- DB 异常统一转：`bizerrors.ConvertRepositoryError(ctx, err)`
- 日志用 `logger.Error(..., logfacade.Err(err))`，**不用 Stack()**

**Interfaces 层（GraphQL Resolver）**：
- 将 `*BusinessError` 通过 `adapter/*_error_adapter.go` 转为 GraphQL 联合错误
- **错误转换前必须记录 `logfacade.Stack(err)`**（唯一允许打堆栈的地方）
- Mutation payload 中错误字段可为 nil，数据字段可为 nil，**不混用**

### RecordNotFound 两种模式

| 场景 | 返回值 | 不存在时 | 示例方法 |
|------|--------|----------|----------|
| 必须存在的记录 | `(value, error)` | 返回 `NotFoundError` | `GetByID`, `GetByName` |
| 可能不存在的查询 | `(value, bool, error)` | 返回 `(value, false, nil)` | `FindIDByExternalID` |

**禁止在 Repository 层直接返回 `bizerrors.ModelNotFound`。**

### 判断流程

```
Repository.Find()
    ├─ sql.ErrNoRows → return (nil, nil) 或 (nil, NotFoundError)
    ├─ 其他 DB 错误           → return (nil, RepositoryError)
    └─ 成功                   → return (entity, nil)

App.UseCase()
    ├─ err != nil            → ConvertRepositoryError → BusinessError(SYSTEM_ERROR)
    ├─ entity == nil         → NewErrorFromContext    → BusinessError(NOT_FOUND.XXX)
    └─ 成功                  → return entity
```

### 日志与堆栈规则

| 层 | 使用方式 | 是否用 Stack() |
|----|----------|----------------|
| Repository | 不打印错误日志 | 否 |
| Application | `logger.Error(..., logfacade.Err(err))` | 否 |
| Interfaces (错误转换点) | `logger.Error(..., logfacade.Err(err), logfacade.Stack(err))` | **是** |

## Repository 层规范

Refer to @modelcraft-go/ai-metadata/2-development/repo-develop.md for repository development details. Triggered by globs: `internal/infrastructure/**/*.go`

### 核心约束

- 接收 `querier` 接口（`dbgen.Querier`），支持事务和非事务
- 使用 `ExecWithErrorHandling` / `QueryWithSQLErrorHandling` 包装 DB 操作
- 返回 `shared.RepositoryError`，不返回 `*BusinessError`
- 使用辅助函数处理 `sql.Null*` 类型（`NullStrToPtr`、`PtrToNullStr` 等）
- 不在 Repository 层开启事务（事务由 Application 层管理）

## 日志规范

- 使用 `logfacade`，默认用 `Infof`/`Errorf`
- 禁止裸用 `log`，统一使用 `logfacade` 包
- 结构化字段只允许用于错误或 `pkg/logfacade/constant.go` 中的常量
- **所有关键分支必须打日志**（if/else、switch）
- 复杂对象用 `bizutils.MarshalToStringIgnoreErr` 输出

```go
import "modelcraft/pkg/logfacade"
import "modelcraft/pkg/bizutils"

logger := logfacade.GetDefault()
logger.Infof("operation started, modelID=%s", modelID)
logger.Errorf("operation failed, err=%v", err)
logger.Error("operation failed", logfacade.Err(err))
logger.Infof("input: %s", bizutils.MarshalToStringIgnoreErr(input))
```

## Context 工具（ctxutils）

```go
import "modelcraft/pkg/ctxutils"

// 用户 ID
userID, err := ctxutils.GetUserIDFromContext(ctx)

// 租户/组织名
orgName, err := ctxutils.GetOrgNameFromContext(ctx)

// 权限列表（格式: "resource:action"）
permissions, err := ctxutils.GetPermissionsFromContext(ctx)

// HTTP 请求元数据
requestID := ctxutils.GetRequestID(ctx)
```

**Getter 函数在值缺失时返回 `error`，必须检查错误。**

## 并发与上下文

### Context 规范
- 所有 I/O、RPC、DB 调用必须传 `context.Context`
- 入口函数尽早创建 `ctx`

### Goroutine 规范
- **禁止裸用 `go func`，必须使用 `bizutils.GoWithCtx`**

```go
import "modelcraft/pkg/bizutils"

// ✅ 正确：携带上下文、自动捕获 panic
bizutils.GoWithCtx(ctx, func(ctx context.Context) {
    // 异步逻辑，ctx 已从父 goroutine 传递
})

// ❌ 错误：裸 goroutine
go func() { /* panic 会导致进程崩溃 */ }()
```

- 需要退出条件与取消信号
- 生产者/消费者使用 `context` 或 `channel`
- 共享状态用 `sync.Mutex` 或 `sync/atomic`

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
```

## 事务规范

Refer to @modelcraft-go/ai-metadata/2-development/code-style.md for transaction patterns.

| 原则 | 说明 |
|------|------|
| **应用层管理** | 事务在 UseCase/Application Service 层开启和提交 |
| **Repository 无感知** | Repository 接收 `dbgen.Querier`，不关心是否在事务中 |
| **defer 保证回滚** | 使用 `defer tx.Rollback()` 确保异常时回滚 |
| **显式提交** | 只有成功时才调用 `tx.Commit()` |

```go
// ✅ 正确：defer + 显式 commit
func (uc *UseCase) Execute(ctx context.Context) error {
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil { return err }
    defer tx.Rollback()

    if err := step1(); err != nil { return err }
    if err := step2(); err != nil { return err }
    return tx.Commit()
}
```

## 数据库（sqlc）

- 所有数据库操作使用 **sqlc** 生成类型安全的 Go 代码
- SQL 查询定义在 `db/queries/*.sql`
- 生成的代码位于 `internal/infrastructure/dbgen/`
- **禁止使用 ORM**
- **禁止运行 `task regenerate-gql`**（会删除自定义 resolver 实现）
- 修改 `.graphql` 文件后运行 `task generate-gql`

## 三条 API 通道

| 通道 | Schema 来源 | 用途 |
|------|-------------|------|
| 设计时 GraphQL | `api/graph/schema/*.graphql` | 模型/字段/枚举/项目/集群 CRUD |
| REST (OpenAPI) | `api/openapi/*.yaml` | 认证、组织管理、Webhook |
| 运行时 GraphQL | 运行时动态生成 | 用户数据查询/变更 |

**业务域功能只走设计时 GraphQL，禁止添加到 REST API。**

## 性能与内存

- 预分配切片容量
- 避免在热路径中频繁分配
- 小对象用值，大对象用指针

```go
items := make([]Item, 0, 64)
```

## 接口与抽象

- 接口要小（`io.Reader` 风格）
- 面向调用方定义接口
- 值对象可通过 Go 结构体嵌入复用（如 `ProjectScope`）

## 测试规范

- 核心业务必须有单测
- 使用表驱动测试
- 断言清晰、覆盖边界

```go
func TestParse(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"normal", "a", "a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Parse(tc.input)
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}
```

## 安全编码

- 所有 SQL 必须使用参数化查询（sqlc 自动处理）
- 动态 `order by`/`group by` 必须白名单校验
- 不允许硬编码密钥或凭据

## 代码格式化

- 使用 **gofumpt** 格式化代码（通过 `just lint-fix`）
- 不使用 `panic` 处理错误
- 接口定义在使用方（非实现方）
- 优先组合而非继承


## 参考文档

| 主题 | 文件 |
|------|------|
| 架构分层详解 | Refer to @modelcraft-go/ai-metadata/2-development/architecture.md |
| 代码风格规范 | Refer to @modelcraft-go/ai-metadata/2-development/code-style.md |
| 错误处理规范 | Refer to @modelcraft-go/ai-metadata/2-development/error-handling.md |
| Repository 层开发规范 | Refer to @modelcraft-go/ai-metadata/2-development/repo-develop.md |
| 错误码定义 | @modelcraft-go/pkg/bizerrors/common_errors.go |
| BusinessError 结构 | @modelcraft-go/pkg/bizerrors/business_error.go |
| RepositoryError / Sentinel | @modelcraft-go/internal/domain/shared/repository_error.go |
| SQL 错误分类 | @modelcraft-go/internal/infrastructure/repository/sql_error_analyzer.go |
| GraphQL 错误适配器 | @modelcraft-go/internal/interfaces/graphql/adapter/*_error_adapter.go |
