package rbac

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rbacdomain "modelcraft/internal/domain/rbac"
)

type mockPermissionRepo struct {
	byModel map[string][]*rbacdomain.EndUserPermission
	refs    map[string]bool
}

func newMockPermissionRepo(initial []*rbacdomain.EndUserPermission) *mockPermissionRepo {
	repo := &mockPermissionRepo{
		byModel: map[string][]*rbacdomain.EndUserPermission{},
		refs:    map[string]bool{},
	}
	for _, p := range initial {
		repo.byModel[p.ModelID] = append(repo.byModel[p.ModelID], clonePermission(p))
	}
	return repo
}

func (m *mockPermissionRepo) CreatePermission(_ context.Context, p *rbacdomain.EndUserPermission) error {
	m.byModel[p.ModelID] = append(m.byModel[p.ModelID], clonePermission(p))
	return nil
}

func (m *mockPermissionRepo) GetPermissionByID(
	_ context.Context,
	orgName, id string,
) (*rbacdomain.EndUserPermission, error) {
	for _, list := range m.byModel {
		for _, p := range list {
			if p.OrgName == orgName && p.ID == id {
				return clonePermission(p), nil
			}
		}
	}
	return nil, shared.NewNotFoundError("not found")
}

func (m *mockPermissionRepo) ListPermissionsByProject(
	_ context.Context,
	orgName, projectSlug string,
) ([]*rbacdomain.EndUserPermission, error) {
	result := make([]*rbacdomain.EndUserPermission, 0)
	for _, list := range m.byModel {
		for _, p := range list {
			if p.OrgName == orgName && p.ProjectSlug == projectSlug {
				result = append(result, clonePermission(p))
			}
		}
	}
	return result, nil
}

