package enduser

import "context"

// EndUserRepository defines persistence operations for end-users.
// Operations target mc_meta.end_user_users and are tenant-scoped by org.
type EndUserRepository interface {
	// Save creates a new end-user (INSERT, unique conflict returns shared.ErrTypeDuplicatedKey).
	Save(ctx context.Context, user *EndUser) error

	// GetByID retrieves a user by ID under org scope (returns (nil, nil) when not found).
	GetByID(ctx context.Context, orgName, id string) (*EndUser, error)

	// GetByUsername retrieves a user by username under org scope (returns (nil, nil) when not found).
	GetByUsername(ctx context.Context, orgName, username string) (*EndUser, error)

	// UpdateStatus updates the is_forbidden field under org scope.
	UpdateStatus(ctx context.Context, orgName, id string, isForbidden bool) error

	// Delete physically deletes a user record under org scope.
	Delete(ctx context.Context, orgName, id string) error

	// ListWithTotal retrieves users with pagination and optional search.
	// Returns (users, totalCount, error).
	ListWithTotal(ctx context.Context, query ListEndUsersQuery) ([]*EndUser, int64, error)
}

// ListEndUsersQuery defines list query parameters.
type ListEndUsersQuery struct {
	OrgName string // org scope key (required)
	Search  string // username fuzzy search (optional)
	First   int    // page size, default 20, max 100
	After   string // cursor (ID of the last item from previous page, optional)
}
