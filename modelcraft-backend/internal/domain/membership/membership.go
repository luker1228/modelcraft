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
	MembershipStatusInvited   MembershipStatus = "invited"
)

// Membership 用户-组织关联实体
// 角色信息通过 user_roles 表管理，不在此实体中存储
type Membership struct {
	ID        string           // UUID
	UserID    string           // 用户 ID
	OrgName   string           // 组织名称（主键引用）
	Status    MembershipStatus // 成员状态
	InvitedBy string           // 邀请人用户 ID（可为空）
	InvitedAt *time.Time       // 邀请时间（可为空）
	JoinedAt  *time.Time       // 加入时间（可为空）
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
		m.Status != MembershipStatusSuspended &&
		m.Status != MembershipStatusInvited {
		return fmt.Errorf("membership status must be one of: active, suspended, invited")
	}
	return nil
}

// NewMembership 创建成员关系（直接加入，状态为 active）
// 注意：角色通过 user_roles 表管理，不在此处指定
func NewMembership(id, userID, orgName string) (*Membership, error) {
	now := time.Now()
	m := &Membership{
		ID:        id,
		UserID:    userID,
		OrgName:   orgName,
		Status:    MembershipStatusActive,
		JoinedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// NewInvitation 创建邀请关系（状态为 invited）
// 注意：角色通过 user_roles 表管理，不在此处指定
func NewInvitation(id, userID, orgName, invitedBy string) (*Membership, error) {
	now := time.Now()
	m := &Membership{
		ID:        id,
		UserID:    userID,
		OrgName:   orgName,
		Status:    MembershipStatusInvited,
		InvitedBy: invitedBy,
		InvitedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// AcceptInvitation 接受邀请，状态从 invited 变为 active
func (m *Membership) AcceptInvitation() error {
	if m.Status != MembershipStatusInvited {
		return fmt.Errorf("can only accept invitation when status is 'invited', current: %s", m.Status)
	}
	now := time.Now()
	m.Status = MembershipStatusActive
	m.JoinedAt = &now
	m.UpdatedAt = now
	return nil
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
