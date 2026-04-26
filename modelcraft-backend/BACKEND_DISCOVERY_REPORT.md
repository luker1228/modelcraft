# ModelCraft 后端架构详细发现报告

生成时间：2026-04-17
探索范围：Runtime 模块、RBAC 权限设计、数据库层结构、PostgreSQL 连接方式

---

## 目录

1. [核心架构概览](#核心架构概览)
2. [Runtime 模块设计](#runtime-模块设计)
3. [RBAC 和权限管理](#rbac-和权限管理)
4. [认证系统](#认证系统)
5. [数据库层结构](#数据库层结构)
6. [数据隔离机制](#数据隔离机制)
7. [多租户设计](#多租户设计)
8. [关键路径汇总](#关键路径汇总)

---

## 核心架构概览

### 项目定位

ModelCraft 是**面向开发团队的低代码数据模型管理平台**，核心价值链为：

```
设计态 (Design-time)                运行态 (Runtime)
─────────────────────               ────────────────────
用户在 ModelCraft 中                 客户应用通过 GraphQL
定义模型、字段、关联关系               消费数据，CRUD 操作
        │                                    ▲
        ▼                                    │
  同步到目标 MySQL 数据库 ────────────────────┘
```

**两个独立 GraphQL 入口**：
- **设计态 GraphQL**: `/org/modelcraft/design/graphql` - 静态 Schema，管理元数据（模型、字段、集群）
- **运行态 GraphQL**: `/:orgName/:projectSlug/:db/:model` - 动态 Schema，查询和操作客户数据

---

## Runtime 模块设计

### 1. Runtime 模块位置

```
modelcraft-backend/
├── internal/
│   ├── domain/modelruntime/        # 领域层
│   │   ├── runtimemodel.go         # 运行时模型定义
│   │   ├── model_resolver.go       # GraphQL Schema 生成和查询执行
│   │   ├── model_repository.go     # 模型仓储接口
│   │   ├── graphql_*.go            # GraphQL 相关工具（输入类型、标量、字段条件等）
│   │   ├── graphql_request_context.go  # 请求级上下文管理
│   │   ├── relation_loader.go      # 关系加载（处理 N+1 问题）
│   │   └── ...
│   └── app/modelruntime/           # 应用层
│       ├── graphql_app.go          # GraphQL 应用服务
│       └── commands.go             # 命令定义
│
└── internal/infrastructure/database/
    └── dml/                        # 数据操作层
        ├── client_db_repo_impl.go  # 客户端 DB 仓储实现
        ├── sql_mapper.go           # SQL 映射（WHERE 条件 -> SQL）
        ├── query_parser.go         # 查询解析器（处理复杂条件）
        └── ...
```

### 2. Runtime 模型结构 (`RuntimeModel`)

```go
type RuntimeModel struct {
    OrgName      string                   // 组织名称
    ProjectSlug  string                   // 项目标识
    Name         string                   // 模型名称
    Title        string                   // 模型标题
    Description  string                   // 模型描述
    DatabaseName string                   // 数据库名称
    DisplayField *string                  // 用于 _displayName 的字段名
    Fields       map[string]*RuntimeField // 字段映射（字段名 -> 字段定义）
}

// RuntimeField 是 modeldesign.FieldDefinition 的别名
```

### 3. 动态 GraphQL Schema 生成流程

```
请求入口: Execute(ctx, orgName, projectSlug, databaseName, modelName, query)
    │
    ▼
GetSchema()  // 从缓存或新建 GraphQL Schema
    │
    ├─ 检查 graphqlSchemaManager 缓存
    ├─ 若无缓存，通过 modelRepo 从设计态数据库获取 RuntimeModel
    └─ 调用 graphqlSchemaManager.NewSchemaFrom() 生成 Schema
         │
         ▼
    newGraphqlModelResolver()
         │
         ├─ createModelType()        // 生成模型类型（包含所有字段）
         ├─ createRootQuery()        // 生成 Query（findUnique, findFirst, findMany, count 等）
         ├─ generateModelTypeSkipRelation()  // 生成不含关系字段的模型（避免递归）
         └─ createRootMutation()     // 生成 Mutation（create, update, delete 等）
    │
    ▼
缓存 Schema 到 graphqlSchemaManager
    │
    ▼
创建请求级上下文
    ├─ 从 ClusterConnectionManager 获取客户数据库连接
    ├─ 执行 USE `databaseName` 切换到指定数据库
    └─ 创建 ClientDB 仓储实例
    │
    ▼
通过 graphql.Do() 执行查询
    │
    ├─ Resolver 从 p.Context 读取 graphqlRequestContext（包含 ClientRepo、dataloader）
    ├─ 执行实际 SQL 查询
    └─ 处理关系字段（通过 dataloader 聚合 N+1 查询）
```

### 4. Schema 类型结构 vs 请求级状态的分离

**关键设计原则**：Schema 类型结构（纯数据结构）和请求级状态（DB连接、数据加载）完全解耦。

```go
// Schema 生成阶段（无状态）
newGraphqlModelResolver(ctx, model, modelRepo, lfkRepo)
    └─ 返回 *graphql.Schema（只含类型定义和 Resolve 函数闭包，不持有 context）

// 请求执行阶段（有状态）
Execute(ctx, ..., cmd)
    ├─ 获取 ClientDB 连接
    ├─ 注入到 context: ctx = WithGraphqlRequestContext(ctx, clientRepo)
    └─ graphql.Do(Params{
        Schema: *gschema,  // 无状态的 Schema
        Context: reqCtx,   // 包含请求级状态的 context
    })

// Resolver 闭包读取请求级状态
Resolve: func(p graphql.ResolveParams) (interface{}, error) {
    rctx, _ := getGraphqlRequestContext(p.Context)  // 从 context 取出请求级上下文
    result := rctx.ClientRepo.FindUnique(p.Context, input)  // 使用请求级 DB 连接
    return result, nil
}
```

**优势**：
- Schema 可安全缓存，跨请求复用
- 每个请求有独立的 DB 连接、dataloader 实例
- 避免并发冲突

### 5. 动态 SQL 执行（CRUD 操作）

#### FindUnique 查询流程

```go
// 1. 输入转换为 SQL
convertFindUniqueInputToSQL(ctx, input)
    ├─ 获取 WHERE 条件：input.Where（map[string]interface{}）
    ├─ 转换为 goqu 表达式：convertWhereToExpression(where)
    │   ├─ 检查是否包含复杂操作符（AND, OR, NOT 等）
    │   ├─ 若包含，调用 ParseAndConvert() 处理复杂语法树
    │   └─ 否则，直接使用 goqu.Ex(where) 转换
    ├─ 构建 SQL: dialectWrapper.Select(...).From(tableName).Where(whereExpr)
    └─ 返回 (sql string, args []any)

// 2. 执行 SQL 并获取结果
row := c.stdDB.QueryRowx(sql, args...)
err = row.MapScan(result)  // 自动映射为 map[string]any

// 3. 后处理
result = convertBytesToString(result)  // 字节转字符串
```

#### WHERE 条件解析

**简单条件**（向后兼容）：
```graphql
{
  "id": "123",
  "name": "Alice"
}
```
→ SQL: `WHERE id = '123' AND name = 'Alice'`

**复杂条件**（支持操作符）：
```graphql
{
  "$and": [
    { "age": { "$gte": 18 } },
    { "city": { "$in": ["Beijing", "Shanghai"] } }
  ]
}
```
→ 通过 `ParseAndConvert()` 构建 AST，支持操作符：
- 逻辑操作符：`$and`, `$or`, `$not`
- 比较操作符：`$eq`, `$ne`, `$gt`, `$gte`, `$lt`, `$lte`, `$in`, `$nin`, `$like`, `$between` 等

#### 其他查询类型

| 查询类型 | 功能 | SQL 生成 |
|---------|------|---------|
| `findFirst` | 查找第一条记录（带 WHERE/ORDER BY/LIMIT 1） | `SELECT * FROM table WHERE ... ORDER BY ... LIMIT 1` |
| `findMany` | 查找多条记录（带分页） | `SELECT * FROM table WHERE ... LIMIT ? OFFSET ?` |
| `findManyIn` | 批量查找（IN 子句，处理 N+1） | `SELECT * FROM table WHERE refKey IN (...)` |
| `count` | 计数 | `SELECT COUNT(*) FROM table WHERE ...` |

#### 变更操作 (Mutation)

- `create`: INSERT 语句
- `update`: UPDATE 语句（支持按主键或唯一字段定位）
- `delete`: DELETE 语句
- `updateMany`: 批量更新
- `deleteMany`: 批量删除

### 6. 处理 N+1 问题：DataLoader

**问题场景**：
```graphql
query {
  findMany {
    items {
      id
      name
      author {  # 这会导致 N+1 问题（每条记录加载作者）
        id
        name
      }
    }
  }
}
```

**解决方案**：使用 `graphql-go/dataloader`

```go
type graphqlRequestContext struct {
    ClientRepo ClientDatabaseRepository
    relationLoaders map[string]*dataloader.Loader[string, map[string]any]  // 关系 loader 缓存
}

// 请求级上下文
func WithGraphqlRequestContext(ctx context.Context, clientRepo ClientDatabaseRepository) context.Context {
    rctx := newGraphqlRequestContext(clientRepo)
    return context.WithValue(ctx, graphqlRequestContextKey{}, rctx)
}

// 懒初始化 loader
func (rctx *graphqlRequestContext) getOrCreateLoader(tableName, referenceKey string) *dataloader.Loader[...] {
    key := tableName + "/" + referenceKey
    if l, ok := rctx.relationLoaders[key]; ok {
        return l  // 同一请求内复用
    }
    l := newRelationBatchLoader(rctx.ClientRepo, tableName, referenceKey)
    rctx.relationLoaders[key] = l
    return l
}
```

**工作原理**：
1. 每条记录的 `author` 字段 resolver 调用 `loader.Load(authorID)`
2. graphql-go 广度优先遍历，所有同层的 `Load()` 调用被聚合
3. 同一请求内相同的 loader 实例，相同的 Load 调用被合并
4. 最终执行单条 SQL: `SELECT * FROM authors WHERE id IN (id1, id2, id3, ...)`

---

## RBAC 和权限管理

### 1. 权限系统概览

```
User ──── Membership ────▶ Organization
                    │
                    ▼
                 UserRole ────▶ Role ────▶ Permission[]
```

### 2. 核心实体

#### Membership（用户-组织关联）
```go
type Membership struct {
    ID         string    // UUID
    UserID     string    // 用户 ID
    OrgID      string    // 组织 ID
    OrgName    string    // 组织名称（冗余，避免 JOIN）
    Status     string    // active | suspended | invited
    InvitedBy  string    // 邀请人 ID
    InvitedAt  *time.Time
    JoinedAt   *time.Time
}

// 生命周期：invited → active ←→ suspended
```

#### Role（角色定义）
```go
type Role struct {
    ID          string   // 角色 ID
    OrgName     string   // 所属组织
    Name        string   // 角色标识（如 "owner", "editor", "viewer"）
    DisplayName string   // 显示名称
    IsSystem    bool     // 系统角色不可删除
    Permissions []string // 权限列表
}
```

#### Permission（权限值对象）
```
格式："{resource}:{action}"
示例：
  - model:read      # 读取模型
  - model:write     # 编写模型
  - model:create    # 创建模型
  - model:delete    # 删除模型
  - project:*       # 项目下所有操作（通配符）
  - *               # 超级权限
```

### 3. 权限检查流程

```
GraphQL 请求
    │
    ▼
@hasPermission(action: "model:read") 指令触发
    │
    ├─ validateContext()
    │   ├─ 从 context 获取 userID
    │   └─ 从 context 获取 orgName
    │
    ├─ 尝试从 JWT context 读取权限（快速路径）
    │   └─ checkContextPermission()
    │       ├─ middleware.CheckPermission(permissions, action)
    │       └─ 返回访问结果
    │
    └─ 若 JWT context 无权限，查询数据库（回退路径）
        └─ checkDatabasePermission()
            ├─ userRoleService.CheckPermission(ctx, userID, orgName, action)
            └─ 返回访问结果
```

### 4. JWT Claims 结构

ModelCraft 使用两种 JWT 类型：

#### ModelCraftClaims（包含认证和授权信息）
```go
type ModelCraftClaims struct {
    jwt.RegisteredClaims

    // 用户身份
    UserID     string `json:"user_id"`         // ModelCraft 内部 UUID
    ExternalID string `json:"external_id"`     // 外部 IdP 用户 ID（AuthProvider）
    Name       string `json:"name"`
    Email      string `json:"email"`

    // 组织
    Organization string `json:"organization"`  // 当前组织名

    // 权限
    Roles       []string `json:"roles"`        // 角色名列表（如 ["owner"]）
    Permissions []string `json:"permissions"` // 权限列表（如 ["model:read", "model:write"]）

    // 多组织成员关系
    Memberships        []MembershipClaimInfo `json:"memberships,omitempty"`  // 最多 10 个
    HasMoreMemberships bool                  `json:"hasMoreMemberships,omitempty"`

    Issuer string `json:"iss"`  // 总是 "modelcraft"
}

type MembershipClaimInfo struct {
    OrgName     string `json:"orgName"`
    DisplayName string `json:"displayName"`
    Role        string `json:"role"`
    JoinedAt    int64  `json:"joinedAt"`  // Unix 时间戳（毫秒）
}
```

#### UserClaims（简化版，仅用于简单认证）
```go
type UserClaims struct {
    jwt.RegisteredClaims
    UserID string `json:"user_id"`
}
```

---

## 认证系统

### 1. 认证流程

```
客户端 (AuthProvider)
    │
    ├─ 用户登录 AuthProvider
    ├─ AuthProvider 签发 JWT Token（RS256 公钥签名）
    │
    ▼
ModelCraft 后端
    │
    ├─ 接收 JWT Token（Bearer Authorization 头）
    │
    ├─ validateModelCraftJWT()
    │   ├─ 使用 ModelCraft 密钥验证签名（HMAC-SHA256）
    │   └─ 提取 UserClaims（userID）
    │
    ├─ 从 JWT 提取 ExternalID（AuthProvider 用户 ID）
    │
    ├─ 查找或创建本地 User 记录（通过 ExternalID 关联）
    │
    └─ 注入 context，传递给下游
```

### 2. AuthProvider 集成

**核心配置**：
```go
type ProjectAuthConfig struct {
    OrgName      string                    // 组织名
    ProjectSlug  string                    // 项目标识
    Provider     ProviderType              // 认证提供者：auth_provider, keycloak, oidc
    Enabled      bool                      // 是否启用
    Config       map[string]interface{}    // Provider 专属配置
}
```

**AuthProvider 提供者支持**：
- 完整实现：✅ AuthProvider
- 预留结构：❌ Keycloak（未实现）
- 预留结构：❌ OIDC（未实现）

### 3. 中间件认证

```go
// ChiJWTAuthMiddleware 验证 JWT 或 API Key
func ChiJWTAuthMiddleware(config *JWTAuthConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearerToken(r)  // 从 Authorization 头提取

            // API Key 路径
            if strings.HasPrefix(token, "mc_") {
                userID, err := config.APIKeyVerifier.VerifyAPIKey(r.Context(), token)
                if err != nil || userID == "" {
                    http.Error(w, "Unauthorized", http.StatusUnauthorized)
                    return
                }
                ctx := ctxutils.SetUserID(r.Context(), userID)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }

            // JWT 路径
            claims, err := validateModelCraftJWT(config.ModelCraftSecret, token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := ctxutils.SetUserID(r.Context(), claims.UserID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

## 数据库层结构

### 1. 数据库表设计

```
modelcraft-backend/db/schema/mysql/
├── 01_project.sql              # 项目表
├── 02_database_cluster.sql     # 数据库集群配置表
├── 03_model_domain.sql         # 模型定义表
├── 04_auth.sql                 # 认证配置表
├── 05_organizations.sql        # 组织表
├── 06_users.sql                # 用户表
├── 07_roles_permissions.sql    # 角色和权限表
├── 08_refresh_tokens.sql       # 刷新 Token 表
├── 09_api_keys.sql             # API Key 表
└── 10_security_audit_logs.sql  # 审计日志表
```

### 2. 关键表结构

#### organizations（组织表）
```sql
CREATE TABLE `organizations` (
  `name` VARCHAR(36) PRIMARY KEY,              -- 唯一标识符
  `display_name` VARCHAR(255),                 -- UI 显示名称
  `owner_id` VARCHAR(36),                      -- 创建者用户 ID
  `status` VARCHAR(20) DEFAULT 'active',       -- active | suspended | deleted
  `created_at` DATETIME(3),
  `updated_at` DATETIME(3)
);
```

#### projects（项目表 - 复合主键）
```sql
CREATE TABLE `projects` (
  `org_name` VARCHAR(36) NOT NULL,              -- 复合主键之一
  `slug` VARCHAR(64) NOT NULL,                  -- 复合主键之二
  `title` VARCHAR(255) NOT NULL,
  `description` TEXT,
  `status` VARCHAR(20) DEFAULT 'active',
  `cluster_id` VARCHAR(36),                     -- 关联的集群 ID（1:1）
  `created_at` DATETIME(3),
  `updated_at` DATETIME(3),
  
  PRIMARY KEY (`org_name`, `slug`)              -- 复合主键
);
```

#### database_clusters（数据库集群表）
```sql
CREATE TABLE `database_clusters` (
  `id` VARCHAR(36) PRIMARY KEY,
  `org_name` VARCHAR(36) NOT NULL,              -- 所属组织
  `project_slug` VARCHAR(64) NOT NULL,          -- 所属项目
  `title` VARCHAR(255) NOT NULL,
  `host` VARCHAR(255) NOT NULL,                 -- 数据库主机
  `port` BIGINT DEFAULT 3306,                   -- 数据库端口
  `username` VARCHAR(255) NOT NULL,             -- 数据库用户名
  `password` TEXT NOT NULL,                     -- 加密的数据库密码
  `connection_timeout` INT DEFAULT 5,           -- 连接超时（秒）
  `max_open_conns` BIGINT DEFAULT 100,          -- 连接池：最大打开连接数
  `max_idle_conns` BIGINT DEFAULT 10,           -- 连接池：最大空闲连接数
  `conn_max_lifetime` BIGINT DEFAULT 3600,      -- 连接池：连接最大生命周期
  `status` VARCHAR(20) DEFAULT 'active',
  `version` BIGINT DEFAULT 1,
  `created_at` DATETIME(3),
  `updated_at` DATETIME(3),
  
  UNIQUE KEY `idx_cluster_project_unique` (`org_name`, `project_slug`)  -- 1:1 约束
);
```

#### models（模型定义表）
```sql
CREATE TABLE `models` (
  `id` VARCHAR(36) PRIMARY KEY,
  `org_name` VARCHAR(36) NOT NULL,
  `project_slug` VARCHAR(64) NOT NULL,
  `name` VARCHAR(64) NOT NULL,                  -- 模型名称
  `title` VARCHAR(255) NOT NULL,
  `description` TEXT,
  `database_name` VARCHAR(64) NOT NULL,         -- 目标数据库名称
  `storage_type` VARCHAR(100) NOT NULL,
  `display_field` VARCHAR(64),                  -- 用于 _displayName 的字段
  `version` BIGINT DEFAULT 1,
  `status` VARCHAR(50) DEFAULT 'draft',         -- draft | published | archived
  `deployment_status` VARCHAR(50) DEFAULT 'pending',
  `group_id` VARCHAR(36),
  `last_sync_at` DATETIME(3),
  `sync_error` TEXT,
  `created_at` DATETIME(3),
  `updated_at` DATETIME(3),
  
  UNIQUE KEY `idx_models_name` (`org_name`, `project_slug`, `database_name`, `name`)
);
```

#### users（用户表 - 混合认证）
```sql
CREATE TABLE `users` (
  `id` VARCHAR(36) PRIMARY KEY,                  -- 内部 UUID
  `external_id` VARCHAR(255),                    -- 外部 IdP 用户 ID（AuthProvider）
  `name` VARCHAR(255) NOT NULL,                  -- 用户名
  `phone` VARCHAR(32),                           -- 手机号（本地注册用）
  `password_hash` VARCHAR(255),                  -- bcrypt 哈希（本地注册用）
  `display_name` VARCHAR(255),
  `created_at` DATETIME(3),
  `updated_at` DATETIME(3),
  
  UNIQUE INDEX `uk_phone` (`phone`),
  UNIQUE INDEX `uk_user_name` (`name`),
  INDEX `idx_external_id` (`external_id`)
);
```

#### user_organizations（用户-组织关联）
```sql
CREATE TABLE `user_organizations` (
  `id` VARCHAR(36) PRIMARY KEY,
  `user_id` VARCHAR(36) NOT NULL,
  `org_name` VARCHAR(36) NOT NULL,
  `status` VARCHAR(20) DEFAULT 'active',        -- active | suspended | invited
  `invited_by` VARCHAR(36),
  `invited_at` DATETIME(3),
  `joined_at` DATETIME(3),
  `created_at` DATETIME(3),
  `updated_at` DATETIME(3),
  
  UNIQUE KEY `uk_user_org` (`user_id`, `org_name`)
);
```

---

## 数据隔离机制

### 1. 多租户隔离架构

```
ModelCraft 元数据数据库（共享）
    │
    ├─ organizations（组织隔离键）
    ├─ users / user_organizations
    ├─ roles / permissions
    ├─ projects（依赖 org_name）
    ├─ database_clusters（依赖 org_name, project_slug）
    └─ models（依赖 org_name, project_slug）
    
           ▼

客户端数据库实例（每个 Project 一个）
    │
    ├─ 客户自行管理的数据库
    ├─ ModelCraft 自动创建表
    ├─ Runtime GraphQL 查询在此执行
    └─ 数据完全隔离（不同 Project 不同数据库）
```

### 2. 数据隔离实现

#### 应用层隔离（设计态）
```go
// 模型定义等元数据存储在 ModelCraft 共享数据库，通过字段隔离
// 必需字段：org_name, project_slug, database_name

// 查询示例
models := modelRepo.GetByProjectKey(ctx, orgName, projectSlug, databaseName)
    // WHERE org_name = ? AND project_slug = ? AND database_name = ?
```

#### 数据库层隔离（运行态）
```go
// 客户数据存储在各自的数据库实例中，物理隔离

// 获取连接过程
conn, err := ClusterConnectionManager.GetConnectionWithDatabase(
    ctx, orgName, projectSlug, databaseName,
)
    │
    ├─ 从 database_clusters 表查询 Cluster 配置（按 orgName, projectSlug）
    ├─ 建立 TCP 连接到客户数据库（host, port, username, password）
    ├─ 执行 `USE `databaseName`` 切换到指定数据库
    └─ 返回 *sql.DB
```

### 3. 数据隔离的两个维度

| 维度 | 元数据（设计态） | 客户数据（运行态） |
|------|-----------------|------------------|
| 存储位置 | ModelCraft 共享 MySQL 数据库 | 客户自管理的 MySQL 数据库实例 |
| 隔离级别 | 应用层隔离（通过 WHERE 条件） | 物理隔离（不同 host/port/db） |
| 隔离键 | org_name, project_slug, database_name | 数据库实例（host:port） + database_name |
| 访问时机 | 设计态 GraphQL 读取元数据 | 运行态 GraphQL 执行 SQL 查询 |

---

## 多租户设计

### 1. 多租户层级

```
Organization (顶层租户容器)
    │
    ├─ Project 1（资源隔离单元）
    │   ├─ Cluster（数据库连接配置）
    │   └─ Models
    │       ├─ Model A（对应数据库 db1 中的表 tableA）
    │       └─ Model B（对应数据库 db2 中的表 tableB）
    │
    ├─ Project 2
    │   └─ ...
    │
    └─ User 1 → Role → Permissions
       User 2 → Role → Permissions
       ...
```

### 2. 租户路由

**设计态 URL 路由**：
```
GET /org/{orgName}/design/graphql
    │
    ├─ orgName 从 URL 提取
    └─ 后续所有查询在此 org 作用域内
```

**运行态 URL 路由**：
```
POST /{orgName}/{projectSlug}/{databaseName}/{modelName}
    │
    ├─ orgName 从 URL 提取
    ├─ projectSlug 从 URL 提取
    ├─ databaseName 从 URL 提取
    ├─ modelName 从 URL 提取
    │
    └─ 路由处理器
        ├─ 验证 orgName 与 JWT token 中的 organization 匹配
        ├─ 验证用户对此 project 有 model:read 权限
        ├─ 获取客户数据库连接（指定 databaseName）
        └─ 执行 GraphQL 查询
```

### 3. JWT Context 中的租户信息

```go
type ModelCraftClaims struct {
    UserID       string  // 用户 ID
    Organization string  // 当前组织（从 JWT 中提取）
    Permissions  []string
    Memberships  []MembershipClaimInfo  // 用户的多个组织成员身份
}

// 权限检查时使用
@hasPermission(action: "model:read")
    │
    ├─ 从 context 获取 orgName（从 JWT.Organization 或 URL 提取）
    ├─ 检查 permissions 中是否包含 "model:read"
    └─ 通过则执行，否则返回 PERMISSION_DENIED
```

---

## PostgreSQL 连接方式

### 当前状态

**Current**: ✅ 支持 MySQL
**PostgreSQL**: ❌ 未支持

### 架构分析

虽然当前只支持 MySQL，但架构预留了扩展性：

#### 连接工厂设计
```go
// connection_factory.go
type ConnectionFactory struct {
    SqlDB *sql.DB
}

// 未来可扩展为支持多种 DB 方言
```

#### SQL 生成使用了 Goqu 库

```go
import "github.com/doug-martin/goqu/v9"

// Goqu 支持多种 SQL 方言
dialectWrapper := goqu.Dialect("mysql")  // 当前使用
// 未来可改为：
// dialectWrapper := goqu.Dialect("postgres")
```

#### 数据库驱动

```go
// 当前驱动
import _ "github.com/go-sql-driver/mysql"
db, err := sql.Open("mysql", dsn)

// 未来 PostgreSQL 驱动（已有 Go 库支持）
import _ "github.com/lib/pq"
db, err := sql.Open("postgres", dsn)
```

### 迁移到 PostgreSQL 的关键点

1. **修改连接字符串生成**（`cluster_connection_manager.go`）
   - MySQL DSN: `user:password@tcp(host:port)/?charset=utf8mb4`
   - PostgreSQL DSN: `postgres://user:password@host:port/dbname`

2. **修改 SQL 方言**（`sql_mapper.go`）
   ```go
   // 改为 postgres
   dialectWrapper := goqu.Dialect("postgres")
   ```

3. **类型映射**（`runtimemodel.go`）
   - MySQL `VARCHAR` → PostgreSQL `VARCHAR` or `TEXT`
   - MySQL `DATETIME` → PostgreSQL `TIMESTAMP`
   - MySQL `BIGINT` → PostgreSQL `BIGINT`

4. **保存密码的字段类型**
   - MySQL `TEXT` → PostgreSQL `TEXT`（兼容）

---

## 关键文件路径汇总

### Runtime 模块

| 层级 | 文件路径 | 功能 |
|-----|--------|------|
| 领域 | `internal/domain/modelruntime/runtimemodel.go` | 运行时模型定义 |
| 领域 | `internal/domain/modelruntime/model_resolver.go` | GraphQL Schema 生成和查询执行核心逻辑 |
| 领域 | `internal/domain/modelruntime/model_repository.go` | 模型仓储接口 |
| 领域 | `internal/domain/modelruntime/graphql_request_context.go` | 请求级上下文管理（DB 连接、dataloader） |
| 领域 | `internal/domain/modelruntime/graphql_*.go` | GraphQL 工具类（输入类型、标量、字段条件等） |
| 领域 | `internal/domain/modelruntime/relation_loader.go` | 关系加载（N+1 问题处理） |
| 应用 | `internal/app/modelruntime/graphql_app.go` | GraphQL 应用服务（Execute、GetSchema） |
| 基础设施 | `internal/infrastructure/database/dml/client_db_repo_impl.go` | 客户端 DB 操作实现 |
| 基础设施 | `internal/infrastructure/database/dml/sql_mapper.go` | SQL 映射（WHERE → SQL） |
| 基础设施 | `internal/infrastructure/database/dml/query_parser.go` | 查询解析器（复杂条件支持） |

### RBAC 和权限

| 层级 | 文件路径 | 功能 |
|-----|--------|------|
| 领域 | `internal/domain/membership/membership.go` | 用户-组织关联实体 |
| 领域 | `internal/domain/role/role.go` | 角色实体 |
| 领域 | `internal/domain/permission/permission.go` | 权限值对象 |
| 认证 | `internal/domain/auth/modelcraft_claims.go` | ModelCraft JWT Claims |
| 认证 | `internal/domain/auth/user_claims.go` | 简化 JWT Claims |
| 认证 | `internal/domain/auth/project_auth_config.go` | 项目级认证配置 |
| 中间件 | `internal/middleware/chi_jwt_auth.go` | JWT 认证中间件 |
| GraphQL | `internal/interfaces/graphql/project/directives.go` | @hasPermission 指令实现 |

### 数据库层

| 文件路径 | 功能 |
|--------|------|
| `internal/infrastructure/repository/cluster_connection_manager.go` | 集群连接管理（连接池、缓存） |
| `internal/infrastructure/repository/connection_factory.go` | 连接工厂 |
| `db/schema/mysql/01_project.sql` | 项目表 |
| `db/schema/mysql/02_database_cluster.sql` | 数据库集群配置表 |
| `db/schema/mysql/03_model_domain.sql` | 模型定义表 |
| `db/schema/mysql/05_organizations.sql` | 组织表 |
| `db/schema/mysql/06_users.sql` | 用户表 |
| `db/schema/mysql/07_roles_permissions.sql` | 角色和权限表 |

### GraphQL Schema

| 文件路径 | 功能 |
|--------|------|
| `api/graph/project/schema/model.graphql` | 模型相关 Schema（设计态） |
| `api/graph/project/schema/base.graphql` | 基础类型定义 |
| `api/graph/project/schema/cluster.graphql` | 集群 Schema |
| `api/graph/project/schema/field.graphql` | 字段 Schema |
| `api/graph/project/schema/enum.graphql` | 枚举 Schema |

---

## 关键发现和设计要点

### 1. Runtime 执行的核心原理

✅ **Schema 和状态分离**
- GraphQL Schema 只是类型结构和 Resolve 函数，不持有状态
- 每次请求注入独立的 context，包含 DB 连接和 dataloader
- 这使得 Schema 可安全缓存和跨请求复用

✅ **动态 SQL 生成**
- 使用 Goqu 库生成 SQL（避免手写 SQL 字符串）
- 支持简单条件和复杂操作符（AND, OR, $in, $like 等）
- 全面的 CRUD 操作支持

✅ **N+1 问题处理**
- 使用 graphql-go/dataloader 聚合批量查询
- 同一请求内的关系字段加载被合并为单条 IN 查询

### 2. RBAC 权限模型

✅ **两层权限检查**
1. **快速路径**：从 JWT Claims 中读取权限（内存查询，无 DB I/O）
2. **回退路径**：若 JWT 中无权限，查询数据库

✅ **权限格式**
- `{resource}:{action}` 格式
- 支持通配符 `*`
- 支持资源级权限聚合

### 3. 数据隔离的双重机制

✅ **应用层隔离（元数据）**
- 所有表都有 `org_name, project_slug` 字段
- 查询时在 WHERE 条件中筛选

✅ **物理隔离（客户数据）**
- 每个 Project 对应一个外部数据库实例
- 完全独立的 host:port:database
- 即使是同一 MySQL server，不同 Project 用不同数据库名

### 4. 用户定义模型数据的存储位置

❓ **当前设计**：
- 元数据（模型定义、字段、关系）→ ModelCraft MySQL（共享）
- 用户数据（实际业务数据）→ 客户自管理的 MySQL 数据库实例（隔离）

✅ **隔离保证**：
- 不同 Project 的数据在物理上完全分离（不同数据库实例）
- 即使在同一 MySQL server，也通过 `USE database_name` 隔离

### 5. 当前认证系统的局限

✅ 已实现：
- AuthProvider 集成
- JWT 验证（HMAC-SHA256）
- 本地密码注册（手机号 + bcrypt）
- 混合认证模式

❌ 未实现：
- Keycloak（预留结构）
- 通用 OIDC（预留结构）
- PostgreSQL 支持（预留架构）

---

## 总结

ModelCraft 后端采用了**分离的两态架构**：

1. **设计态**：静态 GraphQL Schema，操作元数据
   - 在 ModelCraft 共享数据库中执行
   - 通过应用层筛选实现多租户隔离

2. **运行态**：动态 GraphQL Schema，操作客户数据
   - 在客户自管理的数据库实例中执行
   - 通过物理隔离实现完全的数据独立

**关键设计优势**：
- ✅ GraphQL Schema 可缓存，支持高并发
- ✅ 请求级状态完全隔离，无竞态条件
- ✅ 数据隔离多维度保证（应用层 + 数据库层）
- ✅ 权限检查快速路径（JWT）+ 回退路径（数据库）
- ✅ 灵活的查询语言（支持简单条件和复杂操作符）

