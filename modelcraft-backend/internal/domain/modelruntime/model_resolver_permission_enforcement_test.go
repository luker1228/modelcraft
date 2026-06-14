package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── helper: fullCapturingRepo ───────────────────────────────────────────────

// fullCapturingRepo captures WHERE clauses from every read/write operation so
// row-filter injection can be verified independently of a real database.
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

func (r *fullCapturingRepo) FindUnique(_ context.Context, input *FindUniqueInput) (map[string]any, error) {
	r.capturedFindUniqueWhere = input.Where
	return map[string]any{"id": "x"}, nil
}

func (r *fullCapturingRepo) FindFirst(_ context.Context, input *FindFirstInput) (map[string]any, error) {
	r.capturedFindFirstWhere = input.Where
	return map[string]any{"id": "x"}, nil
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

type allowingRLSUsingGuard struct{}

func (g allowingRLSUsingGuard) ValidateInput(_ context.Context, _ string, _ Action, _ map[string]any) error {
	return nil
}

func (g allowingRLSUsingGuard) ResolveUsingFilter(_ context.Context, _ string, _ Action) (*RawSQLFilter, error) {
	return &RawSQLFilter{SQL: "owner_id = ?", Params: []any{"u_123"}}, nil
}

// ─── helper: taskModelWithoutOwner ───────────────────────────────────────────

// taskModelWithoutOwner returns a RuntimeModel that has NO FormatEndUserRef field.
func taskModelWithoutOwner() *RuntimeModel {
	return &RuntimeModel{
		Name:  "Task",
		Title: "任务",
		Fields: map[string]*RuntimeField{
			"id": {
				Name:      "id",
				Title:     "ID",
				Type:      &modeldesign.FieldType{Format: modeldesign.FormatUUID},
				IsPrimary: true,
				IsUnique:  true,
			},
			"title": {
				Name:  "title",
				Title: "标题",
				Type:  &modeldesign.FieldType{Format: modeldesign.FormatString},
			},
		},
	}
}

// ─── helper: permission builders ─────────────────────────────────────────────

func selectAllPerm() *ResolvedModelPermissions {
	return &ResolvedModelPermissions{
		Select: ActionPermission{Allowed: true},
		Insert: ActionPermission{Allowed: true},
		Update: ActionPermission{Allowed: true},
		Delete: ActionPermission{Allowed: true},
	}
}

func selfScopePerm() *ResolvedModelPermissions {
	return &ResolvedModelPermissions{
		Select: ActionPermission{Allowed: true, IsSelf: true},
		Insert: ActionPermission{Allowed: true, IsSelf: true},
		Update: ActionPermission{Allowed: true, IsSelf: true},
		Delete: ActionPermission{Allowed: true, IsSelf: true},
	}
}

func allScopePerm() *ResolvedModelPermissions {
	return &ResolvedModelPermissions{
		Select: ActionPermission{Allowed: true, IsSelf: false},
		Insert: ActionPermission{Allowed: true, IsSelf: false},
		Update: ActionPermission{Allowed: true, IsSelf: false},
		Delete: ActionPermission{Allowed: true, IsSelf: false},
	}
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

// ─── helpers ─────────────────────────────────────────────────────────────────

// hasPermissionError returns true when any error message contains "permission" (case-insensitive).
func hasPermissionError(result *graphql.Result) bool {
	for _, e := range result.Errors {
		if strings.Contains(strings.ToLower(e.Error()), "permission") {
			return true
		}
	}
	return false
}

// doQuery executes a raw GraphQL string against the built schema with the given context.
func doQuery(schema *graphql.Schema, ctx context.Context, query string) *graphql.Result {
	return graphql.Do(graphql.Params{
		Schema:        *schema,
		Context:       ctx,
		RequestString: query,
	})
}

// ─── Part 1: Action Gate ─────────────────────────────────────────────────────

// TestPermissionEnforcement_ActionGate_ZeroPerms verifies that an end-user with
// all permissions set to false receives a "permission" error for every operation.
func TestPermissionEnforcement_ActionGate_ZeroPerms(t *testing.T) {
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)

	zeroPerm := &ResolvedModelPermissions{} // all false

	operations := []struct {
		name  string
		query string
	}{
		{
			name:  "findMany",
			query: `{ findMany { items { id title } } }`,
		},
		{
			name:  "findUnique",
			query: `{ findUnique(where: { id: "x" }) { item { id } } }`,
		},
		{
			name:  "findFirst",
			query: `{ findFirst { item { id } } }`,
		},
		{
			name:  "aggregate",
			query: `{ aggregate(_count: { _all: true }) { _count { _all } } }`,
		},
		{
			name:  "count",
			query: `{ count { count } }`,
		},
		{
			name:  "create",
			query: `mutation { create(data: { title: "t" }) { id } }`,
		},
		{
			name:  "update",
			query: `mutation { update(where: { id: "x" }, data: { title: "t" }) { success } }`,
		},
		{
			name:  "delete",
			query: `mutation { delete(where: { id: "x" }) { success } }`,
		},
		{
			name:  "createMany",
			query: `mutation { createMany(data: [{ title: "t" }]) { count } }`,
		},
		{
			name:  "updateMany",
			query: `mutation { updateMany(where: {}, data: { title: "t" }, take: 1) { count } }`,
		},
		{
			name:  "deleteMany",
			query: `mutation { deleteMany(where: {}, take: 1) { count } }`,
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			repo := &fullCapturingRepo{}
			ctx := WithGraphqlRequestContext(
				context.Background(), repo, "org-1", "proj-1", "user-abc", "",
				zeroPerm,
			)

			result := doQuery(schema, ctx, op.query)
			assert.True(
				t,
				len(result.Errors) > 0,
				"expected permission error for %s, got no errors", op.name,
			)
			assert.True(
				t,
				hasPermissionError(result),
				"expected 'permission' in error message for %s, got: %v", op.name, result.Errors,
			)
		})
	}
}

