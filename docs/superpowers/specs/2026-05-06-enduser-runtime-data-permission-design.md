# End-User Runtime 数据权限体系设计

**日期**: 2026-05-06  
**状态**: 待实现  
**范围**: modelruntime 执行层接入 RBAC，为 end-user 请求执行 Action Gate + Row Filter

---

## 1. 背景与问题

ModelCraft 已有完整的 RBAC 设计层：permission bundle → role → end-user 授权链路，
`RowScope`（SELF/DEPT/DEPT_AND_CHILDREN/ALL）和 `RowPolicy`（select/insert/update/delete 粒度）均已定义。

但 **modelruntime 执行层完全未接入 RBAC**：end-user 发起 GraphQL 请求时，
不论其被授予何种角色，均可对 project 内任意 model 执行任意 CRUD 操作，
且能读写所有行数据。

---

## 2. V1 边界

| 维度 | V1 范围 | 排除 |
|------|---------|------|
| Action Gate | select / insert / update / delete | export（v2） |
| Row Filter | SELF（注入 WHERE）、ALL（不注入） | DEPT、DEPT_AND_CHILDREN（v2） |
| Column Policy | 不实现 | v2 |
| 隐式角色 | 不处理 | v2 |
| Owner 字段识别 | 自动识别 `FormatEndUserRef` 类型字段 | 无需配置 |
| 权限查询时机 | per-request，按 model 粒度查询 | 不缓存至 token |
| Tenant Admin | 跳过所有检查（`endUserID == ""`） | — |

---

## 3. 整体架构

```
GraphQL 请求 (end-user)
        │
        ▼
graphql_app.go::Execute()
  ├─ 1. 解析 modelLocator，获取 RuntimeModel（含 model.ID）
  ├─ 2. 若 CurrentEndUserID != ""：
  │      └─ EndUserPermissionService.Resolve(orgName, projectSlug, endUserID, model.ID)
  │            └─ 查 RBAC → 合并 bundle → 返回 ResolvedModelPermissions
  └─ 3. 注入 graphqlRequestContext（含 EndUserPerms）
        │
        ▼
  model_resolver（query / mutation 闭包）
  ├─ Action Gate：EndUserPerms.CheckAction(action) → PermissionDenied or continue
  └─ Row Filter：
       SELECT/UPDATE/DELETE → 若 IsSelf，注入 WHERE <EndUserRef字段> = $endUserID
       INSERT → 若 IsSelf，自动写入 EndUserRef字段 = $endUserID（已有逻辑，保持不变）
```

---

## 4. 新增组件

| 组件 | 位置 | 职责 |
|------|------|------|
| `ResolvedModelPermissions` | `domain/modelruntime/` | 单次请求的权限快照（4 action × allowed/isSelf） |
| `ActionPermission` | `domain/modelruntime/` | 单个操作的权限状态 |
| `EndUserPermissionService` interface | `domain/modelruntime/` | 依赖倒置，隔离 RBAC 细节 |
| `endUserPermissionServiceImpl` | `app/modelruntime/` | 查 RBAC repo → 合并 bundle → 返回快照 |
| `FindPermissionsByEndUserAndModel` | `domain/rbac/repository.go` | 新增 repo 方法 |
| Action Gate 调用 | `domain/modelruntime/model_resolver.go` | 各 execute* 方法头部 |
| Row Filter 注入 | `domain/modelruntime/model_resolver.go` | 各 execute* 方法 WHERE 条件 |

---

## 5. 数据类型设计

### 5.1 `domain/modelruntime/` 新增

