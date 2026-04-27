package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/propagation"

	"modelcraft-gateway/internal/auth"
	"modelcraft-gateway/internal/middleware"
)

// RESTHandler proxies protected REST endpoints to the backend.
// It validates the Gateway-issued JWT, extracts the userID from claims,
// injects it as X-User-ID, and forwards the request via reverse proxy.
// The backend trusts X-User-ID because it is only reachable via the Gateway.
type RESTHandler struct {
	proxy   *httputil.ReverseProxy
	authSvc *auth.Service
}

func NewRESTHandler(backendURL string, authSvc *auth.Service) (*RESTHandler, error) {
	target, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}
	return &RESTHandler{
		proxy:   httputil.NewSingleHostReverseProxy(target),
		authSvc: authSvc,
	}, nil
}

func (h *RESTHandler) Handle(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		restWriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing access token")
		return
	}

	claims, err := h.authSvc.VerifyAccessToken(token)
	if err != nil {
		restWriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
		return
	}

	// Strip Gateway JWT; backend identifies the caller via X-User-ID only.
	r.Header.Del("Authorization")
	r.Header.Set("X-User-ID", claims.UserID)

	// Propagate or generate a request ID for end-to-end tracing.
	if reqID := chiMiddleware.GetReqID(r.Context()); reqID != "" {
		r.Header.Set("X-Request-Id", reqID)
	}

	// Propagate the original client request ID for cross-layer tracing.
	if clientReqID := middleware.GetClientRequestID(r); clientReqID != "" {
		r.Header.Set("X-Client-Request-Id", clientReqID)
	}

	// Inject W3C traceparent/tracestate from the active OTel span so the backend
	// can parse trace_id and span_id from the Traceparent header.
	propagation.TraceContext{}.Inject(r.Context(), propagation.HeaderCarrier(r.Header))

	h.proxy.ServeHTTP(w, r)
}

type restErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func restWriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(restErrorResponse{Code: code, Message: message})
}
