package adapter

import (
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
)

// BizCodeToResourceType 将业务错误码映射到 GraphQL ResourceType 枚举。
// 仅处理 NOT_FOUND.* 前缀的错误码；未能匹配的统一兜底为 ResourceTypeUnknown。
//
// 调用方应先通过 err.Info().IsNotFoundError() 确认错误类型，再调用本函数。
func BizCodeToResourceType(code string) generated.ResourceType {
	switch code {
	case bizerrors.ModelNotFound.GetCode():
		return generated.ResourceTypeModel
	case bizerrors.ProjectNotFound.GetCode():
		return generated.ResourceTypeProject
	case bizerrors.ClusterNotFound.GetCode():
		return generated.ResourceTypeCluster
	case bizerrors.EnumNotFound.GetCode():
		return generated.ResourceTypeEnum
	case bizerrors.GroupNotFound.GetCode():
		return generated.ResourceTypeGroup
	case bizerrors.UserNotFound.GetCode():
		return generated.ResourceTypeUser
	case bizerrors.ProfileNotFound.GetCode():
		return generated.ResourceTypeProfile
	case bizerrors.OrganizationNotFound.GetCode():
		return generated.ResourceTypeOrganization
	case bizerrors.RoleNotFound.GetCode():
		return generated.ResourceTypeRole
	case bizerrors.EndUserNotFound.GetCode():
		return generated.ResourceTypeEndUser
	case bizerrors.EndUserPermissionNotFound.GetCode():
		return generated.ResourceTypeEndUserPermission
	case bizerrors.EndUserPermissionBundleNotFound.GetCode():
		return generated.ResourceTypeEndUserPermissionBundle
	case bizerrors.EndUserPermissionBundleSnapshotNotFound.GetCode():
		return generated.ResourceTypeEndUserPermissionBundleSnapshot
	case bizerrors.EndUserRoleNotFound.GetCode():
		return generated.ResourceTypeEndUserRole
	case bizerrors.EndUserNotFoundInProject.GetCode():
		return generated.ResourceTypeEndUserInProject
	default:
		// NOT_FOUND（无子类型）或未收录的细粒度资源（FIELD、FK、MEMBERSHIP 等）
		// 统一兜底为 UNKNOWN；extensions.code 仍保留原始细粒度码供排障使用
		return generated.ResourceTypeUnknown
	}
}
