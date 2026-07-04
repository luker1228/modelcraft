package modeldesign

import (
	"context"
	"errors"
	"modelcraft/internal/domain/cluster"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// MockModelRepository is a mock for modeldesign.ModelRepository
type MockModelRepository struct {
	mock.Mock
}

func (m *MockModelRepository) Save(ctx context.Context, orgName string, model *modeldesign.DataModel) error {
	args := m.Called(ctx, orgName, model)
	return args.Error(0)
}

func (m *MockModelRepository) Update(ctx context.Context, model *modeldesign.DataModel) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *MockModelRepository) UpdateWithVersion(
	ctx context.Context,
	model *modeldesign.DataModel,
	originalVersion int64,
) (int64, error) {
	args := m.Called(ctx, model, originalVersion)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockModelRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockModelRepository) GetByID(
	ctx context.Context,
	id string,
	opts ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	args := m.Called(ctx, id, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.DataModel), args.Error(1)
}

func (m *MockModelRepository) GetByName(
	ctx context.Context,
	orgName string,
	databaseName string,
	name string,
	projectId string,
	opts ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	args := m.Called(ctx, orgName, databaseName, name, projectId, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.DataModel), args.Error(1)
}

func (m *MockModelRepository) FindByDeploymentStatus(
	ctx context.Context,
	statuses ...modeldesign.DeploymentStatus,
) ([]modeldesign.DataModel, error) {
	args := m.Called(ctx, statuses)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]modeldesign.DataModel), args.Error(1)
}

func (m *MockModelRepository) GetMetaByIDs(
	ctx context.Context,
	orgName, projectSlug string,
	ids []string,
) ([]*modeldesign.DataModel, error) {
	args := m.Called(ctx, orgName, projectSlug, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.DataModel), args.Error(1)
}

func (m *MockModelRepository) Query(
	ctx context.Context,
	queryObj modeldesign.ModelQuery,
) ([]modeldesign.DataModel, int, error) {
	args := m.Called(ctx, queryObj)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]modeldesign.DataModel), args.Int(1), args.Error(2)
}

func (m *MockModelRepository) ListDatabaseCatalog(
	ctx context.Context,
	orgName, projectSlug, search string,
	page, pageSize int,
) ([]string, int, error) {
	args := m.Called(ctx, orgName, projectSlug, search, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]string), args.Int(1), args.Error(2)
}

func (m *MockModelRepository) AddFields(
	ctx context.Context, orgName string, field []*modeldesign.FieldDefinition,
) error {
	args := m.Called(ctx, orgName, field)
	return args.Error(0)
}

func (m *MockModelRepository) AddRelationField(
	ctx context.Context, orgName string, field *modeldesign.FieldDefinition,
) error {
	args := m.Called(ctx, orgName, field)
	return args.Error(0)
}

func (m *MockModelRepository) GetFieldByModelID(
	ctx context.Context,
	modelID string,
	name string,
) (*modeldesign.FieldDefinition, error) {
	args := m.Called(ctx, modelID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.FieldDefinition), args.Error(1)
}

func (m *MockModelRepository) GetFieldsByModelID(
	ctx context.Context,
	modelID string,
) ([]*modeldesign.FieldDefinition, error) {
	args := m.Called(ctx, modelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.FieldDefinition), args.Error(1)
}

func (m *MockModelRepository) UpdateField(ctx context.Context, field *modeldesign.FieldDefinition) error {
	args := m.Called(ctx, field)
	return args.Error(0)
}

func (m *MockModelRepository) BulkUpdateFields(ctx context.Context, field []*modeldesign.FieldDefinition) error {
	args := m.Called(ctx, field)
	return args.Error(0)
}

func (m *MockModelRepository) UpdateFieldsStatus(
	ctx context.Context,
	requests ...modeldesign.UpdateFieldsStatusRequest,
) error {
	args := m.Called(ctx, requests)
	return args.Error(0)
}

func (m *MockModelRepository) DeleteFields(ctx context.Context, modelID string, names []string) error {
	args := m.Called(ctx, modelID, names)
	return args.Error(0)
}

func (m *MockModelRepository) BulkDeleteFields(ctx context.Context, requests ...modeldesign.DeleteFieldRequest) error {
	args := m.Called(ctx, requests)
	return args.Error(0)
}

func (m *MockModelRepository) GetTailFieldDisplayOrder(ctx context.Context, modelID string) (string, error) {
	args := m.Called(ctx, modelID)
	return args.String(0), args.Error(1)
}

// MockDeployRepo is a mock for modeldesign.DeployRepo
type MockDeployRepo struct {
	mock.Mock
}

func (m *MockDeployRepo) DeployModelToCreate(ctx context.Context, dataModel *modeldesign.DataModel) error {
	args := m.Called(ctx, dataModel)
	return args.Error(0)
}

func (m *MockDeployRepo) DeployModelToDrop(ctx context.Context, dataModel *modeldesign.DataModel) error {
	args := m.Called(ctx, dataModel)
	return args.Error(0)
}

