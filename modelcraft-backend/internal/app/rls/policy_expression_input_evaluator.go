package rls

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"

	domainrls "modelcraft/internal/domain/rls"
)

type PolicyExpressionInputEvaluator struct {
	env *cel.Env
}

func NewPolicyExpressionInputEvaluator() *PolicyExpressionInputEvaluator {
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		panic(fmt.Sprintf("create CEL env: %v", err))
	}

	return &PolicyExpressionInputEvaluator{env: env}
}

func (e *PolicyExpressionInputEvaluator) ValidateInput(
	_ context.Context,
	expr string,
	input map[string]any,
	userCtx *domainrls.UserContext,
) error {
	if expr == "" {
		return fmt.Errorf("RLS CHECK violation: empty input check")
	}

	if IsLegacyJSONExpression(expr) {
		return NewPolicyExecutor().ValidateCheck(context.Background(), domainrls.JsonExpr(expr), input, &domainrls.AuthContext{
			EndUserID: resolveUserID(userCtx),
			Variables: buildAuthEvalContext(userCtx),
		})
	}

	ast, issues := e.env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("RLS CHECK violation: %w", issues.Err())
	}

	program, err := e.env.Program(ast)
	if err != nil {
		return fmt.Errorf("RLS CHECK violation: %w", err)
	}

	out, _, err := program.Eval(map[string]any{
		"input": input,
		"auth":  buildAuthEvalContext(userCtx),
	})
	if err != nil {
		return fmt.Errorf("RLS CHECK violation: %w", err)
	}

	allowed, ok := out.Value().(bool)
	if !ok {
		return fmt.Errorf("RLS CHECK violation: expression returned %T", out.Value())
	}
	if !allowed {
		return fmt.Errorf("RLS CHECK violation: expression evaluated to false")
	}
	return nil
}

func buildAuthEvalContext(userCtx *domainrls.UserContext) map[string]any {
	if userCtx == nil {
		return map[string]any{
			"userid":   "",
			"username": "",
			"roles":    []string{},
		}
	}
	return map[string]any{
		"userid":   userCtx.UserIDValue(),
		"username": userCtx.UserName,
		"roles":    userCtx.Roles,
	}
}

func resolveUserID(userCtx *domainrls.UserContext) string {
	if userCtx == nil {
		return ""
	}
	return fmt.Sprint(userCtx.UserIDValue())
}
