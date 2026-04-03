package org

import (
	"testing"
)

// TestIsValidOrgName tests the organization name validation function.
// Valid format: Slug with 6-24 characters
// - Must start with a lowercase letter [a-z]
// - Can contain lowercase letters, numbers, and underscores [a-z0-9_]
// - No hyphens allowed
// - Total length: 6-24 characters
func TestIsValidOrgName(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		expected bool
		reason   string
	}{
		// Valid cases (6-24 characters, no hyphens)
		{"min length", "my_org", true, "minimum length: 6 characters"},
		{"simple name", "company", true, "7 characters, simple lowercase"},
		{"with underscore", "my_company", true, "10 characters with underscore"},
		{"with number", "company123", true, "10 characters with numbers"},
		{"mixed format", "acme_corp_2024", true, "14 characters: letters, underscores, numbers"},
		{"max length", "this_is_a_long_name_24", true, "exactly 24 characters"},
		{"23 chars", "this_is_a_long_name_23", true, "23 characters"},
		{"with multiple underscores", "my_awesome_company", true, "18 characters with multiple underscores"},
		{"start letter end number", "company1", true, "8 characters: starts with letter, ends with number"},

		// Invalid cases - too short
		{"single char", "a", false, "too short: 1 character (min 6)"},
		{"two chars", "ab", false, "too short: 2 characters (min 6)"},
		{"five chars", "abcde", false, "too short: 5 characters (min 6)"},
		{"five with underscore", "ab_cd", false, "too short: 5 characters (min 6)"},

		// Invalid cases - hyphens not allowed
		{"with hyphen", "my-company", false, "hyphens not allowed"},
		{"mixed hyphen format", "acme-corp-2024", false, "hyphens not allowed"},
		{"start with hyphen", "-company", false, "starts with hyphen"},
		{"end with hyphen", "company-", false, "ends with hyphen"},
		{"only hyphens", "------", false, "only hyphens"},
		{"hyphen in middle", "my-awesome-company", false, "hyphens not allowed"},

		// Invalid cases - other format violations
		{"uppercase", "My_Company", false, "contains uppercase letters"},
		{"start with number", "123company", false, "starts with number"},
		{"space", "my company", false, "contains space"},
		{"dot", "my.company", false, "contains dot"},
		{"special char", "my_company!", false, "contains special character"},
		{"empty", "", false, "empty string"},

		// Invalid cases - too long
		{"too long", "this_is_a_very_long_organization", false, "exceeds 24 characters (32 chars)"},
		{"25 chars", "this_is_a_long_name_23456", false, "25 characters (over limit)"},

		// Edge cases
		{"only numbers", "123456", false, "starts with number"},
		{"mixed case", "MyCompany", false, "mixed case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidOrgName(tt.orgName)
			if result != tt.expected {
				t.Errorf("isValidOrgName(%q) = %v, want %v (reason: %s)",
					tt.orgName, result, tt.expected, tt.reason)
			}
		})
	}
}

// TestIsValidOrgName_RandomGenerated tests validation of randomly generated names.
func TestIsValidOrgName_RandomGenerated(t *testing.T) {
	// Examples from the random generator (using underscores instead of hyphens)
	generatedNames := []string{
		"swift_labs_42",
		"cosmic_cloud_789",
		"quantum_dev_156",
		"stellar_forge_321",
		"bright_hub_518",
		"noble_labs_198",
		"cyber_nest_387",
		"clever_cloud_648",
		"stellar_ventures_371", // Max length from generator (20 chars)
	}

	for _, name := range generatedNames {
		t.Run(name, func(t *testing.T) {
			if !isValidOrgName(name) {
				t.Errorf("isValidOrgName(%q) = false, expected true (randomly generated name should be valid)",
					name)
			}
		})
	}
}

// TestIsValidOrgName_LengthBoundary tests length boundary conditions.
func TestIsValidOrgName_LengthBoundary(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected bool
	}{
		{"1 char", 1, false},  // Too short
		{"2 chars", 2, false}, // Too short
		{"5 chars", 5, false}, // Too short
		{"6 chars", 6, true},  // Min valid length
		{"10 chars", 10, true},
		{"23 chars", 23, true},
		{"24 chars", 24, true},  // Max valid length
		{"25 chars", 25, false}, // Too long
		{"30 chars", 30, false}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate test string of specified length using valid characters
			var orgName string
			if tt.length < 6 {
				// Short strings
				orgName = "a"
				for i := 1; i < tt.length; i++ {
					orgName += "b"
				}
			} else {
				// Valid length: "a" + (length-2) * "x_" + "z"
				orgName = "a"
				for i := 0; i < tt.length-2; i++ {
					if i%2 == 0 {
						orgName += "x"
					} else {
						orgName += "_"
					}
				}
				orgName += "z"
			}

			result := isValidOrgName(orgName)
			if result != tt.expected {
				t.Errorf("isValidOrgName(%q) [len=%d] = %v, want %v",
					orgName, tt.length, result, tt.expected)
			}
		})
	}
}
