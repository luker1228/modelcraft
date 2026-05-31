package modeldatabase

import (
	"context"
	"errors"
	"modelcraft/internal/app/modeldesign"
	domaindb "modelcraft/internal/domain/modeldatabase"
	domainmodel "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/ctxutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeModelDatabaseRepo struct {
	byID map[string]*domaindb.ModelDatabase
}

func (f *fakeModelDatabaseRepo) Create(context.Context, *domaindb.ModelDatabase) error { return nil }
func (f *fakeModelDatabaseRepo) GetByID(_ context.Context, orgName, projectSlug, id string) (*domaindb.ModelDatabase, error) {
	db := f.byID[id]
	if db == nil || db.OrgName != orgName || db.ProjectSlug != projectSlug {
		return nil, shared.NewNotFoundError("model database not found")
	}
	return db, nil
}

func (f *fakeModelDatabaseRepo) GetByName(context.Context, string, string, string) (*domaindb.ModelDatabase, error) {
	return nil, shared.NewNotFoundError("not found")
}

func (f *fakeModelDatabaseRepo) List(context.Context, string, string) ([]*domaindb.ModelDatabase, error) {
	return nil, nil
}

func (f *fakeModelDatabaseRepo) Update(context.Context, string, string, *domaindb.ModelDatabase) error {
	return nil
}
func (f *fakeModelDatabaseRepo) Delete(context.Context, string, string, string) error { return nil }

type fakeSyncJobRepo struct {
	jobs             map[string]*domaindb.ModelDatabaseSyncJob
	runningByDB      map[string]*domaindb.ModelDatabaseSyncJob
	created          []*domaindb.ModelDatabaseSyncJob
	updatedSnapshots []*domaindb.ModelDatabaseSyncJob
}

func newFakeSyncJobRepo() *fakeSyncJobRepo {
	return &fakeSyncJobRepo{
		jobs:        map[string]*domaindb.ModelDatabaseSyncJob{},
		runningByDB: map[string]*domaindb.ModelDatabaseSyncJob{},
	}
}

func (f *fakeSyncJobRepo) Create(ctx context.Context, job *domaindb.ModelDatabaseSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	f.created = append(f.created, &cloned)
	if job.Status == domaindb.ModelDatabaseSyncJobStatusPending || job.Status == domaindb.ModelDatabaseSyncJobStatusRunning {
		f.runningByDB[job.DatabaseID] = &cloned
	}
	return nil
}

func (f *fakeSyncJobRepo) GetByID(ctx context.Context, orgName, projectSlug, jobID string) (*domaindb.ModelDatabaseSyncJob, error) {
	job := f.jobs[jobID]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, shared.NewNotFoundError("sync job not found")
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncJobRepo) GetActiveByDatabase(
	ctx context.Context,
	orgName, projectSlug, databaseID string,
	staleBefore time.Time,
) (*domaindb.ModelDatabaseSyncJob, error) {
	job := f.runningByDB[databaseID]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, nil
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncJobRepo) FailStalePendingJobs(_ context.Context, _ time.Time) error {
	return nil
}

func (f *fakeSyncJobRepo) Update(ctx context.Context, job *domaindb.ModelDatabaseSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	if job.Status == domaindb.ModelDatabaseSyncJobStatusPending || job.Status == domaindb.ModelDatabaseSyncJobStatusRunning {
		f.runningByDB[job.DatabaseID] = &cloned
	} else {
		delete(f.runningByDB, job.DatabaseID)
	}
	f.updatedSnapshots = append(f.updatedSnapshots, &cloned)
	return nil
}

type fakeBackgroundRunner struct {
	run func(context.Context, func(context.Context))
}

func (f *fakeBackgroundRunner) Go(ctx context.Context, fn func(context.Context)) {
	if f.run != nil {
		f.run(ctx, fn)
	}
}

type fakeReverseEngineerService struct {
	tables        []string
	listErr       error
	schemas       map[string]tableSchemaResult
	importResults map[string]*modeldesign.ImportModelResult
	importErrs    map[string]error
}

type tableSchemaResult struct {
	ModelName  string
	SchemaJSON string
	Err        error
}

func (f *fakeReverseEngineerService) ListTables(
	ctx context.Context,
	orgName, projectSlug, databaseName string,
	excludeExisting bool,
	limit, offset int,
) (*modeldesign.ListTablesResult, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &modeldesign.ListTablesResult{Tables: f.tables, TotalCount: len(f.tables)}, nil
}

func (f *fakeReverseEngineerService) ImportModel(ctx context.Context, cmd modeldesign.ImportModelCommand) (*modeldesign.ImportModelResult, error) {
	if err := f.importErrs[cmd.TableName]; err != nil {
		return nil, err
	}
	return f.importResults[cmd.TableName], nil
}

