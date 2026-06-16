# Merge model_database_sync_job into model_sync_job — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate `model_database_sync_job` and `model_sync_job` into a single table and service, with a new `startModelSync` batch mutation and `modelSyncJobs` query.

**Architecture:** Add `batch_id`, `database_id`, `database_name` (already exists) to `model_sync_job`. Extend `SyncModelsAppService` to absorb `ModelDatabaseSyncAppService` capabilities (group assignment, stale job recovery, database-id lookup). New GraphQL schema adds `startModelSync` / `modelSyncJobs`; old fields kept with `@deprecated`. Frontend adds new hooks and migrates call sites.

**Tech Stack:** Go (sqlc, gqlgen), MySQL, GraphQL, Next.js/TypeScript, Apollo Client

---

## File Map

### Backend — DB / SQL
| Action | File |
|--------|------|
| Modify | `db/schema/mysql/18_model_sync_job.sql` |
| Create | `db/schema/mysql/19_model_sync_job_v2.sql` (migration ALTER) |
| Modify | `db/queries/model_sync_job.sql` |
| Regenerate | `internal/infrastructure/dbgen/model_sync_job.sql.go` (via `just generate-sqlc`) |
| Regenerate | `internal/infrastructure/dbgen/models.go` (via `just generate-sqlc`) |

### Backend — Domain
| Action | File |
|--------|------|
| Modify | `internal/domain/modeldatabase/model_sync_job.go` |
| Modify | `internal/domain/modeldatabase/repository.go` |

### Backend — Infrastructure
| Action | File |
|--------|------|
| Modify | `internal/infrastructure/repository/sql_model_sync_job_repository.go` |

### Backend — App Service
| Action | File |
|--------|------|
| Modify | `internal/app/modeldatabase/sync_models_app.go` |
| Modify | `internal/app/modeldatabase/sync_models_app_test.go` |

### Backend — GraphQL Schema
| Action | File |
|--------|------|
| Modify | `api/graph/project/schema/model.graphql` |
| Modify | `api/graph/project/schema/database.graphql` |
| Regenerate | `internal/interfaces/graphql/project/generated/` (via `just generate-gql`) |

### Backend — GraphQL Resolvers
| Action | File |
|--------|------|
| Modify | `internal/interfaces/graphql/project/model.resolvers.go` |
| Modify | `internal/interfaces/graphql/project/resolver.go` |
| Modify | `internal/interfaces/http/routes.go` |

### Frontend
| Action | File |
|--------|------|
| Modify | `src/api-client/model/graphql-docs.ts` |
| Modify | `src/api-client/project/model-database-graphql-docs.ts` |
| Create | `src/web/hooks/model/use-model-sync.ts` |
| Modify | `src/web/hooks/model-database/use-model-databases.ts` |
| Modify | `src/app/org/[orgName]/project/[projectSlug]/databases/page.tsx` |
| Modify | `src/app/org/[orgName]/project/[projectSlug]/databases/_components/batch-register-sync.ts` |
| Sync | `src/contract/` (via `front-contract-pull` skill) |
| Regenerate | `src/generated/graphql.ts` (via codegen) |

---

## Task 1: DB Migration — add batch_id and database_id to model_sync_job

**Files:**
- Create: `modelcraft-backend/db/schema/mysql/19_model_sync_job_v2.sql`
- Modify: `modelcraft-backend/db/queries/model_sync_job.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- modelcraft-backend/db/schema/mysql/19_model_sync_job_v2.sql
-- Add batch_id and database_id to model_sync_job

ALTER TABLE `model_sync_job`
  ADD COLUMN `batch_id`    VARCHAR(36) NOT NULL DEFAULT '' COMMENT '批次 ID，同批次多条 job 共享' AFTER `id`,
  ADD COLUMN `database_id` VARCHAR(36) NOT NULL DEFAULT '' COMMENT '关联 model_database.id' AFTER `batch_id`,
  ADD INDEX `idx_model_sync_job_batch` (`batch_id`),
  ADD INDEX `idx_model_sync_job_database_id` (`database_id`);
```

> DEFAULT '' allows existing rows to not break NOT NULL. New code always populates these.

- [ ] **Step 2: Apply migration to local dev DB**

```bash
cd modelcraft-backend
just db-migrate   # or the equivalent: check `just --list` for the atlas/migrate command
```

Expected: migration applies without error.

- [ ] **Step 3: Update `model_sync_job.sql` queries to include new columns**

Replace the contents of `modelcraft-backend/db/queries/model_sync_job.sql`:

```sql
-- name: CreateModelSyncJob :exec
INSERT INTO model_sync_job (
  id, batch_id, database_id, org_name, project_slug, database_name, table_names,
  status, total_tables, processed_tables, created_models, synced_models,
  failed_count, failed_tables, started_at, finished_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelSyncJobByID :one
SELECT * FROM model_sync_job
WHERE id = ? AND org_name = ? AND project_slug = ?
LIMIT 1;

-- name: GetModelSyncJobsByIDs :many
SELECT * FROM model_sync_job
WHERE org_name = ? AND project_slug = ? AND id IN (/*SLICE:ids*/?)
ORDER BY created_at DESC;

-- name: GetModelSyncJobsByBatchID :many
SELECT * FROM model_sync_job
WHERE org_name = ? AND project_slug = ? AND batch_id = ?
ORDER BY created_at DESC;

-- name: GetActiveModelSyncJobByDatabase :one
SELECT * FROM model_sync_job
WHERE org_name = ?
  AND project_slug = ?
  AND database_id = ?
  AND status IN ('pending', 'running')
  AND updated_at > ?
ORDER BY created_at DESC
LIMIT 1;

-- name: FailStaleModelSyncJobs :exec
UPDATE model_sync_job
SET status = 'failed',
    finished_at = NOW(3),
    updated_at = NOW(3)
WHERE status IN ('pending', 'running')
  AND updated_at <= ?;

-- name: UpdateModelSyncJob :exec
UPDATE model_sync_job
SET status = ?,
    total_tables = ?,
    processed_tables = ?,
    created_models = ?,
    synced_models = ?,
    failed_count = ?,
    failed_tables = ?,
    started_at = ?,
    finished_at = ?,
    updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND project_slug = ?;
```

> Note: `GetActiveModelSyncJobByDatabase` now uses `database_id` instead of `database_name`.
> The old `GetActiveModelSyncJobByDatabase` using `database_name` is replaced.

- [ ] **Step 4: Regenerate sqlc**

```bash
cd modelcraft-backend
just generate-sqlc
```

Expected: `internal/infrastructure/dbgen/model_sync_job.sql.go` and `models.go` updated — `ModelSyncJob` struct now has `BatchID` and `DatabaseID` fields. No compile errors.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/db/schema/mysql/19_model_sync_job_v2.sql \
        modelcraft-backend/db/queries/model_sync_job.sql \
        modelcraft-backend/internal/infrastructure/dbgen/