```go
// ResolvedModelPermissions 单次请求的权限快照。
// 在 Execute() 入口解析一次，注入 graphqlRequestContext，resolver 只读。
// nil 表示 tenant admin 请求，跳过所有检查。
type ResolvedModelPermissions struct {
    Select ActionPermission
    Insert ActionPermission
    Update ActionPermission
    Delete ActionPermission
}

// ActionPermission 单个操作的权限状态。
type ActionPermission struct {
    Allowed bool
    IsSelf  bool // true = rowScope=SELF，需注入 WHERE <EndUserRef> = $endUserID
}

// CheckAction 默认拒绝原则。nil receiver = tenant admin，直接放行。
func (p *ResolvedModelPermissions) CheckAction(action Action) error {
    if p == nil {
        return nil
    }
    if !p.Get(action).Allowed {
        return bizerrors.NewError(bizerrors.PermissionDenied, string(action))
    }
    return nil
}

// Get 返回指定 action 的权限状态。
func (p *ResolvedModelPermissions) Get(action Action) ActionPermission {
    switch action {
    case ActionSelect:
        return p.Select
    case ActionInsert:
        return p.Insert
    case ActionUpdate:
        return p.Update
    case ActionDelete:
        return p.Delete
    default:
        return ActionPermission{Allowed: false}
    }
}

// Action 操作类型，与 rbac.Action 对齐，在 domain/modelruntime 内独立定义避免循环依赖。
type Action string

const (
    ActionSelect Action = "SELECT"
    ActionInsert Action = "INSERT"
    ActionUpdate Action = "UPDATE"
    ActionDelete Action = "DELETE"
)

// EndUserPermissionService 依赖倒置接口，app 层实现。
// domain/modelruntime 只依赖此接口，不感知 rbac 包细节。
type EndUserPermissionService interface {
    // Resolve 查询并合并 end-user 在指定 model 上的有效权限。
    // endUserID 为空时（tenant admin）直接返回 nil。
    Resolve(ctx context.Context, orgName, projectSlug, endUserID, modelID string) (*ResolvedModelPermissions, error)
}
```

### 5.2 `graphqlRequestContext` 扩展

```go
type graphqlRequestContext struct {
    ClientRepo       ClientDatabaseRepository
    relationLoaders  map[string]*dataloader.Loader[string, map[string]any]
    OrgName          string
    ProjectSlug      string
    CurrentEndUserID string
    EndUserPerms     *ResolvedModelPermissions // nil = tenant admin，跳过所有检查
}
```

### 5.3 `domain/rbac/repository.go` 新增方法

```go
type Repository interface {
    // ... 原有方法 ...

    // FindPermissionsByEndUserAndModel 查询指定 end-user 在某 model 上的
    // 所有有效权限点（跨所有 bundle）。仅查该 model，不全量拉取。
    FindPermissionsByEndUserAndModel(
        ctx context.Context,
        orgName, projectSlug, endUserID, modelID string,
    ) ([]*EndUserPermission, error)
}
```

---

## 6. RBAC 查询 SQL

```sql
SELECT bp.model_id, bp.row_policy, bp.column_policy
FROM end_user_project_access eupa
JOIN role_bundles rb       ON rb.role_id    = eupa.role_id
JOIN bundle_permissions bp ON bp.bundle_id  = rb.bundle_id
WHERE eupa.org_name     = ?
  AND eupa.project_slug = ?
  AND eupa.end_user_id  = ?
  AND bp.model_id       = ?
```

返回结果交由已有的 `rbac.EffectivePermissionSet{}.Merge()` 合并，
取各 action 最宽 `rowScope`（ALL > SELF），映射为 `ResolvedModelPermissions`：

- `RowScopeAll`  → `ActionPermission{Allowed: true, IsSelf: false}`
- `RowScopeSelf` → `ActionPermission{Allowed: true, IsSelf: true}`
- 无权限记录     → `ActionPermission{Allowed: false}`

---

## 7. Execute() 权限注入

`graphql_app.go::Execute()` 在构建 `graphqlRequestContext` 之前：

```go
// 解析 end-user 权限（仅 end-user 请求，tenant admin 跳过）
var endUserPerms *modelruntime.ResolvedModelPermissions
if endUserID != "" {
    endUserPerms, err = s.permService.Resolve(ctx, orgName, projectSlug, endUserID, model.ID)
    if err != nil {
        return nil, err
    }
}
// 注入 graphqlRequestContext
reqCtx := modelruntime.WithGraphqlRequestContext(ctx, clientRepo, orgName, projectSlug, endUserID, endUserPerms)
```

`GraphqlAppService` 新增字段：
```go
type GraphqlAppService struct {
    modelRepo            modelruntime.ModelRepository
    graphqlSchemaManager *modelruntime.GraphqlSchemaManager
    permService          modelruntime.EndUserPermissionService // 新增
}
```

> **注意**：权限解析依赖 `model.ID`，因此 `Execute()` 需在权限解析前完成 model 加载。
> 当前 `GetSchema()` 内部已加载 model，需重构为先加载 model 再复用于 schema 和权限解析两个步骤。

---

## 8. Action Gate

每个 `execute*` 方法头部统一调用，nil receiver 自动放行（tenant admin）：

| 方法 | Gate Action |
|------|------------|
| `executeCreateOne` / `executeCreateMany` | `ActionInsert` |
| `executeUpdateOne` / `executeUpdateMany` | `ActionUpdate` |
| `executeDeleteOne` / `executeDeleteMany` | `ActionDelete` |
| `executeFindMany` / `executeFindUnique` / `executeFindFirst` | `ActionSelect` |
| `executeAggregate` | `ActionSelect` |

