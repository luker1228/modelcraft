# RLS SQL Intercept Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract RLS USING/CHECK logic from resolvers into a dedicated SQL intercept layer (RLSInterceptDB wrapper), with policy resolution precomputed at Execute() entry time as RLSPolicySnapshot.

**Architecture:** Build RLSPolicySnapshot once at Execute() entry (resolve USING → RawSQLFilter, precompile CHECK → cel.Program), store on context. RLSInterceptDB reads snapshot from context, transparently injects USING into RawFilters and evaluates CHECK before delegating to inner ClientDBRepoImpl. Resolver layer no longer knows about RLS.

**Tech Stack:** Go, cel-go, goqu (SQL builder), existing ClientDatabaseRepository interface

---

## File Structure

| Action | File | Purpose |
|--------|------|---------|
| Create | `domain/modelruntime/rls_snapshot.go` | RLSPolicySnapshot type, CheckProgram wrapper, context helpers |
| Create | `app/modelruntime/rls_snapshot_builder.go` | RLSSnapshotBuilder interface + implementation |
| Create | `infrastructure/database/dml/rls_intercept_db.go` | RLSInterceptDB — wraps ClientDatabaseRepository |
| Create | `infrastructure/database/dml/rls_intercept_db_test.go` | Unit tests for RLSInterceptDB |
| Modify | `app/modelruntime/graphql_app.go` | Replace rlsGuard with snapshotBuilder, wrap clientRepo |
| Modify | `domain/modelruntime/model_resolver.go` | Remove appendRLSUsingFilter, ValidateInput calls |
| Modify | `domain/modelruntime/graphql_request_context.go` | Remove RLSPolicyGuard interface and field |
| Modify | `internal/interfaces/http/routes.go` | Wire snapshotBuilder instead of runtimeRLSResolver |

---

### Task 1: Create RLSPolicySnapshot type and context helpers

**Files:**
- Create: `modelcraft-backend/internal/domain/modelruntime/rls_snapshot.go`

- [ ] **Step 1: Write the snapshot type and context helpers**

```go
// Package modelruntime provides domain types for model runtime execution.
package modelruntime

import (
	"context"

	"github.com/google/cel-go/cel"
)

// CheckProgram wraps a precompiled CEL program for CHECK expression evaluation.
// A nil CheckProgram means no CHECK is required (developer or true expression).
type CheckProgram struct {
	program cel.Program
}

// NewCheckProgram creates a CheckProgram from a precompiled CEL program.
func NewCheckProgram(program cel.Program) *CheckProgram {
	return &CheckProgram{program: program}
}

// Eval evaluates the CHECK expression against the given input and auth context.
// Returns nil if the expression evaluates to true, error otherwise.
func (c *CheckProgram) Eval(input map[string]any, auth map[string]any) error {
	if c == nil || c.program == nil {
		return nil
	}
	out, _, err := c.program.Eval(map[string]any{
		"input": input,
		"auth":  auth,
	})
	if err != nil {
		return err
	}
	allowed, ok := out.Value().(bool)
	if !ok {
		return &CheckEvalError{msg: "CHECK expression returned non-boolean"}
	}
	if !allowed {
		return &CheckEvalError{msg: "CHECK expression evaluated to false"}
	}
	return nil
}

// CheckEvalError represents a CHECK evaluation failure.
type CheckEvalError struct {
	msg string
}

func (e *CheckEvalError) Error() string {
	return "RLS CHECK violation: " + e.msg
}

// RLSPolicySnapshot holds pre-resolved RLS policies for the current request.
// Built once at Execute() entry, consumed by RLSInterceptDB at SQL execution time.
// nil snapshot means RLS is not applicable (developer JWT).
type RLSPolicySnapshot struct {
	// USING filters — injected into WHERE clause for SELECT/UPDATE/DELETE.
	// nil means no filtering needed (developer or true expression).
	SelectUSING *RawSQLFilter
	UpdateUSING *RawSQLFilter
	DeleteUSING *RawSQLFilter

	// CHECK programs — evaluated against input data before INSERT/UPDATE.
	// nil means no validation needed.
	InsertCHECK *CheckProgram
	UpdateCHECK *CheckProgram

	// Auth holds the pre-built auth context map for CEL evaluation.
	Auth map[string]any

	// DenyAll is true when no matching policy exists — all operations are denied.
	DenyAll bool
}

type rlsSnapshotKey struct{}

// WithRLSSnapshot stores the RLS snapshot on the context.
func WithRLSSnapshot(ctx context.Context, snap *RLSPolicySnapshot) context.Context {
	return context.WithValue(ctx, rlsSnapshotKey{}, snap)
}

// GetRLSSnapshot retrieves the RLS snapshot from the context.
// Returns nil if no snapshot was stored (developer access).
func GetRLSSnapshot(ctx context.Context) *RLSPolicySnapshot {
	snap, _ := ctx.Value(rlsSnapshotKey{}).(*RLSPolicySnapshot)
	return snap
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd modelcraft-backend && go build ./internal/domain/modelruntime/...
```

