---
paths:
  - "**/*.go"
---
# 禁止项规则

以下为强制禁止项，违反即视为不符合规范。

补充规范请参考：See skill: `coding-standards` for comprehensive Go idioms and patterns.

## 日志
- 禁止裸用 `log`，必须使用 `logfacade` 包
- 日志必须打印英文
- 不允许打印 debug 日志
- **日志字段 key 必须使用 `logfacade` 包定义的常量，禁止硬编码字符串**
- **打印错误日志时，必须使用 `logger.Error(..., logfacade.Err(err))` 进行结构化记录**
- **堆栈跟踪 (`logfacade.Stack()`) 只在以下场景使用：**
  1. **错误转换时**（MUST）：当错误被转换成其他类型（如 GraphQL error、HTTP error）前，必须记录堆栈
  2. **顶层中间件**（MUST）：panic recovery、全局错误处理等最终错误处理点
  3. **禁止在 service/repository 层使用 `Stack()`**（MUST NOT）：业务逻辑层只使用 `Err(err)`
  4. **必须同时使用 `Err()` 和 `Stack()`**（MUST）：记录堆栈时，必须同时包含错误消息和堆栈跟踪作为独立字段

### 日志字段 Key 常量

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

✅ 允许：
```go
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

❌ 禁止：
```go
// 禁止：硬编码 key 字符串
logger.Info("Processing request",
    logfacade.String("method", r.Method),      // 错误！应使用 logfacade.MethodKey
    logfacade.String("path", r.URL.Path),      // 错误！应使用 logfacade.PathKey
    logfacade.Int("status_code", 200),         // 错误！应使用 logfacade.StatusCodeKey
)
```

### 什么时候使用堆栈跟踪

#### 场景 1: 错误转换点

当错误被捕获并转换为不同的错误类型时（如 business error → GraphQL error，domain error → HTTP error）：

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

#### 场景 2: 顶层中间件

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
                        logfacade.String("method", r.Method),
                        logfacade.String("path", r.URL.Path),
                    )
                    http.Error(w, "Internal Server Error", 500)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

### 示例

✅ 允许：
```go
import "modelcraft/pkg/logfacade"

logger := logfacade.GetDefault()
logger.Infof("operation started, modelID=%s", modelID)

// 正确：Service 层使用 Err() 记录错误
if err != nil {
    logger.Error("operation failed", logfacade.Err(err))
}

// 正确：错误转换点使用 Err() + Stack()
func (r *resolver) CreateProject(ctx context.Context, input Input) (*Payload, error) {
    project, err := r.service.Create(ctx, input)
    if err != nil {
        // 错误即将转换为 GraphQL error，记录完整堆栈
        r.logger.Error("Failed to create project",
            logfacade.Err(err),
            logfacade.Stack(err),  // 转换前记录堆栈
        )
        return &Payload{Error: toGraphQLError(err)}, nil
    }
    return &Payload{Project: project}, nil
}

// 正确：顶层中间件记录堆栈
func PanicRecovery(logger logfacade.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    err := pkgerrors.Errorf("panic: %v", rec)
                    logger.Error("Panic recovered",
                        logfacade.Err(err),
                        logfacade.Stack(err),  // 顶层处理记录堆栈
                    )
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}

// 正确：Service 层不使用 Stack()
func (s *ProjectService) Create(ctx context.Context, input Input) (*Project, error) {
    if err := s.validator.Validate(input); err != nil {
        // ✅ Service 层只用 Err()
        s.logger.Error("Validation failed",
            logfacade.Err(err),
            logfacade.String("name", input.Name),
        )
        return nil, pkgerrors.Wrap(err, "validation failed")
    }
    return project, nil
}
```

❌ 禁止：
```go
import "log"

log.Printf("开始处理")
log.Println("debug: foo=%v", foo)

// 禁止：在格式化字符串中直接打印错误
logger.Errorf("operation failed: %v", err)  // 错误！
logger.Errorf("save model failed, err=%v", err)  // 错误！

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

// 禁止：单独使用 Stack() 而不使用 Err()
logger.Error("Operation failed", logfacade.Stack(err))  // 错误！

