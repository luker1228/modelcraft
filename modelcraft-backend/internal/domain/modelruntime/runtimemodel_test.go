package modelruntime

import (
	"modelcraft/internal/domain/modeldesign"
	"testing"
)

// TestRuntimeModelGetUniqueField tests RuntimeModel.getUniqueField method
func TestRuntimeModelGetUniqueField(t *testing.T) {
	tests := []struct {
		name      string
		model     *RuntimeModel
		wantCount int
	}{
		{
			name: "model with multiple unique fields",
			model: &RuntimeModel{
				Name: "TestModel",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:     "id",
						IsUnique: true,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatUUID,
						},
					},
					"email": {
						Name:     "email",
						IsUnique: true,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatString,
						},
					},
					"username": {
						Name:     "username",
						IsUnique: true,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatString,
						},
					},
					"name": {
						Name:     "name",
						IsUnique: false,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatString,
						},
					},
				},
			},
			wantCount: 3,
		},
		{
			name: "model with single unique field",
			model: &RuntimeModel{
				Name: "SingleUniqueModel",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:     "id",
						IsUnique: true,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatUUID,
						},
					},
					"name": {
						Name:     "name",
						IsUnique: false,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatString,
						},
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "model without unique fields",
			model: &RuntimeModel{
				Name: "NoUniqueModel",
				Fields: map[string]*RuntimeField{
					"name": {
						Name:     "name",
						IsUnique: false,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatString,
						},
					},
					"age": {
						Name:     "age",
						IsUnique: false,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatInteger,
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "model with nil fields",
			model: &RuntimeModel{
				Name:   "NilFieldsModel",
				Fields: nil,
			},
			wantCount: 0,
		},
		{
			name: "model with empty fields",
			model: &RuntimeModel{
				Name:   "EmptyFieldsModel",
				Fields: map[string]*RuntimeField{},
			},
			wantCount: 0,
		},
		{
			name: "all fields are unique",
			model: &RuntimeModel{
				Name: "AllUniqueModel",
				Fields: map[string]*RuntimeField{
					"id": {
						Name:     "id",
						IsUnique: true,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatUUID,
						},
					},
					"email": {
						Name:     "email",
						IsUnique: true,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatString,
						},
					},
					"username": {
						Name:     "username",
						IsUnique: true,
						Type: &modeldesign.FieldType{
							Format: modeldesign.FormatString,
						},
					},
				},
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When: calling getUniqueField
			uniqueFields := tt.model.getUniqueField()

			// Then: correct number of unique fields should be returned
			if len(uniqueFields) != tt.wantCount {
				t.Errorf("getUniqueField() returned %d unique fields, want %d", len(uniqueFields), tt.wantCount)
			}

			// Verify all returned fields are actually unique
			for _, field := range uniqueFields {
				if !field.IsUnique {
					t.Errorf("getUniqueField() returned non-unique field: %s", field.Name)
				}
			}
		})
	}
}

// TestGetGraphqlTypeBy tests getGraphqlTypeBy function
func TestGetGraphqlTypeBy(t *testing.T) {
	tests := []struct {
		name       string
		formatType modeldesign.FormatType
		wantErr    bool
	}{
		{
			name:       "string format",
			formatType: modeldesign.FormatString,
			wantErr:    false,
		},
		{
			name:       "uuid format",
			formatType: modeldesign.FormatUUID,
			wantErr:    false,
		},
		{
			name:       "number format",
			formatType: modeldesign.FormatNumber,
			wantErr:    false,
		},
		{
			name:       "decimal format",
			formatType: modeldesign.FormatDecimal,
			wantErr:    false,
		},
		{
			name:       "integer format",
			formatType: modeldesign.FormatInteger,
			wantErr:    false,
		},
		{
			name:       "boolean format",
			formatType: modeldesign.FormatBoolean,
			wantErr:    false,
		},
		{
			name:       "datetime format",
			formatType: modeldesign.FormatDateTime,
			wantErr:    false,
		},
		{
			name:       "date format",
			formatType: modeldesign.FormatDate,
			wantErr:    false,
		},
		{
			name:       "time format",
			formatType: modeldesign.FormatTime,
			wantErr:    false,
		},
		{
			name:       "enum format",
			formatType: modeldesign.FormatEnum,
			wantErr:    false,
		},
		{
			name:       "unknown format",
			formatType: "UNKNOWN_FORMAT",
			wantErr:    true,
		},
		{
			name:       "empty format",
			formatType: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When: getting GraphQL type by format
			scalar, err := getGraphqlTypeBy(tt.formatType)

			// Then: result should match expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("getGraphqlTypeBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if scalar == nil {
					t.Error("getGraphqlTypeBy() returned nil scalar for valid format")
				}
			}
		})
	}
}

