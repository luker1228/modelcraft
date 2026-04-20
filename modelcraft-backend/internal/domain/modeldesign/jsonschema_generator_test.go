package modeldesign

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getXMC 从字段 map 中提取 x-mc 对象，供测试辅助使用。
// 如果不存在则返回空 map。
func getXMC(field map[string]interface{}) map[string]interface{} {
	raw, ok := field["x-mc"]
	if !ok || raw == nil {
		return map[string]interface{}{}
	}
	xmc, ok := raw.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	return xmc
}

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

	// Verify custom properties (model-level info stays at root)
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
	assert.Equal(t, true, getXMC(idField)["isPrimary"])
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
	assert.Equal(t, "email", getXMC(emailField)["validateRule"])
	assert.Equal(t, true, getXMC(emailField)["nullable"])

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
	assert.Equal(t, "2020-01-01", getXMC(dateField)["minDate"])
	assert.Equal(t, "2030-12-31", getXMC(dateField)["maxDate"])

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

	// x-mc.enum.labelFieldName must always be present for enum fields
	statusXMC := getXMC(statusField)
	enumRaw, hasXEnum := statusXMC["enum"]
	require.True(t, hasXEnum)
	enumMeta, ok := enumRaw.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "status_label", enumMeta["labelFieldName"])

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
	assert.Equal(t, float64(10), getXMC(priceField)["precision"])
	assert.Equal(t, float64(2), getXMC(priceField)["scale"])
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
// Metadata["x-relation"] produces "x-mc.relation" in the generated JSON Schema.
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

	// user_id field should carry x-mc.relation
	userIDField, ok := properties["user_id"].(map[string]interface{})
	require.True(t, ok)

	xmc := getXMC(userIDField)
	xRelationRaw, ok := xmc["relation"]
	require.True(t, ok, "user_id x-mc must have relation")
	xRelation, ok := xRelationRaw.(map[string]interface{})
	require.True(t, ok, "user_id x-mc must have relation as map")
	assert.Equal(t, "users_db", xRelation["databaseName"])
	assert.Equal(t, "User", xRelation["modelName"])
	assert.Equal(t, fkID, xRelation["belongsToFkId"])
	assert.Equal(t, "MANY_TO_ONE", xRelation["relationType"])
	assert.Equal(t, "normal", xRelation["relationDirection"])

	// note field must NOT have x-mc.relation
	noteField, ok := properties["note"].(map[string]interface{})
	require.True(t, ok)
	noteXMC := getXMC(noteField)
	_, hasRelation := noteXMC["relation"]
	assert.False(t, hasRelation, "regular field must NOT have x-mc.relation")
}

// TestJSONSchemaGenerator_XRelation_MissingMetadata verifies that a field with
// BelongsToFKID but without Metadata["x-relation"] still emits relation ids/type metadata.
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
				// Metadata is nil — relation should still contain fk/type/direction metadata.
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

	xmc := getXMC(userIDField)
	xRelationRaw, hasRelation := xmc["relation"]
	require.True(t, hasRelation, "x-mc.relation must be present")
	xRelation, ok := xRelationRaw.(map[string]interface{})
	require.True(t, ok, "x-mc.relation must be an object")
	assert.Equal(t, fkID, xRelation["belongsToFkId"])
	assert.Equal(t, "MANY_TO_ONE", xRelation["relationType"])
	assert.Equal(t, "normal", xRelation["relationDirection"])
}

// TestJSONSchemaGenerator_XRelation_WithRelateFKID verifies that a RELATION field
// with RelateFKID and Metadata["x-relation"] also emits x-mc.relation.
func TestJSONSchemaGenerator_XRelation_WithRelateFKID(t *testing.T) {
	relateFKID := "fk-rel-1"
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "user-model-id",
			ModelLocator: ModelLocator{ModelName: "User", DatabaseName: "app_db"},
			Title:        "User",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "orders",
				Title:        "Orders",
				Type:         GetFieldTypeByFormat(FormatRelation),
				RelateFKID:   &relateFKID,
				DisplayOrder: "a0",
				Metadata: map[string]any{
					"x-relation": map[string]string{
						"databaseName": "app_db",
						"modelName":    "Order",
						"direction":    "reverse",
						"cardinality":  "one-to-many",
					},
				},
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))

	properties := schema["properties"].(map[string]interface{})
	ordersField := properties["orders"].(map[string]interface{})
	xmc := getXMC(ordersField)
	xRelationRaw, hasRelation := xmc["relation"]
	require.True(t, hasRelation, "RELATION field with RelateFKID should emit x-mc.relation")
	xRelation, ok := xRelationRaw.(map[string]interface{})
	require.True(t, ok, "x-mc.relation must be an object")
	assert.Equal(t, relateFKID, xRelation["relateFkId"])
	assert.Equal(t, "ONE_TO_MANY", xRelation["relationType"])
	assert.Equal(t, "reverse", xRelation["relationDirection"])
}

