package adapter

import (
	"modelcraft/internal/interfaces/graphql/project/generated"
	"strings"
	"time"

	rbacdomain "modelcraft/internal/domain/rbac"
)

// ─── Domain → GraphQL DTO ─────────────────────────────────────────────────────

// ToEndUserPermissionDTO converts domain EndUserPermission to GraphQL DTO.
func ToEndUserPermissionDTO(p *rbacdomain.EndUserPermission) *generated.EndUserPermission {
	if p == nil {
		return nil
	}
	var displayName *string
	if p.Name != "" {
		n := p.Name
		displayName = &n
	}
	action, rowScope := deriveLegacyActionAndScope(p)
	dto := &generated.EndUserPermission{
		ID:           p.ID,
		ModelID:      p.ModelID,
		DatabaseName: p.DatabaseName,
		ModelName:    p.ModelName,
		Action:       action,
		ColumnPolicy: ToColumnPolicyDTO(p.ColumnPolicy),
		RowScope:     rowScope,
		Preset:       toPresetDTO(p.Preset),
		DisplayName:  displayName,
		Description:  p.Description,
		CreatedAt:    time.Now(), // populated from DB if needed
		UpdatedAt:    time.Now(),
	}

	return dto
}

func toPresetDTO(preset *rbacdomain.PermissionPreset) *generated.EndUserPermissionPreset {
	if preset == nil {
		return nil
	}
	value := generated.EndUserPermissionPreset(*preset)
	return &value
}

func deriveLegacyActionAndScope(p *rbacdomain.EndUserPermission) (generated.RbacAction, generated.RowScopeType) {
	if p == nil || p.RowPolicy == nil {
		return generated.RbacActionSelect, generated.RowScopeTypeAll
	}
	policy := p.RowPolicy
	policy.Normalize()

	if policy.Select.Allowed {
		return generated.RbacActionSelect, scopeToLegacyRowScope(policy.Select.Scope)
	}
	if policy.Insert.Allowed {
		return generated.RbacActionInsert, scopeToLegacyRowScope(policy.Insert.Scope)
	}
	if policy.Update.Allowed {
		return generated.RbacActionUpdate, scopeToLegacyRowScope(policy.Update.Scope)
	}
	if policy.Delete.Allowed {
		return generated.RbacActionDelete, scopeToLegacyRowScope(policy.Delete.Scope)
	}
	return generated.RbacActionSelect, generated.RowScopeTypeAll
}

func scopeToLegacyRowScope(scope rbacdomain.PolicyScope) generated.RowScopeType {
	if scope == rbacdomain.ScopeCustom {
		return generated.RowScopeTypeSelf
	}
	return generated.RowScopeTypeAll
}

