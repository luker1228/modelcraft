## 1. Database Schema

- [x] 1.1 Update `db/schema/mysql/03_model_domain.sql`: remove `tags` column from `models` table, add `group_id VARCHAR(36) NULL` column to `models` table
- [x] 1.2 Add `model_groups` table to `db/schema/mysql/03_model_domain.sql` between `models` and `field_definitions` with `id`, `org_name`, `project_slug`, `name`, `display_order`, `created_at`, `updated_at`, unique index on `(org_name, project_slug, name)`, and order index on `(org_name, project_slug, display_order)`
- [ ] 1.3 Run `task db:migrate-up` to apply schema changes

## 2. Domain Layer

- [x] 2.1 Write unit tests for `ModelGroup` domain entity and `GroupName` value object (validation: `^[a-z][a-z0-9_]*$`, max 64 chars)
- [x] 2.2 Create `ModelGroup` domain entity in `internal/domain/modeldesign/` with fields: `ID`, `OrgName`, `ProjectSlug`, `Name`, `DisplayOrder`, `CreatedAt`, `UpdatedAt`
- [x] 2.3 Remove `Tags []string` from `ModelMeta` struct; add `GroupID *string` field
- [x] 2.4 Write unit tests for `ModelGroupRepository` interface methods
- [x] 2.5 Define `ModelGroupRepository` interface in `internal/domain/modeldesign/` with methods: `Create`, `FindByID`, `FindByName`, `ListByProject`, `Update`, `Delete`, `UpdateModelsGroup`
- [x] 2.6 Write unit tests for lexicographic fractional index utility (midpoint computation, tail/head insertion)
- [x] 2.7 Implement lexicographic fractional index utility in `pkg/` or `internal/domain/modeldesign/`

## 3. Infrastructure Layer

- [x] 3.1 Write unit tests for `ModelGroupRepository` MySQL implementation
- [x] 3.2 Implement `ModelGroupRepository` MySQL implementation in `internal/infrastructure/` mapping `model_groups` table to domain entity
- [x] 3.3 Update `ModelRepository` MySQL implementation to handle `group_id` column (read/write `GroupID` field on model)
- [x] 3.4 Register `ModelGroupRepository` in the dependency injection / wire setup

## 4. Application Layer

- [x] 4.1 Write unit tests for `CreateGroupUseCase` (happy path, duplicate name, invalid name)
- [x] 4.2 Implement `CreateGroupUseCase` in `internal/app/`: validate name, check uniqueness, assign tail `display_order`, persist
- [x] 4.3 Write unit tests for `RenameGroupUseCase` (happy path, not found, duplicate name, invalid name)
- [x] 4.4 Implement `RenameGroupUseCase`: validate name, check uniqueness, single UPDATE
- [x] 4.5 Write unit tests for `DeleteGroupUseCase` (with models, empty, not found)
- [x] 4.6 Implement `DeleteGroupUseCase`: set `group_id = NULL` on models, delete group, single transaction
- [x] 4.7 Write unit tests for `ReorderGroupUseCase` (to head, between two groups, to tail)
- [x] 4.8 Implement `ReorderGroupUseCase`: compute lexicographic midpoint, single UPDATE; detect collision and renumber if needed
- [x] 4.9 Write unit tests for `MoveModelToGroupUseCase` (assign group, move to ungrouped, cross-project rejected)
- [x] 4.10 Implement `MoveModelToGroupUseCase`: validate group belongs to same project, update `group_id`
- [x] 4.11 Write unit tests for `ListGroupsUseCase` (with models, empty project)
- [x] 4.12 Implement `ListGroupsUseCase`: query groups ordered by `display_order`, append virtual ungrouped with models where `group_id IS NULL`

## 5. GraphQL Schema

- [x] 5.1 Add `ModelGroup` type to `api/graph/schema/model.graphql` with fields: `id: ID!`, `name: String!`, `isVirtual: Boolean!`, `displayOrder: String!`, `models: [Model!]!`
- [x] 5.2 Add `group: ModelGroup!` field to `Model` type in `api/graph/schema/model.graphql`
- [x] 5.3 Add group queries to schema: `modelGroups(projectSlug: String!): [ModelGroup!]!`
- [x] 5.4 Add group mutation types and payloads: `CreateGroupPayload`, `RenameGroupPayload`, `DeleteGroupPayload`, `ReorderGroupPayload`, `MoveModelToGroupPayload`
- [x] 5.5 Add group error union types: `CreateGroupError` (GroupAlreadyExists, InvalidGroupName), `RenameGroupError` (GroupAlreadyExists, InvalidGroupName, GroupNotFound), `DeleteGroupError` (GroupNotFound), `ReorderGroupError` (GroupNotFound), `MoveModelToGroupError` (GroupNotFound)
- [x] 5.6 Add group mutations to schema: `createGroup`, `renameGroup`, `deleteGroup`, `reorderGroup`, `moveModelToGroup`
- [x] 5.7 Run `task generate-gql` to regenerate GraphQL code

## 6. GraphQL Resolvers

- [x] 6.1 Write unit tests for group mutation resolvers (createGroup, renameGroup, deleteGroup, reorderGroup, moveModelToGroup)
- [x] 6.2 Implement group mutation resolvers in `internal/interfaces/graphql/` wiring to application use cases
- [x] 6.3 Implement `modelGroups` query resolver
- [x] 6.4 Implement `Model.group` field resolver (returns virtual ungrouped when `GroupID` is nil)
- [x] 6.5 Implement `ModelGroup.models` field resolver
- [x] 6.6 Add group error adapter in `internal/interfaces/graphql/adapter/` converting `bizerrors` codes to GraphQL error types

## 7. Cleanup

- [x] 7.1 Remove `Tags` field from all places it is referenced (domain, infrastructure mapper, any resolver or DTO)
- [x] 7.2 Run `task check-all` (fmt + lint + vet + test) and fix any issues
- [ ] 7.3 Run `task auto-test` integration tests and verify group operations end-to-end
- [x] 7.4 Run `task build` to confirm clean build
