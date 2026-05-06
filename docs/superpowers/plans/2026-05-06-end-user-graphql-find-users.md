# End-User findUsers / me GraphQL Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `findUsers` and `me` queries to the project GraphQL schema, accessible from both the tenant developer endpoint and the end-user endpoint, with the `@hasPermission` directive enforcing a default-deny policy for end-users (only explicitly `allowEndUser: true` operations are accessible).

**Architecture:**
- `@hasPermission` directive gains a new `allowEndUser: Boolean! = false` argument; end-user callers hitting an operation where `allowEndUser = false` get a `PERMISSION_DENIED` error.
- `findUsers` and `me` are added to `api/graph/project/schema/end_user.graphql` with `@hasPermission(action: "end-user:read", allowEndUser: true)`.
- Two new resolvers call `MetaUserAppService` (already implemented in `internal/app/enduser/meta_user_app_service.go`). The `Resolver` struct gains a `MetaUserAppService` field.
- No gateway or route changes needed — `/graphql/end-user/org/{orgName}/project/{projectSlug}` already exists, strips `/end-user/`, and injects `X-User-Type: end_user`.

**Tech Stack:** Go, gqlgen v0.17.83, chi router, existing `MetaUserAppService`.

---

## File Map

| Action | File | Purpose |
|--------|------|---------|
| Modify | `modelcraft-backend/api/graph/project/schema/base.graphql` | Add `allowEndUser` argument to `@hasPermission` directive declaration |
| Modify | `modelcraft-backend/api/graph/project/schema/end_user.graphql` | Add `Tuser`, `TuserWhereInput`, `UserFindManyResult`, `UserFindOneResult` types; add `findUsers` and `me` queries |
| Modify | `modelcraft-backend/internal/interfaces/graphql/project/directives.go` | Update `HasPermission` method signature, flip end-user default-deny logic |
| Modify | `modelcraft-backend/internal/interfaces/graphql/project/resolver.go` | Add `MetaUserAppService *appEnduser.MetaUserAppService` field |
| Create | `modelcraft-backend/internal/interfaces/graphql/project/meta_user.resolvers.go` | Implement `FindUsers` and `Me` resolvers |
| Modify | `modelcraft-backend/internal/interfaces/http/routes.go` | Wire `MetaUserAppService` into `projectResolver` |
| Run | `just generate-gql` | Regenerate gqlgen types and interfaces after schema changes |

---

## Task 1: Update `@hasPermission` directive declaration in schema

**Files:**
- Modify: `modelcraft-backend/api/graph/project/schema/base.graphql`

- [ ] **Step 1: Edit directive declaration**

Replace line 7 in `modelcraft-backend/api/graph/project/schema/base.graphql`:

```graphql
# Before
directive @hasPermission(action: String!) on FIELD_DEFINITION

# After
directive @hasPermission(action: String!, allowEndUser: Boolean! = false) on FIELD_DEFINITION
```

- [ ] **Step 2: Commit**

```bash
cd modelcraft-backend
git add api/graph/project/schema/base.graphql
git commit -m "feat(graphql): add allowEndUser arg to @hasPermission directive declaration"
```

---

## Task 2: Add `findUsers` / `me` types and queries to schema

**Files:**
- Modify: `modelcraft-backend/api/graph/project/schema/end_user.graphql`

- [ ] **Step 1: Append new types and queries to end_user.graphql**

Add the following block at the end of `modelcraft-backend/api/graph/project/schema/end_user.graphql` (after the existing `extend type Query` block):