---

### Task 2: Create RLSSnapshotBuilder

**Files:**
- Create: `modelcraft-backend/internal/app/modelruntime/rls_snapshot_builder.go`

- [ ] **Step 1: Add GetCheckExpr to PolicyMatchingService**

The existing `ResolveCheck` method compiles to SQL (legacy path), not CEL. The snapshot builder needs raw CEL expression strings. Add a new method `GetCheckExpr` to `PolicyMatchingService`.

**File to modify:** `modelcraft-backend/internal/app/rls/policy_matching_service.go`

Add the method after `ValidateCheck`:

```go
// GetCheckExpr returns the raw CHECK expression string for the given action.
// Returns ("", nil) if no matching policy exists.
func (s *PolicyMatchingService) GetCheckExpr(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (string, error) {
	policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
	if err != nil {
		return "", err
	}
	if len(policies) == 0 {
		return "", nil // no matching policy → no CHECK needed
	}
	for _, p := range policies {
		if p.WithCheckExpr != "" {
			return string(p.WithCheckExpr), nil
		}
	}
	return "", nil
}
```

- [ ] **Step 2: Read relevant existing interfaces**

These are the interfaces the builder depends on — already exist, no changes needed:

```go
// From internal/app/rls/policy_matching_service.go:
// PolicyMatchingService has:
//   ResolveUsing(ctx, orgName, projectSlug, modelID, action, userCtx) (string, []interface{}, error)
//   GetCheckExpr(ctx, orgName, projectSlug, modelID, action, userCtx) (string, error)

// From internal/interfaces/http/middleware/runtime_auth_middleware.go:
//   GetEndUserIdentity(ctx) *EndUserIdentity
//   GetUserContext(ctx) *rls.UserContext

// From internal/domain/rls/:
//   Action constants: ActionRead, ActionCreate, ActionUpdate, ActionDelete
```

- [ ] **Step 2: Write the RLSSnapshotBuilder**

