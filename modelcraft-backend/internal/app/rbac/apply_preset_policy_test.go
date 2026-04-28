package rbac

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"sort"
	"testing"

	domainproject "modelcraft/internal/domain/project"
	rbacdomain "modelcraft/internal/domain/rbac"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPermissionRepo struct {
	byModel map[string][]*rbacdomain.EndUserPermission
}

func newMockPermissionRepo(initial []*rbacdomain.EndUserPermission) *mockPermissionRepo {
	repo := &mockPermissionRepo{byModel: map[string][]*rbacdomain.EndUserPermission{}}
	for _, p := range initial {
		repo.byModel[p.ModelID] = append(repo.byModel[p.ModelID], clonePermission(p))
	}
	return repo
}

func (m *mockPermissionRepo) CreatePermission(
	_ context.Context,
	p *rbacdomain.EndUserPermission,
) error {
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

func (m *mockPermissionRepo) UpdatePermission(
	_ context.Context,
	p *rbacdomain.EndUserPermission,
) error {
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

func (m *mockPermissionRepo) DeletePermission(
	_ context.Context,
	orgName, id string,
) error {
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

func (m *mockPermissionRepo) DeletePresetPermissionsByModel(
	_ context.Context,
	orgName, modelID string,
) error {
	list := m.byModel[modelID]
	kept := make([]*rbacdomain.EndUserPermission, 0, len(list))
	for _, p := range list {
		if p.OrgName == orgName && p.Type == rbacdomain.PermissionTypePreset {
			continue
		}
		kept = append(kept, p)
	}
	m.byModel[modelID] = kept
	return nil
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

func TestApplyPresetPolicy(t *testing.T) {
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

	t.Run("正常 apply READ_WRITE_ALL", func(t *testing.T) {
		svc := makeService(makeModelWithOwner(false), []*rbacdomain.EndUserPermission{
			makeCustomPermission(orgName, projectSlug, modelID, "custom-1"),
			makePresetPermission(
				orgName,
				projectSlug,
				modelID,
				"preset-old-1",
				rbacdomain.PresetReadAll,
			),
			makePresetPermission(
				orgName,
				projectSlug,
				modelID,
				"preset-old-2",
				rbacdomain.PresetReadWriteOwner,
			),
		})

		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: domainproject.ProjectScope{
				OrgName:     orgName,
				ProjectSlug: projectSlug,
			},
			ModelID: modelID,
			Preset:  rbacdomain.PresetReadWriteAll,
		})
		require.NoError(t, err)
		require.Len(t, perms, 2)
		assert.Equal(t, "custom-1", perms[0].Name)
		assert.Equal(t, "preset:READ_WRITE_ALL", perms[1].Name)
		assert.Equal(t, rbacdomain.PermissionTypePreset, perms[1].Type)
	})

	t.Run("正常 apply READ_ALL（无 END_USER_REF 也可）", func(t *testing.T) {
		svc := makeService(makeModelWithOwner(false), []*rbacdomain.EndUserPermission{
			makeCustomPermission(orgName, projectSlug, modelID, "custom-1"),
		})

		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: domainproject.ProjectScope{
				OrgName:     orgName,
				ProjectSlug: projectSlug,
			},
			ModelID: modelID,
			Preset:  rbacdomain.PresetReadAll,
		})
		require.NoError(t, err)
		require.Len(t, perms, 2)
		assert.Equal(t, "preset:READ_ALL", perms[1].Name)
	})

	t.Run("apply READ_WRITE_OWNER，模型有 END_USER_REF 成功", func(t *testing.T) {
		svc := makeService(makeModelWithOwner(true), nil)

		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: domainproject.ProjectScope{
				OrgName:     orgName,
				ProjectSlug: projectSlug,
			},
			ModelID: modelID,
			Preset:  rbacdomain.PresetReadWriteOwner,
		})
		require.NoError(t, err)
		require.Len(t, perms, 1)
		assert.Equal(t, "preset:READ_WRITE_OWNER", perms[0].Name)
	})

	t.Run("apply READ_WRITE_OWNER，模型无 END_USER_REF 返回错误", func(t *testing.T) {
		origin := []*rbacdomain.EndUserPermission{
			makeCustomPermission(orgName, projectSlug, modelID, "custom-keep"),
		}
		svc := makeService(makeModelWithOwner(false), origin)

		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: domainproject.ProjectScope{
				OrgName:     orgName,
				ProjectSlug: projectSlug,
			},
			ModelID: modelID,
			Preset:  rbacdomain.PresetReadWriteOwner,
		})
		require.Error(t, err)
		assert.Nil(t, perms)
		var bizErr *bizerrors.BusinessError
		require.True(t, bizerrors.As(err, &bizErr))
		assert.Equal(
			t,
			bizerrors.EndUserPresetRequiresOwnerField.GetCode(),
			bizErr.Info().GetCode(),
		)

		kept, listErr := svc.rbacRepo.ListPermissionsByModel(context.Background(), orgName, modelID)
		require.NoError(t, listErr)
		require.Len(t, kept, 1)
		assert.Equal(t, "custom-keep", kept[0].Name)
	})

	t.Run("apply 后 CUSTOM 权限点不受影响", func(t *testing.T) {
		svc := makeService(makeModelWithOwner(true), []*rbacdomain.EndUserPermission{
			makeCustomPermission(orgName, projectSlug, modelID, "custom-1"),
			makeCustomPermission(orgName, projectSlug, modelID, "custom-2"),
			makePresetPermission(
				orgName,
				projectSlug,
				modelID,
				"preset-old",
				rbacdomain.PresetReadAll,
			),
		})

		perms, err := svc.ApplyPresetPolicy(context.Background(), ApplyPresetPolicyCommand{
			ProjectScope: domainproject.ProjectScope{
				OrgName:     orgName,
				ProjectSlug: projectSlug,
			},
			ModelID: modelID,
			Preset:  rbacdomain.PresetReadAllWriteOwner,
		})
		require.NoError(t, err)
		require.Len(t, perms, 3)
		assert.Equal(t, "custom-1", perms[0].Name)
		assert.Equal(t, "custom-2", perms[1].Name)
		assert.Equal(t, "preset:READ_ALL_WRITE_OWNER", perms[2].Name)
	})

	t.Run("重复 apply 同一预设幂等", func(t *testing.T) {
		svc := makeService(makeModelWithOwner(true), nil)
		cmd := ApplyPresetPolicyCommand{
			ProjectScope: domainproject.ProjectScope{
				OrgName:     orgName,
				ProjectSlug: projectSlug,
			},
			ModelID: modelID,
			Preset:  rbacdomain.PresetReadWriteOwner,
		}

		_, err := svc.ApplyPresetPolicy(context.Background(), cmd)
		require.NoError(t, err)
		perms, err := svc.ApplyPresetPolicy(context.Background(), cmd)
		require.NoError(t, err)
		require.Len(t, perms, 1)
		assert.Equal(t, "preset:READ_WRITE_OWNER", perms[0].Name)
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
		assert.Equal(t, rbacdomain.ScopeAll, rp.Select.Scope)
		assert.Equal(t, rbacdomain.ScopeAll, rp.Insert.Scope)
		assert.Equal(t, rbacdomain.ScopeAll, rp.Update.Scope)
		assert.Equal(t, rbacdomain.ScopeAll, rp.Delete.Scope)
		assert.NoError(t, rp.Validate())
	})

	t.Run("READ_ALL", func(t *testing.T) {
		rp, err := expandPreset(rbacdomain.PresetReadAll, "")
		require.NoError(t, err)
		require.NotNil(t, rp)
		assert.True(t, rp.Select.Allowed)
		assert.False(t, rp.Insert.Allowed)
		assert.False(t, rp.Update.Allowed)
		assert.False(t, rp.Delete.Allowed)
		assert.Equal(t, rbacdomain.ScopeAll, rp.Select.Scope)
		assert.NoError(t, rp.Validate())
	})

	t.Run("READ_WRITE_OWNER", func(t *testing.T) {
		rp, err := expandPreset(rbacdomain.PresetReadWriteOwner, "owner")
		require.NoError(t, err)
		require.NotNil(t, rp)
		assert.True(t, rp.Select.Allowed)
		assert.True(t, rp.Insert.Allowed)
		assert.True(t, rp.Update.Allowed)
		assert.True(t, rp.Delete.Allowed)
		assert.Equal(t, rbacdomain.ScopeCustom, rp.Select.Scope)
		assert.Equal(t, rbacdomain.ScopeCustom, rp.Insert.Scope)
		assert.Equal(t, rbacdomain.ScopeCustom, rp.Update.Scope)
		assert.Equal(t, rbacdomain.ScopeCustom, rp.Delete.Scope)
		assert.NoError(t, rp.Validate())
	})

	t.Run("READ_ALL_WRITE_OWNER", func(t *testing.T) {
		rp, err := expandPreset(rbacdomain.PresetReadAllWriteOwner, "owner")
		require.NoError(t, err)
		require.NotNil(t, rp)
		assert.True(t, rp.Select.Allowed)
		assert.True(t, rp.Insert.Allowed)
		assert.True(t, rp.Update.Allowed)
		assert.True(t, rp.Delete.Allowed)
		assert.Equal(t, rbacdomain.ScopeAll, rp.Select.Scope)
		assert.Equal(t, rbacdomain.ScopeCustom, rp.Insert.Scope)
		assert.Equal(t, rbacdomain.ScopeCustom, rp.Update.Scope)
		assert.Equal(t, rbacdomain.ScopeCustom, rp.Delete.Scope)
		assert.NoError(t, rp.Validate())
	})
}

