package auth

// Issuer JWT 签发者
type Issuer string

const (
	// IssuerPlatform 是统一的 JWT 签发者标识，所有 token 均使用此值。
	IssuerPlatform Issuer = "mc-platform"

	// Deprecated: IssuerDeveloper 已废弃，由 IssuerPlatform 替代。硬切后不再有效。
	IssuerDeveloper Issuer = "mc-developer" //nolint:deadcode

	// Deprecated: IssuerEndUser 已废弃，由 IssuerPlatform 替代。硬切后不再有效。
	IssuerEndUser Issuer = "mc-enduser" //nolint:deadcode
)

// IsValid 判断是否为合法的 Issuer。只有 IssuerPlatform 有效。
func (i Issuer) IsValid() bool {
	return i == IssuerPlatform
}
