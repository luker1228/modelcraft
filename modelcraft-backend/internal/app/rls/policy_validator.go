package rls

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/rls"
	"strconv"
)

// PolicyValidator PolicyValidator 实现
type PolicyValidator struct{}

// NewPolicyValidator 创建 PolicyValidator
func NewPolicyValidator() *PolicyValidator {
	return &PolicyValidator{}
}

// Validate 校验 JSON 表达式合法性
func (v *PolicyValidator) Validate(ctx context.Context, expr rls.JsonExpr,
	exprType rls.ExprType, modelSchema rls.ModelSchema,
	authSchema rls.AuthSchemaProvider,
) []rls.ValidationError {
	var errors []rls.ValidationError

	// 1. JSON 合法性校验
	var root interface{}
	if err := json.Unmarshal([]byte(expr), &root); err != nil {
		return []rls.ValidationError{{
			Path:    "",
			Message: "Invalid JSON: " + err.Error(),
			Code:    "INVALID_JSON",
		}}
	}

	// 2. 常量简写 true/false 直接通过
	if _, ok := root.(bool); ok {
		return nil
	}

	// 3. 递归校验
	obj, ok := root.(map[string]interface{})
	if !ok {
		return []rls.ValidationError{{
			Path:    "",
			Message: "Expression must be an object or boolean constant",
			Code:    "INVALID_STRUCTURE",
		}}
	}

	errors = append(errors, v.validateNode(ctx, obj, "", exprType, modelSchema, authSchema)...)

	return errors
}

//nolint:gocognit,funlen // recursive validation over expression tree is intentionally explicit
func (v *PolicyValidator) validateNode(
	ctx context.Context, node map[string]interface{},
	path string, exprType rls.ExprType, modelSchema rls.ModelSchema,
	authSchema rls.AuthSchemaProvider,
) []rls.ValidationError {
	var errors []rls.ValidationError

	for key, value := range node {
		currentPath := path
		if currentPath != "" {
			currentPath += "."
		}
		currentPath += key

		switch key {
		case "_and", "_or":
			// 逻辑操作符：值必须是数组
			arr, ok := value.([]interface{})
			if !ok {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_and/_or value must be an array",
					Code:    "INVALID_OPERATOR",
				})
				continue
			}
			for i, item := range arr {
				itemObj, ok := item.(map[string]interface{})
				if !ok {
					errors = append(errors, rls.ValidationError{
						Path:    currentPath + "[" + strconv.Itoa(i) + "]",
						Message: "Array item must be an object",
						Code:    "INVALID_ARRAY_ITEM",
					})
					continue
				}
				errors = append(errors, v.validateNode(ctx, itemObj,
					currentPath+"["+strconv.Itoa(i)+"]", exprType, modelSchema, authSchema)...)
			}

		case "_not":
			// _not: 值必须是对象
			obj, ok := value.(map[string]interface{})
			if !ok {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_not value must be an object",
					Code:    "INVALID_OPERATOR",
				})
				continue
			}
			errors = append(errors, v.validateNode(ctx, obj, currentPath, exprType, modelSchema, authSchema)...)

		case "_exists":
			// _exists: 仅 PREDICATE 允许，CHECK 不允许
			if exprType.IsCheck() {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_exists is not allowed in CHECK expressions (insertCheck/updateCheck)",
					Code:    "EXISTS_IN_CHECK",
				})
				continue
			}
			// 校验 _exists 结构
			existsObj, ok := value.(map[string]interface{})
			if !ok {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_exists value must be an object with 'model' and 'where'",
					Code:    "INVALID_EXISTS",
				})
				continue
			}
			// 校验 model 字段
			if modelName, ok := existsObj["model"].(string); ok {
				if modelName == "" {
					errors = append(errors, rls.ValidationError{
						Path:    currentPath + ".model",
						Message: "model name cannot be empty",
						Code:    "EMPTY_MODEL_NAME",
					})
				}
			} else {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_exists must have a 'model' field of type string",
					Code:    "MISSING_MODEL",
				})
			}

		case "_auth":
			// _auth: 校验变量名是否在白名单
			varName, ok := value.(string)
			if !ok {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_auth value must be a string",
					Code:    "INVALID_AUTH_REF",
				})
				continue
			}
			if !authSchema.IsValidRef(varName) {
				errors = append(errors, rls.ValidationError{
					Path: currentPath,
					Message: fmt.Sprintf(
						"Unknown auth variable '%s'. Declare it in project auth_schema first.",
						varName,
					),
					Code: "UNKNOWN_AUTH_VAR",
				})
			}

		case "_ref":
			// _ref: 仅 PREDICATE 允许
			if exprType.IsCheck() {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_ref is not allowed in CHECK expressions (insertCheck/updateCheck)",
					Code:    "REF_IN_CHECK",
				})
				continue
			}
			// 校验 _ref 格式
			refValue, ok := value.(string)
			if !ok || refValue == "" {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_ref value must be a non-empty string",
					Code:    "INVALID_REF",
				})
			}

		case "_const":
			// _const: 必须是布尔值
			if _, ok := value.(bool); !ok {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "_const value must be a boolean",
					Code:    "INVALID_CONST",
				})
			}

		default:
			// 字段比较：校验字段名是否存在
			if !modelSchema.HasField(key) {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: fmt.Sprintf("Unknown field '%s'", key),
					Code:    "UNKNOWN_FIELD",
				})
				continue
			}
			// 校验比较操作符
			compObj, ok := value.(map[string]interface{})
			if !ok {
				errors = append(errors, rls.ValidationError{
					Path:    currentPath,
					Message: "Field comparison value must be an object",
					Code:    "INVALID_COMPARISON",
				})
				continue
			}
			// 校验支持的比较操作符
			validOps := map[string]bool{
				"_eq": true, "_neq": true, "_gt": true, "_gte": true,
				"_lt": true, "_lte": true, "_in": true, "_nin": true, "_is_null": true,
			}
			for op := range compObj {
				if !validOps[op] {
					errors = append(errors, rls.ValidationError{
						Path:    currentPath + "." + op,
						Message: fmt.Sprintf("Unknown comparison operator '%s'", op),
						Code:    "UNKNOWN_OPERATOR",
					})
				}
			}
		}
	}

	return errors
}

// Ensure PolicyValidator implements rls.PolicyValidator
var _ rls.PolicyValidator = (*PolicyValidator)(nil)
