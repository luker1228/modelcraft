package auth

import (
	"time"
)

// ProviderType represents the authentication provider type.
type ProviderType string

const (
	// ProviderCasdoor represents the Casdoor authentication provider
	ProviderCasdoor ProviderType = "casdoor"

	// ProviderKeycloak represents the Keycloak authentication provider (future implementation)
	ProviderKeycloak ProviderType = "keycloak"

	// ProviderOIDC represents a generic OIDC authentication provider (future implementation)
	ProviderOIDC ProviderType = "oidc"
)

// IsValid checks if the provider type is one of the supported values.
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderCasdoor, ProviderKeycloak, ProviderOIDC:
		return true
	default:
		return false
	}
}

// String returns the string representation of the provider type.
func (p ProviderType) String() string {
	return string(p)
}

// ProjectAuthConfig represents the authentication configuration for a project.
type ProjectAuthConfig struct {
	// ID is the primary key
	ID int64

	// OrgName is the organization name (from projects table composite key)
	OrgName string

	// ProjectSlug is the project slug (from projects table composite key)
	ProjectSlug string

	// Provider is the authentication provider type
	Provider ProviderType

	// Enabled indicates whether authentication is enabled for this project
	Enabled bool

	// Config contains provider-specific configuration as a map
	// For Casdoor: endpoint, client_id, client_secret, organization, application, certificate
	// For Keycloak: realm, auth_server_url, jwks_uri, client_id
	// For OIDC: issuer, jwks_uri, client_id
	Config map[string]interface{}

	// CreatedAt is the record creation timestamp
	CreatedAt time.Time

	// UpdatedAt is the record last update timestamp
	UpdatedAt time.Time
}

// Validate validates the project auth configuration.
func (c *ProjectAuthConfig) Validate() error {
	if c.OrgName == "" {
		return ErrInvalidConfig("org_name is required")
	}

	if c.ProjectSlug == "" {
		return ErrInvalidConfig("project_slug is required")
	}

	if !c.Provider.IsValid() {
		return ErrInvalidConfig("invalid provider type: " + string(c.Provider))
	}

	if c.Config == nil {
		return ErrInvalidConfig("config is required")
	}

	// Provider-specific validation
	switch c.Provider {
	case ProviderCasdoor:
		return ValidateCasdoorConfig(c.Config)
	case ProviderKeycloak:
		return ValidateKeycloakConfig(c.Config)
	case ProviderOIDC:
		return ValidateOIDCConfig(c.Config)
	default:
		return ErrInvalidConfig("unsupported provider: " + string(c.Provider))
	}
}