```graphql
# ============================================
# System User Query Types (findUsers / me)
# Used by EndUserRef field selectors and tenant management.
# Tenant callers: developer JWT via /graphql/org/{orgName}/project/{projectSlug}/
# End-user callers: end-user JWT via /graphql/end-user/org/{orgName}/project/{projectSlug}/
#
# Note: "User" (not "Tuser") is used here because this is a fixed system model name.
# The Tuser prefix exists only for custom models whose names may start with a digit,
# making them invalid as GraphQL type names. "user" is always a valid identifier.
# ============================================

# User is the public projection of an end-user (no tenant fields).
type User {
  id: ID!
  username: String!
  createdAt: Time!
}

# ---- Filter inputs ----

input StringFilter {
  eq: String
  contains: String
  startsWith: String
  in: [String!]
}

input IDFilter {
  eq: ID
  in: [ID!]
}

input DateTimeFilter {
  eq: String
  gte: String
  lte: String
}

input UserWhereInput {
  id: IDFilter
  username: StringFilter
  createdAt: DateTimeFilter
}

# ---- Result types ----

type UserFindManyResult {
  items: [User!]!
  totalCount: Int
  reqId: String!
}

type UserFindOneResult {
  item: User
  reqId: String!
}

# ---- Queries ----

extend type Query {
  # findUsers: list users with optional filtering and pagination.
  # Accessible to both tenant developers and end-users (allowEndUser: true).
  # Pagination: skip default 0 max 1000; take default 20 max 50.
  findUsers(
    where: UserWhereInput
    skip: Int
    take: Int
  ): UserFindManyResult! @hasPermission(action: "end-user:read", allowEndUser: true)

  # me: returns the currently authenticated caller's user profile.
  # Accessible to end-users only (tenant callers have no end-user identity).
  me: UserFindOneResult! @hasPermission(action: "end-user:read", allowEndUser: true)
}
```

- [ ] **Step 2: Commit**

```bash
cd modelcraft-backend
git add api/graph/project/schema/end_user.graphql
git commit -m "feat(graphql): add Tuser types and findUsers/me queries to project schema"
```

---

## Task 3: Regenerate gqlgen code

**Files:**
- Run: `just generate-gql`

After schema changes, gqlgen must regenerate `internal/interfaces/graphql/project/generated/`.

- [ ] **Step 1: Run code generation**

```bash
cd modelcraft-backend
just generate-gql
```

Expected: no errors. New types `Tuser`, `TuserWhereInput`, `StringFilter`, `IDFilter`, `DateTimeFilter`, `UserFindManyResult`, `UserFindOneResult` appear in `internal/interfaces/graphql/project/generated/models_gen.go`. New resolver stubs for `FindUsers` and `Me` appear in `internal/interfaces/graphql/project/end_user.resolvers.go` (or a new file if gqlgen creates one).

- [ ] **Step 2: Verify build still compiles**

```bash
cd modelcraft-backend
go build ./...
```

Expected: compile errors for unimplemented `FindUsers` and `Me` resolver methods (this is fine — we implement them in Task 5). If there are other errors (e.g., signature mismatch for `HasPermission`), fix them before proceeding.

---