func (f *fakeReverseEngineerService) BuildSchemaForTable(
	ctx context.Context,
	cmd modeldesign.ImportModelCommand,
) (*modeldesign.TableSchemaBuildResult, error) {
	result := f.schemas[cmd.TableName]
	if result.Err != nil {
		return nil, result.Err
	}
	return &modeldesign.TableSchemaBuildResult{
		ModelName:  result.ModelName,
		SchemaJSON: result.SchemaJSON,
	}, nil
}

type fakeModelRepository struct {
	modelsByName map[string]*domainmodel.DataModel
}

func (f *fakeModelRepository) Save(context.Context, string, *domainmodel.DataModel) error { return nil }
func (f *fakeModelRepository) Update(context.Context, *domainmodel.DataModel) error       { return nil }
func (f *fakeModelRepository) UpdateWithVersion(context.Context, *domainmodel.DataModel, int64) (int64, error) {
	return 0, nil
}
func (f *fakeModelRepository) Delete(context.Context, string) error { return nil }
func (f *fakeModelRepository) GetByID(context.Context, string, ...*domainmodel.ModelQueryOptions) (*domainmodel.DataModel, error) {
	return nil, shared.NewNotFoundError("not found")
}

func (f *fakeModelRepository) GetByName(
	ctx context.Context,
	orgName, databaseName, name, projectID string,
	opts ...*domainmodel.ModelQueryOptions,
) (*domainmodel.DataModel, error) {
	key := databaseName + ":" + name
	model := f.modelsByName[key]
	if model == nil {
		return nil, shared.NewNotFoundError("not found")
	}
	return model, nil
}

func (f *fakeModelRepository) FindByDeploymentStatus(context.Context, ...domainmodel.DeploymentStatus) ([]domainmodel.DataModel, error) {
	return nil, nil
}

func (f *fakeModelRepository) GetMetaByIDs(context.Context, string, string, []string) ([]*domainmodel.DataModel, error) {
	return nil, nil
}

func (f *fakeModelRepository) Query(context.Context, domainmodel.ModelQuery) ([]domainmodel.DataModel, int, error) {
	return nil, 0, nil
}

func (f *fakeModelRepository) ListDatabaseCatalog(context.Context, string, string, string, int, int) ([]string, int, error) {
	return nil, 0, nil
}

func (f *fakeModelRepository) AddFields(context.Context, string, []*domainmodel.FieldDefinition) error {
	return nil
}

func (f *fakeModelRepository) AddRelationField(context.Context, string, *domainmodel.FieldDefinition) error {
	return nil
}

func (f *fakeModelRepository) GetFieldByModelID(context.Context, string, string) (*domainmodel.FieldDefinition, error) {
	return nil, nil
}

func (f *fakeModelRepository) GetFieldsByModelID(context.Context, string) ([]*domainmodel.FieldDefinition, error) {
	return nil, nil
}

func (f *fakeModelRepository) GetTailFieldDisplayOrder(context.Context, string) (string, error) {
	return "", nil
}

func (f *fakeModelRepository) UpdateField(context.Context, *domainmodel.FieldDefinition) error {
	return nil
}

func (f *fakeModelRepository) BulkUpdateFields(context.Context, []*domainmodel.FieldDefinition) error {
	return nil
}

func (f *fakeModelRepository) UpdateFieldsStatus(context.Context, ...domainmodel.UpdateFieldsStatusRequest) error {
	return nil
}
func (f *fakeModelRepository) DeleteFields(context.Context, string, []string) error { return nil }
func (f *fakeModelRepository) BulkDeleteFields(context.Context, ...domainmodel.DeleteFieldRequest) error {
	return nil
}

type fakeSchemaSyncService struct {
	errs  map[string]error
	calls []string
}

func (f *fakeSchemaSyncService) SyncModelSchemaFromJSON(
	ctx context.Context,
	modelID string,
	schemaJSON string,
	deleteExtraFields bool,
) (*modeldesign.SyncModelSchemaResult, error) {
	f.calls = append(f.calls, modelID)
	if err := f.errs[modelID]; err != nil {
		return nil, err
	}
	return &modeldesign.SyncModelSchemaResult{}, nil
}

type fakeGroupService struct {
	group      *domainmodel.ModelGroup
	ensureErr  error
	moveErrs   map[string]error
	moveCalls  []string
	ensureCall int
}

func (f *fakeGroupService) EnsureImportGroup(ctx context.Context, orgName, projectSlug string) (*domainmodel.ModelGroup, error) {
	f.ensureCall++
	if f.ensureErr != nil {
		return nil, f.ensureErr
	}
	return f.group, nil
}

func (f *fakeGroupService) MoveModelToGroup(ctx context.Context, modelID string, groupID *string) error {
	f.moveCalls = append(f.moveCalls, modelID)
	if groupID != nil {
		if err := f.moveErrs[*groupID+":"+modelID]; err != nil {
			return err
		}
	}
	return nil
}

