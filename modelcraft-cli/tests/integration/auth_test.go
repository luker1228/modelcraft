package integration

import (
	"testing"
)

// ---------------------------------------------------------------------------
// auth login
// ---------------------------------------------------------------------------

func TestAuthLogin_Success(t *testing.T) {
	projects := []map[string]any{{"slug": "sales", "title": "Sales"}}
	srv := newAuthServer(t, projects, nil)
	defer srv.Close()

	cp := credPath(t)
	stdout, _, code := mc(t,
		"auth", "login",
		"--server", srv.URL,
		"--org", "acme",
		"--username", "alice",
		"--password", "secret",
		"--credentials", cp,
	)

	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	if data["server"] != srv.URL {
		t.Errorf("data.server = %v, want %s", data["server"], srv.URL)
	}
	if data["orgName"] != "acme" {
		t.Errorf("data.orgName = %v, want acme", data["orgName"])
	}
}

func TestAuthLogin_MissingRequiredFlags(t *testing.T) {
	cp := credPath(t)
	// Omit --password
	stdout, _, code := mc(t,
		"auth", "login",
		"--server", "http://localhost",
		"--org", "acme",
		"--username", "alice",
		"--credentials", cp,
	)

	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "MISSING_REQUIRED_FLAG")
}

func TestAuthLogin_DoesNotAutoSelectProject(t *testing.T) {
	projects := []map[string]any{{"slug": "sales", "title": "Sales"}}
	srv := newAuthServer(t, projects, nil)
	defer srv.Close()

	cp := credPath(t)
	_, _, code := mc(t,
		"auth", "login",
		"--server", srv.URL,
		"--org", "acme",
		"--username", "alice",
		"--password", "secret",
		"--credentials", cp,
	)
	if code != 0 {
		t.Fatalf("login failed with code %d", code)
	}

	// currentProject must NOT be set automatically after login.
	stdout, _, code2 := mc(t, "auth", "status", "--credentials", cp)
	if code2 != 0 {
		t.Fatalf("status failed with code %d, stdout: %s", code2, stdout)
	}
	v := mustJSON(t, stdout)
	data, _ := v["data"].(map[string]any)
	if cur, _ := data["currentProject"].(string); cur != "" {
		t.Errorf("currentProject should be empty after login, got %q", cur)
	}
}

// ---------------------------------------------------------------------------
// auth status
// ---------------------------------------------------------------------------

func TestAuthStatus_NoCredentials(t *testing.T) {
	cp := credPath(t)
	stdout, _, code := mc(t, "auth", "status", "--credentials", cp)

	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "UNAUTHENTICATED")
}

func TestAuthStatus_WithValidCredentials(t *testing.T) {
	srv := newAuthServer(t, nil, nil)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "auth", "status", "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)
	data, _ := v["data"].(map[string]any)
	if data["orgName"] != "acme" {
		t.Errorf("data.orgName = %v, want acme", data["orgName"])
	}
}

// ---------------------------------------------------------------------------
// auth switch-project
// ---------------------------------------------------------------------------

func TestAuthSwitchProject_Success(t *testing.T) {
	cp := credPath(t)
	writeCredJSON(t, cp, map[string]any{
		"server":  "http://localhost",
		"orgName": "acme",
		"userId":  "u1",
		"projects": []map[string]any{
			{"slug": "sales", "title": "Sales"},
			{"slug": "hr", "title": "HR"},
		},
	})

	stdout, _, code := mc(t, "auth", "switch-project", "hr", "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)
	data, _ := v["data"].(map[string]any)
	if data["currentProject"] != "hr" {
		t.Errorf("data.currentProject = %v, want hr", data["currentProject"])
	}
}

func TestAuthSwitchProject_UnknownSlugRejected(t *testing.T) {
	cp := credPath(t)
	writeCredJSON(t, cp, map[string]any{
		"server":  "http://localhost",
		"orgName": "acme",
		"projects": []map[string]any{
			{"slug": "sales", "title": "Sales"},
		},
	})

	stdout, _, code := mc(t, "auth", "switch-project", "nonexistent", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "PROJECT_NOT_FOUND")
}

// ---------------------------------------------------------------------------
// auth logout
// ---------------------------------------------------------------------------

func TestAuthLogout_Success(t *testing.T) {
	srv := newAuthServer(t, nil, nil)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "auth", "logout", "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)
	data, _ := v["data"].(map[string]any)
	if v, _ := data["loggedOut"].(bool); !v {
		t.Errorf("data.loggedOut should be true")
	}
}

func TestAuthLogout_NoSession(t *testing.T) {
	cp := credPath(t)
	stdout, _, code := mc(t, "auth", "logout", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "UNAUTHENTICATED")
}
