# ModelCraft 后端探索 - 执行摘要

## 🎯 核心发现

### 1. Runtime 模块如何执行动态查询（CRUD 操作）

**关键设计**：**Schema 和状态分离**

```
GraphQL Schema（无状态）              请求级上下文（有状态）
├─ 类型定义                          ├─ ClientDB（数据库连接）
├─ Resolver 函数                     ├─ DataLoader 实例
└─ 可跨请求缓存                      └─ 每次请求隔离创建

                    ▼ Resolver 闭包通过 context 读取 ▼
```

**执行流程**：

1. **Schema 构建**（首次）
   - 从设计态数据库获取 RuntimeModel（模型定义 + 字段）
   - `graphqlModelResolver.newGraphqlSchema()`
     - `createModelType()` → 生成包含所有字段的模型类型
     - `createRootQuery()` → findUnique, findFirst, findMany, count
     - `createRootMutation()` → create, update, delete

2. **缓存 Schema**
   - 使用 `graphqlSchemaManager` 缓存，跨请求复用

3. **请求执行**
   ```go
   // 每次请求：
   clientDB := ClusterConnectionManager.GetConnectionWithDatabase(
       ctx, orgName, projectSlug, databaseName
   )
   ctx = WithGraphqlRequestContext(ctx, clientDB)  // 注入请求级状态
   result := graphql.Do(Params{
       Schema: *gschema,           // 无状态 Schema
       Context: ctx,               // 包含 DB 连接的 context
   })
   ```

4. **SQL 生成和执行**
   ```
   GraphQL 输入 → Goqu SQL 映射器 → SQL string + args
                  ↓
   支持简单条件：{"id": "123"}
   支持复杂条件：{"$and": [{"age": {"$gte": 18}}]}
                  ↓
   clientDB.QueryRowx(sql, args...) → map[string]any → JSON 返回
   ```

**关键源文件**：
- `internal/domain/modelruntime/model_resolver.go` - Schema 生成核心
- `internal/app/modelruntime/graphql_app.go` - Execute 方法
- `internal/infrastructure/database/dml/sql_mapper.go` - SQL 映射

---

### 2. 当前的认证/授权模型

**认证流程**：
```
用户 → Casdoor 登录 → 签发 JWT (Casdoor 公钥)
                ↓
       ModelCraft 中间件 (chi_jwt_auth.go)
                ↓
       验证签名 → 提取 UserID/ExternalID
                ↓
       注入 context → 下游使用
```

**JWT Claims 包含什么**：
```go
ModelCraftClaims {
    // 用户身份
    UserID     string    // ModelCraft 内部 UUID
    ExternalID string    // Casdoor 用户 ID
    Name       string
    Email      string

    // 组织和权限
    Organization string      // 当前组织
    Roles        []string    // ["owner", "editor"]
    Permissions  []string    // ["model:read", "model:write", "project:*"]

    // 多组织成员身份（最多 10 个）
    Memberships  []MembershipClaimInfo
    HasMoreMemberships bool

    // 标准 JWT 字段
    ExpiresAt  time.Time
    Issuer     "modelcraft"  // 固定值
}
```

**权限检查**（两层策略）：
```
@hasPermission(action: "model:read")
    ↓
快速路径（无 DB I/O）：从 JWT context 读取权限
    ✅ 如果 JWT 中包含权限列表 → 直接检查 → 快速返回
    
回退路径（有 DB I/O）：查询数据库
    ❌ 如果 JWT 中无权限 → 查询 user_roles / roles / permissions 表 → 检查
```

**权限格式**：`{resource}:{action}`
```
model:read          # 读取模型
model:write         # 写入模型
model:create        # 创建模型
model:delete        # 删除模型
project:*           # Project 下所有操作（通配符）
*                   # 超级权限
```

**关键源文件**：
- `internal/middleware/chi_jwt_auth.go` - JWT 验证
- `internal/domain/auth/modelcraft_claims.go` - Claims 结构
- `internal/interfaces/graphql/project/directives.go` - @hasPermission 指令

---

### 3. 数据隔离的实现方式

**两个维度**：

#### 应用层隔离（元数据）
```sql
-- ModelCraft 共享数据库中的所有表都有这些字段
org_name      VARCHAR(36)  -- 隔离键 1
project_slug  VARCHAR(64)  -- 隔离键 2
database_name VARCHAR(64)  -- 隔离键 3（仅限 models 表）

-- 查询时必须加上 WHERE 条件
SELECT * FROM models 
WHERE org_name = ? 
  AND project_slug = ? 
  AND database_name = ?
```

#### 物理隔离（客户数据）
```
Project A ────┐
              ├─ Customer MySQL Instance 1
Project B ────┤  (host1:port1, database1)
              │

Project C ────┐
              ├─ Customer MySQL Instance 2
Project D ────┤  (host2:port2, database2)
              │
```

**隔离保证**：
- Project A 的数据在 Customer Instance 1 的 database1
- Project B 的数据在 Customer Instance 1 的 database1
- Project C 的数据完全在不同的 MySQL server（Instance 2）
- 不同 Project 的数据**物理分离**，即使在同一 MySQL server 也用不同数据库名

**关键源文件**：
- `internal/infrastructure/repository/cluster_connection_manager.go` - 连接获取
- `db/schema/mysql/` - 所有元数据表

---

### 4. 用户定义模型数据是否存储在同一 database instance 中？

**答案**：❌ **不一定**

**设计**：
- **元数据**（模型定义、字段、关系）→ ModelCraft **共享** MySQL 实例
- **用户数据**（实际业务数据）→ 客户各自管理的 MySQL 实例（可不同）

**场景示例**：