// TestPermissionEnforcement_ActionGate_TenantAdmin verifies that a tenant admin
// (nil EndUserPerms) is never blocked by the action gate.
func TestPermissionEnforcement_ActionGate_TenantAdmin(t *testing.T) {
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)

	operations := []struct {
		name  string
		query string
	}{
		{
			name:  "findMany",
			query: `{ findMany { items { id title } } }`,
		},
		{
			name:  "findUnique",
			query: `{ findUnique(where: { id: "x" }) { item { id } } }`,
		},
		{
			name:  "findFirst",
			query: `{ findFirst { item { id } } }`,
		},
		{
			name:  "aggregate",
			query: `{ aggregate(_count: { _all: true }) { _count { _all } } }`,
		},
		{
			name:  "count",
			query: `{ count { count } }`,
		},
		{
			name:  "create",
			query: `mutation { create(data: { title: "t" }) { id } }`,
		},
		{
			name:  "update",
			query: `mutation { update(where: { id: "x" }, data: { title: "t" }) { success } }`,
		},
		{
			name:  "delete",
			query: `mutation { delete(where: { id: "x" }) { success } }`,
		},
		{
			name:  "createMany",
			query: `mutation { createMany(data: [{ title: "t" }]) { count } }`,
		},
		{
			name:  "updateMany",
			query: `mutation { updateMany(where: {}, data: { title: "t" }, take: 1) { count } }`,
		},
		{
			name:  "deleteMany",
			query: `mutation { deleteMany(where: {}, take: 1) { count } }`,
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			repo := &fullCapturingRepo{}
			// nil EndUserPerms = tenant admin
			ctx := WithGraphqlRequestContext(
				context.Background(), repo, "org-1", "proj-1", "", "",
				nil,
			)

			result := doQuery(schema, ctx, op.query)
			assert.False(
				t,
				hasPermissionError(result),
				"tenant admin must not receive permission error for %s, got: %v", op.name, result.Errors,
			)
		})
	}
}

// TestPermissionEnforcement_ActionGate_AllowedPerm verifies that an end-user
// with the correct permission set can execute the corresponding operation.
func TestPermissionEnforcement_ActionGate_AllowedPerm(t *testing.T) {
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)

	t.Run("findMany allowed when Select.Allowed=true", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", "user-abc", "",
			selectAllPerm(),
		)

		result := doQuery(schema, ctx, `{ findMany { items { id title } } }`)
		assert.False(t, hasPermissionError(result),
			"findMany must succeed when Select.Allowed=true, got errors: %v", result.Errors)
	})

	t.Run("create allowed when Insert.Allowed=true", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", "user-abc", "",
			selectAllPerm(),
		)

		result := doQuery(schema, ctx, `mutation { create(data: { title: "t" }) { id } }`)
		assert.False(t, hasPermissionError(result),
			"create must succeed when Insert.Allowed=true, got errors: %v", result.Errors)
	})
}

// ─── Part 2: Row Filter ───────────────────────────────────────────────────────

