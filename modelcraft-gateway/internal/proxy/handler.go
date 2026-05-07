package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"

	"modelcraft-gateway/internal/auth"
	"modelcraft-gateway/internal/middleware"
)

type contextKey string

const (
	userIDContextKey   contextKey = "user_id"
	userTypeContextKey contextKey = "user_type"
)

const userTypeEndUser = "end_user"

// Handler is a reverse proxy that forwards requests to the Go backend.
// After validating the Bearer token, it injects X-User-ID (from JWT claims)
// and strips the Authorization header before forwarding upstream.
type Handler struct {
	backendURL    *url.URL
	authService   *auth.Service
	internalToken string
	reverseProxy  *httputil.ReverseProxy
}

func NewHandler(backendURL, internalToken string, authService *auth.Service) (*Handler, error) {
	parsed, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		backendURL:    parsed,
		authService:   authService,
		internalToken: internalToken,
	}
	h.reverseProxy = &httputil.ReverseProxy{
		Director: h.director,
	}
	return h, nil
}

// director rewrites the incoming request to target the upstream backend.
func (h *Handler) director(req *http.Request) {
	req.URL.Scheme = h.backendURL.Scheme
	req.URL.Host = h.backendURL.Host

	// Strip Gateway JWT; backend identifies the caller via X-User-ID only.
	req.Header.Del("Authorization")

	// Inject the authenticated user ID so the backend can identify the caller.
	if userID, ok := req.Context().Value(userIDContextKey).(string); ok && userID != "" {
		req.Header.Set("X-User-ID", userID)
		middleware.LoggerFromCtx(req.Context()).Debug("gateway: injected X-User-ID into upstream request",
			zap.String("user_id", userID),
			zap.String("upstream", req.URL.String()),
		)
	}

	// Inject user type so the backend can distinguish EndUser from tenant callers.
	if userType, ok := req.Context().Value(userTypeContextKey).(string); ok && userType != "" {
		req.Header.Set("X-User-Type", userType)
	}

	// Propagate the internal request ID for end-to-end tracing.
	if reqID := chiMiddleware.GetReqID(req.Context()); reqID != "" {
		req.Header.Set("X-Request-Id", reqID)
	}

	// Propagate the original client request ID for cross-layer tracing.
	if clientReqID := middleware.GetClientRequestID(req); clientReqID != "" {
		req.Header.Set("X-Client-Request-Id", clientReqID)
	}

	// Inject W3C traceparent/tracestate from the active OTel span so the backend
	// can parse trace_id and span_id from the Traceparent header.
	propagation.TraceContext{}.Inject(req.Context(), propagation.HeaderCarrier(req.Header))

	// Preserve the original host for virtual-hosting setups.
	req.Host = h.backendURL.Host
}

// GraphQLOrgHandler validates the gateway JWT, then proxies to /graphql/org/{orgName}/.
func (h *Handler) GraphQLOrgHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.extractAndVerify(w, r)
	if !ok {
		return
	}
	ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
	middleware.LoggerFromCtx(ctx).Info("gateway: GraphQL org request authenticated",
		zap.String("user_id", claims.UserID),
		zap.String("path", r.URL.Path),
	)
	h.reverseProxy.ServeHTTP(w, r.WithContext(ctx))
}

// GraphQLProjectHandler validates the gateway JWT, then proxies to
// /graphql/org/{orgName}/project/{projectSlug}/.
func (h *Handler) GraphQLProjectHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.extractAndVerify(w, r)
	if !ok {
		return
	}
	ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
	// Propagate user type so the backend @hasPermission directive can identify EndUser callers.
	for _, aud := range claims.Audience {
		if aud == userTypeEndUser {
			ctx = context.WithValue(ctx, userTypeContextKey, userTypeEndUser)
			break
		}
	}
	middleware.LoggerFromCtx(ctx).Info("gateway: GraphQL project request authenticated",
		zap.String("user_id", claims.UserID),
		zap.String("path", r.URL.Path),
	)
	h.reverseProxy.ServeHTTP(w, r.WithContext(ctx))
}

// GraphQLEndUserProjectHandler validates the end-user JWT and proxies to the
// same project GraphQL backend as GraphQLProjectHandler, rewriting the path to
// strip the /end-user/ segment.
//
// Incoming:  /graphql/end-user/org/{orgName}/project/{projectSlug}
// Upstream:  /graphql/org/{orgName}/project/{projectSlug}
func (h *Handler) GraphQLEndUserProjectHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.extractAndVerify(w, r)
	if !ok {
		return
	}
	ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
	// End-user route always injects X-User-Type: end_user so the backend can bypass RBAC.
	ctx = context.WithValue(ctx, userTypeContextKey, userTypeEndUser)
	middleware.LoggerFromCtx(ctx).Info("gateway: GraphQL end-user project request authenticated",
		zap.String("user_id", claims.UserID),
		zap.String("path", r.URL.Path),
	)
	// Rewrite: /graphql/end-user/org/... → /graphql/org/...
	r = r.WithContext(ctx)
	r.URL.Path = strings.Replace(r.URL.Path, "/graphql/end-user/", "/graphql/", 1)
	h.reverseProxy.ServeHTTP(w, r)
}

// GraphQLEndUserOrgHandler validates the end-user JWT and proxies to the
// same org GraphQL backend as GraphQLOrgHandler, rewriting the path to
// strip the /end-user/ segment.
//
// Incoming:  /graphql/end-user/org/{orgName}
// Upstream:  /graphql/org/{orgName}
func (h *Handler) GraphQLEndUserOrgHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.extractAndVerify(w, r)
	if !ok {
		return
	}
	ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
	ctx = context.WithValue(ctx, userTypeContextKey, userTypeEndUser)
	middleware.LoggerFromCtx(ctx).Info("gateway: GraphQL end-user org request authenticated",
		zap.String("user_id", claims.UserID),
		zap.String("path", r.URL.Path),
	)
	// Rewrite: /graphql/end-user/org/... → /graphql/org/...
	r = r.WithContext(ctx)
	r.URL.Path = strings.Replace(r.URL.Path, "/graphql/end-user/", "/graphql/", 1)
	h.reverseProxy.ServeHTTP(w, r)
}

func (h *Handler) extractAndVerify(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		middleware.LoggerFromCtx(r.Context()).Warn("gateway: missing Authorization header",
			zap.String("path", r.URL.Path),
		)
		writeError(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header required")
		return nil, false
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.authService.VerifyAccessToken(tokenStr)
	if err != nil {
		middleware.LoggerFromCtx(r.Context()).Warn("gateway: invalid or expired access token",
			zap.Error(err),
			zap.String("path", r.URL.Path),
		)
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired access token")
		return nil, false
	}
	return claims, true
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"code":"` + code + `","message":"` + message + `"}`))
}
