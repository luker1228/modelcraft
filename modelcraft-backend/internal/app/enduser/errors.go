package enduser

import "modelcraft/pkg/bizerrors"

// ---------------------------------------------------------------------------
// End-User 领域错误定义
// 注意：这些定义正常应添加到 pkg/bizerrors/common_errors.go
// 这里临时定义以便 App 层编译通过，需要 Domain/Infra 层 owner 将其迁移到正确位置
// ---------------------------------------------------------------------------

var (
	// EndUserNotFound 终端用户不存在
	EndUserNotFound = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeNotFound + ".END_USER",
		EnMessage: "End user not found: {0}",
		ZhMessage: "终端用户不存在: {0}",
	}

	// EndUserConflict 终端用户已存在（用户名冲突）
	EndUserConflict = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeConflict + ".END_USER",
		EnMessage: "End user already exists: {0}",
		ZhMessage: "终端用户已存在: {0}",
	}

	// EndUserParamInvalid 参数无效（用户名/密码格式错误）
	EndUserParamInvalid = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeParamInvalid + ".END_USER",
		EnMessage: "Invalid end user parameter: {0}",
		ZhMessage: "终端用户参数无效: {0}",
	}

	// EndUserClusterNotConfigured Project 未配置数据库集群
	EndUserClusterNotConfigured = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeOperationFailed + ".CLUSTER_NOT_CONFIGURED",
		EnMessage: "Database cluster not configured for this project: {0}",
		ZhMessage: "项目未配置数据库集群: {0}",
	}

	// EndUserAccountDisabled 账号已禁用
	EndUserAccountDisabled = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeOperationFailed + ".ACCOUNT_DISABLED",
		EnMessage: "End user account is disabled",
		ZhMessage: "终端用户账号已禁用",
	}

	// ErrBuiltinUserCannotBeDeleted is returned when attempting to delete a builtin admin EndUser.
	ErrBuiltinUserCannotBeDeleted = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeOperationFailed + ".BUILTIN_USER_DELETE",
		EnMessage: "Built-in admin user cannot be deleted",
		ZhMessage: "内置管理员账号不可删除",
	}

	// ErrBuiltinUserCannotBeDisabled is returned when attempting to disable a builtin admin EndUser.
	ErrBuiltinUserCannotBeDisabled = bizerrors.ErrorDefinition{
		Code:      bizerrors.ErrorTypeOperationFailed + ".BUILTIN_USER_DISABLE",
		EnMessage: "Built-in admin user cannot be disabled",
		ZhMessage: "内置管理员账号不可禁用",
	}
)
