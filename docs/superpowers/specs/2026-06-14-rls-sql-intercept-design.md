# RLS SQL Intercept 设计

> 2026-06-14 | feature/rls-context-open-data-api

## 动机

当前 RLS（Row Level Security）逻辑分散在 `model_resolver.go` 的各个 execute 方法中。
每个操作（FindMany、CreateOne、UpdateMany 等）都各自调用：

1. `appendRLSUsingFilter()` — 将 CEL USING 表达式注入 `RawFilters`
2. `RLSPolicyGuard.ValidateInput()` — 校验 CEL CHECK 表达式

这种分散式设计导致：
- Resolver 层知道太多 RLS 细节
- "SQL 执行前拦截" 这个核心概念没有明确的代码表达
- 难以测试、难以扩展

## 核心思路

在请求入口（`Execute()`）一次性完成 RLS 策略解析与预编译，构建 `RLSPolicySnapshot`，
存入 context。SQL 执行层通过 `RLSInterceptDB` wrapper 从 context 读取 snapshot，
在 SQL 执行前透明注入 USING 过滤条件或校验 CHECK 表达式。

**DenyAll 在入口短路**：如果没有任何匹配策略，在 `Execute()` 中直接返回
`PermissionDenied`，不进入 GraphQL 执行。

## 架构

```
Execute()
  │
  ├─ buildRLSSnapshot(orgName, projectSlug, modelID)
  │   ├─ 从 ctx 获取 end-user identity + UserContext
  │   ├─ 若 developer → 返回 nil（不应用 RLS）
  │   ├─ 调用 PolicyMatchingService 解析各 action 的 USING/CHECK
  │   ├─ USING → RawSQLFilter（3 个 action：Select / Update / Delete）
  │   ├─ CHECK → cel.Program（预编译，2 个 action：Insert / Update）
  │   └─ 无匹配策略 → DenyAll
  │
  ├─ DenyAll? → return PermissionDenied（短路）
  │
  ├─ clientRepo = RLSInterceptDB(ClientDBRepoImpl)
  ├─ ctx = WithRLSSnapshot(ctx, snap)
  └─ graphql.Do(schema, query)
        │
        Resolver（不感知 RLS）
          │
          RLSInterceptDB
            ├─ 读操作 → snap.SelectUSING → input.RawFilters
            ├─ INSERT  → snap.InsertCHECK.Eval(input.Data)
            ├─ UPDATE  → snap.UpdateUSING → RawFilters
            │            snap.UpdateCHECK.Eval(input.Data)
            └─ DELETE  → snap.DeleteUSING → RawFilters
              │
              inner.ClientDBRepoImpl → sql_mapper.go → db.Query()
```

## 组件

### 1. RLSPolicySnapshot（`domain/modelruntime/rls_snapshot.go`）

```go
type RLSPolicySnapshot struct {
    SelectUSING *RawSQLFilter  // SELECT 的 USING WHERE 条件
    UpdateUSING *RawSQLFilter  // UPDATE 的 USING WHERE 条件
    DeleteUSING *RawSQLFilter  // DELETE 的 USING WHERE 条件
    InsertCHECK cel.Program    // INSERT 的预编译 CHECK 表达式
    UpdateCHECK cel.Program    // UPDATE 的预编译 CHECK 表达式
    DenyAll     bool           // 无匹配策略，拒绝所有操作
}

func WithRLSSnapshot(ctx context.Context, snap *RLSPolicySnapshot) context.Context
func getRLSSnapshot(ctx context.Context) *RLSPolicySnapshot
```

- USINGS 为 nil 表示无需注入（developer 或 true 表达式）
- CHECK 为 nil 表示无需校验
- `DenyAll = true` 时，`Execute()` 直接短路，snapshot 不进入 context

### 2. RLSSnapshotBuilder（`app/modelruntime/rls_snapshot_builder.go`）

```go
type RLSSnapshotBuilder interface {
    Build(ctx context.Context, orgName, projectSlug, modelID string) (*modelruntime.RLSPolicySnapshot, error)
}
```

