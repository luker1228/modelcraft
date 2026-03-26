---
description: Repository 层开发规范 - 用于 internal/infrastructure/*.go 开发
globs: internal/infrastructure/**/*.go
alwaysApply: false
priority: high
---

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
    "modelcraft/pkg/*"                       
)

// ❌ 禁止依赖
import (
    "modelcraft/internal/app/modeldesign"            // ❌ Application 层
    "modelcraft/internal/interfaces/graphql"         // ❌ Interfaces 层
)
```
## 错误处理规范

### 规则 1: RecordNotFound 的两种处理模式

#### 模式 A: 返回 `(value, error)` - 不存在时返回错误

**适用场景**: 必须存在的记录,不存在是错误情况 (如 `GetByID`)

```go
// ✅ 正确: 直接返回错误 (包括 NotFoundError)
func (r *SqlModelRepo) GetByID(ctx context.Context, id string) (*DataModel, error) {
    var row dbgen.Model
    err := QueryWithSQLErrorHandling(func() error {
        var e error
        row, e = r.q.GetModelByID(ctx, id)
        return e
    })
    if err != nil {
        // QueryWithSQLErrorHandling 已将 sql.ErrNoRows 转换为 shared.NewNotFoundError
        // 直接返回,由 Application 层判断如何处理
        return nil, err  // ← 关键: 所有错误(包括 NotFound)都直接返回
    }
    return toDomain(row), nil
}

// ❌ 错误: 检查 NotFound 后返回 (nil, nil)
func (r *SqlModelRepo) GetByID(ctx context.Context, id string) (*DataModel, error) {
    row, err := r.q.GetModelByID(ctx, id)
    if err != nil {
        if shared.IsNotFoundError(err) {
            // ❌ 不要返回 (nil, nil)，应该直接返回 error
            return nil, nil
        }
        return nil, err
    }
    return toDomain(row), nil
}
```

#### 模式 B: 返回 `(value, bool, error)` - 不存在是预期情况

**适用场景**: 记录不存在是合法状态,不应视为错误 (如通过外部 ID 查询可能不存在的映射关系)

```go
// ✅ 正确: 经典 Go 模式 - 返回 (value, found, error)
// FindIDByExternalID retrieves the internal user ID by external authentication provider ID.
// Returns ("", false, nil) if no user matches the given externalID.
// Returns ("", false, err) on system failure.
func (r *SqlUserRepo) FindIDByExternalID(ctx context.Context, externalID string) (string, bool, error) {
    var userID string
    err := QueryWithSQLErrorHandling(func() error {
        var e error
        userID, e = r.q.GetUserIDByExternalID(ctx, externalID)
        return e
    })
    
    if err != nil {
        // 不存在是预期情况,不是错误
        if shared.IsNotFoundError(err) {
            return "", false, nil  // ← 关键: 返回 false 表示未找到,不返回 error
        }
        // 其他错误 (DB 故障等) 才返回 error
        return "", false, err
    }
    
    return userID, true, nil  // 找到记录
}

// Application 层使用示例
func (uc *LoginUseCase) Execute(ctx context.Context, externalID string) error {
    userID, found, err := uc.userRepo.FindIDByExternalID(ctx, externalID)
    if err != nil {
        // 处理系统错误
        return bizerrors.ConvertRepositoryError(ctx, err)
    }
    if !found {
        // 处理"未找到"的业务逻辑 (例如创建新用户)
        return uc.createNewUser(ctx, externalID)
    }
    // 使用找到的 userID
    return uc.processLogin(ctx, userID)
}
```

**两种模式的选择标准**:

| 场景 | 返回值 | 不存在时返回 | 示例 |
|------|--------|------------|------|
| **必须存在的记录** | `(value, error)` | `(nil, NotFoundError)` | `GetByID`, `GetByName` |
| **可能不存在的查询** | `(value, bool, error)` | `(value, false, nil)` | `FindIDByExternalID`, `FindMapping` |

**判断依据**: 
- 记录不存在需要**返回业务错误**给前端 → 用模式 A
- 记录不存在需要**执行后续逻辑**处理 → 用模式 B
```

### 规则 2: 使用 `shared.RepositoryError`,不使用 `*BusinessError`

```go
// ✅ 正确: 返回 RepositoryError
func (r *SqlModelRepo) Save(ctx context.Context, model *DataModel) error {
    params := toCreateParams(model)
    return ExecWithErrorHandling(func() error {
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

### 规则 3: 使用辅助函数处理错误

项目提供了三个核心辅助函数 (`internal/infrastructure/repository/sql_error_analyzer.go`):

```go
// 1. WrapSQLError - 包装任意 SQL 错误
err := r.q.CreateModel(ctx, params)
return WrapSQLError(err)  // 自动分类错误类型

// 2. ExecWithErrorHandling - 用于写操作 (INSERT/UPDATE/DELETE)
return ExecWithErrorHandling(func() error {
    return r.q.CreateModel(ctx, params)
})

