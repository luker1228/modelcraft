---
name: modelruntime-dev
description: |
  ModelCraft modelruntime 模块开发指南。负责 GraphQL 运行时引擎的设计、实现与维护。
  当用户涉及以下任何内容时，必须使用此 skill：
  - 修改或新增 modelruntime 模块代码（internal/domain/modelruntime/, internal/app/modelruntime/）
  - 关联字段查询（many-to-one / one-to-many）、N+1 问题、dataloader
  - GraphQL Schema 构建、Schema 缓存、resolver 生命周期
  - graphqlRequestContext、ClientDatabaseRepository、FindManyIn
  - RuntimeModel 字段、LogicalForeignKey、关系字段解析
  - 动态 GraphQL Schema 生成（根据数据库模型动态构建）
---

# ModelRuntime 模块开发指南

## 模块职责

`modelruntime` 是 ModelCraft 的 GraphQL 运行时引擎：根据用户定义的数据模型（`RuntimeModel`）**动态生成** GraphQL Schema，执行 findMany/findUnique/findFirst/aggregate/count/CRUD 操作，并解析 many-to-one / one-to-many 关联关系。

**关键路径**：
```
HTTP 请求
  → GraphqlAppService.Execute()           [internal/app/modelruntime/graphql_app.go]
    → GetSchema() → GraphqlSchemaManager  [internal/domain/modelruntime/graphqlschema_manager.go]
    → WithGraphqlRequestContext()         [internal/domain/modelruntime/graphql_request_context.go]
    → graphql.Do(schema, reqCtx)
      → graphqlModelResolver 闭包          [internal/domain/modelruntime/model_resolver.go]
        → getGraphqlRequestContext(p.Context)
        → rctx.ClientRepo.FindXxx()
```

---

## 核心设计原则：Schema 与请求状态解耦

**这是本模块最重要的架构约束，修改时必须严格遵守。**

### 问题背景

`graphqlModelResolver` 构建 Schema 时会注册大量 resolver 闭包。这些闭包如果捕获请求级状态（DB 连接、dataloader），在 Schema 被缓存后会导致：
- 跨请求数据泄露（A 请求的 DB 连接被 B 请求的查询使用）
- Dataloader batch 队列混合（不同请求的 FK 值合并查询）
- Data race

### 解决方案：两层分离

```
可缓存（Schema 类型结构）              请求级（每次 graphql.Do 独立）
─────────────────────────────         ─────────────────────────────
graphqlModelResolver                  graphqlRequestContext
  model                                ClientRepo     ← DB 连接
  enumConfigMap                        relationLoaders ← dataloader map
  inputTypeGenerator
  modelRepo
  lfkRepo
  // 不持有任何 context 字段
```

**graphqlModelResolver 是无状态的 Schema 构建器**，不持有任何 `context.Context` 字段。构建阶段所需的 ctx（用于日志和 Repository 查询）作为参数从 `newGraphqlSchema(ctx)` 开始沿调用链透传：

```
newGraphqlSchema(ctx)
  → createModelType(ctx)
    → generateModelType(ctx, ...)
      → createRelationField(ctx, ...)
        → lfkRepo.GetByID(ctx, ...)     ← 用传入的 ctx
        → modelRepo.GetByID(ctx, ...)
```

**Resolver 闭包内获取请求级状态的唯一方式**：
```go
rctx, _ := getGraphqlRequestContext(p.Context)
// 然后使用 rctx.ClientRepo.FindXxx(p.Context, ...)
```

**App 层注入（每次请求调用一次）**：
```go
reqCtx := modelruntime.WithGraphqlRequestContext(ctx, clientRepo)
graphql.Do(graphql.Params{Context: reqCtx, ...})
```

### Schema 缓存安全性

`GraphqlSchemaManager.StoreSchema` / `GetByName` 目前是 TODO（空实现）。**实现时可以安全缓存**：`graphqlModelResolver` 不持有任何 context 或请求级资源，缓存 `*graphql.Schema`（其中 resolver 闭包捕获了 `r *graphqlModelResolver`）是完全安全的。多个并发请求可以共享同一个 Schema 对象。

