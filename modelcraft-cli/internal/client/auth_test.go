package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthClientLoginPopulatesServerAndOrg(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cli/end-user/auth/login" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"requestId":"r1","userId":"u1","accessToken":"a1","refreshToken":"rt1","expiresAt":"2026-05-10T10:00:00Z","projects":[]}`))
	}))
	defer srv.Close()

	c := AuthClient{HTTPClient: srv.Client()}
	creds, err := c.Login(context.Background(), srv.URL, "acme", "alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if creds.Server != srv.URL || creds.OrgName != "acme" {
		t.Fatalf("unexpected creds: %+v", creds)
	}
}

func TestAuthClientRefreshReturnsUnauthenticatedOn401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"UNAUTHORIZED","message":"invalid token"}`))
	}))
	defer srv.Close()

	c := AuthClient{HTTPClient: srv.Client()}
	_, err := c.Refresh(context.Background(), srv.URL, "acme", "bad-token")
	if err == nil {
		t.Fatal("Refresh() error = nil, want error")
	}
}
