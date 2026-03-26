# 错误处理规范

> 本文档覆盖错误码体系设计、各层错误职责划分、RecordNotFound 处理约定。
> 代码参考优先看源文件和测试，不重复誊写实现细节。

## 错误包体系

项目有两套错误包，职责不同，**不可混用**：

| 包 | 路径 | 用途 |
|----|------|------|
| `bizerrors` | `pkg/bizerrors/` | 业务错误，跨层传递，最终暴露给客户端 |
| `shared.RepositoryError` | `internal/domain/shared/repository_error.go` | Repository 层技术错误，不暴露给客户端 |

禁止直接使用标准库 `errors`，所有通用错误包装使用 `pkg/bizerrors`（内部封装了 `github.com/pkg/errors`）：

```go
// pkg/bizerrors/errors.go
var (
    New       = pkgerrors.New
    Wrap      = pkgerrors.Wrap
    Wrapf     = pkgerrors.Wrapf
    // ...
)
```

## 错误码体系（`pkg/bizerrors`）

### 错误码格式

```
ErrorType.DOMAIN
```

- **ErrorType**（前缀）决定错误性质和 HTTP 状态码
- **DOMAIN**（后缀）标识资源归属

```go
// pkg/bizerrors/definition.go
ErrorTypeNotFound       = "NOT_FOUND"         // → HTTP 404
ErrorTypeParamInvalid   = "PARAM_INVALID"     // → HTTP 400
ErrorTypeOperationFailed = "OPERATION_FAILED" // → HTTP 403
ErrorTypeConflict       = "CONFLICT"          // → HTTP 409
ErrorTypeSystemError    = "SYSTEM_ERROR"      // → HTTP 500
```

完整的错误码→HTTP 映射和类型判断见：
→ `pkg/bizerrors/business_error.go` `GetHTTPStatusCode()`
→ `pkg/bizerrors/business_error_test.go` `TestHTTPStatusCodes`

### 已定义的业务错误码

完整列表见 `pkg/bizerrors/common_errors.go`，示例：

```
NOT_FOUND.MODEL       NOT_FOUND.PROJECT     NOT_FOUND.CLUSTER
CONFLICT.MODEL        CONFLICT.PROJECT      CONFLICT.FIELD
OPERATION_FAILED.PROJECT  OPERATION_FAILED.ENUM  OPERATION_FAILED.FK
PARAM_INVALID.GROUP   PARAM_INVALID.FK      PARAM_INVALID.FK.FIELD_COUNT
SYSTEM_ERROR
```

### 创建错误

```go
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

→ 完整用法见 `pkg/bizerrors/business_error_test.go` `TestErrorTypeJudgment`

---

## 各层错误设计

### Infrastructure 层（Repository）

**职责**：将数据库原始错误转换为 `shared.RepositoryError`，**不使用** `BusinessError`。

```
数据库错误（sqlc/sql）
    → AnalyzeSQLError()       // 分类为 RepositoryErrorType
    → shared.RepositoryError  // 结构化技术错误
```

两个核心 sentinel error，支持 `errors.Is()` 检查：

```go
// internal/domain/shared/repository_error.go
var (
    ErrRecordNotFound = errors.New("record not found")
    ErrDuplicateKey   = errors.New("duplicate key")
)
```

SQL 错误分类逻辑（按 MySQL 错误码和消息匹配）：
→ `internal/infrastructure/repository/sql_error_analyzer.go`

便捷函数：
```go
repository.IsNotFoundError(err)    // 检查是否为 NOT_FOUND
repository.IsDuplicateKeyError(err) // 检查是否为重复键
```

**规则：Repository 层返回 `error`，不返回 `*BusinessError`。**

### Application 层

**职责**：将 Repository 错误或领域校验结果转换为 `*BusinessError`，上传给接口层。

常见转换模式：

```go
// 1. 查询不存在 → 业务 NotFound
model, err := s.modelRepo.FindByName(ctx, name)
if err != nil {
    return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, name)
}
if model == nil {
    return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, name)
}

// 2. Repository 系统错误 → BusinessError SystemError
if err != nil {
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
}

