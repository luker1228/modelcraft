package modeldesign

import (
	"modelcraft/internal/domain/project"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEnumOption_Validate(t *testing.T) {
	tests := []struct {
		name        string
		option      EnumOption
		wantErr     bool
		errContains string
	}{
		{
			name: "valid option",
			option: EnumOption{
				Code:  "ACTIVE",
				Label: "Active",
				Order: 1,
			},
			wantErr: false,
		},
		{
			name: "valid option with description",
			option: EnumOption{
				Code:        "PENDING",
				Label:       "Pending",
				Order:       2,
				Description: "Order is pending",
			},
			wantErr: false,
		},
		{
			name: "empty code",
			option: EnumOption{
				Code:  "",
				Label: "Active",
				Order: 1,
			},
			wantErr:     true,
			errContains: "enum option code cannot be empty",
		},
		{
			name: "whitespace only code",
			option: EnumOption{
				Code:  "   ",
				Label: "Active",
				Order: 1,
			},
			wantErr:     true,
			errContains: "enum option code cannot be empty",
		},
		{
			name: "empty label",
			option: EnumOption{
				Code:  "ACTIVE",
				Label: "",
				Order: 1,
			},
			wantErr:     true,
			errContains: "enum option label cannot be empty",
		},
		{
			name: "whitespace only label",
			option: EnumOption{
				Code:  "ACTIVE",
				Label: "   ",
				Order: 1,
			},
			wantErr:     true,
			errContains: "enum option label cannot be empty",
		},
		{
			name: "zero order is valid",
			option: EnumOption{
				Code:  "ACTIVE",
				Label: "Active",
				Order: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := tt.option.Validate()

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnumOption_String(t *testing.T) {
	tests := []struct {
		name     string
		option   EnumOption
		expected string
	}{
		{
			name: "standard option",
			option: EnumOption{
				Code:  "ACTIVE",
				Label: "Active",
			},
			expected: "EnumOption(ACTIVE:Active)",
		},
		{
			name: "option with empty fields",
			option: EnumOption{
				Code:  "",
				Label: "",
			},
			expected: "EnumOption(:)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := tt.option.String()

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnumDefinition_Validate(t *testing.T) {
	validOptions := []EnumOption{
		{Code: "ACTIVE", Label: "Active", Order: 1},
		{Code: "INACTIVE", Label: "Inactive", Order: 2},
	}

	tests := []struct {
		name        string
		enum        EnumDefinition
		wantErr     bool
		errContains string
	}{
		{
			name: "valid enum definition",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
				Name:         "order_status",
				DisplayName:  "Order Status",
				Options:      validOptions,
			},
			wantErr: false,
		},
		{
			name: "valid enum with multi-select",
			enum: EnumDefinition{
				ProjectScope:  project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
				Name:          "tags",
				DisplayName:   "Tags",
				Options:       validOptions,
				IsMultiSelect: true,
			},
			wantErr: false,
		},
		{
			name: "missing org name",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "", ProjectSlug: "ecommerce"},
				Name:         "order_status",
				DisplayName:  "Order Status",
				Options:      validOptions,
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name: "whitespace only org name",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "   ", ProjectSlug: "ecommerce"},
				Name:         "order_status",
				DisplayName:  "Order Status",
				Options:      validOptions,
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name: "missing project slug",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: ""},
				Name:         "order_status",
				DisplayName:  "Order Status",
				Options:      validOptions,
			},
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name: "missing name",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
				Name:         "",
				DisplayName:  "Order Status",
				Options:      validOptions,
			},
			wantErr:     true,
			errContains: "enum name cannot be empty",
		},
		{
			name: "missing display name",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
				Name:         "order_status",
				DisplayName:  "",
				Options:      validOptions,
			},
			wantErr:     true,
			errContains: "enum display name cannot be empty",
		},
		{
			name: "no options",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
				Name:         "order_status",
				DisplayName:  "Order Status",
				Options:      []EnumOption{},
			},
			wantErr:     true,
			errContains: "enum must have at least one option",
		},
		{
			name: "invalid option",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
				Name:         "order_status",
				DisplayName:  "Order Status",
				Options: []EnumOption{
					{Code: "", Label: "Active", Order: 1},
				},
			},
			wantErr:     true,
			errContains: "invalid enum option",
		},
		{
			name: "duplicate option codes",
			enum: EnumDefinition{
				ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
				Name:         "order_status",
				DisplayName:  "Order Status",
				Options: []EnumOption{
					{Code: "ACTIVE", Label: "Active", Order: 1},
					{Code: "ACTIVE", Label: "Active Again", Order: 2},
				},
			},
			wantErr:     true,
			errContains: "duplicate enum option code: ACTIVE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := tt.enum.Validate()

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnumDefinition_GetOptionByCode(t *testing.T) {
	// Given
	enum := EnumDefinition{
		Options: []EnumOption{
			{Code: "ACTIVE", Label: "Active", Order: 1},
			{Code: "INACTIVE", Label: "Inactive", Order: 2},
			{Code: "PENDING", Label: "Pending", Order: 3},
		},
	}

	tests := []struct {
		name        string
		code        string
		wantErr     bool
		expectedOpt *EnumOption
		errContains string
	}{
		{
			name:    "existing code - first",
			code:    "ACTIVE",
			wantErr: false,
			expectedOpt: &EnumOption{
				Code:  "ACTIVE",
				Label: "Active",
				Order: 1,
			},
		},
		{
			name:    "existing code - middle",
			code:    "INACTIVE",
			wantErr: false,
			expectedOpt: &EnumOption{
				Code:  "INACTIVE",
				Label: "Inactive",
				Order: 2,
			},
		},
		{
			name:    "existing code - last",
			code:    "PENDING",
			wantErr: false,
			expectedOpt: &EnumOption{
				Code:  "PENDING",
				Label: "Pending",
				Order: 3,
			},
		},
		{
			name:        "non-existing code",
			code:        "DELETED",
			wantErr:     true,
			errContains: "enum option not found: DELETED",
		},
		{
			name:        "empty code",
			code:        "",
			wantErr:     true,
			errContains: "enum option not found:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			opt, err := enum.GetOptionByCode(tt.code)

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOpt.Code, opt.Code)
				assert.Equal(t, tt.expectedOpt.Label, opt.Label)
				assert.Equal(t, tt.expectedOpt.Order, opt.Order)
			}
		})
	}
}