func (m *MockDeployRepo) DeployModelToAddFields(
	ctx context.Context,
	dataModel *modeldesign.DataModel,
	addFields []*modeldesign.FieldDefinition,
) error {
	args := m.Called(ctx, dataModel, addFields)
	return args.Error(0)
}

func (m *MockDeployRepo) DeployModelToRemoveFields(
	ctx context.Context,
	dataModel *modeldesign.DataModel,
	fieldKeys []string,
) error {
	args := m.Called(ctx, dataModel, fieldKeys)
	return args.Error(0)
}

func (m *MockDeployRepo) CheckTableExists(ctx context.Context, dataModel *modeldesign.DataModel) (bool, error) {
	args := m.Called(ctx, dataModel)
	return args.Bool(0), args.Error(1)
}

// MockClusterRepository is a mock for cluster.DatabaseClusterRepository
type MockClusterRepository struct {
	mock.Mock
}

func (m *MockClusterRepository) Create(ctx context.Context, c *cluster.DatabaseCluster) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockClusterRepository) Update(
	ctx context.Context, orgName, projectSlug string, c *cluster.DatabaseCluster,
) error {
	args := m.Called(ctx, orgName, projectSlug, c)
	return args.Error(0)
}

func (m *MockClusterRepository) GetByID(ctx context.Context, orgName, id string) (*cluster.DatabaseCluster, error) {
	args := m.Called(ctx, orgName, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cluster.DatabaseCluster), args.Error(1)
}

func (m *MockClusterRepository) List(
	ctx context.Context,
	orgName string,
	projectSlug string,
	status ...cluster.ClusterStatus,
) ([]*cluster.DatabaseCluster, error) {
	args := m.Called(ctx, orgName, projectSlug, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cluster.DatabaseCluster), args.Error(1)
}

func (m *MockClusterRepository) Delete(ctx context.Context, orgName, projectSlug, id string) error {
	args := m.Called(ctx, orgName, projectSlug, id)
	return args.Error(0)
}

func (m *MockClusterRepository) GetByProjectKey(
	ctx context.Context,
	orgName string,
	projectSlug string,
) (*cluster.DatabaseCluster, error) {
	args := m.Called(ctx, orgName, projectSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cluster.DatabaseCluster), args.Error(1)
}

func (m *MockClusterRepository) ExistsByProjectKey(ctx context.Context, orgName, projectSlug string) (bool, error) {
	args := m.Called(ctx, orgName, projectSlug)
	return args.Bool(0), args.Error(1)
}

func (m *MockClusterRepository) ListUpdatedAfter(
	ctx context.Context,
	orgName string,
	projectSlug string,
	updatedAfter time.Time,
	status ...cluster.ClusterStatus,
) ([]*cluster.DatabaseCluster, error) {
	args := m.Called(ctx, orgName, projectSlug, updatedAfter, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cluster.DatabaseCluster), args.Error(1)
}

// ============================================================================
// Helper Functions
// ============================================================================

// newTestContext creates a context with the required HTTP request context value
func newTestContext() context.Context {
	ctx := context.Background()
	ctx = ctxutils.SetRequestID(ctx, "test-req-id")
	ctx = ctxutils.SetLang(ctx, "en")
	ctx = ctxutils.SetContextValue(ctx, ctxutils.ContextKeyOrgName, "test-org")
	return ctx
}

// newTestModel creates a DataModel for testing with a system field so Validate() passes
func newTestModel(id, projectSlug, modelName, databaseName string) *modeldesign.DataModel {
	now := time.Now()
	locator, _ := modeldesign.NewModelLocator("test-org", projectSlug, databaseName, modelName)
	fields := modeldesign.GetSystemFields()
	// Set ModelID and ModelLocator on system fields so field validation passes
	for _, f := range fields {
		f.ModelID = id
		f.ModelLocator = locator
		f.CreatedAt = now
		f.UpdatedAt = now
	}
	return &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID:               id,
			ModelLocator:     *locator,
			Title:            modelName,
			Description:      "Test model",
			StorageType:      "mysql",
			Version:          1,
			Status:           "draft",
			DeploymentStatus: modeldesign.DeploymentPending,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		Fields: fields,
	}
}

// newTestService creates a ModelDesignAppService with mocked dependencies (no db)
func newTestService(
	modelRepo modeldesign.ModelRepository,
	deployRepo modeldesign.DeployRepo,
	clusterRepo cluster.DatabaseClusterRepository,
) *ModelDesignAppService {
	return NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:   modelRepo,
		DeployRepo:  deployRepo,
		ClusterRepo: clusterRepo,
	})
}

// ============================================================================
// Tests: UpdateModelMeta (uses UpdateModelMetaCommand)
// ============================================================================

