package modeldatabase

import (
	"context"
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

type ModelSyncJobRepository interface {
	Create(ctx context.Context, job *ModelSyncJob) error
	GetByID(ctx context.Context, orgName, projectSlug, jobID string) (*ModelSyncJob, error)
	GetActiveByDatabase(ctx context.Context, orgName, projectSlug, databaseName string) (*ModelSyncJob, error)
	Update(ctx context.Context, job *ModelSyncJob) error
}
