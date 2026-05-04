package auth

// Issuer JWT 签发者
type Issuer string

const (
	IssuerDeveloper Issuer = "mc-developer" // 开发者 JWT
	IssuerEndUser   Issuer = "mc-enduser"   // 终端用户 JWT
)

// IsValid 判断是否为合法的 Issuer
func (i Issuer) IsValid() bool {
	switch i {
	case IssuerDeveloper, IssuerEndUser:
		return true
	}
	return false
}
