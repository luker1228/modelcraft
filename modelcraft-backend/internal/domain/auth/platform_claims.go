package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PlatformClaims 是统一 JWT 的 payload 结构，适用于所有用户类型（tenant / end-user）。
// aud 始终为 AudiencePlatform；is_admin 字段区分管理员与普通用户。
// is_admin 由 Gateway 注入 X-Is-Admin header，供后端鉴权使用。
type PlatformClaims struct {
	UserID  string `json:"user_id"`
	OrgName string `json:"org_name"`
	IsAdmin bool   `json:"is_admin"`
	Key     string `json:"key"` // APISIX jwt-auth Consumer key
	jwt.RegisteredClaims
}

var ErrPlatformClaimsInvalid = errors.New("invalid platform claims")

// Validate 校验 PlatformClaims 的完整性和有效性。
func (c *PlatformClaims) Validate() error {
	if c.UserID == "" {
		return fmt.Errorf("%w: user_id is required", ErrPlatformClaimsInvalid)
	}
	if c.OrgName == "" {
		return fmt.Errorf("%w: org_name is required", ErrPlatformClaimsInvalid)
	}
	if c.Issuer != string(IssuerPlatform) {
		return fmt.Errorf("%w: expected issuer %q, got %q", ErrPlatformClaimsInvalid, IssuerPlatform, c.Issuer)
	}
	if c.ExpiresAt == nil || c.ExpiresAt.Before(time.Now()) {
		return ErrTokenExpired
	}
	return nil
}

// IsExpired 检查 token 是否已过期。
func (c *PlatformClaims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return !time.Now().Before(c.ExpiresAt.Time)
}
