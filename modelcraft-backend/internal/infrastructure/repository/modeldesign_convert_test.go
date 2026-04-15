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

// --- ModelToDomain ---

func TestModelToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	groupID := "group-1"

	t.Run("all fields set", func(t *testing.T) {
		row := dbgen.Model{
			ID:               "model-id",
			OrgName:          "my-org",
			ProjectSlug:      "my-project",
			Name:             "users",
			Title:            "Users",
			Description:      sql.NullString{String: "user table", Valid: true},
			StorageType:      "table",
			DatabaseName:     "main_db",
			Version:          sql.NullInt64{Int64: 3, Valid: true},
			Status:           sql.NullString{String: "published", Valid: true},
			GroupID:          sql.NullString{String: groupID, Valid: true},
			DeploymentStatus: sql.NullString{String: "success", Valid: true},
			LastSyncAt:       sql.NullTime{Time: now, Valid: true},
			SyncError:        sql.NullString{String: "", Valid: false},
			CreatedAt:        sql.NullTime{Time: now, Valid: true},
			UpdatedAt:        sql.NullTime{Time: now, Valid: true},
		}

		m := repository.ModelToDomain(row)

		assert.Equal(t, "model-id", m.ID)
		assert.Equal(t, "my-project", m.ProjectSlug)
		assert.Equal(t, "users", m.ModelName)
		assert.Equal(t, "Users", m.Title)
		assert.Equal(t, "user table", m.Description)
		assert.Equal(t, "table", m.StorageType)
		assert.Equal(t, "main_db", m.DatabaseName)
		assert.Equal(t, int64(3), m.Version)
		assert.Equal(t, "published", m.Status)
		require.NotNil(t, m.GroupID)
		assert.Equal(t, groupID, *m.GroupID)
		assert.Equal(t, modeldesign.DeploymentStatus("success"), m.DeploymentStatus)
		require.NotNil(t, m.LastSyncAt)
		assert.Equal(t, now, *m.LastSyncAt)
		assert.Equal(t, "", m.SyncError)
		assert.Equal(t, now, m.CreatedAt)
		assert.Equal(t, now, m.UpdatedAt)
	})

	t.Run("nullable fields are NULL", func(t *testing.T) {
		row := dbgen.Model{
			ID:           "model-id",
			OrgName:      "my-org",
			ProjectSlug:  "my-project",
			Name:         "orders",
			Title:        "Orders",
			StorageType:  "table",
			DatabaseName: "db",
		}

		m := repository.ModelToDomain(row)

		assert.Equal(t, "", m.Description)
		assert.Equal(t, int64(0), m.Version)
		assert.Equal(t, "", m.Status)
		assert.Nil(t, m.GroupID)
		assert.Equal(t, modeldesign.DeploymentStatus(""), m.DeploymentStatus)
		assert.Nil(t, m.LastSyncAt)
		assert.Equal(t, "", m.SyncError)
	})
}

// --- ModelToCreateParams ---

func TestModelToCreateParams(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	groupID := "group-42"
	lastSync := now

	t.Run("full model", func(t *testing.T) {
		m := &modeldesign.DataModel{
			ModelMeta: modeldesign.ModelMeta{
				ID: "mid",
				ModelLocator: modeldesign.ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "org-1", ProjectSlug: "proj"},
					ModelName:    "users",
					DatabaseName: "db",
				},
				Title:            "Users",
				Description:      "desc",
				StorageType:      "table",
				Version:          2,
				Status:           "draft",
				GroupID:          &groupID,
				DeploymentStatus: modeldesign.DeploymentStatus("pending"),
				LastSyncAt:       &lastSync,
				SyncError:        "some error",
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		}
		orgName := "org-1"

		p := repository.ModelToCreateParams(m, orgName)

		assert.Equal(t, "mid", p.ID)
		assert.Equal(t, "org-1", p.OrgName)
		assert.Equal(t, "proj", p.ProjectSlug)
		assert.Equal(t, "users", p.Name)
		assert.Equal(t, "Users", p.Title)
		assert.True(t, p.Description.Valid)
		assert.Equal(t, "desc", p.Description.String)
		assert.Equal(t, "table", p.StorageType)
		assert.Equal(t, "db", p.DatabaseName)
		assert.True(t, p.Version.Valid)
		assert.Equal(t, int64(2), p.Version.Int64)
		assert.True(t, p.Status.Valid)
		assert.Equal(t, "draft", p.Status.String)
		assert.True(t, p.GroupID.Valid)
		assert.Equal(t, groupID, p.GroupID.String)
		assert.True(t, p.DeploymentStatus.Valid)
		assert.Equal(t, "pending", p.DeploymentStatus.String)
		assert.True(t, p.LastSyncAt.Valid)
		assert.Equal(t, lastSync, p.LastSyncAt.Time)
		assert.True(t, p.SyncError.Valid)
		assert.Equal(t, "some error", p.SyncError.String)
	})

	t.Run("minimal model, nullable fields empty", func(t *testing.T) {
		m := &modeldesign.DataModel{
			ModelMeta: modeldesign.ModelMeta{
				ID: "mid2",
				ModelLocator: modeldesign.ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "org", ProjectSlug: "proj"},
					ModelName:    "items",
					DatabaseName: "db",
				},
				Title:       "Items",
				StorageType: "table",
			},
		}

		p := repository.ModelToCreateParams(m, "org")

		assert.False(t, p.Description.Valid)
		assert.False(t, p.Version.Valid)
		assert.False(t, p.Status.Valid)
		assert.False(t, p.GroupID.Valid)
		assert.False(t, p.DeploymentStatus.Valid)
		assert.False(t, p.LastSyncAt.Valid)
		assert.False(t, p.SyncError.Valid)
	})
}

