package modeldatabase

import (
	"context"
	"errors"
	"testing"
	"time"

	appmodeldesign "modelcraft/internal/app/modeldesign"
	domaindb "modelcraft/internal/domain/modeldatabase"
	domainmodel "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/ctxutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeSyncModelsJobRepo struct {
	jobs       map[string]*domaindb.ModelSyncJob
	activeByDB map[string]*domaindb.ModelSyncJob // key: databaseName
	snapshots  []*domaindb.ModelSyncJob
}

func newFakeSyncModelsJobRepo() *fakeSyncModelsJobRepo {
	return &fakeSyncModelsJobRepo{
		jobs:       make(map[string]*domaindb.ModelSyncJob),
		activeByDB: make(map[string]*domaindb.ModelSyncJob),
	}
}

func (f *fakeSyncModelsJobRepo) Create(_ context.Context, job *domaindb.ModelSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	if job.Status == domaindb.ModelSyncJobStatusPending || job.Status == domaindb.ModelSyncJobStatusRunning {
		f.activeByDB[job.DatabaseName] = &cloned
	}
	return nil
}

func (f *fakeSyncModelsJobRepo) GetByID(
	_ context.Context, orgName, projectSlug, jobID string,
) (*domaindb.ModelSyncJob, error) {
	job := f.jobs[jobID]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, shared.NewNotFoundError("sync job not found")
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncModelsJobRepo) GetActiveByDatabase(
	_ context.Context, orgName, projectSlug, databaseName string, _ time.Time,
) (*domaindb.ModelSyncJob, error) {
	job := f.activeByDB[databaseName]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, nil
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncModelsJobRepo) Update(_ context.Context, job *domaindb.ModelSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	f.snapshots = append(f.snapshots, &cloned)
	if job.Status == domaindb.ModelSyncJobStatusPending || job.Status == domaindb.ModelSyncJobStatusRunning {
		f.activeByDB[job.DatabaseName] = &cloned
	} else {
		delete(f.activeByDB, job.DatabaseName)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────

type fakeSyncModelsReverseEngineer struct {
	tableDefs  map[string]*appmodeldesign.TableDefinitionResult // key: tableName
	tableErrs  map[string]error
	listTables []string
	listErr    error
	importErrs map[string]error
}

func (f *fakeSyncModelsReverseEngineer) ListTables(
	_ context.Context, _, _, _ string, _ bool, _, _ int,
) (*appmodeldesign.ListTablesResult, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &appmodeldesign.ListTablesResult{Tables: f.listTables, TotalCount: len(f.listTables)}, nil
}

func (f *fakeSyncModelsReverseEngineer) ImportModel(
	_ context.Context, cmd appmodeldesign.ImportModelCommand,
) (*appmodeldesign.ImportModelResult, error) {
	if f.importErrs != nil {
		if err := f.importErrs[cmd.TableName]; err != nil {
			return nil, err
		}
	}
	return &appmodeldesign.ImportModelResult{ModelID: "model-" + cmd.TableName, ModelName: cmd.TableName}, nil
}

func (f *fakeSyncModelsReverseEngineer) GetTableDefinition(
	_ context.Context, _, _, _, tableName string,
) (*appmodeldesign.TableDefinitionResult, error) {
	if f.tableErrs != nil {
		if err := f.tableErrs[tableName]; err != nil {
			return nil, err
		}
	}
	if f.tableDefs != nil {
		if def, ok := f.tableDefs[tableName]; ok {
			return def, nil
		}
	}
	return &appmodeldesign.TableDefinitionResult{}, nil
}

// ─────────────────────────────────────────────────────────────────────────────

type fakeSyncModelsModelRepo struct {
	modelsByName map[string]*domainmodel.DataModel // key: databaseName:modelName
}

func (f *fakeSyncModelsModelRepo) GetByName(
	_ context.Context, _, databaseName, name, _ string,
	_ ...*domainmodel.ModelQueryOptions,
) (*domainmodel.DataModel, error) {
	key := databaseName + ":" + name
	m := f.modelsByName[key]
	if m == nil {
		return nil, shared.NewNotFoundError("model not found")
	}
	return m, nil
}

// ─────────────────────────────────────────────────────────────────────────────

type fakeFieldSyncer struct {
	errs      map[string]error
	syncCalls []string // modelIDs called
}

func (f *fakeFieldSyncer) SyncFieldsFromDB(
	_ context.Context, modelID string, _ []*domainmodel.FieldDefinition,
) error {
	f.syncCalls = append(f.syncCalls, modelID)
	if f.errs != nil {
		if err := f.errs[modelID]; err != nil {
			return err
		}
	}
	return nil
}

// ── helper ───────────────────────────────────────────────────────────────────

func syncModelsProjectCtx() context.Context {
	return ctxutils.SetProjectSlug(ctxutils.SetOrgName(context.Background(), "org-a"), "proj-a")
}

func syncNow() time.Time {
	return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
}

// syncRunner runs the background func synchronously in tests.
func syncRunner() *fakeBackgroundRunner {
	return &fakeBackgroundRunner{run: func(ctx context.Context, fn func(context.Context)) { fn(ctx) }}
}

func newSyncModelsService(
	jobRepo *fakeSyncModelsJobRepo,
	re *fakeSyncModelsReverseEngineer,
	modelRepo *fakeSyncModelsModelRepo,
	fieldSyncer *fakeFieldSyncer,
) *SyncModelsAppService {
	return NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo:     jobRepo,
		ReverseEngineer: re,
		ModelRepo:       modelRepo,
		FieldSyncer:     fieldSyncer,
		Runner:          syncRunner(),
		Now:             syncNow,
	})
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestSyncModels_RejectsActiveJob(t *testing.T) {
	ctx := syncModelsProjectCtx()
	jobRepo := newFakeSyncModelsJobRepo()
	// Pre-seed an active job for the database "mydb"
	jobRepo.activeByDB["mydb"] = &domaindb.ModelSyncJob{
		ID:           "job-existing",
		OrgName:      "org-a",
		ProjectSlug:  "proj-a",
		DatabaseName: "mydb",
		Status:       domaindb.ModelSyncJobStatusRunning,
	}

	svc := newSyncModelsService(
		jobRepo,
		&fakeSyncModelsReverseEngineer{},
		&fakeSyncModelsModelRepo{modelsByName: map[string]*domainmodel.DataModel{}},
		&fakeFieldSyncer{},
	)

	_, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "mydb",
		SyncAll:      true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestSyncModels_RejectsNoTablesAndNoSyncAll(t *testing.T) {
	ctx := syncModelsProjectCtx()
	svc := newSyncModelsService(
		newFakeSyncModelsJobRepo(),
		&fakeSyncModelsReverseEngineer{},
		&fakeSyncModelsModelRepo{modelsByName: map[string]*domainmodel.DataModel{}},
		&fakeFieldSyncer{},
	)

	_, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "mydb",
		// Neither TableNames nor SyncAll
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must specify")
}

func TestSyncModels_RejectsBothTablesAndSyncAll(t *testing.T) {
	ctx := syncModelsProjectCtx()
	svc := newSyncModelsService(
		newFakeSyncModelsJobRepo(),
		&fakeSyncModelsReverseEngineer{},
		&fakeSyncModelsModelRepo{modelsByName: map[string]*domainmodel.DataModel{}},
		&fakeFieldSyncer{},
	)

	_, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "mydb",
		TableNames:   []string{"orders"},
		SyncAll:      true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot specify both")
}

func TestSyncModels_CreatesNewModels(t *testing.T) {
	ctx := syncModelsProjectCtx()
	jobRepo := newFakeSyncModelsJobRepo()

	re := &fakeSyncModelsReverseEngineer{
		tableDefs: map[string]*appmodeldesign.TableDefinitionResult{
			"orders": {Fields: []*domainmodel.FieldDefinition{}},
		},
		importErrs: map[string]error{},
	}
	// No existing model for "orders"
	modelRepo := &fakeSyncModelsModelRepo{modelsByName: map[string]*domainmodel.DataModel{}}
	fieldSyncer := &fakeFieldSyncer{}

	svc := newSyncModelsService(jobRepo, re, modelRepo, fieldSyncer)

	job, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "mydb",
		TableNames:   []string{"orders"},
	})
	require.NoError(t, err)
	require.NotNil(t, job)

	saved, err := jobRepo.GetByID(ctx, "org-a", "proj-a", job.ID)
	require.NoError(t, err)

	assert.Equal(t, domaindb.ModelSyncJobStatusSucceeded, saved.Status)
	assert.Equal(t, 1, saved.CreatedModels)
	assert.Equal(t, 0, saved.SyncedModels)
	assert.Equal(t, 0, saved.FailedCount)
}

