package core

import "fmt"

// FieldDefinition 统一字段定义
type FieldDefinition struct {
	Key         string           `json:"key"`  // 必填字段
	Type        FieldType        `json:"type"` // 必填字段
	NonNull     bool             `json:"nonNull"`
	Required    bool             `json:"required"`
	Description string           `json:"description"`
	Default     any              `json:"default,omitempty"`
	Validation  *ValidationRules `json:"validation,omitempty"`
}

// NewFieldDefinition 创建字段定义
func NewFieldDefinition(key string, fieldType FieldType) (*FieldDefinition, error) {
	if key == "" {
		return nil, fmt.Errorf("key is required and cannot be empty")
	}
	if !fieldType.IsValid() {
		return nil, fmt.Errorf("invalid field type: %s", fieldType)
	}
	return &FieldDefinition{
		Key:         key,
		Type:        fieldType,
		NonNull:     true,
		Required:    false,
		Description: "",
		Default:     nil,
		Validation:  nil,
	}, nil
}

// ValidationRules 验证规则
type ValidationRules struct {
	MinLength *int     `json:"minLength,omitempty"`
	MaxLength *int     `json:"maxLength,omitempty"`
	Pattern   *string  `json:"pattern,omitempty"`
	Minimum   *float64 `json:"minimum,omitempty"`
	Maximum   *float64 `json:"maximum,omitempty"`
}

// GetJSONType 获取JSON Schema类型
func (f *FieldDefinition) GetJSONType() string {
	return f.Type.GetJSONType()
}

// Validate 验证字段定义的有效性
func (f *FieldDefinition) Validate() error {
	if f.Key == "" {
		return fmt.Errorf("field key is required and cannot be empty")
	}
	if !f.Type.IsValid() {
		return fmt.Errorf("invalid field type: %s", f.Type)
	}
	// 验证默认值与类型的兼容性
	if f.Default != nil {
		if err := f.validateDefaultValue(); err != nil {
			return err
		}
	}
	return nil
}

// validateDefaultValue 验证默认值与字段类型的兼容性
func (f *FieldDefinition) validateDefaultValue() error {
	if f.Default == nil {
		return nil
	}

	switch f.Type {
	case FieldTypeString, FieldTypeID:
		if _, ok := f.Default.(string); !ok {
			return fmt.Errorf("default value for %s field must be string, got %T", f.Type, f.Default)
		}
	case FieldTypeInteger:
		switch f.Default.(type) {
		case int, int32, int64:
			// 整数类型都可以接受
		default:
			return fmt.Errorf("default value for %s field must be integer, got %T", f.Type, f.Default)
		}
	case FieldTypeFloat:
		switch f.Default.(type) {
		case float32, float64, int, int32, int64:
			// 浮点数和整数都可以接受
		default:
			return fmt.Errorf("default value for %s field must be number, got %T", f.Type, f.Default)
		}
	case FieldTypeBoolean:
		if _, ok := f.Default.(bool); !ok {
			return fmt.Errorf("default value for %s field must be boolean, got %T", f.Type, f.Default)
		}
	}
	return nil
}

// WithNonNull 链式设置非空属性
func (f *FieldDefinition) WithNonNull(nonNull bool) *FieldDefinition {
	f.NonNull = nonNull
	return f
}

// WithRequired 链式设置必填属性
func (f *FieldDefinition) WithRequired(required bool) *FieldDefinition {
	f.Required = required
	return f
}

// WithDescription 链式设置描述
func (f *FieldDefinition) WithDescription(description string) *FieldDefinition {
	f.Description = description
	return f
}

// WithDefault 链式设置默认值
func (f *FieldDefinition) WithDefault(value any) *FieldDefinition {
	f.Default = value
	return f
}

// WithValidation 链式设置验证规则
func (f *FieldDefinition) WithValidation(rules *ValidationRules) *FieldDefinition {
	f.Validation = rules
	return f
}
