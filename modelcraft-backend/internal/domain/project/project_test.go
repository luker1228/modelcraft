package project

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewProject(t *testing.T) {
	tests := []struct {
		name        string
		projectSlug string
		title       string
		description string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid project",
			projectSlug: "ecommerce",
			title:       "E-Commerce Platform",
			description: "Online shopping platform",
			wantErr:     false,
		},
		{
			name:        "invalid project with hyphens",
			projectSlug: "my-project-123",
			title:       "My Project",
			description: "Test project",
			wantErr:     true,
			errContains: "lowercase letters/digits/underscores only",
		},
		{
			name:        "empty name",
			projectSlug: "",
			title:       "Test",
			description: "",
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name:        "empty title",
			projectSlug: "test",
			title:       "",
			description: "",
			wantErr:     true,
			errContains: "project title is required",
		},
		{
			name:        "name too short",
			projectSlug: "ab",
			title:       "Test",
			description: "",
			wantErr:     true,
			errContains: "3-64 characters",
		},
		{
			name:        "name starts with number",
			projectSlug: "123project",
			title:       "Test",
			description: "",
			wantErr:     true,
			errContains: "start with a letter",
		},
		{
			name:        "name with uppercase",
			projectSlug: "MyProject",
			title:       "Test",
			description: "",
			wantErr:     true,
			errContains: "lowercase",
		},
		{
			name:        "valid name with underscore",
			projectSlug: "my_project",
			title:       "Test",
			description: "",
			wantErr:     false,
		},
		{
			name:        "valid name with underscore and digits",
			projectSlug: "my_project_123",
			title:       "Test",
			description: "",
			wantErr:     false,
		},
		{
			name:        "name with special characters",
			projectSlug: "project@123",
			title:       "Test",
			description: "",
			wantErr:     true,
			errContains: "lowercase letters/digits/underscores only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := NewProject("built-in", tt.projectSlug, tt.title, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tt.projectSlug, project.Slug)
				assert.Equal(t, "built-in", project.OrgName)
				assert.Equal(t, tt.title, project.Title)
				assert.Equal(t, tt.description, project.Description)
				assert.Equal(t, ProjectStatusActive, project.Status)
				assert.False(t, project.CreatedAt.IsZero())
				assert.False(t, project.UpdatedAt.IsZero())
			}
		})
	}
}

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name        string
		project     *Project
		wantErr     bool
		errContains string
	}{
		{
			name: "valid project",
			project: &Project{
				OrgName:     "built-in",
				Slug:        "test_project",
				Title:       "Test Project",
				Description: "Description",
				Status:      ProjectStatusActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			project: &Project{
				OrgName:     "built-in",
				Slug:        "test_project",
				Title:       "Test Project",
				Description: "Description",
				Status:      "invalid",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:     true,
			errContains: "status MUST be either 'active' or 'archived'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()

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

func TestProject_UpdateMetadata(t *testing.T) {
	t.Run("update title and description", func(t *testing.T) {
		project := &Project{
			OrgName:     "built-in",
			Slug:        "test_project",
			Title:       "Original Title",
			Description: "Original Description",
			Status:      ProjectStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		oldUpdatedAt := project.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		err := project.UpdateMetadata("New Title", "New Description")
		assert.NoError(t, err)
		assert.Equal(t, "New Title", project.Title)
		assert.Equal(t, "New Description", project.Description)
		assert.True(t, project.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("update title only", func(t *testing.T) {
		project := &Project{
			OrgName:     "built-in",
			Slug:        "test_project",
			Title:       "Original Title",
			Description: "Original Description",
			Status:      ProjectStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		oldUpdatedAt := project.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		err := project.UpdateMetadata("New Title", "")
		assert.NoError(t, err)
		assert.Equal(t, "New Title", project.Title)
		assert.Equal(t, "Original Description", project.Description)
		assert.True(t, project.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("update description only", func(t *testing.T) {
		project := &Project{
			OrgName:     "built-in",
			Slug:        "test_project",
			Title:       "Original Title",
			Description: "Original Description",
			Status:      ProjectStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		oldUpdatedAt := project.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		err := project.UpdateMetadata("", "New Description")
		assert.NoError(t, err)
		assert.Equal(t, "Original Title", project.Title)
		assert.Equal(t, "New Description", project.Description)
		assert.True(t, project.UpdatedAt.After(oldUpdatedAt))
	})
}

func TestProject_Archive(t *testing.T) {
	project := &Project{
		OrgName:     "built-in",
		Slug:        "test_project",
		Title:       "Test",
		Description: "Test",
		Status:      ProjectStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.True(t, project.IsActive())

	oldUpdatedAt := project.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	project.Archive()
	assert.Equal(t, ProjectStatusArchived, project.Status)
	assert.False(t, project.IsActive())
	assert.True(t, project.UpdatedAt.After(oldUpdatedAt))
}

func TestProject_Activate(t *testing.T) {
	project := &Project{
		OrgName:     "built-in",
		Slug:        "test_project",
		Title:       "Test",
		Description: "Test",
		Status:      ProjectStatusArchived,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.False(t, project.IsActive())

	oldUpdatedAt := project.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	project.Activate()
	assert.Equal(t, ProjectStatusActive, project.Status)
	assert.True(t, project.IsActive())
	assert.True(t, project.UpdatedAt.After(oldUpdatedAt))
}

func TestIsValidProjectSlug(t *testing.T) {
	tests := []struct {
		name        string
		projectSlug string
		valid       bool
	}{
		{"valid lowercase", "project", true},
		{"valid with numbers", "project123", true},
		{"invalid with hyphens", "my-project-123", false},
		{"valid with underscore", "my_project", true},
		{"valid with underscore and digits", "my_project_123", true},
		{"valid with multiple underscores", "my_test_project", true},
		{"too short", "ab", false},
		{"too long", "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz012", false},
		{"starts with number", "1project", false},
		{"starts with hyphen", "-project", false},
		{"starts with underscore", "_project", false},
		{"uppercase", "Project", false},
		{"with spaces", "my project", false},
		{"with special chars", "project@123", false},
		{"minimum length", "abc", true},
		{"maximum length", "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxy1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidProjectSlug(tt.projectSlug)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestProject_SetCluster(t *testing.T) {
	t.Run("set valid cluster ID", func(t *testing.T) {
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		oldUpdatedAt := project.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		clusterID := "cluster-123"
		err := project.SetCluster(clusterID)
		assert.NoError(t, err)
		assert.NotNil(t, project.ClusterID)
		assert.Equal(t, clusterID, *project.ClusterID)
		assert.True(t, project.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("set empty cluster ID should fail", func(t *testing.T) {
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := project.SetCluster("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cluster ID cannot be empty")
		assert.Nil(t, project.ClusterID)
	})

	t.Run("update existing cluster ID", func(t *testing.T) {
		oldClusterID := "cluster-old"
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: &oldClusterID,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		newClusterID := "cluster-new"
		err := project.SetCluster(newClusterID)
		assert.NoError(t, err)
		assert.NotNil(t, project.ClusterID)
		assert.Equal(t, newClusterID, *project.ClusterID)
	})
}

func TestProject_UnsetCluster(t *testing.T) {
	clusterID := "cluster-123"
	project := &Project{
		OrgName:   "built-in",
		Slug:      "test_project",
		Title:     "Test Project",
		ClusterID: &clusterID,
		Status:    ProjectStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.True(t, project.HasCluster())

	oldUpdatedAt := project.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	project.UnsetCluster()
	assert.Nil(t, project.ClusterID)
	assert.False(t, project.HasCluster())
	assert.True(t, project.UpdatedAt.After(oldUpdatedAt))
}

func TestProject_GetClusterID(t *testing.T) {
	t.Run("get cluster ID when set", func(t *testing.T) {
		clusterID := "cluster-123"
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: &clusterID,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result := project.GetClusterID()
		assert.Equal(t, clusterID, result)
	})

	t.Run("get cluster ID when nil returns empty string", func(t *testing.T) {
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: nil,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result := project.GetClusterID()
		assert.Equal(t, "", result)
	})
}

func TestProject_HasCluster(t *testing.T) {
	t.Run("has cluster returns true when set", func(t *testing.T) {
		clusterID := "cluster-123"
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: &clusterID,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.True(t, project.HasCluster())
	})

	t.Run("has cluster returns false when nil", func(t *testing.T) {
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: nil,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.False(t, project.HasCluster())
	})

	t.Run("has cluster returns false when empty string", func(t *testing.T) {
		emptyClusterID := ""
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: &emptyClusterID,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.False(t, project.HasCluster())
	})
}

func TestProject_ValidateClusterID(t *testing.T) {
	t.Run("validate with nil cluster ID succeeds", func(t *testing.T) {
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: nil,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := project.Validate()
		assert.NoError(t, err)
	})

	t.Run("validate with valid cluster ID succeeds", func(t *testing.T) {
		clusterID := "cluster-123"
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: &clusterID,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := project.Validate()
		assert.NoError(t, err)
	})

	t.Run("validate with empty string cluster ID fails", func(t *testing.T) {
		emptyClusterID := ""
		project := &Project{
			OrgName:   "built-in",
			Slug:      "test_project",
			Title:     "Test Project",
			ClusterID: &emptyClusterID,
			Status:    ProjectStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := project.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cluster ID cannot be empty if provided")
	})
}