func TestSyncModels_SyncsExistingModel(t *testing.T) {
	ctx := syncModelsProjectCtx()
	jobRepo := newFakeSyncModelsJobRepo()

	re := &fakeSyncModelsReverseEngineer{
		tableDefs: map[string]*appmodeldesign.TableDefinitionResult{
			"orders": {Fields: []*domainmodel.FieldDefinition{}},
		},
		importErrs: map[string]error{},
	}
	// "orders" table normalizes to "orders"; key = "mydb:orders"
	existingModel := &domainmodel.DataModel{
		ModelMeta: domainmodel.ModelMeta{
			ID: "model-orders-123",
		},
	}
	modelRepo := &fakeSyncModelsModelRepo{
		modelsByName: map[string]*domainmodel.DataModel{
			"mydb:orders": existingModel,
		},
	}
	fieldSyncer := &fakeFieldSyncer{}

	svc := newSyncModelsService(jobRepo, re, modelRepo, fieldSyncer)

	job, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "mydb",
		TableNames:   []string{"orders"},
	})
	require.NoError(t, err)
	require.NotNil(t, job)

	saved, err := jobRepo.GetByID(ctx, "org-a", "proj-a", job.ID)
	require.NoError(t, err)

	assert.Equal(t, domaindb.ModelSyncJobStatusSucceeded, saved.Status)
	assert.Equal(t, 0, saved.CreatedModels)
	assert.Equal(t, 1, saved.SyncedModels)
	assert.Equal(t, 0, saved.FailedCount)

	// fieldSyncer must have been called with the correct modelID
	require.Len(t, fieldSyncer.syncCalls, 1)
	assert.Equal(t, "model-orders-123", fieldSyncer.syncCalls[0])
}

