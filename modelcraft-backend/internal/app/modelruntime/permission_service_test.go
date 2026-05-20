package modelruntime_test

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/rbac"
	"testing"

	appruntimeimport "modelcraft/internal/app/modelruntime"
)

// ─── stubModelRepo ────────────────────────────────────────────────────────────

// stubModelRepo 用于测试的 modeldesign.ModelRepository stub，仅实现 GetByID，其余 no-op。
type stubModelRepo struct {
	models map[string]*modeldesign.DataModel
}

func (s *stubModelRepo) GetByID(
	_ context.Context, id string, _ ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	if m, ok := s.models[id]; ok {
		return m, nil
	}
	return nil, nil
}

func (s *stubModelRepo) Save(_ context.Context, _ string, _ *modeldesign.DataModel) error {
	return nil
}

func (s *stubModelRepo) Update(_ context.Context, _ *modeldesign.DataModel) error { return nil }

func (s *stubModelRepo) UpdateWithVersion(_ context.Context, _ *modeldesign.DataModel, _ int64) (int64, error) {
	return 0, nil
}

func (s *stubModelRepo) Delete(_ context.Context, _ string) error { return nil }

func (s *stubModelRepo) GetByName(
	_ context.Context, _, _, _, _ string, _ ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	return nil, nil
}

func (s *stubModelRepo) FindByDeploymentStatus(
	_ context.Context, _ ...modeldesign.DeploymentStatus,
) ([]modeldesign.DataModel, error) {
	return nil, nil
}

func (s *stubModelRepo) GetMetaByIDs(_ context.Context, _, _ string, _ []string) ([]*modeldesign.DataModel, error) {
	return nil, nil
}

func (s *stubModelRepo) Query(_ context.Context, _ modeldesign.ModelQuery) ([]modeldesign.DataModel, int, error) {
	return nil, 0, nil
}

func (s *stubModelRepo) ListDatabaseCatalog(_ context.Context, _, _, _ string, _, _ int) ([]string, int, error) {
	return nil, 0, nil
}

func (s *stubModelRepo) AddFields(_ context.Context, _ string, _ []*modeldesign.FieldDefinition) error {
	return nil
}

func (s *stubModelRepo) AddRelationField(_ context.Context, _ string, _ *modeldesign.FieldDefinition) error {
	return nil
}

func (s *stubModelRepo) GetFieldByModelID(_ context.Context, _, _ string) (*modeldesign.FieldDefinition, error) {
	return nil, nil
}

func (s *stubModelRepo) GetFieldsByModelID(_ context.Context, _ string) ([]*modeldesign.FieldDefinition, error) {
	return nil, nil
}

