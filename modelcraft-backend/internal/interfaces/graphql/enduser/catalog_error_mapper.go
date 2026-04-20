package endusergraphql

import (
	"context"

	"modelcraft/internal/interfaces/graphql/enduser/generated"
	"modelcraft/pkg/bizerrors"
)

func (r *queryResolver) mapDatabaseCatalogError(
	ctx context.Context,
	requestID string,
	err error,
) *generated.GetDatabaseCatalogPayload {
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.ProjectNotFound{
				Message: "Internal server error (requestId: " + requestID + ")",
			},
		}
	}

	code := bizErr.Info().GetCode()
	switch code {
	case bizerrors.ProjectNotFound.Code:
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.ProjectNotFound{Message: bizErr.Msg()},
		}
	case bizerrors.AuthUnauthorized.Code:
		return &generated.GetDatabaseCatalogPayload{
			Error: generated.Unauthorized{Message: bizErr.Msg()},
		}
	default:
		return &generated.GetDatabaseCatalogPayload{
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
