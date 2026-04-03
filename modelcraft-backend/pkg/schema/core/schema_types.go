package core

import "github.com/graphql-go/graphql"

// JSONSchemaResult JSON Schema生成结果
type JSONSchemaResult struct {
	Schema map[string]interface{} `json:"modeldesign"`
	Raw    string                 `json:"raw"`
}

// GraphQLSchemaResult GraphQL Schema生成结果
type GraphQLSchemaResult struct {
	Schema *graphql.Schema `json:"-"`
	SDL    string          `json:"sdl"`
}

// SchemaGenerator 统一生成器接口
type SchemaGenerator interface {
	GenerateJSONSchema(model *ModelDefinition) (*JSONSchemaResult, error)
	GenerateGraphQLSchema(model *ModelDefinition) (*GraphQLSchemaResult, error)
}
