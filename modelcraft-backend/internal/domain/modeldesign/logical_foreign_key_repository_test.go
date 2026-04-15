package modeldesign

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogicalForeignKeyRepository 用于测试的 mock 实现
type MockLogicalForeignKeyRepository struct {
	mock.Mock
}

func (m *MockLogicalForeignKeyRepository) Save(ctx context.Context, lf *LogicalForeignKey) error {
	args := m.Called(ctx, lf)
	return args.Error(0)
}

func (m *MockLogicalForeignKeyRepository) DeleteByPairID(ctx context.Context, orgName, pairID string) error {
	args := m.Called(ctx, orgName, pairID)
	return args.Error(0)
}

func (m *MockLogicalForeignKeyRepository) FindByModel(
	ctx context.Context, orgName, modelID string,
) ([]*LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, modelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) FindByPairID(
	ctx context.Context, orgName, pairID string,
) ([]*LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, pairID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) FindByBelongsToField(
	ctx context.Context, orgName, lfID string,
) ([]*LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, lfID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) FindByRelateField(
	ctx context.Context, orgName, lfID string,
) ([]*LogicalForeignKey, error) {
	args := m.Called(ctx, orgName, lfID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) GetByID(ctx context.Context, id string) (*LogicalForeignKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LogicalForeignKey), args.Error(1)
}

func (m *MockLogicalForeignKeyRepository) BindBelongsToFields(
	ctx context.Context, orgName, modelID, lfID string, fieldNames []string,
) error {
	args := m.Called(ctx, orgName, modelID, lfID, fieldNames)
	return args.Error(0)
}

// 确保 MockLogicalForeignKeyRepository 实现了 LogicalForeignKeyRepository 接口
var _ LogicalForeignKeyRepository = (*MockLogicalForeignKeyRepository)(nil)

func TestLogicalForeignKeyRepository_InterfaceCompliance(t *testing.T) {
	// 验证 mock 实现了接口（编译时检查）
	var _ LogicalForeignKeyRepository = (*MockLogicalForeignKeyRepository)(nil)
}

func TestLogicalForeignKeyRepository_Save(t *testing.T) {
	mockRepo := new(MockLogicalForeignKeyRepository)
	ctx := context.Background()
	lf := &LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    DirectionNormal,
		ModelID:      "model-order",
		RefModelID:   "model-user",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockRepo.On("Save", ctx, lf).Return(nil)
	err := mockRepo.Save(ctx, lf)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogicalForeignKeyRepository_DeleteByPairID(t *testing.T) {
	mockRepo := new(MockLogicalForeignKeyRepository)
	ctx := context.Background()

	mockRepo.On("DeleteByPairID", ctx, "test-org", "pair-001").Return(nil)
	err := mockRepo.DeleteByPairID(ctx, "test-org", "pair-001")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogicalForeignKeyRepository_FindByModel(t *testing.T) {
	mockRepo := new(MockLogicalForeignKeyRepository)
	ctx := context.Background()
	expected := []*LogicalForeignKey{
		{ID: "lf-001", PairID: "pair-001", Direction: DirectionNormal, ModelID: "model-order"},
	}

	mockRepo.On("FindByModel", ctx, "test-org", "model-order").Return(expected, nil)
	result, err := mockRepo.FindByModel(ctx, "test-org", "model-order")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "lf-001", result[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestLogicalForeignKeyRepository_FindByPairID(t *testing.T) {
	mockRepo := new(MockLogicalForeignKeyRepository)
	ctx := context.Background()
	expected := []*LogicalForeignKey{
		{ID: "lf-001", PairID: "pair-001", Direction: DirectionNormal},
		{ID: "lf-002", PairID: "pair-001", Direction: DirectionReverse},
	}

	mockRepo.On("FindByPairID", ctx, "test-org", "pair-001").Return(expected, nil)
	result, err := mockRepo.FindByPairID(ctx, "test-org", "pair-001")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestLogicalForeignKeyRepository_FindByBelongsToField(t *testing.T) {
	mockRepo := new(MockLogicalForeignKeyRepository)
	ctx := context.Background()

	mockRepo.On("FindByBelongsToField", ctx, "test-org", "lf-001").Return([]*LogicalForeignKey{}, nil)
	result, err := mockRepo.FindByBelongsToField(ctx, "test-org", "lf-001")
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestLogicalForeignKeyRepository_FindByRelateField(t *testing.T) {
	mockRepo := new(MockLogicalForeignKeyRepository)
	ctx := context.Background()

	mockRepo.On("FindByRelateField", ctx, "test-org", "lf-001").Return([]*LogicalForeignKey{}, nil)
	result, err := mockRepo.FindByRelateField(ctx, "test-org", "lf-001")
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}
