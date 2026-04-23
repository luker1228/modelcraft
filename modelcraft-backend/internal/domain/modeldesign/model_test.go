package modeldesign

import (
	"modelcraft/internal/domain/project"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModelLocator_String(t *testing.T) {
	tests := []struct {
		name     string
		locator  ModelLocator
		expected string
	}{
		{
			name: "standard locator",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "Order",
				DatabaseName: "main_db",
			},
			expected: "acme.ecommerce.main_db.Order",
		},
		{
			name: "empty fields",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "",
					ProjectSlug: "",
				},
				ModelName:    "",
				DatabaseName: "",
			},
			expected: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := tt.locator.String()

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModelLocator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		locator     ModelLocator
		wantErr     bool
		errContains string
	}{
		{
			name: "valid locator",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "Order",
				DatabaseName: "main_db",
			},
			wantErr: false,
		},
		{
			name: "missing org name",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "Order",
				DatabaseName: "main_db",
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name: "missing project slug",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "",
				},
				ModelName:    "Order",
				DatabaseName: "main_db",
			},
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name: "missing model name",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "",
				DatabaseName: "main_db",
			},
			wantErr:     true,
			errContains: "ModelName cant be blank",
		},
		{
			name: "missing database name",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "Order",
				DatabaseName: "",
			},
			wantErr:     true,
			errContains: "DatabaseName cant be blank",
		},
		{
			name: "all fields missing",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "",
					ProjectSlug: "",
				},
				ModelName:    "",
				DatabaseName: "",
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := tt.locator.Validate()

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewModelLocator(t *testing.T) {
	tests := []struct {
		name         string
		orgName      string
		projectSlug  string
		databaseName string
		modelName    string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid locator",
			orgName:      "acme",
			projectSlug:  "ecommerce",
			databaseName: "main_db",
			modelName:    "Order",
			wantErr:      false,
		},
		{
			name:         "missing org name",
			orgName:      "",
			projectSlug:  "ecommerce",
			databaseName: "main_db",
			modelName:    "Order",
			wantErr:      true,
			errContains:  "OrgName cant be blank",
		},
		{
			name:         "missing project slug",
			orgName:      "acme",
			projectSlug:  "",
			databaseName: "main_db",
			modelName:    "Order",
			wantErr:      true,
			errContains:  "ProjectSlug cant be blank",
		},
		{
			name:         "missing database name",
			orgName:      "acme",
			projectSlug:  "ecommerce",
			databaseName: "",
			modelName:    "Order",
			wantErr:      true,
			errContains:  "DatabaseName cant be blank",
		},
		{
			name:         "missing model name",
			orgName:      "acme",
			projectSlug:  "ecommerce",
			databaseName: "main_db",
			modelName:    "",
			wantErr:      true,
			errContains:  "ModelName cant be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			locator, err := NewModelLocator(tt.orgName, tt.projectSlug, tt.databaseName, tt.modelName)

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, locator)
				assert.Equal(t, tt.orgName, locator.OrgName)
				assert.Equal(t, tt.projectSlug, locator.ProjectSlug)
				assert.Equal(t, tt.databaseName, locator.DatabaseName)
				assert.Equal(t, tt.modelName, locator.ModelName)
			}
		})
	}
}

func TestModelLocator_GetFullPath(t *testing.T) {
	tests := []struct {
		name     string
		locator  ModelLocator
		expected string
	}{
		{
			name: "standard path",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "Order",
				DatabaseName: "main_db",
			},
			expected: "acme.ecommerce.main_db.Order",
		},
		{
			name: "with underscores",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "my_org",
					ProjectSlug: "my_project",
				},
				ModelName:    "user_profile",
				DatabaseName: "app_db",
			},
			expected: "my_org.my_project.app_db.user_profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := tt.locator.GetFullPath()

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModelLocator_GetDatabasePath(t *testing.T) {
	tests := []struct {
		name     string
		locator  ModelLocator
		expected string
	}{
		{
			name: "standard database path",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "Order",
				DatabaseName: "main_db",
			},
			expected: "acme.ecommerce.main_db",
		},
		{
			name: "with underscores",
			locator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "my_org",
					ProjectSlug: "my_project",
				},
				ModelName:    "user_profile",
				DatabaseName: "app_db",
			},
			expected: "my_org.my_project.app_db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := tt.locator.GetDatabasePath()

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataModel_GetField(t *testing.T) {
	// Given
	field1 := &FieldDefinition{Name: "id"}
	field2 := &FieldDefinition{Name: "name"}
	field3 := &FieldDefinition{Name: "email"}

	model := &DataModel{
		Fields: []*FieldDefinition{field1, field2, field3},
	}

	tests := []struct {
		name          string
		fieldName     string
		expectedField *FieldDefinition
	}{
		{
			name:          "existing field - first",
			fieldName:     "id",
			expectedField: field1,
		},
		{
			name:          "existing field - middle",
			fieldName:     "name",
			expectedField: field2,
		},
		{
			name:          "existing field - last",
			fieldName:     "email",
			expectedField: field3,
		},
		{
			name:          "non-existing field",
			fieldName:     "phone",
			expectedField: nil,
		},
		{
			name:          "empty field name",
			fieldName:     "",
			expectedField: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := model.GetField(tt.fieldName)

			// Then
			assert.Equal(t, tt.expectedField, result)
		})
	}
}

