# Checklist Review: `sql_modelruntime_repository.go`

> File: `modelcraft-backend/internal/infrastructure/repository/sql_modelruntime_repository.go`
> Review date: 2026-04-15

---

## Summary

The file is 141 lines implementing `modelruntime.ModelRepository` via sqlc. It exposes two public methods (`GetByID`, `GetByName`) and one private helper (`getFields`). The code is generally clean and idiomatic, but the review surfaces **3 confirmed bug patterns** and **2 design concerns**.

---

## Checklist Items

### ✅ PASS — Compile-time interface guard present
```go
var _ modelruntime.ModelRepository = (*SqlModelRuntimeRepository)(nil)
```
Drift between interface and implementation is caught at compile time. Good.

---

### ✅ PASS — Error wrapping via `bizerrors`
`getFields` wraps sub-errors with context strings (`"getFields: fetch fields"`, etc.). Callers can trace error origin without scanning call stacks.

---

### ✅ PASS — SafeQuerier wraps raw SQL errors
`NewSqlModelRuntimeRepository` stores `dbgenwrap.NewSafeQuerier(q)` so that all `sql.ErrNoRows`, duplicate key, and connection errors are automatically converted to typed `RepositoryError` values before bubbling up.

---

### ❌ BUG — `GetByID` / `GetByName` do NOT check for "not found"

**Location:** Lines 25–28, 44–50

```go
row, err := r.q.GetModelByID(ctx, id)
if err != nil {
    return nil, err
}
// ← no "row.ID == """ guard here
```

The companion `SqlModelDesignRepository.GetByID` (lines 344–351 of `sql_modeldesign_repository.go`) explicitly checks `row.ID == ""` after a successful query call and returns `shared.NewNotFoundError(...)` in that case. The runtime repository skips this check entirely.

**Impact:** If the underlying DB driver returns a zero-value `dbgen.Model{}` instead of `sql.ErrNoRows` (e.g., on some MySQL connector versions when using `QueryRowContext` without scanning error propagation), the caller receives a non-nil `RuntimeModel` with all empty fields (`Name`, `DatabaseName`, `OrgName` all `""`). This propagates silently into GraphQL schema generation and may produce an anonymous, unconfigured schema rather than a clear 404-style error.

**Fix:**
```go
if row.ID == "" {
    return nil, shared.NewNotFoundError("model not found by id: " + id)
}
```
Apply the same pattern used in `SqlModelDesignRepository`.

---

### ❌ BUG — Silent enum miss: field attached to a missing enum definition goes silently unset

**Location:** Lines 110–114

```go
if fieldRow.EnumName.Valid && fieldRow.EnumName.String != "" {
    if ed, ok := enumMap[fieldRow.EnumName.String]; ok {
        fd.Enum = ed
    }
    // ← no `else` — no error if enum is absent from enumMap
}
```

`GetEnumsByNames` is scoped to `org_name + project_slug + names IN (...)`. If an enum was deleted but the `EnumName` column on the field was not cleared (e.g., due to a failed migration or race condition), the enum lookup silently fails. `fd.Enum` stays `nil`.

**Impact:** Downstream code in `modelruntime` that assumes `fd.Enum != nil` for `FormatEnum` fields will either panic or return incorrect GraphQL responses. Since `getGraphqlTypeBy` maps `FormatEnum → graphql.String`, runtime schema construction may succeed but enum validation/label resolution fails without any observable error at the repo layer.

**Fix:** Add an explicit error (or at least a structured log) when an expected enum is absent from `enumMap`:
```go
if ed, ok := enumMap[fieldRow.EnumName.String]; ok {
    fd.Enum = ed
} else {
    return nil, bizerrors.Errorf("getFields: enum %q referenced by field %q not found", fieldRow.EnumName.String, fieldRow.Name)
}
```

---

### ❌ BUG — Enum deduplication not applied; `GetEnumsByNames` called with duplicate names

**Location:** Lines 73–79

```go
enumNames := make([]string, 0)
for _, f := range fieldRows {
    if f.EnumName.Valid && f.EnumName.String != "" {
        enumNames = append(enumNames, f.EnumName.String)
    }
}
```

If two fields share the same enum (e.g., `status_enum`), the name appears twice in `enumNames`. The SQL `IN (?)` query is then sent with duplicates, which is benign for correctness (MySQL deduplicates `IN` conditions internally) but wastes a slot in the query and can confuse query-plan analysis or logging.

More importantly: if the project ever switches to a database that enforces `IN` distinctness (or if sqlc changes the slice-expansion behavior), this becomes a subtle SQL error.

**Fix:**
```go
seen := make(map[string]struct{})
for _, f := range fieldRows {
    if f.EnumName.Valid && f.EnumName.String != "" {
        if _, ok := seen[f.EnumName.String]; !ok {
            seen[f.EnumName.String] = struct{}{}
            enumNames = append(enumNames, f.EnumName.String)
        }
    }
}
```

---

### ⚠️ DESIGN CONCERN — `DbgenModelToRuntimeModel` silently discards `row.Description` when `Valid == false`

**Location:** Lines 133–136

```go
Description:  row.Description.String,
```

`sql.NullString.String` is `""` when `Valid == false`. This is consistent with the existing test (`TestDbgenModelToRuntimeModel / null description maps to empty string`). However, `ModelToDomain` in the design repository also does the same (`Description: row.Description.String`). The pattern is project-wide, so this is not a bug in isolation — but worth noting because a future migration that makes Description `NOT NULL DEFAULT ''` would be the correct fix to close the semantic gap between "null" and "empty string" at the storage layer.

No action required now; document the invariant if it is not already noted.

---

### ⚠️ DESIGN CONCERN — `getFields` is not a method on `FindManyIn` path; N+1 is not addressed here

**Location:** `getFields` is called once per `GetByID` / `GetByName`.

The `modelruntime` domain has a `FindManyIn` interface on `graphql_repository.go` specifically to batch-load relations and avoid N+1. `sql_modelruntime_repository.go` implements a *schema-loading* path (not a data-fetching path), so a single-model load with 2 queries (`GetFieldsByModelID` + `GetEnumsByNames`) is appropriate and not a bug. However, if `GetByName` is ever called in a loop (e.g., loading all models for a project's schema), the per-model `getFields` call will cause N+1 on the enum fetch. There is no batch equivalent today.

This is a scalability concern, not a correctness bug. Track it if model count is expected to grow.

---

## Findings Summary Table

| # | Severity | Category | Description | Line(s) |
|---|----------|----------|-------------|---------|
| 1 | **Bug** | Error handling | `GetByID`/`GetByName` skip "not found" guard after `row.ID == ""` | 25–28, 44–50 |
| 2 | **Bug** | Silent failure | Missing enum silently leaves `fd.Enum = nil` with no error | 110–114 |
| 3 | **Bug** | Data integrity | Enum names not deduplicated before `GetEnumsByNames` SQL call | 73–79 |
| 4 | Warning | Design | `NullString` → empty string conflation is consistent but lossy | 133 |
| 5 | Warning | Scalability | No batch path for `getFields`; N+1 risk if called in a loop | 64–120 |
