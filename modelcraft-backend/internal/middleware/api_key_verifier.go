package middleware

import "context"

// APIKeyVerifier API Key verification interface.
// Implemented by APIKeyService in Phase A6.
type APIKeyVerifier interface {
	VerifyAPIKey(ctx context.Context, rawKey string) (userID string, err error)
}
