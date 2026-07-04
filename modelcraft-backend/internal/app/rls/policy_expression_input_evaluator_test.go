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
		`input.owner_id == auth.userid && input.status in ["draft", "pending"]`,
		map[string]any{"owner_id": "u_123", "status": "draft"},
		&domainrls.UserContext{UserIDStr: "u_123"},
	)
	require.NoError(t, err)
}

func TestPolicyExpressionInputEvaluator_DeniesMismatchedInput(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.owner_id == auth.userid`,
		map[string]any{"owner_id": "u_999"},
		&domainrls.UserContext{UserIDStr: "u_123"},
	)
	require.ErrorContains(t, err, "RLS CHECK violation")
}

func TestPolicyExpressionInputEvaluator_UpdateUsesPatchOnly(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.status == "draft"`,
		map[string]any{"title": "renamed"},
		&domainrls.UserContext{UserIDStr: "u_123"},
	)
	require.ErrorContains(t, err, "no such key")
}

// WITH CHECK: auth 字段覆盖

func TestPolicyExpressionInputEvaluator_NumericUserID(t *testing.T) {
	// auth.userid 数值型时，CEL 能与 input 的 int 字段正确比较
	evaluator := NewPolicyExpressionInputEvaluator()
	uid := int64(42)
	err := evaluator.ValidateInput(
		context.Background(),
		`input.user_id == auth.userid`,
		map[string]any{"user_id": int64(42)},
		&domainrls.UserContext{UserIDNum: &uid},
	)
	require.NoError(t, err)
}

func TestPolicyExpressionInputEvaluator_AuthUsername(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.author == auth.username`,
		map[string]any{"author": "alice"},
		&domainrls.UserContext{UserIDStr: "u_1", UserName: "alice"},
	)
	require.NoError(t, err)
}

func TestPolicyExpressionInputEvaluator_RoleCheck(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`"admin" in auth.roles`,
		map[string]any{},
		&domainrls.UserContext{UserIDStr: "u_1", Roles: []string{"editor", "admin"}},
	)
	require.NoError(t, err)
}

func TestPolicyExpressionInputEvaluator_RoleCheckDenied(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`"admin" in auth.roles`,
		map[string]any{},
		&domainrls.UserContext{UserIDStr: "u_1", Roles: []string{"viewer"}},
	)
	require.ErrorContains(t, err, "RLS CHECK violation")
}

// WITH CHECK: string 方法（cel-go 原生支持，与 USING SQL 编译路径不同）

func TestPolicyExpressionInputEvaluator_StartsWith(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.slug.startsWith("proj-")`,
		map[string]any{"slug": "proj-alpha"},
		&domainrls.UserContext{UserIDStr: "u_1"},
	)
	require.NoError(t, err)
}

func TestPolicyExpressionInputEvaluator_StartsWithDenied(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.slug.startsWith("proj-")`,
		map[string]any{"slug": "other-alpha"},
		&domainrls.UserContext{UserIDStr: "u_1"},
	)
	require.ErrorContains(t, err, "RLS CHECK violation")
}

// WITH CHECK: 空表达式 / 编译错误

func TestPolicyExpressionInputEvaluator_EmptyExpressionDenied(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		``,
		map[string]any{},
		&domainrls.UserContext{UserIDStr: "u_1"},
	)
	require.ErrorContains(t, err, "RLS CHECK violation")
}

func TestPolicyExpressionInputEvaluator_InvalidExpressionErrors(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(
		context.Background(),
		`input.owner_id ===`,
		map[string]any{"owner_id": "u_1"},
		&domainrls.UserContext{UserIDStr: "u_1"},
	)
	require.ErrorContains(t, err, "RLS CHECK violation")
}
