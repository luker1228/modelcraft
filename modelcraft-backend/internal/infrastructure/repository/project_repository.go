package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"time"
)

// ProjectToDomain converts a dbgen.Project row to a domain Project entity.
func ProjectToDomain(row dbgen.Project) *project.Project {
	var createdAt, updatedAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}

	return &project.Project{
		OrgName:     row.OrgName,
		Slug:        row.Slug,
		Title:       row.Title,
		Description: row.Description.String,
		LoginURL:    row.LoginUrl.String,
		ClusterID:   NullStrToPtr(row.ClusterID),
		Status:      project.ProjectStatus(row.Status),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// ProjectToCreateParams converts a domain Project to dbgen.CreateProjectParams.
func ProjectToCreateParams(p *project.Project) dbgen.CreateProjectParams {
	return dbgen.CreateProjectParams{
		OrgName:     p.OrgName,
		Slug:        p.Slug,
		Title:       p.Title,
		Description: sql.NullString{String: p.Description, Valid: p.Description != ""},
		LoginUrl:    sql.NullString{String: p.LoginURL, Valid: p.LoginURL != ""},
		ClusterID:   PtrToNullStr(p.ClusterID),
		Status:      string(p.Status),
	}
}

// ProjectToUpdateParams converts a domain Project to dbgen.UpdateProjectParams.
func ProjectToUpdateParams(p *project.Project) dbgen.UpdateProjectParams {
	return dbgen.UpdateProjectParams{
		Title:       p.Title,
		Description: sql.NullString{String: p.Description, Valid: p.Description != ""},
		LoginUrl:    sql.NullString{String: p.LoginURL, Valid: p.LoginURL != ""},
		ClusterID:   PtrToNullStr(p.ClusterID),
		Slug:        p.Slug,
		OrgName:     p.OrgName,
	}
}

// SqlProjectRepository is the sqlc-based implementation of project.ProjectRepository.
type SqlProjectRepository struct {
	q dbgen.Querier
}

// NewSqlProjectRepository creates a SqlProjectRepository.
func NewSqlProjectRepository(q dbgen.Querier) project.ProjectRepository {
	return &SqlProjectRepository{q: q}
}

// Create creates a new project.
func (r *SqlProjectRepository) Create(ctx context.Context, p *project.Project) error {
	return ExecWithErrorHandling(func() error {
		return r.q.CreateProject(ctx, ProjectToCreateParams(p))
	})
}

// GetByNameAndOrg retrieves a project by slug and org (primary key lookup).
// Returns nil, shared.NewNotFoundError if not found.
func (r *SqlProjectRepository) GetByNameAndOrg(ctx context.Context, slug, orgName string) (*project.Project, error) {
	var row dbgen.Project
	if err := QueryWithSQLErrorHandling(func() error {
		var err error
		row, err = r.q.GetProjectBySlugAndOrg(ctx, dbgen.GetProjectBySlugAndOrgParams{
			Slug:    slug,
			OrgName: orgName,
		})
		return err
	}); err != nil {
		return nil, err
	}
	if row.Slug == "" {
		return nil, shared.NewNotFoundError("project not found: " + orgName + "/" + slug)
	}
	return ProjectToDomain(row), nil
}

// GetByClusterID retrieves a project by org and cluster ID.
// Returns nil, shared.NewNotFoundError if not found.
func (r *SqlProjectRepository) GetByClusterID(
	ctx context.Context, orgName, clusterID string,
) (*project.Project, error) {
	var row dbgen.Project
	err := QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetProjectByClusterID(ctx, dbgen.GetProjectByClusterIDParams{
			OrgName:   orgName,
			ClusterID: sql.NullString{String: clusterID, Valid: true},
		})
		return e
	})
	if err != nil {
		return nil, err
	}
	if row.Slug == "" {
		return nil, shared.NewNotFoundError("project not found for cluster: " + clusterID)
	}
	return ProjectToDomain(row), nil
}

// List retrieves all projects, optionally filtered by status.
func (r *SqlProjectRepository) List(ctx context.Context, status *project.ProjectStatus) ([]*project.Project, error) {
	if status != nil {
		return r.listByStatus(ctx, *status)
	}
	var rows []dbgen.Project
	if err := QueryWithSQLErrorHandling(func() error {
		var err error
		rows, err = r.q.ListProjects(ctx)
		return err
	}); err != nil {
		return nil, err
	}
	return projectRowsToDomain(rows), nil
}

func (r *SqlProjectRepository) listByStatus(
	ctx context.Context, status project.ProjectStatus,
) ([]*project.Project, error) {
	var rows []dbgen.Project
	if err := QueryWithSQLErrorHandling(func() error {
		var err error
		rows, err = r.q.ListProjects(ctx)
		return err
	}); err != nil {
		return nil, err
	}
	var result []*project.Project
	for _, row := range rows {
		if row.Status == string(status) {
			result = append(result, ProjectToDomain(row))
		}
	}
	return result, nil
}

// ListByOrg retrieves projects for a specific org, optionally filtered by status.
func (r *SqlProjectRepository) ListByOrg(
	ctx context.Context, orgName string, status *project.ProjectStatus,
) ([]*project.Project, error) {
	var rows []dbgen.Project
	if err := QueryWithSQLErrorHandling(func() error {
		var err error
		rows, err = r.q.ListProjectsByOrg(ctx, orgName)
		return err
	}); err != nil {
		return nil, err
	}

	if status == nil {
		return projectRowsToDomain(rows), nil
	}

	var result []*project.Project
	for _, row := range rows {
		if row.Status == string(*status) {
			result = append(result, ProjectToDomain(row))
		}
	}
	return result, nil
}

// Update updates an existing project.
func (r *SqlProjectRepository) Update(ctx context.Context, p *project.Project) error {
	return ExecWithErrorHandling(func() error {
		return r.q.UpdateProject(ctx, ProjectToUpdateParams(p))
	})
}

// Archive archives a project (status → archived).
func (r *SqlProjectRepository) Archive(ctx context.Context, slug, orgName string) error {
	return ExecWithErrorHandling(func() error {
		return r.q.ArchiveProject(ctx, dbgen.ArchiveProjectParams{
			Slug:    slug,
			OrgName: orgName,
		})
	})
}

// ExistsByName checks if a project slug exists in the org.
func (r *SqlProjectRepository) ExistsByName(ctx context.Context, slug, orgName string) (bool, error) {
	var count int64
	if err := QueryWithSQLErrorHandling(func() error {
		var err error
		count, err = r.q.ExistsProjectBySlug(ctx, dbgen.ExistsProjectBySlugParams{
			Slug:    slug,
			OrgName: orgName,
		})
		return err
	}); err != nil {
		return false, err
	}
	return count > 0, nil
}

func projectRowsToDomain(rows []dbgen.Project) []*project.Project {
	result := make([]*project.Project, len(rows))
	for i, row := range rows {
		result[i] = ProjectToDomain(row)
	}
	return result
}
