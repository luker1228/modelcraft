package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImpersonationEnabled(t *testing.T) {
	tests := []struct {
		imp Impersonation
		want bool
	}{
		{Impersonation{}, false},
		{Impersonation{UserID: "u1"}, true},
		{Impersonation{UserName: "alice"}, true},
		{Impersonation{Roles: "admin"}, true},
	}
	for _, tt := range tests {
		if got := tt.imp.Enabled(); got != tt.want {
			t.Errorf("Impersonation%v.Enabled() = %v, want %v", tt.imp, got, tt.want)
		}
	}
}

func TestExecuteInjectsImpersonationHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("X-MC-Auth-Userid-Str"); v != "user_abc" {
			t.Errorf("X-MC-Auth-Userid-Str = %q, want %q", v, "user_abc")
		}
		if v := r.Header.Get("X-MC-Auth-Username"); v != "alice" {
			t.Errorf("X-MC-Auth-Username = %q, want %q", v, "alice")
		}
		if v := r.Header.Get("X-MC-Auth-Roles"); v != "admin,manager" {
			t.Errorf("X-MC-Auth-Roles = %q, want %q", v, "admin,manager")
		}
		_, _ = w.Write([]byte(`{"data":null}`))
	}))
	defer srv.Close()

	c := GraphQLClient{
		HTTPClient: srv.Client(),
		Impersonation: Impersonation{
			UserID:   "user_abc",
			UserName: "alice",
			Roles:    "admin,manager",
		},
	}
	if err := c.Execute(context.Background(), srv.URL, "tok", "{ __typename }", nil, nil); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestExecuteInjectsNumericUserIDAsIntHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("X-MC-Auth-Userid-Int"); v != "42" {
			t.Errorf("X-MC-Auth-Userid-Int = %q, want %q", v, "42")
		}
		if v := r.Header.Get("X-MC-Auth-Userid-Str"); v != "" {
			t.Errorf("X-MC-Auth-Userid-Str should be empty, got %q", v)
		}
		_, _ = w.Write([]byte(`{"data":null}`))
	}))
	defer srv.Close()

	c := GraphQLClient{
		HTTPClient:    srv.Client(),
		Impersonation: Impersonation{UserID: "42"},
	}
	if err := c.Execute(context.Background(), srv.URL, "tok", "{ __typename }", nil, nil); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestExecuteWithoutImpersonationSendsNoAuthHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range []string{"X-MC-Auth-Userid-Int", "X-MC-Auth-Userid-Str", "X-MC-Auth-Username", "X-MC-Auth-Roles"} {
			if v := r.Header.Get(h); v != "" {
				t.Errorf("%s should be empty, got %q", h, v)
			}
		}
		_, _ = w.Write([]byte(`{"data":null}`))
	}))
	defer srv.Close()

	c := GraphQLClient{HTTPClient: srv.Client()}
	if err := c.Execute(context.Background(), srv.URL, "tok", "{ __typename }", nil, nil); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
