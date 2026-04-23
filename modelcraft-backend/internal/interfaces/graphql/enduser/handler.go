package endusergraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/enduser/generated"
	"modelcraft/pkg/ctxutils"
	"net/http"
	"strings"

	playgroundpkg "modelcraft/pkg/graphql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
)

func injectRequestIDMiddleware(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	resp := next(ctx)
	requestID := ctxutils.GetRequestID(ctx)
	if requestID == "" {
		return resp
	}
	if resp.Extensions == nil {
		resp.Extensions = make(map[string]any)
	}
	resp.Extensions["requestId"] = requestID
	return resp
}

// EndUserGraphQLHandler creates a handler for the end-user GraphQL endpoint.
func EndUserGraphQLHandler(resolver *Resolver) http.HandlerFunc {
	config := generated.Config{Resolvers: resolver}
	h := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	h.AroundResponses(injectRequestIDMiddleware)
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// EndUserPlaygroundHandler creates a handler for the end-user GraphQL playground.
func EndUserPlaygroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "orgName")
		if orgName == "" {
			orgName = "default"
		}
		projectSlug := chi.URLParam(r, "projectSlug")
		if projectSlug == "" {
			projectSlug = "default"
		}
		endpoint := strings.TrimSuffix(r.URL.Path, "/")
		if endpoint == "" {
			endpoint = "/graphql/end-user/org/" + orgName + "/project/" + projectSlug
		}
		ginHandler := playgroundpkg.Handler(playgroundpkg.PlaygroundConfig{
			Endpoint: endpoint,
			Title:    "GraphQL Playground - End-User API (" + orgName + "/" + projectSlug + ")",
		})
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		ginHandler(c)
	}
}
