# cluster-management Specification

## Purpose
TBD - created by archiving change implement-cluster-graphql-typed-errors. Update Purpose after archive.
## Requirements
### Requirement: One-to-One Project-Cluster Constraint

The system SHALL enforce a one-to-one relationship between Projects and DatabaseClusters at the database level, ensuring each project has at most one cluster and each cluster belongs to at most one project.

#### Scenario: Database prevents multiple clusters per project

- **WHEN** an attempt is made to create a second DatabaseCluster for a project that already has one
- **THEN** the database SHALL reject the operation with a unique constraint violation
- **AND** the error SHALL reference the `idx_cluster_project_unique (org_name, project_name)` constraint
- **AND** the application layer SHALL catch this error and return a typed error to the client

#### Scenario: Application validates cluster uniqueness before creation

- **WHEN** `ClusterService.CreateCluster()` is called
- **THEN** the service SHALL check if a cluster already exists for the target project
- **AND** if a cluster exists, SHALL return `ClusterAlreadyExistsForProject` error
- **AND** the check SHALL occur before attempting database insertion
- **AND** the error message SHALL include the project name and existing cluster name

#### Scenario: Cluster creation succeeds for project without cluster

- **WHEN** `ClusterService.CreateCluster()` is called for a project with no existing cluster
- **THEN** the cluster SHALL be created successfully
- **AND** the unique constraint SHALL be satisfied
- **AND** the cluster SHALL be associated with the project via `(org_name, project_name)` fields

### Requirement: Cluster Repository Project Lookup

The ClusterRepository SHALL provide methods to query clusters by project key and validate project-cluster relationships.

#### Scenario: Repository retrieves cluster by project key

- **WHEN** `ClusterRepository.GetByProjectKey(ctx, orgName, projectName)` is called
- **THEN** the repository SHALL return the DatabaseCluster associated with the project
- **AND** the query SHALL use the `idx_cluster_project_unique` index for performance
- **AND** return NULL if no cluster exists for the project
- **AND** the method SHALL handle soft-deleted clusters (exclude deleted_at IS NOT NULL)

#### Scenario: Repository checks if project has cluster

- **WHEN** `ClusterRepository.ExistsByProjectKey(ctx, orgName, projectName)` is called
- **THEN** the repository SHALL return true if a cluster exists for the project
- **AND** return false if no cluster exists
- **AND** the query SHALL be optimized (COUNT or EXISTS query, not full object retrieval)

### Requirement: Cluster Service One-to-One Validation

The ClusterService SHALL validate one-to-one relationship constraints during cluster creation and prevent violation attempts.

#### Scenario: Service rejects creation for project with existing cluster

- **WHEN** `ClusterService.CreateCluster()` is called for a project that already has a cluster
- **THEN** the service SHALL return a business error with code `OPERATION_DENIED.CLUSTER`
- **AND** the error message SHALL be "Project {orgName}/{projectName} already has a cluster: {existingClusterName}"
- **AND** the error SHALL include a suggestion: "Please delete the existing cluster first or update it instead"
- **AND** no database insertion SHALL be attempted

#### Scenario: Service handles constraint violation gracefully

- **WHEN** a unique constraint violation occurs during cluster creation (race condition)
- **THEN** the service SHALL catch the database error
- **AND** SHALL return a typed `ClusterAlreadyExistsForProject` error to the client
- **AND** the error SHALL include helpful context (project name, conflict details)
- **AND** SHALL log the constraint violation for debugging

### Requirement: Cluster GraphQL Error Types

The Cluster GraphQL API SHALL provide typed error responses for one-to-one constraint violations.

#### Scenario: Create cluster returns typed error for existing cluster

- **WHEN** client calls `createDatabaseCluster` for a project that already has a cluster
- **THEN** the response SHALL include `ClusterAlreadyExistsForProject` in the `error` field
- **AND** the error SHALL implement the `Error` interface with `message` field
- **AND** the error SHALL include `suggestion` field with guidance
- **AND** the `cluster` field SHALL be null
- **AND** the error type SHALL be distinguishable via `__typename`

#### Scenario: GraphQL schema defines ClusterAlreadyExistsForProject type

