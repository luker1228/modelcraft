package modeldatabase

import (
	"context"
	"modelcraft/internal/app/modeldesign"
	domaindb "modelcraft/internal/domain/modeldatabase"
	domainmodel "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"
	"time"

	"github.com/google/uuid"
)

const (
	importGroupName    = "数据库导入"
	defaultStalePeriod = 30 * time.Minute
)

type modelDatabaseSyncRunner interface {
	Go(ctx context.Context, fn func(context.Context))
}

type modelDatabaseReverseEngineer interface {
	ListTables(
		ctx context.Context,
		orgName, projectSlug, databaseName string,
		excludeExisting bool,
		limit, offset int,
	) (*modeldesign.ListTablesResult, error)
	ImportModel(ctx context.Context, cmd modeldesign.ImportModelCommand) (*modeldesign.ImportModelResult, error)
	BuildSchemaForTable(
		ctx context.Context,
		cmd modeldesign.ImportModelCommand,
	) (*modeldesign.TableSchemaBuildResult, error)
}

type modelDatabaseSchemaSync interface {
	SyncModelSchemaFromJSON(
		ctx context.Context,
		modelID string,
		schemaJSON string,
		deleteExtraFields bool,
	) (*modeldesign.SyncModelSchemaResult, error)
}

type modelDatabaseGroupService interface {
	EnsureImportGroup(ctx context.Context, orgName, projectSlug string) (*domainmodel.ModelGroup, error)
	MoveModelToGroup(ctx context.Context, modelID string, groupID *string) error
}

type backgroundRunner struct{}

func (backgroundRunner) Go(ctx context.Context, fn func(context.Context)) {
	bizutils.GoWithCtx(ctx, fn)
}

type ModelDatabaseSyncAppServiceDeps struct {
	ModelDatabaseRepo domaindb.ModelDatabaseRepository
	SyncJobRepo       domaindb.ModelDatabaseSyncJobRepository
	ReverseEngineer   modelDatabaseReverseEngineer
	ModelRepo         domainmodel.ModelRepository
	SchemaSync        modelDatabaseSchemaSync
	GroupService      modelDatabaseGroupService
	Runner            modelDatabaseSyncRunner
	Now               func() time.Time
}

type ModelDatabaseSyncAppService struct {
	modelDatabaseRepo domaindb.ModelDatabaseRepository
	syncJobRepo       domaindb.ModelDatabaseSyncJobRepository
	reverseEngineer   modelDatabaseReverseEngineer
	modelRepo         domainmodel.ModelRepository
	schemaSync        modelDatabaseSchemaSync
	groupService      modelDatabaseGroupService
	runner            modelDatabaseSyncRunner
	now               func() time.Time
}

func NewModelDatabaseSyncAppService(deps ModelDatabaseSyncAppServiceDeps) *ModelDatabaseSyncAppService {
	runner := deps.Runner
	if runner == nil {
		runner = backgroundRunner{}
	}
	nowFn := deps.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	return &ModelDatabaseSyncAppService{
		modelDatabaseRepo: deps.ModelDatabaseRepo,
		syncJobRepo:       deps.SyncJobRepo,
		reverseEngineer:   deps.ReverseEngineer,
		modelRepo:         deps.ModelRepo,
		schemaSync:        deps.SchemaSync,
		groupService:      deps.GroupService,
		runner:            runner,
		now:               nowFn,
	}
}

func (s *ModelDatabaseSyncAppService) StartSync(ctx context.Context, databaseID string) (*domaindb.ModelDatabaseSyncJob, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	if _, err := s.modelDatabaseRepo.GetByID(ctx, orgName, projectSlug, databaseID); err != nil {
		return nil, err
	}
	active, err := s.syncJobRepo.GetActiveByDatabase(ctx, orgName, projectSlug, databaseID, s.now().Add(-defaultStalePeriod))
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, bizerrors.NewError(bizerrors.Conflict, "sync job already running for database")
	}

	now := s.now()
	job := &domaindb.ModelDatabaseSyncJob{
		ID:          uuid.NewString(),
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		DatabaseID:  databaseID,
		Status:      domaindb.ModelDatabaseSyncJobStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.syncJobRepo.Create(ctx, job); err != nil {
		return nil, err
	}
	s.runner.Go(ctx, func(runCtx context.Context) {
		_ = s.RunSyncJob(runCtx, job.ID)
	})
	return job, nil
}

// RecoverStaleJobs 将超过 stalePeriod 未更新的 pending/running job 全部标记为 failed。
// 通常在服务启动时调用，清理因进程崩溃遗留的僵尸 job。
func (s *ModelDatabaseSyncAppService) RecoverStaleJobs(ctx context.Context) error {
	staleBefore := s.now().Add(-defaultStalePeriod)
	return s.syncJobRepo.FailStalePendingJobs(ctx, staleBefore)
}

