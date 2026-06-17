package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// version
// ---------------------------------------------------------------------------

func TestVersion_OutputsOKEnvelope(t *testing.T) {
	stdout, _, code := mc(t, "version")
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	// version field must be present (may be "dev" in tests)
	if _, ok := data["version"]; !ok {
		t.Errorf("data.version is missing")
	}
}

// ---------------------------------------------------------------------------
// Error envelope shape
// ---------------------------------------------------------------------------

func TestErrorEnvelope_HasRequiredFields(t *testing.T) {
	// Trigger an error by calling a command without credentials.
	cp := credPath(t)
	stdout, _, code := mc(t, "auth", "status", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}

	v := mustJSON(t, stdout)
	if ok, _ := v["ok"].(bool); ok {
		t.Fatal("expected ok=false")
	}

	errObj, _ := v["error"].(map[string]any)
	for _, field := range []string{"code", "message", "retryable", "suggestion"} {
		if _, ok := errObj[field]; !ok {
			t.Errorf("error.%s is missing", field)
		}
	}
}

// ---------------------------------------------------------------------------
// Exit codes
// ---------------------------------------------------------------------------

func TestExitCode_UnauthenticatedIs3(t *testing.T) {
	cp := credPath(t)
	_, _, code := mc(t, "auth", "status", "--credentials", cp)
	if code != 3 {
		t.Errorf("exit code = %d, want 3 (UNAUTHENTICATED)", code)
	}
}

func TestExitCode_InvalidArgumentIs2(t *testing.T) {
	cp := credPath(t)
	writeValidCreds(t, cp, "http://localhost", "dev")
	_, _, code := mc(t,
		"run", "dev.maindb.User",
		"--where", "not-json",
		"--credentials", cp,
	)
	if code != 2 {
		t.Errorf("exit code = %d, want 2 (INVALID_JSON_FLAG)", code)
	}
}

func TestExitCode_NoProjectContextIs5(t *testing.T) {
	cp := credPath(t)
	writeCredJSON(t, cp, map[string]any{
		"server":      "http://localhost",
		"orgName":     "acme",
		"accessToken": "at1",
		"expiresAt":   futureExpiry(),
	})
	_, _, code := mc(t, "run", "maindb.User", "{ findMany { id } }", "--credentials", cp)
	if code != 5 {
		t.Errorf("exit code = %d, want 5 (NO_PROJECT_CONTEXT)", code)
	}
}

// ---------------------------------------------------------------------------
// Upstream error propagation
// ---------------------------------------------------------------------------

func TestUpstreamError_401_MapsToUnauthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"UNAUTHORIZED","message":"token expired"}`))
	}))
	defer srv.Close()

	cp := credPath(t)
	writeCredJSON(t, cp, map[string]any{
		"server":         srv.URL,
		"orgName":        "acme",
		"accessToken":    "expired",
		"currentProject": "dev",
	})

	stdout, _, code := mc(t, "run", "dev.maindb.User", "{ findMany { id } }", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "UNAUTHENTICATED")
}

func TestUpstreamError_404_MapsToNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"NOT_FOUND","message":"model not found"}`))
	}))
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "run", "dev.maindb.User", "{ findMany { id } }", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "NOT_FOUND")
}

func TestUpstreamError_503_MapsToServiceUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"message":"service down"}`))
	}))
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "run", "dev.maindb.User", "{ findMany { id } }", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "SERVICE_UNAVAILABLE")
}

// ---------------------------------------------------------------------------
// Unknown command
// ---------------------------------------------------------------------------

func TestUnknownCommand_ReturnsInvalidArgument(t *testing.T) {
	stdout, _, code := mc(t, "notacommand")
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "INVALID_ARGUMENT")
}
