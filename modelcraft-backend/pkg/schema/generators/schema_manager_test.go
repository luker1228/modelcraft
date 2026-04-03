package generators

import (
	"modelcraft/pkg/schema/core"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// normalizeSDL 标准化SDL字符串，用于对比
func normalizeSDL(sdl string) string {
	// 1. 按行处理
	lines := strings.Split(sdl, "\n")
	var processedLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// 保留所有非空行，包括注释
			processedLines = append(processedLines, trimmed)
		}
	}

	// 2. 重新组合
	result := strings.Join(processedLines, "\n")

	// 3. 标准化空格，但保持结构
	result = regexp.MustCompile(`[ \t]+`).ReplaceAllString(result, " ")

	return strings.TrimSpace(result)
}

// assertSDLEqual SDL对比辅助函数
func assertSDLEqual(t *testing.T, expected, actual string) {
	normalizedExpected := normalizeSDL(expected)
	normalizedActual := normalizeSDL(actual)
	assert.Equal(t, normalizedExpected, normalizedActual, "SDL schemas should match")
}

func TestSchemaManager_GenerateJSONSchema(t *testing.T) {
	manager := NewSchemaManager()

	model := createTestModel()

	result, err := manager.GenerateJSONSchema(model)
	if err != nil {
		t.Fatalf("GenerateJSONSchema() error = %v", err)
	}

	if result == nil {
		t.Fatal("GenerateJSONSchema() returned nil result")
	}

	if result.Schema == nil {
		t.Error("JSON Schema is nil")
	}

	if result.Raw == "" {
		t.Error("Raw JSON is empty")
	}
}

func TestSchemaManager_GenerateGraphQLSchema(t *testing.T) {
	manager := NewSchemaManager()

	model := createTestModel()

	result, err := manager.GenerateGraphQLSchema(model)
	if err != nil {
		t.Fatalf("GenerateGraphQLSchema() error = %v", err)
	}

	if result == nil {
		t.Fatal("GenerateGraphQLSchema() returned nil result")
	}

	if result.Schema == nil {
		t.Error("GraphQL Schema is nil")
	}

	if result.SDL == "" {
		t.Error("SDL is empty")
	}

	// SDL 内容直观对比校验
	expectedSDL := strings.Join([]string{
		"type User {",
		"  id: ID! # User ID",
		"  name: String! # User name",
		"  age: Int # User age",
		"  email: String! # User email",
		"  active: Boolean # User active status",
		"}",
		"",
		"type UpdateUserResult {",
		"  success: Boolean!",
		"  updatedObj: User",
		"}",
		"",
		"type Query {",
		"  getUser(id: ID!): User",
		"  listUser(limit: Int, offset: Int): [User]",
		"}",
		"",
		"type Mutation {",
		"  createUser(id: ID!, name: String!, age: Int, email: String!, " +
			"active: Boolean): User",
		"  updateUser(id: ID!, returnUpdatedObj: Boolean, name: String, age: Int, " +
			"email: String, active: Boolean): UpdateUserResult",
		"  deleteUser(id: ID!): Boolean",
		"}",
	}, "\n")

	assertSDLEqual(t, expectedSDL, result.SDL)
}

func TestSchemaManager_GenerateBothSchemas(t *testing.T) {
	manager := NewSchemaManager()

	model := createTestModel()

	jsonResult, graphqlResult, err := manager.GenerateBothSchemas(model)
	if err != nil {
		t.Fatalf("GenerateBothSchemas() error = %v", err)
	}

	// 验证JSON Schema结果
	if jsonResult == nil {
		t.Error("JSON Schema result is nil")
	} else {
		if jsonResult.Schema == nil {
			t.Error("JSON Schema is nil")
		}
		if jsonResult.Raw == "" {
			t.Error("Raw JSON is empty")
		}

		// JSON Schema 内容验证
		expectedJSON := `{
			"$modeldesign": "http://json-modeldesign.org/draft-07/modeldesign#",
			"type": "object",
			"title": "User",
			"description": "User modeldesign for testing",
			"additionalProperties": false,
			"properties": {
				"id": {
					"type": "string",
					"description": "User ID"
				},
				"name": {
					"type": "string",
					"description": "User name",
					"default": "Anonymous"
				},
				"age": {
					"type": "integer",
					"description": "User age",
					"minimum": 0,
					"maximum": 150
				},
				"email": {
					"type": "string",
					"description": "User email",
					"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
				},
				"active": {
					"type": "boolean",
					"description": "User active status",
					"default": true
				}
			},
			"required": ["id", "name", "email"]
		}`

		assert.JSONEq(t, expectedJSON, jsonResult.Raw, "JSON schemas should match")
	}

	// 验证GraphQL Schema结果
	if graphqlResult == nil {
		t.Error("GraphQL Schema result is nil")
	} else {
		if graphqlResult.Schema == nil {
			t.Error("GraphQL Schema is nil")
		}
		if graphqlResult.SDL == "" {
			t.Error("SDL is empty")
		}

		// SDL 内容验证
		expectedSDL := strings.Join([]string{
			"type User {",
			"  id: ID! # User ID",
			"  name: String! # User name",
			"  age: Int # User age",
			"  email: String! # User email",
			"  active: Boolean # User active status",
			"}",
			"",
			"type UpdateUserResult {",
			"  success: Boolean!",
			"  updatedObj: User",
			"}",
			"",
			"type Query {",
			"  getUser(id: ID!): User",
			"  listUser(limit: Int, offset: Int): [User]",
			"}",
			"",
			"type Mutation {",
			"  createUser(id: ID!, name: String!, age: Int, email: String!, " +
				"active: Boolean): User",
			"  updateUser(id: ID!, returnUpdatedObj: Boolean, name: String, age: Int, " +
				"email: String, active: Boolean): UpdateUserResult",

			"  deleteUser(id: ID!): Boolean",
			"}",
		}, "\n")

		assertSDLEqual(t, expectedSDL, graphqlResult.SDL)
	}
}