func TestModelDesignAppService_UpdateModelMeta(t *testing.T) {
	ctx := newTestContext()

	t.Run("success - update title and description", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
		mockModelRepo.On("Update", ctx, mock.AnythingOfType("*modeldesign.DataModel")).Return(nil)

		newTitle := "Updated Title"
		newDesc := "Updated Description"
		cmd := UpdateModelMetaCommand{
			ProjectSlug: "project-1",
			Title:       &newTitle,
			Description: &newDesc,
		}

		err := service.UpdateModelMeta(ctx, "model-1", cmd)

		assert.NoError(t, err)
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("model not found", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		mockModelRepo.On("GetByID", ctx, "nonexistent", mock.Anything).Return(nil, nil)

		cmd := UpdateModelMetaCommand{
			ProjectSlug: "project-1",
		}

		err := service.UpdateModelMeta(ctx, "nonexistent", cmd)

		assert.Error(t, err)
		var bizErr *bizerrors.BusinessError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.ModelNotFound.GetCode(), bizErr.Info().GetCode())
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("repo error on get", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		repoErr := shared.NewRepositoryError(shared.ErrTypeUnknown, "db error")
		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(nil, repoErr)

		cmd := UpdateModelMetaCommand{
			ProjectSlug: "project-1",
		}

		err := service.UpdateModelMeta(ctx, "model-1", cmd)

		assert.Error(t, err)
		mockModelRepo.AssertExpectations(t)
	})
}

// ============================================================================
// Tests: DeleteModelSync
// ============================================================================

func TestModelDesignAppService_DeleteModelSync(t *testing.T) {
	ctx := newTestContext()

	t.Run("success", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		mockDeployRepo := new(MockDeployRepo)
		service := newTestService(mockModelRepo, mockDeployRepo, nil)

		existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
		mockDeployRepo.On("DeployModelToDrop", ctx, existingModel).Return(nil)
		mockModelRepo.On("Delete", ctx, "model-1").Return(nil)

		err := service.DeleteModelSync(ctx, "model-1", "project-1", true)

		assert.NoError(t, err)
		mockModelRepo.AssertExpectations(t)
		mockDeployRepo.AssertExpectations(t)
	})

	t.Run("model not found by repo error", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		mockModelRepo.On("GetByID", ctx, "nonexistent", mock.Anything).Return(nil, nil)

		err := service.DeleteModelSync(ctx, "nonexistent", "project-1", false)

		assert.Error(t, err)
		var bizErr *bizerrors.BusinessError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.ModelNotFound.GetCode(), bizErr.Info().GetCode())
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("model nil after get", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(nil, nil)

		err := service.DeleteModelSync(ctx, "model-1", "project-1", false)

		assert.Error(t, err)
		var bizErr *bizerrors.BusinessError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.ModelNotFound.GetCode(), bizErr.Info().GetCode())
		mockModelRepo.AssertExpectations(t)
	})
}

// ============================================================================
// Tests: QueryModelsWithCommand (uses ModelQueryCommand)
// ============================================================================

