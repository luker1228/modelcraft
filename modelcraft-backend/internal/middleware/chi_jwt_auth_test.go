package middleware

import (
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/httpheader"
	"net/http"
	"net/http/httptest"
	"testing"
)

// sentinel handler that records whether it was reached and the userID in context.
func sentinelHandler(t *testing.T, reached *bool, gotUserID *string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*reached = true
		if uid, err := ctxutils.GetUserIDFromContext(r.Context()); err == nil {
			*gotUserID = uid
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestChiJWTAuthMiddleware_GatewayTrustedRequest(t *testing.T) {
	// Gateway-trusted request: X-User-ID injected, no Authorization header.
	// Backend MUST accept the request and propagate the userID.
	reached := false
	gotUserID := ""
	mw := ChiJWTAuthMiddleware()
	handler := mw(sentinelHandler(t, &reached, &gotUserID))

	req := httptest.NewRequest(http.MethodPost, "/graphql/org/acme/", nil)
	req.Header.Set(httpheader.XUserID, "user-abc-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !reached {
		t.Fatal("expected downstream handler to be reached")
	}
	if gotUserID != "user-abc-123" {
		t.Fatalf("expected userID 'user-abc-123' in context, got '%s'", gotUserID)
	}
}

func TestChiJWTAuthMiddleware_DirectBearerTokenRejected(t *testing.T) {
	// Direct bearer token: no X-User-ID header, only Authorization: Bearer <token>.
	// Backend MUST reject this request — direct JWT validation is no longer supported.
	reached := false
	gotUserID := ""
	mw := ChiJWTAuthMiddleware()
	handler := mw(sentinelHandler(t, &reached, &gotUserID))

	req := httptest.NewRequest(http.MethodPost, "/graphql/org/acme/", nil)
	req.Header.Set(httpheader.Authorization, "Bearer sometoken.payload.sig")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if reached {
		t.Fatal("downstream handler must NOT be reached for a direct bearer token request")
	}
}

func TestChiJWTAuthMiddleware_NoCredentials(t *testing.T) {
	// Request with neither X-User-ID nor Authorization header must be rejected.
	reached := false
	gotUserID := ""
	mw := ChiJWTAuthMiddleware()
	handler := mw(sentinelHandler(t, &reached, &gotUserID))

	req := httptest.NewRequest(http.MethodPost, "/graphql/org/acme/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if reached {
		t.Fatal("downstream handler must NOT be reached with no credentials")
	}
}

