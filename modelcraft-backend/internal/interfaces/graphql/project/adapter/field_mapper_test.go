package adapter

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertFormatType2Domain_TableDriven(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		format  generated.FormatType
		want    modeldesign.FormatType
		wantErr bool
	}{
		{name: "string", format: generated.FormatTypeString, want: modeldesign.FormatString},
		{name: "integer", format: generated.FormatTypeInteger, want: modeldesign.FormatInteger},
		{name: "number", format: generated.FormatTypeNumber, want: modeldesign.FormatNumber},
		{name: "boolean", format: generated.FormatTypeBoolean, want: modeldesign.FormatBoolean},
		{name: "datetime", format: generated.FormatTypeDatetime, want: modeldesign.FormatDateTime},
		{name: "date", format: generated.FormatTypeDate, want: modeldesign.FormatDate},
		{name: "time", format: generated.FormatTypeTime, want: modeldesign.FormatTime},
		{name: "uuid", format: generated.FormatTypeUUID, want: modeldesign.FormatUUID},
		{name: "decimal", format: generated.FormatTypeDecimal, want: modeldesign.FormatDecimal},
		{name: "relation", format: generated.FormatTypeRelation, want: modeldesign.FormatRelation},
		{name: "enum", format: generated.FormatTypeEnum, want: modeldesign.FormatEnum},
		{name: "end_user_ref", format: generated.FormatTypeEndUserRef, want: modeldesign.FormatEndUserRef},
		{name: "unknown", format: generated.FormatType("UNKNOWN"), wantErr: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := convertFormatType2Domain(tc.format)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFieldMapper_ConvertAddFieldInputToDTO_TableDriven(t *testing.T) {
	t.Parallel()

	type addFieldDTOCheckFunc func(
		t *testing.T,
		gotName, gotTitle string,
		gotFormat modeldesign.FormatType,
		gotNonNull, gotRequired, gotUnique, gotArray bool,
		gotDescription string,
		gotEnumName, gotFKID *string,
	)

	testCases := []struct {
		name  string
		input *generated.AddFieldInput
		check addFieldDTOCheckFunc
	}{
		{
			name: "enum field keeps enum name and pointer bools",
			input: &generated.AddFieldInput{
				Name:           "status",
				Title:          "Status",
				Format:         generated.FormatTypeEnum,
				NonNull:        boolPtr(true),
				Required:       boolPtr(true),
				IsUnique:       boolPtr(false),
				IsArray:        boolPtr(true),
				Description:    stringPtr("enum field"),
				RelateEnumName: stringPtr("StatusEnum"),
			},
			check: func(
				t *testing.T,
				gotName, gotTitle string,
				gotFormat modeldesign.FormatType,
				gotNonNull, gotRequired, gotUnique, gotArray bool,
				gotDescription string,
				gotEnumName, gotFKID *string,
			) {
				require.Equal(t, "status", gotName)
				require.Equal(t, "Status", gotTitle)
				require.Equal(t, modeldesign.FormatEnum, gotFormat)
				require.True(t, gotNonNull)
				require.True(t, gotRequired)
				require.False(t, gotUnique)
				require.True(t, gotArray)
				require.Equal(t, "enum field", gotDescription)
				require.NotNil(t, gotEnumName)
				require.Equal(t, "StatusEnum", *gotEnumName)
				require.Nil(t, gotFKID)
			},
		},
		{
			name: "relation field keeps fk id and drops enum name",
			input: &generated.AddFieldInput{
				Name:           "user",
				Title:          "User",
				Format:         generated.FormatTypeRelation,
				RelateFkID:     stringPtr("fk-user"),
				RelateEnumName: stringPtr("ShouldBeDropped"),
			},
			check: func(
				t *testing.T,
				gotName, gotTitle string,
				gotFormat modeldesign.FormatType,
				gotNonNull, gotRequired, gotUnique, gotArray bool,
				gotDescription string,
				gotEnumName, gotFKID *string,
			) {
				require.Equal(t, "user", gotName)
				require.Equal(t, "User", gotTitle)
				require.Equal(t, modeldesign.FormatRelation, gotFormat)
				require.False(t, gotNonNull)
				require.False(t, gotRequired)
				require.False(t, gotUnique)
				require.False(t, gotArray)
				require.Equal(t, "", gotDescription)
				require.Nil(t, gotEnumName)
				require.NotNil(t, gotFKID)
				require.Equal(t, "fk-user", *gotFKID)
			},
		},
		{
			name: "defaults for optional bool pointers",
			input: &generated.AddFieldInput{
				Name:   "nickname",
				Title:  "Nickname",
				Format: generated.FormatTypeString,
			},
			check: func(
				t *testing.T,
				gotName, gotTitle string,
				gotFormat modeldesign.FormatType,
				gotNonNull, gotRequired, gotUnique, gotArray bool,
				gotDescription string,
				gotEnumName, gotFKID *string,
			) {
				require.Equal(t, "nickname", gotName)
				require.Equal(t, "Nickname", gotTitle)
				require.Equal(t, modeldesign.FormatString, gotFormat)
				require.False(t, gotNonNull)
				require.False(t, gotRequired)
				require.False(t, gotUnique)
				require.False(t, gotArray)
				require.Equal(t, "", gotDescription)
				require.Nil(t, gotEnumName)
				require.Nil(t, gotFKID)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := FieldMapper.ConvertAddFieldInputToDTO(tc.input)
			require.NoError(t, err)
			require.NotNil(t, got)

			tc.check(
				t,
				got.Name,
				got.Title,
				got.Format,
				got.NonNull,
				got.Required,
				got.IsUnique,
				got.IsArray,
				got.Description,
				got.RelateEnumName,
				got.RelateFKID,
			)
		})
	}
}

func TestFieldMapper_ConvertAddFieldInputToDTO_NilInput(t *testing.T) {
	t.Parallel()

	got, err := FieldMapper.ConvertAddFieldInputToDTO(nil)
	require.Error(t, err)
	require.Nil(t, got)
}

func FuzzFieldMapper_ConvertAddFieldInputToDTO(f *testing.F) {
	f.Add("name", "title", "desc", uint8(0), true, false, true, false, uint8(255))
	f.Add("", "", "", uint8(5), false, false, false, false, uint8(0))

	formats := []generated.FormatType{
		generated.FormatTypeString,
		generated.FormatTypeInteger,
		generated.FormatTypeNumber,
		generated.FormatTypeBoolean,
		generated.FormatTypeDatetime,
		generated.FormatTypeDate,
		generated.FormatTypeTime,
		generated.FormatTypeUUID,
		generated.FormatTypeDecimal,
		generated.FormatTypeRelation,
		generated.FormatTypeEnum,
		generated.FormatTypeEndUserRef,
	}

	f.Fuzz(func(
		t *testing.T,
		name string,
		title string,
		description string,
		formatIndex uint8,
		nonNull bool,
		required bool,
		isUnique bool,
		isArray bool,
		flags uint8,
	) {
		format := formats[int(formatIndex)%len(formats)]
		in := &generated.AddFieldInput{
			Name:   name,
			Title:  title,
			Format: format,
		}

		if flags&1 != 0 {
			in.NonNull = &nonNull
		}
		if flags&2 != 0 {
			in.Required = &required
		}
		if flags&4 != 0 {
			in.IsUnique = &isUnique
		}
		if flags&8 != 0 {
			in.IsArray = &isArray
		}
		if flags&16 != 0 {
			in.Description = &description
		}

		enumName := "Enum_" + name
		if flags&32 != 0 {
			in.RelateEnumName = &enumName
		}

		fkID := "fk_" + title
		if flags&64 != 0 {
			in.RelateFkID = &fkID
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()

		got, err := FieldMapper.ConvertAddFieldInputToDTO(in)
		require.NoError(t, err)
		require.NotNil(t, got)

		require.Equal(t, name, got.Name)
		require.Equal(t, title, got.Title)

		if in.NonNull == nil {
			require.False(t, got.NonNull)
		} else {
			require.Equal(t, *in.NonNull, got.NonNull)
		}

		if in.Required == nil {
			require.False(t, got.Required)
		} else {
			require.Equal(t, *in.Required, got.Required)
		}

		if in.IsUnique == nil {
			require.False(t, got.IsUnique)
		} else {
			require.Equal(t, *in.IsUnique, got.IsUnique)
		}

		if in.IsArray == nil {
			require.False(t, got.IsArray)
		} else {
			require.Equal(t, *in.IsArray, got.IsArray)
		}

		if format == generated.FormatTypeEnum && in.RelateEnumName != nil {
			require.NotNil(t, got.RelateEnumName)
			require.Equal(t, *in.RelateEnumName, *got.RelateEnumName)
		} else {
			require.Nil(t, got.RelateEnumName)
		}
	})
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
