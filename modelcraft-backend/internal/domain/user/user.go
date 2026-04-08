package user

import (
	"fmt"
	"time"
)

// User 用户实体
type User struct {
	ID           string      // ModelCraft 内部 UUID
	ExternalID   string      // 外部认证提供者用户 ID（来自 JWT.sub，通常为 Casdoor 用户 ID）
	Name         string      // 用户姓名
	Phone        PhoneNumber // 用户手机号（值对象）
	PasswordHash string      // 密码哈希（仅手机号+密码注册的用户有值）
	CreatedAt    time.Time   // 创建时间
	UpdatedAt    time.Time   // 更新时间
}

// Validate 验证用户实体
func (u *User) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("user ID is required")
	}
	return nil
}

// NewUser 通过手机号+密码创建用户实体
// Name 自动设置为手机号脱敏格式 (如 "138****8000")
func NewUser(id string, phone PhoneNumber, passwordHash string) (*User, error) {
	if phone.IsZero() {
		return nil, fmt.Errorf("phone number is required")
	}
	if passwordHash == "" {
		return nil, fmt.Errorf("password hash is required")
	}
	now := time.Now()
	user := &User{
		ID:           id,
		Name:         phone.Masked(),
		Phone:        phone,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

// NewOAuthUser 通过外部认证提供者（OAuth）创建用户实体
func NewOAuthUser(id, externalID, name, phone string) (*User, error) {
	if externalID == "" {
		return nil, fmt.Errorf("external ID is required")
	}
	now := time.Now()
	var phoneVO PhoneNumber
	if phone != "" {
		var err error
		phoneVO, err = NewPhoneNumber(phone)
		if err != nil {
			// OAuth 用户的手机号可能为空或非标准格式，不强制校验
			phoneVO = PhoneNumber{}
		}
	}
	user := &User{
		ID:         id,
		ExternalID: externalID,
		Name:       name,
		Phone:      phoneVO,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}
