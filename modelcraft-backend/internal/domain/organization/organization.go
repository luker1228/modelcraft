package organization

import (
	"fmt"
	"regexp"
	"time"
)

// OrgStatus 组织状态
type OrgStatus string

const (
	OrgStatusActive    OrgStatus = "active"
	OrgStatusSuspended OrgStatus = "suspended"
	OrgStatusDeleted   OrgStatus = "deleted"
)

// Organization 组织实体
// 多租户逻辑容器，用于隔离项目、集群、模型等资源
type Organization struct {
	Name        string    // 唯一标识符（来自 AuthProvider 组织名称），也是主键
	DisplayName string    // 可选的 UI 显示名称
	OwnerID     string    // 组织创建者/所有者的用户 ID
	Phone       string    // Org 注册手机号，全局唯一
	Status      OrgStatus // 组织状态
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// orgNamePattern 组织名称格式：小写字母、数字、下划线、连字符，以字母开头
var orgNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// isValidOrgName 验证组织名称格式
// - 2-64 字符
// - 小写字母、数字、下划线、连字符
// - 必须以字母开头
func isValidOrgName(name string) bool {
	if len(name) < 2 || len(name) > 64 {
		return false
	}
	return orgNamePattern.MatchString(name)
}

// Validate 验证组织实体
func (o *Organization) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("organization name is required")
	}
	if !isValidOrgName(o.Name) {
		return fmt.Errorf(
			"organization name must be 2-64 characters, " +
				"lowercase letters/digits/underscores/hyphens only, " +
				"and start with a letter",
		)
	}
	if o.OwnerID == "" {
		return fmt.Errorf("organization owner ID is required")
	}
	if o.Status != OrgStatusActive && o.Status != OrgStatusSuspended && o.Status != OrgStatusDeleted {
		return fmt.Errorf("organization status must be one of: active, suspended, deleted")
	}
	return nil
}

// NewOrganization 创建组织实体
func NewOrganization(name, displayName, ownerID, phone string) (*Organization, error) {
	now := time.Now()
	org := &Organization{
		Name:        name,
		DisplayName: displayName,
		OwnerID:     ownerID,
		Phone:       phone,
		Status:      OrgStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := org.Validate(); err != nil {
		return nil, err
	}
	return org, nil
}

// UpdateDisplayName 更新显示名称
func (o *Organization) UpdateDisplayName(displayName string) {
	o.DisplayName = displayName
	o.UpdatedAt = time.Now()
}

// Suspend 挂起组织
func (o *Organization) Suspend() {
	o.Status = OrgStatusSuspended
	o.UpdatedAt = time.Now()
}

// Activate 激活组织
func (o *Organization) Activate() {
	o.Status = OrgStatusActive
	o.UpdatedAt = time.Now()
}

// MarkDeleted 标记为删除（软删除）
func (o *Organization) MarkDeleted() {
	o.Status = OrgStatusDeleted
	o.UpdatedAt = time.Now()
}

// IsActive 判断组织是否活跃
func (o *Organization) IsActive() bool {
	return o.Status == OrgStatusActive
}