```go
func (m *graphqlModelResolver) executeCreateOne(p graphql.ResolveParams) (interface{}, error) {
    rctx, _ := getGraphqlRequestContext(p.Context)
    if err := rctx.EndUserPerms.CheckAction(ActionInsert); err != nil {
        return nil, err
    }
    // ... 原有逻辑不变
}
```

---

## 9. Row Filter（IsSelf WHERE 注入）

### EndUserRef 字段发现

```go
// graphqlModelResolver helper，遍历一次即可
func (m *graphqlModelResolver) endUserRefFieldName() string {
    for _, f := range m.model.Fields {
        if f.Type != nil && f.Type.Format == modeldesign.FormatEndUserRef {
            return f.Name
        }
    }
    return "" // 该 model 无 owner 字段，SELF scope 无法注入，视为 ALL
}
```

### 各操作注入规则

| 操作 | IsSelf 注入逻辑 |
|------|----------------|
| `executeCreateOne` / `executeCreateMany` | 已有 force-inject 逻辑，**无需改动** |
| `executeFindMany` | `input.Where[ownerField] = endUserID` |
| `executeFindUnique` / `executeFindFirst` | `input.Where[ownerField] = endUserID`；找不到返回 nil（不泄露记录存在性） |
| `executeUpdateOne` | `input.Where[ownerField] = endUserID`（防止越权修改他人记录） |
| `executeDeleteOne` | `input.Where[ownerField] = endUserID`（防止越权删除他人记录） |
| `executeUpdateMany` / `executeDeleteMany` | 同 updateOne / deleteOne |

注入代码模式（以各 action 对应的权限字段为准）：

```go
// SELECT（findMany / findUnique / findFirst）
if rctx.EndUserPerms != nil && rctx.EndUserPerms.Select.IsSelf {
    if ownerField := m.endUserRefFieldName(); ownerField != "" {
        input.Where[ownerField] = rctx.CurrentEndUserID
    }
}

// UPDATE（updateOne / updateMany）
if rctx.EndUserPerms != nil && rctx.EndUserPerms.Update.IsSelf {
    if ownerField := m.endUserRefFieldName(); ownerField != "" {
        input.Where[ownerField] = rctx.CurrentEndUserID
    }
}

// DELETE（deleteOne / deleteMany）
if rctx.EndUserPerms != nil && rctx.EndUserPerms.Delete.IsSelf {
    if ownerField := m.endUserRefFieldName(); ownerField != "" {
        input.Where[ownerField] = rctx.CurrentEndUserID
    }
}
```

---

## 10. 错误处理

| 场景 | 行为 |
|------|------|
| Action 被拒绝 | `bizerrors.PermissionDenied` → GraphQL error 返回给客户端 |
| RBAC repo 查询失败 | 返回 internal error，请求中断 |
| model 无 EndUserRef 字段但 rowScope=SELF | 降级为 ALL（无法注入 WHERE，不报错） |
| endUserID 为空（tenant admin） | `EndUserPerms = nil`，所有 gate 和 filter 跳过 |

---

## 11. 不在 V1 范围内

- **Column policy**（字段级可见性 / 可编辑性）
- **DEPT / DEPT_AND_CHILDREN rowScope**（需部门层级树支撑）
- **隐式角色接入**（`isImplicit=true` 的内置角色）
- **权限变更实时推送**（TTL 内 token 不感知变更，当前 per-request 查询已规避此问题）
- **Export action**

---

## 12. 文件改动清单

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `domain/modelruntime/permission.go`（新建） | 新增 | `ResolvedModelPermissions`、`ActionPermission`、`EndUserPermissionService` |
| `domain/modelruntime/graphql_request_context.go` | 修改 | 新增 `EndUserPerms` 字段，更新构造函数签名 |
| `domain/rbac/repository.go` | 修改 | 新增 `FindPermissionsByEndUserAndModel` 方法 |
| `app/modelruntime/permission_service.go`（新建） | 新增 | `endUserPermissionServiceImpl` 实现 |
| `app/modelruntime/graphql_app.go` | 修改 | 注入 `permService`，Execute() 内权限解析，重构 model 加载顺序 |
| `domain/modelruntime/model_resolver.go` | 修改 | 各 `execute*` 方法加 Action Gate + Row Filter |
| `infrastructure/repository/rbac_repo.go`（或同目录） | 修改 | 实现新 repo 方法，编写 SQL |
