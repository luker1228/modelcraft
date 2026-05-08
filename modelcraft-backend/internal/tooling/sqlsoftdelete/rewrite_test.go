package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteFile_AddsDeletedAtToSelect(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}
	src, err := os.ReadFile(filepath.Join("testdata", "rewrite", "list_models.sql"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	out, changed, err := RewriteFile(policy, src)
	if err != nil {
		t.Fatalf("RewriteFile() error = %v", err)
	}
	if !changed {
		t.Fatal("expected rewrite to change SQL")
	}
	got := string(out)
	if !strings.Contains(got, "`models`.`deleted_at` = 0") {
		t.Fatalf("expected rewritten SQL to contain deleted_at predicate, got: %s", got)
	}
}

func TestRewriteFile_RewritesDeleteIntoSoftDeleteUpdate(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}
	src, err := os.ReadFile(filepath.Join("testdata", "rewrite", "delete_model.sql"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	out, changed, err := RewriteFile(policy, src)
	if err != nil {
		t.Fatalf("RewriteFile() error = %v", err)
	}
	if !changed {
		t.Fatal("expected delete rewrite to change SQL")
	}
	got := string(out)
	if !(strings.Contains(got, "UPDATE `models`") || strings.Contains(got, "UPDATE models")) ||
		!strings.Contains(got, "CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)") {
		t.Fatalf("unexpected soft-delete SQL: %s", got)
	}
	if !strings.Contains(got, "`delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)") {
		t.Fatalf("expected delete_token rewrite, got: %s", got)
	}
	if !strings.Contains(got, "AND `models`.`deleted_at` = 0") {
		t.Fatalf("expected active-state guard, got: %s", got)
	}
}

func TestRewriteFile_IsIdempotent(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}
	src, err := os.ReadFile(filepath.Join("testdata", "rewrite", "delete_model.sql"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	once, _, err := RewriteFile(policy, src)
	if err != nil {
		t.Fatalf("RewriteFile() first pass error = %v", err)
	}
	twice, changed, err := RewriteFile(policy, once)
	if err != nil {
		t.Fatalf("RewriteFile() second pass error = %v", err)
	}
	if changed {
		t.Fatal("expected second rewrite to be no-op")
	}
	if string(once) != string(twice) {
		t.Fatal("expected idempotent output")
	}
}
