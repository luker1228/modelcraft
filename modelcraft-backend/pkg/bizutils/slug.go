package bizutils

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// GenerateSlug converts a display name to a URL-friendly slug.
// Rules:
// - Convert to lowercase
// - Convert Chinese characters to Pinyin initials
// - Remove all special characters except letters and numbers
// - Remove hyphens, underscores, and spaces
// - Return empty string if input is empty or contains only special chars
//
// Examples:
//   - "My Company" -> "mycompany"
//   - "我的公司" -> "wdgs"
//   - "Swift Labs 42" -> "swiftlabs42"
func GenerateSlug(displayName string) string {
	if displayName == "" {
		return ""
	}

	var result strings.Builder

	// Process character by character
	for _, r := range displayName {
		// Keep ASCII letters and numbers as-is
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(unicode.ToLower(r))
			continue
		}

		// Convert Chinese characters to Pinyin initials
		if unicode.Is(unicode.Han, r) {
			pinyinList := pinyin.Pinyin(string(r), pinyin.NewArgs())
			if len(pinyinList) > 0 && len(pinyinList[0]) > 0 {
				// Take first letter of pinyin
				initial := pinyinList[0][0]
				if len(initial) > 0 {
					result.WriteByte(byte(unicode.ToLower(rune(initial[0]))))
				}
			}
			continue
		}

		// Skip all other characters (spaces, hyphens, underscores, special chars)
	}

	slug := result.String()

	// Remove any remaining non-alphanumeric characters
	slug = regexp.MustCompile(`[^a-z0-9]`).ReplaceAllString(slug, "")

	return slug
}

// GenerateSlugWithLength generates a slug with length constraints.
// If the slug is too short, it adds a random alphanumeric suffix.
// If the slug is too long, it truncates to maxLen.
//
// Parameters:
//   - displayName: input text to convert to slug
//   - minLen: minimum length required (typically 6 for organization names)
//   - maxLen: maximum length allowed (typically 24 for organization names)
//
// Returns:
//   - A slug that meets the length constraints
func GenerateSlugWithLength(displayName string, minLen, maxLen int) string {
	slug := GenerateSlug(displayName)

	// Truncate if too long
	if len(slug) > maxLen {
		slug = slug[:maxLen]
	}

	// Pad with random suffix if too short
	if len(slug) < minLen {
		needed := minLen - len(slug)
		suffix := generateRandomAlphanumeric(needed)
		slug += suffix
	}

	return slug
}

// generateRandomAlphanumeric generates a random lowercase alphanumeric string
// of the specified length.
func generateRandomAlphanumeric(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// crypto/rand failure is extremely rare, fallback to timestamp-based char
			result[i] = charset[(i*7)%len(charset)]
			continue
		}
		result[i] = charset[n.Int64()]
	}

	return string(result)
}
