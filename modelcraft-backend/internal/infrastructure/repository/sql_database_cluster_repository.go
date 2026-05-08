package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/cluster"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// DatabaseClusterToDomain converts a dbgen.DatabaseCluster row to a domain DatabaseCluster entity.
// It returns an error if the encrypted password cannot be loaded.
func DatabaseClusterToDomain(row dbgen.DatabaseCluster) (*cluster.DatabaseCluster, error) {
	pwd, err := cluster.NewByEncrypt(row.Password)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "DatabaseClusterToDomain: load password for cluster %s", row.ID)
	}

	var createdAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}

	var updatedAt time.Time
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &cluster.DatabaseCluster{
		ID:                row.ID,
		OrgName:           row.OrgName,
		ProjectSlug:       row.ProjectSlug,
		Title:             row.Title,
		Description:       row.Description.String,
		Host:              row.Host,
		Port:              int(row.Port),
		Username:          row.Username,
		Password:          *pwd,
		ConnectionTimeout: int(row.ConnectionTimeout),
		Status:            cluster.ClusterStatus(row.Status.String),
		Version:           row.Version.Int64,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}, nil
}

// DatabaseClusterToCreateParams converts a domain DatabaseCluster to dbgen.CreateDatabaseClusterParams.
func DatabaseClusterToCreateParams(entity *cluster.DatabaseCluster) dbgen.CreateDatabaseClusterParams {
	return dbgen.CreateDatabaseClusterParams{
		ID:                entity.ID,
		OrgName:           entity.OrgName,
		ProjectSlug:       entity.ProjectSlug,
		Title:             entity.Title,
		Description:       sql.NullString{String: entity.Description, Valid: entity.Description != ""},
		Host:              entity.Host,
		Port:              int64(entity.Port),
		Username:          entity.Username,
		Password:          entity.Password.GetPassword(),
		ConnectionTimeout: int32(entity.ConnectionTimeout),
		Status:            sql.NullString{String: string(entity.Status), Valid: entity.Status != ""},
		Version:           sql.NullInt64{Int64: entity.Version, Valid: true},
	}
}

// DatabaseClusterToUpdateParams converts a domain DatabaseCluster to dbgen.UpdateDatabaseClusterWithVersionParams.
func DatabaseClusterToUpdateParams(
	orgName, projectSlug string,
	entity *cluster.DatabaseCluster,
) dbgen.UpdateDatabaseClusterWithVersionParams {
	return dbgen.UpdateDatabaseClusterWithVersionParams{
		Title:             entity.Title,
		Description:       sql.NullString{String: entity.Description, Valid: entity.Description != ""},
		Host:              entity.Host,
		Port:              int64(entity.Port),
		Username:          entity.Username,
		Password:          entity.Password.GetPassword(),
		ConnectionTimeout: int32(entity.ConnectionTimeout),
		Status:            sql.NullString{String: string(entity.Status), Valid: entity.Status != ""},
		ID:                entity.ID,
		OrgName:           orgName,
		ProjectSlug:       projectSlug,
		Version:           sql.NullInt64{Int64: entity.Version, Valid: true},
	}
}

// SqlDatabaseClusterRepository is the sqlc-based implementation of cluster.DatabaseClusterRepository.
type SqlDatabaseClusterRepository struct {
	q dbgen.Querier
}

