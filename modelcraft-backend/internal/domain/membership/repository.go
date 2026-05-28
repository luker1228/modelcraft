package membership

import (
	"context"
	"time"
)

// MembershipWithDetails 包含成员关系及其关联的组织和角色详细信息
type MembershipWithDetails struct {
	OrgName     string    // 组织名称（唯一标识符）
	DisplayName string    // 组织显示名称
	RoleName    string    // 角色名称
	IsAdmin     bool      // 是否为管理员
	JoinedAt    time.Time // 加入时间
}

// MembershipWithUserName 包含成员关系及用户姓名，通过 LEFT JOIN users 表一次性获取
type MembershipWithUserName struct {
	Membership *Membership // 成员关系实体
	UserName   string      // 用户姓名，来自 users 表
}

// MembershipRepository 成员关系仓储接口
type MembershipRepository interface {
	// Create 创建成员关系
	Create(ctx context.Context, membership *Membership) error

	// GetByID 根据 UUID 获取成员关系
	GetByID(ctx context.Context, id string) (*Membership, error)

	// GetByUserAndOrg 根据用户 ID 和组织名称获取成员关系
	GetByUserAndOrg(ctx context.Context, userID, orgName string) (*Membership, error)

	// ListByOrg 获取组织的所有成员
	ListByOrg(ctx context.Context, orgName string) ([]*Membership, error)

	// ListByOrgWithUserName 获取组织的所有成员，并通过 LEFT JOIN users 表带出用户姓名
	ListByOrgWithUserName(ctx context.Context, orgName string) ([]*MembershipWithUserName, error)

	// ListByUser 获取用户的所有成员关系
	ListByUser(ctx context.Context, userID string) ([]*Membership, error)

	// CountByUser returns the number of organizations the user belongs to.
	CountByUser(ctx context.Context, userID string) (int64, error)

	// ListByUserWithDetails 获取用户的所有成员关系，包含组织和角色详细信息（用于 token exchange）
	// 返回的 memberships 按加入时间倒序排列，限制为 limit 条
	ListByUserWithDetails(ctx context.Context, userID string, limit int) ([]*MembershipWithDetails, error)

	// Update 更新成员关系（角色变更、状态变更等）
	Update(ctx context.Context, membership *Membership) error

	// Delete 删除成员关系（将用户从组织中移除）
	Delete(ctx context.Context, id string) error
}