func TestModelDesignAppService_QueryModelsWithCommand(t *testing.T) {
	ctx := newTestContext()

	t.Run("success - returns models", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		model1 := *newTestModel("model-1", "project-1", "model_a", "db_1")
		model2 := *newTestModel("model-2", "project-1", "model_b", "db_1")
		expectedModels := []modeldesign.DataModel{model1, model2}

		// The command will be converted to a domain query with defaults applied
		mockModelRepo.On("Query", ctx, mock.MatchedBy(func(q modeldesign.ModelQuery) bool {
			return q.ProjectSlug == "project-1" &&
				q.DatabaseName == "db_1" &&
				q.Page == 1 &&
				q.PageSize == 20
		})).Return(expectedModels, 2, nil)

		cmd := ModelQueryCommand{
			ProjectSlug:  "project-1",
			DatabaseName: "db_1",
		}

		models, total, err := service.QueryModelsWithCommand(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, models, 2)
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("success - with custom pagination", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		expectedModels := []modeldesign.DataModel{}
		mockModelRepo.On("Query", ctx, mock.MatchedBy(func(q modeldesign.ModelQuery) bool {
			return q.Page == 2 && q.PageSize == 10
		})).Return(expectedModels, 0, nil)

		cmd := ModelQueryCommand{
			ProjectSlug:  "project-1",
			DatabaseName: "db_1",
			Page:         2,
			PageSize:     10,
		}

		models, total, err := service.QueryModelsWithCommand(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Empty(t, models)
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("validation fails - missing required fields", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		cmd := ModelQueryCommand{
			ProjectSlug: "project-1",
			// Missing DatabaseName
		}

		_, _, err := service.QueryModelsWithCommand(ctx, cmd)

		assert.Error(t, err)
		// Should not call the repo if validation fails
		mockModelRepo.AssertNotCalled(t, "Query")
	})
}

// ============================================================================
// Tests: GetModelByID (uses GetModelOptions)
// ============================================================================

func TestModelDesignAppService_GetModelByID(t *testing.T) {
	ctx := newTestContext()

	t.Run("success", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		expectedModel := newTestModel("model-1", "project-1", "test_model", "db_1")
		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(expectedModel, nil)

		opts := NewGetModelOptions()
		model, err := service.GetModelByID(ctx, "model-1", opts)

		assert.NoError(t, err)
		assert.NotNil(t, model)
		assert.Equal(t, "model-1", model.ID)
		assert.Equal(t, "test_model", model.ModelName)
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("not found - repo error", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		mockModelRepo.On("GetByID", ctx, "nonexistent", mock.Anything).Return(nil, nil)

		opts := NewGetModelOptions()
		model, err := service.GetModelByID(ctx, "nonexistent", opts)

		assert.Error(t, err)
		assert.Nil(t, model)
		var bizErr *bizerrors.BusinessError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.ModelNotFound.GetCode(), bizErr.Info().GetCode())
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("not found - nil result", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(nil, nil)

		opts := NewGetModelOptions()
		model, err := service.GetModelByID(ctx, "model-1", opts)

		assert.Error(t, err)
		assert.Nil(t, model)
		var bizErr *bizerrors.BusinessError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.ModelNotFound.GetCode(), bizErr.Info().GetCode())
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("fills external FK relation metadata for END_USER_REF owner", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		mockFKRepo := new(MockLogicalForeignKeyRepository)
		service := NewModelDesignAppService(ModelDesignAppServiceDeps{
			ModelRepo: mockModelRepo,
			FKRepo:    mockFKRepo,
		})

		ownerFKID := "fk-owner-1"
		ownerType := modeldesign.GetFieldTypeByFormat(modeldesign.FormatEndUserRef)
		ownerField := &modeldesign.FieldDefinition{
			ModelID:       "model-1",
			Name:          "owner",
			Title:         "Owner",
			Type:          ownerType,
			BelongsToFKID: &ownerFKID,
		}

		testModel := newTestModel("model-1", "project-1", "orders", "db_1")
		testModel.Fields = []*modeldesign.FieldDefinition{ownerField}

		mockModelRepo.On("GetByID", mock.Anything, "model-1", mock.Anything).Return(testModel, nil).Once()
		mockFKRepo.On("GetByID", mock.Anything, ownerFKID).Return(&modeldesign.LogicalForeignKey{
			ID:              ownerFKID,
			PairID:          "pair-owner-1",
			OrgName:         "test-org",
			Direction:       modeldesign.DirectionNormal,
			ModelID:         "model-1",
			ModelName:       "orders",
			RefModelID:      "",
			RefModelName:    "end_user_users",
			RefDatabaseName: "mc_meta",
			RefTableName:    "end_user_users",
			SourceFields:    []string{"owner"},
			TargetFields:    []string{"id"},
			IsDeletable:     false,
		}, nil).Once()

		opts := NewGetModelOptions()
		opts.GetFields = true
		model, err := service.GetModelByID(ctx, "model-1", opts)

		require.NoError(t, err)
		require.NotNil(t, model)
		require.Len(t, model.Fields, 1)

		relationRaw, ok := model.Fields[0].Metadata["x-relation"]
		require.True(t, ok)
		relation, ok := relationRaw.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "mc_meta", relation["databaseName"])
		assert.Equal(t, "end_user_users", relation["modelName"])
		assert.Equal(t, "normal", relation["direction"])
		assert.Equal(t, "many-to-one", relation["cardinality"])

		mockFKRepo.AssertExpectations(t)
		mockModelRepo.AssertExpectations(t)
	})
}

// ============================================================================
// Tests: GetFieldsByModelID (uses GetFieldsCommand)
// ============================================================================

func TestModelDesignAppService_GetFieldsByModelID(t *testing.T) {
	ctx := newTestContext()

	t.Run("success", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		expectedFields := []*modeldesign.FieldDefinition{
			{Name: "id", Title: "ID"},
			{Name: "name", Title: "Name"},
		}
		mockModelRepo.On("GetFieldsByModelID", ctx, "model-1").Return(expectedFields, nil)

		cmd := GetFieldsCommand{ModelID: "model-1"}
		fields, err := service.GetFieldsByModelID(ctx, cmd)

		assert.NoError(t, err)
		assert.Len(t, fields, 2)
		assert.Equal(t, "id", fields[0].Name)
		assert.Equal(t, "name", fields[1].Name)
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("model not found", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		mockModelRepo.On("GetFieldsByModelID", ctx, "nonexistent").Return(nil, nil)

		cmd := GetFieldsCommand{ModelID: "nonexistent"}
		fields, err := service.GetFieldsByModelID(ctx, cmd)

		assert.NoError(t, err)
		assert.Nil(t, fields)
		mockModelRepo.AssertExpectations(t)
	})
}

// ============================================================================
// Tests: UpdateFieldSync (uses UpdateFieldCommand)
// ============================================================================

func TestModelDesignAppService_UpdateFieldSync(t *testing.T) {
	ctx := newTestContext()

	t.Run("success", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		now := time.Now()
		locator, err := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
		require.NoError(t, err)
		existingField := &modeldesign.FieldDefinition{
			ModelID:      "model-1",
			ModelLocator: locator,
			Name:         "username",
			Title:        "Old Title",
			Description:  "Old Desc",
			Type:         &modeldesign.FieldType{Format: modeldesign.FormatString, SchemaType: "string"},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
		mockModelRepo.On("GetFieldByModelID", ctx, "model-1", "username").Return(existingField, nil)
		mockModelRepo.On("UpdateField", ctx, mock.AnythingOfType("*modeldesign.FieldDefinition")).Return(nil)

		cmd := UpdateFieldCommand{
			ModelID:     "model-1",
			FieldName:   "username",
			Title:       "New Title",
			Description: "New Description",
		}

		err = service.UpdateFieldSync(ctx, cmd)

		assert.NoError(t, err)
		mockModelRepo.AssertExpectations(t)
	})

	t.Run("field not found", func(t *testing.T) {
		mockModelRepo := new(MockModelRepository)
		service := newTestService(mockModelRepo, nil, nil)

		existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
		mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
		mockModelRepo.On("GetFieldByModelID", ctx, "model-1", "nonexistent").Return(nil, nil)

		cmd := UpdateFieldCommand{
			ModelID:   "model-1",
			FieldName: "nonexistent",
			Title:     "Title",
		}

		err := service.UpdateFieldSync(ctx, cmd)

		assert.Error(t, err)
		var bizErr *bizerrors.BusinessError
		assert.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.FieldNotFound.GetCode(), bizErr.Info().GetCode())
		mockModelRepo.AssertExpectations(t)
	})
}

