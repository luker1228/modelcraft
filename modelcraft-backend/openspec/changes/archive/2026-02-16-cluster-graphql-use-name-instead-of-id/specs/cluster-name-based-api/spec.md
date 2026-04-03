## ADDED Requirements

### Requirement: Name-Based Cluster Identification

Cluster GraphQL operations MUST use `name: String!` as the primary identifier instead of `id: ID!`, improving API usability and aligning with developer expectations.

**Rationale**: Human-readable names are more intuitive for developers, reduce friction in CLI/IaC workflows, and align with how resources are conceptually referenced in practice.

**Related**: `cluster-management` spec, `graphql-conventions`, `api-design-guidelines`

#### Scenario: Get cluster by name

- **WHEN** client queries `databaseCluster(projectName: "my-project", name: "prod-mysql")`
- **THEN** the system MUST look up the cluster by (projectName, name) composite key
- **AND** return the cluster if found
- **AND** the `id` field MUST still be present in the DatabaseCluster type
- **AND** return ClusterNotFound error if cluster does not exist

#### Scenario: Update cluster by name

- **WHEN** client executes `updateDatabaseCluster(projectName: "my-project", name: "prod-mysql", input: {...})`
- **THEN** the system MUST identify the cluster by (projectName, name)
- **AND** apply the updates to the correct cluster
- **AND** return the updated cluster with its ID preserved
- **AND** return ClusterNotFound error if cluster does not exist

#### Scenario: Delete cluster by name

- **WHEN** client executes `deleteDatabaseCluster(projectName: "my-project", name: "prod-mysql")`
- **THEN** the system MUST identify and delete the cluster by (projectName, name)
- **AND** return `success: true` if deletion succeeds
- **AND** return ClusterNotFound error if cluster does not exist

#### Scenario: Test connection with cluster name

- **WHEN** client executes `testDatabaseConnection(input: {projectName: "my-project", name: "prod-mysql"})`
- **THEN** the system MUST look up the cluster by (projectName, name)
- **AND** test the connection using the cluster's stored connection info
- **AND** return success/failure with connection time
- **AND** return ClusterNotFound error if cluster does not exist

#### Scenario: List databases by cluster name

- **WHEN** client queries `listDatabases(input: {projectName: "my-project", name: "prod-mysql"})`
- **THEN** the system MUST identify the cluster by (projectName, name)
- **AND** return the list of databases for that cluster
- **AND** return ClusterNotFound error if cluster does not exist

### Requirement: Name Uniqueness Enforcement

Cluster names MUST be unique within the scope of a project, enforced at both application and database levels.

**Rationale**: Ensures unambiguous cluster identification. Prevents naming conflicts. Enables name to serve as a reliable identifier.

**Related**: `database-schema` spec, `cluster-validation` requirements

#### Scenario: Create cluster with duplicate name fails

- **WHEN** a cluster named "prod-mysql" exists in project "my-project"
- **AND** client attempts to create another cluster with the same name in the same project
- **THEN** the system MUST return `ClusterAlreadyExists` error
- **AND** the error message MUST include the conflicting name
- **AND** the suggestion MUST be "Please use a different cluster name within this project"

#### Scenario: Name uniqueness is case-sensitive

- **WHEN** a cluster named "Prod-MySQL" exists in project "my-project"
- **AND** client attempts to create a cluster named "prod-mysql" (different case)
- **THEN** the system MUST allow the creation (case-sensitive uniqueness)
- **AND** both clusters can coexist with different names

#### Scenario: Database-level unique constraint

- **WHEN** the clusters table is inspected
- **THEN** there MUST be a unique constraint on (project_name, name)
- **AND** the constraint MUST be enforced at the database level
- **AND** concurrent creation attempts MUST be prevented by the constraint

### Requirement: Query Schema Changes

The GraphQL schema MUST replace ID-based parameters with name-based parameters for all cluster queries.

**Rationale**: Consistency across query operations. Simplified API surface. Removal of redundant query variants.

**Related**: `cluster.graphql` schema definition

#### Scenario: databaseCluster query signature

- **WHEN** introspecting the `databaseCluster` query
- **THEN** it MUST have signature: `databaseCluster(projectName: String!, name: String!): GetClusterPayload!`
- **AND** it MUST NOT have an `id` parameter
- **AND** it MUST have the `@hasPermission(action: "cluster:read")` directive

#### Scenario: databaseClusterByName query is removed