---

## Dataloader：解决 N+1 问题

### 工作原理（graphql-go 广度优先执行）

graphql-go 的执行分两个阶段：

```
Phase 1 — 广度优先收集（串行）
  items[0].tl resolver → loader.Load(ctx, "v1") → Thunk1  ← 只入队
  items[1].tl resolver → loader.Load(ctx, "v2") → Thunk2  ← 只入队
  items[2].tl resolver → loader.Load(ctx, "v3") → Thunk3  ← 只入队

Phase 2 — dethunk
  Thunk1() → dataloader 触发 → SELECT * FROM slave WHERE id IN (v1,v2,v3)
  Thunk2() → 从缓存返回
  Thunk3() → 从缓存返回
```

resolver 返回 `func() (interface{}, error)` 类型时，graphql-go 将其识别为 thunk 并延迟执行，**这是 dataloader 在串行执行模型下能工作的关键**。

### 实现位置

| 文件 | 职责 |
|------|------|
| `relation_loader.go` | `newRelationBatchLoader()` — 创建 dataloader，batch 函数调用 `FindManyIn` |
| `graphql_request_context.go` | `getOrCreateLoader()` — 懒初始化，per-(tableName/referenceKey) 复用 |
| `model_resolver.go` | `createManyToOneResolverFromFK()` — 调用 `loader.Load()` 返回 Thunk |

### Many-to-one Resolver 模式

```go
func (r *graphqlModelResolver) createManyToOneResolverFromFK(...) graphql.FieldResolveFn {
    foreignKey := lf.SourceFields[0]
    referenceKey := lf.TargetFields[0]

    return func(p graphql.ResolveParams) (interface{}, error) {
        rctx, _ := getGraphqlRequestContext(p.Context)

        // 获取 FK 值 ...
        fkStr, ok := toString(foreignKeyValue)
        if !ok {
            // fallback 到单条查询
        }

        // 懒获取 loader（同一请求内复用）
        loader := rctx.getOrCreateLoader(refModelName, referenceKey)

        // 返回 Thunk，不立即执行
        thunk := loader.Load(p.Context, fkStr)
        return func() (interface{}, error) {
            result, err := thunk()
            if err != nil {
                return nil, nil  // 悬空外键按 LEFT JOIN 语义返回 nil
            }
            return result, nil
        }, nil
    }
}
```

### 悬空外键处理

FK 值在目标表中不存在时（如 `test = "aa"` 但 `slave` 表无此记录），按 **LEFT JOIN 语义** 返回 `nil`，**不报错**。这是有意设计，对应 SQL 的 `LEFT JOIN` 行为。

---

## 关键文件速查

### Domain 层 (`internal/domain/modelruntime/`)

| 文件 | 说明 |
|------|------|
| `model_resolver.go` | 核心，约 1600 行。Schema 生成 + 所有 resolver 实现 |
| `graphql_request_context.go` | 请求级状态容器，`WithGraphqlRequestContext` / `getGraphqlRequestContext` |
| `relation_loader.go` | `newRelationBatchLoader` — dataloader batch 函数 |
| `graphqlschema_manager.go` | Schema 管理器，`NewSchemaFrom` / 缓存接口（TODO） |
| `graphql_repository.go` | `ClientDatabaseRepository` 接口定义 |
| `graphql_input.go` | 所有 Input 类型定义（FindManyInput、FindManyInInput 等） |
| `runtimemodel.go` | `RuntimeModel` / `RuntimeField` 数据结构 |

### App 层 (`internal/app/modelruntime/`)

| 文件 | 说明 |
|------|------|
| `graphql_app.go` | `Execute()` 入口，创建 clientRepo，注入 `graphqlRequestContext` |

### Infrastructure 层

