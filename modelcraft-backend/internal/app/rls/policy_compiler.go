package rls

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"modelcraft/internal/domain/rls"
)

// PolicyCompiler PolicyCompiler 实现
type PolicyCompiler struct{}

// NewPolicyCompiler 创建 PolicyCompiler
func NewPolicyCompiler() *PolicyCompiler {
	return &PolicyCompiler{}
}

// Compile 将 JSON 表达式编译为 CompiledPolicy
func (c *PolicyCompiler) Compile(ctx context.Context, expr rls.JsonExpr) (*rls.CompiledPolicy, error) {
	var root interface{}
	if err := json.Unmarshal([]byte(expr), &root); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// 处理布尔常量
	if b, ok := root.(bool); ok {
		if b {
			return &rls.CompiledPolicy{SQL: "1=1", Params: nil}, nil
		}
		return &rls.CompiledPolicy{SQL: "1=0", Params: nil}, nil
	}

	obj, ok := root.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expression must be an object or boolean constant")
	}

	return c.compileNode(ctx, obj, "")
}

func (c *PolicyCompiler) compileNode(ctx context.Context, node map[string]interface{}, path string) (*rls.CompiledPolicy, error) {
	var conditions []string
	var params []interface{}

	for key, value := range node {
		switch key {
		case "_and":
			arr, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("_and value must be an array")
			}
			var andConditions []string
			for _, item := range arr {
				itemObj, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("_and array item must be an object")
				}
				compiled, err := c.compileNode(ctx, itemObj, path)
				if err != nil {
					return nil, err
				}
				andConditions = append(andConditions, "("+compiled.SQL+")")
				params = append(params, compiled.Params...)
			}
			if len(andConditions) > 0 {
				conditions = append(conditions, "("+strings.Join(andConditions, " AND ")+")")
			}

		case "_or":
			arr, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("_or value must be an array")
			}
			var orConditions []string
			for _, item := range arr {
				itemObj, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("_or array item must be an object")
				}
				compiled, err := c.compileNode(ctx, itemObj, path)
				if err != nil {
					return nil, err
				}
				orConditions = append(orConditions, "("+compiled.SQL+")")
				params = append(params, compiled.Params...)
			}
			if len(orConditions) > 0 {
				conditions = append(conditions, "("+strings.Join(orConditions, " OR ")+")")
			}

		case "_not":
			obj, ok := value.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("_not value must be an object")
			}
			compiled, err := c.compileNode(ctx, obj, path)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, "NOT ("+compiled.SQL+")")
			params = append(params, compiled.Params...)

		case "_const":
			if b, ok := value.(bool); ok {
				if b {
					conditions = append(conditions, "1=1")
				} else {
					conditions = append(conditions, "1=0")
				}
			}

		default:
			// 字段比较
			compObj, ok := value.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("field comparison must be an object")
			}
			fieldCond, fieldParams, err := c.compileFieldComparison(key, compObj)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, fieldCond)
			params = append(params, fieldParams...)
		}
	}

	if len(conditions) == 0 {
		return &rls.CompiledPolicy{SQL: "1=1", Params: nil}, nil
	}

	return &rls.CompiledPolicy{
		SQL:    strings.Join(conditions, " AND "),
		Params: params,
	}, nil
}

func (c *PolicyCompiler) compileFieldComparison(fieldName string, compObj map[string]interface{}) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	for op, value := range compObj {
		switch op {
		case "_eq":
			sql, param := c.compileValue(fieldName, value)
			conditions = append(conditions, sql)
			if param != nil {
				params = append(params, param)
			}
		case "_neq":
			sql, param := c.compileValue(fieldName, value)
			conditions = append(conditions, "NOT "+sql)
			if param != nil {
				params = append(params, param)
			}
		case "_gt":
			conditions = append(conditions, fmt.Sprintf("%s > ?", fieldName))
			params = append(params, value)
		case "_gte":
			conditions = append(conditions, fmt.Sprintf("%s >= ?", fieldName))
			params = append(params, value)
		case "_lt":
			conditions = append(conditions, fmt.Sprintf("%s < ?", fieldName))
			params = append(params, value)
		case "_lte":
			conditions = append(conditions, fmt.Sprintf("%s <= ?", fieldName))
			params = append(params, value)
		case "_is_null":
			if b, ok := value.(bool); ok && b {
				conditions = append(conditions, fmt.Sprintf("%s IS NULL", fieldName))
			} else {
				conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", fieldName))
			}
		case "_in":
			arr, ok := value.([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("_in value must be an array")
			}
			placeholders := make([]string, len(arr))
			for i := range arr {
				placeholders[i] = "?"
				params = append(params, arr[i])
			}
			conditions = append(conditions, fmt.Sprintf("%s IN (%s)", fieldName, strings.Join(placeholders, ", ")))
		case "_nin":
			arr, ok := value.([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("_nin value must be an array")
			}
			placeholders := make([]string, len(arr))
			for i := range arr {
				placeholders[i] = "?"
				params = append(params, arr[i])
			}
			conditions = append(conditions, fmt.Sprintf("%s NOT IN (%s)", fieldName, strings.Join(placeholders, ", ")))
		}
	}

	if len(conditions) == 0 {
		return "1=1", nil, nil
	}

	return strings.Join(conditions, " AND "), params, nil
}

func (c *PolicyCompiler) compileValue(fieldName string, value interface{}) (string, interface{}) {
	switch v := value.(type) {
	case map[string]interface{}:
		// 特殊值引用，如 {"_auth": "uid"}
		if authVar, ok := v["_auth"].(string); ok {
			return fmt.Sprintf("%s = ?", fieldName), map[string]string{"_auth": authVar}
		}
		if ref, ok := v["_ref"].(string); ok {
			return fmt.Sprintf("%s = %s", fieldName, ref), nil
		}
		return fmt.Sprintf("%s = ?", fieldName), v
	default:
		return fmt.Sprintf("%s = ?", fieldName), v
	}
}

// Ensure PolicyCompiler implements rls.PolicyCompiler
var _ rls.PolicyCompiler = (*PolicyCompiler)(nil)
