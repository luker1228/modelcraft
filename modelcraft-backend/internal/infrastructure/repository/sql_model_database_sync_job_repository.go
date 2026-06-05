package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	domaindb "modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"
)

type SqlModelDatabaseSyncJobRepository struct {
	q dbgen.Querier
}

func NewSqlModelDatabaseSyncJobRepository(q dbgen.Querier) domaindb.ModelDatabaseSyncJobRepository {
	return &SqlModelDatabaseSyncJobRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

func (r *SqlModelDatabaseSyncJobRepository) Create(ctx context.Context, job *domaindb.ModelDatabaseSyncJob) error {
	failedTables, err := marshalSyncFailedTables(job.FailedTables)
	if err != nil {
		return err
	}
	return r.q.CreateModelDatabaseSyncJob(ctx, dbgen.CreateModelDatabaseSyncJobParams{
		ID:              job.ID,
		OrgName:         job.OrgName,
		ProjectSlug:     job.ProjectSlug,
		DatabaseID:      job.DatabaseID,
		Status:          dbgen.ModelDatabaseSyncJobStatus(job.Status),
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

func (r *SqlModelDatabaseSyncJobRepository) GetByID(
	ctx context.Context,
	orgName, projectSlug, jobID string,
) (*domaindb.ModelDatabaseSyncJob, error) {
	row, err := r.q.GetModelDatabaseSyncJobByID(ctx, dbgen.GetModelDatabaseSyncJobByIDParams{
		ID:          jobID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // nil is the canonical "not found" sentinel for this repository method
		}
		return nil, err
	}
	return modelDatabaseSyncJobToDomain(row)
}

func (r *SqlModelDatabaseSyncJobRepository) GetActiveByDatabase(
	ctx context.Context,
	orgName, projectSlug, databaseID string,
	staleBefore time.Time,
) (*domaindb.ModelDatabaseSyncJob, error) {
	row, err := r.q.GetActiveModelDatabaseSyncJobByDatabase(ctx, dbgen.GetActiveModelDatabaseSyncJobByDatabaseParams{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseID:   databaseID,
		UpdatedAfter: staleBefore,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil // nil is the canonical "not found" sentinel for this repository method
		}
		return nil, err
	}
	return modelDatabaseSyncJobToDomain(row)
}

func (r *SqlModelDatabaseSyncJobRepository) FailStalePendingJobs(
	ctx context.Context,
	staleBefore time.Time,
) error {
	return r.q.FailStaleSyncJobs(ctx, staleBefore)
}

func (r *SqlModelDatabaseSyncJobRepository) Update(ctx context.Context, job *domaindb.ModelDatabaseSyncJob) error {
	failedTables, err := marshalSyncFailedTables(job.FailedTables)
	if err != nil {
		return err
	}
	return r.q.UpdateModelDatabaseSyncJob(ctx, dbgen.UpdateModelDatabaseSyncJobParams{
		Status:          dbgen.ModelDatabaseSyncJobStatus(job.Status),
		TotalTables:     int32(job.TotalTables),
		ProcessedTables: int32(job.ProcessedTables),
		CreatedModels:   int32(job.CreatedModels),
		SyncedModels:    int32(job.SyncedModels),
		FailedCount:     int32(job.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       timeToNull(job.StartedAt),
		FinishedAt:      timeToNull(job.FinishedAt),
		ID:              job.ID,
	})
}

func modelDatabaseSyncJobToDomain(row dbgen.ModelDatabaseSyncJob) (*domaindb.ModelDatabaseSyncJob, error) {
	failedTables, err := unmarshalSyncFailedTables(row.FailedTables)
	if err != nil {
		return nil, err
	}

	return &domaindb.ModelDatabaseSyncJob{
		ID:              row.ID,
		OrgName:         row.OrgName,
		ProjectSlug:     row.ProjectSlug,
		DatabaseID:      row.DatabaseID,
		Status:          domaindb.ModelDatabaseSyncJobStatus(row.Status),
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

func marshalSyncFailedTables(items []domaindb.ModelDatabaseSyncFailedTable) (json.RawMessage, error) {
	if items == nil {
		items = []domaindb.ModelDatabaseSyncFailedTable{}
	}
	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func unmarshalSyncFailedTables(data json.RawMessage) ([]domaindb.ModelDatabaseSyncFailedTable, error) {
	if len(data) == 0 {
		return []domaindb.ModelDatabaseSyncFailedTable{}, nil
	}
	var items []domaindb.ModelDatabaseSyncFailedTable
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func timeToNull(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func nullToTimePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

var _ domaindb.ModelDatabaseSyncJobRepository = (*SqlModelDatabaseSyncJobRepository)(nil)