// --- FieldDefinitionToDomain ---

func TestFieldDefinitionToDomain(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	t.Run("string field, no JSON fields", func(t *testing.T) {
		row := dbgen.FieldDefinition{
			ModelID:      "model-1",
			Name:         "username",
			ModelName:    "users",
			DatabaseName: "db",
			Title:        "Username",
			Description:  sql.NullString{String: "the username", Valid: true},
			Format:       "STRING",
			NonNull:      sql.NullBool{Bool: true, Valid: true},
			Required:     sql.NullBool{Bool: true, Valid: true},
			IsUnique:     sql.NullBool{Bool: false, Valid: true},
			IsPrimary:    sql.NullBool{Bool: false, Valid: true},
			Status:       "active",
			DisplayOrder: "a0",
			CreatedAt:    sql.NullTime{Time: now, Valid: true},
			UpdatedAt:    sql.NullTime{Time: now, Valid: true},
		}

		fd, err := repository.FieldDefinitionToDomain(row)
		require.NoError(t, err)

		assert.Equal(t, "model-1", fd.ModelID)
		assert.Equal(t, "username", fd.Name)
		assert.Equal(t, "users", fd.ModelLocator.ModelName)
		assert.Equal(t, "db", fd.ModelLocator.DatabaseName)
		assert.Equal(t, "Username", fd.Title)
		assert.Equal(t, "the username", fd.Description)
		require.NotNil(t, fd.Type)
		assert.Equal(t, modeldesign.FormatString, fd.Type.Format)
		assert.True(t, fd.NonNull)
		assert.True(t, fd.Required)
		assert.False(t, fd.IsUnique)
		assert.False(t, fd.IsPrimary)
		assert.Equal(t, modeldesign.StatusType("active"), fd.Status)
		assert.Equal(t, "a0", fd.DisplayOrder)
		assert.Equal(t, now, fd.CreatedAt)
		assert.Equal(t, now, fd.UpdatedAt)
		assert.Nil(t, fd.Validation)
		assert.Nil(t, fd.Metadata)
	})

	t.Run("field with validation JSON", func(t *testing.T) {
		validJSON, _ := json.Marshal(modeldesign.ValidationConfig{MaxLength: ptrInt(100)})
		v := json.RawMessage(validJSON)
		row := dbgen.FieldDefinition{
			ModelID:      "model-1",
			Name:         "email",
			ModelName:    "users",
			DatabaseName: "db",
			Title:        "Email",
			Format:       "STRING",
			Status:       "active",
			Validation:   &v,
		}

		fd, err := repository.FieldDefinitionToDomain(row)
		require.NoError(t, err)
		require.NotNil(t, fd.Validation)
		require.NotNil(t, fd.Validation.MaxLength)
		assert.Equal(t, 100, *fd.Validation.MaxLength)
	})

	t.Run("field with metadata JSON", func(t *testing.T) {
		metaJSON := []byte(`{"key":"value"}`)
		m := json.RawMessage(metaJSON)
		row := dbgen.FieldDefinition{
			ModelID:      "model-1",
			Name:         "notes",
			ModelName:    "users",
			DatabaseName: "db",
			Title:        "Notes",
			Format:       "STRING",
			Status:       "active",
			Metadata:     &m,
		}

		fd, err := repository.FieldDefinitionToDomain(row)
		require.NoError(t, err)
		require.NotNil(t, fd.Metadata)
		assert.Equal(t, "value", fd.Metadata["key"])
	})

	t.Run("field with enumDisplay in metadata", func(t *testing.T) {
		metaJSON := []byte(`{"key":"value","enumDisplay":{"enabled":true,"labelFieldName":"status_text"}}`)
		m := json.RawMessage(metaJSON)
		row := dbgen.FieldDefinition{
			ModelID:      "model-1",
			Name:         "status",
			ModelName:    "users",
			DatabaseName: "db",
			Title:        "Status",
			Format:       "ENUM",
			Status:       "active",
			Metadata:     &m,
		}

		fd, err := repository.FieldDefinitionToDomain(row)
		require.NoError(t, err)
		require.NotNil(t, fd.Metadata)
		assert.Equal(t, "value", fd.Metadata["key"])
		enumDisplay, ok := fd.Metadata["enumDisplay"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "status_text", enumDisplay["labelFieldName"])
	})

	t.Run("field with enum relation id for ENUM_LABEL", func(t *testing.T) {
		relationID := "relation-123"
		row := dbgen.FieldDefinition{
			ModelID:        "model-1",
			Name:           "status_label",
			ModelName:      "users",
			DatabaseName:   "db",
			Title:          "Status Label",
			Format:         "ENUM_LABEL",
			Status:         "active",
			EnumRelationID: sql.NullString{String: relationID, Valid: true},
		}

		fd, err := repository.FieldDefinitionToDomain(row)
		require.NoError(t, err)
		require.NotNil(t, fd.EnumRelationID)
		assert.Equal(t, relationID, *fd.EnumRelationID)
	})

	t.Run("invalid format returns error", func(t *testing.T) {
		row := dbgen.FieldDefinition{
			ModelID:      "model-1",
			Name:         "bad",
			ModelName:    "users",
			DatabaseName: "db",
			Title:        "Bad",
			Format:       "INVALID_FORMAT",
			Status:       "active",
		}

		_, err := repository.FieldDefinitionToDomain(row)
		assert.Error(t, err)
	})
}

