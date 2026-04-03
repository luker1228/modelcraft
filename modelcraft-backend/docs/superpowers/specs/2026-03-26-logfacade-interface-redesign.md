# logfacade Interface Redesign

**Date:** 2026-03-26  
**Status:** Approved

---

## Problem

The current `Logger` interface has two overlapping logging styles and an awkward context-propagation mechanism:

1. **Dual formats** — both structured (`Info(msg, fields...)`) and formatted (`Infof(format, args...)`) exist as top-level methods. Callers must decide which style to use, creating inconsistency.
2. **Context via chain** — context is attached via `WithContext(ctx)`, which returns a new logger. Callers routinely write `logger.WithContext(ctx).Infof(...)`, adding noise to every log call.
3. **Dead weight** — `DeepInfof` (spew-based deep dump) is a debug utility with no production use case; it bloats the interface.

The goal is to reduce the interface to a single, predictable shape that eliminates the `WithContext` chaining pattern and removes unused methods.

---

## Design

### New Interface

```go
type Logger interface {
    // Structured logging — context is first arg, followed by message and optional fields
    Debug(ctx context.Context, msg string, fields ...Field)
    Info(ctx context.Context, msg string, fields ...Field)
    Warn(ctx context.Context, msg string, fields ...Field)
    Error(ctx context.Context, msg string, fields ...Field)
    Fatal(ctx context.Context, msg string, fields ...Field)

    // Formatted logging — context is first arg, followed by printf-style format and args
    Debugf(ctx context.Context, format string, args ...any)
    Infof(ctx context.Context, format string, args ...any)
    Warnf(ctx context.Context, format string, args ...any)
    Errorf(ctx context.Context, format string, args ...any)
    Fatalf(ctx context.Context, format string, args ...any)

    // With returns a new Logger with the given fields pre-attached to every subsequent call
    With(fields ...Field) Logger

    // Sync flushes any buffered log entries; call before program exit
    Sync() error
}
```

### What Changes

| Before | After |
|--------|-------|
| `logger.WithContext(ctx).Info("msg", fields...)` | `logger.Info(ctx, "msg", fields...)` |
| `logger.WithContext(ctx).Infof("fmt", args...)` | `logger.Infof(ctx, "fmt", args...)` |
| `logger.WithContext(ctx).With(fields...).Info("msg")` | `logger.With(fields...).Info(ctx, "msg")` |
| `logger.Infof("fmt", args...)` (no ctx) | `logger.Infof(ctx, "fmt", args...)` |
| `logger.DeepInfof("msg", obj)` | _(removed — use spew directly in test/debug code)_ |

### What Is Removed

- `WithContext(ctx) Logger` — eliminated; context is now passed at each call site
- `DeepInfof(msg string, args ...any)` — removed from the interface and implementation
- `Debugf` — added (was missing from the old interface; only `Infof/Warnf/Errorf` existed); implements standard fmt.Sprintf behavior
- `Fatalf` — added (was missing from the old interface); formats message then calls fatal-level logging (which exits)

### Context Extraction

The implementation (`zapLogger`) extracts `request_id` and other context fields inside each log method call, replacing the extraction that previously happened in `WithContext`. The behavior is identical — the extraction point just moves from logger construction to log emission.

**nil context handling:** If `ctx` is `nil`, context field extraction is skipped gracefully (no panic). This allows for use cases outside HTTP handlers (e.g., `main()`, background tasks, test code) where a context may not be available. Behavior is the same as extracting from an empty context.

### `With` Semantics

`With` is unchanged in behavior: it returns a child logger with the given fields attached to all subsequent calls. The child logger's log methods also accept `ctx` as the first argument.

```go
// Attach fixed fields to a component-scoped logger
repoLogger := logger.With(logfacade.String("component", "model-repo"))
repoLogger.Info(ctx, "fetching model", logfacade.String("id", id))
```

---

## Migration

### Scope

The interface change is breaking. Every caller of the old methods must be updated. Based on codebase analysis:

- ~165 calls to `Infof` (no ctx today → add ctx)
- ~107 calls to `Info` with fields (reorder ctx to front)
- ~20+ calls to `WithContext` (remove the chain, pass ctx to method)
- `DeepInfof` usages → remove or replace with direct `spew` calls

### Strategy

1. Update `Logger` interface in `interface.go`
2. Update `zapLogger` implementation in `zap_logger.go` — adapt all method signatures to accept `ctx` as first param; adjust context extraction to happen at call site instead of logger construction
3. Update all callers across the codebase — mechanical find-and-replace with per-call verification
4. Delete `WithContext` and `DeepInfof` from interface and implementation
5. Run `go build ./...` to identify remaining compile errors
6. Fix callers until build succeeds
7. Run test suite to verify no behavior changes

### Implementation Notes

- **Debugf / Fatalf:** Both use `fmt.Sprintf(format, args...)` internally, identical to Infof/Warnf/Errorf; Fatalf also exits the process after logging (same as Fatal)
- **Field merging:** The `With()` child logger merges its stored fields with context-extracted fields on each log call (no special ordering needed; zap handles this)
- **Global logger:** `logfacade.GetLogger(ctx)` is unchanged — it returns the default logger (optionally extracted from context). The returned logger is then called with `ctx` passed explicitly to each log method

### Caller Pattern Changes (Reference)

```go
// BEFORE
logger := logfacade.GetLogger(ctx)
logger.WithContext(ctx).Infof("created project: slug=%s", slug)
logger.WithContext(ctx).Error("failed", logfacade.Err(err), logfacade.Stack(err))

// AFTER
logger := logfacade.GetLogger(ctx)
logger.Infof(ctx, "created project: slug=%s", slug)
logger.Error(ctx, "failed", logfacade.Err(err), logfacade.Stack(err))
```

```go
// BEFORE
logfacade.GetLogger(ctx).WithContext(ctx).With(logfacade.String("mod", "auth")).Info("login")

// AFTER
logfacade.GetLogger(ctx).With(logfacade.String("mod", "auth")).Info(ctx, "login")
```

---

## Non-Goals

- No change to `Field`, field constructors (`String`, `Int`, `Err`, `Stack`, etc.), or `Config`
- No change to `GetLogger`, `SetDefault`, `WithLogger` global helpers
- No change to the underlying zap implementation beyond adapting method signatures
- No new log levels

---

## Testing

- All log method signatures are changed, so all call sites in tests must also be updated (in `*_test.go` files and test helpers)
- Existing unit tests in `interface_test.go` and `example_test.go` are updated to the new signatures
- `go build ./...` must pass with zero errors after migration
- No behavior change in log output — only call-site syntax changes
