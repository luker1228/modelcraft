# Repository 层开发规范

> **触发条件**: 当开发或修改 `internal/infrastructure/**/*.go` 文件时适用本规范。

## 核心原则

### 1. 依赖规则

```go
// ✅ 允许依赖
import (
    "modelcraft/internal/domain/modeldesign"         // Domain 实体和接口
    "modelcraft/internal/domain/shared"              // Sentinel errors
    "modelcraft/internal/infrastructure/dbgen"       // sqlc 生成代码
    "modelcraft/internal/infrastructure/sqlerr"      // 错误处理 & 类型转换
    "modelcraft/pkg/*"                       
)

// ❌ 禁止依赖
import (
    "modelcraft/internal/app/modeldesign"            // ❌ Application 层
    "modelcraft/internal/interfaces/graphql"         // ❌ Interfaces 层
)
```

### 2. Go Wrapper 架构

Repository 层使用多层 Go Wrapper 处理横切关注点，避免在业务代码中重复处理：

```
Application 层
    ↓ (调用 Repository 方法)
Repository 实现 (sql_*_repository.go)
    ├─ sqlerr 包 (错误分类、类型转换)
    ├─ sqlcLogger wrapper (SQL 日志)
    └─ TxManager wrapper (事务控制)
    ↓
dbgen.Querier (sqlc 生成的接口)
    ↓
*sql.DB / *sql.Tx (数据库驱动)
```

---

## 错误处理规范

### 规则 1: Go Wrapper 已处理错误分类，Repository 直接返回

`sqlerr.QueryWithSQLErrorHandling` / `sqlerr.ExecWithErrorHandling` 已经完成了所有错误分类工作：
- `sql.ErrNoRows` → 自动转换为 `shared.NotFoundError`
- MySQL 错误码 → 自动转换为对应的 `shared.RepositoryError`

Repository 层**不需要手动检查 `IsNotFoundError`**，直接返回 wrapper 的结果即可。

#### 模式 A: 返回 `(value, error)` — 不存在时由 wrapper 返回 NotFoundError

**适用场景**: 必须存在的记录 (如 `GetByID`, `GetByName`)

```go
// ✅ 正确: wrapper 已处理所有错误分类，直接返回
func (r *SqlRoleRepository) GetByID(ctx context.Context, id string) (*Role, error) {
    var row dbgen.Role
    if err := sqlerr.QueryWithSQLErrorHandling(func() error {
        var e error
        row, e = r.q.GetRoleByID(ctx, id)
        return e
    }); err != nil {
        return nil, err  // ← wrapper 已将 sql.ErrNoRows 转为 NotFoundError，直接透传
    }
    return RoleToDomain(row), nil
}

// ❌ 错误: 手动检查 IsNotFoundError（与 wrapper 重复）
func (r *SqlRoleRepository) GetByID(ctx context.Context, id string) (*Role, error) {
    var row dbgen.Role
    err := sqlerr.QueryWithSQLErrorHandling(func() error {
        var e error
        row, e = r.q.GetRoleByID(ctx, id)
        return e
    })
    if err != nil {
        if sqlerr.IsNotFoundError(err) {
            // ❌ 多余: wrapper 已经返回了 NotFoundError，无需再次检查和包装
            return nil, shared.NewNotFoundError("role not found by id: " + id)
        }
        return nil, bizerrors.Wrapf(err, "failed to get role by id: %s", id)
    }
    return RoleToDomain(row), nil
}

// ❌ 错误: 绕过 wrapper，手动调用 sqlc 并检查错误
func (r *SqlRoleRepository) GetByID(ctx context.Context, id string) (*Role, error) {
    row, err := r.q.GetRoleByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil  // ❌ 不要返回 (nil, nil)
        }
        return nil, err
    }
    return RoleToDomain(row), nil
}
```

#### 模式 B: 返回 `(value, bool, error)` — 不存在是预期情况

**适用场景**: 记录不存在是合法状态，需要区分 "不存在" 和 "出错" (如查询可能不存在的映射关系)

> 这是唯一需要在 Repository 层检查 `IsNotFoundError` 的场景 —— 因为需要将 NotFoundError 转换为 `(zero, false, nil)`。