// --- FieldDefinitionToCreateParams ---

func TestFieldDefinitionToCreateParams(t *testing.T) {
	t.Run("full field with validation", func(t *testing.T) {
		maxLen := 255
		fd := &modeldesign.FieldDefinition{
			ModelID: "model-1",
			ModelLocator: &modeldesign.ModelLocator{
				ModelName:    "users",
				DatabaseName: "db",
			},
			Name:         "email",
			EnumName:     "status_enum",
			Title:        "Email",
			Description:  "user email",
			Type:         &modeldesign.FieldType{Format: modeldesign.FormatString},
			NonNull:      true,
			Required:     true,
			IsUnique:     true,
			IsPrimary:    false,
			Status:       modeldesign.FieldStatusDeploySuccess,
			Validation:   &modeldesign.ValidationConfig{MaxLength: &maxLen},
			DisplayOrder: "a1",
			Metadata: map[string]any{
				"hint": "enter email",
				"enumDisplay": map[string]any{
					"labelFieldName": "status_text",
				},
			},
		}

		p, err := repository.FieldDefinitionToCreateParams(fd, "test-org")
		require.NoError(t, err)

		assert.Equal(t, "model-1", p.ModelID)
		assert.Equal(t, "test-org", p.OrgName)
		assert.Equal(t, "users", p.ModelName)
		assert.Equal(t, "db", p.DatabaseName)
		assert.Equal(t, "email", p.Name)
		assert.True(t, p.EnumName.Valid)
		assert.Equal(t, "status_enum", p.EnumName.String)
		assert.Equal(t, "Email", p.Title)
		assert.True(t, p.Description.Valid)
		assert.Equal(t, "user email", p.Description.String)
		assert.Equal(t, "STRING", p.Format)
		assert.True(t, p.NonNull.Valid && p.NonNull.Bool)
		assert.True(t, p.Required.Valid && p.Required.Bool)
		assert.True(t, p.IsUnique.Valid && p.IsUnique.Bool)
		assert.False(t, p.IsPrimary.Bool)
		assert.Equal(t, "deploy_success", p.Status)
		assert.Equal(t, "a1", p.DisplayOrder)
		// Validation and Metadata are JSON-marshaled
		assert.NotEmpty(t, p.Validation)
		assert.NotEmpty(t, p.Metadata)
		require.NotNil(t, p.Metadata)

		var meta map[string]any
		require.NoError(t, json.Unmarshal(*p.Metadata, &meta))
		assert.Equal(t, "enter email", meta["hint"])
		enumDisplayRaw, ok := meta["enumDisplay"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "status_text", enumDisplayRaw["labelFieldName"])
	})

	t.Run("field without optional fields", func(t *testing.T) {
		fd := &modeldesign.FieldDefinition{
			ModelID: "model-1",
			ModelLocator: &modeldesign.ModelLocator{
				ModelName:    "items",
				DatabaseName: "db",
			},
			Name:   "name",
			Title:  "Name",
			Type:   &modeldesign.FieldType{Format: modeldesign.FormatString},
			Status: modeldesign.FieldStatusInit,
		}

		p, err := repository.FieldDefinitionToCreateParams(fd, "test-org")
		require.NoError(t, err)

		assert.False(t, p.EnumName.Valid)
		assert.False(t, p.Description.Valid)
		// NonNull/Required/IsUnique/IsPrimary are always valid (persisted as bool)
		assert.True(t, p.NonNull.Valid)
		assert.True(t, p.Required.Valid)
		// Nil JSON fields produce nil (not "null" bytes)
		assert.Nil(t, p.Validation)
		assert.Nil(t, p.Metadata)
		assert.Nil(t, p.Metadata)
	})
}

// helper

func ptrInt(v int) *int { return &v }