// ✅ 正确：同时使用 Err() 和 Stack()
logger.Error("Operation failed",
    logfacade.Err(err),
    logfacade.Stack(err),
)
```

### 原理

堆栈跟踪成本高且冗长，应仅在关键点记录：(1) 错误转换边界，原始错误上下文会丢失；(2) 顶层处理器，代表捕获诊断信息的最后机会。这可防止日志中出现重复堆栈跟踪，同时确保在重要位置保留完整的错误上下文。

## 错误
- 禁止裸用 `errors`，必须使用 `pkg/bizerrors`

### 示例

✅ 允许：
```go
import pkgerrors "modelcraft/pkg/bizerrors"

if err != nil {
	return nil, pkgerrors.Wrapf(err, "get user %s", id)
}
```

❌ 禁止：
```go
import "errors"

return nil, errors.New("invalid input")
```

## 并发调用
- 禁止裸用 `go func`，必须使用 `bizutils.GoWithCtx`

### 示例

✅ 允许：
```go
import "modelcraft/pkg/bizutils"

bizutils.GoWithCtx(ctx, func(ctx context.Context) {
	// do work
})
```

❌ 禁止：
```go
go func() {
	// do work
}()
```

## 注释
- **禁止使用行尾注释**，注释必须独占一行或多行
- **所有导出的（Exported）标识符必须写注释**（符合 Go 官方规范）
  - 导出的包、类型、常量、变量、函数、方法必须有文档注释
  - 注释必须以标识符名称开头，形成完整句子
  - 注释应该说明"是什么"和"为什么"，而不是"怎么做"

### 示例

✅ 允许：
```go
// 检查用户权限
if !hasPermission {
	return errors.New("permission denied")
}

// MaxRetryCount 表示最大重试次数
const MaxRetryCount = 3

// User 表示系统用户
type User struct {
	ID   string
	Name string
}

// Create creates a new user with the given name and returns the user ID.
// It returns an error if the name is empty or already exists.
func (s *UserService) Create(ctx context.Context, name string) (string, error) {
	// 实现代码
}

// ProjectRepository defines the interface for project data access.
// Implementations should ensure thread-safety for concurrent operations.
type ProjectRepository interface {
	// FindByID retrieves a project by its unique identifier.
	// It returns nil if the project is not found.
	FindByID(ctx context.Context, id string) (*Project, error)
}
```

❌ 禁止：
```go
if !hasPermission { // 检查权限
	return errors.New("permission denied")
}

const MaxRetryCount = 3 // 最大重试次数

type User struct {
	ID   string // 用户ID
	Name string // 用户名称
}

// 缺少注释
func (s *UserService) Create(ctx context.Context, name string) (string, error) {
	// 实现代码
}

// 注释没有以标识符名称开头
// This interface is for project repository
type ProjectRepository interface {
	FindByID(ctx context.Context, id string) (*Project, error)
}

// 注释过于简单，只重复了代码
// Create creates
func (s *UserService) Create(ctx context.Context, name string) (string, error) {
	// 实现代码
}
```

### 注释规范说明

1. **包注释**：每个包都应该有包注释，位于 package 语句之前
   ```go
   // Package modeldesign provides domain models and services for design-time model management.
   // It implements the core business logic for creating, updating, and managing models.
   package modeldesign
   ```

2. **类型注释**：导出的类型必须有注释，说明类型的用途
   ```go
   // ModelService handles business logic for model management.
   // It orchestrates operations between repositories and domain services.
   type ModelService struct {
       repo ModelRepository
   }
   ```

3. **函数/方法注释**：说明功能、参数含义、返回值、可能的错误
   ```go
   // FindByName retrieves a model by its name within the specified project.
   // It returns nil if no model is found with the given name.
   // An error is returned only for system failures, not for "not found" cases.
   func (r *ModelRepository) FindByName(ctx context.Context, projectID, name string) (*Model, error)
   ```

4. **常量/变量注释**：说明用途和约束
   ```go
   // DefaultTimeout is the default timeout duration for database operations.
   const DefaultTimeout = 30 * time.Second

   // ErrModelNotFound is returned when a requested model does not exist.
   var ErrModelNotFound = errors.New("model not found")
   ```

注释应该放在被注释代码的上方，用完整的句子描述代码意图，而不是简单地重复代码逻辑。

## 类型转换
- **禁止使用 Go 原生类型断言 `x.(T)`**，必须使用 `github.com/spf13/cast` 包进行类型转换
- 需要区分"值缺失"与"类型错误"时，使用带 `E` 后缀的函数（返回 error）；只关心转换结果时，使用不带 `E` 的函数（零值降级）

### 规则

| 场景 | 使用方式 |
|------|----------|
| 必须有值，类型错误需报错 | `cast.ToStringE(v)` + `bizerrors.Wrap(err, ...)` |
| 可选值，类型错误降级为零值 | `cast.ToString(v)`，再判断空值 |

### 示例

✅ 允许：
```go
import "github.com/spf13/cast"

