package rbac

import "modelcraft/pkg/bizerrors"

// EndUserRole 业务角色（Project 维度）
// 与 Casbin `roles` 表完全独立：roles = Org 维度系统角色，end_user_roles = Project 维度业务角色
type EndUserRole struct {
	OrgName     string
	ProjectSlug string
	ID          string
	Name        string
	Description *string
	// IsImplicit 内置隐式角色标志
	// true: 角色落库，但 end_user_role_users 表中不为每个用户插入关联行；
	//       鉴权时由系统自动注入（Step 3）；不可删除；name 不可更新
	IsImplicit bool
	// IsProtected 受保护角色标志
	// true: 角色可见、可显式分配给用户；但不可删除、不可改名、不可修改权限包关联。
	// 典型用途：每个 Project 自动创建的 admin 角色。
	IsProtected bool
	// Bundles 按需填充
	Bundles []*EndUserPermissionBundle
}

// GuardDelete 隐式/受保护角色删除保护（领域规则）
func (r *EndUserRole) GuardDelete() error {
	if r.IsImplicit {
		return bizerrors.NewValidationError(
			"rbac.role.implicit_protected: implicit role cannot be deleted",
		)
	}
	if r.IsProtected {
		return bizerrors.NewValidationError(
			"rbac.role.protected: protected role cannot be deleted",
		)
	}
	return nil
}

// GuardUpdate 隐式/受保护角色更新保护：name/description 不可更新（领域规则）
func (r *EndUserRole) GuardUpdate() error {
	if r.IsImplicit {
		return bizerrors.NewValidationError(
			"rbac.role.implicit_name_immutable: implicit role name cannot be updated",
		)
	}
	if r.IsProtected {
		return bizerrors.NewValidationError(
			"rbac.role.protected_immutable: protected role cannot be updated",
		)
	}
	return nil
}

// GuardBundleModify 受保护角色权限包修改保护（领域规则）
// implicit 角色允许修改权限包；protected 角色不允许。
func (r *EndUserRole) GuardBundleModify() error {
	if r.IsProtected {
		return bizerrors.NewValidationError(
			"rbac.role.protected_bundle_immutable: protected role bundles cannot be modified",
		)
	}
	return nil
}