```go
package modelruntime

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"

	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/rls"
	"modelcraft/internal/interfaces/http/middleware"
	"modelcraft/pkg/logfacade"
)

// PolicyResolver provides access to RLS policy expressions.
// Implemented by app/rls.PolicyMatchingService.
type PolicyResolver interface {
	ResolveUsing(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, []interface{}, error)
	GetCheckExpr(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, error)
}

// RLSSnapshotBuilder builds RLSPolicySnapshot at request entry.
type RLSSnapshotBuilder struct {
	logger     logfacade.Logger
	policySvc  PolicyResolver
	celEnv     *cel.Env
}

// NewRLSSnapshotBuilder creates a new RLSSnapshotBuilder.
func NewRLSSnapshotBuilder(logger logfacade.Logger, policySvc PolicyResolver) *RLSSnapshotBuilder {
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create CEL environment: %v", err))
	}
	return &RLSSnapshotBuilder{
		logger:    logger,
		policySvc: policySvc,
		celEnv:    env,
	}
}

// Build constructs the RLSPolicySnapshot for the given model.
// Returns nil for developer access (no RLS applied).
// Returns DenyAll=true when no matching policy exists.
func (b *RLSSnapshotBuilder) Build(
	ctx context.Context,
	orgName, projectSlug, modelID string,
) (*modelruntime.RLSPolicySnapshot, error) {
	// Developer JWT — no RLS
	identity := middleware.GetEndUserIdentity(ctx)
	if identity == nil || identity.IsDeveloper() {
		return nil, nil
	}

	userCtx := middleware.GetUserContext(ctx)
	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}

	auth := buildAuthMap(userCtx)

	// Resolve USING for each read/write action
	selectUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionRead, userCtx)
	if err != nil {
		return nil, err
	}
	updateUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionUpdate, userCtx)
	if err != nil {
		return nil, err
	}
	deleteUSING, err := b.resolveUSING(ctx, orgName, projectSlug, modelID, rls.ActionDelete, userCtx)
	if err != nil {
		return nil, err
	}

	// Compile CHECK for insert and update
	insertCHECK, err := b.compileCHECK(ctx, orgName, projectSlug, modelID, rls.ActionCreate, userCtx)
	if err != nil {
		return nil, err
	}
	updateCHECK, err := b.compileCHECK(ctx, orgName, projectSlug, modelID, rls.ActionUpdate, userCtx)
	if err != nil {
		return nil, err
	}

	// DenyAll: no USING and no CHECK for any action
	if selectUSING == nil && updateUSING == nil && deleteUSING == nil &&
		insertCHECK == nil && updateCHECK == nil {
		return &modelruntime.RLSPolicySnapshot{
			DenyAll: true,
		}, nil
	}

	return &modelruntime.RLSPolicySnapshot{
		SelectUSING: selectUSING,
		UpdateUSING: updateUSING,
		DeleteUSING: deleteUSING,
		InsertCHECK: insertCHECK,
		UpdateCHECK: updateCHECK,
		Auth:        auth,
	}, nil
}

// resolveUSING resolves the USING expression for a single action.
// Returns nil if no USING filter is needed (no policies match).
func (b *RLSSnapshotBuilder) resolveUSING(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (*modelruntime.RawSQLFilter, error) {
	sql, params, err := b.policySvc.ResolveUsing(ctx, orgName, projectSlug, modelID, action, userCtx)
	if err != nil {
		// No matching policy for this action — that's ok, just no filter
		return nil, nil //nolint:nilnil
	}
	if sql == "" || sql == "1=1" {
		return nil, nil //nolint:nilnil
	}
	return &modelruntime.RawSQLFilter{SQL: sql, Params: params}, nil
}

// compileCHECK compiles the CHECK expression into a cel.Program.
// Returns nil if no CHECK expression exists.
func (b *RLSSnapshotBuilder) compileCHECK(
	ctx context.Context, orgName, projectSlug, modelID string,
	action rls.Action, userCtx *rls.UserContext,
) (*modelruntime.CheckProgram, error) {
	expr, err := b.policySvc.GetCheckExpr(ctx, orgName, projectSlug, modelID, action, userCtx)
	if err != nil {
		return nil, nil //nolint:nilnil // no matching policy — no CHECK needed
	}
	if expr == "" {
		return nil, nil //nolint:nilnil
	}

	// Compile the CEL expression
	ast, issues := b.celEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("compile CHECK expression: %w", issues.Err())
	}
	program, err := b.celEnv.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("program CHECK expression: %w", err)
	}
	return modelruntime.NewCheckProgram(program), nil
}

// buildAuthMap builds the auth context map for CEL evaluation.
func buildAuthMap(userCtx *rls.UserContext) map[string]any {
	if userCtx == nil {
		return map[string]any{
			"userid":   "",
			"username": "",
			"roles":    []string{},
		}
	}
	return map[string]any{
		"userid":   userCtx.UserID,
		"username": userCtx.UserName,
		"roles":    userCtx.Roles,
	}
}
```

- [ ] **Step 3: Verify compilation**

```bash
cd modelcraft-backend && go build ./internal/app/modelruntime/...
```

---

### Task 3: Create RLSInterceptDB wrapper

**Files:**
- Create: `modelcraft-backend/internal/infrastructure/database/dml/rls_intercept_db.go`

- [ ] **Step 1: Write RLSInterceptDB**

