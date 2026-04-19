package rls

// AuthVariable 认证变量定义
type AuthVariable struct {
	Name   string      `json:"name"`   // 如 "tenant_id"
	Source string      `json:"source"` // JWT 路径，如 "jwt.tenant_id"
	Type   AuthVarType `json:"type"`   // uuid | string | integer
}

// AuthVarType 认证变量类型
type AuthVarType string

const (
	AuthVarTypeUUID    AuthVarType = "UUID"
	AuthVarTypeString  AuthVarType = "STRING"
	AuthVarTypeInteger AuthVarType = "INTEGER"
)
