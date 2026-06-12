package rls

import "strings"

// UserContext 从 Header 提取的用户上下文
type UserContext struct {
	UserID   string   `json:"userId"`
	UserName string   `json:"userName"`
	Roles    []string `json:"roles"`
}

// HasRole 判断是否包含指定 role
// role="" 总是匹配（默认策略）
func (uc *UserContext) HasRole(role string) bool {
	if role == "" {
		return true
	}
	for _, r := range uc.Roles {
		if strings.TrimSpace(r) == role {
			return true
		}
	}
	return false
}

// ResolveVariable 解析表达式变量
func (uc *UserContext) ResolveVariable(name string) string {
	switch name {
	case "uid", "user_id":
		return uc.UserID
	case "user_name":
		return uc.UserName
	default:
		return ""
	}
}
