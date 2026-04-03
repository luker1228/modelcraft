package modeldesign

import (
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// FieldValidator 字段验证器
type FieldValidator struct{}

// NewFieldValidator 创建字段验证器
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{}
}

// ValidateUniqueConstraint 验证唯一约束
// isUnique=true的string字段必须设置maxLength（UUIDV7除外，因为它有固定长度且天然有序）
func (v *FieldValidator) ValidateUniqueConstraint(field *FieldDefinition) error {
	if !field.IsUnique {
		return nil
	}

	// UUIDV7有固定长度且天然有序，不需要maxLength验证
	if field.Type.Format == FormatUUID {
		return nil
	}

	// 只对string类型的唯一字段进行maxLength检查
	if field.Type.SchemaType == SchemaTypeString {
		if field.Validation == nil || field.Validation.MaxLength == nil {
			return bizerrors.Errorf("isUnique字段必须设置maxLength")
		}
	}

	return nil
}

// storageHintStringTypes 字符串类系format允许的storageHint前缀
var storageHintStringTypes = []string{"VARCHAR", "CHAR", "TEXT", "MEDIUMTEXT", "LONGTEXT", "TINYTEXT"}

// storageHintNumericTypes 数值类系format允许的storageHint前缀
var storageHintNumericTypes = []string{
	"INT", "BIGINT", "TINYINT", "SMALLINT", "MEDIUMINT",
	"DOUBLE", "FLOAT", "DECIMAL", "NUMERIC",
}

// storageHintDateTimeTypes 日期时间类系format允许的storageHint前缀
var storageHintDateTimeTypes = []string{"DATE", "DATETIME", "TIMESTAMP", "TIME"}

// ValidateStorageHint 验证storageHint兼容性
func (v *FieldValidator) ValidateStorageHint(field *FieldDefinition) error {
	if field.StorageHint == nil || *field.StorageHint == "" {
		return nil
	}

	hint := *field.StorageHint
	format := field.Type.Format

	// ENUM/ENUM_ARRAY 存储格式固定，不允许自定义storageHint
	if format == FormatEnum || format == FormatEnumArray {
		return bizerrors.Errorf("format='%s' 不支持自定义storageHint", format)
	}

	// 验证storageHint与format的类型族兼容性
	if err := validateStorageHintFormatCompatibility(hint, format); err != nil {
		return err
	}

	// VARCHAR类型必须有maxLength且不超过255
	if hint == "VARCHAR" {
		if field.Validation == nil || field.Validation.MaxLength == nil {
			return bizerrors.Errorf("storageHint='VARCHAR'必须设置maxLength")
		}
		if *field.Validation.MaxLength > 255 {
			return bizerrors.Errorf("storageHint='VARCHAR'但maxLength>255")
		}
	}

	return nil
}

// validateStorageHintFormatCompatibility 验证storageHint与format类型族兼容性
func validateStorageHintFormatCompatibility(hint string, format FormatType) error {
	// 提取storageHint的基础类型（去掉括号参数，如 VARCHAR(64) -> VARCHAR）
	baseType := extractBaseStorageType(hint)

	switch {
	case isStringFormat(format):
		if !containsPrefix(storageHintStringTypes, baseType) {
			return bizerrors.Errorf(
				"format='%s' 与storageHint='%s'不兼容: 字符串类型只能使用 %v",
				format, hint, storageHintStringTypes,
			)
		}
	case isNumericFormat(format):
		if !containsPrefix(storageHintNumericTypes, baseType) {
			return bizerrors.Errorf(
				"format='%s' 与storageHint='%s'不兼容: 数值类型只能使用 %v",
				format, hint, storageHintNumericTypes,
			)
		}
	case isDateTimeFormat(format):
		if !containsPrefix(storageHintDateTimeTypes, baseType) {
			return bizerrors.Errorf(
				"format='%s' 与storageHint='%s'不兼容: 日期时间类型只能使用 %v",
				format, hint, storageHintDateTimeTypes,
			)
		}
	}

	return nil
}

// extractBaseStorageType 提取storageHint的基础类型名（去掉括号参数）
// 例如：VARCHAR(64) -> VARCHAR，DECIMAL(10,2) -> DECIMAL
func extractBaseStorageType(hint string) string {
	for i, ch := range hint {
		if ch == '(' {
			return hint[:i]
		}
	}
	return hint
}

// containsPrefix 检查target是否在允许列表中（大小写不敏感）
func containsPrefix(allowed []string, target string) bool {
	upper := toUpperASCII(target)
	for _, a := range allowed {
		if a == upper {
			return true
		}
	}
	return false
}

// toUpperASCII 将ASCII字母转为大写
func toUpperASCII(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		b[i] = c
	}
	return string(b)
}

// isStringFormat 判断format是否属于字符串类型族
func isStringFormat(format FormatType) bool {
	return format == FormatString || format == FormatUUID
}

// isNumericFormat 判断format是否属于数值类型族
func isNumericFormat(format FormatType) bool {
	return format == FormatInteger || format == FormatNumber || format == FormatDecimal || format == FormatBoolean
}

