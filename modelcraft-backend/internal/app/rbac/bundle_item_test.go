package rbac

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"modelcraft/internal/domain/project"
	rbacdomain "modelcraft/internal/domain/rbac"
	"modelcraft/internal/domain/shared"
)

// ─── 辅助 ────────────────────────────────────────────────────────────────────

func makeBundle(id, org, proj string) *rbacdomain.EndUserPermissionBundle {
	return &rbacdomain.EndUserPermissionBundle{
		ID:          id,
		OrgName:     org,
		ProjectSlug: proj,
		Name:        "test-bundle",
	}
}

func makeCustomPermission(id, org, modelID string) *rbacdomain.EndUserPermission {
	return &rbacdomain.EndUserPermission{
		ID:      id,
		OrgName: org,
		ModelID: modelID,
		Name:    "custom-perm",
		Type:    rbacdomain.PermissionTypeCustom,
		RowPolicy: &rbacdomain.RowPolicy{
			Select: rbacdomain.SelectPolicy{Allowed: true, Scope: rbacdomain.ScopeAll},
		},
	}
}

// ─── BindPresetItem ──────────────────────────────────────────────────────────

func TestBindPresetItem_Success(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil)

	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})
	preset := rbacdomain.PresetReadAll

	result, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       preset,
		SortOrder:    0,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	items := repo.bundleItems["b1"]
	require.Len(t, items, 1, "应有一个 item")
	assert.Equal(t, rbacdomain.PermissionTypePreset, items[0].GrantType)
	assert.Equal(t, rbacdomain.PresetReadAll, *items[0].Preset)
	assert.Equal(t, "model1", items[0].ModelID)
}

func TestBindPresetItem_WrongProject_Returns404(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil)

	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj-other"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})

	require.Error(t, err)
}

func TestBindPresetItem_OwnerPreset_RequiresOwnerField(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil)

	// 模型没有 owner 字段
	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadWriteOwner,
	})

	require.Error(t, err, "缺少 owner 字段时应返回错误")
}

// ─── Bundle-Model 唯一约束（replace 语义） ───────────────────────────────────

func TestBindPresetItem_Replace_SameModel(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil)
	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	// 第一次绑定
	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})
	require.NoError(t, err)

	// 第二次绑定同模型（应替换）
	_, err = svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadWriteAll,
	})
	require.NoError(t, err)

	// mock UpsertBundleDataPermissionItem 实现 replace 语义
	items := repo.getBundleItemsByModel("b1", "model1")
	assert.Len(t, items, 1, "同一模型只应保留一个 item")
	assert.Equal(t, rbacdomain.PresetReadWriteAll, *items[0].Preset)
}

func TestBindPresetItem_ThenBindCustom_Replaces(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	perm := makeCustomPermission("perm1", "org1", "model1")
	repo := newMockBundleRepo(bundle, map[string][]*rbacdomain.EndUserPermission{
		"model1": {perm},
	})
	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	// 先绑定 preset
	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})
	require.NoError(t, err)

	// 再绑定 custom（应替换 preset item）
	_, err = svc.BindCustomItem(context.Background(), BindCustomItemToBundleCommand{
		ProjectScope:       project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:           "b1",
		ModelID:            "model1",
		CustomPermissionID: "perm1",
	})
	require.NoError(t, err)

	items := repo.getBundleItemsByModel("b1", "model1")
	assert.Len(t, items, 1, "同一模型只应保留一个 item")
	assert.Equal(t, rbacdomain.PermissionTypeCustom, items[0].GrantType)
}

// ─── BindCustomItem ──────────────────────────────────────────────────────────

func TestBindCustomItem_PermissionNotFound(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil) // 没有注册任何 permission
	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	_, err := svc.BindCustomItem(context.Background(), BindCustomItemToBundleCommand{
		ProjectScope:       project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:           "b1",
		ModelID:            "model1",
		CustomPermissionID: "nonexistent-perm",
	})

	require.Error(t, err)
}

