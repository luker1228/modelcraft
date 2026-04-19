package rls

// AuthSchema Project 级别认证变量配置
type AuthSchema struct {
	ProjectID string         `json:"projectId"`
	Variables []AuthVariable `json:"variables"`
}

// GetVariable 获取指定名称的变量（uid 内置，不返回）
func (a *AuthSchema) GetVariable(name string) *AuthVariable {
	if name == "uid" {
		return &AuthVariable{Name: "uid", Source: "jwt.user_id", Type: AuthVarTypeUUID}
	}
	for _, v := range a.Variables {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

// IsValidRef 判断变量引用是否合法
func (a *AuthSchema) IsValidRef(name string) bool {
	return a.GetVariable(name) != nil
}