| 文件 | 说明 |
|------|------|
| `internal/infrastructure/database/dml/client_db_repo_impl.go` | `ClientDatabaseRepository` 实现，含 `FindManyIn` |
| `internal/infrastructure/database/dml/sql_mapper.go` | SQL 生成，`convertFindManyInInputToSQL` 用 goqu `IN` |

---

## graphqlModelResolver 结构体字段说明

```go
// 无状态的 Schema 构建器，不持有任何 context.Context 字段。
// 可以安全缓存（闭包不捕获任何可变状态）。
type graphqlModelResolver struct {
    model              *RuntimeModel    // 当前模型定义
    enumConfigMap      map[string]*graphqlEnumConfig
    inputTypeGenerator *inputTypeGenerator
    modelRepo          ModelRepository
    lfkRepo            modeldesign.LogicalForeignKeyRepository
    // 无 ctx、无 clientRepo、无 relationLoaders —— 全部是请求级的
}
```

**构建阶段**：ctx 作为参数从 `newGraphqlSchema(ctx)` 透传到所有构建方法。  
**执行阶段**：从 `p.Context` 取 `rctx`，用 `rctx.ClientRepo` 和 `rctx.getOrCreateLoader()`。

---

## 新增/修改功能的标准模式

### 新增一个查询操作

1. 在 `model_resolver.go` 中新增 `executeXxx(p graphql.ResolveParams)` 方法
2. 方法内第一行：`rctx, _ := getGraphqlRequestContext(p.Context)`
3. 使用 `rctx.ClientRepo.Xxx(p.Context, input)`
4. 新增对应的 `createXxxField()` 方法，Resolve 闭包内调用 `executeXxx`
5. 在 `createRootQuery(ctx, modelType)` 或 `createRootMutation` 中注册

### 修改 Schema 构建逻辑

Schema 构建方法（`createModelType(ctx)`、`generateModelType(ctx, ...)`、`createRelationField(ctx, ...)` 等）签名均含 `ctx context.Context` 参数，从 `newGraphqlSchema(ctx)` 透传而来。**绝对不能在构建方法内使用 `p`**（构建阶段没有 `p`），也不能在结构体上存 ctx 字段。

### 新增关联关系类型

关联关系的 resolver 必须：
1. 从 `p.Context` 取 `rctx`
2. 通过 `rctx.getOrCreateLoader()` 获取 dataloader
3. 返回 Thunk（`func() (interface{}, error)`）而非直接返回结果

---

## 并发安全说明

| 组件 | 安全性 | 原因 |
|------|--------|------|
| `graphql.Schema` | 并发安全 | 构建后只读，resolver 闭包不捕获可变状态 |
| `graphqlModelResolver` | **并发安全** | 无状态，不持有任何 context 或可变字段，可跨请求共享 |
| `graphqlRequestContext` | 单请求内安全 | graphql-go 串行执行 resolver，无并发写 |
| `dataloader.Loader` | 安全 | 库内部用 channel 同步 |
| `ClientDBRepoImpl` | 安全 | 底层 `*sqlx.DB` 支持并发 |

---

## 常见陷阱

1. **在 `createXxxField` 闭包外用 `p.Context`**：`p` 在 Schema 构建阶段不存在，会编译报错。构建方法接收 `ctx context.Context` 参数，用传入的 ctx，不要从 `p` 取。

2. **Schema 缓存后 resolver 状态泄露**：若在 Schema 构建时（而非 resolver 执行时）初始化 dataloader 或持有 DB 连接，缓存后会跨请求共享。始终在 resolver 闭包内通过 `getGraphqlRequestContext(p.Context)` 获取。

3. **Many-to-one resolver 直接返回结果而非 Thunk**：会退化为 N+1 查询，失去 dataloader 批量效果。resolver 必须返回 `func() (interface{}, error)`。

4. **`FindManyIn` 的 IN 列表为空**：`sql_mapper.go` 对空列表返回错误，调用前需确保 `len(values) > 0`。
