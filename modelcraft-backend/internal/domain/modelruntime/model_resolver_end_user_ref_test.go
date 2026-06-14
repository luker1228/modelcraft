package modelruntime

import (
	"context"
	"testing"

	"modelcraft/internal/domain/modeldesign"

	"github.com/graphql-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturingClientRepo wraps mockClientDatabaseRepository and captures the last CreateOne call.
type capturingClientRepo struct {
	mockClientDatabaseRepository
	capturedCreateInput     *CreateOneInput
	capturedCreateManyInput *CreateManyInput
	capturedUpdateOneInput  *UpdateOneInput
}

func (c *capturingClientRepo) CreateOne(_ context.Context, input *CreateOneInput) (string, error) {
	c.capturedCreateInput = input
	return "new-record-id", nil
}

func (c *capturingClientRepo) CreateMany(_ context.Context, input *CreateManyInput) (interface{}, error) {
	c.capturedCreateManyInput = input
	return map[string]any{"count": len(input.Data)}, nil
}

func (c *capturingClientRepo) UpdateOne(_ context.Context, input *UpdateOneInput) (map[string]any, error) {
	c.capturedUpdateOneInput = input
	return map[string]any{"id": "record-id"}, nil
}

// taskModelWithOwner is a RuntimeModel that has an END_USER_REF "owner" field.
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

// TestEndUserRefOwnerInjection_ViaRealResolver 通过真实的 GraphQL schema 执行路径验证
// END_USER_REF 字段的 owner 注入逻辑：
//   - EndUser 上下文：context 中的 endUserID 强制覆盖客户端提交的 owner 值（防止 spoofing）
//   - Tenant admin 上下文（无 EndUser 身份）：客户端提交的 owner 值原样使用
//
// 如果删除 executeCreateOne 中的注入块，这两个测试均会失败 ——
// 第一个用例断言的 "end-user-uuid-123" 将变成 "attacker-uuid"。
func TestEndUserRefOwnerInjection_ViaRealResolver(t *testing.T) {
	model := taskModelWithOwner()

	t.Run("injects owner from context, ignoring client-supplied value (EndUser context)", func(t *testing.T) {
		repo := &capturingClientRepo{}
		schema := buildSchemaFor(t, model)

		// Simulate EndUser request: CurrentEndUserID is set in rctx
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", "end-user-uuid-123", "",
			nil,
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			// Attacker tries to supply their own owner value
			RequestString: `mutation {
				create(data: { title: "hello", owner: "attacker-uuid" }) {
					id
				}
			}`,
		})

		require.Empty(t, result.Errors, "mutation must succeed without errors")
		require.NotNil(t, repo.capturedCreateInput, "CreateOne must have been called")
		assert.Equal(
			t,
			"end-user-uuid-123",
			repo.capturedCreateInput.Data["owner"],
			"owner must be force-injected from context endUserID, not the attacker-supplied value",
		)
	})

	t.Run("preserves client-supplied owner when no EndUser identity (tenant admin context)", func(t *testing.T) {
		repo := &capturingClientRepo{}
		schema := buildSchemaFor(t, model)

		// Simulate tenant admin request: CurrentEndUserID is empty
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", "", "",
			nil,
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				create(data: { title: "hello", owner: "chosen-end-user-uuid" }) {
					id
				}
			}`,
		})

		require.Empty(t, result.Errors, "mutation must succeed without errors")
		require.NotNil(t, repo.capturedCreateInput, "CreateOne must have been called")
		assert.Equal(
			t,
			"chosen-end-user-uuid",
			repo.capturedCreateInput.Data["owner"],
			"tenant admin may supply owner explicitly when no EndUser identity is present",
		)
	})
}

// TestEndUserRefOwnerInjection_CreateMany_ViaRealResolver verifies that the
// createMany mutation cannot be used to spoof the owner field in batch creates.
func TestEndUserRefOwnerInjection_CreateMany_ViaRealResolver(t *testing.T) {
	model := taskModelWithOwner()

	t.Run("injects owner from context in each createMany item (EndUser context)", func(t *testing.T) {
		repo := &capturingClientRepo{}
		schema := buildSchemaFor(t, model)

		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", "end-user-uuid-123", "",
			nil,
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				createMany(data: [
					{ title: "record1", owner: "attacker-uuid" }
				]) { count }
			}`,
		})

		require.Empty(t, result.Errors, "createMany mutation must succeed without errors")
		require.NotNil(t, repo.capturedCreateManyInput, "CreateMany must have been called")
		require.Len(t, repo.capturedCreateManyInput.Data, 1)
		assert.Equal(
			t,
			"end-user-uuid-123",
			repo.capturedCreateManyInput.Data[0]["owner"],
			"owner must be force-injected from context endUserID in every createMany item",
		)
	})
}

