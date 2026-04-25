package projectgraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/ctxutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appEnduser "modelcraft/internal/app/enduser"
)

func TestListProjectEndUsers_MissingProjectContext(t *testing.T) {
	SetEndUserManagementAppService(&appEnduser.EndUserManagementAppService{})
	t.Cleanup(func() { SetEndUserManagementAppService(nil) })

	resolver := &queryResolver{&Resolver{}}
	ctx := ctxutils.SetOrgName(context.Background(), "org-a")

	payload, err := resolver.ListProjectEndUsers(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, payload)

	clusterNotFound, ok := payload.Error.(*generated.ClusterNotFound)
	require.True(t, ok)
	assert.Equal(t, "projectSlug not found in context", clusterNotFound.Message)
}

func TestGrantEndUserProjectAccess_ParamValidation(t *testing.T) {
	resolver := &mutationResolver{&Resolver{}}

	payload, err := resolver.GrantEndUserProjectAccess(context.Background(), generated.GrantEndUserProjectAccessInput{
		EndUserID:          "",
		PermissionBundleID: "bundle-1",
	})
	require.NoError(t, err)
	require.NotNil(t, payload)

	invalidInput, ok := payload.Error.(*generated.InvalidInput)
	require.True(t, ok)
	assert.Equal(t, "endUserId is required", invalidInput.Message)
}

func TestGrantEndUserProjectAccess_ServiceNotInitialized(t *testing.T) {
	resolver := &mutationResolver{&Resolver{}}

	payload, err := resolver.GrantEndUserProjectAccess(context.Background(), generated.GrantEndUserProjectAccessInput{
		EndUserID:          "end-user-1",
		PermissionBundleID: "bundle-1",
	})
	require.NoError(t, err)
	require.NotNil(t, payload)

	invalidInput, ok := payload.Error.(*generated.InvalidInput)
	require.True(t, ok)
	assert.Equal(t, "end-user access service not initialized", invalidInput.Message)
}

func TestGrantEndUserProjectAccess_MissingContext(t *testing.T) {
	resolver := &mutationResolver{&Resolver{EndUserAccessAppService: &appEnduser.EndUserProjectAccessAppService{}}}

	payload, err := resolver.GrantEndUserProjectAccess(context.Background(), generated.GrantEndUserProjectAccessInput{
		EndUserID:          "end-user-1",
		PermissionBundleID: "bundle-1",
	})
	require.NoError(t, err)
	require.NotNil(t, payload)

	invalidInput, ok := payload.Error.(*generated.InvalidInput)
	require.True(t, ok)
	assert.Contains(t, invalidInput.Message, "failed to get organization name from context")
}

func TestListProjectEndUserAccess_ParamValidation(t *testing.T) {
	resolver := &queryResolver{&Resolver{EndUserAccessAppService: &appEnduser.EndUserProjectAccessAppService{}}}
	ctx := ctxutils.SetProjectSlug(ctxutils.SetOrgName(context.Background(), "org-a"), "project-a")
	first := int32(0)

	payload, err := resolver.ListProjectEndUserAccess(ctx, &generated.ListProjectEndUserAccessInput{First: &first})
	require.NoError(t, err)
	require.NotNil(t, payload)

	invalidInput, ok := payload.Error.(*generated.InvalidInput)
	require.True(t, ok)
	assert.Equal(t, "first must be greater than 0", invalidInput.Message)
}
