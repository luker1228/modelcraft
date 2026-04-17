package enduser

import (
	"fmt"
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptCost is the bcrypt cost factor (12 is production recommended).
	bcryptCost = 12

	// minPasswordLength is the minimum password length.
	minPasswordLength = 8
)

// HashedPassword is a value object that encapsulates a bcrypt-hashed password.
type HashedPassword struct {
	Hash      string // bcrypt hash string
	Algorithm string // always "bcrypt"
}

// NewHashedPasswordFromPlain creates a HashedPassword from plaintext (calls bcrypt with cost=12).
func NewHashedPasswordFromPlain(plain string) (HashedPassword, error) {
	if err := ValidatePasswordStrength(plain); err != nil {
		return HashedPassword{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return HashedPassword{}, fmt.Errorf("failed to hash password: %w", err)
	}

	return HashedPassword{
		Hash:      string(hash),
		Algorithm: "bcrypt",
	}, nil
}

// NewHashedPasswordFromHash reconstructs a HashedPassword from an existing hash (no re-hashing).
func NewHashedPasswordFromHash(hash string) HashedPassword {
	return HashedPassword{
		Hash:      hash,
		Algorithm: "bcrypt",
	}
}

// Verify checks if the plaintext password matches the hash.
func (p HashedPassword) Verify(plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.Hash), []byte(plain))
	return err == nil
}

// ValidatePasswordStrength validates that a password meets minimum strength requirements:
// - At least 8 characters
// - Contains at least one letter
// - Contains at least one digit
func ValidatePasswordStrength(plain string) error {
	if len(plain) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	hasLetter := false
	hasDigit := false
	for _, r := range plain {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
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

// usernameRegex matches valid usernames: 3-64 characters, alphanumeric with underscore and hyphen.
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)

// ValidateUsername validates that a username meets format requirements.
func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username must be 3-64 characters and contain only letters, numbers, underscores, or hyphens")
	}
	return nil
}
