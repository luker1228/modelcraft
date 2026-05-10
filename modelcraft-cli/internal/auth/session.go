package auth

import (
	"context"
	"time"

	"modelcraft-cli/internal/config"
	"modelcraft-cli/internal/output"
)

type RefreshClient interface {
	Refresh(ctx context.Context, server, orgName, refreshToken string) (*config.Credentials, error)
}

type Manager struct {
	Now func() time.Time
}

func (m Manager) EnsureFresh(ctx context.Context, creds config.Credentials, client RefreshClient) (config.Credentials, error) {
	now := time.Now()
	if m.Now != nil {
		now = m.Now()
	}

	// Explicit bearer-token usage (for example via MC_ACCESS_TOKEN) is caller-managed.
	if creds.AccessToken != "" && creds.RefreshToken == "" {
		return creds, nil
	}

	if creds.ExpiresAt.After(now.Add(60 * time.Second)) {
		return creds, nil
	}

	if creds.RefreshToken == "" {
		return config.Credentials{}, output.NewCLIError(
			"TOKEN_EXPIRED",
			"Access token has expired.",
			true,
			"Run 'mc auth login'.",
			nil,
		)
		}

	fresh, err := client.Refresh(ctx, creds.Server, creds.OrgName, creds.RefreshToken)
	if err != nil {
		return config.Credentials{}, err
	}
	if fresh == nil {
		return config.Credentials{}, output.NewCLIError(
			"UNKNOWN_ERROR",
			"Refresh request returned no session.",
			false,
			"Run 'mc auth login'.",
			nil,
		)
	}

	fresh.CurrentProject = creds.CurrentProject
	return *fresh, nil
}

func SwitchProject(creds config.Credentials, slug string) (config.Credentials, error) {
	for _, project := range creds.Projects {
		if project.Slug == slug {
			creds.CurrentProject = slug
			return creds, nil
		}
	}

	return config.Credentials{}, output.NewCLIError(
		"PROJECT_NOT_FOUND",
		"Project is not accessible for the current user.",
		false,
		"Run 'mc catalog projects' to inspect available projects.",
		map[string]any{"project": slug},
	)
}
