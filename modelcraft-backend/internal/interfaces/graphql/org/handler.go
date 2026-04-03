package orggraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/ctxutils"
	"net/http"

	playgroundpkg "modelcraft/pkg/graphql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
)

// injectRequestIDMiddleware adds requestId to GraphQL response extensions
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

// OrgGraphQLHandler creates GraphQL handler for org domain
func OrgGraphQLHandler(resolver *Resolver) http.HandlerFunc {
	hasPermissionDirective := NewHasPermissionDirective(resolver.UserRoleService)
	config := generated.Config{Resolvers: resolver}
	config.Directives.HasPermission = hasPermissionDirective.HasPermission
	h := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	h.AroundResponses(injectRequestIDMiddleware)
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// OrgPlaygroundHandler serves GraphQL Playground for org domain
func OrgPlaygroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "orgName")
		if orgName == "" {
			orgName = "default"
		}
		endpoint := "/org/" + orgName + "/graphql"
		ginHandler := playgroundpkg.Handler(playgroundpkg.PlaygroundConfig{
			Endpoint: endpoint,
			Title:    "GraphQL Playground - Org API (" + orgName + ")",
		})
		c, _ := gin.CreateTestContext(w)
		c.Request = r
		ginHandler(c)
	}
}
