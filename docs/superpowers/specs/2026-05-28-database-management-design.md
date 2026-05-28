# Database Management Design

**Date:** 2026-05-28  
**Status:** Approved  
**Scope:** v1 — 基础接管：选择 + 设置模式

---

## 背景与目标

### 问题

当前模型编辑器（ModelSidebar）可以访问集群上的所有 MySQL database，不符合预期。应该只能访问被"接管"的 database。

### 目标

1. 提供独立的"数据库管理"页面，让开发者显式接管（注册）集群上的 database
2. 每个 database 可标记为**自建**（可读写）或**托管**（只读）
3. ModelSidebar 只展示已接管的 database，并根据模式限制操作项

### 核心概念

| 概念 | 说明 |
|------|------|
| **Cluster** | 一个 MySQL 连接（host/port/user/pass），1 Project 对应 1 Cluster |
| **Database** | Cluster 上的某个 MySQL database（`SHOW DATABASES` 的结果） |
| **自建（self_hosted）** | 可读写，支持新建/导入模型 |
| **托管（managed）** | 只读，仅支持同步模型（增量同步 + 结构跟随，v2 实现） |

---

## Section 1：数据模型

### 新增表：`model_database`

```sql
CREATE TABLE model_database (
  id            VARCHAR(36)   NOT NULL,
  org_name      VARCHAR(64)   NOT NULL,
  project_slug  VARCHAR(64)   NOT NULL,
  cluster_id    VARCHAR(36)   NOT NULL,
  name          VARCHAR(64)   NOT NULL,       -- MySQL database 原始名，只读
  title         VARCHAR(128)  NOT NULL,       -- 用户设置的友好名称
  description   TEXT          NOT NULL DEFAULT '',
  mode          ENUM('self_hosted','managed') NOT NULL,
  delete_token  VARCHAR(36)   NOT NULL DEFAULT '',
  deleted_at    DATETIME,
  created_at    DATETIME      NOT NULL,
  updated_at    DATETIME      NOT NULL,

  PRIMARY KEY (id),
  UNIQUE KEY uq_project_name (project_slug, name, delete_token),
  INDEX idx_cluster (cluster_id)
);
```

**字段说明：**

| 字段 | 说明 |
|------|------|
| `name` | 来自 `SHOW DATABASES` 的原始库名，注册后不可修改 |
| `title` | 友好显示名，默认等于 `name`，用户可编辑 |
| `description` | 可选描述 |
| `mode` | `self_hosted` / `managed` |
| `delete_token` | 软删除 token，配合联合唯一索引允许重复接管 |

### 模型关联

- 现有 `Model.databaseName` 字段（裸字符串）通过 `(project_slug, database_name)` 与 `model_database.name` 做逻辑关联
- v1 不加外键约束，避免历史数据迁移风险

---

## Section 2：后端 API

### 新文件：`api/graph/project/schema/database.graphql`

```graphql
enum DatabaseMode {
  SELF_HOSTED
  MANAGED
}

type ModelDatabase {
  id: ID!
  name: String!          # MySQL 原始库名，只读
  title: String!         # 友好名称
  description: String!
  mode: DatabaseMode!
  createdAt: Time!
  updatedAt: Time!
}

# 从 Cluster 实时拉取的原始 database
type RawDatabase {
  name: String!
  isRegistered: Boolean!  # 是否已在 model_database 表中
}

extend type Query {
  modelDatabases: [ModelDatabase!]!
  clusterRawDatabases: [RawDatabase!]!
}

extend type Mutation {
  registerModelDatabase(input: RegisterModelDatabaseInput!): RegisterModelDatabaseResult!
  updateModelDatabase(id: ID!, input: UpdateModelDatabaseInput!): ModelDatabase!
  unregisterModelDatabase(id: ID!): Boolean!
}

input RegisterModelDatabaseInput {
  name: String!
  title: String!
  description: String
  mode: DatabaseMode!
}

input UpdateModelDatabaseInput {
  title: String
  description: String
  mode: DatabaseMode
}

union RegisterModelDatabaseResult = ModelDatabase | InvalidInput | ResourceNotFound
```

### Domain 层

