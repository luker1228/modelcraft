package rls

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/rls"
)

// SetModelRLSPolicyInput 设置 RLS 策略输入
type SetModelRLSPolicyInput struct {
	ModelID         string       `json:"modelId"`
	SelectPredicate rls.JsonExpr `json:"selectPredicate"`
	InsertCheck     rls.JsonExpr `json:"insertCheck"`
	UpdatePredicate rls.JsonExpr `json:"updatePredicate"`
	UpdateCheck     rls.JsonExpr `json:"updateCheck"`
	DeletePredicate rls.JsonExpr `json:"deletePredicate"`
}

// ValidateRLSExprInput 校验 RLS 表达式输入
type ValidateRLSExprInput struct {
	ModelID  string       `json:"modelId"`
	Expr     rls.JsonExpr `json:"expr"`
	ExprType rls.ExprType `json:"exprType"`
}

// ValidationErrorResult 校验错误结果
type ValidationErrorResult struct {
	Errors []rls.ValidationError `json:"errors"`
}

// ApplyRLSPresetInput 应用 RLS 预设输入
type ApplyRLSPresetInput struct {
	ModelID string        `json:"modelId"`
	Preset  rls.RLSPreset `json:"preset"`
}

// DataModelWrapper 包装 DataModel 用于校验
type DataModelWrapper struct {
	Model  *modeldesign.ModelMeta
	Fields []*modeldesign.FieldDefinition
}

// HasField 检查模型是否有指定字段
func (w *DataModelWrapper) HasField(fieldName string) bool {
	for _, f := range w.Fields {
		if f.Name == fieldName {
			return true
		}
	}
	return false
}

// GetField 获取指定字段
func (w *DataModelWrapper) GetField(fieldName string) *modeldesign.FieldDefinition {
	for _, f := range w.Fields {
		if f.Name == fieldName {
			return f
		}
	}
	return nil
}
