package rls

// FixedAuthContext provides the built-in auth variables supported by runtime RLS.
// New expressions should use auth.userid/auth.username/auth.roles.
// Legacy aliases remain accepted in backend for compatibility with persisted policies.
type FixedAuthContext struct{}

// DefaultAuthContext returns the fixed auth context provider used by runtime RLS.
func DefaultAuthContext() FixedAuthContext {
	return FixedAuthContext{}
}

// IsValidRef reports whether an auth variable is supported.
func (FixedAuthContext) IsValidRef(name string) bool {
	switch name {
	case "userid", "username", "roles":
		return true
	case "uid", "user_id", "user_name":
		return true
	default:
		return false
	}
}
