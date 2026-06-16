package modeldatabase

import (
	"context"
	"time"
)

// ModelDatabaseRepository 项目数据库注册仓储接口
type ModelDatabaseRepository interface {
	// Create 创建数据库注册记录
	Create(ctx context.Context, db *ModelDatabase) error
	// GetByID 按 ID 查询（需 orgName + projectSlug 隔离）
	GetByID(ctx context.Context, orgName, projectSlug, id string) (*ModelDatabase, error)
	// GetByName 按数据库名查询
	GetByName(ctx context.Context, orgName, projectSlug, name string) (*ModelDatabase, error)
	// List 列出项目下所有已注册的数据库
	List(ctx context.Context, orgName, projectSlug string) ([]*ModelDatabase, error)
	// Update 更新数据库注册记录
	Update(ctx context.Context, orgName, projectSlug string, db *ModelDatabase) error
	// UpdateLatestSyncJobID 更新最近同步任务 ID
	UpdateLatestSyncJobID(ctx context.Context, orgName, projectSlug, databaseID, jobID string) error
	// Delete 软删除数据库注册记录
	Delete(ctx context.Context, orgName, projectSlug, id string) error
}

type ModelDatabaseSyncJobRepository interface {
	Create(ctx context.Context, job *ModelDatabaseSyncJob) error
	GetByID(ctx context.Context, orgName, projectSlug, jobID string) (*ModelDatabaseSyncJob, error)
	// GetActiveByDatabase 返回指定数据库的活跃 job（pending/running），仅返回 staleBefore 之后有心跳的。
	GetActiveByDatabase(
		ctx context.Context,
		orgName, projectSlug, databaseID string,
		staleBefore time.Time,
	) (*ModelDatabaseSyncJob, error)
	// FailStalePendingJobs 将所有 updated_at <= staleBefore 的 pending/running job 标记为 failed。
	FailStalePendingJobs(ctx context.Context, staleBefore time.Time) error
	Update(ctx context.Context, job *ModelDatabaseSyncJob) error
}

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