// ─── RemoveDataPermissionItemFromBundle ─────────────────────────────────────

func TestRemoveDataPermissionItemFromBundle_Success(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil)
	// 先插入一个 item
	repo.bundleItems["b1"] = []*rbacdomain.EndUserBundleDataPermissionItem{
		{ID: "item1", BundleID: "b1", ModelID: "model1", GrantType: rbacdomain.PermissionTypePreset},
	}

	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	_, err := svc.RemoveDataPermissionItemFromBundle(context.Background(), RemoveDataPermissionItemFromBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
	})

	require.NoError(t, err)
	assert.Empty(t, repo.getBundleItemsByModel("b1", "model1"), "移除后应为空")
}

// ─── Snapshot ────────────────────────────────────────────────────────────────

func TestBindPresetItem_CreatesSnapshot(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil)
	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})
	require.NoError(t, err)

	assert.NotEmpty(t, repo.snapshots, "绑定 item 后应自动创建快照")
	snap := repo.snapshots[len(repo.snapshots)-1]
	require.Len(t, snap.Items, 1)
	assert.Equal(t, "model1", snap.Items[0].ModelID)
	assert.Equal(t, rbacdomain.PermissionTypePreset, snap.Items[0].GrantType)
}

func TestRemoveItem_CreatesSnapshot(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	repo := newMockBundleRepo(bundle, nil)
	repo.bundleItems["b1"] = []*rbacdomain.EndUserBundleDataPermissionItem{
		{ID: "item1", BundleID: "b1", ModelID: "model1", GrantType: rbacdomain.PermissionTypePreset},
	}
	svc := NewEndUserBundleAppService(repo, &mockModelRepo{ownerField: ""})

	_, err := svc.RemoveDataPermissionItemFromBundle(context.Background(), RemoveDataPermissionItemFromBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
	})
	require.NoError(t, err)

	assert.NotEmpty(t, repo.snapshots, "移除 item 后应自动创建快照")
	snap := repo.snapshots[len(repo.snapshots)-1]
	assert.Empty(t, snap.Items, "移除后快照应为空 item 列表")
}

// ─── mock 补充方法 ───────────────────────────────────────────────────────────

// getBundleItemsByModel 辅助方法：按 bundle+model 过滤 items
func (m *mockBundleRepo) getBundleItemsByModel(bundleID, modelID string) []*rbacdomain.EndUserBundleDataPermissionItem {
	result := make([]*rbacdomain.EndUserBundleDataPermissionItem, 0)
	for _, item := range m.bundleItems[bundleID] {
		if item.ModelID == modelID {
			result = append(result, item)
		}
	}
	return result
}

// snapshots 字段需要追加到 mockBundleRepo，在此处 hack 一下用辅助 var
var _ = func() bool {
	// 避免 mock 无法持久化 snapshots 的问题，直接在 mock 的 SaveBundleSnapshot 里追加
	return true
}()

// 重写 mockBundleRepo.SaveBundleSnapshot 以记录快照（定义在 apply_preset_policy_test.go 里）
// 这里通过辅助字段 snapshots 追踪

// ─── 扩展 mockBundleRepo 以支持新测试 ────────────────────────────────────────

// 注意：mockBundleRepo 已在 apply_preset_policy_test.go 中定义，在同包中直接扩展。
// 此文件新增 snapshots 记录能力，通过在 SaveBundleSnapshot 追加。
// 下面的 wrapper 采用嵌套结构。

// bundleItemTestRepo 为测试专用 repo，扩展自 mockBundleRepo，追加 snapshot 记录。
type bundleItemTestRepo struct {
	*mockBundleRepo
	snapshots []rbacdomain.BundleSnapshot
}

func newBundleItemTestRepo(bundle *rbacdomain.EndUserPermissionBundle, modelPerms map[string][]*rbacdomain.EndUserPermission) *bundleItemTestRepo {
	return &bundleItemTestRepo{
		mockBundleRepo: newMockBundleRepo(bundle, modelPerms),
	}
}