func TestJSONSchemaGenerator_RelationTypeByCardinality(t *testing.T) {
	relateFKID := "fk-rel-2"
	model := &DataModel{
		ModelMeta: ModelMeta{
			ID:           "user-model-id",
			ModelLocator: ModelLocator{ModelName: "User", DatabaseName: "app_db"},
			Title:        "User",
			StorageType:  "mysql",
		},
		Fields: []*FieldDefinition{
			{
				Name:         "orders",
				Title:        "Orders",
				Type:         GetFieldTypeByFormat(FormatRelation),
				RelateFKID:   &relateFKID,
				DisplayOrder: "a0",
				Metadata: map[string]any{
					"x-relation": map[string]string{
						"databaseName": "app_db",
						"modelName":    "Order",
						"direction":    "reverse",
						"cardinality":  "one-to-many",
					},
				},
			},
			{
				Name:         "owner",
				Title:        "Owner",
				Type:         GetFieldTypeByFormat(FormatRelation),
				RelateFKID:   &relateFKID,
				DisplayOrder: "a1",
				Metadata: map[string]any{
					"x-relation": map[string]string{
						"databaseName": "app_db",
						"modelName":    "User",
						"direction":    "normal",
						"cardinality":  "many-to-one",
					},
				},
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))

	properties := schema["properties"].(map[string]interface{})

	ordersField := properties["orders"].(map[string]interface{})
	assert.Equal(t, "array", ordersField["type"], "one-to-many relation should be rendered as array")
	items, hasItems := ordersField["items"].(map[string]interface{})
	require.True(t, hasItems, "one-to-many relation should include items schema")
	assert.Equal(t, "object", items["type"])
	assert.Equal(t, "relation-multi-readonly", getXMC(ordersField)["widget"],
		"one-to-many relation should use readonly multi relation widget")

	ownerField := properties["owner"].(map[string]interface{})
	assert.Equal(t, "object", ownerField["type"], "many-to-one relation should remain object")
}

// ─── 新增 x-mc 测试 ────────────────────────────────────────────────────────────

// TestXMC_Widget_EnumSelect 验证 FormatEnum 和 FormatEnumArray 都得到 widget="enum-select"
func TestXMC_Widget_EnumSelect(t *testing.T) {
	enumDef := &EnumDefinition{
		Name:    "Status",
		Options: []EnumOption{{Code: "A", Label: "A"}},
	}
	for _, format := range []FormatType{FormatEnum, FormatEnumArray} {
		field := &FieldDefinition{
			Name:         "f",
			Title:        "F",
			Type:         GetFieldTypeByFormat(format),
			DisplayOrder: "a0",
			Enum:         enumDef,
		}
		generator := NewJSONSchemaGenerator()
		model := makeMinimalModel([]*FieldDefinition{field})
		schemaJSON, err := generator.GenerateSchema(model)
		require.NoError(t, err, "format=%s", format)

		var schema map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))
		props := schema["properties"].(map[string]interface{})
		fField := props["f"].(map[string]interface{})
		assert.Equal(t, "enum-select", getXMC(fField)["widget"], "format=%s should yield widget=enum-select", format)
	}
}

// TestXMC_Widget_Date 验证 FormatDate → widget="date"
func TestXMC_Widget_Date(t *testing.T) {
	field := &FieldDefinition{
		Name: "d", Title: "D",
		Type:         GetFieldTypeByFormat(FormatDate),
		DisplayOrder: "a0",
	}
	xmc := generateAndGetFieldXMC(t, field)
	assert.Equal(t, "date", xmc["widget"])
}

// TestXMC_Widget_DatetimeLocal 验证 FormatDateTime → widget="datetime-local"
func TestXMC_Widget_DatetimeLocal(t *testing.T) {
	field := &FieldDefinition{
		Name: "dt", Title: "DT",
		Type:         GetFieldTypeByFormat(FormatDateTime),
		DisplayOrder: "a0",
	}
	xmc := generateAndGetFieldXMC(t, field)
	assert.Equal(t, "datetime-local", xmc["widget"])
}

