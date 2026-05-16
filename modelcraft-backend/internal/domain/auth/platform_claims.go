package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PlatformClaims 是统一 JWT 的 payload 结构，适用于所有用户类型（tenant / end-user）。
// 通过标准 aud 字段区分 token 的受众类型（目前仅携带，端点级鉴权后续实现）。
type PlatformClaims struct {
	UserID          string            `json:"user_id"`
	OrgName         string            `json:"org_name"`
	Key             string            `json:"key"` // APISIX jwt-auth Consumer key
	EndUserAdminIDs map[string]string `json:"end_user_admin_ids,omitempty"`
	jwt.RegisteredClaims
}

// EndUserAdminIDs: orgName → end-user super-admin ID (tenant tokens only).
// Used by the gateway to inject X-End-User-Admin-ID when a tenant admin
// accesses end-user org endpoints.

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
