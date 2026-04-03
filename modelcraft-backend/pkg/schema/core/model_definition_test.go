package core

import (
	"testing"
)

func TestModelDefinition_GetRequiredFields(t *testing.T) {
	model := &ModelDefinition{
		Name: "User",
		Fields: []*FieldDefinition{
			{Key: "name", Type: FieldTypeString, Required: true},
			{Key: "email", Type: FieldTypeString, Required: true},
			{Key: "age", Type: FieldTypeInteger, Required: false},
		},
	}

	required := model.GetRequiredFields()
	expected := []string{"name", "email"}

	if len(required) != len(expected) {
		t.Errorf("GetRequiredFields() length = %v, want %v", len(required), len(expected))
		return
	}

	for i, field := range required {
		if field != expected[i] {
			t.Errorf("GetRequiredFields()[%d] = %v, want %v", i, field, expected[i])
		}
	}
}

func TestModelDefinition_GetFieldByKey(t *testing.T) {
	model := &ModelDefinition{
		Name: "User",
		Fields: []*FieldDefinition{
			{Key: "name", Type: FieldTypeString},
			{Key: "age", Type: FieldTypeInteger},
		},
	}

	// 测试存在的字段
	field := model.GetFieldByKey("name")
	if field == nil {
		t.Error("GetFieldByKey('name') returned nil")
		return
	}
	if field.Key != "name" || field.Type != "string" {
		t.Errorf("GetFieldByKey('name') = %+v, want Key='name' Type='string'", field)
	}

	// 测试不存在的字段
	field = model.GetFieldByKey("nonexistent")
	if field != nil {
		t.Errorf("GetFieldByKey('nonexistent') = %+v, want nil", field)
	}
}

func TestModelDefinition_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		model *ModelDefinition
		valid bool
	}{
		{
			name: "valid modeldesign",
			model: &ModelDefinition{
				Name: "User",
				Fields: []*FieldDefinition{
					{Key: "name", Type: FieldTypeString},
					{Key: "age", Type: FieldTypeInteger},
				},
			},
			valid: true,
		},
		{
			name: "empty name",
			model: &ModelDefinition{
				Name: "",
				Fields: []*FieldDefinition{
					{Key: "name", Type: FieldTypeString},
				},
			},
			valid: false,
		},
		{
			name: "no fields",
			model: &ModelDefinition{
				Name:   "User",
				Fields: []*FieldDefinition{},
			},
			valid: false,
		},
		{
			name: "invalid field",
			model: &ModelDefinition{
				Name: "User",
				Fields: []*FieldDefinition{
					{Key: "", Type: FieldTypeString},
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.IsValid()
			if result != tt.valid {
				t.Errorf("IsValid() = %v, want %v", result, tt.valid)
			}
		})
	}
}

func TestModelDefinition_FieldCount(t *testing.T) {
	model := &ModelDefinition{
		Name: "User",
		Fields: []*FieldDefinition{
			{Key: "name", Type: FieldTypeString},
			{Key: "age", Type: FieldTypeInteger},
			{Key: "email", Type: FieldTypeString},
		},
	}

	count := model.FieldCount()
	if count != 3 {
		t.Errorf("FieldCount() = %v, want %v", count, 3)
	}
}
