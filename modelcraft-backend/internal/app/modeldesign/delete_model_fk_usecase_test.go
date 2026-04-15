package modeldesign

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Task 4.9: Tests for model deletion cleanup (orphaned reverse FK rows)
// ---------------------------------------------------------------------------

// TestDeleteModelSync_CleansUpOrphanedFKRows verifies that after model deletion,
// orphaned reverse FK rows (where ref_model_id = deleted model) are also deleted.
func TestDeleteModelSync_CleansUpOrphanedFKRows(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockFKRepo := new(MockLogicalForeignKeyRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockTxMgr := new(MockTxManager)
	mockClusterRepo := new(MockClusterRepository)

	svc := NewModelDesignAppService(mockDeployRepo, mockModelRepo, mockClusterRepo, mockTxMgr)
	svc.WithFKRepo(mockFKRepo)

	deletedModelID := "model-deleted"
	model := makeModel(deletedModelID, "DeletedModel")

	// Two orphaned FK rows (same pair)
	orphanRow1 := &modeldesign.LogicalForeignKey{
		ID:           "fk-orphan-1",
		PairID:       "pair-orphan",
		Direction:    modeldesign.DirectionReverse,
		ModelID:      deletedModelID,
		RefModelID:   "model-other",
		SourceFields: []string{"id"},
		TargetFields: []string{"user_id"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockModelRepo.On("GetByID", ctx, deletedModelID, mock.Anything).Return(model, nil)
	mockDeployRepo.On("DeployModelToDrop", ctx, model).Return(nil)
	mockModelRepo.On("Delete", ctx, deletedModelID).Return(nil)
	mockFKRepo.On("FindByModel", ctx, "", deletedModelID).Return(
		[]*modeldesign.LogicalForeignKey{orphanRow1}, nil,
	)
	mockFKRepo.On("DeleteByPairID", ctx, "", "pair-orphan").Return(nil)

	err := svc.DeleteModelSync(ctx, deletedModelID, "proj1", true)
	assert.NoError(t, err)
	mockFKRepo.AssertCalled(t, "DeleteByPairID", ctx, "", "pair-orphan")
}

// TestDeleteModelSync_NoOrphanedRows verifies that deletion proceeds normally
// when there are no orphaned FK rows.
func TestDeleteModelSync_NoOrphanedRows(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockFKRepo := new(MockLogicalForeignKeyRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockTxMgr := new(MockTxManager)
	mockClusterRepo := new(MockClusterRepository)

	svc := NewModelDesignAppService(mockDeployRepo, mockModelRepo, mockClusterRepo, mockTxMgr)
	svc.WithFKRepo(mockFKRepo)

	deletedModelID := "model-clean"
	model := makeModel(deletedModelID, "CleanModel")

	mockModelRepo.On("GetByID", ctx, deletedModelID, mock.Anything).Return(model, nil)
	mockDeployRepo.On("DeployModelToDrop", ctx, model).Return(nil)
	mockModelRepo.On("Delete", ctx, deletedModelID).Return(nil)
	mockFKRepo.On("FindByModel", ctx, "", deletedModelID).Return([]*modeldesign.LogicalForeignKey{}, nil)

	err := svc.DeleteModelSync(ctx, deletedModelID, "proj1", true)
	assert.NoError(t, err)
	mockFKRepo.AssertNotCalled(t, "DeleteByPairID")
}

// TestDeleteModelSync_NoFKRepo verifies that deletion works without fkRepo (backward compat).
func TestDeleteModelSync_NoFKRepo(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockTxMgr := new(MockTxManager)
	mockClusterRepo := new(MockClusterRepository)

	// No fkRepo set
	svc := NewModelDesignAppService(mockDeployRepo, mockModelRepo, mockClusterRepo, mockTxMgr)

	deletedModelID := "model-no-fk"
	model := makeModel(deletedModelID, "NoFKModel")

	mockModelRepo.On("GetByID", ctx, deletedModelID, mock.Anything).Return(model, nil)
	mockDeployRepo.On("DeployModelToDrop", ctx, model).Return(nil)
	mockModelRepo.On("Delete", ctx, deletedModelID).Return(nil)

	err := svc.DeleteModelSync(ctx, deletedModelID, "proj1", true)
	assert.NoError(t, err)
}