// TestXMC_Widget_Time 验证 FormatTime → widget="time"
func TestXMC_Widget_Time(t *testing.T) {
	field := &FieldDefinition{
		Name: "t", Title: "T",
		Type:         GetFieldTypeByFormat(FormatTime),
		DisplayOrder: "a0",
	}
	xmc := generateAndGetFieldXMC(t, field)
	assert.Equal(t, "time", xmc["widget"])
}

// TestXMC_Widget_Textarea 验证 storageHint=TEXT → widget="textarea"
func TestXMC_Widget_Textarea(t *testing.T) {
	hint := "TEXT"
	field := &FieldDefinition{
		Name: "content", Title: "Content",
		Type:         GetFieldTypeByFormat(FormatString),
		StorageHint:  &hint,
		DisplayOrder: "a0",
	}
	xmc := generateAndGetFieldXMC(t, field)
	assert.Equal(t, "textarea", xmc["widget"])
}

// TestXMC_Widget_RelationSelector 验证 BelongsToFKID != nil → widget="relation-selector"
func TestXMC_Widget_RelationSelector(t *testing.T) {
	fkID := "fk-1"
	field := &FieldDefinition{
		Name: "user_id", Title: "User ID",
		Type:          GetFieldTypeByFormat(FormatString),
		BelongsToFKID: &fkID,
		DisplayOrder:  "a0",
	}
	xmc := generateAndGetFieldXMC(t, field)
	assert.Equal(t, "relation-selector", xmc["widget"])
	relRaw, ok := xmc["relation"]
	require.True(t, ok, "x-mc.relation must be present")
	rel, ok := relRaw.(map[string]interface{})
	require.True(t, ok, "x-mc.relation must be an object")
	assert.Equal(t, "MANY_TO_ONE", rel["relationType"])
	assert.Equal(t, "normal", rel["relationDirection"])
}

// TestXMC_Widget_None 验证普通 STRING 字段不写 widget 键
func TestXMC_Widget_None(t *testing.T) {
	field := &FieldDefinition{
		Name: "name", Title: "Name",
		Type:         GetFieldTypeByFormat(FormatString),
		DisplayOrder: "a0",
	}
	xmc := generateAndGetFieldXMC(t, field)
	_, hasWidget := xmc["widget"]
	assert.False(t, hasWidget, "plain STRING field must not have widget key in x-mc")
}

// TestXMC_BaseFields 验证 isPrimary/isUnique/displayOrder/nullable 始终出现在 x-mc
func TestXMC_BaseFields(t *testing.T) {
	field := &FieldDefinition{
		Name: "f", Title: "F",
		Type:         GetFieldTypeByFormat(FormatString),
		IsPrimary:    true,
		IsUnique:     true,
		NonNull:      false, // nullable = true
		DisplayOrder: "b1",
	}
	xmc := generateAndGetFieldXMC(t, field)
	assert.Equal(t, true, xmc["isPrimary"])
	assert.Equal(t, true, xmc["isUnique"])
	assert.Equal(t, "b1", xmc["displayOrder"])
	assert.Equal(t, "STRING", xmc["format"])
	assert.Equal(t, true, xmc["nullable"])
}

// TestXMC_Enum_Metadata_DefaultLabelField 验证无 metadata.enumDisplay 时使用默认 labelFieldName
func TestXMC_Enum_Metadata_DefaultLabelField(t *testing.T) {
	field := &FieldDefinition{
		Name: "status", Title: "Status",
		Type:         GetFieldTypeByFormat(FormatEnum),
		DisplayOrder: "a0",
		Enum: &EnumDefinition{
			Name:          "Status",
			DisplayName:   "Status",
			IsMultiSelect: false,
			Options: []EnumOption{
				{Code: "A", Label: "Alpha", Description: "first"},
				{Code: "B", Label: "Beta", Description: "second"},
			},
		},
	}
	xmc := generateAndGetFieldXMC(t, field)
	enumRaw, ok := xmc["enum"]
	require.True(t, ok, "x-mc.enum must be present for ENUM fields")
	enumMeta, ok := enumRaw.(map[string]interface{})
	require.True(t, ok, "x-mc.enum must be an object")
	assert.Equal(t, "status_label", enumMeta["labelFieldName"])
}

