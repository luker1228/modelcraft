# RLS CEL Policy Expression Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace user-authored RLS JSON predicates with CEL policy expressions where `using` compiles to MySQL row filters and `check` evaluates write input before create/update.

**Architecture:** Add a focused CEL policy-expression package under `modelcraft-backend/internal/app/rls` that parses, validates, compiles, and evaluates the supported CEL subset. Keep persisted GraphQL/DB field names `usingExpr` and `withCheckExpr` for compatibility, route legacy JSON expressions through the existing compiler, and wire new CEL behavior into RLS policy validation, dry-run, matching, and model runtime writes. Update the frontend RLS drawer to present CEL as `Using Filter` and `Input Check`.

**Tech Stack:** Go 1.25, `github.com/google/cel-go`, gqlgen GraphQL, goqu/MySQL DML mapping, Next.js/React/TypeScript, Vitest, ESLint.

---

## Scope Check

This plan implements one coherent feature: CEL-based RLS policy expressions. It touches backend expression semantics, runtime enforcement, GraphQL dry-run payloads, and the frontend editor because those pieces must change together for a working feature. It does not rename database or GraphQL fields; `withCheckExpr` remains the storage/API field for phase one.

## File Structure

Backend files:

- Create `modelcraft-backend/internal/app/rls/policy_expression_types.go` for expression modes, dry-run DTOs, and shared helpers.
- Create `modelcraft-backend/internal/app/rls/policy_expression_validator.go` for CEL parse/check and supported-subset validation.
- Create `modelcraft-backend/internal/app/rls/policy_expression_sql_compiler.go` for `using` CEL AST to MySQL `WHERE` SQL plus params.
- Create `modelcraft-backend/internal/app/rls/policy_expression_input_evaluator.go` for `check(input, auth)` evaluation.
- Create tests beside each file in `modelcraft-backend/internal/app/rls`.
- Modify `modelcraft-backend/internal/app/rls/policy_matching_service.go` so `ResolveUsing` compiles CEL using filters and legacy JSON filters.
- Modify `modelcraft-backend/internal/interfaces/runtime/rls_resolver.go` so create/update check validates input maps instead of only checking for `1=0`.
- Modify `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go` to carry an optional RLS guard.
- Modify `modelcraft-backend/internal/app/modelruntime/graphql_app.go` and `modelcraft-backend/internal/domain/modelruntime/model_resolver.go` to apply input checks before create/update execution.
- Modify `modelcraft-backend/internal/infrastructure/database/dml/sql_mapper.go` and `modelcraft-backend/internal/domain/modelruntime/graphql_input.go` to support raw RLS SQL filters for `using`.
- Modify `modelcraft-backend/internal/interfaces/graphql/project/rls.resolvers.go` and `modelcraft-backend/api/graph/project/schema/rls.graphql` to return dry-run details.
- Regenerate GraphQL code with `go run github.com/99designs/gqlgen`.

Frontend files:

- Modify `modelcraft-front/src/api-client/rls-policy/graphql-docs.ts` for dry-run result fields.
- Modify `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyManage.ts` to display dry-run details.
- Modify `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsExpressionEditor.tsx` for CEL labels, validation state, and dry-run result display.
- Modify `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/PolicyEditorDialog.tsx` for labels and placeholders.
- Modify `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.ts` and its test for CEL syntax expectations.

---

### Task 1: Add CEL Dependency And Expression Types

**Files:**
- Modify: `modelcraft-backend/go.mod`
- Modify: `modelcraft-backend/go.sum`
- Create: `modelcraft-backend/internal/app/rls/policy_expression_types.go`
- Test: `modelcraft-backend/internal/app/rls/policy_expression_types_test.go`

- [ ] **Step 1: Add the CEL dependency**

Run:

```bash
go get github.com/google/cel-go@latest
```

Expected: `go.mod` contains `github.com/google/cel-go` and `go.sum` is updated.

- [ ] **Step 2: Write the failing helper tests**

Create `modelcraft-backend/internal/app/rls/policy_expression_types_test.go`:

```go
package rls

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolicyExpressionModeAllowedRoot(t *testing.T) {
	require.True(t, PolicyExpressionModeUsing.AllowsRoot("row"))
	require.True(t, PolicyExpressionModeUsing.AllowsRoot("auth"))
	require.False(t, PolicyExpressionModeUsing.AllowsRoot("input"))

	require.True(t, PolicyExpressionModeCheck.AllowsRoot("input"))
	require.True(t, PolicyExpressionModeCheck.AllowsRoot("auth"))
	require.False(t, PolicyExpressionModeCheck.AllowsRoot("row"))
}

func TestIsLegacyJSONExpression(t *testing.T) {
	require.True(t, IsLegacyJSONExpression(`{"owner_id":{"equals":"{{user_id}}"}}`))
	require.True(t, IsLegacyJSONExpression(`true`))
	require.False(t, IsLegacyJSONExpression(`row.owner_id == auth.user_id`))
	require.False(t, IsLegacyJSONExpression(` input.owner_id == auth.user_id `))
}
```

- [ ] **Step 3: Run the helper tests and verify they fail**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionModeAllowedRoot|TestIsLegacyJSONExpression' -count=1
```

Expected: FAIL with undefined `PolicyExpressionModeUsing` or `IsLegacyJSONExpression`.

- [ ] **Step 4: Add expression type helpers**

Create `modelcraft-backend/internal/app/rls/policy_expression_types.go`:

```go
package rls

import "strings"

type PolicyExpressionMode string

const (
	PolicyExpressionModeUsing PolicyExpressionMode = "using"
	PolicyExpressionModeCheck PolicyExpressionMode = "check"
)

func (m PolicyExpressionMode) AllowsRoot(root string) bool {
	switch m {
	case PolicyExpressionModeUsing:
		return root == "row" || root == "auth"
	case PolicyExpressionModeCheck:
		return root == "input" || root == "auth"
	default:
		return false
	}
}

type PolicyExpressionDryRunResult struct {
	Valid      bool
	SQL        string
	Params     []any
	Result     *bool
	Errors     []PolicyExpressionError
}

type PolicyExpressionError struct {
	Path    string
	Message string
	Code    string
}

func IsLegacyJSONExpression(expr string) bool {
	trimmed := strings.TrimSpace(expr)
	return trimmed == "true" || trimmed == "false" || strings.HasPrefix(trimmed, "{")
}
```

- [ ] **Step 5: Run tests and commit**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionModeAllowedRoot|TestIsLegacyJSONExpression' -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/go.mod modelcraft-backend/go.sum modelcraft-backend/internal/app/rls/policy_expression_types.go modelcraft-backend/internal/app/rls/policy_expression_types_test.go
git commit -m "feat: add RLS CEL expression types"
```

---

### Task 2: Validate CEL Expressions And Enforce The Supported Subset

**Files:**
- Create: `modelcraft-backend/internal/app/rls/policy_expression_validator.go`
- Test: `modelcraft-backend/internal/app/rls/policy_expression_validator_test.go`

- [ ] **Step 1: Write failing validator tests**

Create `modelcraft-backend/internal/app/rls/policy_expression_validator_test.go`:

