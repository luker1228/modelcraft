# Implementation Tasks

## 1. Database Schema Changes
- [x] 1.1 Update `db/schema/mysql/01_project.sql` to add `cluster_id VARCHAR(36) NULL` field
- [x] 1.2 Update `db/schema/mysql/01_project.sql` to add index `KEY idx_project_cluster (cluster_id)`
- [x] 1.3 Update `db/schema/mysql/02_database_cluster.sql` to add unique constraint `UNIQUE KEY idx_cluster_project_unique (org_name, project_name)`
- [ ] 1.4 Test schema changes by running `task db:migrate-up` on development database
- [ ] 1.5 Verify constraints are created correctly by checking MySQL schema
- [ ] 1.6 Document rollback procedure (DROP CONSTRAINT, DROP COLUMN)

## 2. Domain Model Updates
- [x] 2.1 Add `ClusterID *string` field to `Project` struct in `internal/domain/project/project.go`
- [x] 2.2 Update `Project.Validate()` method to validate cluster reference (if provided)
- [x] 2.3 Add `SetCluster(clusterID string)` method to `Project` entity
- [x] 2.4 Add `GetClusterID()` method to return cluster ID (handling nil)
- [x] 2.5 Write unit tests for new Project methods
- [x] 2.6 Update `DatabaseCluster` validation logic (if needed for one-to-one constraint)
- [x] 2.7 Write unit tests for cluster constraint validation

## 3. Repository Layer Updates
- [x] 3.1 Update `ProjectRepository` interface to support cluster_id in Create/Update operations
- [x] 3.2 Update sqlc model in `internal/infrastructure/repository/project_repository.go` to include `ClusterID` field
- [x] 3.3 Update `Create()` method to handle cluster_id
- [x] 3.4 Update `Update()` method to handle cluster_id changes
- [x] 3.5 Add `GetByClusterID(ctx, orgName, clusterID)` method with multi-tenant isolation (requires orgName)
- [x] 3.6 Update `ClusterRepository` to validate one-to-one constraint before creation
- [x] 3.7 Add `GetByProjectKey(ctx, orgName, projectName)` method to ClusterRepository
- [ ] 3.8 Write unit tests for repository changes

## 4. Application Service Layer Updates
- [ ] 4.1 Update `ProjectService.CreateProject()` to accept optional `clusterID` parameter
- [ ] 4.2 Add validation: if `clusterID` provided, verify cluster exists and belongs to same org/project
- [ ] 4.3 Update `ProjectService.UpdateProject()` to support cluster assignment/change
- [ ] 4.4 Add business rule: prevent assigning cluster if cluster already assigned to another project
- [x] 4.5 Update `ClusterService.CreateCluster()` to validate one-project-one-cluster constraint
- [x] 4.6 Add business rule: prevent creating second cluster for a project that already has one
- [x] 4.7 Write unit tests for service layer changes (use mocks for repositories)
- [ ] 4.8 Add integration tests for create/update scenarios with cluster assignment

## 5. GraphQL Schema Updates
- [x] 5.1 Add `clusterId: String` field to `Project` type in `api/graph/schema/project.graphql`
- [x] 5.2 Add `clusterInfo: DatabaseCluster` field to `Project` type for convenient access
- [x] 5.3 Add `clusterId: String` parameter to `CreateProjectInput` type (optional)
- [x] 5.4 Add `clusterId: String` parameter to `UpdateProjectInput` type (optional)
- [x] 5.5 Run `task generate-gql` to regenerate GraphQL code
- [x] 5.6 Review generated types in `internal/interfaces/graphql/generated/model_gen.go`

## 6. GraphQL Resolver Implementation
- [x] 6.1 Update `project.resolvers.go` to resolve `clusterId` field from Project entity
- [ ] 6.2 Implement `clusterInfo` resolver to fetch cluster by ID (use DataLoader if needed)
- [ ] 6.3 Update `CreateProject` resolver to accept and validate `clusterId` input
- [ ] 6.4 Update `UpdateProject` resolver to accept and validate `clusterId` input
- [ ] 6.5 Add error handling for cluster-not-found scenarios
- [x] 6.6 Update project mapper in `internal/interfaces/graphql/adapter/project_mapper.go`
- [ ] 6.7 Write unit tests for resolver changes

## 7. Testing
- [x] 7.1 Write integration tests for creating project with cluster assignment
- [ ] 7.2 Write integration tests for updating project cluster assignment
- [x] 7.3 Write integration tests for constraint validation (prevent multiple clusters per project)
- [x] 7.4 Write integration tests for backward compatibility (projects without cluster_id)
- [x] 7.5 Write integration tests for GraphQL API changes
- [ ] 7.6 Write integration tests for multi-tenant isolation (orgName required in GetByClusterID)
- [ ] 7.7 Update existing tests that may be affected by schema changes
- [x] 7.8 Run unit tests: `task test-unit`
- [ ] 7.9 Run integration tests: Start server with `task run`, then execute `task auto-test`
- [ ] 7.10 Manual testing with GraphQL Playground

## 8. Documentation Updates
- [ ] 8.1 Update `CLAUDE.md` to reflect new one-to-one relationship
- [ ] 8.2 Update database schema documentation in `db/README.md`
- [ ] 8.3 Update architecture documentation in `docs/00-overview/architecture.md`
- [ ] 8.4 Update GraphQL API examples in `tests/graphql/graphql_queries.md`
- [ ] 8.5 Update project domain documentation in `docs/01-common/domain-models.md`

## 9. Deployment and Validation
- [ ] 9.1 Review all changes with team
- [ ] 9.2 Apply database migration to development: `task db:migrate-up`
- [ ] 9.3 Validate functionality: `task run` then `task auto-test`
- [ ] 9.4 Apply database migration to Docker environment: `task db:migrate-up` (if staging exists)
- [ ] 9.5 Deploy code to Docker environment: `task deploy-docker`
- [ ] 9.6 Run integration tests: `task run` (in Docker) then `task auto-test-env ENV=docker`
- [ ] 9.7 Prepare production deployment plan
- [ ] 9.8 Apply database migration to production: `task db:migrate-up`
- [ ] 9.9 Deploy code to production
- [ ] 9.10 Monitor for errors and validate functionality
- [ ] 9.11 Verify unique constraints are working as expected

## Dependencies
- Task 2 depends on Task 1 (database schema must be finalized)
- Task 3 depends on Task 2 (repositories need updated domain models)
- Task 4 depends on Task 3 (services need updated repositories)
- Task 5 and 6 can be done in parallel with Task 4
- Task 7 depends on all implementation tasks (2-6) being complete
- Task 8 can be done in parallel with Task 7
- Task 9 depends on all previous tasks being complete

## Parallelizable Work
- Tasks 5-6 (GraphQL) can be done in parallel with Task 4 (Services)
- Task 8 (Documentation) can start once Task 5 (Schema) is complete