func (s *stubModelRepo) GetTailFieldDisplayOrder(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (s *stubModelRepo) UpdateField(_ context.Context, _ *modeldesign.FieldDefinition) error {
	return nil
}

func (s *stubModelRepo) BulkUpdateFields(_ context.Context, _ []*modeldesign.FieldDefinition) error {
	return nil
}

func (s *stubModelRepo) UpdateFieldsStatus(_ context.Context, _ ...modeldesign.UpdateFieldsStatusRequest) error {
	return nil
}

func (s *stubModelRepo) DeleteFields(_ context.Context, _ string, _ []string) error { return nil }

func (s *stubModelRepo) BulkDeleteFields(_ context.Context, _ ...modeldesign.DeleteFieldRequest) error {
	return nil
}

// ─── stubRBACRepo ─────────────────────────────────────────────────────────────

// stubRBACRepo 用于测试的 rbac.EndUserPermissionRepository stub。
// bundleIDsByExplicitRoles: GetBundleIDsByUserExplicitRoles 返回的 bundle ID 列表。
// customPerms: GetPermissionsByBundleIDs 返回的 CUSTOM 权限点（模拟 bundle 展开结果）。
// bundleItems: bundleID → 该 bundle 下的 PRESET data permission items。
type stubRBACRepo struct {
	bundleIDsByExplicitRoles []string
	customPerms              []*rbac.EndUserPermission
	bundleItems              map[string][]*rbac.EndUserBundleDataPermissionItem
}

func (s *stubRBACRepo) FindPermissionsByEndUserAndModel(
	_ context.Context, _, _, _, _ string,
) ([]*rbac.EndUserPermission, error) {
	return nil, nil
}

func (s *stubRBACRepo) GetBundleIDsByUserDirect(_ context.Context, _, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *stubRBACRepo) GetBundleIDsByUserExplicitRoles(_ context.Context, _, _, _ string) ([]string, error) {
	return s.bundleIDsByExplicitRoles, nil
}

func (s *stubRBACRepo) GetBundleIDsByImplicitRoles(_ context.Context, _, _ string) ([]string, error) {
	return nil, nil
}

func (s *stubRBACRepo) GetPermissionsByBundleIDs(
	_ context.Context, _ string, _ []string,
) ([]*rbac.EndUserPermission, error) {
	return s.customPerms, nil
}

func (s *stubRBACRepo) ListBundleDataPermissionItems(
	_ context.Context, bundleID string,
) ([]*rbac.EndUserBundleDataPermissionItem, error) {
	if s.bundleItems != nil {
		return s.bundleItems[bundleID], nil
	}
	return nil, nil
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

func (s *stubRBACRepo) GetPermissionByModelTypeName(
	_ context.Context, _, _ string, _ rbac.PermissionType, _ string,
) (*rbac.EndUserPermission, error) {
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

func (s *stubRBACRepo) UpsertBundleDataPermissionItem(
	_ context.Context, _ *rbac.EndUserBundleDataPermissionItem,
) error {
	return nil
}

func (s *stubRBACRepo) RemoveBundleDataPermissionItem(_ context.Context, _, _ string) error {
	return nil
}

func (s *stubRBACRepo) GetBundleDataPermissionItemByBundleAndModel(
	_ context.Context, _, _ string,
) (*rbac.EndUserBundleDataPermissionItem, error) {
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

func (s *stubRBACRepo) CreateRole(_ context.Context, _ *rbac.EndUserRole) error { return nil }

func (s *stubRBACRepo) GetRoleByID(_ context.Context, _, _ string) (*rbac.EndUserRole, error) {
	return nil, nil
}

func (s *stubRBACRepo) ListRolesByProject(_ context.Context, _, _ string) ([]*rbac.EndUserRole, error) {
	return nil, nil
}

func (s *stubRBACRepo) UpdateRole(_ context.Context, _ *rbac.EndUserRole) error { return nil }
func (s *stubRBACRepo) DeleteRole(_ context.Context, _, _ string) error         { return nil }

func (s *stubRBACRepo) AssignBundleToRole(_ context.Context, _, _, _, _ string) error { return nil }

func (s *stubRBACRepo) RevokeBundleFromRole(_ context.Context, _, _ string) error { return nil }

func (s *stubRBACRepo) ListBundlesByRole(_ context.Context, _ string) ([]*rbac.EndUserPermissionBundle, error) {
	return nil, nil
}

func (s *stubRBACRepo) GrantBundleToUser(_ context.Context, _, _, _, _ string) error { return nil }

func (s *stubRBACRepo) RevokeBundleFromUser(_ context.Context, _, _, _, _ string) error { return nil }

func (s *stubRBACRepo) AssignRoleToUser(_ context.Context, _, _, _, _ string) error { return nil }

func (s *stubRBACRepo) RevokeRoleFromUser(_ context.Context, _, _, _, _ string) error { return nil }

func (s *stubRBACRepo) ListProjectEndUserRoleUsers(
	_ context.Context, _ rbac.ListProjectEndUserRoleUsersQuery,
) ([]*rbac.ProjectEndUserRoleUser, int64, error) {
	return nil, 0, nil
}

func (s *stubRBACRepo) IsUserBuiltin(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// newModelWithOwnerField 创建含 owner 字段的 DataModel 用于测试。
func newModelWithOwnerField(modelID, ownerFieldName string) *modeldesign.DataModel {
	ownerType, _ := modeldesign.NewFieldFormat(modeldesign.FormatEndUserRef)
	return &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{ID: modelID},
		Fields: []*modeldesign.FieldDefinition{
			{Name: ownerFieldName, Type: ownerType},
		},
	}
}

// makeRowPolicy 创建并规范化 RowPolicy。
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

// presetItem 构造一个 PRESET grant_type 的 EndUserBundleDataPermissionItem。
func presetItem(modelID string, preset rbac.PermissionPreset) *rbac.EndUserBundleDataPermissionItem {
	p := preset
	return &rbac.EndUserBundleDataPermissionItem{
		ModelID:   modelID,
		GrantType: rbac.PermissionTypePreset,
		Preset:    &p,
	}
}

// ─── existing tests (CUSTOM path) ────────────────────────────────────────────

func TestEndUserPermissionService_Resolve_TenantAdmin(t *testing.T) {
	svc := appruntimeimport.NewEndUserPermissionService(&stubRBACRepo{}, nil)
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "", "model-id")
	if err != nil {
		t.Fatal(err)
	}
	if perms != nil {
		t.Error("tenant admin (empty endUserID) should return nil permissions")
	}
}

func TestEndUserPermissionService_Resolve_NoPermissions(t *testing.T) {
	svc := appruntimeimport.NewEndUserPermissionService(&stubRBACRepo{}, nil)
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
	const (
		modelID  = "model-id"
		bundleID = "bundle-select-all"
	)
	rp := makeRowPolicy(true, rbac.ScopeAll, false, "", false, "", false, "")
	stub := &stubRBACRepo{
		bundleIDsByExplicitRoles: []string{bundleID},
		customPerms: []*rbac.EndUserPermission{
			{ModelID: modelID, RowPolicy: rp},
		},
	}
	svc := appruntimeimport.NewEndUserPermissionService(stub, &stubModelRepo{})
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", modelID)
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
	const (
		modelID  = "model-id"
		bundleID = "bundle-self"
	)
	rp := makeRowPolicy(true, rbac.ScopeCustom, true, rbac.ScopeCustom, false, "", false, "")
	stub := &stubRBACRepo{
		bundleIDsByExplicitRoles: []string{bundleID},
		customPerms: []*rbac.EndUserPermission{
			{ModelID: modelID, RowPolicy: rp},
		},
	}
	svc := appruntimeimport.NewEndUserPermissionService(stub, &stubModelRepo{})
	perms, err := svc.Resolve(context.Background(), "org1", "proj1", "user1", modelID)
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

// ─── new tests: PRESET path (RED — will fail until fix is applied) ────────────

// TestEndUserPermissionService_Resolve_PresetReadAll 验证：
// bundle 中配置了 PRESET=READ_ALL 的 data permission item，
// Resolve 应返回 Select.Allowed=true，Insert/Update/Delete.Allowed=false。
func TestEndUserPermissionService_Resolve_PresetReadAll(t *testing.T) {
	const (
		modelID  = "model-preset-1"
		bundleID = "bundle-preset-1"
		userID   = "user-preset-1"
	)

	stub := &stubRBACRepo{
		bundleIDsByExplicitRoles: []string{bundleID},
		bundleItems: map[string][]*rbac.EndUserBundleDataPermissionItem{
			bundleID: {presetItem(modelID, rbac.PresetReadAll)},
		},
	}
	svc := appruntimeimport.NewEndUserPermissionService(stub, &stubModelRepo{})

	perms, err := svc.Resolve(context.Background(), "org1", "proj1", userID, modelID)
	if err != nil {
		t.Fatal(err)
	}
	if perms == nil {
		t.Fatal("expected non-nil perms")
	}
	if !perms.Select.Allowed {
		t.Error("PRESET=READ_ALL: expected Select.Allowed = true")
	}
	if perms.Select.IsSelf {
		t.Error("PRESET=READ_ALL: expected Select.IsSelf = false (scope=ALL)")
	}
	if perms.Insert.Allowed {
		t.Error("PRESET=READ_ALL: expected Insert.Allowed = false")
	}
	if perms.Update.Allowed {
		t.Error("PRESET=READ_ALL: expected Update.Allowed = false")
	}
	if perms.Delete.Allowed {
		t.Error("PRESET=READ_ALL: expected Delete.Allowed = false")
	}
}

// TestEndUserPermissionService_Resolve_PresetReadWriteOwner 验证：
// bundle 中配置了 PRESET=READ_WRITE_OWNER，且 model 有 owner 字段，
// Resolve 应返回所有操作 Allowed=true 且 IsSelf=true（owner-scoped）。
func TestEndUserPermissionService_Resolve_PresetReadWriteOwner(t *testing.T) {
	const (
		modelID    = "model-preset-2"
		bundleID   = "bundle-preset-2"
		userID     = "user-preset-2"
		ownerField = "owner"
	)

	stub := &stubRBACRepo{
		bundleIDsByExplicitRoles: []string{bundleID},
		bundleItems: map[string][]*rbac.EndUserBundleDataPermissionItem{
			bundleID: {presetItem(modelID, rbac.PresetReadWriteOwner)},
		},
	}
	modelRepo := &stubModelRepo{
		models: map[string]*modeldesign.DataModel{
			modelID: newModelWithOwnerField(modelID, ownerField),
		},
	}
	svc := appruntimeimport.NewEndUserPermissionService(stub, modelRepo)

	perms, err := svc.Resolve(context.Background(), "org1", "proj1", userID, modelID)
	if err != nil {
		t.Fatal(err)
	}
	if perms == nil {
		t.Fatal("expected non-nil perms")
	}
	if !perms.Select.Allowed {
		t.Error("PRESET=READ_WRITE_OWNER: expected Select.Allowed = true")
	}
	if !perms.Select.IsSelf {
		t.Error("PRESET=READ_WRITE_OWNER: expected Select.IsSelf = true")
	}
	if !perms.Insert.Allowed {
		t.Error("PRESET=READ_WRITE_OWNER: expected Insert.Allowed = true")
	}
	if !perms.Insert.IsSelf {
		t.Error("PRESET=READ_WRITE_OWNER: expected Insert.IsSelf = true")
	}
	if !perms.Update.Allowed {
		t.Error("PRESET=READ_WRITE_OWNER: expected Update.Allowed = true")
	}
	if !perms.Update.IsSelf {
		t.Error("PRESET=READ_WRITE_OWNER: expected Update.IsSelf = true")
	}
	if !perms.Delete.Allowed {
		t.Error("PRESET=READ_WRITE_OWNER: expected Delete.Allowed = true")
	}
	if !perms.Delete.IsSelf {
		t.Error("PRESET=READ_WRITE_OWNER: expected Delete.IsSelf = true")
	}
}