// TestEndUserRefOwnerInjection_UpdateOne_ViaRealResolver verifies that the
// updateOne mutation cannot be used to reassign the owner field.
func TestEndUserRefOwnerInjection_UpdateOne_ViaRealResolver(t *testing.T) {
	model := taskModelWithOwner()

	t.Run("injects owner from context, ignoring client-supplied value (EndUser context)", func(t *testing.T) {
		repo := &capturingClientRepo{}
		schema := buildSchemaFor(t, model)

		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", "end-user-uuid-123", "",
			nil,
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				update(where: { id: "some-record-id" }, data: { title: "new title", owner: "attacker-uuid" }) {
					success
				}
			}`,
		})

		require.Empty(t, result.Errors, "updateOne mutation must succeed without errors")
		require.NotNil(t, repo.capturedUpdateOneInput, "UpdateOne must have been called")
		assert.Equal(
			t,
			"end-user-uuid-123",
			repo.capturedUpdateOneInput.Data["owner"],
			"owner must be force-injected from context endUserID, not the attacker-supplied value",
		)
	})

	t.Run("preserves client-supplied owner when no EndUser identity (tenant admin context)", func(t *testing.T) {
		repo := &capturingClientRepo{}
		schema := buildSchemaFor(t, model)

		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", "", "",
			nil,
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				update(where: { id: "some-record-id" }, data: { title: "new title", owner: "chosen-end-user-uuid" }) {
					success
				}
			}`,
		})

		require.Empty(t, result.Errors, "mutation must succeed without errors")
		require.NotNil(t, repo.capturedUpdateOneInput, "UpdateOne must have been called")
		assert.Equal(
			t,
			"chosen-end-user-uuid",
			repo.capturedUpdateOneInput.Data["owner"],
			"tenant admin may supply owner explicitly when no EndUser identity is present",
		)
	})
}

