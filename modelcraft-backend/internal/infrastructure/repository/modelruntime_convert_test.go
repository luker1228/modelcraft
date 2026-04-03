package repository_test

import (
	"database/sql"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDbgenModelToRuntimeModel verifies the dbgen.Model → RuntimeModel converter.
func TestDbgenModelToRuntimeModel(t *testing.T) {
	t.Run("basic fields mapped correctly", func(t *testing.T) {
		row := dbgen.Model{
			ID:           "model-1",
			OrgName:      "my-org",
			ProjectSlug:  "my-project",
			Name:         "users",
			Title:        "Users",
			Description:  sql.NullString{String: "user table", Valid: true},
			DatabaseName: "main_db",
		}

		m := repository.DbgenModelToRuntimeModel(row)

		assert.Equal(t, "my-org", m.OrgName)
		assert.Equal(t, "my-project", m.ProjectSlug)
		assert.Equal(t, "users", m.Name)
		assert.Equal(t, "Users", m.Title)
		assert.Equal(t, "user table", m.Description)
		assert.Equal(t, "main_db", m.DatabaseName)
	})

	t.Run("null description maps to empty string", func(t *testing.T) {
		row := dbgen.Model{
			ID:           "model-2",
			OrgName:      "org",
			ProjectSlug:  "proj",
			Name:         "orders",
			Title:        "Orders",
			DatabaseName: "db",
		}

		m := repository.DbgenModelToRuntimeModel(row)

		assert.Equal(t, "", m.Description)
	})
}