## Task 4: Update `HasPermission` directive to default-deny end-users

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/directives.go`

The gqlgen-generated `DirectiveRoot.HasPermission` signature now includes `allowEndUser bool`. We need to update the implementation to match.

- [ ] **Step 1: Update the HasPermission method signature and logic**

Replace the entire `HasPermission` method in `modelcraft-backend/internal/interfaces/graphql/project/directives.go`:

```go
// HasPermission checks if the user has the required permission before executing the field resolver.
// End-user callers are denied by default; only operations with allowEndUser=true are accessible.
func (d *HasPermissionDirective) HasPermission(
	ctx context.Context,
	obj interface{},
	next graphql.Resolver,
	action string,
	allowEndUser bool,
) (interface{}, error) {
	logger := logfacade.GetLogger(ctx)

	// Validate authentication and organization context
	userID, orgName, err := d.validateContext(ctx, logger)
	if err != nil {
		return nil, err
	}

	// Validate action format
	if action == "" {
		logger.Errorf(ctx, "@hasPermission directive: empty action parameter")
		return nil, newGQLError("Invalid permission directive configuration", "INVALID_DIRECTIVE")
	}

	// End-user callers: default-deny. Only operations explicitly marked allowEndUser=true are accessible.
	if ctxutils.IsEndUser(ctx) {
		if !allowEndUser {
			logger.Infof(ctx, "@hasPermission directive: end-user access denied (allowEndUser=false) user=%s action=%s",
				userID, action)
			return nil, newPermissionDeniedError(action, userID, orgName, "end-user-default-deny")
		}
		logger.Infof(ctx, "@hasPermission directive: end-user read access granted (allowEndUser=true) user=%s action=%s",
			userID, action)
		return next(ctx)
	}

	// Developer callers: OPTIMIZATION: Try to get permissions from JWT context first (no DB query)
	if permissions, err := ctxutils.GetPermissionsFromContext(ctx); err == nil && permissions != nil {
		logger.Infof(
			ctx, "@hasPermission directive: using permissions from JWT context (source: jwt, count: %d)",
			len(permissions),
		)
		return d.checkContextPermission(ctx, next, logger, userID, orgName, action, permissions)
	}

	// FALLBACK: Query database if permissions not in context
	return d.checkDatabasePermission(ctx, next, logger, userID, orgName, action)
}
```

Note: `isReadAction` is no longer used; remove it from the file:

```go
// DELETE this function entirely:
// func isReadAction(action string) bool {
//     return strings.HasSuffix(action, ":read") || strings.HasSuffix(action, ":list")
// }
```

- [ ] **Step 2: Remove the `strings` import if now unused**

Check that `strings` is still imported elsewhere in the file. If `isReadAction` was the only user of `strings`, remove it from the import block:

```go
import (
    "context"
    "fmt"
    "modelcraft/internal/middleware"
    "modelcraft/pkg/ctxutils"
    "modelcraft/pkg/logfacade"
    // Remove "strings" if isReadAction was its only user

    "github.com/99designs/gqlgen/graphql"
    "github.com/vektah/gqlparser/v2/gqlerror"

    appPermission "modelcraft/internal/app/permission"
)
```

- [ ] **Step 3: Verify build**

```bash
cd modelcraft-backend
go build ./internal/interfaces/graphql/project/...
```

Expected: SUCCESS.

- [ ] **Step 4: Commit**

```bash
cd modelcraft-backend
git add internal/interfaces/graphql/project/directives.go
git commit -m "feat(graphql): directive default-deny end-users unless allowEndUser=true"
```

---

## Task 5: Add `MetaUserAppService` to `Resolver` struct

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/resolver.go`

- [ ] **Step 1: Add the field**

In `modelcraft-backend/internal/interfaces/graphql/project/resolver.go`, add `MetaUserAppService` to the `Resolver` struct:

```go
// End-User
EndUserMgmtAppService *appEnduser.EndUserManagementAppService
MetaUserAppService    *appEnduser.MetaUserAppService   // ← add this line
```

The import `appEnduser "modelcraft/internal/app/enduser"` is already present.

- [ ] **Step 2: Wire it in routes.go**

In `modelcraft-backend/internal/interfaces/http/routes.go`, find the `projectResolver` construction inside `SetupProjectGraphQLRoutesOnChi` (around line 464) and add:

```go
projectResolver := &projectgraphql.Resolver{
    // ... existing fields ...
    EndUserMgmtAppService:    handlers.EndUserMgmtAppService,
    MetaUserAppService:       appEnduser.NewMetaUserAppService(handlers.PrivateDBManager), // ← add
    // ... existing fields ...
}
```

