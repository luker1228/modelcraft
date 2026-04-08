package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPhoneNumber(t *testing.T) {
	t.Run("valid mainland China numbers", func(t *testing.T) {
		validNumbers := []string{
			"13800138000",
			"15912345678",
			"18611112222",
			"17700001111",
			"19988887777",
		}
		for _, num := range validNumbers {
			phone, err := NewPhoneNumber(num)
			require.NoError(t, err, "expected %s to be valid", num)
			assert.Equal(t, num, phone.String())
		}
	})

	t.Run("invalid phone numbers", func(t *testing.T) {
		invalidNumbers := []struct {
			input string
			desc  string
		}{
			{"", "empty string"},
			{"1380013800", "10 digits"},
			{"138001380001", "12 digits"},
			{"12345678901", "starts with 12"},
			{"10800138000", "starts with 10"},
			{"23800138000", "starts with 2"},
			{"abcdefghijk", "letters"},
			{"1380013800a", "contains letter"},
		}
		for _, tc := range invalidNumbers {
			_, err := NewPhoneNumber(tc.input)
			assert.Error(t, err, "expected %q (%s) to be invalid", tc.input, tc.desc)
			assert.Contains(t, err.Error(), "invalid phone number format: must be 11-digit mainland China number")
		}
	})
}

func TestPhoneNumber_Masked(t *testing.T) {
	t.Run("masks middle digits", func(t *testing.T) {
		phone, err := NewPhoneNumber("13800138000")
		require.NoError(t, err)
		assert.Equal(t, "138****8000", phone.Masked())
	})

	t.Run("masks different number", func(t *testing.T) {
		phone, err := NewPhoneNumber("15912345678")
		require.NoError(t, err)
		assert.Equal(t, "159****5678", phone.Masked())
	})
}

func TestPhoneNumber_IsZero(t *testing.T) {
	t.Run("zero value is zero", func(t *testing.T) {
		var phone PhoneNumber
		assert.True(t, phone.IsZero())
	})

	t.Run("valid phone is not zero", func(t *testing.T) {
		phone, err := NewPhoneNumber("13800138000")
		require.NoError(t, err)
		assert.False(t, phone.IsZero())
	})
}
