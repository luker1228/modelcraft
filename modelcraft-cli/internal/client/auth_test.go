package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthClientWhoamiPopulatesCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cli/end-user/auth/whoami" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer mc_pat_test" {
			t.Fatalf("unexpected Authorization: %s", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"userId":"u1","orgName":"acme","projects":[{"slug":"sales","title":"Sales"}]}`))
	}))
	defer srv.Close()

	c := AuthClient{HTTPClient: srv.Client()}
	creds, err := c.Whoami(context.Background(), srv.URL, "mc_pat_test")
	if err != nil {
		t.Fatalf("Whoami() error = %v", err)
	}
	if creds.Server != srv.URL || creds.OrgName != "acme" || creds.UserID != "u1" {
		t.Fatalf("unexpected creds: %+v", creds)
	}
}

func TestAuthClientWhoamiReturnsUnauthenticatedOn401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"UNAUTHORIZED","message":"invalid token"}`))
	}))
	defer srv.Close()

	c := AuthClient{HTTPClient: srv.Client()}
	_, err := c.Whoami(context.Background(), srv.URL, "bad-token")
	if err == nil {
		t.Fatal("Whoami() error = nil, want error")
	}
}
