package modeldatabase

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaindb "modelcraft/internal/domain/modeldatabase"
	domainmodel "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"

	appmodeldesign "modelcraft/internal/app/modeldesign"
)

// SyncTarget is one database to sync, with optional table filter.
type SyncTarget struct {
	DatabaseID string
	TableNames []string // nil or empty = full sync
}

// syncModelsDBRepo is the subset of ModelDatabaseRepository used here.
type syncModelsDBRepo interface {
	GetByID(ctx context.Context, orgName, projectSlug, id string) (*domaindb.ModelDatabase, error)
	GetByName(ctx context.Context, orgName, projectSlug, name string) (*domaindb.ModelDatabase, error)
	UpdateLatestSyncJobID(ctx context.Context, orgName, projectSlug, databaseID, jobID string) error
}

// syncModelsGroupService is the subset of ModelGroupAppService used here.
type syncModelsGroupService interface {
	EnsureImportGroup(ctx context.Context, orgName, projectSlug string) (*domainmodel.ModelGroup, error)
	MoveModelToGroup(ctx context.Context, modelID string, groupID *string) error
}

// SyncModelsFromDBCommand is the input for triggering a sync job.
type SyncModelsFromDBCommand struct {
	DatabaseName string
	TableNames   []string // nil or empty means full sync (only valid with SyncAll=true)
	SyncAll      bool
}

// syncModelsReverseEngineer is the subset of ReverseEngineerAppService used here.
type syncModelsReverseEngineer interface {
	ListTables(
		ctx context.Context,
		orgName, projectSlug, databaseName string,
		excludeExisting bool,
		limit, offset int,
	) (*appmodeldesign.ListTablesResult, error)
	ImportModel(ctx context.Context, cmd appmodeldesign.ImportModelCommand) (*appmodeldesign.ImportModelResult, error)
	GetTableDefinition(
		ctx context.Context,
		orgName, projectSlug, databaseName, tableName string,
	) (*appmodeldesign.TableDefinitionResult, error)
}

// syncModelsFieldSyncer is the subset of ModelDesignAppService used here.
type syncModelsFieldSyncer interface {
	SyncFieldsFromDB(ctx context.Context, modelID string, dbFields []*domainmodel.FieldDefinition) error
}

// syncModelsModelRepo is the subset of ModelRepository used here.
type syncModelsModelRepo interface {
	GetByName(
		ctx context.Context,
		orgName, databaseName, name, projectSlug string,
		opts ...*domainmodel.ModelQueryOptions,
	) (*domainmodel.DataModel, error)
}

// SyncModelsAppServiceDeps holds dependencies for SyncModelsAppService.
type SyncModelsAppServiceDeps struct {
	SyncJobRepo     domaindb.ModelSyncJobRepository
	DBRepo          syncModelsDBRepo         // may be nil (degraded mode)
	ReverseEngineer syncModelsReverseEngineer
	ModelRepo       syncModelsModelRepo
	FieldSyncer     syncModelsFieldSyncer
	GroupService    syncModelsGroupService   // may be nil
	Runner          modelDatabaseSyncRunner
	Now             func() time.Time
}

// SyncModelsAppService orchestrates syncModelsFromDB async jobs.
type SyncModelsAppService struct {
	syncJobRepo     domaindb.ModelSyncJobRepository
	dbRepo          syncModelsDBRepo
	reverseEngineer syncModelsReverseEngineer
	modelRepo       syncModelsModelRepo
	fieldSyncer     syncModelsFieldSyncer
	groupService    syncModelsGroupService
	runner          modelDatabaseSyncRunner
	now             func() time.Time
}

// NewSyncModelsAppService creates a SyncModelsAppService with defaults applied.
func NewSyncModelsAppService(deps SyncModelsAppServiceDeps) *SyncModelsAppService {
	runner := deps.Runner
	if runner == nil {
		runner = backgroundRunner{}
	}
	nowFn := deps.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	return &SyncModelsAppService{
		syncJobRepo:     deps.SyncJobRepo,
		dbRepo:          deps.DBRepo,
		reverseEngineer: deps.ReverseEngineer,
		modelRepo:       deps.ModelRepo,
		fieldSyncer:     deps.FieldSyncer,
		groupService:    deps.GroupService,
		runner:          runner,
		now:             nowFn,
	}
}

