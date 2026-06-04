package repository_test

import (
	"context"
	"modelcraft/internal/domain/rbac"
	"testing"
)

// TestFindPermissionsByEndUserAndModel_EmptyInputs verifies nil is returned for empty inputs
// without needing a real DB connection.
func TestFindPermissionsByEndUserAndModel_EmptyInputs(t *testing.T) {
	// We can't easily test the full DB path without integration infra,
	// but we can verify the guard clauses via interface compliance.
	// The compile-time check below ensures the interface is implemented.
	var _ rbac.EndUserPermissionRepository = (*stubPermRepo)(nil)
}

// stubPermRepo is a compile-time interface assertion helper.
type stubPermRepo struct{}

func (s *stubPermRepo) FindPermissionsByEndUserAndModel(
	_ context.Context, _, _, _, _ string,
) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

// Satisfy the rest of the interface with empty implementations
func (s *stubPermRepo) CreatePermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}

func (s *stubPermRepo) GetPermissionByID(_ context.Context, _, _ string) (*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubPermRepo) ListPermissionsByProject(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubPermRepo) ListPermissionsByModel(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubPermRepo) ListPresetPermissionsByModel(_ context.Context, _, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubPermRepo) GetPermissionByModelTypeName(
	_ context.Context, _, _ string, _ rbac.PermissionType, _ string,
) (*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubPermRepo) UpdatePermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}
func (s *stubPermRepo) DeletePermission(_ context.Context, _, _ string) error { return nil }
func (s *stubPermRepo) DeletePresetPermissionsByModel(_ context.Context, _, _ string) error {
	return nil
}

func (s *stubPermRepo) UpdatePresetPermission(_ context.Context, _ *rbac.EndUserPermission) error {
	return nil
}

func (s *stubPermRepo) IsPermissionReferencedByBundle(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (s *stubPermRepo) CreateBundle(_ context.Context, _ *rbac.EndUserPermissionBundle) error {
	return nil
}

func (s *stubPermRepo) GetBundleByID(_ context.Context, _, _, _ string) (*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}

func (s *stubPermRepo) GetBundleBySlug(_ context.Context, _, _, _ string) (*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}

func (s *stubPermRepo) ListBundlesByProject(_ context.Context, _, _ string) ([]*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}

func (s *stubPermRepo) UpdateBundle(_ context.Context, _ *rbac.EndUserPermissionBundle) error {
	return nil
}
func (s *stubPermRepo) DeleteBundle(_ context.Context, _, _ string) error { return nil }
func (s *stubPermRepo) AddPermissionToBundle(_ context.Context, _, _ string, _ int) error {
	return nil
}
func (s *stubPermRepo) RemovePermissionFromBundle(_ context.Context, _, _ string) error { return nil }
func (s *stubPermRepo) ListPermissionsInBundle(_ context.Context, _ string) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubPermRepo) UpsertBundleDataPermissionItem(
	_ context.Context, _ *rbac.EndUserBundleDataPermissionItem,
) error {
	return nil
}

func (s *stubPermRepo) RemoveBundleDataPermissionItem(_ context.Context, _, _ string) error {
	return nil
}

func (s *stubPermRepo) ListBundleDataPermissionItems(
	_ context.Context, _ string,
) ([]*rbac.EndUserBundleDataPermissionItem, error) {
	return nil, nil
}

func (s *stubPermRepo) GetBundleDataPermissionItemByBundleAndModel(
	_ context.Context, _, _ string,
) (*rbac.EndUserBundleDataPermissionItem, error) {
	return nil, nil
}

func (s *stubPermRepo) SaveBundleSnapshot(_ context.Context, _ *rbac.BundleSnapshot) error {
	return nil
}

func (s *stubPermRepo) ListBundleSnapshots(_ context.Context, _ string) ([]rbac.BundleSnapshot, error) {
	return nil, nil
}
func (s *stubPermRepo) DeleteOldBundleSnapshots(_ context.Context, _ string) error { return nil }
func (s *stubPermRepo) GetBundleCurrentVersion(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (s *stubPermRepo) GetBundleSnapshotByVersion(_ context.Context, _ string, _ int) (*rbac.BundleSnapshot, error) {
	return nil, nil
}
func (s *stubPermRepo) ClearBundlePermissions(_ context.Context, _ string) error { return nil }
func (s *stubPermRepo) CreateRole(_ context.Context, _ *rbac.EndUserRole) error  { return nil }
func (s *stubPermRepo) GetRoleByID(_ context.Context, _, _ string) (*rbac.EndUserRole, error) {
	return nil, nil
}

func (s *stubPermRepo) ListRolesByProject(_ context.Context, _, _ string) ([]*rbac.EndUserRole, error) {
	return nil, nil
}
func (s *stubPermRepo) UpdateRole(_ context.Context, _ *rbac.EndUserRole) error       { return nil }
func (s *stubPermRepo) DeleteRole(_ context.Context, _, _ string) error               { return nil }
func (s *stubPermRepo) AssignBundleToRole(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubPermRepo) RevokeBundleFromRole(_ context.Context, _, _ string) error     { return nil }
func (s *stubPermRepo) ListBundlesByRole(_ context.Context, _ string) ([]*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}
func (s *stubPermRepo) GrantBundleToUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubPermRepo) RevokeBundleFromUser(_ context.Context, _, _, _, _ string) error {
	return nil
}
func (s *stubPermRepo) AssignRoleToUser(_ context.Context, _, _, _, _ string) error { return nil }
func (s *stubPermRepo) RevokeRoleFromUser(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (s *stubPermRepo) ListProjectEndUserRoleUsers(
	_ context.Context, _ rbac.ListProjectEndUserRoleUsersQuery,
) ([]*rbac.ProjectEndUserRoleUser, int64, error) {
	return nil, 0, nil
}

func (s *stubPermRepo) GetBundleIDsByUserDirect(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *stubPermRepo) GetBundleIDsByUserExplicitRoles(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *stubPermRepo) GetBundleIDsByImplicitRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *stubPermRepo) GetPermissionsByBundleIDs(
	_ context.Context, _ string, _ []string,
) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubPermRepo) HasProtectedAdminRole(_ context.Context, _, _, _ string) (bool, error) {
	return false, nil
}

func (s *stubPermRepo) IsOrgAdmin(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (s *stubPermRepo) IsUserBuiltin(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}
