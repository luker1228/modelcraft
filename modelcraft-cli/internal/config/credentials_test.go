package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestSaveAndLoadCredentials(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")

	creds := Credentials{
		Server:       "https://gateway.example.com",
		OrgName:      "acme",
		UserID:       "user-1",
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC),
		Projects:     []AccessibleProject{{Slug: "sales", Title: "Sales"}},
	}

	if err := Save(path, creds); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.OrgName != "acme" || got.CurrentProject != "" {
		t.Fatalf("unexpected credentials: %+v", got)
	}
}

func TestSaveCreatesPrivateFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-style file mode bits are not stable on windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "credentials.json")

	if err := Save(path, Credentials{OrgName: "acme"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("credentials mode = %o, want 600", mode)
	}
}