// StartSync validates input, checks for active jobs, creates the job, and fires the background runner.
func (s *SyncModelsAppService) StartSync(
	ctx context.Context,
	cmd SyncModelsFromDBCommand,
) (*domaindb.ModelSyncJob, error) {
	if err := validateSyncModelsCommand(cmd); err != nil {
		return nil, err
	}

	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}

	// Resolve database ID if dbRepo is available.
	var databaseID string
	if s.dbRepo != nil {
		db, err := s.dbRepo.GetByName(ctx, orgName, projectSlug, cmd.DatabaseName)
		if err != nil {
			return nil, err
		}
		databaseID = db.ID
	}

	staleBefore := s.now().Add(-defaultStalePeriod)
	if s.dbRepo != nil {
		active, err := s.syncJobRepo.GetActiveByDatabase(ctx, orgName, projectSlug, databaseID, staleBefore)
		if err != nil {
			return nil, err
		}
		if active != nil {
			return nil, bizerrors.NewError(bizerrors.Conflict, "sync job already running for database "+cmd.DatabaseName)
		}
	} else {
		// Degraded mode: check by database name (backward compat).
		active, err := s.syncJobRepo.GetActiveByDatabase(ctx, orgName, projectSlug, cmd.DatabaseName, staleBefore)
		if err != nil {
			return nil, err
		}
		if active != nil {
			return nil, bizerrors.NewError(bizerrors.Conflict, "sync job already running for database "+cmd.DatabaseName)
		}
	}

	tableNames := cmd.TableNames
	if tableNames == nil {
		tableNames = []string{}
	}
	now := s.now()
	job := &domaindb.ModelSyncJob{
		ID:           uuid.NewString(),
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseID:   databaseID,
		DatabaseName: cmd.DatabaseName,
		TableNames:   tableNames,
		Status:       domaindb.ModelSyncJobStatusPending,
		FailedTables: []domaindb.ModelSyncFailedTable{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.syncJobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	s.runner.Go(ctx, func(runCtx context.Context) {
		if err := s.RunSyncJob(runCtx, job.ID); err != nil {
			logfacade.GetLogger(runCtx).Error(
				runCtx,
				"model sync job failed",
				logfacade.String("job_id", job.ID),
				logfacade.Err(err),
			)
		}
	})
	return job, nil
}

// GetJob returns the current state of a sync job.
func (s *SyncModelsAppService) GetJob(ctx context.Context, jobID string) (*domaindb.ModelSyncJob, error) {
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

// StartModelSync starts one sync job per target. All jobs share the same batchId.
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
		// Look up database name
		db, dberr := s.dbRepo.GetByID(ctx, orgName, projectSlug, target.DatabaseID)
		if dberr != nil {
			return "", nil, dberr
		}

		// Active job check
		active, aerr := s.syncJobRepo.GetActiveByDatabase(ctx, orgName, projectSlug, target.DatabaseID, staleBefore)
		if aerr != nil {
			return "", nil, aerr
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
		if cerr := s.syncJobRepo.Create(ctx, job); cerr != nil {
			return "", nil, cerr
		}
		if s.dbRepo != nil {
			if uerr := s.dbRepo.UpdateLatestSyncJobID(ctx, orgName, projectSlug, target.DatabaseID, job.ID); uerr != nil {
				logfacade.GetLogger(ctx).Warn(
					ctx, "failed to update latest_sync_job_id",
					logfacade.String("database_id", target.DatabaseID),
					logfacade.String("job_id", job.ID),
					logfacade.Err(uerr),
				)
			}
		}
		jobs = append(jobs, job)
	}

	for _, job := range jobs {
		job := job
		s.runner.Go(ctx, func(runCtx context.Context) {
			if runErr := s.RunSyncJob(runCtx, job.ID); runErr != nil {
				logfacade.GetLogger(runCtx).Error(
					runCtx, "model sync job failed",
					logfacade.String("job_id", job.ID),
					logfacade.Err(runErr),
				)
			}
		})
	}
	return batchID, jobs, nil
}

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

// RecoverStaleJobs marks all stale pending/running jobs as failed.
func (s *SyncModelsAppService) RecoverStaleJobs(ctx context.Context) error {
	staleBefore := s.now().Add(-defaultStalePeriod)
	return s.syncJobRepo.FailStalePendingJobs(ctx, staleBefore)
}

// RunSyncJob executes the sync job in the background goroutine.
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

	// Ensure import group
	var group *domainmodel.ModelGroup
	if s.groupService != nil {
		group, err = s.groupService.EnsureImportGroup(ctx, orgName, projectSlug)
		if err != nil {
			logger.Error(
				ctx, "model sync job: EnsureImportGroup failed",
				logfacade.String("job_id", jobID),
				logfacade.Err(err),
			)
			return s.failJob(ctx, job, err)
		}
	}

	// Determine which tables to process
	var tableNames []string
	if len(job.TableNames) > 0 {
		tableNames = job.TableNames
	} else {
		tableResult, err := s.reverseEngineer.ListTables(ctx, orgName, projectSlug, job.DatabaseName, false, 0, 0)
		if err != nil {
			logger.Error(
				ctx,
				"model sync job: ListTables failed",
				logfacade.String("job_id", jobID),
				logfacade.Err(err),
			)
			return s.failJob(ctx, job, err)
		}
		tableNames = tableResult.Tables
	}

	job.TotalTables = len(tableNames)
	job.UpdatedAt = s.now()
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		return err
	}

	for _, tableName := range tableNames {
		if err := s.processTable(ctx, job, tableName, group); err != nil {
			logger.Error(
				ctx,
				"model sync job: processTable fatal",
				logfacade.String("job_id", jobID),
				logfacade.String("table", tableName),
				logfacade.Err(err),
			)
			return err
		}
	}

	finishedAt := s.now()
	job.FinishedAt = &finishedAt
	job.UpdatedAt = finishedAt

	switch {
	case job.FailedCount == 0:
		job.Status = domaindb.ModelSyncJobStatusSucceeded
	default:
		job.Status = domaindb.ModelSyncJobStatusFailed
	}
	return s.syncJobRepo.Update(ctx, job)
}

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

