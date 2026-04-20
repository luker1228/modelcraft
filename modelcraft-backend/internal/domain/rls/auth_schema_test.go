package rls

import "testing"

func TestAuthSchemaGetVariable(t *testing.T) {
	schema := &AuthSchema{
		ProjectID: "test-project",
		Variables: []AuthVariable{
			{Name: "tenant_id", Source: "jwt.tenant_id", Type: AuthVarTypeUUID},
			{Name: "role", Source: "jwt.role", Type: AuthVarTypeString},
		},
	}

	tests := []struct {
		name     string
		varName  string
		expected *AuthVariable
	}{
		{"uid builtin", "uid", &AuthVariable{Name: "uid", Source: "jwt.user_id", Type: AuthVarTypeUUID}},
		{"tenant_id", "tenant_id", &AuthVariable{Name: "tenant_id", Source: "jwt.tenant_id", Type: AuthVarTypeUUID}},
		{"role", "role", &AuthVariable{Name: "role", Source: "jwt.role", Type: AuthVarTypeString}},
		{"unknown", "unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.GetVariable(tt.varName)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetVariable() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("GetVariable() = nil, want %v", tt.expected)
				return
			}
			if result.Name != tt.expected.Name || result.Source != tt.expected.Source {
				t.Errorf("GetVariable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAuthSchemaIsValidRef(t *testing.T) {
	schema := &AuthSchema{
		ProjectID: "test-project",
		Variables: []AuthVariable{
			{Name: "tenant_id", Source: "jwt.tenant_id", Type: AuthVarTypeUUID},
		},
	}

	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"uid builtin", "uid", true},
		{"tenant_id", "tenant_id", true},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schema.IsValidRef(tt.varName)
			if result != tt.expected {
				t.Errorf("IsValidRef() = %v, want %v", result, tt.expected)
			}
		})
	}
}
