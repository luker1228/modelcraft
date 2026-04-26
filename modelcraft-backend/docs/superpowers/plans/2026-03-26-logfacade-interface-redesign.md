# logfacade Interface Redesign Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Simplify the logfacade Logger interface by making context a required first parameter on all logging methods, removing WithContext chaining, and eliminating DeepInfof.

**Architecture:** 
- Phase 1: Update core interface and implementation (pkg/logfacade)
- Phase 2: Update all callers in layers: application, domain, infrastructure, middleware, GraphQL resolvers, HTTP handlers
- Phase 3: Verify build and tests pass

**Tech Stack:** Go, zap logging, context.Context

---

## File Structure

**Files to modify:**

**Core logfacade package:**
- `pkg/logfacade/interface.go` — update Logger interface signature
- `pkg/logfacade/zap_logger.go` — update all 12 log methods to accept ctx as first param
- `pkg/logfacade/global.go` — may need minor updates if any helper functions call log methods
- `pkg/logfacade/logger.go` — factory function; check for any internal logger calls
- `pkg/logfacade/interface_test.go` — update test calls to match new signatures
- `pkg/logfacade/example_test.go` — update example calls

**Callers (51 files across application, domain, infrastructure, middleware, interfaces):**
- See inventory: ~51 files that call logger methods
- Pattern: find `.WithContext(ctx).Method(...)` or `.Method(msg, fields)` and rewrite to `.Method(ctx, msg, fields)` or `.Method(ctx, format, args)`
- Affected directories: `internal/app/`, `internal/domain/`, `internal/infrastructure/`, `internal/interfaces/`, `internal/middleware/`, `pkg/`, `cmd/`

**No files to create** — only modifications.

---

## Chunk 1: Core Interface and Implementation

### Task 1: Update Logger interface signature

**Files:**
- Modify: `pkg/logfacade/interface.go:26-105` (Logger interface + methods)

- [ ] **Step 1: Read the current interface**

Run:
```bash
head -n 105 /root/modelcraft_project/modelcraft-go/pkg/logfacade/interface.go | tail -n 80
```

- [ ] **Step 2: Replace the interface definition**

Replace the Logger interface to add `ctx context.Context` as first param on all methods and remove `WithContext`, remove `DeepInfof`:

```go
type Logger interface {
	// Debug records a debug-level log message with context.
	Debug(ctx context.Context, msg string, fields ...Field)

	// Info records an info-level log message with context.
	Info(ctx context.Context, msg string, fields ...Field)

	// Warn records a warn-level log message with context.
	Warn(ctx context.Context, msg string, fields ...Field)

	// Error records an error-level log message with context.
	Error(ctx context.Context, msg string, fields ...Field)

	// Fatal records a fatal-level log message with context and exits.
	Fatal(ctx context.Context, msg string, fields ...Field)

	// Debugf records a debug-level log message using printf-style formatting with context.
	Debugf(ctx context.Context, format string, args ...any)

	// Infof records an info-level log message using printf-style formatting with context.
	Infof(ctx context.Context, format string, args ...any)

	// Warnf records a warn-level log message using printf-style formatting with context.
	Warnf(ctx context.Context, format string, args ...any)

	// Errorf records an error-level log message using printf-style formatting with context.
	Errorf(ctx context.Context, format string, args ...any)

	// Fatalf records a fatal-level log message using printf-style formatting with context and exits.
	Fatalf(ctx context.Context, format string, args ...any)

	// With creates a child logger with the given fields pre-attached to all subsequent log calls.
	With(fields ...Field) Logger

	// Sync flushes any buffered log entries. Call before program exit.
	Sync() error
}
```

- [ ] **Step 3: Verify the old methods are removed**

In interface.go, search for and delete:
- Line with `WithContext(ctx context.Context) Logger`
- Line with `DeepInfof(msg string, args ...any)`
- Any docstrings for these methods

- [ ] **Step 4: Commit**

```bash
cd /root/modelcraft_project/modelcraft-go
git add pkg/logfacade/interface.go
git commit -m "refactor: update Logger interface with ctx as first param, remove WithContext and DeepInfof"
```

---

### Task 2: Update zapLogger implementation

