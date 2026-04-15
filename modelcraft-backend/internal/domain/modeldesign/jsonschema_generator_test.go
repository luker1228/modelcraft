package modeldesign

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSchemaGenerator_GenerateSchema_BasicTypes(t *testing.T) {
	// Create a test model with basic field types
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "test-model-id",
			ModelLocator: ModelLocator{ModelName: "TestModel", DatabaseName: "test-db"},
			Title:        "Test Model",
			Description:  "A test model for JSON Schema generation",
			StorageType:  "mysql",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		Fields: []*FieldDefinition{
			{
				Name:         "id",
				Title:        "ID",
				Description:  "Primary key",
				Type:         GetFieldTypeByFormat(FormatUUID),
				Required:     true,
				NonNull:      true,
				IsPrimary:    true,
				DisplayOrder: "a0",
			},
			{
				Name:        "name",
				Title:       "Name",
				Description: "User name",
				Type:        GetFieldTypeByFormat(FormatString),
				Required:    true,
				NonNull:     true,
				Validation: &ValidationConfig{
					MinLength: intPtr(1),
					MaxLength: intPtr(100),
				},
				DisplayOrder: "a1",
			},
			{
				Name:         "age",
				Title:        "Age",
				Description:  "User age",
				Type:         GetFieldTypeByFormat(FormatInteger),
				NonNull:      true,
				DisplayOrder: "a2",
				Validation: &ValidationConfig{
					Minimum: float64Ptr(0),
					Maximum: float64Ptr(150),
				},
			},
			{
				Name:         "email",
				Title:        "Email",
				Description:  "Email address",
				Type:         GetFieldTypeByFormat(FormatString),
				DisplayOrder: "a3",
				Validation: &ValidationConfig{
					Pattern: stringPtr("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"),
					Rule:    EmailRule,
				},
			},
			{
				Name:         "active",
				Title:        "Active",
				Description:  "Is active user",
				Type:         GetFieldTypeByFormat(FormatBoolean),
				NonNull:      true,
				DisplayOrder: "a4",
			},
		},
	}

	// Generate JSON Schema
	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	// Parse JSON Schema
	var schema map[string]interface{}
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	require.NoError(t, err)

	// Verify basic structure
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema["$schema"])
	assert.Equal(t, "object", schema["type"])
	assert.Equal(t, "Test Model", schema["title"])
	assert.Equal(t, "A test model for JSON Schema generation", schema["description"])

	// Verify custom properties
	assert.Equal(t, "TestModel", schema["x-modelName"])
	assert.Equal(t, "test-db", schema["x-databaseName"])

	// Verify required fields
	required, ok := schema["required"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, len(required))
	assert.Contains(t, required, "id")
	assert.Contains(t, required, "name")

	// Verify properties
	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Check ID field
	idField, ok := properties["id"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", idField["type"])
	assert.Equal(t, "uuid", idField["format"])
	assert.Equal(t, true, idField["x-isPrimary"])
	assert.Equal(t, true, idField["readOnly"], "isPrimary field must have readOnly: true")

	// Check name field with validation
	nameField, ok := properties["name"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", nameField["type"])
	assert.Equal(t, float64(1), nameField["minLength"])
	assert.Equal(t, float64(100), nameField["maxLength"])

	// Check age field
	ageField, ok := properties["age"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "integer", ageField["type"])
	assert.Equal(t, float64(0), ageField["minimum"])
	assert.Equal(t, float64(150), ageField["maximum"])

	// Check email field
	emailField, ok := properties["email"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", emailField["type"])
	assert.NotNil(t, emailField["pattern"])
	assert.Equal(t, "email", emailField["x-validateRule"])
	assert.Equal(t, true, emailField["nullable"])

	// Check active field
	activeField, ok := properties["active"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "boolean", activeField["type"])
}

func TestJSONSchemaGenerator_GenerateSchema_DateTimeTypes(t *testing.T) {
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "test-model-id",
			ModelLocator: ModelLocator{ModelName: "EventModel", DatabaseName: "test-db"},
			Title:        "Event Model",
			Description:  "A model with date/time fields",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "eventDate",
				Title:        "Event Date",
				Type:         GetFieldTypeByFormat(FormatDate),
				DisplayOrder: "a0",
				Validation: &ValidationConfig{
					MinDate: stringPtr("2020-01-01"),
					MaxDate: stringPtr("2030-12-31"),
				},
			},
			{
				Name:         "eventTime",
				Title:        "Event Time",
				Type:         GetFieldTypeByFormat(FormatTime),
				DisplayOrder: "a1",
			},
			{
				Name:         "createdAt",
				Title:        "Created At",
				Type:         GetFieldTypeByFormat(FormatDateTime),
				DisplayOrder: "a2",
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	require.NoError(t, err)

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Check date field
	dateField, ok := properties["eventDate"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", dateField["type"])
	assert.Equal(t, "date", dateField["format"])
	assert.Equal(t, "2020-01-01", dateField["x-minDate"])
	assert.Equal(t, "2030-12-31", dateField["x-maxDate"])

	// Check time field
	timeField, ok := properties["eventTime"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", timeField["type"])
	assert.Equal(t, "time", timeField["format"])

	// Check datetime field
	datetimeField, ok := properties["createdAt"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", datetimeField["type"])
	assert.Equal(t, "date-time", datetimeField["format"])
}

func TestJSONSchemaGenerator_GenerateSchema_EnumFields(t *testing.T) {
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "test-model-id",
			ModelLocator: ModelLocator{ModelName: "ProductModel", DatabaseName: "test-db"},
			Title:        "Product Model",
			Description:  "A model with enum fields",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "status",
				Title:        "Status",
				Type:         GetFieldTypeByFormat(FormatEnum),
				DisplayOrder: "a0",
				Enum: &EnumDefinition{
					ID:            "status-enum-id",
					Name:          "ProductStatus",
					DisplayName:   "Product Status",
					Description:   "Status of the product",
					IsMultiSelect: false,
					Options: []EnumOption{
						{Code: "ACTIVE", Label: "Active", Description: "Product is active"},
						{Code: "INACTIVE", Label: "Inactive", Description: "Product is inactive"},
						{Code: "DISCONTINUED", Label: "Discontinued", Description: "Product is discontinued"},
					},
				},
			},
			{
				Name:         "tags",
				Title:        "Tags",
				Type:         GetFieldTypeByFormat(FormatEnumArray),
				DisplayOrder: "a1",
				Enum: &EnumDefinition{
					ID:            "tags-enum-id",
					Name:          "ProductTag",
					DisplayName:   "Product Tag",
					IsMultiSelect: true,
					Options: []EnumOption{
						{Code: "electronics", Label: "Electronics"},
						{Code: "clothing", Label: "Clothing"},
						{Code: "food", Label: "Food"},
					},
				},
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	require.NoError(t, err)

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Check enum field
	statusField, ok := properties["status"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", statusField["type"])

	enumValues, ok := statusField["enum"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(enumValues))
	assert.Contains(t, enumValues, "ACTIVE")
	assert.Contains(t, enumValues, "INACTIVE")
	assert.Contains(t, enumValues, "DISCONTINUED")

	// Check x-enum metadata
	xEnum, ok := statusField["x-enum"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ProductStatus", xEnum["name"])
	assert.Equal(t, false, xEnum["isMultiSelect"])

	options, ok := xEnum["options"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(options))

	// Check enum array field
	tagsField, ok := properties["tags"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "array", tagsField["type"])

	items, ok := tagsField["items"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", items["type"])

	itemEnums, ok := items["enum"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 3, len(itemEnums))
}

func TestJSONSchemaGenerator_GenerateSchema_DecimalField(t *testing.T) {
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "test-model-id",
			ModelLocator: ModelLocator{ModelName: "PriceModel", DatabaseName: "test-db"},
			Title:        "Price Model",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "price",
				Title:        "Price",
				Type:         GetFieldTypeByFormat(FormatDecimal),
				DisplayOrder: "a0",
				Validation: &ValidationConfig{
					Precision: intPtr(10),
					Scale:     intPtr(2),
					Minimum:   float64Ptr(0),
				},
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	err = json.Unmarshal([]byte(schemaJSON), &schema)
	require.NoError(t, err)

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	priceField, ok := properties["price"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "number", priceField["type"])
	assert.Equal(t, float64(10), priceField["x-precision"])
	assert.Equal(t, float64(2), priceField["x-scale"])
	assert.Equal(t, float64(0), priceField["minimum"])
}

// TestJSONSchemaGenerator_GenerateSchema_RequiredIsArrayWhenNoRequiredFields regression test.
//
// When no fields are marked as required, the generated JSON Schema must contain
// "required": [] (an empty JSON array), NOT "required": null.
//
// Root cause: Go's nil slice marshals to JSON null, whereas an initialised empty
// slice marshals to [].  RJSF calls required.includes() in JavaScript, which
// throws "Cannot read properties of null (reading 'includes')" when the value
// is null instead of an array.
func TestJSONSchemaGenerator_GenerateSchema_RequiredIsArrayWhenNoRequiredFields(t *testing.T) {
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "test-model-id",
			ModelLocator: ModelLocator{ModelName: "NoRequiredModel", DatabaseName: "test-db"},
			Title:        "No Required Model",
			Description:  "A model where no fields are marked as required",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "notes",
				Title:        "Notes",
				Type:         GetFieldTypeByFormat(FormatString),
				Required:     false, // explicitly not required
				DisplayOrder: "a0",
			},
			{
				Name:         "score",
				Title:        "Score",
				Type:         GetFieldTypeByFormat(FormatNumber),
				Required:     false,
				DisplayOrder: "a1",
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	// The raw JSON must contain "required":[] — not "required":null.
	// This is the exact contract that RJSF depends on.
	// Note: json.Marshal produces compact JSON without spaces.
	assert.Contains(t, schemaJSON, `"required":[]`,
		"required must be an empty JSON array, not null, so RJSF can call required.includes()")

	// Also verify via unmarshalling: the field must be a non-nil slice.
	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))

	rawRequired, exists := schema["required"]
	require.True(t, exists, "required key must be present in the schema")
	require.NotNil(t, rawRequired, "required must not be null")

	required, ok := rawRequired.([]interface{})
	require.True(t, ok, "required must deserialise as a JSON array, got %T", rawRequired)
	assert.Empty(t, required, "required array must be empty when no fields are required")
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

// TestJSONSchemaGenerator_ReadOnly verifies that isPrimary and RELATION fields
// are marked with "readOnly": true in the generated JSON Schema.
func TestJSONSchemaGenerator_ReadOnly(t *testing.T) {
	relateFKID := "fk_user"
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "test-model-id",
			ModelLocator: ModelLocator{ModelName: "OrderModel", DatabaseName: "test-db"},
			Title:        "Order Model",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "id",
				Title:        "ID",
				Type:         GetFieldTypeByFormat(FormatUUID),
				IsPrimary:    true,
				DisplayOrder: "a0",
			},
			{
				Name:         "user",
				Title:        "User",
				Type:         GetFieldTypeByFormat(FormatRelation),
				RelateFKID:   &relateFKID,
				DisplayOrder: "a1",
			},
			{
				Name:         "note",
				Title:        "Note",
				Type:         GetFieldTypeByFormat(FormatString),
				DisplayOrder: "a2",
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// isPrimary field → readOnly: true
	idField := properties["id"].(map[string]interface{})
	assert.Equal(t, true, idField["readOnly"], "primary key field must have readOnly: true")

	// RELATION field → readOnly: true
	userField := properties["user"].(map[string]interface{})
	assert.Equal(t, true, userField["readOnly"], "RELATION field must have readOnly: true")

	// regular string field → no readOnly
	noteField := properties["note"].(map[string]interface{})
	_, hasReadOnly := noteField["readOnly"]
	assert.False(t, hasReadOnly, "regular field must NOT have readOnly")
}

// TestJSONSchemaGenerator_XRelation verifies that a field with BelongsToFKID and
// Metadata["x-relation"] produces "x-relation" in the generated JSON Schema.
func TestJSONSchemaGenerator_XRelation(t *testing.T) {
	fkID := "fk-123"
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "order-model-id",
			ModelLocator: ModelLocator{ModelName: "Order", DatabaseName: "shop_db"},
			Title:        "Order",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:          "user_id",
				Title:         "User ID",
				Type:          GetFieldTypeByFormat(FormatString),
				BelongsToFKID: &fkID,
				DisplayOrder:  "a0",
				Metadata: map[string]any{
					"x-relation": map[string]string{
						"databaseName": "users_db",
						"modelName":    "User",
					},
				},
			},
			{
				Name:         "note",
				Title:        "Note",
				Type:         GetFieldTypeByFormat(FormatString),
				DisplayOrder: "a1",
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))

	properties, ok := schema["properties"].(map[string]interface{})
	require.True(t, ok)

	// user_id field should carry x-relation
	userIDField, ok := properties["user_id"].(map[string]interface{})
	require.True(t, ok)

	xRelation, ok := userIDField["x-relation"].(map[string]interface{})
	require.True(t, ok, "user_id field must have x-relation")
	assert.Equal(t, "users_db", xRelation["databaseName"])
	assert.Equal(t, "User", xRelation["modelName"])
	assert.Equal(t, fkID, userIDField["x-belongsToFkId"], "x-belongsToFkId must also be present")

	// note field must NOT have x-relation
	noteField, ok := properties["note"].(map[string]interface{})
	require.True(t, ok)
	_, hasXRelation := noteField["x-relation"]
	assert.False(t, hasXRelation, "regular field must NOT have x-relation")
}

// TestJSONSchemaGenerator_XRelation_MissingMetadata verifies that a field with
// BelongsToFKID but without Metadata["x-relation"] does NOT emit x-relation.
func TestJSONSchemaGenerator_XRelation_MissingMetadata(t *testing.T) {
	fkID := "fk-456"
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "order-model-id",
			ModelLocator: ModelLocator{ModelName: "Order", DatabaseName: "shop_db"},
			Title:        "Order",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:          "user_id",
				Title:         "User ID",
				Type:          GetFieldTypeByFormat(FormatString),
				BelongsToFKID: &fkID,
				DisplayOrder:  "a0",
				// Metadata is nil — x-relation should NOT appear
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))

	properties := schema["properties"].(map[string]interface{})
	userIDField := properties["user_id"].(map[string]interface{})

	_, hasXRelation := userIDField["x-relation"]
	assert.False(t, hasXRelation, "x-relation must not appear when Metadata is nil")
	assert.Equal(t, fkID, userIDField["x-belongsToFkId"], "x-belongsToFkId must still be present")
}
