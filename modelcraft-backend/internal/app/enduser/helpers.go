package enduser

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// 密码强度要求：至少 8 位，含字母 + 数字
var (
	passwordMinLength = 8
	passwordHasLetter = regexp.MustCompile(`[a-zA-Z]`)
	passwordHasDigit  = regexp.MustCompile(`[0-9]`)
	usernamePattern   = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)
)

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	if len(password) < passwordMinLength {
		return fmt.Errorf("password must be at least %d characters", passwordMinLength)
	}
	if !passwordHasLetter.MatchString(password) {
		return fmt.Errorf("password must contain at least one letter")
	}
	if !passwordHasDigit.MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}
	return nil
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("username must be 3-64 characters, alphanumeric, underscore or hyphen only")
	}
	return nil
}

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	// cost=12 是生产推荐值
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword 验证密码是否匹配
func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// GenerateUUID 生成 UUID（使用 crypto/rand）
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// GenerateOpaqueToken 生成 opaque refresh token
func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