// ============================================================================
// Tests: Command type validation
// ============================================================================

func TestCreateModelCommand_Fields(t *testing.T) {
	cmd := CreateModelCommand{
		ProjectSlug:  "project-1",
		Name:         "test_model",
		Title:        "Test Model",
		Description:  "A test model",
		StorageType:  "mysql",
		DatabaseName: "test_db",
	}

	assert.Equal(t, "project-1", cmd.ProjectSlug)
	assert.Equal(t, "test_model", cmd.Name)
	assert.Equal(t, "Test Model", cmd.Title)
	assert.Equal(t, "A test model", cmd.Description)
	assert.Equal(t, "mysql", cmd.StorageType)
	assert.Equal(t, "test_db", cmd.DatabaseName)
}

func TestUpdateModelMetaCommand_Fields(t *testing.T) {
	title := "New Title"
	desc := "New Description"
	cmd := UpdateModelMetaCommand{
		ProjectSlug: "project-1",
		Title:       &title,
		Description: &desc,
	}

	assert.Equal(t, "project-1", cmd.ProjectSlug)
	assert.Equal(t, &title, cmd.Title)
	assert.Equal(t, &desc, cmd.Description)
}

func TestModelQueryCommand_Fields(t *testing.T) {
	cmd := ModelQueryCommand{
		ProjectSlug:  "project-1",
		DatabaseName: "db_1",
		Name:         "model_a",
		Page:         1,
		PageSize:     20,
	}

	assert.Equal(t, "project-1", cmd.ProjectSlug)
	assert.Equal(t, "db_1", cmd.DatabaseName)
	assert.Equal(t, "model_a", cmd.Name)
	assert.Equal(t, 1, cmd.Page)
	assert.Equal(t, 20, cmd.PageSize)
}

// ============================================================================
// Tests: AddFieldSync — DisplayOrder assignment
// ============================================================================

// newTestField creates a minimal valid FieldDefinition for testing.
func newTestField(modelID, name string, locator *modeldesign.ModelLocator) *modeldesign.FieldDefinition {
	fieldType, _ := modeldesign.NewFieldFormat(modeldesign.FormatString)
	return &modeldesign.FieldDefinition{
		ModelID:      modelID,
		ModelLocator: locator,
		Name:         name,
		Title:        name,
		Type:         fieldType,
		Status:       modeldesign.FieldStatusInit,
		Metadata:     map[string]any{},
	}
}

// MockEnumAssocRepo is a minimal stub for FieldEnumAssociationRepository.
type MockEnumAssocRepo struct {
	mock.Mock
}

func (m *MockEnumAssocRepo) Create(ctx context.Context, assoc *modeldesign.FieldEnumAssociation) error {
	args := m.Called(ctx, assoc)
	return args.Error(0)
}

func (m *MockEnumAssocRepo) FindByField(
	ctx context.Context, modelID, fieldName string,
) (*modeldesign.FieldEnumAssociation, error) {
	return nil, nil
}

func (m *MockEnumAssocRepo) FindByEnumName(
	ctx context.Context, orgName, projectID, enumName string,
) ([]*modeldesign.FieldEnumAssociation, error) {
	return nil, nil
}

func (m *MockEnumAssocRepo) FindByModelID(
	ctx context.Context, modelID string,
) ([]*modeldesign.FieldEnumAssociation, error) {
	return nil, nil
}

func (m *MockEnumAssocRepo) Delete(ctx context.Context, modelID, fieldName string) error {
	return nil
}

func (m *MockEnumAssocRepo) DeleteByModelID(ctx context.Context, modelID string) error {
	return nil
}

