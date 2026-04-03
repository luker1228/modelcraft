## ADDED Requirements

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

## MODIFIED Requirements

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
