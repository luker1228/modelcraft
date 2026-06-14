package rls

import (
	"context"
	"testing"

	domainrls "modelcraft/internal/domain/rls"

	"github.com/stretchr/testify/require"
)

type testModelSchema struct{ fields map[string]bool }

func (s testModelSchema) HasField(name string) bool { return s.fields[name] }

func (s testModelSchema) GetFieldNames() []string {
	names := make([]string, 0, len(s.fields))
	for name := range s.fields {
		names = append(names, name)
	}
	return names
}

type testAuthSchema struct{ refs map[string]bool }

func (s testAuthSchema) IsValidRef(name string) bool {
	return name == "userid" || name == "username" || name == "roles" || name == "uid" || name == "user_id" || s.refs[name]
}

func TestPolicyExpressionValidator_UsingAcceptsRowAndAuth(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(
		context.Background(),
		PolicyExpressionModeUsing,
		`row.owner_id == auth.userid && row.status in ["draft", "pending"]`,
		testModelSchema{fields: map[string]bool{"owner_id": true, "status": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Empty(t, errs)
}

func TestPolicyExpressionValidator_RejectsWrongRootForMode(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(
		context.Background(),
		PolicyExpressionModeUsing,
		`input.owner_id == auth.userid`,
		testModelSchema{fields: map[string]bool{"owner_id": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Len(t, errs, 1)
	require.Equal(t, "INVALID_CONTEXT", errs[0].Code)
}

func TestPolicyExpressionValidator_RejectsUnknownField(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(
		context.Background(),
		PolicyExpressionModeCheck,
		`input.missing == auth.userid`,
		testModelSchema{fields: map[string]bool{"owner_id": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Len(t, errs, 1)
	require.Equal(t, "UNKNOWN_FIELD", errs[0].Code)
}

func TestPolicyExpressionValidator_RejectsUnsupportedFunctionCall(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(
		context.Background(),
		PolicyExpressionModeUsing,
		`size(row.title) > 0`,
		testModelSchema{fields: map[string]bool{"title": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Len(t, errs, 1)
	require.Equal(t, "UNSUPPORTED_CALL", errs[0].Code)
}

func TestPolicyExpressionValidator_ValidateRoutesLegacyJSONToOldValidator(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.Validate(
		context.Background(),
		domainrls.JsonExpr(`{"owner_id":{"_eq":{"_auth":"uid"}}}`),
		domainrls.ExprTypeInsertCheck,
		testModelSchema{fields: map[string]bool{"owner_id": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Empty(t, errs)
}
