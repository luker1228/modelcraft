# Design Document: Project-Cluster One-to-One Relationship

## Context

ModelCraft currently implements a one-to-many relationship between Projects and Clusters, where a single project can be associated with multiple database clusters. This design was created to provide flexibility for advanced use cases where a project might need to distribute data across multiple clusters.

### Current State
- **Project**: Has `(org_name, name)` as composite primary key
- **Cluster**: Has `id` as primary key, references project via `(org_name, project_name)`
- **Model**: References cluster via `cluster_name` (not ID)
- **Constraint**: `UNIQUE KEY idx_cluster_name (org_name, project_name, name)` allows multiple clusters per project

### Stakeholders
- Backend developers maintaining the platform
- Frontend developers building UIs
- Operations team managing deployments
- End users creating projects and models

### Problem Statement
Analysis of real-world usage shows:
1. 95%+ of projects use only one cluster
2. Multi-cluster support adds unnecessary complexity to:
   - Cluster selection logic when accessing models
   - UI workflows (forcing users to choose cluster)
   - Query optimization and caching strategies
3. The flexible design contradicts the principle of "do the simplest thing that works"

## Goals / Non-Goals

### Goals
1. **Simplify Data Model**: Establish one-to-one relationship between Project and Cluster
2. **Maintain Backward Compatibility**: Existing projects continue to work without data loss
3. **Improve Query Performance**: Enable direct cluster lookup via project reference
4. **Reduce Cognitive Load**: Remove cluster selection complexity from user workflows
5. **Enforce Constraints**: Prevent accidental creation of multiple clusters per project

### Non-Goals
1. **Multi-cluster support**: Explicitly removing support for multi-cluster projects
2. **Data distribution**: Not implementing cluster sharding or cross-cluster queries
3. **Cluster migration**: Not building tools to move data between clusters
4. **Dynamic cluster switching**: Not supporting runtime cluster changes for existing data

## Decisions

### Decision 1: Bidirectional Reference with Dual Constraints

**Choice**: Implement bidirectional reference between Project and Cluster with database-level constraints.

**Rationale**:
- **Project → Cluster**: Add `cluster_id` to `projects` table for direct lookup (nullable for backward compatibility)
- **Cluster → Project**: Keep `(org_name, project_name)` for reverse lookup and foreign key semantics
- **Constraint**: Add `UNIQUE (org_name, project_name)` on `database_clusters` to enforce one-to-one at database level