func TestEnumDefinition_HasOptionCode(t *testing.T) {
	// Given
	enum := EnumDefinition{
		Options: []EnumOption{
			{Code: "ACTIVE", Label: "Active", Order: 1},
			{Code: "INACTIVE", Label: "Inactive", Order: 2},
		},
	}

	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "existing code",
			code:     "ACTIVE",
			expected: true,
		},
		{
			name:     "another existing code",
			code:     "INACTIVE",
			expected: true,
		},
		{
			name:     "non-existing code",
			code:     "PENDING",
			expected: false,
		},
		{
			name:     "empty code",
			code:     "",
			expected: false,
		},
		{
			name:     "case sensitive",
			code:     "active",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := enum.HasOptionCode(tt.code)

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnumDefinition_ValidateCodes(t *testing.T) {
	// Given
	enum := EnumDefinition{
		Options: []EnumOption{
			{Code: "ACTIVE", Label: "Active", Order: 1},
			{Code: "INACTIVE", Label: "Inactive", Order: 2},
			{Code: "PENDING", Label: "Pending", Order: 3},
		},
	}

	tests := []struct {
		name        string
		codes       []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "all valid codes",
			codes:   []string{"ACTIVE", "INACTIVE"},
			wantErr: false,
		},
		{
			name:    "single valid code",
			codes:   []string{"PENDING"},
			wantErr: false,
		},
		{
			name:    "all existing codes",
			codes:   []string{"ACTIVE", "INACTIVE", "PENDING"},
			wantErr: false,
		},
		{
			name:    "empty list",
			codes:   []string{},
			wantErr: false,
		},
		{
			name:        "one invalid code",
			codes:       []string{"DELETED"},
			wantErr:     true,
			errContains: "invalid enum code: DELETED",
		},
		{
			name:        "mix of valid and invalid",
			codes:       []string{"ACTIVE", "DELETED"},
			wantErr:     true,
			errContains: "invalid enum code: DELETED",
		},
		{
			name:        "empty string code",
			codes:       []string{""},
			wantErr:     true,
			errContains: "invalid enum code:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := enum.ValidateCodes(tt.codes)

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnumDefinition_Update(t *testing.T) {
	validOptions := []EnumOption{
		{Code: "ACTIVE", Label: "Active", Order: 1},
		{Code: "INACTIVE", Label: "Inactive", Order: 2},
	}

	newOptions := []EnumOption{
		{Code: "NEW1", Label: "New 1", Order: 1},
		{Code: "NEW2", Label: "New 2", Order: 2},
	}

	tests := []struct {
		name                string
		initialDisplayName  string
		initialDesc         string
		initialOptions      []EnumOption
		updateDisplayName   *string
		updateDesc          *string
		updateOptions       []EnumOption
		wantErr             bool
		errContains         string
		expectedDisplayName string
		expectedDesc        string
		expectedOptions     []EnumOption
	}{
		{
			name:                "update display name only",
			initialDisplayName:  "Old Title",
			initialDesc:         "Old Desc",
			initialOptions:      validOptions,
			updateDisplayName:   stringPtr("New Title"),
			updateDesc:          nil,
			updateOptions:       nil,
			wantErr:             false,
			expectedDisplayName: "New Title",
			expectedDesc:        "Old Desc",
			expectedOptions:     validOptions,
		},
		{
			name:                "update description only",
			initialDisplayName:  "Old Title",
			initialDesc:         "Old Desc",
			initialOptions:      validOptions,
			updateDisplayName:   nil,
			updateDesc:          stringPtr("New Desc"),
			updateOptions:       nil,
			wantErr:             false,
			expectedDisplayName: "Old Title",
			expectedDesc:        "New Desc",
			expectedOptions:     validOptions,
		},
		{
			name:                "update options only",
			initialDisplayName:  "Old Title",
			initialDesc:         "Old Desc",
			initialOptions:      validOptions,
			updateDisplayName:   nil,
			updateDesc:          nil,
			updateOptions:       newOptions,
			wantErr:             false,
			expectedDisplayName: "Old Title",
			expectedDesc:        "Old Desc",
			expectedOptions:     newOptions,
		},
		{
			name:                "update all fields",
			initialDisplayName:  "Old Title",
			initialDesc:         "Old Desc",
			initialOptions:      validOptions,
			updateDisplayName:   stringPtr("New Title"),
			updateDesc:          stringPtr("New Desc"),
			updateOptions:       newOptions,
			wantErr:             false,
			expectedDisplayName: "New Title",
			expectedDesc:        "New Desc",
			expectedOptions:     newOptions,
		},
		{
			name:                "update nothing",
			initialDisplayName:  "Old Title",
			initialDesc:         "Old Desc",
			initialOptions:      validOptions,
			updateDisplayName:   nil,
			updateDesc:          nil,
			updateOptions:       nil,
			wantErr:             false,
			expectedDisplayName: "Old Title",
			expectedDesc:        "Old Desc",
			expectedOptions:     validOptions,
		},
		{
			name:               "empty display name",
			initialDisplayName: "Old Title",
			initialDesc:        "Old Desc",
			initialOptions:     validOptions,
			updateDisplayName:  stringPtr(""),
			updateDesc:         nil,
			updateOptions:      nil,
			wantErr:            true,
			errContains:        "enum display name cannot be empty",
		},
		{
			name:               "whitespace only display name",
			initialDisplayName: "Old Title",
			initialDesc:        "Old Desc",
			initialOptions:     validOptions,
			updateDisplayName:  stringPtr("   "),
			updateDesc:         nil,
			updateOptions:      nil,
			wantErr:            true,
			errContains:        "enum display name cannot be empty",
		},
		{
			name:               "empty options list",
			initialDisplayName: "Old Title",
			initialDesc:        "Old Desc",
			initialOptions:     validOptions,
			updateDisplayName:  nil,
			updateDesc:         nil,
			updateOptions:      []EnumOption{},
			wantErr:            true,
			errContains:        "enum must have at least one option",
		},
		{
			name:               "invalid option in update",
			initialDisplayName: "Old Title",
			initialDesc:        "Old Desc",
			initialOptions:     validOptions,
			updateDisplayName:  nil,
			updateDesc:         nil,
			updateOptions: []EnumOption{
				{Code: "", Label: "Invalid", Order: 1},
			},
			wantErr:     true,
			errContains: "invalid enum option",
		},
		{
			name:               "duplicate codes in update",
			initialDisplayName: "Old Title",
			initialDesc:        "Old Desc",
			initialOptions:     validOptions,
			updateDisplayName:  nil,
			updateDesc:         nil,
			updateOptions: []EnumOption{
				{Code: "DUP", Label: "First", Order: 1},
				{Code: "DUP", Label: "Second", Order: 2},
			},
			wantErr:     true,
			errContains: "duplicate enum option code: DUP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			oldTime := time.Now().Add(-1 * time.Hour)
			enum := &EnumDefinition{
				DisplayName: tt.initialDisplayName,
				Description: tt.initialDesc,
				Options:     tt.initialOptions,
				UpdatedAt:   oldTime,
			}

			// When
			err := enum.Update(tt.updateDisplayName, tt.updateDesc, tt.updateOptions)

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDisplayName, enum.DisplayName)
				assert.Equal(t, tt.expectedDesc, enum.Description)
				assert.Equal(t, tt.expectedOptions, enum.Options)
				assert.True(t, enum.UpdatedAt.After(oldTime))
			}
		})
	}
}

