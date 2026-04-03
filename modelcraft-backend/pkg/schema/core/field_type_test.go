package core

import (
	"testing"

	"github.com/graphql-go/graphql"
)

func TestFieldType_String(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  string
	}{
		{"string type", FieldTypeString, "string"},
		{"integer type", FieldTypeInteger, "integer"},
		{"float type", FieldTypeFloat, "float"},
		{"boolean type", FieldTypeBoolean, "boolean"},
		{"id type", FieldTypeID, "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldType.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFieldType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  bool
	}{
		{"valid string", FieldTypeString, true},
		{"valid integer", FieldTypeInteger, true},
		{"valid float", FieldTypeFloat, true},
		{"valid boolean", FieldTypeBoolean, true},
		{"valid id", FieldTypeID, true},
		{"invalid type", FieldType("invalid"), false},
		{"empty type", FieldType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldType.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFieldType_GetJSONType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  string
	}{
		{"string to string", FieldTypeString, "string"},
		{"integer to integer", FieldTypeInteger, "integer"},
		{"float to number", FieldTypeFloat, "number"},
		{"boolean to boolean", FieldTypeBoolean, "boolean"},
		{"id to string", FieldTypeID, "string"},
		{"invalid to string", FieldType("invalid"), "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldType.GetJSONType()
			if result != tt.expected {
				t.Errorf("GetJSONType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFieldType_GetGraphQLType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  graphql.Type
	}{
		{"string to String", FieldTypeString, graphql.String},
		{"integer to Int", FieldTypeInteger, graphql.Int},
		{"float to Float", FieldTypeFloat, graphql.Float},
		{"boolean to Boolean", FieldTypeBoolean, graphql.Boolean},
		{"id to ID", FieldTypeID, graphql.ID},
		{"invalid to String", FieldType("invalid"), graphql.String},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldType.GetGraphQLType()
			if result != tt.expected {
				t.Errorf("GetGraphQLType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFieldType_GetSDLType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  string
	}{
		{"string to String", FieldTypeString, "String"},
		{"integer to Int", FieldTypeInteger, "Int"},
		{"float to Float", FieldTypeFloat, "Float"},
		{"boolean to Boolean", FieldTypeBoolean, "Boolean"},
		{"id to ID", FieldTypeID, "ID"},
		{"invalid to String", FieldType("invalid"), "String"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldType.GetSDLType()
			if result != tt.expected {
				t.Errorf("GetSDLType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFieldType_GetDefaultValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  interface{}
	}{
		{"string default", FieldTypeString, ""},
		{"integer default", FieldTypeInteger, 0},
		{"float default", FieldTypeFloat, 0.0},
		{"boolean default", FieldTypeBoolean, false},
		{"id default", FieldTypeID, ""},
		{"invalid default", FieldType("invalid"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldType.GetDefaultValue()
			if result != tt.expected {
				t.Errorf("GetDefaultValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAllFieldTypes(t *testing.T) {
	types := AllFieldTypes()
	expected := []FieldType{
		FieldTypeString,
		FieldTypeInteger,
		FieldTypeFloat,
		FieldTypeBoolean,
		FieldTypeID,
	}

	if len(types) != len(expected) {
		t.Errorf("AllFieldTypes() length = %v, want %v", len(types), len(expected))
		return
	}

	for i, fieldType := range types {
		if fieldType != expected[i] {
			t.Errorf("AllFieldTypes()[%d] = %v, want %v", i, fieldType, expected[i])
		}
	}
}

func TestParseFieldType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  FieldType
		shouldErr bool
	}{
		{"valid string", "string", FieldTypeString, false},
		{"valid integer", "integer", FieldTypeInteger, false},
		{"valid float", "float", FieldTypeFloat, false},
		{"valid boolean", "boolean", FieldTypeBoolean, false},
		{"valid id", "id", FieldTypeID, false},
		{"invalid type", "invalid", "", true},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFieldType(tt.input)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("ParseFieldType() should return error for input %v", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFieldType() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseFieldType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMustParseFieldType(t *testing.T) {
	// 测试成功情况
	result := MustParseFieldType("string")
	if result != FieldTypeString {
		t.Errorf("MustParseFieldType() = %v, want %v", result, FieldTypeString)
	}

	// 测试panic情况
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseFieldType() should panic for invalid input")
		}
	}()

	MustParseFieldType("invalid")
}

func TestInvalidFieldTypeError(t *testing.T) {
	err := &InvalidFieldTypeError{Type: "invalid"}
	expected := "invalid field type: invalid"

	if err.Error() != expected {
		t.Errorf("InvalidFieldTypeError.Error() = %v, want %v", err.Error(), expected)
	}
}
