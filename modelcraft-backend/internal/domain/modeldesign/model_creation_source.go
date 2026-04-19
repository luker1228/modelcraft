package modeldesign

// ModelCreationSource 模型创建来源
type ModelCreationSource string

const (
	ModelCreationSourceNew      ModelCreationSource = "NEW"      // 新建模型，自动生成 owner 字段
	ModelCreationSourceImported ModelCreationSource = "IMPORTED" // 导入模型，不生成 owner 字段
)
