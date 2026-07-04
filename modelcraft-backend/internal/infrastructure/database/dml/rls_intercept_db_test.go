package dml

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"testing"

	"github.com/google/cel-go/cel"
)

// mockClientDB implements ClientDatabaseRepository for testing.
type mockClientDB struct {
	findManyCalled bool
	findManyInput  *modelruntime.FindManyInput
	findManyResult []map[string]any

	createOneCalled bool
	createOneInput  *modelruntime.CreateOneInput
	createOneResult string

	updateManyCalled bool
	updateManyInput  *modelruntime.UpdateManyInput
}

func (m *mockClientDB) FindMany(ctx context.Context, input *modelruntime.FindManyInput) ([]map[string]any, error) {
	m.findManyCalled = true
	m.findManyInput = input
	return m.findManyResult, nil
}

func (m *mockClientDB) FindUnique(ctx context.Context, input *modelruntime.FindUniqueInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) FindFirst(ctx context.Context, input *modelruntime.FindFirstInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) ListByCursor(
	ctx context.Context, input *modelruntime.ListByCursorInput,
) ([]map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) FindManyIn(ctx context.Context, input *modelruntime.FindManyInInput) ([]map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) Aggregate(ctx context.Context, input *modelruntime.AggregateInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) Count(ctx context.Context, input *modelruntime.CountInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) CreateOne(ctx context.Context, input *modelruntime.CreateOneInput) (string, error) {
	m.createOneCalled = true
	m.createOneInput = input
	return m.createOneResult, nil
}

func (m *mockClientDB) UpdateOne(ctx context.Context, input *modelruntime.UpdateOneInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) DeleteOne(ctx context.Context, input *modelruntime.DeleteOneInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDB) CreateMany(ctx context.Context, input *modelruntime.CreateManyInput) (interface{}, error) {
	return nil, nil
}

func (m *mockClientDB) UpdateMany(ctx context.Context, input *modelruntime.UpdateManyInput) (interface{}, error) {
	m.updateManyCalled = true
	m.updateManyInput = input
	return nil, nil
}

func (m *mockClientDB) DeleteMany(ctx context.Context, input *modelruntime.DeleteManyInput) (interface{}, error) {
	return nil, nil
}

func makeTrueCheckProgram(t *testing.T) *modelruntime.CheckProgram {
	t.Helper()
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		t.Fatal(err)
	}
	ast, issues := env.Compile("true")
	if issues != nil && issues.Err() != nil {
		t.Fatal(issues.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}
	return modelruntime.NewCheckProgram(program)
}

func makeFalseCheckProgram(t *testing.T) *modelruntime.CheckProgram {
	t.Helper()
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		t.Fatal(err)
	}
	ast, issues := env.Compile("false")
	if issues != nil && issues.Err() != nil {
		t.Fatal(issues.Err())
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}
	return modelruntime.NewCheckProgram(program)
}

func TestRLSInterceptDB_FindMany_InjectsUSING(t *testing.T) {
	mock := &mockClientDB{findManyResult: []map[string]any{{"id": "1"}}}
	db := NewRLSInterceptDB(mock)

	snap := &modelruntime.RLSPolicySnapshot{
		SelectFilter: &modelruntime.RawSQLFilter{SQL: "owner_id = ?", Params: []any{"u_123"}},
	}
	ctx := modelruntime.WithRLSSnapshot(context.Background(), snap)

	input := &modelruntime.FindManyInput{
		TableName:  "posts",
		Where:      map[string]any{"status": "draft"},
		RawFilters: []modelruntime.RawSQLFilter{},
		Limit:      10,
	}

	result, err := db.FindMany(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if !mock.findManyCalled {
		t.Fatal("inner FindMany was not called")
	}
	if len(mock.findManyInput.RawFilters) != 1 {
		t.Fatalf("expected 1 RawFilter, got %d", len(mock.findManyInput.RawFilters))
	}
	if mock.findManyInput.RawFilters[0].SQL != "owner_id = ?" {
		t.Errorf("expected SQL 'owner_id = ?', got '%s'", mock.findManyInput.RawFilters[0].SQL)
	}
}

func TestRLSInterceptDB_FindMany_NoSnapshot(t *testing.T) {
	mock := &mockClientDB{findManyResult: []map[string]any{{"id": "1"}}}
	db := NewRLSInterceptDB(mock)

	input := &modelruntime.FindManyInput{
		TableName:  "posts",
		RawFilters: []modelruntime.RawSQLFilter{},
		Limit:      10,
	}

	_, err := db.FindMany(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.findManyInput.RawFilters) != 0 {
		t.Errorf("expected 0 RawFilters, got %d", len(mock.findManyInput.RawFilters))
	}
}

func TestRLSInterceptDB_CreateOne_CHECKAcepted(t *testing.T) {
	mock := &mockClientDB{createOneResult: "new-id"}
	db := NewRLSInterceptDB(mock)

	snap := &modelruntime.RLSPolicySnapshot{
		CreateChecks: []*modelruntime.CheckProgram{makeTrueCheckProgram(t)},
		Auth:         map[string]any{"userid": "u_123"},
	}
	ctx := modelruntime.WithRLSSnapshot(context.Background(), snap)

	input := &modelruntime.CreateOneInput{
		TableName: "posts",
		Data:      map[string]any{"title": "hello"},
	}

	result, err := db.CreateOne(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "new-id" {
		t.Errorf("expected 'new-id', got '%s'", result)
	}
	if !mock.createOneCalled {
		t.Fatal("inner CreateOne was not called")
	}
}

func TestRLSInterceptDB_CreateOne_CHECKRejected(t *testing.T) {
	mock := &mockClientDB{createOneResult: "new-id"}
	db := NewRLSInterceptDB(mock)

	snap := &modelruntime.RLSPolicySnapshot{
		CreateChecks: []*modelruntime.CheckProgram{makeFalseCheckProgram(t)},
		Auth:         map[string]any{"userid": "u_123"},
	}
	ctx := modelruntime.WithRLSSnapshot(context.Background(), snap)

	input := &modelruntime.CreateOneInput{
		TableName: "posts",
		Data:      map[string]any{"title": "hello"},
	}

	_, err := db.CreateOne(ctx, input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if mock.createOneCalled {
		t.Fatal("inner CreateOne should NOT be called on CHECK rejection")
	}
}

func TestRLSInterceptDB_UpdateMany_USINGAndCHECK(t *testing.T) {
	mock := &mockClientDB{}
	db := NewRLSInterceptDB(mock)

	snap := &modelruntime.RLSPolicySnapshot{
		UpdateFilter: &modelruntime.RawSQLFilter{SQL: "owner_id = ?", Params: []any{"u_123"}},
		UpdateChecks: []*modelruntime.CheckProgram{makeTrueCheckProgram(t)},
		Auth:         map[string]any{"userid": "u_123"},
	}
	ctx := modelruntime.WithRLSSnapshot(context.Background(), snap)

	input := &modelruntime.UpdateManyInput{
		TableName:  "posts",
		Data:       map[string]any{"status": "published"},
		RawFilters: []modelruntime.RawSQLFilter{},
		Take:       5,
	}

	_, err := db.UpdateMany(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.updateManyCalled {
		t.Fatal("inner UpdateMany was not called")
	}
	if len(mock.updateManyInput.RawFilters) != 1 {
		t.Fatalf("expected 1 RawFilter, got %d", len(mock.updateManyInput.RawFilters))
	}
}
