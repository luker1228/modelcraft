package enduser

import "errors"

// Domain layer sentinel errors.
// These errors are defined in the Domain layer and do not depend on bizerrors.
// Application layer translates these to appropriate BusinessErrors.

var (
	// ErrInvalidCredentials indicates username not found or password mismatch.
	// Used to prevent user enumeration (same error for both cases).
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrAccountDisabled indicates the account is disabled (is_forbidden=true).
	ErrAccountDisabled = errors.New("account is disabled")

	// ErrInvalidRefreshToken indicates the refresh token is not found, expired, or revoked.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrUsernameExists indicates the username is already taken.
	ErrUsernameExists = errors.New("username already exists")
)