```go
// ✅ 正确: 拦截 NotFound 转为 (zero, false, nil)
func (r *SqlUserRepo) FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error) {
    var userID string

    err := sqlerr.QueryWithSQLErrorHandling(func() error {
        var e error
        userID, e = r.q.FindIDByExternalID(ctx, externalID)
        return e
    })

    if err != nil {
        if sqlerr.IsNotFoundError(err) {
            return "", false, nil  // ← 不存在是预期情况，不是错误
        }
        return "", false, bizerrors.Wrapf(err, "failed to find user id by external id: %s", externalID)
    }

    return userID, true, nil
}
```

**两种模式的选择标准**:

| 场景 | 返回值 | 不存在时返回 | 是否检查 IsNotFoundError | 示例 |
|------|--------|------------|------------------------|------|
| **必须存在的记录** | `(value, error)` | `(nil, NotFoundError)` | 不检查，wrapper 直接返回 | `GetByID`, `GetByName` |
| **可能不存在的查询** | `(value, bool, error)` | `(zero, false, nil)` | 检查，转为 false | `FindIDByExternalID` |

**判断依据**: 
- 记录不存在需要**返回业务错误**给前端 → 用模式 A（wrapper 自动处理）
- 记录不存在需要**执行后续逻辑**处理 → 用模式 B（拦截 NotFound 转为 bool）

### 规则 2: 使用 `shared.RepositoryError`，不使用 `*BusinessError`

```go
// ✅ 正确: 返回 RepositoryError (通过 sqlerr 包装)
func (r *SqlModelRepo) Save(ctx context.Context, model *DataModel) error {
    params := toCreateParams(model)
    return sqlerr.ExecWithErrorHandling(func() error {
        return r.q.CreateModel(ctx, params)
    })
}

// ❌ 错误: 返回 BusinessError
func (r *SqlModelRepo) Save(ctx context.Context, model *DataModel) error {
    err := r.q.CreateModel(ctx, params)
    if err != nil {
        // ❌ Repository 层不应该创建 BusinessError
        return bizerrors.WrapError(err, bizerrors.SystemError, "failed to save model")
    }
    return nil
}
```

### 规则 3: 使用 `sqlerr` 包辅助函数处理错误

若 Repository 构造函数已使用 `dbgenwrap.NewSafeQuerier(q)`，则 `r.q` 的方法在 `safe_querier_gen.go` 中已统一执行 `WrapSQLErrorInPlace`。此时不要再写 `sqlerr.WrapSQLError(r.q.XXX(...))`，避免重复包装。

```go
// ✅ 正确：SafeQuerier 已包装，直接返回
func (r *SqlEndUserPermissionRepository) CreatePermission(ctx context.Context, p *rbac.EndUserPermission) error {
    params := toCreateParams(p)
    return r.q.CreateEndUserPermission(ctx, params)
}

// ❌ 错误：重复包装
return sqlerr.WrapSQLError(r.q.CreateEndUserPermission(ctx, params))
```

错误处理辅助函数位于 `internal/infrastructure/sqlerr/sqlerr.go`：

```go
// 1. WrapSQLError - 包装任意 SQL 错误（仅在 err 未经过 SafeQuerier 时）
err := rawQ.CreateModel(ctx, params)
return sqlerr.WrapSQLError(err)  // 自动分类错误类型

// 2. ExecWithErrorHandling - 用于写操作 (INSERT/UPDATE/DELETE)
return sqlerr.ExecWithErrorHandling(func() error {
    return r.q.CreateModel(ctx, params)
})

// 3. QueryWithSQLErrorHandling - 用于读操作 (SELECT)
var row dbgen.Model
err := sqlerr.QueryWithSQLErrorHandling(func() error {
    var e error
    row, e = r.q.GetModelByID(ctx, id)
    return e
})

// 4. WrapSQLErrorInPlace - 用于 named return + defer 场景
func (r *SqlModelRepo) GetByID(ctx context.Context, id string) (result *DataModel, err error) {
    defer sqlerr.WrapSQLErrorInPlace(&err)
    // ... 直接使用 err，defer 会自动包装
}
```

**错误分类逻辑** (`sqlerr.AnalyzeSQLError`):
- `sql.ErrNoRows` → `shared.NewNotFoundError` (sentinel error)
- MySQL 错误 1062 (Duplicate entry) → `shared.ErrTypeDuplicatedKey`
- MySQL 错误 1451/1452 (FK constraint) → `shared.ErrTypeConstraint`
- MySQL 错误 1064 (Syntax error) → `shared.ErrTypeDML`
- MySQL 错误 1146 (Table not found) → `shared.ErrTypeNotFound`
- 连接/超时/死锁等 → 根据错误消息模式匹配分类