- **WHEN** introspecting the GraphQL schema
- **THEN** the `databaseClusterByName` query MUST NOT exist
- **AND** clients MUST use the updated `databaseCluster` query with name parameter

#### Scenario: databaseClusters list query unchanged

- **WHEN** introspecting the `databaseClusters` query
- **THEN** it MUST have signature: `databaseClusters(projectName: String!, input: DatabaseClusterQueryInput): DatabaseClusterConnection!`
- **AND** it MUST remain unchanged (already uses projectName scoping)

### Requirement: Mutation Schema Changes

The GraphQL schema MUST replace ID-based parameters with name-based parameters for all cluster mutations.

**Rationale**: Consistency with query operations. Improved usability for update/delete operations.

**Related**: `cluster.graphql` schema definition

#### Scenario: updateDatabaseCluster mutation signature

- **WHEN** introspecting the `updateDatabaseCluster` mutation
- **THEN** it MUST have signature: `updateDatabaseCluster(projectName: String!, name: String!, input: UpdateDatabaseClusterInput!): UpdateClusterPayload!`
- **AND** it MUST NOT have an `id` parameter
- **AND** it MUST have the `@hasPermission(action: "cluster:update")` directive

#### Scenario: deleteDatabaseCluster mutation signature

- **WHEN** introspecting the `deleteDatabaseCluster` mutation
- **THEN** it MUST have signature: `deleteDatabaseCluster(projectName: String!, name: String!): DeleteClusterPayload!`
- **AND** it MUST NOT have an `id` parameter
- **AND** it MUST have the `@hasPermission(action: "cluster:delete")` directive

#### Scenario: createDatabaseCluster mutation unchanged

- **WHEN** introspecting the `createDatabaseCluster` mutation
- **THEN** it MUST have signature: `createDatabaseCluster(input: CreateDatabaseClusterInput!): CreateClusterPayload!`
- **AND** it MUST remain unchanged (already uses name in input)

### Requirement: Input Type Changes

Input types that reference clusters by ID MUST be updated to use name instead.

**Rationale**: Consistency across all input types. Complete migration from ID to name paradigm.

**Related**: `cluster.graphql` input type definitions

#### Scenario: TestDatabaseConnectionInput uses name

- **WHEN** introspecting the `TestDatabaseConnectionInput` input type
- **THEN** it MUST have an optional `name: String` field
- **AND** it MUST NOT have an `id` field
- **AND** it MUST have optional `projectName: String` field
- **AND** it MUST have optional `connectionInfo: DatabaseConnectionInput` field

#### Scenario: ListDatabasesInput uses name

- **WHEN** introspecting the `ListDatabasesInput` input type
- **THEN** it MUST have a required `name: String!` field (cluster name)
- **AND** it MUST have a required `projectName: String!` field
- **AND** it MUST NOT have a `clusterId` or `id` field

### Requirement: DatabaseCluster Type Retains ID

The `DatabaseCluster` type MUST retain the `id` field for Node interface compliance and backward compatibility.

**Rationale**: Maintains GraphQL Node interface contract. Allows internal systems to use IDs if needed. Provides flexibility for future relay-style pagination.

**Related**: GraphQL Node interface specification

#### Scenario: DatabaseCluster type has both id and name

- **WHEN** introspecting the `DatabaseCluster` type
- **THEN** it MUST implement the `Node` interface
- **AND** it MUST have a required `id: ID!` field
- **AND** it MUST have a required `name: String!` field
- **AND** it MUST have a required `projectName: String!` field
- **AND** all other fields MUST remain unchanged

#### Scenario: ID field is still queryable

- **WHEN** client queries a cluster
- **THEN** the response MUST include the `id` field if requested
- **AND** the `id` value MUST be unique across all clusters
- **AND** the `id` MUST remain stable for the cluster's lifetime

### Requirement: Service Layer Name-Based Lookups

Service layer methods MUST support efficient name-based cluster lookups scoped by project.

**Rationale**: Ensures resolvers can efficiently query by name. Maintains consistent service layer API.

**Related**: `internal/service/cluster.go`, `cluster-service` spec

#### Scenario: GetByName service method

- **WHEN** resolver calls `clusterService.GetByName(ctx, projectName, name)`
- **THEN** the service MUST query the database using (projectName, name) as composite key
- **AND** return the cluster entity if found
- **AND** return `ClusterNotFound` error if not found
- **AND** query performance MUST be comparable to ID-based lookup