func (s *ModelDatabaseSyncAppService) GetJob(ctx context.Context, jobID string) (*domaindb.ModelDatabaseSyncJob, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	return s.syncJobRepo.GetByID(ctx, orgName, projectSlug, jobID)
}

func (s *ModelDatabaseSyncAppService) RunSyncJob(ctx context.Context, jobID string) error {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	job, err := s.syncJobRepo.GetByID(ctx, orgName, projectSlug, jobID)
	if err != nil {
		return err
	}
	db, err := s.modelDatabaseRepo.GetByID(ctx, orgName, projectSlug, job.DatabaseID)
	if err != nil {
		return s.failJob(ctx, job, err)
	}

	now := s.now()
	job.Status = domaindb.ModelDatabaseSyncJobStatusRunning
	job.StartedAt = &now
	job.UpdatedAt = now
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		return err
	}

	tableResult, err := s.reverseEngineer.ListTables(ctx, orgName, projectSlug, db.Name, false, 0, 0)
	if err != nil {
		return s.failJob(ctx, job, err)
	}
	job.TotalTables = tableResult.TotalCount
	job.UpdatedAt = s.now()
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		return err
	}

	group, err := s.groupService.EnsureImportGroup(ctx, orgName, projectSlug)
	if err != nil {
		return s.failJob(ctx, job, err)
	}

	for _, tableName := range tableResult.Tables {
		if err := s.processTable(ctx, job, db, group, tableName); err != nil {
			return err
		}
	}

	finishedAt := s.now()
	job.FinishedAt = &finishedAt
	job.UpdatedAt = finishedAt
	switch {
	case job.FailedCount == 0:
		job.Status = domaindb.ModelDatabaseSyncJobStatusSucceeded
	case job.CreatedModels > 0 || job.SyncedModels > 0:
		job.Status = domaindb.ModelDatabaseSyncJobStatusPartialSuccess
	default:
		job.Status = domaindb.ModelDatabaseSyncJobStatusFailed
	}
	return s.syncJobRepo.Update(ctx, job)
}

func (s *ModelDatabaseSyncAppService) processTable(
	ctx context.Context,
	job *domaindb.ModelDatabaseSyncJob,
	db *domaindb.ModelDatabase,
	group *domainmodel.ModelGroup,
	tableName string,
) error {
	buildResult, err := s.reverseEngineer.BuildSchemaForTable(ctx, modeldesign.ImportModelCommand{
		OrgName:      job.OrgName,
		ProjectSlug:  job.ProjectSlug,
		DatabaseName: db.Name,
		TableName:    tableName,
	})
	if err != nil {
		return s.recordTableFailure(ctx, job, tableName, err)
	}

	existingModel, err := s.modelRepo.GetByName(
		ctx,
		job.OrgName,
		db.Name,
		buildResult.ModelName,
		job.ProjectSlug,
	)
	if err != nil && !shared.IsNotFoundError(err) {
		return s.recordTableFailure(ctx, job, tableName, err)
	}

	if existingModel == nil {
		importResult, importErr := s.reverseEngineer.ImportModel(ctx, modeldesign.ImportModelCommand{
			OrgName:      job.OrgName,
			ProjectSlug:  job.ProjectSlug,
			DatabaseName: db.Name,
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
	} else {
		if _, err := s.schemaSync.SyncModelSchemaFromJSON(ctx, existingModel.ID, buildResult.SchemaJSON, false); err != nil {
			return s.recordTableFailure(ctx, job, tableName, err)
		}
		job.SyncedModels++
	}

	job.ProcessedTables++
	job.UpdatedAt = s.now()
	return s.syncJobRepo.Update(ctx, job)
}

func (s *ModelDatabaseSyncAppService) recordTableFailure(
	ctx context.Context,
	job *domaindb.ModelDatabaseSyncJob,
	tableName string,
	err error,
) error {
	job.ProcessedTables++
	job.FailedCount++
	job.FailedTables = append(job.FailedTables, domaindb.ModelDatabaseSyncFailedTable{
		TableName: tableName,
		Message:   err.Error(),
	})
	job.UpdatedAt = s.now()
	return s.syncJobRepo.Update(ctx, job)
}

func (s *ModelDatabaseSyncAppService) failJob(
	ctx context.Context,
	job *domaindb.ModelDatabaseSyncJob,
	err error,
) error {
	now := s.now()
	job.Status = domaindb.ModelDatabaseSyncJobStatusFailed
	job.FinishedAt = &now
	job.UpdatedAt = now
	if updateErr := s.syncJobRepo.Update(ctx, job); updateErr != nil {
		return updateErr
	}
	return err
}