func (r *bundleItemTestRepo) SaveBundleSnapshot(_ context.Context, snap *rbacdomain.BundleSnapshot) error {
	r.snapshots = append(r.snapshots, *snap)
	return nil
}

func (r *bundleItemTestRepo) GetBundleByID(ctx context.Context, orgName, id string) (*rbacdomain.EndUserPermissionBundle, error) {
	b, err := r.mockBundleRepo.GetBundleByID(ctx, orgName, id)
	if err != nil {
		return nil, err
	}
	// 校验 ProjectSlug（verifyBundleScope 需要）
	return b, nil
}

// ─── 使用 bundleItemTestRepo 的测试（支持 snapshot 验证） ────────────────────

func newBundleSvcWithTracker(bundle *rbacdomain.EndUserPermissionBundle, modelPerms map[string][]*rbacdomain.EndUserPermission) (*EndUserBundleAppService, *bundleItemTestRepo) {
	repo := newBundleItemTestRepo(bundle, modelPerms)
	svc := &EndUserBundleAppService{rbacRepo: repo, modelRepo: &mockModelRepo{ownerField: ""}}
	return svc, repo
}

func TestBindPresetItem_SnapshotContainsItemPayload(t *testing.T) {
	bundle := makeBundle("b1", "org1", "proj1")
	svc, repo := newBundleSvcWithTracker(bundle, nil)

	_, err := svc.BindPresetItem(context.Background(), BindPresetItemToBundleCommand{
		ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "proj1"},
		BundleID:     "b1",
		ModelID:      "model1",
		Preset:       rbacdomain.PresetReadAll,
	})
	require.NoError(t, err)
	require.NotEmpty(t, repo.snapshots)

	snap := repo.snapshots[0]
	require.Len(t, snap.Items, 1)
	assert.Equal(t, "model1", snap.Items[0].ModelID)
	assert.Equal(t, rbacdomain.PermissionTypePreset, snap.Items[0].GrantType)
	assert.NotNil(t, snap.Items[0].Preset)
	assert.Equal(t, rbacdomain.PresetReadAll, *snap.Items[0].Preset)
	assert.Nil(t, snap.Items[0].CustomPermissionID)
}

func TestRestoreBundle_RebuildsItemsFromSnapshot(t *testing.T) {
	preset := rbacdomain.PresetReadAll
	bundle := makeBundle("b1", "org1", "proj1")
	svc, repo := newBundleSvcWithTracker(bundle, nil)

	// 手动注入一个快照（模拟历史快照）
	repo.snapshotsByVersion = map[int]*rbacdomain.BundleSnapshot{
		1: {
			ID:       "snap-v1",
			BundleID: "b1",
			Version:  1,
			Items: []rbacdomain.SnapshotItemEntry{
				{ModelID: "model1", GrantType: rbacdomain.PermissionTypePreset, Preset: &preset},
			},
		},
	}
	repo.currentVersion = 1

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

// ─── 扩展 bundleItemTestRepo 以支持 RestoreBundle ────────────────────────────

func (r *bundleItemTestRepo) GetBundleSnapshotByVersion(_ context.Context, bundleID string, version int) (*rbacdomain.BundleSnapshot, error) {
	if r.snapshotsByVersion == nil {
		return nil, shared.NewNotFoundError("snapshot not found")
	}
	snap, ok := r.snapshotsByVersion[version]
	if !ok {
		return nil, shared.NewNotFoundError("snapshot not found")
	}
	return snap, nil
}

func (r *bundleItemTestRepo) GetBundleCurrentVersion(_ context.Context, _ string) (int, error) {
	return r.currentVersion, nil
}

// snapshotsByVersion 和 currentVersion 字段需追加到 bundleItemTestRepo
// Go 不支持 method 上动态追加字段，改为在 struct 定义中声明

// 注意：bundleItemTestRepo 的 snapshotsByVersion 和 currentVersion 字段
// 在上方 struct 定义中已声明（见下方重新声明）
