package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"requestId":"r1","userId":"u1","accessToken":"a1","refreshToken":"rt1","expiresAt":"2026-05-10T12:00:00Z","projects":[{"slug":"sales","title":"Sales"}]}`))
	}))
	defer srv.Close()

	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")

	cmd := NewRootCommand(BuildInfo{})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "login", "--server", srv.URL, "--org", "acme", "--username", "alice", "--password", "secret", "--credentials", credPath})

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
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	_ = os.WriteFile(credPath, []byte(`{"server":"https://gateway.example.com","orgName":"acme","projects":[{"slug":"sales","title":"Sales"}]}`), 0o600)

	cmd := NewRootCommand(BuildInfo{})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "switch-project", "hr", "--credentials", credPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected switch-project error")
	}
}