实现逻辑：
1. 从 ctx 获取 `EndUserIdentity`
2. 若 `IsDeveloper()` → 返回 nil（RLS 不适用）
3. 从 ctx 获取 `UserContext`（HTTP header 注入的身份变量）
4. 调用 `PolicyMatchingService.ResolveUsing()` 解析 3 个 action 的 USING → `RawSQLFilter`
5. 调用匹配服务获取 CHECK 原始表达式字符串（需新增 `ResolveCheck` 方法或复用现有接口）
6. 用 `cel.Env` 预编译 CHECK 表达式为 `cel.Program`
7. 若无任何匹配策略 → 返回 `DenyAll = true`

### 3. RLSInterceptDB（`infrastructure/database/dml/rls_intercept_db.go`）

```go
type RLSInterceptDB struct {
    inner ClientDatabaseRepository
}

func NewRLSInterceptDB(inner ClientDatabaseRepository) *RLSInterceptDB
```

实现 `ClientDatabaseRepository` 的全部 13 个方法，按操作类型处理：

| 方法 | USING 注入 | CHECK 校验 |
|------|-----------|-----------|
| FindUnique / FindFirst / FindMany / ListByCursor / Aggregate / Count | SelectUSING | — |
| FindManyIn | — | — |
| CreateOne / CreateMany | — | InsertCHECK |
| UpdateOne / UpdateMany | UpdateUSING | UpdateCHECK |
| DeleteOne / DeleteMany | DeleteUSING | — |

CHECK 校验示例：

```go
func (r *RLSInterceptDB) CreateOne(ctx context.Context, input *CreateOneInput) (string, error) {
    snap := getRLSSnapshot(ctx)
    if snap != nil && snap.InsertCHECK != nil {
        out, _, err := snap.InsertCHECK.Eval(map[string]any{
            "input": input.Data,
            "auth":  getAuthContext(ctx),
        })
        if err != nil || !out.Value().(bool) {
            return "", bizerrors.NewError(bizerrors.PermissionDenied, "RLS CHECK violation")
        }
    }
    return r.inner.CreateOne(ctx, input)
}
```

USING 注入示例：

```go
func (r *RLSInterceptDB) FindMany(ctx context.Context, input *FindManyInput) ([]map[string]any, error) {
    snap := getRLSSnapshot(ctx)
    if snap != nil && snap.SelectUSING != nil {
        input.RawFilters = append(input.RawFilters, *snap.SelectUSING)
    }
    return r.inner.FindMany(ctx, input)
}
```

### 4. ClientDatabaseRepository 接口（不变）

`RLSInterceptDB` 直接实现现有接口，无需修改。

## Execute() 变更（`app/modelruntime/graphql_app.go`）

```go
func (s *GraphqlAppService) Execute(ctx context.Context, ...) (*graphql.Result, error) {
    // ... 现有：GetSchema, 创建 DB 连接, 提取 endUserID, 解析 endUserPerms ...

    clientRepo := dml.NewRLSInterceptDB(dml.NewClientDB(clientSqlDB))

    // ★ 新增：构建 RLS snapshot
    snap, err := s.snapshotBuilder.Build(ctx, orgName, projectSlug, modelID)
    if err != nil {
        return nil, err
    }
    if snap != nil && snap.DenyAll {
        return nil, bizerrors.NewError(bizerrors.PermissionDenied, "RLS: no matching policy")
    }

    // 注入 snapshot 到 ctx（intercept 层消费）
    ctx = modelruntime.WithRLSSnapshot(ctx, snap)

    reqCtx := modelruntime.WithGraphqlRequestContext(
        ctx, clientRepo, orgName, projectSlug, endUserID, endUserAdminID, endUserPerms,
    )
    // ... graphql.Do ...
}
```

**变更点：**
- `clientRepo` 用 `RLSInterceptDB` 包裹
- 调用 `snapshotBuilder.Build()` 构建 snapshot
- DenyAll 在此短路
- snapshot 注入 ctx
- **删除** `rlsGuard RuntimeRLSPolicyGuard` 字段、`WithRLSPolicyGuard()` 调用

## Resolver 清理（`domain/modelruntime/model_resolver.go`）

**删除：**
- `appendRLSUsingFilter()` 方法（约 20 行）
- 所有 execute 方法中调用 `appendRLSUsingFilter()` 的行
- 所有 execute 方法中调用 `rctx.RLSPolicyGuard.ValidateInput()` 的行