func (m *mockPermissionRepo) ListPermissionsByModel(
	_ context.Context,
	orgName, modelID string,
) ([]*rbacdomain.EndUserPermission, error) {
	list := m.byModel[modelID]
	result := make([]*rbacdomain.EndUserPermission, 0, len(list))
	for _, p := range list {
		if p.OrgName == orgName {
			result = append(result, clonePermission(p))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (m *mockPermissionRepo) ListPresetPermissionsByModel(
	_ context.Context,
	orgName, modelID string,
) ([]*rbacdomain.EndUserPermission, error) {
	list := m.byModel[modelID]
	result := make([]*rbacdomain.EndUserPermission, 0, len(list))
	for _, p := range list {
		if p.OrgName == orgName && p.Type == rbacdomain.PermissionTypePreset {
			result = append(result, clonePermission(p))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (m *mockPermissionRepo) UpdatePermission(_ context.Context, p *rbacdomain.EndUserPermission) error {
	for modelID, list := range m.byModel {
		for i, item := range list {
			if item.ID == p.ID && item.OrgName == p.OrgName {
				m.byModel[modelID][i] = clonePermission(p)
				return nil
			}
		}
	}
	return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "not found")
}

func (m *mockPermissionRepo) DeletePermission(_ context.Context, orgName, id string) error {
	for modelID, list := range m.byModel {
		for i, item := range list {
			if item.ID == id && item.OrgName == orgName {
				m.byModel[modelID] = append(list[:i], list[i+1:]...)
				return nil
			}
		}
	}
	return shared.NewRepositoryError(shared.ErrTypeNoRowsAffected, "not found")
}

func (m *mockPermissionRepo) DeletePresetPermissionsByModel(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockPermissionRepo) UpdatePresetPermission(_ context.Context, p *rbacdomain.EndUserPermission) error {
	return m.UpdatePermission(context.Background(), p)
}

func (m *mockPermissionRepo) IsPermissionReferencedByBundle(_ context.Context, permissionID string) (bool, error) {
	return m.refs[permissionID], nil
}

type mockModelRepo struct {
	model *modeldesign.DataModel
	err   error
}

func (m *mockModelRepo) GetByID(
	_ context.Context,
	_ string,
	_ ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.model, nil
}

func TestApplyPresetPolicy_Reconcile(t *testing.T) {
	orgName := "org-a"
	projectSlug := "proj-a"
	modelID := "model-a"

	makeService := func(
		model *modeldesign.DataModel,
		initial []*rbacdomain.EndUserPermission,
	) *EndUserPermissionAppService {
		return &EndUserPermissionAppService{
			rbacRepo:  newMockPermissionRepo(initial),
			modelRepo: &mockModelRepo{model: model},
		}
	}

	t.Run("模型级 reconcile 补齐可适配预设并保留 CUSTOM", func(t *testing.T) {
		svc := makeService(makeModelWithOwner(false), []*rbacdomain.EndUserPermission{
			makeCustomPermission(orgName, projectSlug, modelID, "custom-1"),
		})

		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
			ModelID:      modelID,
			Preset:       nil,
		})
		require.NoError(t, err)
		require.Len(t, perms, 3)

		names := []string{perms[0].Name, perms[1].Name, perms[2].Name}
		sort.Strings(names)
		assert.Equal(t, []string{"custom-1", "preset:READ_ALL", "preset:READ_WRITE_ALL"}, names)
	})

	t.Run("显式 OWNER 预设且缺少 owner 字段时报错", func(t *testing.T) {
		svc := makeService(makeModelWithOwner(false), nil)
		preset := rbacdomain.PresetReadWriteOwner
		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
			ModelID:      modelID,
			Preset:       &preset,
		})
		require.Error(t, err)
		assert.Nil(t, perms)
		var bizErr *bizerrors.BusinessError
		require.True(t, bizerrors.As(err, &bizErr))
		assert.Equal(t, bizerrors.EndUserPresetRequiresOwnerField.GetCode(), bizErr.Info().GetCode())
	})

	t.Run("toUpdate 原地更新并保持 permission_id", func(t *testing.T) {
		existing := makePresetPermission(orgName, projectSlug, modelID, "legacy-name", rbacdomain.PresetReadAll)
		existing.ID = "perm-fixed-id"

		svc := makeService(makeModelWithOwner(false), []*rbacdomain.EndUserPermission{existing})
		preset := rbacdomain.PresetReadAll
		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
			ModelID:      modelID,
			Preset:       &preset,
		})
		require.NoError(t, err)
		require.Len(t, perms, 1)
		assert.Equal(t, "perm-fixed-id", perms[0].ID)
		assert.Equal(t, "preset:READ_ALL", perms[0].Name)
	})

	t.Run("toDelete 若被引用则阻断", func(t *testing.T) {
		target := makePresetPermission(
			orgName,
			projectSlug,
			modelID,
			"preset:READ_WRITE_OWNER",
			rbacdomain.PresetReadWriteOwner,
		)
		target.ID = "preset-owner-id"
		svc := makeService(makeModelWithOwner(false), []*rbacdomain.EndUserPermission{target})
		svc.rbacRepo.(*mockPermissionRepo).refs[target.ID] = true

		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
			ModelID:      modelID,
			Preset:       nil,
		})
		require.Error(t, err)
		assert.Nil(t, perms)
		var bizErr *bizerrors.BusinessError
		require.True(t, bizerrors.As(err, &bizErr))
		assert.Equal(t, bizerrors.PresetDeleteBlockedByBundle.GetCode(), bizErr.Info().GetCode())
	})
}

func TestExpandPreset(t *testing.T) {
	t.Run("READ_WRITE_ALL", func(t *testing.T) {
		rp, err := expandPreset(rbacdomain.PresetReadWriteAll, "")
		require.NoError(t, err)
		require.NotNil(t, rp)
		assert.True(t, rp.Select.Allowed)
		assert.True(t, rp.Insert.Allowed)
		assert.True(t, rp.Update.Allowed)
		assert.True(t, rp.Delete.Allowed)
		assert.NoError(t, rp.Validate())
	})

	t.Run("owner preset requires owner field", func(t *testing.T) {
		_, err := expandPreset(rbacdomain.PresetReadWriteOwner, "")
		require.Error(t, err)
	})
}

