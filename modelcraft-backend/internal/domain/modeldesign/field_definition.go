package modeldesign

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// StatusType 字段状态枚举
type StatusType string

const (
	FieldStatusInit          StatusType = "init"           // 创建时的初始状态
	FieldStatusDeploySuccess StatusType = "deploy_success" // deploy成功后的更新状态
	FieldStatusToDelete      StatusType = "to_delete"      // 待删除状态
)

// FieldDefinition 字段定义实体
type FieldDefinition struct {
	ModelID        string            `json:"modelId"`
	Name           string            `json:"name"`
	ModelLocator   *ModelLocator     `json:"modelLocator"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Type           *FieldType        `json:"type"`                  // 使用FieldFormat值对象指针（JSON序列化为format字符串）
	StorageHint    *string           `json:"storageHint,omitempty"` // 存储优化提示
	NonNull        bool              `json:"nonNull"`               // 字段是否非空，默认true表示不可为空
	Required       bool              `json:"required"`              // 字段是否必填，默认false表示非必填
	IsUnique       bool              `json:"isUnique"`
	IsPrimary      bool              `json:"isPrimary"`
	IsDeprecated   bool              `json:"isDeprecated"` // 是否已废弃
	IsArray        bool              `json:"isArray"`      // ENUM 是否为多选
	Status         StatusType        `json:"status"`
	Validation     *ValidationConfig `json:"validation"`
	DisplayOrder   string            `json:"displayOrder"`             // 字典序排序键（lexicographic fractional index）
	Enum           *EnumDefinition   `json:"enum,omitempty"`           // 关联的枚举详情（查询时加载）
	EnumName       string            `json:"enumName,omitempty"`       // format=ENUM 时使用
	EnumRelationID *string           `json:"enumRelationId,omitempty"` // format=ENUM_LABEL 时使用
	BelongsToFKID  *string           `json:"belongsToFkId,omitempty"`  // FK 列字段引用的逻辑外键 ID（model_id 侧）
	RelateFKID     *string           `json:"relateFkId,omitempty"`     // RELATION 格式字段引用的逻辑外键 ID
	Metadata       map[string]any    `json:"metadata"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// ValidationConfig 字段验证属性类型
type ValidationConfig struct {
	// 字符串类型特有属性
	MaxLength *int    `json:"maxLength,omitempty"` // 最大长度限制
	MinLength *int    `json:"minLength,omitempty"` // 最小长度限制
	Pattern   *string `json:"pattern,omitempty"`   // 正则表达式模式

	// 数字类型特有属性
	Maximum   *float64 `json:"maximum,omitempty"`   // 最大值
	Minimum   *float64 `json:"minimum,omitempty"`   // 最小值
	Precision *int     `json:"precision,omitempty"` // 精度（decimal类型）
	Scale     *int     `json:"scale,omitempty"`     // 小数位数（decimal类型）

	// 日期/时间类型特有属性
	MinDate *string `json:"minDate,omitempty"` // 最小日期（ISO 8601 YYYY-MM-DD）
	MaxDate *string `json:"maxDate,omitempty"` // 最大日期（ISO 8601 YYYY-MM-DD）
	MinTime *string `json:"minTime,omitempty"` // 最小时间（HH:MM:SS）
	MaxTime *string `json:"maxTime,omitempty"` // 最大时间（HH:MM:SS）

	// 数组类型特有属性
	MaxItems   *int         `json:"maxItems,omitempty"` // 数组最大元素数
	MinItems   *int         `json:"minItems,omitempty"` // 数组最小元素数
	EnumValues []string     `json:"enumValues,omitempty"`
	Rule       ValidateRule `json:"rule,omitempty"`
}

// Validate 验证字段定义
func (fd *FieldDefinition) Validate() error {
	// 验证必填字段
	if fd.Name == "" {
		return bizerrors.NewError(bizerrors.ParamInvalid, "field name is required")
	}
	// 验证字段key格式（只允许字母、数字、下划线，且不能以数字开头）
	if !isValidFieldName(fd.Name) {
		return bizerrors.NewError(
			bizerrors.ParamInvalid,
			fmt.Sprintf("field name '%s' 格式无效，只允许字母、数字、下划线，且不能以数字开头", fd.Name),
		)
	}

	err := fd.validate()
	if err != nil {
		var bizErr *bizerrors.BusinessError
		if bizerrors.As(err, &bizErr) {
			return bizErr
		}
		return bizerrors.NewError(bizerrors.ParamInvalid,
			fmt.Sprintf("field '%s' validation failed: %s", fd.Name, err.Error()))
	}
	return nil
}

func (fd *FieldDefinition) validate() error {
	if fd.ModelID == "" {
		return bizerrors.Errorf("ModelID不能为空")
	}
	if fd.ModelLocator == nil {
		return bizerrors.Errorf("ModelLocator不能为空")
	}
	if err := fd.ModelLocator.Validate(); err != nil {
		return bizerrors.Wrapf(err, "ModelLocator验证失败")
	}
	if fd.Title == "" {
		return bizerrors.Errorf("Title不能为空")
	}
	if fd.Type == nil {
		return bizerrors.Errorf("字段必须有Type")
	}
	if fd.Type.Format == "" {
		return bizerrors.Errorf("Type必须有format")
	}
	if fd.Type.SchemaType == "" {
		return bizerrors.Errorf("Type缺少SchemaType")
	}

	if err := fd.validateEnumField(); err != nil {
		return err
	}

	// belongs_to_fk_id 和 relate_fk_id 互斥
	if fd.BelongsToFKID != nil && fd.RelateFKID != nil {
		return bizerrors.Errorf("belongs_to_fk_id and relate_fk_id are mutually exclusive")
	}

	// RELATION 格式字段必须有 relate_fk_id
	if fd.Type.Format == FormatRelation && fd.RelateFKID == nil {
		return bizerrors.Errorf("RELATION format field must have relate_fk_id")
	}

	// 根据字段类型进行特定验证
	switch fd.Type.SchemaType {
	case SchemaTypeString:
		return fd.validateStringField()
	case SchemaTypeNumber:
		return fd.validateNumberField()
	case SchemaTypeArray:
		return fd.validateArrayField()
	case SchemaTypeBoolean:
		return fd.validateBooleanField()
	}
	return nil
}

// 私有验证方法
// validateStringField 验证字符串字段
func (fd *FieldDefinition) validateStringField() error {
	if fd.Validation == nil {
		return nil
	}

	validateProps := fd.Validation
	// 验证长度限制
	if validateProps.MinLength != nil && *validateProps.MinLength < 0 {
		return bizerrors.Errorf("最小长度不能为负数")
	}
	if validateProps.MaxLength != nil && *validateProps.MaxLength < 0 {
		return bizerrors.Errorf("最大长度不能为负数")
	}
	if validateProps.MinLength != nil &&
		validateProps.MaxLength != nil &&
		*validateProps.MinLength > *validateProps.MaxLength {
		return bizerrors.Errorf("最小长度不能大于最大长度")
	}

	// 验证正则表达式
	if validateProps.Pattern != nil {
		if _, err := regexp.Compile(*validateProps.Pattern); err != nil {
			return bizerrors.Errorf("正则表达式格式无效: %w", err)
		}
	}

	return nil
}

// validateNumberField 验证数字字段
func (fd *FieldDefinition) validateNumberField() error {
	if fd.Validation != nil {
		// 验证数值范围
		validateProps := fd.Validation
		if validateProps.Minimum != nil &&
			validateProps.Maximum != nil &&
			*validateProps.Minimum > *validateProps.Maximum {
			return bizerrors.Errorf("最小值不能大于最大值")
		}
	}

	return nil
}

// validateArrayField 验证数组字段
func (fd *FieldDefinition) validateArrayField() error {
	if fd.Validation != nil {
		// 验证数组长度限制
		validateProps := fd.Validation
		if validateProps.MinItems != nil && *validateProps.MinItems < 0 {
			return bizerrors.Errorf("最小元素数量不能为负数")
		}
		if validateProps.MaxItems != nil && *validateProps.MaxItems < 0 {
			return bizerrors.Errorf("最大元素数量不能为负数")
		}
		if validateProps.MinItems != nil &&
			validateProps.MaxItems != nil &&
			*validateProps.MinItems > *validateProps.MaxItems {
			return bizerrors.Errorf("最小元素数量不能大于最大元素数量")
		}
	}
	return nil
}

// validateBooleanField 验证布尔字段
func (fd *FieldDefinition) validateBooleanField() error {
	return nil
}

// validateEnumField 验证枚举字段与枚举标签字段不变量
func (fd *FieldDefinition) validateEnumField() error {
	if fd.Type == nil {
		return nil
	}

	enumRelationID := ""
	if fd.EnumRelationID != nil {
		enumRelationID = *fd.EnumRelationID
	}

	switch fd.Type.Format {
	case FormatEnum, FormatEnumArray:
		// 容错：ENUM 忽略 enumRelationId
		fd.EnumRelationID = nil
		if fd.EnumName == "" {
			return bizerrors.NewError(bizerrors.ParamInvalid, "relateEnumName is required when format=ENUM")
		}
	case FormatEnumLabel:
		// 容错：ENUM_LABEL 忽略 relateEnumName
		fd.EnumName = ""
		if enumRelationID == "" {
			return bizerrors.NewError(bizerrors.ParamInvalid, "enumRelationId is required when format=ENUM_LABEL")
		}
	default:
		// 容错：非 ENUM/ENUM_LABEL 忽略关联参数
		fd.EnumName = ""
		fd.EnumRelationID = nil
	}
	return nil
}

// IsRelationField 判断是否为关联关系字段
func (fd *FieldDefinition) IsRelationField() bool {
	if fd.Type == nil {
		return false
	}
	return fd.Type.Format == FormatRelation
}

// IsEnumField 判断是否为枚举字段
func (fd *FieldDefinition) IsEnumField() bool {
	if fd.Type == nil {
		return false
	}
	return fd.Type.Format == FormatEnum || fd.Type.Format == FormatEnumArray
}

// IsEnumLabelField 判断是否为枚举标签虚拟字段
func (fd *FieldDefinition) IsEnumLabelField() bool {
	if fd.Type == nil {
		return false
	}
	return fd.Type.Format == FormatEnumLabel
}

// IsEnumArrayField 判断是否为多选枚举字段（ENUM format with IsArray=true, or legacy ENUM_ARRAY)
func (fd *FieldDefinition) IsEnumArrayField() bool {
	if fd.Type == nil {
		return false
	}
	// Legacy: FormatEnumArray is always multi-select
	if fd.Type.Format == FormatEnumArray {
		return true
	}
	// New: FormatEnum with IsArray=true is multi-select
	return fd.Type.Format == FormatEnum && fd.IsArray
}

// IsStringifiable 判断字段值是否可以转为字符串用于 __label 显示
// 可字符串化的类型：STRING、UUID、INTEGER、NUMBER、DECIMAL、DATE、DATETIME、TIME、ENUM、BOOLEAN
// 不可字符串化的类型：RELATION（对象）、ENUM_LABEL（虚拟字段）、ENUM_ARRAY（数组）
func (fd *FieldDefinition) IsStringifiable() bool {
	if fd.Type == nil {
		return false
	}
	switch fd.Type.Format {
	case FormatString, FormatUUID, FormatInteger, FormatNumber, FormatDecimal,
		FormatDate, FormatDateTime, FormatTime, FormatEnum, FormatBoolean:
		return true
	default:
		return false
	}
}

// Update 更新字段信息
// 空字符串表示不更新该字段，nil 指针同理
func (fd *FieldDefinition) Update(title, description string, validation *ValidationConfig) {
	if title != "" {
		fd.Title = title
	}
	if description != "" {
		fd.Description = description
	}
	if validation != nil {
		fd.Validation = validation
	}
}

// Deprecate 将字段标记为废弃。若已废弃则幂等返回 nil。
func (fd *FieldDefinition) Deprecate() {
	fd.IsDeprecated = true
}

// Undeprecate 解除字段的废弃状态。若未废弃则幂等返回 nil。
func (fd *FieldDefinition) Undeprecate() {
	fd.IsDeprecated = false
}

// isValidFieldName 验证字段key格式
func isValidFieldName(key string) bool {
	// 只允许字母、数字、下划线，且不能以数字开头
	pattern := `^[a-zA-Z_][a-zA-Z0-9_]*$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

// SchemaType 字段基础类型枚举
type SchemaType string

const (
	SchemaTypeBoolean SchemaType = "boolean"
	SchemaTypeNumber  SchemaType = "number"
	SchemaTypeString  SchemaType = "string"
	SchemaTypeArray   SchemaType = "array"
	SchemaTypeObject  SchemaType = "object"
)

// ValidateRule 验证规则类型
type ValidateRule string

const (
	EmailRule = "email"
	URLRule   = "url"
	PhoneRule = "phone"
)

// FormatType 字段格式枚举
type FormatType string

const (
	// 基于string
	FormatString   FormatType = "STRING"
	FormatUUID     FormatType = "UUID"
	FormatDate     FormatType = "DATE"
	FormatDateTime FormatType = "DATETIME"
	FormatTime     FormatType = "TIME"

	// 基于number
	FormatNumber  FormatType = "NUMBER"
	FormatInteger FormatType = "INTEGER"
	FormatDecimal FormatType = "DECIMAL"

	// 基于boolean
	FormatBoolean FormatType = "BOOLEAN"

	// 特殊类型（保留用于关系字段）
	FormatRelation FormatType = "RELATION"

	// 枚举类型
	FormatEnum      FormatType = "ENUM"       // 单选枚举
	FormatEnumArray FormatType = "ENUM_ARRAY" // 多选枚举

	// 虚拟字段类型
	FormatEnumLabel FormatType = "ENUM_LABEL" // 枚举标签虚拟字段
)

// FieldType 字段类型结构（值对象）
// 包含字段的完整类型信息，避免重复查询映射表
type FieldType struct {
	SchemaType SchemaType `json:"-"` // 基础Schema类型，不序列化
	Format     FormatType `json:"-"` // 具体格式类型，不序列化
	Title      string     `json:"-"` // 显示名称，不序列化
}

// NewFieldFormat 创建字段格式
func NewFieldFormat(format FormatType) (*FieldType, error) {
	fieldFormat := GetFieldTypeByFormat(format)
	if fieldFormat == nil {
		return nil, bizerrors.Errorf("unsupported format type: %s", format)
	}

	return fieldFormat, nil
}

// GetType 获取Schema类型（保留向后兼容）
func (ft *FieldType) GetType() SchemaType {
	return ft.SchemaType
}

// GetFormat 获取格式类型（保留向后兼容）
func (ft *FieldType) GetFormat() FormatType {
	return ft.Format
}

// String 实现Stringer接口
func (ft *FieldType) String() string {
	return string(ft.Format)
}

// Equals 比较两个FieldFormat是否相等
func (ft FieldType) Equals(other FieldType) bool {
	return ft.Format == other.Format &&
		ft.SchemaType == other.SchemaType &&
		ft.Title == other.Title
}

// MarshalJSON 实现自定义JSON序列化，序列化为format字符串
func (ff FieldType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ff.Format))
}

// UnmarshalJSON 实现自定义JSON反序列化，从format字符串构造完整对象
func (ff *FieldType) UnmarshalJSON(data []byte) error {
	var formatStr string
	if err := json.Unmarshal(data, &formatStr); err != nil {
		return bizerrors.Errorf("invalid format value: %w", err)
	}

	format := FormatType(formatStr)
	fieldType := GetFieldTypeByFormat(format)
	if fieldType == nil {
		return bizerrors.Errorf(
			"unsupported format type: %s (valid types: STRING, UUID, DATE, DATETIME, TIME, "+
				"NUMBER, INTEGER, DECIMAL, BOOLEAN, RELATION, ENUM, ENUM_ARRAY, ENUM_LABEL)",
			formatStr,
		)
	}

	*ff = *fieldType
	return nil
}

// 全局字段类型映射表
var fieldTypeMap map[FormatType]*FieldType

// 初始化字段类型映射
func init() {
	fieldTypeMap = map[FormatType]*FieldType{
		FormatString:   {SchemaType: SchemaTypeString, Format: FormatString, Title: "字符串"},
		FormatUUID:     {SchemaType: SchemaTypeString, Format: FormatUUID, Title: "UUIDV7（天然有序）"},
		FormatDate:     {SchemaType: SchemaTypeString, Format: FormatDate, Title: "日期"},
		FormatDateTime: {SchemaType: SchemaTypeString, Format: FormatDateTime, Title: "日期时间"},
		FormatTime:     {SchemaType: SchemaTypeString, Format: FormatTime, Title: "时间"},

		FormatBoolean: {SchemaType: SchemaTypeBoolean, Format: FormatBoolean, Title: "布尔值"},

		// 数字格式
		FormatNumber:   {SchemaType: SchemaTypeNumber, Format: FormatNumber, Title: "数字"},
		FormatInteger:  {SchemaType: SchemaTypeNumber, Format: FormatInteger, Title: "整数"},
		FormatDecimal:  {SchemaType: SchemaTypeNumber, Format: FormatDecimal, Title: "精确小数"},
		FormatRelation: {SchemaType: SchemaTypeObject, Format: FormatRelation, Title: "关联"},

		// 枚举格式
		FormatEnum:      {SchemaType: SchemaTypeString, Format: FormatEnum, Title: "枚举(单选)"},
		FormatEnumArray: {SchemaType: SchemaTypeArray, Format: FormatEnumArray, Title: "枚举(多选)"},

		// 虚拟字段格式
		FormatEnumLabel: {SchemaType: SchemaTypeObject, Format: FormatEnumLabel, Title: "枚举标签(虚拟字段)"},
	}
}

// GetFieldTypeByFormat 根据format获取FieldFormat，如果获取不到返回nil
// 返回副本的指针，防止修改映射表
func GetFieldTypeByFormat(format FormatType) *FieldType {
	if fieldType, exists := fieldTypeMap[format]; exists {
		copyValue := *fieldType
		return &copyValue
	}
	return nil
}

// GetAllSupportedFieldTypes 获取所有支持的字段类型
// 返回映射表的副本，防止被修改
func getAllSupportedFieldTypes() map[FormatType]*FieldType {
	result := make(map[FormatType]*FieldType, len(fieldTypeMap))
	for format, fieldType := range fieldTypeMap {
		copyValue := *fieldType
		result[format] = &copyValue
	}
	return result
}
