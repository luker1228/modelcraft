// Package modeldesign provides application-layer use cases for model design operations.
package modeldesign

import (
	"context"
	"database/sql"
	"modelcraft/pkg/logfacade"

	entity "modelcraft/internal/domain/modeldesign"
)

// clusterConnector abstracts cluster DB connection retrieval for testability.
type clusterConnector interface {
	GetConnection(ctx context.Context, orgName, projectSlug string) (*sql.DB, error)
}

// ActualSchemaQueryUseCase queries the actual database schema for a model.
// It handles connectivity failures gracefully by returning CLUSTER_UNREACHABLE instead of an error.
type ActualSchemaQueryUseCase struct {
	schemaService  entity.ActualSchemaService
	clusterManager clusterConnector
}

// NewActualSchemaQueryUseCase creates a new ActualSchemaQueryUseCase.
func NewActualSchemaQueryUseCase(
	schemaService entity.ActualSchemaService,
	clusterManager clusterConnector,
) *ActualSchemaQueryUseCase {
	return &ActualSchemaQueryUseCase{
		schemaService:  schemaService,
		clusterManager: clusterManager,
	}
}

// Query fetches the actual database schema for the given model.
// When the cluster cannot be reached, it returns a CLUSTER_UNREACHABLE result without an error.
// When the table does not exist, it returns a TABLE_MISSING result without an error.
func (uc *ActualSchemaQueryUseCase) Query(
	ctx context.Context,
	model *entity.DataModel,
	orgName string,
) (*entity.ActualSchemaResult, error) {
	logger := logfacade.GetLogger(ctx)

	db, err := uc.clusterManager.GetConnection(ctx, orgName, model.ProjectSlug)
	if err != nil {
		logger.Infof(ctx, "cluster unreachable for project=%s, orgName=%s: %v", model.ProjectSlug, orgName, err)
		return &entity.ActualSchemaResult{Status: entity.DbTableClusterUnreachable}, nil
	}

	// Filter out virtual fields — they have no DB column to query.
	var queryFields []*entity.FieldDefinition
	for _, f := range model.Fields {
		if !f.IsEnumLabelField() {
			queryFields = append(queryFields, f)
		}
	}

	result, err := uc.schemaService.QueryActualSchema(
		ctx,
		db,
		model.DatabaseName,
		model.ModelName,
		queryFields,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}
