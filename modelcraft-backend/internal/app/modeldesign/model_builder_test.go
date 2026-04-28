package modeldesign

import (
	"context"
	"testing"

	domainmodel "modelcraft/internal/domain/modeldesign"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildModelFromTable_DoesNotInjectSystemFieldsButKeepsPrimaryKeyFlags(t *testing.T) {
	result, err := BuildModelFromTable(
		"users",
		"",
		[]domainmodel.TableColumn{
			{
				Name:     "id",
				DataType: "VARCHAR",
				Length:   36,
				Nullable: false,
				Comment:  "primary key",
			},
			{
				Name:     "username",
				DataType: "VARCHAR",
				Length:   64,
				Nullable: false,
			},
		},
		[]string{"id"},
		"org",
		"project",
		"mc_private_project",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Model)
	require.Len(t, result.Model.Fields, 2)

	var idCount int
	var ownerFound bool
	var idField *domainmodel.FieldDefinition
	for _, field := range result.Model.Fields {
		if field.Name == "id" {
			idCount++
			idField = field
		}
		if field.Name == "owner" {
			ownerFound = true
		}
	}

	assert.Equal(t, 1, idCount, "id should only appear once during import")
	assert.False(t, ownerFound, "import should not inject owner system field")
	require.NotNil(t, idField)
	assert.True(t, idField.IsPrimary, "import should preserve primary key flag")
	assert.True(t, idField.IsUnique, "import should preserve primary key uniqueness")
	assert.True(t, idField.NonNull, "import should preserve primary key non-null")
}

func TestNewModelFromCommand_SetsCreatedViaNew(t *testing.T) {
	model, err := newModelFromCommand(context.Background(), CreateModelCommand{
		OrgName:      "org",
		ProjectSlug:  "project",
		Name:         "orders",
		Title:        "Orders",
		StorageType:  "mysql",
		DatabaseName: "db_1",
	})
	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Equal(t, domainmodel.ModelCreationSourceNew, model.CreatedVia)
}

func TestBuildModelFromTable_SetsCreatedViaImported(t *testing.T) {
	result, err := BuildModelFromTable(
		"orders",
		"",
		[]domainmodel.TableColumn{{Name: "id", DataType: "VARCHAR", Length: 36, Nullable: false}},
		[]string{"id"},
		"org",
		"project",
		"db_1",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Model)
	assert.Equal(t, domainmodel.ModelCreationSourceImported, result.Model.CreatedVia)
}
