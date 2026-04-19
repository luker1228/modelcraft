package rls

import (
	"context"
	"encoding/json"
	"fmt"

	"modelcraft/internal/domain/rls"
)

// PolicyExecutor PolicyExecutor 实现
type PolicyExecutor struct{}

// NewPolicyExecutor 创建 PolicyExecutor
func NewPolicyExecutor() *PolicyExecutor {
	return &PolicyExecutor{}
}

// ToSQL 将编译后的策略绑定运行时 authCtx 生成最终 SQL
func (e *PolicyExecutor) ToSQL(ctx context.Context, compiled *rls.CompiledPolicy,
	authCtx *rls.AuthContext) (string, []interface{}, error) {

	if compiled == nil {
		return "", nil, fmt.Errorf("compiled policy is nil")
	}

	// 如果没有参数，直接返回 SQL
	if len(compiled.Params) == 0 {
		return compiled.SQL, nil, nil
	}

	// 替换参数
	var boundParams []interface{}
	for _, param := range compiled.Params {
		switch v := param.(type) {
		case map[string]string:
			// 处理 _auth 引用
			if authVar, ok := v["_auth"]; ok {
				value, err := e.resolveAuthVar(authCtx, authVar)
				if err != nil {
					return "", nil, err
				}
				boundParams = append(boundParams, value)
			} else {
				boundParams = append(boundParams, param)
			}
		default:
			boundParams = append(boundParams, param)
		}
	}

	return compiled.SQL, boundParams, nil
}

// resolveAuthVar 解析 _auth 变量
func (e *PolicyExecutor) resolveAuthVar(authCtx *rls.AuthContext, varName string) (interface{}, error) {
	switch varName {
	case "uid":
		if authCtx.EndUserID == "" {
			return nil, fmt.Errorf("auth context missing endUserId")
		}
		return authCtx.EndUserID, nil
	default:
		// 从扩展变量中查找
		if authCtx.Variables != nil {
			if value, ok := authCtx.Variables[varName]; ok {
				return value, nil
			}
		}
		return nil, fmt.Errorf("unknown auth variable: %s", varName)
	}
}

// ValidateCheck 应用层校验 CHECK 约束
func (e *PolicyExecutor) ValidateCheck(ctx context.Context, expr rls.JsonExpr,
	rowData map[string]interface{}, authCtx *rls.AuthContext) error {

	var root interface{}
	if err := json.Unmarshal([]byte(expr), &root); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// 处理布尔常量
	if b, ok := root.(bool); ok {
		if b {
			return nil
		}
		return fmt.Errorf("CHECK constraint violated: false")
	}

	obj, ok := root.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expression must be an object or boolean constant")
	}

	return e.validateCheckNode(ctx, obj, rowData, authCtx)
}

func (e *PolicyExecutor) validateCheckNode(ctx context.Context, node map[string]interface{},
	rowData map[string]interface{}, authCtx *rls.AuthContext) error {

	for key, value := range node {
		switch key {
		case "_and":
			arr, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("_and value must be an array")
			}
			for _, item := range arr {
				itemObj, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("_and array item must be an object")
				}
				if err := e.validateCheckNode(ctx, itemObj, rowData, authCtx); err != nil {
					return err
				}
			}

		case "_or":
			arr, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("_or value must be an array")
			}
			for _, item := range arr {
				itemObj, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("_or array item must be an object")
				}
				if err := e.validateCheckNode(ctx, itemObj, rowData, authCtx); err == nil {
					return nil // 只要有一个满足就通过
				}
			}
			return fmt.Errorf("CHECK constraint violated: _or conditions not met")

		case "_not":
			obj, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("_not value must be an object")
			}
			if err := e.validateCheckNode(ctx, obj, rowData, authCtx); err == nil {
				return fmt.Errorf("CHECK constraint violated: _not condition is true")
			}

		case "_const":
			if b, ok := value.(bool); ok && !b {
				return fmt.Errorf("CHECK constraint violated: _const is false")
			}

		default:
			// 字段比较
			compObj, ok := value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("field comparison must be an object")
			}
			if err := e.validateFieldComparison(key, compObj, rowData, authCtx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *PolicyExecutor) validateFieldComparison(fieldName string, compObj map[string]interface{},
	rowData map[string]interface{}, authCtx *rls.AuthContext) error {

	fieldValue, exists := rowData[fieldName]

	for op, expectedValue := range compObj {
		switch op {
		case "_eq":
			resolvedValue, err := e.resolveValue(expectedValue, authCtx)
			if err != nil {
				return err
			}
			if !exists || !e.valuesEqual(fieldValue, resolvedValue) {
				return fmt.Errorf("CHECK constraint violated: %s != %v", fieldName, resolvedValue)
			}
		case "_neq":
			resolvedValue, err := e.resolveValue(expectedValue, authCtx)
			if err != nil {
				return err
			}
			if exists && e.valuesEqual(fieldValue, resolvedValue) {
				return fmt.Errorf("CHECK constraint violated: %s == %v", fieldName, resolvedValue)
			}
		case "_is_null":
			isNull, ok := expectedValue.(bool)
			if !ok {
				return fmt.Errorf("_is_null value must be a boolean")
			}
			if isNull && exists && fieldValue != nil {
				return fmt.Errorf("CHECK constraint violated: %s is not null", fieldName)
			}
			if !isNull && (!exists || fieldValue == nil) {
				return fmt.Errorf("CHECK constraint violated: %s is null", fieldName)
			}
		}
	}

	return nil
}

func (e *PolicyExecutor) resolveValue(value interface{}, authCtx *rls.AuthContext) (interface{}, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		// 处理 _auth 引用
		if authVar, ok := v["_auth"].(string); ok {
			return e.resolveAuthVar(authCtx, authVar)
		}
		return v, nil
	default:
		return v, nil
	}
}

func (e *PolicyExecutor) valuesEqual(a, b interface{}) bool {
	// 简单的值比较，可以根据需要扩展
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case float64:
		// JSON 数字解析为 float64
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int:
			return av == float64(bv)
		case int64:
			return av == float64(bv)
		}
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case nil:
		return b == nil
	}
	return false
}

// Ensure PolicyExecutor implements rls.PolicyExecutor
var _ rls.PolicyExecutor = (*PolicyExecutor)(nil)