- 新目录：`internal/domain/database/`
- 实体：`ModelDatabase`
- Repository 接口：`List`, `GetByName`, `Create`, `Update`, `Delete`
- `ClusterService.ListDatabases()` 执行 `SHOW DATABASES` 并过滤系统库（`information_schema`, `mysql`, `performance_schema`, `sys`）

### Adapter 层

- 新增 `internal/adapter/mysql/database_adapter.go`（sqlc 生成）
- 新增 `internal/interfaces/graphql/resolver/database_resolver.go`

---

## Section 3：前端页面

### 路由

```
/org/[orgName]/project/[projectSlug]/databases   ← 新增
```

### 导航栏变更（AppLayout.tsx）

```
数据建模
├── 数据模型   (/model-editor)
├── 数据库     (/databases)   ← 新增，紧跟数据模型之后
└── 枚举管理   (/enums)
```

### 页面结构

**顶部 Header**
- 标题"数据库管理"
- 右上角"接管数据库"按钮 → 打开注册 Dialog

**主体：Table 列表**

| 列 | 内容 |
|----|------|
| 名称 | `title`（正文），下方小字展示原始 `name` |
| 描述 | `description`，截断显示 |
| 模式 | Badge：自建（绿色）/ 托管（蓝色）|
| 操作 | 编辑图标 / 更多菜单（取消接管）|

### 注册 Dialog

1. 调用 `clusterRawDatabases`，Select 下拉只展示 `isRegistered=false` 的 database
2. 友好名称默认填充原始库名，可修改
3. 描述（可选）
4. 模式单选：自建（默认）/ 托管，含简短说明文字
5. 操作按钮：取消 / 确认接管

### 编辑 Sheet

可编辑字段：`title`、`description`、`mode`  
只读展示：`name`

### ModelSidebar 改动

| 改动 | 说明 |
|------|------|
| 数据库列表来源 | 从"所有 database"改为查询 `modelDatabases`（已接管） |
| 下拉展示 | 在 database 名称旁显示模式 Badge |
| `mode=managed` 时 | 隐藏"新建模型"和"导入模型"，仅保留"同步模型"（v1 占位，v2 实现） |

---

## 范围边界

### v1 包含

- `model_database` 表 + Atlas 迁移
- GraphQL API：`modelDatabases`、`clusterRawDatabases`、`registerModelDatabase`、`updateModelDatabase`、`unregisterModelDatabase`
- 前端：数据库管理独立页面 + 注册/编辑 Dialog
- ModelSidebar 过滤逻辑 + managed 模式操作限制

### v2 延后

- 托管数据库"同步模型"后端实现（增量同步 + 结构跟随）
- 同步日志、同步进度、同步状态展示
- 每个 database 的已同步模型覆盖详情

---

## 文件变更清单

### 后端

| 文件 | 操作 |
|------|------|
| `api/graph/project/schema/database.graphql` | 新建 |
| `internal/domain/database/model_database.go` | 新建 |
| `internal/domain/database/repository.go` | 新建 |
| `internal/adapter/mysql/database_adapter.go` | 新建（sqlc） |
| `internal/interfaces/graphql/resolver/database_resolver.go` | 新建 |
| `internal/domain/cluster/cluster_service.go` | 新增 `ListDatabases()` |
| Atlas 迁移文件 | 新建 |

### 前端

| 文件 | 操作 |
|------|------|
| `src/app/org/[orgName]/project/[projectSlug]/databases/page.tsx` | 新建 |
| `src/app/org/[orgName]/project/[projectSlug]/databases/_components/DatabaseListTable.tsx` | 新建 |
| `src/app/org/[orgName]/project/[projectSlug]/databases/_components/RegisterDatabaseDialog.tsx` | 新建 |
| `src/app/org/[orgName]/project/[projectSlug]/databases/_components/EditDatabaseSheet.tsx` | 新建 |
| `src/app/org/[orgName]/project/[projectSlug]/databases/_hooks/useModelDatabases.ts` | 新建 |
| `src/web/components/features/layout/AppLayout.tsx` | 修改（加导航项） |
| `src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx` | 修改（过滤 + 操作限制） |
| `src/api-client/project/graphql-docs.ts` | 修改（加新 GraphQL 文档） |
