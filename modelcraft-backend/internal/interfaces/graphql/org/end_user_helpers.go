package orggraphql

import (
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/bizerrors"
	"strings"
)

func convertOrgCreateEndUserError(err *bizerrors.BusinessError) generated.CreateEndUserError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.EndUserConflict.GetCode(), bizerrors.Conflict.GetCode():
		return &generated.EndUserAlreadyExists{Message: err.Msg()}
	case bizerrors.EndUserParamInvalid.GetCode(), bizerrors.ParamInvalid.GetCode():
		if strings.Contains(strings.ToLower(err.Msg()), "password") {
			suggestion := "Use at least 8 characters containing letters and digits"
			return &generated.EndUserPasswordTooWeak{Message: err.Msg(), Suggestion: &suggestion}
		}
		return &generated.InvalidInput{Message: err.Msg()}
	default:
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

func convertOrgUpdateEndUserError(err *bizerrors.BusinessError) generated.UpdateEndUserError {
	if err == nil {
		return nil
	}

	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.EndUserNotFound{Message: err.Msg()}
	case bizerrors.EndUserParamInvalid.GetCode(), bizerrors.ParamInvalid.GetCode():
		return &generated.InvalidInput{Message: err.Msg()}
	default:
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

func convertOrgDeleteEndUserError(err *bizerrors.BusinessError) generated.DeleteEndUserError {
	if err == nil {
		return nil
	}
	return &generated.EndUserNotFound{Message: err.Msg()}
}

func convertOrgListEndUsersError(err *bizerrors.BusinessError) generated.ListEndUsersError {
	if err == nil {
		return nil
	}
	return &generated.InvalidInput{Message: err.Msg()}
}

func toOrgOptionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}
