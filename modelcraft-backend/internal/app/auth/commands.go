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
	Phone            string
	Password         string
	UserName         string // 用户名（3-32字符，以字母/下划线/连字符开头）
	OrganizationName string // 可选，组织名 slug（6-24字符，小写字母/数字/下划线），为空时自动生成
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
	// IdentifierType 标识符类型：PHONE 或 USERNAME，默认为 USERNAME
	IdentifierType IdentifierType
	// Password 密码
	Password string
}

// LoginResult 登录成功后返回给 Gateway
type LoginResult struct {
	UserID       string
	UserName     string // 用户显示名
	OrgName      string // 用户首个组织名（如有）
	AccessToken  string // JWT access token
	RefreshToken string // 明文，Gateway 存入 httpOnly Cookie
	ExpiresIn    int    // access token TTL（秒）
}

// RefreshCommand Gateway 刷新时传入
type RefreshCommand struct {
	RefreshToken string // 明文（从 cookie 中取出）
}

// RefreshResult 刷新成功后返回给 Gateway
type RefreshResult struct {
	AccessToken  string // 新签发的 JWT
	RefreshToken string // 新明文 token，Gateway 写入 cookie
	ExpiresIn    int    // access token TTL（秒）
}

// LogoutCommand Gateway 登出时传入
type LogoutCommand struct {
	RefreshToken string // 明文（从 cookie 中取出）
}

// LoginEndUserCommand EndUser 登录命令，支持 username 或 phone。
type LoginEndUserCommand struct {
	OrgName        string
	Identifier     string         // 登录标识符（用户名或手机号）
	IdentifierType IdentifierType // "USERNAME" 或 "PHONE"，默认 USERNAME
	Password       string
}

// LoginEndUserResult EndUser 登录成功后返回。
type LoginEndUserResult struct {
	UserID       string
	OrgName      string
	AccessToken  string
	RefreshToken string // 明文，由 handler 写入 httpOnly Cookie
	ExpiresAt    time.Time
}

// RefreshEndUserCommand EndUser 刷新 token 命令。
type RefreshEndUserCommand struct {
	OrgName      string // 用于查找 tenant DB（用户隔离）
	RefreshToken string // 明文（从 cookie 中取出）
}

// RefreshEndUserResult EndUser 刷新 token 成功后返回。
type RefreshEndUserResult struct {
	UserID       string
	OrgName      string
	AccessToken  string
	RefreshToken string // 新明文 token
	ExpiresAt    time.Time
}

// GetEndUserMeCommand 获取当前 EndUser 信息命令（身份从 Bearer JWT 解析）。
type GetEndUserMeCommand struct {
	OrgName string
	UserID  string
}
