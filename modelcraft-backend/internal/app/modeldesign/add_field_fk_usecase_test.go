package modeldesign

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogicalForeignKeyRepository is a mock for modeldesign.LogicalForeignKeyRepository.
type MockLogicalForeignKeyRepository struct {
	mock.Mock
}

func (m *MockLogicalForeignKeyRepository) Save(ctx context.Context, fk *modeldesign.LogicalForeignKey) error {
	args := m.Called(ctx, fk)
	return args.Error(0)
}

func (m *MockLogicalForeignKeyRepository) DeleteByPairID(ctx context.Context, orgName, pairID string) error {
	args := m.Called(ctx, orgName, pairID)
	return args.Error(0)
}

func (m *MockLogicalForeignKeyRepository) FindByModel(
	ctx context.Context, orgName, modelID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, modelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) FindByPairID(
	ctx context.Context, orgName, pairID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, pairID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) FindByBelongsToField(
	ctx context.Context, orgName, fkID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, fkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) FindByRelateField(
	ctx context.Context, orgName, fkID string,
) ([]*modeldesign.LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, fkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*modeldesign.LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) GetByID(
	ctx context.Context, id string,
) (*modeldesign.LogicalForeignKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modeldesign.LogicalForeignKey), args.Error(1)
}

// MockTxManager is a mock for repository.TxManager.
type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, q dbgen.Querier) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Ensure MockTxManager implements repository.TxManager
var _ repository.TxManager = (*MockTxManager)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeLocator(modelName, dbName, projectSlug string) *modeldesign.ModelLocator {
	return &modeldesign.ModelLocator{
		ProjectScope: project.ProjectScope{OrgName: "test-org", ProjectSlug: projectSlug},
		ModelName:    modelName,
		DatabaseName: dbName,
	}
}

func makeModel(id, name string) *modeldesign.DataModel {
	m := &modeldesign.DataModel{}
	m.ID = id
	m.ModelName = name
	m.ProjectSlug = "proj1"
	m.DatabaseName = "db1"
	m.Fields = []*modeldesign.FieldDefinition{}
	return m
}

func makeFKRow(
	id, pairID, modelID, refModelID string,
	dir modeldesign.LogicalFKDirection,
) *modeldesign.LogicalForeignKey {
	return &modeldesign.LogicalForeignKey{
		ID:           id,
		PairID:       pairID,
		Direction:    dir,
		ModelID:      modelID,
		RefModelID:   refModelID,
		SourceFields: []string{"user_id"},
		TargetFields: []string{"id"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Task 4.5: Tests for AddFieldUseCase — RELATION format requires relate_fk_id
// ---------------------------------------------------------------------------

// TestAddRelationField_RequiresRelateFkId verifies that adding a RELATION-format
// field without relate_fk_id is rejected at the domain validation level.
func TestAddRelationField_RequiresRelateFkId(t *testing.T) {
	ft, _ := modeldesign.NewFieldFormat(modeldesign.FormatRelation)
	field := &modeldesign.FieldDefinition{
		ModelID:      "model-1",
		Name:         "user",
		Title:        "User",
		Type:         ft,
		ModelLocator: makeLocator("Order", "db1", "proj1"),
		RelateFKID:   nil, // missing!
	}
	err := field.Validate()
	assert.Error(t, err, "RELATION format field without relate_fk_id should fail validation")
	assert.Contains(t, err.Error(), "relate_fk_id")
}

// TestAddRelationField_WithRelateFkId verifies that a RELATION-format field
// with relate_fk_id passes domain validation.
func TestAddRelationField_WithRelateFkId(t *testing.T) {
	ft, _ := modeldesign.NewFieldFormat(modeldesign.FormatRelation)
	fkID := "fk-normal-1"
	field := &modeldesign.FieldDefinition{
		ModelID:      "model-1",
		Name:         "user",
		Title:        "User",
		Type:         ft,
		ModelLocator: makeLocator("Order", "db1", "proj1"),
		RelateFKID:   &fkID,
	}
	err := field.Validate()
	assert.NoError(t, err, "RELATION format field with relate_fk_id should pass validation")
}

// TestAddRelationField_RejectsFKIDMismatch verifies that AddFieldFKService rejects
// a RELATION field whose relate_fk_id references a FK row that belongs to a
// different model.
func TestAddRelationField_RejectsFKIDMismatch(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockFKRepo := new(MockLogicalForeignKeyRepository)

	svc := newAddFieldFKService(mockModelRepo, mockFKRepo)

	// FK row belongs to a different model
	fkRow := makeFKRow("fk-normal-1", "pair-1", "model-WRONG", "model-2", modeldesign.DirectionNormal)
	mockFKRepo.On("FindByModel", ctx, "test-org", "model-1").Return(
		[]*modeldesign.LogicalForeignKey{fkRow}, nil,
	)

	fkID := "fk-normal-1"
	err := svc.ValidateRelateFKID(ctx, "test-org", "model-1", &fkID)
	// The FK row ID matches but belongs to a different model — should be rejected
	assert.Error(t, err)
}

// TestAddRelationField_AcceptsMatchingFKID verifies that ValidateRelateFKID
// succeeds when the FK row's model_id matches the field's model.
func TestAddRelationField_AcceptsMatchingFKID(t *testing.T) {
	ctx := context.Background()
	mockModelRepo := new(MockModelRepository)
	mockFKRepo := new(MockLogicalForeignKeyRepository)

	svc := newAddFieldFKService(mockModelRepo, mockFKRepo)

	fkRow := makeFKRow("fk-normal-1", "pair-1", "model-1", "model-2", modeldesign.DirectionNormal)
	mockFKRepo.On("FindByModel", ctx, "test-org", "model-1").Return(
		[]*modeldesign.LogicalForeignKey{fkRow}, nil,
	)

	fkID := "fk-normal-1"
	err := svc.ValidateRelateFKID(ctx, "test-org", "model-1", &fkID)
	assert.NoError(t, err)
}

// TestAddRelationField_NilRelateFkIdSkipsValidation verifies that nil relate_fk_id
// skips FK validation (non-RELATION fields).
func TestAddRelationField_NilRelateFkIdSkipsValidation(t *testing.T) {
	ctx := context.Background()
	svc := newAddFieldFKService(new(MockModelRepository), new(MockLogicalForeignKeyRepository))

	err := svc.ValidateRelateFKID(ctx, "test-org", "model-1", nil)
	assert.NoError(t, err)
}