```go
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

func (s testAuthSchema) IsValidRef(name string) bool { return name == "user_id" || s.refs[name] }

func TestPolicyExpressionValidator_UsingAcceptsRowAndAuth(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(context.Background(), PolicyExpressionModeUsing,
		`row.owner_id == auth.user_id && row.status in ["draft", "pending"]`,
		testModelSchema{fields: map[string]bool{"owner_id": true, "status": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Empty(t, errs)
}

func TestPolicyExpressionValidator_RejectsWrongRootForMode(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(context.Background(), PolicyExpressionModeUsing,
		`input.owner_id == auth.user_id`,
		testModelSchema{fields: map[string]bool{"owner_id": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Len(t, errs, 1)
	require.Equal(t, "INVALID_CONTEXT", errs[0].Code)
}

func TestPolicyExpressionValidator_RejectsUnknownField(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(context.Background(), PolicyExpressionModeCheck,
		`input.missing == auth.user_id`,
		testModelSchema{fields: map[string]bool{"owner_id": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Len(t, errs, 1)
	require.Equal(t, "UNKNOWN_FIELD", errs[0].Code)
}

func TestPolicyExpressionValidator_RejectsUnsupportedFunctionCall(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.ValidateCEL(context.Background(), PolicyExpressionModeUsing,
		`size(row.title) > 0`,
		testModelSchema{fields: map[string]bool{"title": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Len(t, errs, 1)
	require.Equal(t, "UNSUPPORTED_CALL", errs[0].Code)
}

func TestPolicyExpressionValidator_ValidateRoutesLegacyJSONToOldValidator(t *testing.T) {
	validator := NewPolicyExpressionValidator()
	errs := validator.Validate(context.Background(),
		domainrls.JsonExpr(`{"owner_id":{"_eq":{"_auth":"uid"}}}`),
		domainrls.ExprTypeInsertCheck,
		testModelSchema{fields: map[string]bool{"owner_id": true}},
		testAuthSchema{refs: map[string]bool{}},
	)
	require.Empty(t, errs)
}
```

- [ ] **Step 2: Run validator tests and verify they fail**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionValidator' -count=1
```

Expected: FAIL with undefined `NewPolicyExpressionValidator`.

- [ ] **Step 3: Implement the CEL validator wrapper**

Create `modelcraft-backend/internal/app/rls/policy_expression_validator.go`:

```go
package rls

import (
	"context"
	"fmt"
	domainrls "modelcraft/internal/domain/rls"
	"strings"

	"github.com/google/cel-go/cel"
)

type PolicyExpressionValidator struct {
	legacy *PolicyValidator
}

func NewPolicyExpressionValidator() *PolicyExpressionValidator {
	return &PolicyExpressionValidator{legacy: NewPolicyValidator()}
}

func (v *PolicyExpressionValidator) Validate(
	ctx context.Context,
	expr domainrls.JsonExpr,
	exprType domainrls.ExprType,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []domainrls.ValidationError {
	if IsLegacyJSONExpression(string(expr)) {
		return v.legacy.Validate(ctx, expr, exprType, modelSchema, authSchema)
	}
	mode := PolicyExpressionModeUsing
	if exprType.IsCheck() {
		mode = PolicyExpressionModeCheck
	}
	errs := v.ValidateCEL(ctx, mode, string(expr), modelSchema, authSchema)
	out := make([]domainrls.ValidationError, 0, len(errs))
	for _, e := range errs {
		out = append(out, domainrls.ValidationError{Path: e.Path, Message: e.Message, Code: e.Code})
	}
	return out
}

func (v *PolicyExpressionValidator) ValidateCEL(
	_ context.Context,
	mode PolicyExpressionMode,
	expr string,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []PolicyExpressionError {
	env, err := cel.NewEnv(
		cel.Variable("row", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return []PolicyExpressionError{{Message: err.Error(), Code: "CEL_ENV_ERROR"}}
	}
	parsed, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return []PolicyExpressionError{{Message: issues.Err().Error(), Code: "SYNTAX_ERROR"}}
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return []PolicyExpressionError{{Message: issues.Err().Error(), Code: "TYPE_ERROR"}}
	}
	if checked.OutputType() != cel.BoolType {
		return []PolicyExpressionError{{Message: "policy expression must return bool", Code: "NON_BOOLEAN_RESULT"}}
	}
	return validatePolicyExpressionSource(mode, expr, modelSchema, authSchema)
}

func validatePolicyExpressionSource(
	mode PolicyExpressionMode,
	expr string,
	modelSchema domainrls.ModelSchema,
	authSchema domainrls.AuthSchemaProvider,
) []PolicyExpressionError {
	var errs []PolicyExpressionError
	unsupportedCalls := []string{"size(", ".all(", ".exists(", ".map(", ".filter("}
	for _, token := range unsupportedCalls {
		if strings.Contains(expr, token) {
			errs = append(errs, PolicyExpressionError{
				Message: fmt.Sprintf("unsupported CEL construct %q", token),
				Code:    "UNSUPPORTED_CALL",
			})
		}
	}
	for _, root := range []string{"row.", "input."} {
		if strings.Contains(expr, root) && !mode.AllowsRoot(strings.TrimSuffix(root, ".")) {
			errs = append(errs, PolicyExpressionError{
				Path:    strings.TrimSuffix(root, "."),
				Message: fmt.Sprintf("%s is not allowed in %s expression", strings.TrimSuffix(root, "."), mode),
				Code:    "INVALID_CONTEXT",
			})
		}
	}
	for _, ref := range extractPolicyRefs(expr, "row.") {
		if !modelSchema.HasField(ref) {
			errs = append(errs, PolicyExpressionError{Path: "row." + ref, Message: "unknown model field " + ref, Code: "UNKNOWN_FIELD"})
		}
	}
	for _, ref := range extractPolicyRefs(expr, "input.") {
		if !modelSchema.HasField(ref) {
			errs = append(errs, PolicyExpressionError{Path: "input." + ref, Message: "unknown model field " + ref, Code: "UNKNOWN_FIELD"})
		}
	}
	for _, ref := range extractPolicyRefs(expr, "auth.") {
		if !authSchema.IsValidRef(ref) {
			errs = append(errs, PolicyExpressionError{Path: "auth." + ref, Message: "unknown auth field " + ref, Code: "UNKNOWN_AUTH_FIELD"})
		}
	}
	return errs
}

func extractPolicyRefs(expr string, prefix string) []string {
	parts := strings.Split(expr, prefix)
	seen := map[string]bool{}
	var refs []string
	for i := 1; i < len(parts); i++ {
		name := readIdentifier(parts[i])
		if name != "" && !seen[name] {
			seen[name] = true
			refs = append(refs, name)
		}
	}
	return refs
}

func readIdentifier(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
			continue
		}
		break
	}
	return b.String()
}
```

- [ ] **Step 4: Wire the new validator into app service construction**

Find RLS app service construction in `modelcraft-backend/internal/interfaces/http/routes.go`. Replace the old validator construction for `RLSPolicyAppService` with:

```go
rlsPolicyValidator := rls.NewPolicyExpressionValidator()
```

Use that variable wherever `NewModelRLSPolicyAppService` or equivalent RLS policy app service constructor currently receives `rls.NewPolicyValidator()`.

- [ ] **Step 5: Run tests and commit**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionValidator' -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/internal/app/rls/policy_expression_validator.go modelcraft-backend/internal/app/rls/policy_expression_validator_test.go modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "feat: validate RLS CEL expressions"
```

---

### Task 3: Compile CEL Using Expressions To MySQL WHERE

**Files:**
- Create: `modelcraft-backend/internal/app/rls/policy_expression_sql_compiler.go`
- Test: `modelcraft-backend/internal/app/rls/policy_expression_sql_compiler_test.go`

- [ ] **Step 1: Write failing SQL compiler tests**

Create `modelcraft-backend/internal/app/rls/policy_expression_sql_compiler_test.go`:

```go
package rls

import (
	"context"
	"testing"

	domainrls "modelcraft/internal/domain/rls"

	"github.com/stretchr/testify/require"
)

func TestPolicyExpressionSQLCompiler_CompilesEqualityAndIn(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	compiled, err := compiler.CompileUsing(context.Background(),
		`row.owner_id == auth.user_id && row.status in ["draft", "pending"]`,
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.NoError(t, err)
	require.Equal(t, "(owner_id = ? AND status IN (?, ?))", compiled.SQL)
	require.Equal(t, []interface{}{"u_123", "draft", "pending"}, compiled.Params)
}

func TestPolicyExpressionSQLCompiler_CompilesOrAndNot(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	compiled, err := compiler.CompileUsing(context.Background(),
		`row.owner_id == auth.user_id || !(row.status == "archived")`,
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.NoError(t, err)
	require.Equal(t, "(owner_id = ? OR NOT (status = ?))", compiled.SQL)
	require.Equal(t, []interface{}{"u_123", "archived"}, compiled.Params)
}

func TestPolicyExpressionSQLCompiler_RejectsInputRoot(t *testing.T) {
	compiler := NewPolicyExpressionSQLCompiler()
	_, err := compiler.CompileUsing(context.Background(),
		`input.owner_id == auth.user_id`,
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.ErrorContains(t, err, "input is not allowed")
}
```

