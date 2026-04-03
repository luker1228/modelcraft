package modeldesign

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchemaParser_ParseSchema_BasicTypes(t *testing.T) {
	parser := NewJSONSchemaParser(context.Background())

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"title": "Test Model",
		"description": "A test model",
		"x-modelName": "test_model",
		"x-clusterName": "test-cluster",
		"x-databaseName": "test-db",
		"properties": {
			"name": {
				"type": "string",
				"title": "Name",
				"description": "User name"
			},
			"age": {
				"type": "integer",
				"title": "Age"
			},
			"score": {
				"type": "number",
				"title": "Score"
			},
			"active": {
				"type": "boolean",
				"title": "Active"
			}
		},
		"required": ["name"]
	}`

	model, err := parser.ParseSchema(schema)
	require.NoError(t, err)
	require.NotNil(t, model)

	// 验证模型元数据
	assert.Equal(t, "Test Model", model.Title)
	assert.Equal(t, "A test model", model.Description)
	assert.Equal(t, "test_model", model.ModelName)
	assert.Equal(t, "test-db", model.DatabaseName)

	// 验证字段数量
	assert.Len(t, model.Fields, 4)

	// 验证字段类型
	nameField := model.GetField("name")
	require.NotNil(t, nameField)
	assert.Equal(t, FormatString, nameField.Type.Format)
	assert.True(t, nameField.Required)

	ageField := model.GetField("age")
	require.NotNil(t, ageField)
	assert.Equal(t, FormatInteger, ageField.Type.Format)
	assert.False(t, ageField.Required)

	scoreField := model.GetField("score")
	require.NotNil(t, scoreField)
	assert.Equal(t, FormatNumber, scoreField.Type.Format)

	activeField := model.GetField("active")
	require.NotNil(t, activeField)
	assert.Equal(t, FormatBoolean, activeField.Type.Format)
}

func TestJSONSchemaParser_ParseSchema_Formats(t *testing.T) {
	parser := NewJSONSchemaParser(context.Background())

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"title": "Test Model",
		"x-modelName": "test_model",
		"x-clusterName": "test-cluster",
		"x-databaseName": "test-db",
		"properties": {
			"id": {
				"type": "string",
				"format": "uuid",
				"title": "ID"
			},
			"birthDate": {
				"type": "string",
				"format": "date",
				"title": "Birth Date"
			},
			"createdAt": {
				"type": "string",
				"format": "date-time",
				"title": "Created At"
			},
			"startTime": {
				"type": "string",
				"format": "time",
				"title": "Start Time"
			}
		},
		"required": []
	}`

	model, err := parser.ParseSchema(schema)
	require.NoError(t, err)

	idField := model.GetField("id")
	assert.Equal(t, FormatUUID, idField.Type.Format)

	birthDateField := model.GetField("birthDate")
	assert.Equal(t, FormatDate, birthDateField.Type.Format)

	createdAtField := model.GetField("createdAt")
	assert.Equal(t, FormatDateTime, createdAtField.Type.Format)

	startTimeField := model.GetField("startTime")
	assert.Equal(t, FormatTime, startTimeField.Type.Format)
}

func TestJSONSchemaParser_ParseSchema_ValidationRules(t *testing.T) {
	parser := NewJSONSchemaParser(context.Background())

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"title": "Test Model",
		"x-modelName": "test_model",
		"x-clusterName": "test-cluster",
		"x-databaseName": "test-db",
		"properties": {
			"username": {
				"type": "string",
				"title": "Username",
				"minLength": 3,
				"maxLength": 20,
				"pattern": "^[a-zA-Z0-9]+$"
			},
			"price": {
				"type": "number",
				"title": "Price",
				"minimum": 0,
				"maximum": 1000
			},
			"amount": {
				"type": "number",
				"title": "Amount",
				"x-precision": 10,
				"x-scale": 2
			}
		},
		"required": []
	}`

	model, err := parser.ParseSchema(schema)
	require.NoError(t, err)

	usernameField := model.GetField("username")
	require.NotNil(t, usernameField.Validation)
	assert.Equal(t, 3, *usernameField.Validation.MinLength)
	assert.Equal(t, 20, *usernameField.Validation.MaxLength)
	assert.Equal(t, "^[a-zA-Z0-9]+$", *usernameField.Validation.Pattern)

	priceField := model.GetField("price")
	require.NotNil(t, priceField.Validation)
	assert.Equal(t, 0.0, *priceField.Validation.Minimum)
	assert.Equal(t, 1000.0, *priceField.Validation.Maximum)

	amountField := model.GetField("amount")
	assert.Equal(t, FormatDecimal, amountField.Type.Format)
	require.NotNil(t, amountField.Validation)
	assert.Equal(t, 10, *amountField.Validation.Precision)
	assert.Equal(t, 2, *amountField.Validation.Scale)
}

func TestJSONSchemaParser_ParseSchema_Nullable(t *testing.T) {
	parser := NewJSONSchemaParser(context.Background())

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"title": "Test Model",
		"x-modelName": "test_model",
		"x-clusterName": "test-cluster",
		"x-databaseName": "test-db",
		"properties": {
			"required_field": {
				"type": "string",
				"title": "Required Field"
			},
			"nullable_field": {
				"type": "string",
				"title": "Nullable Field",
				"nullable": true
			}
		},
		"required": []
	}`

	model, err := parser.ParseSchema(schema)
	require.NoError(t, err)

	requiredField := model.GetField("required_field")
	assert.True(t, requiredField.NonNull)

	nullableField := model.GetField("nullable_field")
	assert.False(t, nullableField.NonNull)
}