- **WHEN** GraphQL schema is introspected
- **THEN** the schema SHALL include `ClusterAlreadyExistsForProject` type
- **AND** the type SHALL implement the `Error` interface
- **AND** the type SHALL have `message: String!` field
- **AND** the type SHALL have `suggestion: String` field (optional)
- **AND** the type SHALL be included in `CreateDatabaseClusterError` union

### Requirement: Typed GraphQL Error Responses for Cluster Operations

The Cluster GraphQL API SHALL provide structured, typed error responses for all mutations and queries, enabling clients to handle specific error scenarios programmatically.

#### Scenario: Get cluster with non-existent project returns typed project not found error

- **WHEN** client calls `databaseCluster` query with a project ID that does not exist
- **THEN** the response SHALL include a `ProjectNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the project ID in the message)
- **AND** the `cluster` field SHALL be null
- **AND** the error type SHALL be distinguishable via `__typename`

#### Scenario: Get cluster by ID returns typed not found error

- **WHEN** client calls `databaseCluster` query with a cluster ID that does not exist
- **THEN** the response SHALL include a `ClusterNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the cluster ID in the message)
- **AND** the `cluster` field SHALL be null
- **AND** the error type SHALL be distinguishable via `__typename`

#### Scenario: Get cluster by name returns typed not found error

- **WHEN** client calls `databaseClusterByName` query with a cluster name that does not exist in the project
- **THEN** the response SHALL include a `ClusterNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the cluster name in the message)
- **AND** the `cluster` field SHALL be null

#### Scenario: Create cluster with non-existent project returns typed project not found error

- **WHEN** client calls `createDatabaseCluster` mutation with a project ID that does not exist
- **THEN** the response SHALL include a `ProjectNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the project ID in the message)
- **AND** the `cluster` field SHALL be null
- **AND** project validation SHALL occur before other validations

#### Scenario: Create cluster with duplicate name returns typed conflict error

- **WHEN** client calls `createDatabaseCluster` mutation with a cluster name that already exists within the same project
- **THEN** the response SHALL include a `ClusterAlreadyExists` in the `error` field
- **AND** the error SHALL contain `message` field with descriptive text (including cluster name)
- **AND** the error SHALL contain `suggestion` field with guidance
- **AND** the `cluster` field SHALL be null
- **AND** the error type SHALL be distinguishable via `__typename`

#### Scenario: Create cluster with invalid input returns typed validation error

- **WHEN** client calls `createDatabaseCluster` mutation with invalid input (e.g., empty name, invalid port, missing connection info)
- **THEN** the response SHALL include an `InvalidClusterInput` in the `error` field
- **AND** the error SHALL contain `message` field describing which input is invalid
- **AND** the error MAY contain `suggestion` field with guidance
- **AND** the `cluster` field SHALL be null

#### Scenario: Create cluster with connection failure returns typed connection error

- **WHEN** client calls `createDatabaseCluster` mutation and the database connection test fails
- **THEN** the response SHALL include a `DatabaseConnectionFailed` in the `error` field
- **AND** the error SHALL contain `message` field describing the connection failure
- **AND** the error SHALL contain `suggestion` field with troubleshooting guidance
- **AND** the `cluster` field SHALL be null

#### Scenario: Update non-existent cluster returns typed not found error

- **WHEN** client calls `updateDatabaseCluster` mutation with a cluster ID that does not exist
- **THEN** the response SHALL include a `ClusterNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the cluster ID in the message)
- **AND** the `cluster` field SHALL be null

#### Scenario: Update cluster with non-existent project returns typed project not found error

- **WHEN** client calls `updateDatabaseCluster` mutation with a project ID that does not exist
- **THEN** the response SHALL include a `ProjectNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the project ID in the message)
- **AND** the `cluster` field SHALL be null
- **AND** project validation SHALL occur before cluster lookup

#### Scenario: Update cluster with invalid input returns typed validation error

- **WHEN** client calls `updateDatabaseCluster` mutation with invalid input (e.g., invalid port, malformed host)
- **THEN** the response SHALL include an `InvalidClusterInput` in the `error` field
- **AND** the error SHALL contain `message` field describing which input is invalid
- **AND** the error MAY contain `suggestion` field with guidance
- **AND** the `cluster` field SHALL be null

#### Scenario: Update cluster with connection failure returns typed connection error