`handlers.PrivateDBManager` is already present in `DesignHandlers` (it's used for `PrivateDBManager` field in the resolver). `NewMetaUserAppService` is defined in `internal/app/enduser/meta_user_app_service.go`.

The import `appEnduser "modelcraft/internal/app/enduser"` is already present in `routes.go`.

- [ ] **Step 3: Verify build**

```bash
cd modelcraft-backend
go build ./...
```

Expected: SUCCESS (resolver stubs from gqlgen not yet implemented → compile error for `FindUsers`/`Me`). If there are missing method errors, proceed to Task 6.

- [ ] **Step 4: Commit**

```bash
cd modelcraft-backend
git add internal/interfaces/graphql/project/resolver.go internal/interfaces/http/routes.go
git commit -m "feat(graphql): wire MetaUserAppService into project resolver"
```

---

## Task 6: Implement `FindUsers` and `Me` resolvers

**Files:**
- Create: `modelcraft-backend/internal/interfaces/graphql/project/meta_user.resolvers.go`

> **Note:** gqlgen may have already created stub implementations in `end_user.resolvers.go` during Task 3. If so, move/replace those stubs with the code below in a new dedicated file. Do NOT edit gqlgen-generated files directly — create `meta_user.resolvers.go` as a hand-written file in the same package.

- [ ] **Step 1: Create `meta_user.resolvers.go`**

```go
package projectgraphql

import (
	"context"
	"time"

	appEnduser "modelcraft/internal/app/enduser"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"errors"
)

// FindUsers is the resolver for the findUsers field.
// Accessible to both tenant developers and end-users (allowEndUser=true on the schema).
// Pagination: skip default 0 max 1000, take default 20 max 50 (enforced by MetaUserAppService).
func (r *queryResolver) FindUsers(ctx context.Context, where *generated.UserWhereInput, skip *int, take *int) (*generated.UserFindManyResult, error) {
	reqID := ctxutils.GetRequestID(ctx)

	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil || orgName == "" {
		return nil, newGQLError("organization context required", "MISSING_ORGANIZATION")
	}

	cmd := appEnduser.MetaUserFindManyCommand{
		OrgName: orgName,
	}
	if skip != nil {
		cmd.Skip = *skip
	}
	if take != nil {
		cmd.Take = *take
	}
	if where != nil {
		cmd.Where = convertUserWhereInput(where)
	}

	result, err := r.MetaUserAppService.FindMany(ctx, cmd)
	if err != nil {
		var bizErr *bizerrors.BusinessError
		if errors.As(err, &bizErr) {
			return nil, newGQLError(bizErr.Message, string(bizErr.Code))
		}
		return nil, err
	}

	items := make([]*generated.User, 0, len(result.Items))
	for _, dto := range result.Items {
		items = append(items, &generated.User{
			ID:        dto.ID,
			Username:  dto.Username,
			CreatedAt: dto.CreatedAt,
		})
	}

	return &generated.UserFindManyResult{
		Items: items,
		ReqID: reqID,
	}, nil
}

// Me is the resolver for the me field.
// Returns the currently authenticated end-user's profile.
// For tenant developers: returns an error (no end-user identity in context).
func (r *queryResolver) Me(ctx context.Context) (*generated.UserFindOneResult, error) {
	reqID := ctxutils.GetRequestID(ctx)

	if !ctxutils.IsEndUser(ctx) {
		return nil, newGQLError("me is only available for end-user callers", "INVALID_CALLER")
	}

	dto, err := r.MetaUserAppService.GetMe(ctx)
	if err != nil {
		var bizErr *bizerrors.BusinessError
		if errors.As(err, &bizErr) {
			return nil, newGQLError(bizErr.Message, string(bizErr.Code))
		}
		return nil, err
	}
	if dto == nil {
		return &generated.UserFindOneResult{ReqID: reqID}, nil
	}

	return &generated.UserFindOneResult{
		Item:  &generated.User{ID: dto.ID, Username: dto.Username, CreatedAt: dto.CreatedAt},
		ReqID: reqID,
	}, nil
}

// convertUserWhereInput maps the generated GraphQL where input to the app-layer filter type.
func convertUserWhereInput(w *generated.UserWhereInput) *appEnduser.MetaUserFindManyFilter {
	if w == nil {
		return nil
	}
	f := &appEnduser.MetaUserFindManyFilter{}

	if w.ID != nil {
		f.IDEq = w.ID.Eq
		if len(w.ID.In) > 0 {
			f.IDIn = w.ID.In
		}
	}
	if w.Username != nil {
		f.UsernameEq = w.Username.Eq
		f.UsernameContains = w.Username.Contains
		f.UsernameStartsWith = w.Username.StartsWith
		if len(w.Username.In) > 0 {
			f.UsernameIn = w.Username.In
		}
	}
	if w.CreatedAt != nil {
		f.CreatedAtEq = parseTimePtr(w.CreatedAt.Eq)
		f.CreatedAtGte = parseTimePtr(w.CreatedAt.Gte)
		f.CreatedAtLte = parseTimePtr(w.CreatedAt.Lte)
	}
	return f
}

// parseTimePtr parses an ISO-8601 string pointer to a *time.Time.
// Returns nil if the input is nil or unparseable.
func parseTimePtr(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}
```

> **Important:** `generated.User.CreatedAt` will be of type `time.Time` (gqlgen maps the `Time` scalar to `time.Time`). Confirm the exact type after running `just generate-gql` in Task 3 by checking `internal/interfaces/graphql/project/generated/models_gen.go`.

- [ ] **Step 2: Check generated stub conflicts**

After `just generate-gql`, gqlgen may have added stub implementations for `FindUsers` / `Me` to `end_user.resolvers.go`. If so, **delete those stubs** from `end_user.resolvers.go` (they panic with "not implemented") so the hand-written file is the only implementation.

- [ ] **Step 3: Verify build**

```bash
cd modelcraft-backend
go build ./...
```

Expected: SUCCESS with no errors.

- [ ] **Step 4: Run linter**

```bash
cd modelcraft-backend
just lint
```

Expected: no new lint errors. If there are import issues (`time` unused, etc.), fix them.

- [ ] **Step 5: Commit**

```bash
cd modelcraft-backend
git add internal/interfaces/graphql/project/meta_user.resolvers.go internal/interfaces/graphql/project/end_user.resolvers.go
git commit -m "feat(graphql): implement FindUsers and Me resolvers via MetaUserAppService"
```

---

## Task 7: Manual smoke test

- [ ] **Step 1: Start the backend**

```bash
cd modelcraft-backend
just run force=true
```

- [ ] **Step 2: Test `findUsers` as a tenant developer**

```bash
curl -s -X POST http://localhost:8080/graphql/org/lukeid/project/first/ \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <your-developer-user-id>" \
  -d '{"query": "{ findUsers(take: 5) { items { id username } reqId } }"}'
```

Expected: `{ "data": { "findUsers": { "items": [...], "reqId": "..." } } }`

- [ ] **Step 3: Test `findUsers` as an end-user (via gateway)**

```bash
# First obtain an end-user token via /api/end-user/auth/login
TOKEN="<end-user-jwt>"
curl -s -X POST http://localhost:3000/graphql/end-user/org/lukeid/project/first \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "{ findUsers(take: 5) { items { id username } reqId } }"}'
```

Expected: `{ "data": { "findUsers": { "items": [...], "reqId": "..." } } }`

- [ ] **Step 4: Verify `listProjectEndUsers` is blocked for end-users**

```bash
TOKEN="<end-user-jwt>"
curl -s -X POST http://localhost:3000/graphql/end-user/org/lukeid/project/first \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "{ listProjectEndUsers { connection { nodes { id } } } }"}'
```

Expected: `{ "errors": [{ "message": "Permission denied...", "extensions": { "code": "PERMISSION_DENIED" } }] }`

- [ ] **Step 5: Test `me` as an end-user**

```bash
TOKEN="<end-user-jwt>"
curl -s -X POST http://localhost:3000/graphql/end-user/org/lukeid/project/first \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "{ me { item { id username } reqId } }"}'
```

Expected: returns the current end-user's id and username.

- [ ] **Step 6: Commit (if any fixups needed)**

```bash
cd modelcraft-backend
git add -A
git commit -m "fix(graphql): smoke test fixups for findUsers/me"
```

---

## Self-Review Notes

1. **Spec coverage:**
   - ✅ `allowEndUser` directive argument: Task 1 + Task 4
   - ✅ `findUsers` / `me` schema types: Task 2
   - ✅ gqlgen regeneration: Task 3
   - ✅ Default-deny logic for end-users: Task 4
   - ✅ Resolver implementations using `MetaUserAppService`: Task 5 + Task 6
   - ✅ Route wiring (no change needed — gateway already rewrites path): Task 5 note
   - ✅ Smoke tests covering tenant + end-user + blocked case: Task 7

2. **Type consistency:**
   - `MetaUserFindManyCommand`, `MetaUserFindManyFilter`, `MetaUserDTO` — all defined in `internal/app/enduser/commands.go`
   - `MetaUserAppService.FindMany` / `GetMe` — defined in `internal/app/enduser/meta_user_app_service.go`
   - `generated.User`, `generated.UserFindManyResult`, `generated.UserFindOneResult`, `generated.UserWhereInput` — generated by gqlgen in Task 3

3. **Potential issue:** `generated.IDFilter.In` and `generated.StringFilter.In` field types — after gqlgen runs, confirm field names match (`In []string` vs `In []*string`). The `convertUserWhereInput` function uses `w.ID.In` directly; adjust if the generated type uses pointers.
