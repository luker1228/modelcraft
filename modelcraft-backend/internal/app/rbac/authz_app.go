package rbac

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modeldesign"

	rbacdomain "modelcraft/internal/domain/rbac"
)

// EndUserAuthzService 鉴权应用服务（编排 prd/rbac/03-auth-flow.md 中的 5 步鉴权流程）
type EndUserAuthzService struct {
	rbacRepo  rbacdomain.EndUserPermissionRepository
	modelRepo modelRepository
}

// NewEndUserAuthzService 创建鉴权服务。可选传入 modelRepo 以支持 PRESET item 运行时展开。
func NewEndUserAuthzService(
	rbacRepo rbacdomain.EndUserPermissionRepository,
	modelRepo ...modeldesign.ModelRepository,
) *EndUserAuthzService {
	svc := &EndUserAuthzService{rbacRepo: rbacRepo}
	if len(modelRepo) > 0 {
		svc.modelRepo = modelRepo[0]
	}
	return svc
}

// GetEffectivePermissions 获取用户在指定 Project 下的有效权限集。
//
// 步骤：
//
//	Step 1: 查 end_user_user_bundles（直接授权）
//	Step 2: 查 end_user_role_users → end_user_role_bundles（显式角色，单次 JOIN）
//	Step 3: 查 end_user_roles WHERE is_implicit=true（隐式角色，对所有认证用户执行）
//	Step 4: 展开所有 data permission items（CUSTOM 加载实体，PRESET 运行时展开）
//	Step 5: 合并取并集（rowScope 取最宽泛范围）
func (s *EndUserAuthzService) GetEffectivePermissions(
	ctx context.Context,
	q GetEffectivePermissionsQuery,
) (rbacdomain.EffectivePermissionSet, error) {
	directIDs, err := s.rbacRepo.GetBundleIDsByUserDirect(ctx, q.UserID, q.OrgName, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step1 failed for user %s: %w", q.UserID, err)
	}

	explicitIDs, err := s.rbacRepo.GetBundleIDsByUserExplicitRoles(ctx, q.UserID, q.OrgName, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step2 failed for user %s: %w", q.UserID, err)
	}

	implicitIDs, err := s.rbacRepo.GetBundleIDsByImplicitRoles(ctx, q.OrgName, q.ProjectSlug)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step3 failed for project %s/%s: %w", q.OrgName, q.ProjectSlug, err)
	}

	allBundleIDs := deduplicateStrings(append(append(directIDs, explicitIDs...), implicitIDs...))
	if len(allBundleIDs) == 0 {
		return rbacdomain.EffectivePermissionSet{}, nil
	}

	permissions, err := s.expandBundleItemsToPermissions(ctx, q.OrgName, allBundleIDs)
	if err != nil {
		return nil, fmt.Errorf("rbac authz step4 failed: %w", err)
	}

	eps := rbacdomain.EffectivePermissionSet{}
	return eps.Merge(permissions), nil
}

// expandBundleItemsToPermissions 将 bundle ID 列表展开为权限列表。
// CUSTOM item 加载 custom permission 实体；PRESET item 运行时展开（需 modelRepo）。
func (s *EndUserAuthzService) expandBundleItemsToPermissions(
	ctx context.Context,
	orgName string,
	bundleIDs []string,
) ([]*rbacdomain.EndUserPermission, error) {
	perms, err := s.rbacRepo.GetPermissionsByBundleIDs(ctx, orgName, bundleIDs)
	if err != nil {
		return nil, err
	}
	if s.modelRepo == nil {
		return perms, nil
	}
	presetPerms, err := s.expandPresetItems(ctx, orgName, bundleIDs)
	if err != nil {
		return nil, err
	}
	return append(perms, presetPerms...), nil
}

// expandPresetItems 遍历 bundle 列表，对 PRESET item 运行时展开为虚拟权限点。
func (s *EndUserAuthzService) expandPresetItems(
	ctx context.Context,
	orgName string,
	bundleIDs []string,
) ([]*rbacdomain.EndUserPermission, error) {
	result := make([]*rbacdomain.EndUserPermission, 0)
	for _, bundleID := range bundleIDs {
		items, err := s.rbacRepo.ListBundleDataPermissionItems(ctx, bundleID)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item.GrantType != rbacdomain.PermissionTypePreset || item.Preset == nil {
				continue
			}
			perm, expandErr := s.expandPresetItem(ctx, orgName, bundleID, item)
			if expandErr != nil {
				continue // owner field 缺失等情况跳过，不中断其他 item
			}
			result = append(result, perm)
		}
	}
	return result, nil
}

// expandPresetItem 将单个 PRESET item 展开为虚拟 EndUserPermission（不落库）。
func (s *EndUserAuthzService) expandPresetItem(
	ctx context.Context,
	orgName, bundleID string,
	item *rbacdomain.EndUserBundleDataPermissionItem,
) (*rbacdomain.EndUserPermission, error) {
	ownerField, _ := s.tryGetOwnerField(ctx, item.ModelID)
	rowPolicy, err := expandPreset(*item.Preset, ownerField)
	if err != nil {
		return nil, err
	}
	presetCopy := *item.Preset
	return &rbacdomain.EndUserPermission{
		ID:           fmt.Sprintf("preset:%s:%s:%s", bundleID, item.ModelID, presetCopy),
		OrgName:      orgName,
		ModelID:      item.ModelID,
		Name:         presetPermissionName(presetCopy),
		Type:         rbacdomain.PermissionTypePreset,
		Preset:       &presetCopy,
		ColumnPolicy: nil,
		RowPolicy:    rowPolicy,
	}, nil
}

func (s *EndUserAuthzService) tryGetOwnerField(ctx context.Context, modelID string) (string, error) {
	if s.modelRepo == nil {
		return "", nil
	}
	model, err := s.modelRepo.GetByID(ctx, modelID, modeldesign.NewModelQueryOptions().WithFields())
	if err != nil || model == nil {
		return "", err
	}
	if model.GetOwnerField() == nil {
		return "", nil
	}
	return model.GetOwnerField().Name, nil
}

// deduplicateStrings 简单去重工具函数。
func deduplicateStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}