- [ ] **Step 2: Run SQL compiler tests and verify they fail**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionSQLCompiler' -count=1
```

Expected: FAIL with undefined `NewPolicyExpressionSQLCompiler`.

- [ ] **Step 3: Implement minimal supported SQL compiler**

Create `modelcraft-backend/internal/app/rls/policy_expression_sql_compiler.go` with this public surface:

```go
package rls

import (
	"context"
	"fmt"
	domainrls "modelcraft/internal/domain/rls"
	"strings"
)

type PolicyExpressionSQLCompiler struct{}

func NewPolicyExpressionSQLCompiler() *PolicyExpressionSQLCompiler {
	return &PolicyExpressionSQLCompiler{}
}

func (c *PolicyExpressionSQLCompiler) CompileUsing(
	_ context.Context,
	expr string,
	userCtx *domainrls.UserContext,
) (*domainrls.CompiledPolicy, error) {
	if strings.Contains(expr, "input.") {
		return nil, fmt.Errorf("input is not allowed in using expression")
	}
	if IsLegacyJSONExpression(expr) {
		return NewPolicyCompiler().Compile(context.Background(), domainrls.JsonExpr(expr), userCtx)
	}
	compiled, params, err := compileSimpleCELWhere(expr, userCtx)
	if err != nil {
		return nil, err
	}
	return &domainrls.CompiledPolicy{SQL: compiled, Params: params}, nil
}
```

Implement `compileSimpleCELWhere` in the same file. It must support the exact first-phase CEL subset:

```go
func compileSimpleCELWhere(expr string, userCtx *domainrls.UserContext) (string, []interface{}, error) {
	parser := newCELWhereParser(expr, userCtx)
	return parser.parse()
}
```

Add these private parser declarations in the same file:

```go
type celTokenKind string

const (
	celTokenIdentifier celTokenKind = "identifier"
	celTokenString     celTokenKind = "string"
	celTokenNumber     celTokenKind = "number"
	celTokenBool       celTokenKind = "bool"
	celTokenNull       celTokenKind = "null"
	celTokenOperator   celTokenKind = "operator"
	celTokenLParen     celTokenKind = "("
	celTokenRParen     celTokenKind = ")"
	celTokenLBracket   celTokenKind = "["
	celTokenRBracket   celTokenKind = "]"
	celTokenComma      celTokenKind = ","
	celTokenEOF        celTokenKind = "eof"
)

type celToken struct {
	kind  celTokenKind
	value string
}

type celWhereParser struct {
	tokens  []celToken
	pos     int
	userCtx *domainrls.UserContext
}

func newCELWhereParser(expr string, userCtx *domainrls.UserContext) *celWhereParser {
	return &celWhereParser{tokens: tokenizeCELWhere(expr), userCtx: userCtx}
}

func (p *celWhereParser) parse() (string, []interface{}, error) {
	sql, params, err := p.parseOr()
	if err != nil {
		return "", nil, err
	}
	if p.peek().kind != celTokenEOF {
		return "", nil, fmt.Errorf("unexpected token %q", p.peek().value)
	}
	return sql, params, nil
}
```

Implement the private parser methods in the same file with these exact responsibilities:

- `tokenizeCELWhere(expr string) []celToken`: reads identifiers such as `row.owner_id`, string literals, numeric literals, `true`, `false`, `null`, `&&`, `||`, `!`, comparison operators, `in`, parentheses, brackets, and commas.
- `parseOr()`, `parseAnd()`, `parseUnary()`, `parseComparison()`: enforce precedence in this order from lowest to highest: `||`, `&&`, `!`, comparisons and `in`.
- `parseOperand()`: resolves `row.<field>` into a SQL field reference, resolves `auth.user_id` and `auth.user_name` into bound parameter values from `domainrls.UserContext`, and parses string/number/bool/null/array literals into bound values.
- Every literal and auth-derived value must become a `?` parameter. Field names are emitted as raw identifiers only after `PolicyExpressionValidator` has validated the expression; dynamic field names are rejected.
- Boolean groups must be parenthesized, matching test expectations such as `(owner_id = ? AND status IN (?, ?))`.

- [ ] **Step 4: Run SQL compiler tests and commit**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionSQLCompiler' -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/internal/app/rls/policy_expression_sql_compiler.go modelcraft-backend/internal/app/rls/policy_expression_sql_compiler_test.go
git commit -m "feat: compile RLS CEL using filters"
```

---

### Task 4: Evaluate CEL Input Check Expressions

**Files:**
- Create: `modelcraft-backend/internal/app/rls/policy_expression_input_evaluator.go`
- Test: `modelcraft-backend/internal/app/rls/policy_expression_input_evaluator_test.go`

- [ ] **Step 1: Write failing input evaluator tests**

Create `modelcraft-backend/internal/app/rls/policy_expression_input_evaluator_test.go`:

```go
package rls

import (
	"context"
	"testing"

	domainrls "modelcraft/internal/domain/rls"

	"github.com/stretchr/testify/require"
)

func TestPolicyExpressionInputEvaluator_AllowsMatchingCreateInput(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(context.Background(),
		`input.owner_id == auth.user_id && input.status in ["draft", "pending"]`,
		map[string]any{"owner_id": "u_123", "status": "draft"},
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.NoError(t, err)
}

func TestPolicyExpressionInputEvaluator_DeniesMismatchedInput(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(context.Background(),
		`input.owner_id == auth.user_id`,
		map[string]any{"owner_id": "u_999"},
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.ErrorContains(t, err, "RLS CHECK violation")
}

func TestPolicyExpressionInputEvaluator_UpdateUsesPatchOnly(t *testing.T) {
	evaluator := NewPolicyExpressionInputEvaluator()
	err := evaluator.ValidateInput(context.Background(),
		`input.status == "draft"`,
		map[string]any{"title": "renamed"},
		&domainrls.UserContext{UserID: "u_123"},
	)
	require.ErrorContains(t, err, "no such key")
}
```

- [ ] **Step 2: Run input evaluator tests and verify they fail**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionInputEvaluator' -count=1
```

Expected: FAIL with undefined `NewPolicyExpressionInputEvaluator`.

- [ ] **Step 3: Implement input evaluator**

Create `modelcraft-backend/internal/app/rls/policy_expression_input_evaluator.go`:

```go
package rls

import (
	"context"
	"fmt"
	domainrls "modelcraft/internal/domain/rls"

	"github.com/google/cel-go/cel"
)

type PolicyExpressionInputEvaluator struct{}

func NewPolicyExpressionInputEvaluator() *PolicyExpressionInputEvaluator {
	return &PolicyExpressionInputEvaluator{}
}