func TestListVirtualPresetsByModel(t *testing.T) {
	t.Run("无 owner 字段时仅返回非 owner 预设", func(t *testing.T) {
		svc := &EndUserPermissionAppService{modelRepo: &mockModelRepo{model: makeModelWithOwner(false)}}
		presets, err := svc.ListVirtualPresetsByModel(context.Background(), "model-a")
		require.NoError(t, err)
		assert.Equal(t, []rbacdomain.PermissionPreset{
			rbacdomain.PresetReadWriteAll,
			rbacdomain.PresetReadAll,
		}, presets)
	})

	t.Run("有 owner 字段时返回全部可适配预设", func(t *testing.T) {
		svc := &EndUserPermissionAppService{modelRepo: &mockModelRepo{model: makeModelWithOwner(true)}}
		presets, err := svc.ListVirtualPresetsByModel(context.Background(), "model-a")
		require.NoError(t, err)
		assert.Equal(t, []rbacdomain.PermissionPreset{
			rbacdomain.PresetReadWriteAll,
			rbacdomain.PresetReadAll,
			rbacdomain.PresetReadWriteOwner,
			rbacdomain.PresetReadAllWriteOwner,
		}, presets)
	})
}

func clonePermission(p *rbacdomain.EndUserPermission) *rbacdomain.EndUserPermission {
	if p == nil {
		return nil
	}
	clone := *p
	if p.Preset != nil {
		preset := *p.Preset
		clone.Preset = &preset
	}
	if p.RowPolicy != nil {
		rp := *p.RowPolicy
		clone.RowPolicy = &rp
	}
	return &clone
}

func makeCustomPermission(orgName, projectSlug, modelID, name string) *rbacdomain.EndUserPermission {
	return &rbacdomain.EndUserPermission{
		ID:          name + "-id",
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
		Name:        name,
		Type:        rbacdomain.PermissionTypeCustom,
		RowPolicy: &rbacdomain.RowPolicy{
			Select: rbacdomain.SelectPolicy{Allowed: true, Scope: rbacdomain.ScopeAll},
			Insert: rbacdomain.InsertPolicy{Allowed: false},
			Update: rbacdomain.UpdatePolicy{Allowed: false},
			Delete: rbacdomain.DeletePolicy{Allowed: false},
		},
	}
}

func makePresetPermission(
	orgName, projectSlug, modelID, name string,
	preset rbacdomain.PermissionPreset,
) *rbacdomain.EndUserPermission {
	return &rbacdomain.EndUserPermission{
		ID:          name + "-id",
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ModelID:     modelID,
		Name:        name,
		Type:        rbacdomain.PermissionTypePreset,
		RowPolicy: &rbacdomain.RowPolicy{
			Select: rbacdomain.SelectPolicy{Allowed: true, Scope: rbacdomain.ScopeAll},
			Insert: rbacdomain.InsertPolicy{Allowed: false},
			Update: rbacdomain.UpdatePolicy{Allowed: false},
			Delete: rbacdomain.DeletePolicy{Allowed: false},
		},
		Preset: &preset,
	}
}

func makeModelWithOwner(withOwner bool) *modeldesign.DataModel {
	m := &modeldesign.DataModel{}
	if !withOwner {
		return m
	}
	ownerType := modeldesign.GetFieldTypeByFormat(modeldesign.FormatEndUserRef)
	m.Fields = []*modeldesign.FieldDefinition{{Name: "owner", Type: ownerType}}
	return m
}

type mockBundleRepo struct {
	bundle      *rbacdomain.EndUserPermissionBundle
	modelPerms  map[string][]*rbacdomain.EndUserPermission
	bundlePerm  map[string][]string
	bundleItems map[string][]*rbacdomain.EndUserBundleDataPermissionItem

	createPermissionCalls int
	addPermissionCalls    int
}

func newMockBundleRepo(
	bundle *rbacdomain.EndUserPermissionBundle,
	modelPerms map[string][]*rbacdomain.EndUserPermission,
) *mockBundleRepo {
	cloned := map[string][]*rbacdomain.EndUserPermission{}
	for modelID, items := range modelPerms {
		copyItems := make([]*rbacdomain.EndUserPermission, 0, len(items))
		for _, item := range items {
			copyItems = append(copyItems, clonePermission(item))
		}
		cloned[modelID] = copyItems
	}
	return &mockBundleRepo{
		bundle:      bundle,
		modelPerms:  cloned,
		bundlePerm:  map[string][]string{},
		bundleItems: map[string][]*rbacdomain.EndUserBundleDataPermissionItem{},
	}
}

