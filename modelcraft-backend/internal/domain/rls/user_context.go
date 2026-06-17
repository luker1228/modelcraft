package rls

import (
	"fmt"
	"strings"
)

// UserContext 从 Header 提取的用户上下文
type UserContext struct {
	UserIDNum *int64  `json:"userIdNum"` // 数值型 userId（与 UserIDStr 互斥）
	UserIDStr string  `json:"userIdStr"` // 字符串型 userId（与 UserIDNum 互斥）
	UserName  string  `json:"userName"`
	Roles     []string `json:"roles"`
	UseAdmin  bool     `json:"useAdmin"` // 是否请求 admin 级别访问
}

// UserIDValue 返回 userId 的正确 Go 类型：int64 或 string。
// 下游 SQL 参数绑定和 CEL 求值均使用此方法以保持类型安全。
func (uc *UserContext) UserIDValue() interface{} {
	if uc.UserIDNum != nil {
		return *uc.UserIDNum
	}
	return uc.UserIDStr
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
	case "userid", "uid", "user_id":
		return fmt.Sprint(uc.UserIDValue())
	case "username", "user_name":
		return uc.UserName
	default:
		return ""
	}
}
