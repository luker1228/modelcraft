package membership

import (
	"fmt"
	"time"
)

// MembershipStatus 成员关系状态
type MembershipStatus string

const (
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusSuspended MembershipStatus = "suspended"
)

// Membership 用户-组织关联实体
// 角色信息通过 user_roles 表管理，不在此实体中存储
type Membership struct {
	ID        string           // UUID
	UserID    string           // 用户 ID
	OrgName   string           // 组织名称（主键引用）
	IsAdmin   bool             // 是否为管理员（可访问管理后台）
	Status    MembershipStatus // 成员状态：active | suspended
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate 验证成员关系实体
func (m *Membership) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("membership ID is required")
	}
	if m.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if m.OrgName == "" {
		return fmt.Errorf("organization name is required")
	}
	if m.Status != MembershipStatusActive &&
		m.Status != MembershipStatusSuspended {
		return fmt.Errorf("membership status must be one of: active, suspended")
	}
	return nil
}

// NewMembership 创建成员关系（直接加入，状态为 active）
// 注意：角色通过 user_roles 表管理，不在此处指定
func NewMembership(id, userID, orgName string, isAdmin bool) (*Membership, error) {
	now := time.Now()
	m := &Membership{
		ID:        id,
		UserID:    userID,
		OrgName:   orgName,
		IsAdmin:   isAdmin,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// Suspend 挂起成员
func (m *Membership) Suspend() {
	m.Status = MembershipStatusSuspended
	m.UpdatedAt = time.Now()
}

// IsActive 判断成员是否活跃
func (m *Membership) IsActive() bool {
	return m.Status == MembershipStatusActive
}
