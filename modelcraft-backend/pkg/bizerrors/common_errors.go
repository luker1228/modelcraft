package bizerrors

import "fmt"

// 通用错误定义
var (
	NotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound,
		EnMessage: "Resource not found",
		ZhMessage: "资源不存在",
	}

	ParamInvalid = ErrorDefinition{
		Code:      ErrorTypeParamInvalid,
		EnMessage: "Invalid parameter: {0}",
		ZhMessage: "参数无效: {0}",
	}

	OperationDenied = ErrorDefinition{
		Code:      ErrorTypeOperationFailed,
		EnMessage: "Operation failed: {0}",
		ZhMessage: "操作失败: {0}",
	}

	Conflict = ErrorDefinition{
		Code:      ErrorTypeConflict,
		EnMessage: "Resource conflict: {0}",
		ZhMessage: "资源冲突: {0}",
	}

	SystemError = ErrorDefinition{
		Code:      ErrorTypeSystemError,
		EnMessage: "System error: {0}",
		ZhMessage: "系统错误: {0}",
	}

	// AuthUnauthorized 表示认证失败（缺少有效凭证、token 过期/无效等）
	AuthUnauthorized = ErrorDefinition{
		Code:      ErrorTypeUnauthorized,
		EnMessage: "Authentication required: {0}",
		ZhMessage: "认证失败: {0}",
	}

	// AuthenticationFailed 登录失败（手机号不存在、密码错误）
	AuthenticationFailed = ErrorDefinition{
		Code:      ErrorTypeAuthentication,
		EnMessage: "Authentication failed: {0}",
		ZhMessage: "认证失败: {0}",
	}

	// AuthParamInvalid 注册/登录参数校验失败
	AuthParamInvalid = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".AUTH",
		EnMessage: "Invalid auth parameter: {0}",
		ZhMessage: "认证参数无效: {0}",
	}
)

// 定义Model领域错误
var (
	ModelNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".MODEL",
		EnMessage: "Model Resource not found, Id={0}",
		ZhMessage: "模型资源不存在，标识{0}",
	}

	ModelAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".MODEL",
		EnMessage: "Model already exists: {0}",
		ZhMessage: "模型资源已存在: {0}",
	}

	// ModelTableAlreadyExists is returned when the underlying database table already exists
	// before the model is created. Users should use importModel to import the existing table.
	ModelTableAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".MODEL.TABLE",
		EnMessage: "Table '{0}' already exists in database '{1}'",
		ZhMessage: "表 '{0}' 在数据库 '{1}' 中已存在",
	}
)

// 定义Project领域错误
var (
	ProjectNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".PROJECT",
		EnMessage: "Project not found: {0}",
		ZhMessage: "项目不存在: {0}",
	}

	ProjectAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".PROJECT",
		EnMessage: "Project already exists: {0}",
		ZhMessage: "项目已存在: {0}",
	}

	CannotDeleteDefaultProject = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".PROJECT",
		EnMessage: "Cannot delete the default project",
		ZhMessage: "无法删除默认项目",
	}
)

// 定义ModelGroup领域错误
var (
	GroupNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".GROUP",
		EnMessage: "Group not found: {0}",
		ZhMessage: "分组不存在: {0}",
	}

	GroupAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".GROUP",
		EnMessage: "Group already exists: {0}",
		ZhMessage: "分组已存在: {0}",
	}

	InvalidGroupName = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".GROUP",
		EnMessage: "Invalid group name: {0}",
		ZhMessage: "无效的分组名称: {0}",
	}
)

// 定义Field领域错误
var (
	FieldNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".FIELD",
		EnMessage: "Field not found: {0}",
		ZhMessage: "字段不存在: {0}",
	}

	FieldHasDependencies = ErrorDefinition{
		Code: ErrorTypeOperationFailed + ".RELATION",
		EnMessage: "Cannot delete field '{0}' because it has dependent relations. " +
			"Please delete the following relations first: {1}",
		ZhMessage: "无法删除字段 '{0}'，因为存在依赖的关联关系。请先删除其依赖的关联关系: {1}",
	}
)