**不变：**
- `CheckAction()` — 权限检查仍在 resolver 层
- `BuildRowFilter()` — RBAC rowScope 过滤仍在 resolver 层
- `enforceOwnerOnCreate()` — owner 字段注入仍在 resolver 层

## graphql_request_context.go 清理

**删除：**
- `RLSPolicyGuard` 接口定义
- `WithRLSPolicyGuard()` 函数
- `graphqlRequestContext.RLSPolicyGuard` 字段

## 数据流

```
HTTP Request
  │  X-User-Type, X-User-Id, X-User-Roles, ...
  │
  ▼
Middleware: 解析 header → ctx (EndUserIdentity, UserContext)
  │
  ▼
HandleQuery: orgName, projectSlug, db, model 从 URL 提取 → runtimeContext
  │
  ▼
GraphqlAppService.Execute()
  │  1. snapshotBuilder.Build(ctx, orgName, projectSlug, modelID)
  │     ├─ 从 ctx 读 EndUserIdentity / UserContext
  │     ├─ PolicyMatchingService.ResolveUsing() ×3 → RawSQLFilter
  │     └─ cel.Env.Compile() ×2 → cel.Program
  │  2. WithRLSSnapshot(ctx, snap)
  │  3. graphql.Do(schema, query)
  │
  ▼
Resolver (executeFindMany / executeCreateOne / ...)
  │  权限检查、owner 注入（不变）
  │  调用 clientRepo.FindMany(ctx, input)（不传 RLS 参数）
  │
  ▼
RLSInterceptDB
  │  从 ctx 读 snapshot
  │  注入 USING → RawFilters / Eval CHECK
  │  委托 inner.FindMany(ctx, input)
  │
  ▼
ClientDBRepoImpl
  │  convertFindManyInputToSQL() → sql_mapper.go
  │  applyRawFilters(ds, input.RawFilters)  ← USING 在这里生效
  │  db.Queryx(sql, args)
```

## 错误处理

| 场景 | 处理位置 | 返回 |
|------|---------|------|
| 无匹配策略 (DenyAll) | `Execute()` 入口 | `PermissionDenied` |
| CHECK 表达式编译失败 | `snapshotBuilder.Build()` | 内部错误 |
| CHECK 校验不通过 | `RLSInterceptDB` | `PermissionDenied: RLS CHECK violation` |
| USING 解析失败 | `snapshotBuilder.Build()` | 内部错误 |
| Developer JWT | `snapshotBuilder.Build()` | 返回 nil snapshot，RLS 不生效 |

## 测试策略

### 单元测试

1. **RLSInterceptDB** — mock `ClientDatabaseRepository` 作为 inner
   - 验证 USING 注入到 `input.RawFilters`
   - 验证 CHECK 校验拦截
   - 验证 nil snapshot（developer）时不拦截
   - 验证 DenyAll 场景

2. **RLSSnapshotBuilder** — mock `PolicyMatchingService`
   - 验证 developer 返回 nil
   - 验证 DenyAll 场景
   - 验证预编译 CHECK 程序

### 集成测试

- 端到端验证 end-user 查询走完整 RLS 拦截链路

### 删除的测试

- `model_resolver.go` 中 `appendRLSUsingFilter` 相关的测试用例

## 迁移步骤

1. 新建 `RLSPolicySnapshot` 类型 + context helper
2. 新建 `RLSSnapshotBuilder` 接口 + 实现
3. 新建 `RLSInterceptDB`
4. 修改 `GraphqlAppService.Execute()` — 构建 snapshot + 使用 `RLSInterceptDB`
5. 清理 `model_resolver.go` — 删除 RLS 相关代码
6. 清理 `graphql_request_context.go` — 删除 `RLSPolicyGuard` 接口
7. 清理 `graphql_app.go` — 删除 `rlsGuard` 字段
8. 更新 wire 注入
9. 运行测试，确保无回归

## 不在此范围的变更

- RBAC rowScope 过滤（`BuildRowFilter`）— 保持在 resolver 层
- 权限检查（`CheckAction`）— 保持在 resolver 层
- `FindManyIn` RLS 支持 — 后续迭代考虑
- CEL 表达式编辑 / 策略 CRUD — 已有独立设计
