package rls

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/rls"
	"strings"
)

// RLS expression operator constants (backward-compatible).
const (
	oldOpAnd   = "_and"
	oldOpOr    = "_or"
	oldOpNot   = "_not"
	oldOpConst = "_const"
	oldOpAuth  = "_auth"
)

// PolicyCompiler RLS 表达式编译器实现
type PolicyCompiler struct{}

// NewPolicyCompiler 创建 PolicyCompiler
func NewPolicyCompiler() *PolicyCompiler {
	return &PolicyCompiler{}
}

// Compile 将 JSON 表达式编译为 CompiledPolicy
// 支持新旧两种操作符命名，支持 {{user_id}} / {{user_name}} 变量替换
func (c *PolicyCompiler) Compile(ctx context.Context, expr rls.JsonExpr, userCtx *rls.UserContext) (*rls.CompiledPolicy, error) {
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

	return c.compileNode(ctx, obj, userCtx)
}

func (c *PolicyCompiler) compileNode(ctx context.Context, node map[string]interface{}, userCtx *rls.UserContext) (*rls.CompiledPolicy, error) {
	var conditions []string
	var params []interface{}

	for key, value := range node {
		switch c.normalizeOp(key) {
		case "AND":
			arr, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("AND value must be an array")
			}
			var andConds []string
			for _, item := range arr {
				itemObj, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("AND array item must be an object")
				}
				compiled, err := c.compileNode(ctx, itemObj, userCtx)
				if err != nil {
					return nil, err
				}
				andConds = append(andConds, compiled.SQL)
				params = append(params, compiled.Params...)
			}
			if len(andConds) > 0 {
				conditions = append(conditions, "("+strings.Join(andConds, " AND ")+")")
			}

		case "OR":
			arr, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("OR value must be an array")
			}
			var orConds []string
			for _, item := range arr {
				itemObj, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("OR array item must be an object")
				}
				compiled, err := c.compileNode(ctx, itemObj, userCtx)
				if err != nil {
					return nil, err
				}
				orConds = append(orConds, compiled.SQL)
				params = append(params, compiled.Params...)
			}
			if len(orConds) > 0 {
				conditions = append(conditions, "("+strings.Join(orConds, " OR ")+")")
			}

		case "NOT":
			obj, ok := value.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("NOT value must be an object")
			}
			compiled, err := c.compileNode(ctx, obj, userCtx)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, "NOT ("+compiled.SQL+")")
			params = append(params, compiled.Params...)

		case oldOpConst:
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
				// 尝试简单布尔常量
				if b, ok := value.(bool); ok {
					if b {
						conditions = append(conditions, "1=1")
					} else {
						conditions = append(conditions, "1=0")
					}
					continue
				}
				return nil, fmt.Errorf("field comparison must be an object for field %q", key)
			}
			fieldCond, fieldParams, err := c.compileFieldComparison(key, compObj, userCtx)
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

// normalizeOp 统一逻辑操作符名（兼容旧 _and/_or/_not 和新 AND/OR/NOT）
func (c *PolicyCompiler) normalizeOp(op string) string {
	switch op {
	case oldOpAnd:
		return "AND"
	case oldOpOr:
		return "OR"
	case oldOpNot:
		return "NOT"
	default:
		return op
	}
}

