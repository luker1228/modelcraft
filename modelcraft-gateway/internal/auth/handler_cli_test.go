package auth

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCLIEndUserLoginReturnsRefreshTokenInJSONBody(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/end-user/auth/login" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r1","userId":"u1","accessToken":"a1","refreshToken":"rt1","expiresAt":"2026-05-09T12:00:00Z","projects":[]}`))
	}))
	defer backend.Close()

	h := NewHandler(nil, backend.URL, backend.Client(), "")
	req := httptest.NewRequest(http.MethodPost, "/api/cli/end-user/auth/login", bytes.NewBufferString(`{"orgName":"acme","username":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()

	h.CLIEndUserLogin(rec, req)

	if got := rec.Header().Get("Set-Cookie"); got != "" {
		t.Fatalf("did not expect refresh cookie, got %q", got)
	}
	want := `{"requestId":"r1","userId":"u1","accessToken":"a1","refreshToken":"rt1","expiresAt":"2026-05-09T12:00:00Z","projects":[]}` + "\n"
	if rec.Body.String() != want {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestCLIEndUserRefreshReadsRefreshTokenFromBody(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, _ := io.ReadAll(r.Body)
		want := `{"orgName":"acme","refreshToken":"rt1"}`
		if string(buf) != want {
			t.Fatalf("unexpected request body: %s", buf)
		}
		_, _ = w.Write([]byte(`{"requestId":"r2","userId":"u1","accessToken":"a2","refreshToken":"rt2","expiresAt":"2026-05-09T13:00:00Z","projects":[]}`))
	}))
	defer backend.Close()

	h := NewHandler(nil, backend.URL, backend.Client(), "")
	req := httptest.NewRequest(http.MethodPost, "/api/cli/end-user/auth/refresh", bytes.NewBufferString(`{"orgName":"acme","refreshToken":"rt1"}`))
	rec := httptest.NewRecorder()

	h.CLIEndUserRefresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Set-Cookie"); got != "" {
		t.Fatalf("did not expect refresh cookie, got %q", got)
	}
}