func (s *SyncModelsAppService) recordTableFailure(
	ctx context.Context,
	job *domaindb.ModelSyncJob,
	tableName string,
	err error,
) error {
	job.ProcessedTables++
	job.FailedCount++
	job.FailedTables = append(job.FailedTables, domaindb.ModelSyncFailedTable{
		TableName: tableName,
		Message:   err.Error(),
	})
	job.UpdatedAt = s.now()
	return s.syncJobRepo.Update(ctx, job)
}

func (s *SyncModelsAppService) failJob(
	ctx context.Context,
	job *domaindb.ModelSyncJob,
	err error,
) error {
	logfacade.GetLogger(ctx).Error(
		ctx,
		"model sync job failed",
		logfacade.String("job_id", job.ID),
		logfacade.String("database_name", job.DatabaseName),
		logfacade.Err(err),
	)
	now := s.now()
	job.Status = domaindb.ModelSyncJobStatusFailed
	job.FinishedAt = &now
	job.UpdatedAt = now
	if updateErr := s.syncJobRepo.Update(ctx, job); updateErr != nil {
		return updateErr
	}
	return err
}

// validateSyncModelsCommand enforces the tableNames/syncAll mutual exclusion.
func validateSyncModelsCommand(cmd SyncModelsFromDBCommand) error {
	hasTableNames := len(cmd.TableNames) > 0
	if hasTableNames && cmd.SyncAll {
		return bizerrors.NewError(bizerrors.ParamInvalid, "cannot specify both tableNames and syncAll")
	}
	if !hasTableNames && !cmd.SyncAll {
		return bizerrors.NewError(bizerrors.ParamInvalid, "must specify either tableNames or syncAll=true")
	}
	return nil
}
