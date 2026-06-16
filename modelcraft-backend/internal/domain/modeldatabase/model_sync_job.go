package modeldatabase

import (
	"time"
)

type ModelSyncJobStatus string

const (
	ModelSyncJobStatusPending        ModelSyncJobStatus = "pending"
	ModelSyncJobStatusRunning        ModelSyncJobStatus = "running"
	ModelSyncJobStatusSucceeded      ModelSyncJobStatus = "succeeded"
	ModelSyncJobStatusPartialSuccess ModelSyncJobStatus = "partial_success"
	ModelSyncJobStatusFailed         ModelSyncJobStatus = "failed"
)

type ModelSyncFailedTable struct {
	TableName string `json:"tableName"`
	Message   string `json:"message"`
}

type ModelSyncJob struct {
	ID              string
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
