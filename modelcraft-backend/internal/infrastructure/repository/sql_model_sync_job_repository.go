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

func (r *SqlModelSyncJobRepository) GetActiveByDatabase(
	ctx context.Context, orgName, projectSlug, databaseName string, staleBefore time.Time,
) (*domaindb.ModelSyncJob, error) {
	row, err := r.q.GetActiveModelSyncJobByDatabase(ctx, dbgen.GetActiveModelSyncJobByDatabaseParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		DatabaseID:  databaseName,
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

func (r *SqlModelSyncJobRepository) FailStalePendingJobs(ctx context.Context, staleBefore time.Time) error {
	return r.q.FailStaleModelSyncJobs(ctx, staleBefore)
}

// Ensure interface is fully implemented
var _ domaindb.ModelSyncJobRepository = (*SqlModelSyncJobRepository)(nil)
