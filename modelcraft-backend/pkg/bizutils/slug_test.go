package bizutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple english text",
			input:    "My Company",
			expected: "mycompany",
		},
		{
			name:     "text with spaces",
			input:    "Swift Labs 42",
			expected: "swiftlabs42",
		},
		{
			name:     "text with special characters",
			input:    "Acme Corp!@#",
			expected: "acmecorp",
		},
		{
			name:     "chinese characters",
			input:    "我的公司",
			expected: "wdgs",
		},
		{
			name:     "mixed chinese and english",
			input:    "我的Company",
			expected: "wdcompany",
		},
		{
			name:     "text with numbers",
			input:    "Tech 2024",
			expected: "tech2024",
		},
		{
			name:     "text with hyphens and underscores",
			input:    "My-Company_Name",
			expected: "mycompanyname",
		},
		{
			name:     "japanese characters",
			input:    "テスト",
			expected: "", // Japanese romanization not supported; non-CJK chars produce empty slug
		},
		{
			name:     "already lowercase",
			input:    "mycompany",
			expected: "mycompany",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "!@#$%",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSlugWithLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minLen   int
		maxLen   int
		expected string
		hasError bool
	}{
		{
			name:     "within length bounds",
			input:    "MyCompany",
			minLen:   6,
			maxLen:   24,
			expected: "mycompany",
			hasError: false,
		},
		{
			name:     "too short - pad with random suffix",
			input:    "abc",
			minLen:   6,
			maxLen:   24,
			expected: "", // will be checked for length
			hasError: false,
		},
		{
			name:     "too long - truncate",
			input:    "verylongcompanynamethatexceedsmaximumlength",
			minLen:   6,
			maxLen:   24,
			expected: "verylongcompanynamethate",
			hasError: false,
		},
		{
			name:     "exact minimum length",
			input:    "abcdef",
			minLen:   6,
			maxLen:   24,
			expected: "abcdef",
			hasError: false,
		},
		{
			name:     "exact maximum length",
			input:    "abcdefghijklmnopqrstuvwx",
			minLen:   6,
			maxLen:   24,
			expected: "abcdefghijklmnopqrstuvwx",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlugWithLength(tt.input, tt.minLen, tt.maxLen)

			// Verify length constraints
			assert.GreaterOrEqual(t, len(result), tt.minLen, "slug should meet minimum length")
			assert.LessOrEqual(t, len(result), tt.maxLen, "slug should not exceed maximum length")

			// Verify specific expected values (except for random padding cases)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