func TestEnumDefinition_Clone(t *testing.T) {
	// Given
	now := time.Now()
	original := &EnumDefinition{
		ID:           "enum-123",
		ProjectScope: project.ProjectScope{OrgName: "myorg", ProjectSlug: "ecommerce"},
		Name:         "order_status",
		DisplayName:  "Order Status",
		Description:  "Status of orders",
		Options: []EnumOption{
			{Code: "ACTIVE", Label: "Active", Order: 1},
			{Code: "INACTIVE", Label: "Inactive", Order: 2},
		},
		IsMultiSelect: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// When
	cloned := original.Clone()

	// Then
	assert.NotSame(t, original, cloned)
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.OrgName, cloned.OrgName)
	assert.Equal(t, original.ProjectSlug, cloned.ProjectSlug)
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.DisplayName, cloned.DisplayName)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, original.IsMultiSelect, cloned.IsMultiSelect)
	assert.Equal(t, original.CreatedAt, cloned.CreatedAt)
	assert.Equal(t, original.UpdatedAt, cloned.UpdatedAt)

	// Verify options are deeply copied
	assert.Equal(t, original.Options, cloned.Options)
	assert.NotSame(t, &original.Options[0], &cloned.Options[0])

	// Verify modifying clone doesn't affect original
	cloned.DisplayName = "Modified Title"
	cloned.Options[0].Label = "Modified Label"
	assert.NotEqual(t, original.DisplayName, cloned.DisplayName)
	assert.NotEqual(t, original.Options[0].Label, cloned.Options[0].Label)
}

func TestEnumDefinition_String(t *testing.T) {
	tests := []struct {
		name     string
		enum     EnumDefinition
		expected string
	}{
		{
			name: "standard enum",
			enum: EnumDefinition{
				Name:        "order_status",
				DisplayName: "Order Status",
			},
			expected: "Enum(order_status:Order Status)",
		},
		{
			name: "empty fields",
			enum: EnumDefinition{
				Name:        "",
				DisplayName: "",
			},
			expected: "Enum(:)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := tt.enum.String()

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}