func TestDataModel_GetModelLocator(t *testing.T) {
	// Given
	model := &DataModel{
		ModelMeta: ModelMeta{
			ModelLocator: ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "acme",
					ProjectSlug: "ecommerce",
				},
				ModelName:    "Order",
				DatabaseName: "main_db",
			},
		},
	}

	// When
	locator := model.GetModelLocator()

	// Then
	assert.NotNil(t, locator)
	assert.Equal(t, "acme", locator.OrgName)
	assert.Equal(t, "ecommerce", locator.ProjectSlug)
	assert.Equal(t, "Order", locator.ModelName)
	assert.Equal(t, "main_db", locator.DatabaseName)
}

func TestDataModel_GetFieldsByNames(t *testing.T) {
	// Given
	field1 := &FieldDefinition{Name: "id"}
	field2 := &FieldDefinition{Name: "name"}
	field3 := &FieldDefinition{Name: "email"}
	field4 := &FieldDefinition{Name: "age"}

	model := &DataModel{
		Fields: []*FieldDefinition{field1, field2, field3, field4},
	}

	tests := []struct {
		name           string
		fieldNames     []string
		expectedFields []*FieldDefinition
	}{
		{
			name:           "single field",
			fieldNames:     []string{"name"},
			expectedFields: []*FieldDefinition{field2},
		},
		{
			name:           "multiple fields",
			fieldNames:     []string{"id", "email"},
			expectedFields: []*FieldDefinition{field1, field3},
		},
		{
			name:           "all fields",
			fieldNames:     []string{"id", "name", "email", "age"},
			expectedFields: []*FieldDefinition{field1, field2, field3, field4},
		},
		{
			name:           "non-existing field",
			fieldNames:     []string{"phone"},
			expectedFields: []*FieldDefinition{},
		},
		{
			name:           "mix of existing and non-existing",
			fieldNames:     []string{"name", "phone", "email"},
			expectedFields: []*FieldDefinition{field2, field3},
		},
		{
			name:           "empty list",
			fieldNames:     []string{},
			expectedFields: nil,
		},
		{
			name:           "nil list",
			fieldNames:     nil,
			expectedFields: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := model.GetFieldsByNames(tt.fieldNames)

			// Then
			assert.Equal(t, tt.expectedFields, result)
		})
	}
}

func TestDataModel_IsProtectedSystemModel(t *testing.T) {
	tests := []struct {
		name  string
		model *DataModel
		want  bool
	}{
		{
			name: "protected end user users model",
			model: &DataModel{ModelMeta: ModelMeta{
				CreatedVia: ModelCreationSourceImported,
				ModelLocator: ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "p1"},
					DatabaseName: "mc_meta",
					ModelName:    "end_user_users",
				},
			}},
			want: true,
		},
		{
			name: "protected end user accounts model",
			model: &DataModel{ModelMeta: ModelMeta{
				CreatedVia: ModelCreationSourceImported,
				ModelLocator: ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "p1"},
					DatabaseName: "mc_meta",
					ModelName:    "end_user_accounts",
				},
			}},
			want: true,
		},
		{
			name: "new end user users model is not protected",
			model: &DataModel{ModelMeta: ModelMeta{
				CreatedVia: ModelCreationSourceNew,
				ModelLocator: ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "p1"},
					DatabaseName: "mc_meta",
					ModelName:    "end_user_users",
				},
			}},
			want: false,
		},
		{
			name: "imported normal db model is not protected",
			model: &DataModel{ModelMeta: ModelMeta{
				CreatedVia: ModelCreationSourceImported,
				ModelLocator: ModelLocator{
					ProjectScope: project.ProjectScope{OrgName: "org1", ProjectSlug: "p1"},
					DatabaseName: "app_db",
					ModelName:    "users",
				},
			}},
			want: false,
		},
		{
			name:  "nil model",
			model: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.model.IsProtectedSystemModel())
		})
	}
}