func (e *PolicyExpressionInputEvaluator) ValidateInput(
	_ context.Context,
	expr string,
	input map[string]any,
	userCtx *domainrls.UserContext,
) error {
	if expr == "" {
		return fmt.Errorf("RLS CHECK violation: empty input check")
	}
	if IsLegacyJSONExpression(expr) {
		return NewPolicyExecutor().ValidateCheck(context.Background(), domainrls.JsonExpr(expr), input, &domainrls.AuthContext{
			EndUserID: userCtx.UserID,
			Variables: map[string]interface{}{
				"user_id":   userCtx.UserID,
				"user_name": userCtx.UserName,
				"roles":     userCtx.Roles,
			},
		})
	}
	env, err := cel.NewEnv(
		cel.Variable("input", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("auth", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return err
	}
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return issues.Err()
	}
	program, err := env.Program(ast)
	if err != nil {
		return err
	}
	out, _, err := program.Eval(map[string]any{
		"input": input,
		"auth": map[string]any{
			"user_id":   userCtx.UserID,
			"user_name": userCtx.UserName,
			"roles":     userCtx.Roles,
		},
	})
	if err != nil {
		return fmt.Errorf("RLS CHECK violation: %w", err)
	}
	allowed, ok := out.Value().(bool)
	if !ok {
		return fmt.Errorf("RLS CHECK violation: expression did not return bool")
	}
	if !allowed {
		return fmt.Errorf("RLS CHECK violation: input rejected")
	}
	return nil
}
```

- [ ] **Step 4: Run input evaluator tests and commit**

Run:

```bash
go test ./internal/app/rls -run 'TestPolicyExpressionInputEvaluator' -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/internal/app/rls/policy_expression_input_evaluator.go modelcraft-backend/internal/app/rls/policy_expression_input_evaluator_test.go
git commit -m "feat: evaluate RLS CEL input checks"
```

---

### Task 5: Wire CEL Compiler And Evaluator Into Policy Matching

**Files:**
- Modify: `modelcraft-backend/internal/app/rls/policy_matching_service.go`
- Modify: `modelcraft-backend/internal/app/rls/policy_matching_service_test.go`
- Modify: `modelcraft-backend/internal/domain/rls/policy_compiler.go`

- [ ] **Step 1: Add failing matching tests for CEL using and check**

Append to `modelcraft-backend/internal/app/rls/policy_matching_service_test.go`:

```go
func TestResolveUsing_WithCELExpression(t *testing.T) {
	ctx := context.Background()
	repo := &mockPolicyRepo{policies: []*rls.Policy{{
		PolicyName: "owner_read",
		Action:     rls.ActionRead,
		Role:       "user",
		UsingExpr:  `row.owner_id == auth.user_id`,
	}}}
	svc := NewPolicyMatchingService(repo, NewPolicyExpressionSQLCompiler(), NewPolicyExpressionInputEvaluator())

	sql, params, err := svc.ResolveUsing(ctx, "my-org", "my-proj", "model-1", rls.ActionRead, &rls.UserContext{
		UserID: "u_123",
		Roles:  []string{"user"},
	})

	require.NoError(t, err)
	require.Equal(t, "(owner_id = ?)", sql)
	require.Equal(t, []interface{}{"u_123"}, params)
}

func TestValidateCheck_WithCELExpression(t *testing.T) {
	ctx := context.Background()
	repo := &mockPolicyRepo{policies: []*rls.Policy{{
		PolicyName:    "owner_create",
		Action:        rls.ActionCreate,
		Role:          "user",
		WithCheckExpr: `input.owner_id == auth.user_id`,
	}}}
	svc := NewPolicyMatchingService(repo, NewPolicyExpressionSQLCompiler(), NewPolicyExpressionInputEvaluator())

	err := svc.ValidateCheck(ctx, "my-org", "my-proj", "model-1", rls.ActionCreate,
		map[string]any{"owner_id": "u_123"},
		&rls.UserContext{UserID: "u_123", Roles: []string{"user"}},
	)

	require.NoError(t, err)
}
```

- [ ] **Step 2: Run matching tests and verify they fail**

Run:

```bash
go test ./internal/app/rls -run 'TestResolveUsing_WithCELExpression|TestValidateCheck_WithCELExpression' -count=1
```

Expected: FAIL because `NewPolicyMatchingService` still takes the old compiler signature and `ValidateCheck` does not exist.

- [ ] **Step 3: Update matching service constructor and methods**

Modify `modelcraft-backend/internal/app/rls/policy_matching_service.go`:

```go
type usingCompiler interface {
	CompileUsing(ctx context.Context, expr string, userCtx *rls.UserContext) (*rls.CompiledPolicy, error)
}

type inputCheckEvaluator interface {
	ValidateInput(ctx context.Context, expr string, input map[string]any, userCtx *rls.UserContext) error
}

type PolicyMatchingService struct {
	repo           PolicyRepository
	usingCompiler  usingCompiler
	checkEvaluator inputCheckEvaluator
}

func NewPolicyMatchingService(repo PolicyRepository, usingCompiler usingCompiler, checkEvaluator inputCheckEvaluator) *PolicyMatchingService {
	return &PolicyMatchingService{repo: repo, usingCompiler: usingCompiler, checkEvaluator: checkEvaluator}
}
```

In `ResolveUsing`, replace `s.compiler.Compile(ctx, expr, userCtx)` with:

```go
compiled, err := s.usingCompiler.CompileUsing(ctx, string(expr), userCtx)
```

Add:

```go
func (s *PolicyMatchingService) ValidateCheck(
	ctx context.Context,
	orgName, projectSlug, modelID string,
	action rls.Action,
	input map[string]any,
	userCtx *rls.UserContext,
) error {
	policies, err := s.repo.ListByAction(ctx, orgName, projectSlug, modelID, action, userCtx.Roles)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return fmt.Errorf("RLS deny: no matching policy for action=%s", action)
	}
	for _, p := range policies {
		if p.WithCheckExpr == "" {
			continue
		}
		if err := s.checkEvaluator.ValidateInput(ctx, string(p.WithCheckExpr), input, userCtx); err == nil {
			return nil
		}
	}
	return fmt.Errorf("RLS CHECK violation: no matching input check policy passed")
}
```

Keep `ResolveCheck` temporarily for callers that still compile legacy SQL checks during the transition, but mark it as deprecated in a comment and route it through `ValidateCheck` in Task 6.

- [ ] **Step 4: Update service construction**

Modify `modelcraft-backend/internal/interfaces/http/routes.go`:

```go
rlsUsingCompiler := rls.NewPolicyExpressionSQLCompiler()
rlsCheckEvaluator := rls.NewPolicyExpressionInputEvaluator()
rlsMatchingSvc := rls.NewPolicyMatchingService(policyRepo, rlsUsingCompiler, rlsCheckEvaluator)
```

- [ ] **Step 5: Run matching tests and commit**

Run:

```bash
go test ./internal/app/rls -run 'TestResolveUsing|TestValidateCheck|TestResolveCheck' -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/internal/app/rls/policy_matching_service.go modelcraft-backend/internal/app/rls/policy_matching_service_test.go modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "feat: wire CEL RLS policy matching"
```

---

### Task 6: Enforce Input Check Before Runtime Create And Update

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go`
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver_end_user_ref_test.go`
- Modify: `modelcraft-backend/internal/app/modelruntime/graphql_app.go`
- Modify: `modelcraft-backend/internal/interfaces/runtime/rls_resolver.go`

- [ ] **Step 1: Add failing modelruntime tests for write guard short-circuit**

Append to `modelcraft-backend/internal/domain/modelruntime/model_resolver_end_user_ref_test.go`:

```go
type denyingRLSPolicyGuard struct{}

func (g denyingRLSPolicyGuard) ValidateInput(_ context.Context, _ string, _ Action, _ map[string]any) error {
	return fmt.Errorf("RLS CHECK violation: input rejected")
}

func TestRLSInputCheck_CreateOne_DeniesBeforeRepoCall(t *testing.T) {
	repo := &capturingClientRepo{}
	schema := buildSchemaFor(t, taskModelWithOwner())
	ctx := WithGraphqlRequestContext(
		context.Background(), repo, "org-1", "project-1", "u_123", "",
		&ResolvedModelPermissions{Insert: ActionPermission{Allowed: true}},
	)
	ctx = WithRLSPolicyGuard(ctx, denyingRLSPolicyGuard{})

	result := graphql.Do(graphql.Params{
		Schema:  *schema,
		Context: ctx,
		RequestString: `mutation {
			create(data: { title: "hello", owner: "u_123" }) { id }
		}`,
	})

	require.NotEmpty(t, result.Errors)
	require.Contains(t, result.Errors[0].Message, "RLS CHECK violation")
	require.Nil(t, repo.capturedCreateInput)
}

