package rls

// EndUserIdentity 终端用户身份信息
type EndUserIdentity struct {
	EndUserID string `json:"endUserId"`
	Issuer    string `json:"issuer"` // mc-developer 或 mc-enduser
}

// IsEndUser 判断是否为合法的 Runtime 调用者（EndUser JWT）
func (e *EndUserIdentity) IsEndUser() bool {
	return e.Issuer == "mc-enduser"
}

// IsDeveloper 判断是否为 Developer JWT
func (e *EndUserIdentity) IsDeveloper() bool {
	return e.Issuer == "mc-developer"
}
