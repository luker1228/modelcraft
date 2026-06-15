package rls

import "context"

// PolicyProvider 提供 RLS 策略的接口
// 由 modeldesign.DataModel 实现，避免直接依赖
type PolicyProvider interface {
	// GetPolicy 返回模型的 RLS 策略，如果没有则返回 nil
	GetPolicy() *ModelRLSPolicy
}

// ModelRLSPolicy RLS 策略实体（五件套 JsonExpr）
// 定义在 rls 包中以避免循环依赖
type ModelRLSPolicy struct {
	ModelID         string   `json:"modelId"`
	SelectPredicate JsonExpr `json:"selectPredicate"` // SELECT USING
	InsertCheck     JsonExpr `json:"insertCheck"`     // INSERT WITH CHECK
	UpdatePredicate JsonExpr `json:"updatePredicate"` // UPDATE USING
	UpdateCheck     JsonExpr `json:"updateCheck"`     // UPDATE WITH CHECK
	DeletePredicate JsonExpr `json:"deletePredicate"` // DELETE USING
}

// RLSResolver RLS 过滤器解析器
type RLSResolver interface {
	// Resolve 根据 runtime end-user 语义和 Model 策略解析 RLSFilter
	// - 上下文中无 EndUserID → nil / deny（取决于具体实现）
	// - model.getPolicy() == nil → DENY ALL（无 Policy = Default Deny）
	// - 否则 → RLSFilter { 五件套 JsonExpr, endUserId }
	Resolve(ctx context.Context, model PolicyProvider) (*RLSFilter, error)
}

// DenyAllFilter 返回 DENY ALL 过滤器（全 false）
var DenyAllFilter = &RLSFilter{
	SelectPredicate: JsonExpr(`false`),
	InsertCheck:     JsonExpr(`false`),
	UpdatePredicate: JsonExpr(`false`),
	UpdateCheck:     JsonExpr(`false`),
	DeletePredicate: JsonExpr(`false`),
	FieldName:       "owner",
}