func TestXMC_EnumLabelFieldName_FromMetadata(t *testing.T) {
	field := &FieldDefinition{
		Name: "status", Title: "Status",
		Type:         GetFieldTypeByFormat(FormatEnum),
		DisplayOrder: "a0",
		Metadata: map[string]any{
			"enumDisplay": map[string]any{
				"labelFieldName": "enum_display_label",
			},
		},
		Enum: &EnumDefinition{
			Name:          "Status",
			DisplayName:   "Status",
			IsMultiSelect: false,
			Options: []EnumOption{
				{Code: "A", Label: "Alpha"},
			},
		},
	}

	xmc := generateAndGetFieldXMC(t, field)
	enumRaw, ok := xmc["enum"]
	require.True(t, ok, "x-mc.enum must be present")
	enumMeta, ok := enumRaw.(map[string]interface{})
	require.True(t, ok, "x-mc.enum must be an object")
	assert.Equal(t, "enum_display_label", enumMeta["labelFieldName"])
}

// TestXMC_Relation_Metadata 验证 x-mc.relation.databaseName 和 modelName 正确
func TestXMC_Relation_Metadata(t *testing.T) {
	fkID := "fk-99"
	field := &FieldDefinition{
		Name: "org_id", Title: "Org ID",
		Type:          GetFieldTypeByFormat(FormatString),
		BelongsToFKID: &fkID,
		DisplayOrder:  "a0",
		Metadata: map[string]any{
			"x-relation": map[string]string{
				"databaseName": "org_db",
				"modelName":    "Org",
			},
		},
	}
	xmc := generateAndGetFieldXMC(t, field)
	relRaw, ok := xmc["relation"]
	require.True(t, ok, "x-mc.relation must be present")
	rel, ok := relRaw.(map[string]interface{})
	require.True(t, ok, "x-mc.relation must be a map")
	assert.Equal(t, "org_db", rel["databaseName"])
	assert.Equal(t, "Org", rel["modelName"])
}

// TestXMC_Validation_Fields 验证 minDate/maxDate/precision/scale/validateRule 出现在 x-mc
func TestXMC_Validation_Fields(t *testing.T) {
	// Use date field for minDate/maxDate
	dateField := &FieldDefinition{
		Name: "event_date", Title: "Date",
		Type:         GetFieldTypeByFormat(FormatDate),
		DisplayOrder: "a0",
		Validation: &ValidationConfig{
			MinDate: stringPtr("2020-01-01"),
			MaxDate: stringPtr("2030-12-31"),
		},
	}
	dateXMC := generateAndGetFieldXMC(t, dateField)
	assert.Equal(t, "2020-01-01", dateXMC["minDate"])
	assert.Equal(t, "2030-12-31", dateXMC["maxDate"])

	decimalField := &FieldDefinition{
		Name: "price", Title: "Price",
		Type:         GetFieldTypeByFormat(FormatDecimal),
		DisplayOrder: "a0",
		Validation: &ValidationConfig{
			Precision: intPtr(10),
			Scale:     intPtr(4),
		},
	}
	decimalXMC := generateAndGetFieldXMC(t, decimalField)
	assert.Equal(t, float64(10), decimalXMC["precision"])
	assert.Equal(t, float64(4), decimalXMC["scale"])

	ruleField := &FieldDefinition{
		Name: "email", Title: "Email",
		Type:         GetFieldTypeByFormat(FormatString),
		DisplayOrder: "a0",
		Validation: &ValidationConfig{
			Rule: EmailRule,
		},
	}
	ruleXMC := generateAndGetFieldXMC(t, ruleField)
	assert.Equal(t, "email", ruleXMC["validateRule"])
}

