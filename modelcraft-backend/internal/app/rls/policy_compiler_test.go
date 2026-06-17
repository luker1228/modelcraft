package rls

import (
	"context"
	"modelcraft/internal/domain/rls"
	"testing"
)

func TestCompile_Equals_VariableSubstitution(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{"tenant_id": {"equals": "{{user_id}}"}}`)
	userCtx := &rls.UserContext{UserIDStr: "customer_123"}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "tenant_id = ?"
	if result.SQL != expected {
		t.Errorf("expected SQL %q, got %q", expected, result.SQL)
	}
	if len(result.Params) != 1 || result.Params[0] != "customer_123" {
		t.Errorf("expected params [customer_123], got %v", result.Params)
	}
}

func TestCompile_AND_With_Two_Fields(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{
		"AND": [
			{"tenant_id": {"equals": "{{user_id}}"}},
			{"status": {"equals": "active"}}
		]
	}`)
	userCtx := &rls.UserContext{UserIDStr: "123"}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "(tenant_id = ? AND status = ?)"
	if result.SQL != expected {
		t.Errorf("expected SQL %q, got %q", expected, result.SQL)
	}
	if len(result.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(result.Params))
	}
}

func TestCompile_Contains(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{"name": {"contains": "{{user_name}}"}}`)
	userCtx := &rls.UserContext{UserName: "zhangsan"}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SQL != "name LIKE ?" {
		t.Errorf("expected 'name LIKE ?', got %q", result.SQL)
	}
	if len(result.Params) != 1 || result.Params[0] != "%zhangsan%" {
		t.Errorf("expected params [%%zhangsan%%], got %v", result.Params)
	}
}

func TestCompile_OldOp_BackwardCompatible(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{
		"_and": [
			{"field_a": {"_eq": "value1"}},
			{"field_b": {"_gt": 10}}
		]
	}`)
	userCtx := &rls.UserContext{}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "(field_a = ? AND field_b > ?)"
	if result.SQL != expected {
		t.Errorf("expected SQL %q, got %q", expected, result.SQL)
	}
	if len(result.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(result.Params))
	}
}

func TestCompile_BooleanTrue(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	result, err := compiler.Compile(ctx, rls.JsonExpr(`true`), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SQL != "1=1" {
		t.Errorf("expected '1=1', got %q", result.SQL)
	}
}

func TestCompile_BooleanFalse(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	result, err := compiler.Compile(ctx, rls.JsonExpr(`false`), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SQL != "1=0" {
		t.Errorf("expected '1=0', got %q", result.SQL)
	}
}

func TestCompile_OR_With_Two_Conditions(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{
		"OR": [
			{"role": {"equals": "admin"}},
			{"role": {"equals": "manager"}}
		]
	}`)
	userCtx := &rls.UserContext{}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "(role = ? OR role = ?)"
	if result.SQL != expected {
		t.Errorf("expected SQL %q, got %q", expected, result.SQL)
	}
}

func TestCompile_StartsWith_EndsWith(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{
		"AND": [
			{"name": {"startsWith": "pre_"}},
			{"name": {"endsWith": "_suffix"}}
		]
	}`)
	userCtx := &rls.UserContext{}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "(name LIKE ? AND name LIKE ?)"
	if result.SQL != expected {
		t.Errorf("expected SQL %q, got %q", expected, result.SQL)
	}
	if len(result.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(result.Params))
	}
	if result.Params[0] != "pre_%" {
		t.Errorf("expected params[0]='pre_%%', got %v", result.Params[0])
	}
	if result.Params[1] != "%_suffix" {
		t.Errorf("expected params[1]='%%_suffix', got %v", result.Params[1])
	}
}

func TestCompile_EqualsNull(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{"deleted_at": {"equals": null}}`)
	userCtx := &rls.UserContext{}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SQL != "deleted_at IS NULL" {
		t.Errorf("expected 'deleted_at IS NULL', got %q", result.SQL)
	}
}

func TestCompile_InArray(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{"status": {"in": ["active", "draft", "pending"]}}`)
	userCtx := &rls.UserContext{}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SQL != "status IN (?, ?, ?)" {
		t.Errorf("expected 'status IN (?, ?, ?)', got %q", result.SQL)
	}
	if len(result.Params) != 3 {
		t.Errorf("expected 3 params, got %d", len(result.Params))
	}
}

func TestCompile_NilUserContext(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{"field": {"equals": "{{user_id}}"}}`)
	// userCtx is nil

	result, err := compiler.Compile(ctx, expr, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without UserContext, {{user_id}} should NOT be replaced
	if result.SQL != "field = ?" {
		t.Errorf("expected 'field = ?', got %q", result.SQL)
	}
}

func TestCompile_OldAuthReference(t *testing.T) {
	compiler := NewPolicyCompiler()
	ctx := context.Background()

	expr := rls.JsonExpr(`{"owner_id": {"_auth": "uid"}}`)
	userCtx := &rls.UserContext{UserIDStr: "customer_456"}

	result, err := compiler.Compile(ctx, expr, userCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "owner_id = ?"
	if result.SQL != expected {
		t.Errorf("expected SQL %q, got %q", expected, result.SQL)
	}
	if len(result.Params) != 1 || result.Params[0] != "customer_456" {
		t.Errorf("expected params [customer_456], got %v", result.Params)
	}
}
