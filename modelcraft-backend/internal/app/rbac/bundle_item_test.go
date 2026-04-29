package rbac

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rbacdomain "modelcraft/internal/domain/rbac"
)

// ─── 测试专用 repo（扩展 mockBundleRepo，支持 snapshot 追踪和 RestoreBundle）────

type bundleItemTestRepo struct {
	*mockBundleRepo
	capturedSnapshots  []*rbacdomain.BundleSnapshot
	snapshotsByVersion map[int]*rbacdomain.BundleSnapshot
	currentVersion     int
}

func newBundleItemRepo(bundle *rbacdomain.EndUserPermissionBundle) *bundleItemTestRepo {
	return &bundleItemTestRepo{
		mockBundleRepo:     newMockBundleRepo(bundle, nil),
		snapshotsByVersion: map[int]*rbacdomain.BundleSnapshot{},
	}
}

func newBundleItemRepoWithPerms(
	bundle *rbacdomain.EndUserPermissionBundle,
	modelPerms map[string][]*rbacdomain.EndUserPermission,
) *bundleItemTestRepo {
	return &bundleItemTestRepo{
		mockBundleRepo:     newMockBundleRepo(bundle, modelPerms),
		snapshotsByVersion: map[int]*rbacdomain.BundleSnapshot{},
	}
}

func (r *bundleItemTestRepo) SaveBundleSnapshot(_ context.Context, snap *rbacdomain.BundleSnapshot) error {
	r.capturedSnapshots = append(r.capturedSnapshots, snap)
	return nil
}

func (r *bundleItemTestRepo) GetBundleCurrentVersion(_ context.Context, _ string) (int, error) {
	return r.currentVersion, nil
}

func (r *bundleItemTestRepo) GetBundleSnapshotByVersion(
	_ context.Context, _ string, version int,
) (*rbacdomain.BundleSnapshot, error) {
	snap, ok := r.snapshotsByVersion[version]
	if !ok {
		return nil, shared.NewNotFoundError("snapshot not found")
	}
	return snap, nil
}

// 覆盖 GetBundleByID：增加 ProjectSlug 验证（verifyBundleScope 需要）
func (r *bundleItemTestRepo) GetBundleByID(
	ctx context.Context, orgName, id string,
) (*rbacdomain.EndUserPermissionBundle, error) {
	return r.mockBundleRepo.GetBundleByID(ctx, orgName, id)
}

func (r *bundleItemTestRepo) itemsByModel(bundleID, modelID string) []*rbacdomain.EndUserBundleDataPermissionItem {
	var result []*rbacdomain.EndUserBundleDataPermissionItem
	for _, item := range r.bundleItems[bundleID] {
		if item.ModelID == modelID {
			result = append(result, item)
		}
	}
	return result
}

func (r *bundleItemTestRepo) lastSnapshot() *rbacdomain.BundleSnapshot {
	if len(r.capturedSnapshots) == 0 {
		return nil
	}
	return r.capturedSnapshots[len(r.capturedSnapshots)-1]
}

// ─── 测试专用 model repo（实现 modelRepository 接口）────────────────────────

type itemTestModelRepo struct {
	ownerFieldName string
}

func (m *itemTestModelRepo) GetByID(
	_ context.Context, _ string, _ ...*modeldesign.ModelQueryOptions,
) (*modeldesign.DataModel, error) {
	model := &modeldesign.DataModel{}
	if m.ownerFieldName != "" {
		_ = model // ownerField 通过 GetOwnerField() 返回
	}
	return makeModelWithOwner(m.ownerFieldName != ""), nil
}

// ─── 辅助 ────────────────────────────────────────────────────────────────────

func makeTestBundle(id, org, proj string) *rbacdomain.EndUserPermissionBundle {
	return &rbacdomain.EndUserPermissionBundle{
		ID:          id,
		OrgName:     org,
		ProjectSlug: proj,
		Name:        "test-bundle",
	}
}

func makeTestCustomPerm(id, org, modelID string) *rbacdomain.EndUserPermission {
	return makeCustomPermission(org, "proj1", modelID, id)
}

