package modeldesign

import (
	"modelcraft/internal/domain/project"
	"strings"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// ModelLocator 模型定位器 - 用于唯一定位一个模型。
// 包含完整的项目上下文（OrgName + ProjectSlug）以确保数据完整性。
type ModelLocator struct {
	project.ProjectScope        // 嵌入项目作用域，包含 OrgName 和 ProjectSlug
	DatabaseName         string // 数据库名称
	ModelName            string // 模型名称
}

func (l *ModelLocator) String() string {
	return strings.Join([]string{l.OrgName, l.ProjectSlug, l.DatabaseName, l.ModelName}, ".")
}

// Validate 验证模型定位器的必填字段
// 检查 OrgName、ProjectSlug、DatabaseName、ModelName 是否为空
func (l *ModelLocator) Validate() error {
	if err := l.ProjectScope.Validate(); err != nil {
		return err
	}
	if l.DatabaseName == "" {
		return bizerrors.Errorf("DatabaseName cant be blank")
	}
	if l.ModelName == "" {
		return bizerrors.Errorf("ModelName cant be blank")
	}
	return nil
}

// NewModelLocator 创建模型定位器并验证必填字段
// 参数: orgName - 组织名称, projectSlug - 项目标识符, databaseName - 数据库名称, modelName - 模型名称
// 返回: ModelLocator 值和验证错误
func NewModelLocator(orgName, projectSlug, databaseName, modelName string) (*ModelLocator, error) {
	locator := ModelLocator{
		ProjectScope: project.ProjectScope{
			OrgName:     orgName,
			ProjectSlug: projectSlug,
		},
		DatabaseName: databaseName,
		ModelName:    modelName,
	}
	if err := locator.Validate(); err != nil {
		return &locator, err
	}
	return &locator, nil
}

// GetFullPath 获取模型完整路径
// 返回格式: org_name.project_slug.database_name.model_name
func (l *ModelLocator) GetFullPath() string {
	return strings.Join([]string{l.OrgName, l.ProjectSlug, l.DatabaseName, l.ModelName}, ".")
}

// GetDatabasePath 获取数据库完整路径
// 返回格式: org_name.project_slug.database_name
func (l *ModelLocator) GetDatabasePath() string {
	return strings.Join([]string{l.OrgName, l.ProjectSlug, l.DatabaseName}, ".")
}

// ModelMeta 模型元数据
type ModelMeta struct {
	ID               string           `json:"id"`
	ModelLocator                      // 嵌入模型定位器
	Title            string           `json:"title"`
	Description      string           `json:"description"`
	StorageType      string           `json:"storageType"`
	DisplayField     *string          `json:"displayField"` // 用于 runtime _label 解析的字段名
	Version          int64            `json:"version"`
	Status           string           `json:"status"`
	GroupID          *string          `json:"groupId"`
	DeploymentStatus DeploymentStatus `json:"deploymentStatus"`
	LastSyncAt       *time.Time       `json:"lastSyncAt"`
	SyncError        string           `json:"syncError"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

// DataModel 模型定义实体
// 由模型元数据和字段定义两部分组成
type DataModel struct {
	ModelMeta
	Fields []*FieldDefinition `json:"fields"`
}

// GetField 根据字段名获取单个字段定义
// fieldName: 要查找的字段名称
// 返回: 匹配的字段定义，如果未找到则返回nil
func (m *DataModel) GetField(fieldName string) *FieldDefinition {
	for _, field := range m.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// GetModelLocator 获取模型定位器
// 返回包含模型名称、集群名称和数据库名称的定位器实例
func (m *DataModel) GetModelLocator() *ModelLocator {
	return &m.ModelLocator
}

// GetFieldsByNames 根据字段名列表获取多个字段定义
// fieldNames: 要查找的字段名称列表
// 返回: 匹配的字段定义列表，保持原有顺序
func (m *DataModel) GetFieldsByNames(fieldNames []string) []*FieldDefinition {
	if len(fieldNames) == 0 {
		return nil
	}

	// 使用 map 提高查找效率
	nameSet := make(map[string]bool, len(fieldNames))
	for _, name := range fieldNames {
		nameSet[name] = true
	}

	result := make([]*FieldDefinition, 0, len(fieldNames))
	for _, field := range m.Fields {
		if nameSet[field.Name] {
			result = append(result, field)
		}
	}

	return result
}

// Update 更新模型属性
// title: 新的标题，传nil表示不更新
// description: 新的描述，传nil表示不更新
func (m *DataModel) Update(title, description *string) error {
	now := time.Now()

	if title != nil {
		m.Title = *title
	}

	if description != nil {
		m.Description = *description
	}

	m.UpdatedAt = now

	return nil
}

// UpdateDisplayField 更新 displayField
// displayField: 新的显示字段名称，传 nil 表示清空
func (m *DataModel) UpdateDisplayField(displayField *string) {
	m.DisplayField = displayField
	m.UpdatedAt = time.Now()
}

// ValidateDisplayField 验证 displayField 是否有效
// 如果 displayField 为 nil 或空字符串，返回 nil（视为未设置）
// 如果 displayField 不为空，验证：
// 1. 该字段必须存在于模型的字段集合中
// 2. 该字段必须是可字符串化的类型（非 RELATION / ENUM_LABEL 等虚拟字段）
func (m *DataModel) ValidateDisplayField() error {
	if m.DisplayField == nil || *m.DisplayField == "" {
		return nil
	}

	fieldName := *m.DisplayField
	field := m.GetField(fieldName)
	if field == nil {
		return bizerrors.Errorf("displayField '%s' not found in model fields", fieldName)
	}

	// 检查字段类型是否可字符串化
	if !field.IsStringifiable() {
		return bizerrors.Errorf("displayField '%s' is not stringifiable (type: %s)", fieldName, field.Type.Format)
	}

	return nil
}

// ValidateMeta 验证模型元数据必填字段
func (m *ModelMeta) ValidateMeta() error {
	if err := m.ModelLocator.Validate(); err != nil {
		return err
	}
	if m.Title == "" {
		return bizerrors.Errorf("Title cant be blank")
	}
	if len(m.Title) > 255 {
		return bizerrors.Errorf("Title exceeds 255 characters")
	}
	if m.StorageType == "" {
		return bizerrors.Errorf("StorageType cant be blank")
	}
	return nil
}

// Validate 验证模型定义
// 检查模型的必填字段和字段定义的有效性
// 返回: 验证失败时返回具体的错误信息
func (m *DataModel) Validate() error {
	if err := m.ModelMeta.ValidateMeta(); err != nil {
		return err
	}
	if len(m.Fields) == 0 {
		return bizerrors.Errorf("Fields cant be blank")
	}
	if len(m.Fields) != 0 {
		if err := m.validateDuplicateFields(); err != nil {
			return err
		}
		for _, field := range m.Fields {
			err := field.Validate()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *DataModel) validateDuplicateFields() error {
	if len(m.Fields) == 0 {
		return nil
	}

	fieldNames := make(map[string]bool)
	for _, field := range m.Fields {
		if fieldNames[field.Name] {
			return bizerrors.Errorf("字段名称 '%s' 重复", field.Name)
		}
		fieldNames[field.Name] = true
	}
	return nil
}

// GetBizUniqueName 获取业务唯一名称
// 返回格式为: OrgName.ProjectSlug.DatabaseName.ModelName 的唯一标识符
func (m *DataModel) GetBizUniqueName() string {
	return strings.Join([]string{m.OrgName, m.ProjectSlug, m.DatabaseName, m.ModelName}, ".")
}

// AddFields 向模型添加字段定义
// fields: 要添加的字段定义列表
// 会自动设置字段的ModelID和ModelLocator属性
func (m *DataModel) AddFields(fields []*FieldDefinition) {
	for _, field := range fields {
		field.ModelID = m.ID
		field.ModelLocator = m.GetModelLocator()
	}
	m.Fields = append(m.Fields, fields...)
}