- **WHEN** client calls `updateDatabaseCluster` mutation with new connection info and the database connection test fails
- **THEN** the response SHALL include a `DatabaseConnectionFailed` in the `error` field
- **AND** the error SHALL contain `message` field describing the connection failure
- **AND** the error SHALL contain `suggestion` field with troubleshooting guidance
- **AND** the `cluster` field SHALL be null

#### Scenario: Delete non-existent cluster returns typed not found error

- **WHEN** client calls `deleteDatabaseCluster` mutation with a cluster ID that does not exist
- **THEN** the response SHALL include a `ClusterNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the cluster ID in the message)
- **AND** the `success` field SHALL be false

#### Scenario: Delete cluster with non-existent project returns typed project not found error

- **WHEN** client calls `deleteDatabaseCluster` mutation with a project ID that does not exist
- **THEN** the response SHALL include a `ProjectNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the project ID in the message)
- **AND** the `success` field SHALL be false
- **AND** project validation SHALL occur before cluster lookup

#### Scenario: Test connection for non-existent cluster returns typed not found error

- **WHEN** client calls `testDatabaseConnection` mutation with a cluster ID that does not exist
- **THEN** the response SHALL include a `ClusterNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the cluster ID in the message)
- **AND** the `success` field SHALL be false

#### Scenario: Test connection with non-existent project returns typed project not found error

- **WHEN** client calls `testDatabaseConnection` mutation with a project ID that does not exist
- **THEN** the response SHALL include a `ProjectNotFound` in the `error` field
- **AND** the error SHALL contain `message` field (including the project ID in the message)
- **AND** the `success` field SHALL be false
- **AND** project validation SHALL occur before cluster lookup

#### Scenario: Test connection failure returns typed connection error

- **WHEN** client calls `testDatabaseConnection` mutation and the connection test fails (invalid credentials, unreachable host, etc.)
- **THEN** the response SHALL include a `DatabaseConnectionFailed` in the `error` field
- **AND** the error SHALL contain `message` field describing the connection failure
- **AND** the error SHALL contain `suggestion` field with troubleshooting guidance (e.g., "Please verify host, port, username, and password are correct")
- **AND** the `success` field SHALL be false

#### Scenario: Successful operation returns null error

- **WHEN** any cluster operation succeeds without errors
- **THEN** the response SHALL have the `error` field as null
- **AND** the data field SHALL contain the expected result (cluster, success status, etc.)

### Requirement: Error Interface Implementation

All cluster-specific errors SHALL implement a common `Error` interface with a `message` field for consistent error handling.

#### Scenario: All error types implement Error interface

- **WHEN** any cluster operation returns an error
- **THEN** the error SHALL implement the `Error` interface
- **AND** the error SHALL have a `message: String!` field with a human-readable description

#### Scenario: Client queries error message without knowing specific type

- **WHEN** client queries errors using the `Error` interface
- **THEN** client SHALL be able to retrieve `message` for any error type
- **AND** client can use fragments for specific error types to get additional fields like `suggestion`

### Requirement: Cluster-Specific Error Types

The API SHALL define specific error types for Cluster domain operations with clear messages and optional suggestions.

#### Scenario: ClusterAlreadyExists provides clear message and suggestion

- **WHEN** `ClusterAlreadyExists` is returned
- **THEN** it SHALL include `message: String!` describing the conflict (including cluster name)
- **AND** it SHALL include `suggestion: String` with guidance (e.g., "Please use a different cluster name within this project")

#### Scenario: ClusterNotFound provides clear message

- **WHEN** `ClusterNotFound` is returned
- **THEN** it SHALL include `message: String!` describing the not found error (including the cluster ID or name in message text)

#### Scenario: InvalidClusterInput provides validation details

- **WHEN** `InvalidClusterInput` is returned
- **THEN** it SHALL include `message: String!` describing the validation failure
- **AND** it MAY include `suggestion: String` with guidance for fixing the input (e.g., "Port must be between 1 and 65535")

#### Scenario: DatabaseConnectionFailed provides troubleshooting guidance

- **WHEN** `DatabaseConnectionFailed` is returned
- **THEN** it SHALL include `message: String!` describing the connection failure
- **AND** it SHALL include `suggestion: String` with troubleshooting guidance (e.g., "Please verify host, port, username, and password are correct")

### Requirement: Union Types for Operation-Specific Errors

Each cluster operation SHALL define a union type containing all possible error types for that specific operation.

#### Scenario: GetCluster query has specific error union

- **WHEN** client introspects `GetClusterError` union type
- **THEN** it SHALL include `ClusterNotFound` type
- **AND** it SHALL include `ProjectNotFound` type
- **AND** these SHALL be the only possible error types for databaseCluster query

#### Scenario: CreateCluster mutation has specific error union

- **WHEN** client introspects `CreateClusterError` union type
- **THEN** it SHALL include `ClusterAlreadyExists` type
- **AND** it SHALL include `ClusterAlreadyExistsForProject` type
- **AND** it SHALL include `InvalidClusterInput` type
- **AND** it SHALL include `DatabaseConnectionFailed` type
- **AND** it SHALL include `ProjectNotFound` type
- **AND** these SHALL be the only possible error types for createDatabaseCluster operation

#### Scenario: UpdateCluster mutation has specific error union

- **WHEN** client introspects `UpdateClusterError` union type
- **THEN** it SHALL include `ClusterNotFound` type
- **AND** it SHALL include `InvalidClusterInput` type
- **AND** it SHALL include `DatabaseConnectionFailed` type
- **AND** it SHALL include `ProjectNotFound` type
- **AND** these SHALL be the only possible error types for updateDatabaseCluster operation

#### Scenario: DeleteCluster mutation has specific error union

- **WHEN** client introspects `DeleteClusterError` union type
- **THEN** it SHALL include `ClusterNotFound` type
- **AND** it SHALL include `ProjectNotFound` type
- **AND** these SHALL be the only possible error types for deleteDatabaseCluster operation

#### Scenario: TestConnection mutation has specific error union

- **WHEN** client introspects `TestConnectionError` union type
- **THEN** it SHALL include `ClusterNotFound` type
- **AND** it SHALL include `DatabaseConnectionFailed` type
- **AND** it SHALL include `ProjectNotFound` type
- **AND** these SHALL be the only possible error types for testDatabaseConnection operation

### Requirement: Backward Compatible Payload Structure

All cluster operation payloads SHALL maintain existing fields while adding typed error fields, ensuring zero breaking changes for existing clients.

#### Scenario: GetClusterPayload includes error field

- **WHEN** `databaseCluster` query completes
- **THEN** the payload SHALL include `cluster: DatabaseCluster` field (nullable, null on error)
- **AND** the payload SHALL include `error: GetClusterError` field (null on success)
- **AND** existing clients using only `cluster` field SHALL continue to work

#### Scenario: CreateClusterPayload includes error field

- **WHEN** `createDatabaseCluster` mutation completes
- **THEN** the payload SHALL include `cluster: DatabaseCluster` field (nullable, null on error)
- **AND** the payload SHALL include `error: CreateClusterError` field (null on success)
- **AND** existing clients SHALL continue to work

#### Scenario: UpdateClusterPayload includes error field

- **WHEN** `updateDatabaseCluster` mutation completes
- **THEN** the payload SHALL include `cluster: DatabaseCluster` field (nullable, null on error)
- **AND** the payload SHALL include `error: UpdateClusterError` field (null on success)
- **AND** existing clients SHALL continue to work

#### Scenario: DeleteClusterPayload includes both success and error fields

- **WHEN** `deleteDatabaseCluster` mutation completes
- **THEN** the payload SHALL include `success: Boolean!` field (backward compatible)
- **AND** the payload SHALL include `error: DeleteClusterError` field (null on success)
- **AND** the `success` field SHALL be false when error is present
- **AND** the `success` field SHALL be true when error is null

#### Scenario: TestConnectionPayload includes both success and error fields

- **WHEN** `testDatabaseConnection` mutation completes
- **THEN** the payload SHALL include `success: Boolean!` field (backward compatible)
- **AND** the payload SHALL include `connectionTime: Float` field (nullable)
- **AND** the payload SHALL include `error: TestConnectionError` field (null on success)
- **AND** the `success` field SHALL be false when error is present
- **AND** the `success` field SHALL be true when error is null

### Requirement: Cluster Creation Validation

The system SHALL validate cluster creation requests to ensure uniqueness within project scope and prevent constraint violations.

#### Scenario: Validate cluster name uniqueness within project

- **WHEN** `ClusterService.CreateCluster()` is called
- **THEN** validation SHALL check cluster name is unique within `(org_name, project_name)` scope
- **AND** validation SHALL check project does not already have a cluster (new constraint)
- **AND** validation SHALL verify project exists before cluster creation
- **AND** validation SHALL occur in this order: project exists → project has no cluster → cluster name unique
- **AND** appropriate typed errors SHALL be returned for each validation failure

#### Scenario: Cluster creation enforces database constraints

- **WHEN** a new cluster is created
- **THEN** the database SHALL enforce:
  - `UNIQUE (org_name, project_name, name)` - cluster name unique within project
  - `UNIQUE (org_name, project_name)` - one cluster per project (NEW)
- **AND** constraint violations SHALL be caught and translated to typed errors
- **AND** the application SHALL handle both constraints independently

### Requirement: Database Schema Constraints

The `database_clusters` table SHALL define unique constraints to enforce both cluster name uniqueness and one-to-one project-cluster relationship.

#### Scenario: Cluster table has one-to-one constraint

- **WHEN** database schema is inspected
- **THEN** the `database_clusters` table SHALL have:
  - `UNIQUE KEY idx_cluster_name (org_name, project_name, name)` - existing constraint
  - `UNIQUE KEY idx_cluster_project_unique (org_name, project_name)` - NEW constraint
- **AND** both constraints SHALL be active and enforced
- **AND** the `idx_cluster_project_unique` constraint SHALL prevent multiple clusters per project

#### Scenario: One cluster per project constraint enforced at database level

- **WHEN** an INSERT statement attempts to create a second cluster for a project
- **THEN** the database SHALL reject the operation with error code 1062 (MySQL duplicate entry)
- **AND** the error message SHALL reference `idx_cluster_project_unique` constraint
- **AND** the transaction SHALL be rolled back

### Requirement: Error Adapter Integration

The GraphQL resolver layer SHALL integrate with the existing `pkg/bizerrors` error classification system to convert domain errors to typed GraphQL errors.

#### Scenario: bizerrors CONFLICT.CLUSTER maps to ClusterAlreadyExists

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `CONFLICT.CLUSTER`
- **THEN** the error adapter SHALL convert it to `ClusterAlreadyExists` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`
- **AND** SHALL populate `suggestion` with helpful guidance

