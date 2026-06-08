package graphqlutil

import (
	"context"
	"encoding/json"
	"modelcraft/pkg/ctxutils"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/require"
)

func TestInjectRequestIDExtension(t *testing.T) {
	ctx := ctxutils.NewHttpContext(context.Background(), &ctxutils.HttpRequestContext{
		RequestId: "test-request-id",
	})

	resp := InjectRequestIDExtension(ctx, func(context.Context) *graphql.Response {
		return &graphql.Response{}
	})

	require.NotNil(t, resp.Extensions)

	raw, ok := resp.Extensions["requestId"]
	require.True(t, ok)

	var got string
	data, err := json.Marshal(raw)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &got))
	require.Equal(t, "test-request-id", got)
}
