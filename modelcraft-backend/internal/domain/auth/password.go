package auth

import (
	"context"
	"fmt"
	"unicode"
)

const minPasswordLength = 8

// ValidatePasswordStrength validates that a password meets minimum strength requirements.
func ValidatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	hasLetter := false
	hasDigit := false
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

// PasswordHasher defines the interface for hashing and verifying passwords.
// Infrastructure layer provides the concrete implementation (e.g., bcrypt).
type PasswordHasher interface {
	// Hash returns a hashed representation of the given password.
	Hash(ctx context.Context, password string) (string, error)

	// Verify checks if the given password matches the hash.
	// Returns nil on success, error on mismatch or failure.
	Verify(ctx context.Context, password, hash string) error
}
