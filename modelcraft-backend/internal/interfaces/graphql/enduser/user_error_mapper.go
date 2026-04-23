package endusergraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/enduser/generated"
	"modelcraft/pkg/bizerrors"
)

func (r *queryResolver) mapRuntimeUserErrorForGet(
	ctx context.Context,
	requestID string,
	err error,
) *generated.GetRuntimeUserPayload {
	_ = ctx
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		return &generated.GetRuntimeUserPayload{
			Error: generated.ProjectNotFound{
				Message: "Internal server error (requestId: " + requestID + ")",
			},
		}
	}

	switch bizErr.Info().GetCode() {
	case bizerrors.EndUserNotFound.Code:
		return &generated.GetRuntimeUserPayload{
			Error: generated.EndUserNotFound{Message: bizErr.Msg()},
		}
	case bizerrors.ProjectNotFound.Code, bizerrors.EndUserClusterNotConfigured.Code:
		return &generated.GetRuntimeUserPayload{
			Error: generated.ProjectNotFound{Message: bizErr.Msg()},
		}
	case bizerrors.AuthUnauthorized.Code:
		return &generated.GetRuntimeUserPayload{
			Error: generated.Unauthorized{Message: bizErr.Msg()},
		}
	default:
		return &generated.GetRuntimeUserPayload{
			Error: generated.InvalidInput{Message: bizErr.Msg()},
		}
	}
}

func (r *queryResolver) mapRuntimeUserErrorForList(
	ctx context.Context,
	requestID string,
	err error,
) *generated.ListRuntimeUsersPayload {
	_ = ctx
	bizErr, ok := err.(*bizerrors.BusinessError)
	if !ok {
		return &generated.ListRuntimeUsersPayload{
			Error: generated.ProjectNotFound{
				Message: "Internal server error (requestId: " + requestID + ")",
			},
		}
	}

	switch bizErr.Info().GetCode() {
	case bizerrors.EndUserNotFound.Code:
		return &generated.ListRuntimeUsersPayload{
			Error: generated.EndUserNotFound{Message: bizErr.Msg()},
		}
	case bizerrors.ProjectNotFound.Code, bizerrors.EndUserClusterNotConfigured.Code:
		return &generated.ListRuntimeUsersPayload{
			Error: generated.ProjectNotFound{Message: bizErr.Msg()},
		}
	case bizerrors.AuthUnauthorized.Code:
		return &generated.ListRuntimeUsersPayload{
			Error: generated.Unauthorized{Message: bizErr.Msg()},
		}
	default:
		return &generated.ListRuntimeUsersPayload{
			Error: generated.InvalidInput{Message: bizErr.Msg()},
		}
	}
}
