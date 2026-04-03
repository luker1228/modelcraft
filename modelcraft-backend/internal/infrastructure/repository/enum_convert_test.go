package repository_test

import (
	"database/sql"
	"encoding/json"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- EnumDefinitionToDomain ---

func TestEnumDefinitionToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	options := []modeldesign.EnumOption{
		{Code: "A", Label: "Option A", Order: 1},
		{Code: "B", Label: "Option B", Order: 2},
	}
	optionsJSON, _ := json.Marshal(options)

	t.Run("all fields set", func(t *testing.T) {
		row := dbgen.ModelEnum{
			ID:            "enum-1",
			OrgName:       "my-org",
			ProjectSlug:   "my-project",
			Name:          "status",
			DisplayName:   "Status",
			Description:   sql.NullString{String: "status enum", Valid: true},
			Options:       json.RawMessage(optionsJSON),
			IsMultiSelect: sql.NullBool{Bool: true, Valid: true},
			CreatedAt:     sql.NullTime{Time: now, Valid: true},
			UpdatedAt:     sql.NullTime{Time: now, Valid: true},
		}

		ed, err := repository.EnumDefinitionToDomain(row)
		require.NoError(t, err)

		assert.Equal(t, "enum-1", ed.ID)
		assert.Equal(t, "my-org", ed.OrgName)
		assert.Equal(t, "my-project", ed.ProjectSlug)
		assert.Equal(t, "status", ed.Name)
		assert.Equal(t, "Status", ed.DisplayName)
		assert.Equal(t, "status enum", ed.Description)
		assert.True(t, ed.IsMultiSelect)
		require.Len(t, ed.Options, 2)
		assert.Equal(t, "A", ed.Options[0].Code)
		assert.Equal(t, "B", ed.Options[1].Code)
		assert.Equal(t, now, ed.CreatedAt)
		assert.Equal(t, now, ed.UpdatedAt)
	})

	t.Run("nullable fields are NULL", func(t *testing.T) {
		row := dbgen.ModelEnum{
			ID:          "enum-2",
			OrgName:     "org",
			ProjectSlug: "proj",
			Name:        "category",
			DisplayName: "Category",
			Options:     json.RawMessage(optionsJSON),
		}

		ed, err := repository.EnumDefinitionToDomain(row)
		require.NoError(t, err)

		assert.Equal(t, "", ed.Description)
		assert.False(t, ed.IsMultiSelect)
	})

	t.Run("invalid options JSON returns error", func(t *testing.T) {
		row := dbgen.ModelEnum{
			ID:      "enum-bad",
			Name:    "bad",
			Options: json.RawMessage(`not valid json`),
		}

		_, err := repository.EnumDefinitionToDomain(row)
		assert.Error(t, err)
	})
}

// --- EnumDefinitionToCreateParams ---

func TestEnumDefinitionToCreateParams(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	options := []modeldesign.EnumOption{
		{Code: "Y", Label: "Yes", Order: 1},
		{Code: "N", Label: "No", Order: 2},
	}

	t.Run("full enum definition", func(t *testing.T) {
		ed := &modeldesign.EnumDefinition{
			ID:            "enum-1",
			ProjectScope:  project.ProjectScope{OrgName: "org-1", ProjectSlug: "proj-1"},
			Name:          "confirm",
			DisplayName:   "Confirmation",
			Description:   "yes or no",
			Options:       options,
			IsMultiSelect: false,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		p, err := repository.EnumDefinitionToCreateParams(ed)
		require.NoError(t, err)

		assert.Equal(t, "enum-1", p.ID)
		assert.Equal(t, "org-1", p.OrgName)
		assert.Equal(t, "proj-1", p.ProjectSlug)
		assert.Equal(t, "confirm", p.Name)
		assert.Equal(t, "Confirmation", p.DisplayName)
		assert.True(t, p.Description.Valid)
		assert.Equal(t, "yes or no", p.Description.String)
		assert.False(t, p.IsMultiSelect.Bool)

		var parsedOptions []modeldesign.EnumOption
		require.NoError(t, json.Unmarshal(p.Options, &parsedOptions))
		require.Len(t, parsedOptions, 2)
		assert.Equal(t, "Y", parsedOptions[0].Code)
	})

	t.Run("empty description", func(t *testing.T) {
		ed := &modeldesign.EnumDefinition{
			ID:           "enum-2",
			ProjectScope: project.ProjectScope{OrgName: "org", ProjectSlug: "proj"},
			Name:         "type",
			DisplayName:  "Type",
			Options:      options,
		}

		p, err := repository.EnumDefinitionToCreateParams(ed)
		require.NoError(t, err)

		assert.False(t, p.Description.Valid)
	})
}

// --- FieldEnumAssociationToDomain ---

func TestFieldEnumAssociationToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	t.Run("all fields set", func(t *testing.T) {
		row := dbgen.ModelFieldEnumAssociation{
			ModelID:      "model-1",
			FieldName:    "status",
			OrgName:      "my-org",
			ProjectSlug:  "proj",
			EnumName:     "status_enum",
			DatabaseName: "main_db",
			CreatedAt:    sql.NullTime{Time: now, Valid: true},
			UpdatedAt:    sql.NullTime{Time: now, Valid: true},
		}

		assoc := repository.FieldEnumAssociationToDomain(row)

		assert.Equal(t, "model-1", assoc.ModelID)
		assert.Equal(t, "status", assoc.FieldName)
		assert.Equal(t, "my-org", assoc.OrgName)
		assert.Equal(t, "proj", assoc.ProjectSlug)
		assert.Equal(t, "status_enum", assoc.EnumName)
		assert.Equal(t, "main_db", assoc.DatabaseName)
		assert.Equal(t, now, assoc.CreatedAt)
		assert.Equal(t, now, assoc.UpdatedAt)
	})

	t.Run("nullable times are zero", func(t *testing.T) {
		row := dbgen.ModelFieldEnumAssociation{
			ModelID:      "model-2",
			FieldName:    "category",
			ProjectSlug:  "proj",
			EnumName:     "cat_enum",
			DatabaseName: "db",
		}

		assoc := repository.FieldEnumAssociationToDomain(row)
		assert.True(t, assoc.CreatedAt.IsZero())
		assert.True(t, assoc.UpdatedAt.IsZero())
	})
}

// --- FieldEnumAssociationToCreateParams ---

func TestFieldEnumAssociationToCreateParams(t *testing.T) {
	assoc := &modeldesign.FieldEnumAssociation{
		ModelID:      "model-1",
		FieldName:    "status",
		ProjectScope: project.ProjectScope{OrgName: "org-1", ProjectSlug: "proj-1"},
		EnumName:     "status_enum",
		DatabaseName: "main_db",
	}

	p := repository.FieldEnumAssociationToCreateParams(assoc)

	assert.Equal(t, "model-1", p.ModelID)
	assert.Equal(t, "status", p.FieldName)
	assert.Equal(t, "proj-1", p.ProjectSlug)
	assert.Equal(t, "status_enum", p.EnumName)
	assert.Equal(t, "main_db", p.DatabaseName)
}
