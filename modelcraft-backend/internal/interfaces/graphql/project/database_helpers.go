package projectgraphql

import (
	domainmodeldatabase "modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/interfaces/graphql/project/generated"
)

func modelDatabaseToGQL(db *domainmodeldatabase.ModelDatabase) *generated.ModelDatabase {
	return &generated.ModelDatabase{
		ID:          db.ID,
		Name:        db.Name,
		Title:       db.Title,
		Description: db.Description,
		Mode:        domainDatabaseModeToGQL(db.Mode),
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}
}

func domainDatabaseModeToGQL(m domainmodeldatabase.DatabaseMode) generated.DatabaseMode {
	if m == domainmodeldatabase.DatabaseModeManaged {
		return generated.DatabaseModeManaged
	}
	return generated.DatabaseModeSelfHosted
}

func gqlDatabaseModeToDomain(m generated.DatabaseMode) domainmodeldatabase.DatabaseMode {
	if m == generated.DatabaseModeManaged {
		return domainmodeldatabase.DatabaseModeManaged
	}
	return domainmodeldatabase.DatabaseModeSelfHosted
}

func domainDatabaseModeFromGQLPtr(m *generated.DatabaseMode) *domainmodeldatabase.DatabaseMode {
	if m == nil {
		return nil
	}
	mode := gqlDatabaseModeToDomain(*m)
	return &mode
}

func derefStringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func modelDatabaseSyncJobToGQL(job *domainmodeldatabase.ModelDatabaseSyncJob) *generated.ModelDatabaseSyncJob {
	failedTables := make([]*generated.ModelDatabaseSyncFailedTable, 0, len(job.FailedTables))
	for _, item := range job.FailedTables {
		failedTables = append(failedTables, &generated.ModelDatabaseSyncFailedTable{
			TableName: item.TableName,
			Message:   item.Message,
		})
	}

	return &generated.ModelDatabaseSyncJob{
		ID:              job.ID,
		DatabaseID:      job.DatabaseID,
		Status:          domainSyncJobStatusToGQL(job.Status),
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

func domainSyncJobStatusToGQL(
	status domainmodeldatabase.ModelDatabaseSyncJobStatus,
) generated.ModelDatabaseSyncJobStatus {
	switch status {
	case domainmodeldatabase.ModelDatabaseSyncJobStatusPending:
		return generated.ModelDatabaseSyncJobStatusPending
	case domainmodeldatabase.ModelDatabaseSyncJobStatusRunning:
		return generated.ModelDatabaseSyncJobStatusRunning
	case domainmodeldatabase.ModelDatabaseSyncJobStatusSucceeded:
		return generated.ModelDatabaseSyncJobStatusSucceeded
	case domainmodeldatabase.ModelDatabaseSyncJobStatusPartialSuccess:
		return generated.ModelDatabaseSyncJobStatusPartialSuccess
	default:
		return generated.ModelDatabaseSyncJobStatusFailed
	}
}
