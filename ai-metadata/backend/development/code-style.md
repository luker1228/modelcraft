# 代码风格规范

> **优先级: 高** - 定义 ModelCraft Go 后端的编码风格、命名约定及关键模式。

## 命名约定

- **包名**: 小写，单个单词
- **接口**: 名词形式（如 `ModelRepository`、`QueryExecutor`）
- **领域服务**: 行为导向（如 `ModelService`）
- **应用服务**: 用例导向（如 `CreateModelUseCase`）

## 错误处理

使用结构化业务错误（`pkg/bizerrors`），**不使用**标准 `errors` 包：

```go
import "modelcraft/pkg/bizerrors"

// 映射到 GraphQL ModelNotFound 错误
return bizerrors.New("NOT_FOUND.MODEL", "model not found", ...)
```

**标准错误码：**

| 错误码 | GraphQL 类型 |
|--------|--------------|
| `NOT_FOUND.{RESOURCE}` | `{Resource}NotFound` |
| `PARAM_INVALID.{RESOURCE}` | `Invalid{Resource}Input` |
| `CONFLICT.{RESOURCE}` | `{Resource}AlreadyExists` |
| `OPERATION_DENIED.{RESOURCE}` | `CannotDelete{Resource}` |
| `SYSTEM_ERROR` | 系统内部错误 |

**领域示例：** `NOT_FOUND.PROJECT`、`CONFLICT.MODEL`、`OPERATION_DENIED.MODEL`

**重要规则：**
- 数据不存在时返回 `nil`（而非 error）
- GraphQL 错误适配器：`internal/interfaces/graphql/adapter/*_error_adapter.go`

> 完整错误处理模式请参考 [error-handling.md](./error-handling.md)。

## 日志

使用 `pkg/logfacade`；错误使用结构化字段或 `pkg/logfacade/constant.go` 中的常量：

```go
import "modelcraft/pkg/logfacade"

logger := logfacade.GetDefault()
logger.Infof("Operation started, modelID: %s", modelID)
logger.Errorf("Operation failed: %v", err)
```

**复杂对象日志：**
```go
import "modelcraft/pkg/bizutils"

logger.Infof("Object data: %s", bizutils.MarshalToStringIgnoreErr(obj))
```

> 完整日志模式请参考 [logging.md](./logging.md)。

## GraphQL 注意事项

- **NEVER** 使用 `task regenerate-gql` —— 会删除自定义 resolver 实现
- 只在修改 `api/graph/schema/*.graphql` 后才运行 `task generate-gql`
- `internal/interfaces/graphql/` 中的生成文件可以安全编辑

## 数据库模式

- 所有数据库操作使用 **sqlc** 生成类型安全的 Go 代码
- SQL 查询定义在 `db/queries/*.sql` 文件中
- 生成的代码位于 `internal/infrastructure/database/sqlc/`
- 事务在应用服务层管理

### sqlc 工作流程

```bash
# 1. 编写 SQL 查询 (db/queries/*.sql)
# 2. 生成 Go 代码
task sqlc:generate

# 3. 使用生成的类型安全函数
```

### SQL 查询示例

```sql
-- db/queries/model.sql

-- name: GetModelByID :one
SELECT * FROM models WHERE id = ? LIMIT 1;

-- name: ListModelsByProjectID :many
SELECT * FROM models WHERE project_id = ? ORDER BY created_at DESC;

-- name: CreateModel :execresult
INSERT INTO models (id, name, project_id, created_at) 
VALUES (?, ?, ?, ?);
```

### 使用生成的代码

```go
// 使用 sqlc 生成的类型安全查询
model, err := queries.GetModelByID(ctx, modelID)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil  // 数据不存在返回 nil
    }
    return nil, err
}
```

### 注意事项

- **不使用** ORM (如 sqlc)，统一使用 sqlc
- Repository 层封装 sqlc 生成的查询函数
- 复杂查询使用 JOIN，避免 N+1 问题
- 使用预编译语句，sqlc 自动处理参数绑定

### 事务最佳实践

#### 核心原则

| 原则 | 说明 |
|------|------|
| **应用层管理** | 事务在 UseCase/Application Service 层开启和提交 |
| **Repository 无感知** | Repository 层不关心是否在事务中，接收 `DBTX` 接口 |
| **defer 保证回滚** | 使用 `defer tx.Rollback()` 确保异常时回滚 |
| **显式提交** | 只有成功时才调用 `tx.Commit()` |

#### 事务使用模式

```go
// UseCase 层 - 事务管理
func (uc *CreateModelUseCase) Execute(ctx context.Context, input CreateModelInput) (*Model, error) {
    // 1. 开启事务
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    // 2. defer 保证回滚 (commit 后 rollback 是 no-op)
    defer tx.Rollback()

    // 3. 创建事务作用域的 queries
    queries := sqlc.New(tx)

    // 4. 执行多个数据库操作
    model, err := uc.modelRepo.Create(ctx, queries, input)
    if err != nil {
        return nil, err  // 自动回滚
    }

    err = uc.auditRepo.Log(ctx, queries, "model_created", model.ID)
    if err != nil {
        return nil, err  // 自动回滚
    }

    // 5. 全部成功，提交事务
    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return model, nil
}
```

#### Repository 层 - 接收 DBTX 接口