func TestExpandPreset_OwnerField(t *testing.T) {
	t.Run("owner preset requires owner field", func(t *testing.T) {
		_, err := expandPreset(rbacdomain.PresetReadWriteOwner, "")
		require.Error(t, err)
		var bizErr *bizerrors.BusinessError
		require.True(t, bizerrors.As(err, &bizErr))
		assert.Equal(t, bizerrors.EndUserPresetRequiresOwnerField.GetCode(), bizErr.Info().GetCode())
	})

	t.Run("non-owner preset ignores owner field", func(t *testing.T) {
		rp, err := expandPreset(rbacdomain.PresetReadAll, "owner")
		require.NoError(t, err)
		require.NotNil(t, rp)
		assert.True(t, rp.Select.Allowed)
		assert.False(t, rp.Insert.Allowed)
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

func makeCustomPermission(
	orgName, projectSlug, modelID, name string,
) *rbacdomain.EndUserPermission {
	return &rbacdomain.EndUserPermission{
		ID:           name + "-id",
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		ModelID:      modelID,
		Name:         name,
		Type:         rbacdomain.PermissionTypeCustom,
		ColumnPolicy: nil,
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
		ID:           name + "-id",
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		ModelID:      modelID,
		Name:         name,
		Type:         rbacdomain.PermissionTypePreset,
		ColumnPolicy: nil,
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
