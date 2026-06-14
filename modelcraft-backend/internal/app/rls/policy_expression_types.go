package rls

import "strings"

type PolicyExpressionMode string

const (
	PolicyExpressionModeUsing PolicyExpressionMode = "using"
	PolicyExpressionModeCheck PolicyExpressionMode = "check"
)

func (m PolicyExpressionMode) AllowsRoot(root string) bool {
	switch m {
	case PolicyExpressionModeUsing:
		return root == "row" || root == "auth"
	case PolicyExpressionModeCheck:
		return root == "input" || root == "auth"
	default:
		return false
	}
}

type PolicyExpressionDryRunResult struct {
	Valid  bool
	SQL    string
	Params []any
	Result *bool
	Errors []PolicyExpressionError
}

type PolicyExpressionError struct {
	Path    string
	Message string
	Code    string
}

func IsLegacyJSONExpression(expr string) bool {
	trimmed := strings.TrimSpace(expr)
	return trimmed == "true" || trimmed == "false" || strings.HasPrefix(trimmed, "{")
}