// 定义数据库集群相关错误
var (
	ClusterAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".CLUSTER",
		EnMessage: "Database cluster already exists: {0}",
		ZhMessage: "数据库集群已存在: {0}",
	}

	ClusterNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".CLUSTER",
		EnMessage: "Database cluster not found: {0}",
		ZhMessage: "数据库集群不存在: {0}",
	}

	DatabaseConnectionFailed = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".DB_CONNECTION",
		EnMessage: "Database connection failed: {0}",
		ZhMessage: "数据库连接失败: {0}",
	}

	ProjectAlreadyHasCluster = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".CLUSTER",
		EnMessage: "Project '{0}' already has a cluster. Each project can only have one cluster.",
		ZhMessage: "项目 '{0}' 已经有集群。每个项目只能有一个集群。",
	}
)

// 定义Enum领域错误
var (
	EnumNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".ENUM",
		EnMessage: "Enum not found: {0}",
		ZhMessage: "枚举不存在: {0}",
	}

	EnumAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".ENUM",
		EnMessage: "Enum already exists: {0}",
		ZhMessage: "枚举已存在: {0}",
	}

	CannotDeleteReferencedEnum = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".ENUM",
		EnMessage: "Cannot delete enum '{0}', it is referenced by fields: {1}",
		ZhMessage: "无法删除枚举 '{0}'，它被以下字段引用: {1}",
	}
)

var (
	FieldAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".FIELD",
		EnMessage: "Field already exists: {0}",
		ZhMessage: "字段已经存在：{0}",
	}
	RecordNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".RECORD",
		EnMessage: "Record Resource not found",
		ZhMessage: "记录不存在",
	}

	ConnectFailed = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".CONNECT",
		EnMessage: "Connect Database Failed details:{0}",
		ZhMessage: "连接数据库失败，详情: {0}",
	}

	TimeOut = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".TIME_OUT",
		EnMessage: "Request timeout, Please retry or improve index",
		ZhMessage: "请求超时，请重试或者优化索引",
	}

	DuplicateKey = ErrorDefinition{
		Code:      ErrorTypeConflict + ".RECORD",
		EnMessage: "Duplicate key, details:{0}",
		ZhMessage: "重复数据，详情：{0}",
	}

	DatabaseAccessDenied = ErrorDefinition{
		Code:      ErrorTypeAuthentication + ".DATABASE",
		ZhMessage: "Database auth failed, Please check your account and password",
		EnMessage: "数据库认证失败，请检查用户名和密码",
	}
)

// 定义用户领域错误
var (
	UserNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".USER",
		EnMessage: "User not found: {0}",
		ZhMessage: "用户不存在: {0}",
	}

	UserAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".USER",
		EnMessage: "User already exists: {0}",
		ZhMessage: "用户已存在: {0}",
	}
)

// 定义用户资料（Profile）领域错误
var (
	ProfileNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".PROFILE",
		EnMessage: "Profile not found: {0}",
		ZhMessage: "用户资料不存在: {0}",
	}

	InvalidProfileInput = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".PROFILE",
		EnMessage: "Invalid profile input: {0}",
		ZhMessage: "用户资料参数无效: {0}",
	}
)

// 定义组织领域错误
var (
	OrganizationNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".ORGANIZATION",
		EnMessage: "Organization not found: {0}",
		ZhMessage: "组织不存在: {0}",
	}

	OrganizationAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".ORGANIZATION",
		EnMessage: "Organization already exists: {0}",
		ZhMessage: "组织已存在: {0}",
	}

	OrganizationLimitExceeded = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".ORGANIZATION",
		EnMessage: "Organization limit exceeded: {0}",
		ZhMessage: "组织数量超出限制: {0}",
	}
)

