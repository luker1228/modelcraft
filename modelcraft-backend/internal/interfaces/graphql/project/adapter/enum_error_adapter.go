package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// EnumErrorAdapter handles conversion of domain errors to GraphQL errors for enum operations
type EnumErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewEnumErrorAdapter creates a new enum error adapter
func NewEnumErrorAdapter(ctx context.Context) *EnumErrorAdapter {
	return &EnumErrorAdapter{
		ctx:    ctx,
		logger: logfacade.GetLogger(ctx),
	}
}

// ConvertToGetEnumError converts business error to GetEnumError union type
func (a *EnumErrorAdapter) ConvertToGetEnumError(err *bizerrors.BusinessError) generated.GetEnumError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.EnumNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.EnumNotFound{
			Message: err.Msg(),
		}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for GetEnum: %s", err.Info().GetCode())
		return &generated.EnumNotFound{
			Message: err.Msg(),
		}
	}
}

// ConvertToCreateEnumError converts business error to CreateEnumError union type
func (a *EnumErrorAdapter) ConvertToCreateEnumError(err *bizerrors.BusinessError) generated.CreateEnumError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.EnumAlreadyExists.GetCode():
		suggestion := "Please use a different enum name within this project"
		return &generated.EnumAlreadyExists{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	case bizerrors.ParamInvalid.GetCode():
		gqlErr := &generated.InvalidEnumInput{
			Message: err.Msg(),
		}
		// Add suggestion if available from error detail
		if err.Detail() != "" {
			detail := err.Detail()
			gqlErr.Suggestion = &detail
		} else {
			suggestion := "Please check enum name format, options, and ensure option codes are unique"
			gqlErr.Suggestion = &suggestion
		}
		return gqlErr
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for CreateEnum: %s", err.Info().GetCode())
		// Return as InvalidEnumInput for unknown errors
		return &generated.InvalidEnumInput{
			Message: err.Msg(),
		}
	}
}

// ConvertToUpdateEnumError converts business error to UpdateEnumError union type
func (a *EnumErrorAdapter) ConvertToUpdateEnumError(err *bizerrors.BusinessError) generated.UpdateEnumError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.EnumNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.EnumNotFound{
			Message: err.Msg(),
		}
	case bizerrors.ParamInvalid.GetCode():
		gqlErr := &generated.InvalidEnumInput{
			Message: err.Msg(),
		}
		if err.Detail() != "" {
			detail := err.Detail()
			gqlErr.Suggestion = &detail
		} else {
			suggestion := "Please check enum options and ensure option codes are unique"
			gqlErr.Suggestion = &suggestion
		}
		return gqlErr
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for UpdateEnum: %s", err.Info().GetCode())
		return &generated.EnumNotFound{
			Message: err.Msg(),
		}
	}
}

// ConvertToDeleteEnumError converts business error to DeleteEnumError union type
func (a *EnumErrorAdapter) ConvertToDeleteEnumError(err *bizerrors.BusinessError) generated.DeleteEnumError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{
			Message: err.Msg(),
		}
	case bizerrors.EnumNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.EnumNotFound{
			Message: err.Msg(),
		}
	case bizerrors.CannotDeleteReferencedEnum.GetCode():
		suggestion := "Please remove the enum from these fields before deleting"
		return &generated.CannotDeleteReferencedEnum{
			Message:    err.Msg(),
			Suggestion: &suggestion,
		}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for DeleteEnum: %s", err.Info().GetCode())
		return &generated.EnumNotFound{
			Message: err.Msg(),
		}
	}
}