func (m *mockBundleRepo) CreateBundle(_ context.Context, _ *rbacdomain.EndUserPermissionBundle) error {
	return nil
}

func (m *mockBundleRepo) GetBundleByID(
	_ context.Context,
	orgName, _, id string,
) (*rbacdomain.EndUserPermissionBundle, error) {
	if m.bundle == nil || m.bundle.OrgName != orgName || m.bundle.ID != id {
		return nil, shared.NewNotFoundError("bundle not found")
	}
	copied := *m.bundle
	return &copied, nil
}

func (m *mockBundleRepo) ListBundlesByProject(
	_ context.Context,
	_, _ string,
) ([]*rbacdomain.EndUserPermissionBundle, error) {
	return nil, nil
}

func (m *mockBundleRepo) GetBundleBySlug(
	_ context.Context,
	_, _, _ string,
) (*rbacdomain.EndUserPermissionBundle, error) {
	return nil, nil
}

func (m *mockBundleRepo) UpdateBundle(_ context.Context, _ *rbacdomain.EndUserPermissionBundle) error {
	return nil
}

func (m *mockBundleRepo) DeleteBundle(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockBundleRepo) GetPermissionByID(
	_ context.Context,
	orgName, id string,
) (*rbacdomain.EndUserPermission, error) {
	p := m.findPermissionByID(id)
	if p == nil || p.OrgName != orgName {
		return nil, shared.NewNotFoundError("permission not found")
	}
	return clonePermission(p), nil
}

func (m *mockBundleRepo) UpsertBundleDataPermissionItem(
	_ context.Context,
	item *rbacdomain.EndUserBundleDataPermissionItem,
) error {
	if item == nil {
		return nil
	}
	list := m.bundleItems[item.BundleID]
	for i, existing := range list {
		if existing.ModelID == item.ModelID {
			copied := *item
			list[i] = &copied
			m.bundleItems[item.BundleID] = list
			return nil
		}
	}
	copied := *item
	m.bundleItems[item.BundleID] = append(list, &copied)
	return nil
}

func (m *mockBundleRepo) RemoveBundleDataPermissionItem(
	_ context.Context,
	bundleID, modelID string,
) error {
	list := m.bundleItems[bundleID]
	if len(list) == 0 {
		return nil
	}
	next := make([]*rbacdomain.EndUserBundleDataPermissionItem, 0, len(list))
	for _, item := range list {
		if item.ModelID == modelID {
			continue
		}
		next = append(next, item)
	}
	m.bundleItems[bundleID] = next
	return nil
}

func (m *mockBundleRepo) ListBundleDataPermissionItems(
	_ context.Context,
	bundleID string,
) ([]*rbacdomain.EndUserBundleDataPermissionItem, error) {
	list := m.bundleItems[bundleID]
	result := make([]*rbacdomain.EndUserBundleDataPermissionItem, 0, len(list))
	for _, item := range list {
		copied := *item
		result = append(result, &copied)
	}
	return result, nil
}

func (m *mockBundleRepo) AddPermissionToBundle(
	_ context.Context,
	bundleID, permissionID string,
	_ int,
) error {
	m.addPermissionCalls++
	for _, id := range m.bundlePerm[bundleID] {
		if id == permissionID {
			return shared.NewDuplicateKeyError("duplicate bundle permission")
		}
	}
	m.bundlePerm[bundleID] = append(m.bundlePerm[bundleID], permissionID)
	return nil
}

func (m *mockBundleRepo) RemovePermissionFromBundle(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockBundleRepo) ListPermissionsInBundle(
	_ context.Context,
	bundleID string,
) ([]*rbacdomain.EndUserPermission, error) {
	ids := m.bundlePerm[bundleID]
	result := make([]*rbacdomain.EndUserPermission, 0, len(ids))
	for _, id := range ids {
		if p := m.findPermissionByID(id); p != nil {
			result = append(result, clonePermission(p))
		}
	}
	return result, nil
}

