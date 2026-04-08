package user

import (
	"fmt"
	"regexp"
)

// phoneRegex validates 11-digit mainland China phone numbers.
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// PhoneNumber is a value object representing a validated mainland China phone number.
type PhoneNumber struct {
	value string
}

// NewPhoneNumber creates a PhoneNumber after validating format.
// Returns error if the phone number is not a valid 11-digit mainland China number.
func NewPhoneNumber(value string) (PhoneNumber, error) {
	if !phoneRegex.MatchString(value) {
		return PhoneNumber{}, fmt.Errorf("invalid phone number format: must be 11-digit mainland China number")
	}
	return PhoneNumber{value: value}, nil
}

// String returns the raw phone number.
func (p PhoneNumber) String() string {
	return p.value
}

// Masked returns the phone number in masked format, e.g. "138****1234".
func (p PhoneNumber) Masked() string {
	if len(p.value) != 11 {
		return p.value
	}
	return p.value[:3] + "****" + p.value[7:]
}

// IsZero returns true if the PhoneNumber has not been set.
func (p PhoneNumber) IsZero() bool {
	return p.value == ""
}
