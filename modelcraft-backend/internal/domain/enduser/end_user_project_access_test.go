package enduser

import "testing"

func TestNewEndUserProjectAccess(t *testing.T) {
	t.Run("should create access", func(t *testing.T) {
		access, err := NewEndUserProjectAccess(
			"access-1",
			"org-a",
			"project-a",
			"user-1",
			"bundle-1",
			"creator-1",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if access.ID != "access-1" {
			t.Fatalf("unexpected id: %s", access.ID)
		}
		if access.OrgName != "org-a" {
			t.Fatalf("unexpected org name: %s", access.OrgName)
		}
		if access.ProjectSlug != "project-a" {
			t.Fatalf("unexpected project slug: %s", access.ProjectSlug)
		}
		if access.EndUserID != "user-1" {
			t.Fatalf("unexpected end user id: %s", access.EndUserID)
		}
		if access.PermissionBundleID != "bundle-1" {
			t.Fatalf("unexpected bundle id: %s", access.PermissionBundleID)
		}
	})

	t.Run("should reject empty required fields", func(t *testing.T) {
		cases := []struct {
			name             string
			id               string
			orgName          string
			projectSlug      string
			endUserID        string
			permissionBundle string
		}{
			{name: "id", id: "", orgName: "org", projectSlug: "project", endUserID: "user", permissionBundle: "bundle"},
			{name: "org", id: "id", orgName: "", projectSlug: "project", endUserID: "user", permissionBundle: "bundle"},
			{name: "project", id: "id", orgName: "org", projectSlug: "", endUserID: "user", permissionBundle: "bundle"},
			{name: "user", id: "id", orgName: "org", projectSlug: "project", endUserID: "", permissionBundle: "bundle"},
			{name: "bundle", id: "id", orgName: "org", projectSlug: "project", endUserID: "user", permissionBundle: ""},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := NewEndUserProjectAccess(
					tc.id,
					tc.orgName,
					tc.projectSlug,
					tc.endUserID,
					tc.permissionBundle,
					"creator",
				)
				if err == nil {
					t.Fatalf("expected validation error")
				}
			})
		}
	})
}

func TestEndUserProjectAccess_UpdatePermissionBundle(t *testing.T) {
	access, err := NewEndUserProjectAccess("access-1", "org-a", "project-a", "user-1", "bundle-1", "creator-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := access.UpdatePermissionBundle("bundle-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if access.PermissionBundleID != "bundle-2" {
		t.Fatalf("unexpected bundle id: %s", access.PermissionBundleID)
	}

	if err := access.UpdatePermissionBundle(""); err == nil {
		t.Fatalf("expected validation error")
	}
}
