package core

// ModelDefinition 模型定义
type ModelDefinition struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Fields      []*FieldDefinition `json:"fields"`
}

// GetRequiredFields 获取必填字段列表
func (m *ModelDefinition) GetRequiredFields() []string {
	var required []string
	for _, field := range m.Fields {
		if field.Required {
			required = append(required, field.Key)
		}
	}
	return required
}

// GetFieldByKey 根据键获取字段定义
func (m *ModelDefinition) GetFieldByKey(key string) *FieldDefinition {
	for _, field := range m.Fields {
		if field.Key == key {
			return field
		}
	}
	return nil
}

// IsValid 验证模型定义是否有效
func (m *ModelDefinition) IsValid() bool {
	if m.Name == "" {
		return false
	}

	if len(m.Fields) == 0 {
		return false
	}

	// 检查字段键是否重复，并验证每个字段
	keyMap := make(map[string]bool)
	for _, field := range m.Fields {
		if err := field.Validate(); err != nil {
			return false
		}

		if keyMap[field.Key] {
			return false // 重复的键
		}
		keyMap[field.Key] = true
	}

	return true
}

// FieldCount 获取字段数量
func (m *ModelDefinition) FieldCount() int {
	return len(m.Fields)
}
