package enduser

import "context"

// EndUserRepository defines persistence operations for end-users.
// Operations target mc_meta.end_user_users and are tenant-scoped by org/project.
type EndUserRepository interface {
	// Save creates a new end-user (INSERT, unique conflict returns shared.ErrTypeDuplicatedKey).
	Save(ctx context.Context, user *EndUser) error

	// GetByID retrieves a user by ID (returns (nil, nil) when not found).
	GetByID(ctx context.Context, id string) (*EndUser, error)

	// GetByUsername retrieves a user by username (returns (nil, nil) when not found).
	GetByUsername(ctx context.Context, username string) (*EndUser, error)

	// UpdateStatus updates the is_forbidden field (checks RowsAffected, returns NotFoundError on 0 rows).
	UpdateStatus(ctx context.Context, id string, isForbidden bool) error

	// Delete physically deletes a user record (checks RowsAffected, returns NotFoundError on 0 rows).
	Delete(ctx context.Context, id string) error

	// ListWithTotal retrieves users with pagination and optional search.
	// Returns (users, totalCount, error).
	ListWithTotal(ctx context.Context, query ListEndUsersQuery) ([]*EndUser, int64, error)
}

// ListEndUsersQuery defines list query parameters.
type ListEndUsersQuery struct {
	Search string // username fuzzy search (optional)
	First  int    // page size, default 20, max 100
	After  string // cursor (ID of the last item from previous page, optional)
}
