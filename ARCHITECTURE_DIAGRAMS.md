# ModelCraft 后端 - 架构可视化

## 1️⃣ 两态架构全景图

```
┌────────────────────────────────────────────────────────────────────────────┐
│                          ModelCraft 架构总览                                │
└────────────────────────────────────────────────────────────────────────────┘

                     ┌─────────────────────────────────┐
                     │      ModelCraft 用户界面         │
                     │    (Web 前端 / 移动应用)        │
                     └──────────────┬──────────────────┘
                                    │
                   ┌────────────────┴────────────────┐
                   │                                 │
                   ▼                                 ▼
        ┌─────────────────────┐        ┌─────────────────────┐
        │   设计态 GraphQL     │        │   运行态 GraphQL    │
        ├─────────────────────┤        ├─────────────────────┤
        │ /design/graphql     │        │ /:orgName/:project  │
        │ （静态 Schema）     │        │ /：db/:model        │
        │                     │        │ （动态 Schema）     │
        └────────┬────────────┘        └────────┬────────────┘
                 │                               │
        ┌────────▼───────────────┐      ┌────────▼──────────────┐
        │ JWT 验证 + 权限检查    │      │ JWT 验证 + 权限检查   │
        │ (chi_jwt_auth)        │      │ (@hasPermission)      │
        └────────┬───────────────┘      └────────┬──────────────┘
                 │                               │
                 ▼                               ▼
        ┌─────────────────────┐        ┌─────────────────────┐
        │  设计态 Repository  │        │ 运行态 App Service  │
        │  (模型/字段/集群)   │        │ (graphql_app.go)    │
        └────────┬────────────┘        └────────┬────────────┘
                 │                               │
                 │                      ┌────────▼────────────┐
                 │                      │ ClusterManager      │
                 │                      │ 获取客户 DB 连接    │
                 │                      └────────┬────────────┘
                 │                               │
                 ▼                               ▼
        ┌──────────────────────────────────────────────────┐
        │   ModelCraft 元数据数据库 (MySQL)                │
        │  ┌─────────────┬─────────────┬──────────────┐   │
        │  │ org_name    │ project_slug│ database_name│ ←─┼─ WHERE 条件隔离
        │  └─────────────┴─────────────┴──────────────┘   │
        │                                                   │
        │  organizations, projects, models, users, roles   │
        │  permissions, user_organizations, ...            │
        └──────────────────────────────────────────────────┘
                                │
                                ▼
                    ┌─────────────────────────┐
                    │  客户各自的数据库实例    │
                    │  (多个 MySQL 实例)      │
                    │                         │
                    │  ┌──────────────────┐  │
                    │  │ Instance 1       │  │
                    │  │ host1:port1      │  │
                    │  │ ├─ database_a    │  │  Project A 数据
                    │  │ └─ database_b    │  │  Project B 数据
                    │  └──────────────────┘  │
                    │                         │
                    │  ┌──────────────────┐  │
                    │  │ Instance 2       │  │
                    │  │ host2:port2      │  │
                    │  │ └─ database_c    │  │  Project C 数据
                    │  └──────────────────┘  │
                    │                         │
                    └─────────────────────────┘
```

---

## 2️⃣ Runtime GraphQL 执行流程

