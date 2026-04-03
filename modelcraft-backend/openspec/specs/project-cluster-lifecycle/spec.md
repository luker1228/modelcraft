# Spec Delta: project-management

## MODIFIED Requirements

### Requirement: Project Creation Requires Cluster Connection

Creating a project MUST atomically create an associated database cluster in the same transaction.
A project without a cluster is not a valid state.

**Rationale**: A project without a cluster cannot store any model data. Requiring cluster creation
at project creation time prevents projects from existing in an incomplete state.

#### Scenario: Create project with valid cluster connection info succeeds

- **GIVEN** a valid project name and cluster connection info
- **WHEN** client calls `createProject` with `clusterInput` containing valid connection details
- **THEN** both the project and cluster SHALL be created atomically in a single transaction
- **AND** the response SHALL include the created project with its `cluster` field populated
- **AND** the connection SHALL be tested before creation (unless `skipConnectionTest: true`)

#### Scenario: Create project with connection test failure returns error

- **WHEN** client calls `createProject` with `clusterInput` and the database connection test fails
- **THEN** the response SHALL include a `DatabaseConnectionFailed` error
- **AND** neither the project nor the cluster SHALL be created
- **AND** the error SHALL include `message` and `suggestion` fields

#### Scenario: Create project with skipConnectionTest skips validation

- **WHEN** client calls `createProject` with `skipConnectionTest: true`
- **THEN** the connection info SHALL NOT be tested before creation
- **AND** the project and cluster SHALL be created regardless of whether the connection is reachable
- **AND** the response SHALL include the created project with its `cluster` field populated

#### Scenario: Create project with missing clusterInput fails validation

- **WHEN** client calls `createProject` without `clusterInput`
- **THEN** the request SHALL be rejected with an `InvalidProjectInput` error
- **AND** neither the project nor the cluster SHALL be created

#### Scenario: Create project transaction rolls back on cluster creation failure

- **WHEN** cluster creation fails after project record is inserted (e.g., DB constraint violation)
- **THEN** the project record SHALL also be rolled back
- **AND** no partial state SHALL persist in the database

### Requirement: Project Type Exposes Cluster as Direct Field

The `Project` GraphQL type SHALL expose its cluster as a direct `cluster` field, removing
`clusterId` and `clusterInfo` fields.

**Rationale**: Cluster is a sub-entity of Project, not an independent resource. Exposing it as a
direct field reflects the correct domain relationship and simplifies client code.

#### Scenario: Query project returns cluster field

- **WHEN** client queries `project(name: "my-project") { cluster { name connectionInfo { host } } }`
- **THEN** the response SHALL include the cluster object nested under the project
- **AND** the cluster SHALL include `name`, `title`, `description`, `connectionInfo`, `status`, `version`

#### Scenario: Project type no longer exposes clusterId or clusterInfo

- **WHEN** client introspects the `Project` type
- **THEN** `clusterId` field SHALL NOT exist
- **AND** `clusterInfo` field SHALL NOT exist
- **AND** `cluster: DatabaseCluster!` SHALL be present

### Requirement: Delete Project Cascades to Cluster

Deleting a project MUST also delete its associated cluster in the same transaction.

**Rationale**: Cluster has no independent existence. When the project is deleted, its cluster
configuration has no remaining purpose.

#### Scenario: Delete project also deletes its cluster

- **GIVEN** a project with an associated cluster
- **WHEN** client calls `deleteProject(name: "my-project")`
- **THEN** both the project and its cluster SHALL be deleted atomically
- **AND** the response SHALL include `success: true`

#### Scenario: Delete project with no cluster succeeds

- **GIVEN** a project without an associated cluster (legacy data)
- **WHEN** client calls `deleteProject(name: "my-project")`
- **THEN** only the project SHALL be deleted
- **AND** the response SHALL include `success: true`

### Requirement: CreateProjectError Union Includes DatabaseConnectionFailed

The `CreateProjectError` union SHALL include `DatabaseConnectionFailed` to handle connection
test failures during project creation.

#### Scenario: CreateProjectError union includes connection failure type

- **WHEN** client introspects `CreateProjectError` union type
- **THEN** it SHALL include `ProjectAlreadyExists`
- **AND** it SHALL include `InvalidProjectInput`
- **AND** it SHALL include `DatabaseConnectionFailed`

## REMOVED Requirements

### Requirement: Project Supports Optional Cluster Association via clusterId

*(Previously, `Project` had `clusterId: String` for optional cluster association.
This is replaced by the mandatory `cluster: DatabaseCluster!` field.)*
