package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintFile_FailsWhenSelectOmitsDeletedAt(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}

	src, err := os.ReadFile(filepath.Join("testdata", "lint", "missing_deleted_at.sql"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	findings, err := LintFile(policy, "missing_deleted_at.sql", src)
	if err != nil {
		t.Fatalf("LintFile() error = %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("expected missing deleted_at finding")
	}

	rendered := RenderFindings(findings)
	if !strings.Contains(rendered, "missing deleted_at predicate for models") {
		t.Fatalf("unexpected findings: %s", rendered)
	}
}

func TestLintFile_PassesWithIncludeDeletedAnnotation(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}

	src, err := os.ReadFile(filepath.Join("testdata", "lint", "include_deleted.sql"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	findings, err := LintFile(policy, "include_deleted.sql", src)
	if err != nil {
		t.Fatalf("LintFile() error = %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %+v", findings)
	}
}

func TestLintFile_FailsWhenJoinTableMissesDeletedAt(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}

	src, err := os.ReadFile(filepath.Join("testdata", "lint", "join_missing_child_filter.sql"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	findings, err := LintFile(policy, "join_missing_child_filter.sql", src)
	if err != nil {
		t.Fatalf("LintFile() error = %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %+v", findings)
	}
	if findings[0].Message != "missing deleted_at predicate for organizations" {
		t.Fatalf("unexpected finding: %+v", findings[0])
	}
}

func TestLintFile_FailsPhysicalDeleteOnSoftDeleteTable(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}

	src := []byte("-- name: DeleteModel :exec\nDELETE FROM models WHERE id = ?;")
	findings, err := LintFile(policy, "delete_model.sql", src)
	if err != nil {
		t.Fatalf("LintFile() error = %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %+v", findings)
	}
	if findings[0].Message != "physical DELETE on soft-delete table models" {
		t.Fatalf("unexpected finding: %+v", findings[0])
	}
}

func TestRenderFindings_DeterministicOrder(t *testing.T) {
	findings := []Finding{
		{File: "b.sql", Query: "-- name: B :many", Message: "m2"},
		{File: "a.sql", Query: "-- name: A :many", Message: "m1"},
		{File: "a.sql", Query: "-- name: A :many", Message: "m0"},
	}

	rendered := RenderFindings(findings)
	expected := strings.Join([]string{
		"a.sql -- name: A :many: m0",
		"a.sql -- name: A :many: m1",
		"b.sql -- name: B :many: m2",
		"",
	}, "\n")

	if rendered != expected {
		t.Fatalf("RenderFindings() mismatch\nexpected:\n%s\ngot:\n%s", expected, rendered)
	}
}
