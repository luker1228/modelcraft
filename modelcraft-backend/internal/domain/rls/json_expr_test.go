package rls

import "testing"

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
