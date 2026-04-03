package repository_test

import (
	"database/sql"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProjectToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	clusterID := "cluster-123"

	t.Run("full row with all fields set", func(t *testing.T) {
		row := dbgen.Project{
			OrgName:     "my-org",
			Slug:        "my-project",
			Title:       "My Project",
			Description: sql.NullString{String: "desc", Valid: true},
			LoginUrl:    sql.NullString{String: "https://example.com", Valid: true},
			ClusterID:   sql.NullString{String: clusterID, Valid: true},
			Status:      "active",
			CreatedAt:   sql.NullTime{Time: now, Valid: true},
			UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		}

		p := repository.ProjectToDomain(row)

		assert.Equal(t, "my-org", p.OrgName)
		assert.Equal(t, "my-project", p.Slug)
		assert.Equal(t, "My Project", p.Title)
		assert.Equal(t, "desc", p.Description)
		assert.Equal(t, "https://example.com", p.LoginURL)
		assert.NotNil(t, p.ClusterID)
		assert.Equal(t, clusterID, *p.ClusterID)
		assert.Equal(t, project.ProjectStatusActive, p.Status)
		assert.Equal(t, now, p.CreatedAt)
		assert.Equal(t, now, p.UpdatedAt)
	})

	t.Run("nullable fields are NULL", func(t *testing.T) {
		row := dbgen.Project{
			OrgName: "my-org",
			Slug:    "my-project",
			Title:   "My Project",
			Status:  "active",
		}

		p := repository.ProjectToDomain(row)

		assert.Equal(t, "", p.Description)
		assert.Equal(t, "", p.LoginURL)
		assert.Nil(t, p.ClusterID)
	})
}

func TestProjectToRow(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	clusterID := "cluster-456"

	t.Run("full entity with cluster", func(t *testing.T) {
		p := &project.Project{
			OrgName:     "org",
			Slug:        "proj",
			Title:       "Proj",
			Description: "some desc",
			LoginURL:    "https://login.example.com",
			ClusterID:   &clusterID,
			Status:      project.ProjectStatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		params := repository.ProjectToCreateParams(p)

		assert.Equal(t, "org", params.OrgName)
		assert.Equal(t, "proj", params.Slug)
		assert.Equal(t, "some desc", params.Description.String)
		assert.True(t, params.Description.Valid)
		assert.Equal(t, "https://login.example.com", params.LoginUrl.String)
		assert.True(t, params.LoginUrl.Valid)
		assert.Equal(t, clusterID, params.ClusterID.String)
		assert.True(t, params.ClusterID.Valid)
	})

	t.Run("entity without cluster", func(t *testing.T) {
		p := &project.Project{
			OrgName: "org",
			Slug:    "proj",
			Title:   "Proj",
			Status:  project.ProjectStatusActive,
		}

		params := repository.ProjectToCreateParams(p)

		assert.False(t, params.Description.Valid)
		assert.False(t, params.LoginUrl.Valid)
		assert.False(t, params.ClusterID.Valid)
	})
}
