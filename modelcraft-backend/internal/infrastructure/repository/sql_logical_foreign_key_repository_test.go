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

// MockQuerierForFK is a mock for logicalFKQuerier used in FK repository tests.
type MockQuerierForFK struct {
	mock.Mock
}

func (m *MockQuerierForFK) BindBelongsToFKIDToFields(
	ctx context.Context, arg dbgen.BindBelongsToFKIDToFieldsParams,
) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerierForFK) CreateLogicalForeignKey(ctx context.Context, arg dbgen.CreateLogicalForeignKeyParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerierForFK) DeleteLogicalForeignKeyByPairID(
	ctx context.Context, arg dbgen.DeleteLogicalForeignKeyByPairIDParams,
) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerierForFK) FindLogicalForeignKeysByModelID(
	ctx context.Context, arg dbgen.FindLogicalForeignKeysByModelIDParams,
) ([]dbgen.LogicalForeignKey, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.LogicalForeignKey), args.Error(1)
}

func (m *MockQuerierForFK) FindLogicalForeignKeysByPairID(
	ctx context.Context, arg dbgen.FindLogicalForeignKeysByPairIDParams,
) ([]dbgen.LogicalForeignKey, error) {
	args := m.Called(ctx, arg)
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
	ctx context.Context, arg dbgen.FindFieldsByBelongsToFKIDParams,
) ([]dbgen.FindFieldsByBelongsToFKIDRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.FindFieldsByBelongsToFKIDRow), args.Error(1)
}

func (m *MockQuerierForFK) FindFieldsByRelateFKID(
	ctx context.Context, arg dbgen.FindFieldsByRelateFKIDParams,
) ([]dbgen.FindFieldsByRelateFKIDRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.FindFieldsByRelateFKIDRow), args.Error(1)
}

func newTestLFRow(id, pairID, direction, modelID, modelName, refModelID, refModelName string) dbgen.LogicalForeignKey {
	sf, _ := json.Marshal([]string{"userId"})
	tf, _ := json.Marshal([]string{"id"})
	return dbgen.LogicalForeignKey{
		ID:              id,
		PairID:          pairID,
		Direction:       dbgen.LogicalForeignKeysDirection(direction),
		ModelID:         modelID,
		ModelName:       modelName,
		RefModelID:      sql.NullString{String: refModelID, Valid: refModelID != ""},
		RefModelName:    refModelName,
		RefDatabaseName: sql.NullString{},
		RefTableName:    sql.NullString{},
		SourceFields:    json.RawMessage(sf),
		TargetFields:    json.RawMessage(tf),
		IsDeletable:     true,
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
		IsDeletable:  true,
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

	expectedArg := dbgen.DeleteLogicalForeignKeyByPairIDParams{PairID: "pair-001", OrgName: "test-org"}
	mockQ.On("DeleteLogicalForeignKeyByPairID", ctx, expectedArg).Return(nil)
	err := repo.DeleteByPairID(ctx, "test-org", "pair-001")
	assert.NoError(t, err)
	mockQ.AssertExpectations(t)
}

func TestSqlLogicalForeignKeyRepository_FindByModel(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	rows := []dbgen.LogicalForeignKey{newTestLFRow("lf-001", "pair-001", "normal", "model-order", "Order", "model-user", "User")}
	expectedArg := dbgen.FindLogicalForeignKeysByModelIDParams{OrgName: "test-org", ModelID: "model-order"}
	mockQ.On("FindLogicalForeignKeysByModelID", ctx, expectedArg).Return(rows, nil)

	result, err := repo.FindByModel(ctx, "test-org", "model-order")
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
	expectedArg := dbgen.FindLogicalForeignKeysByPairIDParams{OrgName: "test-org", PairID: "pair-001"}
	mockQ.On("FindLogicalForeignKeysByPairID", ctx, expectedArg).Return(rows, nil)

	result, err := repo.FindByPairID(ctx, "test-org", "pair-001")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockQ.AssertExpectations(t)
}

func TestSqlLogicalForeignKeyRepository_FindByBelongsToField_NoFields(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	expectedArg := dbgen.FindFieldsByBelongsToFKIDParams{
		BelongsToFkID: sql.NullString{String: "lf-001", Valid: true},
		OrgName:       "test-org",
	}
	mockQ.On("FindFieldsByBelongsToFKID", ctx, expectedArg).Return([]dbgen.FindFieldsByBelongsToFKIDRow{}, nil)

	result, err := repo.FindByBelongsToField(ctx, "test-org", "lf-001")
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockQ.AssertExpectations(t)
}

func TestSqlLogicalForeignKeyRepository_FindByRelateField_NoFields(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	expectedArg := dbgen.FindFieldsByRelateFKIDParams{
		RelateFkID: sql.NullString{String: "lf-001", Valid: true},
		OrgName:    "test-org",
	}
	mockQ.On("FindFieldsByRelateFKID", ctx, expectedArg).Return([]dbgen.FindFieldsByRelateFKIDRow{}, nil)

	result, err := repo.FindByRelateField(ctx, "test-org", "lf-001")
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockQ.AssertExpectations(t)
}

func TestSqlLogicalForeignKeyRepository_BindBelongsToFields(t *testing.T) {
	mockQ := new(MockQuerierForFK)
	repo := &SqlLogicalForeignKeyRepository{q: mockQ}
	ctx := context.Background()

	expectedArg := dbgen.BindBelongsToFKIDToFieldsParams{
		BelongsToFkID: sql.NullString{String: "lf-normal-1", Valid: true},
		OrgName:       "test-org",
		ModelID:       "model-order",
		FieldNames:    []string{"user_id"},
	}
	mockQ.On("BindBelongsToFKIDToFields", ctx, expectedArg).Return(nil)

	err := repo.BindBelongsToFields(ctx, "test-org", "model-order", "lf-normal-1", []string{"user_id"})
	assert.NoError(t, err)
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
		IsDeletable:  true,
	}

	params, err := LogicalForeignKeyToCreateParams(lf)
	assert.NoError(t, err)
	assert.Equal(t, "lf-001", params.ID)
	assert.Equal(t, "pair-001", params.PairID)
	assert.Equal(t, dbgen.LogicalForeignKeysDirectionNormal, params.Direction)
	assert.Equal(t, "model-order", params.ModelID)
	assert.Equal(t, "Order", params.ModelName)
	assert.Equal(t, sql.NullString{String: "model-user", Valid: true}, params.RefModelID)
	assert.Equal(t, "User", params.RefModelName)
	assert.True(t, params.IsDeletable)

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
	assert.True(t, lf.IsDeletable)
}