### 规则 4: 模式 A 直接返回 wrapper 错误，不再手动包装

由于 Go Wrapper 已完成错误分类，模式 A 的方法直接返回即可。
模式 B 的方法因为需要区分 NotFound 和真实错误，仍需 `bizerrors.Wrapf` 包装非 NotFound 错误：

```go
// ✅ 模式 A: 直接返回 wrapper 错误
func (r *SqlRoleRepository) GetByName(ctx context.Context, name string) (*Role, error) {
    var row dbgen.Role
    if err := sqlerr.QueryWithSQLErrorHandling(func() error {
        var e error
        row, e = r.q.GetRoleByName(ctx, name)
        return e
    }); err != nil {
        return nil, err  // wrapper 已分类，直接返回
    }
    return RoleToDomain(row), nil
}

// ✅ 模式 B: 拦截 NotFound，非 NotFound 用 bizerrors.Wrapf 添加上下文
func (r *SqlUserRepo) FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error) {
    // ...
    if err != nil {
        if sqlerr.IsNotFoundError(err) {
            return "", false, nil
        }
        return "", false, bizerrors.Wrapf(err, "failed to find user id by external id: %s", externalID)
    }
    return userID, true, nil
}
```

---

## Go Wrapper 规范

### sqlcLogger — SQL 日志 Wrapper

`sqlcLogger` 是 `DBTX` 接口的 Go Wrapper 实现，对所有 SQL 操作透明地添加日志记录（`internal/infrastructure/repository/sqlc_logger.go`）。

```go
// 创建带日志的数据库连接
loggedDB := NewSqlcLogger(db, SqlcLogInfo, 100*time.Millisecond)

// 传给 sqlc 生成的 queries（对 Repository 透明）
queries := dbgen.New(loggedDB)
```

**四个日志级别**（仿 GORM）:

| 级别 | 行为 |
|------|------|
| `SqlcLogSilent` | 关闭所有日志 |
| `SqlcLogError` | 仅记录错误查询 |
| `SqlcLogWarn` | 记录错误 + 超过慢查询阈值的查询 |
| `SqlcLogInfo` | 记录所有查询 |

**特性**:
- 实现 `DBTX` 接口（`ExecContext`、`QueryContext`、`QueryRowContext`、`PrepareContext`）
- 自动记录耗时、参数、错误
- `QueryRowContext` 因 `*sql.Row` 延迟执行特性，仅记录 dispatch
- SQL 多行文本自动折叠为单行 (`cleanSQL`)

**规则: Repository 代码不需要感知 sqlcLogger 的存在。日志由 Wrapper 自动处理。**

### TxManager — 事务 Wrapper

`TxManager` 提供事务管理的 Go Wrapper，采用 **显式 Querier 传递** 模式（`internal/infrastructure/repository/tx_manager.go`）：

```go
// 接口定义
type TxManager interface {
    WithTx(ctx context.Context, fn func(ctx context.Context, q dbgen.Querier) error) error
}

// 使用方式 (Application 层)
func (s *ModelDesignAppService) CreateModel(ctx context.Context, cmd CreateModelCommand) error {
    return s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
        // q 绑定到事务，Repository 无感知
        modelRepo := repository.NewSqlModelDesignRepository(q)
        
        if err := modelRepo.Save(ctx, orgName, model); err != nil {
            return err
        }
        // ... 更多操作，同一事务
        return nil
    })
}
```

**特性**:
- 自动 Begin/Commit/Rollback
- panic recovery 时自动 Rollback
- 回调函数接收绑定到事务的 `dbgen.Querier`
- Repository 接收 `Querier` 接口，对事务/非事务无感知

### ConnectionFactory — 连接工厂

`ConnectionFactory` 持有数据库连接，用于统一构造 Repository（`internal/infrastructure/repository/connection_factory.go`）：

```go
type ConnectionFactory struct {
    SqlDB *sql.DB
}

func NewConnectionFactory(sqlDB *sql.DB) *ConnectionFactory {
    return &ConnectionFactory{SqlDB: sqlDB}
}
```

---

## Repository 接口实现规范

### 规则 5: 接收 `Querier` 接口，不直接依赖 `*sql.DB`