func TestRLSInputCheck_UpdateOne_DeniesBeforeRepoCall(t *testing.T) {
	repo := &capturingClientRepo{}
	schema := buildSchemaFor(t, taskModelWithOwner())
	ctx := WithGraphqlRequestContext(
		context.Background(), repo, "org-1", "project-1", "u_123", "",
		&ResolvedModelPermissions{Update: ActionPermission{Allowed: true}},
	)
	ctx = WithRLSPolicyGuard(ctx, denyingRLSPolicyGuard{})

	result := graphql.Do(graphql.Params{
		Schema:  *schema,
		Context: ctx,
		RequestString: `mutation {
			update(where: { id: "record-id" }, data: { title: "renamed" }) { success }
		}`,
	})

	require.NotEmpty(t, result.Errors)
	require.Contains(t, result.Errors[0].Message, "RLS CHECK violation")
	require.Nil(t, repo.capturedUpdateOneInput)
}
```

- [ ] **Step 2: Run write guard tests and verify they fail**

Run:

```bash
go test ./internal/domain/modelruntime -run 'TestRLSInputCheck' -count=1
```

Expected: FAIL with undefined `WithRLSPolicyGuard` or missing interface.

- [ ] **Step 3: Add runtime guard interface to request context**

Modify `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go`:

```go
type RLSPolicyGuard interface {
	ValidateInput(ctx context.Context, modelID string, action Action, input map[string]any) error
}

type graphqlRequestContext struct {
	ClientRepo       ClientDatabaseRepository
	relationLoaders  map[string]*dataloader.Loader[string, map[string]any]
	OrgName          string
	ProjectSlug      string
	CurrentEndUserID string
	EndUserAdminID    string
	EndUserPerms      *ResolvedModelPermissions
	RLSPolicyGuard    RLSPolicyGuard
}

func WithRLSPolicyGuard(ctx context.Context, guard RLSPolicyGuard) context.Context {
	rctx, ok := getGraphqlRequestContext(ctx)
	if !ok || rctx == nil {
		return ctx
	}
	next := *rctx
	next.RLSPolicyGuard = guard
	return context.WithValue(ctx, graphqlRequestContextKey{}, &next)
}
```

- [ ] **Step 4: Call guard before create/update repo calls**

In `executeCreateOne`, after UUID and owner injection and before `input.Id = ...`, add:

```go
if rctx.RLSPolicyGuard != nil {
	if err := rctx.RLSPolicyGuard.ValidateInput(p.Context, m.model.ID, ActionInsert, input.Data); err != nil {
		return nil, err
	}
}
```

In `executeCreateMany`, after owner injection for each `dataItem` and before `CreateMany`, add:

```go
if rctx.RLSPolicyGuard != nil {
	for _, dataItem := range input.Data {
		if err := rctx.RLSPolicyGuard.ValidateInput(p.Context, m.model.ID, ActionInsert, dataItem); err != nil {
			return nil, err
		}
	}
}
```

In `executeUpdateOne`, after owner injection and before row filter merge, add:

```go
if rctx.RLSPolicyGuard != nil {
	if err := rctx.RLSPolicyGuard.ValidateInput(p.Context, m.model.ID, ActionUpdate, input.Data); err != nil {
		return nil, err
	}
}
```

In `executeUpdateMany`, after owner injection and before row filter merge, add the same `ActionUpdate` validation for `input.Data`.

- [ ] **Step 5: Adapt `interfaces/runtime.RLSResolver` to implement the guard**

Modify `modelcraft-backend/internal/interfaces/runtime/rls_resolver.go`:

```go
type MatchingService interface {
	ResolveUsing(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, userCtx *rls.UserContext) (string, []interface{}, error)
	ValidateCheck(ctx context.Context, orgName, projectSlug, modelID string, action rls.Action, input map[string]any, userCtx *rls.UserContext) error
}

func (r *RLSResolver) ValidateInput(ctx context.Context, modelID string, action modelruntime.Action, input map[string]any) error {
	domainAction := rls.ActionCreate
	if action == modelruntime.ActionUpdate {
		domainAction = rls.ActionUpdate
	}
	userCtx := middleware.GetUserContext(ctx)
	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}
	rctx, ok := getRuntimeContext(ctx)
	if !ok {
		return fmt.Errorf("RLS CHECK violation: runtime context missing")
	}
	return r.matchingSvc.ValidateCheck(ctx, rctx.OrgName, rctx.ProjectSlug, modelID, domainAction, input, userCtx)
}
```

Keep `ValidateInsert` and `ValidateUpdate` as wrappers that call `ValidateInput`:

```go
func (r *RLSResolver) ValidateInsert(ctx context.Context, modelID string, input map[string]interface{}) error {
	return r.ValidateInput(ctx, modelID, modelruntime.ActionInsert, input)
}

func (r *RLSResolver) ValidateUpdate(ctx context.Context, modelID string, input map[string]interface{}) error {
	return r.ValidateInput(ctx, modelID, modelruntime.ActionUpdate, input)
}
```

- [ ] **Step 6: Wire the guard into `GraphqlAppService.Execute`**

Modify `modelcraft-backend/internal/app/modelruntime/graphql_app.go` by adding an optional field and constructor arg:

```go
type RuntimeRLSPolicyGuard interface {
	ValidateInput(ctx context.Context, modelID string, action modelruntime.Action, input map[string]any) error
}

type GraphqlAppService struct {
	modelRepo            modelruntime.ModelRepository
	graphqlSchemaManager *modelruntime.GraphqlSchemaManager
	permService          modelruntime.EndUserPermissionService
	rlsGuard             RuntimeRLSPolicyGuard
}
```

After `reqCtx := modelruntime.WithGraphqlRequestContext(...)`, add:

```go
if s.rlsGuard != nil {
	reqCtx = modelruntime.WithRLSPolicyGuard(reqCtx, s.rlsGuard)
}
```

Update `NewGraphqlAppService` to accept `rlsGuard RuntimeRLSPolicyGuard` and store it.

Update call site in `modelcraft-backend/internal/interfaces/http/routes.go`:

```go
runtimeRLSResolver := runtime.NewRLSResolver(logfacade.GetLogger(context.Background()), rlsMatchingSvc)
graphqlAppService := modelruntime.NewGraphqlAppService(modelRuntimeRepo, lfkRepo, permService, runtimeRLSResolver)
```

- [ ] **Step 7: Run write guard tests and commit**

Run:

```bash
go test ./internal/domain/modelruntime -run 'TestRLSInputCheck|TestEndUserRefOwnerInjection' -count=1
go test ./internal/interfaces/runtime ./internal/app/modelruntime -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go modelcraft-backend/internal/domain/modelruntime/model_resolver.go modelcraft-backend/internal/domain/modelruntime/model_resolver_end_user_ref_test.go modelcraft-backend/internal/app/modelruntime/graphql_app.go modelcraft-backend/internal/interfaces/runtime/rls_resolver.go modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "feat: enforce RLS input checks before writes"
```

---

### Task 7: Add Raw SQL RLS Filters To Runtime DML Inputs

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_input.go`
- Modify: `modelcraft-backend/internal/infrastructure/database/dml/sql_mapper.go`
- Test: `modelcraft-backend/internal/infrastructure/database/dml/sql_mapper_test.go`

- [ ] **Step 1: Add failing DML mapper tests for raw RLS filters**

Append to `modelcraft-backend/internal/infrastructure/database/dml/sql_mapper_test.go`:

