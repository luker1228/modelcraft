package modeldesign

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizerrors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Task 4.7: Tests for RemoveFieldUseCase
// ---------------------------------------------------------------------------

// makeRelationField creates a RELATION-format FieldDefinition with a relate_fk_id.
func makeRelationField(name, modelID, fkID string) *modeldesign.FieldDefinition {
	ft, _ := modeldesign.NewFieldFormat(modeldesign.FormatRelation)
	return &modeldesign.FieldDefinition{
		ModelID:      modelID,
		Name:         name,
		Title:        name,
		Type:         ft,
		ModelLocator: makeLocator("SomeModel", "db1", "proj1"),
		RelateFKID:   &fkID,
		Status:       modeldesign.FieldStatusDeploySuccess,
	}
}

// makeBelongsToField creates a FK-column FieldDefinition with a belongs_to_fk_id.
func makeBelongsToField(name, modelID, fkID string) *modeldesign.FieldDefinition {
	ft, _ := modeldesign.NewFieldFormat(modeldesign.FormatString)
	return &modeldesign.FieldDefinition{
		ModelID:       modelID,
		Name:          name,
		Title:         name,
		Type:          ft,
		ModelLocator:  makeLocator("SomeModel", "db1", "proj1"),
		BelongsToFKID: &fkID,
		Status:        modeldesign.FieldStatusDeploySuccess,
	}
}

// makeModelWithFields creates a DataModel with given fields.
func makeModelWithFields(id, name string, fields ...*modeldesign.FieldDefinition) *modeldesign.DataModel {
	m := &modeldesign.DataModel{}
	m.ID = id
	m.ModelName = name
	m.OrgName = "test-org"
	m.ProjectSlug = "proj1"
	m.DatabaseName = "db1"
	m.Fields = fields
	return m
}

// TestRemoveRelationField_AllowedFreely verifies that removing a RELATION-format
// field (relate_fk_id field) is allowed without any FK checks.
func TestRemoveRelationField_AllowedFreely(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockFKRepo := new(MockLogicalForeignKeyRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockTxMgr := new(MockTxManager)
	mockClusterRepo := new(MockClusterRepository)

	svc := NewModelDesignAppService(mockDeployRepo, mockModelRepo, mockClusterRepo, mockTxMgr)
	svc.WithFKRepo(mockFKRepo)

	relateField := makeRelationField("user_rel", "model-1", "fk-normal-1")
	model := makeModelWithFields("model-1", "Order", relateField)

	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(model, nil)
	mockModelRepo.On("BulkDeleteFields", ctx, mock.Anything).Return(nil)

	err := svc.RemoveFieldSync(ctx, RemoveFieldCommand{ModelID: "model-1", FieldName: "user_rel"})
	assert.NoError(t, err)
	// FK repo should NOT be consulted for RELATION fields
	mockFKRepo.AssertNotCalled(t, "FindByRelateField")
}

// TestRemoveBelongsToField_BlockedWhenRelateFieldsExist verifies that removing a
// belongs_to_fk_id field is blocked when RELATION fields still reference the FK.
func TestRemoveBelongsToField_BlockedWhenRelateFieldsExist(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockFKRepo := new(MockLogicalForeignKeyRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockTxMgr := new(MockTxManager)
	mockClusterRepo := new(MockClusterRepository)

	svc := NewModelDesignAppService(mockDeployRepo, mockModelRepo, mockClusterRepo, mockTxMgr)
	svc.WithFKRepo(mockFKRepo)

	fkID := "fk-normal-1"
	belongsField := makeBelongsToField("user_id", "model-1", fkID)
	model := makeModelWithFields("model-1", "Order", belongsField)

	// A RELATION field references this FK
	existingRelateField := makeRelationField("user_rel", "model-1", fkID)
	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(model, nil)
	// FindByRelateField returns a non-empty list (FK is still in use)
	mockFKRepo.On("FindByRelateField", ctx, "test-org", fkID).Return(
		[]*modeldesign.LogicalForeignKey{
			{ID: "fk-rel", PairID: "pair-1", ModelID: "model-1"},
		}, nil,
	)
	_ = existingRelateField

	err := svc.RemoveFieldSync(ctx, RemoveFieldCommand{ModelID: "model-1", FieldName: "user_id"})
	assert.Error(t, err)
	be, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.Equal(t, bizerrors.FKPairHasRelateFields.GetCode(), be.Info().GetCode())
}

// TestRemoveBelongsToField_DeletesFKPairWhenNoRelateFields verifies that removing a
// belongs_to_fk_id field cascades FK pair deletion when no RELATION fields remain.
func TestRemoveBelongsToField_DeletesFKPairWhenNoRelateFields(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockFKRepo := new(MockLogicalForeignKeyRepository)
	mockDeployRepo := new(MockDeployRepo)
	mockTxMgr := new(MockTxManager)
	mockClusterRepo := new(MockClusterRepository)

	svc := NewModelDesignAppService(mockDeployRepo, mockModelRepo, mockClusterRepo, mockTxMgr)
	svc.WithFKRepo(mockFKRepo)

	fkID := "fk-normal-1"
	pairID := "pair-1"
	belongsField := makeBelongsToField("user_id", "model-1", fkID)
	model := makeModelWithFields("model-1", "Order", belongsField)

	fkRow := &modeldesign.LogicalForeignKey{
		ID:           fkID,
		PairID:       pairID,
		Direction:    modeldesign.DirectionNormal,
		ModelID:      "model-1",
		RefModelID:   "model-2",
		SourceFields: []string{"user_id"},
		TargetFields: []string{"id"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockModelRepo.On("GetByID", ctx, "model-1", mock.Anything).Return(model, nil)
	// No RELATION fields reference this FK
	mockFKRepo.On("FindByRelateField", ctx, "test-org", fkID).Return([]*modeldesign.LogicalForeignKey{}, nil)
	mockFKRepo.On("FindByModel", ctx, "test-org", "model-1").Return([]*modeldesign.LogicalForeignKey{fkRow}, nil)
	mockFKRepo.On("DeleteByPairID", ctx, "test-org", pairID).Return(nil)
	mockModelRepo.On("UpdateFieldsStatus", ctx, mock.Anything).Return(nil)
	mockDeployRepo.On("DeployModelToRemoveFields", ctx, model, []string{"user_id"}).Return(nil)
	mockModelRepo.On("DeleteFields", ctx, "model-1", []string{"user_id"}).Return(nil)

	err := svc.RemoveFieldSync(ctx, RemoveFieldCommand{ModelID: "model-1", FieldName: "user_id"})
	assert.NoError(t, err)
	mockFKRepo.AssertCalled(t, "DeleteByPairID", ctx, "test-org", pairID)
}
