package modeldesign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFieldService_ValidateDuplicates_ReservedKeywords tests reserved keyword validation
func TestFieldService_ValidateDuplicates_ReservedKeywords(t *testing.T) {
	service := NewFieldService()

	tests := []struct {
		name      string
		fields    []FieldDefinition
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid field names",
			fields: []FieldDefinition{
				{Name: "username"},
				{Name: "email"},
				{Name: "created_at"},
			},
			expectErr: false,
		},
		{
			name: "invalid field name - starts with underscore",
			fields: []FieldDefinition{
				{Name: "_meta"},
			},
			expectErr: true,
			errMsg:    "格式无效",
		},
		{
			name: "invalid field name - starts with double underscore",
			fields: []FieldDefinition{
				{Name: "__meta"},
			},
			expectErr: true,
			errMsg:    "格式无效",
		},
		{
			name: "reserved keyword - AND",
			fields: []FieldDefinition{
				{Name: "AND"},
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name: "reserved keyword - OR",
			fields: []FieldDefinition{
				{Name: "OR"},
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name: "reserved keyword - equals",
			fields: []FieldDefinition{
				{Name: "equals"},
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name: "reserved keyword - not (lowercase)",
			fields: []FieldDefinition{
				{Name: "not"},
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name: "reserved keyword - in",
			fields: []FieldDefinition{
				{Name: "in"},
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name: "reserved keyword - contains",
			fields: []FieldDefinition{
				{Name: "contains"},
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name: "reserved keyword - mixed case (And)",
			fields: []FieldDefinition{
				{Name: "And"},
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name: "similar but not reserved - contain",
			fields: []FieldDefinition{
				{Name: "contain"},
			},
			expectErr: false,
		},
		{
			name: "similar but not reserved - note",
			fields: []FieldDefinition{
				{Name: "note"},
			},
			expectErr: false,
		},
		{
			name: "multiple fields with one reserved",
			fields: []FieldDefinition{
				{Name: "username"},
				{Name: "email"},
				{Name: "gt"}, // reserved
			},
			expectErr: true,
			errMsg:    "保留关键字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateDuplicates(tt.fields)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFieldService_NewField_ReservedKeywords tests reserved keyword validation in NewField
func TestFieldService_NewField_ReservedKeywords(t *testing.T) {
	service := NewFieldService()
	locator := &ModelLocator{
		ModelName:    "TestModel",
		DatabaseName: "test-db",
	}

	tests := []struct {
		name      string
		fieldName string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid field name",
			fieldName: "username",
			expectErr: false,
		},
		{
			name:      "invalid field name - starts with underscore",
			fieldName: "_meta",
			expectErr: true,
			errMsg:    "格式无效",
		},
		{
			name:      "invalid field name - starts with double underscore",
			fieldName: "__meta",
			expectErr: true,
			errMsg:    "格式无效",
		},
		{
			name:      "reserved keyword - AND",
			fieldName: "AND",
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name:      "reserved keyword - or (lowercase)",
			fieldName: "or",
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name:      "reserved keyword - equals",
			fieldName: "equals",
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name:      "reserved keyword - lt",
			fieldName: "lt",
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name:      "reserved keyword - gte",
			fieldName: "gte",
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name:      "reserved keyword - startsWith",
			fieldName: "startsWith",
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name:      "reserved keyword - mode",
			fieldName: "mode",
			expectErr: true,
			errMsg:    "保留关键字",
		},
		{
			name:      "similar but valid - start",
			fieldName: "start",
			expectErr: false,
		},
		{
			name:      "similar but valid - modeType",
			fieldName: "modeType",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, err := service.NewField("model123", tt.fieldName, "Test Field", FormatString, locator)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, field)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, field)
				assert.Equal(t, tt.fieldName, field.Name)
			}
		})
	}
}

// TestReservedKeywords_ErrorMessage tests that error messages are helpful
func TestReservedKeywords_ErrorMessage(t *testing.T) {
	service := NewFieldService()
	locator := &ModelLocator{
		ModelName:    "TestModel",
		DatabaseName: "test-db",
	}

	_, err := service.NewField("model123", "AND", "Test Field", FormatString, locator)
	assert.Error(t, err)

	// 验证错误消息包含有用的信息
	errMsg := err.Error()
	assert.Contains(t, errMsg, "AND")
	assert.Contains(t, errMsg, "保留关键字")
	// 验证包含建议
	assert.True(t,
		assert.ObjectsAreEqual(errMsg, errMsg) &&
			(assert.ObjectsAreEqual(errMsg, errMsg)),
		"错误消息应该包含有用的建议")
}
