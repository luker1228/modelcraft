package middleware

import (
	"strings"
)

// CheckPermission checks if a permission list grants the required permission.
// Supports three matching modes:
//  1. Global wildcard: "*" matches any permission
//  2. Resource wildcard: "resource:*" matches any action on that resource
//  3. Exact match: "resource:action" matches exactly
func CheckPermission(userPermissions []string, required string) bool {
	for _, perm := range userPermissions {
		if perm == "*" {
			return true
		}
		if perm == required {
			return true
		}
		if strings.HasSuffix(perm, ":*") {
			permResource := strings.TrimSuffix(perm, ":*")
			requiredParts := strings.SplitN(required, ":", 2)
			if len(requiredParts) == 2 && requiredParts[0] == permResource {
				return true
			}
		}
	}
	return false
}
