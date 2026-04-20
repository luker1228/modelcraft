package modeldesign

import (
	"context"
	domainmodel "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/infrastructure/dbgen"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogicalFKAppService_CreateLogicalForeignKey_BindsBelongsToFields(t *testing.T) {
	ctx := context.Background()

	mockFKRepo := new(MockLogicalForeignKeyRepository)
	mockModelRepo := new(MockModelRepository)
	mockTxManager := new(MockTxManager)

	sourceModel := &domainmodel.DataModel{
		Fields: []*domainmodel.FieldDefinition{
			{Name: "user_id"},
		},
	}
	sourceModel.ID = "model-order"
	sourceModel.ModelName = "Order"
	sourceModel.ProjectSlug = "proj-1"

	refModel := &domainmodel.DataModel{
		Fields: []*domainmodel.FieldDefinition{
			{Name: "id"},
		},
	}
	refModel.ID = "model-user"
	refModel.ModelName = "User"
	refModel.ProjectSlug = "proj-1"

	mockModelRepo.
		On("GetByID", ctx, "model-order", mock.Anything).
		Return(sourceModel, nil).Once()
	mockModelRepo.
		On("GetByID", ctx, "model-user", mock.Anything).
		Return(refModel, nil).Once()

	mockFKRepo.On("Save", ctx, mock.AnythingOfType("*modeldesign.LogicalForeignKey")).
		Return(nil).Twice()
	mockFKRepo.
		On(
			"BindBelongsToFields",
			ctx,
			"test-org",
			"model-order",
			mock.AnythingOfType("string"),
			[]string{"user_id"},
		).
		Return(nil).Once()

	mockTxManager.
		On("WithTx", ctx, mock.AnythingOfType("func(context.Context, dbgen.Querier) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context, dbgen.Querier) error)
			err := fn(ctx, nil)
			require.NoError(t, err)
		}).
		Return(nil).Once()

	svc := NewLogicalFKAppService(mockFKRepo, mockModelRepo, mockTxManager)
	svc.fkRepoFactory = func(_ dbgen.Querier) domainmodel.LogicalForeignKeyRepository {
		return mockFKRepo
	}

	result, err := svc.CreateLogicalForeignKey(ctx, CreateLogicalForeignKeyCommand{
		OrgName:      "test-org",
		ModelID:      "model-order",
		RefModelID:   "model-user",
		SourceFields: []string{"user_id"},
		TargetFields: []string{"id"},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "model-order", result.ModelID)
	assert.Equal(t, "model-user", result.RefModelID)

	mockModelRepo.AssertExpectations(t)
	mockFKRepo.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}
