package user

import (
	"fmt"
	"time"
)

// User 用户实体
type User struct {
	ID         string    // ModelCraft 内部 UUID
	ExternalID string    // 外部认证提供者用户 ID（来自 JWT.sub，通常为 Casdoor 用户 ID）
	Name       string    // 用户姓名（来自 Casdoor）
	Phone      string    // 用户手机号（来自 Casdoor）
	CreatedAt  time.Time // 创建时间
	UpdatedAt  time.Time // 更新时间
}

// Validate 验证用户实体
func (u *User) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	if u.ExternalID == "" {
		return fmt.Errorf("external ID is required")
	}
	return nil
}

// NewUser 创建用户实体
func NewUser(id, externalID, name, phone string) (*User, error) {
	now := time.Now()
	user := &User{
		ID:         id,
		ExternalID: externalID,
		Name:       name,
		Phone:      phone,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}