| 组织 | 项目 | 数据库实例 | 数据库名 | 存储位置 |
|-----|------|----------|---------|---------|
| Org1 | Proj-A | Customer Instance 1 | db_a | Instance 1 的 db_a |
| Org1 | Proj-B | Customer Instance 1 | db_b | Instance 1 的 db_b |
| Org1 | Proj-C | Customer Instance 2 | db_c | Instance 2 的 db_c |

**隔离等级**：
- **最弱**：同一 MySQL server，不同数据库名（应用层 `USE database_name` 隔离）
- **最强**：完全独立的 MySQL server（主机、端口、数据库名都不同）

**连接获取**：
```go
// 步骤 1：根据 orgName + projectSlug 查询 database_clusters 表
cluster := repo.GetByProjectKey(orgName, projectSlug)
// 返回：{ host: "host1", port: 3306, username: "user", password: "xxx" }

// 步骤 2：建立 TCP 连接到客户数据库
conn := createConnection(cluster.ConnectionInfo)

// 步骤 3：切换到指定数据库
conn.ExecContext(ctx, "USE `databaseName`")

// 步骤 4：后续 SQL 查询在此数据库中执行
```

---

## 🗂️ 关键文件路径

### Runtime 模块
```
runtime 核心逻辑
├── internal/domain/modelruntime/
│   ├── model_resolver.go               ★ Schema 生成 + 查询执行
│   ├── graphql_request_context.go      ★ 请求级上下文（DB 连接、dataloader）
│   ├── relation_loader.go              ★ N+1 问题处理（dataloader）
│   ├── runtimemodel.go                 - 模型结构定义
│   └── graphql_*.go                    - 工具类
│
├── internal/app/modelruntime/
│   └── graphql_app.go                  ★ Execute/GetSchema 入口
│
└── internal/infrastructure/database/dml/
    ├── client_db_repo_impl.go          ★ CRUD 实现（FindUnique/FindMany/Create/Update/Delete）
    ├── sql_mapper.go                   ★ SQL 生成（WHERE 条件转 SQL）
    └── query_parser.go                 - 复杂条件解析
```

### 认证和权限
```
auth & rbac
├── internal/domain/auth/
│   ├── modelcraft_claims.go            ★ JWT Claims 结构
│   ├── user_claims.go                  - 简化 Claims
│   └── project_auth_config.go          - Provider 配置
│
├── internal/middleware/
│   └── chi_jwt_auth.go                 ★ JWT 中间件
│
├── internal/domain/
│   ├── membership/membership.go        - 用户-组织关联
│   ├── role/role.go                    - 角色定义
│   └── permission/permission.go        - 权限值对象
│
└── internal/interfaces/graphql/project/
    └── directives.go                   ★ @hasPermission 指令
```

### 数据库和集群
```
database & cluster
├── internal/infrastructure/repository/
│   └── cluster_connection_manager.go   ★ 连接管理（连接池、缓存）
│
└── db/schema/mysql/
    ├── 01_project.sql                  - 项目表
    ├── 02_database_cluster.sql         - 集群配置表
    ├── 03_model_domain.sql             - 模型定义表
    ├── 05_organizations.sql            - 组织表
    ├── 06_users.sql                    - 用户表
    └── 07_roles_permissions.sql        - 角色权限表
```

---

## 🔌 PostgreSQL 连接方式

**当前状态**：
- ✅ MySQL 完整实现
- ❌ PostgreSQL 未支持
- ⏳ 预留架构支持未来扩展

**架构扩展性**：
```
connection_factory.go       - 未来可支持多方言
    ↓
goqu (SQL 生成器)          - 已支持 postgres dialect
    ├─ goqu.Dialect("mysql")
    └─ goqu.Dialect("postgres")  // 未来可用

sql.Open("mysql", dsn)     - 改为 sql.Open("postgres", dsn)
    ↓
驱动库                      - github.com/lib/pq（已有 Go 库）
```

**迁移要点**：
1. 连接字符串格式改变
2. SQL 方言改为 "postgres"
3. 类型映射调整（DATETIME → TIMESTAMP 等）
4. 数据库操作方式改变（USE 不存在）

---

## 📊 架构对比

| 方面 | 设计态（Design-time） | 运行态（Runtime） |
|------|----------------------|------------------|
| GraphQL 入口 | `/org/{orgName}/design/graphql` | `/{orgName}/{projectSlug}/{db}/{model}` |
| Schema 类型 | 静态（.graphql 文件定义） | 动态（根据模型定义生成） |
| 数据库 | ModelCraft 共享 MySQL | 客户各自管理的 MySQL 实例 |
| 隔离方式 | 应用层（WHERE 条件） | 物理隔离（不同数据库实例） |
| 目标 | 管理元数据 | CRUD 客户业务数据 |
| 部署 | 必须与后端一起 | 可独立部署 |

---

## 💡 设计优势总结

✅ **灵活性**
- Schema 缓存支持高并发
- 请求级状态隔离避免竞态条件
- 支持简单和复杂查询条件

✅ **安全性**
- 应用层 + 数据库层双重数据隔离
- 权限检查两层策略（快速 + 回退）
- JWT 签名验证 + 权限检查

✅ **多租户**
- 完善的组织-项目-集群层级关系
- 每个 Project 可绑定独立的数据库实例
- 混合支持集中存储和分布式存储

✅ **可维护性**
- 清晰的分层架构（Domain/App/Infrastructure）
- Repository 模式便于单元测试
- Goqu SQL 生成避免手写 SQL

---

## 📝 完整报告

详细版本已保存至：
```
modelcraft-backend/BACKEND_DISCOVERY_REPORT.md
```

包含：
- 完整的 Runtime 执行流程图
- 所有表结构详细说明
- 完整的代码示例
- 未实现的功能列表
- PostgreSQL 迁移指南