// 必填字段：类型错误需报错
schemaType, err := cast.ToStringE(schemaMap["type"])
if err != nil {
    return nil, bizerrors.Wrap(err, "invalid type")
}

// 必填字段：值为空也需报错
title := cast.ToString(schemaMap["title"])
if title == "" {
    return nil, bizerrors.New("required metadata 'title' is missing")
}

// 可选字段：类型错误静默降级为零值
description := cast.ToString(schemaMap["description"])

// 数值类型
port, err := cast.ToIntE(config["port"])
if err != nil {
    return nil, bizerrors.Wrap(err, "invalid port")
}
```

❌ 禁止：
```go
// 禁止：原生类型断言，panic 风险
schemaType := schemaMap["type"].(string)

// 禁止：ok 模式虽然安全，但代码冗长，且项目统一用 cast
schemaType, ok := schemaMap["type"].(string)
if !ok {
    return nil, bizerrors.New("invalid type")
}
```

## Context 处理
- **禁止直接使用 `context.Value()` 提取值**，必须统一使用 `pkg/ctxutils` 包
- **Interface 层**：使用 `pkg/ctxutils` 提取 context 值，作为显式参数传递给下层
- **Application/Domain 层**：接收 orgName、userID 等作为**显式参数**，禁止从 context 提取❌ 禁止：

✅ 允许：
```go
// internal/interfaces/graphql/*.resolvers.go
import "modelcraft/pkg/ctxutils"

func (r *mutationResolver) CreateModel(ctx context.Context, input Input) (*Payload, error) {
	// ✅ 使用 pkg/ctxutils 提取 context
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("organization context not found: %w", err)
	}
```
禁止：
```go
// ❌ 错误 1：直接使用 context.Value() 提取
func (r *mutationResolver) CreateModel(ctx context.Context, input Input) (*Payload, error) {
	orgName := ctx.Value("org_name").(string)  // 禁止！必须用 ctxutils
	...
}
```

## sqlc 自定义类型
- **JSON 类型字段必须使用 `db:"type:json"` 标签**，不可省略
- 自定义类型必须实现 `sql.Scanner` 和 `driver.Valuer` 接口
- `Scan` 方法必须先检查 `fmt.Stringer` 接口，再处理标准类型

### 示例

✅ 允许：
```go
// StringSlice 自定义类型用于处理字符串数组
type StringSlice []string