// NewSqlDatabaseClusterRepository creates a new SqlDatabaseClusterRepository backed by the
// provided sqlc Querier. Returns a cluster.DatabaseClusterRepository interface value.
func NewSqlDatabaseClusterRepository(q dbgen.Querier) cluster.DatabaseClusterRepository {
	return &SqlDatabaseClusterRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

// Create persists a new database cluster.
func (r *SqlDatabaseClusterRepository) Create(ctx context.Context, entity *cluster.DatabaseCluster) error {
	return r.q.CreateDatabaseCluster(ctx, DatabaseClusterToCreateParams(entity))
}

// Update persists changes to an existing cluster using optimistic locking.
// It returns an error if no row was updated due to a version mismatch.
func (r *SqlDatabaseClusterRepository) Update(
	ctx context.Context, orgName, projectSlug string, entity *cluster.DatabaseCluster,
) error {
	result, err := r.q.UpdateDatabaseClusterWithVersion(
		ctx, DatabaseClusterToUpdateParams(orgName, projectSlug, entity),
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return bizerrors.Wrapf(err, "Update cluster: check rows affected")
	}
	if rows == 0 {
		return bizerrors.Errorf(
			"database cluster not found or version mismatch: id=%s version=%d", entity.ID, entity.Version,
		)
	}

	entity.Version++
	return nil
}

// GetByID retrieves a cluster by primary key, scoped to the given org.
// Returns ErrRecordNotFound if the cluster does not exist.
func (r *SqlDatabaseClusterRepository) GetByID(
	ctx context.Context, orgName, id string,
) (*cluster.DatabaseCluster, error) {
	row, err := r.q.GetDatabaseClusterByID(ctx, dbgen.GetDatabaseClusterByIDParams{
		ID:      id,
		OrgName: orgName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("database cluster not found: " + id)
		}
		return nil, err
	}
	return DatabaseClusterToDomain(row)
}

// GetByProjectKey retrieves the cluster associated with the given org and project.
// Returns ErrRecordNotFound if no cluster exists for that project key.
func (r *SqlDatabaseClusterRepository) GetByProjectKey(
	ctx context.Context, orgName, projectSlug string,
) (*cluster.DatabaseCluster, error) {
	row, err := r.q.GetDatabaseClusterByProjectKey(ctx, dbgen.GetDatabaseClusterByProjectKeyParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			msg := "database cluster not found for project: " + orgName + "/" + projectSlug
			return nil, shared.NewNotFoundError(msg)
		}
		return nil, err
	}
	return DatabaseClusterToDomain(row)
}

// List returns all clusters within a project.
func (r *SqlDatabaseClusterRepository) List(
	ctx context.Context, orgName, projectSlug string,
) ([]*cluster.DatabaseCluster, error) {
	rows, err := r.q.ListDatabaseClusters(ctx, dbgen.ListDatabaseClustersParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	return clusterRowsToDomain(rows)
}

// Delete removes a cluster by ID, scoped to org and project.
func (r *SqlDatabaseClusterRepository) Delete(
	ctx context.Context, orgName, projectSlug, id string,
) error {
	return r.q.DeleteDatabaseCluster(ctx, dbgen.DeleteDatabaseClusterParams{
		ID:          id,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
}

// ExistsByProjectKey reports whether a cluster is already registered for the given project.
func (r *SqlDatabaseClusterRepository) ExistsByProjectKey(
	ctx context.Context, orgName, projectSlug string,
) (bool, error) {
	count, err := r.q.ExistsDatabaseClusterByProjectKey(ctx, dbgen.ExistsDatabaseClusterByProjectKeyParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListUpdatedAfter returns clusters updated after the given time.
// orgName is required, projectSlug is optional (pass empty string to omit).
func (r *SqlDatabaseClusterRepository) ListUpdatedAfter(
	ctx context.Context,
	orgName, projectSlug string,
	updatedAfter time.Time,
	status ...cluster.ClusterStatus,
) ([]*cluster.DatabaseCluster, error) {
	projectSlugFilter := sql.NullString{String: projectSlug, Valid: projectSlug != ""}

	rows, err := r.q.ListDatabaseClustersUpdatedAfter(ctx, dbgen.ListDatabaseClustersUpdatedAfterParams{
		UpdatedAt:         sql.NullTime{Time: updatedAfter, Valid: true},
		OrgName:           orgName,
		ProjectSlugFilter: projectSlugFilter,
	})
	if err != nil {
		return nil, err
	}

	entities, err := clusterRowsToDomain(rows)
	if err != nil {
		return nil, err
	}

	if len(status) == 0 {
		return entities, nil
	}

	allowedStatus := make(map[cluster.ClusterStatus]struct{}, len(status))
	for _, s := range status {
		allowedStatus[s] = struct{}{}
	}

	filtered := entities[:0]
	for _, e := range entities {
		if _, ok := allowedStatus[e.Status]; ok {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

// clusterRowsToDomain converts a slice of dbgen rows to domain entities.
func clusterRowsToDomain(rows []dbgen.DatabaseCluster) ([]*cluster.DatabaseCluster, error) {
	entities := make([]*cluster.DatabaseCluster, 0, len(rows))
	for _, row := range rows {
		entity, err := DatabaseClusterToDomain(row)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

// compile-time interface check
var _ cluster.DatabaseClusterRepository = (*SqlDatabaseClusterRepository)(nil)