func TestDataModel_Update(t *testing.T) {
	tests := []struct {
		name             string
		initialTitle     string
		initialDesc      string
		updateTitle      *string
		updateDesc       *string
		expectedTitle    string
		expectedDesc     string
		expectTimeChange bool
	}{
		{
			name:             "update title only",
			initialTitle:     "Old Title",
			initialDesc:      "Old Desc",
			updateTitle:      stringPtr("New Title"),
			updateDesc:       nil,
			expectedTitle:    "New Title",
			expectedDesc:     "Old Desc",
			expectTimeChange: true,
		},
		{
			name:             "update description only",
			initialTitle:     "Old Title",
			initialDesc:      "Old Desc",
			updateTitle:      nil,
			updateDesc:       stringPtr("New Desc"),
			expectedTitle:    "Old Title",
			expectedDesc:     "New Desc",
			expectTimeChange: true,
		},
		{
			name:             "update both title and description",
			initialTitle:     "Old Title",
			initialDesc:      "Old Desc",
			updateTitle:      stringPtr("New Title"),
			updateDesc:       stringPtr("New Desc"),
			expectedTitle:    "New Title",
			expectedDesc:     "New Desc",
			expectTimeChange: true,
		},
		{
			name:             "update nothing",
			initialTitle:     "Old Title",
			initialDesc:      "Old Desc",
			updateTitle:      nil,
			updateDesc:       nil,
			expectedTitle:    "Old Title",
			expectedDesc:     "Old Desc",
			expectTimeChange: true,
		},
		{
			name:             "update to empty title",
			initialTitle:     "Old Title",
			initialDesc:      "Old Desc",
			updateTitle:      stringPtr(""),
			updateDesc:       nil,
			expectedTitle:    "",
			expectedDesc:     "Old Desc",
			expectTimeChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			oldTime := time.Now().Add(-1 * time.Hour)
			model := &DataModel{
				ModelMeta: ModelMeta{
					Title:       tt.initialTitle,
					Description: tt.initialDesc,
					UpdatedAt:   oldTime,
				},
			}

			// When
			err := model.Update(tt.updateTitle, tt.updateDesc)

			// Then
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTitle, model.Title)
			assert.Equal(t, tt.expectedDesc, model.Description)
			if tt.expectTimeChange {
				assert.True(t, model.UpdatedAt.After(oldTime))
			}
		})
	}
}

