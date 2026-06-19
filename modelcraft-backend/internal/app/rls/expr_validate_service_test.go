package rls

import (
	"context"
	"testing"

	domainrls "modelcraft/internal/domain/rls"
	"modelcraft/internal/domain/modeldesign"

	"github.com/stretchr/testify/require"
)

// mockModelRepo is a minimal ModelRepository stub for ValidateAndDryRun tests.
type mockModelRepo struct {
	model *modeldesign.DataModel
	err   error
}

func (m *mockModelRepo) GetByID(_ context.Context, _ string, _ ...*modeldesign.ModelQueryOptions) (*modeldesign.DataModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.model, nil
}

// Stubs for unused interface methods — ModelRepository has many methods but
// ValidateAndDryRun only calls GetByID.
func (m *mockModelRepo) Save(_ context.Context, _ string, _ *modeldesign.DataModel) error               { return nil }
func (m *mockModelRepo) Update(_ context.Context, _ *modeldesign.DataModel) error                       { return nil }
func (m *mockModelRepo) UpdateWithVersion(_ context.Context, _ *modeldesign.DataModel, _ int64) (int64, error) {
	return 0, nil
}
func (m *mockModelRepo) Delete(_ context.Context, _ string) error                                       { return nil }
func (m *mockModelRepo) GetByName(_ context.Context, _, _, _, _ string, _ ...*modeldesign.ModelQueryOptions) (*modeldesign.DataModel, error) {
	return nil, nil
}
func (m *mockModelRepo) FindByDeploymentStatus(_ context.Context, _ ...modeldesign.DeploymentStatus) ([]modeldesign.DataModel, error) {
	return nil, nil
}
func (m *mockModelRepo) GetMetaByIDs(_ context.Context, _, _ string, _ []string) ([]*modeldesign.DataModel, error) {
	return nil, nil
}
func (m *mockModelRepo) Query(_ context.Context, _ modeldesign.ModelQuery) ([]modeldesign.DataModel, int, error) {
	return nil, 0, nil
}
func (m *mockModelRepo) ListDatabaseCatalog(_ context.Context, _, _, _ string, _, _ int) ([]string, int, error) {
	return nil, 0, nil
}
func (m *mockModelRepo) AddFields(_ context.Context, _ string, _ []*modeldesign.FieldDefinition) error {
	return nil
}
func (m *mockModelRepo) AddRelationField(_ context.Context, _ string, _ *modeldesign.FieldDefinition) error {
	return nil
}
func (m *mockModelRepo) GetFieldByModelID(_ context.Context, _, _ string) (*modeldesign.FieldDefinition, error) {
	return nil, nil
}
func (m *mockModelRepo) GetFieldsByModelID(_ context.Context, _ string) ([]*modeldesign.FieldDefinition, error) {
	return nil, nil
}
func (m *mockModelRepo) GetTailFieldDisplayOrder(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (m *mockModelRepo) UpdateField(_ context.Context, _ *modeldesign.FieldDefinition) error { return nil }
func (m *mockModelRepo) BulkUpdateFields(_ context.Context, _ []*modeldesign.FieldDefinition) error {
	return nil
}
func (m *mockModelRepo) UpdateFieldsStatus(_ context.Context, _ ...modeldesign.UpdateFieldsStatusRequest) error {
	return nil
}
func (m *mockModelRepo) DeleteFields(_ context.Context, _ string, _ []string) error { return nil }
func (m *mockModelRepo) BulkDeleteFields(_ context.Context, _ ...modeldesign.DeleteFieldRequest) error {
	return nil
}

func TestValidateAndDryRun_PredicateCompilesSQL(t *testing.T) {
	svc := NewRLSExprValidateService(&mockModelRepo{model: &modeldesign.DataModel{}})

	result := svc.ValidateAndDryRun(
		context.Background(),
		"org", "proj", "model-1",
		domainrls.ExprTypeSelectPredicate,
		`row.user_id == auth.userid`,
		nil,
		&domainrls.UserContext{UserIDStr: "u_123"},
	)

	require.True(t, result.Valid)
	require.NotEmpty(t, result.SQL)
	require.Equal(t, "user_id = ?", result.SQL)
	require.Equal(t, []any{"u_123"}, result.Params)
	require.Nil(t, result.Result)
}

func TestValidateAndDryRun_CheckEvaluatesTrue(t *testing.T) {
	svc := NewRLSExprValidateService(&mockModelRepo{model: &modeldesign.DataModel{}})

	result := svc.ValidateAndDryRun(
		context.Background(),
		"org", "proj", "model-1",
		domainrls.ExprTypeInsertCheck,
		`input.user_id == auth.userid`,
		map[string]any{"user_id": "u_123"},
		&domainrls.UserContext{UserIDStr: "u_123"},
	)

	require.True(t, result.Valid)
	require.Empty(t, result.SQL)
	require.NotNil(t, result.Result)
	require.True(t, *result.Result)
}

func TestValidateAndDryRun_CheckEvaluatesFalse(t *testing.T) {
	svc := NewRLSExprValidateService(&mockModelRepo{model: &modeldesign.DataModel{}})

	result := svc.ValidateAndDryRun(
		context.Background(),
		"org", "proj", "model-1",
		domainrls.ExprTypeInsertCheck,
		`input.user_id == auth.userid`,
		map[string]any{"user_id": "someone_else"},
		&domainrls.UserContext{UserIDStr: "u_123"},
	)

	// "evaluated to false" is a valid dry-run result (not a compile error)
	require.True(t, result.Valid)
	require.Empty(t, result.Errors)
	require.NotNil(t, result.Result)
	require.False(t, *result.Result)
}

func TestValidateAndDryRun_ModelNotFound(t *testing.T) {
	svc := NewRLSExprValidateService(&mockModelRepo{model: nil})

	result := svc.ValidateAndDryRun(
		context.Background(),
		"org", "proj", "nonexistent",
		domainrls.ExprTypeSelectPredicate,
		`row.user_id == auth.userid`,
		nil,
		&domainrls.UserContext{UserIDStr: "u_123"},
	)

	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	require.Equal(t, "MODEL_NOT_FOUND", result.Errors[0].Code)
	require.Equal(t, "modelId", result.Errors[0].Path)
}

func TestValidateAndDryRun_PredicateInvalidExpression(t *testing.T) {
	svc := NewRLSExprValidateService(&mockModelRepo{model: &modeldesign.DataModel{}})

	result := svc.ValidateAndDryRun(
		context.Background(),
		"org", "proj", "model-1",
		domainrls.ExprTypeSelectPredicate,
		`input.user_id == auth.userid`, // input.* not allowed in predicate
		nil,
		&domainrls.UserContext{UserIDStr: "u_123"},
	)

	require.False(t, result.Valid)
	require.Len(t, result.Errors, 1)
	require.Equal(t, "DRY_RUN_FAILED", result.Errors[0].Code)
}

func TestValidateAndDryRun_CheckWithNilSampleInput(t *testing.T) {
	svc := NewRLSExprValidateService(&mockModelRepo{model: &modeldesign.DataModel{}})

	result := svc.ValidateAndDryRun(
		context.Background(),
		"org", "proj", "model-1",
		domainrls.ExprTypeInsertCheck,
		`input.status == "draft"`,
		nil, // nil sampleInput should default to empty map, not panic
		&domainrls.UserContext{UserIDStr: "u_123"},
	)

	require.False(t, result.Valid)
	require.NotNil(t, result.Result)
	require.False(t, *result.Result) // empty input → status != "draft" → false
}

func TestValidateAndDryRun_NilUserCtx(t *testing.T) {
	svc := NewRLSExprValidateService(&mockModelRepo{model: &modeldesign.DataModel{}})

	result := svc.ValidateAndDryRun(
		context.Background(),
		"org", "proj", "model-1",
		domainrls.ExprTypeSelectPredicate,
		`row.user_id == auth.userid`,
		nil,
		nil, // nil userCtx should not panic
	)

	require.True(t, result.Valid)
	require.Equal(t, "user_id = ?", result.SQL)
}
