package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/infrastructure/dbgen"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuerier is a mock for dbgen.Querier used in FK repository tests.
type MockQuerierForFK struct {
	mock.Mock
}

func (m *MockQuerierForFK) CreateLogicalForeignKey(ctx context.Context, arg dbgen.CreateLogicalForeignKeyParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerierForFK) DeleteLogicalForeignKeyByPairID(ctx context.Context, pairID string) error {
	args := m.Called(ctx, pairID)
	return args.Error(0)
}

func (m *MockQuerierForFK) FindLogicalForeignKeysByModelID(
	ctx context.Context, modelID string,
) ([]dbgen.LogicalForeignKey, error) {
	args := m.Called(ctx, modelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.LogicalForeignKey), args.Error(1)
}

func (m *MockQuerierForFK) FindLogicalForeignKeysByPairID(
	ctx context.Context, pairID string,
) ([]dbgen.LogicalForeignKey, error) {
	args := m.Called(ctx, pairID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.LogicalForeignKey), args.Error(1)
}

func (m *MockQuerierForFK) FindLogicalForeignKeysByRefModelID(
	ctx context.Context, refModelID string,
) ([]dbgen.LogicalForeignKey, error) {
	args := m.Called(ctx, refModelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.LogicalForeignKey), args.Error(1)
}

func (m *MockQuerierForFK) GetLogicalForeignKeyByID(ctx context.Context, id string) (dbgen.LogicalForeignKey, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(dbgen.LogicalForeignKey), args.Error(1)
}

func (m *MockQuerierForFK) FindFieldsByBelongsToFKID(
	ctx context.Context, belongsToFkID sql.NullString,
) ([]dbgen.FindFieldsByBelongsToFKIDRow, error) {
	args := m.Called(ctx, belongsToFkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.FindFieldsByBelongsToFKIDRow), args.Error(1)
}

func (m *MockQuerierForFK) FindFieldsByRelateFKID(
	ctx context.Context, relateFkID sql.NullString,
) ([]dbgen.FindFieldsByRelateFKIDRow, error) {
	args := m.Called(ctx, relateFkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.FindFieldsByRelateFKIDRow), args.Error(1)
}

func newTestLFRow(id, pairID, direction, modelID, modelName, refModelID, refModelName string) dbgen.LogicalForeignKey {
	sf, _ := json.Marshal([]string{"userId"})
	tf, _ := json.Marshal([]string{"id"})
	return dbgen.LogicalForeignKey{
		ID:           id,
		PairID:       pairID,
		Direction:    dbgen.LogicalForeignKeysDirection(direction),
		ModelID:      modelID,
		ModelName:    modelName,
		RefModelID:   refModelID,
		RefModelName: refModelName,
		SourceFields: json.RawMessage(sf),
		TargetFields: json.RawMessage(tf),
	}
}

func TestSqlLogicalForeignKeyRepository_Save(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	lf := &modeldesign.LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    modeldesign.DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId"},
		TargetFields: []string{"id"},
	}

	mockQ.On("CreateLogicalForeignKey", ctx, mock.AnythingOfType("dbgen.CreateLogicalForeignKeyParams")).Return(nil)
	err := repo.Save(ctx, lf)
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestSqlLogicalForeignKeyRepository_DeleteByPairID(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	mockQ.On("DeleteLogicalForeignKeyByPairID", ctx, "pair-001").Return(nil)
	err := repo.DeleteByPairID(ctx, "pair-001")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestSqlLogicalForeignKeyRepository_FindByModel(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	rows := []dbgen.LogicalForeignKey{
		newTestLFRow("lf-001", "pair-001", "normal", "model-order", "Order", "model-user", "User"),
	}

	mockQ.On("FindLogicalForeignKeysByModelID", ctx, "model-order").Return(rows, nil)
	result, err := repo.FindByModel(ctx, "model-order")
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "lf-001", result[0].ID)
	assert.Equal(t, modeldesign.DirectionNormal, result[0].Direction)
	assert.Equal(t, "Order", result[0].ModelName)
	assert.Equal(t, "User", result[0].RefModelName)
	mockQ.AssertExpectations(t)
}

func TestSqlLogicalForeignKeyRepository_FindByPairID(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	rows := []dbgen.LogicalForeignKey{
		newTestLFRow("lf-001", "pair-001", "normal", "model-order", "Order", "model-user", "User"),
		newTestLFRow("lf-002", "pair-001", "reverse", "model-user", "User", "model-order", "Order"),
	}

	mockQ.On("FindLogicalForeignKeysByPairID", ctx, "pair-001").Return(rows, nil)
	result, err := repo.FindByPairID(ctx, "pair-001")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockQ.AssertExpectations(t)
}

func TestLogicalForeignKeyToCreateParams(t *testing.T) {
	lf := &modeldesign.LogicalForeignKey{
		ID:           "lf-001",
		PairID:       "pair-001",
		Direction:    modeldesign.DirectionNormal,
		ModelID:      "model-order",
		ModelName:    "Order",
		RefModelID:   "model-user",
		RefModelName: "User",
		SourceFields: []string{"userId", "companyId"},
		TargetFields: []string{"id", "companyId"},
	}

	params, err := LogicalForeignKeyToCreateParams(lf)
	assert.NoError(t, err)
	assert.Equal(t, "lf-001", params.ID)
	assert.Equal(t, "pair-001", params.PairID)
	assert.Equal(t, dbgen.LogicalForeignKeysDirectionNormal, params.Direction)
	assert.Equal(t, "model-order", params.ModelID)
	assert.Equal(t, "Order", params.ModelName)
	assert.Equal(t, "model-user", params.RefModelID)
	assert.Equal(t, "User", params.RefModelName)
	assert.NotNil(t, params.SourceFields)
	assert.NotNil(t, params.TargetFields)

	var sf, tf []string
	assert.NoError(t, json.Unmarshal(params.SourceFields, &sf))
	assert.NoError(t, json.Unmarshal(params.TargetFields, &tf))
	assert.Equal(t, []string{"userId", "companyId"}, sf)
	assert.Equal(t, []string{"id", "companyId"}, tf)
}

func TestLogicalForeignKeyToDomain(t *testing.T) {
	row := newTestLFRow("lf-001", "pair-001", "normal", "model-order", "Order", "model-user", "User")
	lf, err := LogicalForeignKeyToDomain(row)
	assert.NoError(t, err)
	assert.Equal(t, "lf-001", lf.ID)
	assert.Equal(t, "pair-001", lf.PairID)
	assert.Equal(t, modeldesign.DirectionNormal, lf.Direction)
	assert.Equal(t, "model-order", lf.ModelID)
	assert.Equal(t, "Order", lf.ModelName)
	assert.Equal(t, "model-user", lf.RefModelID)
	assert.Equal(t, "User", lf.RefModelName)
	assert.Equal(t, []string{"userId"}, lf.SourceFields)
	assert.Equal(t, []string{"id"}, lf.TargetFields)
}
