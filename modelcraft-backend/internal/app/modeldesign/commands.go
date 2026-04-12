package modeldesign

import "modelcraft/internal/domain/modeldesign"

// CreateModelCommand 创建模型命令
type CreateModelCommand struct {
	OrgName      string  // 组织名称（从 URL 路径或 JWT 中提取）
	ProjectSlug  string  // 项目标识符
	Name         string
	Title        string
	Description  string
	StorageType  string
	DatabaseName string
	DisplayField *string // 用于 runtime __label 解析的字段名
}

// UpdateModelMetaCommand 更新模型元数据命令
type UpdateModelMetaCommand struct {
	OrgName     string  // 组织名称（从 URL 路径或 JWT 中提取）
	ProjectSlug string  // 项目标识符
	Title       *string
	Description *string
	DisplayField *string // 用于 runtime __label 解析的字段名（nil 表示不更新）
}

// ModelQueryCommand 模型查询命令
type ModelQueryCommand struct {
	OrgName          string // 组织名称（从 URL 路径或 JWT 中提取）
	ProjectSlug      string // 项目标识符
	DatabaseName     string
	Name             string
	Title            string
	StorageType      string
	Status           string
	IncludeFields    bool
	IncludeRelations bool
	Page             int
	PageSize         int
}

// ============================================================================
// Field Commands
// ============================================================================

// AddFieldCommand 添加字段命令
// Note: Fields contains domain FieldDefinitions as conversion from GraphQL/HTTP
// input is complex and happens in the interface layer via adapter/mapper
type AddFieldCommand struct {
	ModelID string                         // 模型ID
	Fields  []*modeldesign.FieldDefinition // 要添加的字段列表
}

// UpdateFieldCommand 更新字段命令
type UpdateFieldCommand struct {
	ModelID          string                        // 模型ID
	FieldName        string                        // 字段名称
	Title            string                        // 新的显示名称（空字符串表示不更新）
	Description      string                        // 新的描述（空字符串表示不更新）
	ValidationConfig *modeldesign.ValidationConfig // 验证配置（nil表示不更新）
}

// DeprecateFieldCommand 废弃字段命令（幂等：已废弃时直接成功）
type DeprecateFieldCommand struct {
	ModelID   string // 模型ID
	FieldName string // 字段名称
}

// UndeprecateFieldCommand 解除字段废弃命令（幂等：未废弃时直接成功）
type UndeprecateFieldCommand struct {
	ModelID   string // 模型ID
	FieldName string // 字段名称
}

// RemoveFieldCommand 删除字段命令
type RemoveFieldCommand struct {
	ModelID   string // 模型ID
	FieldName string // 字段名称
}

// GetFieldsCommand 获取字段列表命令
type GetFieldsCommand struct {
	ModelID string // 模型ID
}

// GetModelOptions 获取模型选项
type GetModelOptions struct {
	GetFields bool
}

// NewGetModelOptions 创建默认获取模型选项
func NewGetModelOptions() *GetModelOptions {
	return &GetModelOptions{
		GetFields: true,
	}
}

// ============================================================================
// ModelGroup Commands
// ============================================================================

// CreateGroupCommand creates a new model group within a project.
type CreateGroupCommand struct {
	OrgName     string
	ProjectSlug string
	Name        string
}

// RenameGroupCommand renames an existing model group.
type RenameGroupCommand struct {
	OrgName     string
	ProjectSlug string
	GroupID     string
	NewName     string
}

// DeleteGroupCommand deletes a model group, cascading its models to ungrouped.
type DeleteGroupCommand struct {
	GroupID string
}

// ReorderGroupCommand moves a group to a new position.
// AfterGroupID is the ID of the group that the moved group should follow.
// Set AfterGroupID to nil to move the group to the head of the list.
type ReorderGroupCommand struct {
	OrgName      string
	ProjectSlug  string
	GroupID      string
	AfterGroupID *string
}

// MoveModelToGroupCommand assigns a model to a group.
// Set GroupID to nil to move the model to the virtual ungrouped state.
type MoveModelToGroupCommand struct {
	ModelID string
	GroupID *string
}

// ============================================================================
// Enum Commands
// ============================================================================

// CreateEnumCommand 创建枚举命令
type CreateEnumCommand struct {
	OrgName       string                   // 组织名称
	ProjectSlug   string                   // 项目标识符
	Name          string                   // 枚举标识名称（项目内唯一）
	DisplayName   string                   // 显示名称
	Description   string                   // 描述
	Options       []modeldesign.EnumOption // 枚举选项列表
	IsMultiSelect bool                     // 是否支持多选
}

// UpdateEnumCommand 更新枚举命令
type UpdateEnumCommand struct {
	OrgName     string                   // 组织名称
	ProjectSlug string                   // 项目标识符
	Name        string                   // 枚举标识名称（用于查找）
	DisplayName *string                  // 新的显示名称（nil表示不更新）
	Description *string                  // 新的描述（nil表示不更新）
	Options     []modeldesign.EnumOption // 新的选项列表（nil表示不更新）
}

// DeleteEnumCommand 删除枚举命令
type DeleteEnumCommand struct {
	OrgName     string // 组织名称
	ProjectSlug string // 项目标识符
	Name        string // 枚举标识名称
}

// GetEnumCommand 获取枚举命令
type GetEnumCommand struct {
	OrgName     string // 组织名称
	ProjectSlug string // 项目标识符
	Name        string // 枚举标识名称
}

// ListEnumsCommand 列举枚举命令
type ListEnumsCommand struct {
	OrgName     string // 组织名称
	ProjectSlug string // 项目标识符
}

// GetEnumReferencesCommand 获取枚举引用命令
type GetEnumReferencesCommand struct {
	OrgName     string // 组织名称
	ProjectSlug string // 项目标识符
	Name        string // 枚举标识名称
}

// ValidateEnumCodesCommand 验证枚举代码命令
type ValidateEnumCodesCommand struct {
	OrgName     string   // 组织名称
	ProjectSlug string   // 项目标识符
	EnumName    string   // 枚举名称
	Codes       []string // 待验证的代码列表
}

// ============================================================================
// LogicalForeignKey Commands
// ============================================================================

// CreateLogicalForeignKeyCommand 创建逻辑外键命令
type CreateLogicalForeignKeyCommand struct {
	ProjectSlug  string   // 项目标识符
	ModelID      string   // 拥有 FK 列的模型 ID
	RefModelID   string   // 被引用的模型 ID
	SourceFields []string // FK 列（属于 ModelID 的字段名）
	TargetFields []string // 被引用列（属于 RefModelID 的字段名）
}

// DeleteLogicalForeignKeyCommand 删除逻辑外键命令
type DeleteLogicalForeignKeyCommand struct {
	ProjectSlug string // 项目标识符
	PairID      string // FK 对的 pair_id
}

// ListLogicalForeignKeysCommand 查询模型的逻辑外键列表命令
type ListLogicalForeignKeysCommand struct {
	ProjectSlug string // 项目标识符
	ModelID     string // 模型 ID
}
