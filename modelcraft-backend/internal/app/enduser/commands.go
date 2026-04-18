package enduser

import "time"

// --- Authentication Commands ---

// RegisterCommand represents a self-registration request from an end-user.
type RegisterCommand struct {
	OrgName     string
	ProjectSlug string
	Username    string
	Password    string
}

// RegisterResult represents the result of a successful registration.
type RegisterResult struct {
	UserID       string
	RefreshToken string // opaque token plaintext (returned only once)
	ExpiresAt    time.Time
}

// LoginCommand represents a login request from an end-user.
type LoginCommand struct {
	OrgName     string
	ProjectSlug string
	Username    string
	Password    string
}

// LoginResult represents the result of a successful login.
// Alias to RegisterResult as they have identical structure.
type LoginResult = RegisterResult

// LogoutCommand represents a logout request from an end-user.
type LogoutCommand struct {
	OrgName      string
	ProjectSlug  string
	RefreshToken string // opaque token plaintext
}

// RefreshCommand represents a token refresh request from an end-user.
type RefreshCommand struct {
	OrgName      string
	ProjectSlug  string
	RefreshToken string // opaque token plaintext
}

// RefreshResult represents the result of a successful token refresh.
// Alias to RegisterResult as they have identical structure.
type RefreshResult = RegisterResult

// GetMeCommand represents a request to get the current end-user's profile.
type GetMeCommand struct {
	OrgName     string
	ProjectSlug string
	UserID      string // parsed from JWT by BFF, passed via X-End-User-Id header
}

// --- User Management Commands (by developers) ---

// CreateEndUserCommand represents a request from a developer to create an end-user.
type CreateEndUserCommand struct {
	OrgName     string
	ProjectSlug string
	Username    string
	Password    string
	CreatedBy   string // developer user_id from mc_meta
}

// ListEndUsersCommand represents a request to list end-users with pagination.
type ListEndUsersCommand struct {
	OrgName     string
	ProjectSlug string
	Search      string // username fuzzy search (optional)
	First       int    // page size, default 20
	After       string // cursor (ID of last item from previous page)
}

// ListEndUsersResult represents the result of listing end-users.
type ListEndUsersResult struct {
	Items       []*EndUserDTO
	TotalCount  int64
	HasNextPage bool
	EndCursor   string
}

// EndUserItem represents an end-user in list results.
type EndUserItem struct {
	ID          string
	Username    string
	IsForbidden bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpdateEndUserStatusCommand represents a request to update an end-user's status.
type UpdateEndUserStatusCommand struct {
	OrgName     string
	ProjectSlug string
	UserID      string
	IsForbidden bool
}

// DeleteEndUserCommand represents a request to delete an end-user.
type DeleteEndUserCommand struct {
	OrgName     string
	ProjectSlug string
	UserID      string
}

// --- Result Types for User Management (required by end_user_app_service.go) ---

// CreateEndUserResult represents the result of creating an end-user.
type CreateEndUserResult struct {
	ID          string
	Username    string
	IsForbidden bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EndUserDTO represents an end-user data transfer object.
type EndUserDTO struct {
	ID          string
	Username    string
	IsForbidden bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