func newItemSvc(repo *bundleItemTestRepo, hasOwnerField bool) *EndUserBundleAppService {
	return &EndUserBundleAppService{
		rbacRepo:  repo,
		modelRepo: &itemTestModelRepo{ownerFieldName: map[bool]string{true: "owner", false: ""}[hasOwnerField]},
	}
}

// ─── 5.1: preset/custom item 绑定、bundle-model 唯一约束 ────────────────────

func TestBindPresetItem_Success(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	svc := newItemSvc(repo, false)

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})

	require.NoError(t, err)
	items := repo.bundleItems["b1"]
	require.Len(t, items, 1)
	assert.Equal(t, rbacdomain.PermissionTypePreset, items[0].GrantType)
	assert.Equal(t, rbacdomain.PresetReadAll, *items[0].Preset)
	assert.Equal(t, "model1", items[0].ModelID)
}

func TestBindPresetItem_WrongProject_Returns404(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	svc := newItemSvc(repo, false)

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj-other"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})

	require.Error(t, err, "跨 project 操作应返回错误")
}

func TestBindPresetItem_OwnerPreset_WithoutOwnerField_Fails(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	svc := newItemSvc(repo, false) // 无 owner 字段

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadWriteOwner,
	})

	require.Error(t, err, "缺少 owner 字段时绑定 OWNER 预设应失败")
}

func TestBindPresetItem_OwnerPreset_WithOwnerField_Succeeds(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	svc := newItemSvc(repo, true) // 有 owner 字段

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadWriteOwner,
	})

	require.NoError(t, err)
	items := repo.bundleItems["b1"]
	require.Len(t, items, 1)
	assert.Equal(t, rbacdomain.PresetReadWriteOwner, *items[0].Preset)
}

func TestBindPreset_ThenBindPreset_SameModel_Replaces(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	svc := newItemSvc(repo, false)

	cmd := BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	}
	_, err := svc.BindPresetItem(context.Background(), cmd)
	require.NoError(t, err)

	cmd.Preset = rbacdomain.PresetReadWriteAll
	_, err = svc.BindPresetItem(context.Background(), cmd)
	require.NoError(t, err)

	items := repo.itemsByModel("b1", "model1")
	assert.Len(t, items, 1, "同一模型只应保留一个 item")
	assert.Equal(t, rbacdomain.PresetReadWriteAll, *items[0].Preset)
}

func TestBindPreset_ThenBindCustom_SameModel_Replaces(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	perm := makeTestCustomPerm("perm1", "org1", "model1")
	repo := newBundleItemRepoWithPerms(bundle, map[string][]*rbacdomain.EndUserPermission{
		"model1": {perm},
	})
	svc := newItemSvc(repo, false)

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})
	require.NoError(t, err)

	_, err = svc.BindCustomItem(context.Background(), BindCustomItemToBundleCommand{
		ProjectScope:       project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:           "b1",
		ModelID:            "model1",
		CustomPermissionID: perm.ID, // 使用 perm.ID 而非硬编码
	})
	require.NoError(t, err)

	items := repo.itemsByModel("b1", "model1")
	assert.Len(t, items, 1, "preset→custom 替换后仍应只有一个 item")
	assert.Equal(t, rbacdomain.PermissionTypeCustom, items[0].GrantType)
}

func TestBindCustomItem_PermissionNotFound_Fails(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	svc := newItemSvc(repo, false)

	_, err := svc.BindCustomItem(context.Background(), BindCustomItemToBundleCommand{
		ProjectScope:       project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:           "b1",
		ModelID:            "model1",
		CustomPermissionID: "nonexistent",
	})

	require.Error(t, err)
}

