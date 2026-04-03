# Spec: Database Cluster - Project Scope

## Overview

This spec modifies the existing Database Cluster capability to operate within project scope. Clusters now belong to projects, and cluster names are unique within each project.

## MODIFIED Requirements

### Requirement: Cluster Belongs to Project

Every database cluster MUST belong to exactly one project.

**Changes from Previous:**
- Add `project_id` field to DatabaseCluster entity
- Cluster name uniqueness scoped to project (not globally unique)
- Cluster queries filtered by project

**Acceptance Criteria:**
- Cluster has required `project_id` field
- Cluster name unique within project, can duplicate across projects
- Cluster full identifier format: `project_id.cluster_name`

#### Scenario: Create cluster in specific project

**Given** project "ecommerce" exists
**When** user creates cluster "prod-db" in project "ecommerce" with:
- name: "prod-db"
- host: "db.example.com"
- port: 3306

**Then** cluster is created successfully
**And** cluster.projectId equals "ecommerce"
**And** cluster is visible when querying project "ecommerce"

#### Scenario: Create cluster in non-existent project

**Given** no project exists with ID "nonexistent"
**When** user attempts to create cluster in project "nonexistent"
**Then** creation fails with error "project not found"

#### Scenario: Create duplicate cluster name in same project

**Given** project "ecommerce" exists
**And** cluster "prod-db" exists in project "ecommerce"
**When** user attempts to create another cluster "prod-db" in project "ecommerce"
**Then** creation fails with error "cluster name already exists in project"

#### Scenario: Create same cluster name in different projects

**Given** projects "ecommerce" and "crm" exist
**And** cluster "prod-db" exists in project "ecommerce"
**When** user creates cluster "prod-db" in project "crm"
**Then** cluster is created successfully
**And** both clusters exist independently
**And** clusters have different project_id values

---

### Requirement: Query Clusters by Project

Users MUST be able to query clusters filtered by project.

**Changes from Previous:**
- Add projectId parameter to list/search operations
- Filter results by project_id in repository layer

**Acceptance Criteria:**
- Can list clusters in specific project
- Clusters from other projects are not visible in query results
- Cluster lookup by name searches within project scope

#### Scenario: List clusters in specific project

**Given** clusters exist:
- "prod-db" in project "ecommerce"
- "staging-db" in project "ecommerce"
- "prod-db" in project "crm"

**When** user lists clusters with projectId "ecommerce"
**Then** response contains 2 clusters: "prod-db" and "staging-db"
**And** response does not contain "prod-db" from project "crm"

#### Scenario: Get cluster by name within project

**Given** cluster "prod-db" exists in both projects "ecommerce" and "crm"
**When** user gets cluster "prod-db" in project "ecommerce"
**Then** returns cluster from project "ecommerce"
**And** does not return cluster from project "crm"


**Given** clusters exist in multiple projects
**When** user lists clusters without specifying projectId
**Then** only clusters in "default" project are returned
**And** clusters from other projects are excluded

---

### Requirement: Update Cluster Unique Constraint

Cluster name uniqueness MUST be enforced at project scope, not globally.

**Changes from Previous:**
- Remove global unique constraint on cluster name
- Add composite unique constraint on (project_id, name)

**Acceptance Criteria:**
- Same cluster name can exist in multiple projects
- Duplicate cluster names within same project are rejected
- Database constraint enforces uniqueness

#### Scenario: Database enforces project-scoped uniqueness

**Given** database has cluster "prod-db" in project "ecommerce"
**When** direct database insert attempts to add cluster "prod-db" in project "ecommerce"
**Then** database raises unique constraint violation error
**And** insert is rolled back

---


## ADDED Requirements

### Requirement: Delete Project Cascades to Clusters

When a project is deleted, all clusters in that project MUST be deleted.

**Acceptance Criteria:**
- Deleting project cascades to all clusters in that project
- Clusters in other projects are unaffected
- Foreign key constraint enforces cascade
- Deleting cluster with active models may be restricted (existing behavior)

#### Scenario: Delete project removes all clusters

**Given** project "old-app" exists
**And** clusters "db1", "db2" exist in project "old-app"
**And** cluster "prod-db" exists in project "active-app"
**When** user deletes project "old-app" (with force flag)
**Then** project "old-app" is deleted
**And** clusters "db1" and "db2" are deleted
**And** cluster "prod-db" in project "active-app" still exists

---

### Requirement: Cluster Connection Management Per Project

Connection pools MUST be managed with awareness of project context for better resource isolation.

**Acceptance Criteria:**
- Connection manager can identify connections by project and cluster
- Metrics and monitoring can be aggregated by project
- Connection pool limits can be configured per project (future)

#### Scenario: List active connections by project

