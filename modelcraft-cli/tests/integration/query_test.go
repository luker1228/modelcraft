package integration

import (
	"testing"
)

// ---------------------------------------------------------------------------
// query
// ---------------------------------------------------------------------------

func TestQuery_ReturnsItems(t *testing.T) {
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
		"query", "dev.maindb.User",
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
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestQuery_WithWhereFilter(t *testing.T) {
	gqlData := map[string]any{
		"findMany": []any{
			map[string]any{"id": "1", "name": "Alice"},
		},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t,
		"query", "dev.maindb.User",
		"--where", `{"name":"Alice"}`,
		"--credentials", cp,
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)
}

func TestQuery_InvalidJSON_Where(t *testing.T) {
	cp := credPath(t)
	writeValidCreds(t, cp, "http://localhost", "dev")

	stdout, _, code := mc(t,
		"query", "dev.maindb.User",
		"--where", "not-json",
		"--credentials", cp,
	)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "INVALID_JSON_FLAG")
}

func TestQuery_MetaContainsPathSegments(t *testing.T) {
	gqlData := map[string]any{"findMany": []any{}}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "query", "dev.maindb.User", "--credentials", cp)
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

func TestQuery_NoProjectContext(t *testing.T) {
	cp := credPath(t)
	// Credentials without currentProject; use db.Model path (2-segment)
	writeCredJSON(t, cp, map[string]any{
		"server":      "http://localhost",
		"orgName":     "acme",
		"accessToken": "at1",
		"expiresAt":   futureExpiry(),
	})

	stdout, _, code := mc(t, "query", "maindb.User", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "NO_PROJECT_CONTEXT")
}

// ---------------------------------------------------------------------------
// get
// ---------------------------------------------------------------------------

func TestGet_ReturnsSingleRecord(t *testing.T) {
	gqlData := map[string]any{
		"findUnique": map[string]any{"id": "42", "name": "Alice"},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t,
		"get", "dev.maindb.User",
		"--where", `{"id":"42"}`,
		"--credentials", cp,
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	if data["id"] != "42" {
		t.Errorf("data.id = %v, want 42", data["id"])
	}
}

func TestGet_MissingWhereFlag(t *testing.T) {
	cp := credPath(t)
	writeValidCreds(t, cp, "http://localhost", "dev")

	stdout, _, code := mc(t, "get", "dev.maindb.User", "--credentials", cp)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "MISSING_REQUIRED_FLAG")
}

// ---------------------------------------------------------------------------
// count
// ---------------------------------------------------------------------------

func TestCount_ReturnsNumber(t *testing.T) {
	gqlData := map[string]any{"count": float64(7)}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t, "count", "dev.maindb.User", "--credentials", cp)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)

	data, _ := v["data"].(map[string]any)
	if count, _ := data["count"].(float64); count != 7 {
		t.Errorf("data.count = %v, want 7", data["count"])
	}
}

func TestCount_WithWhereFilter(t *testing.T) {
	gqlData := map[string]any{"count": float64(3)}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t,
		"count", "dev.maindb.User",
		"--where", `{"active":true}`,
		"--credentials", cp,
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)
}

// ---------------------------------------------------------------------------
// aggregate
// ---------------------------------------------------------------------------

func TestAggregate_ReturnsStats(t *testing.T) {
	gqlData := map[string]any{
		"aggregate": map[string]any{
			"amount": map[string]any{"sum": 1000.0, "avg": 250.0, "count": 4},
		},
	}
	srv := newGraphQLServer(t, gqlData)
	defer srv.Close()

	cp := credPath(t)
	writeValidCreds(t, cp, srv.URL, "dev")

	stdout, _, code := mc(t,
		"aggregate", "dev.maindb.Order",
		"--fields", "amount",
		"--credentials", cp,
	)
	if code != 0 {
		t.Fatalf("exit code = %d, stdout: %s", code, stdout)
	}
	v := mustJSON(t, stdout)
	assertOK(t, v)
}

func TestAggregate_InvalidJSON_Where(t *testing.T) {
	cp := credPath(t)
	writeValidCreds(t, cp, "http://localhost", "dev")

	stdout, _, code := mc(t,
		"aggregate", "dev.maindb.Order",
		"--where", "{bad json}",
		"--credentials", cp,
	)
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	v := mustJSON(t, stdout)
	assertErrorCode(t, v, "INVALID_JSON_FLAG")
}
