package adapter

import (
	appAuth "modelcraft/internal/app/auth"
	domainauth "modelcraft/internal/domain/auth"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"strconv"
)

type apiKeyMapper struct{}

// APIKeyMapperInstance is the singleton mapper for APIKey conversions.
var APIKeyMapperInstance = &apiKeyMapper{}

// ConvertAPIKeyToGraphQL converts a domain APIKey to the GraphQL APIKey type.
func (m *apiKeyMapper) ConvertAPIKeyToGraphQL(key *domainauth.APIKey) *generated.APIKey {
	if key == nil {
		return nil
	}
	gql := &generated.APIKey{
		ID:        key.ID,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		RoleIDs:   convertRoleIDsToGraphQL(key.RoleIDs),
		CreatedAt: key.CreatedAt,
	}
	gql.LastUsedAt = key.LastUsedAt
	gql.ExpiresAt = key.ExpiresAt
	gql.RevokedAt = key.RevokedAt
	return gql
}

// ConvertAPIKeysToGraphQL converts a slice of domain APIKeys to GraphQL APIKey types.
func (m *apiKeyMapper) ConvertAPIKeysToGraphQL(keys []*domainauth.APIKey) []*generated.APIKey {
	result := make([]*generated.APIKey, len(keys))
	for i, k := range keys {
		result[i] = m.ConvertAPIKeyToGraphQL(k)
	}
	return result
}

// ConvertCreateResultToGraphQL converts a CreateAPIKeyResult to the GraphQL CreateAPIKeyResult type.
func (m *apiKeyMapper) ConvertCreateResultToGraphQL(result *appAuth.CreateAPIKeyResult) *generated.CreateAPIKeyResult {
	if result == nil {
		return nil
	}
	return &generated.CreateAPIKeyResult{
		ID:        result.Key.ID,
		Name:      result.Key.Name,
		Key:       result.PlainKey,
		KeyPrefix: result.Key.KeyPrefix,
		RoleIDs:   convertRoleIDsToGraphQL(result.Key.RoleIDs),
		CreatedAt: result.Key.CreatedAt,
	}
}

func convertRoleIDsToGraphQL(roleIDs []int) []string {
	if len(roleIDs) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		result = append(result, strconv.Itoa(roleID))
	}
	return result
}
