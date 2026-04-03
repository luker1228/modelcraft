package auth

import (
	"fmt"
	"modelcraft/pkg/bizerrors"
)

// ErrInvalidConfig creates a PARAM_INVALID error for invalid authentication configuration.
func ErrInvalidConfig(message string) error {
	return bizerrors.NewError(bizerrors.ParamInvalid, fmt.Sprintf("invalid auth config: %s", message))
}

// ErrProviderNotConfigured creates a SYSTEM_ERROR for missing provider configuration.
func ErrProviderNotConfigured(projectID string) error {
	return bizerrors.NewError(
		bizerrors.SystemError,
		fmt.Sprintf("authentication provider not configured for project: %s", projectID),
	)
}

// ErrProviderInitFailed creates a SYSTEM_ERROR for provider initialization failure.
func ErrProviderInitFailed(provider string, err error) error {
	return bizerrors.WrapError(err, bizerrors.SystemError, fmt.Sprintf("failed to initialize %s provider", provider))
}
