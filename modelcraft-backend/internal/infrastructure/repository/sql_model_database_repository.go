package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
)

// SqlModelDatabaseRepository is the sqlc-based implementation of modeldatabase.ModelDatabaseRepository.
type SqlModelDatabaseRepository struct {
	q dbgen.Querier
}

// NewSqlModelDatabaseRepository creates a new SqlModelDatabaseRepository backed by the
// provided sqlc Querier. Returns a modeldatabase.ModelDatabaseRepository interface value.
func NewSqlModelDatabaseRepository(q dbgen.Querier) modeldatabase.ModelDatabaseRepository {
	return &SqlModelDatabaseRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Create persists a new model database registration.
func (r *SqlModelDatabaseRepository) Create(ctx context.Context, db *modeldatabase.ModelDatabase) error {
	return r.q.CreateModelDatabase(ctx, dbgen.CreateModelDatabaseParams{
		ID:          db.ID,
		OrgName:     db.OrgName,
		ProjectSlug: db.ProjectSlug,
		ClusterID:   db.ClusterID,
		Name:        db.Name,
		Title:       db.Title,
		Description: sql.NullString{String: db.Description, Valid: db.Description != ""},
		Mode:        dbgen.ModelDatabaseMode(db.Mode),
	})
}

// GetByID retrieves a model database by primary key, scoped to org and project.
// Returns NotFoundError if the record does not exist.
func (r *SqlModelDatabaseRepository) GetByID(
	ctx context.Context, orgName, projectSlug, id string,
) (*modeldatabase.ModelDatabase, error) {
	row, err := r.q.GetModelDatabaseByID(ctx, dbgen.GetModelDatabaseByIDParams{
		ID:          id,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("model database not found: " + id)
		}
		return nil, err
	}
	return modelDatabaseToDomain(row), nil
}

// GetByName retrieves a model database by database name, scoped to org and project.
// Returns NotFoundError if the record does not exist.
func (r *SqlModelDatabaseRepository) GetByName(
	ctx context.Context, orgName, projectSlug, name string,
) (*modeldatabase.ModelDatabase, error) {
	row, err := r.q.GetModelDatabaseByName(ctx, dbgen.GetModelDatabaseByNameParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Name:        name,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("model database not found: " + name)
		}
		return nil, err
	}
	return modelDatabaseToDomain(row), nil
}

// List returns all active model databases within a project.
func (r *SqlModelDatabaseRepository) List(
	ctx context.Context, orgName, projectSlug string,
) ([]*modeldatabase.ModelDatabase, error) {
	rows, err := r.q.ListModelDatabasesByProject(ctx, dbgen.ListModelDatabasesByProjectParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*modeldatabase.ModelDatabase, 0, len(rows))
	for _, row := range rows {
		result = append(result, modelDatabaseToDomain(row))
	}
	return result, nil
}

// Update persists changes to an existing model database registration.
func (r *SqlModelDatabaseRepository) Update(
	ctx context.Context, orgName, projectSlug string, db *modeldatabase.ModelDatabase,
) error {
	return r.q.UpdateModelDatabase(ctx, dbgen.UpdateModelDatabaseParams{
		Title:       db.Title,
		Description: sql.NullString{String: db.Description, Valid: db.Description != ""},
		Mode:        dbgen.ModelDatabaseMode(db.Mode),
		ID:          db.ID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
}

// Delete soft-deletes a model database registration by ID.
// It first calls GetByID to verify the record exists (DeleteModelDatabase is :exec and
// cannot return RowsAffected), then performs the soft-delete.
func (r *SqlModelDatabaseRepository) Delete(
	ctx context.Context, orgName, projectSlug, id string,
) error {
	if _, err := r.GetByID(ctx, orgName, projectSlug, id); err != nil {
		return err
	}
	return r.q.DeleteModelDatabase(ctx, dbgen.DeleteModelDatabaseParams{
		ID:          id,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
}

// modelDatabaseToDomain converts a dbgen.ModelDatabase row to a domain ModelDatabase entity.
func modelDatabaseToDomain(row dbgen.ModelDatabase) *modeldatabase.ModelDatabase {
	return &modeldatabase.ModelDatabase{
		ID:          row.ID,
		OrgName:     row.OrgName,
		ProjectSlug: row.ProjectSlug,
		ClusterID:   row.ClusterID,
		Name:        row.Name,
		Title:       row.Title,
		Description: row.Description.String,
		Mode:        modeldatabase.DatabaseMode(row.Mode),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// compile-time interface check
var _ modeldatabase.ModelDatabaseRepository = (*SqlModelDatabaseRepository)(nil)
