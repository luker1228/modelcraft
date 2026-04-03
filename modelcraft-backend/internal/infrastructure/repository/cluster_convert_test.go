package repository_test

import (
	"database/sql"
	"modelcraft/internal/domain/cluster"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/crypto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupClusterTestCipher initializes the default AES cipher required by cluster.NewByEncrypt.
// It must be called at the top of any test that exercises password conversion.
func setupClusterTestCipher(t *testing.T) {
	t.Helper()
	require.NoError(t, crypto.InitDefaultAESCipher("12345678901234567890123456789012"))
}

// TestDatabaseClusterToDomain verifies that dbgen.DatabaseCluster rows are correctly converted
// to domain DatabaseCluster entities, covering all field mappings and nullable field handling.
func TestDatabaseClusterToDomain(t *testing.T) {
	setupClusterTestCipher(t)

	now := time.Now().Truncate(time.Millisecond)

	t.Run("all fields set", func(t *testing.T) {
		row := dbgen.DatabaseCluster{
			ID:                "cluster-1",
			OrgName:           "my-org",
			ProjectSlug:       "my-project",
			Title:             "Production DB",
			Description:       sql.NullString{String: "main cluster", Valid: true},
			Host:              "db.example.com",
			Port:              3306,
			Username:          "admin",
			Password:          "encrypted-secret",
			ConnectionTimeout: 10,
			Status:            sql.NullString{String: "active", Valid: true},
			Version:           sql.NullInt64{Int64: 3, Valid: true},
			CreatedAt:         sql.NullTime{Time: now, Valid: true},
			UpdatedAt:         sql.NullTime{Time: now, Valid: true},
		}

		entity, err := repository.DatabaseClusterToDomain(row)
		require.NoError(t, err)

		assert.Equal(t, "cluster-1", entity.ID)
		assert.Equal(t, "my-org", entity.OrgName)
		assert.Equal(t, "my-project", entity.ProjectSlug)
		assert.Equal(t, "Production DB", entity.Title)
		assert.Equal(t, "main cluster", entity.Description)
		assert.Equal(t, "db.example.com", entity.Host)
		assert.Equal(t, 3306, entity.Port)
		assert.Equal(t, "admin", entity.Username)
		assert.Equal(t, "encrypted-secret", entity.Password.GetPassword())
		assert.Equal(t, 10, entity.ConnectionTimeout)
		assert.Equal(t, cluster.ClusterStatusActive, entity.Status)
		assert.Equal(t, int64(3), entity.Version)
		assert.Equal(t, now, entity.CreatedAt)
		assert.Equal(t, now, entity.UpdatedAt)
	})

	t.Run("nullable fields NULL", func(t *testing.T) {
		row := dbgen.DatabaseCluster{
			ID:                "cluster-2",
			OrgName:           "org",
			ProjectSlug:       "proj",
			Title:             "Dev DB",
			Host:              "localhost",
			Port:              5432,
			Username:          "dev",
			Password:          "enc-pwd",
			ConnectionTimeout: 5,
		}

		entity, err := repository.DatabaseClusterToDomain(row)
		require.NoError(t, err)

		assert.Equal(t, "", entity.Description)
		assert.Equal(t, cluster.ClusterStatus(""), entity.Status)
		assert.Equal(t, int64(0), entity.Version)
		assert.True(t, entity.CreatedAt.IsZero())
		assert.True(t, entity.UpdatedAt.IsZero())
	})

	t.Run("password round-trips via GetPassword", func(t *testing.T) {
		row := dbgen.DatabaseCluster{
			ID:                "cluster-3",
			OrgName:           "org",
			ProjectSlug:       "proj",
			Title:             "Title",
			Host:              "host",
			Port:              3306,
			Username:          "user",
			Password:          "stored-encrypted-value",
			ConnectionTimeout: 5,
		}

		entity, err := repository.DatabaseClusterToDomain(row)
		require.NoError(t, err)
		assert.Equal(t, "stored-encrypted-value", entity.Password.GetPassword())
	})
}

// TestDatabaseClusterToCreateParams verifies that a domain DatabaseCluster is correctly
// mapped to dbgen.CreateDatabaseClusterParams with proper type conversions.
func TestDatabaseClusterToCreateParams(t *testing.T) {
	setupClusterTestCipher(t)

	t.Run("full cluster", func(t *testing.T) {
		pwd, err := cluster.NewByEncrypt("encrypted-pass")
		require.NoError(t, err)

		entity := &cluster.DatabaseCluster{
			ID:                "cluster-1",
			OrgName:           "my-org",
			ProjectSlug:       "my-project",
			Title:             "Production DB",
			Description:       "main cluster",
			Host:              "db.example.com",
			Port:              3306,
			Username:          "admin",
			Password:          *pwd,
			ConnectionTimeout: 10,
			Status:            cluster.ClusterStatusActive,
			Version:           1,
		}

		params := repository.DatabaseClusterToCreateParams(entity)

		assert.Equal(t, "cluster-1", params.ID)
		assert.Equal(t, "my-org", params.OrgName)
		assert.Equal(t, "my-project", params.ProjectSlug)
		assert.Equal(t, "Production DB", params.Title)
		assert.True(t, params.Description.Valid)
		assert.Equal(t, "main cluster", params.Description.String)
		assert.Equal(t, "db.example.com", params.Host)
		assert.Equal(t, int64(3306), params.Port)
		assert.Equal(t, "admin", params.Username)
		assert.Equal(t, "encrypted-pass", params.Password)
		assert.Equal(t, int32(10), params.ConnectionTimeout)
		assert.True(t, params.Status.Valid)
		assert.Equal(t, "active", params.Status.String)
		assert.True(t, params.Version.Valid)
		assert.Equal(t, int64(1), params.Version.Int64)
	})

	t.Run("empty description maps to invalid NullString", func(t *testing.T) {
		pwd, err := cluster.NewByEncrypt("enc")
		require.NoError(t, err)

		entity := &cluster.DatabaseCluster{
			ID:          "cluster-2",
			OrgName:     "org",
			ProjectSlug: "proj",
			Title:       "Title",
			Host:        "host",
			Port:        3306,
			Username:    "user",
			Password:    *pwd,
		}

		params := repository.DatabaseClusterToCreateParams(entity)

		assert.False(t, params.Description.Valid)
	})
}
