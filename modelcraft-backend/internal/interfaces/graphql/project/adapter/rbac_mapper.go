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
	return &generated.EndUserPermission{
		ID:           p.ID,
		ModelID:      p.ModelID,
		Action:       generated.RbacAction(p.Action),
		ColumnPolicy: ToColumnPolicyDTO(p.ColumnPolicy),
		RowScope:     generated.RowScopeType(p.RowScope),
		DisplayName:  displayName,
		Description:  p.Description,
		CreatedAt:    time.Now(), // populated from DB if needed
		UpdatedAt:    time.Now(),
	}
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
	return &generated.EndUserPermissionBundle{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		Permissions: entries,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
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
