package auth

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
