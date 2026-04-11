package profile

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxNicknameLength  = 32
	maxAvatarURLLength = 512
	maxBioLength       = 256
)

const (
	defaultNicknamePrefix = "user_"
	nicknameRandomLength  = 6
)

const nicknameAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Profile 用户资料聚合（最小字段）
type Profile struct {
	ID        string
	UserID    string
	Nickname  string
	AvatarURL *string
	Bio       *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpdatePatch 定义 profile 的部分更新字段（PATCH 语义）
type UpdatePatch struct {
	Nickname  *string
	AvatarURL *string
	Bio       *string
}

// IsEmpty 判断 patch 是否未包含任何更新字段。
func (p UpdatePatch) IsEmpty() bool {
	return p.Nickname == nil && p.AvatarURL == nil && p.Bio == nil
}

// NewProfile 创建 profile 实体。
func NewProfile(id, userID, nickname string, avatarURL, bio *string) (*Profile, error) {
	now := time.Now()
	profile := &Profile{
		ID:        id,
		UserID:    userID,
		Nickname:  nickname,
		AvatarURL: avatarURL,
		Bio:       bio,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := profile.Validate(); err != nil {
		return nil, err
	}

	return profile, nil
}

// NewInitialProfile 创建用户初始 profile。
// 当 nickname 为空时，自动生成默认昵称：user_ + 6位大写字母数字。
func NewInitialProfile(id, userID, nickname string, avatarURL, bio *string) (*Profile, error) {
	effectiveNickname := strings.TrimSpace(nickname)
	if effectiveNickname == "" {
		effectiveNickname = generateDefaultNickname()
	}

	return NewProfile(id, userID, effectiveNickname, avatarURL, bio)
}

func generateDefaultNickname() string {
	randomBytes := make([]byte, nicknameRandomLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return defaultNicknamePrefix + "000000"
	}

	for i := range randomBytes {
		randomBytes[i] = nicknameAlphabet[int(randomBytes[i])%len(nicknameAlphabet)]
	}

	return defaultNicknamePrefix + string(randomBytes)
}

// Validate 校验 profile 聚合不变量。
func (p *Profile) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("profile id is required")
	}
	if p.UserID == "" {
		return fmt.Errorf("profile userID is required")
	}

	nickname := strings.TrimSpace(p.Nickname)
	if nickname == "" {
		return fmt.Errorf("profile nickname is required")
	}
	if utf8.RuneCountInString(nickname) > maxNicknameLength {
		return fmt.Errorf("profile nickname must be <= %d characters", maxNicknameLength)
	}

	if p.AvatarURL != nil && utf8.RuneCountInString(*p.AvatarURL) > maxAvatarURLLength {
		return fmt.Errorf("profile avatarURL must be <= %d characters", maxAvatarURLLength)
	}

	if p.Bio != nil && utf8.RuneCountInString(*p.Bio) > maxBioLength {
		return fmt.Errorf("profile bio must be <= %d characters", maxBioLength)
	}

	return nil
}

// ApplyPatch 将 patch 更新应用到当前 profile，并进行校验。
func (p *Profile) ApplyPatch(patch UpdatePatch) error {
	if patch.Nickname != nil {
		p.Nickname = *patch.Nickname
	}
	if patch.AvatarURL != nil {
		p.AvatarURL = patch.AvatarURL
	}
	if patch.Bio != nil {
		p.Bio = patch.Bio
	}
	p.UpdatedAt = time.Now()

	return p.Validate()
}