#### Scenario: UpdateByName service method

- **WHEN** resolver calls `clusterService.UpdateByName(ctx, projectName, name, input)`
- **THEN** the service MUST locate the cluster by (projectName, name)
- **AND** validate the update input
- **AND** apply the updates to the cluster
- **AND** return the updated cluster
- **AND** return appropriate errors (ClusterNotFound, InvalidClusterInput)

#### Scenario: DeleteByName service method

- **WHEN** resolver calls `clusterService.DeleteByName(ctx, projectName, name)`
- **THEN** the service MUST locate the cluster by (projectName, name)
- **AND** perform soft delete (set deletedAt timestamp)
- **AND** return success
- **AND** return ClusterNotFound error if cluster doesn't exist

### Requirement: Database Indexing for Performance

The database MUST have appropriate indexes to ensure name-based lookups are performant.

**Rationale**: Name lookups must be as fast as ID lookups. Prevents performance regression.

**Related**: `database-schema` spec, database migration files

#### Scenario: Composite index on (project_name, name)

- **WHEN** inspecting the clusters table indexes
- **THEN** there MUST be an index on (project_name, name)
- **AND** the index MUST exclude soft-deleted rows (WHERE deleted_at IS NULL)
- **AND** the index MUST support efficient equality lookups

#### Scenario: Query performance comparable to ID lookups

- **WHEN** executing a name-based lookup query
- **THEN** the execution plan MUST use the (project_name, name) index
- **AND** the query time MUST be comparable to ID-based lookups (within 10% margin)
- **AND** no full table scans occur

### Requirement: Test Coverage for Name-Based Operations

All cluster GraphQL tests MUST be updated to use name-based identifiers and maintain comprehensive coverage.

**Rationale**: Validates the complete name-based workflow. Ensures error cases are properly handled.

**Related**: `tests/design/cluster/test_cluster_graphql.py`, `python-testing-guidelines`

#### Scenario: Test create and retrieve by name

- **WHEN** test creates a cluster with name "test-cluster"
- **AND** test queries the cluster using `databaseCluster(projectName: "...", name: "test-cluster")`
- **THEN** the correct cluster MUST be returned
- **AND** the cluster's ID MUST be present in the response
- **AND** the cluster's name MUST match "test-cluster"

#### Scenario: Test update by name

- **WHEN** test creates a cluster with name "test-cluster"
- **AND** test updates it using `updateDatabaseCluster(projectName: "...", name: "test-cluster", input: {title: "Updated"})`
- **THEN** the update MUST succeed
- **AND** subsequent queries MUST return the updated title
- **AND** the cluster name MUST remain unchanged

#### Scenario: Test delete by name

- **WHEN** test creates a cluster with name "test-cluster"
- **AND** test deletes it using `deleteDatabaseCluster(projectName: "...", name: "test-cluster")`
- **THEN** the deletion MUST succeed
- **AND** subsequent queries for that name MUST return ClusterNotFound error

#### Scenario: Test error case - cluster not found

- **WHEN** test queries a non-existent cluster name
- **THEN** the response MUST include `error` field with `__typename: "ClusterNotFound"`
- **AND** the error message MUST indicate the cluster was not found
- **AND** the `cluster` field MUST be null

#### Scenario: Test cleanup fixtures use name

- **WHEN** test cleanup fixtures run after test completion
- **THEN** they MUST track clusters by (projectName, name) tuples
- **AND** cleanup deletion MUST use name-based deletion
- **AND** all test clusters MUST be properly cleaned up

### Requirement: Backward Compatibility Considerations

This is a BREAKING CHANGE. The migration path MUST be clearly documented.

**Rationale**: Existing API clients will break. Clear communication and migration guidance are essential.

**Related**: `api-versioning` policy, `migration-guides`

#### Scenario: Existing ID-based queries fail with clear errors

- **WHEN** an existing client sends a query with `id` parameter
- **THEN** the GraphQL validation MUST fail
- **AND** the error message MUST indicate that `id` parameter is not recognized
- **AND** the error SHOULD suggest using `name` parameter instead

#### Scenario: Migration guide is provided

- **WHEN** developers consult the migration documentation
- **THEN** they MUST find a comprehensive guide covering:
  - Overview of the breaking change
  - Step-by-step migration instructions
  - Before/after code examples
  - Common pitfalls and solutions
- **AND** the guide MUST be prominently linked in the changelog