```go
func TestConvertFindManyInputToSQL_WithRawRLSFilter(t *testing.T) {
	input := &modelruntime.FindManyInput{
		TableName: "posts",
		Where:     map[string]any{"status": "draft"},
		RawFilters: []modelruntime.RawSQLFilter{{
			SQL:    "owner_id = ?",
			Params: []any{"u_123"},
		}},
		Limit: 10,
	}

	sql, args, err := convertFindManyInputToSQL(context.Background(), input)

	require.NoError(t, err)
	require.Contains(t, sql, "WHERE")
	require.Contains(t, sql, "owner_id = ?")
	require.Equal(t, []any{"draft", "u_123", uint(10)}, args)
}

func TestConvertUpdateOneInputToSQL_WithRawRLSFilter(t *testing.T) {
	input := &modelruntime.UpdateOneInput{
		TableName: "posts",
		Where:     map[string]any{"id": "post-1"},
		Data:      map[string]any{"title": "new"},
		RawFilters: []modelruntime.RawSQLFilter{{
			SQL:    "owner_id = ?",
			Params: []any{"u_123"},
		}},
	}

	sql, args, err := convertUpdateOneInputToSQL(context.Background(), input)

	require.NoError(t, err)
	require.Contains(t, sql, "owner_id = ?")
	require.Contains(t, sql, "id")
	require.Contains(t, fmt.Sprint(args), "u_123")
}
```

- [ ] **Step 2: Run DML tests and verify they fail**

Run:

```bash
go test ./internal/infrastructure/database/dml -run 'TestConvert.*RawRLSFilter' -count=1
```

Expected: FAIL with undefined `RawSQLFilter`.

- [ ] **Step 3: Add raw filter fields to runtime input structs**

Modify `modelcraft-backend/internal/domain/modelruntime/graphql_input.go`:

```go
type RawSQLFilter struct {
	SQL    string
	Params []any
}
```

Update the input structs to include `RawFilters []RawSQLFilter` as shown:

```go
type FindUniqueInput struct {
	TableName  string
	Selection  *Selection
	Where      map[string]any
	RawFilters []RawSQLFilter
}

type FindManyInput struct {
	TableName     string
	Selection     *Selection
	Where         map[string]any
	RawFilters    []RawSQLFilter
	OrderBy       []OrderBy
	Limit         uint
	ExplicitLimit bool
	Offset        uint
}

type ListByCursorInput struct {
	TableName           string
	Selection           *Selection
	Where               map[string]any
	RawFilters          []RawSQLFilter
	SortField           string
	SortDirection       string
	InsertionOrderField string
	After               *CursorData
	Limit               uint
}

type FindFirstInput struct {
	TableName  string
	Selection  *Selection
	Where      map[string]any
	RawFilters []RawSQLFilter
}

type UpdateOneInput struct {
	TableName  string
	UpdatedObj bool
	Where      map[string]any
	RawFilters []RawSQLFilter
	Data       map[string]any
}

type DeleteOneInput struct {
	DeletedObj  bool
	TableName   string
	Where       map[string]any
	RawFilters  []RawSQLFilter
}

type UpdateManyInput struct {
	TableName  string
	Where      map[string]any
	RawFilters []RawSQLFilter
	Data       map[string]any
	Take       uint
}

type DeleteManyInput struct {
	TableName  string
	Where      map[string]any
	RawFilters []RawSQLFilter
	Take       uint
}

type AggregateInput struct {
	TableName  string
	Where      map[string]any
	RawFilters []RawSQLFilter
	Count      map[string]bool
	Avg        map[string]bool
	Sum        map[string]bool
	Min        map[string]bool
	Max        map[string]bool
}

type CountInput struct {
	TableName  string
	Where      map[string]any
	RawFilters []RawSQLFilter
	Select     map[string]bool
}
```

- [ ] **Step 4: Apply raw filters in SQL mapper**

Modify `modelcraft-backend/internal/infrastructure/database/dml/sql_mapper.go` by adding:

```go
func rawFiltersToExpressions(filters []modelruntime.RawSQLFilter) ([]goqu.Expression, error) {
	exprs := make([]goqu.Expression, 0, len(filters))
	for _, filter := range filters {
		if strings.TrimSpace(filter.SQL) == "" {
			return nil, bizerrors.Errorf("raw RLS filter SQL cannot be empty")
		}
		exprs = append(exprs, goqu.L(filter.SQL, filter.Params...))
	}
	return exprs, nil
}
```

In each `convert*InputToSQL` function that has `Where`, after converting the structured `Where`, append raw filters:

```go
rawExprs, err := rawFiltersToExpressions(input.RawFilters)
if err != nil {
	return "", nil, err
}
whereExprs := []goqu.Expression{whereExpr}
whereExprs = append(whereExprs, rawExprs...)
ds = ds.Where(whereExprs...)
```

For find/list/count functions that allow no `Where`, apply raw filters even when structured `Where` is empty.

- [ ] **Step 5: Run DML mapper tests and commit**

Run:

```bash
go test ./internal/infrastructure/database/dml -run 'TestConvert.*RawRLSFilter|TestConvertFindManyInputToSQL|TestConvertUpdateOneInputToSQL' -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/internal/domain/modelruntime/graphql_input.go modelcraft-backend/internal/infrastructure/database/dml/sql_mapper.go modelcraft-backend/internal/infrastructure/database/dml/sql_mapper_test.go
git commit -m "feat: support raw RLS filters in runtime DML"
```

---

### Task 8: Apply Using Filters To Runtime Read, Update, And Delete

**Files:**
- Modify: `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go`
- Modify: `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`
- Test: `modelcraft-backend/internal/domain/modelruntime/model_resolver_permission_enforcement_test.go`

- [ ] **Step 1: Add failing tests for raw using filter propagation**

Append to `modelcraft-backend/internal/domain/modelruntime/model_resolver_permission_enforcement_test.go`:

```go
type allowingRLSUsingGuard struct{}

func (g allowingRLSUsingGuard) ValidateInput(_ context.Context, _ string, _ Action, _ map[string]any) error {
	return nil
}

func (g allowingRLSUsingGuard) ResolveUsingFilter(_ context.Context, _ string, _ Action) (*RawSQLFilter, error) {
	return &RawSQLFilter{SQL: "owner_id = ?", Params: []any{"u_123"}}, nil
}

func TestRLSUsingFilter_FindMany_AttachesRawFilter(t *testing.T) {
	repo := &fullCapturingRepo{}
	schema := buildSchemaFor(t, taskModelWithoutOwner())
	ctx := WithGraphqlRequestContext(context.Background(), repo, "org-1", "project-1", "u_123", "",
		&ResolvedModelPermissions{Select: ActionPermission{Allowed: true}},
	)
	ctx = WithRLSPolicyGuard(ctx, allowingRLSUsingGuard{})

	result := graphql.Do(graphql.Params{
		Schema:  *schema,
		Context: ctx,
		RequestString: `query {
			findMany(where: { title: { equals: "draft" } }) { items { id title } }
		}`,
	})

	require.Empty(t, result.Errors)
	require.Len(t, repo.capturedFindManyRawFilters, 1)
	require.Equal(t, "owner_id = ?", repo.capturedFindManyRawFilters[0].SQL)
}
```

Before this test, update `fullCapturingRepo` in the same test file:

