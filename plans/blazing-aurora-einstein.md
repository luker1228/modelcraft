# Plan: Unify Runtime GraphQL Endpoint

## Goal

Change the runtime GraphQL endpoint from the current pattern to a unified RESTful pattern:

**Before:**
```
POST /org/{orgName}/project/{projectSlug}/db/{database}/model/{modelName}
GET  /org/{orgName}/project/{projectSlug}/db/{database}/model/{modelName}  (Playground)
```

**After:**
```
POST /org/{orgName}/project/{slug}/db/{db}/model/{model}/graphql
GET  /org/{orgName}/project/{slug}/db/{db}/model/{model}/graphql  (Playground)
```

Changes:
1. Shorten URL param names: `{projectSlug}` → `{slug}`, `{database}` → `{db}`, `{modelName}` → `{model}`
2. Add `/graphql` suffix to the endpoint (matches design-time convention of `/graphql` as terminator)

## Files to Modify

### Backend (3 files)

1. **`modelcraft-go/internal/interfaces/http/routes.go`** (line 424)
   - Change route path and comments
   - Route: `/org/{orgName}/project/{projectSlug}/db/{database}/model/{modelName}` → `/org/{orgName}/project/{slug}/db/{db}/model/{model}/graphql`

2. **`modelcraft-go/internal/interfaces/runtime/handler.go`**
   - `HandlePlayground` (line 42-57): Update `chi.URLParam` keys and playground endpoint construction
   - `HandleQuery` (line 71-86): Update `chi.URLParam` keys and error messages

3. **`modelcraft-go/cmd/server/main.go`** (line 195, 210)
   - Update startup log to reflect new endpoint pattern

### Frontend (5 files)

4. **`modelcraft-front/src/bff/apollo/clients.ts`** (line 19)
   - `buildRuntimeEndpoint`: update URL path and param names

5. **`modelcraft-front/src/components/model-editor/DynamicModelTable.tsx`** — already uses `buildRuntimeEndpoint`, no direct URL construction (verify only)

6. **`modelcraft-front/src/components/cms/FormRenderer.tsx`** — same as above (verify only)

7. **`modelcraft-front/src/web/components/model-editor/DynamicModelTable.tsx`** — same as above (verify only)

8. **`modelcraft-front/src/web/components/cms/FormRenderer.tsx`** — same as above (verify only)

## Implementation Steps

### Step 1: Backend route registration
- In `routes.go:424`: Change route path to `/org/{orgName}/project/{slug}/db/{db}/model/{model}/graphql`
- Update function comments (lines 408-412)

### Step 2: Backend handler param extraction
- In `handler.go`: Change all `chi.URLParam` calls:
  - `"projectSlug"` → `"slug"`
  - `"database"` → `"db"`
  - `"modelName"` → `"model"`
- Update playground endpoint in `HandlePlayground` to match new URL pattern
- Update error messages

### Step 3: Backend startup log
- In `main.go:195,210`: Update log messages with new endpoint pattern

### Step 4: Frontend URL builder
- In `clients.ts:19`: Update `buildRuntimeEndpoint` to use new path `/org/${orgName}/project/${projectSlug}/db/${databaseName}/model/${modelName}/graphql`

### Step 5: Verify frontend consumers
- Confirm that DynamicModelTable.tsx and FormRenderer.tsx (both `src/` and `src/web/`) all use `buildRuntimeEndpoint` — no direct URL construction to change.

## Verification
- `just lint` (backend)
- `cd modelcraft-front && npm run lint` (frontend)
