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
		&domainrls.UserContext{UserIDStr: "u_123"},
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
		&domainrls.UserContext{UserIDStr: "u_123"},
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
		&domainrls.UserContext{UserIDStr: "u_123"},
	)
	require.ErrorContains(t, err, "input is not allowed")
}

// 大小比较：> / >= / < / <= 映射到对应 SQL 运算符，参数类型保持原始数值类型。
// 注意：CEL 数字字面量无小数点时解析为 int64，带小数点时解析为 float64。
func TestPolicyExpressionSQLCompiler_ComparisonOperators(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	ctx := context.Background()
	userCtx := &domainrls.UserContext{UserIDStr: "u_1"}

	tests := []struct {
		name         string
		expr         string
		wantSQL      string
		wantParams   []interface{}
	}{
		{
			name:       "greater than integer",
			expr:       `row.age > 18`,
			wantSQL:    "age > ?",
			wantParams: []interface{}{int64(18)},
		},
		{
			name:       "greater than or equal integer",
			expr:       `row.score >= 60`,
			wantSQL:    "score >= ?",
			wantParams: []interface{}{int64(60)},
		},
		{
			name:       "less than integer",
			expr:       `row.priority < 5`,
			wantSQL:    "priority < ?",
			wantParams: []interface{}{int64(5)},
		},
		{
			name:       "less than or equal integer",
			expr:       `row.rank <= 100`,
			wantSQL:    "rank <= ?",
			wantParams: []interface{}{int64(100)},
		},
		{
			name:       "greater than float",
			expr:       `row.ratio > 0.5`,
			wantSQL:    "ratio > ?",
			wantParams: []interface{}{float64(0.5)},
		},
		{
			name:       "range check with AND",
			// 闭区间 [18, 65]：两个比较通过 && 组合
			expr:       `row.age >= 18 && row.age <= 65`,
			wantSQL:    "(age >= ? AND age <= ?)",
			wantParams: []interface{}{int64(18), int64(65)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compiled, err := compiler.CompileUsing(ctx, tc.expr, userCtx)
			require.NoError(t, err)
			require.Equal(t, tc.wantSQL, compiled.SQL)
			require.Equal(t, tc.wantParams, compiled.Params)
		})
	}
}

func TestPolicyExpressionSQLCompiler_NotEqualAndNullChecks(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	ctx := context.Background()
	userCtx := &domainrls.UserContext{UserIDStr: "u_1"}

	tests := []struct {
		name       string
		expr       string
		wantSQL    string
		wantParams []interface{}
	}{
		{
			name:       "not equal string",
			expr:       `row.status != "deleted"`,
			wantSQL:    "status <> ?",
			wantParams: []interface{}{"deleted"},
		},
		{
			name:       "equal null becomes IS NULL",
			expr:       `row.deleted_at == null`,
			wantSQL:    "deleted_at IS NULL",
			wantParams: nil,
		},
		{
			name:       "not equal null becomes IS NOT NULL",
			expr:       `row.deleted_at != null`,
			wantSQL:    "deleted_at IS NOT NULL",
			wantParams: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compiled, err := compiler.CompileUsing(ctx, tc.expr, userCtx)
			require.NoError(t, err)
			require.Equal(t, tc.wantSQL, compiled.SQL)
			require.Equal(t, tc.wantParams, compiled.Params)
		})
	}
}

func TestPolicyExpressionSQLCompiler_EmptyInList(t *testing.T) {
	// 空 in 列表应生成恒假条件 "1=0"，确保不产生语法错误的 IN ()
	compiler := NewPolicyExpressionSQLCompiler()
	compiled, err := compiler.CompileUsing(
		context.Background(),
		`row.status in []`,
		&domainrls.UserContext{UserIDStr: "u_1"},
	)
	require.NoError(t, err)
	require.Equal(t, "1=0", compiled.SQL)
	require.Empty(t, compiled.Params)
}

func TestPolicyExpressionSQLCompiler_NumericUserID(t *testing.T) {
	// auth.userid 使用数值型 UserIDNum 时，参数应为 int64，而非字符串
	compiler := NewPolicyExpressionSQLCompiler()
	uid := int64(42)
	compiled, err := compiler.CompileUsing(
		context.Background(),
		`row.user_id == auth.userid`,
		&domainrls.UserContext{UserIDNum: &uid},
	)
	require.NoError(t, err)
	require.Equal(t, "user_id = ?", compiled.SQL)
	require.Equal(t, []interface{}{int64(42)}, compiled.Params)
}

func TestPolicyExpressionSQLCompiler_AuthUsername(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	compiled, err := compiler.CompileUsing(
		context.Background(),
		`row.author == auth.username`,
		&domainrls.UserContext{UserIDStr: "u_1", UserName: "alice"},
	)
	require.NoError(t, err)
	require.Equal(t, "author = ?", compiled.SQL)
	require.Equal(t, []interface{}{"alice"}, compiled.Params)
}
