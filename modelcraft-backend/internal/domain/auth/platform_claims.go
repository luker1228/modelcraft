package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PlatformClaims 是统一 JWT 的 payload 结构，适用于所有用户类型（平台管理员 / 端用户）。
// 通过 Scope 字段区分 token 的访问范围，废弃了原 mc-developer / mc-enduser 双 issuer 设计。
type PlatformClaims struct {
	UserID  string `json:"user_id"`
	OrgName string `json:"org_name"`
	// Scope 标识 token 的访问范围："org" | "project" | "service_key"（预留）
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

const (
	// TokenScopeOrg 表示 Org 级别 token，可访问 org 管理接口
	TokenScopeOrg = "org"
	// TokenScopeProject 表示 Project 级别 token，可访问 project 数据接口
	TokenScopeProject = "project"
	// TokenScopeServiceKey 预留，本期不签发
	TokenScopeServiceKey = "service_key"
)

var ErrPlatformClaimsInvalid = errors.New("invalid platform claims")

// Validate 校验 PlatformClaims 的完整性和有效性。
func (c *PlatformClaims) Validate() error {
	if c.UserID == "" {
		return fmt.Errorf("%w: user_id is required", ErrPlatformClaimsInvalid)
	}
	if c.OrgName == "" {
		return fmt.Errorf("%w: org_name is required", ErrPlatformClaimsInvalid)
	}
	validScope := c.Scope == TokenScopeOrg || c.Scope == TokenScopeProject || c.Scope == TokenScopeServiceKey
	if !validScope {
		return fmt.Errorf("%w: invalid scope %q, must be org|project|service_key", ErrPlatformClaimsInvalid, c.Scope)
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
