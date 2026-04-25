package enduser

import "context"

// AccessibleProject represents one project that an end-user can access in an org.
type AccessibleProject struct {
	ProjectSlug  string
	ProjectTitle string
}

// EndUserProjectAccessRepository defines persistence operations for end-user project access grants.
type EndUserProjectAccessRepository interface {
	// Grant creates a project access grant.
	Grant(ctx context.Context, access *EndUserProjectAccess) error

	// GetByID retrieves a project access grant by ID under org+project scope.
	// Returns (nil, nil) when not found.
	GetByID(ctx context.Context, orgName, projectSlug, accessID string) (*EndUserProjectAccess, error)

	// UpdatePermissionBundle updates permission_bundle_id for a grant.
	UpdatePermissionBundle(ctx context.Context, orgName, projectSlug, accessID, permissionBundleID string) error

	// Revoke deletes a project access grant.
	Revoke(ctx context.Context, orgName, projectSlug, accessID string) error

	// RemoveByEndUserID deletes all access grants for an end-user in one org.
	RemoveByEndUserID(ctx context.Context, orgName, endUserID string) error

	// PermissionBundleExists checks whether the bundle exists in current org+project scope.
	PermissionBundleExists(ctx context.Context, orgName, projectSlug, permissionBundleID string) (bool, error)

	// ListWithTotal lists project access grants with cursor pagination and optional username search.
	ListWithTotal(ctx context.Context, query ListEndUserProjectAccessQuery) ([]*EndUserProjectAccess, int64, error)

	// ListAccessibleProjectsByUserID lists all accessible projects for a user under one org.
	ListAccessibleProjectsByUserID(ctx context.Context, orgName, endUserID string) ([]AccessibleProject, error)

	// HasProjectAccess returns true when the user has access to the given org+project.
	HasProjectAccess(ctx context.Context, orgName, endUserID, projectSlug string) (bool, error)
}

// ListEndUserProjectAccessQuery defines list query parameters.
type ListEndUserProjectAccessQuery struct {
	OrgName     string
	ProjectSlug string
	Search      string
	First       int
	After       string
}
