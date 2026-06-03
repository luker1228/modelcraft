package enduser

import (
	"time"
)

// --- User Management Commands (by developers) ---

// CreateEndUserCommand represents a request from a developer to create an end-user.
type CreateEndUserCommand struct {
	OrgName     string
	ProjectSlug string
	Username    string
	Password    string
	Phone       string // 手机号（必填）
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

// ResetEndUserPasswordCommand represents a request to reset an end-user's password.
type ResetEndUserPasswordCommand struct {
	OrgName     string
	UserID      string
	NewPassword string
}

// --- Result Types for User Management (required by end_user_app_service.go) ---

// CreateEndUserResult represents the result of creating an end-user.
type CreateEndUserResult struct {
	ID          string
	Username    string
	IsForbidden bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EndUserDTO represents an end-user data transfer object.
type EndUserDTO struct {
	ID          string
	Username    string
	IsForbidden bool
	IsAdmin     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GetEndUserCommand represents a request to get a single end-user.
type GetEndUserCommand struct {
	OrgName     string
	ProjectSlug string
	UserID      string
}

// AccessibleProjectItem represents one project that an end-user can access in an org.
type AccessibleProjectItem struct {
	Slug        string
	Title       string
	Description string
	Status      string // "active" | "archived"
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// --- Runtime Meta/User Query Commands ---

// MetaUserFindOneCommand 通过唯一条件（id 或 username）查询单个 user。
// OrgName 始终由上下文注入，禁止客户端传入。
// meta/user 操作 Org 级表 (end_user_users)，不需要 ProjectSlug。
type MetaUserFindOneCommand struct {
	OrgName  string // 从中间件上下文注入
	ID       string // 唯一条件之一（id 或 username 必须提供其中一个）
	Username string // 唯一条件之一
}

// MetaUserFindManyFilter 白名单过滤字段。
type MetaUserFindManyFilter struct {
	IDEq               *string
	IDIn               []string
	UsernameEq         *string
	UsernameContains   *string
	UsernameStartsWith *string
	UsernameIn         []string
	CreatedAtEq        *string
	CreatedAtGte       *string
	CreatedAtLte       *string
}

// MetaUserFindManyCommand 受限列表查询命令（cursor 分页）。
// 排序固定为 created_at DESC, id DESC（最新优先）。
// first 默认 20，最大 50；After 为空表示第一页。
type MetaUserFindManyCommand struct {
	OrgName string // 从中间件上下文注入
	Where   *MetaUserFindManyFilter
	After   string // cursor（Base64 编码），空字符串表示第一页
	First   int    // page size，默认 20，最大 50
}

// MetaUserDTO runtime meta/user 查询结果 DTO（不含租户字段）。
type MetaUserDTO struct {
	ID        string
	Username  string
	CreatedAt time.Time
}

// MetaUserFindManyResult findMany 查询结果（cursor 分页）。
type MetaUserFindManyResult struct {
	Items      []*MetaUserDTO
	NextCursor string // 下一页 cursor（Base64 编码）；空字符串表示无更多数据
	HasMore    bool
}

// CreateUserCommand 创建统一用户（管理员或普通用户）。
type CreateUserCommand struct {
	OrgName  string
	Username string
	Password string
	Phone    string
	IsAdmin  bool
}

// CreateUserResult 创建用户的返回结果。
type CreateUserResult struct {
	ID          string
	Username    string
	IsAdmin     bool
	IsForbidden bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
