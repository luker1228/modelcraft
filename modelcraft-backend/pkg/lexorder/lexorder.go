// Package lexorder provides lexicographic fractional indexing for ordered collections.
// It enables efficient single-row updates for drag-and-drop reordering by computing
// string keys that sort correctly with standard ORDER BY.
//
// Keys use printable ASCII characters '!' (33) to '~' (126) as the digit alphabet.
// A midpoint between any two distinct strings can always be computed without precision loss.
package lexorder

import (
	"fmt"
	"strings"
)

const (
	// base is the size of the digit alphabet.
	base = 94 // '~' - '!' + 1
	// offset maps a digit value (0..base-1) to an ASCII character.
	offset = '!'
)

// charToDigit converts an ASCII character to its digit value (0-based).
func charToDigit(c byte) int {
	return int(c) - int(offset)
}

// digitToChar converts a digit value (0-based) to its ASCII character.
func digitToChar(d int) byte {
	return byte(d + int(offset))
}

// InitialOrder returns the default display_order for the first item in an empty project.
func InitialOrder() string {
	return string(digitToChar(base / 2))
}

// Midpoint computes a lexicographic string strictly between prev and next.
//   - If prev is empty, returns a value strictly less than next.
//   - If next is empty, returns a value strictly greater than prev.
//   - If both are empty, returns InitialOrder().
//
// Error is returned only if prev >= next (invalid input).
func Midpoint(prev, next string) (string, error) {
	switch {
	case prev == "" && next == "":
		return InitialOrder(), nil
	case next == "":
		// Append a mid-alphabet char: prev + mid gives something > prev.
		return prev + string(digitToChar(base/2)), nil
	case prev == "":
		// We need something < next.
		// Try decrementing the last character.
		return decrementOne(next)
	}

	if prev >= next {
		return "", fmt.Errorf("lexorder: prev %q must be less than next %q", prev, next)
	}

	return computeMidpoint(prev, next), nil
}

// decrementOne returns a string lexicographically less than s.
// It tries to decrement the last character; if that character is already at minimum,
// it shortens the string (removes the trailing minimum char).
func decrementOne(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("lexorder: cannot decrement empty string")
	}
	// Find last character that is not at minimum.
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] > offset {
			// Decrement this character and truncate the rest.
			return s[:i] + string(s[i]-1) + string(digitToChar(base/2)), nil
		}
	}
	// All characters are at minimum — prepend a min char halved.
	// This shouldn't happen in normal usage, but handle gracefully.
	return string(digitToChar(base / 4)), nil
}

// computeMidpoint computes the arithmetic average of two strings treated as base-94 fractions.
// Both strings must satisfy a < b lexicographically.
func computeMidpoint(a, b string) string {
	// Normalize lengths by padding with the zero digit (offset char).
	length := len(a)
	if len(b) > length {
		length = len(b)
	}

	aDigits := toDigits(a, length+1)
	bDigits := toDigits(b, length+1)

	// Compute (a + b) / 2 in base-94.
	sum := make([]int, length+2)
	for i := length; i >= 0; i-- {
		s := aDigits[i] + bDigits[i] + sum[i+1]
		sum[i] = s / base
		sum[i+1] = s % base
	}
	// Wait — this adds, we need average. Use standard long division by 2 instead.
	return average(aDigits, bDigits)
}

// average computes floor((a + b) / 2) for two digit slices of equal length.
func average(a, b []int) string {
	n := len(a)
	result := make([]int, n)
	carry := 0

	// Add a and b.
	sum := make([]int, n+1)
	for i := n - 1; i >= 0; i-- {
		s := a[i] + b[i] + carry
		carry = s / base
		sum[i+1] = s % base
	}
	sum[0] = carry

	// Divide by 2.
	rem := 0
	for i := 0; i <= n; i++ {
		cur := sum[i] + rem*base
		result2 := cur / 2
		rem = cur % 2
		if i > 0 {
			result[i-1] = result2
		}
	}

	// The remainder propagates: if rem != 0 at the end, we need an extra digit.
	// We ignore the fractional part — floor division is fine.

	// Convert back to string, trimming trailing zero-digits (offset chars would be '!').
	// But DON'T trim if they are needed to be > a.
	raw := toStr(result)

	// Ensure result > a (it should be by construction, but verify).
	// If result == a, append a mid char to push it above.
	if raw == toStr(a) {
		raw += string(digitToChar(base / 2))
	}

	return raw
}

// toDigits converts a string to a digit slice of the given length, zero-padding on the right.
func toDigits(s string, length int) []int {
	d := make([]int, length)
	for i := 0; i < len(s) && i < length; i++ {
		d[i] = charToDigit(s[i])
	}
	return d
}

// toStr converts a digit slice back to a string, trimming trailing zero-digits.
func toStr(digits []int) string {
	end := len(digits)
	for end > 1 && digits[end-1] == 0 {
		end--
	}
	var sb strings.Builder
	for i := 0; i < end; i++ {
		sb.WriteByte(digitToChar(digits[i]))
	}
	return sb.String()
}

// Renumber generates n evenly-spaced display_order values for a full renumber.
// Use when lexicographic collision is detected (extremely rare).
func Renumber(n int) ([]string, error) {
	if n == 0 {
		return nil, nil
	}

	result := make([]string, n)
	for i := 0; i < n; i++ {
		// Distribute evenly across a two-digit keyspace for robustness.
		// Total keyspace is base * base positions.
		total := base * base
		pos := (i+1)*total/(n+1) + 1
		if pos >= total {
			pos = total - 1
		}
		hi := pos / base
		lo := pos % base
		if hi == 0 {
			result[i] = string(digitToChar(lo))
		} else {
			result[i] = string([]byte{digitToChar(hi), digitToChar(lo)})
		}
	}

	// Verify strictly ascending.
	for i := 1; i < len(result); i++ {
		if result[i] <= result[i-1] {
			return nil, fmt.Errorf("lexorder: renumber produced non-ascending sequence at index %d: %q <= %q",
				i, result[i], result[i-1])
		}
	}

	return result, nil
}
