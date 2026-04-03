package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectScope_Validate(t *testing.T) {
	tests := []struct {
		name        string
		scope       ProjectScope
		wantErr     bool
		errContains string
	}{
		{
			name: "valid scope",
			scope: ProjectScope{
				OrgName:     "acme",
				ProjectSlug: "ecommerce",
			},
			wantErr: false,
		},
		{
			name: "missing org name",
			scope: ProjectScope{
				OrgName:     "",
				ProjectSlug: "ecommerce",
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name: "missing project slug",
			scope: ProjectScope{
				OrgName:     "acme",
				ProjectSlug: "",
			},
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name:        "both fields missing",
			scope:       ProjectScope{},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scope.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewProjectScope(t *testing.T) {
	tests := []struct {
		name        string
		orgName     string
		projectSlug string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid scope",
			orgName:     "acme",
			projectSlug: "ecommerce",
			wantErr:     false,
		},
		{
			name:        "missing org name",
			orgName:     "",
			projectSlug: "ecommerce",
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name:        "missing project slug",
			orgName:     "acme",
			projectSlug: "",
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name:        "both fields missing",
			orgName:     "",
			projectSlug: "",
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, err := NewProjectScope(tt.orgName, tt.projectSlug)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.orgName, scope.OrgName)
				assert.Equal(t, tt.projectSlug, scope.ProjectSlug)
			}
		})
	}
}

func TestProjectScope_GetFullPath(t *testing.T) {
	tests := []struct {
		name     string
		scope    ProjectScope
		expected string
	}{
		{
			name: "standard path",
			scope: ProjectScope{
				OrgName:     "acme",
				ProjectSlug: "ecommerce",
			},
			expected: "acme.ecommerce",
		},
		{
			name: "with underscores",
			scope: ProjectScope{
				OrgName:     "my_org",
				ProjectSlug: "my_project",
			},
			expected: "my_org.my_project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scope.GetFullPath()
			assert.Equal(t, tt.expected, result)
		})
	}
}
