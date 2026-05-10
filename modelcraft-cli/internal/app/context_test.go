package app

import (
	"testing"
	"time"

	"modelcraft-cli/internal/config"
)

func TestResolveContextAppliesEnvironmentOverrides(t *testing.T) {
	t.Setenv("MC_SERVER", "https://override.example.com")
	t.Setenv("MC_ORG", "override-org")
	t.Setenv("MC_ACCESS_TOKEN", "override-token")
	t.Setenv("MC_PROJECT", "sales")

	creds := config.Credentials{
		Server:       "https://gateway.example.com",
		OrgName:      "acme",
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC),
		Projects:     []config.AccessibleProject{{Slug: "sales", Title: "Sales"}},
	}

	got, err := ResolveContext(creds, "")
	if err != nil {
		t.Fatalf("ResolveContext() error = %v", err)
	}
	if got.Server != "https://override.example.com" {
		t.Fatalf("Server = %q, want override", got.Server)
	}
	if got.OrgName != "override-org" {
		t.Fatalf("OrgName = %q, want override", got.OrgName)
	}
	if got.AccessToken != "override-token" {
		t.Fatalf("AccessToken = %q, want override", got.AccessToken)
	}
	if got.RefreshToken != "" {
		t.Fatalf("RefreshToken = %q, want empty for caller-managed token", got.RefreshToken)
	}
	if got.CurrentProject != "sales" {
		t.Fatalf("CurrentProject = %q, want sales", got.CurrentProject)
	}
}

func TestResolveContextAllowsExplicitProjectWithoutStoredProjectList(t *testing.T) {
	creds := config.Credentials{}

	got, err := ResolveContext(creds, "finance")
	if err != nil {
		t.Fatalf("ResolveContext() error = %v", err)
	}
	if got.CurrentProject != "finance" {
		t.Fatalf("CurrentProject = %q, want finance", got.CurrentProject)
	}
}
