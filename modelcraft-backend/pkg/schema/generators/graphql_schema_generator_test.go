package generators

import (
	"modelcraft/pkg/schema/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GraphQLTestCase GraphQL测试案例
type GraphQLTestCase struct {
	Name        string
	Model       *core.ModelDefinition
	ExpectedSDL string
}

func TestGraphQLSchemaGenerator_GenerateGraphQLSchema(t *testing.T) {
	generator := NewGraphQLSchemaGenerator()

	testCases := []GraphQLTestCase{
		{
			Name: "基本字段类型测试",
			Model: &core.ModelDefinition{
				Name:        "Product",
				Description: "Product modeldesign for testing",
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
			ExpectedSDL: `
type Product {
  id: ID! # Product ID
  name: String! # Product name
  price: Float! # Product price
  inStock: Boolean # In stock status
}

type UpdateProductResult {
  success: Boolean!
  updatedObj: Product
}

type Query {
  getProduct(id: ID!): Product
  listProduct(limit: Int, offset: Int): [Product]
}

type Mutation {
  createProduct(id: ID!, name: String!, price: Float!, inStock: Boolean): Product
  updateProduct(id: ID!, returnUpdatedObj: Boolean, name: String, price: Float, inStock: Boolean): UpdateProductResult
  deleteProduct(id: ID!): Boolean
}
`,
		},
		{
			Name: "验证规则测试",
			Model: &core.ModelDefinition{
				Name:        "User",
				Description: "User modeldesign with validation",
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
						Key:         "age",
						Type:        core.FieldTypeInteger,
						Required:    false,
						Description: "User age",
						Validation: &core.ValidationRules{
							Minimum: floatPtr(0),
							Maximum: floatPtr(150),
						},
					},
				},
			},
			ExpectedSDL: `
type User {
  username: String! # Username
  age: Int # User age
}

type UpdateUserResult {
  success: Boolean!
  updatedObj: User
}

type Query {
  getUser(id: ID!): User
  listUser(limit: Int, offset: Int): [User]
}

type Mutation {
  createUser(username: String!, age: Int): User
  updateUser(id: ID!, returnUpdatedObj: Boolean, username: String, age: Int): UpdateUserResult
  deleteUser(id: ID!): Boolean
}
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := generator.GenerateGraphQLSchema(tc.Model)
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

			// SDL 内容对比
			assertSDLEqual(t, tc.ExpectedSDL, result.SDL)
		})
	}
}

func TestGraphQLSchemaGenerator_InvalidModel(t *testing.T) {
	generator := NewGraphQLSchemaGenerator()

	// 测试无效模型
	invalidModel := &core.ModelDefinition{
		Name:   "", // 空名称
		Fields: []*core.FieldDefinition{},
	}

	result, err := generator.GenerateGraphQLSchema(invalidModel)
	if err == nil {
		t.Error("GenerateGraphQLSchema() should return error for invalid modeldesign")
	}
	if result != nil {
		t.Error("GenerateGraphQLSchema() should return nil result for invalid modeldesign")
	}
}

func TestGraphQLSchemaGenerator_SDLStructure(t *testing.T) {
	generator := NewGraphQLSchemaGenerator()

	model := &core.ModelDefinition{
		Name:        "TestModel",
		Description: "Test modeldesign for SDL structure",
		Fields: []*core.FieldDefinition{
			{
				Key:         "id",
				Type:        core.FieldTypeID,
				Required:    true,
				Description: "Test ID",
			},
		},
	}

	result, err := generator.GenerateGraphQLSchema(model)
	if err != nil {
		t.Fatalf("GenerateGraphQLSchema() error = %v", err)
	}

	// 验证SDL包含必要的结构
	sdl := result.SDL

	// 验证类型定义
	assert.Contains(t, sdl, "type TestModel {", "SDL should contain type definition")
	assert.Contains(t, sdl, "id: ID!", "SDL should contain required ID field")

	// 验证Query类型
	assert.Contains(t, sdl, "type Query {", "SDL should contain Query type")
	assert.Contains(t, sdl, "getTestModel(id: ID!): TestModel", "SDL should contain get query")
	assert.Contains(t, sdl, "listTestModel(limit: Int, offset: Int): [TestModel]", "SDL should contain list query")

	// 验证Mutation类型
	assert.Contains(t, sdl, "type Mutation {", "SDL should contain Mutation type")
	assert.Contains(t, sdl, "createTestModel(", "SDL should contain create mutation")
	assert.Contains(t, sdl, "updateTestModel(", "SDL should contain update mutation")
	assert.Contains(t, sdl, "deleteTestModel(id: ID!): Boolean", "SDL should contain delete mutation")
}
