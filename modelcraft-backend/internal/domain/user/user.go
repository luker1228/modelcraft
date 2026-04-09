package user

import (
	"fmt"
	"hash/fnv"
	"strings"
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

var registerNameAdjectives = [...]string{
	"brisk", "calm", "clever", "daring", "eager", "fancy", "gentle", "happy",
	"jolly", "kind", "lucky", "merry", "noble", "quick", "royal", "smart",
}

var registerNameNouns = [...]string{
	"aurora", "bamboo", "cloud", "dolphin", "ember", "falcon", "glade", "harbor",
	"island", "jungle", "kitten", "legend", "meteor", "nebula", "orchid", "phoenix",
}

// NewUser 通过手机号+密码创建用户实体
// Name 自动设置为随机风格且稳定可复现的 displayName（基于 userID 生成）
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
		Name:         generateRegisterDisplayName(id),
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

func generateRegisterDisplayName(userID string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(userID))
	sum := h.Sum64()

	adjective := registerNameAdjectives[int(sum%uint64(len(registerNameAdjectives)))]
	noun := registerNameNouns[int((sum>>8)%uint64(len(registerNameNouns)))]

	compactID := strings.ToLower(strings.ReplaceAll(userID, "-", ""))
	if compactID == "" {
		compactID = "000000"
	}
	if len(compactID) < 6 {
		compactID = strings.Repeat("0", 6-len(compactID)) + compactID
	}
	suffix := compactID[len(compactID)-6:]

	return fmt.Sprintf("%s_%s_%s", adjective, noun, suffix)
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
