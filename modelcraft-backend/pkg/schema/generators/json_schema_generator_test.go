package generators

import (
	"encoding/json"
	"modelcraft/pkg/schema/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCase JSON Schema测试案例
type TestCase struct {
	Name     string
	Model    *core.ModelDefinition
	Expected string
}

// assertJSONEqual 简单的JSON对比辅助函数
func assertJSONEqual(t *testing.T, expected, actual string) {
	assert.JSONEq(t, expected, actual, "JSON schemas should be equal")
}

// TestJSONSchemaGenerator_GenerateJSONSchema 测试JSON Schema生成器
func TestJSONSchemaGenerator_GenerateJSONSchema(t *testing.T) {
	generator := NewJSONSchemaGenerator()

	testCases := []TestCase{
		{
			Name: "基本字段测试",
			Model: &core.ModelDefinition{
				Name:        "User",
				Description: "User modeldesign for testing",
				Fields: []*core.FieldDefinition{
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
				},
			},
			Expected: `{
				"$modeldesign": "http://json-modeldesign.org/draft-07/modeldesign#",
				"type": "object",
				"title": "User",
				"description": "User modeldesign for testing",
				"additionalProperties": false,
				"properties": {
					"name": {
						"type": "string",
						"description": "User name",
						"default": "Anonymous"
					},
					"age": {
						"type": "integer",
						"description": "User age"
					},
					"email": {
						"type": "string",
						"description": "User email",
						"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
					}
				},
				"required": ["name", "email"]
			}`,
		},
		{
			Name: "所有字段类型测试",
			Model: &core.ModelDefinition{
				Name:        "Product",
				Description: "Product modeldesign with all field types",
				Fields: []*core.FieldDefinition{
					{
						Key:         "id",
						Type:        core.FieldTypeID,
						Required:    true,
						Description: "Product ID",
					},
					{
						Key:         "name",
						Type:        core.FieldTypeString,
						Required:    true,
						Description: "Product name",
					},
					{
						Key:         "price",
						Type:        core.FieldTypeFloat,
						Required:    true,
						Description: "Product price",
					},
					{
						Key:         "inStock",
						Type:        core.FieldTypeBoolean,
						Required:    false,
						Description: "In stock status",
						Default:     true,
					},
				},
			},
			Expected: `{
				"$modeldesign": "http://json-modeldesign.org/draft-07/modeldesign#",
				"type": "object",
				"title": "Product",
				"description": "Product modeldesign with all field types",
				"additionalProperties": false,
				"properties": {
					"id": {
						"type": "string",
						"description": "Product ID"
					},
					"name": {
						"type": "string",
						"description": "Product name"
					},
					"price": {
						"type": "number",
						"description": "Product price"
					},
					"inStock": {
						"type": "boolean",
						"description": "In stock status",
						"default": true
					}
				},
				"required": ["id", "name", "price"]
			}`,
		},
		{
			Name: "验证规则测试",
			Model: &core.ModelDefinition{
				Name:        "Account",
				Description: "Account modeldesign with validation rules",
				Fields: []*core.FieldDefinition{
					{
						Key:         "username",
						Type:        core.FieldTypeString,
						Required:    true,
						Description: "Username",
						Validation: &core.ValidationRules{
							MinLength: intPtr(3),
							MaxLength: intPtr(20),
							Pattern:   stringPtr("^[a-zA-Z0-9_]+$"),
						},
					},
					{
						Key:         "score",
						Type:        core.FieldTypeFloat,
						Required:    false,
						Description: "User score",
						Validation: &core.ValidationRules{
							Minimum: floatPtr(0.0),
							Maximum: floatPtr(100.0),
						},
					},
				},
			},
			Expected: `{
				"$modeldesign": "http://json-modeldesign.org/draft-07/modeldesign#",
				"type": "object",
				"title": "Account",
				"description": "Account modeldesign with validation rules",
				"additionalProperties": false,
				"properties": {
					"username": {
						"type": "string",
						"description": "Username",
						"minLength": 3,
						"maxLength": 20,
						"pattern": "^[a-zA-Z0-9_]+$"
					},
					"score": {
						"type": "number",
						"description": "User score",
						"minimum": 0.0,
						"maximum": 100.0
					}
				},
				"required": ["username"]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := generator.GenerateJSONSchema(tc.Model)
			if err != nil {
				t.Fatalf("GenerateJSONSchema() error = %v", err)
			}

			if result == nil {
				t.Fatal("GenerateJSONSchema() returned nil result")
			}

			// 使用JSON对比
			assertJSONEqual(t, tc.Expected, result.Raw)
		})
	}
}

func TestJSONSchemaGenerator_InvalidModel(t *testing.T) {
	generator := NewJSONSchemaGenerator()

	// 测试无效模型
	invalidModel := &core.ModelDefinition{
		Name:   "", // 空名称
		Fields: []*core.FieldDefinition{},
	}

	result, err := generator.GenerateJSONSchema(invalidModel)
	if err == nil {
		t.Error("GenerateJSONSchema() should return error for invalid modeldesign")
	}
	if result != nil {
		t.Error("GenerateJSONSchema() should return nil result for invalid modeldesign")
	}
}

func TestJSONSchemaBuilder_ValidationRules(t *testing.T) {
	builder := NewJSONSchemaBuilder()

	// 测试带验证规则的字段
	field := &core.FieldDefinition{
		Key:  "username",
		Type: core.FieldTypeString,
		Validation: &core.ValidationRules{
			MinLength: intPtr(3),
			MaxLength: intPtr(20),
			Pattern:   stringPtr("^[a-zA-Z0-9_]+$"),
		},
	}

	builder.SetBasicInfo("Test", "Test modeldesign")
	builder.AddField(field)
	schema := builder.Build()

	// 将schema转换为JSON字符串进行对比
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal modeldesign: %v", err)
	}

	expected := `{
		"$modeldesign": "http://json-modeldesign.org/draft-07/modeldesign#",
		"type": "object",
		"title": "Test",
		"description": "Test modeldesign",
		"additionalProperties": false,
		"properties": {
			"username": {
				"type": "string",
				"description": "",
				"minLength": 3,
				"maxLength": 20,
				"pattern": "^[a-zA-Z0-9_]+$"
			}
		}
	}`

	assertJSONEqual(t, expected, string(schemaBytes))
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