// Value 实现 driver.Valuer 接口
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var bytes []byte

	// 参考 sqlc 官方 datatypes 包的做法，先检查 fmt.Stringer 接口
	if str, ok := value.(fmt.Stringer); ok {
		bytes = []byte(str.String())
	} else {
		switch v := value.(type) {
		case []byte:
			if len(v) == 0 {
				*s = nil
				return nil
			}
			bytes = v
		case string:
			if v == "" {
				*s = nil
				return nil
			}
			bytes = []byte(v)
		default:
			return fmt.Errorf("cannot scan %T into StringSlice", value)
		}
	}

	if len(bytes) == 0 {
		*s = nil
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// ModelRelationPO 结构体定义 - JSON 字段必须有 type 标签
type ModelRelationPO struct {
	ID                 string      `db:"primaryKey" json:"id"`
	ModelId            string      `json:"modelId"`
	Name               string      `json:"name"`
	RelationType       string      `json:"relation_type"`
	ModelName          string      `json:"modelName"`
	SourceFields       StringSlice `db:"type:json" json:"source_fields"`       // ✅ 必须有 db:"type:json"
	TargetFields       StringSlice `db:"type:json" json:"target_fields"`       // ✅ 必须有 db:"type:json"`
	ThroughTable       *string     `json:"through_table"`
	CreatedAt          time.Time   `db:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time   `db:"autoUpdateTime" json:"updated_at"`
}
```

❌ 禁止：
```go
// ❌ 错误 1：JSON 字段缺少 db:"type:json" 标签
type ModelRelationPO struct {
	SourceFields StringSlice `json:"source_fields"` // 禁止！必须添加 db:"type:json"
}

// ❌ 错误 2：Scan 方法没有先检查 fmt.Stringer 接口
func (s *StringSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:  // 禁止！应该先检查 fmt.Stringer
		return json.Unmarshal(v, s)
	}
}

// ❌ 错误 3：Scan 方法处理了非标准类型（应只处理 []byte, string）
func (s *StringSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case []interface{}:  // 禁止！不是 sqlc 标准做法
		for _, item := range v {
			*s = append(*s, item.(string))
		}
	}
}
```

### 原理

1. **`db:"type:json"` 标签**：明确告诉 sqlc 字段类型，确保 MySQL 驱动返回正确的数据类型，避免 `&[]` 等异常类型
2. **`fmt.Stringer` 接口**：sqlc 官方 `datatypes` 包推荐做法，处理未知类型时先尝试调用 `String()` 方法
3. **只处理标准类型**：`[]byte` 和 `string` 是数据库驱动返回的标准类型，其他类型应返回错误而非尝试转换

## sqlc 常见坑

### 1. 禁止使用 `db.Model`

`db.Model` 内置 `DeletedAt` 字段，会自动启用软删除，可能导致查询结果不符合预期。

✅ 允许：
```go
type UserPO struct {
    ID        string    `db:"primaryKey" json:"id"`
    CreatedAt time.Time `db:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `db:"autoUpdateTime" json:"updated_at"`
}
```

❌ 禁止：
```go
type UserPO struct {
    db.Model  // 禁止！会自动启用软删除
    Name string
}
```

### 2. struct 更新零值问题

使用 struct 进行 `Updates()` 时，零值字段（`""`, `0`, `false`, `nil`）**不会被更新**。

✅ 允许：
```go
// 方式1：使用 map 更新零值
db.Model(&user).Updates(map[string]interface{}{
    "name":   "",
    "status": 0,
    "active": false,
})

// 方式2：使用 Select 明确指定字段
db.Model(&user).Select("Name", "Status", "Active").Updates(user)
```

❌ 禁止：
```go
// 零值字段不会被更新！
user.Name = ""
user.Status = 0
user.Active = false
db.Model(&user).Updates(user)  // 禁止！Name/Status/Active 不会被更新
```

### 3. 预加载避免 N+1 查询

查询关联数据时必须使用 `Preload`，否则会产生 N+1 查询问题。

✅ 允许：
```go
// 预加载关联数据
var models []ModelPO
db.Preload("Fields").Preload("Fields.Relation").Find(&models)
```

❌ 禁止：
```go
// N+1 查询问题
var models []ModelPO
db.Find(&models)
for _, m := range models {
    db.Where("model_id = ?", m.ID).Find(&m.Fields)  // 禁止！每次循环都查询
}
```

### 4. 外键定义格式

外键定义时，`foreignKey` 是当前模型的字段，`references` 是关联模型的字段。

```go
type FieldDefinitionPO struct {
    ModelID string           `db:"primaryKey"`
    Name    string           `db:"primaryKey"`
    // foreignKey: 当前模型用于关联的字段
    // references: 关联模型中对应的字段
    Relation *ModelRelationPO `db:"foreignKey:model_id,name;references:model_id,name"`
}

type ModelRelationPO struct {
    ModelID   string `db:"column:model_id"`   // 对应 references 中的 model_id
    ModelName string `db:"column:model_name"` // 对应 references 中的 name
}
```

### 5. 查询条件不要用 map

`Where(map)` 的行为可能不符合预期，建议使用明确条件。

✅ 允许：
```go
db.Where("org_name = ? AND project_slug = ?", orgName, projectSlug).Find(&models)
```

❌ 禁止：
```go
db.Where(map[string]interface{}{
    "org_name": orgName,
    "project_slug": projectSlug,
}).Find(&models)  // 禁止！行为可能不符合预期
```


