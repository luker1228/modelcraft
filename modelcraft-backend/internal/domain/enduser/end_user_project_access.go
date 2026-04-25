package enduser

import (
	"fmt"
	"strings"
	"time"
)

// EndUserProjectAccess represents an end-user's access grant in a project.
type EndUserProjectAccess struct {
	ID                 string
	OrgName            string
	ProjectSlug        string
	EndUserID          string
	PermissionBundleID string
	PermissionName     string
	GrantedBy          string
	GrantedAt          time.Time
	EndUser            *EndUser
}

// NewEndUserProjectAccess creates a new access grant with validation.
func NewEndUserProjectAccess(
	id, orgName, projectSlug, endUserID, permissionBundleID, grantedBy string,
) (*EndUserProjectAccess, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("access ID is required")
	}
	if strings.TrimSpace(orgName) == "" {
		return nil, fmt.Errorf("org name is required")
	}
	if strings.TrimSpace(projectSlug) == "" {
		return nil, fmt.Errorf("project slug is required")
	}
	if strings.TrimSpace(endUserID) == "" {
		return nil, fmt.Errorf("end user ID is required")
	}
	if strings.TrimSpace(permissionBundleID) == "" {
		return nil, fmt.Errorf("permission bundle ID is required")
	}

	now := time.Now()
	return &EndUserProjectAccess{
		ID:                 id,
		OrgName:            orgName,
		ProjectSlug:        projectSlug,
		EndUserID:          endUserID,
		PermissionBundleID: permissionBundleID,
		GrantedBy:          strings.TrimSpace(grantedBy),
		GrantedAt:          now,
	}, nil
}

// UpdatePermissionBundle updates the permission bundle bound to this access grant.
func (a *EndUserProjectAccess) UpdatePermissionBundle(permissionBundleID string) error {
	if strings.TrimSpace(permissionBundleID) == "" {
		return fmt.Errorf("permission bundle ID is required")
	}
	a.PermissionBundleID = permissionBundleID
	return nil
}