// 3. 参数校验失败 → PARAM_INVALID
if input.Name == "" {
    return nil, bizerrors.NewValidationError("name is required")
}
```

→ `pkg/bizerrors/business_error.go` `ConvertRepositoryError()`

**规则：Application 层只返回 `*BusinessError` 或 `error`，不暴露 `RepositoryError`。**

#### ✅ 正确示例

```go
// internal/app/modeldesign/model_app.go
func (s *ModelDesignAppService) GetModel(ctx context.Context, id string) (*modeldesign.DataModel, error) {
    model, err := s.modelRepo.FindByID(ctx, id)
    if err != nil {
        // DB 异常 → 统一转为 SYSTEM_ERROR，不暴露技术细节
        return nil, bizerrors.ConvertRepositoryError(ctx, err)
    }
    if model == nil {
        // nil 结果在 Application 层赋予业务语义
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, id)
    }
    return model, nil
}
```

#### ❌ 错误示例

```go
func (s *ModelDesignAppService) GetModel(ctx context.Context, id string) (*modeldesign.DataModel, error) {
    model, err := s.modelRepo.FindByID(ctx, id)

    // ❌ 直接透传 RepositoryError，客户端会看到技术内部信息
    if err != nil {
        return nil, err
    }

    // ❌ 忽略 nil 检查，调用方拿到 nil 指针会 panic
    return model, nil
}
```

### Interfaces 层（GraphQL Resolver）

**职责**：将 `*BusinessError` 转换为 GraphQL 联合错误类型，记录日志（含堆栈）。

```go
// internal/interfaces/graphql/*.resolvers.go 标准模式
func (r *mutationResolver) CreateModel(ctx context.Context, input ...) (*generated.CreateModelPayload, error) {
    result, err := r.modelService.CreateModel(ctx, cmd)
    if err != nil {
        if bizErr, ok := err.(*bizerrors.BusinessError); ok {
            // 错误转换前，记录完整堆栈（MUST）
            r.logger.Error("Failed to create model",
                logfacade.Err(err),
                logfacade.Stack(err),
            )
            adapter := adapter.NewModelErrorAdapter(ctx)
            return &generated.CreateModelPayload{
                Error: adapter.ConvertToCreateError(bizErr),
            }, nil  // 返回 nil，错误放入 payload
        }
        return nil, err  // 非 BusinessError 的系统异常
    }
    return &generated.CreateModelPayload{Model: toGQL(result)}, nil
}
```

错误适配器将错误码 switch 映射到 GraphQL 联合类型：
→ `internal/interfaces/graphql/adapter/model_error_adapter.go`

**规则：接口层是唯一记录 `logfacade.Stack()` 的地方（错误转换点）。**

---

## RecordNotFound 处理

这是最容易出错的地方，规则如下：

### Repository 层：返回 nil，不返回 error

```go
// ✅ 正确：记录不存在返回 (nil, nil)
func (r *sqlModelRepo) FindByName(ctx context.Context, name string) (*DataModel, error) {
    var row dbgen.Model
    err := r.db.Where("name = ?", name).First(&row).Error
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil  // 不存在 → nil，由调用方决定语义
    }
    if err != nil {
        return nil, repository.WrapSQLError(err)
    }
    return toDomain(row), nil
}
```

#### ❌ 错误示例

```go
func (r *sqlModelRepo) FindByName(ctx context.Context, name string) (*DataModel, error) {
    var row dbgen.Model
    err := r.db.Where("name = ?", name).First(&row).Error

    // ❌ 在 Repository 层直接返回 BusinessError，越层处理业务语义
    if errors.Is(err, sql.ErrNoRows) {
        return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, name)
    }
    return toDomain(row), nil
}
```

`sql.ErrNoRows` 同理（见 `sql_error_analyzer.go`：`AnalyzeSQLError` 对 `sql.ErrNoRows` 返回 `nil`）。

sqlc logger 也配置了忽略 `RecordNotFound` 日志噪音：
→ `internal/infrastructure/repository/sql_error_analyzer.go`

### Application 层：nil 结果转换为 BusinessError

```go
// ✅ 正确：检查 nil 并转换语义
model, err := s.modelRepo.FindByName(ctx, name)
if err != nil {
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
}
if model == nil {
    // nil 结果在这里才转换为业务错误
    return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, name)
}
```

### 判断流程

```
Repository.Find()
    ├─ sql.ErrNoRows → return (nil, nil)
    ├─ 其他 DB 错误           → return (nil, RepositoryError)
    └─ 成功                   → return (entity, nil)

App.UseCase()
    ├─ err != nil            → ConvertRepositoryError → BusinessError(SYSTEM_ERROR)
    ├─ entity == nil         → NewErrorFromContext    → BusinessError(NOT_FOUND.XXX)
    └─ 成功                  → return entity
```

**禁止在 Repository 层直接返回 `bizerrors.ModelNotFound`。**

---

## 日志与堆栈规则

| 层 | 使用方式 | 是否用 Stack() |
|----|----------|----------------|
| Repository | 不打印错误日志（返回 error 即可） | 否 |
| Application | `logger.Error(..., logfacade.Err(err))` | 否 |
| Interfaces (错误转换点) | `logger.Error(..., logfacade.Err(err), logfacade.Stack(err))` | **是** |
| 顶层中间件 (panic recovery) | `logger.Error(..., logfacade.Err(err), logfacade.Stack(err))` | **是** |

→ 完整日志规则见 `.codebuddy/rules/code-style/coding-style.md`

---

## 参考索引

| 主题 | 文件 |
|------|------|
| 错误码定义 | `pkg/bizerrors/common_errors.go` |
| BusinessError 结构 | `pkg/bizerrors/business_error.go` |
| 错误类型/HTTP 映射测试 | `pkg/bizerrors/business_error_test.go` |
| RepositoryError / Sentinel | `internal/domain/shared/repository_error.go` |
| SQL 错误分类 | `internal/infrastructure/repository/sql_error_analyzer.go` |
| Repository 层辅助函数 | `internal/infrastructure/repository/error_helper.go` |
| GraphQL 错误适配器 | `internal/interfaces/graphql/adapter/*_error_adapter.go` |
| GraphQL 设计模式 | `.codebuddy/rules/api-design/graphql-patterns.md` |