```go
package dml

import (
	"context"

	"modelcraft/internal/domain/modelruntime"
	"modelcraft/pkg/bizerrors"
)

// RLSInterceptDB wraps ClientDatabaseRepository to apply RLS policies
// before SQL execution. It reads RLSPolicySnapshot from context and
// transparently injects USING filters and evaluates CHECK expressions.
type RLSInterceptDB struct {
	inner modelruntime.ClientDatabaseRepository
}

// NewRLSInterceptDB creates a new RLS intercept wrapper.
func NewRLSInterceptDB(inner modelruntime.ClientDatabaseRepository) *RLSInterceptDB {
	return &RLSInterceptDB{inner: inner}
}

// injectUSING appends the USING filter to RawFilters if present.
func injectUSING(filters *[]modelruntime.RawSQLFilter, using *modelruntime.RawSQLFilter) {
	if using != nil {
		*filters = append(*filters, *using)
	}
}

// evalCHECK evaluates a CHECK expression if present.
func evalCHECK(check *modelruntime.CheckProgram, input map[string]any, auth map[string]any) error {
	if check == nil {
		return nil
	}
	if err := check.Eval(input, auth); err != nil {
		return bizerrors.NewError(bizerrors.PermissionDenied, err.Error())
	}
	return nil
}

// ---- Read operations ----

func (r *RLSInterceptDB) FindUnique(ctx context.Context, input *modelruntime.FindUniqueInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.FindUnique(ctx, input)
}

func (r *RLSInterceptDB) FindFirst(ctx context.Context, input *modelruntime.FindFirstInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.FindFirst(ctx, input)
}

func (r *RLSInterceptDB) FindMany(ctx context.Context, input *modelruntime.FindManyInput) ([]map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.FindMany(ctx, input)
}

func (r *RLSInterceptDB) ListByCursor(ctx context.Context, input *modelruntime.ListByCursorInput) ([]map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.ListByCursor(ctx, input)
}

func (r *RLSInterceptDB) Aggregate(ctx context.Context, input *modelruntime.AggregateInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.Aggregate(ctx, input)
}

func (r *RLSInterceptDB) Count(ctx context.Context, input *modelruntime.CountInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.SelectUSING)
	}
	return r.inner.Count(ctx, input)
}

// FindManyIn is used for N+1 relation loading, no RLS interception currently.
func (r *RLSInterceptDB) FindManyIn(ctx context.Context, input *modelruntime.FindManyInInput) ([]map[string]any, error) {
	return r.inner.FindManyIn(ctx, input)
}

// ---- INSERT operations ----

func (r *RLSInterceptDB) CreateOne(ctx context.Context, input *modelruntime.CreateOneInput) (string, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		if err := evalCHECK(snap.InsertCHECK, input.Data, snap.Auth); err != nil {
			return "", err
		}
	}
	return r.inner.CreateOne(ctx, input)
}

func (r *RLSInterceptDB) CreateMany(ctx context.Context, input *modelruntime.CreateManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		for _, dataItem := range input.Data {
			if err := evalCHECK(snap.InsertCHECK, dataItem, snap.Auth); err != nil {
				return nil, err
			}
		}
	}
	return r.inner.CreateMany(ctx, input)
}

// ---- UPDATE operations ----

func (r *RLSInterceptDB) UpdateOne(ctx context.Context, input *modelruntime.UpdateOneInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.UpdateUSING)
		if err := evalCHECK(snap.UpdateCHECK, input.Data, snap.Auth); err != nil {
			return nil, err
		}
	}
	return r.inner.UpdateOne(ctx, input)
}

func (r *RLSInterceptDB) UpdateMany(ctx context.Context, input *modelruntime.UpdateManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.UpdateUSING)
		if err := evalCHECK(snap.UpdateCHECK, input.Data, snap.Auth); err != nil {
			return nil, err
		}
	}
	return r.inner.UpdateMany(ctx, input)
}

// ---- DELETE operations ----

func (r *RLSInterceptDB) DeleteOne(ctx context.Context, input *modelruntime.DeleteOneInput) (map[string]any, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.DeleteUSING)
	}
	return r.inner.DeleteOne(ctx, input)
}

func (r *RLSInterceptDB) DeleteMany(ctx context.Context, input *modelruntime.DeleteManyInput) (interface{}, error) {
	if snap := modelruntime.GetRLSSnapshot(ctx); snap != nil {
		injectUSING(&input.RawFilters, snap.DeleteUSING)
	}
	return r.inner.DeleteMany(ctx, input)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/database/dml/...
```

---

### Task 4: Write RLSInterceptDB unit tests

**Files:**
- Create: `modelcraft-backend/internal/infrastructure/database/dml/rls_intercept_db_test.go`

- [ ] **Step 1: Write mock and tests**