**Files:**
- Modify: `pkg/logfacade/zap_logger.go` (all log methods)

- [ ] **Step 1: Update Debug method**

Find the current Debug method (around line 94) and replace:

```go
func (z *ZapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	if ctx == nil {
		ctx = context.Background()
	}
	contextFields := z.extractContextFields(ctx)
	allFields := append(contextFields, z.convertFields(fields...)...)
	z.logger.Debug(msg, allFields...)
}
```

- [ ] **Step 2: Update Info method**

Find the current Info method and replace:

```go
func (z *ZapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	if ctx == nil {
		ctx = context.Background()
	}
	contextFields := z.extractContextFields(ctx)
	allFields := append(contextFields, z.convertFields(fields...)...)
	z.logger.Info(msg, allFields...)
}
```

- [ ] **Step 3: Update Warn method**

```go
func (z *ZapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	if ctx == nil {
		ctx = context.Background()
	}
	contextFields := z.extractContextFields(ctx)
	allFields := append(contextFields, z.convertFields(fields...)...)
	z.logger.Warn(msg, allFields...)
}
```

- [ ] **Step 4: Update Error method**

```go
func (z *ZapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	if ctx == nil {
		ctx = context.Background()
	}
	contextFields := z.extractContextFields(ctx)
	allFields := append(contextFields, z.convertFields(fields...)...)
	z.logger.Error(msg, allFields...)
}
```

- [ ] **Step 5: Update Fatal method**

```go
func (z *ZapLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	if ctx == nil {
		ctx = context.Background()
	}
	contextFields := z.extractContextFields(ctx)
	allFields := append(contextFields, z.convertFields(fields...)...)
	z.logger.Fatal(msg, allFields...)
}
```

- [ ] **Step 6: Add Debugf method**

Insert before Infof (Debugf was missing from the old interface):

```go
func (z *ZapLogger) Debugf(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Debug(msg, contextFields...)
}
```

- [ ] **Step 7: Update Infof method**

Find the current Infof method and replace:

```go
func (z *ZapLogger) Infof(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Info(msg, contextFields...)
}
```

- [ ] **Step 8: Update Warnf method**

```go
func (z *ZapLogger) Warnf(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Warn(msg, contextFields...)
}
```

- [ ] **Step 9: Update Errorf method**

```go
func (z *ZapLogger) Errorf(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Error(msg, contextFields...)
}
```

- [ ] **Step 10: Add Fatalf method**

Insert after Errorf:

```go
func (z *ZapLogger) Fatalf(ctx context.Context, format string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	contextFields := z.extractContextFields(ctx)
	z.logger.Fatal(msg, contextFields...)
}
```

- [ ] **Step 11: Delete WithContext method**

Find and delete the entire `WithContext` method implementation.

- [ ] **Step 12: Delete DeepInfof method**

Find and delete the entire `DeepInfof` method implementation (lines ~81-90).

- [ ] **Step 13: Verify no compile errors**

Run:
```bash
cd /root/modelcraft_project/modelcraft-go
go build ./pkg/logfacade
```

Expected: Success with zero errors.

**⚠️ IMPORTANT: Intermediate Build State**
After Task 1 is committed and Task 2 begins, `zap_logger.go` does not yet implement the new interface. This is temporary and expected — the build will fail until Task 2 is complete. Do NOT be alarmed. Continue with Task 2 to completion, then the project will compile again.

- [ ] **Step 14: Commit**

```bash
git add pkg/logfacade/zap_logger.go
git commit -m "refactor: update zapLogger methods with ctx first param, add Debugf/Fatalf, remove WithContext/DeepInfof"
```

---

### Task 3: Update logfacade tests

**Files:**
- Modify: `pkg/logfacade/interface_test.go` (test method calls)
- Modify: `pkg/logfacade/example_test.go` (example method calls)

- [ ] **Step 1: Read interface_test.go**

Run:
```bash
cat /root/modelcraft_project/modelcraft-go/pkg/logfacade/interface_test.go
```

- [ ] **Step 2: Update all logger calls in interface_test.go**