// TestAddFieldSync_PhysicField_SetsDisplayOrder verifies that display_order is
// computed and assigned before a physical field is persisted.
func TestAddFieldSync_PhysicField_SetsDisplayOrder(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumRepo := new(MockEnumAssocRepo)

	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:     mockModelRepo,
		DeployRepo:    mockDeployRepo,
		EnumAssocRepo: mockEnumRepo,
	})

	locator, _ := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
	newField := newTestField("model-1", "user_name", locator)

	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("P", nil)
	mockModelRepo.On("AddFields", ctx, "test-org", mock.MatchedBy(func(fields []*modeldesign.FieldDefinition) bool {
		return len(fields) == 1 && fields[0].DisplayOrder != ""
	})).Return(nil)
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(nil)
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{newField},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, newField.DisplayOrder)
	assert.Greater(t, newField.DisplayOrder, "P")
	mockModelRepo.AssertExpectations(t)
}

// TestAddFieldSync_MultipleFields_SequentialDisplayOrders verifies that when
// adding multiple fields at once, each gets a strictly ascending display_order.
func TestAddFieldSync_MultipleFields_SequentialDisplayOrders(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumRepo := new(MockEnumAssocRepo)

	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:     mockModelRepo,
		DeployRepo:    mockDeployRepo,
		EnumAssocRepo: mockEnumRepo,
	})

	locator, _ := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")

	f1 := newTestField("model-1", "field_a", locator)
	f2 := newTestField("model-1", "field_b", locator)
	f3 := newTestField("model-1", "field_c", locator)

	var capturedFields []*modeldesign.FieldDefinition
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("N", nil)
	mockModelRepo.On("AddFields", ctx, "test-org", mock.MatchedBy(func(fields []*modeldesign.FieldDefinition) bool {
		capturedFields = fields
		return len(fields) == 3
	})).Return(nil)
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(nil)
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{f1, f2, f3},
	})

	require.NoError(t, err)
	require.Len(t, capturedFields, 3)
	assert.Greater(t, capturedFields[0].DisplayOrder, "N")
	assert.Greater(t, capturedFields[1].DisplayOrder, capturedFields[0].DisplayOrder)
	assert.Greater(t, capturedFields[2].DisplayOrder, capturedFields[1].DisplayOrder)
	mockModelRepo.AssertExpectations(t)
}

func TestModelDesignAppService_AddFieldsWithResults_PartialSuccess(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:     mockModelRepo,
		DeployRepo:    mockDeployRepo,
		EnumAssocRepo: mockEnumAssocRepo,
	})

	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
	locator := &existingModel.ModelLocator

	validField := newTestField("model-1", "email", locator)
	enumType, _ := modeldesign.NewFieldFormat(modeldesign.FormatEnum)
	invalidEnumField := &modeldesign.FieldDefinition{
		ModelID:      "model-1",
		ModelLocator: locator,
		Name:         "level",
		Title:        "level",
		Type:         enumType,
		Status:       modeldesign.FieldStatusInit,
	}

	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil).Twice()
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("", nil).Once()
	mockModelRepo.On("AddFields", ctx, "test-org", mock.Anything).Return(nil).Once()
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(nil).Once()
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil).Once()

	results, err := svc.AddFieldsWithResults(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{validField, invalidEnumField},
	})
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.True(t, results[0].Success)
	assert.Nil(t, results[0].Err)
	assert.Equal(t, "email", results[0].Name)

	assert.False(t, results[1].Success)
	assert.NotNil(t, results[1].Err)
	assert.Equal(t, "level", results[1].Name)
	assert.Contains(t, results[1].Err.Error(), "relateEnumName is required when format=ENUM")

	mockModelRepo.AssertExpectations(t)
	mockDeployRepo.AssertExpectations(t)
}

