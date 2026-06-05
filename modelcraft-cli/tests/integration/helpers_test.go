// Package integration provides end-to-end tests that compile and run the mc
// binary against a local httptest server. Each test is self-contained: it
// spins up its own mock HTTP server, writes a temporary credentials file, and
// invokes the real binary.
//
// Run:
//
//	go test ./tests/integration/... -v
//
// The binary is rebuilt once per test run (TestMain).
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// binPath holds the path to the compiled mc binary, set during TestMain.
var binPath string

// TestMain compiles the binary once before all integration tests run.
func TestMain(m *testing.M) {
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..")

	bin := filepath.Join(repoRoot, "bin", "mc-inttest")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	out, err := exec.Command("go", "build", "-o", bin, repoRoot).CombinedOutput()
	if err != nil {
		panic("failed to build mc binary:\n" + string(out))
	}
	binPath = bin

	code := m.Run()
	_ = os.Remove(bin)
	os.Exit(code)
}

// mc runs the mc binary with the given args and returns stdout, stderr, exit code.
func mc(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(binPath, args...) //nolint:gosec
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	_ = cmd.Run()
	return outBuf.String(), errBuf.String(), cmd.ProcessState.ExitCode()
}

// mustJSON parses stdout as JSON and fails the test if it is invalid.
func mustJSON(t *testing.T, s string) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nraw: %s", err, s)
	}
	return v
}

// assertOK fails if the envelope does not carry ok=true.
func assertOK(t *testing.T, v map[string]any) {
	t.Helper()
	if ok, _ := v["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got: %v", v)
	}
}

// assertErrorCode fails if ok != false or the error code doesn't match.
func assertErrorCode(t *testing.T, v map[string]any, wantCode string) {
	t.Helper()
	if ok, _ := v["ok"].(bool); ok {
		t.Fatalf("expected ok=false, got ok=true: %v", v)
	}
	errObj, _ := v["error"].(map[string]any)
	if errObj == nil {
		t.Fatalf("missing 'error' field: %v", v)
	}
	if code, _ := errObj["code"].(string); code != wantCode {
		t.Fatalf("error.code = %q, want %q\nfull: %v", code, wantCode, v)
	}
}

// futureExpiry returns an RFC 3339 timestamp one day from now (no refresh needed).
func futureExpiry() string {
	return time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
}

// writeCredJSON writes a JSON credentials file to path with base fields merged
// with any extras provided by the caller.
func writeCredJSON(t *testing.T, path string, fields map[string]any) {
	t.Helper()
	b, _ := json.Marshal(fields)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("writeCredJSON: %v", err)
	}
}

// credPath returns a standard temp credentials path for a test.
func credPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "credentials.json")
}

// writeValidCreds writes a ready-to-use credentials file pointing to serverURL.
// The token is not expired so no refresh will be attempted.
func writeValidCreds(t *testing.T, path, serverURL, project string) {
	t.Helper()
	writeCredJSON(t, path, map[string]any{
		"server":         serverURL,
		"orgName":        "acme",
		"userId":         "u1",
		"accessToken":    "at-valid",
		"refreshToken":   "rt-valid",
		"expiresAt":      futureExpiry(),
		"currentProject": project,
		"projects":       []map[string]any{{"slug": project, "title": "Test Project"}},
	})
}

// newAuthServer returns an httptest.Server whose handler:
//   - GET  /api/cli/end-user/auth/whoami  → whoamiResp with projects
//   - POST /api/cli/end-user/auth/logout  → 204
//   - POST /api/cli/end-user/auth/refresh → refreshResp
//   - all other paths                     → graphqlHandler (if non-nil)
func newAuthServer(t *testing.T, projects []map[string]any, graphqlHandler http.HandlerFunc) *httptest.Server {
	t.Helper()
	whoami := map[string]any{
		"userId":   "u1",
		"orgName":  "acme",
		"isAdmin":  false,
		"projects": projects,
	}
	refresh := map[string]any{
		"userId":       "u1",
		"orgName":      "acme",
		"accessToken":  "at-refreshed",
		"refreshToken": "rt-refreshed",
		"expiresAt":    futureExpiry(),
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/cli/end-user/auth/whoami":
			_ = json.NewEncoder(w).Encode(whoami)
		case r.Method == http.MethodPost && r.URL.Path == "/api/cli/end-user/auth/refresh":
			_ = json.NewEncoder(w).Encode(refresh)
		case r.Method == http.MethodPost && r.URL.Path == "/api/cli/end-user/auth/logout":
			w.WriteHeader(http.StatusNoContent)
		default:
			if graphqlHandler != nil {
				graphqlHandler(w, r)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))
}

// newGraphQLServer returns an httptest.Server that responds to every request
// with {"data": data}.
func newGraphQLServer(t *testing.T, data map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
}