// 3. QueryWithSQLErrorHandling - 用于读操作 (SELECT)
var row dbgen.Model
err := QueryWithSQLErrorHandling(func() error {
    var e error
    row, e = r.q.GetModelByID(ctx, id)
    return e
})
```

**错误分类逻辑** (`AnalyzeSQLError`):
- `sql.ErrNoRows` → `shared.ErrRecordNotFound` (sentinel error)
- MySQL 错误 1062 (Duplicate entry) → `shared.ErrTypeDuplicatedKey`
- MySQL 错误 1451/1452 (FK constraint) → `shared.ErrTypeConstraint`
- 其他错误 → 根据错误消息模式匹配分类

---

## Repository 接口实现规范

### 规则 5: 接收 querier 接口,不直接依赖 `*sql.DB`

```go
// ✅ 正确: 使用 querier 接口 (支持事务和非事务)
type SqlModelRepository struct {
    q dbgen.Querier  // 接口，可以是 *sql.DB 或 *sql.Tx
}

func NewSqlModelRepository(q dbgen.Querier) modeldesign.ModelRepository {
    return &SqlModelRepository{q: q}
}

// ❌ 错误: 直接依赖 *sql.DB
type SqlModelRepository struct {
    db *sql.DB  // ❌ 无法支持事务
}
```

**好处**: Application 层可以传入 `*sql.Tx` 实现事务,Repository 层无感知。

---

### 规则 8: 处理 `sql.Null*` 类型

使用项目提供的辅助函数 (`internal/infrastructure/repository/sql_error_analyzer.go`):

```go
// ✅ 正确: 使用辅助函数
func toDomain(row dbgen.Model) *DataModel {
    return &DataModel{
        ID:          row.ID,
        Name:        row.Name,
        Description: NullStrToPtr(row.Description),  // sql.NullString → *string
        Version:     NullInt32ToPtr(row.Version),    // sql.NullInt32 → *int32
        CreatedAt:   row.CreatedAt.Time,             // sql.NullTime → time.Time
        IsActive:    NullBoolToBool(row.IsActive),   // sql.NullBool → bool (false for NULL)
    }
}

func toCreateParams(model *DataModel) dbgen.CreateModelParams {
    return dbgen.CreateModelParams{
        ID:          model.ID,
        Name:        model.Name,
        Description: PtrToNullStr(model.Description),  // *string → sql.NullString
        Version:     PtrToNullInt32(model.Version),    // *int32 → sql.NullInt32
        IsActive:    BoolToNullBool(model.IsActive),   // bool → sql.NullBool
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
- `NullStrToPtr` / `PtrToNullStr` - `sql.NullString` ↔ `*string`
- `NullTimeToPtr` / `PtrToNullTime` - `sql.NullTime` ↔ `*time.Time`
- `NullInt64ToPtr` / `PtrToNullInt64` - `sql.NullInt64` ↔ `*int64`
- `NullInt32ToPtr` / `PtrToNullInt32` - `sql.NullInt32` ↔ `*int32`
- `NullBoolToPtr` / `BoolToNullBool` - `sql.NullBool` ↔ `*bool` / `bool`

### 规则 9: 处理 JSON 字段

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

### 规则 10: Repository 不管理事务

```go
// ✅ 正确: Repository 接收 querier，不关心是否在事务中
func (r *SqlModelRepo) Save(ctx context.Context, model *DataModel) error {
    params := toCreateParams(model)
    return ExecWithErrorHandling(func() error {
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

**事务由 Application 层管理**:

```go
// internal/app/modeldesign/model_usecase.go
func (uc *CreateModelUseCase) Execute(ctx context.Context, input CreateModelInput) error {
    tx, err := uc.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    queries := dbgen.New(tx)  // 使用事务创建 queries
    
    // Repository 无感知，只是接收了一个 querier
    _, err = uc.modelRepo.Save(ctx, queries, model)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## 检查清单

开发 Repository 层时,请确认以下事项:

- [ ] **选择正确的 RecordNotFound 处理模式**:
  - [ ] 必须存在的记录 → 返回 `(value, error)`,不存在时返回 `NotFoundError`
  - [ ] 可能不存在的查询 → 返回 `(value, bool, error)`,不存在时返回 `(value, false, nil)`
- [ ] 使用 `ExecWithErrorHandling` / `QueryWithSQLErrorHandling` 包装 DB 操作
- [ ] 返回 `shared.RepositoryError`,不返回 `*bizerrors.BusinessError`
- [ ] Repository 层对于 `(value, error)` 模式不检查 `IsNotFoundError` (由 Application 层检查)
- [ ] Repository 层对于 `(value, bool, error)` 模式检查 `IsNotFoundError` 并返回 `(value, false, nil)`
- [ ] 接收 `querier` 接口,支持事务和非事务
- [ ] 定义最小 `querier` 接口 (便于测试 mock)
- [ ] 使用辅助函数处理 `sql.Null*` 类型 (`NullStrToPtr` 等)
- [ ] 不在 Repository 层开启事务
- [ ] 不打印错误日志 (只返回 error)
- [ ] 方法命名遵循约定 (`GetByID`, `FindByXXX`, `Save` 等)