// TestPermissionEnforcement_RowFilter_SelfScope verifies that SELF scope injects
// "owner = $endUserID" into WHERE for every relevant read/write operation.
func TestPermissionEnforcement_RowFilter_SelfScope(t *testing.T) {
	const endUserID = "user-abc"
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)
	perms := selfScopePerm()

	t.Run("findMany injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		result := doQuery(schema, ctx, `{ findMany { items { id title } } }`)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedFindManyWhere["owner"],
			"findMany WHERE must contain owner=%s", endUserID)
	})

	t.Run("listByCursor preserves where and injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		q := `{ listByCursor(where: { title: { contains: "foo" } },` +
			` sortField: "id", sortDirection: "asc", limit: 20) { items { id } } }`
		result := doQuery(schema, ctx, q)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedListByCursorWhere["owner"],
			"listByCursor WHERE must contain owner=%s", endUserID)
		titleCond, ok := repo.capturedListByCursorWhere["title"].(map[string]any)
		require.True(t, ok, "listByCursor WHERE must preserve user title condition")
		assert.Equal(t, "foo", titleCond["contains"])
	})

	t.Run("listByPage preserves where and injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		q := `{ listByPage(where: { title: { contains: "foo" } },` +
			` orderBy: [{ id: asc }], pageIndex: 1, pageSize: 20) { items { id } total } }`
		result := doQuery(schema, ctx, q)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedListByPageWhere["owner"],
			"listByPage WHERE must contain owner=%s", endUserID)
		titleCond, ok := repo.capturedListByPageWhere["title"].(map[string]any)
		require.True(t, ok, "listByPage WHERE must preserve user title condition")
		assert.Equal(t, "foo", titleCond["contains"])
	})

	t.Run("findUnique injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		result := doQuery(schema, ctx, `{ findUnique(where: { id: "x" }) { item { id } } }`)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedFindUniqueWhere["owner"],
			"findUnique WHERE must contain owner=%s", endUserID)
	})

	t.Run("findFirst injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		result := doQuery(schema, ctx, `{ findFirst { item { id } } }`)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedFindFirstWhere["owner"],
			"findFirst WHERE must contain owner=%s", endUserID)
	})

	t.Run("update injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		result := doQuery(schema, ctx,
			`mutation { update(where: { id: "x" }, data: { title: "t" }) { success } }`)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedUpdateOneWhere["owner"],
			"update WHERE must contain owner=%s", endUserID)
	})

	t.Run("delete injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		result := doQuery(schema, ctx,
			`mutation { delete(where: { id: "x" }) { success } }`)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedDeleteOneWhere["owner"],
			"delete WHERE must contain owner=%s", endUserID)
	})

	t.Run("updateMany injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		result := doQuery(schema, ctx,
			`mutation { updateMany(where: {}, data: { title: "t" }, take: 1) { count } }`)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedUpdateManyWhere["owner"],
			"updateMany WHERE must contain owner=%s", endUserID)
	})

	t.Run("deleteMany injects owner", func(t *testing.T) {
		repo := &fullCapturingRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", endUserID, "", perms,
		)
		result := doQuery(schema, ctx,
			`mutation { deleteMany(where: {}, take: 1) { count } }`)
		require.Empty(t, result.Errors)
		assert.Equal(t, endUserID, repo.capturedDeleteManyWhere["owner"],
			"deleteMany WHERE must contain owner=%s", endUserID)
	})
}

// TestPermissionEnforcement_RowFilter_AllScope verifies that ALL scope (IsSelf=false)
// does NOT inject an owner filter.
func TestPermissionEnforcement_RowFilter_AllScope(t *testing.T) {
	const endUserID = "user-abc"
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)

	repo := &fullCapturingRepo{}
	ctx := WithGraphqlRequestContext(
		context.Background(), repo, "org-1", "proj-1", endUserID, "",
		allScopePerm(),
	)

	result := doQuery(schema, ctx, `{ findMany { items { id title } } }`)
	require.Empty(t, result.Errors)
	_, ownerPresent := repo.capturedFindManyWhere["owner"]
	assert.False(t, ownerPresent,
		"ALL scope must NOT inject owner into findMany WHERE")
}

// TestPermissionEnforcement_RowFilter_TenantAdmin verifies that a tenant admin
// (nil EndUserPerms) does NOT receive an owner filter injection.
func TestPermissionEnforcement_RowFilter_TenantAdmin(t *testing.T) {
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)

	repo := &fullCapturingRepo{}
	// nil EndUserPerms = tenant admin
	ctx := WithGraphqlRequestContext(
		context.Background(), repo, "org-1", "proj-1", "", "",
		nil,
	)

	result := doQuery(schema, ctx, `{ findMany { items { id title } } }`)
	require.Empty(t, result.Errors)
	_, ownerPresent := repo.capturedFindManyWhere["owner"]
	assert.False(t, ownerPresent,
		"tenant admin must NOT have owner injected into findMany WHERE")
}

// TestPermissionEnforcement_RowFilter_NoEndUserRefField verifies that SELF scope
// does NOT inject anything when the model has no FormatEndUserRef field.
func TestPermissionEnforcement_RowFilter_NoEndUserRefField(t *testing.T) {
	const endUserID = "user-abc"
	// Model without an owner field
	model := taskModelWithoutOwner()
	schema := buildSchemaFor(t, model)

	repo := &fullCapturingRepo{}
	ctx := WithGraphqlRequestContext(
		context.Background(), repo, "org-1", "proj-1", endUserID, "",
		selfScopePerm(),
	)

	result := doQuery(schema, ctx, `{ findMany { items { id title } } }`)
	require.Empty(t, result.Errors)
	_, ownerPresent := repo.capturedFindManyWhere["owner"]
	assert.False(t, ownerPresent,
		"SELF scope must NOT inject owner when model has no EndUserRef field")
}