Find all calls like `logger.WithContext(ctx)...` or `logger.Stack(...)` and update:
- `logger.WithContext(ctx).Error("msg", logfacade.Err(err), logfacade.Stack(err))` → `logger.Error(ctx, "msg", logfacade.Err(err), logfacade.Stack(err))`
- `logger.Error("msg", fields...)` → `logger.Error(ctx, "msg", fields...)`

Use find-and-replace:
- Replace `.WithContext(ctx).` with just `.` (remove WithContext chain)
- Then add `ctx` as first argument to each method call

- [ ] **Step 3: Read example_test.go**

Run:
```bash
cat /root/modelcraft_project/modelcraft-go/pkg/logfacade/example_test.go
```

- [ ] **Step 4: Update all logger calls in example_test.go**

Same pattern as interface_test.go:
- `logger.WithContext(ctx).Infof(...)` → `logger.Infof(ctx, ...)`
- `logger.Info("msg", fields...)` → `logger.Info(ctx, "msg", fields...)`

- [ ] **Step 5: Run logfacade tests**

Run:
```bash
cd /root/modelcraft_project/modelcraft-go
go test ./pkg/logfacade -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/logfacade/interface_test.go pkg/logfacade/example_test.go
git commit -m "test: update logfacade tests with new method signatures"
```

---

### Task 3a: Update test files across codebase

**Strategy:** Test files across the codebase (~73 files) will also have logger calls. These will be updated incrementally as compile errors surface. However, you can proactively fix the main ones:

**High-impact test files to update proactively:**

- [ ] **Step 1: Find test files with logger calls**

Run:
```bash
find /root/modelcraft_project/modelcraft-go -name "*_test.go" -exec grep -l "logger\.\(Info\|Error\|Warn\|Debug\|Infof\|WithContext\)" {} \; | head -20
```

- [ ] **Step 2: For each test file, apply the same pattern**

```bash
# Remove WithContext chaining in test files
find /root/modelcraft_project/modelcraft-go -name "*_test.go" -exec sed -i 's/\.WithContext(ctx)\././' {} \;
```

- [ ] **Step 3: Update remaining test calls**

For test files, scan for remaining calls and add `ctx` as first param:
```bash
grep -r "logger\.Info(\|logger\.Error(\|logger\.Warn(" /root/modelcraft_project/modelcraft-go --include="*_test.go" | head -10
```

Update each by hand or with targeted sed rules.

- [ ] **Step 4: Do NOT commit test files yet**

Test file changes will be committed together with their corresponding main code file commits in Chunk 2, to keep commits logically grouped (e.g., app layer changes + app tests together).

---

## Chunk 2: Update Callers (51 files)

### Task 4: Update infrastructure layer callers (11 files)

**Files to update (grep in these for `.WithContext(` or `.Info(`, `.Error(`, etc.):**
- `internal/infrastructure/database/ddl/introspector.go`
- `internal/infrastructure/repository/gorm_logger.go`
- `internal/infrastructure/repository/membership_model.go`
- `internal/infrastructure/repository/organization_model.go`
- `internal/infrastructure/repository/role_model.go`
- `internal/infrastructure/repository/sqlc_logger.go`
- `internal/infrastructure/repository/user_model.go`
- `internal/infrastructure/auth/casbin_enforcer.go`
- `internal/infrastructure/auth/auth_provider_provider.go`
- `internal/infrastructure/auth_provider/client.go`
- `internal/infrastructure/database/dml/client_db_repo_impl.go`

- [ ] **Step 1: Find and replace `.WithContext(` calls**

For each file above, run:
```bash
grep -n "\.WithContext(" <file>
```

For each line found, replace:
- `logger.WithContext(ctx).Infof(...)` → `logger.Infof(ctx, ...)`
- `logger.WithContext(ctx).Info(...)` → `logger.Info(ctx, ...)`
- etc. for all log levels

- [ ] **Step 2: Find and replace inline fields (old style)**

For each file above, run:
```bash
grep -n "\.Info(" <file> | head -5
```

For each line with inline fields, reorder `ctx` to first param:
- `logger.Info("msg", logfacade.String("k", v))` → `logger.Info(ctx, "msg", logfacade.String("k", v))`
- `logger.Error("msg", logfacade.Err(err))` → `logger.Error(ctx, "msg", logfacade.Err(err))`