```go
type fullCapturingRepo struct {
	mockClientDatabaseRepository
	capturedFindManyWhere          map[string]any
	capturedFindManyRawFilters     []RawSQLFilter
	capturedListByCursorWhere      map[string]any
	capturedListByCursorRawFilters []RawSQLFilter
	capturedListByPageWhere        map[string]any
	capturedFindUniqueWhere        map[string]any
	capturedFindFirstWhere         map[string]any
	capturedUpdateOneWhere         map[string]any
	capturedUpdateOneRawFilters    []RawSQLFilter
	capturedDeleteOneWhere         map[string]any
	capturedDeleteOneRawFilters    []RawSQLFilter
	capturedUpdateManyWhere        map[string]any
	capturedUpdateManyRawFilters   []RawSQLFilter
	capturedDeleteManyWhere        map[string]any
	capturedDeleteManyRawFilters   []RawSQLFilter
}

func (r *fullCapturingRepo) FindMany(_ context.Context, input *FindManyInput) ([]map[string]any, error) {
	r.capturedFindManyWhere = input.Where
	r.capturedFindManyRawFilters = input.RawFilters
	r.capturedListByPageWhere = input.Where
	return []map[string]any{}, nil
}

func (r *fullCapturingRepo) ListByCursor(_ context.Context, input *ListByCursorInput) ([]map[string]any, error) {
	r.capturedListByCursorWhere = input.Where
	r.capturedListByCursorRawFilters = input.RawFilters
	return []map[string]any{}, nil
}

func (r *fullCapturingRepo) UpdateOne(_ context.Context, input *UpdateOneInput) (map[string]any, error) {
	r.capturedUpdateOneWhere = input.Where
	r.capturedUpdateOneRawFilters = input.RawFilters
	return map[string]any{"id": "x"}, nil
}

func (r *fullCapturingRepo) DeleteOne(_ context.Context, input *DeleteOneInput) (map[string]any, error) {
	r.capturedDeleteOneWhere = input.Where
	r.capturedDeleteOneRawFilters = input.RawFilters
	return map[string]any{"id": "x"}, nil
}

func (r *fullCapturingRepo) UpdateMany(_ context.Context, input *UpdateManyInput) (any, error) {
	r.capturedUpdateManyWhere = input.Where
	r.capturedUpdateManyRawFilters = input.RawFilters
	return map[string]any{"count": 1}, nil
}

func (r *fullCapturingRepo) DeleteMany(_ context.Context, input *DeleteManyInput) (any, error) {
	r.capturedDeleteManyWhere = input.Where
	r.capturedDeleteManyRawFilters = input.RawFilters
	return map[string]any{"count": 1}, nil
}
```

- [ ] **Step 2: Run using filter propagation test and verify it fails**

Run:

```bash
go test ./internal/domain/modelruntime -run 'TestRLSUsingFilter' -count=1
```

Expected: FAIL because `ResolveUsingFilter` is not defined on `RLSPolicyGuard`.

- [ ] **Step 3: Extend guard interface and attach filters in resolvers**

Modify `modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go`:

```go
type RLSPolicyGuard interface {
	ValidateInput(ctx context.Context, modelID string, action Action, input map[string]any) error
	ResolveUsingFilter(ctx context.Context, modelID string, action Action) (*RawSQLFilter, error)
}
```

Add helper in `modelcraft-backend/internal/domain/modelruntime/model_resolver.go`:

```go
func (m *graphqlModelResolver) appendRLSUsingFilter(
	ctx context.Context,
	rctx *graphqlRequestContext,
	modelID string,
	action Action,
	filters []RawSQLFilter,
) ([]RawSQLFilter, error) {
	if rctx.RLSPolicyGuard == nil {
		return filters, nil
	}
	filter, err := rctx.RLSPolicyGuard.ResolveUsingFilter(ctx, modelID, action)
	if err != nil {
		return filters, err
	}
	if filter == nil {
		return filters, nil
	}
	return append(filters, *filter), nil
}
```

Call this helper before repository calls for:

- `FindUnique`
- `FindFirst`
- `FindMany`
- list by cursor
- list by page count query
- `UpdateOne`
- `UpdateMany`
- `DeleteOne`
- `DeleteMany`
- `Count`
- `Aggregate`

For update/delete, append the raw filter after existing `BuildRowFilter` merge and before `ClientRepo`.

- [ ] **Step 4: Implement using filter resolution in runtime RLS resolver**

Modify `modelcraft-backend/internal/interfaces/runtime/rls_resolver.go`:

```go
func (r *RLSResolver) ResolveUsingFilter(ctx context.Context, modelID string, action modelruntime.Action) (*modelruntime.RawSQLFilter, error) {
	domainAction := rls.ActionRead
	switch action {
	case modelruntime.ActionUpdate:
		domainAction = rls.ActionUpdate
	case modelruntime.ActionDelete:
		domainAction = rls.ActionDelete
	}
	userCtx := middleware.GetUserContext(ctx)
	if userCtx == nil {
		userCtx = &rls.UserContext{}
	}
	rctx, ok := getRuntimeContext(ctx)
	if !ok {
		return nil, fmt.Errorf("RLS USING violation: runtime context missing")
	}
	sql, params, err := r.matchingSvc.ResolveUsing(ctx, rctx.OrgName, rctx.ProjectSlug, modelID, domainAction, userCtx)
	if err != nil {
		return nil, err
	}
	return &modelruntime.RawSQLFilter{SQL: sql, Params: params}, nil
}
```

- [ ] **Step 5: Run runtime using tests and commit**

Run:

```bash
go test ./internal/domain/modelruntime -run 'TestRLSUsingFilter|TestPermission' -count=1
go test ./internal/interfaces/runtime -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/internal/domain/modelruntime/graphql_request_context.go modelcraft-backend/internal/domain/modelruntime/model_resolver.go modelcraft-backend/internal/domain/modelruntime/model_resolver_permission_enforcement_test.go modelcraft-backend/internal/interfaces/runtime/rls_resolver.go
git commit -m "feat: apply RLS using filters in runtime"
```

---

### Task 9: Extend GraphQL Dry Run To Return SQL Or Boolean Result

**Files:**
- Modify: `modelcraft-backend/api/graph/project/schema/rls.graphql`
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/rls.resolvers.go`
- Modify: `modelcraft-backend/internal/app/rls/model_rls_policy_app_service.go`
- Modify generated files under `modelcraft-backend/internal/interfaces/graphql/project/generated/`

- [ ] **Step 1: Update schema with dry-run fields**

Modify `modelcraft-backend/api/graph/project/schema/rls.graphql`:

```graphql
type ValidateRLSExprPayload {
  result: ValidationResult!
  error: ValidateRLSExprError
  dryRun: RLSExprDryRun
}

type RLSExprDryRun {
  sql: String
  params: [String!]
  result: Boolean
}

input ValidateRLSExprInput {
  modelId: ID!
  exprType: RLSExprType!
  expression: String!
  sampleInput: String
}
```

`sampleInput` is a JSON object string used only for check expressions.

- [ ] **Step 2: Regenerate gqlgen code**

Run:

```bash
go run github.com/99designs/gqlgen
```

Expected: generated project GraphQL files update without manual edits in generated code.

- [ ] **Step 3: Add failing resolver/app tests**

Create or extend `modelcraft-backend/internal/interfaces/graphql/project/rls_resolvers_test.go` with resolver-level tests that assert:

```go
func TestValidateRLSExpr_UsingDryRunReturnsSQL(t *testing.T) {
	// Input: row.owner_id == auth.user_id
	// Expected: payload.DryRun.SQL contains "owner_id = ?"
	// Expected: payload.DryRun.Params contains "u_123" when resolver context has user id u_123
}

func TestValidateRLSExpr_CheckDryRunReturnsBoolean(t *testing.T) {
	// Input: input.owner_id == auth.user_id, sampleInput {"owner_id":"u_123"}
	// Expected: payload.DryRun.Result == true
}
```

- [ ] **Step 4: Implement dry-run app method**

Add to `modelcraft-backend/internal/app/rls/model_rls_policy_app_service.go`:

```go
func (s *ModelRLSPolicyAppService) DryRunExpr(
	ctx context.Context,
	orgName, projectSlug, modelID string,
	exprType domainrls.ExprType,
	expression string,
	sampleInput map[string]any,
	userCtx *domainrls.UserContext,
) PolicyExpressionDryRunResult {
	// Validate expression first.
	// For predicate exprType: compile using SQL and return SQL/params.
	// For check exprType: evaluate sampleInput and return boolean result.
}
```

Use `PolicyExpressionSQLCompiler` for predicate types and `PolicyExpressionInputEvaluator` for check types.

- [ ] **Step 5: Populate dry-run payload in resolver**

Modify `ValidateRLSExpr` in `modelcraft-backend/internal/interfaces/graphql/project/rls.resolvers.go`:

```go
var sampleInput map[string]any
if input.SampleInput != nil && *input.SampleInput != "" {
	if err := json.Unmarshal([]byte(*input.SampleInput), &sampleInput); err != nil {
		// return valid=false with INVALID_SAMPLE_INPUT
	}
}
dryRun := r.RLSPolicyAppService.DryRunExpr(ctx, orgName, projectSlug, input.ModelID,
	domainRLS.ExprType(exprType), input.Expression, sampleInput, middleware.GetUserContext(ctx))
