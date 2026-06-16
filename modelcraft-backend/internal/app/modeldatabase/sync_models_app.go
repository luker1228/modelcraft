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
	ReverseEngineer syncModelsReverseEngineer
	ModelRepo       syncModelsModelRepo
	FieldSyncer     syncModelsFieldSyncer
	Runner          modelDatabaseSyncRunner
	Now             func() time.Time
}

// SyncModelsAppService orchestrates syncModelsFromDB async jobs.
type SyncModelsAppService struct {
	syncJobRepo     domaindb.ModelSyncJobRepository
	reverseEngineer syncModelsReverseEngineer
	modelRepo       syncModelsModelRepo
	fieldSyncer     syncModelsFieldSyncer
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
		reverseEngineer: deps.ReverseEngineer,
		modelRepo:       deps.ModelRepo,
		fieldSyncer:     deps.FieldSyncer,
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

	staleBefore := s.now().Add(-defaultStalePeriod)
	active, err := s.syncJobRepo.GetActiveByDatabase(ctx, orgName, projectSlug, cmd.DatabaseName, staleBefore)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, bizerrors.NewError(bizerrors.Conflict, "sync job already running for database "+cmd.DatabaseName)
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
		if err := s.processTable(ctx, job, tableName); err != nil {
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
	case job.CreatedModels > 0 || job.SyncedModels > 0:
		job.Status = domaindb.ModelSyncJobStatusPartialSuccess
	default:
		job.Status = domaindb.ModelSyncJobStatusFailed
	}
	return s.syncJobRepo.Update(ctx, job)
}

func (s *SyncModelsAppService) processTable(
	ctx context.Context,
	job *domaindb.ModelSyncJob,
	tableName string,
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
		_, importErr := s.reverseEngineer.ImportModel(ctx, appmodeldesign.ImportModelCommand{
			OrgName:      orgName,
			ProjectSlug:  projectSlug,
			DatabaseName: databaseName,
			TableName:    tableName,
		})
		if importErr != nil {
			return s.recordTableFailure(ctx, job, tableName, importErr)
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