- [ ] **Step 3: Fix each file systematically**

For infrastructure files, apply bulk changes:
```bash
cd /root/modelcraft_project/modelcraft-go

# Replace .WithContext(ctx). with . (remove the chaining)
# Then manually verify each log method call has ctx as first param

# Example for one file:
sed -i 's/\.WithContext(ctx)\././' internal/infrastructure/database/ddl/introspector.go

# Then manually check for any `Info(msg` or `Error(msg` calls and add ctx
```

- [ ] **Step 4: Verify infrastructure layer builds**

Run:
```bash
cd /root/modelcraft_project/modelcraft-go
go build ./internal/infrastructure/...
```

Expected: Compile errors for remaining `.Info(msg, ...)`-style calls (next step fixes those).

- [ ] **Step 5: Fix remaining callers in infrastructure**

For any remaining compile errors, add `ctx` as first param:
```go
// Old: logger.Info("msg", fields...)
// New: logger.Info(ctx, "msg", fields...)
```

Use find-and-replace with care or edit files individually.

- [ ] **Step 6: Verify build again**

Run:
```bash
go build ./internal/infrastructure/...
```

Expected: Success.

- [ ] **Step 7: Commit infrastructure changes**

```bash
git add internal/infrastructure/
git commit -m "refactor: update infrastructure layer logfacade calls with ctx first param"
```

---

### Task 5: Update middleware layer callers (5 files)

**Files to update:**
- `internal/middleware/chi_http_context.go`
- `internal/middleware/chi_jwt_auth.go`
- `internal/middleware/chi_logger.go`
- `internal/middleware/chi_permission_test.go`
- `internal/middleware/chi_tenant.go`

- [ ] **Step 1: Apply systematic replacements**

```bash
cd /root/modelcraft_project/modelcraft-go

# Remove WithContext chaining in middleware files
for file in internal/middleware/chi_*.go; do
  sed -i 's/\.WithContext(ctx)\././' "$file"
done
```

- [ ] **Step 2: Fix remaining method calls**

For each file, find calls like `.Info("msg"` or `.Error("msg"` and add `ctx` as first param.

Run per file:
```bash
grep -n "\.Info(\|\.Error(\|\.Warn(\|\.Debug(\|\.Fatal(" internal/middleware/chi_http_context.go
```

Replace manually:
- `logger.Info("msg")` → `logger.Info(ctx, "msg")`
- `logger.Error("msg", fields...)` → `logger.Error(ctx, "msg", fields...)`

- [ ] **Step 3: Verify middleware builds**

Run:
```bash
go build ./internal/middleware
```

Expected: Success.

- [ ] **Step 4: Commit middleware changes**

```bash
git add internal/middleware/
git commit -m "refactor: update middleware layer logfacade calls with ctx first param"
```

---

### Task 6: Update application layer callers (13 files)

**Files to update:**
- `internal/app/auth/token_service.go`
- `internal/app/cluster/cluster_app.go`
- `internal/app/modeldesign/actual_schema_usecase.go`
- `internal/app/modeldesign/model_app.go`
- `internal/app/modeldesign/repair_usecase.go`
- `internal/app/modeldesign/reverse_engineer_app.go`
- `internal/app/modelruntime/graphql_app.go`
- `internal/app/organization/create_organization_service.go`
- `internal/app/permission/permission_service.go`
- `internal/app/permission/role_service.go`
- `internal/app/permission/system_role_syncer.go`
- `internal/app/permission/user_role_service.go`
- `internal/app/project/project_service.go`

- [ ] **Step 1: Apply systematic replacements across app layer**

```bash
cd /root/modelcraft_project/modelcraft-go

# Remove WithContext chaining in app files
find internal/app -name "*.go" -not -name "*_test.go" | while read file; do
  sed -i 's/\.WithContext(ctx)\././' "$file"
done
```

- [ ] **Step 2: Fix remaining method calls**

For each file, identify calls like:
- `logger.Infof("fmt", args...)` (no ctx) → `logger.Infof(ctx, "fmt", args...)`
- `logger.Info("msg", fields...)` → `logger.Info(ctx, "msg", fields...)`

Run:
```bash
grep -n "logger\.\(Info\|Error\|Warn\|Debug\|Fatal\|Infof\|Errorf\|Warnf\|Debugf\|Fatalf\)(" internal/app/auth/token_service.go
```

Manually update each call.

- [ ] **Step 3: Verify application layer builds**

Run:
```bash
go build ./internal/app/...
```

Expected: Success or compile errors pointing to remaining old-style calls.

- [ ] **Step 4: Fix any remaining errors**

For each compile error, add `ctx` as first param or reorder params.

- [ ] **Step 5: Commit application layer changes**

```bash
git add internal/app/
git commit -m "refactor: update application layer logfacade calls with ctx first param"
```

---

### Task 7: Update domain layer callers (2 files)

**Files to update:**
- `internal/domain/modeldesign/jsonschema_parser.go`
- `internal/domain/modelruntime/model_resolver.go`

- [ ] **Step 1: Update domain layer files**

```bash
cd /root/modelcraft_project/modelcraft-go

# Remove WithContext chaining
sed -i 's/\.WithContext(ctx)\././' internal/domain/modeldesign/jsonschema_parser.go
sed -i 's/\.WithContext(ctx)\././' internal/domain/modelruntime/model_resolver.go
```

- [ ] **Step 2: Fix remaining calls**

For each file:
```bash
grep -n "logger\.\(Info\|Error\|Warn\|Debug\)(" internal/domain/modeldesign/jsonschema_parser.go
```

Add `ctx` as first param.

- [ ] **Step 3: Verify domain builds**

Run:
```bash
go build ./internal/domain/...
```

Expected: Success.

- [ ] **Step 4: Commit domain layer changes**

```bash
git add internal/domain/
git commit -m "refactor: update domain layer logfacade calls with ctx first param"
```

---

### Task 8: Update GraphQL interface layer callers (10 files)

**Files to update:**
- `internal/interfaces/graphql/org/adapter/cluster_error_adapter.go`
- `internal/interfaces/graphql/org/adapter/project_error_adapter.go`
- `internal/interfaces/graphql/org/directives.go`
- `internal/interfaces/graphql/org/project.resolvers.go`
- `internal/interfaces/graphql/project/adapter/cluster_error_adapter.go`
- `internal/interfaces/graphql/project/adapter/enum_error_adapter.go`
- `internal/interfaces/graphql/project/adapter/fk_error_adapter.go`
- `internal/interfaces/graphql/project/adapter/model_error_adapter.go`
- `internal/interfaces/graphql/project/cluster.resolvers.go`
- `internal/interfaces/graphql/project/directives.go`

- [ ] **Step 1: Update all GraphQL resolver files**

```bash
cd /root/modelcraft_project/modelcraft-go

# Remove WithContext chaining in all graphql interface files
find internal/interfaces/graphql -name "*.go" -not -name "*_test.go" | while read file; do
  sed -i 's/\.WithContext(ctx)\././' "$file"
done
```

- [ ] **Step 2: Fix remaining calls**

For each resolver file, add `ctx` as first param to all log calls.

- [ ] **Step 3: Verify GraphQL layer builds**

Run:
```bash
go build ./internal/interfaces/graphql/...
```

Expected: Success.

- [ ] **Step 4: Commit GraphQL layer changes**

```bash
git add internal/interfaces/graphql/
git commit -m "refactor: update GraphQL interface layer logfacade calls with ctx first param"
```

---

### Task 9: Update HTTP interface layer callers (7 files)

**Files to update:**
- `internal/interfaces/http/handlers/auth/token_handler.go`
- `internal/interfaces/http/handlers/org/create_handler.go`
- `internal/interfaces/http/handlers/user/handler.go`
- `internal/interfaces/http/handlers/webhook/auth_provider_handler.go`
- `internal/interfaces/http/routes.go`
- `internal/interfaces/http/server.go`
- `internal/interfaces/runtime/handler.go`

- [ ] **Step 1: Update HTTP handler files**

```bash
cd /root/modelcraft_project/modelcraft-go

# Remove WithContext chaining in HTTP interface files
find internal/interfaces/http -name "*.go" -not -name "*_test.go" | while read file; do
  sed -i 's/\.WithContext(ctx)\././' "$file"
done

# Also update runtime
sed -i 's/\.WithContext(ctx)\././' internal/interfaces/runtime/handler.go
```

- [ ] **Step 2: Fix remaining calls**

For each handler file, add `ctx` as first param.

- [ ] **Step 3: Verify HTTP layer builds**

Run:
```bash
go build ./internal/interfaces/http/...
go build ./internal/interfaces/runtime/...
```

Expected: Success.

- [ ] **Step 4: Commit HTTP layer changes**

```bash
git add internal/interfaces/
git commit -m "refactor: update HTTP interface layer logfacade calls with ctx first param"
```

---

### Task 10: Update remaining package and cmd callers (4 files)

**Files to update:**
- `pkg/config/config.go`
- `pkg/logfacade/logger.go`
- `pkg/bizutils/utils.go`
- `cmd/server/main.go`

- [ ] **Step 1: Update pkg files**

```bash
cd /root/modelcraft_project/modelcraft-go

# Remove WithContext in pkg files
sed -i 's/\.WithContext(ctx)\././' pkg/config/config.go
sed -i 's/\.WithContext(ctx)\././' pkg/logfacade/logger.go
sed -i 's/\.WithContext(ctx)\././' pkg/bizutils/utils.go
sed -i 's/\.WithContext(ctx)\././' cmd/server/main.go
```

- [ ] **Step 2: Fix remaining calls**

For each file, add `ctx` as first param to log calls.

- [ ] **Step 3: Verify package builds**

Run:
```bash
go build ./pkg/...
go build ./cmd/...
```

Expected: Success.

- [ ] **Step 4: Commit package/cmd changes**

```bash
git add pkg/ cmd/
git commit -m "refactor: update pkg and cmd logfacade calls with ctx first param"
```

---

## Chunk 3: Verification and Cleanup

### Task 11: Full build and test

**Files:**
- Verify entire project builds and tests pass

- [ ] **Step 1: Run full build**

Run:
```bash
cd /root/modelcraft_project/modelcraft-go
go build ./...
```

Expected: Success with zero errors.

- [ ] **Step 2: Run all tests**

Run:
```bash
cd /root/modelcraft_project/modelcraft-go
go test ./... -v
```

Expected: All tests pass.

- [ ] **Step 3: Check for any remaining DeepInfof references**

Run:
```bash
grep -r "DeepInfof" /root/modelcraft_project/modelcraft-go --include="*.go" | grep -v "Binary\|vendor"
```

Expected: No results (or only in comments about removal).

- [ ] **Step 4: Verify WithContext is gone**

Run:
```bash
grep -r "\.WithContext(" /root/modelcraft_project/modelcraft-go --include="*.go" | grep -v "vendor" | grep -v "Binary"
```

Expected: No results (except possibly in git diff output).

- [ ] **Step 5: Final commit**

```bash
cd /root/modelcraft_project/modelcraft-go
git status
```

If any remaining changes:
```bash
git add -A
git commit -m "refactor: complete logfacade interface redesign - all callers updated"
```

---

## Testing Strategy

1. **Unit tests:** Existing tests in `pkg/logfacade/` verify interface behavior
2. **Integration:** Full build verifies no type mismatches across all ~51 caller files
3. **Functionality:** Existing log output behavior unchanged; only call-site syntax changes
4. **No new tests needed** — this is a mechanical refactor with zero behavior change

---

## Rollback Plan

If a step fails at any point:
1. All changes are committed frequently (per task), so you can `git reset HEAD~1` to undo the last commit
2. If multiple commits need reverting: `git log --oneline` to find the before-state, then `git reset <commit>`
3. Revert to last known-good state before logfacade changes: `git log --oneline | grep logfacade` to find first commit, then reset to before it

---

## Notes

- **Generated files:** `internal/interfaces/http/generated/server.gen.go` is auto-generated (do not edit by hand)
- **Pattern consistency:** All callers follow one pattern: `logger.Method(ctx, msg, fields...)`  or `logger.Method(ctx, format, args...)`
- **No WithContext chains:** All removed — context is always passed as first arg
- **No DeepInfof calls:** All removed — debug code should use `spew` directly or simple Printf