func projectContext() context.Context {
	return ctxutils.SetProjectSlug(ctxutils.SetOrgName(context.Background(), "org-a"), "proj-a")
}

func TestStartModelDatabaseSync_RejectsExistingActiveJob(t *testing.T) {
	ctx := projectContext()

	dbRepo := &fakeModelDatabaseRepo{byID: map[string]*domaindb.ModelDatabase{
		"db-1": {ID: "db-1", OrgName: "org-a", ProjectSlug: "proj-a", Name: "orders"},
	}}
	jobRepo := newFakeSyncJobRepo()
	jobRepo.runningByDB["db-1"] = &domaindb.ModelDatabaseSyncJob{
		ID:          "job-1",
		OrgName:     "org-a",
		ProjectSlug: "proj-a",
		DatabaseID:  "db-1",
		Status:      domaindb.ModelDatabaseSyncJobStatusRunning,
	}

	svc := NewModelDatabaseSyncAppService(ModelDatabaseSyncAppServiceDeps{
		ModelDatabaseRepo: dbRepo,
		SyncJobRepo:       jobRepo,
		Runner:            &fakeBackgroundRunner{},
	})

	_, err := svc.StartSync(ctx, "db-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already")
}

func TestRunSyncJob_ContinuesAfterPerTableFailures(t *testing.T) {
	ctx := projectContext()
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	dbRepo := &fakeModelDatabaseRepo{byID: map[string]*domaindb.ModelDatabase{
		"db-1": {ID: "db-1", OrgName: "org-a", ProjectSlug: "proj-a", Name: "orders"},
	}}
	jobRepo := newFakeSyncJobRepo()
	reverseSvc := &fakeReverseEngineerService{
		tables: []string{"new_table", "existing_table", "broken_table"},
		schemas: map[string]tableSchemaResult{
			"new_table":      {ModelName: "new_table", SchemaJSON: `{"type":"object"}`},
			"existing_table": {ModelName: "existing_table", SchemaJSON: `{"type":"object"}`},
			"broken_table":   {Err: errors.New("introspect failed")},
		},
		importResults: map[string]*modeldesign.ImportModelResult{
			"new_table": {ModelID: "model-new", ModelName: "new_table", FieldsCount: 2},
		},
		importErrs: map[string]error{},
	}
	modelRepo := &fakeModelRepository{
		modelsByName: map[string]*domainmodel.DataModel{
			"orders:existing_table": {
				ModelMeta: domainmodel.ModelMeta{
					ID: "model-existing",
					ModelLocator: domainmodel.ModelLocator{
						ProjectScope: project.ProjectScope{OrgName: "org-a", ProjectSlug: "proj-a"},
						DatabaseName: "orders",
						ModelName:    "existing_table",
					},
				},
			},
		},
	}
	syncSvc := &fakeSchemaSyncService{}
	groupSvc := &fakeGroupService{
		group:    &domainmodel.ModelGroup{ID: "group-1", Name: "数据库导入"},
		moveErrs: map[string]error{},
	}

	svc := NewModelDatabaseSyncAppService(ModelDatabaseSyncAppServiceDeps{
		ModelDatabaseRepo: dbRepo,
		SyncJobRepo:       jobRepo,
		ReverseEngineer:   reverseSvc,
		ModelRepo:         modelRepo,
		SchemaSync:        syncSvc,
		GroupService:      groupSvc,
		Runner:            &fakeBackgroundRunner{},
		Now:               func() time.Time { return now },
	})

	job := &domaindb.ModelDatabaseSyncJob{
		ID:          "job-1",
		OrgName:     "org-a",
		ProjectSlug: "proj-a",
		DatabaseID:  "db-1",
		Status:      domaindb.ModelDatabaseSyncJobStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, jobRepo.Create(ctx, job))

	err := svc.RunSyncJob(ctx, "job-1")
	require.NoError(t, err)

	saved, err := jobRepo.GetByID(ctx, "org-a", "proj-a", "job-1")
	require.NoError(t, err)
	assert.Equal(t, domaindb.ModelDatabaseSyncJobStatusPartialSuccess, saved.Status)
	assert.Equal(t, 3, saved.TotalTables)
	assert.Equal(t, 3, saved.ProcessedTables)
	assert.Equal(t, 1, saved.CreatedModels)
	assert.Equal(t, 1, saved.SyncedModels)
	assert.Equal(t, 1, saved.FailedCount)
	require.Len(t, saved.FailedTables, 1)
	assert.Equal(t, "broken_table", saved.FailedTables[0].TableName)
	assert.Equal(t, []string{"model-new"}, groupSvc.moveCalls)
	assert.Equal(t, []string{"model-existing"}, syncSvc.calls)
}