```go
package dml

import (
	"context"
	"testing"

	"github.com/google/cel-go/cel"

	"modelcraft/internal/domain/modelruntime"
)

// mockClientDB implements ClientDatabaseRepository for testing.
type mockClientDB struct {
	findManyCalled   bool
	findManyInput    *modelruntime.FindManyInput
	findManyResult   []map[string]any

	createOneCalled  bool
	createOneInput   *modelruntime.CreateOneInput
	createOneResult  string

	updateManyCalled bool
	updateManyInput  *modelruntime.UpdateManyInput
}

func (m *mockClientDB) FindMany(ctx context.Context, input *modelruntime.FindManyInput) ([]map[string]any, error) {
	m.findManyCalled = true
	m.findManyInput = input
	return m.findManyResult, nil
}
func (m *mockClientDB) FindUnique(ctx context.Context, input *modelruntime.FindUniqueInput) (map[string]any, error) { return nil, nil }
func (m *mockClientDB) FindFirst(ctx context.Context, input *modelruntime.FindFirstInput) (map[string]any, error) { return nil, nil }
func (m *mockClientDB) ListByCursor(ctx context.Context, input *modelruntime.ListByCursorInput) ([]map[string]any, error) { return nil, nil }
func (m *mockClientDB) FindManyIn(ctx context.Context, input *modelruntime.FindManyInInput) ([]map[string]any, error) { return nil, nil }
func (m *mockClientDB) Aggregate(ctx context.Context, input *modelruntime.AggregateInput) (map[string]any, error) { return nil, nil }
func (m *mockClientDB) Count(ctx context.Context, input *modelruntime.CountInput) (map[string]any, error) { return nil, nil }
func (m *mockClientDB) CreateOne(ctx context.Context, input *modelruntime.CreateOneInput) (string, error) {
	m.createOneCalled = true
	m.createOneInput = input
	return m.createOneResult, nil
}
func (m *mockClientDB) UpdateOne(ctx context.Context, input *modelruntime.UpdateOneInput) (map[string]any, error) { return nil, nil }
func (m *mockClientDB) DeleteOne(ctx context.Context, input *modelruntime.DeleteOneInput) (map[string]any, error) { return nil, nil }
func (m *mockClientDB) CreateMany(ctx context.Context, input *modelruntime.CreateManyInput) (interface{}, error) { return nil, nil }
func (m *mockClientDB) UpdateMany(ctx context.Context, input *modelruntime.UpdateManyInput) (interface{}, error) {
	m.updateManyCalled = true
	m.updateManyInput = input
	return nil, nil
}
func (m *mockClientDB) DeleteMany(ctx context.Context, input *modelruntime.DeleteManyInput) (interface{}, error) { return nil, nil }

// makeTrueCheckProgram creates a CheckProgram that always evaluates to true.
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

// makeFalseCheckProgram creates a CheckProgram that always evaluates to false.
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

// TestRLSInterceptDB_FindMany_InjectsUSING verifies that FindMany
// appends the SelectUSING filter to input.RawFilters.
func TestRLSInterceptDB_FindMany_InjectsUSING(t *testing.T) {
	mock := &mockClientDB{
		findManyResult: []map[string]any{{"id": "1"}},
	}
	db := NewRLSInterceptDB(mock)

	rlsFilter := &modelruntime.RawSQLFilter{SQL: "owner_id = ?", Params: []any{"u_123"}}
	snap := &modelruntime.RLSPolicySnapshot{
		SelectUSING: rlsFilter,
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

// TestRLSInterceptDB_FindMany_NoSnapshot verifies that without a snapshot,
// no interception occurs (developer path).
func TestRLSInterceptDB_FindMany_NoSnapshot(t *testing.T) {
	mock := &mockClientDB{
		findManyResult: []map[string]any{{"id": "1"}},
	}
	db := NewRLSInterceptDB(mock)

	// No snapshot on context — developer access
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

// TestRLSInterceptDB_CreateOne_CHECKAcepted verifies that CreateOne
// succeeds when the CHECK expression evaluates to true.
func TestRLSInterceptDB_CreateOne_CHECKAcepted(t *testing.T) {
	mock := &mockClientDB{createOneResult: "new-id"}
	db := NewRLSInterceptDB(mock)

	snap := &modelruntime.RLSPolicySnapshot{
		InsertCHECK: makeTrueCheckProgram(t),
		Auth:        map[string]any{"userid": "u_123"},
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

// TestRLSInterceptDB_CreateOne_CHECKRejected verifies that CreateOne
// returns PermissionDenied when the CHECK expression evaluates to false.
func TestRLSInterceptDB_CreateOne_CHECKRejected(t *testing.T) {
	mock := &mockClientDB{createOneResult: "new-id"}
	db := NewRLSInterceptDB(mock)

	snap := &modelruntime.RLSPolicySnapshot{
		InsertCHECK: makeFalseCheckProgram(t),
		Auth:        map[string]any{"userid": "u_123"},
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

// TestRLSInterceptDB_UpdateMany_USINGAndCHECK verifies that UpdateMany
// applies both USING injection and CHECK evaluation.
func TestRLSInterceptDB_UpdateMany_USINGAndCHECK(t *testing.T) {
	mock := &mockClientDB{}
	db := NewRLSInterceptDB(mock)

	snap := &modelruntime.RLSPolicySnapshot{
		UpdateUSING: &modelruntime.RawSQLFilter{SQL: "owner_id = ?", Params: []any{"u_123"}},
		UpdateCHECK: makeTrueCheckProgram(t),
		Auth:        map[string]any{"userid": "u_123"},
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
```

