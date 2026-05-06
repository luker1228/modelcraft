package projectgraphql

import (
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/interfaces/graphql/project/generated"
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