// ToEndUserPermissionBundleDTO converts domain bundle to GraphQL DTO.
func ToEndUserPermissionBundleDTO(b *rbacdomain.EndUserPermissionBundle) *generated.EndUserPermissionBundle {
	if b == nil {
		return nil
	}
	entries := make([]*generated.EndUserBundlePermissionEntry, 0, len(b.Permissions))
	for i, p := range b.Permissions {
		entries = append(entries, &generated.EndUserBundlePermissionEntry{
			SortOrder:  int32(i),
			Permission: ToEndUserPermissionDTO(p),
		})
	}

	// Item-centric data permission items
	itemDTOs := make(
		[]*generated.EndUserBundleDataPermissionItem,
		0,
		len(b.Items),
	)
	for _, item := range b.Items {
		itemDTOs = append(itemDTOs, ToDataPermissionItemDTO(item))
	}

	// 计算 currentVersion（最新快照的版本号，无快照时为 0）
	var currentVersion int32
	if len(b.Snapshots) > 0 {
		currentVersion = int32(b.Snapshots[0].Version)
	}

	// 转换快照列表
	snapshots := make([]*generated.EndUserPermissionBundleSnapshot, 0, len(b.Snapshots))
	for _, s := range b.Snapshots {
		snapshots = append(snapshots, toBundleSnapshotDTO(&s))
	}

	return &generated.EndUserPermissionBundle{
		ID:                  b.ID,
		Slug:                b.Slug,
		Name:                b.Name,
		Description:         b.Description,
		DataPermissionItems: itemDTOs,
		Permissions:         entries,
		CurrentVersion:      currentVersion,
		Snapshots:           snapshots,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// ToDataPermissionItemDTO converts domain item to GraphQL DTO.
func ToDataPermissionItemDTO(
	item *rbacdomain.EndUserBundleDataPermissionItem,
) *generated.EndUserBundleDataPermissionItem {
	if item == nil {
		return nil
	}
	dto := &generated.EndUserBundleDataPermissionItem{
		ID:        item.ID,
		BundleID:  item.BundleID,
		ModelID:   item.ModelID,
		GrantType: generated.DataPermissionGrantType(item.GrantType),
		SortOrder: int32(item.SortOrder),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
	if item.Preset != nil {
		v := generated.EndUserPermissionPreset(*item.Preset)
		dto.Preset = &v
	}
	if item.CustomPermissionID != nil {
		dto.CustomPermissionID = item.CustomPermissionID
	}
	return dto
}

// toBundleSnapshotDTO converts a domain BundleSnapshot to its GraphQL DTO.
// The snapshot's `permissions` field contains only ID references; the `permission` object
// itself is not populated here since it requires a DB lookup for deleted permission handling.
// In practice, the field resolver in gqlgen resolves `permission` from the struct directly.
func toBundleSnapshotDTO(s *rbacdomain.BundleSnapshot) *generated.EndUserPermissionBundleSnapshot {
	if s == nil {
		return nil
	}
	permEntries := make([]*generated.EndUserPermissionSnapshotEntry, 0, len(s.Permissions))
	for _, p := range s.Permissions {
		permEntries = append(permEntries, &generated.EndUserPermissionSnapshotEntry{
			SortOrder:    int32(p.SortOrder),
			PermissionID: p.PermissionID,
			Permission:   nil,
		})
	}

	// Item-centric snapshot entries
	itemEntries := make(
		[]*generated.EndUserPermissionSnapshotItemEntry,
		0,
		len(s.Items),
	)
	for _, item := range s.Items {
		entry := &generated.EndUserPermissionSnapshotItemEntry{
			ModelID:   item.ModelID,
			GrantType: generated.DataPermissionGrantType(item.GrantType),
			SortOrder: int32(item.SortOrder),
		}
		if item.Preset != nil {
			v := generated.EndUserPermissionPreset(*item.Preset)
			entry.Preset = &v
		}
		entry.CustomPermissionID = item.CustomPermissionID
		itemEntries = append(itemEntries, entry)
	}

	var restoredFrom *int32
	if s.RestoredFrom != nil {
		v := int32(*s.RestoredFrom)
		restoredFrom = &v
	}
	return &generated.EndUserPermissionBundleSnapshot{
		Version:      int32(s.Version),
		CreatedAt:    s.CreatedAt,
		CreatedBy:    s.CreatedBy,
		RestoredFrom: restoredFrom,
		Items:        itemEntries,
		Permissions:  permEntries,
	}
}

// ToEndUserRoleDTO converts domain role to GraphQL DTO.
func ToEndUserRoleDTO(r *rbacdomain.EndUserRole) *generated.EndUserRole {
	if r == nil {
		return nil
	}
	entries := make([]*generated.EndUserRoleBundleEntry, 0, len(r.Bundles))
	for _, b := range r.Bundles {
		entries = append(entries, &generated.EndUserRoleBundleEntry{
			Bundle:     ToEndUserPermissionBundleDTO(b),
			AssignedAt: time.Now(),
		})
	}
	return &generated.EndUserRole{
		ID:                r.ID,
		Name:              r.Name,
		Description:       r.Description,
		IsImplicit:        r.IsImplicit,
		PermissionBundles: entries,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// ToColumnPolicyDTO converts domain ColumnPolicy to GraphQL DTO.
func ToColumnPolicyDTO(cp *rbacdomain.ColumnPolicy) *generated.ColumnPolicy {
	if cp == nil {
		// 默认策略：全列可见
		return &generated.ColumnPolicy{
			DefaultMode: generated.ColumnAccessModeVisible,
			Rules:       []*generated.ColumnRule{},
		}
	}
	rules := make([]*generated.ColumnRule, 0, len(cp.Rules))
	for _, r := range cp.Rules {
		rule := &generated.ColumnRule{
			FieldName: r.FieldName,
			Mode:      generated.ColumnAccessMode(r.Mode),
		}
		if r.MaskPattern != "" {
			rule.MaskPattern = &r.MaskPattern
		}
		rules = append(rules, rule)
	}
	return &generated.ColumnPolicy{
		DefaultMode: generated.ColumnAccessMode(cp.DefaultMode),
		Rules:       rules,
	}
}

// ─── GraphQL Input → Domain ───────────────────────────────────────────────────

// ToRbacActionDomain converts GraphQL RbacAction (uppercase "SELECT") to domain Action (lowercase "select").
// GraphQL enum values are uppercase by convention; DB ENUM is lowercase.
func ToRbacActionDomain(a generated.RbacAction) rbacdomain.Action {
	return rbacdomain.Action(strings.ToLower(string(a)))
}

// ToRowScopeDomain converts GraphQL RowScopeType to domain RowScope.
func ToRowScopeDomain(r generated.RowScopeType) rbacdomain.RowScope {
	return rbacdomain.RowScope(string(r))
}

// ToColumnPolicyDomain converts GraphQL ColumnPolicyInput to domain ColumnPolicy.
func ToColumnPolicyDomain(input *generated.ColumnPolicyInput) *rbacdomain.ColumnPolicy {
	if input == nil {
		return nil
	}
	rules := make([]rbacdomain.ColumnRule, 0, len(input.Rules))
	for _, r := range input.Rules {
		rule := rbacdomain.ColumnRule{
			FieldName: r.FieldName,
			Mode:      rbacdomain.ColumnAccessMode(r.Mode),
		}
		if r.MaskPattern != nil {
			rule.MaskPattern = *r.MaskPattern
		}
		rules = append(rules, rule)
	}
	return &rbacdomain.ColumnPolicy{
		DefaultMode: rbacdomain.ColumnAccessMode(input.DefaultMode),
		Rules:       rules,
	}
}

// ToEndUserPermissionsDTO converts a slice of domain permissions to DTOs.
func ToEndUserPermissionsDTO(perms []*rbacdomain.EndUserPermission) []*generated.EndUserPermission {
	result := make([]*generated.EndUserPermission, 0, len(perms))
	for _, p := range perms {
		result = append(result, ToEndUserPermissionDTO(p))
	}
	return result
}

// ToEndUserRolesDTO converts a slice of domain roles to DTOs.
func ToEndUserRolesDTO(roles []*rbacdomain.EndUserRole) []*generated.EndUserRole {
	result := make([]*generated.EndUserRole, 0, len(roles))
	for _, r := range roles {
		result = append(result, ToEndUserRoleDTO(r))
	}
	return result
}

// ToEndUserBundlesDTO converts a slice of domain bundles to DTOs.
func ToEndUserBundlesDTO(bundles []*rbacdomain.EndUserPermissionBundle) []*generated.EndUserPermissionBundle {
	result := make([]*generated.EndUserPermissionBundle, 0, len(bundles))
	for _, b := range bundles {
		result = append(result, ToEndUserPermissionBundleDTO(b))
	}
	return result
}
