package rbac

import "modelcraft/pkg/bizerrors"

// EndUserPermissionBundle 权限包（系统中唯一正式授权单位）
// 用户和角色只能关联权限包，不能直接关联权限点
type EndUserPermissionBundle struct {
	OrgName     string
	ProjectSlug string
	ID          string
	Name        string
	Description *string
	// Permissions 旧字段（兼容）：按需填充。
	Permissions []*EndUserPermission
	// Items 新字段：bundle 的正式数据权限绑定项。
	Items []*EndUserBundleDataPermissionItem
	// Snapshots 历史版本快照，按需填充，最多 5 个（按 version DESC）
	Snapshots []BundleSnapshot
}

// Validate 校验权限包合法性
func (b *EndUserPermissionBundle) Validate() error {
	if b.Name == "" {
		return bizerrors.NewValidationError("rbac.bundle.name_required: bundle name is required")
	}
	return nil
}