// 定义角色领域错误
var (
	RoleNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".ROLE",
		EnMessage: "Role not found: {0}",
		ZhMessage: "角色不存在: {0}",
	}

	RoleAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".ROLE",
		EnMessage: "Role already exists: {0}",
		ZhMessage: "角色已存在: {0}",
	}

	CannotDeleteSystemRole = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".ROLE",
		EnMessage: "Cannot delete system role: {0}",
		ZhMessage: "无法删除系统角色: {0}",
	}
)

// 定义成员关系领域错误
var (
	MembershipNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".MEMBERSHIP",
		EnMessage: "Membership not found for user in organization",
		ZhMessage: "用户不属于该组织",
	}

	MembershipAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".MEMBERSHIP",
		EnMessage: "User is already a member of this organization",
		ZhMessage: "用户已是该组织成员",
	}

	PermissionDenied = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".PERMISSION",
		EnMessage: "Permission denied: {0}",
		ZhMessage: "权限不足: {0}",
	}
)

// 定义 LogicalForeignKey 领域错误
var (
	FKNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".FK",
		EnMessage: "Logical foreign key not found: {0}",
		ZhMessage: "逻辑外键不存在: {0}",
	}

	FKColumnsNotFound = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".FK",
		EnMessage: "FK columns not found on model: {0}",
		ZhMessage: "FK 列在模型中不存在: {0}",
	}

	FKPairHasRelateFields = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".FK",
		EnMessage: "Cannot delete FK pair: RELATION fields still reference it: {0}",
		ZhMessage: "无法删除 FK 对：仍有 RELATION 字段引用它: {0}",
	}

	FKFieldCountMismatch = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".FK.FIELD_COUNT",
		EnMessage: "source_fields and target_fields count must match",
		ZhMessage: "source_fields 和 target_fields 数量必须一致",
	}

	FKNotDeletable = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".FK.NOT_DELETABLE",
		EnMessage: "Logical foreign key cannot be deleted: {0}",
		ZhMessage: "逻辑外键不可删除: {0}",
	}
)

// 定义 APIKey 领域错误
var (
	APIKeyNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".API_KEY",
		EnMessage: "API key not found: {0}",
		ZhMessage: "API Key 不存在: {0}",
	}

	APIKeyLimitExceeded = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".API_KEY",
		EnMessage: "API key limit exceeded: maximum {0} keys per user",
		ZhMessage: "API Key 数量超出限制：每用户最多 {0} 个",
	}
)

// 定义 EndUser（终端用户）领域错误
var (
	// EndUserInvalidCredentials 凭证无效（用户名不存在或密码错误，统一返回防止用户枚举）
	EndUserInvalidCredentials = ErrorDefinition{
		Code:      "INVALID_CREDENTIALS.END_USER",
		EnMessage: "Invalid credentials",
		ZhMessage: "用户名或密码错误",
	}

	// EndUserInvalidRefreshToken 刷新令牌无效（未找到、已过期或已撤销）
	EndUserInvalidRefreshToken = ErrorDefinition{
		Code:      "INVALID_REFRESH_TOKEN.END_USER",
		EnMessage: "Invalid or expired refresh token",
		ZhMessage: "刷新令牌无效或已过期",
	}

	// EndUserAccountDisabled 账号已被禁用
	EndUserAccountDisabled = ErrorDefinition{
		Code:      "ACCOUNT_DISABLED.END_USER",
		EnMessage: "Account is disabled",
		ZhMessage: "账号已被禁用",
	}

	// EndUserConflict 终端用户已存在（用户名重复）
	EndUserConflict = ErrorDefinition{
		Code:      ErrorTypeConflict + ".END_USER",
		EnMessage: "End user already exists: {0}",
		ZhMessage: "终端用户已存在: {0}",
	}

	// EndUserParamInvalid 终端用户参数无效（用户名格式错误、密码强度不足等）
	EndUserParamInvalid = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".END_USER",
		EnMessage: "Invalid end user parameter: {0}",
		ZhMessage: "终端用户参数无效: {0}",
	}

	// EndUserNotFound 终端用户不存在
	EndUserNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".END_USER",
		EnMessage: "End user not found: {0}",
		ZhMessage: "终端用户不存在: {0}",
	}

	// EndUserClusterNotConfigured 项目未配置数据库集群
	EndUserClusterNotConfigured = ErrorDefinition{
		Code:      "CLUSTER_NOT_CONFIGURED.END_USER",
		EnMessage: "Database cluster not configured for this project",
		ZhMessage: "该项目未配置数据库集群",
	}

	// EndUserPrivateDBNotInitialized 私有库未初始化（mc_private_{projectSlug} 不存在）
	EndUserPrivateDBNotInitialized = ErrorDefinition{
		Code:      "PRIVATE_DB_NOT_INITIALIZED.END_USER",
		EnMessage: "Private database is not initialized for this project",
		ZhMessage: "该项目私有库尚未初始化",
	}

	// EndUserNoProjectAccess 账号无任何项目访问权限
	EndUserNoProjectAccess = ErrorDefinition{
		Code:      "NO_PROJECT_ACCESS.END_USER",
		EnMessage: "No accessible projects for this account",
		ZhMessage: "该账号暂无可访问项目",
	}

	// EndUserProjectAccessDenied 当前项目未授权
	EndUserProjectAccessDenied = ErrorDefinition{
		Code:      "PROJECT_ACCESS_DENIED.END_USER",
		EnMessage: "No access to project: {0}",
		ZhMessage: "无该项目访问权限: {0}",
	}
)