```go
// ✅ 正确: 使用 Querier 接口 (支持事务和非事务)
type SqlModelDesignRepository struct {
    q dbgen.Querier  // 接口，可以是 *sql.DB 或 *sql.Tx
}

func NewSqlModelDesignRepository(q dbgen.Querier) modeldesign.ModelRepository {
    return &SqlModelDesignRepository{q: q}
}

// ❌ 错误: 直接依赖 *sql.DB
type SqlModelDesignRepository struct {
    db *sql.DB  // ❌ 无法支持事务
}
```

**好处**: TxManager 可传入绑定到事务的 Querier，Repository 层无感知。

### 规则 6: 编译期接口满足检查

每个 Repository 实现文件末尾必须添加编译期检查：

```go
// Compile-time interface satisfaction checks.
var (
    _ organization.OrganizationRepository = (*SqlOrganizationRepository)(nil)
    _ user.UserRepository                 = (*SqlUserRepository)(nil)
)
```

---

### 规则 7: 处理 `sql.Null*` 类型

使用 `sqlerr` 包提供的辅助函数（`internal/infrastructure/sqlerr/sqlerr.go`）：

```go
// ✅ 正确: 使用辅助函数
func toDomain(row dbgen.Model) *DataModel {
    return &DataModel{
        ID:          row.ID,
        Name:        row.Name,
        Description: sqlerr.NullStrToPtr(row.Description),  // sql.NullString → *string
        Version:     sqlerr.NullInt32ToPtr(row.Version),     // sql.NullInt32 → *int32
        CreatedAt:   row.CreatedAt.Time,                     // sql.NullTime → time.Time
        IsActive:    sqlerr.NullBoolToBool(row.IsActive),    // sql.NullBool → bool (false for NULL)
    }
}

func toCreateParams(model *DataModel) dbgen.CreateModelParams {
    return dbgen.CreateModelParams{
        ID:          model.ID,
        Name:        model.Name,
        Description: sqlerr.PtrToNullStr(model.Description),  // *string → sql.NullString
        Version:     sqlerr.PtrToNullInt32(model.Version),     // *int32 → sql.NullInt32
        IsActive:    sqlerr.BoolToNullBool(model.IsActive),    // bool → sql.NullBool
    }
}

// ❌ 错误: 手动处理 (重复代码)
func toDomain(row dbgen.Model) *DataModel {
    var desc *string
    if row.Description.Valid {  // ❌ 重复的 null 检查逻辑
        desc = &row.Description.String
    }
    return &DataModel{Description: desc}
}
```

**提供的辅助函数**:
- `NullStrToPtr` / `PtrToNullStr` — `sql.NullString` ↔ `*string`
- `NullTimeToPtr` / `PtrToNullTime` — `sql.NullTime` ↔ `*time.Time`
- `NullInt64ToPtr` / `PtrToNullInt64` — `sql.NullInt64` ↔ `*int64`
- `NullInt32ToPtr` / `PtrToNullInt32` — `sql.NullInt32` ↔ `*int32`
- `NullBoolToPtr` / `BoolToNullBool` — `sql.NullBool` ↔ `*bool` / `bool`
- `NullBoolToBool` — `sql.NullBool` → `bool` (NULL 返回 `false`)

### 规则 8: 处理 JSON 字段

```go
// ✅ 正确: JSON Marshal/Unmarshal with error handling
func toCreateParams(lf *LogicalForeignKey) (dbgen.CreateLogicalForeignKeyParams, error) {
    sourceFieldsJSON, err := json.Marshal(lf.SourceFields)
    if err != nil {
        return dbgen.CreateLogicalForeignKeyParams{}, fmt.Errorf("marshal source_fields: %w", err)
    }
    targetFieldsJSON, err := json.Marshal(lf.TargetFields)
    if err != nil {
        return dbgen.CreateLogicalForeignKeyParams{}, fmt.Errorf("marshal target_fields: %w", err)
    }
    return dbgen.CreateLogicalForeignKeyParams{
        SourceFields: json.RawMessage(sourceFieldsJSON),
        TargetFields: json.RawMessage(targetFieldsJSON),
    }, nil
}

func toDomain(row dbgen.LogicalForeignKey) (*LogicalForeignKey, error) {
    var sourceFields []string
    if len(row.SourceFields) > 0 {
        if err := json.Unmarshal(row.SourceFields, &sourceFields); err != nil {
            return nil, fmt.Errorf("unmarshal source_fields: %w", err)
        }
    }
    return &LogicalForeignKey{SourceFields: sourceFields}, nil
}
```