// isDateTimeFormat 判断format是否属于日期时间类型族
func isDateTimeFormat(format FormatType) bool {
	return format == FormatDate || format == FormatDateTime || format == FormatTime
}

// ValidateFormatCompatibility 验证Format与SchemaType一致性
func (v *FieldValidator) ValidateFormatCompatibility(field *FieldDefinition) error {
	// 这里可以添加更多的Format与SchemaType一致性检查
	// 例如：确保FormatInteger的SchemaType是NUMBER

	return nil
}

// ValidateAll 执行所有验证
func (v *FieldValidator) ValidateAll(field *FieldDefinition) error {
	if err := v.ValidateUniqueConstraint(field); err != nil {
		return err
	}

	if err := v.ValidateStorageHint(field); err != nil {
		return err
	}

	if err := v.ValidateFormatCompatibility(field); err != nil {
		return err
	}

	if err := v.ValidateDateTimeConfig(field); err != nil {
		return err
	}

	if err := v.ValidateEnumField(field); err != nil {
		return err
	}

	if err := v.ValidateIsArray(field); err != nil {
		return err
	}

	return nil
}

// ValidateIsArray 验证 isArray 字段规则
// 规则：
// - format=ENUM, isArray=false: valid
// - format=ENUM, isArray=true: valid，但若 field.Enum != nil 且 field.Enum.IsMultiSelect=false，则报错
// - format!=ENUM, isArray=true: invalid（ENUM_ARRAY 除外，它是 legacy 格式）
// - format!=ENUM, isArray=false: valid
func (v *FieldValidator) ValidateIsArray(field *FieldDefinition) error {
	if field.Type == nil {
		return nil
	}

	format := field.Type.Format

	// ENUM_ARRAY 是 legacy 格式，不受 isArray 标志影响
	if format == FormatEnumArray {
		return nil
	}

	// 非 ENUM 格式：isArray=true 不允许
	if format != FormatEnum {
		if field.IsArray {
			return bizerrors.Errorf("enum array not allowed for non-enum format")
		}
		return nil
	}

	// format=ENUM 的情况
	if !field.IsArray {
		return nil
	}

	// format=ENUM, isArray=true：检查关联的枚举是否支持多选
	if field.Enum != nil && !field.Enum.IsMultiSelect {
		return bizerrors.Errorf("enum does not support multi-select")
	}

	return nil
}

// ValidateEnumField 验证枚举字段配置
func (v *FieldValidator) ValidateEnumField(field *FieldDefinition) error {
	format := field.Type.Format

	// 只对ENUM和ENUM_ARRAY格式进行验证
	if format != FormatEnum && format != FormatEnumArray {
		return nil
	}

	// ENUM和ENUM_ARRAY字段的枚举关联由应用层(enumConfig)管理
	// 此处不再强制要求enumName字段
	// 字段可以使用 ValidationConfig.EnumValues (简单枚举) 或通过关联表管理

	return nil
}

// ValidateDateTimeConfig 验证日期/时间字段的配置
func (v *FieldValidator) ValidateDateTimeConfig(field *FieldDefinition) error {
	if field.Validation == nil {
		return nil
	}

	validation := field.Validation
	format := field.Type.Format

	switch format {
	case FormatDate, FormatDateTime:
		return validateDateFieldConfig(validation)
	case FormatTime:
		return validateTimeFieldConfig(validation)
	default:
		return nil
	}
}

func validateDateFieldConfig(validation *ValidationConfig) error {
	if validation.MinDate != nil {
		if _, err := parseDate(*validation.MinDate); err != nil {
			return bizerrors.Errorf("minDate格式无效,期望 YYYY-MM-DD,得到 '%s': %w", *validation.MinDate, err)
		}
	}
	if validation.MaxDate != nil {
		if _, err := parseDate(*validation.MaxDate); err != nil {
			return bizerrors.Errorf("maxDate格式无效,期望 YYYY-MM-DD,得到 '%s': %w", *validation.MaxDate, err)
		}
	}
	if validation.MinDate != nil && validation.MaxDate != nil {
		minDate, _ := parseDate(*validation.MinDate)
		maxDate, _ := parseDate(*validation.MaxDate)
		if minDate.After(maxDate) {
			return bizerrors.Errorf("minDate不能大于maxDate")
		}
	}

	return nil
}

func validateTimeFieldConfig(validation *ValidationConfig) error {
	if validation.MinTime != nil {
		if _, err := parseTime(*validation.MinTime); err != nil {
			return bizerrors.Errorf("minTime格式无效,期望 HH:MM:SS,得到 '%s': %w", *validation.MinTime, err)
		}
	}
	if validation.MaxTime != nil {
		if _, err := parseTime(*validation.MaxTime); err != nil {
			return bizerrors.Errorf("maxTime格式无效,期望 HH:MM:SS,得到 '%s': %w", *validation.MaxTime, err)
		}
	}
	// 注意:不验证 minTime <= maxTime,因为允许跨午夜的时间范围
	return nil
}