func (c *PolicyCompiler) compileFieldComparison(
	fieldName string, compObj map[string]interface{}, userCtx *rls.UserContext,
) (string, []interface{}, error) {
	// 处理旧 _auth 引用: {"owner_id": {"_auth": "uid"}} → owner_id = ?
	if authVar, ok := compObj[oldOpAuth].(string); ok {
		resolved := userCtx.ResolveVariable(authVar)
		return fmt.Sprintf("%s = ?", fieldName), []interface{}{resolved}, nil
	}

	var conditions []string
	var params []interface{}

	for op, value := range compObj {
		switch c.normalizeFieldOp(op) {
		case "equals":
			if value == nil {
				conditions = append(conditions, fmt.Sprintf("%s IS NULL", fieldName))
			} else {
				resolved := c.resolveValue(value, userCtx)
				conditions = append(conditions, fmt.Sprintf("%s = ?", fieldName))
				params = append(params, resolved)
			}

		case "not":
			resolved := c.resolveValue(value, userCtx)
			conditions = append(conditions, fmt.Sprintf("%s != ?", fieldName))
			params = append(params, resolved)

		case "gt":
			conditions = append(conditions, fmt.Sprintf("%s > ?", fieldName))
			params = append(params, c.resolveValue(value, userCtx))

		case "gte":
			conditions = append(conditions, fmt.Sprintf("%s >= ?", fieldName))
			params = append(params, c.resolveValue(value, userCtx))

		case "lt":
			conditions = append(conditions, fmt.Sprintf("%s < ?", fieldName))
			params = append(params, c.resolveValue(value, userCtx))

		case "lte":
			conditions = append(conditions, fmt.Sprintf("%s <= ?", fieldName))
			params = append(params, c.resolveValue(value, userCtx))

		case "in":
			arr, ok := value.([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("in value must be an array")
			}
			placeholders := make([]string, len(arr))
			for i := range arr {
				placeholders[i] = "?"
				params = append(params, c.resolveValue(arr[i], userCtx))
			}
			conditions = append(conditions, fmt.Sprintf("%s IN (%s)", fieldName, strings.Join(placeholders, ", ")))

		case "nin":
			arr, ok := value.([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("nin value must be an array")
			}
			placeholders := make([]string, len(arr))
			for i := range arr {
				placeholders[i] = "?"
				params = append(params, c.resolveValue(arr[i], userCtx))
			}
			conditions = append(conditions, fmt.Sprintf("%s NOT IN (%s)", fieldName, strings.Join(placeholders, ", ")))

		case "isNull":
			if b, ok := value.(bool); ok && b {
				conditions = append(conditions, fmt.Sprintf("%s IS NULL", fieldName))
			} else {
				conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", fieldName))
			}

		case "contains":
			resolved := c.resolveValue(value, userCtx)
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
			params = append(params, "%"+fmt.Sprint(resolved)+"%")

		case "startsWith":
			resolved := c.resolveValue(value, userCtx)
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
			params = append(params, fmt.Sprint(resolved)+"%")

		case "endsWith":
			resolved := c.resolveValue(value, userCtx)
			conditions = append(conditions, fmt.Sprintf("%s LIKE ?", fieldName))
			params = append(params, "%"+fmt.Sprint(resolved))

		default:
			return "", nil, fmt.Errorf("unknown field operator: %s", op)
		}
	}

	if len(conditions) == 0 {
		return "1=1", nil, nil
	}

	return strings.Join(conditions, " AND "), params, nil
}

// normalizeFieldOp 统一字段操作符名（兼容旧 _eq/_neq/... 和新 equals/not/...）
func (c *PolicyCompiler) normalizeFieldOp(op string) string {
	switch op {
	case "_eq":
		return "equals"
	case "_neq":
		return "not"
	case "_gt":
		return "gt"
	case "_gte":
		return "gte"
	case "_lt":
		return "lt"
	case "_lte":
		return "lte"
	case "_in":
		return "in"
	case "_nin":
		return "nin"
	case "_is_null":
		return "isNull"
	default:
		return op
	}
}

// resolveValue 解析值中的变量占位符
// 支持 {{user_id}} / {{user_name}} 字符串替换和 old _auth 引用
func (c *PolicyCompiler) resolveValue(value interface{}, userCtx *rls.UserContext) interface{} {
	if userCtx == nil {
		return value
	}
	switch v := value.(type) {
	case string:
		// {{user_id}} / {{user_name}} 替换
		v = strings.ReplaceAll(v, "{{user_id}}", userCtx.UserID)
		v = strings.ReplaceAll(v, "{{user_name}}", userCtx.UserName)
		return v
	case map[string]interface{}:
		// 兼容旧 _auth 引用: {"_auth": "uid"} → UserContext 解析
		if authVar, ok := v["_auth"].(string); ok {
			return userCtx.ResolveVariable(authVar)
		}
		return v
	default:
		return v
	}
}

// Ensure PolicyCompiler implements rls.PolicyCompiler
var _ rls.PolicyCompiler = (*PolicyCompiler)(nil)
