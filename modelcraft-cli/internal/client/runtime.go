package client

import (
	"context"
	"fmt"
	"strings"
)

// Run executes a raw GraphQL query string against the model-scoped runtime endpoint.
// gqlBody is the bare GraphQL query, e.g. `{ findMany(take: 5) { id name } }`.
// The response data field is returned as-is so callers can render any shape.
func (c GraphQLClient) Run(ctx context.Context, server, org, project, db, model, token, gqlBody string) (map[string]any, error) {
	endpoint := modelEndpoint(server, org, project, db, model)

	var data map[string]any
	if err := c.Execute(ctx, endpoint, token, gqlBody, nil, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// modelEndpoint builds the model-scoped runtime GraphQL endpoint URL.
func modelEndpoint(server, org, project, db, model string) string {
	return fmt.Sprintf("%s/end-user/graphql/org/%s/project/%s/db/%s/model/%s", strings.TrimRight(server, "/"), org, project, db, model)
}
