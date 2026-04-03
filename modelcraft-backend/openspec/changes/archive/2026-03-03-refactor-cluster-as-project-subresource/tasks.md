# Tasks: Refactor Cluster as Project Sub-Resource

## Overview

Refactor the Project-Cluster relationship so Cluster is managed as a sub-resource of Project.
This is a breaking change. Coordinate frontend and backend deployment.

## Prerequisites

- [ ] `refactor-project-cluster-to-one-to-one` change is fully deployed (DB one-to-one constraint exists)
- [ ] Historical data cleaned up manually (projects without clusters removed)

---

## Phase 1: GraphQL Schema

### Task 1.1 — Update `project.graphql` ✅

- Remove `clusterId: String` field from `Project` type
- Remove `clusterInfo: DatabaseCluster` field from `Project` type
- Add `cluster: DatabaseCluster!` field to `Project` type
- Add `clusterInput: ClusterConnectionInput!` to `CreateProjectInput`
- Add `skipConnectionTest: Boolean` to `CreateProjectInput`
- Add `ClusterConnectionInput` input type (name, title, description, connectionInfo)
- Add `DatabaseConnectionFailed` to `CreateProjectError` union

**Validation**: Run `task generate-gql` — should succeed without errors.

### Task 1.2 — Merge `cluster.graphql` into `project.graphql` and delete `cluster.graphql` ✅

重构后 cluster 不再是独立资源，相关定义移入 `project.graphql` 以体现从属关系。

从 `cluster.graphql` 保留并迁移到 `project.graphql` 的内容：
- `DatabaseCluster` type
- `ClusterStatus` enum
- `DatabaseConnectionInfo` type
- `Database` / `DatabaseConnection` / `DatabaseEdge` types
- `ListDatabasesInput` input type
- `DatabaseConnectionInput` input type
- `TestDatabaseConnectionInput` input type
- Error types: `ClusterAlreadyExists`, `ClusterNotFound`, `InvalidClusterInput`, `DatabaseConnectionFailed`, `ClusterAlreadyExistsForProject`
- Error unions: `GetClusterError`, `UpdateClusterError`, `TestConnectionError`
- Payload types: `GetClusterPayload`, `UpdateClusterPayload`, `DeleteClusterPayload`, `TestConnectionPayload`
- `databaseCluster` query
- `listDatabases` query
- `testDatabaseConnection` mutation
- 新增 `updateProjectCluster` mutation
- 新增 `UpdateClusterConnectionInput` input type (title, description, connectionInfo, skipConnectionTest)

从 `cluster.graphql` **不迁移**（直接删除）的内容：
- `createDatabaseCluster` mutation 及 `CreateDatabaseClusterInput`、`CreateClusterError`、`CreateClusterPayload`
- `deleteDatabaseCluster` mutation
- `updateDatabaseCluster` mutation 及 `UpdateDatabaseClusterInput`
- `databaseClusters` query 及 `DatabaseClusterConnection`、`DatabaseClusterEdge`、`DatabaseClusterQueryInput`

删除 `api/graph/schema/cluster.graphql`。

**Validation**: Run `task generate-gql` — should succeed without errors.

### Task 1.3 — Regenerate GraphQL code ✅

```bash
task generate-gql
```

Verify the generated resolver interfaces reflect the new schema.

---

## Phase 2: Application Layer

### Task 2.1 — Update `CreateProjectCommand` ✅

File: `internal/app/project/project_service.go` (or equivalent use case file)

- Add `ClusterInput CreateClusterInput` field to `CreateProjectCommand`
- Add `SkipConnectionTest bool` field to `CreateProjectCommand`
- Define `CreateClusterInput` struct with: Name, Title, Description, Host, Port, Username, Password, ConnectionTimeout

Write unit tests first:
- Test: create project + cluster atomically succeeds
- Test: connection test failure rolls back both project and cluster
- Test: `skipConnectionTest: true` skips connection validation

### Task 2.2 — Implement atomic create in `CreateProject` use case ✅

In the same DB transaction:
1. Check project name uniqueness
2. If `!SkipConnectionTest`: call connection test; return `DatabaseConnectionFailed` on failure
3. Create project record
4. Create cluster record (linked to project)
5. Update `project.cluster_id`

**Validation**: Unit tests from 2.1 pass.

### Task 2.3 — Add `UpdateProjectClusterUseCase` ✅

File: `internal/app/cluster/cluster_app.go` (or new file)

- Define `UpdateProjectClusterCommand` struct
- Implement: lookup project → lookup cluster by project → optionally test connection → update cluster
- Return `DatabaseConnectionFailed` if connection test fails and `skipConnectionTest` is false

Write unit tests first:
- Test: update succeeds with valid connection info
- Test: connection test failure returns error without updating
- Test: `skipConnectionTest: true` skips test and updates
- Test: project not found returns `ProjectNotFound`

**Validation**: Unit tests pass.

### Task 2.4 — Update `DeleteProject` use case to cascade delete cluster ✅

