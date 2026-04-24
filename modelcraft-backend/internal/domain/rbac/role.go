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
	// Bundles 按需填充
	Bundles []*EndUserPermissionBundle
}

// GuardDelete 隐式角色删除保护（领域规则）
func (r *EndUserRole) GuardDelete() error {
	if r.IsImplicit {
		return bizerrors.NewValidationError(
			"rbac.role.implicit_protected: implicit role cannot be deleted",
		)
	}
	return nil
}

// GuardUpdate 隐式角色更新保护：name 不可更新（领域规则）
func (r *EndUserRole) GuardUpdate() error {
	if r.IsImplicit {
		return bizerrors.NewValidationError(
			"rbac.role.implicit_name_immutable: implicit role name cannot be updated",
		)
	}
	return nil
}
