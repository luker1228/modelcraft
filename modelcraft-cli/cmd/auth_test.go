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

func TestAuthLoginUsesDevcloudDefaultServer(t *testing.T) {
	cmd := newAuthLoginCommand()

	flag := cmd.Flags().Lookup("server")
	if flag == nil {
		t.Fatal("server flag must exist")
	}
	if got, want := flag.DefValue, "http://lukemxjia.devcloud.woa.com:9080"; got != want {
		t.Fatalf("server default = %q, want %q", got, want)
	}
}

func TestAuthLoginPersistsSingleProfileWithoutSelectingProject(t *testing.T) {
	// Login via PAT (whoami endpoint). Server returns one project — must NOT auto-select.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"userId":"u1","orgName":"acme","projects":[{"slug":"sales","title":"Sales"}]}`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")

	cmd := NewRootCommand(BuildInfo{})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "login", "--server", srv.URL, "--token", "mc_pat_test", "--credentials", credPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	b, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if bytes.Contains(b, []byte(`"currentProject":"sales"`)) {
		t.Fatalf("login must not set currentProject automatically: %s", b)
	}
}

func TestAuthSwitchProjectRejectsUnknownProject(t *testing.T) {
	// switch-project validates via backend myProjects query.
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
	}
	b, _ := json.Marshal(creds)
	_ = os.WriteFile(credPath, b, 0o600)

	cmd := NewRootCommand(BuildInfo{})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "switch-project", "hr", "--credentials", credPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected switch-project error for unknown project")
	}
}

