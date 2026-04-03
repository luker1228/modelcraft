package adapter

import (
	"context"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

// APIKeyErrorAdapter converts domain errors to GraphQL API key error union types.
type APIKeyErrorAdapter struct {
	ctx    context.Context
	logger logfacade.Logger
}

// NewAPIKeyErrorAdapter creates a new APIKeyErrorAdapter.
func NewAPIKeyErrorAdapter(ctx context.Context) *APIKeyErrorAdapter {
	return &APIKeyErrorAdapter{ctx: ctx, logger: logfacade.GetLogger(ctx)}
}

// ConvertToCreateError converts a business error to a CreateAPIKeyError union type.
func (a *APIKeyErrorAdapter) ConvertToCreateError(err *bizerrors.BusinessError) generated.CreateAPIKeyError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.APIKeyLimitExceeded.GetCode():
		return &generated.APIKeyLimitExceeded{Message: err.Msg()}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for CreateApiKey: %s", err.Info().GetCode())
		return &generated.APIKeyInvalidInput{Message: err.Msg()}
	}
}

// ConvertToRevokeError converts a business error to a RevokeAPIKeyError union type.
func (a *APIKeyErrorAdapter) ConvertToRevokeError(err *bizerrors.BusinessError) generated.RevokeAPIKeyError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.APIKeyNotFound.GetCode():
		return &generated.APIKeyNotFound{Message: err.Msg()}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for RevokeApiKey: %s", err.Info().GetCode())
		return &generated.APIKeyNotFound{Message: err.Msg()}
	}
}

// ConvertToUpdateError converts a business error to an UpdateAPIKeyError union type.
func (a *APIKeyErrorAdapter) ConvertToUpdateError(err *bizerrors.BusinessError) generated.UpdateAPIKeyError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.APIKeyNotFound.GetCode():
		return &generated.APIKeyNotFound{Message: err.Msg()}
	default:
		a.logger.Errorf(a.ctx, "Unknown error code for UpdateApiKey: %s", err.Info().GetCode())
		return &generated.APIKeyInvalidInput{Message: err.Msg()}
	}
}
