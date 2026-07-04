package middleware

import "context"

// IsOrgAdminFn checks whether an end-user has org-admin status.
// Used by the PAT whoami handler to report admin status in the whoami response.
type IsOrgAdminFn func(ctx context.Context, orgName, userID string) (bool, error)
