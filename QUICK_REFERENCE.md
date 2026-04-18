# ModelCraft 后端 - 快速查询表

## 🎯 常见问题速查

### Q1: Runtime GraphQL 如何动态执行查询？
**A**: Schema 和状态分离设计
- **Schema**（无状态）: 首次构建后缓存，跨请求复用
- **状态**（有状态）: 每次请求创建独立 context，包含 DB 连接和 dataloader
- **执行**: graphql.Do() 通过 context 将 Resolver 连接到数据库

**关键文件**:
- `model_resolver.go` - Schema 构建
- `graphql_request_context.go` - 状态管理

---

### Q2: JWT Token 包含哪些信息？
**A**: ModelCraftClaims 结构

```go
UserID           // 用户 ID
ExternalID       // Casdoor 用户 ID
Name / Email
Organization     // 当前组织
Roles[]          // ["owner", "editor"]
Permissions[]    // ["model:read", "model:write"]
Memberships[]    // 多组织成员身份（最多 10 个）
ExpiresAt
Issuer = "modelcraft"
```

**来源**: `internal/domain/auth/modelcraft_claims.go`

---

### Q3: 权限检查如何工作？
**A**: 两层策略

1. **快速路径**（无 DB）: 从 JWT Claims 直接检查
2. **回退路径**（有 DB）: 查询数据库的 user_roles → roles → permissions

**关键**:
- 权限格式: `{resource}:{action}` (如 `model:read`)
- 指令: `@hasPermission(action: "model:read")`
- 实现: `internal/interfaces/graphql/project/directives.go`

---

### Q4: 数据如何隔离？
**A**: 两个维度

| 维度 | 位置 | 隔离方式 | 表字段 |
|------|------|--------|--------|
| 元数据 | ModelCraft 共享 DB | 应用层（WHERE） | org_name, project_slug |
| 用户数据 | 客户数据库实例 | 物理隔离（不同 host/port） | 无（在客户 DB） |

**连接获取**:
```go
// 1. 查询 database_clusters 表（按 org_name, project_slug）
// 2. 建立连接到客户数据库
// 3. 执行 USE `databaseName`
// 4. 执行用户查询
```

---

### Q5: 不同 Project 的数据存储在同一数据库实例吗？
**A**: 不一定

**可能的部署方式**:
- **方式 1**（集中）: Project A 和 B 在同一 MySQL Instance，不同数据库名
- **方式 2**（分布）: Project A 和 B 在完全不同的 MySQL Instance

**隔离保证**: 
- 即使在同一 Instance，也通过 `USE database_name` 隔离
- Project 的 database_clusters 配置指定具体的 host:port:database

---

### Q6: 如何处理 N+1 查询问题？
**A**: DataLoader 方案

```go
// 请求级 context 持有 dataloader 实例 map
graphqlRequestContext {
    relationLoaders map[string]*dataloader.Loader
}

// 同一请求内的相同关系字段加载被聚合
// SELECT * FROM authors WHERE id IN (id1, id2, id3)
```

**关键文件**:
- `graphql_request_context.go` - loader 管理
- `relation_loader.go` - loader 实现

---

## 🗂️ 核心模块位置

| 功能 | 文件 | 说明 |
|------|------|------|
| GraphQL 执行 | `app/modelruntime/graphql_app.go` | Execute/GetSchema 入口 |
| Schema 生成 | `domain/modelruntime/model_resolver.go` | 动态 Schema 构建 |
| 请求上下文 | `domain/modelruntime/graphql_request_context.go` | DB 连接和 loader 管理 |
| SQL 生成 | `infrastructure/database/dml/sql_mapper.go` | WHERE 条件转 SQL |
| CRUD 操作 | `infrastructure/database/dml/client_db_repo_impl.go` | 执行实际 SQL |
| 连接管理 | `infrastructure/repository/cluster_connection_manager.go` | 连接池和缓存 |
| JWT 验证 | `middleware/chi_jwt_auth.go` | 认证中间件 |
| 权限检查 | `interfaces/graphql/project/directives.go` | @hasPermission 指令 |

