package auth

import "time"

// RegisterCommand 手机号+密码注册
type RegisterCommand struct {
	Phone    string
	Password string
}

// RegisterResult 注册成功后返回
type RegisterResult struct {
	UserID string
}

// LoginCommand 手机号+密码登录
type LoginCommand struct {
	Phone    string
	Password string
}

// LoginResult 登录成功后返回给 BFF
type LoginResult struct {
	UserID       string
	RefreshToken string // 明文，BFF 存入 Cookie
	ExpiresAt    time.Time
}

// OAuthLoginCommand BFF 通过 OAuth 登录时传入的用户信息（来自 Casdoor token）
// Deprecated: 保留兼容，新流程使用 LoginCommand
type OAuthLoginCommand struct {
	ExternalID string
	Email      string
	Name       string
}

// RefreshCommand BFF 刷新时传入
type RefreshCommand struct {
	RefreshToken string // 明文
}

// RefreshResult 刷新成功后返回给 BFF
type RefreshResult struct {
	UserID       string
	RefreshToken string // 新明文 token
	ExpiresAt    time.Time
}

// LogoutCommand BFF 登出时传入
type LogoutCommand struct {
	RefreshToken string // 明文
}
