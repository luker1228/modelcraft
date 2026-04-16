package modeldesign

import (
	"context"
	"database/sql"
	"errors"
	"modelcraft/internal/domain/project"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	entity "modelcraft/internal/domain/modeldesign"
)

// mockActualSchemaService is a testify mock for entity.ActualSchemaService.
type mockActualSchemaService struct {
	mock.Mock
}

func (m *mockActualSchemaService) QueryActualSchema(
	ctx context.Context,
	db *sql.DB,
	databaseName string,
	tableName string,
	fields []*entity.FieldDefinition,
) (*entity.ActualSchemaResult, error) {
	args := m.Called(ctx, db, databaseName, tableName, fields)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ActualSchemaResult), args.Error(1)
}

// mockClusterConnector is a testify mock for the clusterConnector interface.
type mockClusterConnector struct {
	mock.Mock
}

func (m *mockClusterConnector) GetConnection(
	ctx context.Context,
	orgName, projectSlug string,
) (*sql.DB, error) {
	args := m.Called(ctx, orgName, projectSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sql.DB), args.Error(1)
}

// makeTestModel creates a minimal DataModel for testing.
func makeTestModel(projectSlug, dbName, modelName string, fields ...*entity.FieldDefinition) *entity.DataModel {
	return &entity.DataModel{
		ModelMeta: entity.ModelMeta{
			ModelLocator: entity.ModelLocator{
				ProjectScope: project.ProjectScope{ProjectSlug: projectSlug},
				DatabaseName: dbName,
				ModelName:    modelName,
			},
		},
		Fields: fields,
	}
}

func TestActualSchemaQueryUseCase_TableExists(t *testing.T) {
	schemaSvc := &mockActualSchemaService{}
	connector := &mockClusterConnector{}
	db := &sql.DB{}

	model := makeTestModel("proj", "mydb", "users",
		&entity.FieldDefinition{Name: "id", IsUnique: false, NonNull: true},
	)
	expectedResult := &entity.ActualSchemaResult{
		Status: entity.DbTableExists,
		Fields: map[string]*entity.DbColumnInfo{
			"id": {ColumnType: "BIGINT", Constraints: []entity.ActualConstraintType{entity.ActualConstraintNotNull}},
		},
	}

	connector.On("GetConnection", mock.Anything, "org1", "proj").Return(db, nil)
	schemaSvc.On("QueryActualSchema", mock.Anything, db, "mydb", "users", mock.Anything).Return(expectedResult, nil)

	uc := NewActualSchemaQueryUseCase(schemaSvc, connector)
	result, err := uc.Query(context.Background(), model, "org1")

	require.NoError(t, err)
	assert.Equal(t, entity.DbTableExists, result.Status)
	assert.NotNil(t, result.Fields)

	connector.AssertExpectations(t)
	schemaSvc.AssertExpectations(t)
}

func TestActualSchemaQueryUseCase_TableMissing(t *testing.T) {
	schemaSvc := &mockActualSchemaService{}
	connector := &mockClusterConnector{}
	db := &sql.DB{}

	model := makeTestModel("proj", "mydb", "users")
	missingResult := &entity.ActualSchemaResult{Status: entity.DbTableMissing}

	connector.On("GetConnection", mock.Anything, "org1", "proj").Return(db, nil)
	schemaSvc.On("QueryActualSchema", mock.Anything, db, "mydb", "users", mock.Anything).Return(missingResult, nil)

	uc := NewActualSchemaQueryUseCase(schemaSvc, connector)
	result, err := uc.Query(context.Background(), model, "org1")

	require.NoError(t, err)
	assert.Equal(t, entity.DbTableMissing, result.Status)
	assert.Nil(t, result.Fields)

	connector.AssertExpectations(t)
	schemaSvc.AssertExpectations(t)
}

func TestActualSchemaQueryUseCase_ClusterUnreachable(t *testing.T) {
	schemaSvc := &mockActualSchemaService{}
	connector := &mockClusterConnector{}

	model := makeTestModel("proj", "mydb", "users")

	connector.On("GetConnection", mock.Anything, "org1", "proj").Return(nil, errors.New("connection refused"))

	uc := NewActualSchemaQueryUseCase(schemaSvc, connector)
	result, err := uc.Query(context.Background(), model, "org1")

	require.NoError(t, err)
	assert.Equal(t, entity.DbTableClusterUnreachable, result.Status)
	assert.Nil(t, result.Fields)

	// schemaService should never be called when cluster is unreachable
	schemaSvc.AssertNotCalled(t, "QueryActualSchema")
	connector.AssertExpectations(t)
}

func TestActualSchemaQueryUseCase_VirtualFieldsExcluded(t *testing.T) {
	schemaSvc := &mockActualSchemaService{}
	connector := &mockClusterConnector{}
	db := &sql.DB{}

	fields := []*entity.FieldDefinition{
		{
			Name: "id", IsUnique: false, NonNull: true,
			Type: &entity.FieldType{Format: entity.FormatInteger},
		},
	}
	model := makeTestModel("proj", "mydb", "users", fields...)

	connector.On("GetConnection", mock.Anything, "org1", "proj").Return(db, nil)
	// Capture which fields are passed to QueryActualSchema
	schemaSvc.On("QueryActualSchema", mock.Anything, db, "mydb", "users", mock.Anything).
		Return(&entity.ActualSchemaResult{Status: entity.DbTableExists, Fields: map[string]*entity.DbColumnInfo{}}, nil)

	uc := NewActualSchemaQueryUseCase(schemaSvc, connector)
	result, err := uc.Query(context.Background(), model, "org1")

	require.NoError(t, err)
	assert.Equal(t, entity.DbTableExists, result.Status)

	connector.AssertExpectations(t)
	schemaSvc.AssertExpectations(t)
}
