---
name: schema-sync-cascade
description: >-
  Use this skill when GraphQL schema or SQLc queries change — including renaming types/inputs/mutations,
  adding/removing fields, changing SQL query signatures, or modifying sqlc YAML config. Triggers on
  phrases like "rename this GraphQL type", "add a field to the mutation", "change this SQL query",
  "update the schema", "modify the input", or any edit to .graphql or .sql files under api/graph/{org,project}/schema/
  or db/queries/. Also use when the user edits gqlgen.org.yml, gqlgen.project.yml, sqlc.yaml, or justfile
  generate targets. When in doubt, use this skill — it prevents stale generated code and broken references.
---

# Schema Sync Cascade

When GraphQL schema or SQLc queries are modified, this skill ensures all references stay in sync and
generated code is fresh. The core principle: **source schema changes must cascade to every consumer**.

## Project Structure

```
modelcraft-backend/                           # Backend
  api/graph/{org,project}/schema/*.graphql    # GraphQL source schemas (edit here)
  internal/interfaces/graphql/*/*.resolvers.go # Resolver implementations
  internal/interfaces/graphql/*/generated/*.go # gqlgen output (auto-generated)
  db/queries/*.sql                            # SQLc query source files
  internal/infrastructure/dbgen/*.go          # sqlc output (auto-generated)
  sqlc.yaml                                   # sqlc config

modelcraft-front/                             # Frontend
  src/graphql/mutations/*.ts                  # GraphQL mutation documents
  src/graphql/queries/*.ts                    # GraphQL query documents
  src/types/index.ts                          # TypeScript type definitions
```

## Workflow

### Step 1: Identify What Changed

Read the user's instruction carefully. Determine:
- **GraphQL**: Which types, inputs, mutations, or queries are being renamed/added/removed?
- **SQLc**: Which SQL queries are changing? Are column names, table names, or return types affected?
- Extract the old name and new name (for renames), or the new structure (for additions).

### Step 2: Find All References (GraphQL)

Search across both codebases in parallel. Do NOT modify generated directories — they are regenerated.

**Backend (`modelcraft-backend/`):**
1. `api/graph/{org,project}/schema/*.graphql` — the source schema itself
2. `internal/interfaces/graphql/*/*.resolvers.go` — resolver method names and `generated.*` type references
3. `tests/` — any test files referencing the old names

**Frontend (`modelcraft-front/`):**
1. `src/graphql/mutations/*.ts` — mutation document strings (`mutation NameName(...) { oldName(...) }`)
2. `src/graphql/queries/*.ts` — query document strings
3. `src/types/index.ts` — TypeScript interfaces matching GraphQL types
4. `src/hooks/*.{ts,tsx}` — custom hooks that import and call mutations/queries
5. `src/app/**/*.tsx` — page components that directly use mutations or read response fields

Use Grep to find all occurrences. Report findings to the user before editing.

### Step 3: Find All References (SQLc)

**Backend only:**
1. `db/queries/*.sql` — the SQL query files
2. `internal/infrastructure/dbgen/*.go` — do NOT edit (auto-generated)
3. Any Go files in `internal/` that call generated sqlc functions (`.Query()` or `.Exec()` methods)
4. Repository interfaces and implementations that wrap sqlc calls

### Step 4: Apply Changes

Edit all non-generated files found in Steps 2-3. For renames:
- GraphQL schema: type/input/mutation name
- Resolver Go code: method name + `generated.NewName` type references
- Frontend mutation strings: the field name in the GraphQL document
- Frontend types: interface name and any variable type annotations
- Frontend hooks/pages: response data access paths (`data?.oldName` -> `data?.newName`)
- SQL query names: `-- name: OldName` -> `-- name: NewName`
- Go callers: `q.OldName()` -> `q.NewName()`

### Step 5: Regenerate Code

Always regenerate after schema changes. Run from `modelcraft-backend/`:

```bash
# After GraphQL schema changes:
just generate-gql

# After SQLc query changes:
just generate-sqlc

# After both (can run in parallel):
just generate-gql && just generate-sqlc
```

### Step 6: Verify

After regeneration:
1. In `modelcraft-backend/`, run `go build ./...` to check for compilation errors
2. If build fails, investigate — the rename likely missed a reference. Go back to Step 2.
3. Do NOT hand-edit generated files under `generated/` or `dbgen/`.

## Key Rules

- **Never edit generated code.** Always modify source files and regenerate.
- **Search both codebases.** GraphQL changes affect frontend too. SQLc changes are backend-only.
- **Regenerate before verifying.** Stale generated code will cause false compilation errors.
- **Response field paths matter.** When renaming a mutation from `updateModel` to `updateModelMeta`,
  every `data?.updateModel` in the frontend must become `data?.updateModelMeta`.