// TestNewEnumLabel tests NewEnumLabel function
func TestNewEnumLabel(t *testing.T) {
	tests := []struct {
		name string
		opt  modeldesign.EnumOption
		want EnumLabel
	}{
		{
			name: "enum option with description",
			opt: modeldesign.EnumOption{
				Code:        "ACTIVE",
				Label:       "Active",
				Description: "User is active",
			},
			want: EnumLabel{
				Code:        "ACTIVE",
				Label:       "Active",
				Description: func() *string { s := "User is active"; return &s }(),
			},
		},
		{
			name: "enum option without description",
			opt: modeldesign.EnumOption{
				Code:        "INACTIVE",
				Label:       "Inactive",
				Description: "",
			},
			want: EnumLabel{
				Code:        "INACTIVE",
				Label:       "Inactive",
				Description: nil,
			},
		},
		{
			name: "enum option with empty strings",
			opt: modeldesign.EnumOption{
				Code:        "",
				Label:       "",
				Description: "",
			},
			want: EnumLabel{
				Code:        "",
				Label:       "",
				Description: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When: creating EnumLabel
			got := NewEnumLabel(tt.opt)

			// Then: values should match expected
			if got.Code != tt.want.Code {
				t.Errorf("NewEnumLabel() Code = %v, want %v", got.Code, tt.want.Code)
			}

			if got.Label != tt.want.Label {
				t.Errorf("NewEnumLabel() Label = %v, want %v", got.Label, tt.want.Label)
			}

			// Check description
			if tt.want.Description == nil {
				if got.Description != nil {
					t.Errorf("NewEnumLabel() Description = %v, want nil", *got.Description)
				}
				return
			}
			if got.Description == nil {
				t.Error("NewEnumLabel() Description is nil, want non-nil")
				return
			}
			if *got.Description != *tt.want.Description {
				t.Errorf("NewEnumLabel() Description = %v, want %v", *got.Description, *tt.want.Description)
			}
		})
	}
}

// TestRuntimeModelFields tests RuntimeModel field access
func TestRuntimeModelFields(t *testing.T) {
	// Given: a runtime model with various fields
	model := &RuntimeModel{
		OrgName:      "testorg",
		ProjectSlug:  "testproject",
		Name:         "TestModel",
		Title:        "Test Model",
		Description:  "A test model for testing",
		DatabaseName: "testdb",
		Fields: map[string]*RuntimeField{
			"id": {
				Name:     "id",
				Title:    "ID",
				IsUnique: true,
				Required: true,
				Type: &modeldesign.FieldType{
					Format: modeldesign.FormatUUID,
				},
			},
		},
	}

	// Then: all fields should be accessible
	if model.OrgName != "testorg" {
		t.Errorf("OrgName = %v, want testorg", model.OrgName)
	}

	if model.ProjectSlug != "testproject" {
		t.Errorf("ProjectSlug = %v, want testproject", model.ProjectSlug)
	}

	if model.Name != "TestModel" {
		t.Errorf("Name = %v, want TestModel", model.Name)
	}

	if model.Title != "Test Model" {
		t.Errorf("Title = %v, want Test Model", model.Title)
	}

	if model.Description != "A test model for testing" {
		t.Errorf("Description = %v, want 'A test model for testing'", model.Description)
	}

	if model.DatabaseName != "testdb" {
		t.Errorf("DatabaseName = %v, want testdb", model.DatabaseName)
	}

	if model.Fields == nil {
		t.Fatal("Fields is nil")
	}

	if len(model.Fields) != 1 {
		t.Errorf("Fields length = %d, want 1", len(model.Fields))
	}

	idField, exists := model.Fields["id"]
	if !exists {
		t.Fatal("id field not found")
	}

	if !idField.IsUnique {
		t.Error("id field should be unique")
	}

	if !idField.Required {
		t.Error("id field should be required")
	}
}