// ValidateDateValue 验证日期值格式 (ISO 8601 YYYY-MM-DD)
func ValidateDateValue(value string) error {
	_, err := parseDate(value)
	if err != nil {
		return bizerrors.Errorf("日期格式无效,期望 YYYY-MM-DD,得到 '%s'", value)
	}
	return nil
}

// ValidateTimeValue 验证时间值格式 (HH:MM:SS)
func ValidateTimeValue(value string) error {
	_, err := parseTime(value)
	if err != nil {
		return bizerrors.Errorf("时间格式无效,期望 HH:MM:SS,得到 '%s'", value)
	}
	return nil
}

// ValidateDateTimeValue 验证日期时间值格式 (ISO 8601 with timezone)
func ValidateDateTimeValue(value string) error {
	_, err := parseDateTime(value)
	if err != nil {
		return bizerrors.Errorf("日期时间格式无效,期望 ISO 8601格式(如 2024-01-15T14:30:00Z),得到 '%s'", value)
	}
	return nil
}

// ValidateDateRange 验证日期值是否在范围内
func ValidateDateRange(value string, minDate, maxDate *string) error {
	parsedValue, err := parseDate(value)
	if err != nil {
		return bizerrors.Errorf("日期格式无效: %w", err)
	}

	if minDate != nil {
		min, err := parseDate(*minDate)
		if err != nil {
			return bizerrors.Errorf("minDate格式无效: %w", err)
		}
		if parsedValue.Before(min) {
			return bizerrors.Errorf("日期必须在 %s 或之后", *minDate)
		}
	}

	if maxDate != nil {
		max, err := parseDate(*maxDate)
		if err != nil {
			return bizerrors.Errorf("maxDate格式无效: %w", err)
		}
		if parsedValue.After(max) {
			return bizerrors.Errorf("日期必须在 %s 或之前", *maxDate)
		}
	}

	return nil
}

// ValidateTimeRange 验证时间值是否在范围内(支持跨午夜范围)
func ValidateTimeRange(value string, minTime, maxTime *string) error {
	parsedValue, err := parseTime(value)
	if err != nil {
		return bizerrors.Errorf("时间格式无效: %w", err)
	}

	if minTime == nil && maxTime == nil {
		return nil
	}

	switch {
	case minTime != nil && maxTime != nil:
		min, _ := parseTime(*minTime)
		max, _ := parseTime(*maxTime)

		// 检查是否为跨午夜的时间范围 (例如 22:00:00 到 06:00:00)
		if min.After(max) {
			// 跨午夜:时间应该 >= min 或 <= max
			within := parsedValue.Equal(min) ||
				parsedValue.After(min) ||
				parsedValue.Equal(max) ||
				parsedValue.Before(max)
			if !within {
				return bizerrors.Errorf("时间必须在 %s 到 %s 之间(跨午夜)", *minTime, *maxTime)
			}
		} else {
			// 正常范围:时间应该 >= min 且 <= max
			outOfRange := parsedValue.Before(min) || parsedValue.After(max)
			atBoundary := parsedValue.Equal(min) || parsedValue.Equal(max)
			if outOfRange && !atBoundary {
				return bizerrors.Errorf("时间必须在 %s 到 %s 之间", *minTime, *maxTime)
			}
		}
	case minTime != nil:
		min, _ := parseTime(*minTime)
		if parsedValue.Before(min) && !parsedValue.Equal(min) {
			return bizerrors.Errorf("时间必须在 %s 或之后", *minTime)
		}
	case maxTime != nil:
		max, _ := parseTime(*maxTime)
		if parsedValue.After(max) && !parsedValue.Equal(max) {
			return bizerrors.Errorf("时间必须在 %s 或之前", *maxTime)
		}
	}

	return nil
}

// parseDate 解析日期字符串 (ISO 8601 YYYY-MM-DD)
func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

// parseTime 解析时间字符串 (HH:MM:SS)
func parseTime(value string) (time.Time, error) {
	return time.Parse("15:04:05", value)
}

// parseDateTime 解析日期时间字符串 (ISO 8601 with timezone)
func parseDateTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// ValidateEnumValue 验证单个枚举值是否在允许的选项中
func ValidateEnumValue(value string, enum *EnumDefinition) error {
	if enum == nil {
		return bizerrors.Errorf("枚举定义不存在")
	}

	// 检查值是否在枚举选项的code中
	for _, option := range enum.Options {
		if option.Code == value {
			return nil
		}
	}

	return bizerrors.Errorf("枚举值 '%s' 不在允许的选项中", value)
}

// ValidateEnumArrayValue 验证枚举数组值是否都在允许的选项中
func ValidateEnumArrayValue(values []string, enum *EnumDefinition) error {
	if enum == nil {
		return bizerrors.Errorf("枚举定义不存在")
	}

	if !enum.IsMultiSelect {
		return bizerrors.Errorf("该枚举不支持多选")
	}

	// 验证每个值都在枚举选项中
	validCodes := make(map[string]bool)
	for _, option := range enum.Options {
		validCodes[option.Code] = true
	}

	for _, value := range values {
		if !validCodes[value] {
			return bizerrors.Errorf("枚举值 '%s' 不在允许的选项中", value)
		}
	}

	return nil
}
