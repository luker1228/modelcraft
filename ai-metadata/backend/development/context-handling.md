# Context 处理规范

> **优先级: 高** - 定义 ModelCraft Go 后端的 Context 使用标准。

## 核心原则

- **禁止直接使用 `context.Value()` 提取值**，必须统一使用 `pkg/ctxutils` 包
- **Interface 层**：使用 `pkg/ctxutils` 提取 context 值，作为显式参数传递给下层
- **Application/Domain 层**：接收 orgName、userID 等作为**显式参数**，禁止从 context 提取

## 正确示例

### Interface 层 - 提取并传递

```go
// internal/interfaces/graphql/*.resolvers.go
import "modelcraft/pkg/ctxutils"

func (r *mutationResolver) CreateModel(ctx context.Context, input Input) (*Payload, error) {
	// ✅ 使用 pkg/ctxutils 提取 context
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("organization context not found: %w", err)
	}

	userID, err := ctxutils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("user context not found: %w", err)
	}

	// ✅ 作为显式参数传递给下层
	model, err := r.service.CreateModel(ctx, orgName, userID, input)
	if err != nil {
		return nil, err
	}

	return &Payload{Model: model}, nil
}
```

### Application 层 - 接收显式参数

```go
// internal/app/*.go
func (s *ModelService) CreateModel(
	ctx context.Context,
	orgName string,    // ✅ 显式参数
	userID string,     // ✅ 显式参数
	input CreateModelInput,
) (*Model, error) {
	// 直接使用参数，不从 context 提取
	return s.repo.Create(ctx, orgName, input)
}
```

## 错误示例

### ❌ 错误 1：直接使用 context.Value() 提取

```go
func (r *mutationResolver) CreateModel(ctx context.Context, input Input) (*Payload, error) {
	orgName := ctx.Value("org_name").(string)  // 禁止！必须用 ctxutils
	...
}
```

### ❌ 错误 2：Application 层从 context 提取

```go
func (s *ModelService) CreateModel(ctx context.Context, input CreateModelInput) (*Model, error) {
	// 禁止！Application 层不应该从 context 提取
	orgName, _ := ctxutils.GetOrgNameFromContext(ctx)
	...
}
```

## 常用函数

### 用户信息

```go
import "modelcraft/pkg/ctxutils"

// 设置用户 ID
ctx = ctxutils.SetUserID(ctx, userID)

// 获取用户 ID
userID, err := ctxutils.GetUserIDFromContext(ctx)
if err != nil {
    return fmt.Errorf("user context not found: %w", err)
}
```

### 租户信息

```go
// 设置组织名
ctx = ctxutils.SetOrgName(ctx, orgName)

// 获取组织名
orgName, err := ctxutils.GetOrgNameFromContext(ctx)
if err != nil {
    return fmt.Errorf("organization context not found: %w", err)
}
```

### 权限信息

```go
// 设置权限列表（格式: "resource:action"，支持通配符 "*"、"model:*"）
ctx = ctxutils.SetPermissions(ctx, []string{"model:read", "cluster:manage"})

// 获取权限列表
permissions, err := ctxutils.GetPermissionsFromContext(ctx)
if err != nil {
    return fmt.Errorf("permissions context not found: %w", err)
}
```

### HTTP 请求元数据

```go
// 获取 HttpRequestContext（由中间件设置）
hrc := ctxutils.FromContext(ctx)

// 直接获取 RequestId
requestID := ctxutils.GetRequestID(ctx)
```

### Schema 缓存控制

```go
// 获取缓存标志（默认 true，可通过 ?useCache=false 关闭）
useCache := ctxutils.GetUseCache(ctx)
```

## 重要规则

1. **Getter 函数在值缺失时返回 `error`**，必须检查错误
2. **`HttpRequestContext` 仅用于 HTTP 层关注点**（tracing、日志），不用于业务逻辑
3. **`ctxutils` 使用自定义 `contextKey` 类型**，避免与其他包的 context key 冲突
4. **Application/Domain 层使用显式参数**，提高可测试性和可维护性

## 分层职责

| 层 | 职责 | 使用方式 |
|----|------|----------|
| Interface | 提取 context 值 | 使用 `ctxutils.GetXxx()` |
| Application | 接收显式参数 | 函数签名包含 `orgName`, `userID` 等 |
| Domain | 接收显式参数 | 函数签名包含 `orgName`, `userID` 等 |

## 设计原则

### 为什么使用显式参数？

1. **可测试性**：测试时不需要构造复杂的 context
2. **可维护性**：函数签名清晰，依赖关系明确
3. **类型安全**：编译时检查参数类型
4. **避免隐式依赖**：业务逻辑不依赖 context 的具体实现

### 示例对比

```go
// ❌ 隐式依赖 - 不推荐
func (s *Service) Create(ctx context.Context, input Input) error {
    orgName, _ := ctxutils.GetOrgNameFromContext(ctx)  // 隐式依赖
    // ...
}

// ✅ 显式参数 - 推荐
func (s *Service) Create(ctx context.Context, orgName string, input Input) error {
    // orgName 作为参数明确传入
    // ...
}
```

测试时的差异：

```go
// ❌ 隐式依赖 - 测试复杂
func TestCreate_Implicit(t *testing.T) {
    ctx := context.Background()
    ctx = ctxutils.SetOrgName(ctx, "test-org")  // 需要构造 context
    s.Create(ctx, input)
}

// ✅ 显式参数 - 测试简单
func TestCreate_Explicit(t *testing.T) {
    ctx := context.Background()
    s.Create(ctx, "test-org", input)  // 直接传参
}
```
