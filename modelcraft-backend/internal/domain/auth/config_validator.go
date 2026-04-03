package auth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
)

// ValidateCasdoorConfig validates Casdoor-specific configuration.
func ValidateCasdoorConfig(config map[string]interface{}) error {
	// Check required fields
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		return ErrInvalidConfig("endpoint is required")
	}

	// Validate endpoint URL format
	if _, err := url.Parse(endpoint); err != nil {
		return ErrInvalidConfig("endpoint must be a valid URL")
	}

	clientID, ok := config["client_id"].(string)
	if !ok || clientID == "" {
		return ErrInvalidConfig("client_id is required")
	}

	organization, ok := config["organization"].(string)
	if !ok || organization == "" {
		return ErrInvalidConfig("organization is required")
	}

	certificate, ok := config["certificate"].(string)
	if !ok || certificate == "" {
		return ErrInvalidConfig("certificate is required")
	}

	// Validate certificate PEM format
	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return ErrInvalidConfig("certificate is not valid PEM format")
	}

	// Validate it's a valid X.509 certificate
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		return ErrInvalidConfig(fmt.Sprintf("invalid X.509 certificate: %v", err))
	}

	return nil
}

// ValidateKeycloakConfig validates Keycloak-specific configuration (future implementation).
func ValidateKeycloakConfig(config map[string]interface{}) error {
	// TODO: Implement when Keycloak provider is added
	return ErrInvalidConfig("keycloak provider not yet implemented")
}

// ValidateOIDCConfig validates generic OIDC configuration (future implementation).
func ValidateOIDCConfig(config map[string]interface{}) error {
	// TODO: Implement when OIDC provider is added
	return ErrInvalidConfig("oidc provider not yet implemented")
}
