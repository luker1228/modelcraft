# Change: Refactor Cluster as Project Sub-Resource

## Change ID

`refactor-cluster-as-project-subresource`

## Why

当前 Project 和 Cluster 的 API 设计存在领域建模错误：

### 问题 1：Cluster 被错误地作为独立资源暴露

数据库层已通过 `UNIQUE KEY (org_name, project_name)` 强制 Cluster 归属于 Project（一对一），
领域模型也包含 `ProjectName` 字段，但 GraphQL API 仍将 Cluster 作为顶层独立资源处理：

- `createDatabaseCluster(input: { projectName: ... })` —— projectName 只是普通字段，没有体现从属
- `databaseClusters(projectName: String!)` —— 复数查询暗示可有多个 cluster，与一对一约束矛盾
- `deleteDatabaseCluster(...)` —— 独立删除 cluster，割裂了 project-cluster 的生命周期

Cluster 不是独立聚合根，它是 Project 的连接配置组成部分，没有 Project 就没有存在意义。

### 问题 2：API 与前端路由语义不一致

前端路由为 `project/{project}/cluster`，明确表示 cluster 是 project 的子资源配置页。
但当前 API 没有对应的"通过 project 管理其 cluster"的操作路径。

### 问题 3：Project 创建后可以处于无 Cluster 的半完成状态

当前允许先创建 project，再单独创建 cluster，导致 project 可以在没有数据库连接的状态下存在，
无法存储任何 model 数据，是一个无意义的半完成状态。

## What Changes

### Mutations

| 旧 API | 新 API | 说明 |
|--------|--------|------|
| `createProject(input: {...})` | `createProject(input: { ...projectFields, clusterInput: ClusterConnectionInput!, skipConnectionTest: Boolean })` | 内嵌 cluster 创建，原子操作 |
| `updateProject(input: {...})` | 不变 | 只更新 project 元数据 |
| `deleteProject(name: String!)` | 不变签名，行为变化 | 级联删除 cluster |
| `createDatabaseCluster(...)` | **移除** | cluster 通过 createProject 创建 |
| `updateDatabaseCluster(projectName, name, input)` | `updateProjectCluster(projectName: String!, input: UpdateClusterConnectionInput!)` | 重命名，去掉冗余的 `name` 参数 |
| `deleteDatabaseCluster(...)` | **移除** | cluster 随 project 级联删除 |
| `testDatabaseConnection(...)` | 保留不变 | 独立的连接测试工具 |

### Queries

| 旧 API | 新 API | 说明 |
|--------|--------|------|
| `databaseCluster(projectName: String!)` | 保留不变 | |
| `databaseClusters(projectName: String!)` | **移除**（破坏性删除） | 一对一关系下无意义 |
| `listDatabases(...)` | 保留不变 | |

### Project 类型字段

| 旧字段 | 新字段 | 说明 |
|--------|--------|------|
| `clusterId: String` | **移除** | cluster 通过 `cluster` 字段访问 |
| `clusterInfo: DatabaseCluster` | **移除** | 语义不清晰 |
| *(不存在)* | `cluster: DatabaseCluster!` | 新增，必填（project 必有 cluster） |

### 生命周期规则

- **创建**：`createProject` 原子创建 project + cluster，cluster 连接信息必填，默认验证连接，可通过 `skipConnectionTest: true` 跳过
- **更新 cluster**：通过 `updateProjectCluster(projectName, input)` 单独更新连接配置
- **删除**：`deleteProject` 级联删除 cluster，不再有独立的 `deleteDatabaseCluster`

## Breaking Changes

- 移除 `createDatabaseCluster` mutation
- 移除 `deleteDatabaseCluster` mutation
- 移除 `databaseClusters` query
- 移除 `Project.clusterId` 字段
- 移除 `Project.clusterInfo` 字段
- `updateDatabaseCluster` 替换为 `updateProjectCluster`（签名变化）
- `createProject` input 新增必填的 `clusterInput` 字段
- `deleteProject` 新增级联删除 cluster 的行为

## Scope

- **GraphQL schema 变更**：`api/graph/schema/project.graphql` 大幅更新；`api/graph/schema/cluster.graphql` 内容合并入 `project.graphql` 后**删除**（cluster 重构后不再是独立资源，两者合并至约 190 行，与其他 schema 文件体量相当）
- **GraphQL resolver 变更**：`internal/interfaces/graphql/project.resolvers.go` 承接全部 project + cluster resolver；`internal/interfaces/graphql/cluster.resolvers.go` 删除
- **Application 层变更**：`createProject` 原子创建 project + cluster；新增 `UpdateProjectCluster` use case；`deleteProject` 级联删除 cluster
- **Domain 层变更**：`Project.ClusterID` 改为必填语义
- **无数据库 schema 变更**（已有一对一约束结构满足需求）

## Dependencies

- 依赖现有的 `refactor-project-cluster-to-one-to-one` 变更已完成（数据库一对一约束已存在）
