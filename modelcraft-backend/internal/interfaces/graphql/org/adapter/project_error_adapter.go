package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// ProjectErrorAdapter handles conversion of domain errors to GraphQL errors
type ProjectErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewProjectErrorAdapter creates a new project error adapter
func NewProjectErrorAdapter(ctx context.Context) *ProjectErrorAdapter {
	return &ProjectErrorAdapter{
		ctx:    ctx,
		logger: logfacade.GetLogger(ctx),
	}
}

// ConvertToCreateProjectErrors converts business error to CreateProjectError union type
func (a *ProjectErrorAdapter) ConvertToCreateProjectErrors(err *bizerrors.BusinessError) generated.CreateProjectError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectAlreadyExists.GetCode():
		suggestion := "Please use a different project ID"
		return &generated.ProjectAlreadyExists{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	case bizerrors.ParamInvalid.GetCode():
		gqlErr := &generated.InvalidProjectInput{
			Message: err.Msg(),
		}
		// Add suggestion if available from error detail
		if err.Detail() != "" {
			detail := err.Detail()
			gqlErr.Suggestion = &detail
		}
		return gqlErr
	case bizerrors.DatabaseConnectionFailed.GetCode():
		suggestion := "Please verify host, port, username, and password are correct"
		return &generated.DatabaseConnectionFailed{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for CreateProject: %s", err.Info().GetCode())
		// Return as InvalidProjectInput for unknown errors
		msg := err.Msg()
		return &generated.InvalidProjectInput{
			Message: msg,
		}
	}
}

// ConvertToUpdateProjectErrors converts business error to UpdateProjectError union type
func (a *ProjectErrorAdapter) ConvertToUpdateProjectErrors(err *bizerrors.BusinessError) generated.UpdateProjectError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.ParamInvalid.GetCode():
		gqlErr := &generated.InvalidProjectInput{
			Message: err.Msg(),
		}
		if err.Detail() != "" {
			detail := err.Detail()
			gqlErr.Suggestion = &detail
		}
		return gqlErr
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for UpdateProject: %s", err.Info().GetCode())
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	}
}

// ConvertToDeleteProjectErrors converts business error to DeleteProjectError union type
func (a *ProjectErrorAdapter) ConvertToDeleteProjectErrors(err *bizerrors.BusinessError) generated.DeleteProjectError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.CannotDeleteDefaultProject.GetCode(), bizerrors.OperationDenied.GetCode():
		return &generated.CannotDeleteDefaultProject{
			Message: err.Msg(),
		}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for DeleteProject: %s", err.Info().GetCode())
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	}
}

// ConvertToGetProjectErrors converts business error to GetProjectError union type
func (a *ProjectErrorAdapter) ConvertToGetProjectErrors(err *bizerrors.BusinessError) generated.GetProjectError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for GetProject: %s", err.Info().GetCode())
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	}
}
