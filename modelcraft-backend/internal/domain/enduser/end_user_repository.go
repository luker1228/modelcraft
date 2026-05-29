package enduser

import (
	"context"
	"time"
)

// AccessibleProject represents one project that an end-user can access in an org.
type AccessibleProject struct {
	ProjectSlug        string
	ProjectTitle       string
	ProjectDescription string
	ProjectStatus      string // raw value: "active" | "archived"
	ProjectCreatedAt   time.Time
	ProjectUpdatedAt   time.Time
}

// EndUserRepository defines persistence operations for end-users.
// Operations target mc_meta.end_user_users and are tenant-scoped by org.
type EndUserRepository interface {
	// Save creates a new end-user (INSERT, unique conflict returns shared.ErrTypeDuplicatedKey).
	Save(ctx context.Context, user *EndUser) error

	// GetByID retrieves a user by ID under org scope (returns (nil, nil) when not found).
	GetByID(ctx context.Context, orgName, id string) (*EndUser, error)

	// GetByUsername retrieves a user by username under org scope (returns (nil, nil) when not found).
	GetByUsername(ctx context.Context, orgName, username string) (*EndUser, error)

	// GetByUsernameGlobal retrieves a user by username without requiring org input.
	// The returned entity carries its orgName and relies on the single-org-per-user invariant.
	GetByUsernameGlobal(ctx context.Context, username string) (*EndUser, error)

	// UpdateStatus updates the is_forbidden field under org scope.
	UpdateStatus(ctx context.Context, orgName, id string, isForbidden bool) error

	// UpdatePassword updates the password hash for a user under org scope.
	UpdatePassword(ctx context.Context, orgName, id string, hashedPassword HashedPassword) error

	// Delete physically deletes a user record under org scope.
	Delete(ctx context.Context, orgName, id string) error

	// ListWithTotal retrieves users with pagination and optional search.
	// Returns (users, totalCount, error).
	ListWithTotal(ctx context.Context, query ListEndUsersQuery) ([]*EndUser, int64, error)

	// ListAccessibleProjectsByRoleAssignment 通过 end_user_role_users + end_user_roles 查询
	// 用户在该 Org 下可访问的 Project 列表（替代旧的 end_user_project_access 路径）。
	ListAccessibleProjectsByRoleAssignment(ctx context.Context, orgName, endUserID string) ([]AccessibleProject, error)

	// ListAllProjectsByOrg 返回 org 下所有未删除的 project（供 org admin 使用）。
	ListAllProjectsByOrg(ctx context.Context, orgName string) ([]AccessibleProject, error)

	// HasProjectAccessByRole 检查用户在指定 org+project 下是否有任意 Role 分配
	HasProjectAccessByRole(ctx context.Context, orgName, endUserID, projectSlug string) (bool, error)
}

// ListEndUsersQuery defines list query parameters.
type ListEndUsersQuery struct {
	OrgName string // org scope key (required)
	Search  string // username fuzzy search (optional)
	First   int    // page size, default 20, max 100
	After   string // cursor (ID of the last item from previous page, optional)
}
