## ADDED Requirements

### Requirement: Project-Cluster One-to-One Association

A Project SHALL maintain an optional reference to a single associated DatabaseCluster, enabling direct cluster lookup and enforcing a one-to-one relationship between projects and clusters.

#### Scenario: Project stores cluster ID reference

- **WHEN** a Project is created or updated with a cluster assignment
- **THEN** the Project entity SHALL store the `cluster_id` as a nullable string field
- **AND** the `cluster_id` SHALL reference a valid DatabaseCluster ID if provided
- **AND** a NULL `cluster_id` SHALL be valid (project without assigned cluster)

#### Scenario: Project provides cluster access method

- **WHEN** application code needs to retrieve a Project's associated cluster
- **THEN** the Project entity SHALL provide a `GetClusterID()` method returning the cluster ID or nil
- **AND** the method SHALL handle NULL values gracefully

#### Scenario: Project validates cluster reference on assignment

- **WHEN** a cluster ID is assigned to a Project
- **THEN** the Project entity SHALL validate that the cluster ID is not empty if provided
- **AND** the validation SHALL occur during `SetCluster()` or `UpdateMetadata()` operations
- **AND** invalid cluster IDs SHALL result in a validation error

### Requirement: Project GraphQL API Cluster Fields

The Project GraphQL type SHALL expose cluster reference fields to enable clients to access cluster information directly from project queries.

#### Scenario: Project type includes cluster ID field

- **WHEN** client queries a Project via GraphQL
- **THEN** the response SHALL include a `clusterId` field (nullable String)
- **AND** the field SHALL contain the cluster UUID if assigned
- **AND** the field SHALL be NULL if no cluster is assigned

#### Scenario: Project type includes cluster info resolver

- **WHEN** client queries a Project with `clusterInfo` field
- **THEN** the response SHALL include full DatabaseCluster object
- **AND** the resolver SHALL fetch cluster by `project.clusterId` if present
- **AND** the field SHALL be NULL if `clusterId` is NULL
- **AND** the resolver MAY use DataLoader for efficient batch loading

#### Scenario: Create project mutation accepts cluster ID input

- **WHEN** client calls `createProject` mutation with `clusterId` in input
- **THEN** the mutation SHALL assign the cluster to the newly created project
- **AND** the mutation SHALL validate that the cluster exists
- **AND** the mutation SHALL validate that the cluster is not already assigned to another project
- **AND** invalid cluster ID SHALL return `InvalidProjectInput` error
- **AND** cluster-already-assigned SHALL return appropriate conflict error

#### Scenario: Update project mutation accepts cluster ID input

- **WHEN** client calls `updateProject` mutation with `clusterId` in input
- **THEN** the mutation SHALL update the project's cluster assignment
- **AND** the mutation SHALL validate that the new cluster exists
- **AND** the mutation SHALL validate that the cluster is not already assigned to another project
- **AND** providing `clusterId: null` SHALL unassign the cluster (set to NULL)
- **AND** invalid cluster ID SHALL return `InvalidProjectInput` error

### Requirement: Project Repository Cluster Support

The ProjectRepository SHALL support CRUD operations on the `cluster_id` field and enable efficient queries by cluster ID.

#### Scenario: Repository creates project with cluster reference

- **WHEN** `ProjectRepository.Create()` is called with a Project containing `ClusterID`
- **THEN** the repository SHALL persist the `cluster_id` field to the database
- **AND** NULL values SHALL be stored as database NULL
- **AND** non-NULL values SHALL be stored as VARCHAR(36)

#### Scenario: Repository updates project cluster reference

- **WHEN** `ProjectRepository.Update()` is called with modified `ClusterID`
- **THEN** the repository SHALL update the `cluster_id` field in the database
- **AND** changing from non-NULL to NULL SHALL be supported (unassigning cluster)
- **AND** changing from NULL to non-NULL SHALL be supported (assigning cluster)
- **AND** the updated_at timestamp SHALL be updated

#### Scenario: Repository retrieves project by cluster ID

- **WHEN** `ProjectRepository.GetByClusterID(ctx, orgName, clusterID)` is called
- **THEN** the repository SHALL return the Project with matching `org_name` and `cluster_id`
- **AND** the query SHALL filter by BOTH `org_name` and `cluster_id` to ensure multi-tenant isolation
- **AND** the query SHALL use the index on `projects.cluster_id` for performance
- **AND** return NULL if no project has the specified cluster ID within the organization

#### Scenario: Repository loads cluster ID in all queries

- **WHEN** any project retrieval method is called (GetByKey, List, etc.)
- **THEN** the repository SHALL include the `cluster_id` field in the result
- **AND** NULL values SHALL be mapped to nil in the domain entity

## MODIFIED Requirements

### Requirement: Project Entity Validation

A Project entity SHALL validate all required fields including title, name, org name, status, login URL, and optionally cluster ID reference when provided.

#### Scenario: Project validates all required fields

- **WHEN** `Project.Validate()` is called
- **THEN** validation SHALL enforce:
  - `OrgName` is not empty
  - `Name` is valid format (2-64 chars, lowercase, starts with letter)
  - `Title` is not empty
  - `Status` is either "active" or "archived"
  - `LoginURL` is valid URL format if provided
  - `ClusterID` is valid UUID format if provided (not empty string)
- **AND** validation error SHALL be returned if any rule fails

#### Scenario: Project allows NULL cluster ID

- **WHEN** `Project.Validate()` is called with `ClusterID` set to nil
- **THEN** validation SHALL pass (NULL cluster is valid)
- **AND** no cluster-related validation SHALL be performed

#### Scenario: Project rejects empty cluster ID string

- **WHEN** `Project.Validate()` is called with `ClusterID` set to empty string ""
- **THEN** validation SHALL fail with error "cluster ID cannot be empty if provided"
- **AND** the validation SHALL distinguish between nil (valid) and "" (invalid)
