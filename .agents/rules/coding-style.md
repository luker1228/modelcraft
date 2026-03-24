---
paths:
  - "**/*.go"
---
# 代码风格禁止项

以下为强制禁止项，违反即视为不符合规范。

**详细规范请参考：**
- Refer to @modelcraft-go/ai-metadata/2-development/code-style.md for comprehensive coding standards
- See skill: `coding-standards` for project-specific patterns

## 日志

- **禁止裸用 `log`**，必须使用 `logfacade` 包
- **禁止打印中文日志**，必须使用英文
- **禁止打印 debug 日志**
- **禁止硬编码日志字段 key**，必须使用 `logfacade` 包定义的常量
- **禁止在格式化字符串中直接打印错误**：`logger.Errorf("failed: %v", err)` ❌
  - 正确：`logger.Error("failed", logfacade.Err(err))` ✅
- **堆栈跟踪规则**：
  - **必须使用** `Stack()`：错误转换点（GraphQL/HTTP adapter）、顶层中间件
  - **禁止使用** `Stack()`：Service/Repository 层
  - **必须同时使用** `Err()` 和 `Stack()`

Refer to @modelcraft-go/ai-metadata/2-development/logging.md for detailed logging patterns.

## 错误

- **禁止裸用 `errors`**，必须使用 `pkg/bizerrors`

```go
// ✅ 正确
import pkgerrors "modelcraft/pkg/bizerrors"
if err != nil {
    return nil, pkgerrors.Wrapf(err, "get user %s", id)
}

// ❌ 错误
import "errors"
return nil, errors.New("invalid input")
```

## 并发

- **禁止裸用 `go func`**，必须使用 `bizutils.GoWithCtx`

```go
// ✅ 正确
import "modelcraft/pkg/bizutils"
bizutils.GoWithCtx(ctx, func(ctx context.Context) {
    // do work
})

// ❌ 错误
go func() {
    // do work
}()
```

## 注释

- **禁止使用行尾注释**，注释必须独占一行或多行
- **所有导出的（Exported）标识符必须写注释**（符合 Go 官方规范）
  - 注释必须以标识符名称开头，形成完整句子
  - 注释应该说明"是什么"和"为什么"，而不是"怎么做"

```go
// ✅ 正确
// MaxRetryCount 表示最大重试次数
const MaxRetryCount = 3

// ❌ 错误
const MaxRetryCount = 3 // 最大重试次数

// ❌ 错误（缺少注释）
func (s *UserService) Create(ctx context.Context, name string) (string, error) {
    // 实现代码
}
```

Refer to @modelcraft-go/ai-metadata/2-development/comments.md for detailed comment standards.

## 类型转换

- **禁止使用 Go 原生类型断言 `x.(T)`**，必须使用 `github.com/spf13/cast` 包

```go
// ✅ 正确
import "github.com/spf13/cast"
schemaType, err := cast.ToStringE(schemaMap["type"])
if err != nil {
    return nil, bizerrors.Wrap(err, "invalid type")
}

// ❌ 错误
schemaType := schemaMap["type"].(string)

// ❌ 错误（ok 模式虽然安全，但项目统一用 cast）
schemaType, ok := schemaMap["type"].(string)
if !ok {
    return nil, bizerrors.New("invalid type")
}
```

Refer to @modelcraft-go/ai-metadata/2-development/type-conversion.md for type conversion patterns.

## Context 处理

- **禁止直接使用 `context.Value()` 提取值**，必须统一使用 `pkg/ctxutils` 包
- **Interface 层**：使用 `pkg/ctxutils` 提取 context 值，作为显式参数传递给下层
- **Application/Domain 层**：接收 orgName、userID 等作为**显式参数**，禁止从 context 提取

```go
// ✅ 正确（Interface 层）
import "modelcraft/pkg/ctxutils"
func (r *mutationResolver) CreateModel(ctx context.Context, input Input) (*Payload, error) {
    orgName, err := ctxutils.GetOrgNameFromContext(ctx)
    if err != nil {
        return nil, fmt.Errorf("organization context not found: %w", err)
    }
    // 作为显式参数传递给下层
    model, err := r.service.CreateModel(ctx, orgName, input)
    ...
}

// ❌ 错误 1：直接使用 context.Value()
orgName := ctx.Value("org_name").(string)

// ❌ 错误 2：Application 层从 context 提取
func (s *ModelService) CreateModel(ctx context.Context, input CreateModelInput) (*Model, error) {
    orgName, _ := ctxutils.GetOrgNameFromContext(ctx)  // 禁止！
    ...
}
```

Refer to @modelcraft-go/ai-metadata/2-development/context-handling.md for context handling patterns.

## sqlc 自定义类型

- **JSON 类型字段必须使用 `db:"type:json"` 标签**，不可省略
- `Scan` 方法必须先检查 `fmt.Stringer` 接口，再处理标准类型

```go
// ✅ 正确
type ModelRelationPO struct {
    SourceFields StringSlice `db:"type:json" json:"source_fields"`
    TargetFields StringSlice `db:"type:json" json:"target_fields"`
}

// ❌ 错误：缺少 db:"type:json"
type ModelRelationPO struct {
    SourceFields StringSlice `json:"source_fields"`
}
```

> Refer to @modelcraft-go/ai-metadata/2-development/sqlc-custom-types.md for sqlc custom type patterns.

## sqlc 常见坑

- **禁止使用 `db.Model`**（会自动启用软删除）
- struct 更新零值问题：使用 map 或 `Select()` 明确指定字段
- 预加载避免 N+1 查询：使用 `Preload()`
- 查询条件不要用 map：使用明确条件字符串

> Refer to @modelcraft-go/ai-metadata/2-development/sqlc-custom-types.md for detailed sqlc pitfalls.
