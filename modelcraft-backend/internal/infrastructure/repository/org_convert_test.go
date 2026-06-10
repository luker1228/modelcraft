package repository_test

import (
	"database/sql"
	"modelcraft/internal/domain/organization"
	"modelcraft/internal/domain/user"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestOrganizationToDomain verifies that dbgen.Organization rows are correctly converted
// to domain Organization entities, covering all field mappings and nullable field handling.
func TestOrganizationToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	t.Run("all fields set", func(t *testing.T) {
		row := dbgen.Organization{
			Name:        "my-org",
			DisplayName: sql.NullString{String: "My Organization", Valid: true},
			OwnerID:     sql.NullString{String: "user-1", Valid: true},
			Status:      "active",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		entity := repository.OrganizationToDomain(row)

		assert.Equal(t, "my-org", entity.Name)
		assert.Equal(t, "My Organization", entity.DisplayName)
		assert.Equal(t, "user-1", entity.OwnerID)
		assert.Equal(t, organization.OrgStatusActive, entity.Status)
		assert.Equal(t, now, entity.CreatedAt)
		assert.Equal(t, now, entity.UpdatedAt)
	})

	t.Run("nullable fields NULL", func(t *testing.T) {
		row := dbgen.Organization{
			Name:      "bare-org",
			Status:    "suspended",
			CreatedAt: now,
			UpdatedAt: now,
		}

		entity := repository.OrganizationToDomain(row)

		assert.Equal(t, "", entity.DisplayName)
		assert.Equal(t, "", entity.OwnerID)
		assert.Equal(t, organization.OrgStatusSuspended, entity.Status)
	})
}

// TestUserToDomain verifies that dbgen.User rows are correctly converted to domain User entities.
func TestUserToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	t.Run("all fields set", func(t *testing.T) {
		row := dbgen.User{
			ID:           "user-1",
			ExternalID:   sql.NullString{String: "ext-abc", Valid: true},
			Name:         "Alice",
			Phone:        "13800001111",
			PasswordHash: "$2a$10$hash",
			DisplayName:  sql.NullString{String: "Alice Wonderland", Valid: true},
			IsAdmin:      true,
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		entity := repository.UserToDomain(row)

		assert.Equal(t, "user-1", entity.ID)
		assert.Equal(t, "ext-abc", entity.ExternalID)
		assert.Equal(t, "Alice", entity.Name)
		assert.Equal(t, "13800001111", entity.Phone.String())
		assert.Equal(t, "$2a$10$hash", entity.PasswordHash)
		assert.True(t, entity.IsAdmin)
		assert.Equal(t, "active", entity.Status)
		assert.Equal(t, now, entity.CreatedAt)
		assert.Equal(t, now, entity.UpdatedAt)
	})

	t.Run("DisplayName NULL is ignored by domain User", func(t *testing.T) {
		// domain user.User has no DisplayName field; the converter should not error.
		row := dbgen.User{
			ID:         "user-2",
			ExternalID: sql.NullString{String: "ext-xyz", Valid: true},
			Name:       "Bob",
			Phone:      "",
			IsAdmin:    false,
			Status:     "active",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		entity := repository.UserToDomain(row)

		assert.IsType(t, &user.User{}, entity)
		assert.Equal(t, "user-2", entity.ID)
		assert.Equal(t, "Bob", entity.Name)
		assert.False(t, entity.IsAdmin)
	})
}