---

## 事务支持规范

### 规则 9: Repository 不管理事务

```go
// ✅ 正确: Repository 接收 Querier，不关心是否在事务中
func (r *SqlModelRepo) Save(ctx context.Context, model *DataModel) error {
    params := toCreateParams(model)
    return sqlerr.ExecWithErrorHandling(func() error {
        return r.q.CreateModel(ctx, params)  // r.q 可能是 *sql.DB 或 *sql.Tx
    })
}

// ❌ 错误: Repository 内部开启事务
func (r *SqlModelRepo) SaveWithRelations(ctx context.Context, model *DataModel) error {
    tx, err := r.db.BeginTx(ctx, nil)  // ❌ 事务应该在 Application 层管理
    if err != nil {
        return err
    }
    defer tx.Rollback()
    // ... 多个操作
    return tx.Commit()
}
```

**事务由 Application 层通过 TxManager 管理**:

```go
// internal/app/modeldesign/model_app.go
func (s *ModelDesignAppService) CreateModelWithFields(ctx context.Context, cmd CreateModelCommand) error {
    return s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
        // 使用事务 Querier 创建 Repository
        modelRepo := repository.NewSqlModelDesignRepository(q)
        
        // Repository 无感知，只是接收了一个 Querier
        if err := modelRepo.Save(ctx, cmd.OrgName, model); err != nil {
            return err
        }
        
        // 同一事务中的更多操作...
        return nil
    })
    // TxManager 自动 Commit 或 Rollback
}
```

---

## RowsAffected 检查模式

对于 UPDATE/DELETE 操作，检查受影响行数来判断记录是否存在：

```go
// ✅ 正确: 检查 RowsAffected
func (r *SqlModelDesignRepository) Update(ctx context.Context, model *modeldesign.DataModel) error {
    result, err := r.q.UpdateModel(ctx, ModelToUpdateParams(model))
    if err != nil {
        return sqlerr.WrapSQLError(err)
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "Model not found or not updated")
    }

    return nil
}
```

---

## 检查清单

开发 Repository 层时，请确认以下事项:

- [ ] **选择正确的 RecordNotFound 处理模式**:
  - [ ] 必须存在的记录 → 模式 A: `(value, error)`，wrapper 自动返回 NotFoundError，不手动检查
  - [ ] 可能不存在的查询 → 模式 B: `(value, bool, error)`，拦截 NotFoundError 转为 `(zero, false, nil)`
- [ ] 使用 `sqlerr.ExecWithErrorHandling` / `sqlerr.QueryWithSQLErrorHandling` 包装 DB 操作
- [ ] 模式 A 直接返回 wrapper 错误，**不手动检查 `IsNotFoundError`**
- [ ] 返回 `shared.RepositoryError`，不返回 `*bizerrors.BusinessError`
- [ ] 接收 `dbgen.Querier` 接口，支持事务和非事务
- [ ] 文件末尾添加编译期接口满足检查 (`var _ Interface = (*Impl)(nil)`)
- [ ] 使用 `sqlerr` 包辅助函数处理 `sql.Null*` 类型
- [ ] 不在 Repository 层开启事务（事务由 Application 层通过 TxManager 管理）
- [ ] 不打印错误日志（日志由 sqlcLogger Wrapper 自动处理）
- [ ] 方法命名遵循约定 (`GetByID`, `FindByXXX`, `Save` 等)
- [ ] UPDATE/DELETE 操作检查 `RowsAffected`

---

## 参考索引

| 主题 | 文件 |
|------|------|
| 错误分类 & 类型转换 | `internal/infrastructure/sqlerr/sqlerr.go` |
| SQL 日志 Wrapper | `internal/infrastructure/repository/sqlc_logger.go` |
| 事务管理 Wrapper | `internal/infrastructure/repository/tx_manager.go` |
| 连接工厂 | `internal/infrastructure/repository/connection_factory.go` |
| 数据库连接 | `internal/infrastructure/repository/sql_connection.go` |
| Repository 错误辅助 | `internal/infrastructure/repository/error_helper.go` |
| RepositoryError / Sentinel | `internal/domain/shared/repository_error.go` |
| 模型设计 Repository 示例 | `internal/infrastructure/repository/sql_modeldesign_repository.go` |
| 组织/用户 Repository 示例 | `internal/infrastructure/repository/sql_org_repository.go` |
