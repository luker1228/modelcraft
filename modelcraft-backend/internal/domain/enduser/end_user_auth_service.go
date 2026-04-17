package enduser

import "context"

// EndUserAuthService defines the domain service interface for end-user authentication.
// Implementation resides in the Application layer (not in Domain layer, as it requires Repository).
type EndUserAuthService interface {
	// Authenticate verifies credentials and returns (user, error).
	// Error scenarios: invalid credentials / account disabled.
	Authenticate(ctx context.Context, cred Credential) (*EndUser, error)
}