// 定义 RLS 领域错误
var (
	// ModelHasNoOwnerField 模型没有 owner 字段
	ModelHasNoOwnerField = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".RLS.NO_OWNER",
		EnMessage: "Model has no owner field: {0}",
		ZhMessage: "模型缺少归属字段: {0}",
	}

	// EndUserRefAlreadyExists 模型已存在 EndUserRef 字段
	EndUserRefAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".FIELD.END_USER_REF",
		EnMessage: "EndUserRef field already exists in this model",
		ZhMessage: "每个模型只允许一个归属字段",
	}

	// InvalidRLSExpression RLS 表达式无效
	InvalidRLSExpression = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".RLS.EXPR",
		EnMessage: "Invalid RLS expression: {0}",
		ZhMessage: "RLS 表达式无效: {0}",
	}

	// InvalidAuthVariable 认证变量无效
	InvalidAuthVariable = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".RLS.AUTH_VAR",
		EnMessage: "Invalid auth variable: {0}",
		ZhMessage: "认证变量无效: {0}",
	}

	// RLSCheckViolation RLS CHECK 约束违反
	RLSCheckViolation = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".RLS.CHECK_VIOLATION",
		EnMessage: "RLS check violation: {0}",
		ZhMessage: "违反 RLS 策略约束: {0}",
	}

	// PermissionDeniedRLS RLS 权限拒绝
	PermissionDeniedRLS = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".RLS.PERMISSION_DENIED",
		EnMessage: "Permission denied by RLS policy: {0}",
		ZhMessage: "RLS 策略拒绝访问: {0}",
	}
)

