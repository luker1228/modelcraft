# Design: Merge model_database_sync_job into model_sync_job

**Date:** 2026-06-16  
**Status:** Approved

## Background

There are currently two async job tables that do the same thing — sync database tables into models:

- `model_sync_job` — keyed by `database_name` (string), supports `table_names` filter
- `model_database_sync_job` — keyed by `database_id` (FK to `model_database`), always full sync

The goal is to consolidate into a single table and service, with `database_id` as the canonical key and `database_name` stored as a redundant display field.

## Design

### 1. DB: `model_sync_job` — new fields

Add to the existing `model_sync_job` table:

```sql
`batch_id`    VARCHAR(36) NOT NULL COMMENT '批次 ID，同批次多条 job 共享',
`database_id` VARCHAR(36) NOT NULL COMMENT '关联 model_database.id',
-- `database_name` already exists — retained, populated from model_database at job creation time
```

New index:
```sql
INDEX `idx_model_sync_job_batch` (`batch_id`)
```

`model_database_sync_job` table is **retained but frozen** — no new writes. Data expires naturally; can be dropped in a future release.

### 2. GraphQL Schema

#### Input

```graphql
input ModelSyncTargetInput {
  databaseId: ID!
  tableNames: [String!]   # empty = full sync
}
```

#### Mutation

```graphql
type Mutation {
  # New
  startModelSync(targets: [ModelSyncTargetInput!]!): StartModelSyncPayload!

  # Deprecated
  startModelDatabaseSync(databaseID: ID!): StartModelDatabaseSyncPayload! @deprecated(reason: "Use startModelSync")
  syncModelsFromDB(input: SyncModelsFromDBInput!): SyncModelsFromDBPayload! @deprecated(reason: "Use startModelSync")
}
```

#### Payload

```graphql
type StartModelSyncPayload {
  batchId: ID!
  jobs: [ModelSyncJobRef!]!
}

type ModelSyncJobRef {
  databaseId: ID!
  jobId: ID!
}
```

#### Query

```graphql
type Query {
  # jobIds and batchId are mutually exclusive; batchId takes priority if both provided
  modelSyncJobs(jobIds: [ID!], batchId: ID): [ModelSyncJob!]!

  # Deprecated
  modelSyncJob(jobId: ID!): ModelSyncJob @deprecated(reason: "Use modelSyncJobs")
  modelDatabaseSyncJob(jobId: ID!): ModelDatabaseSyncJob @deprecated(reason: "Use modelSyncJobs")
}
```

#### Type

```graphql
type ModelSyncJob {
  id: ID!
  batchId: ID!
  databaseId: ID!
  databaseName: String!
  tableNames: [String!]!
  status: ModelSyncJobStatus!
  totalTables: Int!
  processedTables: Int!
  createdModels: Int!
  syncedModels: Int!
  failedCount: Int!
  failedTables: [ModelSyncFailedTable!]!
  startedAt: Time
  finishedAt: Time
  createdAt: Time!
  updatedAt: Time!
}

type ModelSyncFailedTable {
  tableName: String!
  message: String!
}

enum ModelSyncJobStatus {
  PENDING
  RUNNING
  SUCCEEDED
  PARTIAL_SUCCESS
  FAILED
}
```

### 3. Domain

#### `ModelSyncJob` struct — new fields

```go
BatchID    string
DatabaseID string
// DatabaseName already exists
```

#### `ModelSyncJobRepository` — new methods

```go
GetByIDs(ctx context.Context, orgName, projectSlug string, jobIDs []string) ([]*ModelSyncJob, error)
GetByBatchID(ctx context.Context, orgName, projectSlug, batchID string) ([]*ModelSyncJob, error)
FailStalePendingJobs(ctx context.Context, staleBefore time.Time) error
```

### 4. App Service

`SyncModelsAppService` absorbs all capabilities from `ModelDatabaseSyncAppService`:

**New `StartSync` signature:**
```go
type SyncTarget struct {
    DatabaseID string
    TableNames []string // nil or empty = full sync
}

func (s *SyncModelsAppService) StartSync(
    ctx context.Context,
    targets []SyncTarget,
) (batchID string, jobs []*domaindb.ModelSyncJob, err error)
```

Behavior:
1. Generate one `batchId` (UUID) for the entire call
2. For each target: validate `database_id` exists, check for active job, look up `database_name` from `model_database`, create a `ModelSyncJob` with `batch_id`, `database_id`, `database_name`
3. Fire one background goroutine per job
4. Return `batchId` + `[]ModelSyncJob`

**Migrated from `ModelDatabaseSyncAppService`:**
- `RecoverStaleJobs` — calls `FailStalePendingJobs` on `model_sync_job`
- `EnsureImportGroup` / `MoveModelToGroup` — group assignment logic per table

`ModelDatabaseSyncAppService` is **retained but frozen** — existing deprecated resolvers continue pointing to it; no new code paths added.

### 5. Frontend

- Add `startModelSync` mutation + `modelSyncJobs` query to `api-client/model/graphql-docs.ts`
- Replace call sites of `startModelDatabaseSync` / `modelDatabaseSyncJob` with new interfaces
- Replace `ModelDatabaseSyncJob` type in `use-model-databases.ts` with unified `ModelSyncJob`
- Update MSW mock handlers

## What is NOT changing

- `model_database_sync_job` table: retained, no new writes
- `ModelDatabaseSyncAppService`: retained, frozen
- Deprecated GraphQL fields: kept with `@deprecated` annotations, resolvers still functional
- `model_sync_job` existing rows: no migration needed, `batch_id` and `database_id` will be non-null for new rows only (migration script sets defaults for existing rows)
