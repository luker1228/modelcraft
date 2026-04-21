package adapter

import (
	"context"
	appmodeldesign "modelcraft/internal/app/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// ModelErrorAdapter handles conversion of domain errors to GraphQL errors for model operations
type ModelErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewModelErrorAdapter creates a new model error adapter
func NewModelErrorAdapter(ctx context.Context) *ModelErrorAdapter {
	return &ModelErrorAdapter{
		ctx:    ctx,
		logger: logfacade.GetLogger(ctx),
	}
}

// ConvertToGetError converts business error to GetModelError union type
func (a *ModelErrorAdapter) ConvertToGetError(err *bizerrors.BusinessError) generated.GetModelError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ModelNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ModelNotFound{
			Message: err.Msg(),
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for GetModel: %s", err.Info().GetCode())
		// Return as ModelNotFound for unknown errors
		return &generated.ModelNotFound{
			Message: err.Msg(),
		}
	}
}

// ConvertToCreateError converts business error to CreateModelError union type
func (a *ModelErrorAdapter) ConvertToCreateError(err *bizerrors.BusinessError) generated.CreateModelError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ModelAlreadyExists.GetCode():
		suggestion := "Please use a different model name within this project"
		return &generated.ModelAlreadyExists{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	case bizerrors.ModelTableAlreadyExists.GetCode():
		suggestion := "Table already exists. Use importModel to import the existing table as a model."
		return &generated.ModelTableAlreadyExists{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.ParamInvalid.GetCode():
		gqlErr := &generated.InvalidInput{
			Message: err.Msg(),
		}
		// Add suggestion if available from error detail
		if err.Detail() != "" {
			detail := err.Detail()
			gqlErr.Suggestion = &detail
		}
		return gqlErr
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for CreateModel: %s", err.Info().GetCode())
		// Return as InvalidInput for unknown errors
		msg := err.Msg()
		return &generated.InvalidInput{
			Message: msg,
		}
	}
}

// ConvertToUpdateError converts business error to UpdateModelError union type
func (a *ModelErrorAdapter) ConvertToUpdateError(err *bizerrors.BusinessError) generated.UpdateModelError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ModelNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ModelNotFound{
			Message: err.Msg(),
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.ParamInvalid.GetCode():
		gqlErr := &generated.InvalidInput{
			Message: err.Msg(),
		}
		if err.Detail() != "" {
			detail := err.Detail()
			gqlErr.Suggestion = &detail
		}
		return gqlErr
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for UpdateModel: %s", err.Info().GetCode())
		return &generated.ModelNotFound{
			Message: err.Msg(),
		}
	}
}

// ConvertToDeleteError converts business error to DeleteModelError union type
func (a *ModelErrorAdapter) ConvertToDeleteError(err *bizerrors.BusinessError) generated.DeleteModelError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ModelNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ModelNotFound{
			Message: err.Msg(),
		}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.OperationDenied.GetCode():
		// Handle operation denied errors (e.g., protected system models)
		return &generated.CannotDeleteDeployedModel{
			Message: err.Msg(),
		}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for DeleteModel: %s", err.Info().GetCode())
		return &generated.ModelNotFound{
			Message: err.Msg(),
		}
	}
}

// ConvertToAddFieldsError converts business error to AddFieldsError union type.
func (a *ModelErrorAdapter) ConvertToAddFieldsError(err *bizerrors.BusinessError) generated.AddFieldsError {
	if err == nil {
		return nil
	}
	gqlErr := &generated.InvalidInput{Message: err.Msg()}
	if err.Detail() != "" {
		detail := err.Detail()
		gqlErr.Suggestion = &detail
	}
	return gqlErr
}

// ConvertToUpdateFieldError converts errors to UpdateFieldError union type.
func (a *ModelErrorAdapter) ConvertToUpdateFieldError(err error) generated.UpdateFieldError {
	if err == nil {
		return nil
	}
	if bizerrors.Is(err, appmodeldesign.ErrFieldFormatImmutable) {
		return &generated.FieldFormatImmutable{
			Code:    appmodeldesign.FieldFormatImmutableCode,
			Message: "field format is immutable after creation",
		}
	}
	if bizErr, ok := err.(*bizerrors.BusinessError); ok {
		if bizErr.Info().GetCode() == bizerrors.OperationDenied.GetCode() {
			return &generated.InvalidInput{Message: bizErr.Msg()}
		}
		gqlErr := &generated.InvalidInput{Message: bizErr.Msg()}
		if bizErr.Detail() != "" {
			detail := bizErr.Detail()
			gqlErr.Suggestion = &detail
		}
		return gqlErr
	}
	return &generated.InvalidInput{Message: err.Error()}
}

// ConvertToRemoveFieldError converts errors to RemoveFieldError union type.
func (a *ModelErrorAdapter) ConvertToRemoveFieldError(err error) generated.RemoveFieldError {
	if err == nil {
		return nil
	}
	// Convert error types
	if bizErr, ok := err.(*bizerrors.BusinessError); ok {
		if bizErr.Info().GetCode() == bizerrors.OperationDenied.GetCode() {
			return &generated.InvalidInput{Message: bizErr.Msg()}
		}
		gqlErr := &generated.InvalidInput{Message: bizErr.Msg()}
		if bizErr.Detail() != "" {
			detail := bizErr.Detail()
			gqlErr.Suggestion = &detail
		}
		return gqlErr
	}
	return &generated.InvalidInput{Message: err.Error()}
}