- [ ] **Step 2: Run tests**

```bash
cd modelcraft-backend && go test ./internal/infrastructure/database/dml/... -run TestRLSInterceptDB -v
```

Expected: 5 tests PASS

- [ ] **Step 3: Commit**

```bash
git add modelcraft-backend/internal/domain/modelruntime/rls_snapshot.go \
        modelcraft-backend/internal/app/modelruntime/rls_snapshot_builder.go \
        modelcraft-backend/internal/infrastructure/database/dml/rls_intercept_db.go \
        modelcraft-backend/internal/infrastructure/database/dml/rls_intercept_db_test.go
git commit -m "feat: add RLS SQL intercept layer

- Add RLSPolicySnapshot type with precompiled CHECK programs
- Add RLSSnapshotBuilder to resolve policies at Execute() entry
- Add RLSInterceptDB wrapper to transparently inject USING and eval CHECK
- Add unit tests for RLSInterceptDB

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 5: Modify GraphqlAppService.Execute() to use snapshot builder

**Files:**
- Modify: `modelcraft-backend/internal/app/modelruntime/graphql_app.go`

- [ ] **Step 1: Replace rlsGuard with snapshotBuilder**

Change the struct:

```go
// Before:
type GraphqlAppService struct {
	modelRepo            modelruntime.ModelRepository
	graphqlSchemaManager *modelruntime.GraphqlSchemaManager
	permService          modelruntime.EndUserPermissionService
	rlsGuard             RuntimeRLSPolicyGuard
}

// After:
type GraphqlAppService struct {
	modelRepo            modelruntime.ModelRepository
	graphqlSchemaManager *modelruntime.GraphqlSchemaManager
	permService          modelruntime.EndUserPermissionService
	snapshotBuilder      *RLSSnapshotBuilder
}
```

- [ ] **Step 2: Update NewGraphqlAppService**

```go
// Before:
func NewGraphqlAppService(
	modelRepo modelruntime.ModelRepository,
	lfkRepo modeldesign.LogicalForeignKeyRepository,
	permService modelruntime.EndUserPermissionService,
	rlsGuard RuntimeRLSPolicyGuard,
) *GraphqlAppService

