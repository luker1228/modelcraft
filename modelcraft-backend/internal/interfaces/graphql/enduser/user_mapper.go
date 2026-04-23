package endusergraphql

import (
	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/interfaces/graphql/enduser/generated"
	"time"
)

func toGraphqlRuntimeUser(src *appEnduser.EndUserDTO) *generated.RuntimeUser {
	if src == nil {
		return nil
	}

	return &generated.RuntimeUser{
		ID:          src.ID,
		Username:    src.Username,
		IsForbidden: src.IsForbidden,
		CreatedBy:   nullableString(src.CreatedBy),
		CreatedAt:   src.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   src.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
