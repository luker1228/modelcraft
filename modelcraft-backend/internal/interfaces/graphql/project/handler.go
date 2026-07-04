package projectgraphql

import (
	"modelcraft/internal/interfaces/graphql/project/generated"
	"net/http"

	graphqlutil "modelcraft/internal/interfaces/graphql"
	playgroundpkg "modelcraft/pkg/graphql"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi/v5"
)

// ProjectGraphQLHandler creates a handler for the project domain GraphQL endpoint (internal link).
func ProjectGraphQLHandler(resolver *Resolver) http.HandlerFunc {
	hasPermissionDirective := NewHasPermissionDirective(resolver.UserRoleService)
	config := generated.Config{Resolvers: resolver}
	config.Directives.HasPermission = hasPermissionDirective.HasPermission
	h := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	h.AroundResponses(graphqlutil.InjectRequestIDExtension)
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// ProjectEndUserGraphQLHandler creates a handler for the project domain GraphQL endpoint (EndUser link).
// Uses NewEndUserHasPermissionDirective to enforce allowEndUser gating.
func ProjectEndUserGraphQLHandler(resolver *Resolver) http.HandlerFunc {
	hasPermissionDirective := NewEndUserHasPermissionDirective(resolver.UserRoleService)
	config := generated.Config{Resolvers: resolver}
	config.Directives.HasPermission = hasPermissionDirective.HasPermission
	h := handler.NewDefaultServer(generated.NewExecutableSchema(config))
	h.AroundResponses(graphqlutil.InjectRequestIDExtension)
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// ProjectPlaygroundHandler creates a handler for the project GraphQL playground.
func ProjectPlaygroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "orgName")
		if orgName == "" {
			orgName = "default"
		}
		projectSlug := chi.URLParam(r, "projectSlug")
		if projectSlug == "" {
			projectSlug = "default"
		}
		endpoint := "/org/" + orgName + "/project/" + projectSlug + "/graphql"
		httpHandler := playgroundpkg.HTTPHandler(playgroundpkg.PlaygroundConfig{
			Endpoint: endpoint,
			Title:    "GraphQL Playground - Project API (" + orgName + "/" + projectSlug + ")",
		})
		httpHandler(w, r)
	}
}
