package endusergraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/enduser/generated"
	"modelcraft/pkg/bizerrors"
)

func (r *queryResolver) mapModelDatabaseCatalogError(
	ctx context.Context,
	requestID string,
	err error,
) *generated.GetModelDatabaseCatalogPayload {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		return &generated.GetModelDatabaseCatalogPayload{
			Error: generated.ProjectNotFound{
				Message: "Internal server error (requestId: " + requestID + ")",
			},
		}
	}

	code := bizErr.Info().GetCode()
	switch code {
	case bizerrors.ProjectNotFound.Code:
		return &generated.GetModelDatabaseCatalogPayload{
			Error: generated.ProjectNotFound{Message: bizErr.Msg()},
		}
	case bizerrors.AuthUnauthorized.Code:
		return &generated.GetModelDatabaseCatalogPayload{
			Error: generated.Unauthorized{Message: bizErr.Msg()},
		}
	default:
		return &generated.GetModelDatabaseCatalogPayload{
			Error: generated.InvalidInput{Message: bizErr.Msg()},
		}
	}
}

func (r *queryResolver) mapModelCatalogError(
	ctx context.Context,
	requestID string,
	err error,
) *generated.GetModelCatalogPayload {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		return &generated.GetModelCatalogPayload{
			Error: generated.ProjectNotFound{
				Message: "Internal server error (requestId: " + requestID + ")",
			},
		}
	}

	code := bizErr.Info().GetCode()
	switch code {
	case bizerrors.ProjectNotFound.Code:
		return &generated.GetModelCatalogPayload{
			Error: generated.ProjectNotFound{Message: bizErr.Msg()},
		}
	case bizerrors.AuthUnauthorized.Code:
		return &generated.GetModelCatalogPayload{
			Error: generated.Unauthorized{Message: bizErr.Msg()},
		}
	default:
		return &generated.GetModelCatalogPayload{
			Error: generated.InvalidInput{Message: bizErr.Msg()},
		}
	}
}