// 定义 RBAC（数据行列级权限）领域错误
var (
	// EndUserPermissionNotFound 权限点不存在
	EndUserPermissionNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".RBAC.PERMISSION",
		EnMessage: "End user permission not found: {0}",
		ZhMessage: "权限点不存在: {0}",
	}

	// EndUserPermissionBundleNotFound 权限包不存在
	EndUserPermissionBundleNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".RBAC.BUNDLE",
		EnMessage: "End user permission bundle not found: {0}",
		ZhMessage: "权限包不存在: {0}",
	}

	// EndUserPermissionBundleAlreadyExists 权限包名称重复
	EndUserPermissionBundleAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".RBAC.BUNDLE",
		EnMessage: "End user permission bundle already exists: {0}",
		ZhMessage: "权限包已存在: {0}",
	}

	// EndUserRoleNotFound RBAC 角色不存在
	EndUserRoleNotFound = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".RBAC.ROLE",
		EnMessage: "End user role not found: {0}",
		ZhMessage: "RBAC 角色不存在: {0}",
	}

	// EndUserRoleAlreadyExists RBAC 角色名称重复
	EndUserRoleAlreadyExists = ErrorDefinition{
		Code:      ErrorTypeConflict + ".RBAC.ROLE",
		EnMessage: "End user role already exists: {0}",
		ZhMessage: "RBAC 角色已存在: {0}",
	}

	// EndUserImplicitRoleCannotBeModified 内置隐式角色不可修改或删除
	EndUserImplicitRoleCannotBeModified = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".RBAC.IMPLICIT_ROLE",
		EnMessage: "Implicit role cannot be modified or deleted: {0}",
		ZhMessage: "内置隐式角色不可修改或删除: {0}",
	}

	// EndUserCannotAssignImplicitRole 不可手动分配隐式角色
	EndUserCannotAssignImplicitRole = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".RBAC.ASSIGN_IMPLICIT",
		EnMessage: "Implicit role cannot be manually assigned to users",
		ZhMessage: "内置隐式角色不可手动分配给用户",
	}

	// UserBundleAlreadyAssigned 用户已绑定该权限包
	UserBundleAlreadyAssigned = ErrorDefinition{
		Code:      ErrorTypeConflict + ".RBAC.USER_BUNDLE",
		EnMessage: "User already has this permission bundle assigned",
		ZhMessage: "用户已绑定该权限包",
	}

	// UserRoleAlreadyAssigned 用户已绑定该角色
	UserRoleAlreadyAssigned = ErrorDefinition{
		Code:      ErrorTypeConflict + ".RBAC.USER_ROLE",
		EnMessage: "User already has this role assigned",
		ZhMessage: "用户已绑定该角色",
	}

	// EndUserNotFoundInProject 终端用户在项目中不存在（RBAC 上下文）
	EndUserNotFoundInProject = ErrorDefinition{
		Code:      ErrorTypeNotFound + ".RBAC.END_USER",
		EnMessage: "End user not found in project: {0}",
		ZhMessage: "终端用户在项目中不存在: {0}",
	}

	// EndUserPermissionBundleInUse 权限包已被角色或用户绑定，无法删除
	EndUserPermissionBundleInUse = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".RBAC.BUNDLE_IN_USE",
		EnMessage: "Permission bundle is still in use and cannot be deleted: {0}",
		ZhMessage: "权限包仍被使用中，无法删除: {0}",
	}

	// EndUserPermissionInUse 权限点已被权限包引用，无法删除
	EndUserPermissionInUse = ErrorDefinition{
		Code:      ErrorTypeOperationFailed + ".RBAC.PERMISSION_IN_USE",
		EnMessage: "Permission is still referenced by bundles and cannot be deleted: {0}",
		ZhMessage: "权限点仍被权限包引用，无法删除: {0}",
	}

	// EndUserRowScopeFieldMissing rowScope 要求 Model 上存在特定字段，但字段不存在
	EndUserRowScopeFieldMissing = ErrorDefinition{
		Code:      ErrorTypeParamInvalid + ".RBAC.ROW_SCOPE_FIELD",
		EnMessage: "Row scope '{0}' requires field '{1}' on the model, but it does not exist",
		ZhMessage: "行策略 '{0}' 要求模型存在字段 '{1}'，但该字段不存在",
	}
)

// AllErrorDefinitions 返回所有错误定义（用于测试）
func AllErrorDefinitions() []ErrorDefinition {
	return []ErrorDefinition{
		NotFound,
		ParamInvalid,
		OperationDenied,
		Conflict,
		SystemError,
	}
}

// NewValidationError 创建一个参数验证错误
func NewValidationError(message string, params ...any) error {
	return NewError(ParamInvalid, fmt.Sprintf(message, params...))
}
