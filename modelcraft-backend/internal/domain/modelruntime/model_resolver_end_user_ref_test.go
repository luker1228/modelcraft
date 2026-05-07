package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"testing"

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
			context.Background(), repo, "org-1", "proj-1", "end-user-uuid-123",
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
			context.Background(), repo, "org-1", "proj-1", "",
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
			context.Background(), repo, "org-1", "proj-1", "end-user-uuid-123",
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
			context.Background(), repo, "org-1", "proj-1", "end-user-uuid-123",
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
			context.Background(), repo, "org-1", "proj-1", "",
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