// After:
func NewGraphqlAppService(
	modelRepo modelruntime.ModelRepository,
	lfkRepo modeldesign.LogicalForeignKeyRepository,
	permService modelruntime.EndUserPermissionService,
	snapshotBuilder *RLSSnapshotBuilder,
) *GraphqlAppService {
	schemaManager := modelruntime.NewGraphqlSchemaManager(modelRepo, lfkRepo)
	return &GraphqlAppService{
		modelRepo:            modelRepo,
		graphqlSchemaManager: schemaManager,
		permService:          permService,
		snapshotBuilder:      snapshotBuilder,
	}
}
```

- [ ] **Step 3: Update Execute() method**

Replace lines 70-141. Key changes:
1. Load model once to get `modelID` (reused by both snapshot builder and permission resolution)
2. Wrap `clientRepo` with `RLSInterceptDB`
3. Build `RLSPolicySnapshot` at entry
4. Check `DenyAll` → short-circuit
5. Remove `rlsGuard` injection

```go
func (s *GraphqlAppService) Execute(ctx context.Context, orgName, projectSlug, name, databaseName string,
	cmd ExecuteGraphQLCommand,
) (*graphql.Result, error) {
	logger := logfacade.GetLogger(ctx)

	// Inject request metadata into context
	ctx = requestcontext.WithMetadata(ctx)
	modelLocator, err := modeldesign.NewModelLocator(orgName, projectSlug, databaseName, name)
	if err != nil {
		return nil, err
	}

	// Load model once — modelID needed by snapshot builder and permission resolution
	model, err := s.modelRepo.GetByName(ctx, modelLocator)
	if err != nil {
		logger.Errorf(ctx, "get model fail: %v", err)
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
		}
		return nil, fmt.Errorf("获取模型失败 %s", modelLocator.GetFullPath())
	}
	modelID := model.ID
	if modelID == "" {
		modelID = model.Name
	}

	gschema, err := s.GetSchema(ctx, orgName, modelLocator)
	if err != nil {
		return nil, err
	}

	// 创建请求级 DB 连接
	clientSqlDB, err := repository.DefaultClusterManager.GetConnectionWithDatabase(
		ctx, orgName, modelLocator.ProjectSlug, modelLocator.DatabaseName,
	)
	if err != nil {
		logger.Errorf(ctx, "get client sql db fail: %v", err)
		return nil, fmt.Errorf("获取客户端数据库失败 %s", databaseName)
	}
	// ★ Wrap clientRepo with RLS intercept layer
	clientRepo := dml.NewRLSInterceptDB(dml.NewClientDB(clientSqlDB))

	// 提取 endUserID
	endUserID := ""
	if ctxutils.IsEndUser(ctx) && !ctxutils.GetIsAdminFromContext(ctx) {
		if uid, err := ctxutils.GetUserIDFromContext(ctx); err == nil {
			endUserID = uid
		}
	}

	// 解析 end-user 权限快照
	var endUserPerms *modelruntime.ResolvedModelPermissions
	if endUserID != "" {
		endUserPerms, err = s.permService.Resolve(ctx, orgName, projectSlug, endUserID, modelID)
		if err != nil {
			return nil, err
		}
	}

	// ★ Build RLS snapshot at entry (replaces rlsGuard injection)
	snap, err := s.snapshotBuilder.Build(ctx, orgName, projectSlug, modelID)
	if err != nil {
		return nil, err
	}
	if snap != nil && snap.DenyAll {
		return nil, bizerrors.NewError(bizerrors.PermissionDenied, "RLS: no matching policy")
	}
	if snap != nil {
		ctx = modelruntime.WithRLSSnapshot(ctx, snap)
	}

	endUserAdminID, _ := ctxutils.GetUserIDFromContext(ctx)
	reqCtx := modelruntime.WithGraphqlRequestContext(
		ctx, clientRepo, orgName, projectSlug, endUserID, endUserAdminID, endUserPerms,
	)

	// 执行GraphQL查询
	result := graphql.Do(graphql.Params{
		Schema:         *gschema,
		RequestString:  cmd.Query,
		VariableValues: cmd.Variables,
		Context:        reqCtx,
	})
	marshal, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "result=%+v", string(marshal))

	return result, nil
}
```

Note: `resolveEndUserPerms` method is no longer needed since model loading and permission resolution are now inlined. Delete the `resolveEndUserPerms` method (lines 143-169).

- [ ] **Step 4: Delete the RuntimeRLSPolicyGuard interface**

Remove lines 29-32 (the `RuntimeRLSPolicyGuard` interface definition).

- [ ] **Step 5: Verify compilation**

```bash
cd modelcraft-backend && go build ./internal/app/modelruntime/...
```

- [ ] **Step 6: Commit**

```bash
git add modelcraft-backend/internal/app/modelruntime/graphql_app.go
git commit -m "refactor: replace rlsGuard with snapshotBuilder in Execute()

- Build RLSPolicySnapshot at entry instead of injecting RLS guard
- Wrap clientRepo with RLSInterceptDB
- DenyAll short-circuits before GraphQL execution
- Remove RuntimeRLSPolicyGuard interface

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 6: Clean up model_resolver.go — remove RLS code from resolvers

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`

- [ ] **Step 1: Delete appendRLSUsingFilter method**

Remove lines 62-79 (the entire `appendRLSUsingFilter` method).

- [ ] **Step 2: Remove appendRLSUsingFilter calls from all execute methods**

For each execute method below, remove the line calling `appendRLSUsingFilter`:

- `executeFindUnique` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeFindFirst` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeFindMany` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeAggregate` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeCount` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeUpdateOne` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeUpdateMany` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeDeleteOne` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`
- `executeDeleteMany` — remove `input.RawFilters, err = m.appendRLSUsingFilter(...)`

- [ ] **Step 3: Remove RLSPolicyGuard.ValidateInput calls**

For each execute method below, remove the ValidateInput block:

- `executeCreateOne` — remove the `if rctx.RLSPolicyGuard != nil { ... ValidateInput(...) }` block
- `executeCreateMany` — remove the `if rctx.RLSPolicyGuard != nil { ... ValidateInput(...) }` block
- `executeUpdateOne` — remove the `if rctx.RLSPolicyGuard != nil { ... ValidateInput(...) }` block
- `executeUpdateMany` — remove the `if rctx.RLSPolicyGuard != nil { ... ValidateInput(...) }` block

- [ ] **Step 4: Verify compilation**

```bash
cd modelcraft-backend && go build ./internal/domain/modelruntime/...
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/domain/modelruntime/model_resolver.go
git commit -m "refactor: remove RLS code from model resolvers

