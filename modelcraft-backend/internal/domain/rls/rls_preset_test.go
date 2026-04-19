package rls

import "testing"

func TestRLSPresetIsDangerous(t *testing.T) {
	tests := []struct {
		name     string
		preset   RLSPreset
		expected bool
	}{
		{"READ_WRITE_ALL", RLSPresetReadWriteAll, true},
		{"READ_WRITE_OWNER", RLSPresetReadWriteOwner, false},
		{"READ_ALL_WRITE_OWNER", RLSPresetReadAllWriteOwner, false},
		{"READ_ALL", RLSPresetReadAll, false},
		{"NO_ACCESS", RLSPresetNoAccess, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.preset.IsDangerous()
			if result != tt.expected {
				t.Errorf("IsDangerous() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRLSPresetString(t *testing.T) {
	preset := RLSPresetReadWriteOwner
	if preset.String() != "READ_WRITE_OWNER" {
		t.Errorf("String() = %v, want READ_WRITE_OWNER", preset.String())
	}
}
