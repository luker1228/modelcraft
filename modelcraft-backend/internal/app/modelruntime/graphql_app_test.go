package modelruntime

import (
	"context"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"testing"

	domainmodeldesign "modelcraft/internal/domain/modeldesign"
	domainmodelruntime "modelcraft/internal/domain/modelruntime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRuntimeModelRepo struct {
	model *domainmodelruntime.RuntimeModel
	err   error
}

func (s *stubRuntimeModelRepo) GetByID(ctx context.Context, id string) (*domainmodelruntime.RuntimeModel, error) {
	return s.model, s.err
}

func (s *stubRuntimeModelRepo) GetByName(
	ctx context.Context,
	modelLocator *domainmodeldesign.ModelLocator,
) (*domainmodelruntime.RuntimeModel, error) {
	return s.model, s.err
}

func TestIsMutationOperation(t *testing.T) {
	t.Run("single mutation operation", func(t *testing.T) {
		isMutation, err := isMutationOperation("mutation { create_user(input: {name: \"A\"}) { id } }", "")
		require.NoError(t, err)
		assert.True(t, isMutation)
	})

	t.Run("single query operation", func(t *testing.T) {
		isMutation, err := isMutationOperation("query { users { id } }", "")
		require.NoError(t, err)
		assert.False(t, isMutation)
	})

	t.Run("multiple operations with operationName", func(t *testing.T) {
		query := `query ListUsers { users { id } }
		mutation CreateUser { create_user(input: {name: "A"}) { id } }`
		isMutation, err := isMutationOperation(query, "CreateUser")
		require.NoError(t, err)
		assert.True(t, isMutation)
	})
}

func TestGraphqlAppService_denyManagedModelMutation(t *testing.T) {
	ctx := ctxutils.NewHttpContext(context.Background(), &ctxutils.HttpRequestContext{RequestId: "req-1", Lang: "zh"})
	modelLocator, err := domainmodeldesign.NewModelLocator("test-org", "project-a", "db_1", "customer_orders")
	require.NoError(t, err)

	t.Run("managed model mutation should be denied", func(t *testing.T) {
		repo := &stubRuntimeModelRepo{
			model: &domainmodelruntime.RuntimeModel{
				OrgName:      "test-org",
				ProjectSlug:  "project-a",
				Name:         "customer_orders",
				DatabaseName: "db_1",
				CreatedVia:   domainmodeldesign.ModelCreationSourceImported,
			},
		}
		svc := &GraphqlAppService{modelRepo: repo}

		err = svc.denyManagedModelMutation(ctx, modelLocator, ExecuteGraphQLCommand{
			Query: "mutation { create_customer_orders(input: {name: \"A\"}) { id } }",
		})

		require.Error(t, err)
		var bizErr *bizerrors.BusinessError
		require.ErrorAs(t, err, &bizErr)
		assert.Equal(t, bizerrors.ManagedModelReadOnly.GetCode(), bizErr.Info().GetCode())
	})

	t.Run("managed model query should pass readonly guard", func(t *testing.T) {
		repo := &stubRuntimeModelRepo{
			model: &domainmodelruntime.RuntimeModel{
				OrgName:      "test-org",
				ProjectSlug:  "project-a",
				Name:         "customer_orders",
				DatabaseName: "db_1",
				CreatedVia:   domainmodeldesign.ModelCreationSourceImported,
			},
		}
		svc := &GraphqlAppService{modelRepo: repo}

		err = svc.denyManagedModelMutation(ctx, modelLocator, ExecuteGraphQLCommand{
			Query: "query { findMany_customer_orders { items { id } } }",
		})
		require.NoError(t, err)
	})

	t.Run("self-built model mutation should pass readonly guard", func(t *testing.T) {
		repo := &stubRuntimeModelRepo{
			model: &domainmodelruntime.RuntimeModel{
				OrgName:      "test-org",
				ProjectSlug:  "project-a",
				Name:         "customer_orders",
				DatabaseName: "db_1",
				CreatedVia:   domainmodeldesign.ModelCreationSourceNew,
			},
		}
		svc := &GraphqlAppService{modelRepo: repo}

		err = svc.denyManagedModelMutation(ctx, modelLocator, ExecuteGraphQLCommand{
			Query: "mutation { update_customer_orders(where:{id:\"1\"}, data:{name:\"A\"}) { id } }",
		})
		require.NoError(t, err)
	})
}

func TestGraphqlAppService_Execute_DenyManagedModelMutation(t *testing.T) {
	ctx := ctxutils.NewHttpContext(context.Background(), &ctxutils.HttpRequestContext{RequestId: "req-1", Lang: "zh"})
	repo := &stubRuntimeModelRepo{
		model: &domainmodelruntime.RuntimeModel{
			OrgName:      "test-org",
			ProjectSlug:  "project-a",
			Name:         "customer_orders",
			DatabaseName: "db_1",
			CreatedVia:   domainmodeldesign.ModelCreationSourceImported,
		},
	}
	svc := &GraphqlAppService{modelRepo: repo}

	_, err := svc.Execute(ctx, "test-org", "project-a", "customer_orders", "db_1", ExecuteGraphQLCommand{
		Query: "mutation { create_customer_orders(input: {name: \"A\"}) { id } }",
	})

	require.Error(t, err)
	var bizErr *bizerrors.BusinessError
	require.ErrorAs(t, err, &bizErr)
	assert.Equal(t, bizerrors.ManagedModelReadOnly.GetCode(), bizErr.Info().GetCode())
}
