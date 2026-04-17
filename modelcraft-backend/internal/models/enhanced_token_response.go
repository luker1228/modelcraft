package models

// EnhancedTokenResponse represents the response from token authentication.
// It includes the JWT token and basic user identity information.
type EnhancedTokenResponse struct {
	AccessToken  string `json:"accessToken"`            // ModelCraft JWT
	TokenType    string `json:"tokenType"`              // Always "Bearer"
	ExpiresIn    int    `json:"expiresIn"`              // Token lifetime in seconds
	RefreshToken string `json:"refreshToken,omitempty"` // Future use

	// User identity information
	User UserInfo `json:"user"`
}

// UserInfo represents user identity information
type UserInfo struct {
	ID         string `json:"id"`         // ModelCraft user UUID
	ExternalID string `json:"externalId"` // External provider user ID
	Name       string `json:"name"`
	Email      string `json:"email"`
}
