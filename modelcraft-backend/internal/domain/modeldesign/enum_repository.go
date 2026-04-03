package modeldesign

// EnumRepository 枚举定义仓储接口
type EnumRepository interface {
	// Create 创建枚举定义
	Create(enum *EnumDefinition) error

	// Update 更新枚举定义
	Update(enum *EnumDefinition) error

	// Delete 删除枚举定义（org + project scoped）
	Delete(orgName, projectSlug, name string) error

	// FindByName 根据组织、项目标识符和name查找枚举定义（org + project scoped）
	FindByName(orgName, projectSlug, name string) (*EnumDefinition, error)

	// FindByID 根据ID查找枚举定义
	FindByID(id string) (*EnumDefinition, error)

	// List 列出项目下的所有枚举定义（org + project scoped）
	List(orgName, projectSlug string) ([]*EnumDefinition, error)

	// IsReferencedByFields 检查枚举是否被字段引用（org + project scoped）
	// 返回是否被引用、引用该枚举的字段名列表、错误
	IsReferencedByFields(orgName, projectSlug, name string) (bool, []string, error)

	// ExistsByName 检查项目下指定name的枚举是否存在（org + project scoped）
	ExistsByName(orgName, projectSlug, name string) (bool, error)
}