func TestSyncModels_SingleTableFailureContinues(t *testing.T) {
	ctx := syncModelsProjectCtx()
	jobRepo := newFakeSyncModelsJobRepo()

	// "broken_table" fails introspection; "good_table" succeeds as a new model
	re := &fakeSyncModelsReverseEngineer{
		tableDefs: map[string]*appmodeldesign.TableDefinitionResult{
			"good_table": {Fields: []*domainmodel.FieldDefinition{}},
		},
		tableErrs: map[string]error{
			"broken_table": errors.New("introspect failed"),
		},
		importErrs: map[string]error{},
	}
	modelRepo := &fakeSyncModelsModelRepo{modelsByName: map[string]*domainmodel.DataModel{}}
	fieldSyncer := &fakeFieldSyncer{}

	svc := newSyncModelsService(jobRepo, re, modelRepo, fieldSyncer)

	job, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "mydb",
		TableNames:   []string{"broken_table", "good_table"},
	})
	require.NoError(t, err)
	require.NotNil(t, job)

	saved, err := jobRepo.GetByID(ctx, "org-a", "proj-a", job.ID)
	require.NoError(t, err)

	assert.Equal(t, domaindb.ModelSyncJobStatusPartialSuccess, saved.Status)
	assert.Equal(t, 1, saved.FailedCount)
	assert.Equal(t, 1, saved.CreatedModels)
	require.Len(t, saved.FailedTables, 1)
	assert.Equal(t, "broken_table", saved.FailedTables[0].TableName)
}