func TestRemoveDataPermissionItem_Success(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	preset := rbacdomain.PresetReadAll
	repo.bundleItems["b1"] = []*rbacdomain.EndUserBundleDataPermissionItem{
		{ID: "item1", BundleID: "b1", ModelID: "model1", GrantType: rbacdomain.PermissionTypePreset, Preset: &preset},
	}
	svc := newItemSvc(repo, false)

	_, err := svc.RemoveDataPermissionItemFromBundle(context.Background(), RemoveDataPermissionItemFromBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
	})

	require.NoError(t, err)
	assert.Empty(t, repo.itemsByModel("b1", "model1"), "移除后应为空")
}

// ─── 5.2: snapshot item payload 验证 ────────────────────────────────────────

func TestBindPresetItem_CreatesSnapshotWithItemPayload(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	svc := newItemSvc(repo, false)

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})
	require.NoError(t, err)

	snap := repo.lastSnapshot()
	require.NotNil(t, snap, "应自动创建快照")
	require.Len(t, snap.Items, 1)
	assert.Equal(t, "model1", snap.Items[0].ModelID)
	assert.Equal(t, rbacdomain.PermissionTypePreset, snap.Items[0].GrantType)
	assert.Equal(t, rbacdomain.PresetReadAll, *snap.Items[0].Preset)
	assert.Nil(t, snap.Items[0].CustomPermissionID)
}

func TestBindCustomItem_CreatesSnapshotWithCustomPayload(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	perm := makeTestCustomPerm("perm1", "org1", "model1")
	repo := newBundleItemRepoWithPerms(bundle, map[string][]*rbacdomain.EndUserPermission{
		"model1": {perm},
	})
	svc := newItemSvc(repo, false)

	_, err := svc.BindCustomItem(context.Background(), BindCustomItemToBundleCommand{
		ProjectScope:       project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:           "b1",
		ModelID:            "model1",
		CustomPermissionID: perm.ID,
	})
	require.NoError(t, err)

	snap := repo.lastSnapshot()
	require.NotNil(t, snap)
	require.Len(t, snap.Items, 1)
	assert.Equal(t, rbacdomain.PermissionTypeCustom, snap.Items[0].GrantType)
	assert.Nil(t, snap.Items[0].Preset)
	require.NotNil(t, snap.Items[0].CustomPermissionID)
	assert.Equal(t, perm.ID, *snap.Items[0].CustomPermissionID)
}

func TestRemoveItem_CreatesEmptySnapshot(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	preset := rbacdomain.PresetReadAll
	repo.bundleItems["b1"] = []*rbacdomain.EndUserBundleDataPermissionItem{
		{ID: "item1", BundleID: "b1", ModelID: "model1", GrantType: rbacdomain.PermissionTypePreset, Preset: &preset},
	}
	svc := newItemSvc(repo, false)

	_, err := svc.RemoveDataPermissionItemFromBundle(context.Background(), RemoveDataPermissionItemFromBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
	})
	require.NoError(t, err)

	snap := repo.lastSnapshot()
	require.NotNil(t, snap, "移除后应自动创建快照")
	assert.Empty(t, snap.Items, "移除后快照 item 列表应为空")
}

func TestRestoreBundle_RebuildsItemsFromSnapshot(t *testing.T) {
	bundle := makeTestBundle("b1", "org1", "proj1")
	repo := newBundleItemRepo(bundle)
	preset := rbacdomain.PresetReadAll
	repo.snapshotsByVersion[1] = &rbacdomain.BundleSnapshot{
		ID:       "snap-v1",
		BundleID: "b1",
		Version:  1,
		Items: []rbacdomain.SnapshotItemEntry{
			{ModelID: "model1", GrantType: rbacdomain.PermissionTypePreset, Preset: &preset},
		},
	}
	repo.currentVersion = 1
	svc := newItemSvc(repo, false)

	result, err := svc.RestoreBundle(context.Background(), RestoreBundleCommand{
		OrgName:       "org1",
		BundleID:      "b1",
		TargetVersion: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	items := repo.bundleItems["b1"]
	require.Len(t, items, 1, "回滚后应恢复 1 个 item")
	assert.Equal(t, "model1", items[0].ModelID)
	assert.Equal(t, rbacdomain.PermissionTypePreset, items[0].GrantType)
}
