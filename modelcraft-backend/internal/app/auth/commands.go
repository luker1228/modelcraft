package auth

import "time"

// IdentifierType 登录标识符类型
type IdentifierType string

const (
	// IdentifierTypePhone 手机号登录
	IdentifierTypePhone IdentifierType = "PHONE"
	// IdentifierTypeUsername 用户名登录
	IdentifierTypeUsername IdentifierType = "USERNAME"
)

// RegisterCommand 手机号+密码注册
type RegisterCommand struct {
	Phone    string
	Password string
	UserName string // 用户名（3-32字符，以字母/下划线/连字符开头）
}

// RegisterProfileSnapshot 注册成功后返回的 profile 快照。
type RegisterProfileSnapshot struct {
	ID        string
	UserID    string
	Nickname  string
	AvatarURL *string
	Bio       *string
}

// RegisterResult 注册成功后返回
type RegisterResult struct {
	UserID  string
	OrgName string // 自动创建的个人组织 slug
	Profile RegisterProfileSnapshot
}

// LoginCommand 登录命令 - 支持手机号或用户名
type LoginCommand struct {
	// Identifier 登录标识符（手机号或用户名）
	Identifier string
	// IdentifierType 标识符类型：PHONE 或 USERNAME，默认为 PHONE
	IdentifierType IdentifierType
	// Password 密码
	Password string
	// Deprecated: Phone 保留用于向后兼容，新代码应使用 Identifier + IdentifierType
	Phone string
}

// LoginResult 登录成功后返回给 BFF
type LoginResult struct {
	UserID       string
	UserName     string // 用户显示名
	OrgName      string // 用户首个组织名（如有）
	AccessToken  string // ES256 签发的短期 JWT，Gateway 用公钥验证
	RefreshToken string // 明文，BFF 存入 Cookie
	ExpiresAt    time.Time
}

// OAuthLoginCommand BFF 通过 OAuth 登录时传入的用户信息（来自 AuthProvider token）
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
	AccessToken  string // 新签发的 ES256 JWT
	RefreshToken string // 新明文 token
	ExpiresAt    time.Time
}

// LogoutCommand BFF 登出时传入
type LogoutCommand struct {
	RefreshToken string // 明文
}
