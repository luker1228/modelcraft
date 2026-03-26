# 日志规范

> **优先级: 高** - 定义 ModelCraft Go 后端的日志记录标准和最佳实践。

## 核心原则

- 使用 `pkg/logfacade` 包，禁止裸用 `log`
- 日志必须打印英文
- 不允许打印 debug 日志
- 日志字段 key 必须使用 `logfacade` 包定义的常量
- 打印错误日志时，必须使用结构化字段记录

## 日志字段 Key 常量

`logfacade` 包预定义了常用的字段 key 常量，必须使用这些常量而非硬编码字符串：

```go
// 错误相关
logfacade.ErrorFieldKey   // "error"
logfacade.StackFieldKey   // "stack"

// 请求相关
logfacade.RequestIDKey    // "request_id"
logfacade.MethodKey       // "method"
logfacade.PathKey         // "path"
logfacade.URLKey          // "url"
logfacade.StatusCodeKey   // "status_code"

// 数据库相关
logfacade.SQLKey          // "sql"
logfacade.RowsKey         // "rows"
logfacade.ElapsedKey      // "elapsed"

// 业务相关
logfacade.ProjectKey      // "project"
logfacade.ClusterNameKey  // "cluster_name"
logfacade.HostKey         // "host"
logfacade.PortKey         // "port"
```

## 堆栈跟踪使用规范

堆栈跟踪 (`logfacade.Stack()`) 只在以下场景使用：

### 场景 1: 错误转换时 (MUST)

当错误被转换成其他类型（如 GraphQL error、HTTP error）前，**必须**记录堆栈：

```go
// 接口层转换错误
func (r *resolver) CreateProject(ctx context.Context, input Input) (*Payload, error) {
    project, err := r.service.Create(ctx, input)
    if err != nil {
        // ✅ 转换前记录完整堆栈
        r.logger.Error("Failed to create project",
            logfacade.Err(err),       // 错误消息
            logfacade.Stack(err),     // 完整堆栈
            logfacade.String("name", input.Name),
        )
        // 错误被转换为 GraphQL 类型
        return &Payload{Error: toGraphQLError(err)}, nil
    }
    return &Payload{Project: project}, nil
}
```

### 场景 2: 顶层中间件 (MUST)

处理最终错误响应的中间件（panic recovery、全局错误处理器）：

```go
// Panic recovery 中间件
func PanicRecoveryMiddleware(logger logfacade.Logger) chi.Middlewares {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    err := pkgerrors.Errorf("panic: %v", rec)
                    // ✅ 顶层处理记录堆栈
                    logger.Error("Panic recovered",
                        logfacade.Err(err),
                        logfacade.Stack(err),
                        logfacade.String(logfacade.MethodKey, r.Method),
                        logfacade.String(logfacade.PathKey, r.URL.Path),
                    )
                    http.Error(w, "Internal Server Error", 500)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

### 场景 3: 禁止在 Service/Repository 层使用 (MUST NOT)

业务逻辑层只使用 `Err(err)`，不使用 `Stack()`：

```go
// ✅ 正确：Service 层不使用 Stack()
func (s *ProjectService) Create(ctx context.Context, input Input) (*Project, error) {
    if err := s.validator.Validate(input); err != nil {
        s.logger.Error("Validation failed",
            logfacade.Err(err),
            logfacade.String("name", input.Name),
        )
        return nil, pkgerrors.Wrap(err, "validation failed")
    }
    return project, nil
}

// ❌ 禁止：在 Service 层使用 Stack()
func (s *Service) Create(ctx context.Context) error {
    if err != nil {
        logger.Error("failed",
            logfacade.Err(err),
            logfacade.Stack(err),  // 错误！Service 层不应记录堆栈
        )
    }
}
```

### 场景 4: 必须同时使用 Err() 和 Stack() (MUST)

记录堆栈时，必须同时包含错误消息和堆栈跟踪作为独立字段：

```go
// ✅ 正确：同时使用 Err() 和 Stack()
logger.Error("Operation failed",
    logfacade.Err(err),
    logfacade.Stack(err),
)

// ❌ 禁止：单独使用 Stack() 而不使用 Err()
logger.Error("Operation failed", logfacade.Stack(err))
```

## 原理说明

堆栈跟踪成本高且冗长，应仅在关键点记录：

1. **错误转换边界**：原始错误上下文会丢失
2. **顶层处理器**：代表捕获诊断信息的最后机会

这可防止日志中出现重复堆栈跟踪，同时确保在重要位置保留完整的错误上下文。

## 示例

### ✅ 正确示例

```go
import "modelcraft/pkg/logfacade"

logger := logfacade.GetDefault()

// 基本日志
logger.Infof("operation started, modelID=%s", modelID)

// 错误日志（Service 层）
if err != nil {
    logger.Error("operation failed", logfacade.Err(err))
}

// 错误转换点（Resolver 层）
func (r *resolver) CreateProject(ctx context.Context, input Input) (*Payload, error) {
    project, err := r.service.Create(ctx, input)
    if err != nil {
        r.logger.Error("Failed to create project",
            logfacade.Err(err),
            logfacade.Stack(err),  // 转换前记录堆栈
        )
        return &Payload{Error: toGraphQLError(err)}, nil
    }
    return &Payload{Project: project}, nil
}

// 使用常量 key
logger.Info("Processing request",
    logfacade.String(logfacade.MethodKey, r.Method),
    logfacade.String(logfacade.PathKey, r.URL.Path),
    logfacade.Int(logfacade.StatusCodeKey, 200),
)

logger.Error("SQL execution failed",
    logfacade.Err(err),
    logfacade.String(logfacade.SQLKey, sql),
    logfacade.Duration(logfacade.ElapsedKey, elapsed),
)
```

### ❌ 错误示例

```go
import "log"

// 禁止：使用标准库 log
log.Printf("开始处理")
log.Println("debug: foo=%v", foo)

// 禁止：在格式化字符串中直接打印错误
logger.Errorf("operation failed: %v", err)
logger.Errorf("save model failed, err=%v", err)

// 禁止：硬编码 key 字符串
logger.Info("Processing request",
    logfacade.String("method", r.Method),      // 错误！应使用 logfacade.MethodKey
    logfacade.String("path", r.URL.Path),      // 错误！应使用 logfacade.PathKey
    logfacade.Int("status_code", 200),         // 错误！应使用 logfacade.StatusCodeKey
)

// 禁止：在 Service 层使用 Stack()
func (s *Service) Create(ctx context.Context) error {
    if err != nil {
        logger.Error("failed",
            logfacade.Err(err),
            logfacade.Stack(err),  // 错误！Service 层不应记录堆栈
        )
    }
}

// 禁止：错误转换时缺少 Stack()
func (r *resolver) CreateProject(ctx context.Context) (*Payload, error) {
    project, err := r.service.Create(ctx, input)
    if err != nil {
        logger.Error("Failed", logfacade.Err(err))  // 错误！缺少 Stack()
        return &Payload{Error: toGraphQLError(err)}, nil  // 堆栈信息丢失！
    }
}
```

## 复杂对象日志

使用 `bizutils.MarshalToStringIgnoreErr` 输出复杂对象：

```go
import "modelcraft/pkg/bizutils"

logger.Infof("input: %s", bizutils.MarshalToStringIgnoreErr(input))
```