func TestSchemaManager_ValidateModel(t *testing.T) {
	manager := NewSchemaManager()

	// 测试有效模型
	validModel := createTestModel()
	err := manager.ValidateModel(validModel)
	if err != nil {
		t.Errorf("ValidateModel() error = %v for valid modeldesign", err)
	}

	// 测试无效模型
	invalidModel := &core.ModelDefinition{
		Name:   "", // 空名称
		Fields: []*core.FieldDefinition{},
	}
	err = manager.ValidateModel(invalidModel)
	if err == nil {
		t.Error("ValidateModel() should return error for invalid modeldesign")
	}
}

func TestSchemaManager_InvalidModelHandling(t *testing.T) {
	manager := NewSchemaManager()

	invalidModel := &core.ModelDefinition{
		Name:   "",
		Fields: []*core.FieldDefinition{},
	}

	// 测试JSON Schema生成失败
	jsonResult, err := manager.GenerateJSONSchema(invalidModel)
	if err == nil {
		t.Error("GenerateJSONSchema() should return error for invalid modeldesign")
	}
	if jsonResult != nil {
		t.Error("GenerateJSONSchema() should return nil result for invalid modeldesign")
	}

	// 测试GraphQL Schema生成失败
	graphqlResult, err := manager.GenerateGraphQLSchema(invalidModel)
	if err == nil {
		t.Error("GenerateGraphQLSchema() should return error for invalid modeldesign")
	}
	if graphqlResult != nil {
		t.Error("GenerateGraphQLSchema() should return nil result for invalid modeldesign")
	}

	// 测试GenerateBothSchemas失败
	jsonResult, graphqlResult, err = manager.GenerateBothSchemas(invalidModel)
	if err == nil {
		t.Error("GenerateBothSchemas() should return error for invalid modeldesign")
	}
	if jsonResult != nil {
		t.Error("GenerateBothSchemas() should return nil JSON result for invalid modeldesign")
	}
	if graphqlResult != nil {
		t.Error("GenerateBothSchemas() should return nil GraphQL result for invalid modeldesign")
	}
}

// 创建测试模型的辅助函数
func createTestModel() *core.ModelDefinition {
	return &core.ModelDefinition{
		Name:        "User",
		Description: "User modeldesign for testing",
		Fields: []*core.FieldDefinition{
			{
				Key:         "id",
				Type:        core.FieldTypeID,
				Required:    true,
				Description: "User ID",
			},
			{
				Key:         "name",
				Type:        core.FieldTypeString,
				Required:    true,
				Description: "User name",
				Default:     "Anonymous",
			},
			{
				Key:         "age",
				Type:        core.FieldTypeInteger,
				Required:    false,
				Description: "User age",
				Validation: &core.ValidationRules{
					Minimum: floatPtr(0),
					Maximum: floatPtr(150),
				},
			},
			{
				Key:         "email",
				Type:        core.FieldTypeString,
				Required:    true,
				Description: "User email",
				Validation: &core.ValidationRules{
					Pattern: stringPtr("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"),
				},
			},
			{
				Key:         "active",
				Type:        core.FieldTypeBoolean,
				Required:    false,
				Description: "User active status",
				Default:     true,
			},
		},
	}
}

// 辅助函数 - 使用 json_schema_generator_test.go 中定义的函数
