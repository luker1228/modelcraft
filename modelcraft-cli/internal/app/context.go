package app

import (
	"os"

	"modelcraft-cli/internal/config"
)

func ResolveContext(creds config.Credentials, project string) (config.Credentials, error) {
	resolved := creds

	if server := os.Getenv("MC_SERVER"); server != "" {
		resolved.Server = server
	}
	if org := os.Getenv("MC_ORG"); org != "" {
		resolved.OrgName = org
	}
	if accessToken := os.Getenv("MC_ACCESS_TOKEN"); accessToken != "" {
		resolved.AccessToken = accessToken
	}

	targetProject := project
	if envProject := os.Getenv("MC_PROJECT"); envProject != "" {
		targetProject = envProject
	}
	if targetProject != "" {
		resolved.CurrentProject = targetProject
	}

	return resolved, nil
}
