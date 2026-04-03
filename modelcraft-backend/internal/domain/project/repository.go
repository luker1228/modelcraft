package project

import (
	"context"
)

// ProjectRepository defines the interface for project persistence operations
// Primary key is composite: (org_name, slug)
type ProjectRepository interface {
	// Create creates a new project
	Create(ctx context.Context, project *Project) error

	// GetByNameAndOrg retrieves a project by its slug and organization (primary key lookup)
	GetByNameAndOrg(ctx context.Context, slug, orgName string) (*Project, error)

	// GetByClusterID retrieves a project by its cluster ID with multi-tenant isolation
	// Returns nil if no project has the specified cluster ID within the organization
	GetByClusterID(ctx context.Context, orgName, clusterID string) (*Project, error)

	// List retrieves all projects, optionally filtered by status
	List(ctx context.Context, status *ProjectStatus) ([]*Project, error)

	// ListByOrg retrieves projects for a specific organization, optionally filtered by status
	ListByOrg(ctx context.Context, orgName string, status *ProjectStatus) ([]*Project, error)

	// Update updates an existing project
	Update(ctx context.Context, project *Project) error

	// Archive archives a project (soft delete - status change only)
	Archive(ctx context.Context, slug, orgName string) error

	// ExistsByName checks if a project with the given slug exists in the organization
	ExistsByName(ctx context.Context, slug, orgName string) (bool, error)
}
