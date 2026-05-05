package rls

// EndUserIdentity 终端用户身份信息（统一 token 体系后通过 scope 区分身份）
type EndUserIdentity struct {
	EndUserID string `json:"endUserId"`
	Issuer    string `json:"issuer"` // 统一为 mc-platform
	Scope     string `json:"scope"`  // "org" | "project"
}

// IsEndUser 判断是否为可访问 Runtime 的身份（scope=org 或 scope=project 均可）。
// 统一 token 体系后，issuer 已不足以区分身份，改用 scope 判断。
func (e *EndUserIdentity) IsEndUser() bool {
	return e.Issuer == "mc-platform" && (e.Scope == "org" || e.Scope == "project")
}

// IsDeveloper Deprecated：统一 token 体系后，developer/enduser 概念消失，保留以兼容。
// 下游代码应改为 scope 判断。
func (e *EndUserIdentity) IsDeveloper() bool {
	return e.Issuer == "mc-platform" && e.Scope == "org"
}
