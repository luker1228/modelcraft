package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"modelcraft-cli/internal/config"
	"modelcraft-cli/internal/output"
)

type fakeAuthClient struct {
	refreshed bool
	response  *config.Credentials
	err       error
}

func (f *fakeAuthClient) Refresh(context.Context, string, string, string) (*config.Credentials, error) {
	f.refreshed = true
	if f.err != nil {
		return nil, f.err
	}
	if f.response != nil {
		return f.response, nil
	}
	return &config.Credentials{AccessToken: "fresh", RefreshToken: "fresh-r", ExpiresAt: time.Now().Add(2 * time.Hour)}, nil
}

func TestEnsureFreshRefreshesWhenExpiryIsWithinOneMinute(t *testing.T) {
	mgr := Manager{Now: func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) }}
	creds := config.Credentials{Server: "https://gateway.example.com", OrgName: "acme", RefreshToken: "r1", CurrentProject: "sales", ExpiresAt: time.Date(2026, 5, 9, 12, 0, 30, 0, time.UTC)}
	client := &fakeAuthClient{}

	got, err := mgr.EnsureFresh(context.Background(), creds, client)
	if err != nil {
		t.Fatalf("EnsureFresh() error = %v", err)
	}
	if !client.refreshed || got.AccessToken != "fresh" {
		t.Fatalf("expected refresh to happen, got %+v", got)
	}
	if got.CurrentProject != "sales" {
		t.Fatalf("CurrentProject = %q, want sales", got.CurrentProject)
	}
}

func TestEnsureFreshReturnsExistingCredentialsWhenTokenIsStillFresh(t *testing.T) {
	mgr := Manager{Now: func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) }}
	creds := config.Credentials{AccessToken: "still-valid", ExpiresAt: time.Date(2026, 5, 9, 12, 2, 0, 0, time.UTC)}
	client := &fakeAuthClient{}

	got, err := mgr.EnsureFresh(context.Background(), creds, client)
	if err != nil {
		t.Fatalf("EnsureFresh() error = %v", err)
	}
	if client.refreshed {
		t.Fatal("expected no refresh for fresh token")
	}
	if got.AccessToken != "still-valid" {
		t.Fatalf("AccessToken = %q, want still-valid", got.AccessToken)
	}
}

func TestEnsureFreshTrustsCallerManagedAccessToken(t *testing.T) {
	mgr := Manager{Now: func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) }}
	creds := config.Credentials{AccessToken: "env-token", ExpiresAt: time.Time{}}
	client := &fakeAuthClient{}

	got, err := mgr.EnsureFresh(context.Background(), creds, client)
	if err != nil {
		t.Fatalf("EnsureFresh() error = %v", err)
	}
	if client.refreshed {
		t.Fatal("expected no refresh for caller-managed access token")
	}
	if got.AccessToken != "env-token" {
		t.Fatalf("AccessToken = %q, want env-token", got.AccessToken)
	}
}

func TestEnsureFreshReturnsTokenExpiredWhenRefreshTokenMissing(t *testing.T) {
	mgr := Manager{Now: func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) }}
	creds := config.Credentials{ExpiresAt: time.Date(2026, 5, 9, 11, 59, 59, 0, time.UTC)}

	_, err := mgr.EnsureFresh(context.Background(), creds, &fakeAuthClient{})
	if err == nil {
		t.Fatal("EnsureFresh() error = nil, want TOKEN_EXPIRED")
	}

	var cliErr *output.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != "TOKEN_EXPIRED" {
		t.Fatalf("CLIError.Code = %q, want TOKEN_EXPIRED", cliErr.Code)
	}
}

func TestSwitchProjectRejectsInaccessibleProject(t *testing.T) {
	creds := config.Credentials{Projects: []config.AccessibleProject{{Slug: "sales", Title: "Sales"}}}

	_, err := SwitchProject(creds, "finance")
	if err == nil {
		t.Fatal("SwitchProject() error = nil, want project error")
	}

	var cliErr *output.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != "PROJECT_NOT_FOUND" {
		t.Fatalf("CLIError.Code = %q, want PROJECT_NOT_FOUND", cliErr.Code)
	}
}