func TestRLSInputCheck_CreateOne_DeniesBeforeRepoCall(t *testing.T) {
	repo := &capturingClientRepo{}
	schema := buildSchemaFor(t, taskModelWithOwner())
	ctx := WithGraphqlRequestContext(
		context.Background(), repo, "org-1", "project-1", "u_123", "",
		&ResolvedModelPermissions{Insert: ActionPermission{Allowed: true}},
	)

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

// ─── Owner SELF-scope enforcement (IsSelf=true) ──────────────────────────────

// TestOwnerSelfScopeEnforcement_CreateOne 验证 IsSelf=true 时 create mutation 的 owner 检查。
//
// READ_WRITE_OWNER 场景下：
//   - 调用方传了其他用户的 owner → 应返回 PermissionDenied（拒绝代他人创建）
//   - 调用方传了本人 owner    → 成功，写库的 owner 为当前用户 ID
//   - 调用方不传 owner       → 成功，自动注入当前用户 ID
//   - IsSelf=false 时传他人 owner → 成功（ALL scope 允许写任意 owner，再被注入覆盖）
func TestOwnerSelfScopeEnforcement_CreateOne(t *testing.T) {
	const (
		currentUser = "user-current-abc"
		otherUser   = "user-other-xyz"
	)
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)

	t.Run("IsSelf=true: 传他人 owner 时应返回 PermissionDenied", func(t *testing.T) {
		repo := &capturingClientRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", currentUser, "",
			selfScopePerm(),
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				create(data: { title: "x", owner: "` + otherUser + `" }) { id }
			}`,
		})

		require.NotEmpty(t, result.Errors, "should return permission error when owner != currentUser")
		assert.True(t, hasPermissionError(result),
			"error message must contain 'permission', got: %v", result.Errors)
		assert.Nil(t, repo.capturedCreateInput,
			"CreateOne must NOT be called when permission check fails")
	})

	t.Run("IsSelf=true: 传本人 owner 时应成功，写库 owner 为当前用户", func(t *testing.T) {
		repo := &capturingClientRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", currentUser, "",
			selfScopePerm(),
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				create(data: { title: "x", owner: "` + currentUser + `" }) { id }
			}`,
		})

		require.Empty(t, result.Errors, "should succeed when owner == currentUser")
		require.NotNil(t, repo.capturedCreateInput)
		assert.Equal(t, currentUser, repo.capturedCreateInput.Data["owner"])
	})

	t.Run("IsSelf=true: 不传 owner 时应成功，自动注入当前用户 ID", func(t *testing.T) {
		repo := &capturingClientRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", currentUser, "",
			selfScopePerm(),
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				create(data: { title: "x" }) { id }
			}`,
		})

		require.Empty(t, result.Errors, "should succeed when owner not provided")
		require.NotNil(t, repo.capturedCreateInput)
		assert.Equal(t, currentUser, repo.capturedCreateInput.Data["owner"],
			"owner must be auto-injected from context")
	})

	t.Run("IsSelf=false(ALL scope): 传他人 owner 时应成功（覆盖为当前用户）", func(t *testing.T) {
		repo := &capturingClientRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", currentUser, "",
			allScopePerm(),
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				create(data: { title: "x", owner: "` + otherUser + `" }) { id }
			}`,
		})

		require.Empty(t, result.Errors, "ALL scope must not block owner mismatch")
		require.NotNil(t, repo.capturedCreateInput)
		assert.Equal(t, currentUser, repo.capturedCreateInput.Data["owner"],
			"owner must be overridden to current user even in ALL scope")
	})
}

// TestOwnerSelfScopeEnforcement_CreateMany 验证 createMany 批量写入时的 owner 检查。
func TestOwnerSelfScopeEnforcement_CreateMany(t *testing.T) {
	const (
		currentUser = "user-current-abc"
		otherUser   = "user-other-xyz"
	)
	model := taskModelWithOwner()
	schema := buildSchemaFor(t, model)

	t.Run("IsSelf=true: 批量 create 中含他人 owner 时应返回 PermissionDenied", func(t *testing.T) {
		repo := &capturingClientRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", currentUser, "",
			selfScopePerm(),
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				createMany(data: [
					{ title: "r1", owner: "` + otherUser + `" }
				]) { count }
			}`,
		})

		require.NotEmpty(t, result.Errors, "createMany should return permission error")
		assert.True(t, hasPermissionError(result),
			"error must mention 'permission', got: %v", result.Errors)
		assert.Nil(t, repo.capturedCreateManyInput,
			"CreateMany must NOT be called when permission check fails")
	})

	t.Run("IsSelf=true: 批量 create 不传 owner 时应成功并注入当前用户", func(t *testing.T) {
		repo := &capturingClientRepo{}
		ctx := WithGraphqlRequestContext(
			context.Background(), repo, "org-1", "proj-1", currentUser, "",
			selfScopePerm(),
		)

		result := graphql.Do(graphql.Params{
			Schema:  *schema,
			Context: ctx,
			RequestString: `mutation {
				createMany(data: [
					{ title: "r1" },
					{ title: "r2" }
				]) { count }
			}`,
		})

		require.Empty(t, result.Errors)
		require.NotNil(t, repo.capturedCreateManyInput)
		for i, row := range repo.capturedCreateManyInput.Data {
			assert.Equal(t, currentUser, row["owner"],
				"createMany row[%d] owner must be injected from context", i)
		}
	})
}
