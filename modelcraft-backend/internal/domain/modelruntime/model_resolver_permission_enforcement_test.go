package modelruntime

import (
	"context"
	"strings"
	"testing"

	"modelcraft/internal/domain/modeldesign"

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
		Policies: []ResolvedPolicy{
			{Action: ActionSelect},
			{Action: ActionInsert},
			{Action: ActionUpdate},
			{Action: ActionDelete},
		},
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// hasPermissionError returns true when any error message contains "permission" (case-insensitive).
// It checks both GraphQL errors and structured errors in the data payload.
func hasPermissionError(result *graphql.Result) bool {
	// Check GraphQL errors
	for _, e := range result.Errors {
		if strings.Contains(strings.ToLower(e.Error()), "permission") {
			return true
		}
	}
	// Check structured errors in data payload
	if result.Data != nil {
		if dataMap, ok := result.Data.(map[string]interface{}); ok {
			// Iterate through all operations (findUnique, create, update, delete, etc.)
			for _, v := range dataMap {
				if fieldMap, ok := v.(map[string]interface{}); ok {
					if errField, ok := fieldMap["error"].(map[string]interface{}); ok {
						if typename, ok := errField["__typename"].(string); ok {
							if typename == "PermissionDenied" {
								return true
							}
						}
						if msg, ok := errField["message"].(string); ok {
							if strings.Contains(strings.ToLower(msg), "permission") {
								return true
							}
						}
					}
				}
			}
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
			query: `{ findUnique(where: { id: "x" }) { item { id } error { __typename ... on PermissionDenied { message } } } }`,
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
			query: `mutation { create(data: { title: "t" }) { id error { __typename ... on PermissionDenied { message } } } }`,
		},
		{
			name:  "update",
			query: `mutation { update(where: { id: "x" }, data: { title: "t" }) { success error { __typename ... on PermissionDenied { message } } } }`,
		},
		{
			name:  "delete",
			query: `mutation { delete(where: { id: "x" }) { success error { __typename ... on PermissionDenied { message } } } }`,
		},
		{
			name:  "createMany",
			query: `mutation { createMany(data: [{ title: "t" }]) { count error { __typename ... on PermissionDenied { message } } } }`,
		},
		{
			name:  "updateMany",
			query: `mutation { updateMany(where: {}, data: { title: "t" }, take: 1) { count error { __typename ... on PermissionDenied { message } } } }`,
		},
		{
			name:  "deleteMany",
			query: `mutation { deleteMany(where: {}, take: 1) { count error { __typename ... on PermissionDenied { message } } } }`,
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
			// Check for either GraphQL errors or structured errors in data payload
			hasErrors := len(result.Errors) > 0 || hasPermissionError(result)
			assert.True(
				t,
				hasErrors,
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

// Part 2: Row Filter — removed. RLS USING/CHECK handles all row-level scoping.

// taskModelWithOwner returns a RuntimeModel that has an END_USER_REF "owner" field.
func taskModelWithOwner() *RuntimeModel {
	return &RuntimeModel{
		Name:  "TaskWithOwner",
		Title: "任务（含所有者）",
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
			"owner": {
				Name:  "owner",
				Title: "所有者",
				Type:  &modeldesign.FieldType{Format: modeldesign.FormatEndUserRef},
			},
		},
	}
}

// buildSchemaFor builds a real GraphQL schema from the given RuntimeModel.
func buildSchemaFor(t *testing.T, model *RuntimeModel) *graphql.Schema {
	t.Helper()
	resolver := newGraphqlModelResolver(context.Background(), model, nil, nil)
	schema, err := resolver.newGraphqlSchema(context.Background())
	require.NoError(t, err)
	return schema
}