```go
// Repository 接口 - 使用 DBTX 支持事务
type ModelRepository interface {
    Create(ctx context.Context, q *sqlc.Queries, input CreateModelInput) (*Model, error)
    GetByID(ctx context.Context, q *sqlc.Queries, id string) (*Model, error)
}

// Repository 实现 - 不关心是否在事务中
func (r *modelRepository) Create(ctx context.Context, q *sqlc.Queries, input CreateModelInput) (*Model, error) {
    result, err := q.CreateModel(ctx, sqlc.CreateModelParams{
        ID:        input.ID,
        Name:      input.Name,
        ProjectID: input.ProjectID,
    })
    if err != nil {
        return nil, err
    }
    return mapToModel(result), nil
}
```

#### 事务辅助函数 (推荐)

```go
// pkg/database/tx.go - 事务辅助函数
func WithTx(ctx context.Context, db *sql.DB, fn func(q *sqlc.Queries) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := fn(sqlc.New(tx)); err != nil {
        return err
    }

    return tx.Commit()
}

// 使用辅助函数简化代码
func (uc *CreateModelUseCase) Execute(ctx context.Context, input CreateModelInput) error {
    return database.WithTx(ctx, uc.db, func(q *sqlc.Queries) error {
        _, err := uc.modelRepo.Create(ctx, q, input)
        if err != nil {
            return err
        }
        return uc.auditRepo.Log(ctx, q, "model_created", input.ID)
    })
}
```

#### ❌ 错误做法

```go
// ❌ 错误：Repository 层管理事务
func (r *modelRepository) CreateWithAudit(ctx context.Context, input CreateModelInput) error {
    tx, _ := r.db.BeginTx(ctx, nil)  // Repository 不应该开启事务
    // ...
}

// ❌ 错误：忘记 defer rollback
func (uc *UseCase) Execute(ctx context.Context) error {
    tx, _ := uc.db.BeginTx(ctx, nil)
    // 如果下面代码 panic，事务不会回滚！
    // ...
    return tx.Commit()
}

// ❌ 错误：在错误分支手动回滚
func (uc *UseCase) Execute(ctx context.Context) error {
    tx, _ := uc.db.BeginTx(ctx, nil)
    if err := step1(); err != nil {
        tx.Rollback()  // 冗余，应该用 defer
        return err
    }
    if err := step2(); err != nil {
        tx.Rollback()  // 冗余，应该用 defer
        return err
    }
    return tx.Commit()
}
```

#### ✅ 正确做法

```go
// ✅ 正确：defer + 显式 commit
func (uc *UseCase) Execute(ctx context.Context) error {
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()  // 保证回滚

    if err := step1(); err != nil {
        return err  // 自动回滚
    }
    if err := step2(); err != nil {
        return err  // 自动回滚
    }

    return tx.Commit()  // 成功提交
}
```

## Context 工具（ctxutils）

使用 `pkg/ctxutils` 管理请求上下文中的用户信息、HTTP 请求元数据等。

### 常用函数

```go
import "modelcraft/pkg/ctxutils"

// 用户 ID
ctx = ctxutils.SetUserID(ctx, userID)
userID, err := ctxutils.GetUserIDFromContext(ctx)

// 租户/组织名
ctx = ctxutils.SetOrgName(ctx, orgName)
orgName, err := ctxutils.GetOrgNameFromContext(ctx)

// 权限列表（格式: "resource:action"，支持通配符 "*"、"model:*"）
ctx = ctxutils.SetPermissions(ctx, []string{"model:read", "cluster:manage"})
permissions, err := ctxutils.GetPermissionsFromContext(ctx)

// HTTP 请求元数据（由中间件设置）
hrc := ctxutils.FromContext(ctx)       // 获取 HttpRequestContext
requestID := ctxutils.GetRequestID(ctx) // 直接获取 RequestId

// Schema 缓存控制（默认 true，可通过 ?useCache=false 关闭）
useCache := ctxutils.GetUseCache(ctx)
```

**重要规则：**
- Getter 函数在值缺失时返回 `error`，**必须**检查错误
- `HttpRequestContext` 仅用于 HTTP 层关注点（tracing、日志），不用于业务逻辑
- `ctxutils` 使用自定义 `contextKey` 类型，避免与其他包的 context key 冲突

## 协程（GoWithCtx）

**始终**使用 `bizutils.GoWithCtx` 代替裸 `go` 关键字启动协程：

```go
import "modelcraft/pkg/bizutils"

// ✅ 正确：携带上下文、自动捕获 panic
bizutils.GoWithCtx(ctx, func(ctx context.Context) {
    // 异步逻辑，ctx 已从父 goroutine 传递
})

// ❌ 错误：裸 goroutine，panic 会导致进程崩溃
go func() {
    // ...
}()
```

**GoWithCtx 的作用：**
- 将父 `ctx` 传递给子协程（保留 tracing、用户信息等）
- 自动捕获 `panic`，通过 `logfacade` 记录完整诊断信息：
  - panic 值、完整堆栈（所有协程）、协程数量、时间戳

**典型场景：**

```go
// 后台定时任务
bizutils.GoWithCtx(ctx, func(ctx context.Context) {
    for {
        select {
        case <-ticker.C:
            doSyncWork(ctx)
        case <-stopChan:
            return
        }
    }
})

// 异步事件处理
bizutils.GoWithCtx(ctx, func(ctx context.Context) {
    if err := processEvent(ctx, event); err != nil {
        logger.WithContext(ctx).Errorf("事件处理失败: %v", err)
    }
})
```

## 代码格式化

- 使用 **gofumpt** 格式化代码（通过 `task lint-fix` 自动修复）
- 不使用 `panic` 处理错误
- 接口定义在使用方（非实现方）
- 优先组合而非继承