func TestDataModel_Validate(t *testing.T) {
	validFieldType, _ := NewFieldFormat(FormatString)
	validField := &FieldDefinition{
		Name:        "id",
		Title:       "ID",
		Type:        validFieldType,
		StorageHint: stringPtr("VARCHAR(255)"),
		ModelID:     "model-123",
		ModelLocator: &ModelLocator{
			ProjectScope: project.ProjectScope{
				OrgName:     "acme",
				ProjectSlug: "test",
			},
			ModelName:    "Test",
			DatabaseName: "test_db",
		},
	}

	tests := []struct {
		name        string
		model       *DataModel
		wantErr     bool
		errContains string
	}{
		{
			name: "valid model",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
					Title:       "Order",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{validField},
			},
			wantErr: false,
		},
		{
			name: "missing org name",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
					Title:       "Order",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{validField},
			},
			wantErr:     true,
			errContains: "OrgName cant be blank",
		},
		{
			name: "missing project slug",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
					Title:       "Order",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{validField},
			},
			wantErr:     true,
			errContains: "ProjectSlug cant be blank",
		},
		{
			name: "missing model name",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "",
						DatabaseName: "main_db",
					},
					Title:       "Order",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{validField},
			},
			wantErr:     true,
			errContains: "ModelName cant be blank",
		},
		{
			name: "missing title",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
					Title:       "",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{validField},
			},
			wantErr:     true,
			errContains: "Title cant be blank",
		},
		{
			name: "missing storage type",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
					Title:       "Order",
					StorageType: "",
				},
				Fields: []*FieldDefinition{validField},
			},
			wantErr:     true,
			errContains: "StorageType cant be blank",
		},
		{
			name: "missing database name",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "",
					},
					Title:       "Order",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{validField},
			},
			wantErr:     true,
			errContains: "DatabaseName cant be blank",
		},
		{
			name: "no fields",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
					Title:       "Order",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{},
			},
			wantErr:     true,
			errContains: "Fields cant be blank",
		},
		{
			name: "duplicate field names",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
					Title:       "Order",
					StorageType: "mysql",
				},
				Fields: []*FieldDefinition{
					{
						Name:        "id",
						Title:       "ID",
						Type:        validFieldType,
						StorageHint: stringPtr("VARCHAR(255)"),
						ModelID:     "model-123",
						ModelLocator: &ModelLocator{
							ProjectScope: project.ProjectScope{
								OrgName:     "acme",
								ProjectSlug: "test",
							},
							ModelName:    "Test",
							DatabaseName: "test_db",
						},
					},
					{
						Name:        "id",
						Title:       "ID2",
						Type:        validFieldType,
						StorageHint: stringPtr("VARCHAR(255)"),
						ModelID:     "model-123",
						ModelLocator: &ModelLocator{
							ProjectScope: project.ProjectScope{
								OrgName:     "acme",
								ProjectSlug: "test",
							},
							ModelName:    "Test",
							DatabaseName: "test_db",
						},
					},
				},
			},
			wantErr:     true,
			errContains: "字段名称 'id' 重复",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			err := tt.model.Validate()

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataModel_validateDuplicateFields(t *testing.T) {
	tests := []struct {
		name        string
		fields      []*FieldDefinition
		wantErr     bool
		errContains string
	}{
		{
			name:    "no fields",
			fields:  []*FieldDefinition{},
			wantErr: false,
		},
		{
			name: "unique field names",
			fields: []*FieldDefinition{
				{Name: "id"},
				{Name: "name"},
				{Name: "email"},
			},
			wantErr: false,
		},
		{
			name: "duplicate field names",
			fields: []*FieldDefinition{
				{Name: "id"},
				{Name: "name"},
				{Name: "id"},
			},
			wantErr:     true,
			errContains: "字段名称 'id' 重复",
		},
		{
			name: "multiple duplicates",
			fields: []*FieldDefinition{
				{Name: "id"},
				{Name: "name"},
				{Name: "id"},
				{Name: "name"},
			},
			wantErr:     true,
			errContains: "字段名称",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			model := &DataModel{Fields: tt.fields}

			// When
			err := model.validateDuplicateFields()

			// Then
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataModel_GetBizUniqueName(t *testing.T) {
	tests := []struct {
		name     string
		model    *DataModel
		expected string
	}{
		{
			name: "standard model",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
				},
			},
			expected: "acme.ecommerce.main_db.Order",
		},
		{
			name: "with underscores",
			model: &DataModel{
				ModelMeta: ModelMeta{
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "my_org",
							ProjectSlug: "my_project",
						},
						ModelName:    "user_profile",
						DatabaseName: "app_db",
					},
				},
			},
			expected: "my_org.my_project.app_db.user_profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := tt.model.GetBizUniqueName()

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataModel_AddFields(t *testing.T) {
	tests := []struct {
		name           string
		existingFields []*FieldDefinition
		newFields      []*FieldDefinition
		expectedCount  int
	}{
		{
			name:           "add to empty model",
			existingFields: []*FieldDefinition{},
			newFields: []*FieldDefinition{
				{Name: "id"},
				{Name: "name"},
			},
			expectedCount: 2,
		},
		{
			name: "add to existing fields",
			existingFields: []*FieldDefinition{
				{Name: "id"},
			},
			newFields: []*FieldDefinition{
				{Name: "name"},
				{Name: "email"},
			},
			expectedCount: 3,
		},
		{
			name: "add empty list",
			existingFields: []*FieldDefinition{
				{Name: "id"},
			},
			newFields:     []*FieldDefinition{},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			model := &DataModel{
				ModelMeta: ModelMeta{
					ID: "model-123",
					ModelLocator: ModelLocator{
						ProjectScope: project.ProjectScope{
							OrgName:     "acme",
							ProjectSlug: "ecommerce",
						},
						ModelName:    "Order",
						DatabaseName: "main_db",
					},
				},
				Fields: tt.existingFields,
			}

			// When
			model.AddFields(tt.newFields)

			// Then
			assert.Equal(t, tt.expectedCount, len(model.Fields))

			// Verify that ModelID and ModelLocator are set for new fields
			startIdx := len(tt.existingFields)
			for i := startIdx; i < len(model.Fields); i++ {
				assert.Equal(t, "model-123", model.Fields[i].ModelID)
				assert.Equal(t, &model.ModelLocator, model.Fields[i].ModelLocator)
			}
		})
	}
}
