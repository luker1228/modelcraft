package modelruntime_test

import (
	"context"
	"testing"

	appruntimeimport "modelcraft/internal/app/modelruntime"
	"modelcraft/internal/domain/rbac"
)

// stubRBACRepo minimal stub for testing — only implements FindPermissionsByEndUserAndModel
// and satisfies the rest of the interface with no-ops.
type stubRBACRepo struct {
	permissions []*rbac.EndUserPermission
}

func (s *stubRBACRepo) FindPermissionsByEndUserAndModel(_ context.Context, _, _, _, _ string) ([]*rbac.EndUserPermission, error) {
	return s.permissions, nil
}
func (s *stubRBACRepo) CreatePermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}
func (s *stubRBACRepo) GetPermissionByID(_ context.Context, _, _ string) (*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListPermissionsByProject(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListPermissionsByModel(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListPresetPermissionsByModel(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetPermissionByModelTypeName(_ context.Context, _, _ string, _ rbac.PermissionType, _ string) (*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpdatePermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}
func (s *stubRBACRepo) DeletePermission(_ context.Context, _, _ string) error { return nil }
func (s *stubRBACRepo) DeletePresetPermissionsByModel(_ context.Context, _, _ string) error {
	return nil
}
func (s *stubRBACRepo) UpdatePresetPermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}
func (s *stubRBACRepo) IsPermissionReferencedByBundle(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (s *stubRBACRepo) CreateBundle(_ context.Context, _ *rbac.EndUserPermissionBundle) error {
	return nil
}
func (s *stubRBACRepo) GetBundleByID(_ context.Context, _, _, _ string) (*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleBySlug(_ context.Context, _, _, _ string) (*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListBundlesByProject(_ context.Context, _, _ string) ([]*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpdateBundle(_ context.Context, _ *rbac.EndUserPermissionBundle) error {
	return nil
}
func (s *stubRBACRepo) DeleteBundle(_ context.Context, _, _ string) error { return nil }
func (s *stubRBACRepo) AddPermissionToBundle(_ context.Context, _, _ string, _ int) error {
	return nil
}
func (s *stubRBACRepo) RemovePermissionFromBundle(_ context.Context, _, _ string) error { return nil }
func (s *stubRBACRepo) ListPermissionsInBundle(_ context.Context, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpsertBundleDataPermissionItem(_ context.Context, _ *rbac.EndUserBundleDataPermissionItem) error {
	return nil
}
func (s *stubRBACRepo) RemoveBundleDataPermissionItem(_ context.Context, _, _ string) error {
	return nil
}
func (s *stubRBACRepo) ListBundleDataPermissionItems(_ context.Context, _ string) ([]*rbac.EndUserBundleDataPermissionItem, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleDataPermissionItemByBundleAndModel(_ context.Context, _, _ string) (*rbac.EndUserBundleDataPermissionItem, error) {
	return nil, nil
}
func (s *stubRBACRepo) SaveBundleSnapshot(_ context.Context, _ *rbac.BundleSnapshot) error {
	return nil
}
func (s *stubRBACRepo) ListBundleSnapshots(_ context.Context, _ string) ([]rbac.BundleSnapshot, error) {
	return nil, nil
}
func (s *stubRBACRepo) DeleteOldBundleSnapshots(_ context.Context, _ string) error { return nil }
func (s *stubRBACRepo) GetBundleCurrentVersion(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (s *stubRBACRepo) GetBundleSnapshotByVersion(_ context.Context, _ string, _ int) (*rbac.BundleSnapshot, error) {
	return nil, nil
}
func (s *stubRBACRepo) ClearBundlePermissions(_ context.Context, _ string) error { return nil }
func (s *stubRBACRepo) CreateRole(_ context.Context, _ *rbac.EndUserRole) error  { return nil }
func (s *stubRBACRepo) GetRoleByID(_ context.Context, _, _ string) (*rbac.EndUserRole, error) {
	return nil, nil
}
func (s *stubRBACRepo) ListRolesByProject(_ context.Context, _, _ string) ([]*rbac.EndUserRole, error) {
	return nil, nil
}
func (s *stubRBACRepo) UpdateRole(_ context.Context, _ *rbac.EndUserRole) error { return nil }
func (s *stubRBACRepo) DeleteRole(_ context.Context, _, _ string) error         { return nil }
func (s *stubRBACRepo) AssignBundleToRole(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) RevokeBundleFromRole(_ context.Context, _, _ string) error     { return nil }
func (s *stubRBACRepo) ListBundlesByRole(_ context.Context, _ string) ([]*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubRBACRepo) GrantBundleToUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) RevokeBundleFromUser(_ context.Context, _, _, _, _ string) error {
	return nil
}
func (s *stubRBACRepo) AssignRoleToUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) RevokeRoleFromUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubRBACRepo) ListProjectEndUserRoleUsers(_ context.Context, _ rbac.ListProjectEndUserRoleUsersQuery) ([]*rbac.ProjectEndUserRoleUser, int64, error) {
	return nil, 0, nil
}
func (s *stubRBACRepo) GetBundleIDsByUserDirect(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleIDsByUserExplicitRoles(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetBundleIDsByImplicitRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}
func (s *stubRBACRepo) GetPermissionsByBundleIDs(_ context.Context, _ string, _ []string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

// makeRowPolicy creates a RowPolicy with the given settings and normalizes it.
func makeRowPolicy(
	selectAllowed bool, selectScope rbac.PolicyScope,
	insertAllowed bool, insertScope rbac.PolicyScope,
	updateAllowed bool, updateScope rbac.PolicyScope,
	deleteAllowed bool, deleteScope rbac.PolicyScope,
) *rbac.RowPolicy {
	rp := &rbac.RowPolicy{
		Select: rbac.SelectPolicy{Allowed: selectAllowed, Scope: selectScope},
		Insert: rbac.InsertPolicy{Allowed: insertAllowed, Scope: insertScope},
		Update: rbac.UpdatePolicy{Allowed: updateAllowed, Scope: updateScope},
		Delete: rbac.DeletePolicy{Allowed: deleteAllowed, Scope: deleteScope},
	}
	rp.Normalize()
	return rp
}

func TestEndUserPermissionService_Resolve_TenantAdmin(t *testing.T) {
	svc := appruntimeimport.NewEndUserPermissionService(&stubRBACRepo{})
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if perms != nil {
		t.Error("tenant admin (empty endUserID) should return nil permissions")
	}
}

func TestEndUserPermissionService_Resolve_NoPermissions(t *testing.T) {
	svc := appruntimeimport.NewEndUserPermissionService(&stubRBACRepo{permissions: nil})
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if perms == nil {
		t.Fatal("expected non-nil perms (empty = all denied)")
	}
	if perms.Select.Allowed || perms.Insert.Allowed || perms.Update.Allowed || perms.Delete.Allowed {
		t.Error("no rbac permissions should result in all denied")
	}
}

func TestEndUserPermissionService_Resolve_SelectAll(t *testing.T) {
	rp := makeRowPolicy(true, rbac.ScopeAll, false, "", false, "", false, "")
	stub := &stubRBACRepo{
		permissions: []*rbac.EndUserPermission{
			{ModelID: "model-id", RowPolicy: rp},
		},
	}
	svc := appruntimeimport.NewEndUserPermissionService(stub)
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Select.Allowed {
		t.Error("expected Select.Allowed = true")
	}
	if perms.Select.IsSelf {
		t.Error("scope=all should produce IsSelf=false")
	}
	if perms.Insert.Allowed {
		t.Error("expected Insert.Allowed = false")
	}
}

func TestEndUserPermissionService_Resolve_SelectSelfInsertSelf(t *testing.T) {
	rp := makeRowPolicy(true, rbac.ScopeCustom, true, rbac.ScopeCustom, false, "", false, "")
	stub := &stubRBACRepo{
		permissions: []*rbac.EndUserPermission{
			{ModelID: "model-id", RowPolicy: rp},
		},
	}
	svc := appruntimeimport.NewEndUserPermissionService(stub)
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if !perms.Select.IsSelf {
		t.Error("scope=custom should produce IsSelf=true for Select")
	}
	if !perms.Insert.Allowed {
		t.Error("expected Insert.Allowed = true")
	}
	if !perms.Insert.IsSelf {
		t.Error("scope=custom should produce IsSelf=true for Insert")
	}
}