In the same DB transaction:
1. Look up cluster by project (if exists)
2. Delete cluster record (if exists)
3. Delete project record

Write unit tests first:
- Test: delete project with cluster removes both
- Test: delete project without cluster (legacy) removes only project

**Validation**: Unit tests pass.

---

## Phase 3: GraphQL Resolvers

### Task 3.1 — Implement `Project.Cluster` field resolver ✅

File: `internal/interfaces/graphql/project.resolvers.go`

- Implement `Cluster(ctx, obj *generated.Project) (*generated.DatabaseCluster, error)`
- Fetch cluster by project name using existing `GetClusterByProject` service method
- Remove `ClusterInfo` and `ClusterId` field resolvers (no longer in schema)

**Validation**: `task generate-gql` generates the correct resolver interface.

### Task 3.2 — Implement `updateProjectCluster` mutation resolver ✅

Schema 合并后 cluster resolver 与 project resolver 同在 `project.resolvers.go`（gqlgen 按 schema 文件生成 resolver 文件）。

- Implement `UpdateProjectCluster(ctx, projectName string, input generated.UpdateClusterConnectionInput) (*generated.UpdateClusterPayload, error)`
- Map input to `UpdateProjectClusterCommand`
- Handle errors: `ProjectNotFound`, `DatabaseConnectionFailed`

### Task 3.3 — Remove deleted resolver implementations and `cluster.resolvers.go` ✅

- Remove `CreateDatabaseCluster` resolver
- Remove `DeleteDatabaseCluster` resolver
- Remove `DatabaseClusters` resolver
- Remove `UpdateDatabaseCluster` resolver
- 确认 `cluster.graphql` 删除后，`cluster.resolvers.go` 中剩余 resolver 已迁移至 `project.resolvers.go`，然后删除 `cluster.resolvers.go`

### Task 3.4 — Update `CreateProject` mutation resolver ✅

File: `internal/interfaces/graphql/project.resolvers.go`

- Extract `clusterInput` and `skipConnectionTest` from input
- Map to updated `CreateProjectCommand`
- Handle new `DatabaseConnectionFailed` error in response

### Task 3.5 — Update `DeleteProject` mutation resolver ✅

- No resolver changes needed (signature unchanged)
- Verify cascade delete is triggered from the use case layer

---

## Phase 4: Update Mappers and Adapters

### Task 4.1 — Update `project_mapper.go` ✅

- Remove mapping for `clusterId` / `clusterInfo` fields
- No new mapping needed for `cluster` field (resolved via field resolver, not mapper)

### Task 4.2 — Update `cluster_mapper.go` ✅

- Add `ToUpdateProjectClusterCommand(projectName string, input generated.UpdateClusterConnectionInput) UpdateProjectClusterCommand`
- Remove mappers for deleted operations: `ToCreateClusterCommand`

---

## Phase 5: Clean Up Tests

### Task 5.1 — Update integration tests ✅

- Update any integration tests that call `createDatabaseCluster` → use `createProject` with `clusterInput`
- Update any integration tests that call `deleteDatabaseCluster` → use `deleteProject`
- Update any integration tests that call `databaseClusters` → use `databaseCluster`
- Update any integration tests that call `updateDatabaseCluster` → use `updateProjectCluster`
- Update any integration tests that read `project.clusterId` or `project.clusterInfo` → use `project.cluster`

### Task 5.2 — Run full test suite ✅ (unit tests pass; integration tests pending)

```bash
task test-unit
task auto-test
task check-all
```

All tests must pass.

---

## Phase 6: Validation

### Task 6.1 — Validate openspec

```bash
openspec validate refactor-cluster-as-project-subresource --strict
```

Resolve any issues before marking complete.

### Task 6.2 — Manual smoke test

Using GraphQL Playground (`http://localhost:8080/playground`):

- [ ] `createProject` with `clusterInput` creates both project and cluster
- [ ] `createProject` with `skipConnectionTest: true` skips connection test
- [ ] `createProject` with bad connection info returns `DatabaseConnectionFailed`
- [ ] `project { cluster { name connectionInfo { host } } }` returns cluster nested under project
- [ ] `updateProjectCluster` updates the cluster connection info
- [ ] `deleteProject` removes both project and cluster
- [ ] `databaseCluster(projectName: ...)` still works
- [ ] `createDatabaseCluster` returns "unknown field" error (removed)
- [ ] `databaseClusters` returns "unknown field" error (removed)

---

## Dependencies

```
1.1 → 1.3 (schema must be done before generate)
1.2 → 1.3 (merge cluster.graphql into project.graphql before generate)
1.3 → 3.1, 3.2, 3.3, 3.4, 3.5 (generated interfaces needed)
2.1 → 2.2 (command struct before implementation)
2.3 (independent)
2.4 (independent)
3.x → 4.x (resolvers before mapper updates)
All phases → 5.1, 5.2 (tests after implementation)
5.2 → 6.1, 6.2
```