**Given** clusters exist in projects "ecommerce" and "crm"
**And** active connections exist to clusters in both projects
**When** administrator queries connections for project "ecommerce"
**Then** only connections to clusters in "ecommerce" are returned
**And** connection counts are accurate for the project

---

### Requirement: Cluster IDs Are Project-Scoped

Cluster IDs MUST be unique only within their project, not globally.

**Acceptance Criteria:**
- Cluster ID can be an auto-increment integer within project scope
- Two clusters in different projects can have the same ID
- Cluster lookup requires both project_id and cluster_id
- Full cluster reference format: `{project_id}/{cluster_id}` or composite key

#### Scenario: Same cluster ID in different projects

**Given** projects "ecommerce" and "crm" exist
**When** cluster with ID "1" is created in project "ecommerce"
**And** cluster with ID "1" is created in project "crm"
**Then** both clusters exist independently
**And** clusters are distinct entities
**And** querying cluster "1" requires project context

---

## Database Schema Changes

### Updated Table: database_clusters

```sql
-- Add project_id column
ALTER TABLE `database_clusters`
  ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '所属项目ID' AFTER `id`;

-- Add foreign key constraint
ALTER TABLE `database_clusters`
  ADD CONSTRAINT `fk_cluster_project`
  FOREIGN KEY (`project_id`)
  REFERENCES `projects` (`id`)
  ON DELETE CASCADE;

-- Update unique index to include project_id
ALTER TABLE `database_clusters`
  DROP INDEX `idx_cluster_id`;

ALTER TABLE `database_clusters`
  ADD UNIQUE KEY `idx_cluster_name` (`project_id`, `name`)
  COMMENT '项目内集群名称唯一';

-- Add index for project-based queries
ALTER TABLE `database_clusters`
  ADD KEY `idx_project_status` (`project_id`, `status`);
```

## API Changes

### GraphQL

```graphql
# Updated DatabaseCluster type
type DatabaseCluster {
  id: ID!
  projectId: ID!          # NEW FIELD
  name: String!
  title: String!
  # ... other fields
}

# Updated query
extend type Query {
  # New: filter by project
  clusters(projectId: ID, status: ClusterStatus): [DatabaseCluster!]!

  # Updated: require project for lookups
  cluster(projectId: ID!, name: String!): DatabaseCluster
  clusterById(id: ID!): DatabaseCluster  # ID-based lookup still works
}

# Updated mutation
input CreateDatabaseClusterInput {
  projectId: ID!         # NEW REQUIRED FIELD
  name: String!
  title: String!
  host: String!
  port: Int!
  username: String!
  password: String!
  # ... other fields
}

input UpdateDatabaseClusterInput {
  id: ID!
  # projectId cannot be changed
  title: String
  # ... other updatable fields
}
```

### REST

```
# New project-scoped endpoints
GET    /api/v1/projects/:projectId/clusters
POST   /api/v1/projects/:projectId/clusters
GET    /api/v1/projects/:projectId/clusters/:clusterName
PUT    /api/v1/projects/:projectId/clusters/:clusterId
DELETE /api/v1/projects/:projectId/clusters/:clusterId

# Backward compatible (uses "default" project)
GET    /api/v1/clusters
POST   /api/v1/clusters
GET    /api/v1/clusters/:clusterId
```

## Repository Layer Changes

### Updated Interface: DatabaseClusterRepository

```go
type DatabaseClusterRepository interface {
    // Updated: GetByName now includes projectId
    GetByName(ctx context.Context, projectId, name string) (*DatabaseCluster, error)

    // New: List by project
    ListByProject(ctx context.Context, projectId string, status ...ClusterStatus) ([]*DatabaseCluster, error)

    // New: Check existence in project
    ExistsByNameInProject(ctx context.Context, projectId, name string) (bool, error)

    // Existing methods remain, updated implementations
    Create(ctx context.Context, cluster *DatabaseCluster) error
    Update(ctx context.Context, cluster *DatabaseCluster) error
    GetByID(ctx context.Context, id string) (*DatabaseCluster, error)
    Delete(ctx context.Context, id string) error
    // ... other methods
}
```

## Validation Changes

### Cluster Creation Validation

```go
func (s *ClusterAppService) CreateCluster(
    ctx context.Context,
    projectId, name, title, host string,
    port int,
    username, password string,
) (*cluster.DatabaseCluster, error) {
    // 1. Validate project exists
    _, err := s.projectRepo.GetByID(ctx, projectId)
    if err != nil {
        return nil, bizerrors.Errorf("project not found: %s", projectId)
    }

    // 2. Check name uniqueness within project
    exists, err := s.clusterRepo.ExistsByNameInProject(ctx, projectId, name)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, bizerrors.Errorf("cluster name already exists in project: %s", name)
    }

    // 3. Create cluster
    // ... rest of implementation
}
```
