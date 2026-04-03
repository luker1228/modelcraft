package core

import (
	"testing"
)

func TestFieldDefinition_GetJSONType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  string
	}{
		{"string type", FieldTypeString, "string"},
		{"integer type", FieldTypeInteger, "integer"},
		{"float type", FieldTypeFloat, "number"},
		{"boolean type", FieldTypeBoolean, "boolean"},
		{"id type", FieldTypeID, "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &FieldDefinition{Type: tt.fieldType}
			result := field.GetJSONType()
			if result != tt.expected {
				t.Errorf("GetJSONType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationRules(t *testing.T) {
	minLen := 5
	maxLen := 100
	pattern := "^[a-zA-Z]+$"
	min := 0.0
	max := 100.0

	validation := &ValidationRules{
		MinLength: &minLen,
		MaxLength: &maxLen,
		Pattern:   &pattern,
		Minimum:   &min,
		Maximum:   &max,
	}

	if *validation.MinLength != 5 {
		t.Errorf("MinLength = %v, want %v", *validation.MinLength, 5)
	}
	if *validation.MaxLength != 100 {
		t.Errorf("MaxLength = %v, want %v", *validation.MaxLength, 100)
	}
	if *validation.Pattern != "^[a-zA-Z]+$" {
		t.Errorf("Pattern = %v, want %v", *validation.Pattern, "^[a-zA-Z]+$")
	}
	if *validation.Minimum != 0.0 {
		t.Errorf("Minimum = %v, want %v", *validation.Minimum, 0.0)
	}
	if *validation.Maximum != 100.0 {
		t.Errorf("Maximum = %v, want %v", *validation.Maximum, 100.0)
	}
}

func TestNewFieldDefinition(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		fieldType FieldType
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid field",
			key:       "username",
			fieldType: FieldTypeString,
			wantErr:   false,
		},
		{
			name:      "empty key",
			key:       "",
			fieldType: FieldTypeString,
			wantErr:   true,
			errMsg:    "key is required and cannot be empty",
		},
		{
			name:      "invalid field type",
			key:       "test",
			fieldType: FieldType("invalid"),
			wantErr:   true,
			errMsg:    "invalid field type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, err := NewFieldDefinition(tt.key, tt.fieldType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFieldDefinition() should return error")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("NewFieldDefinition() error = %v, want to contain %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewFieldDefinition() unexpected error = %v", err)
				return
			}

			if field.Key != tt.key {
				t.Errorf("NewFieldDefinition() Key = %v, want %v", field.Key, tt.key)
			}
			if field.Type != tt.fieldType {
				t.Errorf("NewFieldDefinition() Type = %v, want %v", field.Type, tt.fieldType)
			}
		})
	}
}

func TestFieldDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		field   *FieldDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid field",
			field: &FieldDefinition{
				Key:  "username",
				Type: FieldTypeString,
			},
			wantErr: false,
		},
		{
			name: "empty key",
			field: &FieldDefinition{
				Key:  "",
				Type: FieldTypeString,
			},
			wantErr: true,
			errMsg:  "field key is required",
		},
		{
			name: "invalid type",
			field: &FieldDefinition{
				Key:  "test",
				Type: FieldType("invalid"),
			},
			wantErr: true,
			errMsg:  "invalid field type",
		},
		{
			name: "valid with string default",
			field: &FieldDefinition{
				Key:     "name",
				Type:    FieldTypeString,
				Default: "test",
			},
			wantErr: false,
		},
		{
			name: "invalid string default",
			field: &FieldDefinition{
				Key:     "name",
				Type:    FieldTypeString,
				Default: 123,
			},
			wantErr: true,
			errMsg:  "default value for string field must be string",
		},
		{
			name: "valid integer default",
			field: &FieldDefinition{
				Key:     "age",
				Type:    FieldTypeInteger,
				Default: 25,
			},
			wantErr: false,
		},
		{
			name: "invalid integer default",
			field: &FieldDefinition{
				Key:     "age",
				Type:    FieldTypeInteger,
				Default: "25",
			},
			wantErr: true,
			errMsg:  "default value for integer field must be integer",
		},
		{
			name: "valid float default",
			field: &FieldDefinition{
				Key:     "price",
				Type:    FieldTypeFloat,
				Default: 99.99,
			},
			wantErr: false,
		},
		{
			name: "valid boolean default",
			field: &FieldDefinition{
				Key:     "active",
				Type:    FieldTypeBoolean,
				Default: true,
			},
			wantErr: false,
		},
		{
			name: "invalid boolean default",
			field: &FieldDefinition{
				Key:     "active",
				Type:    FieldTypeBoolean,
				Default: "true",
			},
			wantErr: true,
			errMsg:  "default value for boolean field must be boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.field.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() should return error")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want to contain %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestFieldDefinition_ChainMethods(t *testing.T) {
	field, err := NewFieldDefinition("username", FieldTypeString)
	if err != nil {
		t.Fatalf("NewFieldDefinition() failed: %v", err)
	}

	// 测试链式调用
	result := field.
		WithRequired(true).
		WithDescription("User username").
		WithDefault("anonymous").
		WithValidation(&ValidationRules{
			MinLength: intPtr(3),
			MaxLength: intPtr(50),
		})

	// 验证链式调用返回同一实例
	if result != field {
		t.Error("Chain methods should return same instance")
	}

	// 验证设置的值
	if !field.Required {
		t.Error("WithRequired() failed")
	}
	if field.Description != "User username" {
		t.Errorf("WithDescription() = %v, want %v", field.Description, "User username")
	}
	if field.Default != "anonymous" {
		t.Errorf("WithDefault() = %v, want %v", field.Default, "anonymous")
	}
	if field.Validation == nil || *field.Validation.MinLength != 3 {
		t.Error("WithValidation() failed")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

func intPtr(i int) *int {
	return &i
}
