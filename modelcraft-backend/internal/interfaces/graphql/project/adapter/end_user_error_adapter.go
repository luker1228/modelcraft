package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
	"strings"
)

// EndUserErrorAdapter converts BusinessError to End-User GraphQL union errors.
type EndUserErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewEndUserErrorAdapter creates a new EndUserErrorAdapter.
func NewEndUserErrorAdapter(ctx context.Context) *EndUserErrorAdapter {
	return &EndUserErrorAdapter{ctx: ctx, logger: logfacade.GetLogger(ctx)}
}

// ConvertToCreateError converts business error to CreateEndUserError union type.
func (a *EndUserErrorAdapter) ConvertToCreateError(err *bizerrors.BusinessError) generated.CreateEndUserError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.EndUserConflict.GetCode():
		return &generated.EndUserAlreadyExists{Message: err.Msg()}
	case bizerrors.EndUserParamInvalid.GetCode():
		if strings.Contains(strings.ToLower(err.Msg()), "password") {
			suggestion := "Use at least 8 characters containing letters and digits"
			return &generated.EndUserPasswordTooWeak{Message: err.Msg(), Suggestion: &suggestion}
		}
		return &generated.InvalidInput{Message: err.Msg()}
	case bizerrors.EndUserClusterNotConfigured.GetCode(), bizerrors.ClusterNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeCluster}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logger.Errorf(a.ctx, "unknown create end-user error code: %s", err.Info().GetCode())
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToUpdateError converts business error to UpdateEndUserError union type.
func (a *EndUserErrorAdapter) ConvertToUpdateError(err *bizerrors.BusinessError) generated.UpdateEndUserError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
	case bizerrors.EndUserParamInvalid.GetCode():
		return &generated.InvalidInput{Message: err.Msg()}
	case bizerrors.EndUserClusterNotConfigured.GetCode(), bizerrors.ClusterNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeCluster}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logger.Errorf(a.ctx, "unknown update end-user error code: %s", err.Info().GetCode())
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

// ConvertToDeleteError converts business error to DeleteEndUserError union type.
func (a *EndUserErrorAdapter) ConvertToDeleteError(err *bizerrors.BusinessError) generated.DeleteEndUserError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
	case bizerrors.EndUserClusterNotConfigured.GetCode(), bizerrors.ClusterNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeCluster}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeProject}
	default:
		a.logger.Errorf(a.ctx, "unknown delete end-user error code: %s", err.Info().GetCode())
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
	}
}
