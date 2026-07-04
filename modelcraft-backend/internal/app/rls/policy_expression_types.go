package rls

type PolicyExpressionMode string

const (
	PolicyExpressionModeUsing PolicyExpressionMode = "using"
	PolicyExpressionModeCheck PolicyExpressionMode = "check"
)

func (m PolicyExpressionMode) AllowsRoot(root string) bool {
	switch m {
	case PolicyExpressionModeUsing:
		return root == celVarRow || root == celVarAuth
	case PolicyExpressionModeCheck:
		return root == celVarInput || root == celVarAuth
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
