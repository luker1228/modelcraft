package enduser

import "fmt"

// Credential represents input credentials for login/register (not persisted).
type Credential struct {
	Username      string
	PlainPassword string
}

// NewCredential creates and validates a credential.
func NewCredential(username, plainPassword string) (Credential, error) {
	if username == "" {
		return Credential{}, fmt.Errorf("username is required")
	}
	if plainPassword == "" {
		return Credential{}, fmt.Errorf("password is required")
	}
	return Credential{
		Username:      username,
		PlainPassword: plainPassword,
	}, nil
}
