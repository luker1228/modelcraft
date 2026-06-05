package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Credentials struct {
	Server         string `json:"server"`
	OrgName        string `json:"orgName"`
	UserID         string `json:"userId"`
	AccessToken    string `json:"accessToken"`
	CurrentProject string `json:"currentProject,omitempty"`
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".config", "modelcraft", "credentials.json")
	}
	return filepath.Join(home, ".config", "modelcraft", "credentials.json")
}

func Save(path string, creds Credentials) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	body, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(body, '\n'), 0o600)
}

func Load(path string) (Credentials, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Credentials{}, err
	}

	var creds Credentials
	if err := json.Unmarshal(body, &creds); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}