```

Convert params with `fmt.Sprint(param)` for GraphQL `[String!]`.

- [ ] **Step 6: Run GraphQL tests and commit**

Run:

```bash
go test ./internal/interfaces/graphql/project -run 'TestValidateRLSExpr' -count=1
go test ./internal/app/rls -run 'Test.*DryRun|TestPolicyExpression' -count=1
```

Expected: PASS.

Commit:

```bash
git add modelcraft-backend/api/graph/project/schema/rls.graphql modelcraft-backend/internal/interfaces/graphql/project/rls.resolvers.go modelcraft-backend/internal/app/rls/model_rls_policy_app_service.go modelcraft-backend/internal/interfaces/graphql/project/generated
git commit -m "feat: add RLS CEL dry run payload"
```

---

### Task 10: Update Frontend Editor For CEL Policy Expressions

**Files:**
- Modify: `modelcraft-front/src/api-client/rls-policy/graphql-docs.ts`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyManage.ts`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsExpressionEditor.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/PolicyEditorDialog.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.ts`
- Test: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.test.ts`

- [ ] **Step 1: Update failing frontend utility tests**

Modify `rls-expression-utils.test.ts` so syntax examples are CEL:

```ts
it('accepts CEL expressions for visible policy fields', () => {
  expect(validateRlsExpressionSyntax('row.owner_id == auth.user_id')).toEqual({
    valid: true,
    empty: false,
  })
  expect(validateRlsExpressionSyntax('input.status in ["draft", "pending"]')).toEqual({
    valid: true,
    empty: false,
  })
})

it('rejects JSON object syntax for new policy expressions', () => {
  expect(validateRlsExpressionSyntax('{"owner_id":{"equals":"{{user_id}}"}}')).toEqual({
    valid: false,
    empty: false,
    message: '请输入 CEL 表达式，例如 row.owner_id == auth.user_id',
  })
})
```

- [ ] **Step 2: Run frontend utility tests and verify they fail**

Run:

```bash
npm run test -- src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.test.ts
```

Expected: FAIL because current syntax helper expects JSON.

- [ ] **Step 3: Update syntax helper for CEL**

Modify `rls-expression-utils.ts`:

```ts
export function validateRlsExpressionSyntax(value: string): SyntaxResult {
  const trimmed = value.trim()
  if (!trimmed) return { valid: true, empty: true }
  if (trimmed.startsWith('{')) {
    return {
      valid: false,
      empty: false,
      message: '请输入 CEL 表达式，例如 row.owner_id == auth.user_id',
    }
  }
  const hasAllowedRoot = /\b(row|input|auth)\.[A-Za-z_][A-Za-z0-9_]*/.test(trimmed)
  if (!hasAllowedRoot) {
    return {
      valid: false,
      empty: false,
      message: '表达式需要引用 row、input 或 auth',
    }
  }
  return { valid: true, empty: false }
}
```

- [ ] **Step 4: Update labels and placeholders**

Modify `PolicyEditorDialog.tsx`:

```tsx
{showUsingExpr && (
  <RlsExpressionEditor
    label="Using Filter"
    placeholder="例如：row.owner_id == auth.user_id"
    value={usingExpr}
    onChange={setUsingExpr}
    exprType={getRlsExpressionType(action, 'using')}
    onDryRun={onDryRun}
  />
)}

{showCheckExpr && (
  <RlsExpressionEditor
    label="Input Check"
    placeholder="例如：input.owner_id == auth.user_id"
    value={withCheckExpr}
    onChange={setWithCheckExpr}
    exprType={getRlsExpressionType(action, 'check')}
    onDryRun={onDryRun}
  />
)}
```

Modify `RlsExpressionEditor.tsx` props:

```ts
placeholder: string
```

Use the prop in `<Textarea placeholder={placeholder} />`, and change successful local status text to:

```tsx
text: 'CEL 语法初步通过'
```

- [ ] **Step 5: Update dry-run GraphQL document**

Modify `modelcraft-front/src/api-client/rls-policy/graphql-docs.ts`:

```graphql
dryRun {
  sql
  params
  result
}
```

inside `VALIDATE_RLS_EXPR`.

- [ ] **Step 6: Display SQL or boolean dry-run results**

Modify `useRlsPolicyManage.ts` dry-run result mapping:

```ts
const dryRun = payload?.dryRun
if (dryRun?.sql) {
  return {
    success: true,
    message: `${dryRun.sql}${dryRun.params?.length ? ` | params: ${dryRun.params.join(', ')}` : ''}`,
  }
}
if (typeof dryRun?.result === 'boolean') {
  return {
    success: dryRun.result,
    message: dryRun.result ? 'Input Check 返回 true' : 'Input Check 返回 false',
  }
}
```

- [ ] **Step 7: Run frontend tests and lint, then commit**

Run:

```bash
npm run test -- src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.test.ts
npx eslint src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/PolicyEditorDialog.tsx src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsExpressionEditor.tsx src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.ts src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyManage.ts src/api-client/rls-policy/graphql-docs.ts
```

Expected: PASS.

Commit:

```bash
git add modelcraft-front/src/api-client/rls-policy/graphql-docs.ts modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyManage.ts modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsExpressionEditor.tsx modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/PolicyEditorDialog.tsx modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.ts modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.test.ts
git commit -m "feat: update RLS editor for CEL expressions"
```

---

### Task 11: Final Verification

**Files:**
- No source files expected unless verification exposes a bug.

- [ ] **Step 1: Run backend focused tests**

Run:

```bash
go test ./internal/app/rls ./internal/interfaces/runtime ./internal/domain/modelruntime ./internal/infrastructure/database/dml ./internal/interfaces/graphql/project -count=1
```

Expected: PASS.

- [ ] **Step 2: Run frontend focused tests**

Run:

```bash
npm run test -- src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.test.ts
```

Expected: PASS.

- [ ] **Step 3: Run generation checks**

Run:

```bash
go run github.com/99designs/gqlgen
npm run codegen
```

Expected: generated code is current. If `npm run codegen` fails on unrelated cluster contract drift, record the exact existing error in the final handoff and do not modify cluster files unless the user explicitly includes that scope.

- [ ] **Step 4: Run lint/build checks**

Run:

```bash
go test ./...
npx eslint src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/PolicyEditorDialog.tsx src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/RlsExpressionEditor.tsx src/app/org/[orgName]/project/[projectSlug]/access-control/_components/rls-policy/rls-expression-utils.ts src/app/org/[orgName]/project/[projectSlug]/access-control/_hooks/rls-policy/useRlsPolicyManage.ts src/api-client/rls-policy/graphql-docs.ts
```

Expected: PASS, except for known unrelated tests already failing before this work. Any unrelated failure must be named with file path and error text.

- [ ] **Step 5: Commit verification fixes if any**

If verification required fixes, first run:

```bash
git status --short
```

Then stage only the files changed by the verification fix. For example, if the fix touched the SQL compiler and its test:

```bash
git add modelcraft-backend/internal/app/rls/policy_expression_sql_compiler.go modelcraft-backend/internal/app/rls/policy_expression_sql_compiler_test.go
git commit -m "fix: stabilize RLS CEL policy expression integration"
```

If verification did not require a source change, do not create a commit.