```
运行态请求: POST /:orgName/:projectSlug/:database/:model
           Query: { findMany { items { id name } } }
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ 1. 认证 + 权限检查                                      │
├─────────────────────────────────────────────────────────┤
│ • 从 Authorization 头提取 JWT                           │
│ • 验证签名（HMAC-SHA256）                              │
│ • 检查权限：@hasPermission(action: "model:read")       │
│   - 快速路径：JWT context 中有权限                    │
│   - 回退路径：查询数据库确认                          │
│ • 验证 orgName 与 JWT.Organization 匹配               │
└─────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ 2. 获取或构建 GraphQL Schema                            │
├─────────────────────────────────────────────────────────┤
│ • 查询 graphqlSchemaManager 缓存（key: modelLocator）   │
│   ├─ 缓存 hit? → 返回已缓存的 Schema                   │
│   └─ 缓存 miss? → 构建新 Schema（见下图）             │
│                                                         │
│ 构建过程：                                              │
│   1. 从设计态 DB 查询 RuntimeModel（模型定义）        │
│   2. graphqlModelResolver.newGraphqlSchema()            │
│      • createModelType()     → 生成模型类型            │
│      • createRootQuery()     → 生成 Query 操作        │
│      • createRootMutation()  → 生成 Mutation 操作     │
│   3. 缓存到 graphqlSchemaManager                       │
└─────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ 3. 创建请求级上下文                                    │
├─────────────────────────────────────────────────────────┤
│ • ClusterConnectionManager.GetConnectionWithDatabase()  │
│   ├─ 查询 database_clusters 表（org_name, project_slug)│
│   ├─ 读取连接配置（host, port, username, password）    │
│   ├─ 建立 TCP 连接到客户数据库                        │
│   └─ 执行 USE `database` 切换数据库                   │
│                                                         │
│ • WithGraphqlRequestContext(ctx, clientDB)             │
│   ├─ 创建 graphqlRequestContext                       │
│   │  ├─ ClientRepo（DB 连接）                         │
│   │  └─ relationLoaders map（dataloader 缓存）        │
│   └─ 注入 context.WithValue()                         │
└─────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ 4. 执行 GraphQL 查询                                    │
├─────────────────────────────────────────────────────────┤
│ graphql.Do(Params{                                       │
│     Schema: *gschema,            ← 无状态 Schema      │
│     RequestString: query,        ← GraphQL 查询       │
│     VariableValues: variables,   ← 查询变量           │
│     Context: reqCtx,             ← 包含请求级状态    │
│ })                                                       │
│                                                         │
│ • graphql-go 框架执行 resolver 闭包                     │
│ • 每个 resolver 从 p.Context 读取 graphqlRequestContext│
│ • 通过 ClientRepo 执行 SQL 查询                        │
└─────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ 5. SQL 生成与执行                                       │
├─────────────────────────────────────────────────────────┤
│ GraphQL Query → SQL Mapper → SQL + Args → 执行         │
│                                                         │
│ findMany { items { id name where { age: { $gt: 18 } }}} │
│     ↓                                                    │
│ convertFindManyInputToSQL()                             │
│     ↓                                                    │
│ // WHERE 条件转换                                      │
│ whereExpr := convertWhereToExpression(input.Where)     │
│     ├─ 检查是否包含复杂操作符 ($and, $or, $gt 等)    │
│     ├─ 若包含 → ParseAndConvert() 处理 AST            │
│     └─ 否则 → goqu.Ex(where) 直接转换                 │
│     ↓                                                    │
│ dialectWrapper.Select(...).From(tableName).Where(...)   │
│     ↓                                                    │
│ sql: "SELECT * FROM users WHERE age > ?"               │
│ args: [18]                                              │
│     ↓                                                    │
│ clientDB.Queryx(sql, args...)  ← 执行实际 SQL         │
│     ↓                                                    │
│ map[string]any{                                         │
│     "id": 1, "name": "Alice", "age": 25                │
│ }                                                        │
└─────────────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────┐
│ 6. 处理关系字段（N+1 解决方案）                        │
├─────────────────────────────────────────────────────────┤
│ 如果查询包含关系字段：                                  │
│   { items { id name author { id name } } }             │
│                                                         │
│ • Resolver 调用 dataloader.Load(authorID)              │
│ • graphql-go 广度优先遍历，所有 Load() 被聚合         │
│ • 同一请求内相同 loader 复用，Load 调用被合并         │
│ • 执行单条 SQL：                                       │
│   SELECT * FROM authors WHERE id IN (id1, id2, id3)    │
│     ↓                                                    │
│ • 返回 authors 数据填充关系字段                        │
└─────────────────────────────────────────────────────────┘
                │
                ▼
         返回 GraphQL Result
         (JSON 格式)
```

---

## 3️⃣ Schema 和状态分离设计

```
┌─────────────────────────────────────────────────────────┐
│           GraphQL Schema（无状态 - 可缓存）             │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────────────────────────────────┐          │
│  │ ModelType                               │          │
│  │ ├─ id: ID!                              │          │
│  │ ├─ name: String!                        │          │
│  │ ├─ author: Author  ← resolver 闭包      │          │
│  │ └─ ...                                  │          │
│  └─────────────────────────────────────────┘          │
│                                                         │
│  ┌─────────────────────────────────────────┐          │
│  │ Query                                   │          │
│  │ ├─ findUnique(where: ...) ← resolver   │          │
│  │ ├─ findMany(where: ...) ← resolver      │          │
│  │ └─ ...                                  │          │
│  └─────────────────────────────────────────┘          │
│                                                         │
└─────────────────────────────────────────────────────────┘
         │
         │ 缓存在内存中
         │ 跨多个请求复用
         │
         ├─ Request 1 ─────────────────────────────────┐
         │                                              │
         │ ┌────────────────────────────────────────┐  │
         │ │  graphqlRequestContext (Request 1)    │  │
         │ │                                        │  │
         │ │  • ClientDB 连接 1                     │  │
         │ │  • relationLoaders 缓存 1              │  │
         │ │                                        │  │
         │ │  graphql.Do(params) 中：              │  │
         │ │  • Schema: 共享的全局 Schema          │  │
         │ │  • Context: 包含 graphqlRequestContext│  │
         │ │                                        │  │
         │ │  Resolver 闭包执行：                   │  │
         │ │  func(p) {                             │  │
         │ │    rctx, _ := getGraphqlRequestContext│  │
         │ │    result := rctx.ClientDB.FindUnique()│  │
         │ │  }                                      │  │
         │ └────────────────────────────────────────┘  │
         │                                              │
         ├─ Request 2 ─────────────────────────────────┐
         │                                              │
         │ ┌────────────────────────────────────────┐  │
         │ │  graphqlRequestContext (Request 2)    │  │
         │ │                                        │  │
         │ │  • ClientDB 连接 2                     │  │
         │ │  • relationLoaders 缓存 2              │  │
         │ │                                        │  │
         │ │  graphql.Do(params) 中：              │  │
         │ │  • Schema: 同一个共享 Schema          │  │
         │ │  • Context: 包含不同的 graphqlRequestContext
         │ │                                        │  │
         │ │  完全隔离，无竞态条件！               │  │
         │ └────────────────────────────────────────┘  │
         │                                              │
         └──────────────────────────────────────────────┘
```

