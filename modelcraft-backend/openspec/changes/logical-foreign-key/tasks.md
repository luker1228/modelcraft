## 1. Database Schema

- [x] 1.1 Remove `model_relations` table from `db/schema/mysql/03_model_domain.sql`
- [x] 1.2 Add `logical_foreign_keys` table to `db/schema/mysql/03_model_domain.sql` with columns: `id`, `pair_id`, `direction ENUM('normal','reverse')`, `model_id`, `ref_model_id`, `source_fields JSON`, `target_fields JSON`, `created_at`, `updated_at`; indexes on `(pair_id)` and `(model_id)`
- [x] 1.3 Add `belongs_to_fk_id VARCHAR(36) NULL` column to `field_definitions` table
- [x] 1.4 Add `relate_fk_id VARCHAR(36) NULL` column to `field_definitions` table
- [ ] 1.5 Run `task db:migrate-up` to apply schema changes

## 2. Domain Layer

- [x] 2.1 Write unit tests for `LogicalForeignKey` domain entity (validate pair consistency, direction enum, field count parity)
- [x] 2.2 Create `LogicalForeignKey` domain entity in `internal/domain/modeldesign/` with fields: `ID`, `PairID`, `Direction`, `ModelID`, `RefModelID`, `SourceFields []string`, `TargetFields []string`, `CreatedAt`, `UpdatedAt`
- [x] 2.3 Define `LogicalFKDirection` type with constants `DirectionNormal`, `DirectionReverse`
- [x] 2.4 Write unit tests for `LogicalForeignKeyRepository` interface methods
- [x] 2.5 Define `LogicalForeignKeyRepository` interface in `internal/domain/modeldesign/` with methods: `Save`, `DeleteByPairID`, `FindByModel`, `FindByPairID`, `FindByBelongsToField`, `FindByRelateField`
- [x] 2.6 Remove `ModelRelation` domain entity and `ModelRelationRepository` interface
- [x] 2.7 Add `BelongsToFKID *string` and `RelateFKID *string` fields to `FieldDefinition` domain entity
- [x] 2.8 Write unit tests for updated `FieldDefinition` entity (validate that `belongs_to_fk_id` and `relate_fk_id` are mutually exclusive; RELATION format requires `relate_fk_id`)

## 3. Infrastructure Layer

- [x] 3.1 Write unit tests for `SqlLogicalForeignKeyRepository`
- [x] 3.2 Implement `SqlLogicalForeignKeyRepository` in `internal/infrastructure/repository/` using sqlc-generated queries
- [x] 3.3 Add sqlc queries for `logical_foreign_keys`: `CreateLogicalForeignKey`, `DeleteByPairID`, `FindByModelID`, `FindByPairID`
- [x] 3.4 Add sqlc queries for FK-related field lookups: `FindFieldsByBelongsToFKID`, `FindFieldsByRelateFKID`
- [x] 3.5 Run `sqlc generate` to regenerate `internal/infrastructure/dbgen/`
- [x] 3.6 Update `FieldDefinition` infrastructure mapper to handle `belongs_to_fk_id` and `relate_fk_id` columns
- [x] 3.7 Remove `SqlModelRelationRepository` and all related sqlc queries
- [x] 3.8 Register `SqlLogicalForeignKeyRepository` in dependency injection / wire setup; remove `ModelRelationRepository` registration

## 4. Application Layer

- [x] 4.1 Write unit tests for `CreateLogicalForeignKeyUseCase` (happy path, FK columns must exist in both models, source/target field count must match)
- [x] 4.2 Implement `CreateLogicalForeignKeyUseCase` in `internal/app/modeldesign/`: validate that all `source_fields` exist on `model_id` and `target_fields` exist on `ref_model_id`; atomically create both `normal` and `reverse` rows with shared `pair_id`
- [x] 4.3 Write unit tests for `DeleteLogicalForeignKeyUseCase` (happy path, blocked when RELATION fields exist)
- [x] 4.4 Implement `DeleteLogicalForeignKeyUseCase`: check no `relate_fk_id` references on either lf row; delete pair by `pair_id`
- [x] 4.5 Write unit tests for updated `AddFieldUseCase` — `format=RELATION` now requires `relate_fk_id`; `belongs_to_fk_id` is set when field name matches a FK column
- [x] 4.6 Update `AddFieldUseCase` to validate and set `relate_fk_id` for RELATION-format fields; reject if referenced FK row does not exist or `model_id` does not match
- [x] 4.7 Write unit tests for updated `RemoveFieldUseCase` — removing a `belongs_to_fk_id` field cascades FK deletion if no RELATION fields remain; removing blocked if RELATION fields reference the same FK
- [x] 4.8 Update `RemoveFieldUseCase`: if field has `relate_fk_id`, allow deletion freely; if field has `belongs_to_fk_id`, check that no `relate_fk_id` fields reference the FK pair, then delete FK pair, then delete field
- [x] 4.9 Write unit tests for model deletion cleanup — orphaned reverse FK rows are deleted when a model is removed
- [x] 4.10 Update model deletion logic: after model is deleted (DB CASCADE removes its FK rows), query `WHERE ref_model_id = deleted_model_id` and delete any remaining orphaned reverse rows
- [x] 4.11 Remove all `ModelRelation`-related use cases and service methods

