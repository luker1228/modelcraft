package rls

import "testing"

func TestJsonExprIsTrue(t *testing.T) {
	tests := []struct {
		name     string
		expr     JsonExpr
		expected bool
	}{
		{"true bool", JsonExpr(`true`), true},
		{"false bool", JsonExpr(`false`), false},
		{"_const true", JsonExpr(`{"_const":true}`), true},
		{"_const false", JsonExpr(`{"_const":false}`), false},
		{"complex expr", JsonExpr(`{"owner":{"_eq":"123"}}`), false},
		{"invalid json", JsonExpr(`invalid`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expr.IsTrue()
			if result != tt.expected {
				t.Errorf("IsTrue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJsonExprIsFalse(t *testing.T) {
	tests := []struct {
		name     string
		expr     JsonExpr
		expected bool
	}{
		{"true bool", JsonExpr(`true`), false},
		{"false bool", JsonExpr(`false`), true},
		{"_const true", JsonExpr(`{"_const":true}`), false},
		{"_const false", JsonExpr(`{"_const":false}`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expr.IsFalse()
			if result != tt.expected {
				t.Errorf("IsFalse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJsonExprIsOwnerEqualsUser(t *testing.T) {
	tests := []struct {
		name     string
		expr     JsonExpr
		expected bool
	}{
		{"owner equals uid", JsonExpr(`{"owner":{"_eq":{"_auth":"uid"}}}`), true},
		{"other field", JsonExpr(`{"other":{"_eq":{"_auth":"uid"}}}`), false},
		{"not auth", JsonExpr(`{"owner":{"_eq":{"_auth":"other"}}}`), false},
		{"not eq", JsonExpr(`{"owner":{"_neq":{"_auth":"uid"}}}`), false},
		{"complex", JsonExpr(`{"_and":[{"owner":{"_eq":{"_auth":"uid"}}}]]`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expr.IsOwnerEqualsUser()
			if result != tt.expected {
				t.Errorf("IsOwnerEqualsUser() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExprTypeIsPredicate(t *testing.T) {
	tests := []struct {
		exprType ExprType
		expected bool
	}{
		{ExprTypeSelectPredicate, true},
		{ExprTypeUpdatePredicate, true},
		{ExprTypeDeletePredicate, true},
		{ExprTypeInsertCheck, false},
		{ExprTypeUpdateCheck, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.exprType), func(t *testing.T) {
			result := tt.exprType.IsPredicate()
			if result != tt.expected {
				t.Errorf("IsPredicate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExprTypeIsCheck(t *testing.T) {
	tests := []struct {
		exprType ExprType
		expected bool
	}{
		{ExprTypeInsertCheck, true},
		{ExprTypeUpdateCheck, true},
		{ExprTypeSelectPredicate, false},
		{ExprTypeUpdatePredicate, false},
		{ExprTypeDeletePredicate, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.exprType), func(t *testing.T) {
			result := tt.exprType.IsCheck()
			if result != tt.expected {
				t.Errorf("IsCheck() = %v, want %v", result, tt.expected)
			}
		})
	}
}
