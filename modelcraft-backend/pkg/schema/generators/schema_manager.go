package generators

import (
	"fmt"
	"modelcraft/pkg/schema/core"
)

// SchemaManager 统一Schema管理器
type SchemaManager struct {
	jsonGenerator    *JSONSchemaGenerator
	graphqlGenerator *GraphQLSchemaGenerator
}

// NewSchemaManager 创建新的Schema管理器
func NewSchemaManager() *SchemaManager {
	return &SchemaManager{
		jsonGenerator:    NewJSONSchemaGenerator(),
		graphqlGenerator: NewGraphQLSchemaGenerator(),
	}
}

// GenerateJSONSchema 生成JSON Schema
func (m *SchemaManager) GenerateJSONSchema(model *core.ModelDefinition) (*core.JSONSchemaResult, error) {
	return m.jsonGenerator.GenerateJSONSchema(model)
}

// GenerateGraphQLSchema 生成GraphQL Schema
func (m *SchemaManager) GenerateGraphQLSchema(model *core.ModelDefinition) (*core.GraphQLSchemaResult, error) {
	return m.graphqlGenerator.GenerateGraphQLSchema(model)
}

// GenerateBothSchemas 生成两种Schema
func (m *SchemaManager) GenerateBothSchemas(
	model *core.ModelDefinition,
) (*core.JSONSchemaResult, *core.GraphQLSchemaResult, error) {
	jsonResult, err := m.GenerateJSONSchema(model)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate JSON modeldesign: %w", err)
	}

	graphqlResult, err := m.GenerateGraphQLSchema(model)
	if err != nil {
		return jsonResult, nil, fmt.Errorf("failed to generate GraphQL modeldesign: %w", err)
	}

	return jsonResult, graphqlResult, nil
}

// ValidateModel 验证模型定义
func (m *SchemaManager) ValidateModel(model *core.ModelDefinition) error {
	if !model.IsValid() {
		return fmt.Errorf("invalid modeldesign definition: %s", model.Name)
	}
	return nil
}