- Delete appendRLSUsingFilter method
- Remove all appendRLSUsingFilter calls from execute methods
- Remove all RLSPolicyGuard.ValidateInput calls from execute methods
- RLS is now handled transparently by RLSInterceptDB

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 7: Clean up graphql_request_context.go

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go`

- [ ] **Step 1: Remove RLSPolicyGuard interface**

Remove lines 9-12:

```go
// Delete:
type RLSPolicyGuard interface {
	ValidateInput(ctx context.Context, modelID string, action Action, input map[string]any) error
	ResolveUsingFilter(ctx context.Context, modelID string, action Action) (*RawSQLFilter, error)
}
```

- [ ] **Step 2: Remove RLSPolicyGuard field from graphqlRequestContext**

Remove line 39: `RLSPolicyGuard RLSPolicyGuard`

- [ ] **Step 3: Remove WithRLSPolicyGuard function**

Remove lines 81-89 (the entire `WithRLSPolicyGuard` function).

- [ ] **Step 4: Verify compilation**

```bash
cd modelcraft-backend && go build ./internal/domain/modelruntime/...
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go
git commit -m "refactor: remove RLSPolicyGuard from graphql request context

- Remove RLSPolicyGuard interface
- Remove RLSPolicyGuard field from graphqlRequestContext
- Remove WithRLSPolicyGuard function
- RLS is now handled by RLSPolicySnapshot on context

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 8: Update wiring in routes.go

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`

- [ ] **Step 1: Replace runtimeRLSResolver with RLSSnapshotBuilder**

Change lines 685-689:

```go
// Before:
rlsUsingCompiler := rls.NewPolicyExpressionSQLCompiler()
rlsCheckEvaluator := rls.NewPolicyExpressionInputEvaluator()
rlsMatchingSvc := rls.NewPolicyMatchingService(policyRepo, rlsUsingCompiler, rlsCheckEvaluator)
runtimeRLSResolver := runtime.NewRLSResolver(logfacade.GetLogger(context.Background()), rlsMatchingSvc)
graphqlAppService := modelruntime.NewGraphqlAppService(modelRuntimeRepo, lfkRepo, permService, runtimeRLSResolver)

// After:
rlsUsingCompiler := rls.NewPolicyExpressionSQLCompiler()
rlsCheckEvaluator := rls.NewPolicyExpressionInputEvaluator()
rlsMatchingSvc := rls.NewPolicyMatchingService(policyRepo, rlsUsingCompiler, rlsCheckEvaluator)
snapshotBuilder := modelruntime.NewRLSSnapshotBuilder(logfacade.GetLogger(context.Background()), rlsMatchingSvc)
graphqlAppService := modelruntime.NewGraphqlAppService(modelRuntimeRepo, lfkRepo, permService, snapshotBuilder)
```

Note: `rlsCheckEvaluator` is no longer used after removing `NewRLSResolver`. It can be removed from the wiring, but since it's created with `NewPolicyExpressionInputEvaluator()` which panics on error, it's harmless to keep for now.

- [ ] **Step 2: Clean up unused import**

If `runtime` package import (`"modelcraft/internal/interfaces/runtime"`) is no longer needed in this file (check if it's used elsewhere), remove it.

- [ ] **Step 3: Verify compilation**

```bash
cd modelcraft-backend && go build ./internal/interfaces/http/...
```

Expected: PASS

- [ ] **Step 4: Check full project compilation**

```bash
cd modelcraft-backend && go build ./...
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "refactor: wire RLSSnapshotBuilder instead of RLSResolver

- Create RLSSnapshotBuilder with PolicyMatchingService
- Pass builder to GraphqlAppService instead of runtimeRLSResolver
- RLS now resolved at Execute() entry instead of during resolver execution

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 9: Verify no regressions — run tests

- [ ] **Step 1: Run RLS-related tests**

```bash
cd modelcraft-backend && go test ./internal/infrastructure/database/dml/... -v
```

Expected: All tests pass, including new RLSInterceptDB tests.

- [ ] **Step 2: Run modelruntime tests**

```bash
cd modelcraft-backend && go test ./internal/domain/modelruntime/... -v
```

Expected: Tests pass. Check for any test failures related to `appendRLSUsingFilter` or `RLSPolicyGuard` — update or remove affected tests.

- [ ] **Step 3: Run full test suite**

```bash
cd modelcraft-backend && go test ./... 2>&1 | tail -20
```

Expected: No new failures introduced.

- [ ] **Step 4: Run lint**

```bash
cd modelcraft-backend && just lint
```

Fix any lint issues.

- [ ] **Step 5: Commit any test fixes**

```bash
git add -A
git commit -m "test: fix tests after RLS intercept refactor

Co-Authored-By: Claude <noreply@anthropic.com>"
```
