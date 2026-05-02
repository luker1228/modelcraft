package rbac

import (
	"modelcraft/internal/domain/enduser"
	"time"
)

// ProjectEndUserRoleUser 表示 Project 维度下某用户的一条角色分配记录
type ProjectEndUserRoleUser struct {
	ID         string
	EndUser    *enduser.EndUser
	Role       *EndUserRole
	AssignedAt time.Time
}

// ListProjectEndUserRoleUsersQuery 查询参数
type ListProjectEndUserRoleUsersQuery struct {
	OrgName     string
	ProjectSlug string
	Search      string // username 模糊搜索
	RoleID      string // 按 Role 过滤（可选）
	First       int
	After       string // cursor
}