**Alternatives Considered**:
1. **Unidirectional (Cluster → Project only)**:
   - ❌ Requires join query to get project's cluster
   - ❌ Less intuitive API (project doesn't "know" its cluster)

2. **Unidirectional (Project → Cluster only)**:
   - ❌ Lose reverse lookup capability
   - ❌ Require removing `org_name`/`project_name` from cluster (breaking change)

3. **Middle Table**:
   - ❌ Over-engineering for simple one-to-one
   - ❌ Adds unnecessary join complexity

**Trade-offs**:
- ✅ Pro: Efficient queries in both directions
- ✅ Pro: Database enforces one-to-one integrity
- ⚠️ Con: Slight redundancy (bidirectional references)
- ⚠️ Con: Migration requires careful sequencing

### Decision 2: Nullable cluster_id with Gradual Migration

**Choice**: Make `projects.cluster_id` nullable and allow gradual migration.

**Rationale**:
- Zero downtime deployment: existing projects continue working
- Flexibility: new projects can use new pattern immediately
- Safety: no forced data migration on deployment day
- Rollback-friendly: can revert schema without data loss

**Alternatives Considered**:
1. **NOT NULL with mandatory migration**:
   - ❌ Requires all projects to have clusters before deployment
   - ❌ High risk if migration fails
   - ❌ Difficult rollback

2. **Separate projects_v2 table**:
   - ❌ Complexity of managing two schemas
   - ❌ Code paths for both old and new
   - ❌ Data synchronization challenges

**Trade-offs**:
- ✅ Pro: Safe, gradual adoption
- ✅ Pro: Easy rollback
- ⚠️ Con: Temporary state of mixed patterns
- ⚠️ Con: Need to handle NULL in queries

### Decision 3: Keep Model.cluster_name Reference

**Choice**: Do NOT change `models` table to use `cluster_id`; continue using `cluster_name`.

**Rationale**:
- Minimize code changes and testing scope
- `cluster_name` is already unique within project scope: `(org_name, project_name, cluster_name)`
- Model queries already work efficiently with name-based lookup
- Changing to `cluster_id` provides no meaningful performance benefit

**Alternatives Considered**:
1. **Change Model to use cluster_id**:
   - ❌ Requires updating all model queries and joins
   - ❌ Breaking change to existing code
   - ❌ Risk of data inconsistency during migration

2. **Support both cluster_id and cluster_name**:
   - ❌ Adds redundancy
   - ❌ Synchronization complexity

**Trade-offs**:
- ✅ Pro: Minimal code changes
- ✅ Pro: No risk to existing model functionality
- ⚠️ Con: Model queries still require (org, project, name) tuple
- ⚠️ Con: Not using "standard" foreign key pattern

### Decision 4: Database Constraint Enforcement

**Choice**: Use `UNIQUE (org_name, project_name)` constraint on `database_clusters` as primary enforcement mechanism.

**Rationale**:
- Database is authoritative source of truth
- Prevents race conditions and concurrent violations
- Enforcement at lowest level ensures integrity
- Application-level validation is secondary safety net

**Alternatives Considered**:
1. **Application-level validation only**:
   - ❌ Race conditions in concurrent requests
   - ❌ Can be bypassed by direct DB access
   - ❌ Less reliable

2. **Trigger-based validation**:
   - ❌ More complex to maintain
   - ❌ Harder to debug
   - ❌ Performance overhead

**Trade-offs**:
- ✅ Pro: Bulletproof integrity
- ✅ Pro: Simple to understand
- ⚠️ Con: Constraint violations raise DB errors (need good error handling)

## Architecture

### Schema Changes

**Before**:
```sql
CREATE TABLE projects (
  org_name VARCHAR(255) NOT NULL,
  name VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  -- ... other fields
  PRIMARY KEY (org_name, name)
);

CREATE TABLE database_clusters (
  id VARCHAR(36) NOT NULL,
  org_name VARCHAR(255) NOT NULL,
  project_name VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  -- ... other fields
  PRIMARY KEY (id),
  UNIQUE KEY idx_cluster_name (org_name, project_name, name)
);
```

**After**:
```sql
CREATE TABLE projects (
  org_name VARCHAR(255) NOT NULL,
  name VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  cluster_id VARCHAR(36) NULL,  -- NEW: nullable reference to cluster
  -- ... other fields
  PRIMARY KEY (org_name, name),
  KEY idx_project_cluster (cluster_id)  -- NEW: index for cluster lookups
);

CREATE TABLE database_clusters (
  id VARCHAR(36) NOT NULL,
  org_name VARCHAR(255) NOT NULL,
  project_name VARCHAR(64) NOT NULL,
  name VARCHAR(64) NOT NULL,
  -- ... other fields
  PRIMARY KEY (id),
  UNIQUE KEY idx_cluster_name (org_name, project_name, name),
  UNIQUE KEY idx_cluster_project_unique (org_name, project_name)  -- NEW: one-to-one constraint
);
```

### Domain Model Changes

**Before**:
```go
type Project struct {
    OrgName     string
    Name        string
    Title       string
    // ...
}

type DatabaseCluster struct {
    ID          string
    OrgName     string
    ProjectName string
    Name        string
    // ...
}
```

**After**:
```go
type Project struct {
    OrgName     string
    Name        string
    Title       string
    ClusterID   *string  // NEW: nullable pointer
    // ...
}

// DatabaseCluster unchanged
type DatabaseCluster struct {
    ID          string
    OrgName     string
    ProjectName string
    Name        string
    // ...
}
```

### GraphQL API Changes

**Before**:
```graphql
type Project {
  orgName: String!
  name: String!
  title: String!
  # ... other fields
}

input CreateProjectInput {
  orgName: String!
  name: String!
  title: String!
}
```

**After**:
```graphql
type Project {
  orgName: String!
  name: String!
  title: String!
  clusterId: String          # NEW: cluster reference
  clusterInfo: DatabaseCluster  # NEW: convenient resolver
  # ... other fields
}

input CreateProjectInput {
  orgName: String!
  name: String!
  title: String!
  clusterId: String  # NEW: optional cluster assignment
}
```

### Query Patterns

**Getting Project's Cluster (Before)**:
```go
// Required join or separate query
clusters, _ := clusterRepo.ListByProject(ctx, project.OrgName, project.Name)
if len(clusters) > 0 {
    cluster := clusters[0]  // Which one to use?
}
```

**Getting Project's Cluster (After)**:
```go
// Direct lookup
if project.ClusterID != nil {
    cluster, _ := clusterRepo.GetByID(ctx, *project.ClusterID)
}
```

## Risks / Trade-offs

### Risk 1: Existing Projects with Multiple Clusters
**Impact**: Medium
**Probability**: Low (based on usage analysis)

**Mitigation**:
1. Pre-migration analysis script to identify affected projects
2. Manual review and consolidation for projects with multiple clusters
3. Clear communication to users about one-to-one limitation
4. Grace period for users to migrate data if needed

**Rollback**: Remove constraint, revert code changes

### Risk 2: Race Condition During Creation
**Impact**: Low (error message, retry works)
**Probability**: Low (rare concurrent creation)

**Scenario**: Two requests try to create clusters for same project simultaneously.

**Mitigation**:
1. Database constraint prevents corruption
2. Application returns clear error: "Project already has a cluster"
3. Retry logic in client if appropriate

### Risk 3: Data Migration Complexity
**Impact**: Medium
**Probability**: Medium

**Mitigation**:
1. Nullable `cluster_id` allows gradual migration
2. Optional migration script for automated population
3. Backward compatibility ensures no forced migration
4. Clear documentation and runbooks

### Risk 4: Breaking Change for Multi-cluster Users
**Impact**: High (for affected users)
**Probability**: Very Low (~5% of users)

**Mitigation**:
1. Pre-migration notification to all users
2. Analysis script to identify affected users
3. Support for data consolidation or refactoring
4. Extended transition period before enforcing constraint

## Migration Plan

### Phase 1: Schema Preparation (Week 1)
1. Update `db/schema/mysql/01_project.sql` to add `cluster_id` column (nullable)
2. Update `db/schema/mysql/01_project.sql` to add index on `cluster_id`
3. Update `db/schema/mysql/02_database_cluster.sql` to add `idx_cluster_project_unique` constraint
4. Run `task db:migrate-up` to apply schema changes to development database
5. Validate constraint behavior and verify schema structure

### Phase 2: Code Implementation (Week 2-3)
1. Update domain models
2. Update repositories
3. Update services with validation logic
4. Update GraphQL schema and resolvers
5. Comprehensive testing (unit + integration)
6. Deploy code to development: `task deploy-local`

### Phase 3: Data Analysis (Week 3)
1. Run analysis SQL on production data (read-only)
2. Identify projects with multiple clusters
3. Contact affected users
4. Provide migration assistance

### Phase 4: Gradual Rollout (Week 4)
1. Apply schema to staging: `task db:migrate-up` (if staging exists)
2. Deploy code to staging: `task deploy-docker`
3. Run integration tests on staging: `task auto-test-env ENV=docker`
4. Apply schema to production: `task db:migrate-up` during low-traffic window
5. Deploy code to production
6. Monitor for errors

### Phase 5: Data Migration (Week 5+, optional)
1. Run data population SQL to set `cluster_id` for existing projects
2. Validate data consistency
3. Monitor query performance
4. Consider adding NOT NULL constraint (future phase)

### Rollback Steps (if needed)
1. **Before data migration**:
   - Revert code changes
   - Run SQL to drop constraint: `ALTER TABLE database_clusters DROP INDEX idx_cluster_project_unique;`
   - Run SQL to drop column: `ALTER TABLE projects DROP COLUMN cluster_id;`
2. **After data migration**: More complex, requires keeping bidirectional references

## Open Questions

### Q1: Should we add NOT NULL constraint to cluster_id in future?
**Status**: Deferred

**Options**:
- Make cluster required after grace period
- Keep optional forever (allow projects without clusters)

**Decision Timeline**: Revisit after 3 months of production usage

### Q2: Should we build a cluster migration tool?
**Status**: Deferred

**Rationale**: Only needed if users request it; build if demand materializes

### Q3: How to handle projects created before cluster feature existed?
**Status**: Resolved

**Answer**: Allow NULL `cluster_id`; users can assign cluster later via update mutation

### Q4: Should Project creation enforce cluster assignment?
**Status**: Resolved

**Answer**: No, make it optional. Projects can be created first, cluster assigned later.

## Success Metrics

### Technical Metrics
- [ ] Zero data loss during migration
- [ ] <10ms query performance for project-to-cluster lookup
- [ ] 100% test coverage for new constraint logic
- [ ] Zero constraint violation errors in production (after 1 week)

### User Experience Metrics
- [ ] Reduced UI clicks for model creation (no cluster selection needed)
- [ ] Fewer support tickets about "which cluster to use"
- [ ] Positive feedback from developers on simpler API

### Operational Metrics
- [ ] Successful deployment with zero downtime
- [ ] Clean rollback capability maintained for 2 weeks
- [ ] Complete documentation and runbooks

## References

- Database schema: `db/schema/mysql/01_project.sql`, `02_database_cluster.sql`
- Domain models: `internal/domain/project/`, `internal/domain/cluster/`
- GraphQL schema: `api/graph/schema/project.graphql`, `cluster.graphql`
- Related specs: `openspec/specs/project-management/`, `openspec/specs/cluster-management/`
