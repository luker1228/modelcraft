# Change: Refactor Project-Cluster Relationship from One-to-Many to One-to-One

## Why

Currently, a Project can be associated with multiple Clusters (one-to-many relationship), which adds complexity to data model management, querying logic, and cluster selection workflows. In practice, most projects use a single dedicated cluster, making the one-to-many relationship unnecessary.

Simplifying to a one-to-one relationship will:
- Reduce architectural complexity and simplify queries
- Eliminate the need for cluster selection logic when accessing models
- Improve data model clarity and maintainability
- Align the schema with actual usage patterns

## What Changes

### Database Schema Changes
- **projects table**: Add `cluster_id VARCHAR(36) NULL` field to store the associated cluster ID
- **database_clusters table**: Add `UNIQUE KEY idx_cluster_project_unique (org_name, project_name)` constraint to enforce one-to-one relationship at database level
- **models table**: No change - continue using `cluster_name` for backward compatibility
- **Other tables**: No changes to field_definitions, model_relations, model_enums, model_field_enum_associations

### Domain Model Changes
- **Project entity**: Add `ClusterID` field to support direct cluster reference
- **DatabaseCluster entity**: No structural changes, constraints enforced at database level
- **Model entity**: No changes - continue referencing cluster by name

### API Changes
- **Project GraphQL API**:
  - Add `clusterId` field to `Project` type (nullable)
  - Add `clusterInfo` field to `Project` type for convenient cluster access
  - Add `clusterId` parameter to `createProject` and `updateProject` mutations (optional)
- **Cluster GraphQL API**: No breaking changes, continue supporting existing queries

### Repository and Service Changes
- **ProjectRepository**: Update to handle `ClusterID` field in CRUD operations
- **ClusterRepository**: Add validation to enforce one-project-one-cluster constraint
- **ProjectService**: Update create/update logic to validate cluster assignment
- **ClusterService**: Add validation to prevent creating multiple clusters for same project

### Migration Strategy
- **Backward Compatibility**: Existing projects without `cluster_id` continue to work
- **Data Migration**: Optional script to populate `cluster_id` for existing projects based on their models' `cluster_name`
- **Gradual Adoption**: New projects can use `cluster_id`, old projects gradually migrate

## Impact

### Affected Specs
- `project-management` - Add cluster reference requirements
- `cluster-management` - Add one-to-one constraint requirements

### Affected Code
- Database schema: `db/schema/mysql/01_project.sql`, `db/schema/mysql/02_database_cluster.sql`
- Domain entities: `internal/domain/project/project.go`, `internal/domain/cluster/database_cluster.go`
- Repositories: `internal/infrastructure/repository/project_repository.go`, `internal/infrastructure/repository/database_cluster_repository.go`
- Services: `internal/app/project/project_service.go`, `internal/app/cluster/cluster_app.go`
- GraphQL schema: `api/graph/schema/project.graphql`, `api/graph/schema/cluster.graphql`
- GraphQL resolvers: `internal/interfaces/graphql/project.resolvers.go`, `internal/interfaces/graphql/cluster.resolvers.go`
- Mappers: `internal/interfaces/graphql/adapter/project_mapper.go`, `internal/interfaces/graphql/adapter/cluster_mapper.go`

### Breaking Changes
None - All changes are backward compatible:
- `projects.cluster_id` is nullable (existing projects can have NULL)
- Model table continues using `cluster_name` (no code changes required)
- GraphQL API adds optional fields (non-breaking)

### Migration Considerations
- **Existing Data**: Projects with multiple clusters will need manual intervention or automated cleanup before full migration
- **Database Constraints**: The unique constraint on `(org_name, project_name)` in `database_clusters` will prevent creating multiple clusters for the same project going forward
- **Testing**: Comprehensive integration tests needed to ensure backward compatibility

## Dependencies
- Database migration must be applied before code deployment
- Recommend applying migration in maintenance window to assess existing data state

## Risks
- **Data Inconsistency**: Projects with multiple clusters need manual resolution
- **Migration Complexity**: If many projects have multiple clusters, migration requires careful planning
- **Rollback**: Adding NOT NULL constraint later (if desired) requires all projects to have `cluster_id` populated

## Rollback Plan
- Schema changes can be rolled back by:
  1. Remove `cluster_id` column from `projects` table
  2. Drop `idx_cluster_project_unique` constraint from `database_clusters` table
  3. Revert code changes
- Safe rollback window: Before enforcing NOT NULL constraint on `cluster_id`
