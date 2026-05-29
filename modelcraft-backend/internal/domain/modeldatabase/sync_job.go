package modeldatabase

import "time"

type ModelDatabaseSyncJobStatus string

const (
	ModelDatabaseSyncJobStatusPending        ModelDatabaseSyncJobStatus = "pending"
	ModelDatabaseSyncJobStatusRunning        ModelDatabaseSyncJobStatus = "running"
	ModelDatabaseSyncJobStatusSucceeded      ModelDatabaseSyncJobStatus = "succeeded"
	ModelDatabaseSyncJobStatusPartialSuccess ModelDatabaseSyncJobStatus = "partial_success"
	ModelDatabaseSyncJobStatusFailed         ModelDatabaseSyncJobStatus = "failed"
)

type ModelDatabaseSyncFailedTable struct {
	TableName string
	Message   string
}

type ModelDatabaseSyncJob struct {
	ID              string
	OrgName         string
	ProjectSlug     string
	DatabaseID      string
	Status          ModelDatabaseSyncJobStatus
	TotalTables     int
	ProcessedTables int
	CreatedModels   int
	SyncedModels    int
	FailedCount     int
	FailedTables    []ModelDatabaseSyncFailedTable
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
