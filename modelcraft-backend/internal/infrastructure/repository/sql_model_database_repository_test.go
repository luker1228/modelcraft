package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/infrastructure/dbgen"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlModelDatabaseRepository_Create_PopulatesTimestamps(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSqlModelDatabaseRepository(dbgen.New(db))

	entity := &modeldatabase.ModelDatabase{
		ID:          "db-1",
		OrgName:     "org-a",
		ProjectSlug: "project-a",
		ClusterID:   "cluster-1",
		Name:        "test_db",
		Title:       "test_db",
		Description: "",
		Mode:        modeldatabase.DatabaseModeManaged,
	}
	now := time.Date(2026, 5, 29, 10, 11, 12, 123000000, time.UTC)

	mock.ExpectExec("INSERT INTO model_database").
		WithArgs(
			entity.ID,
			entity.OrgName,
			entity.ProjectSlug,
			entity.ClusterID,
			entity.Name,
			entity.Title,
			sql.NullString{},
			dbgen.ModelDatabaseMode(entity.Mode),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rows := sqlmock.NewRows([]string{
		"id",
		"org_name",
		"project_slug",
		"cluster_id",
		"name",
		"title",
		"description",
		"mode",
		"deleted_at",
		"delete_token",
		"created_at",
		"updated_at",
	}).AddRow(
		entity.ID,
		entity.OrgName,
		entity.ProjectSlug,
		entity.ClusterID,
		entity.Name,
		entity.Title,
		sql.NullString{},
		dbgen.ModelDatabaseMode(entity.Mode),
		uint64(0),
		uint64(0),
		now,
		now,
	)

	selectQuery := "SELECT id, org_name, project_slug, cluster_id, name, title," +
		" description, mode, deleted_at, delete_token, created_at, updated_at FROM model_database"
	mock.ExpectQuery(selectQuery).
		WithArgs(entity.ID, entity.OrgName, entity.ProjectSlug).
		WillReturnRows(rows)

	err = repo.Create(context.Background(), entity)
	require.NoError(t, err)
	assert.Equal(t, now, entity.CreatedAt)
	assert.Equal(t, now, entity.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}
