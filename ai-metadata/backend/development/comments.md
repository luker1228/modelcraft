# 注释规范

> **优先级: 高** - 定义 ModelCraft Go 后端的代码注释标准。

## 核心原则

- **禁止使用行尾注释**，注释必须独占一行或多行
- **所有导出的（Exported）标识符必须写注释**（符合 Go 官方规范）
  - 导出的包、类型、常量、变量、函数、方法必须有文档注释
  - 注释必须以标识符名称开头，形成完整句子
  - 注释应该说明"是什么"和"为什么"，而不是"怎么做"

## 正确示例

### 包注释

每个包都应该有包注释，位于 package 语句之前：

```go
// Package modeldesign provides domain models and services for design-time model management.
// It implements the core business logic for creating, updating, and managing models.
package modeldesign
```

### 类型注释

导出的类型必须有注释，说明类型的用途：

```go
// ModelService handles business logic for model management.
// It orchestrates operations between repositories and domain services.
type ModelService struct {
    repo ModelRepository
}

// User 表示系统用户
type User struct {
    ID   string
    Name string
}
```

### 函数/方法注释

说明功能、参数含义、返回值、可能的错误：

```go
// Create creates a new user with the given name and returns the user ID.
// It returns an error if the name is empty or already exists.
func (s *UserService) Create(ctx context.Context, name string) (string, error) {
    // 实现代码
}

// FindByName retrieves a model by its name within the specified project.
// It returns nil if no model is found with the given name.
// An error is returned only for system failures, not for "not found" cases.
func (r *ModelRepository) FindByName(ctx context.Context, projectID, name string) (*Model, error) {
    // 实现代码
}
```

### 常量/变量注释

说明用途和约束：

```go
// MaxRetryCount 表示最大重试次数
const MaxRetryCount = 3

// DefaultTimeout is the default timeout duration for database operations.
const DefaultTimeout = 30 * time.Second

// ErrModelNotFound is returned when a requested model does not exist.
var ErrModelNotFound = errors.New("model not found")
```

### 接口注释

```go
// ProjectRepository defines the interface for project data access.
// Implementations should ensure thread-safety for concurrent operations.
type ProjectRepository interface {
    // FindByID retrieves a project by its unique identifier.
    // It returns nil if the project is not found.
    FindByID(ctx context.Context, id string) (*Project, error)
}
```

### 行内注释（内部逻辑）

```go
// 检查用户权限
if !hasPermission {
    return errors.New("permission denied")
}
```

## 错误示例

### ❌ 错误 1：行尾注释

```go
if !hasPermission { // 检查权限
    return errors.New("permission denied")
}

const MaxRetryCount = 3 // 最大重试次数

type User struct {
    ID   string // 用户ID
    Name string // 用户名称
}
```

### ❌ 错误 2：缺少注释

```go
// 导出的函数缺少注释
func (s *UserService) Create(ctx context.Context, name string) (string, error) {
    // 实现代码
}
```

### ❌ 错误 3：注释没有以标识符名称开头

```go
// This interface is for project repository
type ProjectRepository interface {
    FindByID(ctx context.Context, id string) (*Project, error)
}
```

### ❌ 错误 4：注释过于简单，只重复了代码

```go
// Create creates
func (s *UserService) Create(ctx context.Context, name string) (string, error) {
    // 实现代码
}
```

## 注释风格指南

### 1. 注释应该说明"是什么"和"为什么"

```go
// ✅ 好的注释
// DefaultTimeout is the timeout duration for database operations.
// It prevents long-running queries from blocking the application.
const DefaultTimeout = 30 * time.Second

// ❌ 糟糕的注释
// DefaultTimeout is 30 seconds
const DefaultTimeout = 30 * time.Second
```

### 2. 注释应该解释业务意图

```go
// ✅ 好的注释
// Validate checks if the model name is unique within the project.
// Duplicate names would cause schema generation conflicts.
func (v *ModelValidator) Validate(ctx context.Context, model *Model) error {
    // 实现代码
}

// ❌ 糟糕的注释
// Validate validates the model
func (v *ModelValidator) Validate(ctx context.Context, model *Model) error {
    // 实现代码
}
```

### 3. 注释应该说明特殊行为

```go
// ✅ 好的注释
// FindByID retrieves a model by ID.
// It returns (nil, nil) when the model is not found, not an error.
// An error is returned only for system failures like database connection issues.
func (r *ModelRepository) FindByID(ctx context.Context, id string) (*Model, error) {
    // 实现代码
}
```

### 4. 注释应该说明参数和返回值

```go
// ✅ 好的注释
// Create creates a new model in the specified project.
//
// Parameters:
//   - ctx: request context for cancellation and tracing
//   - projectID: the ID of the parent project
//   - input: model creation parameters
//
// Returns:
//   - *Model: the created model with generated ID and timestamps
//   - error: validation errors or system failures
func (s *ModelService) Create(ctx context.Context, projectID string, input CreateModelInput) (*Model, error) {
    // 实现代码
}
```

## Go 官方注释规范参考

1. 注释应该是完整的句子，以被注释对象的名称开头
2. 公开的（exported）声明必须有文档注释
3. 包注释应该放在 package 语句之前
4. 注释应该解释"为什么"而不是"怎么做"
5. 避免冗余注释，代码能说明的不需要重复

## 工具支持

Go 的文档工具会自动提取注释生成文档：

```bash
# 查看包文档
go doc modelcraft/internal/domain/model

# 查看特定类型/函数文档
go doc modelcraft/internal/domain/model.ModelService
go doc modelcraft/internal/domain/model.ModelService.Create
```

## 注释位置总结

| 元素 | 注释位置 | 示例 |
|------|----------|------|
| 包 | package 语句之前 | `// Package auth ...` |
| 类型 | type 语句之前 | `// User represents ...` |
| 常量 | const 语句之前 | `// MaxRetryCount is ...` |
| 变量 | var 语句之前 | `// ErrNotFound is ...` |
| 函数 | func 语句之前 | `// Create creates ...` |
| 方法 | func 语句之前 | `// Validate validates ...` |
| 接口方法 | 方法签名之前 | `// FindByID retrieves ...` |
| 行内逻辑 | 独占一行，在代码之前 | `// 检查权限` |