// TestGetUniqueFieldWithMixedFields tests getUniqueField with various field combinations
func TestGetUniqueFieldWithMixedFields(t *testing.T) {
	// Given: a model with mixed unique/non-unique fields
	model := &RuntimeModel{
		Name: "MixedModel",
		Fields: map[string]*RuntimeField{
			"id": {
				Name:     "id",
				IsUnique: true,
				Required: true,
				Type: &modeldesign.FieldType{
					Format: modeldesign.FormatUUID,
				},
			},
			"name": {
				Name:     "name",
				IsUnique: false,
				Required: true,
				Type: &modeldesign.FieldType{
					Format: modeldesign.FormatString,
				},
			},
			"email": {
				Name:     "email",
				IsUnique: true,
				Required: false,
				Type: &modeldesign.FieldType{
					Format: modeldesign.FormatString,
				},
			},
			"age": {
				Name:     "age",
				IsUnique: false,
				Required: false,
				Type: &modeldesign.FieldType{
					Format: modeldesign.FormatInteger,
				},
			},
		},
	}

	// When: getting unique fields
	uniqueFields := model.getUniqueField()

	// Then: should return only unique fields
	if len(uniqueFields) != 2 {
		t.Errorf("getUniqueField() returned %d fields, want 2", len(uniqueFields))
	}

	// Check that returned fields are the correct ones
	foundID := false
	foundEmail := false

	for _, field := range uniqueFields {
		if field.Name == "id" {
			foundID = true
			if !field.IsUnique {
				t.Error("id field should be unique")
			}
		}
		if field.Name == "email" {
			foundEmail = true
			if !field.IsUnique {
				t.Error("email field should be unique")
			}
		}
		// Ensure no non-unique fields are included
		if field.Name == "name" || field.Name == "age" {
			t.Errorf("getUniqueField() included non-unique field: %s", field.Name)
		}
	}

	if !foundID {
		t.Error("getUniqueField() did not include id field")
	}

	if !foundEmail {
		t.Error("getUniqueField() did not include email field")
	}
}

// TestGetGraphqlTypeByAllFormats tests getGraphqlTypeBy for all supported formats
func TestGetGraphqlTypeByAllFormats(t *testing.T) {
	// Test all valid formats
	validFormats := []modeldesign.FormatType{
		modeldesign.FormatString,
		modeldesign.FormatUUID,
		modeldesign.FormatNumber,
		modeldesign.FormatDecimal,
		modeldesign.FormatInteger,
		modeldesign.FormatBoolean,
		modeldesign.FormatDateTime,
		modeldesign.FormatDate,
		modeldesign.FormatTime,
		modeldesign.FormatEnum,
	}

	for _, format := range validFormats {
		t.Run(string(format), func(t *testing.T) {
			scalar, err := getGraphqlTypeBy(format)
			if err != nil {
				t.Errorf("getGraphqlTypeBy(%s) unexpected error: %v", format, err)
			}
			if scalar == nil {
				t.Errorf("getGraphqlTypeBy(%s) returned nil scalar", format)
			}
		})
	}
}

// TestEnumLabelPointerHandling tests EnumLabel description pointer handling
func TestEnumLabelPointerHandling(t *testing.T) {
	// Given: enum options with various description scenarios
	tests := []struct {
		name        string
		description string
		wantNil     bool
	}{
		{"empty description", "", true},
		{"non-empty description", "Some description", false},
		{"whitespace description", "   ", false}, // Not empty, but whitespace
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := modeldesign.EnumOption{
				Code:        "TEST",
				Label:       "Test",
				Description: tt.description,
			}

			// When: creating EnumLabel
			label := NewEnumLabel(opt)

			// Then: description pointer should match expectations
			if tt.wantNil {
				if label.Description != nil {
					t.Errorf(
						"NewEnumLabel() Description should be nil for empty description, got %v",
						*label.Description,
					)
				}
				return
			}
			if label.Description == nil {
				t.Error("NewEnumLabel() Description should not be nil for non-empty description")
				return
			}
			if *label.Description != tt.description {
				t.Errorf("NewEnumLabel() Description = %v, want %v", *label.Description, tt.description)
			}
		})
	}
}

// BenchmarkGetUniqueField benchmarks getUniqueField method
func BenchmarkGetUniqueField(b *testing.B) {
	model := &RuntimeModel{
		Name: "BenchModel",
		Fields: map[string]*RuntimeField{
			"id": {
				Name:     "id",
				IsUnique: true,
			},
			"email": {
				Name:     "email",
				IsUnique: true,
			},
			"username": {
				Name:     "username",
				IsUnique: true,
			},
			"name": {
				Name:     "name",
				IsUnique: false,
			},
			"age": {
				Name:     "age",
				IsUnique: false,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.getUniqueField()
	}
}

// BenchmarkGetGraphqlTypeBy benchmarks getGraphqlTypeBy function
func BenchmarkGetGraphqlTypeBy(b *testing.B) {
	formats := []modeldesign.FormatType{
		modeldesign.FormatString,
		modeldesign.FormatInteger,
		modeldesign.FormatNumber,
		modeldesign.FormatBoolean,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		format := formats[i%len(formats)]
		_, _ = getGraphqlTypeBy(format)
	}
}

// BenchmarkNewEnumLabel benchmarks NewEnumLabel function
func BenchmarkNewEnumLabel(b *testing.B) {
	opt := modeldesign.EnumOption{
		Code:        "ACTIVE",
		Label:       "Active",
		Description: "User is active",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewEnumLabel(opt)
	}
}
