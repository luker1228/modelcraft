package integration

import (
	"testing"
)

// ---------------------------------------------------------------------------
// run
// ---------------------------------------------------------------------------

func TestRun_ReturnsData(t *testing.T) {
	gqlData := map[string]any{
		"findMany": []any{
			map[string]any{"id": "1", "name": "Alice"},
			map[string]any{"id": "2", "name": "Bob"},
		},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t,
		"run", "dev.maindb.User", `{ findMany { id name } }`,
		"--credentials", cp,
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	items, _ := data["findMany"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestRun_MetaContainsPathSegments(t *testing.T) {
	gqlData := map[string]any{"findMany": []any{}}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "run", "dev.maindb.User", `{ findMany { id } }`, "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	meta, _ := v["meta"].(map[string]any)
	if meta["project"] != "dev" {
		t.Errorf("meta.project = %v, want dev", meta["project"])
	}
	if meta["database"] != "maindb" {
		t.Errorf("meta.database = %v, want maindb", meta["database"])
	}
	if meta["model"] != "User" {
		t.Errorf("meta.model = %v, want User", meta["model"])
	}
}

func TestRun_NoProjectContext(t *testing.T) {
	cp := credPath(t)
	writeCredJSON(t, cp, map[string]any{
		"server":      "http://localhost",
		"orgName":     "acme",
		"accessToken": "at1",
		"expiresAt":   futureExpiry(),
	})

	stdout, _, code := mc(t, "run", "maindb.User", `{ count }`, "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "NO_PROJECT_CONTEXT")
}

func TestRun_MissingQuery(t *testing.T) {
	cp := credPath(t)
	writeValidCreds(t, cp, "http://localhost", "dev")

	// Only path, no query argument — stdin is empty in test context
	stdout, _, code := mc(t, "run", "dev.maindb.User", "", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code for empty query")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "MISSING_REQUIRED_FLAG")
}