func TestModelDesignAppService_UpdateFieldSync_FormatImmutable(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	svc := newTestService(mockModelRepo, nil, nil)

	now := time.Now()
	locator, err := modeldesign.NewModelLocator("test-org", "project-1", "db_1", "test_model")
	require.NoError(t, err)
	enumType, _ := modeldesign.NewFieldFormat(modeldesign.FormatEnum)
	existingField := &modeldesign.FieldDefinition{
		ModelID:      "model-1",
		ModelLocator: locator,
		Name:         "level",
		Title:        "Level",
		Type:         enumType,
		EnumName:     "CustomerLevel",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
	existingModel.CreatedVia = modeldesign.ModelCreationSourceNew
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockModelRepo.On("GetFieldByModelID", ctx, "model-1", "level").Return(existingField, nil)

	format := modeldesign.FormatString
	err = svc.UpdateFieldSync(ctx, UpdateFieldCommand{
		ModelID:   "model-1",
		FieldName: "level",
		Format:    &format,
	})
	require.Error(t, err)
	assert.True(t, bizerrors.Is(err, ErrFieldFormatImmutable))
}

// ============================================================================
// Tests: addPhysicFields — Deploy failure and rollback failure combined error
// ============================================================================

func TestAddPhysicFields_DeployFailAndRollbackFail_ReturnsCombinedError(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:     mockModelRepo,
		DeployRepo:    mockDeployRepo,
		EnumAssocRepo: mockEnumAssocRepo,
	})

	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
	locator := &existingModel.ModelLocator
	newField := newTestField("model-1", "test_field", locator)

	deployErr := errors.New("deploy to client DB failed")
	rollbackErr := errors.New("rollback delete fields failed")

	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("P", nil)
	mockModelRepo.On("AddFields", ctx, "test-org", mock.Anything).Return(nil)
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(deployErr)
	mockModelRepo.On("DeleteFields", ctx, "model-1", []string{"test_field"}).Return(rollbackErr)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{newField},
	})

	require.Error(t, err)
	// Verify error message contains both deploy failure and rollback failure info
	assert.Contains(t, err.Error(), "部署模型到客户DB失败且回滚失败")
	assert.Contains(t, err.Error(), "model-1")
	// Verify combined error unwraps to both original errors
	assert.True(t, errors.Is(err, deployErr), "combined error should wrap deploy error")
	assert.True(t, errors.Is(err, rollbackErr), "combined error should wrap rollback error")
	mockModelRepo.AssertExpectations(t)
	mockDeployRepo.AssertExpectations(t)
}

func TestAddFieldSync_EnumAssociationCreateFail_ReturnsError(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockEnumAssocRepo := new(MockEnumAssocRepo)

	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:     mockModelRepo,
		DeployRepo:    mockDeployRepo,
		EnumAssocRepo: mockEnumAssocRepo,
	})

	existingModel := newTestModel("model-1", "project-1", "test_model", "db_1")
	locator := &existingModel.ModelLocator
	enumType, _ := modeldesign.NewFieldFormat(modeldesign.FormatEnum)
	enumField := &modeldesign.FieldDefinition{
		ModelID:      "model-1",
		ModelLocator: locator,
		Name:         "status",
		Title:        "status",
		Type:         enumType,
		EnumName:     "status_enum",
		Status:       modeldesign.FieldStatusInit,
		Metadata:     map[string]any{},
	}

	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(existingModel, nil)
	mockModelRepo.On("GetTailFieldDisplayOrder", ctx, "model-1").Return("P", nil)
	mockModelRepo.On("AddFields", ctx, "test-org", mock.Anything).Return(nil)
	mockDeployRepo.On("DeployModelToAddFields", ctx, existingModel, mock.Anything).Return(nil)
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil)
	mockEnumAssocRepo.On("Create", ctx, mock.Anything).Return(errors.New("assoc create failed"))

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-1",
		Fields:  []*modeldesign.FieldDefinition{enumField},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "创建枚举关联失败")
	assert.Contains(t, err.Error(), "assoc create failed")
}

func TestModelDesignAppService_DeleteModelSync_ProtectedSystemModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	service := newTestService(mockModelRepo, mockDeployRepo, nil)

	protectedModel := newTestModel("model-protected", "project-1", "end_user_users", "mc_meta")
	protectedModel.CreatedVia = modeldesign.ModelCreationSourceImported

	mockModelRepo.On("GetByID", ctx, "model-protected", mock.Anything).Return(protectedModel, nil)

	err := service.DeleteModelSync(ctx, "model-protected", "project-1", true)

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.OperationDenied.GetCode(), bizErr.Info().GetCode())
	assert.Contains(t, bizErr.Msg(), "cannot delete protected system model")

	mockModelRepo.AssertExpectations(t)
	mockDeployRepo.AssertNotCalled(t, "DeployModelToDrop", mock.Anything, mock.Anything)
	mockModelRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestModelDesignAppService_AddFieldSync_ProtectedSystemModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)

	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo:  mockModelRepo,
		DeployRepo: mockDeployRepo,
	})

	protectedModel := newTestModel("model-protected", "project-1", "end_user_accounts", "mc_meta")
	protectedModel.CreatedVia = modeldesign.ModelCreationSourceImported

	locator := protectedModel.GetModelLocator()
	newField := newTestField("model-protected", "new_field", locator)

	mockModelRepo.On("GetByID", ctx, "model-protected", mock.Anything).Return(protectedModel, nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-protected",
		Fields:  []*modeldesign.FieldDefinition{newField},
	})

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.OperationDenied.GetCode(), bizErr.Info().GetCode())
	assert.Contains(t, bizErr.Msg(), "cannot add fields to protected system model")

	mockModelRepo.AssertExpectations(t)
	mockModelRepo.AssertNotCalled(t, "AddFields", mock.Anything, mock.Anything, mock.Anything)
	mockDeployRepo.AssertNotCalled(t, "DeployModelToAddFields", mock.Anything, mock.Anything, mock.Anything)
}

