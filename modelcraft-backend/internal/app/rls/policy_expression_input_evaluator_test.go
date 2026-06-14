package rls

import (
	"context"
	"testing"

	domainrls "modelcraft/internal/domain/rls"

	"github.com/stretchr/testify/require"
)

func TestPolicyExpressionInputEvaluator_AllowsMatchingCreateInput(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.owner_id == auth.user_id && input.status in ["draft", "pending"]`,
		map[string]any{"owner_id": "u_123", "status": "draft"},
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.NoError(t, err)
}

func TestPolicyExpressionInputEvaluator_DeniesMismatchedInput(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.owner_id == auth.user_id`,
		map[string]any{"owner_id": "u_999"},
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.ErrorContains(t, err, "RLS CHECK violation")
}

func TestPolicyExpressionInputEvaluator_UpdateUsesPatchOnly(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.status == "draft"`,
		map[string]any{"title": "renamed"},
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.ErrorContains(t, err, "no such key")
}
