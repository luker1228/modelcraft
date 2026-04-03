# Design: Cluster as Project Sub-Resource

## 背景

### 现状分析

当前 Project-Cluster 关系已在数据库层和领域层正确建模为一对一关系：

- `database_clusters` 表有 `UNIQUE KEY idx_cluster_project_unique (org_name, project_name)`
- `DatabaseCluster` 域模型包含 `ProjectName` 字段
- `Project` 域模型有 `ClusterID *string` 字段

**问题在于 GraphQL API 层将 Cluster 暴露为独立资源，与实际领域建模矛盾。**

### 领域建模视角

从 DDD 角度，`DatabaseCluster` 是 `Project` 的**值对象/子实体**，而非独立聚合根：

```
Project (Aggregate Root)
└── DatabaseCluster (sub-entity, 连接配置)
    ├── name
    ├── title
    ├── description
    └── connectionInfo (host, port, username, password)

Models / Enums (通过 project 隔离的下游资源)
```

Project 的生命周期包含 Cluster 的生命周期。没有 Project 的 Cluster 没有业务意义。

---

## API 设计

### 新 GraphQL Schema

#### Project 类型

```graphql
type Project implements Node {
  id: ID!
  name: String!
  title: String!
  description: String!
  loginUrl: String
  status: ProjectStatus!
  orgName: String!
  cluster: DatabaseCluster!   # 必填，project 必有 cluster
  createdAt: String!
  updatedAt: String!
}
```

移除：`clusterId`、`clusterInfo`

#### CreateProject

```graphql
input CreateProjectInput {
  name: String!
  title: String!
  description: String
  loginUrl: String
  clusterInput: ClusterConnectionInput!   # 必填
  skipConnectionTest: Boolean             # 默认 false
}

input ClusterConnectionInput {
  name: String!
  title: String!
  description: String
  connectionInfo: DatabaseConnectionInput!
}
```

**原子操作**：在同一个事务中创建 project 和 cluster，任一失败则全部回滚。

#### UpdateProjectCluster（新增）

```graphql
input UpdateClusterConnectionInput {
  title: String
  description: String
  connectionInfo: DatabaseConnectionInput
  skipConnectionTest: Boolean   # 默认 false
}

extend type Mutation {
  updateProjectCluster(
    projectName: String!
    input: UpdateClusterConnectionInput!
  ): UpdateClusterPayload! @hasPermission(action: "cluster:update")
}
```

移除：`updateDatabaseCluster(projectName, name, input)`

#### DeleteProject（行为变化）

签名不变，但级联删除 cluster：

```graphql
mutation DeleteProject($name: String!) {
  deleteProject(name: $name) {
    success
    error: DeleteProjectError
  }
}
```

实现：在事务中先删 cluster，再删 project（或通过外键级联）。

#### 移除的 API

```graphql
# 移除
createDatabaseCluster(input: CreateDatabaseClusterInput!)
deleteDatabaseCluster(projectName: String!, name: String!)
databaseClusters(projectName: String!, input: DatabaseClusterQueryInput)
updateDatabaseCluster(projectName: String!, name: String!, input: UpdateDatabaseClusterInput!)
```

#### 保留不变

```graphql
databaseCluster(projectName: String!): GetClusterPayload!
listDatabases(input: ListDatabasesInput!): DatabaseConnection!
testDatabaseConnection(input: TestDatabaseConnectionInput!): TestConnectionPayload!
```

---

## 实现路径

### 层次影响分析

| 层次 | 变化程度 | 说明 |
|------|----------|------|
| 数据库 Schema | **无变化** | 已有结构满足需求 |
| Domain 层 | **小改动** | `Project.ClusterID` 改为必填语义；移除独立 cluster 创建入口 |
| Application 层 | **中改动** | `CreateProjectUseCase` 原子创建；新增 `UpdateProjectClusterUseCase`；`DeleteProjectUseCase` 级联删除 |
| GraphQL Schema | **大改动** | 字段增删、mutation 重构 |
| GraphQL Resolver | **大改动** | 对应 schema 变化 |

### Application 层变更

#### CreateProjectUseCase（修改）

```go
type CreateProjectCommand struct {
    OrgName            string
    Name               string
    Title              string
    Description        string
    LoginURL           string
    ClusterInput       CreateClusterInput  // 必填
    SkipConnectionTest bool
}

type CreateClusterInput struct {
    Name              string
    Title             string
    Description       string
    Host              string
    Port              int
    Username          string
    Password          string
    ConnectionTimeout int
}
```

实现逻辑（在事务中）：
1. 验证 project 名称唯一性
2. 若 `!SkipConnectionTest`，测试数据库连接
3. 创建 project 记录
4. 创建 cluster 记录（关联 project）
5. 更新 project.cluster_id

#### UpdateProjectClusterUseCase（新增）

```go
type UpdateProjectClusterCommand struct {
    OrgName            string
    ProjectName        string
    Title              string
    Description        string
    // connection info fields...
    SkipConnectionTest bool
}
```

#### DeleteProjectUseCase（修改）

在事务中：
1. 查找并删除关联的 cluster（若存在）
2. 删除 project

### 错误处理

新增错误场景：

| 场景 | 错误类型 |
|------|----------|
| `createProject` 时连接测试失败 | `DatabaseConnectionFailed`（已有） |
| `createProject` 时 cluster 名称冲突（理论上不会，因为 project 是新建的） | 不需要 |
| `updateProjectCluster` 时 project 不存在 | `ProjectNotFound`（已有） |
| `updateProjectCluster` 时连接测试失败 | `DatabaseConnectionFailed`（已有） |

`CreateProjectError` union 新增 `DatabaseConnectionFailed`：
```graphql
union CreateProjectError = ProjectAlreadyExists | InvalidProjectInput | DatabaseConnectionFailed
```

---

## 迁移策略

### 历史数据

已有的无 cluster 的 project 记录需手动清理（业务方确认历史数据不重要）。

### 客户端迁移

由于是破坏性变更，客户端需要同步迁移：

1. `createProject` 调用需新增 `clusterInput` 字段
2. `updateDatabaseCluster` 调用改为 `updateProjectCluster`（去掉 `name` 参数）
3. `deleteDatabaseCluster` 调用改为通过 `deleteProject` 完成
4. 不再使用 `databaseClusters` 查询，改用 `databaseCluster`
5. `Project.clusterId` / `Project.clusterInfo` 改为 `Project.cluster`

---

## 风险评估

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| `createProject` 事务部分失败（project 创建成功但 cluster 失败） | 低 | 高 | 严格使用数据库事务，任一失败全部回滚 |
| 连接测试超时导致创建体验差 | 中 | 中 | `skipConnectionTest` 选项；合理的超时设置（5s） |
| 客户端未同步迁移 | 中 | 高 | 破坏性变更，需协调前后端同步发布 |
| `task generate-gql` 覆盖手写代码 | 低 | 高 | 严格遵循生成流程，不手写 generated 目录 |