---

## 4️⃣ 多租户隔离架构

```
┌────────────────────────────────────────────────────────┐
│            Organization 顶层租户容器                    │
│                  (org_name: "acme")                    │
├────────────────────────────────────────────────────────┤
│                                                        │
│ ┌──────────────────────────────────────────────────┐ │
│ │  Project: Sales App                              │ │
│ │  (slug: "sales", org_name: "acme")              │ │
│ │                                                   │ │
│ │  ┌──────────────────────────────────────────┐   │ │
│ │  │ DatabaseCluster                          │   │ │
│ │  │ Host: db1.example.com                    │   │ │
│ │  │ Port: 3306                               │   │ │
│ │  │ Database: acme_sales                     │   │ │
│ │  └──────────────────────────────────────────┘   │ │
│ │           │                                      │ │
│ │           ▼                                      │ │
│ │  ┌──────────────────────────────────────────┐   │ │
│ │  │ Models: Users, Orders, Products         │   │ │
│ │  └──────────────────────────────────────────┘   │ │
│ └──────────────────────────────────────────────────┘ │
│                                                        │
│ ┌──────────────────────────────────────────────────┐ │
│ │  Project: Support App                            │ │
│ │  (slug: "support", org_name: "acme")            │ │
│ │                                                   │ │
│ │  ┌──────────────────────────────────────────┐   │ │
│ │  │ DatabaseCluster                          │   │ │
│ │  │ Host: db2.example.com ← 不同的主机      │   │ │
│ │  │ Port: 3306                               │   │ │
│ │  │ Database: acme_support                   │   │ │
│ │  └──────────────────────────────────────────┘   │ │
│ │           │                                      │ │
│ │           ▼                                      │ │
│ │  ┌──────────────────────────────────────────┐   │ │
│ │  │ Models: Tickets, Conversations           │   │ │
│ │  └──────────────────────────────────────────┘   │ │
│ └──────────────────────────────────────────────────┘ │
│                                                        │
│ Users: [Alice, Bob, Charlie]                         │
│   ├─ Alice: Role = Owner    → Permissions: *        │
│   ├─ Bob: Role = Editor      → Permissions: model:* │
│   └─ Charlie: Role = Viewer  → Permissions: model:read
│                                                        │
└────────────────────────────────────────────────────────┘
```

**数据隔离保证**：
- 元数据隔离：WHERE org_name='acme' AND project_slug='sales'
- 客户数据隔离：db1.example.com 的 acme_sales 数据库
- 权限隔离：每个用户有不同的角色和权限

---

## 5️⃣ 权限检查流程

```
┌─────────────────────────────────────────┐
│ GraphQL 操作                             │
│ @hasPermission(action: "model:read")    │
└────────────┬────────────────────────────┘
             │
             ▼
    ┌─────────────────────┐
    │ 验证 Context        │
    │ • UserID            │
    │ • OrgName           │
    └────────┬────────────┘
             │
             ▼
    ╔═════════════════════════════════╗
    ║ 快速路径（推荐）               ║
    ║ 从 JWT Claims 中读取权限        ║
    ║ （无数据库 I/O）               ║
    ╚═════════════╦═══════════════════╝
                  │
          ┌───────┴────────┐
          │                │
          ▼                ▼
    权限存在？         权限不存在？
          │                │
          ✅ 通过          ▼
          │        ╔═══════════════════════╗
          │        ║ 回退路径（降级方案）   ║
          │        ║ 查询数据库验证权限     ║
          │        ║ （有数据库 I/O）      ║
          │        ╚═════════╦═════════════╝
          │                  │
          │          ┌───────┴─────────┐
          │          │                 │
          │          ▼                 ▼
          │     权限存在？         权限不存在？
          │          │                 │
          │          ✅ 通过           ❌ 拒绝
          │          │                 │
          └──────────┬─────────────────┘
                     │
                     ▼
          ┌──────────────────┐
          │ 执行 Resolver    │
          │ 返回数据         │
          └──────────────────┘
```

