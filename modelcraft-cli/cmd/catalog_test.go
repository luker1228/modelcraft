package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogProjectsDoesNotRequireCurrentProject(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	_ = os.WriteFile(credPath, []byte(`{"server":"https://gateway.example.com","orgName":"acme","projects":[{"slug":"sales","title":"Sales"}]}`), 0o600)

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
