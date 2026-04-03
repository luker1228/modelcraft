package auth

import "time"

// LoginCommand BFF 登录时传入的用户信息（来自 Casdoor token）
type LoginCommand struct {
	ExternalID string
	Email      string
	Name       string
}

// LoginResult 登录成功后返回给 BFF
type LoginResult struct {
	UserID       string
	RefreshToken string // 明文，BFF 存入 Cookie
	ExpiresAt    time.Time
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
