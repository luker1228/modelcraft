package rls

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	domainrls "modelcraft/internal/domain/rls"
)

var allowedCELCallFunctions = map[string]struct{}{
	"_&&_": {},
	"_||_": {},
	"!_":   {},
	"_==_": {},
	"_!=_": {},
	"_>_":  {},
	"_>=_": {},
	"_<_":  {},
	"_<=_": {},
	"@in":  {},
	"_in_": {},
}

type PolicyExpressionValidator struct {
	legacy *PolicyValidator
	env    *cel.Env
}

func NewPolicyExpressionValidator() *PolicyExpressionValidator {
	env, err := cel.NewEnv(
		cel.Variable("row", cel.DynType),
		cel.Variable("input", cel.DynType),
		cel.Variable("auth", cel.DynType),
	)
	if err != nil {
		panic(fmt.Sprintf("create CEL env: %v", err))
	}

	return &PolicyExpressionValidator{
		legacy: NewPolicyValidator(),
		env:    env,
	}
}

func (v *PolicyExpressionValidator) Validate(
	ctx context.Context,
	expr domainrls.JsonExpr,
	exprType domainrls.ExprType,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []domainrls.ValidationError {
	if IsLegacyJSONExpression(string(expr)) {
		return v.legacy.Validate(ctx, expr, exprType, modelSchema, authSchema)
	}

	mode, ok := modeForExprType(exprType)
	if !ok {
		return []domainrls.ValidationError{{
			Message: fmt.Sprintf("unsupported expression type %q", exprType),
			Code:    "UNSUPPORTED_EXPR_TYPE",
		}}
	}

	celErrors := v.ValidateCEL(ctx, mode, string(expr), modelSchema, authSchema)
	if len(celErrors) == 0 {
		return nil
	}

	errors := make([]domainrls.ValidationError, 0, len(celErrors))
	for _, err := range celErrors {
		errors = append(errors, domainrls.ValidationError{
			Path:    err.Path,
			Message: err.Message,
			Code:    err.Code,
		})
	}
	return errors
}

func (v *PolicyExpressionValidator) ValidateCEL(
	_ context.Context,
	mode PolicyExpressionMode,
	expr string,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []PolicyExpressionError {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return []PolicyExpressionError{{
			Message: "expression cannot be empty",
			Code:    "INVALID_EXPRESSION",
		}}
	}

	parsed, issues := v.env.Parse(trimmed)
	if issues != nil && issues.Err() != nil {
		return []PolicyExpressionError{{
			Message: issues.Err().Error(),
			Code:    "SYNTAX_ERROR",
		}}
	}

	checked, issues := v.env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return []PolicyExpressionError{{
			Message: issues.Err().Error(),
			Code:    "TYPE_ERROR",
		}}
	}

	if checked.OutputType() != cel.BoolType {
		return []PolicyExpressionError{{
			Message: "policy expression must return bool",
			Code:    "NON_BOOLEAN_RESULT",
		}}
	}

	parsedExpr, err := cel.AstToParsedExpr(parsed)
	if err != nil {
		return []PolicyExpressionError{{
			Message: err.Error(),
			Code:    "INVALID_EXPRESSION",
		}}
	}

	return v.validateExpr(mode, parsedExpr.GetExpr(), modelSchema, authSchema)
}