---

## 6️⃣ 连接管理和缓存

```
                    ClusterConnectionManager
                    ┌────────────────────┐
                    │ connections (Map)  │
                    │ [cluster_id]       │
                    │  → *sql.DB        │
                    └────────┬───────────┘
                             │
                    ┌────────▼──────────┐
                    │  连接生命周期      │
                    └────────┬──────────┘
                             │
                   ┌─────────┴──────────┐
                   │                    │
                   ▼                    ▼
          GetConnection()      GetConnectionWithDatabase()
          ├─ 查询缓存            ├─ 查询缓存
          ├─ 缓存 hit? → 返回    ├─ 缓存 hit? → 返回
          ├─ 缓存 miss?          ├─ 缓存 miss?
          │  ├─ 从 DB 查询         │  ├─ 从 DB 查询
          │  │  cluster 配置       │  │  cluster 配置
          │  ├─ 建立连接           │  ├─ 建立连接
          │  ├─ 设置连接池参数     │  ├─ 执行 USE `database`
          │  └─ 缓存              │  └─ 缓存
          │                        │
          └─────────┬──────────────┘
                    │
                    ▼
          ┌─────────────────────┐
          │ 连接池配置          │
          │ MaxOpenConns: 100   │
          │ MaxIdleConns: 10    │
          │ MaxLifetime: 3600s  │
          └─────────────────────┘
```

---

## 7️⃣ 权限数据模型

```
┌────────────────────────────────────────┐
│            User                        │
│  id: UUID                              │
│  external_id: Casdoor User ID         │
│  name, email                          │
└────────────┬─────────────────────────┘
             │
             │ 1:N
             │
             ▼
┌────────────────────────────────────────┐
│     UserOrganization                   │
│  (Membership)                          │
│  user_id, org_name                    │
│  status: active|suspended|invited      │
└────────────┬─────────────────────────┘
             │
             │ 1:N
             │
             ▼
┌────────────────────────────────────────┐
│        UserRole                        │
│  user_id, role_id                     │
└────────────┬─────────────────────────┘
             │
             │ 1:N
             │
             ▼
┌────────────────────────────────────────┐
│          Role                          │
│  id, org_name, name                    │
│  permissions: []string                │
│  ├─ "model:read"                      │
│  ├─ "model:write"                     │
│  ├─ "project:*"                       │
│  └─ ...                                │
└────────────────────────────────────────┘
```

---

## 8️⃣ HTTP 请求生命周期

```
HTTP Request
    │
    ├─ Headers: Authorization: Bearer <JWT>
    │
    ▼
┌──────────────────────────────────────┐
│ Chi Router                           │
│ GET /:orgName/:projectSlug/:db/:model
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ chi_jwt_auth Middleware              │
│ • 提取 JWT token                     │
│ • 验证签名                           │
│ • 提取 claims                        │
│ • 注入 context: SetUserID()          │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ GraphQL Handler                      │
│ • 从 URL 提取 orgName/project/db/model
│ • 验证 orgName 与 JWT 匹配          │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ @hasPermission Directive             │
│ • 从 context 读取 userID             │
│ • 从 JWT 快速路径检查权限            │
│ • 或查询数据库回退路径               │
│ • 权限检查失败 → PERMISSION_DENIED   │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ GraphQL Execute                      │
│ graphql.Do(Params{...})              │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ Resolver 闭包                         │
│ • 从 context 读取 graphqlRequestContext
│ • 通过 ClientDB 执行 SQL             │
│ • 处理 N+1（dataloader）             │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ 返回 GraphQL Result (JSON)            │
└──────────────────────────────────────┘
```

---

## 总结

| 特性 | 实现 | 优势 |
|------|------|------|
| **Schema 和状态分离** | 缓存 Schema + 请求级 context | 高并发 + 无竞态 |
| **权限检查** | JWT 快速路径 + DB 回退路径 | 性能 + 灵活性 |
| **数据隔离** | 应用层（元数据）+ 物理隔离（用户数据） | 多租户 + 安全 |
| **N+1 解决** | DataLoader per 请求 | 性能优化 |
| **连接管理** | 缓存 + 连接池 | 资源高效 |
