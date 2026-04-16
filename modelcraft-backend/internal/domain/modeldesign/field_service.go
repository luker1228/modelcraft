package modeldesign

import (
	"modelcraft/internal/domain/query"
	"modelcraft/pkg/bizerrors"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"
)

// FieldService 字段领域服务
type FieldService struct{}

func isValidUserFieldName(name string) bool {
	pattern := `^[a-zA-Z][a-zA-Z0-9_]*$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// NewFieldService 创建字段服务
func NewFieldService() *FieldService {
	return &FieldService{}
}

// ValidateDuplicates 验证字段定义列表
func (s *FieldService) ValidateDuplicates(fields []FieldDefinition) error {
	if len(fields) == 0 {
		return nil // 允许空字段列表
	}

	// 检查字段key重复
	keyMap := make(map[string]bool)
	for _, field := range fields {
		if field.Name == "" {
			return bizerrors.Errorf("field: Name cant be blank")
		}
		if !isValidUserFieldName(field.Name) {
			return bizerrors.Errorf(
				"字段名称 '%s' 格式无效，只允许字母、数字、下划线，且必须以字母开头（不允许 '_' 前缀）",
				field.Name,
			)
		}

		// 检查是否为保留关键字
		if query.IsReservedKeyword(field.Name) {
			return bizerrors.Errorf(
				"字段名称 '%s' 是保留关键字，不能用作字段名。保留关键字用于查询操作符，建议使用其他名称，例如 '%s_field' 或 '%s_value'",
				field.Name, field.Name, field.Name,
			)
		}

		if keyMap[field.Name] {
			return bizerrors.Errorf("字段Name '%s' 重复", field.Name)
		}
		keyMap[field.Name] = true
	}

	return nil
}

// GetNamesFromFields 从字段列表中获取字段名称
func (s *FieldService) GetNamesFromFields(fields []*FieldDefinition) []string {
	return lo.Map(fields, func(item *FieldDefinition, index int) string {
		return item.Name
	})
}

// ValidateAddFieldsNotExist 验证字段不重复
func (s *FieldService) ValidateAddFieldsNotExist(existFields []*FieldDefinition, toAddFieldNames ...string) error {
	fieldNames := make(map[string]bool)

	// 收集现有字段名称
	for _, field := range existFields {
		fieldNames[field.Name] = true
	}

	// 检查要添加的字段是否重复
	for _, name := range toAddFieldNames {
		if fieldNames[name] {
			return bizerrors.Errorf("字段名称 '%s' 已存在，不能重复添加", name)
		}
	}

	return nil
}

// GetSupportedFieldFormats 获取指定类型支持的格式
func (s *FieldService) GetSupportedFieldFormats() map[FormatType]*FieldType {
	return getAllSupportedFieldTypes()
}

// NewField 创建字段
func (s *FieldService) NewField(
	modelId, name, title string,
	format FormatType,
	locator *ModelLocator,
) (*FieldDefinition, error) {
	// 检查字段名是否为保留关键字
	if query.IsReservedKeyword(name) {
		return nil, bizerrors.Errorf(
			"字段名称 '%s' 是保留关键字，不能用作字段名。保留关键字用于查询操作符，建议使用其他名称，例如 '%s_field' 或 '%s_value'",
			name, name, name,
		)
	}
	if !isValidUserFieldName(name) {
		return nil, bizerrors.Errorf(
			"字段名称 '%s' 格式无效，只允许字母、数字、下划线，且必须以字母开头（不允许 '_' 前缀）",
			name,
		)
	}

	fieldType := GetFieldTypeByFormat(format)
	if fieldType == nil {
		return nil, bizerrors.Errorf("unknown format type: %s", format)
	}
	now := time.Now()
	return &FieldDefinition{
		ModelID:      modelId,
		ModelLocator: locator,
		Name:         name,
		Type:         fieldType,
		Title:        title,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// ValidContainSystemField 验证是否包含系统保留字段
func ValidContainSystemField(fields []FieldDefinition) error {
	if len(fields) == 0 {
		return nil
	}
	for _, field := range fields {
		if strings.EqualFold(field.Name, "id") {
			return bizerrors.NewError(bizerrors.ParamInvalid, "id is system field")
		}
	}
	return nil
}

// GetSystemFields 获取系统字段列表
func GetSystemFields() []*FieldDefinition {
	return []*FieldDefinition{
		{
			Name:        "id",
			Title:       "PrimaryKey",
			Description: "System Field",
			Type:        GetFieldTypeByFormat(FormatUUID),
			IsPrimary:   true,
			IsUnique:    true,
			NonNull:     true,
		},
	}
}
