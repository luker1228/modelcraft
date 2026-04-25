package projectgraphql

import (
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
)

func toGraphQLEndUserProjectAccess(access *appEnduser.EndUserProjectAccessDTO) *generated.EndUserProjectAccess {
	if access == nil {
		return nil
	}

	endUser := &generated.EndUser{}
	if access.EndUser != nil {
		endUser = &generated.EndUser{
			ID:          access.EndUser.ID,
			Username:    access.EndUser.Username,
			IsForbidden: access.EndUser.IsForbidden,
			CreatedBy:   access.EndUser.CreatedBy,
			CreatedAt:   access.EndUser.CreatedAt,
			UpdatedAt:   access.EndUser.UpdatedAt,
		}
	}

	return &generated.EndUserProjectAccess{
		ID:                   access.ID,
		EndUser:              endUser,
		PermissionBundleID:   access.PermissionBundleID,
		PermissionBundleName: access.PermissionBundleName,
		GrantedBy:            access.GrantedBy,
		GrantedAt:            access.GrantedAt,
	}
}

func convertGrantEndUserProjectAccessError(err *bizerrors.BusinessError) generated.GrantEndUserProjectAccessError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFound.GetCode(), bizerrors.EndUserNotFoundInProject.GetCode():
		return &generated.EndUserNotFound{Message: err.Msg()}
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.EndUserPermissionBundleNotFound{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{Message: err.Msg()}
	case bizerrors.Conflict.GetCode():
		return &generated.EndUserProjectAccessAlreadyExists{Message: err.Msg()}
	case bizerrors.ParamInvalid.GetCode(), bizerrors.EndUserParamInvalid.GetCode():
		return &generated.InvalidInput{Message: err.Msg()}
	default:
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

func convertUpdateEndUserProjectAccessError(err *bizerrors.BusinessError) generated.UpdateEndUserProjectAccessError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.NotFound.GetCode():
		return &generated.EndUserProjectAccessNotFound{Message: err.Msg()}
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return &generated.EndUserPermissionBundleNotFound{Message: err.Msg()}
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{Message: err.Msg()}
	case bizerrors.ParamInvalid.GetCode(), bizerrors.EndUserParamInvalid.GetCode():
		return &generated.InvalidInput{Message: err.Msg()}
	default:
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

func convertRevokeEndUserProjectAccessError(err *bizerrors.BusinessError) generated.RevokeEndUserProjectAccessError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{Message: err.Msg()}
	case bizerrors.NotFound.GetCode():
		return &generated.EndUserProjectAccessNotFound{Message: err.Msg()}
	default:
		return &generated.EndUserProjectAccessNotFound{Message: err.Msg()}
	}
}

func convertListProjectEndUserAccessError(err *bizerrors.BusinessError) generated.ListProjectEndUserAccessError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.ProjectNotFound.GetCode():
		return &generated.ProjectNotFound{Message: err.Msg()}
	default:
		return &generated.InvalidInput{Message: err.Msg()}
	}
}
