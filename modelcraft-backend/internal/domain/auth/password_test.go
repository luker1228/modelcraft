package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePasswordStrength(t *testing.T) {
	t.Run("valid passwords", func(t *testing.T) {
		validPasswords := []string{
			"12345678",         // exactly 8 chars
			"abcdefghij",      // 10 chars
			"MyP@ssw0rd!2024", // complex
		}
		for _, pw := range validPasswords {
			err := ValidatePasswordStrength(pw)
			assert.NoError(t, err, "expected %q to be valid", pw)
		}
	})

	t.Run("too short passwords", func(t *testing.T) {
		shortPasswords := []string{
			"",       // empty
			"1234567", // 7 chars
			"abc",    // 3 chars
			"a",      // 1 char
		}
		for _, pw := range shortPasswords {
			err := ValidatePasswordStrength(pw)
			assert.Error(t, err, "expected %q to be rejected", pw)
			assert.Contains(t, err.Error(), "password must be at least 8 characters")
		}
	})
}
