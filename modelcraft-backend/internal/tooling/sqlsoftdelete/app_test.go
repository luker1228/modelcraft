package sqlsoftdelete

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRun_LintFailureMissingDeletedAt(t *testing.T) {
	tmpDir := t.TempDir()

	queryFile := filepath.Join(tmpDir, "queries", "models.sql")
	if err := os.MkdirAll(filepath.Dir(queryFile), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	sql := "-- name: ListModels :many\nSELECT * FROM models WHERE org_name = ?;\n"
	if err := os.WriteFile(queryFile, []byte(sql), 0o600); err != nil {
		t.Fatalf("WriteFile(query) error = %v", err)
	}

	policyPath := filepath.Join(tmpDir, "soft_delete.yaml")
	lintPaths := "  - \"" + filepath.ToSlash(filepath.Join(tmpDir, "queries", "*.sql")) + "\""
	policy := "default_mode: enabled\nlint_paths:\n" + lintPaths +
		"\nblacklist_tables: []\ndelete_token_tables: []\n"
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"lint", "--config", policyPath}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1, stderr=%q", code, stderr.String())
	}
	if got := stderr.String(); got == "" {
		t.Fatal("stderr should contain lint findings")
	}
	if !bytes.Contains(stderr.Bytes(), []byte("missing deleted_at predicate")) {
		t.Fatalf("stderr should contain missing deleted_at predicate, got=%q", stderr.String())
	}
}