// TestXMC_NoUnknownFields 遍历 x-mc 所有 key，断言均在白名单内
func TestXMC_NoUnknownFields(t *testing.T) {
	whitelist := map[string]bool{
		"widget":       true,
		"isPrimary":    true,
		"isUnique":     true,
		"displayOrder": true,
		"format":       true,
		"nullable":     true,
		"storageHint":  true,
		"validateRule": true,
		"precision":    true,
		"scale":        true,
		"minDate":      true,
		"maxDate":      true,
		"minTime":      true,
		"maxTime":      true,
		"relation":     true,
		"enum":         true,
	}

	fkID := "fk-1"
	hint := "TEXT"
	fields := []*FieldDefinition{
		{
			Name: "f1", Title: "F1",
			Type: GetFieldTypeByFormat(FormatEnum),
			Enum: &EnumDefinition{
				Name:    "S",
				Options: []EnumOption{{Code: "A", Label: "A"}},
			},
			DisplayOrder: "a0",
		},
		{
			Name: "f2", Title: "F2",
			Type:          GetFieldTypeByFormat(FormatString),
			BelongsToFKID: &fkID,
			DisplayOrder:  "a1",
		},
		{
			Name: "f3", Title: "F3",
			Type:         GetFieldTypeByFormat(FormatString),
			StorageHint:  &hint,
			DisplayOrder: "a2",
		},
		{
			Name: "f4", Title: "F4",
			Type:         GetFieldTypeByFormat(FormatDecimal),
			DisplayOrder: "a3",
			Validation: &ValidationConfig{
				Precision: intPtr(10),
				Scale:     intPtr(2),
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	model := makeMinimalModel(fields)
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))
	props := schema["properties"].(map[string]interface{})

	for fieldName, fieldRaw := range props {
		fieldMap := fieldRaw.(map[string]interface{})
		xmc := getXMC(fieldMap)
		for key := range xmc {
			assert.True(t, whitelist[key], "field %s has unknown x-mc key: %s", fieldName, key)
		}
	}
}

// TestXMC_TopLevel_Clean 顶层无 nullable，无任何 x- 开头键（x-mc 除外）
func TestXMC_TopLevel_Clean(t *testing.T) {
	fkID := "fk-1"
	hint := "TEXT"
	fields := []*FieldDefinition{
		{
			Name: "id", Title: "ID",
			Type: GetFieldTypeByFormat(FormatUUID), IsPrimary: true,
			NonNull:      true,
			DisplayOrder: "a0",
		},
		{
			Name: "name", Title: "Name",
			Type:         GetFieldTypeByFormat(FormatString),
			StorageHint:  &hint,
			DisplayOrder: "a1",
		},
		{
			Name: "user_id", Title: "User ID",
			Type:          GetFieldTypeByFormat(FormatString),
			BelongsToFKID: &fkID,
			DisplayOrder:  "a2",
		},
		{
			Name: "event_date", Title: "Date",
			Type:         GetFieldTypeByFormat(FormatDate),
			DisplayOrder: "a3",
			Validation: &ValidationConfig{
				MinDate: stringPtr("2020-01-01"),
				MaxDate: stringPtr("2030-12-31"),
			},
		},
	}

	generator := NewJSONSchemaGenerator()
	model := makeMinimalModel(fields)
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))
	props := schema["properties"].(map[string]interface{})

	for fieldName, fieldRaw := range props {
		fieldMap := fieldRaw.(map[string]interface{})
		for key := range fieldMap {
			// 顶层禁止 nullable
			assert.NotEqual(t, "nullable", key,
				"field %s must not have top-level 'nullable' key", fieldName)
			// 顶层 x- 键只允许 x-mc
			if len(key) >= 2 && key[:2] == "x-" {
				assert.Equal(t, "x-mc", key,
					"field %s has forbidden top-level x- key: %s (only x-mc is allowed)", fieldName, key)
			}
		}
	}
}

// ─── 辅助函数 ────────────────────────────────────────────────────────────────

// makeMinimalModel 创建最小化的 DataModel 用于测试
func makeMinimalModel(fields []*FieldDefinition) *DataModel {
	return &DataModel{
		ModelMeta: ModelMeta{
			ID:           "test-id",
			ModelLocator: ModelLocator{ModelName: "TestModel", DatabaseName: "test-db"},
			Title:        "Test Model",
			StorageType:  "mysql",
		},
		Fields: fields,
	}
}

// generateAndGetFieldXMC 生成 schema 并返回指定字段名（取 Fields[0].Name）的 x-mc map
func generateAndGetFieldXMC(t *testing.T, field *FieldDefinition) map[string]interface{} {
	t.Helper()
	generator := NewJSONSchemaGenerator()
	model := makeMinimalModel([]*FieldDefinition{field})
	schemaJSON, err := generator.GenerateSchema(model)
	require.NoError(t, err)

	var schema map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(schemaJSON), &schema))
	props := schema["properties"].(map[string]interface{})
	fieldMap, ok := props[field.Name].(map[string]interface{})
	require.True(t, ok, "field %s not found in properties", field.Name)
	return getXMC(fieldMap)
}
