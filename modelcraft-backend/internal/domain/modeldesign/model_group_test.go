package modeldesign_test

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGroupName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid names
		{name: "simple lowercase", input: "payment", wantErr: false},
		{name: "with underscore", input: "user_profile", wantErr: false},
		{name: "with numbers", input: "order_v2", wantErr: false},
		{name: "single letter", input: "a", wantErr: false},
		// exactly 64 chars (valid)
		{
			name:    "max length 64",
			input:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: false,
		},

		// Invalid names
		{name: "starts with underscore", input: "_payment", wantErr: true},
		{name: "starts with number", input: "2fa", wantErr: true},
		{name: "uppercase letter", input: "Payment", wantErr: true},
		{name: "hyphen", input: "order-v2", wantErr: true},
		{name: "space", input: "order v2", wantErr: true},
		{name: "empty string", input: "", wantErr: true},
		// 65 chars (invalid)
		{
			name:    "exceeds 64 chars",
			input:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: true,
		},
		{name: "special chars", input: "order@v2", wantErr: true},
		{name: "ungrouped reserved", input: "ungrouped", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modeldesign.ValidateGroupName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewModelGroup(t *testing.T) {
	t.Run("creates valid group", func(t *testing.T) {
		group := &modeldesign.ModelGroup{
			ID:           "group-id-1",
			ProjectScope: project.ProjectScope{OrgName: "my-org", ProjectSlug: "my-project"},
			Name:         "payment",
			DisplayOrder: "a0",
		}
		assert.Equal(t, "payment", group.Name)
		assert.Equal(t, "a0", group.DisplayOrder)
		assert.False(t, group.IsVirtual())
	})
}

func TestUngroupedSentinel(t *testing.T) {
	t.Run("ungrouped group is virtual", func(t *testing.T) {
		g := modeldesign.NewUngroupedGroup()
		assert.Equal(t, modeldesign.UngroupedGroupID, g.ID)
		assert.Equal(t, modeldesign.UngroupedGroupName, g.Name)
		assert.True(t, g.IsVirtual())
	})
}
