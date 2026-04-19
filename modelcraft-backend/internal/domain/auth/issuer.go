package auth

// Issuer JWT 签发者
type Issuer string

const (
	IssuerDeveloper Issuer = "mc-developer"  // 开发者 JWT
	IssuerEndUser   Issuer = "mc-enduser"    // 终端用户 JWT
	IssuerLegacy    Issuer = "modelcraft"    // 兼容旧版（需迁移）
)

// IsValid 判断是否为合法的 Issuer
func (i Issuer) IsValid() bool {
	switch i {
	case IssuerDeveloper, IssuerEndUser, IssuerLegacy:
		return true
	}
	return false
}