func TestJSONSchemaParser_ParseSchema_CustomProperties(t *testing.T) {
	parser := NewJSONSchemaParser(context.Background())

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"title": "Test Model",
		"x-modelName": "test_model",
		"x-clusterName": "test-cluster",
		"x-databaseName": "test-db",
		"properties": {
			"id": {
				"type": "string",
				"title": "ID",
				"x-isPrimary": true,
				"x-isUnique": true,
				"x-displayOrder": 1,
				"x-storageHint": "indexed"
			}
		},
		"required": []
	}`

	model, err := parser.ParseSchema(schema)
	require.NoError(t, err)

	idField := model.GetField("id")
	assert.True(t, idField.IsPrimary)
	assert.True(t, idField.IsUnique)
	assert.Equal(t, "1", idField.DisplayOrder)
	require.NotNil(t, idField.StorageHint)
	assert.Equal(t, "indexed", *idField.StorageHint)
}

func TestJSONSchemaParser_ParseSchema_EnumField(t *testing.T) {
	parser := NewJSONSchemaParser(context.Background())

	schema := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"title": "Test Model",
		"x-modelName": "test_model",
		"x-clusterName": "test-cluster",
		"x-databaseName": "test-db",
		"properties": {
			"status": {
				"type": "string",
				"title": "Status",
				"enum": ["active", "inactive", "pending"]
			},
			"roles": {
				"type": "array",
				"title": "Roles",
				"items": {
					"type": "string",
					"enum": ["admin", "user", "guest"]
				}
			}
		},
		"required": []
	}`

	model, err := parser.ParseSchema(schema)
	require.NoError(t, err)

	statusField := model.GetField("status")
	assert.Equal(t, FormatEnum, statusField.Type.Format)
	require.NotNil(t, statusField.Validation)
	assert.Equal(t, []string{"active", "inactive", "pending"}, statusField.Validation.EnumValues)

	rolesField := model.GetField("roles")
	assert.Equal(t, FormatEnumArray, rolesField.Type.Format)
	require.NotNil(t, rolesField.Validation)
	assert.Equal(t, []string{"admin", "user", "guest"}, rolesField.Validation.EnumValues)
}

func TestJSONSchemaParser_ParseSchema_ErrorCases(t *testing.T) {
	parser := NewJSONSchemaParser(context.Background())

	tests := []struct {
		name        string
		schema      string
		expectedErr string
	}{
		{
			name:        "invalid JSON",
			schema:      `{invalid json}`,
			expectedErr: "invalid JSON",
		},
		{
			name: "missing $schema",
			schema: `{
				"type": "object",
				"title": "Test"
			}`,
			expectedErr: "Missing or invalid '$schema'",
		},
		{
			name: "wrong schema version",
			schema: `{
				"$schema": "http://json-schema.org/draft-04/schema#",
				"type": "object",
				"title": "Test"
			}`,
			expectedErr: "Unsupported schema version",
		},

		{
			name: "missing properties",
			schema: `{
				"$schema": "http://json-schema.org/draft-07/schema#",
				"type": "object",
				"title": "Test",
				"x-modelName": "test",
				"x-clusterName": "cluster",
				"x-databaseName": "db"
			}`,
			expectedErr: "must have 'properties'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseSchema(tt.schema)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestJSONSchemaParser_RoundTrip(t *testing.T) {
	// 创建一个模型
	originalModel := &DataModel{
		ModelMeta: ModelMeta{
			ID: "test-id",
			ModelLocator: ModelLocator{
				ModelName:    "users",
				DatabaseName: "app-db",
			},
			Title:       "Users",
			Description: "User accounts",
			StorageType: "table",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "id",
				Title:        "ID",
				Description:  "User ID",
				Type:         GetFieldTypeByFormat(FormatUUID),
				NonNull:      true,
				Required:     true,
				IsPrimary:    true,
				DisplayOrder: "a0",
			},
			{
				Name:     "name",
				Title:    "Name",
				Type:     GetFieldTypeByFormat(FormatString),
				NonNull:  true,
				Required: true,
				Validation: &ValidationConfig{
					MinLength: intPtrTest(3),
					MaxLength: intPtrTest(50),
				},
				DisplayOrder: "a1",
			},
			{
				Name:     "age",
				Title:    "Age",
				Type:     GetFieldTypeByFormat(FormatInteger),
				NonNull:  false,
				Required: false,
				Validation: &ValidationConfig{
					Minimum: float64PtrTest(0),
					Maximum: float64PtrTest(150),
				},
				DisplayOrder: "a2",
			},
		},
	}

	// 生成JSON Schema
	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(originalModel)
	require.NoError(t, err)

	// 解析回DataModel
	parser := NewJSONSchemaParser(context.Background())
	parsedModel, err := parser.ParseSchema(schemaJSON)
	require.NoError(t, err)

	// 验证关键字段是否一致
	assert.Equal(t, originalModel.ModelName, parsedModel.ModelName)
	assert.Equal(t, originalModel.DatabaseName, parsedModel.DatabaseName)
	assert.Equal(t, originalModel.Title, parsedModel.Title)
	assert.Equal(t, originalModel.Description, parsedModel.Description)
	assert.Len(t, parsedModel.Fields, len(originalModel.Fields))

	// 验证字段类型和验证规则
	for _, originalField := range originalModel.Fields {
		parsedField := parsedModel.GetField(originalField.Name)
		require.NotNil(t, parsedField, "Field %s should exist", originalField.Name)
		assert.Equal(t, originalField.Type.Format, parsedField.Type.Format)
		assert.Equal(t, originalField.Required, parsedField.Required)
		assert.Equal(t, originalField.NonNull, parsedField.NonNull)
	}
}

// Helper functions
func intPtrTest(v int) *int {
	return &v
}

func float64PtrTest(v float64) *float64 {
	return &v
}