git commit -m "feat(db): add batch_id and database_id to model_sync_job"
```

---

## Task 2: Domain — extend ModelSyncJob and Repository interface

**Files:**
- Modify: `modelcraft-backend/internal/domain/modeldatabase/model_sync_job.go`
- Modify: `modelcraft-backend/internal/domain/modeldatabase/repository.go`

- [ ] **Step 1: Add BatchID and DatabaseID to ModelSyncJob domain struct**

In `model_sync_job.go`, replace the `ModelSyncJob` struct:

```go
type ModelSyncJob struct {
	ID              string
	BatchID         string
	DatabaseID      string
	OrgName         string
	ProjectSlug     string
	DatabaseName    string
	TableNames      []string
	Status          ModelSyncJobStatus
	TotalTables     int
	ProcessedTables int
	CreatedModels   int
	SyncedModels    int
	FailedCount     int
	FailedTables    []ModelSyncFailedTable
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
```

- [ ] **Step 2: Extend ModelSyncJobRepository interface**

In `repository.go`, replace the `ModelSyncJobRepository` interface:

```go
type ModelSyncJobRepository interface {
	Create(ctx context.Context, job *ModelSyncJob) error
	GetByID(ctx context.Context, orgName, projectSlug, jobID string) (*ModelSyncJob, error)
	GetByIDs(ctx context.Context, orgName, projectSlug string, jobIDs []string) ([]*ModelSyncJob, error)
	GetByBatchID(ctx context.Context, orgName, projectSlug, batchID string) ([]*ModelSyncJob, error)
	// GetActiveByDatabase returns the active (pending/running) job for the given database_id,
	// only if updated after staleBefore (to exclude zombie jobs).
	GetActiveByDatabase(ctx context.Context, orgName, projectSlug, databaseID string, staleBefore time.Time) (*ModelSyncJob, error)
	// FailStalePendingJobs marks all pending/running jobs with updated_at <= staleBefore as failed.
	FailStalePendingJobs(ctx context.Context, staleBefore time.Time) error
	Update(ctx context.Context, job *ModelSyncJob) error
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd modelcraft-backend
go build ./internal/domain/...
```

Expected: no errors (repository implementations will break — fixed in Task 3).

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/internal/domain/modeldatabase/
git commit -m "feat(domain): add BatchID/DatabaseID to ModelSyncJob, extend repository interface"
```

---

## Task 3: Infrastructure — implement new repository methods

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_model_sync_job_repository.go`

- [ ] **Step 1: Update Create, GetByID, GetActiveByDatabase, Update, modelSyncJobToDomain to use new fields**

Replace the entire file content:

```go
package repository

import (
	"context"
	"encoding/json"
	"time"

	domaindb "modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
)

type SqlModelSyncJobRepository struct {
	q dbgen.Querier
}

func NewSqlModelSyncJobRepository(q dbgen.Querier) domaindb.ModelSyncJobRepository {
	return &SqlModelSyncJobRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

func (r *SqlModelSyncJobRepository) Create(ctx context.Context, job *domaindb.ModelSyncJob) error {
	tableNames, err := marshalModelSyncTableNames(job.TableNames)
	if err != nil {
		return err
	}
	failedTables, err := marshalModelSyncFailedTables(job.FailedTables)
	if err != nil {
		return err
	}
	return r.q.CreateModelSyncJob(ctx, dbgen.CreateModelSyncJobParams{
		ID:              job.ID,
		BatchID:         job.BatchID,
		DatabaseID:      job.DatabaseID,
		OrgName:         job.OrgName,
		ProjectSlug:     job.ProjectSlug,
		DatabaseName:    job.DatabaseName,
		TableNames:      tableNames,
		Status:          dbgen.ModelSyncJobStatus(job.Status),
		TotalTables:     int32(job.TotalTables),
		ProcessedTables: int32(job.ProcessedTables),
		CreatedModels:   int32(job.CreatedModels),
		SyncedModels:    int32(job.SyncedModels),
		FailedCount:     int32(job.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       timeToNull(job.StartedAt),
		FinishedAt:      timeToNull(job.FinishedAt),
	})
}

func (r *SqlModelSyncJobRepository) GetByID(
	ctx context.Context, orgName, projectSlug, jobID string,
) (*domaindb.ModelSyncJob, error) {
	row, err := r.q.GetModelSyncJobByID(ctx, dbgen.GetModelSyncJobByIDParams{
		ID:          jobID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil
		}
		return nil, err
	}
	return modelSyncJobToDomain(row)
}

func (r *SqlModelSyncJobRepository) GetByIDs(
	ctx context.Context, orgName, projectSlug string, jobIDs []string,
) ([]*domaindb.ModelSyncJob, error) {
	if len(jobIDs) == 0 {
		return nil, nil
	}
	rows, err := r.q.GetModelSyncJobsByIDs(ctx, dbgen.GetModelSyncJobsByIDsParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Ids:         jobIDs,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*domaindb.ModelSyncJob, 0, len(rows))
	for _, row := range rows {
		job, err := modelSyncJobToDomain(row)
		if err != nil {
			return nil, err
		}
		result = append(result, job)
	}
	return result, nil
}

func (r *SqlModelSyncJobRepository) GetByBatchID(
	ctx context.Context, orgName, projectSlug, batchID string,
) ([]*domaindb.ModelSyncJob, error) {
	rows, err := r.q.GetModelSyncJobsByBatchID(ctx, dbgen.GetModelSyncJobsByBatchIDParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		BatchID:     batchID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*domaindb.ModelSyncJob, 0, len(rows))
	for _, row := range rows {
		job, err := modelSyncJobToDomain(row)
		if err != nil {
			return nil, err
		}
		result = append(result, job)
	}
	return result, nil
}

func (r *SqlModelSyncJobRepository) GetActiveByDatabase(
	ctx context.Context, orgName, projectSlug, databaseID string, staleBefore time.Time,
) (*domaindb.ModelSyncJob, error) {
	row, err := r.q.GetActiveModelSyncJobByDatabase(ctx, dbgen.GetActiveModelSyncJobByDatabaseParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		DatabaseID:  databaseID,
		UpdatedAt:   staleBefore,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil
		}
		return nil, err
	}
	return modelSyncJobToDomain(row)
}

func (r *SqlModelSyncJobRepository) FailStalePendingJobs(ctx context.Context, staleBefore time.Time) error {
	return r.q.FailStaleModelSyncJobs(ctx, staleBefore)
}

func (r *SqlModelSyncJobRepository) Update(ctx context.Context, job *domaindb.ModelSyncJob) error {
	failedTables, err := marshalModelSyncFailedTables(job.FailedTables)
	if err != nil {
		return err
	}
	return r.q.UpdateModelSyncJob(ctx, dbgen.UpdateModelSyncJobParams{
		Status:          dbgen.ModelSyncJobStatus(job.Status),
		TotalTables:     int32(job.TotalTables),
		ProcessedTables: int32(job.ProcessedTables),
		CreatedModels:   int32(job.CreatedModels),
		SyncedModels:    int32(job.SyncedModels),
		FailedCount:     int32(job.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       timeToNull(job.StartedAt),
		FinishedAt:      timeToNull(job.FinishedAt),
		ID:              job.ID,
		OrgName:         job.OrgName,
		ProjectSlug:     job.ProjectSlug,
	})
}

func modelSyncJobToDomain(row dbgen.ModelSyncJob) (*domaindb.ModelSyncJob, error) {
	tableNames, err := unmarshalModelSyncTableNames(row.TableNames)
	if err != nil {
		return nil, err
	}
	failedTables, err := unmarshalModelSyncFailedTables(row.FailedTables)
	if err != nil {
		return nil, err
	}
	return &domaindb.ModelSyncJob{
		ID:              row.ID,
		BatchID:         row.BatchID,
		DatabaseID:      row.DatabaseID,
		OrgName:         row.OrgName,
		ProjectSlug:     row.ProjectSlug,
		DatabaseName:    row.DatabaseName,
		TableNames:      tableNames,
		Status:          domaindb.ModelSyncJobStatus(row.Status),
		TotalTables:     int(row.TotalTables),
		ProcessedTables: int(row.ProcessedTables),
		CreatedModels:   int(row.CreatedModels),
		SyncedModels:    int(row.SyncedModels),
		FailedCount:     int(row.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       nullToTimePtr(row.StartedAt),
		FinishedAt:      nullToTimePtr(row.FinishedAt),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func marshalModelSyncTableNames(names []string) (json.RawMessage, error) {
	if names == nil {
		names = []string{}
	}
	data, err := json.Marshal(names)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func unmarshalModelSyncTableNames(data json.RawMessage) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, err
	}
	return names, nil
}

func marshalModelSyncFailedTables(items []domaindb.ModelSyncFailedTable) (json.RawMessage, error) {
	if items == nil {
		items = []domaindb.ModelSyncFailedTable{}
	}
	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func unmarshalModelSyncFailedTables(data json.RawMessage) ([]domaindb.ModelSyncFailedTable, error) {
	if len(data) == 0 {
		return []domaindb.ModelSyncFailedTable{}, nil
	}
	var items []domaindb.ModelSyncFailedTable
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

var _ domaindb.ModelSyncJobRepository = (*SqlModelSyncJobRepository)(nil)
```

- [ ] **Step 2: Compile check**

```bash
cd modelcraft-backend
go build ./internal/infrastructure/repository/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add modelcraft-backend/internal/infrastructure/repository/sql_model_sync_job_repository.go
git commit -m "feat(infra): implement GetByIDs, GetByBatchID, FailStalePendingJobs in SqlModelSyncJobRepository"
```

---

## Task 4: App Service — merge ModelDatabaseSyncAppService into SyncModelsAppService

**Files:**
- Modify: `modelcraft-backend/internal/app/modeldatabase/sync_models_app.go`
- Modify: `modelcraft-backend/internal/app/modeldatabase/sync_models_app_test.go`

- [ ] **Step 1: Write failing tests for the new StartSync signature**

Add to `sync_models_app_test.go`:

```go
// TestStartModelSync_BatchCreatesMultipleJobs tests that StartModelSync creates
// one job per target and returns a shared batchId.
func TestStartModelSync_BatchCreatesMultipleJobs(t *testing.T) {
	repo := newFakeSyncModelsJobRepo()
	dbRepo := newFakeModelDatabaseRepo(map[string]*domaindb.ModelDatabase{
		"db-1": {ID: "db-1", Name: "mydb1", OrgName: "org1", ProjectSlug: "proj1"},
		"db-2": {ID: "db-2", Name: "mydb2", OrgName: "org1", ProjectSlug: "proj1"},
	})
	svc := newSyncModelsService(repo, dbRepo)
	ctx := ctxutils.WithOrgName(ctxutils.WithProjectSlug(context.Background(), "proj1"), "org1")

	batchID, jobs, err := svc.StartModelSync(ctx, []SyncTarget{
		{DatabaseID: "db-1", TableNames: nil},
		{DatabaseID: "db-2", TableNames: []string{"orders"}},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, batchID)
	assert.Len(t, jobs, 2)
	assert.Equal(t, batchID, jobs[0].BatchID)
	assert.Equal(t, batchID, jobs[1].BatchID)
	assert.Equal(t, "db-1", jobs[0].DatabaseID)
	assert.Equal(t, "mydb1", jobs[0].DatabaseName)
	assert.Equal(t, "db-2", jobs[1].DatabaseID)
	assert.Equal(t, []string{"orders"}, jobs[1].TableNames)
}

// TestStartModelSync_RejectsActiveJob tests that a target with an existing active job is rejected.
func TestStartModelSync_RejectsActiveJob(t *testing.T) {
	repo := newFakeSyncModelsJobRepo()
	dbRepo := newFakeModelDatabaseRepo(map[string]*domaindb.ModelDatabase{
		"db-1": {ID: "db-1", Name: "mydb1", OrgName: "org1", ProjectSlug: "proj1"},
	})
	// seed an active job for db-1
	repo.activeByDatabaseID["db-1"] = &domaindb.ModelSyncJob{
		ID:         "existing-job",
		DatabaseID: "db-1",
		Status:     domaindb.ModelSyncJobStatusRunning,
	}
	svc := newSyncModelsService(repo, dbRepo)
	ctx := ctxutils.WithOrgName(ctxutils.WithProjectSlug(context.Background(), "proj1"), "org1")

	_, _, err := svc.StartModelSync(ctx, []SyncTarget{{DatabaseID: "db-1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd modelcraft-backend
just test-unit-pkg ./internal/app/modeldatabase/...
```

Expected: compile error or test failure — `StartModelSync`, `SyncTarget`, `fakeModelDatabaseRepo` not yet defined.

- [ ] **Step 3: Implement new SyncTarget type and StartModelSync on SyncModelsAppService**

In `sync_models_app.go`, add `SyncTarget` and the new interface dependency, then add `StartModelSync`:

```go
// SyncTarget is one database to sync, with optional table filter.
type SyncTarget struct {
	DatabaseID string
	TableNames []string // nil or empty = full sync
}

// syncModelsDBRepo is the subset of ModelDatabaseRepository used here.
type syncModelsDBRepo interface {
	GetByID(ctx context.Context, orgName, projectSlug, id string) (*domaindb.ModelDatabase, error)
}
```

Add `dbRepo syncModelsDBRepo` to `SyncModelsAppService` struct and `SyncModelsAppServiceDeps`:

```go
type SyncModelsAppServiceDeps struct {
	SyncJobRepo     domaindb.ModelSyncJobRepository
	DBRepo          syncModelsDBRepo   // NEW
	ReverseEngineer syncModelsReverseEngineer
	ModelRepo       syncModelsModelRepo
	FieldSyncer     syncModelsFieldSyncer
	GroupService    syncModelsGroupService // NEW — see below
	Runner          modelDatabaseSyncRunner
	Now             func() time.Time
}
```

Add `syncModelsGroupService` interface:

```go
// syncModelsGroupService is the subset of ModelGroupAppService used here.
type syncModelsGroupService interface {
	EnsureImportGroup(ctx context.Context, orgName, projectSlug string) (*domainmodel.ModelGroup, error)
	MoveModelToGroup(ctx context.Context, modelID string, groupID *string) error
}
```

Add `StartModelSync` method:

```go
// StartModelSync starts one sync job per target. All jobs share the same batchId.
// Returns an error if any target already has an active job.
func (s *SyncModelsAppService) StartModelSync(
	ctx context.Context,
	targets []SyncTarget,
) (batchID string, jobs []*domaindb.ModelSyncJob, err error) {
	if len(targets) == 0 {
		return "", nil, bizerrors.NewError(bizerrors.ParamInvalid, "targets must not be empty")
	}
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return "", nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return "", nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}

	staleBefore := s.now().Add(-defaultStalePeriod)
	batchID = uuid.NewString()
	now := s.now()

	for _, target := range targets {
		db, err := s.dbRepo.GetByID(ctx, orgName, projectSlug, target.DatabaseID)
		if err != nil {
			return "", nil, err
		}

		active, err := s.syncJobRepo.GetActiveByDatabase(ctx, orgName, projectSlug, target.DatabaseID, staleBefore)
		if err != nil {
			return "", nil, err
		}
		if active != nil {
			return "", nil, bizerrors.NewError(bizerrors.Conflict,
				"sync job already running for database "+target.DatabaseID)
		}

		tableNames := target.TableNames
		if tableNames == nil {
			tableNames = []string{}
		}
		job := &domaindb.ModelSyncJob{
			ID:           uuid.NewString(),
			BatchID:      batchID,
			DatabaseID:   target.DatabaseID,
			OrgName:      orgName,
			ProjectSlug:  projectSlug,
			DatabaseName: db.Name,
			TableNames:   tableNames,
			Status:       domaindb.ModelSyncJobStatusPending,
			FailedTables: []domaindb.ModelSyncFailedTable{},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.syncJobRepo.Create(ctx, job); err != nil {
			return "", nil, err
		}
		jobs = append(jobs, job)
	}

	for _, job := range jobs {
		job := job
		s.runner.Go(ctx, func(runCtx context.Context) {
			if err := s.RunSyncJob(runCtx, job.ID); err != nil {
				logfacade.GetLogger(runCtx).Error(
					runCtx, "model sync job failed",
					logfacade.String("job_id", job.ID),
					logfacade.Err(err),
				)
			}
		})
	}
	return batchID, jobs, nil
}
```

Also add `RecoverStaleJobs`:

```go
// RecoverStaleJobs marks all stale pending/running jobs as failed.
// Call at service startup.
func (s *SyncModelsAppService) RecoverStaleJobs(ctx context.Context) error {
	staleBefore := s.now().Add(-defaultStalePeriod)
	return s.syncJobRepo.FailStalePendingJobs(ctx, staleBefore)
}
```

Also add `GetJobs` for batch/multi-id lookup:

```go
// GetJobs returns jobs by jobIDs or batchID (batchID takes priority).
func (s *SyncModelsAppService) GetJobs(
	ctx context.Context,
	jobIDs []string,
	batchID string,
) ([]*domaindb.ModelSyncJob, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	if batchID != "" {
		return s.syncJobRepo.GetByBatchID(ctx, orgName, projectSlug, batchID)
	}
	if len(jobIDs) == 0 {
		return nil, nil
	}
	return s.syncJobRepo.GetByIDs(ctx, orgName, projectSlug, jobIDs)
}
```

Update `RunSyncJob` to use `db.Name` via `dbRepo.GetByID` instead of `job.DatabaseName` directly — and use `groupService` for model group assignment:

```go
func (s *SyncModelsAppService) RunSyncJob(ctx context.Context, jobID string) error {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	logger := logfacade.GetLogger(ctx)

	job, err := s.syncJobRepo.GetByID(ctx, orgName, projectSlug, jobID)
	if err != nil {
		return err
	}

	now := s.now()
	job.Status = domaindb.ModelSyncJobStatusRunning
	job.StartedAt = &now
	job.UpdatedAt = now
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// Determine tables to process
	var tableNames []string
	if len(job.TableNames) > 0 {
		tableNames = job.TableNames
	} else {
		tableResult, err := s.reverseEngineer.ListTables(ctx, orgName, projectSlug, job.DatabaseName, false, 0, 0)
		if err != nil {
			logger.Error(ctx, "model sync job: ListTables failed",
				logfacade.String("job_id", jobID), logfacade.Err(err))
			return s.failJob(ctx, job, err)
		}
		tableNames = tableResult.Tables
	}

	job.TotalTables = len(tableNames)
	job.UpdatedAt = s.now()
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// Ensure import group (only if groupService is available)
	var group *domainmodel.ModelGroup
	if s.groupService != nil {
		group, err = s.groupService.EnsureImportGroup(ctx, orgName, projectSlug)
		if err != nil {
			logger.Error(ctx, "model sync job: EnsureImportGroup failed",
				logfacade.String("job_id", jobID), logfacade.Err(err))
			return s.failJob(ctx, job, err)
		}
	}

	for _, tableName := range tableNames {
		if err := s.processTable(ctx, job, tableName, group); err != nil {
			logger.Error(ctx, "model sync job: processTable fatal",
				logfacade.String("job_id", jobID), logfacade.String("table", tableName), logfacade.Err(err))
			return err
		}
	}

	finishedAt := s.now()
	job.FinishedAt = &finishedAt
	job.UpdatedAt = finishedAt
	switch {
	case job.FailedCount == 0:
		job.Status = domaindb.ModelSyncJobStatusSucceeded
	case job.CreatedModels > 0 || job.SyncedModels > 0:
		job.Status = domaindb.ModelSyncJobStatusPartialSuccess
	default:
		job.Status = domaindb.ModelSyncJobStatusFailed
	}
	return s.syncJobRepo.Update(ctx, job)
}
```

Update `processTable` signature to accept `group`:

```go
func (s *SyncModelsAppService) processTable(
	ctx context.Context,
	job *domaindb.ModelSyncJob,
	tableName string,
	group *domainmodel.ModelGroup,
) error {
	orgName := job.OrgName
	projectSlug := job.ProjectSlug
	databaseName := job.DatabaseName

	tableDef, err := s.reverseEngineer.GetTableDefinition(ctx, orgName, projectSlug, databaseName, tableName)
	if err != nil {
		return s.recordTableFailure(ctx, job, tableName, err)
	}

	modelName := appmodeldesign.NormalizeModelName(tableName)
	existingModel, err := s.modelRepo.GetByName(ctx, orgName, databaseName, modelName, projectSlug)
	if err != nil && !shared.IsNotFoundError(err) {
		return s.recordTableFailure(ctx, job, tableName, err)
	}

	if existingModel != nil {
		if err := s.fieldSyncer.SyncFieldsFromDB(ctx, existingModel.ID, tableDef.Fields); err != nil {
			return s.recordTableFailure(ctx, job, tableName, err)
		}
		job.SyncedModels++
	} else {
		importResult, importErr := s.reverseEngineer.ImportModel(ctx, appmodeldesign.ImportModelCommand{
			OrgName:      orgName,
			ProjectSlug:  projectSlug,
			DatabaseName: databaseName,
			TableName:    tableName,
		})
		if importErr != nil {
			return s.recordTableFailure(ctx, job, tableName, importErr)
		}
		if group != nil {
			if err := s.groupService.MoveModelToGroup(ctx, importResult.ModelID, &group.ID); err != nil {
				return s.recordTableFailure(ctx, job, tableName, err)
			}
		}
		job.CreatedModels++
	}

	job.ProcessedTables++
	job.UpdatedAt = s.now()
	return s.syncJobRepo.Update(ctx, job)
}
```

- [ ] **Step 4: Add fake helpers and fix test compilation**

In `sync_models_app_test.go`, add `fakeModelDatabaseRepo` and update `newSyncModelsService`:

```go
type fakeModelDatabaseRepo struct {
	dbs map[string]*domaindb.ModelDatabase // key: id
}

func newFakeModelDatabaseRepo(dbs map[string]*domaindb.ModelDatabase) *fakeModelDatabaseRepo {
	return &fakeModelDatabaseRepo{dbs: dbs}
}

func (f *fakeModelDatabaseRepo) GetByID(_ context.Context, _, _, id string) (*domaindb.ModelDatabase, error) {
	db := f.dbs[id]
	if db == nil {
		return nil, shared.NewNotFoundError("model database not found")
	}
	cloned := *db
	return &cloned, nil
}
```

Update `fakeSyncModelsJobRepo` to use `activeByDatabaseID` map:

```go
type fakeSyncModelsJobRepo struct {
	jobs              map[string]*domaindb.ModelSyncJob
	activeByDatabaseID map[string]*domaindb.ModelSyncJob // key: databaseID
	snapshots         []*domaindb.ModelSyncJob
	staleFailed       bool
}

func newFakeSyncModelsJobRepo() *fakeSyncModelsJobRepo {
	return &fakeSyncModelsJobRepo{
		jobs:               make(map[string]*domaindb.ModelSyncJob),
		activeByDatabaseID: make(map[string]*domaindb.ModelSyncJob),
	}
}

func (f *fakeSyncModelsJobRepo) Create(_ context.Context, job *domaindb.ModelSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	if job.Status == domaindb.ModelSyncJobStatusPending || job.Status == domaindb.ModelSyncJobStatusRunning {
		f.activeByDatabaseID[job.DatabaseID] = &cloned
	}
	return nil
}

func (f *fakeSyncModelsJobRepo) GetByID(
	_ context.Context, orgName, projectSlug, jobID string,
) (*domaindb.ModelSyncJob, error) {
	job := f.jobs[jobID]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, shared.NewNotFoundError("sync job not found")
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncModelsJobRepo) GetByIDs(
	_ context.Context, orgName, projectSlug string, jobIDs []string,
) ([]*domaindb.ModelSyncJob, error) {
	var result []*domaindb.ModelSyncJob
	for _, id := range jobIDs {
		job := f.jobs[id]
		if job != nil && job.OrgName == orgName && job.ProjectSlug == projectSlug {
			cloned := *job
			result = append(result, &cloned)
		}
	}
	return result, nil
}

func (f *fakeSyncModelsJobRepo) GetByBatchID(
	_ context.Context, orgName, projectSlug, batchID string,
) ([]*domaindb.ModelSyncJob, error) {
	var result []*domaindb.ModelSyncJob
	for _, job := range f.jobs {
		if job.OrgName == orgName && job.ProjectSlug == projectSlug && job.BatchID == batchID {
			cloned := *job
			result = append(result, &cloned)
		}
	}
	return result, nil
}

func (f *fakeSyncModelsJobRepo) GetActiveByDatabase(
	_ context.Context, orgName, projectSlug, databaseID string, _ time.Time,
) (*domaindb.ModelSyncJob, error) {
	job := f.activeByDatabaseID[databaseID]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, nil
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncModelsJobRepo) FailStalePendingJobs(_ context.Context, _ time.Time) error {
	f.staleFailed = true
	return nil
}

func (f *fakeSyncModelsJobRepo) Update(_ context.Context, job *domaindb.ModelSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	f.snapshots = append(f.snapshots, &cloned)
	if job.Status == domaindb.ModelSyncJobStatusPending || job.Status == domaindb.ModelSyncJobStatusRunning {
		f.activeByDatabaseID[job.DatabaseID] = &cloned
	} else {
		delete(f.activeByDatabaseID, job.DatabaseID)
	}
	return nil
}
```

Update `newSyncModelsService` helper in tests to accept dbRepo:

```go
func newSyncModelsService(repo *fakeSyncModelsJobRepo, dbRepo syncModelsDBRepo) *SyncModelsAppService {
	return NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo:     repo,
		DBRepo:          dbRepo,
		ReverseEngineer: &fakeSyncModelsReverseEngineer{},
		ModelRepo:       &fakeSyncModelsModelRepo{},
		FieldSyncer:     &fakeSyncModelsFieldSyncer{},
		Runner:          syncRunnerFunc(func(_ context.Context, fn func(context.Context)) { fn(context.Background()) }),
	})
}
```

- [ ] **Step 5: Run tests**

```bash
cd modelcraft-backend
just test-unit-pkg ./internal/app/modeldatabase/...
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add modelcraft-backend/internal/app/modeldatabase/
git commit -m "feat(app): merge ModelDatabaseSyncAppService into SyncModelsAppService with batch StartModelSync"
```

---

## Task 5: GraphQL Schema — add new types and deprecate old ones

**Files:**
- Modify: `modelcraft-backend/api/graph/project/schema/model.graphql`
- Modify: `modelcraft-backend/api/graph/project/schema/database.graphql`
- Regenerate: `modelcraft-backend/internal/interfaces/graphql/project/generated/`

- [ ] **Step 1: Update `model.graphql` — add batchId/databaseId to ModelSyncJob, add new query/mutation**

In `model.graphql`, replace the `ModelSyncJob` type, Query extension and Mutation extension:

```graphql
# syncModelsFromDB Job Types
# ============================================

enum ModelSyncJobStatus {
  PENDING
  RUNNING
  SUCCEEDED
  PARTIAL_SUCCESS
  FAILED
}

type ModelSyncFailedTable {
  tableName: String!
  message:   String!
}

type ModelSyncJob {
  id:               ID!
  batchId:          ID!
  databaseId:       ID!
  databaseName:     String!
  tableNames:       [String!]!
  status:           ModelSyncJobStatus!
  totalTables:      Int!
  processedTables:  Int!
  createdModels:    Int!
  syncedModels:     Int!
  failedCount:      Int!
  failedTables:     [ModelSyncFailedTable!]!
  startedAt:        Time
  finishedAt:       Time
  createdAt:        Time!
  updatedAt:        Time!
}

type ModelSyncJobRef {
  databaseId: ID!
  jobId:      ID!
}

type StartModelSyncPayload {
  batchId: ID!
  jobs:    [ModelSyncJobRef!]!
}

input ModelSyncTargetInput {
  databaseId: ID!
  tableNames: [String!]
}

type SyncModelsFromDBPayload {
  jobId: ID!
}

input SyncModelsFromDBInput {
  databaseName: String!
  tableNames:   [String!]
  syncAll:      Boolean
}
```

Update `extend type Query`:

```graphql
extend type Query {
  model(id: ID!, withActualSchema: Boolean): GetModelPayload! @hasPermission(action: "model:read", allowEndUser: true)
  models(input: ModelQueryInput): ModelListResult! @hasPermission(action: "model:read", allowEndUser: true)
  modelByName(name: String!, databaseName: String!): GetModelPayload! @hasPermission(action: "model:read")
  modelJsonSchema(id: ID!): ModelJsonSchema @hasPermission(action: "model:read")
  modelGroups: [ModelGroup!]! @hasPermission(action: "model:read")

  # Unified sync job query: batchId takes priority over jobIds if both provided
  modelSyncJobs(jobIds: [ID!], batchId: ID): [ModelSyncJob!]! @hasPermission(action: "model:read")

  # Deprecated
  modelSyncJob(jobId: ID!): ModelSyncJob @deprecated(reason: "Use modelSyncJobs") @hasPermission(action: "model:read")
}
```

Update `extend type Mutation`:

```graphql
extend type Mutation {
  createModel(input: CreateModelInput!): CreateModelPayload! @hasPermission(action: "model:create")
  updateModelMeta(id: ID!, input: UpdateModelMetaInput!): UpdateModelMetaPayload! @hasPermission(action: "model:update")
  deleteModel(id: ID!, dropTable: Boolean = false): DeleteModelPayload! @hasPermission(action: "model:delete")
  importModel(input: ImportModelInput!): ImportModelPayload! @hasPermission(action: "model:create")
  createModelFromSchema(input: CreateModelFromSchemaInput!): CreateModelFromSchemaPayload! @hasPermission(action: "model:create")

  # Unified sync mutation
  startModelSync(targets: [ModelSyncTargetInput!]!): StartModelSyncPayload! @hasPermission(action: "model:create")

  # Deprecated
  syncModelsFromDB(input: SyncModelsFromDBInput!): SyncModelsFromDBPayload! @deprecated(reason: "Use startModelSync") @hasPermission(action: "model:create")

  # ... keep other existing mutations unchanged
}
```

- [ ] **Step 2: Update `database.graphql` — deprecate modelDatabaseSyncJob and startModelDatabaseSync**

Find and update those two fields in `database.graphql`:

```graphql
# In extend type Query:
modelDatabaseSyncJob(jobId: ID!): ModelDatabaseSyncJob @deprecated(reason: "Use modelSyncJobs") @hasPermission(action: "model:read")

# In extend type Mutation:
startModelDatabaseSync(databaseId: ID!): StartModelDatabaseSyncPayload! @deprecated(reason: "Use startModelSync") @hasPermission(action: "model:create")
```

- [ ] **Step 3: Regenerate GraphQL code**

```bash
cd modelcraft-backend
just generate-gql
```

Expected: generated code updated in `internal/interfaces/graphql/project/generated/`. The generator will report any new unimplemented resolver methods.

- [ ] **Step 4: Compile check**

```bash
go build ./internal/interfaces/graphql/...
```

Expected: compile errors for unimplemented resolvers `StartModelSync` and `ModelSyncJobs` — fixed in Task 6.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/api/graph/project/schema/ \
        modelcraft-backend/internal/interfaces/graphql/project/generated/
git commit -m "feat(gql): add startModelSync mutation, modelSyncJobs query; deprecate old fields"
```

---

## Task 6: GraphQL Resolvers — implement new resolvers

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/model.resolvers.go`
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/resolver.go`
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`

- [ ] **Step 1: Add StartModelSync resolver to model.resolvers.go**

Add after the existing `SyncModelsFromDb` resolver:

```go
// StartModelSync is the resolver for the startModelSync field.
func (r *mutationResolver) StartModelSync(ctx context.Context, targets []*generated.ModelSyncTargetInput) (*generated.StartModelSyncPayload, error) {
	syncTargets := make([]appmodeldatabase.SyncTarget, len(targets))
	for i, t := range targets {
		tableNames := make([]string, len(t.TableNames))
		copy(tableNames, t.TableNames)
		syncTargets[i] = appmodeldatabase.SyncTarget{
			DatabaseID: t.DatabaseID,
			TableNames: tableNames,
		}
	}
	batchID, jobs, err := r.SyncModelsAppService.StartModelSync(ctx, syncTargets)
	if err != nil {
		return nil, err
	}
	refs := make([]*generated.ModelSyncJobRef, len(jobs))
	for i, job := range jobs {
		refs[i] = &generated.ModelSyncJobRef{
			DatabaseID: job.DatabaseID,
			JobID:      job.ID,
		}
	}
	return &generated.StartModelSyncPayload{
		BatchID: batchID,
		Jobs:    refs,
	}, nil
}

// ModelSyncJobs is the resolver for the modelSyncJobs field.
func (r *queryResolver) ModelSyncJobs(ctx context.Context, jobIDs []string, batchID *string) ([]*generated.ModelSyncJob, error) {
	bID := ""
	if batchID != nil {
		bID = *batchID
	}
	jobs, err := r.SyncModelsAppService.GetJobs(ctx, jobIDs, bID)
	if err != nil {
		return nil, err
	}
	result := make([]*generated.ModelSyncJob, len(jobs))
	for i, job := range jobs {
		result[i] = modelSyncJobToGQL(job)
	}
	return result, nil
}
```

Update `modelSyncJobToGQL` to include new fields:

```go
func modelSyncJobToGQL(job *domaindb.ModelSyncJob) *generated.ModelSyncJob {
	failedTables := make([]*generated.ModelSyncFailedTable, len(job.FailedTables))
	for i, ft := range job.FailedTables {
		ft := ft
		failedTables[i] = &generated.ModelSyncFailedTable{
			TableName: ft.TableName,
			Message:   ft.Message,
		}
	}
	tableNames := job.TableNames
	if tableNames == nil {
		tableNames = []string{}
	}
	return &generated.ModelSyncJob{
		ID:              job.ID,
		BatchID:         job.BatchID,
		DatabaseID:      job.DatabaseID,
		DatabaseName:    job.DatabaseName,
		TableNames:      tableNames,
		Status:          generated.ModelSyncJobStatus(job.Status),
		TotalTables:     int32(job.TotalTables),
		ProcessedTables: int32(job.ProcessedTables),
		CreatedModels:   int32(job.CreatedModels),
		SyncedModels:    int32(job.SyncedModels),
		FailedCount:     int32(job.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       job.StartedAt,
		FinishedAt:      job.FinishedAt,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
	}
}
```

- [ ] **Step 2: Wire DBRepo in routes.go**

In `routes.go`, when constructing `SyncModelsAppService`, add `DBRepo`:

```go
syncModelsAppService := appmodeldatabase.NewSyncModelsAppService(appmodeldatabase.SyncModelsAppServiceDeps{
	SyncJobRepo:     syncModelsJobRepo,
	DBRepo:          modelDatabaseRepo,  // ADD THIS
	ReverseEngineer: reverseEngineerService,
	ModelRepo:       modelRepo,
	FieldSyncer:     modelDesignService,
	GroupService:    groupAppService,    // ADD THIS
	Runner:          nil,
	Now:             nil,
})
```

Also call `RecoverStaleJobs` at startup:

```go
// After service creation, recover stale jobs
if err := syncModelsAppService.RecoverStaleJobs(ctx); err != nil {
	logger.Error(ctx, "failed to recover stale model sync jobs", logfacade.Err(err))
}
```

- [ ] **Step 3: Compile and verify**

```bash
cd modelcraft-backend
go build ./...
```

Expected: no errors.

- [ ] **Step 4: Run all unit tests**

```bash
cd modelcraft-backend
just test-unit
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/interfaces/graphql/project/model.resolvers.go \
        modelcraft-backend/internal/interfaces/graphql/project/resolver.go \
        modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "feat(resolver): implement StartModelSync, ModelSyncJobs resolvers; wire DBRepo and GroupService"
```

---

## Task 7: Frontend — sync contract and add new GQL docs

**Files:**
- Sync: `modelcraft-front/src/contract/` (via skill)
- Modify: `modelcraft-front/src/api-client/model/graphql-docs.ts`

- [ ] **Step 1: Sync contract from backend**

Run the `front-contract-pull` skill to pull updated schema into `contract/`.

- [ ] **Step 2: Run codegen**

```bash
cd modelcraft-front
npm run codegen
```

Expected: `src/generated/graphql.ts` updated with `ModelSyncJob` (new fields), `StartModelSyncPayload`, `ModelSyncJobRef`, `ModelSyncTargetInput`, `modelSyncJobs` query.

- [ ] **Step 3: Add new GQL documents to `api-client/model/graphql-docs.ts`**

Add after existing exports:

```typescript
export const MODEL_SYNC_JOBS_QUERY = gql`
  query ModelSyncJobs($jobIds: [ID!], $batchId: ID) {
    modelSyncJobs(jobIds: $jobIds, batchId: $batchId) {
      id
      batchId
      databaseId
      databaseName
      tableNames
      status
      totalTables
      processedTables
      createdModels
      syncedModels
      failedCount
      failedTables {
        tableName
        message
      }
      startedAt
      finishedAt
      createdAt
      updatedAt
    }
  }
`

export const START_MODEL_SYNC_MUTATION = gql`
  mutation StartModelSync($targets: [ModelSyncTargetInput!]!) {
    startModelSync(targets: $targets) {
      batchId
      jobs {
        databaseId
        jobId
      }
    }
  }
`
```

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/contract/ \
        modelcraft-front/src/generated/graphql.ts \
        modelcraft-front/src/api-client/model/graphql-docs.ts
git commit -m "feat(front): sync contract, add StartModelSync and ModelSyncJobs GQL docs"
```

---

## Task 8: Frontend — new hook and migrate call sites

**Files:**
- Create: `modelcraft-front/src/web/hooks/model/use-model-sync.ts`
- Modify: `modelcraft-front/src/web/hooks/model-database/use-model-databases.ts`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/page.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/_components/batch-register-sync.ts`

- [ ] **Step 1: Create `use-model-sync.ts`**

```typescript
// modelcraft-front/src/web/hooks/model/use-model-sync.ts
import { useMutation } from '@apollo/client'
import {
  START_MODEL_SYNC_MUTATION,
  MODEL_SYNC_JOBS_QUERY,
} from '@/api-client/model'
import type {
  StartModelSyncMutation,
  StartModelSyncMutationVariables,
  ModelSyncJobsQuery,
  ModelSyncJobsQueryVariables,
  ModelSyncJob,
} from '@/generated/graphql'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'

export interface SyncTarget {
  databaseId: string
  tableNames?: string[]
}

export interface ModelSyncJobRef {
  databaseId: string
  jobId: string
}

export function useStartModelSync(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug)
  const [mutate, { loading, error }] = useMutation<
    StartModelSyncMutation,
    StartModelSyncMutationVariables
  >(START_MODEL_SYNC_MUTATION, { client })

  const startSync = async (
    targets: SyncTarget[]
  ): Promise<{ batchId: string; jobs: ModelSyncJobRef[] } | null> => {
    const result = await mutate({ variables: { targets } })
    const payload = result.data?.startModelSync
    if (!payload) return null
    return {
      batchId: payload.batchId,
      jobs: payload.jobs.map((j) => ({ databaseId: j.databaseId, jobId: j.jobId })),
    }
  }

  return { startSync, loading, error }
}

export function useFetchModelSyncJobs(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug)

  const fetchByBatch = async (batchId: string): Promise<ModelSyncJob[]> => {
    const result = await client.query<ModelSyncJobsQuery, ModelSyncJobsQueryVariables>({
      query: MODEL_SYNC_JOBS_QUERY,
      variables: { batchId },
      fetchPolicy: 'network-only',
    })
    return result.data.modelSyncJobs ?? []
  }

  const fetchByJobIds = async (jobIds: string[]): Promise<ModelSyncJob[]> => {
    const result = await client.query<ModelSyncJobsQuery, ModelSyncJobsQueryVariables>({
      query: MODEL_SYNC_JOBS_QUERY,
      variables: { jobIds },
      fetchPolicy: 'network-only',
    })
    return result.data.modelSyncJobs ?? []
  }

  return { fetchByBatch, fetchByJobIds }
}
```

- [ ] **Step 2: Update `use-model-databases.ts` — deprecate old sync hooks, add new ones**

Replace `useStartModelDatabaseSync` and `useFetchModelDatabaseSyncJob` bodies to use the new hooks internally (keeping the same exported function signatures for backward compatibility):

```typescript
// In use-model-databases.ts — replace the two functions:

export function useStartModelDatabaseSync(projectSlug: string) {
  const { startSync, loading, error } = useStartModelSync(projectSlug)

  const startDatabaseSync = async (databaseId: string): Promise<ModelDatabaseSyncJob | null> => {
    const result = await startSync([{ databaseId }])
    if (!result || result.jobs.length === 0) return null
    // Return a compatible shape using the first job ref
    const jobRef = result.jobs[0]
    return {
      id: jobRef.jobId,
      databaseId: jobRef.databaseId,
      status: 'PENDING',
      totalTables: 0,
      processedTables: 0,
      createdModels: 0,
      syncedModels: 0,
      failedCount: 0,
      failedTables: [],
      startedAt: null,
      finishedAt: null,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
  }

  return { startSync: startDatabaseSync, loading, error }
}

export function useFetchModelDatabaseSyncJob(projectSlug: string) {
  const { fetchByJobIds } = useFetchModelSyncJobs(projectSlug)

  const fetchJob = async (jobId: string): Promise<ModelDatabaseSyncJob | null> => {
    const jobs = await fetchByJobIds([jobId])
    if (jobs.length === 0) return null
    const job = jobs[0]
    return {
      id: job.id,
      databaseId: job.databaseId,
      status: job.status as ModelDatabaseSyncJobStatus,
      totalTables: job.totalTables,
      processedTables: job.processedTables,
      createdModels: job.createdModels,
      syncedModels: job.syncedModels,
      failedCount: job.failedCount,
      failedTables: job.failedTables,
      startedAt: job.startedAt ?? null,
      finishedAt: job.finishedAt ?? null,
      createdAt: job.createdAt,
      updatedAt: job.updatedAt,
    }
  }

  return { fetchJob }
}
```

Add import at top:

```typescript
import { useStartModelSync, useFetchModelSyncJobs } from '@web/hooks/model/use-model-sync'
```

- [ ] **Step 3: Update batch-register-sync.ts to use new types**

`batch-register-sync.ts` uses `ModelDatabaseSyncJob` type — since `useStartModelDatabaseSync` still returns the same shape, no changes are needed. Verify by checking TypeScript:

```bash
cd modelcraft-front
npx tsc --noEmit
```

Expected: no type errors.

- [ ] **Step 4: Run lint**

```bash
cd modelcraft-front
npm run lint
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/web/hooks/model/use-model-sync.ts \
        modelcraft-front/src/web/hooks/model-database/use-model-databases.ts
git commit -m "feat(front): add useStartModelSync hook; bridge deprecated useStartModelDatabaseSync to new API"
```

---

## Task 9: Frontend — migrate ImportModelDialog to new API

**Files:**
- Modify: `modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx`
- Modify: `modelcraft-front/src/web/hooks/model/use-sync-models-from-db.ts`

- [ ] **Step 1: Update use-sync-models-from-db.ts to use startModelSync**

`ImportModelDialog` uses `useSyncModelsFromDB` (old `syncModelsFromDB` mutation) which takes `databaseName`. We migrate it to `startModelSync` with `databaseId`.

Check what `ImportModelDialog` passes — it likely has access to a `databaseId`. If it only has `databaseName`, leave the old hook intact for now and track as a follow-up. Check the component:

```bash
grep -n "databaseName\|databaseId\|syncModels" modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx | head -30
```

If `databaseId` is available in the component, update `use-sync-models-from-db.ts`:

```typescript
// use-sync-models-from-db.ts
import { useStartModelSync, useFetchModelSyncJobs } from '@web/hooks/model/use-model-sync'
import type { ModelSyncJob } from '@/generated/graphql'
import { useProjectScopedClient } from '@api-client/apollo/develop-client'
import { MODEL_SYNC_JOBS_QUERY } from '@/api-client/model'
import { useQuery } from '@apollo/client'
import type { ModelSyncJobsQuery, ModelSyncJobsQueryVariables } from '@/generated/graphql'

export function useSyncModelsFromDB(projectSlug: string | null | undefined) {
  return useStartModelSync(projectSlug)
}

export function useModelSyncJob(
  jobId: string | null,
  projectSlug: string | null | undefined
) {
  const client = useProjectScopedClient(projectSlug)
  return useQuery<ModelSyncJobsQuery, ModelSyncJobsQueryVariables>(MODEL_SYNC_JOBS_QUERY, {
    client,
    variables: { jobIds: [jobId!] },
    skip: !jobId,
    pollInterval: 2000,
  })
}
```

- [ ] **Step 2: TypeScript check**

```bash
cd modelcraft-front
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/hooks/model/use-sync-models-from-db.ts \
        modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx
git commit -m "feat(front): migrate ImportModelDialog sync to new startModelSync API"
```

---

## Self-Review Checklist

**Spec coverage:**
- [x] DB: `batch_id`, `database_id` added to `model_sync_job` — Task 1
- [x] Domain struct updated — Task 2
- [x] Repository: `GetByIDs`, `GetByBatchID`, `FailStalePendingJobs` — Tasks 2 & 3
- [x] App service: `StartModelSync`, `GetJobs`, `RecoverStaleJobs`, group assignment — Task 4
- [x] GraphQL: `startModelSync`, `modelSyncJobs`, `ModelSyncJob` new fields — Tasks 5 & 6
- [x] GraphQL: deprecated `syncModelsFromDB`, `startModelDatabaseSync`, `modelSyncJob`, `modelDatabaseSyncJob` — Task 5
- [x] Frontend: new GQL docs, `useStartModelSync`, `useFetchModelSyncJobs` — Tasks 7 & 8
- [x] Frontend: bridge deprecated hooks to new API — Task 8
- [x] `databaseName` stored as redundant field from `model_database` at job creation — Task 4 (`StartModelSync`)
- [x] `model_database_sync_job` frozen (no new writes) — not a code task, enforced by removing all callers from new paths

**Type consistency:**
- `SyncTarget` defined in Task 4, used in Task 6 ✓
- `modelSyncJobToGQL` updated in Task 6 to include `BatchID`/`DatabaseID` ✓
- `ModelSyncJobRef.JobID` / `ModelSyncJobRef.DatabaseID` match generated field names ✓ (verify after `just generate-gql`)
- `GetJobs` defined in Task 4, called in Task 6 ✓
