# Spec Delta: cluster-management

## MODIFIED Requirements

### Requirement: Cluster Lifecycle Managed Through Project

A database cluster SHALL only be created and deleted as part of project lifecycle operations.
Independent `createDatabaseCluster` and `deleteDatabaseCluster` mutations SHALL be removed.

**Rationale**: Cluster is a sub-entity of Project. Its lifecycle is fully owned by the Project
aggregate. Exposing independent create/delete operations incorrectly implies cluster independence.

#### Scenario: No independent createDatabaseCluster mutation exists

- **WHEN** client introspects the `Mutation` type
- **THEN** `createDatabaseCluster` SHALL NOT be present
- **AND** cluster creation is only possible through `createProject`

#### Scenario: No independent deleteDatabaseCluster mutation exists

- **WHEN** client introspects the `Mutation` type
- **THEN** `deleteDatabaseCluster` SHALL NOT be present
- **AND** cluster deletion is only possible through `deleteProject`

### Requirement: Cluster Connection Updated via updateProjectCluster

Updating a cluster's connection configuration SHALL use `updateProjectCluster(projectName, input)`,
replacing the old `updateDatabaseCluster(projectName, name, input)`.

**Rationale**: Since there is exactly one cluster per project, the `name` parameter in
`updateDatabaseCluster` is redundant. The new API reflects the one-to-one relationship and
the cluster's sub-resource status.

#### Scenario: Update cluster connection info via updateProjectCluster

- **WHEN** client calls `updateProjectCluster(projectName: "my-project", input: { connectionInfo: {...} })`
- **THEN** the cluster associated with "my-project" SHALL be updated
- **AND** the response SHALL include the updated cluster

#### Scenario: updateProjectCluster with connection test failure returns error

- **WHEN** client calls `updateProjectCluster` with new connection info and the connection test fails
- **THEN** the response SHALL include a `DatabaseConnectionFailed` error
- **AND** the cluster SHALL NOT be updated

#### Scenario: updateProjectCluster with skipConnectionTest skips validation

- **WHEN** client calls `updateProjectCluster` with `skipConnectionTest: true`
- **THEN** the connection SHALL NOT be tested
- **AND** the cluster SHALL be updated with the new connection info regardless

#### Scenario: updateProjectCluster with non-existent project returns error

- **WHEN** client calls `updateProjectCluster(projectName: "nonexistent", input: {...})`
- **THEN** the response SHALL include a `ProjectNotFound` error
- **AND** no update SHALL occur

#### Scenario: No updateDatabaseCluster mutation exists

- **WHEN** client introspects the `Mutation` type
- **THEN** `updateDatabaseCluster` SHALL NOT be present
- **AND** cluster connection updates are only possible through `updateProjectCluster`

### Requirement: No databaseClusters Plural Query

The `databaseClusters` plural query SHALL be removed. Since a project has at most one cluster,
use `databaseCluster(projectName)` to retrieve it.

#### Scenario: No databaseClusters query exists

- **WHEN** client introspects the `Query` type
- **THEN** `databaseClusters` SHALL NOT be present
- **AND** `databaseCluster(projectName: String!)` SHALL be the only cluster retrieval query

### Requirement: skipConnectionTest Option on Cluster Write Operations

All cluster write operations that involve connection info SHALL support a `skipConnectionTest`
boolean flag to bypass connection validation when needed.

**Related**: `project-cluster-lifecycle` spec (createProject), this spec (updateProjectCluster)

#### Scenario: skipConnectionTest defaults to false

- **WHEN** client calls a cluster write operation without specifying `skipConnectionTest`
- **THEN** the connection SHALL be tested by default
- **AND** failure of the connection test SHALL prevent the operation from completing

#### Scenario: skipConnectionTest true bypasses connection test

- **WHEN** client calls a cluster write operation with `skipConnectionTest: true`
- **THEN** the connection SHALL NOT be tested
- **AND** the operation SHALL proceed with the provided connection info

## REMOVED Requirements

### Requirement: Independent Cluster CRUD Operations

*(Previously, clusters could be created, updated, and deleted independently via
`createDatabaseCluster`, `updateDatabaseCluster`, and `deleteDatabaseCluster`.
These are replaced by project-scoped operations.)*

### Requirement: databaseClusters Plural Query

*(Previously, `databaseClusters(projectName: String!)` returned a paginated list.
Removed because a project can only have one cluster; use `databaseCluster` instead.)*
