package adapter

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestModelMapper_ConvertToGraphQLModel_IncludesFieldEnumDefinition(t *testing.T) {
	fieldType, err := modeldesign.NewFieldFormat(modeldesign.FormatEnum)
	require.NoError(t, err)

	now := time.Now()
	enumDef := &modeldesign.EnumDefinition{
		ID: "enum-1",
		ProjectScope: project.ProjectScope{
			OrgName:     "org-a",
			ProjectSlug: "project-a",
		},
		Name:          "StatusEnum",
		DisplayName:   "Status",
		Description:   "status enum",
		IsMultiSelect: false,
		Options: []modeldesign.EnumOption{
			{
				Code:        "ACTIVE",
				Label:       "Active",
				Order:       1,
				Description: "active status",
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	modelEntity := &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID: "model-1",
			ModelLocator: modeldesign.ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "org-a",
					ProjectSlug: "project-a",
				},
				DatabaseName: "db1",
				ModelName:    "user",
			},
			Title:       "User",
			StorageType: "DB",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Fields: []*modeldesign.FieldDefinition{
			{
				ModelID:   "model-1",
				Name:      "status",
				Title:     "status",
				Type:      fieldType,
				EnumName:  "StatusEnum",
				Enum:      enumDef,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	gqlModel, err := ModelMapper.ConvertToGraphQLModel(modelEntity)
	require.NoError(t, err)
	require.Len(t, gqlModel.Fields, 1)
	require.NotNil(t, gqlModel.Fields[0].Enum)
	require.Equal(t, "StatusEnum", gqlModel.Fields[0].Enum.Name)
}

func TestModelMapper_ConvertToGraphQLModel_EndUserRefFormat(t *testing.T) {
	fieldType, err := modeldesign.NewFieldFormat(modeldesign.FormatEndUserRef)
	require.NoError(t, err)

	now := time.Now()
	modelEntity := &modeldesign.DataModel{
		ModelMeta: modeldesign.ModelMeta{
			ID: "model-1",
			ModelLocator: modeldesign.ModelLocator{
				ProjectScope: project.ProjectScope{
					OrgName:     "org-a",
					ProjectSlug: "project-a",
				},
				DatabaseName: "db1",
				ModelName:    "user",
			},
			Title:       "User",
			StorageType: "DB",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Fields: []*modeldesign.FieldDefinition{
			{
				ModelID:   "model-1",
				Name:      "owner",
				Title:     "Owner",
				Type:      fieldType,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	gqlModel, err := ModelMapper.ConvertToGraphQLModel(modelEntity)
	require.NoError(t, err)
	require.Len(t, gqlModel.Fields, 1)
	require.Equal(t, generated.FormatTypeEndUserRef, gqlModel.Fields[0].Format)
}