#### Scenario: bizerrors NOT_FOUND.CLUSTER maps to ClusterNotFound

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `NOT_FOUND.CLUSTER`
- **THEN** the error adapter SHALL convert it to `ClusterNotFound` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`

#### Scenario: bizerrors NOT_FOUND.PROJECT maps to ProjectNotFound

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `NOT_FOUND.PROJECT`
- **THEN** the error adapter SHALL convert it to `ProjectNotFound` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`

#### Scenario: bizerrors PARAM_INVALID maps to InvalidClusterInput

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `PARAM_INVALID`
- **AND** the error is from a cluster operation
- **THEN** the error adapter SHALL convert it to `InvalidClusterInput` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`
- **AND** MAY populate `suggestion` from `bizErr.Detail()` if available

#### Scenario: bizerrors OPERATION_FAILED.DB_CONNECTION maps to DatabaseConnectionFailed

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `OPERATION_FAILED.DB_CONNECTION`
- **THEN** the error adapter SHALL convert it to `DatabaseConnectionFailed` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`
- **AND** SHALL populate `suggestion` with connection troubleshooting guidance

#### Scenario: bizerrors OPERATION_DENIED.CLUSTER maps to ClusterAlreadyExistsForProject

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `OPERATION_DENIED.CLUSTER`
- **AND** the error context indicates a one-to-one project-cluster violation
- **THEN** the error adapter SHALL convert it to `ClusterAlreadyExistsForProject` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`
- **AND** SHALL populate `suggestion` with guidance to delete or update the existing cluster

#### Scenario: Unknown bizerrors code returns safe default error

- **WHEN** a domain service returns a `bizerrors.BusinessError` with an unknown or unmapped code
- **THEN** the error adapter SHALL log a warning
- **AND** SHALL return a safe default error type appropriate to the operation
- **AND** SHALL not expose internal error details in the GraphQL response
- **AND** SHALL populate message from the business error

