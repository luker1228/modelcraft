package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// Connection error suggestion message
const connectionSuggestion = "Please verify host, port, username, and password are correct"

func orgResourceNotFound(message string, resourceType generated.ResourceType) *generated.ResourceNotFound {
	return &generated.ResourceNotFound{Message: message, ResourceType: resourceType}
}

// ClusterErrorAdapter handles conversion of domain errors to GraphQL errors for cluster operations
type ClusterErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewClusterErrorAdapter creates a new cluster error adapter
func NewClusterErrorAdapter(ctx context.Context) *ClusterErrorAdapter {
	return &ClusterErrorAdapter{
		ctx:    ctx,
		logger: logfacade.GetLogger(ctx),
	}
}

// ConvertToGetClusterError converts business error to GetClusterError union type
func (a *ClusterErrorAdapter) ConvertToGetClusterError(err *bizerrors.BusinessError) generated.GetClusterError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeProject)
	case bizerrors.ClusterNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeCluster)
	default:
		a.logger.Errorf(a.ctx, nil, "Unknown error code for GetCluster: %s", err.Info().GetCode())
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeCluster)
	}
}

// ConvertToUpdateClusterError converts business error to UpdateClusterError union type
func (a *ClusterErrorAdapter) ConvertToUpdateClusterError(err *bizerrors.BusinessError) generated.UpdateClusterError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeProject)
	case bizerrors.ClusterNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeCluster)
	case bizerrors.ParamInvalid.GetCode():
		gqlErr := &generated.InvalidInput{Message: err.Msg()}
		if err.Detail() != "" {
			detail := err.Detail()
			gqlErr.Suggestion = &detail
		}
		return gqlErr
	case bizerrors.DatabaseConnectionFailed.GetCode():
		suggestion := connectionSuggestion
		return &generated.DatabaseConnectionFailed{Message: err.Msg(), Suggestion: &suggestion}
	default:
		a.logger.Errorf(a.ctx, nil, "Unknown error code for UpdateCluster: %s", err.Info().GetCode())
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeCluster)
	}
}

// ConvertToDeleteClusterError converts business error to DeleteClusterError union type
func (a *ClusterErrorAdapter) ConvertToDeleteClusterError(err *bizerrors.BusinessError) generated.DeleteClusterError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeProject)
	case bizerrors.ClusterNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeCluster)
	default:
		a.logger.Errorf(a.ctx, nil, "Unknown error code for DeleteCluster: %s", err.Info().GetCode())
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeCluster)
	}
}

// ConvertToTestConnectionError converts business error to TestConnectionError union type
func (a *ClusterErrorAdapter) ConvertToTestConnectionError(err *bizerrors.BusinessError) generated.TestConnectionError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeProject)
	case bizerrors.ClusterNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return orgResourceNotFound(err.Msg(), generated.ResourceTypeCluster)
	case bizerrors.DatabaseConnectionFailed.GetCode():
		suggestion := connectionSuggestion
		return &generated.DatabaseConnectionFailed{Message: err.Msg(), Suggestion: &suggestion}
	default:
		a.logger.Errorf(a.ctx, nil, "Unknown error code for TestConnection: %s", err.Info().GetCode())
		return &generated.DatabaseConnectionFailed{Message: err.Msg()}
	}
}
