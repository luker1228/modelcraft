package enduser_test

import (
	"modelcraft/internal/domain/enduser"
	"testing"
)

func TestNewBuiltinEndUser_SetsIsBuiltin(t *testing.T) {
	hashedPwd, _ := enduser.NewHashedPasswordFromPlain("Password1")
	u, err := enduser.NewBuiltinEndUser("id-1", "myorg", "createdBy-1", hashedPwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !u.IsBuiltin {
		t.Error("expected IsBuiltin to be true")
	}
	if u.Username != "admin" {
		t.Errorf("expected Username=admin, got %s", u.Username)
	}
	if u.IsForbidden {
		t.Error("expected IsForbidden to be false")
	}
}

func TestNewEndUser_IsBuiltinFalseByDefault(t *testing.T) {
	hashedPwd, _ := enduser.NewHashedPasswordFromPlain("Password1")
	u, err := enduser.NewEndUser("id-2", "myorg", "someuser", "creator", hashedPwd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.IsBuiltin {
		t.Error("expected IsBuiltin to be false for normal users")
	}
}
