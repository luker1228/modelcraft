# Cluster Management Capability

## ADDED Requirements

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

#### Scenario: Unknown bizerrors code returns safe default error

- **WHEN** a domain service returns a `bizerrors.BusinessError` with an unknown or unmapped code
- **THEN** the error adapter SHALL log a warning
- **AND** SHALL return a safe default error type appropriate to the operation
- **AND** SHALL not expose internal error details in the GraphQL response
- **AND** SHALL populate message from the business error

## MODIFIED Requirements

None - this change adds new error handling without modifying existing functionality.

## REMOVED Requirements

None - backward compatibility is maintained.
