package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/project"

	"github.com/stretchr/testify/require"
)

func TestEnumMapper_ConvertEnumDefinitionToGraphQL_TableDriven(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *modeldesign.EnumDefinition
	}{
		{
			name: "basic enum",
			in: &modeldesign.EnumDefinition{
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
					{Code: "ACTIVE", Label: "Active", Order: 1, Description: "active status"},
					{Code: "INACTIVE", Label: "Inactive", Order: 2, Description: ""},
				},
				CreatedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
				UpdatedAt: time.Date(2025, 1, 2, 4, 5, 6, 0, time.UTC),
			},
		},
		{
			name: "empty description still keeps non-nil pointer",
			in: &modeldesign.EnumDefinition{
				ID: "enum-2",
				ProjectScope: project.ProjectScope{
					OrgName:     "org-b",
					ProjectSlug: "project-b",
				},
				Name:          "PriorityEnum",
				DisplayName:   "Priority",
				Description:   "",
				IsMultiSelect: true,
				Options: []modeldesign.EnumOption{
					{Code: "P1", Label: "P1", Order: 1, Description: ""},
				},
				CreatedAt: time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC),
				UpdatedAt: time.Date(2025, 2, 3, 5, 6, 7, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := EnumMapper.ConvertEnumDefinitionToGraphQL(tc.in)
			require.NotNil(t, got)

			require.Equal(t, tc.in.ID, got.ID)
			require.Equal(t, tc.in.OrgName, got.OrgName)
			require.Equal(t, tc.in.ProjectSlug, got.ProjectSlug)
			require.Equal(t, tc.in.Name, got.Name)
			require.Equal(t, tc.in.DisplayName, got.DisplayName)
			require.NotNil(t, got.Description)
			require.Equal(t, tc.in.Description, *got.Description)
			require.Equal(t, tc.in.IsMultiSelect, got.IsMultiSelect)

			require.Equal(t, tc.in.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), got.CreatedAt)
			require.Equal(t, tc.in.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), got.UpdatedAt)

			require.Len(t, got.Options, len(tc.in.Options))
			for i, opt := range tc.in.Options {
				require.Equal(t, opt.Code, got.Options[i].Code)
				require.Equal(t, opt.Label, got.Options[i].Label)
				require.Equal(t, opt.Order, got.Options[i].Order)
				require.NotNil(t, got.Options[i].Description)
				require.Equal(t, opt.Description, *got.Options[i].Description)
			}
		})
	}
}

func TestEnumMapper_ConvertEnumDefinitionToGraphQL_NilInput(t *testing.T) {
	t.Parallel()

	got := EnumMapper.ConvertEnumDefinitionToGraphQL(nil)
	require.Nil(t, got)
}

func TestEnumMapper_ConvertEnumDefinitionToGraphQL_Golden(t *testing.T) {
	t.Parallel()

	in := &modeldesign.EnumDefinition{
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
			{Code: "ACTIVE", Label: "Active", Order: 1, Description: "active status"},
			{Code: "INACTIVE", Label: "Inactive", Order: 2, Description: ""},
		},
		CreatedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 2, 4, 5, 6, 0, time.UTC),
	}

	got := EnumMapper.ConvertEnumDefinitionToGraphQL(in)
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	require.NoError(t, err)

	goldenPath := filepath.Join("testdata", "enum_mapper", "basic.golden.json")
	wantJSON, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	require.JSONEq(t, string(wantJSON), string(gotJSON))
}

func FuzzEnumMapper_ConvertEnumDefinitionToGraphQL(f *testing.F) {
	f.Add(
		"enum-1",
		"org-a",
		"project-a",
		"StatusEnum",
		"Status",
		"status enum",
		"ACTIVE",
		"Active",
		int32(1),
		int64(1735787045),
		int64(1735790706),
		true,
	)

	f.Fuzz(func(
		t *testing.T,
		id string,
		orgName string,
		projectSlug string,
		name string,
		displayName string,
		description string,
		code string,
		label string,
		order int32,
		createdUnix int64,
		updatedUnix int64,
		isMultiSelect bool,
	) {
		in := &modeldesign.EnumDefinition{
			ID: id,
			ProjectScope: project.ProjectScope{
				OrgName:     orgName,
				ProjectSlug: projectSlug,
			},
			Name:          name,
			DisplayName:   displayName,
			Description:   description,
			IsMultiSelect: isMultiSelect,
			Options: []modeldesign.EnumOption{
				{Code: code, Label: label, Order: order, Description: description},
			},
			CreatedAt: time.Unix(createdUnix, 0).UTC(),
			UpdatedAt: time.Unix(updatedUnix, 0).UTC(),
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		got := EnumMapper.ConvertEnumDefinitionToGraphQL(in)
		require.NotNil(t, got)

		require.Equal(t, id, got.ID)
		require.Equal(t, orgName, got.OrgName)
		require.Equal(t, projectSlug, got.ProjectSlug)
		require.Equal(t, name, got.Name)
		require.Equal(t, displayName, got.DisplayName)
		require.NotNil(t, got.Description)
		require.Equal(t, description, *got.Description)
		require.Equal(t, isMultiSelect, got.IsMultiSelect)

		require.Equal(t, in.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), got.CreatedAt)
		require.Equal(t, in.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), got.UpdatedAt)

		require.Len(t, got.Options, 1)
		require.Equal(t, code, got.Options[0].Code)
		require.Equal(t, label, got.Options[0].Label)
		require.Equal(t, order, got.Options[0].Order)
		require.NotNil(t, got.Options[0].Description)
		require.Equal(t, description, *got.Options[0].Description)
	})
}
