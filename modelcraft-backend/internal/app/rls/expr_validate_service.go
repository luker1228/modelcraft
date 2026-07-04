package rls

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modeldesign"

	domainrls "modelcraft/internal/domain/rls"
)

// RLSExprValidateService handles RLS expression validation and dry-run.
// It is the V2 replacement for the dead ModelRLSPolicyAppService.ValidateExpr/DryRunExpr.
type RLSExprValidateService struct {
	modelRepo modeldesign.ModelRepository
}

// NewRLSExprValidateService creates a new RLSExprValidateService.
func NewRLSExprValidateService(modelRepo modeldesign.ModelRepository) *RLSExprValidateService {
	return &RLSExprValidateService{modelRepo: modelRepo}
}

// ValidateAndDryRun validates an RLS expression and optionally produces a dry-run result.
// For predicate expressions (using/row.*): compiles CEL -> SQL, returns SQL preview.
// For check expressions (input.*): evaluates CEL with sampleInput, returns boolean.
func (s *RLSExprValidateService) ValidateAndDryRun(
	ctx context.Context,
	_, _ string, // orgName, projectSlug — reserved for future use
	modelID string,
	exprType domainrls.ExprType,
	expression string,
	sampleInput map[string]any,
	userCtx *domainrls.UserContext,
) PolicyExpressionDryRunResult {
	// 1. Check model exists (provides field whitelist context)
	model, err := s.modelRepo.GetByID(ctx, modelID)
	if err != nil || model == nil {
		return PolicyExpressionDryRunResult{
			Valid: false,
			Errors: []PolicyExpressionError{{
				Path:    "modelId",
				Message: "Model not found",
				Code:    "MODEL_NOT_FOUND",
			}},
		}
	}
	_ = model // model existence check only; field whitelist is enforced by the compiler

	// 2. Dry-run: for predicates, compile CEL -> SQL; for checks, evaluate CEL.
	result := PolicyExpressionDryRunResult{Valid: true}

	if exprType.IsPredicate() {
		compiled, err := NewPolicyExpressionSQLCompiler().CompileUsing(ctx, expression, userCtx)
		if err != nil {
			result.Valid = false
			result.Errors = []PolicyExpressionError{{
				Message: err.Error(),
				Code:    "DRY_RUN_FAILED",
			}}
			return result
		}
		result.SQL = compiled.SQL
		result.Params = compiled.Params
		return result
	}

	// Check expression: evaluate with sampleInput
	if sampleInput == nil {
		sampleInput = map[string]any{}
	}
	err = NewPolicyExpressionInputEvaluator().ValidateInput(ctx, expression, sampleInput, userCtx)
	checkResult := err == nil
	result.Result = &checkResult
	if err != nil {
		// "evaluated to false" is a valid dry-run result, not an error.
		// Only non-evaluation failures (compile errors, type errors) set Valid=false.
		if isCheckFalseError(err) {
			return result
		}
		result.Valid = false
		result.Errors = []PolicyExpressionError{{
			Message: fmt.Sprintf("%v", err),
			Code:    "DRY_RUN_FAILED",
		}}
	}
	return result
}

// isCheckFalseError returns true if the error is a CHECK expression evaluating
// to false (a legitimate dry-run outcome) rather than a compile/eval failure.
func isCheckFalseError(err error) bool {
	msg := err.Error()
	return msg == "RLS CHECK violation: expression evaluated to false"
}
