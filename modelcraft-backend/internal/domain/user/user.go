package user

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"slices"
	"strings"
	"time"
)

// User 用户实体
type User struct {
	ID           string      // ModelCraft 内部 UUID
	ExternalID   string      // 外部认证提供者用户 ID（来自 JWT.sub，通常为 AuthProvider 用户 ID）
	Name         string      // 用户姓名
	Phone        PhoneNumber // 用户手机号（值对象）
	PasswordHash string      // 密码哈希（仅手机号+密码注册的用户有值）
	OrgName      string      // 所属 Org，创建时绑定
	IsAdmin      bool        // 是否为管理员（原 user_orgs.is_admin）
	Status       string      // 状态：active | suspended（原 user_orgs.status）
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

// ValidateUserName validates the userName format and reserved words.
// Rules:
//   - Length: 3-32 characters
//   - Pattern: ^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$ (start with letter/underscore/hyphen,
//     followed by alphanumeric/underscore/hyphen)
//   - Reserved words: admin, root, system, etc.
//
// Returns nil if valid, error with description otherwise.
func ValidateUserName(name string) error {
	if len(name) < 3 || len(name) > 32 {
		return fmt.Errorf("userName must be 3-32 characters, got %d", len(name))
	}

	// Pattern: start with letter/underscore/hyphen, then alphanumeric/underscore/hyphen
	pattern := regexp.MustCompile(`^[a-zA-Z_-][a-zA-Z0-9_-]{2,31}$`)
	if !pattern.MatchString(name) {
		return fmt.Errorf("userName must start with letter/underscore/hyphen and contain only alphanumeric/underscore/hyphen") //nolint:lll
	}

	// Reserved words check (case-insensitive)
	lowerName := strings.ToLower(name)
	reservedWords := []string{
		"admin", "administrator", "root", "system", "sys",
		"modelcraft", "support", "help", "api", "www",
		"null", "undefined", "anonymous",
	}
	if slices.Contains(reservedWords, lowerName) {
		return fmt.Errorf("userName '%s' is reserved", name)
	}

	return nil
}

//nolint:unused
var registerNameAdjectives = [...]string{
	"brisk", "calm", "clever", "daring", "eager", "fancy", "gentle", "happy",
	"jolly", "kind", "lucky", "merry", "noble", "quick", "royal", "smart",
}

//nolint:unused
var registerNameNouns = [...]string{
	"aurora", "bamboo", "cloud", "dolphin", "ember", "falcon", "glade", "harbor",
	"island", "jungle", "kitten", "legend", "meteor", "nebula", "orchid", "phoenix",
}

// NewUser 通过用户名+密码创建用户实体（phone 可为空，demo org owner 场景下无手机号）
func NewUser(id, userName string, phone PhoneNumber, passwordHash, orgName string) (*User, error) {
	if passwordHash == "" {
		return nil, fmt.Errorf("password hash is required")
	}
	if orgName == "" {
		return nil, fmt.Errorf("org name is required")
	}
	if err := ValidateUserName(userName); err != nil {
		return nil, err
	}
	now := time.Now()
	user := &User{
		ID:           id,
		Name:         userName,
		Phone:        phone,
		PasswordHash: passwordHash,
		OrgName:      orgName,
		IsAdmin:      true,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

//nolint:unused
func generateRegisterDisplayName(userID string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(userID))
	sum := h.Sum64()

	adjective := registerNameAdjectives[int(sum%uint64(len(registerNameAdjectives)))]
	noun := registerNameNouns[int((sum>>8)%uint64(len(registerNameNouns)))]

	compactID := strings.ToLower(strings.ReplaceAll(userID, "-", ""))

	return fmt.Sprintf("%s-%s-%s", adjective, noun, compactID[:8])
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
		IsAdmin:    false,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}
