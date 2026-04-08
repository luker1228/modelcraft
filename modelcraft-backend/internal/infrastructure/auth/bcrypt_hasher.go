package auth

import (
	"context"

	domainAuth "modelcraft/internal/domain/auth"

	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implements domain auth.PasswordHasher using bcrypt.
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher creates a new BcryptPasswordHasher with default cost.
func NewBcryptPasswordHasher() domainAuth.PasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(_ context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *BcryptPasswordHasher) Verify(_ context.Context, password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Compile-time interface satisfaction check.
var _ domainAuth.PasswordHasher = (*BcryptPasswordHasher)(nil)
