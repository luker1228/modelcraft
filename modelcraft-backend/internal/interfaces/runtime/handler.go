// Package runtime provides HTTP handlers for the model runtime GraphQL API.
// The runtime API serves dynamically generated GraphQL schemas derived from
// model definitions stored in the system.
package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"modelcraft/internal/app/modelruntime"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"net/http"

	playgroundpkg "modelcraft/pkg/graphql"

	"github.com/go-chi/chi/v5"
)

// runtimeContextKey is the context key for runtime context values.
type runtimeContextKey struct{}

// runtimeContext holds request-scoped runtime information.
type runtimeContext struct {
	OrgName     string
	ProjectSlug string
}

// WithRuntimeContext adds runtime context (orgName, projectSlug) to the context.
func WithRuntimeContext(ctx context.Context, orgName, projectSlug string) context.Context {
	rctx := &runtimeContext{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	}
	return context.WithValue(ctx, runtimeContextKey{}, rctx)
}

// RuntimeGraphQLRequest represents a GraphQL request for the runtime API.
type RuntimeGraphQLRequest struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
	OperationName string         `json:"operationName,omitempty"`
}

// ModelRuntimeHandler handles dynamic GraphQL requests for model runtime.
// Unlike the design-time GraphQL (gqlgen with static schema), this handler
// serves a dynamically generated schema derived from the system's own model definitions.
type ModelRuntimeHandler struct {
	graphqlAppService *modelruntime.GraphqlAppService
}

// NewModelRuntimeHandler creates a new ModelRuntimeHandler.
func NewModelRuntimeHandler(graphqlAppService *modelruntime.GraphqlAppService) *ModelRuntimeHandler {
	return &ModelRuntimeHandler{
		graphqlAppService: graphqlAppService,
	}
}

// HandlePlayground serves the GraphQL Playground for the runtime API (GET).
// URL params: orgName, projectSlug, db, model.
func (h *ModelRuntimeHandler) HandlePlayground(w http.ResponseWriter, r *http.Request) {
	orgName := chi.URLParam(r, "orgName")
	projectSlug := chi.URLParam(r, "projectSlug")
	db := chi.URLParam(r, "db")
	model := chi.URLParam(r, "model")

	if orgName == "" || projectSlug == "" || db == "" || model == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "Missing required parameters",
			"message": "orgName, projectSlug, db, and model are required",
		})
		return
	}

	endpoint := fmt.Sprintf("/graphql/org/%s/project/%s/db/%s/model/%s", orgName, projectSlug, db, model)
	title := fmt.Sprintf(
		"GraphQL Playground - %s (%s) [Project: %s, Org: %s]",
		model, db, projectSlug, orgName,
	)

	playgroundpkg.HTTPHandler(playgroundpkg.PlaygroundConfig{
		Endpoint: endpoint,
		Title:    title,
	}).ServeHTTP(w, r)
}

// HandleQuery executes a GraphQL query against the runtime schema (POST).
// URL params: orgName, projectSlug, db, model.
func (h *ModelRuntimeHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	logger := logfacade.GetLogger(r.Context())

	orgName := chi.URLParam(r, "orgName")
	projectSlug := chi.URLParam(r, "projectSlug")
	db := chi.URLParam(r, "db")
	model := chi.URLParam(r, "model")

	if orgName == "" || projectSlug == "" || db == "" || model == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "orgName, projectSlug, db, and model are required",
		})
		return
	}

	var req RuntimeGraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON request body",
		})
		return
	}

	cmd := modelruntime.ExecuteGraphQLCommand{
		Query:         req.Query,
		Variables:     req.Variables,
		OperationName: req.OperationName,
	}

	// Inject runtime context (orgName, projectSlug) into context for RLS resolution
	ctx := WithRuntimeContext(r.Context(), orgName, projectSlug)

	result, err := h.graphqlAppService.Execute(ctx, orgName, projectSlug, model, db, cmd)
	if err != nil {
		statusCode := http.StatusInternalServerError
		var bizErr *bizerrors.BusinessError
		if errors.As(err, &bizErr) {
			statusCode = bizErr.GetHTTPStatusCode()
		}
		if statusCode >= 500 {
			logger.Error(r.Context(), "Runtime GraphQL execution failed", logfacade.Err(err), logfacade.Stack(err))
		}
		requestID := ctxutils.GetRequestID(r.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message":   err.Error(),
			"requestId": requestID,
		})
		return
	}

	// Inject requestId into response extensions (consistent with design-time GraphQL handlers)
	if requestID := ctxutils.GetRequestID(r.Context()); requestID != "" {
		if result.Extensions == nil {
			result.Extensions = make(map[string]interface{})
		}
		result.Extensions["requestId"] = requestID
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
