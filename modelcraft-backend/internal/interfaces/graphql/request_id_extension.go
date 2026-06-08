package graphqlutil

import (
	"context"
	"modelcraft/pkg/ctxutils"

	"github.com/99designs/gqlgen/graphql"
)

// SetRequestIDExtension writes requestId into a GraphQL response extensions map.
func SetRequestIDExtension(ctx context.Context, extensions map[string]any) map[string]any {
	requestID := ctxutils.GetRequestID(ctx)
	if requestID == "" {
		return extensions
	}
	if extensions == nil {
		extensions = make(map[string]any)
	}
	extensions["requestId"] = requestID
	return extensions
}

// InjectRequestIDExtension adds requestId to GraphQL response extensions.
func InjectRequestIDExtension(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	resp := next(ctx)
	resp.Extensions = SetRequestIDExtension(ctx, resp.Extensions)
	return resp
}