---

## 📊 数据库表速查

### 元数据表（ModelCraft 共享 DB）

| 表名 | 用途 | 关键字段 |
|------|------|--------|
| organizations | 组织 | `name (PK)`, owner_id, status |
| projects | 项目 | `(org_name, slug) (PK)`, cluster_id |
| database_clusters | 数据库集群配置 | `id (PK)`, org_name, project_slug, host, port, password |
| models | 模型定义 | `id (PK)`, org_name, project_slug, database_name, name |
| users | 用户 | `id (PK)`, external_id, name, phone |
| user_organizations | 用户-组织关联 | `id (PK)`, user_id, org_name, status |
| roles | 角色 | `id (PK)`, org_name, name, permissions[] |
| user_roles | 用户角色关联 | `user_id`, `role_id` |

### 客户数据表（客户 DB）
由客户数据库实例托管，结构由 ModelCraft 根据模型定义自动创建。

---

## 🔑 权限列表

| 权限 | 含义 |
|-----|------|
| `model:read` | 读取模型 |
| `model:create` | 创建模型 |
| `model:update` | 更新模型 |
| `model:delete` | 删除模型 |
| `model:write` | 写入/编辑模型 |
| `project:*` | Project 所有操作 |
| `*` | 超级权限 |

---

## 🚀 常见代码路径

### 添加新的 GraphQL Query
```
1. 定义在 model_resolver.go 的 createRootQuery()
2. 添加 converter 函数（如 convertFindUniqueInputToSQL）
3. 添加 resolver 函数（executeFindUnique）
```

### 修改权限检查
```
1. 在 GraphQL Schema 添加 @hasPermission(action: "...")
2. 权限检查由 directives.go 自动处理
3. 支持 JWT context 和数据库回退
```

### 支持新的 SQL 方言（如 PostgreSQL）
```
1. cluster_connection_manager.go - 修改连接字符串
2. sql_mapper.go - goqu.Dialect("postgres")
3. runtimemodel.go - 字段类型映射
4. 驱动库 - 改用 github.com/lib/pq
```

---

## 🎓 架构设计模式

### 1. Schema 和状态分离
- ✅ Schema（纯数据结构）可缓存
- ✅ 请求状态（DB 连接）按请求隔离
- ✅ Resolver 闭包通过 context 连接

### 2. Repository 模式
- ✅ Domain 层定义接口（ModelRepository）
- ✅ Infrastructure 层实现（ClientDBRepoImpl）
- ✅ 便于单元测试

### 3. 权限检查两层策略
- ✅ 快速路径：JWT（无 DB I/O）
- ✅ 回退路径：数据库（有 DB I/O）
- ✅ 平衡性能和正确性

### 4. DataLoader 批量加载
- ✅ 解决 N+1 查询问题
- ✅ 请求级实例，同一请求复用
- ✅ 支持广度优先遍历聚合

---

## ⚠️ 注意事项

### 数据隔离
- ❗ 所有查询必须加上 org_name/project_slug WHERE 条件
- ❗ 客户数据在各自的数据库实例中，物理隔离
- ❗ 不要跨租户查询

### 连接管理
- ❗ 使用 ClusterConnectionManager 获取连接，不要手动创建
- ❗ 连接自动缓存和复用
- ❗ 密码自动加密存储

### 权限检查
- ❗ 所有敏感操作都要加 @hasPermission 指令
- ❗ 权限优先从 JWT 读取，回退到数据库
- ❗ 权限验证是第一道防线

---

## 🔗 相关文档

- 完整报告: `BACKEND_DISCOVERY_REPORT.md`
- 执行摘要: `BACKEND_EXECUTIVE_SUMMARY.md`
- 架构文档: `ai-metadata/backend/design/`
- 核心原则: `ai-metadata/backend/design/core-principles.md`
