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

	// GetByIDGlobal retrieves a user by ID without an org filter.
	// The single-org-per-user invariant ensures the returned row carries the owning org.
	GetByIDGlobal(ctx context.Context, id string) (*EndUser, error)

	// GetByUsername retrieves a user by username under org scope (returns (nil, nil) when not found).
	GetByUsername(ctx context.Context, orgName, username string) (*EndUser, error)

	// GetByPhone retrieves a user by phone number under org scope (returns (nil, nil) when not found).
	GetByPhone(ctx context.Context, orgName, phone string) (*EndUser, error)

	// GetByPhoneGlobal retrieves a user by phone without an org filter.
	// Phone numbers are globally unique, so no org scope is needed.
	GetByPhoneGlobal(ctx context.Context, phone string) (*EndUser, error)

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

// APITokenRepository 定义 EndUser PAT 的持久化操作。
// 所有写操作都在系统 DB（非 tenant DB）上执行。
type APITokenRepository interface {
	// Save 插入新 token 记录（id 已由调用方生成）。
	Save(ctx context.Context, token *APIToken) error
	// FindByHash 通过 SHA-256 hash 查找活跃 token（未软删除）。
	// 未找到时返回 (nil, nil)。
	FindByHash(ctx context.Context, hash string) (*APIToken, error)
	// ListByUser 返回指定用户的全部活跃 token 列表（按 created_at DESC）。
	ListByUser(ctx context.Context, orgName, endUserID string) ([]*APIToken, error)
	// SoftDelete 软删除指定 token（设置 deleted_at + delete_token）。
	// 若 token 不属于该用户则返回 error。
	SoftDelete(ctx context.Context, id, orgName, endUserID string) error
	// UpdateLastUsed 异步更新 last_used_at，验证成功后调用。
	UpdateLastUsed(ctx context.Context, id string, at time.Time) error
}
