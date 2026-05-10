package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestQueryCommandReturnsNoProjectContextWhenDatabaseModelLacksFallback(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	_ = os.WriteFile(credPath, []byte(`{"server":"https://gateway.example.com","orgName":"acme","accessToken":"tok"}`), 0o600)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	code := Execute(BuildInfo{}, []string{"query", "maindb.users", "--credentials", credPath}, stdout, stderr)

	if code != 5 {
		t.Fatalf("Execute() code = %d, want 5, stdout=%s", code, stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"code":"NO_PROJECT_CONTEXT"`)) {
		t.Fatalf("expected NO_PROJECT_CONTEXT, got %s", stdout.String())
	}
}
