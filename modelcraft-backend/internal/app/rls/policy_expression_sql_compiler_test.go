package rls

import (
	"context"
	"testing"

	domainrls "modelcraft/internal/domain/rls"

	"github.com/stretchr/testify/require"
)

func TestPolicyExpressionSQLCompiler_CompilesEqualityAndIn(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	compiled, err := compiler.CompileUsing(
		context.Background(),
		`row.owner_id == auth.userid && row.status in ["draft", "pending"]`,
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.NoError(t, err)
	require.Equal(t, "(owner_id = ? AND status IN (?, ?))", compiled.SQL)
	require.Equal(t, []interface{}{"u_123", "draft", "pending"}, compiled.Params)
}

func TestPolicyExpressionSQLCompiler_CompilesOrAndNot(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	compiled, err := compiler.CompileUsing(
		context.Background(),
		`row.owner_id == auth.userid || !(row.status == "archived")`,
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.NoError(t, err)
	require.Equal(t, "(owner_id = ? OR NOT (status = ?))", compiled.SQL)
	require.Equal(t, []interface{}{"u_123", "archived"}, compiled.Params)
}

func TestPolicyExpressionSQLCompiler_RejectsInputRoot(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	_, err := compiler.CompileUsing(
		context.Background(),
		`input.owner_id == auth.userid`,
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.ErrorContains(t, err, "input is not allowed")
}
