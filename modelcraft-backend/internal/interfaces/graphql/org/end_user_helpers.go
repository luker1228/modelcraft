package orggraphql

import (
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/bizerrors"
	"strings"
)

// convertUserWhereInput maps GraphQL UserWhereInput to app-layer filter.
func convertUserWhereInput(w *generated.UserWhereInput) *appEnduser.MetaUserFindManyFilter {
	if w == nil {
		return nil
	}
	f := &appEnduser.MetaUserFindManyFilter{}
	if w.ID != nil {
		f.IDEq = w.ID.Eq
		if len(w.ID.In) > 0 {
			f.IDIn = w.ID.In
		}
	}
	if w.Username != nil {
		f.UsernameEq = w.Username.Eq
		f.UsernameContains = w.Username.Contains
		f.UsernameStartsWith = w.Username.StartsWith
		if len(w.Username.In) > 0 {
			f.UsernameIn = w.Username.In
		}
	}
	if w.CreatedAt != nil {
		f.CreatedAtEq = w.CreatedAt.Eq
		f.CreatedAtGte = w.CreatedAt.Gte
		f.CreatedAtLte = w.CreatedAt.Lte
	}
	return f
}

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
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
	case appEnduser.ErrBuiltinUserCannotBeDisabled.GetCode():
		return &generated.BuiltinUserCannotBeDisabled{Message: err.Msg()}
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
	if err.Info().GetCode() == appEnduser.ErrBuiltinUserCannotBeDeleted.GetCode() {
		return &generated.BuiltinUserCannotBeDeleted{Message: err.Msg()}
	}
	return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
}

func convertOrgListEndUsersError(err *bizerrors.BusinessError) generated.ListEndUsersError {
	if err == nil {
		return nil
	}
	return &generated.InvalidInput{Message: err.Msg()}
}

func convertOrgResetEndUserPasswordError(err *bizerrors.BusinessError) generated.ResetEndUserPasswordError {
	if err == nil {
		return nil
	}
	switch err.Info().GetCode() {
	case bizerrors.EndUserNotFound.GetCode(), bizerrors.NotFound.GetCode():
		return &generated.ResourceNotFound{Message: err.Msg(), ResourceType: generated.ResourceTypeEndUser}
	case appEnduser.ErrBuiltinUserCannotBeDisabled.GetCode():
		return &generated.BuiltinUserCannotBeDisabled{Message: err.Msg()}
	case bizerrors.EndUserParamInvalid.GetCode(), bizerrors.ParamInvalid.GetCode():
		suggestion := "Use at least 8 characters containing letters and digits"
		return &generated.EndUserPasswordTooWeak{Message: err.Msg(), Suggestion: &suggestion}
	default:
		return &generated.InvalidInput{Message: err.Msg()}
	}
}

func toOrgOptionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}
