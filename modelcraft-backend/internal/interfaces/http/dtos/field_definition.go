package dtos

import "modelcraft/internal/domain/modeldesign"

// RelationConfigDTO 关联关系配置DTO
type RelationConfigDTO struct {
	RelationType       string   `json:"relationType"`       // 关联类型
	TargetRelationName string   `json:"targetRelationName"` // 关联模型关联关系名
	ModelName          string   `json:"modelName"`          // 源模型
	TargetModelName    string   `json:"targetModelName"`    // 目标模型
	SourceFields       []string `json:"sourceFields"`       // 引用当前模型的字段，即外键
	TargetFields       []string `json:"targetFields"`       // 引用关联模型的字段
	ThroughTable       *string  `json:"throughTable"`       // 中间表（多对多使用）
}

// ValidationConfigDTO 校验属性配置DTO
type ValidationConfigDTO struct {
	// 字符串类型特有属性
	MaxLength  *int     `json:"maxLength,omitempty"` // 最大长度限制
	MinLength  *int     `json:"minLength,omitempty"` // 最小长度限制
	Pattern    *string  `json:"pattern,omitempty"`   // 正则表达式模式
	EnumValues []string `json:"enum,omitempty"`      // 枚举值列表

	// 数字类型特有属性
	Maximum *float64 `json:"maximum,omitempty"` // 最大值
	Minimum *float64 `json:"minimum,omitempty"` // 最小值

	// 数字类型特有属性
	MaxItems *int `json:"maxItems,omitempty"` // 最大值
	MinItems *int `json:"minItems,omitempty"` // 最小值
}

// EnumOptionDTO 枚举选项DTO
type EnumOptionDTO struct {
	Code        string  `json:"code"`                  // 枚举code
	Label       string  `json:"label"`                 // 枚举显示标签
	Order       int32   `json:"order"`                 // 显示顺序
	Description *string `json:"description,omitempty"` // 选项描述
}

// EnumConfigDTO 枚举配置DTO
type EnumConfigDTO struct {
	EnumName    string          `json:"enumName"`              // 枚举名称
	Options     []EnumOptionDTO `json:"options,omitempty"`     // 枚举选项（创建新枚举时使用）
	Description *string         `json:"description,omitempty"` // 枚举描述
	ConnectEnum bool            `json:"connectEnum"`           // true=关联现有枚举, false=创建新枚举
}

// FieldDefinitionDTO 字段定义DTO
type FieldDefinitionDTO struct {
	Name        string                 `json:"name" binding:"required"`  // 字段键名，不能重复
	Title       string                 `json:"title" binding:"required"` // 字段名称
	Description string                 `json:"description"`              // 字段含义/描述
	Format      modeldesign.FormatType `json:"format"`                   // 字段格式
	StorageHint *string                `json:"storageHint,omitempty"`    // 存储优化提示（如VARCHAR(64)）
	NonNull     bool                   `json:"nonNull"`                  // 字段是否非空，默认true表示不可为空
	Required    bool                   `json:"required"`                 // 字段是否必填，默认false表示非必填
	IsUnique    bool                   `json:"isUnique"`                 // 字段是否唯一
	IsPrimary   bool                   `json:"isPrimary"`
	IsArray     bool                   `json:"isArray"` // ENUM 是否为多选
	// 通用属性
	ValidationConfig *ValidationConfigDTO `json:"validationConfig,omitempty"` // 校验属性配置
	RelationConfig   *RelationConfigDTO   `json:"relationConfig,omitempty"`   // 关联关系配置（保留向后兼容）
	RelateEnumName   *string              `json:"relateEnumName,omitempty"`   // format=ENUM 时必填
	EnumRelationID   *string              `json:"enumRelationId,omitempty"`   // format=ENUM_LABEL 时必填
	RelateFKID       *string              `json:"relateFkId,omitempty"`       // RELATION 格式字段引用的逻辑外键 ID
}
