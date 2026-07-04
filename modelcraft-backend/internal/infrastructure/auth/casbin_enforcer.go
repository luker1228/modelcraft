package auth

import (
	"context"
	"fmt"
	"modelcraft/pkg/logfacade"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

var (
	enforcer     *casbin.Enforcer
	enforcerOnce sync.Once
	enforcerErr  error
)

// GetEnforcer returns the singleton Casbin enforcer instance.
// The enforcer is initialized once with the RBAC model and system role permissions.
func GetEnforcer() (*casbin.Enforcer, error) {
	enforcerOnce.Do(func() {
		enforcer, enforcerErr = initializeEnforcer()
	})
	return enforcer, enforcerErr
}

// initializeEnforcer creates and configures the Casbin enforcer.
func initializeEnforcer() (*casbin.Enforcer, error) {
	logger := logfacade.GetLogger(context.Background())

	// Load Casbin model from string (embedded configuration)
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (p.obj == "*" || r.obj == p.obj) && (p.act == "*" || r.act == p.act)
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		logger.Errorf(context.Background(), err, "Failed to load Casbin model")
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	// Create enforcer without adapter (we'll manage policies programmatically)
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		logger.Errorf(context.Background(), err, "Failed to create Casbin enforcer")
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Note: Casbin v2 has caching enabled by default for performance optimization

	logger.Infof(context.Background(), "Casbin enforcer initialized successfully")

	return e, nil
}

// LoadSystemRolePermissions loads hardcoded system role permissions into the enforcer.
// This should be called during application startup after the enforcer is initialized.
func LoadSystemRolePermissions(e *casbin.Enforcer) error {
	logger := logfacade.GetLogger(context.Background())

	// Clear existing policies to ensure clean state
	e.ClearPolicy()

	// Load system role permissions
	for roleName, permissions := range SystemRolePermissions {
		for _, perm := range permissions {
			// Add policy: role -> obj -> act
			_, err := e.AddPolicy(roleName, perm.Obj, perm.Act)
			if err != nil {
				logger.Errorf(context.Background(), err, "Failed to add policy for role %s", roleName)
				return fmt.Errorf("failed to add policy for role %s: %w", roleName, err)
			}
		}
		logger.Infof(context.Background(), "Loaded %d permissions for system role: %s", len(permissions), roleName)
	}

	logger.Infof(context.Background(), "Successfully loaded all system role permissions into Casbin enforcer")
	return nil
}

// AddUserRole assigns a role to a user in the enforcer.
// This creates a role mapping: user -> role
func AddUserRole(e *casbin.Enforcer, userID, roleName string) error {
	_, err := e.AddGroupingPolicy(userID, roleName)
	if err != nil {
		return fmt.Errorf("failed to add user role mapping: %w", err)
	}
	return nil
}

// RemoveUserRole removes a role assignment from a user in the enforcer.
func RemoveUserRole(e *casbin.Enforcer, userID, roleName string) error {
	_, err := e.RemoveGroupingPolicy(userID, roleName)
	if err != nil {
		return fmt.Errorf("failed to remove user role mapping: %w", err)
	}
	return nil
}

// AddCustomRolePermission adds a permission to a custom role in the enforcer.
func AddCustomRolePermission(e *casbin.Enforcer, roleName, obj, act string) error {
	_, err := e.AddPolicy(roleName, obj, act)
	if err != nil {
		return fmt.Errorf("failed to add custom role permission: %w", err)
	}
	return nil
}

// RemoveCustomRolePermission removes a permission from a custom role in the enforcer.
func RemoveCustomRolePermission(e *casbin.Enforcer, roleName, obj, act string) error {
	_, err := e.RemovePolicy(roleName, obj, act)
	if err != nil {
		return fmt.Errorf("failed to remove custom role permission: %w", err)
	}
	return nil
}

// CheckPermission checks if a user has permission to perform an action on an object.
func CheckPermission(e *casbin.Enforcer, userID, obj, act string) (bool, error) {
	allowed, err := e.Enforce(userID, obj, act)
	if err != nil {
		return false, fmt.Errorf("permission check failed: %w", err)
	}
	return allowed, nil
}
