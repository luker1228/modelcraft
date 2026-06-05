package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCatalogProjectsDoesNotRequireCurrentProject(t *testing.T) {
	// catalog projects now calls the backend — spin up a mock GraphQL server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"myProjects": []map[string]any{
					{"slug": "sales", "title": "Sales"},
				},
			},
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	creds := map[string]any{
		"server":      srv.URL,
		"orgName":     "acme",
		"accessToken": "at-valid",
		"expiresAt":   time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		// No currentProject required
	}
	b, _ := json.Marshal(creds)
	_ = os.WriteFile(credPath, b, 0o600)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	code := Execute(BuildInfo{}, []string{"catalog", "projects", "--credentials", credPath}, stdout, stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, want 0, stdout=%s", code, stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"slug":"sales"`)) {
		t.Fatalf("expected projects output, got %s", stdout.String())
	}
}