func TestModelDesignAppService_RemoveFieldSync_ProtectedSystemModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)

	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{
		ModelRepo: mockModelRepo,
	})

	protectedModel := newTestModel("model-protected", "project-1", "end_user_users", "mc_meta")
	protectedModel.CreatedVia = modeldesign.ModelCreationSourceImported

	mockModelRepo.On("GetByID", ctx, "model-protected", mock.Anything).Return(protectedModel, nil)

	err := svc.RemoveFieldSync(ctx, RemoveFieldCommand{
		ModelID:   "model-protected",
		FieldName: "name",
	})

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.OperationDenied.GetCode(), bizErr.Info().GetCode())
	assert.Contains(t, bizErr.Msg(), "cannot remove fields from protected system model")

	mockModelRepo.AssertExpectations(t)
	mockModelRepo.AssertNotCalled(t, "BulkDeleteFields", mock.Anything, mock.Anything)
}

func TestModelDesignAppService_UpdateModelMeta_ReadOnlyModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	service := newTestService(mockModelRepo, nil, nil)

	importedModel := newTestModel("model-imported", "project-1", "customer_orders", "db_1")
	importedModel.CreatedVia = modeldesign.ModelCreationSourceImported
	importedModel.IsReadOnly = true
	mockModelRepo.On("GetByID", ctx, "model-imported", mock.Anything).Return(importedModel, nil)

	newTitle := "Updated"
	err := service.UpdateModelMeta(ctx, "model-imported", UpdateModelMetaCommand{Title: &newTitle})

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.ManagedModelReadOnly.GetCode(), bizErr.Info().GetCode())
	mockModelRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestModelDesignAppService_AddFieldSync_ReadOnlyModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{ModelRepo: mockModelRepo, DeployRepo: mockDeployRepo})

	importedModel := newTestModel("model-imported", "project-1", "customer_orders", "db_1")
	importedModel.CreatedVia = modeldesign.ModelCreationSourceImported
	importedModel.IsReadOnly = true
	loc := importedModel.GetModelLocator()
	newField := newTestField("model-imported", "extra_col", loc)

	mockModelRepo.On("GetByID", ctx, "model-imported", mock.Anything).Return(importedModel, nil)

	err := svc.AddFieldSync(ctx, AddFieldCommand{
		ModelID: "model-imported",
		Fields:  []*modeldesign.FieldDefinition{newField},
	})

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.ManagedModelReadOnly.GetCode(), bizErr.Info().GetCode())
	mockModelRepo.AssertNotCalled(t, "AddFields", mock.Anything, mock.Anything, mock.Anything)
	mockDeployRepo.AssertNotCalled(t, "DeployModelToAddFields", mock.Anything, mock.Anything, mock.Anything)
}

func TestModelDesignAppService_UpdateFieldSync_ReadOnlyModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{ModelRepo: mockModelRepo})

	importedModel := newTestModel("model-imported", "project-1", "customer_orders", "db_1")
	importedModel.CreatedVia = modeldesign.ModelCreationSourceImported
	importedModel.IsReadOnly = true
	mockModelRepo.On("GetByID", ctx, "model-imported", mock.Anything).Return(importedModel, nil)

	newTitle := "new-title"
	err := svc.UpdateFieldSync(ctx, UpdateFieldCommand{ModelID: "model-imported", FieldName: "name", Title: newTitle})

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.ManagedModelReadOnly.GetCode(), bizErr.Info().GetCode())
	mockModelRepo.AssertNotCalled(t, "GetFieldByModelID", mock.Anything, mock.Anything, mock.Anything)
}

func TestModelDesignAppService_RemoveFieldSync_ReadOnlyModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	svc := NewModelDesignAppService(ModelDesignAppServiceDeps{ModelRepo: mockModelRepo})

	importedModel := newTestModel("model-imported", "project-1", "customer_orders", "db_1")
	importedModel.CreatedVia = modeldesign.ModelCreationSourceImported
	importedModel.IsReadOnly = true
	mockModelRepo.On("GetByID", ctx, "model-imported", mock.Anything).Return(importedModel, nil)

	err := svc.RemoveFieldSync(ctx, RemoveFieldCommand{ModelID: "model-imported", FieldName: "name"})

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.ManagedModelReadOnly.GetCode(), bizErr.Info().GetCode())
	mockModelRepo.AssertNotCalled(t, "UpdateFieldsStatus", mock.Anything, mock.Anything)
}

func TestModelDesignAppService_DeleteModelSync_ReadOnlyModelDenied(t *testing.T) {
	ctx := newTestContext()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	service := newTestService(mockModelRepo, mockDeployRepo, nil)

	importedModel := newTestModel("model-imported", "project-1", "customer_orders", "db_1")
	importedModel.CreatedVia = modeldesign.ModelCreationSourceImported
	importedModel.IsReadOnly = true
	mockModelRepo.On("GetByID", ctx, "model-imported", mock.Anything).Return(importedModel, nil)

	err := service.DeleteModelSync(ctx, "model-imported", "project-1", true)

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.ManagedModelReadOnly.GetCode(), bizErr.Info().GetCode())
	mockDeployRepo.AssertNotCalled(t, "DeployModelToDrop", mock.Anything, mock.Anything)
	mockModelRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
