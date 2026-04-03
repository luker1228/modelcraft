package repository_test

import (
	"database/sql"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- ModelGroupToDomain ---

func TestModelGroupToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	t.Run("all fields set", func(t *testing.T) {
		row := dbgen.ModelGroup{
			ID:           "group-1",
			OrgName:      "my-org",
			ProjectSlug:  "my-project",
			Name:         "payment",
			DisplayOrder: "N",
			CreatedAt:    sql.NullTime{Time: now, Valid: true},
			UpdatedAt:    sql.NullTime{Time: now, Valid: true},
		}

		g := repository.ModelGroupToDomain(row)

		assert.Equal(t, "group-1", g.ID)
		assert.Equal(t, "my-org", g.OrgName)
		assert.Equal(t, "my-project", g.ProjectSlug)
		assert.Equal(t, "payment", g.Name)
		assert.Equal(t, "N", g.DisplayOrder)
		assert.Equal(t, now, g.CreatedAt)
		assert.Equal(t, now, g.UpdatedAt)
	})

	t.Run("nullable times are zero", func(t *testing.T) {
		row := dbgen.ModelGroup{
			ID:          "group-2",
			OrgName:     "org",
			ProjectSlug: "proj",
			Name:        "core",
		}

		g := repository.ModelGroupToDomain(row)

		assert.True(t, g.CreatedAt.IsZero())
		assert.True(t, g.UpdatedAt.IsZero())
	})
}

// --- ModelGroupToCreateParams ---

func TestModelGroupToCreateParams(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	group := &modeldesign.ModelGroup{
		ID:           "group-1",
		ProjectScope: project.ProjectScope{OrgName: "org-1", ProjectSlug: "proj-1"},
		Name:         "payment",
		DisplayOrder: "N",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	p := repository.ModelGroupToCreateParams(group)

	assert.Equal(t, "group-1", p.ID)
	assert.Equal(t, "org-1", p.OrgName)
	assert.Equal(t, "proj-1", p.ProjectSlug)
	assert.Equal(t, "payment", p.Name)
	assert.Equal(t, "N", p.DisplayOrder)
}
