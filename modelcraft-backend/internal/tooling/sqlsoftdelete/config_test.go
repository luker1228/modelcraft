package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPolicy_DefaultEnabledBlacklistAndDeleteToken(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}

	if !policy.SoftDeleteEnabled("models") {
		t.Fatal("models should be soft-delete enabled by default")
	}
	if !policy.SoftDeleteEnabled("MODELS") {
		t.Fatal("MODELS should be checked case-insensitively")
	}
	if policy.SoftDeleteEnabled("refresh_tokens") {
		t.Fatal("refresh_tokens should stay in hard-delete blacklist")
	}
	if policy.SoftDeleteEnabled("REFRESH_TOKENS") {
		t.Fatal("REFRESH_TOKENS should be checked case-insensitively")
	}
	if !policy.NeedsDeleteToken("models") {
		t.Fatal("models should require delete_token for unique-key reuse")
	}
	if !policy.NeedsDeleteToken("MODELS") {
		t.Fatal("MODELS should be checked case-insensitively for delete_token")
	}
	if policy.NeedsDeleteToken("role_permissions") {
		t.Fatal("role_permissions should not require delete_token")
	}
}

func TestLoadPolicy_DefaultModeHandling(t *testing.T) {
	policy := &Policy{DefaultMode: "DISABLED"}
	if policy.SoftDeleteEnabled("models") {
		t.Fatal("disabled default mode should disable soft delete")
	}

	policy.DefaultMode = ""
	if !policy.SoftDeleteEnabled("models") {
		t.Fatal("empty default mode should fallback to enabled")
	}

	policy.DefaultMode = "invalid"
	if !policy.SoftDeleteEnabled("models") {
		t.Fatal("invalid default mode should fallback to enabled")
	}
}

func TestParseAnnotations(t *testing.T) {
	src := []byte("-- @include_deleted\n-- name: ListAllModels :many\nSELECT * FROM models;")
	ann := ParseAnnotations(src)
	if !ann.IncludeDeleted || ann.OnlyDeleted || ann.PhysicalDelete {
		t.Fatalf("unexpected annotation parse result: %+v", ann)
	}
}

func TestParseAnnotations_AllFlagsAndCaseInsensitive(t *testing.T) {
	src := []byte("-- @INCLUDE_DELETED\n-- @Only_Deleted\n-- @physical_delete")
	ann := ParseAnnotations(src)
	if !ann.IncludeDeleted || !ann.OnlyDeleted || !ann.PhysicalDelete {
		t.Fatalf("expected all annotations true, got: %+v", ann)
	}
}

func TestPolicyFileExists(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "..", "db", "soft_delete.yaml")); err != nil {
		t.Fatalf("policy file missing: %v", err)
	}
}