func (v *PolicyExpressionValidator) validateExpr(
	mode PolicyExpressionMode,
	expr *exprpb.Expr,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []PolicyExpressionError {
	if expr == nil {
		return nil
	}

	switch node := expr.ExprKind.(type) {
	case *exprpb.Expr_ConstExpr:
		return nil
	case *exprpb.Expr_IdentExpr:
		return v.validateIdent(mode, node.IdentExpr.GetName())
	case *exprpb.Expr_SelectExpr:
		return v.validateSelect(mode, expr, modelSchema, authSchema)
	case *exprpb.Expr_CallExpr:
		return v.validateCall(mode, node.CallExpr, modelSchema, authSchema)
	case *exprpb.Expr_ListExpr:
		var errors []PolicyExpressionError
		for _, elem := range node.ListExpr.GetElements() {
			errors = append(errors, v.validateExpr(mode, elem, modelSchema, authSchema)...)
		}
		return errors
	default:
		return []PolicyExpressionError{{
			Message: "unsupported CEL expression node",
			Code:    "UNSUPPORTED_EXPR",
		}}
	}
}

func (v *PolicyExpressionValidator) validateIdent(
	mode PolicyExpressionMode,
	name string,
) []PolicyExpressionError {
	if mode.AllowsRoot(name) {
		return nil
	}

	return []PolicyExpressionError{{
		Message: fmt.Sprintf("%q is not available in %s expressions", name, mode),
		Code:    "INVALID_CONTEXT",
	}}
}

func (v *PolicyExpressionValidator) validateSelect(
	mode PolicyExpressionMode,
	expr *exprpb.Expr,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []PolicyExpressionError {
	root, fields, ok := selectPath(expr)
	if !ok {
		return []PolicyExpressionError{{
			Message: "unsupported field access pattern",
			Code:    "UNSUPPORTED_EXPR",
		}}
	}

	if !mode.AllowsRoot(root) {
		return []PolicyExpressionError{{
			Path:    root,
			Message: fmt.Sprintf("%q is not available in %s expressions", root, mode),
			Code:    "INVALID_CONTEXT",
		}}
	}

	if len(fields) == 0 {
		return nil
	}

	switch root {
	case "row", "input":
		if !modelSchema.HasField(fields[0]) {
			return []PolicyExpressionError{{
				Path:    root + "." + fields[0],
				Message: fmt.Sprintf("Unknown field '%s'", fields[0]),
				Code:    "UNKNOWN_FIELD",
			}}
		}
	case "auth":
		if !authSchema.IsValidRef(fields[0]) {
			return []PolicyExpressionError{{
				Path:    root + "." + fields[0],
				Message: fmt.Sprintf("Unknown auth variable '%s'. Supported variables: auth.userid, auth.username, auth.roles.", fields[0]),
				Code:    "UNKNOWN_AUTH_VAR",
			}}
		}
	}

	return nil
}

func (v *PolicyExpressionValidator) validateCall(
	mode PolicyExpressionMode,
	call *exprpb.Expr_Call,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []PolicyExpressionError {
	if _, ok := allowedCELCallFunctions[call.GetFunction()]; !ok {
		return []PolicyExpressionError{{
			Message: fmt.Sprintf("function %q is not supported in policy expressions", call.GetFunction()),
			Code:    "UNSUPPORTED_CALL",
		}}
	}

	var errors []PolicyExpressionError
	if target := call.GetTarget(); target != nil {
		errors = append(errors, v.validateExpr(mode, target, modelSchema, authSchema)...)
	}
	for _, arg := range call.GetArgs() {
		errors = append(errors, v.validateExpr(mode, arg, modelSchema, authSchema)...)
	}
	return errors
}

func modeForExprType(exprType domainrls.ExprType) (PolicyExpressionMode, bool) {
	switch exprType {
	case domainrls.ExprTypeSelectPredicate, domainrls.ExprTypeUpdatePredicate, domainrls.ExprTypeDeletePredicate:
		return PolicyExpressionModeUsing, true
	case domainrls.ExprTypeInsertCheck, domainrls.ExprTypeUpdateCheck:
		return PolicyExpressionModeCheck, true
	default:
		return "", false
	}
}

func selectPath(expr *exprpb.Expr) (string, []string, bool) {
	selectExpr := expr.GetSelectExpr()
	if selectExpr == nil {
		ident := expr.GetIdentExpr()
		if ident == nil {
			return "", nil, false
		}
		return ident.GetName(), nil, true
	}

	root, fields, ok := selectPath(selectExpr.GetOperand())
	if !ok {
		return "", nil, false
	}
	return root, append(fields, selectExpr.GetField()), true
}

var _ domainrls.PolicyValidator = (*PolicyExpressionValidator)(nil)
