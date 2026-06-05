package integration

import (
	"testing"
)

// ---------------------------------------------------------------------------
// catalog projects
// ---------------------------------------------------------------------------

func TestCatalogProjects_ListsFromCredentials(t *testing.T) {
	// catalog projects calls backend myProjects query.
	gqlData := map[string]any{
		"myProjects": []map[string]any{
			{"slug": "sales", "title": "Sales"},
			{"slug": "hr", "title": "HR"},
		},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "sales")

	stdout, _, code := mc(t, "catalog", "projects", "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 projects, got %d", len(items))
	}
}

func TestCatalogProjects_Unauthenticated(t *testing.T) {
	cp := credPath(t)
	stdout, _, code := mc(t, "catalog", "projects", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "UNAUTHENTICATED")
}

// ---------------------------------------------------------------------------
// catalog databases
// ---------------------------------------------------------------------------

func TestCatalogDatabases_Success(t *testing.T) {
	gqlData := map[string]any{
		"modelDatabaseCatalog": map[string]any{
			"data": map[string]any{
				"databases": []any{
					map[string]any{"name": "maindb"},
					map[string]any{"name": "analyticsdb"},
				},
			},
		},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "catalog", "databases", "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 databases, got %d", len(items))
	}
}

func TestCatalogDatabases_NoProjectContext(t *testing.T) {
	cp := credPath(t)
	writeCredJSON(t, cp, map[string]any{
		"server":      "http://localhost",
		"orgName":     "acme",
		"accessToken": "at1",
		"expiresAt":   futureExpiry(),
		// No currentProject, no projects list
	})

	stdout, _, code := mc(t, "catalog", "databases", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "NO_PROJECT_CONTEXT")
}

func TestCatalogDatabases_ProjectOverrideFlag(t *testing.T) {
	gqlData := map[string]any{
		"modelDatabaseCatalog": map[string]any{
			"data": map[string]any{
				"databases": []any{map[string]any{"name": "maindb"}},
			},
		},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	// Credentials have no currentProject, but --project flag supplies it.
	writeCredJSON(t, cp, map[string]any{
		"server":      srv.URL,
		"orgName":     "acme",
		"accessToken": "at1",
		"expiresAt":   futureExpiry(),
		"projects":    []map[string]any{{"slug": "prod", "title": "Prod"}},
	})

	stdout, _, code := mc(t, "catalog", "databases", "--project", "prod", "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	// meta should include the project
	meta, _ := v["meta"].(map[string]any)
	if meta["project"] != "prod" {
		t.Errorf("meta.project = %v, want prod", meta["project"])
	}
}

// ---------------------------------------------------------------------------
// catalog models
// ---------------------------------------------------------------------------

func TestCatalogModels_Success(t *testing.T) {
	gqlData := map[string]any{
		"models": map[string]any{
			"items": []any{
				map[string]any{"name": "User"},
				map[string]any{"name": "Order"},
			},
		},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "catalog", "models",
		"--database", "maindb",
		"--credentials", cp,
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	items, _ := data["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 models, got %d", len(items))
	}
}

func TestCatalogModels_MissingDatabaseFlag(t *testing.T) {
	cp := credPath(t)
	writeValidCreds(t, cp, "http://localhost", "dev")

	stdout, _, code := mc(t, "catalog", "models", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "MISSING_REQUIRED_FLAG")
}
