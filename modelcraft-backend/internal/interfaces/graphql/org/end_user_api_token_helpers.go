package orggraphql

import (
	domainenduser "modelcraft/internal/domain/enduser"
	"modelcraft/internal/interfaces/graphql/org/generated"
)

func toGQLAPIToken(t *domainenduser.APIToken) *generated.EndUserAPIToken {
	return &generated.EndUserAPIToken{
		ID:         t.ID,
		Name:       t.Name,
		CreatedAt:  t.CreatedAt,
		ExpiresAt:  t.ExpiresAt,
		LastUsedAt: t.LastUsedAt,
	}
}