func (m *mockBundleRepo) GrantBundleToUser(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (m *mockBundleRepo) RevokeBundleFromUser(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (m *mockBundleRepo) ListPermissionsByModel(
	_ context.Context,
	orgName, modelID string,
) ([]*rbacdomain.EndUserPermission, error) {
	list := m.modelPerms[modelID]
	result := make([]*rbacdomain.EndUserPermission, 0, len(list))
	for _, p := range list {
		if p.OrgName == orgName {
			result = append(result, clonePermission(p))
		}
	}
	return result, nil
}

func (m *mockBundleRepo) GetPermissionByModelTypeName(
	_ context.Context,
	orgName, modelID string,
	permissionType rbacdomain.PermissionType,
	name string,
) (*rbacdomain.EndUserPermission, error) {
	for _, p := range m.modelPerms[modelID] {
		if p.OrgName == orgName && p.Type == permissionType && p.Name == name {
			return clonePermission(p), nil
		}
	}
	return nil, shared.NewNotFoundError("permission not found")
}

func (m *mockBundleRepo) CreatePermission(_ context.Context, p *rbacdomain.EndUserPermission) error {
	m.createPermissionCalls++
	m.modelPerms[p.ModelID] = append(m.modelPerms[p.ModelID], clonePermission(p))
	return nil
}

func (m *mockBundleRepo) SaveBundleSnapshot(_ context.Context, _ *rbacdomain.BundleSnapshot) error {
	return nil
}

func (m *mockBundleRepo) ListBundleSnapshots(_ context.Context, _ string) ([]rbacdomain.BundleSnapshot, error) {
	return nil, nil
}

func (m *mockBundleRepo) DeleteOldBundleSnapshots(_ context.Context, _ string) error {
	return nil
}

func (m *mockBundleRepo) GetBundleCurrentVersion(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (m *mockBundleRepo) GetBundleSnapshotByVersion(
	_ context.Context, _ string, _ int,
) (*rbacdomain.BundleSnapshot, error) {
	return nil, nil
}

func (m *mockBundleRepo) ClearBundlePermissions(_ context.Context, _ string) error {
	return nil
}

func (m *mockBundleRepo) GetBundleDataPermissionItemByBundleAndModel(
	_ context.Context, _, _ string,
) (*rbacdomain.EndUserBundleDataPermissionItem, error) {
	return nil, nil
}

func (m *mockBundleRepo) findPermissionByID(id string) *rbacdomain.EndUserPermission {
	for _, list := range m.modelPerms {
		for _, p := range list {
			if p.ID == id {
				return p
			}
		}
	}
	return nil
}

func TestAddPresetToBundle(t *testing.T) {
	orgName := "org-a"
	projectSlug := "proj-a"
	bundleID := "bundle-1"
	modelID := "model-a"

	bundle := &rbacdomain.EndUserPermissionBundle{
		ID:          bundleID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Name:        "bundle-main",
	}

	t.Run("已存在预设时复用 permission 且重复请求幂等", func(t *testing.T) {
		existing := makePresetPermission(
			orgName,
			projectSlug,
			modelID,
			"preset:READ_ALL",
			rbacdomain.PresetReadAll,
		)
		existing.ID = "perm-read-all"

		repo := newMockBundleRepo(
			bundle,
			map[string][]*rbacdomain.EndUserPermission{modelID: {existing}},
		)
		svc := &EndUserBundleAppService{
			rbacRepo:  repo,
			modelRepo: &mockModelRepo{model: makeModelWithOwner(false)},
		}

		cmd := AddPresetToBundleCommand{
			ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
			BundleID:     bundleID,
			ModelID:      modelID,
			Preset:       rbacdomain.PresetReadAll,
			SortOrder:    1,
		}
		_, err := svc.AddPresetToBundle(context.Background(), cmd)
		require.NoError(t, err)
		_, err = svc.AddPresetToBundle(context.Background(), cmd)
		require.NoError(t, err)

		assert.Equal(t, 0, repo.createPermissionCalls)
		require.Len(t, repo.bundleItems[bundleID], 1)
		item := repo.bundleItems[bundleID][0]
		assert.Equal(t, modelID, item.ModelID)
		assert.Equal(t, rbacdomain.PermissionTypePreset, item.GrantType)
		require.NotNil(t, item.Preset)
		assert.Equal(t, rbacdomain.PresetReadAll, *item.Preset)
		assert.Nil(t, item.CustomPermissionID)
	})

	t.Run("不存在预设时自动创建并绑定", func(t *testing.T) {
		repo := newMockBundleRepo(bundle, map[string][]*rbacdomain.EndUserPermission{})
		svc := &EndUserBundleAppService{
			rbacRepo:  repo,
			modelRepo: &mockModelRepo{model: makeModelWithOwner(false)},
		}

		_, err := svc.AddPresetToBundle(context.Background(), AddPresetToBundleCommand{
			ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
			BundleID:     bundleID,
			ModelID:      modelID,
			Preset:       rbacdomain.PresetReadWriteAll,
			SortOrder:    2,
		})
		require.NoError(t, err)

		require.Equal(t, 0, repo.createPermissionCalls)
		require.Len(t, repo.bundleItems[bundleID], 1)
		item := repo.bundleItems[bundleID][0]
		assert.Equal(t, modelID, item.ModelID)
		assert.Equal(t, rbacdomain.PermissionTypePreset, item.GrantType)
		require.NotNil(t, item.Preset)
		assert.Equal(t, rbacdomain.PresetReadWriteAll, *item.Preset)
		assert.Nil(t, item.CustomPermissionID)
	})

	t.Run("OWNER 预设缺少 owner 字段时返回结构化错误", func(t *testing.T) {
		repo := newMockBundleRepo(bundle, map[string][]*rbacdomain.EndUserPermission{})
		svc := &EndUserBundleAppService{
			rbacRepo:  repo,
			modelRepo: &mockModelRepo{model: makeModelWithOwner(false)},
		}

		_, err := svc.AddPresetToBundle(context.Background(), AddPresetToBundleCommand{
			ProjectScope: project.ProjectScope{OrgName: orgName, ProjectSlug: projectSlug},
			BundleID:     bundleID,
			ModelID:      modelID,
			Preset:       rbacdomain.PresetReadWriteOwner,
			SortOrder:    3,
		})
		require.Error(t, err)
		var bizErr *bizerrors.BusinessError
		require.True(t, bizerrors.As(err, &bizErr))
		assert.Equal(t, bizerrors.EndUserPresetRequiresOwnerField.GetCode(), bizErr.Info().GetCode())
		assert.Empty(t, repo.bundleItems[bundleID])
		assert.Equal(t, 0, repo.createPermissionCalls)
	})
}

func TestDeletePermission_ReferencedByBundle(t *testing.T) {
	orgName := "org-a"
	modelID := "model-a"

	t.Run("被 bundle 引用时返回 EndUserPermissionInUse 错误", func(t *testing.T) {
		perm := makeCustomPermission(orgName, "proj-a", modelID, "custom-1")
		repo := newMockPermissionRepo([]*rbacdomain.EndUserPermission{perm})
		repo.refs[perm.ID] = true

		svc := &EndUserPermissionAppService{
			rbacRepo:  repo,
			modelRepo: &mockModelRepo{model: makeModelWithOwner(false)},
		}

		err := svc.DeletePermission(context.Background(), DeletePermissionCommand{
			OrgName: orgName,
			ID:      perm.ID,
		})
		require.Error(t, err)
		var bizErr *bizerrors.BusinessError
		require.True(t, bizerrors.As(err, &bizErr))
		assert.Equal(t, bizerrors.EndUserPermissionInUse.GetCode(), bizErr.Info().GetCode())

		// 确认 permission 未被删除
		remaining, _ := repo.ListPermissionsByModel(context.Background(), orgName, modelID)
		assert.Len(t, remaining, 1)
	})

	t.Run("未被引用时正常删除", func(t *testing.T) {
		perm := makeCustomPermission(orgName, "proj-a", modelID, "custom-2")
		repo := newMockPermissionRepo([]*rbacdomain.EndUserPermission{perm})
		// refs 默认 false

		svc := &EndUserPermissionAppService{
			rbacRepo:  repo,
			modelRepo: &mockModelRepo{model: makeModelWithOwner(false)},
		}

		err := svc.DeletePermission(context.Background(), DeletePermissionCommand{
			OrgName: orgName,
			ID:      perm.ID,
		})
		require.NoError(t, err)

		remaining, _ := repo.ListPermissionsByModel(context.Background(), orgName, modelID)
		assert.Empty(t, remaining)
	})
}