## 5. GraphQL Schema

- [x] 5.1 Remove `RelationType` enum, `RelationConfig` type, `RelationConfigInput` from `api/graph/schema/field.graphql`
- [x] 5.2 Remove `relationConfig: RelationConfig` from `Field` type
- [x] 5.3 Remove `relationConfig: RelationConfigInput` from `AddFieldInput`
- [x] 5.4 Add `LogicalForeignKey` type to a new or existing schema file with fields: `id: ID!`, `pairId: String!`, `direction: FKDirection!`, `modelId: String!`, `refModelId: String!`, `sourceFields: [String!]!`, `targetFields: [String!]!`
- [x] 5.5 Add `FKDirection` enum: `NORMAL`, `REVERSE`
- [x] 5.6 Add `relateFkId: String` field to `Field` type (nullable; set for RELATION-format fields)
- [x] 5.7 Add `belongsToFkId: String` field to `Field` type (nullable; set for FK column fields)
- [x] 5.8 Add `CreateLogicalForeignKeyInput` with fields: `modelId: ID!`, `refModelId: ID!`, `sourceFields: [String!]!`, `targetFields: [String!]!`
- [x] 5.9 Add `CreateLogicalForeignKeyPayload` and `DeleteLogicalForeignKeyPayload` with error unions
- [x] 5.10 Add mutations: `createLogicalForeignKey(projectSlug: String!, input: CreateLogicalForeignKeyInput!): CreateLogicalForeignKeyPayload!` and `deleteLogicalForeignKey(projectSlug: String!, pairId: String!): DeleteLogicalForeignKeyPayload!`
- [x] 5.11 Add query: `logicalForeignKeys(projectSlug: String!, modelId: ID!): [LogicalForeignKey!]!`
- [x] 5.12 Update `AddFieldInput`: add optional `relateFkId: String` field; remove `relationConfig`
- [x] 5.13 Run `task generate-gql` to regenerate GraphQL code

## 6. GraphQL Resolvers

- [x] 6.1 Write unit tests for `createLogicalForeignKey` mutation resolver
- [x] 6.2 Implement `createLogicalForeignKey` mutation resolver wiring to `CreateLogicalForeignKeyUseCase`
- [x] 6.3 Write unit tests for `deleteLogicalForeignKey` mutation resolver
- [x] 6.4 Implement `deleteLogicalForeignKey` mutation resolver wiring to `DeleteLogicalForeignKeyUseCase`
- [x] 6.5 Implement `logicalForeignKeys` query resolver
- [x] 6.6 Update `addFields` mutation resolver to pass `relate_fk_id` from input to use case
- [x] 6.7 Remove all `ModelRelation`-related resolvers and error adapters
- [x] 6.8 Add FK error adapter in `internal/interfaces/graphql/adapter/` for FK-specific errors (FKColumnsNotFound, FKPairHasRelateFields, FKNotFound)

## 7. Cleanup & Verification

- [x] 7.1 Remove all remaining `ModelRelation` references (grep for `ModelRelation`, `model_relation`, `RelationType`, `RelationConfig`, `target_relation_name`)
- [x] 7.2 Run `task lint-fix` to fix auto-fixable lint issues
- [x] 7.3 Run `task check-all` (fmt + lint + vet + test) and fix any issues
- [ ] 7.4 Run `task auto-test` integration tests and verify FK operations end-to-end
- [x] 7.5 Run `task build` to confirm clean build